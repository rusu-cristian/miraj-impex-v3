# Impex Line (miraj-impex-v3)

## Project
Impex Line: an internal, Romanian-language web app for a poultry breeder business: daily entry and reports on bird counts, eggs, feed, weights and costs per farm, series and cycle. Users are the owner, managers, vets and viewers (`design/kit/PRODUCT.md`).

## Stack
- Backend (`backend/`, not built yet): Go API service, standard `net/http` ServeMux, pgx/v5 + sqlc, goose migrations, PostgreSQL 16, log/slog, Docker. The full spec is `backend/SPEC.md` (§12 covers the Go layout and the M0–Mn build plan).
- Front end: Vue/Nuxt online-only PWA that calls the API over `/api` (assumed to live in a separate repo; not in this one).
- Design (`design/`): finished design canvas, kit and real-data extracts. It is the source of truth for UI copy and figures. Python scripts there are reference only.

## Commands
Run from `backend/` (Makefile, SPEC §12.3). Full list, tool pins and working rules: `backend/CLAUDE.md`.
test: `make test` (`go test -race ./...`, testcontainers or `TEST_DATABASE_URL`)
lint: `make lint` (`gofmt -l .`, `go vet ./...`, `go tool staticcheck ./...`, `sqlc diff`, no `time.Now()` outside `timex`)
typecheck: `go build ./...`
build: `docker build backend/` (assumed)
dev: `make run` (`go run ./cmd/api serve`); migrations: `go run ./cmd/api migrate up`
generate: `make generate` (`sqlc generate`)

## Structure
- `backend/`: `SPEC.md` (what and how), `CLAUDE.md` (backend working rules and commands), `NOTES-verify.md` (verify-at-implementation findings), `db/migrations/0001_baseline.sql` (PostgreSQL 16 DDL, the former `schema.sql`, never edited). Go module layout in SPEC §12.3.
- `design/kit/`: product, spec, screens, components, data model, decisions (`decisions-current.md`: later `Dnn` wins), open questions.
- `design/canvas/`: boards (`project/*.dc.html`, exact Romanian copy), generator scripts, reference `realdata.py`.
- `design/data/`: workbook extracts (seed and golden values to 18.11.2026) and the scripts that built them.
- `design/system/`: design tokens and components.

## Conventions
- SPEC wins over the kit where it is more precise; later decisions (`Dnn`) win over older ones and over spec text.
- Romanian strings are returned verbatim with diacritics (ș ț with comma below); number/date formats per SPEC §11.3.
- Counts are always derived from events, never stored (D01). One transaction per save with farm advisory locks (SPEC §12.1).
- Golden tests (SPEC §11.7) run with `app.today = 2026-11-18` and a fixed clock.
- `schema.sql` / `0001_baseline.sql` is never edited after M0: changes go in new numbered migrations.
- `internal/db/dbq` is sqlc output: regenerate it, never hand-edit it.

## Security notes
- Secrets come from environment variables; `.env` is never committed. Three DSNs, three roles: `MIGRATE_DATABASE_URL` (owner, migrations only), `DATABASE_URL` (`il_app`), `MAINT_DATABASE_URL` (`il_maint`).
- Auth: `il_session` httpOnly cookie (sha256-hashed in DB, sliding 30 days), argon2id passwords, login throttling, CSRF via Origin/`Sec-Fetch-Site` + JSON Content-Type (SPEC §11.4). Every permission check is server-side (§3.3); client checks are convenience only.
- `audit_log` and `live_event` are append-only for the app role.
- Never logged: request/response bodies, cookies, passwords and one-time passwords, tokens, voice transcripts and audio, Motiv texts (SPEC §11.11).
