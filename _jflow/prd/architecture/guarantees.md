# Guarantees

| ID | Guarantee | Rule | Mechanism | Where | Origin |
|---|---|---|---|---|---|
| GAR-01 | La asignación nunca da a un evaluador su propia propuesta | INV-01, VOT-05 | El reparto la excluye en Go, y el trigger `validate_assignment_constraints` rechaza en la base cualquier asignación que la incluya | api.vote, la base | initial |
| GAR-02 | Una asignación con otra cantidad que m, o con propuestas de otro evento, no se guarda | INV-02 | Trigger `validate_assignment_constraints` (cuando el evento tiene configuración) | la base | initial |
| GAR-03 | Un voto fuera de la asignación o con posición mayor que m no se guarda | INV-03 | Trigger `validate_vote_constraints`, que además calcula el puntaje; `CHECK` de posición mayor que 0 | la base | initial |
| GAR-04 | Un segundo voto sobre la misma propuesta de una asignación falla | INV-04 | Índice único `uq_votes_assignment_attachment` | la base | initial |
| GAR-05 | Una segunda asignación del mismo evaluador en el evento falla | INV-05 | Restricción única `unique_participant_event_assignment`; por ella también falla una segunda apertura concurrente de la votación | la base | initial |
| GAR-06 | La configuración y los resultados son uno por evento | INV-06 | Índices únicos sobre `event_id`; los resultados se reemplazan al recalcular | la base | initial |
| GAR-07 | Abrir la votación guarda la etapa, la configuración y las asignaciones o no guarda nada | VOT-02 | Una transacción | api.vote | initial |
| GAR-08 | Reenviar el ranking reemplaza los votos anteriores enteros, sin mezclar envíos concurrentes | VOT-08, INV-04 | Bloqueo de la fila de la asignación (`FOR UPDATE`), y borrado e inserción de los votos en una transacción | api.vote | initial |
| GAR-09 | Las fechas del evento son coherentes | INV-07, EVE-03 | `CHECK valid_event_dates` (fin ≥ inicio) y trigger `enforce_future_start_date`, que rechaza un inicio pasado al crear o al cambiarlo | la base | initial |
| GAR-10 | Email, identidad de Google y token de recuperación no se repiten | INV-08 | Índices únicos (el email se compara tal cual, distinguiendo mayúsculas) | la base | initial |
| GAR-12 | El cupo no se supera | INV-10, INS-03 | Conteo previo en Go, sin bloqueo: dos inscripciones simultáneas pueden superarlo | api.event | initial |
| GAR-13 | El nombre del evento no se repite | INV-11, EVE-04 | Consulta previa en Go, sin restricción en la base | api.event | initial |
| GAR-14 | La completitud de la asignación y el contador de votos siguen a los votos | INV-12 | Triggers `update_assignment_completion` y `update_attachment_vote_count` después de insertar o borrar un voto | la base | initial |
| GAR-15 | Los resultados no se ven antes de tiempo | RES-05, RES-E1 | Cada operación de resultados comprueba que el evento esté en Resultados | api.vote | initial |
