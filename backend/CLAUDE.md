# Impex Line backend: working rules

Go API service for Impex Line. `SPEC.md` (this directory) is the source of truth; decisions in
`../design/kit/decisions-current.md` win over older spec text (later `Dnn` wins). Module path:
`github.com/rusu-cristian/miraj-impex-v3/backend`. Every command below runs from `backend/`.

## How the work is done
- **Milestone order**: M0 → M1 → … → M12, exactly as the table in SPEC §12.6. Work on one milestone at a time.
  When a milestone is done, **stop and report** before starting the next one.
- **Done for a milestone** means all of these pass:
  - `go build ./...`
  - `gofmt -l .` prints nothing
  - `go vet ./...`
  - `go tool staticcheck ./...`
  - `sqlc diff`
  - `go test -race ./...` against a real PostgreSQL 16 (testcontainers, or `TEST_DATABASE_URL`), with
    `app.today = '2026-11-18'` on every test pool and `timex.FixedClock(2026-11-18T12:00 Europe/Chisinau)`,
    including the milestone's tests and its golden values (SPEC §11.7, §11.8)
  - plus the milestone's "Done when" column in SPEC §12.6.
- **Never change** product rules, API paths, JSON shapes or Romanian messages on your own. On any ambiguity,
  or when something doesn't work in practice, stop and ask; don't guess.
- When the user agrees a change, update `SPEC.md` **in the same commit** as the code (new copy marked *(proposed)*).
- **Verify at implementation**: every SPEC item marked "verify" is checked against the real library API (go doc or
  module source) and recorded in [`NOTES-verify.md`](NOTES-verify.md) (format at the top of that file).
- **Commits**: one commit per milestone, message `backend(M<n>): <summary>`, pushed to branch `team/backend`.
- **Voice (M12)** uses `internal/voice/providers/fake` until the user picks the STT and intent providers (SPEC §13,
  open question 13). Production refuses `fake` (§11.10).

## Standing rules
- `db/migrations/0001_baseline.sql` (the former `schema.sql`) is **never edited**. Schema changes are new numbered
  migrations (`0002_roles.sql`, …). If sqlc or goose seem to need a change in `0001`, stop and ask.
- `internal/db/dbq` is sqlc output: change `db/queries/*.sql` and run `make generate`; never hand-edit it.
- No `time.Now()`, `time.Since` or `time.Until` outside `internal/platform/timex` (test files excepted). Services
  take a `timex.Clock`; auth SQL takes `now` as a parameter (D5). `make lint` greps for all three.
- Never logged (SPEC §11.11): request or response bodies, cookies, `Authorization`, passwords and one-time passwords,
  tokens, voice transcripts and audio, Motiv texts.
- Romanian strings live only in `internal/platform/messages/ro.go`, verbatim, with ș ț (comma below).
- Handlers are thin: `httpx.Decode` → `Validate()` → authorize → service → DTO → `httpx.JSON`; errors leave only
  through `httpx.WriteError` (SPEC §12.1, §12.3).
- Secrets come from the environment (§11.10); `.env` files are never committed.
- Database roles: three DSNs, three roles. `MIGRATE_DATABASE_URL` (owner, `api migrate` only), `DATABASE_URL`
  (`il_app`, `api serve`), `MAINT_DATABASE_URL` (`il_maint`, purge). The roles `il_app` and `il_maint` are created
  by infrastructure (the compose init script, `internal/testdb` in tests, the DBA in production) **before**
  `api migrate up` runs, because `0002_roles.sql` only grants.
- A migration that adds a table or view also grants on it to `il_app` / `il_maint` (there are no default
  privileges, so append-only tables never get update/delete by accident).
- Integration tests use `internal/testdb`: `func TestMain(m *testing.M) { testdb.Main(m) }`, then `testdb.New(t)` per
  test (own database copied from a migrated template, `il_app` pool with `app.today = 2026-11-18`, frozen clock;
  `d.Pool(testdb.RoleMaint)` for another role). Without `TEST_DATABASE_URL` it starts a `postgres:16` container per
  test package; `TEST_DATABASE_URL` must be a **superuser** DSN (e.g. the compose db's `postgres` user), because
  testdb creates databases and switches roles with `SET ROLE` (it never needs or changes the roles' passwords).

## Commands
| Command | What it runs |
|---|---|
| `make generate` | `sqlc generate` (regenerates `internal/db/dbq` from `db/queries`) |
| `make build` | `go build ./...` |
| `make lint` | `gofmt -l .` must print nothing, `go vet ./...`, `go tool staticcheck ./...`, `sqlc diff`, the `time.Now`/`Since`/`Until` grep |
| `make test` | `go test -race ./...` (Docker for testcontainers, or `TEST_DATABASE_URL=postgres://…`) |
| `make run` | `go run ./cmd/api serve` (needs the §11.10 environment) |
| `make migrate` | `go run ./cmd/api migrate up`; `make migrate ARGS=status` for the status |
| `make docker` | build the image (`docker build .`) |
| `make up` / `make down` | start / stop the dev stack (PostgreSQL 16, API, Caddy proxy on `https://localhost:8443`) |
| `make e2e` | login end to end through the reverse proxy |

`make docker|up|down|e2e` come with the Docker and compose setup (`Dockerfile`, `compose.yaml`).

Raw equivalents: `go build ./...`, `go vet ./...`, `go tool staticcheck ./...`, `gofmt -l .`, `go test -race ./...`,
`go test -race -run TestName ./internal/auth`, `sqlc generate`, `sqlc diff`, `go run ./cmd/api serve`,
`go run ./cmd/api migrate up|status`, `go run ./cmd/create-owner --name "…" --login …`.

Tool pins (D2): Go `1.27` with `toolchain go1.27.4` (`go.mod`); staticcheck via the `tool` directive in `go.mod`
(`go tool staticcheck -version`); sqlc CLI v1.31.1; goose is a library (`pressly/goose/v3`), no CLI.
