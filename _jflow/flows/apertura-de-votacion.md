---
flow: apertura-de-votacion
requirements: [EVE-06, EVE-07, VOT-01, VOT-02, VOT-03, VOT-04, VOT-05, NOT-01]
modules: [api.event, api.vote, api.attachment, api.notification, api.email]
---

# Apertura de la votación

El autor pasa el evento de Participación a Votación: se valida la configuración, se reparten las propuestas y todo se guarda junto.

| # | From → To | What happens | Operation / table | Rules |
|---|---|---|---|---|
| 1 | web → api.vote | La web pide la vista previa: inscritos, propuestas, m mínimo, máximo y recomendado | `previewVotingConfiguration` | VOT-01 |
| 2 | web → api.event | El autor confirma con la fecha de cierre y la configuración | `updateEventStage` (`stage: voting`) | EVE-06, EVE-07 |
| 3 | api.event → api.attachment | Se leen las propuestas del evento; sus dueños son los evaluadores, uno por propuesta | `attachments` | VOT-03, INV-13 |
| 4 | api.event → api.vote | Se arma la configuración y se valida m contra k | `voting_configurations` | VOT-02, VOT-04 |
| 5 | api.vote | Se genera la asignación regular en memoria, sin la propuesta propia | — | VOT-05, INV-01 |
| 6 | api.vote → base | Etapa, configuración y asignaciones en una transacción; los triggers validan cada asignación | `events.stage`, `voting_configurations`, `assignments` | GAR-07, GAR-01, GAR-02, GAR-05 |
| 7 | api.event → api.notification, api.email | Aviso de cambio de etapa a cada inscrito, con si puede votar y cuántas propuestas tiene | `notifications` (`stage_changed`), email | NOT-01, NOT-11 |
