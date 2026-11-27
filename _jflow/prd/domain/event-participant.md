---
entity: Inscripción
tables: [event_participants]
describes: Vínculo de un Usuario con un Evento, como su creador o como participante inscrito
---

# Inscripción

## Attributes

| Attribute | Type | Required | Constraints | Personal data | Origin |
|---|---|---|---|---|---|
| evento | referencia | sí | — | no | initial |
| usuario | referencia | sí | una por evento y usuario | no | initial |
| rol en el evento | enum | sí | creator (el autor, al crear el evento) o participant (al inscribirse) | no | initial |
| fecha de inscripción | fecha y hora | sí | — | no | initial |

## Relationships

- Una Inscripción une un Evento con un Usuario.

## Removal

Nunca: no hay forma de desinscribirse.
