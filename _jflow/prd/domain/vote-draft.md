---
entity: Borrador de voto
tables: [vote_drafts]
describes: El ranking a medio armar de un evaluador, para retomarlo
---

# Borrador de voto

## Attributes

| Attribute | Type | Required | Constraints | Personal data | Origin |
|---|---|---|---|---|---|
| evento | referencia | sí | — | no | initial |
| asignación | referencia | sí | uno por asignación y evaluador | no | initial |
| evaluador | referencia | sí | — | no | initial |
| ranking | texto | sí | lista de propuesta y posición, sin validar contra la asignación (solo que cada propuesta sea un identificador válido) | no | initial |

## Relationships

- Un Borrador de voto pertenece a una Asignación y a su evaluador.

## Removal

Se borra con su evento, su asignación o su evaluador.
