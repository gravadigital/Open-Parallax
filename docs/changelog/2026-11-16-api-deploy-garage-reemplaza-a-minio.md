---
date: 2026-11-16
type: technical-change
service: api
---

# Cambio técnico: api y deploy — Garage reemplaza a MinIO

## Requerimiento original

> Migrar Telescope de MinIO a Garage siguiendo la guía base "Garage como reemplazo de MinIO"
> (Garage v2.4.1): MinIO ya no publica imágenes Docker y su repositorio está archivado.

## Resumen

El storage de propuestas pasa de MinIO a Garage, un almacenamiento de objetos compatible con
S3. La API cambia de cliente (`minio-go` → `aws-sdk-go-v2`) y de configuración (`MINIO_*` →
`S3_*`). El stack local levanta Garage en lugar de MinIO más `mc`.

## Cambios

| Área | Antes | Ahora |
|---|---|---|
| Imagen | `pgsty/minio` + `pgsty/mc` (builds de terceros) | `dxflrs/garage:v2.4.1` |
| Cliente Go | `github.com/minio/minio-go/v7` | `github.com/aws/aws-sdk-go-v2/service/s3` |
| `STORAGE_PROVIDER` | `minio` | `s3` (`minio` aborta el arranque con un mensaje) |
| Endpoint | `MINIO_ENDPOINT=minio:9000` + `MINIO_USE_SSL` | `S3_ENDPOINT=http://garage:3900` (el esquema define TLS) |
| Región | `MINIO_REGION=us-east-1` | `S3_REGION=garage`, igual a `s3_region` del servidor |
| Credenciales | Root de MinIO (`minioadmin` / `minioadmin123`) | Clave propia de la API (`S3_ACCESS_KEY_ID` / `S3_SECRET_ACCESS_KEY`, secreto de 16 caracteres o más) |
| Bucket | Lo creaba `minio-init`, y la API también si faltaba | Lo crea Garage con `--default-bucket`. La API solo hace `HeadBucket` y aborta si no existe |
| URLs prefirmadas | Firmadas con el endpoint interno | Firmadas con `S3_PRESIGN_ENDPOINT` (host público) |
| Healthcheck | `mc ready local` | `/garage health -q` |
| Nombre del servicio | `minio` / `telescope-minio` | `storage` / `telescope-storage` |
| Consola | Puerto 9001 | No hay. Se usa el CLI `garage` dentro del contenedor |
| Tests | Sin tests del storage S3 | `s3_storage_integration_test.go` con testcontainers contra Garage |

## Nombre del servicio

El servicio se llama **`storage`**, nunca por la herramienta: `storage` en el stack local
(`http://storage:3900`), `telescope-storage` en los servidores, y las variables del compose son
`STORAGE_*` y `S3_*`. Cambiar de proveedor no cambia el nombre que ve nadie. Si algún día se
publica, el subdominio es `storage.${DOMAIN}`. Hoy no se publica: las descargas pasan por la API.

## Impacto en despliegue

- **El repo de deploy (Telescopio-deploy) se migró en el mismo cambio.** `telescope-storage`
  reemplaza a `telescope-minio` y `telescope-minio-init`, y los scripts de verificación pasan a
  `verify-storage.sh` y `check-file-upload.sh`. La API rechaza `STORAGE_PROVIDER=minio`, así que
  la imagen nueva y el compose nuevo van juntos.
- **Los servidores son de un solo nodo:** `--single-node --default-bucket`, `replication_factor
  = 1` y snapshots de la metadata cada 6 h. No hay redundancia: hacen falta backups del volumen
  `telescope-<env>-storage-data`.
- **No se migraron datos de MinIO:** no había propuestas que conservar. El volumen
  `telescope-<env>-minio` que quede en un servidor ya no lo usa nadie.
- **El subdominio `minio.${DOMAIN}`** que se dio de alta para que no fallara el certificado del
  ingress ya no se usa y se puede dar de baja.
