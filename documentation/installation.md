# Installation

Everything runs from the repository root with `make`. Run `make` alone to list the commands.

## What you need

| | |
|---|---|
| **Docker** with Compose v2 | for the stack: PostgreSQL, Garage (S3 storage), and optionally the api and web themselves |
| **Go 1.26** | only to run or test the api outside Docker — see [`api/go.mod`](../api/go.mod) |
| **Node.js 22** | only to run or test the web outside Docker — see [`web/.nvmrc`](../web/.nvmrc) |
| **make** | comes with Linux and macOS |

No configuration file is needed: every setting has a working local default. See
[configuration.md](configuration.md) for what can be changed.

## Everything in Docker

The quickest way to see it working. Builds both services from the repository:

```sh
make up
```

| | |
|---|---|
| web | http://localhost:3000 |
| api | http://localhost:8080 |

The database starts empty and the api runs the migrations when it starts. A few sample users
and an event are loaded on the first start; they have no password, so to log in, register a
new account from the web.

```sh
make stop     # stop everything, keeping the data — the one to use at the end of the day
make down     # remove the containers, keeping the data
make reset    # remove everything, data included
make logs     # follow the logs; one service with: make logs s=api
```

Garage has no web console. To look at the stored files, use its CLI inside the container:

```sh
docker compose -f deploy/docker-compose.yml exec garage /garage bucket info telescope
docker compose -f deploy/docker-compose.yml exec garage /garage key list
```

`make up` always rebuilds, so code changes are picked up. Configuration changes do not need a
rebuild: the containers read it when they start.

## Each part by hand

For working on the code: the database and Garage in Docker, the api and web running natively
with their own reload. Three terminals:

```sh
make infra    # 1. database + Garage, which creates the bucket
make api      # 2. the api with go run, on :8080
make web      # 3. the web with npm start, on :3000
```

- `make infra` stops the `api` and `web` containers if they are running, so their ports are
  free. It uses the same database as `make up`: switching between the two keeps your data.
- `make api` passes the api the settings that point at the stack. The api's built-in defaults
  do not match it, which is why running `go run` directly fails to connect.
- `make web` installs the dependencies the first time. It reloads on every change; the api does
  not, so restart `make api` after changing Go code.

## Tests

```sh
make test               # api (go vet + go test -race) and web (Jest) — no Docker needed
make test-integration   # api integration tests, against the stack's PostgreSQL and a Garage of their own
```

The integration tests use their own database, `telescope_test`, inside the stack's PostgreSQL,
so they never touch the data you work with. The storage tests start their own throwaway
Garage container with testcontainers, so they need Docker but not the stack's Garage.

CI runs the same unit suites on every push and pull request — see
[`.github/workflows/ci.yml`](../.github/workflows/ci.yml). The integration tests do not run
in CI.

## Troubleshooting

**`port is already allocated`** — something else is using that port. Change it in
`deploy/.env` (`WEB_PORT=3001`, `API_PORT=8081`, …); see
[configuration.md](configuration.md).

**The api does not start: `bucket "telescope" not reachable`** — the api checks the bucket on
start and never creates it. Garage creates it when it starts, so `make logs s=garage` shows
what happened. A `400` in that message means `S3_REGION` does not match Garage's `s3_region`
(`garage`).

**Garage does not start after changing `S3_SECRET_ACCESS_KEY`** — the new secret does not match
the key already in its volume. `make reset` starts over, data included.

**The web loads but shows no data** — the api is not running, or not on the port the web was
built for. `curl http://localhost:8080/health` should answer `"status":"ok"`.

**Logged out after changing `JWT_SECRET`** — expected: existing tokens are no longer valid.
Log in again.

## On a server

This repository only runs Telescope locally. The published images are
`gravadigital/telescope-api` and `gravadigital/telescope-web` on Docker Hub; how a server runs
them is kept in a separate deployment repository. See [configuration.md](configuration.md) for
what a server must set.
