# web — the React single-page app. Every target is also reachable from the root Makefile
# (`make <target> S=web`). Needs Node.js (version in .nvmrc) and Docker.

ROOT    := $(abspath ../..)
REPORTS := $(ROOT)/reports/junit
COMPOSE := docker compose --project-directory $(ROOT) -f $(ROOT)/docker-compose.yml

-include $(ROOT)/.env

API_PORT         ?= 8080
WEB_PORT         ?= 3000
GOOGLE_CLIENT_ID ?=

# Same version as e2e/package.json (@playwright/test): the image carries its browsers.
PLAYWRIGHT_IMAGE := mcr.microsoft.com/playwright:v1.65.1-noble

.PHONY: up down logs check dev test test-unit test-integration test-e2e test-perf lint typecheck fmt migrate build

node_modules: package-lock.json
	npm ci
	@touch $@

e2e/node_modules: e2e/package-lock.json
	cd e2e && npm ci
	@touch $@

dev: node_modules ## npm start (hot reload) against `make dev S=api`
	BROWSER=none PORT=$(WEB_PORT) \
		REACT_APP_API_URL=http://localhost:$(API_PORT) \
		REACT_APP_GOOGLE_CLIENT_ID=$(GOOGLE_CLIENT_ID) \
		npm start

test: test-unit test-integration test-e2e

test-unit: node_modules ## Jest (react-scripts), one pass, JUnit report
	@mkdir -p $(REPORTS)
	CI=true JEST_JUNIT_OUTPUT_DIR=$(REPORTS) JEST_JUNIT_OUTPUT_NAME=web-unit.xml \
		npm test -- --reporters=default --reporters=jest-junit

test-integration:
	@echo "web: no integration tests (the web has no database of its own)"

# Playwright in its official image, against the product up (`make test-e2e` at the root
# brings it up). Host network, so the browser reaches the web and the api as a user does.
test-e2e: e2e/node_modules ## Playwright against `make up`
	@test -n "$(E2E_UP)" || $(MAKE) -C $(ROOT) e2e-up
	@mkdir -p $(REPORTS) $(ROOT)/reports/screenshots
	docker run --rm --network host --ipc host --user $$(id -u):$$(id -g) -e HOME=/tmp \
		-e WEB_URL=http://localhost:$(WEB_PORT) -e API_URL=http://localhost:$(API_PORT) \
		-e PLAYWRIGHT_JUNIT_OUTPUT_FILE=/reports/web-e2e.xml -e SCREENSHOTS_DIR=/screenshots \
		-v $(CURDIR)/e2e:/e2e -v $(REPORTS):/reports -v $(ROOT)/reports/screenshots:/screenshots -w /e2e \
		$(PLAYWRIGHT_IMAGE) npx playwright test

test-perf:
	@echo "web: no performance tests yet"

lint: node_modules
	npm run lint

typecheck: node_modules
	npm run typecheck
	cd e2e && npm ci --silent && npx tsc --noEmit

fmt:
	@echo "web: no formatter configured (there is no Prettier in the project)"

migrate:
	@echo "web: no database"

build:
	$(COMPOSE) build web

# --- The service in the product's stack (the root Makefile runs the whole product) ---

up: ## Build and start this service and what it depends on
	$(COMPOSE) up -d --build --wait web

down: ## Stop and remove this service's container
	$(COMPOSE) rm -s -f web

logs: ## Follow this service's logs
	$(COMPOSE) logs -f web

check: lint typecheck test ## lint + typecheck + every test level of this service
