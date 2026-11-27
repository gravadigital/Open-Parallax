---
entity: Propuesta
tables: [attachments]
describes: Archivo que un inscrito presenta al evento y que otros evalúan
---

# Propuesta

## Attributes

| Attribute | Type | Required | Constraints | Personal data | Origin |
|---|---|---|---|---|---|
| evento | referencia | sí | — | no | initial |
| participante | referencia | sí | una por participante y evento | no | initial |
| nombre original | texto | sí | el del archivo subido | no | initial |
| nombre guardado | texto | sí | `{evento}_{participante}_{instante}{extensión}` | no | initial |
| ubicación | texto | sí | clave del archivo en el almacenamiento | no | initial |
| tamaño | número | sí | hasta el **tamaño máximo de propuesta**; mayor que 0 solo en la base (la api no lo controla) | no | initial |
| tipo | texto | sí | JPG, PNG, GIF, WebP, PDF, TXT, DOC o DOCX, según lo declara quien sube | no | initial |
| descripción | texto | no | hasta 1000 bytes; se guarda sin espacios en los extremos | no | initial |
| fecha de subida | fecha y hora | sí | — | no | initial |

## Derived attributes

| Attribute | Calculation rule | Origin |
|---|---|---|
| votos recibidos | cantidad de Votos sobre la propuesta | initial |

## Relationships

- Una Propuesta pertenece a un Evento y a su participante (Usuario).
- Una Propuesta recibe cero o más Votos y figura en cero o más Asignaciones.

## Removal

Su dueño la elimina durante Participación (INS-07): se borra el registro y después el archivo.
