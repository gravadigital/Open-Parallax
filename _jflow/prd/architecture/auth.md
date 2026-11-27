# Auth

## Credentials

| ID | Credential | Issuer | Obtained | Lifetime and renewal | Revocation | Origin |
|---|---|---|---|---|---|---|
| CRED-01 | Sesión (JWT HS256 con `user_id`, `email`, `role`, firmado con `JWT_SECRET`) | api | Al crear la cuenta, iniciar sesión, o entrar con Google | 24 horas; no se renueva | No se revoca: sigue valiendo hasta vencer, aun tras cambiar la contraseña | initial |
| CRED-02 | Contraseña (hash bcrypt) | el usuario | Al crear la cuenta o por el enlace de recuperación | Sin vencimiento | Se reemplaza con el enlace de recuperación | initial |
| CRED-03 | Enlace de recuperación (token aleatorio de 32 bytes) | api | Al pedir recuperar la contraseña, por email | 1 hora; un solo uso | Se consume al usarse; uno nuevo reemplaza al anterior | initial |
| CRED-04 | Token de acceso de Google | Google | En la web, con el flujo implícito de Google | El de Google | El de Google | initial |

## Use

| Credential | Role | From | Passes through | Verifies | Origin |
|---|---|---|---|---|---|
| CRED-01 | participant, organizer, admin | web | `Authorization: Bearer` | api: firma y vencimiento; el rol se toma del token sin consultar la base | initial |
| CRED-02 | participant, organizer, admin | web | cuerpo de `POST /api/v1/users/authenticate` | api.participant | initial |
| CRED-03 | — | email → web | cuerpo de `POST /api/v1/users/reset-password` | api.participant | initial |
| CRED-04 | — | web | cuerpo de `POST /api/v1/auth/google/verify` y `/register` | api, con Google: audiencia igual a `GOOGLE_CLIENT_ID` y email verificado | initial |

## Roles

| Role | Actors | Allows | Origin |
|---|---|---|---|
| participant | U-02, U-03 | Todo lo que no es exclusivo de los otros roles; gestionar los eventos de los que es autor | initial |
| organizer | U-05 | Además: configurar la votación, generar asignaciones y recalcular los resultados de cualquier evento | initial |
| admin | U-04 | Además: actuar como autor en cualquier evento y ver eventos en Creación ajenos | initial |

## External identity providers

| Provider | Protocol | For which roles | Data it takes | How it is linked to the account | Reference | Origin |
|---|---|---|---|---|---|---|
| Google | OAuth 2.0 (token de acceso; tokeninfo y userinfo) | participant | Identidad (`sub`), email verificado, nombre | Por la identidad de Google; si no existe, por el email de una cuenta existente (ACC-04) | INT-01 | initial |
