---
entity: Notificación
tables: [notifications]
describes: Aviso dentro de la web para un usuario, sobre un evento
---

# Notificación

## Attributes

| Attribute | Type | Required | Constraints | Personal data | Origin |
|---|---|---|---|---|---|
| destinatario | referencia | sí | — | no | initial |
| evento | referencia | sí | — | no | initial |
| tipo | enum | sí | stage_changed, event_cancelled, event_paused, deadline_changed, participant_registered, registration_confirmed, ranking_submitted, file_reminder, vote_reminder | no | initial |
| datos | texto | sí | los que el tipo necesita (etapa, fecha, puesto…) | no | initial |
| leída | fecha y hora | no | — | no | initial |
| fecha | fecha y hora | sí | — | no | initial |

## Relationships

- Una Notificación pertenece a un Usuario y a un Evento.

## Removal

A los 90 días deja de mostrarse y se borra cuando su destinatario lista sus notificaciones. También se borra con su evento o su destinatario.

## States

| State | Type | Meaning | Origin |
|---|---|---|---|
| sin leer | initial | — | initial |
| leída | final | — | initial |

## Transitions

| From | To | Triggered by | Origin |
|---|---|---|---|
| sin leer | leída | BAN-03 | initial |
