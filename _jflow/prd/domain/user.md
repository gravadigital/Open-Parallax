---
entity: Usuario
tables: [users]
describes: Persona con cuenta, con contraseña, con Google o creada al inscribirse en un evento
---

# Usuario

## Attributes

| Attribute | Type | Required | Constraints | Personal data | Origin |
|---|---|---|---|---|---|
| nombre | texto | sí | 2 a 100 caracteres al registrarse o inscribirse; con Google, además, sin tabulaciones ni saltos de línea y sin coincidir (exacto, distingue mayúsculas) con el de ninguna cuenta; fuera de ese caso puede repetirse | sí | initial |
| apellido | texto | no | — | sí | initial |
| email | email | sí | único (distingue mayúsculas); formato validado por la api y, además, por la base (solo letras, dígitos y `._%+-`, dominio con terminación de 2 o más letras) | sí | initial |
| contraseña | texto | no | al menos 8 caracteres y hasta 72 bytes; se guarda su hash bcrypt. Vacía en cuentas de Google y en las creadas al inscribirse, hasta que se define con el enlace de recuperación | no | initial |
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
| con Google y contraseña | final | Cuenta de Google que además definió una contraseña con el enlace (ACC-06) | initial |

## Transitions

| From | To | Triggered by | Origin |
|---|---|---|---|
| sin contraseña | con contraseña | ACC-06 | initial |
| con contraseña | con Google | ACC-04 | initial |
| sin contraseña | con Google | ACC-04 | initial |
| con Google | con Google y contraseña | ACC-06 | initial |
