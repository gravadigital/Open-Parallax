---
prefix: NOT
name: Notificaciones
surfaces: [Web]
covers: Avisos dentro de la web y por email cuando el evento cambia o cuando falta algo, y el email de recuperación de contraseña
entities: [Notificación, Evento, Inscripción]
last_requirement: NOT-11
---

# NOT — Notificaciones

## Requirements

| ID | Event | Channel | Recipient | Content | Origin |
|---|---|---|---|---|---|
| NOT-01 | El evento cambia de etapa | Web y email | Los inscritos, sin el autor | La etapa nueva; en Votación, la fecha estimada de cierre, si puede votar y cuántas propuestas le tocan; en Resultados, su puesto y el total si su propuesta tiene posición. Al pasar a Participación no hay inscritos a quienes avisar. Email (solo la etapa): «[OpenParallax] El evento «{evento}» avanzó a la etapa {etapa}» | initial |
| NOT-02 | El evento se cancela | Web y email | Los inscritos, sin el autor | Email: «[OpenParallax] El evento «{evento}» fue cancelado» | initial |
| NOT-03 | El evento se pausa | Web y email | Los inscritos, sin el autor | Web: que el organizador lo pausó. Email: «[OpenParallax] El evento «{evento}» fue pausado», que dice que durante la pausa no se puede inscribir ni enviar archivos | initial |
| NOT-04 | Se pospone la fecha estimada de cierre | Web y email | Los inscritos, sin el autor | La etapa y la nueva fecha. Email: «[OpenParallax] Cambio de fecha estimada en el evento «{evento}»» | initial |
| NOT-05 | Alguien se inscribe | Web | El autor | La inscripción; mientras no la lee, las nuevas inscripciones del mismo evento se suman a esa misma notificación, que vuelve al principio de la lista con la fecha de la última inscripción | initial |
| NOT-06 | Alguien se inscribe | Web | El inscrito | La confirmación de su inscripción | initial |
| NOT-07 | Un evaluador envía su ranking | Web | Ese evaluador | La confirmación del envío, o de que actualizó uno anterior | initial |
| NOT-08 | El autor envía el recordatorio de propuesta | Web y email | Los inscritos sin propuesta | La fecha estimada de cierre, si hay. Email: «[OpenParallax] Falta tu propuesta en el evento «{evento}»» | initial |
| NOT-09 | El autor envía el recordatorio de ranking | Web y email | Los evaluadores sin ranking | La fecha estimada de cierre, si hay. Email: «[OpenParallax] Falta tu ranking en el evento «{evento}»» | initial |
| NOT-10 | Alguien pide recuperar la contraseña | Email | El dueño de la cuenta | El enlace para definir una contraseña nueva, que vence en 1 hora. Email: «[OpenParallax] Recuperación de contraseña» | initial |
| NOT-11 | Cualquiera de las anteriores | — | — | Los avisos se envían después de la operación y no la cambian si fallan: se registran. Los emails de NOT-01 a NOT-09 se mandan en segundo plano; el de recuperación (NOT-10) se manda dentro del pedido, que espera al servidor SMTP. Por email solo si el envío está habilitado | initial |

## Not notified

- Reanudar un evento pausado.
- Recordatorios automáticos antes de un cierre: solo existen los que envía el autor.
- Preferencias de aviso por email: no existen (la web muestra interruptores que no se guardan, BAN-05).
