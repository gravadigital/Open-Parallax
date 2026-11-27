---
entity: Usuario
tables: [users]
describes: Persona con cuenta, con contraseña, con Google o creada al inscribirse en un evento
---

# Usuario

## Attributes

| Attribute | Type | Required | Constraints | Personal data | Origin |
|---|---|---|---|---|---|
| nombre | texto | sí | 2 a 100 caracteres al registrarse; único solo entre cuentas creadas con Google | sí | initial |
| apellido | texto | no | — | sí | initial |
| email | email | sí | único (distingue mayúsculas); formato validado | sí | initial |
| contraseña | texto | no | al menos 8 caracteres; se guarda su hash bcrypt. Vacía en cuentas de Google y en las creadas al inscribirse | no | initial |
| identidad de Google | texto | no | única | sí | initial |
| rol global | enum | no | participant, organizer o admin; participant al crear la cuenta | no | initial |
| token de recuperación | texto | no | 64 caracteres hexadecimales, único, guardado tal cual | no | initial |
| vencimiento del token de recuperación | fecha y hora | no | 1 hora después de pedirlo | no | initial |

## Relationships

- Un Usuario es autor de cero o más Eventos.
- Un Usuario tiene cero o más Inscripciones, Propuestas, Asignaciones, Votos, Borradores de voto y Notificaciones.

## Removal

Nunca: no hay forma de eliminar una cuenta.

## States

| State | Type | Meaning | Origin |
|---|---|---|---|
| con contraseña | initial | Creada con email y contraseña (ACC-01) o con la contraseña definida por enlace | initial |
| con Google | initial | Creada o vinculada con Google (ACC-03, ACC-04) | initial |
| sin contraseña | initial | Creada al inscribirse en un evento (INS-01) | initial |

## Transitions

| From | To | Triggered by | Origin |
|---|---|---|---|
| sin contraseña | con contraseña | ACC-06 | initial |
| con contraseña | con Google | ACC-04 | initial |
| sin contraseña | con Google | ACC-04 | initial |
