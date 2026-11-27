---
entity: Resultados
tables: [voting_results]
describes: Los rankings calculados de un evento y la calidad de sus evaluadores
---

# Resultados

## Attributes

| Attribute | Type | Required | Constraints | Personal data | Origin |
|---|---|---|---|---|---|
| evento | referencia | sí | uno por evento; se reemplaza al recalcular | no | initial |
| ranking global (G) | texto | sí | por propuesta: puntaje MBC, posición, votos, posición media, si quedó sin evaluar | no | initial |
| calidad de los evaluadores (Q) | texto | sí | por evaluador que completó su ranking | no | initial |
| ranking ajustado (G′) | texto | sí | por propuesta: posición ajustada, ajuste y su motivo (buen evaluador, mal evaluador, no evaluó), si quedó excluida | no | initial |
| participantes | número | sí | mayor que 0 | no | initial |
| votos | número | sí | 0 o más | no | initial |
| propuestas por evaluador (m) | número | sí | — | no | initial |
| fecha de cálculo | fecha y hora | sí | — | no | initial |

## Relationships

- Un registro de Resultados pertenece a un Evento.

## Removal

Nunca.
