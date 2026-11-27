# OpenParallax

[![CI](https://github.com/gravadigital/Open-Parallax/actions/workflows/ci.yml/badge.svg)](https://github.com/gravadigital/Open-Parallax/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Evaluación distribuida entre pares para convocatorias con más propuestas que personas disponibles
para evaluarlas. En lugar de un comité, los participantes ordenan las propuestas de los demás, y
evaluar con cuidado es lo que hace subir la propia propuesta.

```
navegador ── HTTP ──> web ── HTTP + JWT ──> api ──> PostgreSQL
                    (SPA React)             (Go)  ├─> S3 Garage (archivos de las propuestas)
                                                  └─> SMTP (avisos por email)
```

Lo central es el **motor de votación**, una implementación de Merrifield & Saari (2009), *Telescope
time without tears*. Cada participante ordena `m` propuestas que no son la suya; un Modified Borda
Count combina las posiciones en un ranking global, y cada evaluador recibe un puntaje según qué tan
cerca estuvo del consenso. **Quien evalúa bien ve subir su propia propuesta; quien evalúa mal la ve
bajar**: por eso conviene evaluar en serio.

La regla de la que depende el resultado, **nadie evalúa su propia propuesta**, se aplica dos veces:
en el código Go y en triggers de PostgreSQL, así que ninguna escritura que saltee la api puede
romperla.

## Qué hace

Eventos que pasan por `creación → participación → votación → resultados`; inscripción pública con
un enlace para compartir; una propuesta por participante; reparto y ranking configurables;
resultados publicados; y avisos en la web y por email en cada cambio de etapa. Se entra con email y
contraseña, o con Google.

Descripción completa en [documentation/features.md](documentation/features.md) (en inglés).

## Cómo levantarlo

Solo hace falta Docker (con Compose) y `make`. No hay nada que configurar antes:

```bash
make up      # construye y levanta todo, y espera a que cada servicio esté sano
make down    # lo baja, conservando los datos
make logs    # sigue los logs
```

| Qué | Dónde |
|---|---|
| La web | http://localhost:3000 |
| La api | http://localhost:8080 (`/health` para su estado) |
| Mailpit: los emails que manda el producto | http://localhost:8025 |
| pgAdmin (opcional: `docker compose --profile tools up -d pgadmin`) | http://localhost:5050 |

Para usarlo en local, crea una cuenta desde la web («Crear cuenta»): cualquier cuenta puede crear
eventos. Los emails (recuperación de contraseña, avisos) no salen de la máquina: llegan a Mailpit.

La configuración está en `.env` (no se versiona), que se crea copiando
[`.env.example`](.env.example): ahí están todas las variables con su default y para qué sirven. Si
un puerto choca con otro proyecto, cámbialo ahí (`API_PORT=8090`, por ejemplo).

**Login con Google (opcional).** Sin `GOOGLE_CLIENT_ID` solo se entra con email y contraseña. Para
activarlo, crea un *OAuth client ID* de tipo *Web application* en
[Google Cloud Console → Credentials](https://console.cloud.google.com/apis/credentials), agrega
`http://localhost:3000` como *Authorized JavaScript origin* y pon su Client ID en `GOOGLE_CLIENT_ID`
de tu `.env`.

**Roles admin y organizer.** Solo se asignan escribiendo en la base (`users.role`); toda cuenta nueva
es `participant`.

## Comandos

| Comando | Qué hace |
|---|---|
| `make check` | Lint, typecheck y todos los tests (unitarios, de integración y end to end) |
| `make test` | Todos los tests (`S=api` o `S=web` para uno solo); reportes JUnit en `reports/junit/` |
| `make test-e2e` | Levanta el producto con la configuración de test y corre los tests end to end |
| `make dev` | Levanta la base, el storage y Mailpit, y corre api y web a mano (la web con recarga en caliente) |
| `make migrate` | Aplica las migraciones (la api también las aplica al arrancar) |
| `make reset` | Borra todo, datos incluidos |
| `make help` | Lista el resto |

Para trabajar en el código hacen falta también **Go** (versión en
[`services/api/go.mod`](services/api/go.mod)) y **Node.js 22**
([`services/web/.nvmrc`](services/web/.nvmrc)). Más detalle en
[documentation/installation.md](documentation/installation.md).

## Estructura

| Carpeta | Qué es |
|---|---|
| [services/api/](services/api/) | Servicio HTTP en Go: eventos, identidad, archivos, votación |
| [services/web/](services/web/) | La interfaz: una SPA en React |
| [infra/](infra/) | Configuración de Garage y el seed de los tests end to end |
| [docker-compose.yml](docker-compose.yml) | El producto completo, en local |
| [_jflow/](_jflow/) | La definición del producto: qué hace, reglas, decisiones, contratos |
| [docs/](docs/) | La guía de operación y el reporte de adopción |
| [documentation/](documentation/README.md) | La documentación pública de uso e instalación (en inglés) |

## Documentación

- [`docs/operacion.md`](docs/operacion.md): cómo operarlo — salud, procesos, configuración, alertas.
- [`docs/adopcion.md`](docs/adopcion.md): el reporte de la adopción de jiku flow — bugs sospechados,
  comportamientos a confirmar y mejoras sugeridas.
- [`documentation/`](documentation/README.md): uso, instalación, configuración y referencia de la
  api, en inglés.
- [`CHANGELOG.md`](CHANGELOG.md): los cambios por versión.
- La definición del producto (qué hace, reglas y decisiones) está en
  [`_jflow/prd/index.md`](_jflow/prd/index.md), y el contrato HTTP en
  [`_jflow/apis/api.openapi.yaml`](_jflow/apis/api.openapi.yaml). Los requerimientos, con su
  especificación, están en Jiku.
- [`AGENTS.md`](AGENTS.md): la guía para los agentes que trabajan en el repositorio.

## Contribuir

Ver [CONTRIBUTING.md](CONTRIBUTING.md).

## Licencia

[MIT](LICENSE) © Grava.
