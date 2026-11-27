# Contributing to OpenParallax

This document covers how to get the project running, what the codebase expects from a change,
and the rules that are not obvious from reading the code.

## Getting set up

You need **Docker**, **Go 1.26** (see [`services/api/go.mod`](services/api/go.mod)), **Node.js 22** (see
[`services/web/.nvmrc`](services/web/.nvmrc)) and `make`.

```sh
git clone https://github.com/gravadigital/Open-Parallax.git openparallax
cd openparallax
make infra    # database + Garage in Docker
make api      # in another terminal
make web      # in a third
```

No configuration file is needed. Everything is in
[documentation/installation.md](documentation/installation.md).

## How the pieces fit

Read [documentation/README.md](documentation/README.md) first. The short version:

- **`api`** is the only service with state. Gin for HTTP, GORM for the database. All routes
  are declared in [`services/api/cmd/api/main.go`](services/api/cmd/api/main.go).
- **`web`** is a React single-page app built with Create React App. It has no business logic:
  everything it shows comes from the api, through the services in `services/web/src/services/api.ts`.

How to work in the repository is in [`AGENTS.md`](AGENTS.md) and in each service's
`AGENTS.md`; the product definition — rules, decisions, contracts — is in [`_jflow/`](_jflow/), in
Spanish. The api diverges from the usual Go stack on purpose, and a change written for another
stack will not fit.

## Before touching voting or assignments

**Read [`services/api/internal/storage/migrations/004_constraints_and_triggers.go`](services/api/internal/storage/migrations/004_constraints_and_triggers.go)
first.** Part of the rules are PostgreSQL triggers, not Go code: that an assignment has exactly
`m` proposals, that nobody evaluates their own, that only assigned proposals can be voted.

- If a rule exists in both places — conflict of interest does — **change both in the same
  change**.
- A trigger violation reaches the client as a generic `500`. If a user can trigger it, validate
  it in Go first and return a `400` with a useful `code`.
- Test fixtures that insert directly have to respect the triggers, or the insert fails with an
  unhelpful plpgsql error.

The tests in `services/api/internal/domain/vote/voting_service_test.go` are the product's main safety
net. They must keep passing.

## Adding an endpoint

1. Declare it in `services/api/cmd/api/main.go`, **in the right group**. A route outside a group with
   `JWTAuthMiddleware` is public.
2. If the resource belongs to someone, add the permission middleware too. Authenticated is not
   authorised.
3. Answer errors as `{"error", "code"}`, and success wrapped in `data`.
4. Update [`_jflow/apis/api.openapi.yaml`](_jflow/apis/api.openapi.yaml) and
   [documentation/api-reference.md](documentation/api-reference.md).

## Changing the database

Migrations are Go, in `services/api/internal/storage/migrations/`, numbered in sequence and registered in
`migrations.go`. They run when the api starts.

**Adding a field to an entity is not enough**: the tables were created once by `AutoMigrate`,
so an existing database needs a migration with the `ALTER`. Implement `Down` as well.

## Tests

Every change that touches behaviour needs tests.

```sh
make test-unit          # api and web unit tests
make test-integration   # api integration tests, against the local stack's PostgreSQL
make test-e2e           # end-to-end tests of the api and the web, against the product up
make check              # lint + typecheck + every test level
```

| Project | Runner | Needs Docker |
| ------- | ------ | ------------ |
| `api` unit | `go test`, testify, hand-written mocks | no |
| `api` integration | `go test -tags=integration` | yes |
| `web` | Jest + Testing Library | no |
| `api` end to end | `go test -tags=e2e`, an HTTP client against `make up` | yes |
| `web` end to end | Playwright, in its Docker image, against `make up` | yes |

In the api, the mocks live in `mocks_*_test.go` next to the tests. **Adding a method to a
repository interface means adding it to the mock**, or the package stops compiling. Check the
error `code` in handler tests, not only the status.

In the web, query by what the user sees (`getByRole`, `getByText`), not by CSS class.

## Style

- **Go:** `gofmt`. `go vet` runs in CI.
- **Comments explain why**, not what the code does.
- **Comments in English.** The product definition in `_jflow/` and the guides in `docs/` are in
  Spanish.
- No `console.log` left in the web.

## Pull requests

- Branch off `dev` and keep the change focused on one thing.
- Say what problem it solves. If it fixes a bug, say how it reproduced.
- Update `CHANGELOG.md` under `[Unreleased]` for anything a user would notice.
- Make sure `make check` passes. CI runs its lint, typecheck, unit and integration part.

If you are planning something large, open an issue first so we can agree on the approach.

## Upgrading Go or Node

The Docker images pin their base versions, so the build does not depend on what each
machine has cached:

- **Go:** the `go` directive in `services/api/go.mod` and the `FROM golang:X.Y.Z-alpine` in
  `services/api/Dockerfile` must match. The official image runs with `GOTOOLCHAIN=local`, so an image
  older than `go.mod` fails at `go mod download`.
- **Node:** `services/web/.nvmrc` (what CI uses) and the `FROM node:N-alpine` in
  `services/web/docker/Dockerfile` must share the major version.

Change both in the same commit.

## Releases

One version for the whole repository. `scripts/set-version.sh` sets it everywhere, and pushing
a `v1.2.3` tag publishes the images. See the policy in [CHANGELOG.md](CHANGELOG.md).
