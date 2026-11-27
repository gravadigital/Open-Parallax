---
entity: Configuración de votación
tables: [voting_configurations]
describes: Parámetros del reparto y del cálculo de resultados de un evento
---

# Configuración de votación

## Attributes

| Attribute | Type | Required | Constraints | Personal data | Origin |
|---|---|---|---|---|---|
| evento | referencia | sí | una por evento | no | initial |
| propuestas por evaluador (m) | número | sí | 1 a 50 en la base; las cotas de VOT-04 al abrir | no | initial |
| evaluaciones mínimas por propuesta | número | sí | min(3, m) por defecto; de 1 a 20 al abrir y no más que m | no | initial |
| umbral de buen evaluador | número | sí | entre 0 y 1; 0,6 por defecto; supera al malo por al menos 0,1 | no | initial |
| umbral de mal evaluador | número | sí | entre 0 y 1; 0,3 por defecto | no | initial |
| magnitud del ajuste (n) | número | sí | 1 a 10 al abrir; 3 por defecto | no | initial |
| sanción a quien no evalúa | enum | sí | exclusion (por defecto) o penalty | no | initial |
| penalización | número | sí | 1 a 100 puestos; 10 por defecto | no | initial |

## Relationships

- Una Configuración de votación pertenece a un Evento.

## Removal

Nunca.
