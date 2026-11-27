# Services

## Services

| Service | Type | Language | Surfaces | Exposes | Folder | Origin |
|---|---|---|---|---|---|---|
| api | service | go | — | HTTP | services/api | initial |
| web | frontend | react | Web | — | services/web | initial |

## Modules

### api

| Module | Entities | Responsibility | Origin |
|---|---|---|---|
| event | Evento, Inscripción | Eventos y su ciclo de vida, inscripción, recordatorios, visibilidad | initial |
| participant | Usuario | Cuentas, sesiones (JWT), login con Google, recuperación de contraseña, permisos | initial |
| attachment | Propuesta | Propuestas y sus archivos en el almacenamiento (local o S3) | initial |
| vote | Configuración de votación, Asignación, Voto, Borrador de voto, Resultados | Configuración, reparto, rankings, cálculo de MBC, Q y G′ | initial |
| notification | Notificación | Notificaciones dentro de la web y su retención | initial |
| email | — | Envío de emails por SMTP | initial |

## Interactions

| From → To | Purpose | Blocks | Origin |
|---|---|---|---|
| web → api | Todo lo que la web muestra o hace, por HTTP con la sesión del usuario | ACC, EVE, EXP, INS, VOT, RES, NOT, BAN | initial |
| api.event → api.vote | Abrir la votación (configuración y asignaciones) y calcular los resultados al pasar a Resultados | EVE, VOT, RES | initial |
| api.event → api.notification | Avisos de cambio de etapa, cancelación, pausa, cierre, inscripción y recordatorios | NOT | initial |
| api.event → api.email | Emails de cambio de etapa, cancelación, pausa, cierre y recordatorios | NOT | initial |
| api.vote → api.notification | Aviso de ranking enviado | NOT | initial |
| api.participant → api.email | Email de recuperación de contraseña | ACC, NOT | initial |
| api.attachment → api.vote | Descarga de una propuesta por el evaluador que la tiene asignada | INS, VOT | initial |
