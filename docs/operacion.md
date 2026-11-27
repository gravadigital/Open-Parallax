# Operación de OpenParallax

Cómo se opera el producto: cómo saber si está sano, qué procesos corren solos, cómo se configura y
qué conviene alertar. Para el detalle de qué hace el producto, ver la definición en `_jflow/prd/`.

Esta guía describe lo que existe hoy. Lo que pide jiku flow y el producto todavía no tiene (readiness,
logs JSON, OpenTelemetry, métricas) está marcado como «no existe» y figura en
[el reporte de adopción](adopcion.md).

## Servicios y salud

La api responde `GET /health` con su estado, su versión y si llega a la base:

```json
{ "status": "ok", "service": "openparallax-api", "version": "0.1.0", "database": "connected" }
```

Si la base no responde, contesta `503` con `"status": "error"`. No existe `GET /health/ready`, y el
health no revisa el almacenamiento ni el SMTP: el almacenamiento se verifica solo al arrancar (si el
bucket no existe, la api no arranca).

| Servicio | Health | Dependencias que chequea |
|---|---|---|
| api | `GET /health` (puerto 8080) | La base (ping) |
| web | No tiene: nginx sirve la SPA en el puerto 80 | — |
| database (PostgreSQL) | `pg_isready` (healthcheck de Compose) | — |
| storage (Garage) | `/garage health` (healthcheck de Compose) | — |
| mailpit (solo local) | `/mailpit readyz` (healthcheck de Compose) | — |

## Procesos en segundo plano

No hay procesos programados, workers ni colas. Lo único que corre fuera del pedido:

| Proceso | Servicio | Cada cuánto | Qué hace | Métrica de último éxito |
|---|---|---|---|---|
| Migraciones | api | Al arrancar | Aplica las migraciones pendientes (`schema_migrations`). Si una migración borra registros con archivo, borra después sus archivos del almacenamiento; el que no se puede borrar queda en el log (`file of a removed record could not be deleted`, con su clave) y la migración sigue | No existe; si fallan, la api no arranca |
| Envío de emails | api | Después de cada cambio de etapa, cancelación, pausa, cambio de cierre o recordatorio | Manda los emails en segundo plano, sin reintento; si la api se reinicia, se pierden | No existe; los fallos quedan en el log |
| Limpieza de notificaciones | api | Cuando un usuario lista sus notificaciones | Borra las suyas de más de 90 días | No existe |

## Configuración

Todo se configura con variables de entorno (ver `.env.example`). La api solo exige `JWT_SECRET` al
arrancar; la base y el almacenamiento fallan al conectarse si están mal, y el resto (email, Google)
recién falla al usarse. Un número o booleano mal escrito toma su default sin avisar.

| Variable | Servicio | Para qué | Default |
|---|---|---|---|
| `JWT_SECRET` | api | Firma de las sesiones; obligatoria | ninguno (Compose: un valor de desarrollo público) |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE` | api | La base | `localhost`, `5432`, `openparallax`, `openparallax_password`, `openparallax_db`, `disable` |
| `PORT` | api | Puerto HTTP | `8080` |
| `GIN_MODE` | api | `release` en producción; `debug` registra la cadena de conexión con la contraseña y el SQL | `debug` (Compose: `release`) |
| `FRONTEND_URL` | api | Base de los enlaces de los emails | `http://localhost:3000` |
| `CORS_ALLOW_ORIGINS`, `CORS_ALLOW_METHODS`, `CORS_ALLOW_HEADERS` | api | CORS | `*`, los que usa la web |
| `MAX_FILE_SIZE` | api | **Tamaño máximo de propuesta**, en bytes | `10485760` (10 MB) |
| `STORAGE_PROVIDER` | api | `s3` o `local` | `local` (Compose: `s3`) |
| `STORAGE_LOCAL_PATH` | api | Carpeta con `STORAGE_PROVIDER=local` | `./uploads` |
| `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY` | api | El almacenamiento S3; el bucket tiene que existir | `http://localhost:3900`, `garage`, `openparallax`, vacío, vacío |
| `S3_PRESIGN_ENDPOINT` | api | Hoy sin efecto | vacío |
| `GOOGLE_CLIENT_ID` | api, web | Login con Google; vacío lo desactiva | vacío |
| `GOOGLE_TOKENINFO_URL`, `GOOGLE_USERINFO_URL` | api | Endpoints de Google; solo los tests los cambian | los de Google |
| `EMAIL_ENABLED` | api | Enviar emails | `false` (Compose: `true`, a Mailpit) |
| `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_SECURE` | api | El servidor SMTP: siempre STARTTLS (o TLS directo con `SMTP_SECURE=true`) y autenticación | vacío, `587`, vacío, vacío, `false` |
| `EMAIL_FROM`, `EMAIL_FROM_NAME` | api | El remitente | vacío, `OpenParallax` |
| `API_URL` | web (contenedor) | La url de la api como la ve el navegador | vacío (usa `http://localhost:8080`) |
| `UPLOADS_DIR` | api | Solo se registra en el log | `./uploads` |

## Observabilidad

No existe todavía lo que pide jiku flow: ni OpenTelemetry, ni logs JSON, ni métricas, ni un id de
correlación en las respuestas.

- **Logs de la api:** texto en la salida de error, con fecha, nivel y lugar del código (nivel debug
  con `GIN_MODE=debug`), más las líneas de acceso de Gin. El middleware de pedidos registra un
  `request_id`, pero no es único (cambia una vez por segundo) ni se devuelve en la respuesta. Los
  logs incluyen emails de usuarios.
- **Logs de la web:** los de nginx, y una línea al arrancar con la configuración aplicada.

## Alertas recomendadas

Qué conviene alertar. El producto no tiene requisitos no funcionales de disponibilidad ni de tiempos,
así que los umbrales son un punto de partida a ajustar con el uso.

| Alerta | Condición | Umbral |
|---|---|---|
| Servicio caído | `GET /health` de la api no responde o no es `200` (no existe `/health/ready`) | 3 chequeos seguidos, cada 30 s |
| Tasa de errores | Respuestas 5xx sobre el total, desde los logs de acceso o el proxy de entrada | Más del 5 % durante 5 minutos |
| Latencia | Percentil 95 de la duración de los pedidos, desde el proxy de entrada (la api no publica métricas) | Más de 2 s durante 10 minutos |
| Proceso que no corrió | No aplica: no hay procesos programados | — |
| Cola acumulada | No aplica: no hay colas | — |
| Emails que fallan | Líneas «failed to send email to recipient» o «failed to connect» del servicio de email en el log de la api | Cualquiera |

## Secretos del CI

Lo que los pipelines necesitan y es secreto se configura como variable de la plataforma, nunca en el
repositorio. GitHub: *Settings → Secrets and variables → Actions*.

| Variable | Qué es | Cómo se crea | Si falta |
|---|---|---|---|
| `DOCKERHUB_USERNAME` | Usuario de Docker Hub que publica las imágenes `gravadigital/openparallax-{api,web}` | El usuario de la cuenta u organización en Docker Hub | Falla la publicación de imágenes (`dev-images.yml`, `release.yml`); el CI de los checks no lo usa |
| `DOCKERHUB_TOKEN` | Token de acceso de ese usuario | Docker Hub → *Account settings → Personal access tokens → Generate new token*, con permiso *Read & Write* | Igual que el anterior |

## Builds para descargar

Ninguno: el producto se despliega como imágenes. El CI publica en Docker Hub
`gravadigital/openparallax-api` y `gravadigital/openparallax-web` con los tags `dev` y `dev-{sha}`
en cada push a `dev`, y `X.Y.Z`, `X.Y` y `latest` en cada tag `vX.Y.Z`. Los reportes JUnit de cada
corrida del CI quedan como artefactos (`junit-api`, `junit-web`).

## Despliegue

El despliegue a servidores vive en otro repositorio. Lo que necesita una instalación:

- Las imágenes de Docker Hub, con las variables de [Configuración](#configuración): como mínimo
  `JWT_SECRET` (un valor largo y aleatorio), la base, el S3 con el bucket ya creado y privado, y
  `GIN_MODE=release`. La web solo necesita `API_URL` (y `GOOGLE_CLIENT_ID` si se usa Google): la
  misma imagen sirve para cualquier instalación.
- Las migraciones las aplica la api al arrancar, después de conectarse al almacenamiento (también
  `make migrate`; si ahí el almacenamiento no responde, la migración sigue y solo registra en el log
  los archivos que no pudo borrar).
- La migración 029 (#662) deja a cada inscrito una sola propuesta por evento: conserva la más
  reciente y borra las demás con su archivo, sin avisarle. Cada propuesta borrada queda en el log
  (`removed a duplicate proposal`, con evento, participante y clave). Un archivo que no se pudo borrar
  queda en el bucket privado, sin registro que lo use: se puede borrar a mano con la clave del log.
- El primer usuario se crea desde la web («Crear cuenta»). Los roles admin y organizer se asignan
  escribiendo `users.role` en la base.
