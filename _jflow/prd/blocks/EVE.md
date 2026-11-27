---
prefix: EVE
name: Organización de eventos
surfaces: [Web]
covers: Crear un evento, editarlo, avanzarlo por sus etapas, posponer sus cierres, pausarlo, cancelarlo y recordar a los pendientes
entities: [Evento, Inscripción]
last_requirement: EVE-14
last_failure: EVE-E24
---

# EVE — Organización de eventos

## Requirements

| ID | Requirement | Actor | Origin |
|---|---|---|---|
| EVE-01 | Un usuario con sesión crea un evento con nombre (3 a 200 caracteres), descripción (10 a 2000), organizador opcional, cupo (1 a 100, 20 por defecto) y color de portada (uno de 16, «lavender» por defecto); el evento nace en Creación y su autor queda registrado como creador | U-03 | initial |
| EVE-02 | En la web, la creación es un asistente de tres pasos (identificación, cupo, revisión) que guarda un borrador en el navegador; al crear lleva a la gestión del evento con el aviso «Evento creado. Cuando esté listo, abre la inscripción.» | U-03 | initial |
| EVE-03 | El evento tiene fecha de inicio y de fin: el inicio es hoy o después y la duración va de 1 a 365 días. La web no las pide: envía como inicio el día siguiente y como fin el posterior | — | initial |
| EVE-04 | El nombre de un evento es único entre todos los eventos | — | initial |
| EVE-05 | El autor edita nombre, descripción, organizador y cupo mientras el evento está en Creación o Participación, y el color de portada en cualquier etapa; el cupo no puede quedar por debajo de los inscritos | U-03 | initial |
| EVE-06 | El autor avanza el evento una etapa por vez: Creación → Participación → Votación → Resultados, sin saltos ni retrocesos | U-03 | initial |
| EVE-07 | Abrir la inscripción (pasar a Participación) y abrir la votación piden una fecha estimada de cierre de hoy en adelante, que se muestra a los participantes; la web propone 3, 7 o 14 días | — | initial |
| EVE-08 | El autor pospone la fecha estimada de cierre de la etapa en curso (Participación o Votación); nunca la adelanta | U-03 | initial |
| EVE-09 | El autor pausa y reanuda el evento; mientras está pausado nadie se inscribe ni sube propuestas. Un evento cancelado no se pausa | U-03 | initial |
| EVE-10 | El autor cancela el evento, sin vuelta atrás; la api lo permite pero la web no ofrece la acción. Cancelado, el evento sigue aceptando inscripciones, propuestas, rankings y cambios de etapa | U-03 | initial |
| EVE-11 | El autor envía un recordatorio a los pendientes: en Participación, a los inscritos sin propuesta; en Votación, a los evaluadores sin ranking. No con el evento pausado o cancelado | U-03 | initial |
| EVE-12 | En la web, el autor gestiona el evento en una pantalla con las métricas (inscritos, propuestas, rankings enviados, cierre), la tabla de participantes con su email, su propuesta para descargar y si votó, y el paso siguiente de la etapa; otro usuario que entra ahí ve «No tienes permiso para gestionar este evento» | U-03 | initial |
| EVE-13 | Un administrador hace por la api todo lo que hace el autor, en cualquier evento | U-04 | initial |
| EVE-14 | En la web, avanzar de etapa, posponer el cierre y enviar recordatorios se deshabilitan con el evento pausado; el cierre se edita solo en Participación y Votación | — | initial |

## Failure conditions

| ID | Condition | Result | Origin |
|---|---|---|---|
| EVE-E1 | Faltan datos o no cumplen sus límites (nombre, descripción, cupo, etapa, tipo de recordatorio; al editar, también el organizador de hasta 200 caracteres) | `400 INVALID_PAYLOAD` o `VALIDATION_FAILED`, con el detalle | initial |
| EVE-E2 | La fecha de inicio o de fin no tiene formato `AAAA-MM-DD` | `400 INVALID_START_DATE` o `INVALID_END_DATE` | initial |
| EVE-E3 | La fecha de inicio está en el pasado | `400 PAST_START_DATE` | initial |
| EVE-E4 | El fin es anterior al inicio, o la duración es menor a 1 día o mayor a 365 | `400 INVALID_DATE_RANGE`, `DURATION_TOO_SHORT` o `DURATION_TOO_LONG` | initial |
| EVE-E5 | El cupo está fuera de 1 a 100 | `400 INVALID_MAX_PARTICIPANTS` o `MAX_PARTICIPANTS_LIMIT_EXCEEDED` | initial |
| EVE-E6 | El color de portada no es uno de los 16 | `400 INVALID_COVER_COLOR` | initial |
| EVE-E7 | Ya existe un evento con ese nombre | `409 DUPLICATE_EVENT_NAME`; la web vuelve al primer paso: «Ya existe un evento con este nombre. Elige otro.» | initial |
| EVE-E8 | El usuario de la sesión ya no existe o la sesión no trae su identidad | `401 USER_NOT_FOUND` o `UNAUTHORIZED`; `500 INVALID_TOKEN` si la identidad no se puede leer | initial |
| EVE-E9 | Se edita nombre, descripción, organizador o cupo fuera de Creación y Participación | `409 INVALID_UPDATE_STAGE`, con la etapa actual | initial |
| EVE-E10 | El cupo nuevo es menor que los inscritos | `400 MAX_PARTICIPANTS_BELOW_REGISTERED`, con la cantidad actual | initial |
| EVE-E11 | La etapa pedida no existe o no es la siguiente | `400 INVALID_STAGE` o `INVALID_TRANSITION`, con la etapa actual y la pedida | initial |
| EVE-E12 | Falta la fecha estimada de cierre, no tiene formato `AAAA-MM-DD` o es anterior a hoy | `400 MISSING_ESTIMATED_DATE`, `INVALID_DATE_FORMAT` o `INVALID_ESTIMATED_DATE` | initial |
| EVE-E13 | Se cambia el cierre de una etapa que no es la actual | `400 INVALID_STAGE_FOR_EDIT` | initial |
| EVE-E14 | La fecha nueva adelanta el cierre | `400 CANNOT_ADVANCE_DEADLINE`; la web: «El cierre solo se puede posponer…» | initial |
| EVE-E15 | El evento ya está cancelado | `409 ALREADY_CANCELLED` | initial |
| EVE-E16 | Se pausa un evento cancelado | `409 EVENT_CANCELLED` | initial |
| EVE-E17 | Se envía un recordatorio con el evento pausado o cancelado | `409 EVENT_PAUSED_OR_CANCELLED` | initial |
| EVE-E18 | El tipo de recordatorio no corresponde a la etapa (propuesta en Participación, ranking en Votación) | `409 INVALID_EVENT_STAGE`, con la etapa actual | initial |
| EVE-E19 | No hay pendientes a quienes recordar | `409 NO_PENDING_RECIPIENTS` | initial |
| EVE-E20 | Quien no es el autor ni un administrador intenta gestionar el evento | `403 FORBIDDEN` (`{error, message}`) | initial |
| EVE-E21 | El identificador del evento falta o no es válido | `400 MISSING_EVENT_ID` o `INVALID_EVENT_ID`; en las operaciones que controlan permisos, `400 BAD_REQUEST` (`{error, message}`) | initial |
| EVE-E22 | El evento no existe, o está en Creación y quien consulta no es su autor ni un administrador | `404 EVENT_NOT_FOUND`; en las operaciones que controlan permisos, `404 NOT_FOUND` (`{error, message}`) si el evento no existe | initial |
| EVE-E23 | Falla interna al leer o guardar | `500 DB_CREATE_ERROR`, `DB_UPDATE_ERROR`, `UPDATE_ERROR`, `CANCEL_ERROR`, `RETRIEVAL_ERROR` o `PARTICIPANT_CHECK_ERROR` | initial |
| EVE-E24 | Al crear el evento, el organizador supera los 200 caracteres (la creación no lo controla y la base lo rechaza) | `500 DB_CREATE_ERROR` | initial |
