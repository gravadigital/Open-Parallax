import React from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import Card from '../../components/ui/card/Card';
import Callout from '../../components/ui/callout/Callout';
import EmptyState from '../../components/ui/empty-state/EmptyState';
import StatusPill from '../../components/ui/status-pill/StatusPill';
import { ChevronDownIcon } from '../../components/ui/icons/Icons';
import StageTimeline from '../../components/events/stage-timeline/StageTimeline';
import { useAuth } from '../../context/AuthContext';
import { useT } from '../../i18n';
import { daysUntilClose } from '../../domain/dates';
import { openEvents } from '../../domain/events';
import { loginPathFor } from '../../domain/redirect';
import { STAGE_ORDER, stageIndex, stageNameKey, stageStatus } from '../../domain/stages';
import { EventService } from '../../services/api';
import type { EventListItem, EventStage } from '../../types';
import '../../components/ui/visually-hidden.css';
import './HomePage.css';

const FEATURED_LIMIT = 3;
const HOW_ANCHOR = 'como-funciona';
const STEP_KEYS = ['creation', 'participation', 'voting', 'results'] as const;

type LoadState = { status: 'loading' | 'ready' | 'error'; items: EventListItem[] };

const HomePage: React.FC = () => {
  const { t } = useT();
  const { user } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [state, setState] = React.useState<LoadState>({ status: 'loading', items: [] });
  const [stagesOpen, setStagesOpen] = React.useState(false);
  const howTitleRef = React.useRef<HTMLHeadingElement>(null);
  const howSectionRef = React.useRef<HTMLElement>(null);
  const howTitleId = React.useId();
  const requestId = React.useRef(0);

  const load = React.useCallback(() => {
    const id = ++requestId.current;
    setState({ status: 'loading', items: [] });
    EventService.listEvents({ stage: 'participation', limit: 100 })
      .then((result) => {
        if (id === requestId.current) setState({ status: 'ready', items: result.items });
      })
      .catch(() => {
        if (id === requestId.current) setState({ status: 'error', items: [] });
      });
  }, []);

  React.useEffect(() => {
    load();
    return () => {
      requestId.current += 1;
    };
  }, [load]);

  React.useEffect(() => {
    if (location.hash === `#${HOW_ANCHOR}`) {
      howSectionRef.current?.scrollIntoView?.();
      howTitleRef.current?.focus();
    }
  }, [location.hash]);

  const open = React.useMemo(() => openEvents(state.items), [state.items]);
  const featured = open[0];
  const list = open.slice(0, FEATURED_LIMIT);
  const createPath = user ? '/events/create' : loginPathFor('/events/create');

  const stageLabels = STAGE_ORDER.reduce(
    (acc, stage) => ({ ...acc, [stage]: t(stageNameKey(stage)) }),
    {} as Record<EventStage, string>
  );

  const closesText = (event: EventListItem): string | null => {
    const date = event.participation_estimated_end_date;
    if (!date) return null;
    const days = daysUntilClose(date);
    if (days < 0) return null;
    return days === 0 ? t('home.featured.closesToday') : t('home.featured.closesIn', { count: days });
  };

  const loading = state.status === 'loading';
  const featuredCloses = featured ? closesText(featured) : null;

  return (
    <div className="hm-page">
      <section className="hm-hero">
        <div className="hm-hero__copy">
          <p className="hm-eyebrow">{t('home.eyebrow')}</p>
          <h1 className="hm-hero__title">{t('home.title')}</h1>
          <p className="hm-hero__subtitle">{t('home.subtitle')}</p>
          <div className="hm-hero__actions">
            <Link
              to="/events?stage=participation"
              className="ui-button ui-button--lg ui-button--primary"
            >
              {t('home.explore')}
            </Link>
            <Link to={createPath} className="ui-button ui-button--lg ui-button--secondary">
              {t('home.create')}
            </Link>
          </div>
        </div>

        {loading && (
          <div className="hm-hero__featured" aria-hidden="true">
            <Card variant="raised">
              <span className="hm-skeleton hm-skeleton--pill" />
              <span className="hm-skeleton hm-skeleton--title" />
              <span className="hm-skeleton" />
            </Card>
          </div>
        )}

        {state.status === 'ready' && featured && (
        <section className="hm-hero__featured" aria-label={t('home.featured.label')}>
          <article className={`hm-featured-card${stagesOpen ? ' hm-featured-card--stages-open' : ''}`}>
            <div className="hm-featured-card__top">
              <StatusPill tone="success" icon="dot" size="sm">
                {t('home.featured.status')}
              </StatusPill>
              {featuredCloses && <span className="hm-featured-card__closes">{featuredCloses}</span>}
            </div>
            <h3 className="hm-featured__name">{featured.name}</h3>
            {featured.description && <p className="hm-featured__description">{featured.description}</p>}
            <div className="hm-featured-card__stepper">
              <div className="hm-featured-card__stage-head">
                <span className="hm-featured-card__stage-of">
                  {t('events.timeline.stepOf', { number: 2 })}
                </span>
                <button
                  type="button"
                  className="hm-featured-card__stage-toggle"
                  aria-expanded={stagesOpen}
                  onClick={() => setStagesOpen((open) => !open)}
                >
                  {stagesOpen ? t('home.featured.hideStages') : t('home.featured.viewStages')}
                  <ChevronDownIcon
                    className={`hm-featured-card__stage-chevron${
                      stagesOpen ? ' hm-featured-card__stage-chevron--open' : ''
                    }`}
                  />
                </button>
              </div>
              <StageTimeline
                variant="compact"
                current="participation"
                stageLabels={stageLabels}
                nowLabel={t('events.timeline.now')}
                completedLabel={t('events.timeline.completed')}
                pendingLabel={t('events.timeline.pending')}
                stepOfLabel={t('events.timeline.stepOf', { number: 2 })}
              />
              <p className="hm-featured-card__after">
                {t('home.featured.after')}: {stageLabels.voting}
              </p>
              {stagesOpen && (
                <ol className="hm-featured-card__stages">
                  {STAGE_ORDER.map((stage, index) => {
                    const status = stageStatus(stage, 'participation');
                    const stateKey =
                      status === 'completed'
                        ? 'stageDone'
                        : status === 'current'
                          ? 'stageNow'
                          : 'stageNext';
                    return (
                      <li
                        key={stage}
                        className={`hm-featured-card__stage-row hm-featured-card__stage-row--${status}${
                          index === STAGE_ORDER.length - 1
                            ? ' hm-featured-card__stage-row--last'
                            : ''
                        }`}
                      >
                        <span className="hm-featured-card__stage-rail" aria-hidden="true">
                          <span className="hm-featured-card__stage-dot">
                            {status === 'completed' ? '✓' : stageIndex(stage) + 1}
                          </span>
                          <span className="hm-featured-card__stage-line" />
                        </span>
                        <span className="hm-featured-card__stage-body">
                          <span className="hm-featured-card__stage-name-row">
                            <span className="hm-featured-card__stage-name">
                              {stageLabels[stage]}
                            </span>
                            <span className="hm-featured-card__stage-state">
                              {t(`home.featured.${stateKey}` as const)}
                            </span>
                          </span>
                        </span>
                      </li>
                    );
                  })}
                </ol>
              )}
            </div>
            <div className="hm-featured-card__footer">
              <span className="hm-featured-card__capacity">
                {featured.max_participants === null ? (
                  t('events.table.registered', { count: featured.participants_count })
                ) : (
                  <>
                    <b>{featured.participants_count}</b>{' '}
                    {t('home.featured.capacity', { max: featured.max_participants })}
                  </>
                )}
              </span>
              <Link
                to={`/events/${featured.id}`}
                className="ui-button ui-button--md ui-button--primary hm-featured-card__cta"
              >
                {t('home.featured.participate')}
              </Link>
            </div>
          </article>
        </section>
        )}
      </section>

      <div className="hm-divider" aria-hidden="true" />

      <section id={HOW_ANCHOR} className="hm-how" aria-labelledby={howTitleId} ref={howSectionRef}>
        <div className="hm-how__header">
          <div>
            <p className="hm-eyebrow">{t('home.how.eyebrow')}</p>
            <h2 id={howTitleId} className="hm-section-title" tabIndex={-1} ref={howTitleRef}>
              {t('home.how.title')}
            </h2>
          </div>
        </div>
        <ol className="hm-steps">
          {STEP_KEYS.map((key, index) => (
            <li key={key} className="hm-steps__item">
              <span className="hm-step__number">{String(index + 1).padStart(2, '0')}</span>
              <span className="hm-step__name">{t(stageNameKey(key))}</span>
              <span className="hm-step__text">{t(`home.how.${key}` as const)}</span>
            </li>
          ))}
        </ol>
      </section>

      <section className="hm-open" aria-busy={loading || undefined}>
        <div className="hm-open__header">
          <h2 className="hm-section-title hm-open__title">{t('home.open.title')}</h2>
          <Link to="/events" className="hm-open__all">
            {t('home.open.seeAll')}
          </Link>
        </div>
        <div className="hm-open__body">
          {loading && (
            <>
              <p role="status" className="ui-visually-hidden">
                {t('home.open.loading')}
              </p>
              <div className="hm-open__list" aria-hidden="true">
                {[0, 1, 2].map((n) => (
                  <Card key={n} variant="default" padding="compact">
                    <span className="hm-skeleton hm-skeleton--pill" />
                    <span className="hm-skeleton hm-skeleton--title" />
                    <span className="hm-skeleton" />
                  </Card>
                ))}
              </div>
            </>
          )}
          {state.status === 'error' && (
            <Callout tone="error" action={{ label: t('common.retry'), onClick: load }}>
              {t('home.open.error')}
            </Callout>
          )}
          {state.status === 'ready' && list.length === 0 && (
            <EmptyState
              variant="inline"
              headingLevel={3}
              title={t('home.open.empty')}
              action={{ label: t('home.open.emptyAction'), onClick: () => navigate('/events') }}
            />
          )}
          {state.status === 'ready' && list.length > 0 && (
            <ul className="hm-open__list">
              {list.map((event) => (
                <li key={event.id} className="hm-open__item">
                  <Card as="article" variant="subtle" padding="default" className="hm-open__card">
                    <div className="hm-open__card-meta">
                      <StatusPill tone="success" icon="dot" size="sm">
                        {t('home.featured.status')}
                      </StatusPill>
                      <span className="hm-open__card-capacity">
                        {event.max_participants === null
                          ? t('events.table.registered', { count: event.participants_count })
                          : t('home.open.spots', {
                              count: event.participants_count,
                              max: event.max_participants,
                            })}
                      </span>
                    </div>
                    <div className="hm-open__card-details">
                      <h3 className="hm-card__name">{event.name}</h3>
                      {event.description && <p className="hm-card__description">{event.description}</p>}
                    </div>
                    <Link
                      to={`/events/${event.id}`}
                      className="ui-button ui-button--md ui-button--secondary hm-open__card-cta"
                    >
                      {t('home.open.cta')}
                    </Link>
                  </Card>
                </li>
              ))}
            </ul>
          )}
        </div>
      </section>

      <Card variant="feature" className="hm-banner">
        <div className="hm-banner__copy">
          <p className="hm-banner__title">{t('home.banner.title')}</p>
          <p className="hm-banner__text">{t('home.banner.text')}</p>
        </div>
        <Link to={createPath} className="ui-button ui-button--md ui-button--on-band hm-banner__cta">
          {t('home.banner.cta')}
        </Link>
      </Card>
    </div>
  );
};

export default HomePage;
