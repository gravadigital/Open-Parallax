# Entry point for working on OpenParallax locally. Run `make` to list the commands.
#
# Everything runs against the stack in docker-compose.yml, so there is one
# database and one S3 storage (Garage) whether the api runs in a container or with `go run`.
#
# No .env is needed. .env is optional and is read both by docker compose and
# by this Makefile, so a GOOGLE_CLIENT_ID set there reaches every way of running.

COMPOSE := docker compose -f docker-compose.yml

-include .env

# Same defaults as docker-compose.yml. The api's own defaults in Go
# (openparallax / openparallax_db, local file storage) do not match the stack, so
# running it outside Docker needs these passed explicitly.
POSTGRES_DB       ?= openparallax_db
POSTGRES_USER     ?= openparallax
POSTGRES_PASSWORD ?= openparallax_password
POSTGRES_PORT     ?= 5432
S3_ACCESS_KEY_ID     ?= GKopenparallaxlocal01
S3_SECRET_ACCESS_KEY ?= openparallax-local-secret-0123456789
S3_BUCKET            ?= openparallax
STORAGE_PORT         ?= 3900
API_PORT          ?= 8080
WEB_PORT          ?= 3000
JWT_SECRET        ?= openparallax-local-dev-secret-no-usar-en-produccion
GOOGLE_CLIENT_ID  ?=

API_ENV = \
	DB_HOST=localhost DB_PORT=$(POSTGRES_PORT) DB_SSLMODE=disable \
	DB_USER=$(POSTGRES_USER) DB_PASSWORD=$(POSTGRES_PASSWORD) DB_NAME=$(POSTGRES_DB) \
	STORAGE_PROVIDER=s3 S3_ENDPOINT=http://localhost:$(STORAGE_PORT) S3_REGION=garage \
	S3_ACCESS_KEY_ID=$(S3_ACCESS_KEY_ID) S3_SECRET_ACCESS_KEY=$(S3_SECRET_ACCESS_KEY) S3_BUCKET=$(S3_BUCKET) \
	PORT=$(API_PORT) FRONTEND_URL=http://localhost:$(WEB_PORT) GIN_MODE=debug \
	JWT_SECRET='$(JWT_SECRET)' GOOGLE_CLIENT_ID=$(GOOGLE_CLIENT_ID)

.DEFAULT_GOAL := help
.PHONY: help up stop down reset logs infra api web test test-integration

help: ## List the commands
	@awk 'BEGIN {FS = ":.*## "} /^[a-z-]+:.*## / { printf "  make %-18s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# --- The whole stack in Docker ------------------------------------------------

up: ## Build and start everything in Docker
	$(COMPOSE) up -d --build
	@echo
	@echo "  web    http://localhost:$(WEB_PORT)"
	@echo "  api    http://localhost:$(API_PORT)"

stop: ## Stop everything, keeping the data
	$(COMPOSE) stop

down: ## Remove the containers, keeping the data
	$(COMPOSE) down

reset: ## Remove everything, data included
	$(COMPOSE) down -v

logs: ## Follow the logs (one service: make logs s=api)
	$(COMPOSE) logs -f $(s)

# --- Running api and web by hand ----------------------------------------------

# The api and web containers are stopped so `make api` and `make web` can take
# their ports. The storage (Garage) creates the bucket on start, before it reports healthy, so
# --wait returns with the bucket in place: the api checks it and does not create it.
infra: ## Start only the database and the storage, to run api and web by hand
	-@$(COMPOSE) stop api web 2>/dev/null
	$(COMPOSE) up -d --wait database storage

api: ## Run the api with go run, against `make infra`
	cd services/api && $(API_ENV) go run ./cmd/api

web: services/web/node_modules ## Run the web with npm start, against `make api`
	cd services/web && BROWSER=none PORT=$(WEB_PORT) \
		REACT_APP_API_URL=http://localhost:$(API_PORT) \
		REACT_APP_GOOGLE_CLIENT_ID=$(GOOGLE_CLIENT_ID) \
		npm start

# --- Tests --------------------------------------------------------------------

test: services/web/node_modules ## Unit tests of api and web (no Docker needed)
	cd services/api && go vet ./... && go test -race ./...
	cd services/web && CI=true npm test

# Uses its own database in the stack's PostgreSQL, so running the migrations
# under test never touches the data you work with.
test-integration: ## Integration tests of the api, against the stack's database
	$(COMPOSE) up -d --wait database
	@$(COMPOSE) exec -T database psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -tAc \
		"SELECT 1 FROM pg_database WHERE datname = 'openparallax_test'" | grep -q 1 || \
		$(COMPOSE) exec -T database createdb -U $(POSTGRES_USER) openparallax_test
	cd services/api && $(API_ENV) TEST_DB_NAME=openparallax_test go test -tags=integration -count=1 ./...

services/web/node_modules: services/web/package-lock.json
	cd services/web && npm ci
	@touch $@
