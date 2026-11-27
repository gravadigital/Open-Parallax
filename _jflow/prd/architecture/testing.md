# How it is tested

How each entry point of the product is tested end to end, against `make up` and through its real interface: the identity each role comes in with and how each external system is replaced. `jflow.py def-check` requires a row per surface, per exposed interface (`Exposes` in services.md), per role and per external integration.

## Entry points

| Entry point | Type | Service | Actors | How it is tested | Test | Origin |
|---|---|---|---|---|---|---|
| Web | web | web | U-01, U-02, U-03 | Playwright (Chromium, en su imagen oficial) contra la web de `make up`: crea la cuenta en la pantalla de registro, crea el evento con el asistente, abre la inscripción y mira el evento sin sesión; un error de consola falla el test | `services/web/e2e/tests/` | initial |
| api HTTP | HTTP | api | U-01, U-02, U-03, U-04, U-05 | Cliente HTTP en Go contra la api de `make up`, con el token que la api emite a cada usuario al crear la cuenta o iniciar sesión: el ciclo completo de un evento, el administrador, el organizador global, el login con Google y la recuperación de contraseña por email | `services/api/e2e/` | initial |

## Test identities

| Role | How the identity is obtained in tests | Origin |
|---|---|---|
| participant | Cada test crea sus usuarios con `POST /api/v1/users` (o la pantalla de registro) y usa el token que devuelve, o inicia sesión con `POST /api/v1/users/authenticate`. La api es su propio emisor: no hay proveedor externo que reemplazar | initial |
| organizer | Usuario sembrado por `e2e-seed` (`infra/e2e/seed.sql`, perfil de test): `organizer.e2e@openparallax.test`, contraseña `e2e-password-123`; el test inicia sesión con email y contraseña | initial |
| admin | Usuario sembrado por `e2e-seed` (`infra/e2e/seed.sql`, perfil de test): `admin.e2e@openparallax.test`, contraseña `e2e-password-123`; el test inicia sesión con email y contraseña | initial |

## External systems

| System | How it is replaced in tests | Origin |
|---|---|---|
| INT-01 Google Identity (OAuth 2.0) | Google falso (`services/api/e2e/fakegoogle`, servicio `fake-google` de `docker-compose.test.yml`): responde tokeninfo y userinfo para tokens `e2e:{sub}:{email}:{nombre}` con la audiencia `e2e-google-client`; la api apunta a él con `GOOGLE_TOKENINFO_URL` y `GOOGLE_USERINFO_URL`. La ventana de Google en la web no se prueba de punta a punta | initial |
| INT-02 Servidor SMTP | Mailpit de `make up` (STARTTLS con un certificado local en el que la api confía por `SSL_CERT_DIR`); los tests leen los emails con su API (`MAILPIT_URL`) | initial |
| INT-03 Almacenamiento S3 (Garage) | La versión local: Garage de `make up`, con el bucket creado al arrancar. El test de integración del almacenamiento levanta su propio Garage con testcontainers | initial |

## How to run them

`make test-e2e` levanta el producto con `docker-compose.yml` y `docker-compose.test.yml` (Google falso, la api y la web apuntadas a él), espera a que la api responda `/health`, corre `e2e-seed` y después los tests de cada servicio: `make -C services/api test-e2e` (Go, tag `e2e`) y `make -C services/web test-e2e` (Playwright en `mcr.microsoft.com/playwright:v1.65.1-noble`, red del host). Si `make up` estaba corriendo, recrea la api y la web con la configuración de test; un `make up` posterior las vuelve a la normal.

Los datos los crea cada test (nombres y emails únicos por corrida); solo las identidades admin y organizer vienen del seed. Los reportes quedan en `reports/junit/api-e2e.xml` y `reports/junit/web-e2e.xml`. Corren solo en local, nunca en el CI.
