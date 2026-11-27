# The single entry point for running OpenParallax. Run `make` to list the commands.
#
# Each target delegates to services/{api,web}/Makefile; `S=api` (or `S=web`) limits it
# to one service. Everything runs against the stack in docker-compose.yml, so there is
# one database, one S3 storage (Garage) and one Mailpit whether the api runs in a
# container or with `go run`.
#
# No .env is needed. An optional .env (see .env.example) is read both by docker compose
# and by these Makefiles.

SERVICES := api web
S        ?= $(SERVICES)

COMPOSE      := docker compose
COMPOSE_TEST := docker compose -f docker-compose.yml -f docker-compose.test.yml

-include .env

API_PORT     ?= 8080
WEB_PORT     ?= 3000
MAILPIT_PORT ?= 8025

# Runs the target in every selected service, stopping at the first failure.
each = $(foreach s,$(S),$(MAKE) -C services/$(s) $(1) &&) true

.DEFAULT_GOAL := help
.PHONY: help up down stop reset logs infra dev api web build migrate \
	test test-unit test-integration test-e2e test-perf e2e-up lint typecheck fmt check

help: ## List the commands
	@awk 'BEGIN {FS = ":.*## "} /^[a-z0-9-]+:.*## / { printf "  make %-18s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# --- The whole product in Docker -----------------------------------------------

up: ## Build and start the whole product
	$(COMPOSE) up -d --build --wait
	@echo
	@echo "  web      http://localhost:$(WEB_PORT)"
	@echo "  api      http://localhost:$(API_PORT)"
	@echo "  mailpit  http://localhost:$(MAILPIT_PORT)"

down: ## Remove the containers, keeping the data
	$(COMPOSE) down

stop: ## Stop everything, keeping the data
	$(COMPOSE) stop

reset: ## Remove everything, data included
	$(COMPOSE) down -v

logs: ## Follow the logs (one service: make logs s=api)
	$(COMPOSE) logs -f $(s)

build: ## Build the images
	$(call each,build)

migrate: ## Apply the database migrations (the api also applies them when it starts)
	$(COMPOSE) up -d --wait database
	$(MAKE) -C services/api migrate

# --- Running api and web by hand -------------------------------------------------

# The api and web containers are stopped so the hand-run ones can take their ports.
infra: ## Start only the database, the storage and Mailpit
	-@$(COMPOSE) stop api web 2>/dev/null
	$(COMPOSE) up -d --wait database storage mailpit

dev: infra ## Run api (go run) and web (npm start, hot reload) by hand, against `make infra`
	$(MAKE) -j2 api web

api: ## Run only the api by hand (against `make infra`)
	$(MAKE) -C services/api dev

web: ## Run only the web by hand (against `make api`)
	$(MAKE) -C services/web dev

# --- Checks ----------------------------------------------------------------------

lint: ## Static checks
	$(call each,lint)

typecheck: ## Type checks
	$(call each,typecheck)

fmt: ## Format the code
	$(call each,fmt)

test-unit: ## Unit tests → reports/junit/{service}-unit.xml
	$(call each,test-unit)

test-integration: ## Integration tests, against the Compose database → reports/junit/
	$(call each,test-integration)

# Brings the product up with the test overrides (fake Google, seeded test identities)
# and waits until the api answers. Recreates api and web if `make up` was running.
e2e-up:
	$(COMPOSE_TEST) up -d --build --wait database storage mailpit fake-google api web
	@for i in $$(seq 1 120); do curl -sf http://localhost:$(API_PORT)/health >/dev/null && exit 0; sleep 1; done; \
		echo "the api did not answer on http://localhost:$(API_PORT)/health" >&2; exit 1
	$(COMPOSE_TEST) run --rm --no-deps e2e-seed

test-e2e: e2e-up ## End-to-end tests of every entry, against the product up → reports/junit/
	$(call each,test-e2e)

test-perf: ## Performance tests (not part of `make check`)
	$(call each,test-perf)

test: test-unit test-integration test-e2e ## Every test level

check: lint typecheck test ## What must pass before a change is done
