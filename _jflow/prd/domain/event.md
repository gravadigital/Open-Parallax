---
entity: Evento
tables: [events]
describes: Convocatoria que recibe propuestas, las reparte para evaluar y publica el ranking
---

# Evento

## Attributes

| Attribute | Type | Required | Constraints | Personal data | Origin |
|---|---|---|---|---|---|
| nombre | texto | sí | 3 a 200 caracteres; único entre todos los eventos (comparación exacta, solo en la api) | no | initial |
| descripción | texto | sí | 10 a 2000 caracteres | no | initial |
| organizador | texto | no | hasta 200 caracteres (la base; la api lo controla solo al editar); si falta se muestra el nombre del autor | no | initial |
| autor | referencia | sí | el Usuario que lo creó | no | initial |
| fecha de inicio | fecha | sí | hoy (el día local del servidor) o después al crear; no se puede mover al pasado | no | initial |
| fecha de fin | fecha | sí | no anterior al inicio; de 1 a 365 días después al crear | no | initial |
| etapa | enum | sí | creation, participation, voting o results; creation al crear | no | initial |
| cupo | número | sí | 1 a 100; 20 por defecto | no | initial |
| cierre estimado de Participación | fecha | no | se fija al abrir la inscripción, de hoy (el día UTC) en adelante; solo se pospone mientras no pasó | no | initial |
| cierre estimado de Votación | fecha | no | se fija al abrir la votación, de hoy (el día UTC) en adelante; solo se pospone mientras no pasó | no | initial |
| enlace para compartir | texto | sí | `/events/{id}`, único | no | initial |
| pausado | booleano | sí | falso al crear | no | initial |
| cancelado | booleano | sí | falso al crear; no vuelve a falso | no | initial |
| color de portada | enum | sí | uno de 16: lavender, periwinkle, sky, ice, aqua, mint, pistachio, lemon, butter, peach, coral, rose, pink, lilac, sand, mist; lavender por defecto | no | initial |

## Relationships

- Un Evento tiene un autor (Usuario) y cero o más Inscripciones y Propuestas.
- Un Evento tiene a lo sumo una Configuración de votación y un registro de Resultados, y cero o más Asignaciones, Votos y Notificaciones.

## Removal

Nunca: no hay forma de eliminar un evento. Se cancela (EVE-10).

## States

| State | Type | Meaning | Origin |
|---|---|---|---|
| Creación | initial | Solo lo ven su autor y los administradores | initial |
| Participación | intermediate | Inscripción abierta y carga de propuestas | initial |
| Votación | intermediate | Cada evaluador ordena las propuestas asignadas | initial |
| Resultados | final | Resultados calculados y públicos | initial |

## Transitions

| From | To | Triggered by | Origin |
|---|---|---|---|
| Creación | Participación | EVE-06, EVE-07 | initial |
| Participación | Votación | EVE-06, VOT-02 | initial |
| Votación | Resultados | EVE-06, RES-01 | initial |
