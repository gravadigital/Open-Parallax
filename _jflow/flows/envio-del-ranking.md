---
flow: envio-del-ranking
requirements: [VOT-06, VOT-07, VOT-08, INS-09, NOT-07]
modules: [api.vote, api.attachment, api.notification]
---

# Evaluación y envío del ranking

El evaluador abre sus propuestas, arma el ranking con borradores y lo envía.

| # | From → To | What happens | Operation / table | Rules |
|---|---|---|---|---|
| 1 | web → api.vote | El evaluador pide su asignación: las m propuestas como «Propuesta N» | `getAssignment` | VOT-06 |
| 2 | web → api.attachment | Descarga cada propuesta asignada con un nombre neutro | `downloadProposal` | INS-09 |
| 3 | web → api.vote | Cada cambio de orden guarda el borrador | `saveVoteDraft` → `vote_drafts` | VOT-07 |
| 4 | web → api.vote | Envía el ranking completo | `submitRanking` | VOT-08 |
| 5 | api.vote → base | Bloquea la asignación, borra los votos anteriores e inserta los nuevos en una transacción; los triggers validan cada voto, calculan su puntaje y marcan la asignación completa | `assignments` (`FOR UPDATE`), `votes` | GAR-08, GAR-03, GAR-04, GAR-14 |
| 6 | api.vote → api.notification | Confirmación al evaluador | `notifications` (`ranking_submitted`) | NOT-07 |
