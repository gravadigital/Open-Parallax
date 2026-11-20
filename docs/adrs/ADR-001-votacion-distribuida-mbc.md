# ADR-001: Votación distribuida por pares con Modified Borda Count

**Estado:** Aceptado (implementado)
**Fecha:** 2026-09-18 (documentado retroactivamente) · actualizado 2026-11-20 (auditoría contra el paper: asignación regular, sanción por no evaluar, `G'` oficial)
**Detectado desde:** `api`
**Tags:** dominio, algoritmo, núcleo-del-producto

---

## Contexto

Cuando una convocatoria recibe muchas más propuestas de las que un comité puede leer, la
evaluación se convierte en el cuello de botella. Un comité chico no da abasto; uno grande es caro
y lento de coordinar.

La alternativa es repartir la evaluación entre los propios participantes. Eso resuelve la escala
pero abre dos problemas que no existen con un comité:

1. **Conflicto de interés.** Los evaluadores son los propios competidores.
2. **Incentivo a evaluar mal.** Si evaluar no tiene consecuencias, conviene hacerlo rápido y
   superficialmente. Peor: si el ranking propio depende del ajeno, conviene hundir a los demás.

Cualquier sistema de evaluación distribuida que no ataque ambos problemas produce un ranking que
nadie puede defender.

## Decisión

Se implementa el modelo de **Merrifield & Saari (2009)** de revisión por pares distribuida, con
Modified Borda Count, en `internal/domain/vote/voting_service.go`.

El modelo tiene cuatro piezas:

1. **Asignación** — Cada uno de los `n` participantes recibe `m` propuestas de las `k` totales.
   Nadie recibe la propia (`m ≤ k−1`). La asignación es **regular**: como hay un evaluador por
   propuesta (`n = k`), cada propuesta recibe exactamente `m` evaluaciones. Se exige `m ≥ 2·log₂(k)` como condición de convergencia,
   relajada al 60% del máximo posible cuando `k ≤ 10` (la fórmula está pensada para volúmenes
   grandes).

2. **Puntaje de la propuesta (MBC)** — `MBC(f_j) = (1/(v_j(m−1))) · Σ(m − R_i(f_j))`, donde
   `R_i(f_j)` es la posición que el evaluador `i` le dio (1 = mejor) y `v_j` la cantidad de votos
   que recibió `f_j`. Promediar por `v_j` deja el resultado en `[0,1]` y hace que no dependa de
   cuántos evaluadores le tocaron ni de si votaron. Produce el **ranking global `G`**. Los empates
   se rompen por UUID, para que el orden sea determinístico (la cantidad de votos no desempata:
   no dice nada del mérito). Una propuesta sin ningún voto no tiene posición: se marca
   `unevaluated` y se lista aparte, en vez de quedar última con puntaje 0.

3. **Calidad del evaluador** — `Q_i = 1 − (1/⌊m²/2⌋) · Σ|R_i(f_j) − RelativeRank_G(f_j, A(p_i))|`,
   la normalización del paper: `⌊m²/2⌋` es el desvío de un orden exactamente invertido, que da
   `Q_i = 0`; coincidir con el consenso da `Q_i = 1`. Mide cuánto se aparta un evaluador del consenso, **restringido al subconjunto que le tocó**.
   Quien no completa su asignación no tiene `Q_i`: sus votos parciales se descartan y a su
   propuesta se le aplica la sanción configurada por evento (`exclusion`, la del paper y por
   defecto: sale de `G'`; o `penalty`: baja `non_evaluator_penalty` puestos).

4. **Incentivos** — El ranking `G` se ajusta a `G'`: la propuesta de un evaluador con
   `Q_i ≥ quality_good_threshold` sube `n` posiciones; la de uno con `Q_i ≤ quality_bad_threshold`
   baja `n`. **Evaluar bien mejora la posición de tu propia propuesta.** Como en el paper, el
   desplazamiento es **relativo a las propuestas de evaluadores neutros**, que no se mueven: la
   que sube queda delante de la neutral sobre la que cae y la que baja, detrás, así el movimiento
   es exactamente `n` puestos salvo en los extremos de la tabla.

**`G'` es el ranking oficial.** Es el que se publica como "Ranking final" y el que dan las
notificaciones. `G` se conserva como dato (mérito puro) y mide el `Q_i` de cada evaluador.

Los parámetros (`m`, umbrales, magnitud del ajuste, mínimo de evaluaciones) se configuran por
evento en `voting_configurations`.

**Implementado en:**
- `api` — `internal/domain/vote/voting_service.go` es la implementación completa
- `web` — solo presenta: el panel de ranking y el de resultados no calculan nada

## Consecuencias

### Positivas

- **La evaluación escala con la cantidad de propuestas.** El costo por participante es `m`, fijo y
  configurable, sin importar cuántas propuestas haya.
- **El conflicto de interés directo es imposible por construcción**, no por confianza.
- **Evaluar en serio conviene**, porque afecta el resultado propio. El incentivo está alineado con
  la calidad del sistema.
- **El resultado es auditable y reproducible.** Es una función determinística sobre los votos
  emitidos: dos cálculos dan el mismo orden, incluso ante empates.

### Negativas

- **Los parámetros posibles son muy acotados con pocas propuestas.** Con `k = 4`, `m` solo puede
  ser 2 o 3. El producto no sirve bien para convocatorias chicas.
- **La condición de convergencia es una restricción real.** Un organizador que quiera pedir poco
  esfuerzo por participante (bajo `m`) choca con el mínimo del modelo.
- **La calidad `Q_i` es relativa al consenso, no a la verdad.** Un evaluador que acierta contra la
  opinión mayoritaria es penalizado igual que uno que evalúa al azar. El modelo asume que el
  consenso aproxima la calidad — un supuesto del paper, heredado.
- **Es difícil de explicar.** Un participante tiene que entender que evaluar mal lo perjudica, o
  el incentivo no opera. Los resultados explican que el orden incluye el ajuste, y cada autor ve
  el suyo; las instrucciones de votación todavía no lo explican (crítico C6 de la auditoría).
- **El puntaje no siempre sigue el orden de `G'`.** El MBC que se muestra es el mérito sin ajuste.
- **El ajuste revela cómo evaluó cada autor.** Por eso la API solo muestra el ajuste de una
  propuesta (`adjustment`, `adjustment_reason`) a su autor y al organizador del evento.

## Alternativas Consideradas

**No hay registro del rationale original.** Estas son las alternativas que existían objetivamente
para el problema, no las que consta que se evaluaron.

- **Comité centralizado** — No resuelve la escala, que es el problema que originó el producto.
- **Promedio simple de puntajes** — Vulnerable a estrategia: un evaluador puede puntuar bajo a
  todos para elevar el propio. El Borda por ranking forzado lo mitiga: hay que ordenar, no se
  puede hundir a todos por igual.
- **Borda Count clásico** (sin la normalización de la variante "modified") — No produce un puntaje
  comparable entre evaluadores con distinta cantidad de asignaciones.
- **Revisión por pares sin incentivos** — Es el modelo académico tradicional, y su problema
  conocido es exactamente el que el sistema de incentivos ataca.

## Decisiones tomadas (2026-11-20)

- **`G'` es el ranking oficial** (antes "Decisión Pendiente"). Sin eso, las piezas 3 y 4 no
  tenían efecto sobre el resultado.
- **Sanción por no evaluar configurable por evento**, exclusión por defecto.
- **Visibilidad de los ajustes:** el público ve `G'` y una explicación general; cada participante
  ve el ajuste de su propuesta; el organizador ve todos.

## Referencias

- Merrifield, M. R., & Saari, D. G. (2009). *Telescope time without tears: a distributed approach
  to peer review.* Astronomy & Geophysics, 50(4).
- Implementación: `api/internal/domain/vote/voting_service.go`
- Arquitectura: `docs/architectures/api/`
- Requerimientos: F-04 en `docs/prd/requirements.md`
