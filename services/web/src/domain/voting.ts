import type { NonEvaluatorSanction, VotingConfiguration } from '../types';

/** Diferencia mínima entre el umbral bueno y el malo (igual al backend). */
export const MIN_THRESHOLD_GAP = 0.1;

export const DEFAULT_THRESHOLDS = {
  quality_good_threshold: 0.6,
  quality_bad_threshold: 0.3,
  adjustment_magnitude: 3,
} as const;

export type ThresholdIssue = 'not_greater' | 'gap_too_small';

/**
 * Replica la validación del backend en centésimos: `0.7 - 0.6` en JS da
 * `0.0999…` y daría un falso error.
 */
export const validateThresholds = (good: number, bad: number): ThresholdIssue | null => {
  const goodCents = Math.round(good * 100);
  const badCents = Math.round(bad * 100);
  if (goodCents <= badCents) return 'not_greater';
  if (goodCents - badCents < Math.round(MIN_THRESHOLD_GAP * 100)) return 'gap_too_small';
  return null;
};

/** `min_evaluations_per_file` recomendado: min(3, m). */
export const recommendedMinEvaluations = (m: number): number => Math.min(3, m);

/** Sanción por defecto para quien no evalúa: queda fuera del ranking (igual al backend). */
export const DEFAULT_SANCTION: NonEvaluatorSanction = 'exclusion';
export const DEFAULT_NON_EVALUATOR_PENALTY = 10;
export const PENALTY_MIN = 1;
export const PENALTY_MAX = 100;

export const ADJUSTMENT_MIN = 1;
export const ADJUSTMENT_MAX = 10;
export const MIN_EVALUATIONS_MAX = 20;
export const THRESHOLD_STEP = 0.05;

export interface VotingConfigDraft {
  attachments_per_evaluator: number;
  min_evaluations_per_file: number;
  adjustment_magnitude: number;
  quality_good_threshold: number;
  quality_bad_threshold: number;
  non_evaluator_sanction: NonEvaluatorSanction;
  non_evaluator_penalty: number;
}

/** Borrador inicial: m recomendado, mínimo recomendado y defaults del preview. */
export const initialVotingDraft = (p: VotingConfigPreview): VotingConfigDraft => ({
  attachments_per_evaluator: p.recommended_m,
  min_evaluations_per_file: recommendedMinEvaluations(p.recommended_m),
  adjustment_magnitude: p.defaults.adjustment_magnitude,
  quality_good_threshold: p.defaults.quality_good_threshold,
  quality_bad_threshold: p.defaults.quality_bad_threshold,
  non_evaluator_sanction: p.defaults.non_evaluator_sanction ?? DEFAULT_SANCTION,
  non_evaluator_penalty: p.defaults.non_evaluator_penalty ?? DEFAULT_NON_EVALUATOR_PENALTY,
});

export type VotingDraftIssue = ThresholdIssue | 'adjustment_out_of_range' | 'penalty_out_of_range';

export const validateVotingDraft = (d: VotingConfigDraft): VotingDraftIssue | null => {
  const thresholds = validateThresholds(d.quality_good_threshold, d.quality_bad_threshold);
  if (thresholds) return thresholds;
  if (d.adjustment_magnitude < ADJUSTMENT_MIN || d.adjustment_magnitude > ADJUSTMENT_MAX) {
    return 'adjustment_out_of_range';
  }
  if (
    d.non_evaluator_sanction === 'penalty' &&
    (d.non_evaluator_penalty < PENALTY_MIN || d.non_evaluator_penalty > PENALTY_MAX)
  ) {
    return 'penalty_out_of_range';
  }
  return null;
};

const cents = (n: number): number => Math.round(n * 100);

/** true si la configuración coincide con la recomendada (umbrales comparados en centésimos). */
export const isRecommendedConfig = (
  c: Pick<
    VotingConfiguration,
    | 'attachments_per_evaluator'
    | 'min_evaluations_per_file'
    | 'adjustment_magnitude'
    | 'quality_good_threshold'
    | 'quality_bad_threshold'
    | 'non_evaluator_sanction'
  >
): boolean =>
  cents(c.quality_good_threshold) === cents(DEFAULT_THRESHOLDS.quality_good_threshold) &&
  cents(c.quality_bad_threshold) === cents(DEFAULT_THRESHOLDS.quality_bad_threshold) &&
  c.adjustment_magnitude === DEFAULT_THRESHOLDS.adjustment_magnitude &&
  c.min_evaluations_per_file === recommendedMinEvaluations(c.attachments_per_evaluator) &&
  (c.non_evaluator_sanction ?? DEFAULT_SANCTION) === DEFAULT_SANCTION;

/**
 * Respuesta de `GET /api/v1/events/{event_id}/voting-config/preview`.
 * `recommended_m` y `max_m` vienen de ese endpoint; no se recalculan en el cliente (DA-5).
 */
export interface VotingConfigPreview {
  participants_count: number;
  participants_with_proposal: number;
  can_open_voting: boolean;
  min_m: number;
  max_m: number;
  recommended_m: number;
  defaults: {
    quality_good_threshold: number;
    quality_bad_threshold: number;
    adjustment_magnitude: number;
    non_evaluator_sanction?: NonEvaluatorSanction;
    non_evaluator_penalty?: number;
  };
}
