# api — the Go HTTP service. Every target is also reachable from the root Makefile
# (`make <target> S=api`). Needs Go (version in go.mod) and Docker.

ROOT    := $(abspath ../..)
REPORTS := $(ROOT)/reports/junit
COMPOSE := docker compose --project-directory $(ROOT) -f $(ROOT)/docker-compose.yml

-include $(ROOT)/.env

# The same defaults as docker-compose.yml, so the api run by hand reaches the stack.
POSTGRES_DB          ?= openparallax_db
POSTGRES_USER        ?= openparallax
POSTGRES_PASSWORD    ?= openparallax_password
POSTGRES_PORT        ?= 5432
S3_ACCESS_KEY_ID     ?= GKopenparallaxlocal01
S3_SECRET_ACCESS_KEY ?= openparallax-local-secret-0123456789
S3_BUCKET            ?= openparallax
STORAGE_PORT         ?= 3900
API_PORT             ?= 8080
WEB_PORT             ?= 3000
MAILPIT_PORT         ?= 8025
JWT_SECRET           ?= openparallax-local-dev-secret-no-usar-en-produccion
GOOGLE_CLIENT_ID     ?=

API_ENV = \
	DB_HOST=localhost DB_PORT=$(POSTGRES_PORT) DB_SSLMODE=disable \
	DB_USER=$(POSTGRES_USER) DB_PASSWORD=$(POSTGRES_PASSWORD) DB_NAME=$(POSTGRES_DB) \
	STORAGE_PROVIDER=s3 S3_ENDPOINT=http://localhost:$(STORAGE_PORT) S3_REGION=garage \
	S3_ACCESS_KEY_ID=$(S3_ACCESS_KEY_ID) S3_SECRET_ACCESS_KEY=$(S3_SECRET_ACCESS_KEY) S3_BUCKET=$(S3_BUCKET) \
	PORT=$(API_PORT) FRONTEND_URL=http://localhost:$(WEB_PORT) GIN_MODE=debug \
	JWT_SECRET='$(JWT_SECRET)' GOOGLE_CLIENT_ID=$(GOOGLE_CLIENT_ID)

# gotestsum runs `go test` and writes the JUnit report. Pinned; `go run` fetches it once.
GOTESTSUM := go run gotest.tools/gotestsum@v1.13.0

.PHONY: up down logs check dev test test-unit test-integration test-e2e test-perf lint typecheck fmt migrate build

dev: ## Run the api with go run against `make infra` (no hot reload)
	$(API_ENV) go run ./cmd/api

test: test-unit test-integration test-e2e

test-unit: ## Unit tests: no database, no network
	@mkdir -p $(REPORTS)
	$(GOTESTSUM) --format testname --junitfile $(REPORTS)/api-unit.xml -- -race ./...

# Its own database in the stack's PostgreSQL, so the migrations under test never touch
# the data you work with. The S3 test starts its own Garage with testcontainers.
test-integration: ## Integration tests against the Compose database, migrations applied
	@mkdir -p $(REPORTS)
	$(COMPOSE) up -d --wait database
	@$(COMPOSE) exec -T database psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -tAc \
		"SELECT 1 FROM pg_database WHERE datname = 'openparallax_test'" | grep -q 1 || \
		$(COMPOSE) exec -T database createdb -U $(POSTGRES_USER) openparallax_test
	$(API_ENV) TEST_DB_NAME=openparallax_test $(GOTESTSUM) --format testname \
		--junitfile $(REPORTS)/api-integration.xml -- -tags=integration -count=1 -p 1 ./...

# Against the product already up with the test overrides (`make test-e2e` at the root
# brings it up): an HTTP client with real tokens, Mailpit for the email, the fake Google.
test-e2e: ## End-to-end tests of the HTTP interface against `make up`
	@test -n "$(E2E_UP)" || $(MAKE) -C $(ROOT) e2e-up
	@mkdir -p $(REPORTS)
	API_URL=http://localhost:$(API_PORT) MAILPIT_URL=http://localhost:$(MAILPIT_PORT) \
		$(GOTESTSUM) --format testname --junitfile $(REPORTS)/api-e2e.xml -- -tags=e2e -count=1 ./e2e/...

test-perf:
	@echo "api: no performance tests yet"

# go vet also type-checks the test files of each build tag.
lint: ## go vet (the code is not gofmt-clean today; `make fmt` formats it)
	go vet ./...
	go vet -tags=integration ./...
	go vet -tags=e2e ./e2e/...

typecheck: ## Compile the service
	go build ./...

fmt:
	go fmt ./...

migrate: ## Apply the migrations to the Compose database
	$(API_ENV) go run ./cmd/migrate

build:
	$(COMPOSE) build api

# --- The service in the product's stack (the root Makefile runs the whole product) ---

up: ## Build and start this service and what it depends on
	$(COMPOSE) up -d --build --wait api

down: ## Stop and remove this service's container
	$(COMPOSE) rm -s -f api

logs: ## Follow this service's logs
	$(COMPOSE) logs -f api

check: lint typecheck test ## lint + typecheck + every test level of this service
