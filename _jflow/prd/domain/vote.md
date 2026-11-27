---
entity: Voto
tables: [votes]
describes: La posición que un evaluador le da a una propuesta de su asignación
---

# Voto

## Attributes

| Attribute | Type | Required | Constraints | Personal data | Origin |
|---|---|---|---|---|---|
| evento | referencia | sí | — | no | initial |
| asignación | referencia | sí | — | no | initial |
| evaluador | referencia | sí | — | no | initial |
| propuesta | referencia | sí | asignada al evaluador; un voto por asignación y propuesta | no | initial |
| posición | número | sí | de 1 a m; 1 es la mejor | no | initial |
| fecha del voto | fecha y hora | sí | — | no | initial |

## Derived attributes

| Attribute | Calculation rule | Origin |
|---|---|---|
| puntaje | (m − posición + 1) · 100 / m | initial |

## Relationships

- Un Voto pertenece a una Asignación y a una Propuesta.

## Removal

Se reemplazan todos los votos de la asignación cuando el evaluador reenvía su ranking (VOT-08).
