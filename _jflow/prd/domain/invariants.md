# Invariants

| ID | Rule | Entities | Applied in | Origin |
|---|---|---|---|---|
| INV-01 | Nadie evalúa su propia propuesta | Asignación, Propuesta | VOT | initial |
| INV-02 | Una asignación tiene exactamente m propuestas, todas del mismo evento | Asignación, Configuración de votación | VOT | initial |
| INV-03 | Un voto es sobre una propuesta asignada al evaluador, con una posición de 1 a m | Voto, Asignación | VOT | initial |
| INV-04 | Un evaluador vota cada propuesta de su asignación a lo sumo una vez | Voto | VOT | initial |
| INV-05 | Un evaluador tiene a lo sumo una asignación por evento | Asignación | VOT | initial |
| INV-06 | Un evento tiene a lo sumo una configuración de votación y un registro de resultados | Configuración de votación, Resultados | VOT, RES | initial |
| INV-07 | La fecha de fin de un evento no es anterior a la de inicio, y la de inicio no se pone en el pasado | Evento | EVE | initial |
| INV-08 | Dos cuentas no comparten email, identidad de Google ni token de recuperación | Usuario | ACC | initial |
| INV-10 | Un evento no tiene más inscritos que su cupo | Evento, Inscripción | INS, EVE | initial |
| INV-11 | Dos eventos no tienen el mismo nombre | Evento | EVE | initial |
| INV-12 | Una asignación está completa si y solo si su evaluador votó todas sus propuestas, y el contador de votos de cada propuesta es la cantidad de sus votos | Asignación, Voto, Propuesta | VOT | initial |
| INV-13 | Un inscrito tiene a lo sumo una propuesta por evento | Propuesta | INS | #662 |
