# Impex Line backend. Run every target from backend/ (SPEC §12.3).
# Tests need Docker (testcontainers) or TEST_DATABASE_URL pointing at a PostgreSQL 16.

SQLC ?= sqlc
# Arguments for `make migrate`, e.g. `make migrate ARGS=status`.
ARGS ?= up

.PHONY: generate build lint test run migrate

## generate: regenerate internal/db/dbq from db/queries (sqlc).
generate:
	$(SQLC) generate

## build: compile every package.
build:
	go build ./...

## lint: gofmt check, go vet, staticcheck, sqlc diff, and no time.Now/Since/Until outside internal/platform/timex.
lint:
	@out="$$(gofmt -l .)"; if [ -n "$$out" ]; then echo "gofmt -l: these files need gofmt:"; echo "$$out"; exit 1; fi
	go vet ./...
	go tool staticcheck ./...
	$(SQLC) diff
	@out="$$(grep -rnE --include='*.go' --exclude='*_test.go' 'time\.(Now|Since|Until)([^A-Za-z0-9_]|$$)' . | grep -vE '^(\./)?internal/platform/timex/' || true)"; \
	if [ -n "$$out" ]; then echo "time.Now/Since/Until outside internal/platform/timex (inject timex.Clock):"; echo "$$out"; exit 1; fi

## test: every test with the race detector, against a real PostgreSQL 16.
test:
	go test -race ./...

## run: start the API (needs the §11.10 environment).
run:
	go run ./cmd/api serve

## migrate: apply migrations with MIGRATE_DATABASE_URL (ARGS=up|status).
migrate:
	go run ./cmd/api migrate $(ARGS)

# Docker and the dev stack (compose.yaml: PostgreSQL 16, migrate, api, Caddy proxy on https://localhost:8443).
IMAGE ?= impex-api
COMPOSE ?= docker compose

.PHONY: docker up down e2e

## docker: build the API image (Dockerfile, SPEC §11.13).
docker:
	docker build -t $(IMAGE) .

## up: build and start the dev stack; returns once the API is ready behind the proxy. Needs .env (from .env.example),
## or COMPOSE_ENV_FILES naming another env file (e.g. COMPOSE_ENV_FILES=.env.example for a throwaway run).
up:
	@test -f .env || test -n "$$COMPOSE_ENV_FILES" || { echo "missing .env: cp .env.example .env (dev placeholders), then edit it"; exit 1; }
	$(COMPOSE) up -d --build --wait

## down: stop the dev stack (keeps the database volume; `docker compose down -v` also drops it).
down:
	$(COMPOSE) down

## e2e: login end to end through the reverse proxy (scripts/e2e-login.sh), on a started stack.
e2e: up
	./scripts/e2e-login.sh
