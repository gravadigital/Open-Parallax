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
| participante | referencia | sí | una por participante y evento (INV-13) | no | #662 |
| nombre original | texto | sí | el del archivo subido | no | initial |
| nombre guardado | texto | sí | `{evento}_{participante}_{identificador único}{extensión}`: no se repite entre subidas | no | #662 |
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

Su dueño la elimina durante Participación (INS-07): se borra el registro y después el archivo. Al desplegar #662 se eliminaron, con su archivo, las de más de cada inscrito en un evento, salvo la más reciente (DEC-10).
