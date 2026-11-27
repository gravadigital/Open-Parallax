---
flow: inscripcion
requirements: [INS-01, INS-03, NOT-05, NOT-06]
modules: [api.event, api.participant, api.notification]
---

# Inscripción en un evento

Alguien se inscribe con nombre y email; si no tiene cuenta, se crea.

| # | From → To | What happens | Operation / table | Rules |
|---|---|---|---|---|
| 1 | web → api.event | Inscripción con el nombre y el email | `registerParticipant` | INS-01 |
| 2 | api.event | Se valida etapa, pausa, cupo, autor e inscripción previa | `events`, `event_participants` | INS-03, GAR-12 |
| 3 | api.event → api.participant | Se busca la cuenta por email; si no existe, se crea sin contraseña | `users` | INS-01, GAR-10 |
| 4 | api.event → base | Se registra la inscripción | `event_participants` | — |
| 5 | api.event → api.notification | Confirmación al inscrito y aviso al autor, sumado a su aviso sin leer del mismo evento | `notifications` (`registration_confirmed`, `participant_registered`) | NOT-05, NOT-06 |
