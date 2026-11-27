---
prefix: INS
name: Inscripción y propuestas
surfaces: [Web]
covers: Inscribirse en un evento, subir, reemplazar y eliminar la propuesta, y ver y descargar las propuestas
entities: [Inscripción, Propuesta, Usuario]
last_requirement: INS-09
last_failure: INS-E19
---

# INS — Inscripción y propuestas

## Requirements

| ID | Requirement | Actor | Origin |
|---|---|---|---|
| INS-01 | Cualquiera, sin sesión, inscribe un nombre (2 a 100 caracteres) y un email en un evento en Participación: si el email no tiene cuenta, se crea una sin contraseña; si la tiene, esa cuenta queda inscrita. La respuesta devuelve el identificador del usuario | U-01 | initial |
| INS-02 | En la web, inscribirse pide sesión y usa el nombre y el email de la cuenta | U-02 | initial |
| INS-03 | No se inscribe con el evento pausado, con el cupo completo, al autor ni a quien ya está inscrito | — | initial |
| INS-04 | Cualquier usuario con sesión lista los inscritos de un evento, sin el autor, con nombre, email y rol | U-02 | initial |
| INS-05 | Un inscrito sube su propuesta durante Participación, con el evento sin pausar: un archivo de hasta **tamaño máximo de propuesta**, de tipo JPG, PNG, GIF, WebP, PDF, TXT, DOC o DOCX, con una descripción opcional de hasta 1000 caracteres. Una sola por inscrito y evento; el autor no sube | U-02 | initial |
| INS-06 | El autor o un administrador sube, en nombre de un inscrito, su propuesta | U-03 | initial |
| INS-07 | El dueño elimina su propuesta durante Participación; la web la reemplaza eliminándola y subiendo la nueva | U-02 | initial |
| INS-08 | El autor y los administradores ven todas las propuestas del evento; cada inscrito, solo la suya | U-03 | initial |
| INS-09 | Descargan una propuesta su dueño, el autor y los administradores, con su nombre original, y un evaluador que la tiene asignada, durante Votación o Resultados, con un nombre neutro («propuesta-N») | U-02 | initial |

## Failure conditions

| ID | Condition | Result | Origin |
|---|---|---|---|
| INS-E1 | Faltan el nombre o el email de la inscripción, o no cumplen su formato | `400 INVALID_PAYLOAD` | initial |
| INS-E2 | El evento no está en Participación al inscribirse | `400 INVALID_REGISTRATION_STAGE`, con la etapa actual | initial |
| INS-E3 | El evento está pausado al inscribirse o al subir la propuesta | `403 EVENT_PAUSED` | initial |
| INS-E4 | El email es el del autor del evento | `400 CREATOR_CANNOT_REGISTER` | initial |
| INS-E5 | El email ya está inscrito | `409 ALREADY_REGISTERED` | initial |
| INS-E6 | El cupo está completo | `400 MAX_PARTICIPANTS_REACHED` | initial |
| INS-E7 | Se sube o elimina una propuesta fuera de Participación | `400 INVALID_EVENT_STAGE`, con la etapa actual | initial |
| INS-E8 | La subida no trae archivo | `400 NO_FILE` | initial |
| INS-E9 | La descripción supera los 1000 caracteres | `400 DESCRIPTION_TOO_LONG` | initial |
| INS-E10 | El archivo supera el **tamaño máximo de propuesta** | `400 FILE_TOO_LARGE`; la web lo rechaza antes: «El archivo supera los 10 MB.» | initial |
| INS-E11 | El tipo del archivo no está permitido o su nombre no es válido | `400 INVALID_FILE_TYPE` o `INVALID_FILENAME`; la web: «Ese formato no está permitido. Usa JPG, PNG, GIF, WebP, PDF, TXT, DOC o DOCX.» | initial |
| INS-E12 | El autor intenta subir o eliminar una propuesta propia | `403 CREATOR_CANNOT_UPLOAD` o `CREATOR_CANNOT_DELETE` | initial |
| INS-E13 | Quien figura como dueño de la propuesta no está inscrito | `403 NOT_REGISTERED` | initial |
| INS-E14 | El participante no existe | `404 PARTICIPANT_NOT_FOUND` | initial |
| INS-E15 | El inscrito ya tiene una propuesta en el evento | `409 DUPLICATE_ATTACHMENT`, con la propuesta existente | initial |
| INS-E16 | La propuesta o su archivo no existen | `404 ATTACHMENT_NOT_FOUND` o `FILE_NOT_FOUND` | initial |
| INS-E17 | Quien descarga o elimina no tiene permiso sobre la propuesta | `403 FORBIDDEN` | initial |
| INS-E18 | Faltan o no son válidos los identificadores del participante, o quien opera sobre los recursos de un participante no es él, el autor ni un administrador | `400 MISSING_PARAMETERS` o `INVALID_PARTICIPANT_ID`; `403 FORBIDDEN` (`{error, message}`) | initial |
| INS-E19 | Falla interna al crear la cuenta, inscribir, guardar o borrar el archivo, o leer | `500 USER_CREATE_ERROR`, `PARTICIPANT_CHECK_ERROR`, `REGISTRATION_ERROR`, `FILE_SAVE_ERROR`, `DB_SAVE_ERROR`, `EVENT_LOOKUP_ERROR`, `DB_DELETE_ERROR` o `RETRIEVAL_ERROR` | initial |
