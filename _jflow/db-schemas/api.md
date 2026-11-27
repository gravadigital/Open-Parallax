# Database — api

Described from the schema the migrations build on an empty database (`services/api/internal/storage/migrations/`,
applied by the api when it starts or by `make migrate`). It does not follow every convention of the Database spec:
the deviations are listed in the adoption report.

## Base

| Item | Value |
|---|---|
| Owner service | api |
| Engine | PostgreSQL 15 |
| Other services that use it | ninguno |
| Extensions | `uuid-ossp` (identificadores `uuid_generate_v4()`) |
| Migrations | Go, numeradas y registradas en orden en `migrations.go`, cada una en su transacción, registradas en `schema_migrations`. La 002 crea las tablas con `AutoMigrate` de GORM sobre los modelos de `migrations/models.go`. Una migración que borra registros con archivo (la 029) devuelve sus claves y el archivo se borra en el almacenamiento después de confirmarla; el que no se puede borrar queda en el log y la migración no falla (DEC-10) |

## Schema decisions

| Decision | Value |
|---|---|
| Concurrency | No aplica: sin columna de versión. El ranking bloquea la fila de la asignación (GAR-08) |
| Multi-tenancy | No aplica: una sola instalación |
| History | No se guarda historia: los votos se reemplazan al reenviar el ranking y los resultados al recalcular |
| Deletion | Borrado físico; no hay `deleted_at`. Las notificaciones se borran a los 90 días |
| Currency | No aplica |
| Re-delivered tokens | No aplica: el token de recuperación se entrega una sola vez y se guarda tal cual (sin hash) |

## Schema

```dbml
Enum event_stage {
  creation [note: 'Creación']
  participation [note: 'Participación']
  voting [note: 'Votación']
  results [note: 'Resultados']
}

Enum user_role {
  admin [note: 'admin']
  participant [note: 'participant']
  organizer [note: 'organizer']
}

Enum event_participant_role {
  creator [note: 'creador del evento']
  participant [note: 'inscrito']
}

Enum notification_type {
  stage_changed [note: 'NOT-01']
  event_cancelled [note: 'NOT-02']
  event_paused [note: 'NOT-03']
  deadline_changed [note: 'NOT-04']
  participant_registered [note: 'NOT-05']
  registration_confirmed [note: 'NOT-06']
  ranking_submitted [note: 'NOT-07']
  file_reminder [note: 'NOT-08']
  vote_reminder [note: 'NOT-09']
}

// participant
Table users {
  id uuid [pk, default: `uuid_generate_v4()`]
  name text [not null, note: 'nombre · dato personal']
  lastname text [note: 'apellido · dato personal']
  email text [not null, unique, note: 'email · dato personal (índice único idx_users_email, distingue mayúsculas)']
  password_hash text [note: 'contraseña (bcrypt)']
  role user_role [note: 'rol global']
  google_id varchar [unique, note: 'identidad de Google · dato personal']
  password_reset_token varchar(64) [unique, note: 'token de recuperación']
  password_reset_expires_at timestamptz [note: 'vencimiento del token de recuperación']
  created_at timestamptz
  updated_at timestamptz
  Note: 'Entidad: Usuario'
}

// event
Table events {
  id uuid [pk, default: `uuid_generate_v4()`]
  name text [not null, note: 'nombre (único solo en Go: GAR-13)']
  description text [not null, note: 'descripción']
  author_id uuid [not null, ref: > users.id, note: 'autor']
  start_date timestamptz [not null, note: 'fecha de inicio']
  end_date timestamptz [not null, note: 'fecha de fin']
  stage event_stage [not null, default: 'creation', note: 'etapa']
  is_cancelled boolean [default: false, note: 'cancelado']
  is_paused boolean [default: false, note: 'pausado']
  organizer varchar(200) [default: '', note: 'organizador']
  shareable_link varchar(255) [not null, unique, note: 'enlace para compartir']
  max_participants integer [not null, default: 20, note: 'cupo']
  participation_estimated_end_date date [note: 'cierre estimado de Participación']
  voting_estimated_end_date date [note: 'cierre estimado de Votación']
  cover_color varchar(32) [not null, default: 'lavender', note: 'color de portada']
  created_at timestamptz
  updated_at timestamptz
  Note: 'Entidad: Evento'
}

Table event_participants {
  event_id uuid [pk, ref: > events.id]
  user_id uuid [pk, ref: > users.id]
  role event_participant_role [not null, default: 'participant', note: 'rol en el evento']
  joined_at timestamptz [note: 'fecha de inscripción']
  Note: 'Entidad: Inscripción'
}

// attachment
Table attachments {
  id uuid [pk, default: `uuid_generate_v4()`]
  event_id uuid [not null, ref: > events.id]
  participant_id uuid [not null, ref: > users.id, note: 'participante']
  filename text [not null, note: 'nombre guardado']
  original_name text [not null, note: 'nombre original']
  file_path text [not null, note: 'ubicación']
  file_size bigint [not null, note: 'tamaño']
  mime_type varchar(100) [not null, note: 'tipo']
  description text [note: 'descripción']
  vote_count bigint [default: 0, note: 'votos recibidos (GAR-14)']
  uploaded_at timestamptz [note: 'fecha de subida']
  indexes {
    (event_id, participant_id) [unique, name: 'uq_attachments_event_participant', note: 'INV-13 (GAR-16)']
  }
  Note: 'Entidad: Propuesta'
}

// vote
Table voting_configurations {
  id uuid [pk, default: `uuid_generate_v4()`]
  event_id uuid [not null, unique, ref: > events.id]
  attachments_per_evaluator bigint [not null, note: 'm']
  min_evaluations_per_file bigint [not null, default: 3, note: 'evaluaciones mínimas por propuesta']
  quality_good_threshold numeric(3,2) [not null, default: 0.65, note: 'umbral de buen evaluador (la api escribe 0,6 por defecto)']
  quality_bad_threshold numeric(3,2) [not null, default: 0.35, note: 'umbral de mal evaluador (la api escribe 0,3 por defecto)']
  adjustment_magnitude bigint [not null, default: 3, note: 'n']
  non_evaluator_sanction varchar(20) [not null, default: 'exclusion', note: 'sanción a quien no evalúa']
  non_evaluator_penalty integer [not null, default: 10, note: 'penalización']
  use_expertise_matching boolean [default: false, note: 'Técnica: sin uso']
  enable_co_idetection boolean [default: true, note: 'Técnica: sin uso']
  randomization_seed bigint [note: 'Técnica: sin uso']
  assignment_algorithm varchar(50) [default: 'random_balanced', note: 'Técnica: sin uso']
  scoring_algorithm varchar(50) [default: 'modified_borda_count', note: 'Técnica: sin uso']
  created_at timestamptz
  updated_at timestamptz
  Note: 'Entidad: Configuración de votación'
}

Table assignments {
  id uuid [pk, default: `uuid_generate_v4()`]
  event_id uuid [not null, ref: > events.id]
  participant_id uuid [not null, ref: > users.id, note: 'evaluador']
  attachment_ids "uuid[]" [note: 'propuestas']
  assignment_round bigint [default: 1, note: 'Técnica: siempre 1']
  is_completed boolean [default: false, note: 'completa (GAR-14)']
  completed_at timestamptz [note: 'fecha de completitud']
  quality_score numeric(5,4) [note: 'Técnica: nunca se escribe; Q vive en voting_results']
  expertise_match_score numeric(3,2) [note: 'Técnica: sin uso']
  conflict_of_interest boolean [default: false, note: 'Técnica: sin uso']
  created_at timestamptz
  updated_at timestamptz
  indexes {
    (event_id, participant_id) [unique, name: 'unique_participant_event_assignment']
  }
  Note: 'Entidad: Asignación'
}

Table votes {
  id uuid [pk, default: `uuid_generate_v4()`]
  event_id uuid [not null, ref: > events.id]
  assignment_id uuid [not null, ref: > assignments.id]
  voter_id uuid [not null, ref: > users.id, note: 'evaluador']
  attachment_id uuid [not null, ref: > attachments.id, note: 'propuesta']
  rank_position bigint [not null, note: 'posición']
  score numeric(10,4) [note: 'puntaje (lo calcula el trigger)']
  confidence numeric(3,2) [note: 'Técnica: sin uso']
  evaluation_time_seconds bigint [note: 'Técnica: sin uso']
  notes text [note: 'Técnica: sin uso']
  is_quality_vote boolean [note: 'Técnica: sin uso']
  voted_at timestamptz [note: 'fecha del voto']
  indexes {
    (assignment_id, attachment_id) [unique, name: 'uq_votes_assignment_attachment']
  }
  Note: 'Entidad: Voto'
}

Table vote_drafts {
  id uuid [pk, default: `uuid_generate_v4()`]
  event_id uuid [not null, ref: > events.id]
  assignment_id uuid [not null, ref: > assignments.id]
  participant_id uuid [not null, ref: > users.id, note: 'evaluador']
  rankings jsonb [not null, default: '[]', note: 'ranking']
  created_at timestamptz [not null, default: `now()`]
  updated_at timestamptz [not null, default: `now()`]
  indexes {
    (assignment_id, participant_id) [unique, name: 'uq_vote_draft_assignment_participant']
  }
  Note: 'Entidad: Borrador de voto'
}

Table voting_results {
  id uuid [pk, default: `uuid_generate_v4()`]
  event_id uuid [not null, unique, ref: > events.id]
  global_ranking jsonb [not null, note: 'ranking global (G)']
  participant_qualities jsonb [not null, note: 'calidad de los evaluadores (Q)']
  adjusted_ranking jsonb [not null, note: 'ranking ajustado (G′)']
  statistical_metrics jsonb [note: 'Técnica: sin uso']
  total_participants bigint [not null, note: 'participantes']
  total_votes bigint [not null, note: 'votos']
  attachments_per_evaluator bigint [not null, note: 'm']
  algorithm_used varchar(100) [not null, default: 'modified_borda_count']
  quality_adjustments_applied boolean [default: true, note: 'Técnica: sin uso']
  overall_quality_score numeric(5,4) [note: 'Técnica: sin uso']
  good_evaluator_count bigint [default: 0, note: 'Técnica: sin uso']
  bad_evaluator_count bigint [default: 0, note: 'Técnica: sin uso']
  consensus_strength numeric(3,2) [note: 'Técnica: sin uso']
  calculated_at timestamptz [note: 'fecha de cálculo']
  updated_at timestamptz
  Note: 'Entidad: Resultados'
}

// notification
Table notifications {
  id uuid [pk, default: `uuid_generate_v4()`]
  recipient_id uuid [not null, ref: > users.id, note: 'destinatario']
  event_id uuid [not null, ref: > events.id]
  type notification_type [not null, note: 'tipo']
  data jsonb [not null, default: '{}', note: 'datos']
  read_at timestamptz [note: 'leída']
  created_at timestamptz [not null, default: `now()`, note: 'fecha']
  updated_at timestamptz [not null, default: `now()`]
  indexes {
    (recipient_id, created_at) [name: 'idx_notifications_recipient_created']
  }
  Note: 'Entidad: Notificación'
}

// infrastructure
Table schema_migrations {
  id varchar(10) [pk, note: 'Técnica: número de la migración']
  name varchar(255) [not null, note: 'Técnica: nombre de la migración']
  applied_at timestamptz [default: `CURRENT_TIMESTAMP`, note: 'Técnica: cuándo se aplicó; `cmd/migrate -rollback` deshace la última por este orden']
  Note: 'Técnica: migraciones aplicadas'
}
```

Foreign keys have the names GORM generated (`fk_users_authored_events`, `fk_events_votes`…) with no `ON DELETE`, except
those of `notifications` and `vote_drafts`, which are `ON DELETE CASCADE`. On a fresh database `events.is_cancelled`
and `events.is_paused` are nullable (the `NOT NULL` of migrations 018 and 020 only applies to databases that already
had the columns).

## Constraints DBML does not express

| Name | Table | Type | Definition | Guarantee |
|---|---|---|---|---|
| `valid_email` | users | check | `email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'` | — |
| `valid_event_dates` | events | check | `end_date >= start_date` | GAR-09 |
| `valid_cover_color` | events | check | `cover_color` en la paleta de 16 | — |
| `enforce_future_start_date` | events | trigger | `BEFORE INSERT OR UPDATE OF start_date`: rechaza un inicio más de un día en el pasado («start_date cannot be in the past») | GAR-09 |
| `valid_file_size` | attachments | check | `file_size > 0 AND file_size <= 104857600` | — |
| `valid_filename` | attachments | check | `length(filename) > 0` | — |
| `valid_m_parameter` | voting_configurations | check | `attachments_per_evaluator` entre 1 y 50 | — |
| `valid_min_evaluations` | voting_configurations | check | `min_evaluations_per_file > 0` | — |
| `valid_quality_thresholds` | voting_configurations | check | `good > bad AND good - bad >= 0.1` | — |
| `valid_adjustment_magnitude` | voting_configurations | check | entre 0 y 20 | — |
| `valid_non_evaluator_sanction` | voting_configurations | check | `exclusion` o `penalty` | — |
| `valid_non_evaluator_penalty` | voting_configurations | check | entre 1 y 100 | — |
| `valid_quality_score` | assignments | check | nulo o entre 0 y 1 | — |
| `trigger_validate_assignment` | assignments | trigger | `validate_assignment_constraints()`: con configuración, exactamente m propuestas, todas del evento, ninguna del evaluador («Participant cannot be assigned to evaluate their own proposals (conflict of interest)») | GAR-01, GAR-02 |
| `valid_rank_position` | votes | check | `rank_position > 0` | GAR-03 |
| `valid_confidence`, `valid_evaluation_time` | votes | check | rangos de columnas sin uso | — |
| `valid_participant_counts` | voting_results | check | `total_participants > 0 AND total_votes >= 0` | — |
| `valid_evaluator_counts` | voting_results | check | `good_evaluator_count >= 0 AND bad_evaluator_count >= 0 AND good_evaluator_count + bad_evaluator_count <= total_participants` | — |
| `trigger_validate_vote` | votes | trigger | `validate_vote_constraints()`: el evaluador tiene asignación, la propuesta está en ella, posición ≤ m; calcula `score = (m − posición + 1) · 100 / m` | GAR-03 |
| `trigger_update_assignment_completion_insert`, `_delete` | votes | trigger | `update_assignment_completion()`: marca o desmarca `is_completed` y `completed_at` | GAR-14 |
| `trigger_update_vote_count` | votes | trigger | `update_attachment_vote_count()`: mantiene `attachments.vote_count` | GAR-14 |
| `uq_notifications_registered_unread` | notifications | partial unique | `(recipient_id, event_id, type) WHERE read_at IS NULL AND type = 'participant_registered'` | — |
| `idx_notifications_unread` | notifications | partial index | `(recipient_id) WHERE read_at IS NULL` | — |
| `idx_users_google_id` | users | partial index | `(google_id) WHERE google_id IS NOT NULL` | — |
| `check_coverage_constraints(uuid)` | — | function | usada solo por la vista `system_validation` | — |
| `assignment_statistics`, `voting_progress`, `system_validation`, `quality_assessment` | — | views | vistas de diagnóstico; el código no las usa | — |

## Derived attributes

| Entity | Attribute | Stored | By |
|---|---|---|---|
| Propuesta | votos recibidos | sí (`attachments.vote_count`) | trigger `update_attachment_vote_count` |
| Asignación | completa, fecha de completitud | sí (`assignments.is_completed`, `completed_at`) | trigger `update_assignment_completion` |
| Voto | puntaje | sí (`votes.score`) | trigger `validate_vote_constraints` |
