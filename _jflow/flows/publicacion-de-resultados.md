---
flow: publicacion-de-resultados
requirements: [EVE-06, RES-01, RES-02, RES-03, RES-04, RES-05, NOT-01]
modules: [api.event, api.vote, api.notification, api.email]
---

# Publicación de los resultados

El autor pasa el evento a Resultados: se calculan G, Q y G′, se guardan y quedan públicos.

| # | From → To | What happens | Operation / table | Rules |
|---|---|---|---|---|
| 1 | web → api.event | El autor publica (la web avisa si faltan rankings o hay propuestas sin evaluar) | `updateEventStage` (`stage: results`) | EVE-06 |
| 2 | api.event → api.vote | Se calculan G (MBC), Q de cada evaluador y G′ con la sanción a quien no evaluó | `votes`, `assignments`, `voting_configurations` | RES-02, RES-03, RES-04 |
| 3 | api.vote → base | Se guarda el resultado del evento; si el cálculo falla, la etapa queda cambiada y el resultado sin calcular | `voting_results` | RES-01, GAR-06 |
| 4 | api.event → api.notification, api.email | Aviso de cambio de etapa con el puesto de cada inscrito | `notifications` (`stage_changed`), email | NOT-01 |
| 5 | web → api.vote | Cualquiera lee los resultados | `getResults` | RES-05, GAR-15 |
