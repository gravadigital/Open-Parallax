---
prefix: EXP
name: Exploración de eventos
surfaces: [Web]
covers: Ver los eventos públicos, su detalle y cómo compartirlos, y los eventos propios con lo que falta hacer en cada uno
entities: [Evento, Inscripción]
last_requirement: EXP-08
last_failure: EXP-E5
---

# EXP — Exploración de eventos

## Requirements

| ID | Requirement | Actor | Origin |
|---|---|---|---|
| EXP-01 | Cualquiera, sin sesión, lista los eventos que no están en Creación: búsqueda por nombre u organizador sin distinguir mayúsculas, filtro por etapa con la cantidad de eventos de cada etapa, páginas de hasta 100 (10 en la web) | U-01 | initial |
| EXP-02 | Cualquiera ve el detalle de un evento: datos, organizador (el nombre del autor si no se indicó), cupo, inscritos y cantidad de propuestas. Un evento en Creación solo lo ven su autor y los administradores; para el resto no existe | U-01 | initial |
| EXP-03 | Cualquiera obtiene los datos para compartir un evento (título, los primeros 200 caracteres de la descripción, enlace); la web ofrece compartirlo por WhatsApp, X, LinkedIn, Facebook, Instagram, email o copiando el enlace | U-01 | initial |
| EXP-04 | Un usuario ve sus eventos, los que organiza y aquellos en los que participa, del más nuevo al más viejo, con su estado en cada uno: si subió propuesta, si tiene asignación, si envió su ranking y su puesto en los resultados | U-02 | initial |
| EXP-05 | En la web, la página de inicio muestra un evento destacado y hasta 3 eventos con inscripción abierta; la lista de eventos muestra, con sesión, hasta 4 pendientes del usuario («Falta tu archivo», «Te toca votar») | U-01 | initial |
| EXP-06 | En la web, la acción de cada evento depende de quién mira y de la etapa: Gestionar (su autor), Participar, Subir archivo, Votar, Ver resultados o Ver evento; el autor que abre el detalle de su evento va a la gestión | U-02 | initial |
| EXP-07 | En la web, el detalle del evento muestra la cuenta regresiva al cierre estimado de la etapa, la línea de etapas y, para el inscrito, su progreso (propuesta, ranking) | U-02 | initial |
| EXP-08 | En la web, el detalle de un evento cancelado reemplaza toda acción por «Este evento fue cancelado / Ya no se puede participar en este evento»; el de un evento pausado, salvo en Resultados, por «El evento está pausado / Por ahora no se puede inscribir ni subir propuestas. El organizador reanudará el evento.», también en Votación, donde oculta el ranking | U-02 | initial |

## Failure conditions

| ID | Condition | Result | Origin |
|---|---|---|---|
| EXP-E1 | El filtro de etapa no es una etapa conocida (con `creation` la respuesta es una lista vacía) | `400 INVALID_STAGE_FILTER`, con las etapas válidas | initial |
| EXP-E2 | El alcance pedido en «mis eventos» no es `participating` ni `all` | `400 INVALID_PAYLOAD` | initial |
| EXP-E3 | Un usuario pide los eventos de otro: con alcance `participating` (o sin alcance), aunque sea administrador; con alcance `all`, si no es administrador | `401 UNAUTHORIZED_ACCESS` (alcance `participating`) o `403 FORBIDDEN` (alcance `all`) | initial |
| EXP-E4 | El identificador del usuario no es válido | `400 INVALID_USER_ID` | initial |
| EXP-E5 | Falla interna al leer | `500 RETRIEVAL_ERROR` | initial |
