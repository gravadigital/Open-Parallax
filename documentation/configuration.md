# Configuration

OpenParallax has two kinds of configuration: what you may change when running it **locally**,
and what a **server** must set. Locally nothing is required.

## Locally

Every setting has a working default in `docker-compose.yml`. To change one, create
`.env` — it is read both by docker compose and by the `Makefile`, so the same file
applies whether you use `make up` or `make api` / `make web`. It is not versioned.

### Login with Google

The only setting most people will want. Without it, login is by email and password only, and
the web hides the Google button.

```sh
# .env
GOOGLE_CLIENT_ID=xxxxxxxx.apps.googleusercontent.com
```

In the Google Cloud console, the client ID must list `http://localhost:3000` among its
authorised JavaScript origins. No client secret is needed: the web uses the implicit flow, and
the api checks with Google that the token was issued for this client ID and that the email is
verified.

Rebuild afterwards (`make up`, or restart `make web`): the web reads the client ID at build
time.

### Everything else

| Variable | Default | What it is for |
|---|---|---|
| `WEB_PORT` / `API_PORT` | `3000` / `8080` | Published ports of the web and the api |
| `POSTGRES_PORT` | `5432` | Published port of the database |
| `STORAGE_PORT` | `3900` | Published port of the storage's S3 API |
| `JWT_SECRET` | a fixed development value | Signs the session tokens |
| `EMAIL_ENABLED` | `true` | Email goes to the local Mailpit (http://localhost:8025, `MAILPIT_PORT`); `false` sends none. A real server needs the `SMTP_*` variables below |
| `GIN_MODE` | `release` (`debug` with `make api`) | `debug` for verbose api logs |
| `POSTGRES_DB` / `POSTGRES_USER` / `POSTGRES_PASSWORD` | `openparallax_db` / `openparallax` / `openparallax_password` | Database credentials |
| `S3_ACCESS_KEY_ID` / `S3_SECRET_ACCESS_KEY` / `S3_BUCKET` | `GKopenparallaxlocal01` / a fixed development value / `openparallax` | The api's storage key and bucket. Garage creates both on start. The ID needs 8 or more characters (`A-Z a-z 0-9 - _ .`) and the secret 16 or more |
| `STORAGE_RPC_SECRET` / `STORAGE_ADMIN_TOKEN` | fixed development values | The storage's internal secret and its admin token (Garage's `rpc_secret` and `admin_token`) |

Changing the database credentials after the first start has no effect on the existing
database: PostgreSQL only reads them when it initialises an empty volume. The same goes for
Garage: with its volume already created, a different `S3_SECRET_ACCESS_KEY` does not match the
existing key and Garage does not start. `make reset` starts over.

## On a server

Both published images read their configuration from environment variables when the
container starts. Neither carries the settings of any installation, so the same image runs
anywhere: changing a value means recreating the container, not rebuilding the image.

### api

**These must be set.** `JWT_SECRET` has no default in the api; the others have development defaults that are public:

| Variable | What goes in it |
|---|---|
| `JWT_SECRET` | A random secret: `openssl rand -base64 32`. Changing it logs everybody out |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` | The PostgreSQL connection |
| `DB_SSLMODE` | `require` if the database is not on the same private network |
| `STORAGE_PROVIDER` | `s3`. The other value, `local`, writes to the container's disk. `minio` no longer exists: the api refuses to start with it |
| `S3_ENDPOINT` | The URL of the S3 API, with scheme: `http://storage:3900`, `https://storage.example.com` |
| `S3_REGION` | Must match the server's `s3_region`. Garage's default is `garage`, which is also the api's default. A mismatch makes every request fail with 400 |
| `S3_BUCKET`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY` | The bucket and the api's own key, with `read` and `write` on it (never the admin credentials). **The bucket must exist**: the api checks it on start and does not create it, because a bucket created through the S3 API gets only a local alias in Garage |
| `S3_PRESIGN_ENDPOINT` | Optional. The public URL presigned links are signed with, since the signature covers the host. Empty means `S3_ENDPOINT`. Today downloads go through the api, so no link reaches the browser |
| `FRONTEND_URL` | The public URL of the web, used in the links inside emails |
| `CORS_ALLOW_ORIGINS` | The public URL of the web. The default, `*`, accepts any origin |

> If `JWT_SECRET` is missing, the api refuses to start and logs `JWT_SECRET is required`. The
> development value from `make` and the compose file is public: never reuse it on a server.

**Optional:**

| Variable | Default | What it is for |
|---|---|---|
| `PORT` | `8080` | Port the api listens on |
| `GIN_MODE` | `debug` | Set `release` in production. `debug` logs the database connection string, password included |
| `MAX_FILE_SIZE` | `10485760` (10 MB) | Largest proposal accepted, in bytes |
| `GOOGLE_CLIENT_ID` | empty | Enables login with Google. Must be the one the web was built with |
| `EMAIL_ENABLED` | `false` | With `false`, no email is sent and nothing fails |
| `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD` | —, `587`, —, — | The mail server |
| `SMTP_SECURE` | `false` | `true` for TLS from the start (port 465), `false` for STARTTLS (587) |
| `EMAIL_FROM`, `EMAIL_FROM_NAME` | —, `OpenParallax` | The sender |
| `CORS_ALLOW_METHODS`, `CORS_ALLOW_HEADERS` | the ones the web uses | Rarely changed |

Configuration is not validated at startup: a wrong value shows up later, when it is first
used. Check `/health` after deploying — it reports whether the database is reachable.

### web

> **Upgrading from an image built before this change:** those images had the api URL baked in
> and needed no variables. Current images need `API_URL` set on the web container, or the web
> points at `http://localhost:8080`. Add it before pulling the new image.

| Variable | What goes in it |
|---|---|
| `API_URL` | The public URL of the api, **as the browser sees it** — never an internal container name like `http://api:8080`. Empty falls back to `http://localhost:8080` |
| `GOOGLE_CLIENT_ID` | Same value as the api's `GOOGLE_CLIENT_ID`, or empty to hide the Google button |

```sh
docker run -p 80:80 \
  -e API_URL=https://api.example.com \
  -e GOOGLE_CLIENT_ID=xxxxxxxx.apps.googleusercontent.com \
  gravadigital/openparallax-web:1.2.3
```

How it works: the web is a static bundle, and Create React App would bake any
`REACT_APP_*` into it at build time. Instead, when the container starts, a script writes
these two variables into `/config.js`, which `index.html` loads before the app. The file is
served with `Cache-Control: no-store`, so a new value is picked up on the next page load.
The container logs what it wrote: `runtime config: API_URL=...`.

`REACT_APP_API_URL` and `REACT_APP_GOOGLE_CLIENT_ID` still work, but only outside Docker:
`make web` uses them, because there is no container to write `config.js`.
