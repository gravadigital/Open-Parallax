//go:build integration

package storage

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Same version as deploy/docker-compose.yml.
const garageImage = "dxflrs/garage:v2.4.1"

const testGarageToml = `metadata_dir = "/var/lib/garage/meta"
data_dir = "/var/lib/garage/data"
db_engine = "sqlite"
replication_factor = 1
rpc_bind_addr = "[::]:3901"
rpc_public_addr = "127.0.0.1:3901"
rpc_secret = "0000000000000000000000000000000000000000000000000000000000000000"

[s3_api]
s3_region = "garage"
api_bind_addr = "[::]:3900"

[admin]
api_bind_addr = "[::]:3903"
`

const (
	testBucket    = "openparallax-test"
	testAccessKey = "GKopenparallaxtest01"
	testSecretKey = "openparallax-test-secret-0123456789"
)

// startGarage runs a single-node Garage that creates testBucket on start: by the time /health
// answers 200 the bucket exists. It returns the S3 endpoint.
func startGarage(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: garageImage,
			Cmd:   []string{"/garage", "server", "--single-node", "--default-bucket"},
			Env: map[string]string{
				"GARAGE_DEFAULT_ACCESS_KEY": testAccessKey,
				"GARAGE_DEFAULT_SECRET_KEY": testSecretKey,
				"GARAGE_DEFAULT_BUCKET":     testBucket,
			},
			Files: []testcontainers.ContainerFile{{
				Reader: strings.NewReader(testGarageToml), ContainerFilePath: "/etc/garage.toml", FileMode: 0o644,
			}},
			ExposedPorts: []string{"3900/tcp", "3903/tcp"},
			WaitingFor:   wait.ForHTTP("/health").WithPort("3903/tcp").WithStartupTimeout(90 * time.Second),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Terminate(context.Background()) })

	host, err := c.Host(ctx)
	require.NoError(t, err)
	port, err := c.MappedPort(ctx, "3900/tcp")
	require.NoError(t, err)
	return "http://" + host + ":" + port.Port()
}

func testConfig(endpoint string) S3Config {
	return S3Config{
		Endpoint:        endpoint,
		Region:          "garage",
		Bucket:          testBucket,
		AccessKeyID:     testAccessKey,
		SecretAccessKey: testSecretKey,
	}
}

func TestS3Storage_Garage(t *testing.T) {
	endpoint := startGarage(t)
	ctx := context.Background()

	st, err := NewS3Storage(ctx, testConfig(endpoint))
	require.NoError(t, err)

	const key = "event_participant_1.pdf"
	content := "%PDF-1.4 propuesta de prueba"

	t.Run("Put y Get devuelven el mismo contenido", func(t *testing.T) {
		got, err := st.Put(ctx, key, strings.NewReader(content), int64(len(content)), "application/pdf")
		require.NoError(t, err)
		assert.Equal(t, key, got)

		r, err := st.Get(ctx, key)
		require.NoError(t, err)
		defer r.Close()
		body, err := io.ReadAll(r)
		require.NoError(t, err)
		assert.Equal(t, content, string(body))
	})

	t.Run("Put acepta un reader que no es seekable", func(t *testing.T) {
		const k = "no-seekable.txt"
		_, err := st.Put(ctx, k, io.NopCloser(strings.NewReader("hola")), 4, "text/plain")
		require.NoError(t, err)
		info, err := st.GetInfo(ctx, k)
		require.NoError(t, err)
		assert.EqualValues(t, 4, info.Size)
	})

	t.Run("GetInfo conserva el content type", func(t *testing.T) {
		info, err := st.GetInfo(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, "application/pdf", info.ContentType)
		assert.EqualValues(t, len(content), info.Size)
		assert.NotEmpty(t, info.ETag)
		assert.NotEmpty(t, info.LastModified)
	})

	t.Run("Exists distingue presente de ausente", func(t *testing.T) {
		ok, err := st.Exists(ctx, key)
		require.NoError(t, err)
		assert.True(t, ok)

		ok, err = st.Exists(ctx, "no-existe.pdf")
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("Get y GetInfo de una key ausente fallan", func(t *testing.T) {
		_, err := st.Get(ctx, "no-existe.pdf")
		assert.Error(t, err)
		_, err = st.GetInfo(ctx, "no-existe.pdf")
		assert.Error(t, err)
	})

	t.Run("GetURL firma un link de 7 dias que se puede abrir sin credenciales", func(t *testing.T) {
		url, err := st.GetURL(ctx, key)
		require.NoError(t, err)
		assert.Contains(t, url, "X-Amz-Expires=604800")

		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, content, string(body))
	})

	t.Run("Garage no permite lectura anonima por la API S3", func(t *testing.T) {
		resp, err := http.Get(endpoint + "/" + testBucket + "/" + key)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Delete borra y no falla si ya no existe", func(t *testing.T) {
		require.NoError(t, st.Delete(ctx, key))
		ok, err := st.Exists(ctx, key)
		require.NoError(t, err)
		assert.False(t, ok)
		assert.NoError(t, st.Delete(ctx, key))
	})
}

func TestS3Storage_GarageArranqueFallaRapido(t *testing.T) {
	endpoint := startGarage(t)
	ctx := context.Background()

	t.Run("con una region distinta de s3_region", func(t *testing.T) {
		cfg := testConfig(endpoint)
		cfg.Region = "us-east-1"
		_, err := NewS3Storage(ctx, cfg)
		assert.Error(t, err)
	})

	t.Run("si el bucket no existe, sin crearlo", func(t *testing.T) {
		cfg := testConfig(endpoint)
		cfg.Bucket = "no-existe"
		_, err := NewS3Storage(ctx, cfg)
		require.Error(t, err)

		cfg = testConfig(endpoint)
		cfg.Bucket = "no-existe"
		_, err = NewS3Storage(ctx, cfg)
		assert.Error(t, err, "el primer intento no tiene que haber creado el bucket")
	})
}
