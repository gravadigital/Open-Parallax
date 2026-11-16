---
id: file-storage
display_name: Almacenamiento de archivos (local / S3 con Garage)
language: golang
description: Pluggable object storage behind a FileStorage interface, backed by the local filesystem or an S3-compatible service (Garage)
applies_to: [api]
required_by: []
package: github.com/aws/aws-sdk-go-v2/service/s3
---

# File Storage (api)

**No existe en el catálogo.** Es una preocupación propia de este servicio: las propuestas
que suben los participantes se guardan como archivos, y el backend de almacenamiento se
elige por configuración.

## La interfaz

`internal/storage/file_storage.go` define el puerto. Todo lo que toque archivos depende de
esta interfaz, nunca de una implementación concreta:

```go
type FileStorage interface {
    Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (string, error)
    Get(ctx context.Context, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, key string) error
    GetURL(ctx context.Context, key string) (string, error)
    Exists(ctx context.Context, key string) (bool, error)
    GetInfo(ctx context.Context, key string) (*FileInfo, error)
}
```

Es de las pocas partes del servicio que sí recibe `context.Context`. Mantenelo.

## Implementaciones

| Provider | Tipo | Cuándo |
|---|---|---|
| `local` | `LocalStorage` — filesystem bajo `STORAGE_LOCAL_PATH` | Desarrollo |
| `s3` | `S3Storage` — `aws-sdk-go-v2` contra Garage | Stack local y servidores |

Se construye en `main.go` con `storage.NewFileStorage(cfg)`, que despacha según
`STORAGE_PROVIDER` y falla si el valor no es ninguno de los dos. El valor `minio` ya no
existe: falla con un mensaje que indica pasar a `s3`.

### Reglas del provider `s3`

El servidor es [Garage](https://garagehq.deuxfleurs.fr/), que reemplazó a MinIO porque MinIO
dejó de publicar imágenes y su repositorio quedó archivado. El cliente es el SDK de AWS: no se
usa `minio-go`, que pertenece a un proyecto archivado.

- **La API no crea el bucket.** `NewS3Storage` hace `HeadBucket` al arrancar y aborta si el
  bucket no responde (fail-fast). En Garage, un bucket creado con `CreateBucket` desde la API
  S3 queda con alias local: no lo ven otras claves ni se puede publicar. Lo crea la
  infraestructura: `--default-bucket` en el compose local y `garage bucket create` en un
  servidor.
- **La región tiene que coincidir con `s3_region` del servidor** (`garage` por defecto). Si no
  coincide, toda petición falla con 400 y `HeadBucket` lo muestra sin cuerpo, como
  `UnknownError`. Por eso la región sale de `S3_REGION` y nunca va fija en el código.
- **Path style siempre** (`UsePathStyle`). El estilo virtual-host necesitaría un DNS wildcard.
- **`Put` envía siempre el `Content-Type`.** Garage no lo deduce de la extensión. Si llega
  vacío, se usa `application/octet-stream`.
- **Sobre HTTP, el cuerpo tiene que ser seekable**, porque el SDK lo hashea para firmar.
  `multipart.File` lo es. Si llega otro reader, `Put` lo lee entero en memoria. El límite es
  `MAX_FILE_SIZE`.
- **Las URLs prefirmadas se firman con `S3_PRESIGN_ENDPOINT`.** La firma incluye el host, y el
  cliente que firma desactiva el checksum automático (`RequestChecksumCalculation:
  WhenRequired`). Sin eso, un PUT prefirmado recibe `400 InvalidDigest`. Vencen como máximo a
  los 7 días, y dejan de funcionar si la clave que las firmó pierde `read` o se borra.
- **Lo que Garage no soporta no se usa:** bucket policies, ACL, versionado, tags, object lock
  ni notificaciones. Un `If-None-Match` en un PUT **se ignora en silencio**, así que no sirve
  como lock.

## Claves

La clave se arma en el handler de upload, no en el storage:

```go
ext := filepath.Ext(filepath.Base(header.Filename))
secureFilename := fmt.Sprintf("%s_%s_%d%s", eventID, participantID, time.Now().Unix(), ext)
storageKey, err := h.fileStorage.Put(ctx, secureFilename, file, header.Size, contentType)
```

El nombre original **nunca** se usa como clave: se guarda aparte en
`attachments.original_name` y se devuelve en el `Content-Disposition` de la descarga. La
clave que devuelve `Put` se persiste en `attachments.file_path`.

## Seguridad al agregar código

1. **Path traversal.** `LocalStorage.Put` rechaza claves donde `filepath.Clean(key) != key`,
   y el handler normaliza con `filepath.Base` antes de componer. Si construís una clave
   desde entrada del usuario, sanitizala igual.
2. **Tipos permitidos.** El upload valida el Content-Type contra una lista blanca (JPEG,
   PNG, GIF, WebP, PDF, TXT, DOC, DOCX). Ampliarla es una decisión de producto, no un detalle.
3. **Tamaño.** Limitado por `MAX_FILE_SIZE` (10MB por defecto) y además por un CHECK en la
   base (≤100MB).
4. **Compensación ante fallo.** Si el archivo se guardó pero el `INSERT` falla, el handler
   borra el archivo (`attachment_handler.go:268`). Mantené ese patrón: no hay transacción
   que abarque storage y base.

## Limitaciones conocidas

- **`GetURL` del provider `s3` devuelve una URL prefirmada**, pero la descarga no la usa: sirve el
  archivo por streaming a través de la API. Eso hace pasar todo el tráfico de archivos por
  el backend en vez de delegarlo al object storage.
- **Si `io.Copy` falla a mitad del stream de descarga**, el `200` ya se envió: la respuesta
  queda truncada y solo se loguea el error (`attachment_handler.go`).
- **La descarga no incluye a los evaluadores** — `canDownload` permite al dueño, al autor
  del evento y a un `admin`, pero no a quien tiene la propuesta asignada. Ver la deuda
  técnica en el overview.
- **Al eliminar una propuesta se borra primero el registro y después el archivo.** Si falla
  el borrado en el storage, la respuesta es igual `200` y el archivo queda huérfano.
