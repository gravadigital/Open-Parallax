# api

The Go HTTP service of OpenParallax: the whole domain (events, accounts, proposals, voting engine,
notifications) and the only service with state. It exposes the HTTP API under `/api/v1` plus
`GET /health`. Its place in the product is in `_jflow/prd/architecture/services.md`; its contract is
`_jflow/apis/api.openapi.yaml` and its schema `_jflow/db-schemas/api.md`.

Code, comments, identifiers and test names are in English; the texts and error messages the service
shows its users are in the product language (`language` in `_jflow/project.yaml`; see the root
`AGENTS.md`). Today the api answers error `code`s plus English messages, and the web translates the
codes; the emails are in Spanish (`internal/email/templates.go`).

## Stack

| Concern | Choice | Version | Why |
|---|---|---|---|
| Language | Go | 1.26.6 (`go.mod`, same as the Dockerfile image) | — |
| HTTP | Gin, `gin.Default()` (its logger and recovery), gin-contrib/cors | 1.10.1 / 1.7.6 | DEC-T03 |
| Validation | Gin binding tags on the request structs, plus manual checks in the handlers | — | — |
| Persistence | GORM with the pgx Postgres driver; PostgreSQL 15; rules also in plpgsql triggers | 1.30.2 / 1.6.0 | DEC-T01, DEC-T03, DEC-T04 |
| Migrations | Own Go migrations in `internal/storage/migrations/`, applied at startup and by `cmd/migrate` | — | — |
| Auth | golang-jwt (HS256 sessions), bcrypt (`x/crypto`) | 5.3.0 / 0.58.0 | — |
| Object storage | aws-sdk-go-v2 S3 against Garage (or local disk) | s3 1.114.10 | DEC-T10 |
| Email | `net/smtp`, STARTTLS or TLS, AUTH LOGIN | stdlib | — |
| Logging | charmbracelet/log, text to stderr (not JSON, no OpenTelemetry) | 0.4.2 | — |
| Tests | `testing` + testify; testcontainers (Garage) for the S3 test; gotestsum for JUnit | 1.12.1 / 0.45.0 / 1.13.0 | — |

Chosen before the adoption; versions from `go.mod`. Before adding or upgrading a dependency, check
the official documentation for the stable version and its recommended practices.

## Commands

| Command | Runs |
|---|---|
| `make test S=api` | unit (`go test -race ./...`), integration (`-tags=integration`, against the Compose database `openparallax_test`) and end to end (`-tags=e2e ./e2e/...` against the product up); JUnit in `reports/junit/api-*.xml` |
| `make lint S=api` | `go vet` for the default, `integration` and `e2e` build tags |
| `make typecheck S=api` | `go build ./...` |
| `make fmt S=api` | `go fmt ./...` (the code is not gofmt-clean today) |
| `make migrate` | `go run ./cmd/migrate` against the Compose database |
| `make dev S=api` | `go run ./cmd/api` with the environment that points at `make infra` |

## Structure

- `cmd/api/main.go`: config, wiring and every route (groups decide auth: a route outside a group with `JWTAuthMiddleware` is public).
- `cmd/migrate/`: the standalone migration runner (`-rollback`).
- `internal/domain/{event,participant,attachment,vote,notification}/`: entities and pure rules (stages, MBC, Q, G′, the assignment algorithm in `vote/voting_service.go`).
- `internal/handlers/`: HTTP handlers per module, with their unit tests and the hand-written mocks (`mocks_*_test.go`).
- `internal/middleware/auth/`: JWT and the permission middlewares (`permissions.go`).
- `internal/storage/postgres/`: repositories and their integration tests (`*_integration_test.go`).
- `internal/storage/migrations/`: numbered migrations, registered in order in `migrations.go`; `004_constraints_and_triggers.go` holds the triggers.
- `internal/storage/`: file storage (S3 and local).
- `internal/email/`: SMTP service and the email templates.
- `e2e/`: end-to-end tests (build tag `e2e`) and `e2e/fakegoogle/`, the fake Google of the tests.

## Canonical examples

| Kind | File |
|---|---|
| Endpoint with validation and error mapping | `internal/handlers/event_handler.go` (`CreateEvent`) |
| Transaction with a guarantee | `internal/storage/postgres/voting_setup_repository.go` (`OpenVoting`, GAR-07); `vote_repository.go` (ranking replacement under `FOR UPDATE`, GAR-08) |
| Rule in the database | `internal/storage/migrations/004_constraints_and_triggers.go` |
| Unit test of a handler | `internal/handlers/event_handler_test.go` |
| Integration test | `internal/storage/postgres/vote_triggers_integration_test.go` |
| End-to-end test | `e2e/main_path_test.go` |

## Recurring errors

One line per mistake that showed up twice (in verification, in CI or in review).

- A rule enforced both in Go and in a trigger (conflict of interest, assignment size) must change in both places in the same change; a trigger violation reaches the client as a generic 500.
- Adding a method to a repository interface means adding it to the mock in `mocks_repository_test.go`, or the package stops compiling.
