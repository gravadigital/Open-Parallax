import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { renderWithProviders } from '../../../test-utils/renderWithProviders';
import EventResults from './EventResults';
import { DistributedVotingService } from '../../../services/api';
import { ApiError } from '../../../config/api';
import type { AttachmentResult, VotingResults } from '../../../types';

jest.mock('../../../services/api');

const getResults = DistributedVotingService.getDistributedResults as jest.Mock;
const recalculate = DistributedVotingService.recalculateDistributedResults as jest.Mock;

const r = (id: string, pid: string, name: string, file: string, score: number, rank: number): AttachmentResult => ({
  attachment_id: id,
  filename: file,
  participant_id: pid,
  participant_name: name,
  mbc_score: score,
  global_rank: rank,
  adjusted_rank: rank,
  vote_count: 3,
  average_rank: rank,
});

const results: VotingResults = {
  id: 'vr-1',
  event_id: 'e-1',
  global_ranking: [],
  participant_qualities: {},
  adjusted_ranking: [
    r('a-5', 'u-5', 'Eva Torres', 'cielo.gif', 0.22, 5),
    r('a-2', 'u-2', 'Bruno Ríos', 'sol.png', 0.74, 1),
    r('a-1', 'u-1', 'Ana Pérez', 'afiche.pdf', 0.4, 4),
    r('a-3', 'u-3', 'Carla Méndez', 'luna.pdf', 0.68, 2),
    r('a-4', 'u-4', 'Diego Sosa', 'mar.jpg', 0.51, 3),
  ],
  total_participants: 5,
  total_votes: 12,
  attachments_per_evaluator: 3,
  calculated_at: '2026-10-04T18:00:00Z',
};

describe('EventResults', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.spyOn(console, 'error').mockImplementation(() => undefined);
  });

  it('TS-46: ranking final lista, stats y onLoaded', async () => {
    getResults.mockResolvedValue(results);
    const onLoaded = jest.fn();
    renderWithProviders(<EventResults eventId="e-1" currentUserId="u-1" onLoaded={onLoaded} />);
    expect(screen.getByRole('status')).toHaveTextContent('Cargando resultados…');
    expect(await screen.findByText('Ranking final')).toBeInTheDocument();
    expect(screen.getByText('5 participantes · 12 evaluaciones')).toBeInTheDocument();
    const ranking = screen.getByRole('list', { name: 'Ranking final' });
    const items = await within(ranking).findAllByRole('listitem');
    expect(items).toHaveLength(5);
    expect(screen.queryByText(/El puntaje combina/)).not.toBeInTheDocument();
    expect(onLoaded).toHaveBeenCalledWith(results);
  });

  it('muestra aparte, sin posición, las propuestas sin evaluaciones', async () => {
    const none: AttachmentResult = {
      ...r('a-9', 'u-1', 'Ana Pérez', 'nube.pdf', 0, 0),
      vote_count: 0,
      average_rank: 0,
      unevaluated: true,
    };
    getResults.mockResolvedValue({
      ...results,
      adjusted_ranking: [...results.adjusted_ranking.filter((x) => x.attachment_id !== 'a-1'), none],
    });
    renderWithProviders(<EventResults eventId="e-1" currentUserId="u-1" />);
    expect(await screen.findByText('Sin evaluaciones')).toBeInTheDocument();
    expect(
      screen.getByText('Esta propuesta no recibió ninguna evaluación, así que no tiene posición en el ranking.')
    ).toBeInTheDocument();
    expect(screen.getByText('Ana Pérez · nube.pdf (Tú)')).toBeInTheDocument();
  });

  it('muestra aparte las propuestas excluidas porque su autor no evaluó', async () => {
    const out = {
      ...results.adjusted_ranking.find((x) => x.attachment_id === 'a-1')!,
      adjusted_rank: 0,
      excluded: true,
      non_evaluator: true,
    };
    getResults.mockResolvedValue({
      ...results,
      adjusted_ranking: [...results.adjusted_ranking.filter((x) => x.attachment_id !== 'a-1'), out],
    });
    renderWithProviders(<EventResults eventId="e-1" currentUserId="u-1" />);
    expect(await screen.findByText('Fuera del ranking')).toBeInTheDocument();
    expect(
      screen.getByText('Quien presentó esta propuesta no completó su evaluación, así que quedó fuera del ranking.')
    ).toBeInTheDocument();
    expect(screen.getByText(new RegExp(`· ${out.filename} \\(Tú\\)$`))).toBeInTheDocument();
  });

  it('explica que el orden final incluye el ajuste por cómo evaluó cada autor', async () => {
    getResults.mockResolvedValue(results);
    renderWithProviders(<EventResults eventId="e-1" currentUserId={null} />);
    expect(await screen.findByText(/El orden final incluye un ajuste por cómo evaluó cada autor/)).toBeInTheDocument();
    expect(screen.queryByText(/Premio de|Castigo de/)).not.toBeInTheDocument();
  });

  it('muestra al participante el ajuste de su propia propuesta', async () => {
    getResults.mockResolvedValue({
      ...results,
      adjusted_ranking: results.adjusted_ranking.map((x) =>
        x.attachment_id === 'a-1' ? { ...x, adjustment: 3, adjustment_reason: 'good_evaluator' as const } : x
      ),
    });
    renderWithProviders(<EventResults eventId="e-1" currentUserId="u-1" />);
    expect(await screen.findByText('Ajuste de tu propuesta')).toBeInTheDocument();
    expect(
      screen.getByText(/^Tu propuesta recibió un premio de 3 puestos porque tu ranking coincidió con el de la comunidad\. El ajuste se mide/)
    ).toBeInTheDocument();
    expect(screen.getByText('Premio de 3 puestos')).toBeInTheDocument();
  });

  it('avisa cuando la propuesta propia bajó por no completar la evaluación', async () => {
    getResults.mockResolvedValue({
      ...results,
      adjusted_ranking: results.adjusted_ranking.map((x) =>
        x.attachment_id === 'a-1' ? { ...x, adjustment: -10, adjustment_reason: 'non_evaluator' as const } : x
      ),
    });
    renderWithProviders(<EventResults eventId="e-1" currentUserId="u-1" />);
    expect(
      await screen.findByText(/^Tu propuesta recibió un castigo de 10 puestos porque no completaste tu evaluación\./)
    ).toBeInTheDocument();
  });

  it('al organizador le muestra el ajuste de cada propuesta que lo tuvo', async () => {
    getResults.mockResolvedValue({
      ...results,
      adjusted_ranking: results.adjusted_ranking.map((x) =>
        x.attachment_id === 'a-2'
          ? { ...x, adjustment: 3, adjustment_reason: 'good_evaluator' as const }
          : x.attachment_id === 'a-5'
            ? { ...x, adjustment: -3, adjustment_reason: 'bad_evaluator' as const }
            : x
      ),
    });
    renderWithProviders(<EventResults eventId="e-1" currentUserId={null} />);
    expect(await screen.findByText('Premio de 3 puestos')).toBeInTheDocument();
    expect(screen.getByText('Castigo de 3 puestos')).toBeInTheDocument();
    expect(screen.getByText(/su evaluación se alejó de la comunidad/)).toBeInTheDocument();
    expect(screen.queryByText('Ajuste de tu propuesta')).not.toBeInTheDocument();
  });

  it('TS-47: sin resultados no es un error y nunca recalcula', async () => {
    getResults.mockRejectedValueOnce(
      new ApiError({ status: 404, body: { error: 'RESULTS_NOT_CALCULATED', message: 'x' } })
    );
    const { unmount } = renderWithProviders(<EventResults eventId="e-1" currentUserId={null} />);
    expect(await screen.findByText('Los resultados todavía no están disponibles.')).toBeInTheDocument();
    expect(screen.queryByRole('list')).toBeNull();
    expect(screen.queryByRole('alert')).toBeNull();
    unmount();

    getResults.mockResolvedValueOnce({ ...results, adjusted_ranking: [] });
    renderWithProviders(<EventResults eventId="e-1" currentUserId={null} />);
    expect(await screen.findByText('Los resultados todavía no están disponibles.')).toBeInTheDocument();
    expect(recalculate).not.toHaveBeenCalled();
  });

  it('TS-48: error con reintento y sin "pesa más"', async () => {
    getResults.mockRejectedValueOnce(new TypeError('Failed to fetch')).mockResolvedValueOnce(results);
    renderWithProviders(<EventResults eventId="e-1" currentUserId={null} />);
    expect(await screen.findByRole('alert')).toHaveTextContent('No pudimos cargar los resultados.');
    userEvent.click(screen.getByRole('button', { name: 'Reintentar' }));
    expect(await screen.findByText('Ranking final')).toBeInTheDocument();
    expect(document.body.textContent).not.toMatch(/pesa|weight/i);
  });

  it('solo ranking cuando hay 3 o menos propuestas', async () => {
    getResults.mockResolvedValue({ ...results, adjusted_ranking: results.adjusted_ranking.slice(0, 3) });
    renderWithProviders(<EventResults eventId="e-1" currentUserId={null} />);
    expect(await screen.findByText('Ranking final')).toBeInTheDocument();
    const items = await screen.findAllByRole('listitem');
    expect(items).toHaveLength(3);
  });

  it('TS-29: con recalculateIfMissing recalcula una vez y muestra el ranking', async () => {
    getResults.mockRejectedValue(
      new ApiError({ status: 404, body: { error: 'RESULTS_NOT_CALCULATED', message: 'x' } })
    );
    recalculate.mockResolvedValue({ ...results, adjusted_ranking: results.adjusted_ranking.slice(0, 3) });
    renderWithProviders(<EventResults eventId="e-1" currentUserId={null} recalculateIfMissing />);
    expect(await screen.findByText('Ranking final')).toBeInTheDocument();
    expect(recalculate).toHaveBeenCalledTimes(1);
    expect(recalculate).toHaveBeenCalledWith('e-1');
    expect(screen.queryByText('Los resultados todavía no están disponibles.')).toBeNull();
  });

  it('TS-29b: sin la prop no recalcula', async () => {
    getResults.mockRejectedValue(
      new ApiError({ status: 404, body: { error: 'RESULTS_NOT_CALCULATED', message: 'x' } })
    );
    renderWithProviders(<EventResults eventId="e-1" currentUserId={null} />);
    expect(await screen.findByText('Los resultados todavía no están disponibles.')).toBeInTheDocument();
    expect(recalculate).not.toHaveBeenCalled();
  });

  it('TS-29c: si el recálculo falla muestra error con Reintentar', async () => {
    getResults.mockRejectedValue(
      new ApiError({ status: 404, body: { error: 'RESULTS_NOT_CALCULATED', message: 'x' } })
    );
    recalculate.mockRejectedValueOnce(new TypeError('Failed to fetch'));
    renderWithProviders(<EventResults eventId="e-1" currentUserId={null} recalculateIfMissing />);
    expect(await screen.findByRole('alert')).toHaveTextContent('No pudimos cargar los resultados.');
    expect(recalculate).toHaveBeenCalledTimes(1);
    expect(screen.getByRole('button', { name: 'Reintentar' })).toBeInTheDocument();
  });
});
