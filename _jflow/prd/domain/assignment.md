---
entity: Asignación
tables: [assignments]
describes: Las m propuestas que un evaluador tiene que ordenar en un evento
---

# Asignación

## Attributes

| Attribute | Type | Required | Constraints | Personal data | Origin |
|---|---|---|---|---|---|
| evento | referencia | sí | — | no | initial |
| evaluador | referencia | sí | una por evaluador y evento | no | initial |
| propuestas | referencia | sí | exactamente m, del mismo evento, ninguna propia | no | initial |

## Derived attributes

| Attribute | Calculation rule | Origin |
|---|---|---|
| completa | verdadero cuando el evaluador votó todas sus propuestas; se actualiza con cada voto | initial |
| fecha de completitud | instante del último envío del ranking; un reenvío la renueva | initial |

## Relationships

- Una Asignación pertenece a un Evento y a un evaluador (Usuario), y reúne m Propuestas.
- Una Asignación tiene cero o más Votos y a lo sumo un Borrador de voto.

## Removal

Nunca.

## States

| State | Type | Meaning | Origin |
|---|---|---|---|
| pendiente | initial | El evaluador no envió su ranking | initial |
| completa | final | El evaluador envió su ranking | initial |

## Transitions

| From | To | Triggered by | Origin |
|---|---|---|---|
| pendiente | completa | VOT-08 | initial |
