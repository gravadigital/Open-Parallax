# Integration

## Persistence and infrastructure

| Component | Technology | Owner | Used by | Origin |
|---|---|---|---|---|
| Base de datos | PostgreSQL 15, con reglas en funciones y triggers | api | api | initial |
| Almacenamiento de propuestas | S3 (Garage en local), bucket privado; o disco local | api.attachment | api | initial |
| Correo | SMTP con STARTTLS o TLS y autenticación (Mailpit en local) | api.email | api | initial |
| Sesión del navegador | localStorage (token, usuario, idioma, borrador del asistente, último ranking enviado) | web | web | initial |

## Communication

| From → To | Protocol | Type | Purpose | Origin |
|---|---|---|---|---|
| web → api | HTTP + JSON, sesión por `Authorization: Bearer` | síncrona | Toda la funcionalidad | initial |
| web → api | HTTP, consulta periódica cada 60 s | síncrona | Cantidad de notificaciones sin leer | initial |
| api → Google | HTTPS | síncrona | Verificar el token de Google | initial |
| api → SMTP | SMTP | asíncrona, sin reintento | Emails | initial |
| api → S3 | HTTP (API S3) | síncrona | Guardar, leer y borrar archivos | initial |

## External integrations

| ID | System | Direction | Data | When | Source of truth | Reference | Origin |
|---|---|---|---|---|---|---|---|
| INT-01 | Google Identity (OAuth 2.0) | entrada | Identidad, email verificado y nombre del usuario | Al entrar con Google (ACC-03) | Google, para la identidad | https://developers.google.com/identity/protocols/oauth2 | initial |
| INT-02 | Servidor SMTP | salida | Emails de NOT | En cada aviso por email | — | — | initial |
| INT-03 | Almacenamiento S3 (Garage) | salida y entrada | Archivos de las propuestas | Al subir, descargar o borrar una propuesta | El almacenamiento, para el archivo | https://garagehq.deuxfleurs.fr/documentation/ | initial |
