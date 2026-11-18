package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/charmbracelet/log"
	"github.com/gravadigital/openparallax/api/internal/logger"
)

// MaxPresignExpiry is the longest a presigned URL may live: SigV4 caps it at one week and
// Garage enforces it.
const MaxPresignExpiry = 7 * 24 * time.Hour

// S3Config is what S3Storage needs to reach the bucket.
type S3Config struct {
	// Endpoint is the URL the api uses, with scheme (http://storage:3900).
	Endpoint string
	// PresignEndpoint is the URL presigned links are signed with. The signature covers the
	// host, so it must be the one whoever opens the link uses. Empty means Endpoint.
	PresignEndpoint string
	// Region must match the server's s3_region ("garage" by default in Garage).
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
}

// S3Storage implements FileStorage over any S3-compatible service (Garage in every
// environment of OpenParallax).
type S3Storage struct {
	client    *s3.Client
	presigner *s3.PresignClient
	bucket    string
	log       *log.Logger
}

func newS3Client(endpoint string, cfg S3Config, extra ...func(*s3.Options)) *s3.Client {
	opts := s3.Options{
		BaseEndpoint: aws.String(endpoint),
		Region:       cfg.Region,
		// Path style: virtual-host style would need a wildcard DNS for the bucket.
		UsePathStyle: true,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
	}
	for _, f := range extra {
		f(&opts)
	}
	return s3.New(opts)
}

// NewS3Storage connects to the bucket and checks it exists. It never creates it: a bucket
// created through the S3 API gets only a local alias in Garage, invisible to other keys. The
// infrastructure creates it (deploy/docker-compose.yml, or `garage bucket create`).
func NewS3Storage(ctx context.Context, cfg S3Config) (*S3Storage, error) {
	if cfg.Endpoint == "" || cfg.Bucket == "" || cfg.Region == "" {
		return nil, errors.New("S3 endpoint, bucket and region are required")
	}
	if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return nil, errors.New("S3 credentials not configured")
	}

	presignEndpoint := cfg.PresignEndpoint
	if presignEndpoint == "" {
		presignEndpoint = cfg.Endpoint
	}
	signer := newS3Client(presignEndpoint, cfg, func(o *s3.Options) {
		// Without it the SDK embeds a checksum of an empty body in presigned PUTs and the
		// real upload gets 400 InvalidDigest.
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
	})

	st := &S3Storage{
		client:    newS3Client(cfg.Endpoint, cfg),
		presigner: s3.NewPresignClient(signer),
		bucket:    cfg.Bucket,
		log:       logger.Handler("s3-storage"),
	}

	// Fail fast. A region that does not match s3_region shows up here as a 400 with no body.
	if _, err := st.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(cfg.Bucket)}); err != nil {
		return nil, fmt.Errorf("bucket %q not reachable at %s (region %q): %w", cfg.Bucket, cfg.Endpoint, cfg.Region, err)
	}

	st.log.Info("S3 storage initialized", "endpoint", cfg.Endpoint, "bucket", cfg.Bucket, "region", cfg.Region)
	return st, nil
}

// Put stores a file. The content type is always sent: Garage does not infer it from the
// extension, and an object without one is served without Content-Type.
func (s *S3Storage) Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (string, error) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	// Over plain HTTP the SDK hashes the payload to sign it, which needs a seekable body.
	body, ok := reader.(io.ReadSeeker)
	if !ok {
		data, err := io.ReadAll(reader)
		if err != nil {
			return "", fmt.Errorf("failed to read upload: %w", err)
		}
		body, size = bytes.NewReader(data), int64(len(data))
	}

	out, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	s.log.Debug("file stored successfully", "key", key, "size", size, "etag", aws.ToString(out.ETag), "bucket", s.bucket)
	return key, nil
}

// Get retrieves a file. The caller closes it.
func (s *S3Storage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("object not found: %s: %w", key, err)
		}
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	return out.Body, nil
}

// Delete removes a file. Deleting a key that does not exist is not an error (S3 semantics).
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	if _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)}); err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}
	s.log.Debug("file deleted successfully", "key", key, "bucket", s.bucket)
	return nil
}

// GetURL returns a presigned GET link, valid for the longest time S3 allows. It stops working
// at once if the signing key loses read on the bucket or is deleted.
func (s *S3Storage) GetURL(ctx context.Context, key string) (string, error) {
	req, err := s.presigner.PresignGetObject(ctx,
		&s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)},
		s3.WithPresignExpires(MaxPresignExpiry))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return req.URL, nil
}

// Exists checks if a file exists.
func (s *S3Storage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object: %w", err)
	}
	return true, nil
}

// GetInfo returns metadata about a stored file.
func (s *S3Storage) GetInfo(ctx context.Context, key string) (*FileInfo, error) {
	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("file not found: %s", key)
		}
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	info := &FileInfo{
		Key:         key,
		Size:        aws.ToInt64(out.ContentLength),
		ContentType: aws.ToString(out.ContentType),
		ETag:        aws.ToString(out.ETag),
	}
	if out.LastModified != nil {
		info.LastModified = out.LastModified.Format(time.RFC3339)
	}
	return info, nil
}

// isNotFound covers both shapes a missing object takes: a typed NoSuchKey/NotFound when the
// response has a body (GET) and a bare 404 when it has none (HEAD).
func isNotFound(err error) bool {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchKey", "NotFound":
			return true
		}
	}
	var respErr *smithyhttp.ResponseError
	return errors.As(err, &respErr) && respErr.HTTPStatusCode() == http.StatusNotFound
}
