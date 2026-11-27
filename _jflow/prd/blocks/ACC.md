---
prefix: ACC
name: Cuentas y acceso
surfaces: [Web]
covers: Crear una cuenta con email o con Google, iniciar sesión, recuperar la contraseña y mantener la sesión
entities: [Usuario]
last_requirement: ACC-11
last_failure: ACC-E14
---

# ACC — Cuentas y acceso

## Requirements

| ID | Requirement | Actor | Origin |
|---|---|---|---|
| ACC-01 | Un visitante crea una cuenta con nombre (2 a 100 caracteres), apellido opcional, email y contraseña, y queda con la sesión iniciada | U-01 | initial |
| ACC-02 | Un usuario inicia sesión con email y contraseña y recibe una sesión que vale 24 horas, sin renovación | U-02 | initial |
| ACC-03 | Un visitante entra con su cuenta de Google: si ya tiene cuenta, inicia sesión; si es nueva, elige el nombre que lo identifica (3 a 100 caracteres en la web) y se crea la cuenta, sin contraseña | U-01 | initial |
| ACC-04 | Al entrar con Google, si no hay una cuenta con esa identidad de Google pero sí una con el mismo email, la identidad de Google queda vinculada a esa cuenta | — | initial |
| ACC-05 | Un usuario pide recuperar la contraseña con su email y recibe por email un enlace que vence en 1 hora y sirve una sola vez; la respuesta es la misma exista o no la cuenta, pero si existe tarda lo que tarda el envío del email | U-01 | initial |
| ACC-06 | Con el enlace, el usuario define una contraseña nueva; es también como una cuenta creada al inscribirse en un evento define su primera contraseña | U-01 | initial |
| ACC-07 | Toda cuenta nueva tiene el rol global participant; los roles admin y organizer solo se asignan escribiendo en la base | — | initial |
| ACC-08 | Un usuario con sesión consulta el nombre, apellido, email y rol de cualquier usuario por su identificador | U-02 | initial |
| ACC-09 | Las contraseñas tienen al menos 8 caracteres y hasta 72 bytes | — | initial |
| ACC-10 | En la web, cuando la sesión vence, se cierra sin recargar la página; una pantalla que requiere sesión lleva a iniciar sesión y, después, vuelve a donde estaba | U-02 | initial |
| ACC-11 | El nombre elegido al crear una cuenta con Google no puede coincidir con el de otra cuenta | — | initial |

## Failure conditions

| ID | Condition | Result | Origin |
|---|---|---|---|
| ACC-E1 | Faltan datos o no cumplen su formato (nombre, email, contraseña de menos de 8 caracteres, token) | `400 INVALID_PAYLOAD`, con el detalle de cada campo (en la recuperación, sin detalle) | initial |
| ACC-E2 | La contraseña supera los 72 bytes (una de menos de 8 caracteres es ACC-E1) | `400 INVALID_PASSWORD`; la web: «La contraseña no cumple los requisitos. Elige una más segura.» | initial |
| ACC-E3 | Ya hay una cuenta con ese email | `409 EMAIL_ALREADY_EXISTS`; la web: «Ya hay una cuenta con este email. Inicia sesión o recupera tu contraseña.» | initial |
| ACC-E4 | El email no tiene cuenta o la contraseña no es la suya | `401 INVALID_CREDENTIALS`; la web: «El correo o la contraseña no son correctos.» | initial |
| ACC-E5 | La cuenta se creó con Google y no tiene contraseña | `401 OAUTH_ACCOUNT_NO_PASSWORD` | initial |
| ACC-E6 | La cuenta se creó al inscribirse en un evento y todavía no tiene contraseña | `401 ACCOUNT_NO_PASSWORD` | initial |
| ACC-E7 | El enlace de recuperación no existe o ya se usó | `400 INVALID_RESET_TOKEN`; la web: «El enlace no es válido o venció» | initial |
| ACC-E8 | El enlace de recuperación venció | `400 EXPIRED_RESET_TOKEN` | initial |
| ACC-E9 | El token de Google es inválido, venció, se emitió para otra aplicación o su email no está verificado | `401 INVALID_GOOGLE_TOKEN` | initial |
| ACC-E10 | Google no responde, o el login con Google no está configurado | `500 GOOGLE_API_ERROR` | initial |
| ACC-E11 | El nombre elegido con Google ya lo usa otra cuenta | `409 USERNAME_ALREADY_EXISTS`; la web: «Ese nombre ya está en uso. Prueba con otro.» | initial |
| ACC-E12 | El usuario consultado no existe | `404 USER_NOT_FOUND` | initial |
| ACC-E13 | Una operación que requiere sesión llega sin ella, con un formato inválido o con un token inválido o vencido | `401 UNAUTHORIZED` (`{error, message}`); la web cierra la sesión | initial |
| ACC-E14 | Falla interna al crear la cuenta, guardar el enlace o la contraseña, o firmar la sesión | `500 CREATION_ERROR`, `TOKEN_GENERATION_ERROR`, `DB_ERROR` o `INTERNAL_ERROR` | initial |
