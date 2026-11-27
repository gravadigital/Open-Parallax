# Adopción de jiku flow

> Informe de la adopción del 27 de noviembre de 2026 (requisito #622). Es un registro de ese día: no
> se mantiene al día. Lo que decidas hacer con lo que lista entra como requisito de Jiku
> (`/jflow-spec` + `/jflow-build`) o como cambio directo (`/jflow-fix`).

## Resumen

| | |
|---|---|
| Definición | 8 bloques (ACC, EVE, EXP, INS, VOT, RES, NOT, BAN), 10 entidades, 40 operaciones, 12 invariantes, 15 garantías, 23 decisiones (9 de producto, 14 técnicas), 4 flujos |
| Listo para especificar e implementar | Sí: el repositorio cumple el mínimo (M1–M9), verificado |
| Correcciones hechas en la adopción | 15; en el producto, solo las URLs de Google configurables y una traducción que pide un test existente |
| Cargado en Jiku | 15 incidencias de seguridad (ver «Cargado en Jiku») |
| En este informe | 23 bugs, 1 comportamiento a confirmar, 19 defaults sugeridos, 14 otras mejoras |
| Cobertura por ID | 23 de 193 IDs con un test que los nombra: solo los tests end to end de la adopción llevan IDs; los tests anteriores (más de 1.300) no |

## Línea de base

Cómo estaba el repositorio antes de la adopción, corriendo sus propios comandos (los del CI, el
`Makefile` y `web/package.json`).

| Comando de antes | Resultado |
|---|---|
| `./scripts/set-version.sh --check` | Pasa |
| `cd api && go build ./... && go vet ./...` | Pasa |
| `cd api && go test -race ./...` (unitarios) | Pasa: 8 paquetes, 0 fallas |
| `make test-integration` (base de Compose + `go test -tags=integration`) | Pasa: 11 paquetes (y 6 sin tests), 0 fallas |
| `cd web && npm ci && npm run build` | Pasa |
| `cd web && CI=true npm test` (Jest) | Pasa: 91 suites, 813 tests |
| `cd web && npm run lint` y `npm run typecheck` | Pasan (el CI no los corría) |
| `make up` (`deploy/docker-compose.yml`) | Levanta base, Garage, api y web; `/health` responde con la base conectada |

No había tests end to end. El email estaba apagado por defecto y no había Mailpit.

## Correcciones hechas en la adopción

Lo que la adopción cambió para llegar al mínimo de jiku flow. La lógica del producto no cambió: no
había ningún test fallando. Lo único que cambió en el producto: las URLs de Google se volvieron
configurables (sin cambiar su valor) y la web traduce un código de error más, porque un test existente
lo pide.

| Mínimo | Qué se hizo | Cómo se comprobó |
|---|---|---|
| M1 | `api/` y `web/` pasaron a `services/api/` y `services/web/` (con su historia); el Compose, a la raíz; la configuración de Garage, a `infra/storage/`. Se actualizaron todas las rutas: Makefile, CI, `scripts/set-version.sh`, `.gitignore`, plantilla de PR, README, CONTRIBUTING, `documentation/`, comentarios y los dos tests de la web que leen el contrato | Ninguna ruta vieja fuera del CHANGELOG; `make check` |
| M8 | El contrato `docs/apis/api.yaml` pasó a `_jflow/apis/api.openapi.yaml`, anotado con `operationId`, módulo, requerimientos y condiciones de falla; se le agregaron las respuestas y los códigos que la api devuelve y no figuraban (`500` de la asignación, `403` de los borradores, `BAD_REQUEST` y `NOT_FOUND` de los middlewares de permisos, `ACCOUNT_NO_PASSWORD`, `INVALID_USER_ID`). La documentación del framework anterior en `docs/` se borró (queda en la historia de git) | `jflow.py def-check` |
| M2 | `Makefile` en la raíz con todos los targets de jiku flow, que delegan en `services/api/Makefile` y `services/web/Makefile` (`S=` limita a uno). Se conservaron `stop`, `reset`, `infra`, `api` y `web` | `make check` |
| M3 | Reportes JUnit en `reports/junit/`: gotestsum para Go, jest-junit para Jest y el reporter de Playwright | `make check` deja `api-unit`, `api-integration`, `api-e2e`, `web-unit` y `web-e2e` |
| M4 | `.env.example` con todas las variables de Compose, la api y la web, con defaults seguros. El `deploy/.env` local pasó a `.env` | `jflow.py conformance` |
| M5 | Mailpit en Compose, con un certificado local generado una vez (`mailpit-cert`): la api siempre negocia STARTTLS verificando el certificado, y confía en ese por `SSL_CERT_DIR`, sin cambiar código. En local los emails ahora se mandan por defecto, a Mailpit. El esquema ya se construye desde una base vacía con las migraciones | `make up`; un email de recuperación llega a Mailpit |
| M6 | Los tests de integración corren contra la base de Compose (`openparallax_test`) con las migraciones aplicadas | `make test-integration` |
| M7 | El CI corre `make lint typecheck test-unit test-integration` por servicio (antes no corría el lint ni el typecheck de la web ni los tests de integración de la api) y guarda los reportes. La publicación de imágenes solo cambió de rutas | `.github/workflows/ci.yml` |
| M9 | Las URLs con que la api verifica los tokens de Google pasaron a ser configurables (`GOOGLE_TOKENINFO_URL`, `GOOGLE_USERINFO_URL`; por defecto, las de Google) para apuntarla a un Google falso en los tests. `docker-compose.test.yml` agrega ese Google falso y siembra un admin y un organizer de prueba. Tests end to end de la api (cliente HTTP con tokens reales) y de la web (Playwright) | `make test-e2e` |
| — | `services/web/.dockerignore` excluye también `coverage` y `e2e` | `make build` |
| — | Al completar el contrato, el test de la web que cuenta sus códigos de error (TS-23) esperaba 106 y pasaron a ser 108. Con tu respuesta: el test espera 108 y el código `INVALID_USER_ID` tiene ahora su texto en español y en inglés | `make test-unit S=web` |
| M2 | `make down`, `stop` y `reset` también quitan lo que agrega `make test-e2e` | `make test-e2e && make down` |
| M1 | El Dockerfile de la web pasó de `services/web/docker/` a `services/web/` | `make build S=web` |
| M2 | Los Makefiles de cada servicio tienen también `up`, `down`, `logs` y `check`; su `test-e2e`, corrido solo, levanta el producto con la configuración de test | `make -C services/api check` |
| — | Dos comentarios del código citaban IDs del análisis anterior que hoy significan otra cosa (`U-03 / JTBD-03`, `D-11`): ahora citan RES-05, DEC-09 y RES-06 | — |

## Cargado en Jiku

| # | Título | Por qué |
|---|---|---|
| — | Inscripción en eventos sin identificarse, a nombre de cualquier email | Cualquiera puede inscribir a terceros: crea cuentas, inscribe cuentas existentes sin su consentimiento, llena el cupo, les hace llegar los emails del evento y obtiene el identificador del usuario (INS-01). Confirmado como error |
| — | Emails de los inscritos visibles para cualquier usuario | Cualquier cuenta lista nombre y email de los inscritos de cualquier evento (INS-04). Confirmado como error |
| — | El organizador puede votar en nombre de los participantes | El autor y los administradores envían rankings, borradores y propuestas por cualquier inscrito: pueden fabricar el resultado (VOT-09, INS-06). Confirmado como error |
| — | El rol global organizer actúa sobre cualquier evento | Configura la votación, genera asignaciones y recalcula resultados de eventos ajenos, y ve la calidad de los evaluadores (VOT-13, RES-06). Confirmado como error |
| — | Enumeración de cuentas en el registro y el login | Las respuestas distintas revelan si un email tiene cuenta y de qué tipo (ACC-E3, ACC-E4 a ACC-E6). Confirmado como error |
| — | Cualquier usuario lee los datos de cualquier otro | `GET /api/v1/users/{id}` devuelve nombre, email y rol de cualquier cuenta a cualquier sesión (ACC-08); ya figuraba como defecto conocido |
| — | Apropiación de cuentas por la vinculación con Google | Una cuenta creada con un email ajeno, sin verificarlo, queda vinculada al Google de su dueño cuando este entra, y el atacante conserva su contraseña (ACC-04) |
| — | Credenciales y datos personales en los logs | En modo debug (el default del código y de `make dev`) se registra la cadena de conexión con la contraseña y el SQL con hashes y tokens de recuperación; los emails se registran siempre. La web escribe en la consola del navegador el email del usuario al restaurar la sesión y el nombre elegido con Google |
| — | Inyección de cabeceras en los emails por el nombre del evento | El nombre del evento entra sin filtrar en el asunto: un salto de línea agrega cabeceras o cuerpo a los emails que reciben todos los inscritos |
| — | Subida de propuestas sin endurecer | El tipo lo declara quien sube y se devuelve tal cual al descargar, sin `nosniff`; el tamaño se controla después de leer todo el cuerpo, sin límite; el nombre original no se escapa en la descarga |
| — | Sesiones que no se pueden revocar | Una sesión sigue valiendo 24 horas después de cambiar la contraseña o de quitarle un rol: quien recupera una cuenta comprometida sigue expuesto, y un administrador degradado sigue siéndolo hasta que vence |
| — | Sin límite de intentos en el login, el registro y la recuperación | Permite probar contraseñas por fuerza bruta y mandar emails de recuperación en masa a cualquier dirección |
| — | Token de recuperación guardado sin hash | Quien lea la base puede cambiar la contraseña de cualquier cuenta con un pedido vigente; ya figuraba como defecto conocido |
| — | Detalles internos en las respuestas de error | Algunos errores devuelven el mensaje de la base o del parser del token |
| — | Eventos en Creación detectables por su nombre | Crear un evento con el nombre de uno oculto responde que ya existe, y revela que existe |

## Bugs sospechados

No son de seguridad, así que no se cargaron en Jiku. La definición los documenta tal como funcionan hoy.

### B-01 Interruptores de aviso por email que no se guardan

- **Qué pasa:** en Notificaciones, los interruptores «Avisarme también por email» cambian solo en pantalla; al recargar vuelven a sus valores fijos.
- **Por qué parece un error:** confirmado como error en la adopción; las preferencias estaban fuera de alcance y la pantalla las muestra como si existieran.
- **Dónde:** BAN-05; `services/web/src/pages/notifications/NotificationsPage.tsx:27-34,223-238`.
- **Gravedad sugerida:** media: el usuario cree que configuró algo.

### B-02 Recordatorio automático anunciado que no existe

- **Qué pasa:** la tarjeta «Recordatorios» dice «Te avisamos 24 h antes de cada cierre…», pero no hay ningún aviso automático.
- **Por qué parece un error:** confirmado como error en la adopción.
- **Dónde:** BAN-05; `services/web/src/i18n/catalogs/es.ts:55-58`.
- **Gravedad sugerida:** media.

### B-03 Un evento cancelado sigue funcionando

- **Qué pasa:** cancelado, el evento acepta inscripciones, propuestas, rankings y cambios de etapa; solo la pausa los bloquea. La web no ofrece cancelar.
- **Por qué parece un error:** confirmado como error en la adopción.
- **Dónde:** EVE-10; `services/api/internal/handlers/event_handler.go:1086`, `attachment_handler.go:135`.
- **Gravedad sugerida:** media.

### B-04 Una base nueva y una actualizada no tienen el mismo esquema

- **Qué pasa:** la migración 002 crea las tablas con `AutoMigrate` sobre modelos que se siguieron editando, y las posteriores agregan columnas «si no existen»: en una base nueva `events.is_cancelled` e `is_paused` aceptan nulos, y las claves foráneas tienen nombres de GORM. La migración 004 ignora sus propios errores.
- **Por qué parece un error:** el esquema depende de cuándo se creó la base.
- **Dónde:** `services/api/internal/storage/migrations/002_*.go`, `models.go:97-98`, `004_constraints_and_triggers.go:235-238`.
- **Gravedad sugerida:** alta: un cambio de modelo altera bases nuevas sin migración.

### B-05 Los emails distinguen mayúsculas

- **Qué pasa:** registro, login, vínculo con Google e inscripción comparan el email tal cual.
- **Por qué parece un error:** `Ana@x.com` y `ana@x.com` son dos cuentas; el vínculo con Google falla.
- **Dónde:** INV-08, GAR-10; `services/api/internal/storage/postgres/user_repository.go:41,90`.
- **Gravedad sugerida:** media.

### B-06 Publicar sin votos deja los resultados sin calcular para siempre

- **Qué pasa:** si nadie votó, el paso a Resultados se hace igual, el cálculo falla, la página pública responde que no hay resultados y recalcular también falla.
- **Por qué parece un error:** el evento queda cerrado sin resultado posible.
- **Dónde:** RES-01, RES-E2, RES-E5; `services/api/internal/domain/vote/voting_service.go:164-166`.
- **Gravedad sugerida:** media.

### B-07 Reglas sin protección ante pedidos simultáneos

- **Qué pasa:** el cupo, una propuesta por inscrito y el nombre único se controlan en Go sin restricción en la base ni bloqueo; dos pedidos simultáneos pueden romperlos. Una propuesta duplicada hace fallar después la apertura de la votación con un 500. Un email duplicado simultáneo da 500 en vez de 409.
- **Por qué parece un error:** las garantías GAR-11 a GAR-13 no resisten concurrencia.
- **Dónde:** INV-09 a INV-11; `event_handler.go:1150-1181`, `attachment_handler.go:189-202`.
- **Gravedad sugerida:** media.

### B-08 Los emails pueden fallar en silencio

- **Qué pasa:** el envío informa éxito aunque fallen todos los destinatarios; el asunto no se codifica (RFC 2047); no funciona con un servidor sin STARTTLS; se manda sin reintento en segundo plano y se pierde si la api se reinicia.
- **Por qué parece un error:** nadie se entera de que un aviso no llegó.
- **Dónde:** NOT-11; `services/api/internal/email/email_service.go:147-166,209-219`.
- **Gravedad sugerida:** media.

### B-09 Textos con voseo

- **Qué pasa:** los emails dicen «Ingresá», «Si tenés preguntas, comunicate», «Hacé clic», «podés ignorar»; la web, «Pegalo en el chat de Instagram para mandarlo.».
- **Por qué parece un error:** el producto se escribe en español neutro (DEC-08); el test de la web no detecta «pegalo».
- **Dónde:** NOT-01 a NOT-10; `services/api/internal/email/templates.go`, `services/web/src/i18n/catalogs/es.ts:1038`.
- **Gravedad sugerida:** baja.

### B-10 Cuenta regresiva con zona fija y cierre anticipado

- **Qué pasa:** la etiqueta dice «23:59 (ART)» aunque calcula con la hora del navegador; pasada la fecha estimada muestra «LA INSCRIPCIÓN YA CERRÓ» mientras la etapa sigue abierta.
- **Por qué parece un error:** la fecha es estimada y la etapa solo cambia cuando el autor la avanza.
- **Dónde:** EXP-07; `services/web/src/i18n/catalogs/es.ts:692,701`, `components/events/countdown/Countdown.tsx:59-68`.
- **Gravedad sugerida:** media.

### B-11 El identificador de pedido de los logs no es único

- **Qué pasa:** el `request_id` usa un formato literal y es el mismo para todos los pedidos del mismo segundo.
- **Por qué parece un error:** no sirve para seguir un pedido.
- **Dónde:** `services/api/internal/middleware/events/events.go:51-53`.
- **Gravedad sugerida:** baja.

### B-12 Estadísticas de calidad que nunca se calculan

- **Qué pasa:** el avance de la votación informa calidad media y evaluadores buenos y malos sobre una columna que nunca se escribe, con umbrales fijos 0,7 y 0,3; un identificador mal formado da 500.
- **Por qué parece un error:** los números son siempre vacíos o ceros.
- **Dónde:** VOT-10; `distributed_vote_handler.go:1034-1038,1064-1078`.
- **Gravedad sugerida:** baja.

### B-13 Registro con Google con un email existente da 500

- **Qué pasa:** responde `500 INTERNAL_ERROR` en vez de un conflicto; un `GOOGLE_CLIENT_ID` vacío se informa como «Failed to contact Google API».
- **Por qué parece un error:** el error no dice qué pasa.
- **Dónde:** ACC-E10, ACC-E14; `google_auth_handler.go:162-195`.
- **Gravedad sugerida:** baja.

### B-14 El paginado de notificaciones puede saltear avisos

- **Qué pasa:** el cursor es solo la fecha, con `<` estricto: dos avisos del mismo instante en el borde de una página se pierden.
- **Dónde:** BAN-01; `notification_repository.go:85-89`.
- **Gravedad sugerida:** baja.

### B-15 Los borradores de voto no se validan

- **Qué pasa:** el borrador acepta cualquier propuesta, cualquier posición y en cualquier etapa.
- **Dónde:** VOT-07; `vote_draft_handler.go:87-110`.
- **Gravedad sugerida:** baja.

### B-16 La descripción para compartir se corta por bytes

- **Qué pasa:** se toman los primeros 197 bytes, que pueden partir una letra con tilde.
- **Dónde:** EXP-03; `event_handler.go:1962-1963`.
- **Gravedad sugerida:** baja.

### B-17 Código de estado incorrecto al pedir eventos ajenos

- **Qué pasa:** pedir los eventos de otro usuario con alcance `participating` responde 401 en vez de 403.
- **Dónde:** EXP-E3; `user_handler.go:279-287`.
- **Gravedad sugerida:** baja.

### B-18 Reemplazar la propuesta no es atómico

- **Qué pasa:** la web borra la propuesta y después sube la nueva; si la subida falla, el inscrito queda sin propuesta.
- **Dónde:** INS-07; `services/web/src/pages/event-detail/EventDetailPage.tsx:285-320`.
- **Gravedad sugerida:** media.

### B-19 Datos de un usuario quedan en el navegador para el siguiente

- **Qué pasa:** el borrador del asistente de creación y el último ranking enviado quedan en el navegador al cerrar sesión, y el borrador se ofrece al próximo usuario.
- **Dónde:** EVE-02; `CreateEventPage.tsx:130-144`, `domain/ranking.ts:77-78`.
- **Gravedad sugerida:** baja.

### B-20 Datos fijos mostrados como reales en la web

- **Qué pasa:** el pie dice «V1.0» (la versión es 0.1.0); un evento sin cupo se muestra con cupo 20; «Ver demo» lleva al inicio y la pantalla de acceso dice «Proyecto demo · los datos pueden reiniciarse».
- **Dónde:** `AppFooter.tsx:20`, `domain/eventDetail.ts:6,32-39`, `AppHeader.tsx:133-137`, `AuthLayout.tsx:56`.
- **Gravedad sugerida:** baja.

### B-21 Orden y alcance de las listas de eventos

- **Qué pasa:** la lista ordena por etapa solo dentro de la página actual; el inicio no ve eventos abiertos más allá de los primeros 100.
- **Dónde:** EXP-01, EXP-05; `EventsListPage.tsx:178-184`, `HomePage.tsx:41`.
- **Gravedad sugerida:** baja.

### B-22 Documentación pública desactualizada

- **Qué pasa:** `documentation/configuration.md:27-28` e `installation.md:99` dicen que la web lee su configuración al construirse (hoy es al arrancar); `installation.md:30-32` dice que se cargan usuarios de ejemplo (ya no).
- **Gravedad sugerida:** baja.

### B-23 El evento no tiene fechas propias

- **Qué pasa:** la web no pide fechas: manda como inicio el día siguiente y como fin el posterior, y el paso 2 del asistente se llama «Fechas y cupo» aunque solo tiene el cupo.
- **Por qué parece un error:** confirmado como error en la adopción; las fechas del evento no dicen nada real.
- **Dónde:** EVE-02, EVE-03; `services/web/src/domain/eventForm.ts:141-147`, `services/web/src/i18n/catalogs/es.ts:468`.
- **Gravedad sugerida:** media.

## Comportamientos a confirmar

Comportamientos que pueden ser intencionales o no, y que quedaron sin respuesta en la adopción. La
definición los documenta tal cual, sin decisión. Al especificar algo que los toque, conviene decidirlos.

| Comportamiento | Duda | IDs |
|---|---|---|
| Los umbrales de calidad por defecto son 0,6 y 0,3 en la api y 0,65 y 0,35 en la base | Solo cuentan los de la api, que siempre los escribe; se preguntó cuál es el valor buscado y no se sabe | VOT-02 |

## Defaults de jiku flow que el producto no sigue

Los productos nuevos los siguen desde el inicio; este los sigue desde ahora en lo nuevo. Adecuar lo
existente es opcional: si se decide mantener uno a propósito, queda como decisión técnica al especificarlo.

| Default | Hoy | Cambio sugerido | Esfuerzo |
|---|---|---|---|
| Errores RFC 9457 (`application/problem+json`, `code` en snake_case, `detail` en español) | Cuatro formas de error; códigos en MAYÚSCULAS; mensajes en inglés | Una forma única, con un período de convivencia en la web | Alto |
| Validación de toda entrada contra un esquema en el borde | Tags de Gin y chequeos a mano; varios identificadores sin validar terminan en 500; longitudes medidas en bytes | Validar identificadores y cuerpos antes de la lógica, con `400 validation_failed` | Medio |
| Configuración validada al arrancar | Solo `JWT_SECRET`; un número mal escrito toma su default sin avisar | Validar todas las variables y fallar con un mensaje claro | Bajo |
| Contraseñas con Argon2id | bcrypt | Migrar al iniciar sesión | Medio |
| Sesiones del navegador en cookies httpOnly | JWT en `localStorage`, legible por cualquier script, sin CSP | Cookie httpOnly, Secure y SameSite, y una CSP en nginx | Alto |
| Identificadores UUIDv7 | UUIDv4 (`uuid-ossp`) | `uuidv7()` en las tablas nuevas | Bajo |
| Listas con cursor (`limit`, `cursor`) | `page` y `limit` en eventos | Cursor en las listas nuevas | Medio |
| Campos `camelCase` | `snake_case` en todo el contrato | Solo con una versión nueva del contrato | Alto |
| Health `/health` y `/health/ready` con versión, commit y chequeos | Solo `/health`, con otra forma, sin chequear el almacenamiento | Agregar `/health/ready` y la forma estándar; `HEALTHCHECK` en las imágenes | Bajo |
| Observabilidad con OpenTelemetry y logs JSON con `trace_id` | Logs de texto, sin trazas ni métricas, con emails en los logs | SDK de OpenTelemetry en la api y logs JSON sin datos personales | Medio |
| Apagado ordenado y timeouts del servidor | `router.Run` sin timeouts ni manejo de `SIGTERM` | `http.Server` con timeouts y `Shutdown` | Bajo |
| Contrato OpenAPI 3.1 verificado por tests, con `x-roles` | OpenAPI 3.0.3, sin validación en los tests ni `x-roles` | Validar las respuestas de los tests de integración contra el contrato; pasar a 3.1 | Medio |
| Convenciones del esquema | `varchar(n)`, columnas sin `NOT NULL`, sin `created_at`/`updated_at` mantenidos por trigger, nombres de claves de GORM, emails sin normalizar, sin marcas `[pii]` | Aplicarlas en las tablas nuevas y al tocar las existentes | Medio |
| Escáner de secretos y auditoría de dependencias en el CI | No hay | gitleaks y `govulncheck` / `npm audit` en el CI | Bajo |
| Tiempo controlado en los tests (un reloj inyectable) | La api llama a `time.Now()` directamente (vencimientos, fechas de inicio, sesiones); los tests no controlan el reloj | Un reloj inyectable en los servicios que dependen del tiempo | Medio |
| Respuestas por defecto de la entrada: recorte de espacios, `409 invalid_state` para una acción que la etapa no permite, `422 reference_not_found` para una referencia inexistente en el cuerpo | El texto no se recorta (salvo dos campos); una acción que la etapa no permite responde casi siempre `400` (`INVALID_EVENT_STAGE`, `INVALID_REGISTRATION_STAGE`); una propuesta inexistente en el ranking, `404` | Aplicarlo en las operaciones nuevas y al tocar las existentes, junto con el formato de errores | Medio |
| `CHANGELOG.md` con la sección `[Sin publicar]` y cada entrada citando su requerimiento de Jiku (`#622`) | La sección se llama `[Unreleased]` y las entradas citan `REQ-00n` | Renombrar la sección; las entradas nuevas citan su requerimiento | Bajo |
| Imágenes en el registro de la plataforma (GitHub Container Registry), sin secretos | Se publican en Docker Hub, con `DOCKERHUB_USERNAME` y `DOCKERHUB_TOKEN` | Publicar en GHCR con el token del job, o decidir quedarse en Docker Hub | Bajo |
| Documentación para personas en `docs/`, en el idioma del producto | La documentación pública de uso e instalación está en `documentation/`, en inglés | Pasarla a `docs/` en español, o decidir mantenerla así (por ejemplo, por ser un proyecto abierto) | Medio |

## Otras mejoras sugeridas

| Mejora | Por qué | Esfuerzo |
|---|---|---|
| Quitar código muerto de la api: handlers sin ruta (`DeleteEvent`, `RemoveParticipant`, `GetAttachment`, `UpdateVotingConfiguration`, `DeleteVotingConfiguration`), `RequireRole`, el presigner, `optimizer.go`, las vistas de la migración 005 y las columnas sin uso | Confunde al leer y al cambiar el esquema | Medio |
| Quitar código muerto de la web: `Podium`, `RankingList`, `ParticipantsDialog`, `useFetch`, `ApiHealthService`, `getUserById`, `uploadFile`, constantes y claves de catálogo sin uso | Igual | Bajo |
| Retirar las operaciones de configurar la votación y generar asignaciones por separado | La apertura atómica las reemplazó y la de asignaciones no es transaccional | Bajo |
| `GET /api/v1/events` carga todos los eventos y consulta los inscritos de a uno | No escala con muchos eventos | Medio |
| Controlar el tamaño de la subida antes de leer el cuerpo | Hoy se lee todo el archivo y recién después se rechaza | Bajo |
| Tests de `AuthContext` en la web | La sesión no tiene tests propios | Bajo |
| Una `ErrorBoundary` en la web | Un error al renderizar deja la pantalla en blanco | Bajo |
| Migrar de Create React App a Vite | CRA está discontinuado | Alto |
| Corregir las dependencias de la web: librerías de test en `dependencies`, tipos de react-router v5 con el router v7 | Instalación de producción más pesada y tipos que no corresponden | Bajo |
| Formatear la api con `gofmt` y sumarlo al lint | Hoy 7 archivos no están formateados | Bajo |
| El build de la web en el CI ignora los warnings | Los warnings se acumulan sin que nadie los vea | Bajo |
| Que el test de S3 use el Garage de Compose en vez de testcontainers | Una sola forma de levantar la infraestructura de los tests | Bajo |
| `cmd/migrate -rollback` elige la última migración por fecha de aplicación | Con migraciones aplicadas en el mismo instante puede deshacer otra | Bajo |
| Quitar o actualizar las referencias al análisis anterior en comentarios y nombres de tests (`REQ-00n`, `S-0nn`, `D-n`, `DA-n`, `CA-n`, `TS-nn`…) | Citan documentos que la adopción borró (siguen en la historia de git, en `7ff8b00`); los tests se renombran con los IDs nuevos a medida que se tocan | Bajo |

## Contradicciones

- La api pone por defecto los umbrales de calidad en 0,6 y 0,3, y la base en 0,65 y 0,35 (VOT-02, `voting_configurations`).
