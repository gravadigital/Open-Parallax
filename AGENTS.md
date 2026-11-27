# OpenParallax

Distributed peer review for calls with more proposals than evaluators: participants rank each
other's proposals and a Modified Borda Count (Merrifield & Saari, 2009) builds the official ranking.
Adopted into jiku flow in #622 (see `docs/adopcion.md`).

This repository uses **jiku flow**: every change belongs to a Jiku requirement, whose
specification (scope, proposal, acceptance criteria) lives in Jiku; the product definition — what is
built, its rules and why — lives in `_jflow/`. Read this guide before touching code or docs.

## Running

Everything runs with Docker Compose and is driven by `make`. `make up` needs only Docker; the checks
(`make check`) also need Go (version in `services/api/go.mod`) and Node.js 22 (`services/web/.nvmrc`),
as CI has them. The end-to-end tests run Playwright in its Docker image.

| Command | What it does |
|---|---|
| `make up` / `make down` | Start / stop the whole product |
| `make logs` | Follow the logs |
| `make dev` | Database, storage and Mailpit in Docker; api (`go run`) and web (`npm start`, hot reload) by hand |
| `make test` | Every test; JUnit reports in `reports/junit/` |
| `make test-unit` · `make test-integration` · `make test-e2e` | One test level |
| `make lint` · `make typecheck` · `make fmt` | Static checks and formatting |
| `make check` | `lint` + `typecheck` + `test`: what must pass before closing any work. CI runs its fast part (lint, typecheck, unit, integration) |
| `make migrate` | Apply the migrations (the api also applies them when it starts) |
| `make test-e2e` | Brings the product up with `docker-compose.test.yml` (fake Google, seeded admin and organizer) and runs the end-to-end tests |

`S={service}` limits any command to one service: `make test S=api`.

## Services

| Service | Folder | What it is |
|---|---|---|
| api | `services/api/` | Go HTTP service: the whole domain (events, accounts, proposals, voting, notifications); the only one with state |
| web | `services/web/` | React SPA (Create React App), the only surface; no business logic of its own |

Each folder has its own AGENTS.md with the stack, the commands and the canonical examples.

## Where things are

| What | Where |
|---|---|
| Product definition (what is built, rules, decisions) | `_jflow/prd/` — start with `_jflow/prd/index.md` |
| Contracts of each service | `_jflow/apis/` |
| Database schemas | `_jflow/db-schemas/` |
| Flows across services | `_jflow/flows/` |
| Specification of a requirement | In Jiku: `python3 .claude/scripts/jflow.py spec show {id}` |
| Documentation for people | `README.md`, `CHANGELOG.md`, `docs/`: the operations guide and, with external integrators, the integration guide (`operacion.md`, `integracion.md` when the product language is Spanish; `operations.md`, `integration.md` otherwise) |
| Code | `services/{service}/` |
| Test reports | `reports/` (not versioned) |

To load only the part of the definition a task touches:
`python3 .claude/scripts/jflow.py load --block {PREFIX}` (or `--id`, `--entity`, `--module`).

## Answering questions about the product

When asked how the product behaves or why, load the relevant part of the definition
(`jflow.py load …`) and answer citing its IDs (`CAN-E1`, `DEC-04`). When given a new idea, say which
rules and decisions it touches or contradicts, what it would change, and whether it would be a new
requirement (`/jflow-spec`) or a small change (`/jflow-fix`).

## How work is done

- **Every change belongs to a Jiku requirement.** A feature or improvement is specified with
  `/jflow-spec {id}` and implemented with `/jflow-build {id}`. A bug or a small change goes through
  `/jflow-fix {id}`. `/jflow-status` shows what is pending; `/jflow-release` releases a version.
- **Before starting**, `jflow.py git-check`: work starts from an up-to-date, clean `dev`.
- **Branches:** `feat/{id}-{slug}` or `fix/{id}-{slug}`, from `dev`. The MR/PR targets `dev`; `main`
  is production.
- **Tests carry the ID they prove** in brackets: `[CAN-E1] does not cancel an appointment that already started`. What to
  test comes from the definition: requirement, failure condition, invariant, guarantee.
- **The definition is updated in the same branch** as the code that changes it. On `dev`, what is
  documented is what is built.
- **Language** (`.claude/specs/language.md`): talk to the user in their language. What people of the
  product read — `docs/`, `README.md`, `CHANGELOG.md`, everything in Jiku, the texts and error
  messages the product shows its users — is in the product language (`language` in
  `_jflow/project.yaml`, Spanish when unset). `_jflow/` has its structure (headings, columns, fixed
  values) in English and its content in the product language, never translated. Git (commits, MR/PR
  text, branch names) and test names, code, code comments, identifiers and the `AGENTS.md` files are
  in English.
- **Commits** follow Conventional Commits, in English, and name the requirement: `feat(api): cancel
  by WhatsApp (#812)`.
- **The guides stay true:** a change to a command, a URL, a service, a background process or a practice
  updates `README.md`, the operations guide and `AGENTS.md` in the same branch.
- **Observability is not optional for new work:** the engineering default is JSON logs with
  `trace_id`, health with version and dependency checks, and OpenTelemetry
  (`.claude/specs/engineering.md`). The adopted services do not have it yet (`docs/adopcion.md`):
  do not assume it exists.

## Never

- Commit secrets or `.env`.
- Edit a migration already on `dev`: add a new one.
- Change an API contract used by external integrators unless the specification says so.
- Contradict a decision in `_jflow/prd/decisions.md` without reverting it in a requirement.
- Edit the generated indexes by hand (`_jflow/prd/blocks/index.md`, `_jflow/prd/domain/index.md`,
  `_jflow/flows/index.md`): regenerate them with `python3 .claude/scripts/jflow.py def-index`.
- Write code-level detail in Jiku: the client may read it.
