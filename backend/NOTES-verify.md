# Verify-at-implementation notes

SPEC items marked "verify at implementation" are checked against the real library or service and recorded here
(decision D4). One entry per item, newest status wins. Status is `open` until someone checks it, then `verified`.

Entry format:
```
### <short title>                         status: open | verified (YYYY-MM-DD, task)
- SPEC: §<n>
- Library / version: <module or tool> <version>
- Checked: <what was looked at: go doc, module source, registry, a test>
- Finding: <what is true>
- Consequence: <what the code or SPEC does because of it>
```

## M0

### Go version and golang image tags                status: verified (2026-11-27, T1)
- SPEC: §12.2, §11.13
- Library / version: Go 1.27.4; Docker Hub `library/golang`
- Checked: `https://go.dev/dl/?mode=json` (stable: go1.27.4, go1.26.11); local `go version` = go1.27.4 darwin/arm64;
  Docker Hub tags `golang:1.27.4`, `1.27.4-bookworm`, `1.27.4-trixie`, `1.27.4-alpine` all exist (HTTP 200).
- Finding: Go 1.27.4 is the current stable release and has everything §12.2 needs (≥ 1.24 `omitzero`, `crypto/hkdf`;
  ≥ 1.25 `http.CrossOriginProtection`).
- Consequence: `go.mod` has `go 1.27` and `toolchain go1.27.4`. The Dockerfile and CI pin `golang:1.27.4-<variant>`
  (variant is T8's choice).

### Distroless base image tag                       status: verified (2026-11-27, T1)
- SPEC: §11.13
- Library / version: `gcr.io/distroless/static-debian12`, `gcr.io/distroless/static-debian13`
- Checked: the gcr.io registry tag lists.
- Finding: both `static-debian12:nonroot` and `static-debian13:nonroot` exist.
- Consequence: §11.13's `static-debian12:nonroot` is valid as written; moving to `static-debian13` is optional and
  would be a SPEC change. Pin by digest in the Dockerfile if reproducibility is wanted (T8).

### staticcheck support for Go 1.27                status: verified (2026-11-27, T1)
- SPEC: §12.2 (lint), D2
- Library / version: `honnef.co/go/tools` v0.8.2 (staticcheck 2026.2.2, released 2026-10-12, the latest on the proxy)
- Checked: `go get -tool honnef.co/go/tools/cmd/staticcheck@v0.8.2`; `go tool staticcheck -version` →
  `staticcheck 2026.2.2 (0.8.2)`; `go tool staticcheck ./...` runs clean under go1.27.4.
- Finding: works with Go 1.27. Re-check when packages with generics and iterators land (T2+), since the M0 scaffold
  only has `db`.
- Consequence: pinned with a `tool` directive in `go.mod`; run as `go tool staticcheck`.

### http.CrossOriginProtection vs §11.4             status: verified (2026-11-27, T1); acted on in T4
- SPEC: §11.4 (CSRF)
- Library / version: `net/http` in Go 1.27.4 (`$GOROOT/src/net/http/csrf.go`)
- Checked: `go doc net/http.CrossOriginProtection` and the source of `(*CrossOriginProtection).Check`.
- Finding: for non-safe methods it allows `Sec-Fetch-Site: same-origin` **and `none`**, rejects any other value
  unless the Origin is trusted (`AddTrustedOrigin`) or the path is bypassed; when `Sec-Fetch-Site` is absent it
  compares the `Origin` host with the request `Host` (not with `APP_ORIGIN`, scheme ignored); when **both headers
  are absent it allows the request** ("same-origin or not a browser request").
- Consequence: it differs from §11.4 in two ways: §11.4 refuses with `403 FORBIDDEN` when both headers are absent,
  and §11.4 compares `Origin` with `APP_ORIGIN` exactly. So `http.CrossOriginProtection` alone does not implement
  §11.4. Lead's ruling (2026-11-27): §11.4 is the rule; T4 implements it in our own middleware (exact `APP_ORIGIN`
  match; `Sec-Fetch-Site: same-origin` only when `Origin` is absent; 403 when both are absent) plus the Content-Type
  rule. If `http.CrossOriginProtection` is used at all, only in a way that still meets §11.4; T4 records the choice
  here.
- T4 choice (2026-11-27): `http.CrossOriginProtection` is **not used**. `httpx.CSRF(appOrigin)` alone implements
  §11.4 on every method but GET/HEAD/OPTIONS: one `Origin` header equal to `APP_ORIGIN` byte for byte (config
  lower-cases it and, since the M0 review, drops a default port :443/:80 as browsers do; scheme and other ports count), else, with no `Origin`, `Sec-Fetch-Site: same-origin`; anything else,
  including both headers absent, `Sec-Fetch-Site: none` and two `Origin` headers, is `403 FORBIDDEN`; then
  `Content-Type` must parse as `application/json`. Every request `CrossOriginProtection` (with `APP_ORIGIN` trusted)
  refuses is refused by this rule too, so layering it would add nothing. `TestCSRFStricterThanCrossOriginProtection`
  pins the three differences (neither header, `Sec-Fetch-Site: none`, `Origin` matching only `Host`) and fails if a
  later Go changes `CrossOriginProtection` so that the comparison no longer holds. The multipart (guide preview) and
  audio (voice STT) Content-Type exceptions come with those routes.

### pgx date scanner/valuer interface names         status: verified (2026-11-27, T2)
- SPEC: §7.0a (Days), §12.4
- Library / version: `github.com/jackc/pgx/v5` v5.13.0
- Checked: `pgtype/date.go` in the module source; `DateCodec.PlanEncode` / `PlanScan`.
- Finding: the interfaces are `pgtype.DateScanner { ScanDate(pgtype.Date) error }` and
  `pgtype.DateValuer { DateValue() (pgtype.Date, error) }`. `pgtype.Date` = `{Time time.Time; InfinityModifier;
  Valid bool}`; decoded dates carry `Time` at 00:00 UTC. NULL is passed to `ScanDate` as `Date{}` (Valid false); a
  `*T` target gets nil through pgx's pointer handling. A valuer returning `Valid: false` encodes NULL.
- Consequence: `timex.Day` implements both (compile-time assertions in `day.go`). `ScanDate` refuses NULL and
  ±infinity (nullable columns map to `*timex.Day`); `DateValue` refuses the zero Day instead of sending NULL. Binary and
  text round trips are tested with `pgtype.Map` (`TestDayPgx`); T3 maps `date` to `timex.Day` in `sqlc.yaml`.

### shopspring/decimal rounding mode                status: verified (2026-11-27, T2)
- SPEC: §11.2, §5.15 (D82 cases: 168 122 ÷ 186 088 → 90,3; half away from zero from full precision)
- Library / version: `github.com/shopspring/decimal` v1.6.0
- Checked: source of `Decimal.Round` (truncate to places+1, add sign·5, truncate toward zero) and tests in
  `internal/platform/num`.
- Finding: `Round(places)` is half away from zero: −2.5 → −3, 2.5 → 3, −1.45 → −1.5, 90.345… → 90.3 at 1 dp (and
  90.35 → 90.4 when rounded twice, the case §5.15 forbids). `RoundBank` is the half-even variant, not used. `String()`
  drops trailing zeros and never uses exponent notation (`UseScientificNotation` defaults to false; we never set it).
  `Div` keeps `DivisionPrecision` = 16 digits, enough for 1-dp ratios. `NewFromString` caps exponents at
  `MaxDecodeExponent` = 10 000.
- Consequence: `num.Round` calls `Decimal.Round`. `num.Num` JSON rounds to 4 dp. `UnmarshalJSON` also refuses
  exponents beyond ±18 and literals over 40 bytes, so a short "1e9999" can't make later comparisons expensive.

### encoding/json in Go 1.27 is the v2-backed implementation  status: verified (2026-11-27, T2); acted on in T4
- SPEC: §7.0a (Decoding: "one issue naming the JSON path")
- Library / version: Go 1.27.4 standard library (`internal/buildcfg`: `JSONv2: true` in the baseline experiments)
- Checked: `go list -f '{{.GoFiles}}' encoding/json` → `v2_decode.go v2_encode.go …`; source of
  `encoding/json/v2_inject.go` and `ReportErrorsWithLegacySemantics`.
- Finding: `encoding/json` (v1 API) runs on `encoding/json/v2` with v1 options. With the default legacy error
  semantics, an error returned by a type's own `UnmarshalJSON` is returned **verbatim**: a `*json.UnmarshalTypeError`
  from `num.Num` or `timex.Day` comes back without `Field`/`Struct`, so the JSON path is lost. Errors for built-in kinds
  (a string into an `int`) still carry `Field`. Unknown fields are `json: unknown field "x"`. `encoding/json/v2` is
  importable; with `json.ReportErrorsWithLegacySemantics(false)` errors are `*jsonv2.SemanticError` with a
  `JSONPointer`, including errors from Unmarshal methods. `GOEXPERIMENT=nojsonv2` would switch back to the classic
  decoder, which does fill `Field` for method errors.
- Consequence: `num.Num`, `timex.Day` and `timex.LocalTime` return `*json.UnmarshalTypeError` with `Type` set (so
  httpx can tell `bad_date` from `bad_type`), and their tests don't rely on `Field`. T4's `httpx.Decode` must derive
  `issues[0].path` in a way that works for method errors (e.g. decode with `encoding/json/v2` and v1 options, legacy
  errors off, and map `SemanticError.JSONPointer` to the `a.b[0].c` form), with tests for a `num.Num` and a
  `timex.Day` field.
- T4 (2026-11-27), checked with probes against go1.27.4: `jsonv2.Unmarshal(data, &v, json.DefaultOptionsV1(),
  json.ReportErrorsWithLegacySemantics(false), jsonv2.RejectUnknownMembers(true))` returns
  - `*jsonv2.SemanticError` with `JSONPointer` for method errors (`/sections/0/F/mortality`, wrapping the
    `*json.UnmarshalTypeError`, `GoType` = `num.Num` / `timex.Day`), for built-in kinds (a string into an `int`,
    `Err` nil) and for unknown members (`errors.Is(err, jsonv2.ErrUnknownName)`, pointer includes the member name);
  - `*jsontext.SyntacticError` (with a pointer) for malformed JSON, an empty body and a second top-level value
    (`invalid character '{' after top-level value`).
  Everything else is v1 behaviour: names match case-insensitively (`"Login"` fills `login`), duplicate names keep the
  last value, invalid UTF-8 becomes U+FFFD, and a type's `UnmarshalJSONFrom` is called for JSON `null` (so
  `httpx.Optional` sees null). Consequence: `httpx.Decode` reads the body (≤ limit) and calls `jsonv2.Unmarshal` with
  those options; `jsonPath` turns the pointer into `sections[1].F.mortality`, using the document to tell array
  indexes from numeric map keys. Codes: `unknown_field`, `bad_date` (a `timex.Day`/`LocalTime` target), `bad_type`,
  `malformed` (syntax, empty, trailing value), `bad_content_type`. Tests: `TestDecodeErrors`, `TestOptional`.

### net/http ServeMux sets r.Pattern on the request it receives   status: verified (2026-11-27, T4)
- SPEC: §11.11 (access log: the matched route pattern, `r.Pattern`)
- Library / version: `net/http` in Go 1.27.4 (`$GOROOT/src/net/http/server.go`, `(*ServeMux).ServeHTTP`)
- Checked: source (`h, r.Pattern, r.pat, r.matches = mux.findHandler(r)`) and `TestAccessLog` through a real mux.
- Finding: the mux writes `Pattern` into the `*http.Request` it is given; a middleware that calls `r.WithContext`
  between the access log and the mux hands the mux a copy, and the access log then sees an empty pattern.
- Consequence: `httpx.AccessLog` must sit inside `RequestID` and `ClientIP` (which add to the context) and outside
  only middleware that pass `r` through unchanged (`Recover`, `SecurityHeaders`, `CORS`, `CSRF`) before the mux.
  Per-route middleware (auth) runs inside the mux and may replace `r`; it reports the user with `httpx.SetUserID(w,
  id)`. T6 documents the order in code.

### How goose serialises concurrent runs            status: verified (2026-11-27, T3)
- SPEC: §11.9
- Library / version: `github.com/pressly/goose/v3` v3.30.2 (Provider API), driver `github.com/jackc/pgx/v5/stdlib`
- Checked: `goose.NewProvider`, `goose.WithSessionLocker`, source of `lock/postgres.go`; test `TestMigrateConcurrentRuns`.
- Finding: without an option the Provider does **no** locking. `lock.NewPostgresSessionLocker()` holds
  `pg_try_advisory_lock(4097083626)` on a dedicated session for the whole run, retried every 5 s up to 60 times by
  default (`WithLockTimeout(period s, attempts)`); unlock retries 2 s × 30. Each migration file runs in its own
  transaction. `NewProvider` refuses a nil `*sql.DB`, so the newest embedded version is read from the file names
  (`goose.NumericComponent`).
- Consequence: `db.NewMigrator` uses the session locker with `WithLockTimeout(1, 300)` (retry every second, give up
  after 5 min). Two concurrent `api migrate up` runs both exit 0 and each file is applied once
  (`TestMigrateConcurrentRuns`). `db.LatestVersion` feeds /readyz (T6).

### sqlc.yaml keys                                  status: verified (2026-11-27, T3)
- SPEC: §12.4
- Library / version: sqlc CLI v1.31.1 (local Homebrew)
- Checked: `sqlc generate`, `sqlc vet`, `sqlc diff` on `db/migrations` (0001 unchanged, 0002) + `db/queries/auth.sql`.
- Finding: the full schema parses unchanged (views, functions, enums, the exclusion constraint, the generated
  column). Every §12.4 key is valid. Override `db_type` names: `uuid`, `date` and `pg_catalog.numeric` work as
  written; `timestamptz` must be written `timestamptz` (`pg_catalog.timestamptz` is silently ignored). Without a
  timestamptz override, pgx/v5 output uses `pgtype.Timestamptz`. sqlc infers `max(ts)` as `interface{}` and a cast
  `max(ts)::timestamptz` as non-null, so a nullable aggregate needs `coalesce(…)`. A parameter compared with a
  nullable column becomes a pointer unless cast (`sqlc.arg(now)::timestamptz`). A leading `--` comment block before
  the first `-- name:` becomes that query's doc comment (use `/* … */` for file headers). `emit_empty_slices` gives
  `[]T{}` for `:many`.
- Consequence: `sqlc.yaml` is §12.4 with the module path filled in, plus `timestamptz` → `time.Time` /
  `*time.Time` (SPEC §7.0a: instants are `time.Time` in Go). The auth queries use the casts and `coalesce` above.
  No `database:` (database-backed analysis) block was needed.

### testcontainers-go vs Docker 29 API              status: verified (2026-11-27, T3)
- SPEC: §11.8
- Library / version: `github.com/testcontainers/testcontainers-go` and `modules/postgres` v0.45.0 (moby client
  v0.6.2, API module v1.56.1); local Docker Desktop, Engine 29.6.2, API 1.55
- Checked: `postgres.Run(ctx, "postgres:16.16-bookworm", postgres.BasicWaitStrategies())`, `ConnectionString`,
  `testcontainers.TerminateContainer`; `go test -race ./...` starts one container per test package process, reaper
  (Ryuk) included.
- Finding: works with the Docker 29 API (version negotiation, no minimum-API errors). Startup ≈ 1–2 s per package with
  the image cached. With `TEST_DATABASE_URL` set, no container is started.
- Consequence: `internal/testdb` uses the module as above. Two `go test` runs (6 test processes) at the same time
  against one `TEST_DATABASE_URL` passed with no collisions and left no databases behind.

### http.Server timeouts and graceful shutdown       status: verified (2026-11-27, T6)
- SPEC: §11.12, §9.1 ("http.Server timeouts caveat")
- Library / version: `net/http` in Go 1.27.4 (`go doc net/http.Server.Shutdown`)
- Checked: the doc of `Server.Shutdown` and `Serve`; `TestServeGracefulShutdown`, `TestServeShutdownBudget`, `TestRunServe`.
- Finding: `Shutdown` closes the listeners first, then idle connections, then waits for active ones; `Serve` returns
  `ErrServerClosed` at once, so the caller must wait for `Shutdown`. On an expired context it returns the context's
  error and leaves the remaining connections open (`Close` ends them). Hijacked connections (the M12 WebSocket) are
  not waited for: `RegisterOnShutdown` hooks must end them.
- Consequence: `cmd/api/serve.go` sets `ReadHeaderTimeout` 10 s, `ReadTimeout` 30 s, `WriteTimeout` 60 s,
  `IdleTimeout` 120 s (the §9.1 values; `MaxHeaderBytes` stays at the 1 MiB default), and `ErrorLog` through slog at
  warn level. Shutdown on SIGINT/SIGTERM (context cancel): `/readyz` answers 503 `shutdown` first, then `Shutdown`
  with a 20 s budget (from a `context.WithoutCancel` of the signal context), `Close` if the budget runs out (exit 1),
  then both pools close. Requests in flight keep their own context (the server's base context is not the signal
  context), so their database calls finish. There is no drain delay between the 503 and `Shutdown` (§11.12 names
  none); add one if a load balancer needs time to notice.

### Image tags for the Dockerfile, compose and CI   status: verified (2026-11-27, T8)
- SPEC: §11.13, §12.2; plan D2, D3, D7
- Library / version: `golang:1.27.4-bookworm`, `gcr.io/distroless/static-debian12:nonroot`, `postgres:16.16-bookworm`,
  `caddy:2.11.9-alpine`; sqlc v1.31.1 linux/amd64; `actions/checkout` v7.0.1, `actions/setup-go` v7.1.0
- Checked: Docker Hub tag API (golang 1.27.4 variants; postgres 16.16 is the newest 16.x; caddy 2.11.9-alpine has the
  same digest as `caddy:alpine` and `2-alpine`, sha256:003a9780…836f0); the sqlc v1.31.1 release assets (tarball
  sha256 497ae4fc…ca27354, pinned in CI); action tags resolved to commit SHAs with `gh api`. Then `make docker`,
  `make up` and `make e2e` on Docker 29.6.2 / Compose v5.3.1.
- Finding: all tags exist and work together. The image is 37 MB and runs as 65532:65532 (`nonroot`); with no
  environment it exits 1 with one `config: missing required environment variables: …` line; with the env, `/healthz`
  is 200. hadolint v2.15.1 flags `USER nonroot:nonroot` (DL3066, non-numeric user).
- Consequence: the Dockerfile pins tags (no digests, as elsewhere in the project) and uses `USER 65532:65532`, the
  numeric form of §11.13's `nonroot:nonroot`. CI pins actions by SHA and sqlc by checksum. Moving to
  `static-debian13` stays optional and would be a SPEC change.
