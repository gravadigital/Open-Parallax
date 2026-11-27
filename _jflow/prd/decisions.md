# Decisions

## Product

| ID | Topic | Decision | Rejected | Reason | Affects | Origin |
|---|---|---|---|---|---|---|
| DEC-01 | Modelo de evaluación | Revisión distribuida entre pares de Merrifield & Saari (2009): los participantes ordenan propuestas ajenas, Modified Borda Count, calidad del evaluador e incentivo sobre su propia propuesta | — | Una convocatoria con más propuestas que evaluadores no se resuelve con un comité; el incentivo hace que convenga evaluar bien | VOT, RES | initial |
| DEC-02 | Ranking oficial | G′, el ranking ajustado, es el oficial; G se publica como referencia | Dejar sin definir cuál de los dos es el oficial | Sin G′ como oficial, la calidad del evaluador y el incentivo no tenían efecto sobre el resultado | RES | initial |
| DEC-03 | Sanción a quien no evalúa | Configurable por evento: exclusión de su propuesta (por defecto) o penalización en puestos | — | Quien no completa su ranking no aporta evaluación; el autor decide cuán dura es la sanción | VOT, RES | initial |
| DEC-04 | Visibilidad de los ajustes | El público ve G′ y una explicación general; cada participante ve el ajuste de su propuesta; el autor ve todos. La calidad Q no se publica | — | — | RES | initial |
| DEC-05 | Reparto de propuestas | Asignación regular: cada evaluador revisa exactamente m propuestas y cada propuesta recibe exactamente m evaluaciones, nunca de su autor | Garantizar un mínimo por propuesta y completar al azar | Con el reparto anterior, con 100 propuestas y m = 14 cada una recibía entre 4 y 29 evaluaciones, y alguna podía quedar en 0; con m fijo el promedio y la suma del paper dan el mismo MBC | VOT | initial |
| DEC-06 | Rankings completos | Solo cuenta el ranking completo; los votos de quien no lo completó se descartan, y el consenso no se muestra durante la votación | Aceptar rankings parciales | — | VOT, RES | initial |
| DEC-07 | Notificaciones dentro de la web | Replican los emails y suman inscripciones, ranking enviado y recordatorios; se muestran 90 días | — | — | NOT, BAN | initial |
| DEC-08 | Idiomas | Interfaz en español neutro (sin voseo ni modismos rioplatenses) e inglés, con selector que se recuerda en el navegador | — | — | web | initial |
| DEC-09 | Resultados y avance públicos | Los resultados (en Resultados) y el avance de la votación, con quién envió su ranking, se ven sin sesión y solo para leer | — | — | RES, VOT | initial |

## Technical

| ID | Topic | Decision | Rejected | Reason | Affects | Origin |
|---|---|---|---|---|---|---|
| DEC-T01 | Reglas que no pueden romperse | Las invariantes de la votación se aplican también con funciones y triggers de PostgreSQL, además de en Go | — | Ninguna escritura que no pase por la api puede romper el conflicto de interés, que sostiene la legitimidad del resultado | api.vote, INV-01, INV-02, INV-03, INV-12 | initial |
| DEC-T02 | Arquitectura | Dos servicios: la api, único con estado, con todo el dominio, y la web, una SPA sin lógica de negocio propia. Comunicación en un solo sentido, sin broker, colas ni caché | — | El dominio es acotado (decenas de participantes por evento, hasta 100) y no hay integradores externos | api, web | initial |
| DEC-T03 | HTTP y persistencia de la api | Gin 1.10 y GORM 1.30, con el wiring manual en `cmd/api/main.go` | — | — | api | initial |
| DEC-T04 | Contrato de la api | Las structs de dominio son a la vez modelo de persistencia y contrato JSON | — | — | api | initial |
| DEC-T05 | Web | Create React App, SPA sin SSR, con react-router | — | — | web | initial |
| DEC-T06 | Estilos | Un CSS por componente y tokens de diseño como variables CSS, sin framework de estilos ni librería de componentes | — | El lenguaje visual propio no es el de ninguna librería | web | initial |
| DEC-T07 | Sesión vencida en la web | Ante un 401 el cliente HTTP limpia la sesión y emite el evento `auth:logout`, que el contexto de sesión escucha para cerrar la sesión sin recargar | — | Un solo lugar para una reacción que puede ocurrir en cualquier pantalla | web, ACC-10 | initial |
| DEC-T08 | Notificaciones | Se guardan en una tabla en el mismo handler que dispara el email, sin bloquear la operación; la web las consulta por polling cada 60 s. Se guardan tipo y datos, no el texto | WebSocket o Server-Sent Events; cola y worker; guardar el texto renderizado | Sin estado de conexión en una api sin estado y sin comunicación asíncrona (DEC-T02); el texto se arma en el idioma de quien lo lee | api.notification, web, BAN, NOT | initial |
| DEC-T09 | Traducciones | Módulo propio sobre `Intl`, con catálogos `es` y `en`; el tipo de `en` se deriva de `es` para que una clave faltante no compile | react-i18next; FormatJS | Dos idiomas y unas 15 pantallas no justifican una dependencia | web | initial |
| DEC-T10 | Almacenamiento de propuestas | S3 con Garage; la api no crea el bucket y lo verifica al arrancar; el bucket es privado y la api sirve los archivos | MinIO | MinIO dejó de publicar imágenes y su repositorio quedó archivado | api.attachment, INT-03 | initial |
| DEC-T11 | Configuración de la web | La url de la api y el client id de Google entran al arrancar el contenedor (`config.js`), no al construir: la misma imagen sirve para cualquier instalación | — | — | web | initial |
| DEC-T12 | Versión | Una sola versión para todo el monorepo, en `VERSION`, que `scripts/set-version.sh` propaga y verifica; la api la recibe al compilar | — | La api en Go no tiene dónde declarar su versión | api, web | initial |
| DEC-T13 | Publicación | El CI publica las imágenes de api y web en Docker Hub en cada push a dev y en cada tag `vX.Y.Z`; el despliegue a servidores vive en otro repositorio | — | — | api, web | initial |
| DEC-T14 | Login con Google | La web usa el flujo implícito y la api verifica el token de acceso contra tokeninfo (audiencia y email verificado) y userinfo | — | El chequeo de audiencia impide reusar un token emitido para otra aplicación; el de email, porque las cuentas se vinculan por email | api.participant, INT-01 | initial |

## Reverted

| ID | Topic | Said | Reverted by | Reason | Now |
|---|---|---|---|---|---|
