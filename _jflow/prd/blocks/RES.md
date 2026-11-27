---
prefix: RES
name: Resultados
surfaces: [Web]
covers: Calcular el ranking por Modified Borda Count, medir la calidad de cada evaluador, ajustar el ranking oficial y publicarlo
entities: [Resultados, Asignación, Voto, Propuesta]
last_requirement: RES-07
last_failure: RES-E5
---

# RES — Resultados

## Requirements

| ID | Requirement | Actor | Origin |
|---|---|---|---|
| RES-01 | Al pasar el evento a Resultados se calculan y guardan los resultados. Si el cálculo falla (por ejemplo, porque nadie votó), la etapa cambia igual y los resultados quedan sin calcular | — | initial |
| RES-02 | El ranking global G ordena las propuestas por su puntaje MBC, el promedio de (m − posición) / (m − 1) sobre los votos que recibió, de 0 a 1; los empates se ordenan por identificador. Una propuesta sin votos no tiene posición: se marca como no evaluada y va al final | — | initial |
| RES-03 | Quien no completó su ranking no es evaluador: sus votos se descartan. A cada evaluador se le mide la calidad Q = 1 − Σ\|posición que dio − posición relativa en G\| / ⌊m²/2⌋, de 0 (orden inverso al consenso) a 1 | — | initial |
| RES-04 | El ranking ajustado G′ es el oficial: la propuesta de un evaluador con Q mayor o igual al umbral bueno sube n puestos, y la de uno con Q menor o igual al umbral malo baja n, respecto de las propuestas que no se mueven. La de quien no evaluó se excluye (sanción exclusión) o baja los puestos de la penalización (sanción penalización) | — | initial |
| RES-05 | Cualquiera, sin sesión, ve los resultados de un evento en Resultados: G′ y G, con el autor y el nombre del archivo de cada propuesta, puntaje, votos y posición media, y la configuración si la pide; la web muestra solo G′. El ajuste de cada propuesta y su motivo solo los ven el autor, los administradores y el dueño de cada propuesta, pero quién quedó excluido por no evaluar es público (la web lo lista en «Fuera del ranking»); la calidad Q de los evaluadores no se publica | U-01 | initial |
| RES-06 | El autor, un organizador global o un administrador recalcula los resultados de un evento en Resultados, y ve la calidad de cada evaluador | U-03 | initial |
| RES-07 | En la web, en la página del evento en Resultados el inscrito ve su fila marcada «Tú» y, si completó su ranking, ese ranking solo para leer (el que guardó el navegador al enviarlo o, si no está, el último borrador); en sus eventos ve «Puesto N de M». En la gestión, si faltan los resultados, se recalculan | U-02 | initial |

## Failure conditions

| ID | Condition | Result | Origin |
|---|---|---|---|
| RES-E1 | El evento no está en Resultados (también uno en Creación ajeno, que acá no se oculta) | `403` «Results can only be viewed in the results stage», con la etapa actual | initial |
| RES-E2 | Los resultados todavía no se calcularon | `404`, con `error: RESULTS_NOT_CALCULATED` (el código va en `error`, sin `code`) y `message` | initial |
| RES-E3 | El identificador del evento falta o no es válido, o el evento no existe | `400` o `404` «Event not found» | initial |
| RES-E4 | Se recalcula un evento sin configuración de votación | `404` «Voting configuration not found» | initial |
| RES-E5 | El cálculo falla, por ejemplo porque no hay votos | `500` «Failed to calculate results» | initial |

## Examples

- RES-02: con m = 3, una propuesta que recibió las posiciones 1, 1 y 2 suma (2 + 2 + 1) = 5 puntos sobre 3 votos · (3 − 1) = 6: MBC = 0,83.
- RES-03: con m = 3, ⌊m²/2⌋ = 4; quien ordenó sus tres propuestas al revés que G se desvía 2 + 0 + 2 = 4 y su Q es 0.
- RES-04: con n = 3, la propuesta 5.ª en G de un evaluador con Q = 0,8 (umbral bueno 0,6) queda 2.ª en G′.
