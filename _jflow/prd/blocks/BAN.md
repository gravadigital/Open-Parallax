---
prefix: BAN
name: Bandeja de notificaciones
surfaces: [Web]
covers: Ver las notificaciones propias en la campana y en su página, y marcarlas como leídas
entities: [Notificación]
last_requirement: BAN-05
last_failure: BAN-E3
---

# BAN — Bandeja de notificaciones

## Requirements

| ID | Requirement | Actor | Origin |
|---|---|---|---|
| BAN-01 | Un usuario lista sus notificaciones de los últimos 90 días, de la más nueva a la más vieja, de a 1 a 50 (10 en el panel de la campana, 20 en la página), con la cantidad sin leer y el cursor de la página siguiente | U-02 | initial |
| BAN-02 | La web muestra la cantidad sin leer en la campana: la consulta cada 60 segundos con la pestaña visible, al volver a la pestaña y al navegar | U-02 | initial |
| BAN-03 | El usuario marca una notificación como leída (repetirlo no cambia nada) o todas a la vez; abrirla en la web lleva a su evento (la de inscripciones, a su gestión) | U-02 | initial |
| BAN-04 | Una notificación guarda su tipo y sus datos, no su texto: la web lo arma en el idioma elegido; un tipo que la web no conoce, o uno sin los datos que necesita, no se muestra | — | initial |
| BAN-05 | La página de notificaciones muestra interruptores «Avisarme también por email» que solo cambian en pantalla y no se guardan, y una tarjeta «Recordatorios» que anuncia un aviso 24 horas antes de cada cierre que no existe | — | initial |

## Failure conditions

| ID | Condition | Result | Origin |
|---|---|---|---|
| BAN-E1 | El tamaño de página no está entre 1 y 50, o el cursor no es una fecha y hora RFC 3339 | `400 INVALID_PAYLOAD` | initial |
| BAN-E2 | La notificación no existe, es de otro usuario o venció | `404 NOTIFICATION_NOT_FOUND` | initial |
| BAN-E3 | Falla interna al leer o guardar | `500 RETRIEVAL_ERROR` o `DB_UPDATE_ERROR` | initial |
