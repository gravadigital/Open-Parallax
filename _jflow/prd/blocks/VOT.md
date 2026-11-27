---
prefix: VOT
name: Votación distribuida
surfaces: [Web]
covers: Abrir la votación con su configuración, repartir las propuestas entre los evaluadores y que cada uno envíe su ranking
entities: [Configuración de votación, Asignación, Voto, Borrador de voto, Propuesta]
last_requirement: VOT-13
last_failure: VOT-E21
---

# VOT — Votación distribuida

## Requirements

| ID | Requirement | Actor | Origin |
|---|---|---|---|
| VOT-01 | El autor ve, antes de abrir la votación, cuántos inscritos hay y cuántos subieron propuesta, si se puede abrir, el m mínimo, máximo y recomendado, y los valores por defecto de la configuración | U-03 | initial |
| VOT-02 | El autor abre la votación con m propuestas por evaluador y, si quiere, los umbrales de calidad (bueno 0,6 y malo 0,3 por defecto), la magnitud del ajuste n (1 a 10, 3 por defecto), las evaluaciones mínimas por propuesta (min(3, m) por defecto) y la sanción para quien no evalúa (exclusión por defecto, o penalización de 1 a 100 puestos, 10 por defecto). La etapa, la configuración y las asignaciones se guardan juntas o no se guarda nada | U-03 | initial |
| VOT-03 | Abrir la votación exige al menos 3 propuestas. Evalúan los inscritos que subieron propuesta; el resto no evalúa ni es evaluado | — | initial |
| VOT-04 | Con k propuestas, m cumple m ≤ k−1 y m ≥ ⌈2·log₂k⌉, cota que se relaja a ⌈0,6·(k−1)⌉ cuando k ≤ 10. Las evaluaciones mínimas por propuesta no superan m. Los umbrales van de 0 a 1 y el bueno supera al malo por al menos 0,1 | — | initial |
| VOT-05 | La asignación es regular y al azar: cada evaluador recibe m propuestas y cada propuesta tiene m evaluadores, nunca su autor; cada apertura da un reparto distinto | — | initial |
| VOT-06 | Cada evaluador ve su asignación durante Votación y Resultados: las m propuestas como «Propuesta N», con tipo, tamaño y descripción, sin su autor | U-02 | initial |
| VOT-07 | El evaluador guarda un borrador de su ranking (la web lo hace sola, medio segundo después de cada cambio, y otra vez con el orden enviado después de enviarlo) y lo recupera al volver. La api acepta el borrador en cualquier etapa mientras exista la asignación, sin validarlo contra ella | U-02 | initial |
| VOT-08 | El evaluador envía su ranking durante Votación: todas sus propuestas asignadas, cada una una vez, en las posiciones 1 a m (1 es la mejor). Reenviarlo reemplaza entero el anterior. La api lo acepta también con el evento pausado; la web no muestra el ranking con el evento pausado o cancelado, y solo permite reenviarlo si cambió el orden | U-02 | initial |
| VOT-09 | El autor y los administradores, por la api (la web no lo ofrece), ven la asignación, guardan el borrador y envían el ranking en nombre de cualquier inscrito | U-03 | initial |
| VOT-10 | Cualquiera, sin sesión, ve por la api el avance de la votación: votos, asignaciones completas, porcentaje y quién envió su ranking (la web lo muestra solo en la gestión). Un evento en Creación ajeno responde 404 | U-01 | initial |
| VOT-11 | El autor consulta la configuración de votación aplicada al evento | U-03 | initial |
| VOT-12 | Configurar la votación y generar las asignaciones como pasos separados siguen en la api (la web no los usa) para el autor, un organizador global o un administrador. Configurar se acepta en Participación y en Votación; una configuración creada así en Participación se reemplaza sin aviso al abrir la votación, y en Votación ya existe (VOT-E16). Generar asignaciones solo se acepta en Votación, cuando la apertura ya las creó: siempre termina en VOT-E17 | U-03 | initial |
| VOT-13 | Un organizador global configura la votación, genera asignaciones y recalcula los resultados de cualquier evento | U-05 | initial |

## Failure conditions

| ID | Condition | Result | Origin |
|---|---|---|---|
| VOT-E1 | Hay menos de 3 propuestas | `400 INSUFFICIENT_ATTACHMENTS`, con la cantidad actual y el mínimo (3); la web lo anticipa con la vista previa: «Se necesitan al menos 3 participantes con propuesta…», y si lo responde la api muestra «Se necesitan al menos 3 propuestas para abrir la votación.» | initial |
| VOT-E2 | Se abre la votación sin configuración | `400 MISSING_VOTING_CONFIG` | initial |
| VOT-E3 | Los umbrales de calidad no cumplen la regla | `400 INVALID_THRESHOLDS` | initial |
| VOT-E4 | m es mayor que las propuestas que cada evaluador puede recibir | `400 M_EXCEEDS_EVALUABLE` | initial |
| VOT-E5 | m está fuera de las cotas de la regla, o las evaluaciones mínimas por propuesta superan m | `400 MATH_CONSTRAINT_VIOLATION` o `CONFIG_INVALID` | initial |
| VOT-E6 | La sanción o su penalización no son válidas | `400 INVALID_SANCTION` | initial |
| VOT-E7 | Se pide la asignación fuera de Votación y Resultados | `400`, con la etapa actual | initial |
| VOT-E8 | El participante no está inscrito en el evento | `403` «Participant is not registered for this event» | initial |
| VOT-E9 | El participante no tiene asignación | `404 NO_ASSIGNMENT` (asignación) o `ASSIGNMENT_NOT_FOUND` (borrador) | initial |
| VOT-E10 | Se envía el ranking fuera de Votación | `400` «Voting is only allowed during voting stage»; la web: «La votación ya cerró. No se puede modificar el ranking.» | initial |
| VOT-E11 | El ranking no incluye cada propuesta asignada una vez, repite posiciones, no las numera de 1 a m, incluye una propuesta no asignada o de otro evento, o cita otra asignación | `400`, con el motivo (y la cantidad esperada y la recibida si no coinciden) | initial |
| VOT-E12 | La asignación o una propuesta del ranking no existen | `404` | initial |
| VOT-E13 | No hay borrador guardado | `404 DRAFT_NOT_FOUND` | initial |
| VOT-E14 | El borrador o el ranking, o sus identificadores, no son válidos | borrador: `400 INVALID_PAYLOAD`, `INVALID_ATTACHMENT_ID`, `INVALID_PARTICIPANT_ID` o `INVALID_EVENT_ID`; ranking: `400` sin código, «Invalid request payload» o «Invalid … format» | initial |
| VOT-E15 | El evento no tiene configuración de votación | `404 CONFIG_NOT_FOUND` | initial |
| VOT-E16 | Ya hay una configuración (paso separado) | `409 CONFIG_EXISTS` | initial |
| VOT-E17 | Ya hay asignaciones (paso separado) | `409 ASSIGNMENTS_EXIST` | initial |
| VOT-E18 | Se configura fuera de Participación y Votación, o se generan asignaciones fuera de Votación | `400 INVALID_EVENT_STAGE` | initial |
| VOT-E19 | Quien configura o genera asignaciones no es el autor, un organizador global ni un administrador | `403 FORBIDDEN` (`{error, message}`) | initial |
| VOT-E20 | Falla interna al preparar o guardar la votación, el ranking o el borrador, o al leer | `500 VOTING_SETUP_ERROR`, `ATTACHMENTS_ERROR`, `PARTICIPANTS_ERROR`, `DB_SAVE_ERROR`, `GENERATION_FAILED`, `CONFIG_LOOKUP_ERROR`, `RETRIEVAL_ERROR`, `DRAFT_SAVE_ERROR`, `DRAFT_GET_ERROR`, «Failed to save votes», «Failed to get votes» o «Failed to get assignments» (también con un identificador de evento mal formado en las estadísticas) | initial |
| VOT-E21 | Un valor de la configuración está fuera de rango: m fuera de 1 a 50, un umbral fuera de 0 a 1, n fuera de 1 a 10 o las evaluaciones mínimas fuera de 1 a 20 | `400 INVALID_PAYLOAD`, con el detalle | initial |

## Examples

- VOT-04: con k = 4 propuestas, m va de ⌈0,6·3⌉ = 2 a 3; con k = 3, m = 2.
- VOT-08: con m = 3, un ranking válido da a sus tres propuestas las posiciones 1, 2 y 3; uno con 1, 1 y 2 se rechaza (VOT-E11).
