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

## M1

### Base image digests: golang build image          status: verified (2026-11-27, M1 T2)
- SPEC: §11.13, §12.2; plan D2, D7
- Library / version: `golang:1.27.4-bookworm` on Docker Hub (`registry-1.docker.io/library/golang`)
- Checked (2026-11-27): `docker buildx imagetools inspect golang:1.27.4-bookworm` and a registry `HEAD
  /v2/library/golang/manifests/1.27.4-bookworm` with an OCI index `Accept` header; both give the same
  `docker-content-digest`.
- Finding: the tag is an OCI image index (`application/vnd.oci.image.index.v1+json`), digest
  `sha256:ce59b41415e2524ae99926fe1f37ce48ba8c194dcbd2a56a65436cdcefa2bd34` (built 2026-11-24), covering linux/amd64,
  arm64/v8, arm/v7, 386 and ppc64le plus attestations. `docker build` resolves it on Docker 29.6.2 (arm64).
- Consequence: the Dockerfile's build stage is `golang:1.27.4-bookworm@sha256:ce59b414…a2bd34`. This supersedes the
  "tags, no digests" consequence of the M0 image-tags entry for the Dockerfile; compose and CI pins are unchanged.

### Base image digests: distroless runtime image    status: verified (2026-11-27, M1 T2)
- SPEC: §11.13
- Library / version: `gcr.io/distroless/static-debian12:nonroot`
- Checked (2026-11-27): `docker buildx imagetools inspect gcr.io/distroless/static-debian12:nonroot` and a registry
  `HEAD https://gcr.io/v2/distroless/static-debian12/manifests/nonroot` with an OCI index `Accept` header; same digest.
- Finding: OCI image index, digest `sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab`,
  covering linux/amd64, arm64/v8, arm/v7, s390x and ppc64le. The image built on it is unchanged in behaviour: runs as
  65532:65532; with no environment `docker run --rm impex-api` exits 1 with the one `config: missing required
  environment variables: …` line; `make e2e` passes.
- Consequence: the runtime stage is `gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872…ff57f7ab`.

### How to bump the base image digests             status: verified (2026-11-27, M1 T2)
- SPEC: §11.13
- Procedure (from `backend/`), per image:
  1. `docker buildx imagetools inspect <image>:<tag>`: take the top-level `Digest:` and check that `MediaType` is an
     image index (`…oci.image.index…` or `…manifest.list…`). A per-platform digest from `docker pull`/`docker
     images --digests` on one machine is the wrong one: it would pin one architecture.
  2. Optionally cross-check with the registry: `HEAD /v2/<repo>/manifests/<tag>` with
     `Accept: application/vnd.oci.image.index.v1+json`, header `docker-content-digest`.
  3. Replace only the digest after `@` in the `FROM` line; keep the tag (it documents the version; the digest wins).
     A Go version bump changes the golang tag, `go.mod`'s `toolchain` line and CI's `GO_VERSION` together (D2).
  4. `make docker`; `docker run --rm impex-api` must exit 1 with the missing-variables line; `make e2e`
     (`COMPOSE_ENV_FILES=.env.example make e2e` without a `.env`); push and check the CI `docker` job.
  5. Update the two entries above (digest, date).
- When: the distroless `nonroot` tag is rebuilt for Debian security fixes and golang for each patch release, so the
  pins go stale; bump at least once per milestone and on any Go or Debian security release.

### calc readers vs the dense view's app_today() reach   status: verified (2026-11-27, M1 T3; review COR-1)
- SPEC: §4.12, §5.3–§5.8; plan M1 T3 criterion 2 ("never app_today() implicitly"), D4 (Verify does not set app.today)
- Library / version: `0001_baseline.sql` views on PostgreSQL 16.16; sqlc v1.31.1
- Checked: the view definitions; `TestReadFarmDay`, `TestLayDay`, `TestFeedKg` ("after app.today" cases, 31.12.2026
  with app.today 18.11.2026);
  generated `dbq/calc.sql.go`.
- Finding: `v_section_sex_day` (and so `v_farm_sex_day`) lists days only up to `greatest(max(event date),
  app_today())`, so a filter on a later day returns no row rather than the last count. `v_section_sex_delta`
  and `v_bird_event` have no such bound. `v_egg_day` and `v_feed_consumption` read `v_section_sex_day` only for days
  that have an egg or allocation row, so a row dated after that bound gets no count. sqlc types an aggregate cast `::integer` /
  `::numeric` wrapped in `coalesce` as non-null; reusing `sqlc.arg(day)` binds one parameter.
- Consequence: `calc` bird counts (`ReadFarmDay`, `SectionsAliveEnd`, `CycleDeaths`, `Sold`, `FarmBase`) are running
  sums over `v_section_sex_delta` / `v_bird_event` up to the asked day (the §5.3 formula), so they hold for any day
  and on a pool without app.today. Lay and feed (review COR-1) take the hens / birds at the start of the day the
  same way, as a running sum over `v_section_sex_delta` before the day, not from `v_egg_day` / `v_feed_consumption`:
  those join the dense view, so an egg or allocation row dated after `greatest(last event, app_today())` would read
  0 hens / 0 kg (`TestLayDay`, `TestFeedKg` "after app.today", 31.12.2026). Stock still reads `v_feed_stock`
  (stock = the latest row ≤ the day), whose consumption comes from `v_feed_consumption`: an allocation dated after
  that bound would count 0 kg there (open for the M3 stock work; SPEC §3.2 refuses future-dated writes).
  `FarmBase` counts a transfer-in row without `peer_section_id` (allowed for migration rows) as from outside only
  when its entry is a `transfer` entry whose source farm is another farm.

### pgx and context.Canceled                         status: verified (2026-11-27, M1 T1)
- SPEC: §11.11; plan M1-D8 (follow-up COR-3)
- Library / version: `github.com/jackc/pgx/v5` v5.13.0 (`pgconn/errors.go`: `normalizeTimeoutError`, `errTimeout`,
  `contextAlreadyDoneError`; pgxpool acquire)
- Checked: the source, and probes against PostgreSQL 16 (testcontainers) and a silent TCP listener.
- Finding: every cancellation pgx returns satisfies `errors.Is(err, context.Canceled)`, in four shapes:
  - canceled during a query (`select pg_sleep(5)` through the pool, sqlc `:one` included): bare `context.Canceled`
    ("context canceled"); `normalizeTimeoutError` turns the deadline-interrupted read back into the context error;
  - a pool call with an already-canceled context: bare `context.Canceled` from the acquire;
  - a `*pgx.Conn` call with an already-done context: `*pgconn.errTimeout` → `contextAlreadyDoneError` → `Canceled`
    ("timeout: context already done: context canceled"; `pgconn.Timeout(err)` is true);
  - canceled while connecting: `*pgconn.ConnectError` wrapping it ("failed to connect to `user=… database=…`: …:
    failed to receive message: context canceled"; a context canceled before the dial gives "dial error: … operation
    was canceled", still `errors.Is` `Canceled`).
- Consequence: `httpx.WriteError` tests `errors.Is(err, context.Canceled)` and whether the request context AccessLog
  saw is done (the recorder keeps it). If so it writes nothing and the access line is INFO, status 499,
  `error` "context canceled" (the literal, not the wrapped chain); a cancel with the request context still live stays
  ERROR 500. Without AccessLog (no recorder) it can't tell and stays 500. `TestCanceled` uses the `ConnectError`
  shape (no database needed).

### ServeMux: telling 404 from 405                   status: verified (2026-11-27, M1 T1)
- SPEC: §7.0 (user answer Q5); follow-up COR-5
- Library / version: `net/http` in Go 1.27.4 (`(*ServeMux).Handler`, `findHandler`, `matchingMethods`)
- Checked: the source and `TestMuxFallback`.
- Finding: `ServeMux.Handler(r)` returns the matched pattern, or `""` with the mux's own handler: a 405 handler
  (sets `Allow`, e.g. `POST` or `GET, HEAD`, then `http.Error` "Method Not Allowed"), `NotFoundHandler()`, or, for a
  path that needs cleaning (`/api//x`), a 307 redirect (a trailing-slash redirect returns a non-empty pattern). The
  handlers are unexported closures, so the pattern alone can't tell them apart; a catch-all pattern such as `/api/`
  would also swallow the 405s (it matches every method). `Handler` does not touch the request; `ServeHTTP` sets
  `r.Pattern`.
- Consequence: `httpx.MuxFallback(mux, "/api")` wraps the mux (no pattern registered, so the route list and the §3.3
  matrix don't see it). For a path at or under `/api` with an empty pattern it runs the mux's handler against a
  throw-away writer: 404 → the `NOT_FOUND` envelope; 405 → `METHOD_NOT_ALLOWED` with the mux's `Allow` header; any
  other status (the redirect) → the mux as before. Matched paths go to `mux.ServeHTTP` with the same request, so
  `r.Pattern` still reaches the access log. Paths outside `/api` keep the mux's plain text.

### sqlc: a uuid[] column with the uuid override     status: verified (2026-11-27, M1 T1)
- SPEC: §12.4; follow-up ARC-2
- Library / version: sqlc CLI v1.31.1, pgx/v5 v5.13.0
- Checked: `make generate` output for `GetActiveSession` and `TestRequireSessionQueries`.
- Finding: `coalesce(array_agg(f.id order by f.name_key), '{}')::uuid[]` generates `FarmIds []uuid.UUID` (the
  `db_type: uuid` override applies to the element type); pgx scans `{}` as an empty slice.
- Consequence: the session middleware gets the user's farms in the session query; `auth.principalFarms` keeps the
  owner's `FarmIDs` nil and never hands out nil for the others.

### textkey: x/text NFD and the cedilla / comma-below letters   status: verified (2026-11-27, M1 T4)
- SPEC: §10.2 (`textkey.Key`), §4.4 (D111 name keys)
- Library / version: `golang.org/x/text` v0.44.0 (`unicode/norm`), Go 1.27.4 `unicode`
- Checked: `go doc golang.org/x/text/unicode/norm`; `TestKey` / `TestKeyCollisions` in `internal/platform/textkey`.
- Finding: `norm.NFD` splits ş/ţ (U+015F/U+0163) into s/t + U+0327 (cedilla) and ș/ț (U+0219/U+021B) into s/t +
  U+0326 (comma below); both marks are in `unicode.Mn`, as are the breve and circumflex of ă/î/â. Dropping `Mn`
  and lower-casing gives "hirbovat" for "Hîrbovăț", "Hirbovat" and "HÎRBOVAŢ".
- Consequence: `textkey.Key` = NFD, drop `Mn`, lower-case, drop whitespace and the stand-alone unit words "g" and
  "%". The migration fills every `name_key` with it (farms, guides, feed types, series).

### Extracts vs the §10.3 rules and the table constraints   status: verified (2026-11-27, M1 T4)
- SPEC: §10.3
- Library / version: design/data/extracts (DATA06), `migratehistory.BuildPlan`
- Checked: `go test -race ./internal/migratehistory -run 'TestPlan|TestBuildPlanRejects'`; the plan's own
  constraint check (`plan_check.go`) mirrors the 0001 checks and unique indexes before any COPY.
- Finding: everything §10.3 lists reproduces (goldens 5, 7, 9–11, 16, 26, 27, 35, 36 from the plan alone), with
  two extract facts the rules didn't name: (1) Telița also has two rows with g and 0 kg (H5-2 ♂ 22–23.02.2026), so
  the Q6 rule skips 16 rows, not 14; (2) `eggs-hirbovat-prev-daily.json` has three negative ♂ cells (H1.2 = H5.2
  10.09.2025 −75 and 09.11.2025 −6, H4.2 = H7.2 09.11.2025 −10) that `legacy_mortality.deaths >= 0` refuses.
- Consequence: (1) skipped like the Q6 rows (user answer 27.11.2026), SPEC §10.3 names them; (2) skipped (user
  answer Q-T4-1), SPEC §10.3 names them.

### go-cmp, unexported fields and the golden value types   status: verified (2026-11-27, M1 T6)
- SPEC: §11.7, §11.8 (`go-cmp` for diffs)
- Library / version: `github.com/google/go-cmp` v0.7.0; Go 1.27.4 `reflect`, `errors.AsType`
- Checked: `go doc github.com/google/go-cmp/cmp.Equal`; `TestCheck`, `TestGoldenRowsOnHandBuiltLedger` in
  `internal/golden`, and a mutation run (banker's rounding in place of `num.Round` fails goldens 12 and 26 on the
  hand-built ledger: 51,25 → "51.2", 1,05 → "1.0").
- Finding: `cmp.Equal`/`cmp.Diff` panic on a struct with unexported fields unless it has an `Equal(T) bool` method
  or an option allows it; `timex.Day` has unexported fields and no `Equal`, and `decimal.Decimal` compares by
  representation (2303.2 vs 2303.20) under `reflect.DeepEqual`.
- Consequence: the golden rows return plain values only (ints, `calc.PerSex`, days as "YYYY-MM-DD", decimals
  rounded once with `num.Round` and written with `StringFixed` at the §11.7 precision). `Row.Check`
  (`reflect.DeepEqual`, used by `migratehistory.Verify`) and `cmp.Diff` (TestGolden's message) then agree, and the
  precision is part of the expected value (Cocoși "1781.0": 1 781,04 passes, 1 781,1 fails).

### pgx CopyFrom vs sqlc :copyfrom                   status: verified (2026-11-27, M1 T5)
- SPEC: §10.3 ("bulk rows with pgx `CopyFrom` / sqlc `:copyfrom`"), §12.4
- Library / version: sqlc v1.31.1 (`sql_package: pgx/v5`), pgx v5.13.0 (`copy_from.go`, `values.go`)
- Checked: `make generate` output for `db/queries/migratehistory.sql`; pgx source; `TestLoad` (16 tables, ~13 500 rows,
  about 0.5 s with Verify).
- Finding: `:copyfrom` generates `internal/db/dbq/copyfrom.go` (one `Copy…(ctx, []Copy…Params)` per query, an
  iterator over the params) and adds `CopyFrom` to the generated `DBTX` interface; `*pgxpool.Pool` and `pgx.Tx`
  satisfy it. The column list is the INSERT's, so `egg_production.total_eggs` (generated) is simply not listed.
  pgx's `CopyFrom` always uses the binary format; for a column type it doesn't know (our enums) it encodes a
  string-kind value's text bytes, which is the enum's binary form, so `dbq.Sex` etc. need no registration.
  `timex.Day` (DateValuer) and `decimal.NullDecimal` (pgx-shopspring-decimal, registered by `db.PoolConfig`) encode
  as date and numeric. COPY fires the row triggers (`cycle_guide_type_trg`) and checks every constraint.
- Consequence: the migration writes through sqlc `:copyfrom` (compile-time column lists, §10.3 point 3), not
  hand-written `pgx.CopyFrom` calls. Callers of `dbq.New` must pass a pool, a connection or a transaction (all
  implement `CopyFrom`).

### The dev seed vs golden 10                        status: verified (2026-11-27, M1 T7)
- SPEC: §11.9, §11.7 row 10, §5.5; plan M1-D7
- Library / version: `illustrative-events-hirbovat.json` (RF08), `migratehistory.Load` with `SeedDev`
- Checked: a load with every D7 item, including the sexing error carved out of H6.2 ♀ 19.05.2026 deaths (3 → 2 + −1).
- Finding: Verify then fails only golden 10: cycle deaths ♀ 1 520, 5,65 % (want 1 521, 5,66 %). §5.5 deaths are mortality +
  culls; a sexing error is neither, so carving one out of the deaths lowers them. The 5 culls keep the deaths.
- Consequence: the sexing error is not seeded (user answer Q-T7-1, 27.11.2026); everything else in D7 is.

### testdb's seeded template vs Go's test import cycles   status: verified (2026-11-27, M1 T8)
- SPEC: §11.8 (`il_seed` template), plan M1-D5/D6 ("testdb → migratehistory → golden → calc → db")
- Library / version: Go 1.27.4 `go vet`/`go test`; PostgreSQL 16.16 `CREATE DATABASE … TEMPLATE`
- Checked: a probe file in `internal/testdb` importing `migratehistory`, then `go vet ./internal/auth ./internal/calc`;
  `go test -race ./internal/testdb ./internal/golden`.
- Finding: `internal/auth/auth_test.go` (`package auth`) and `internal/calc/readers_test.go` (`package calc`) import
  testdb, and migratehistory imports auth (seeddev.go) and calc (verify.go): "import cycle not allowed in test". So
  testdb can't import migratehistory. Building `il_seed` (copy of the migrated template + Load with the dev seed +
  Verify) takes about 2 s under `-race`; each `NewSeeded` copy about 0.1–0.3 s.
- Consequence: testdb takes the loader as a `testdb.Seeder` through `testdb.MainSeeded(m, seeder)`; the small package
  `internal/testdb/seed` supplies it (`seed.Main(m)`, `seed.Load`). A package that calls `testdb.NewSeeded(t)` uses
  `func TestMain(m *testing.M) { seed.Main(m) }`; packages on `testdb.Main` are unchanged and never pay for the seed.
  The seeded template is `il_t_<pid>_<random>_seed`, built lazily once per process (`sync.OnceValue`), dropped at exit.

## M2

### pgx.RowToStructByName and the `db` tags            status: verified (2026-11-27, M2 T8)
- SPEC: §12.4 (hand-written SQL scanned with `pgx.CollectRows(…, pgx.RowToStructByName[T])`); plan M2-D1
- Library / version: `github.com/jackc/pgx/v5` v5.13.0 (`rows.go`: `RowToStructByName`, `namedStructRowScanner`,
  `lookupNamedStructFields`, `computeNamedStructFields`, `fieldPosByName`)
- Checked: the source, and a throw-away probe (`pgx.CollectRows` on PostgreSQL 16.16 through a testdb pool with the
  registered decimal and `timex.Day` codecs), removed after the run.
- Finding:
  - The match is strict both ways: a struct field without a column fails with `cannot find field b in returned row`,
    a column without a field with `struct doesn't have corresponding row field c` (the `…Lax` variants allow missing
    columns, never extra ones). A `db:"name"` tag matches the column exactly first, then case-insensitively; an
    untagged field matches case-insensitively with `_` ignored (`FooBar` ← `foo_bar`); `db:"-"` skips the field;
    unexported fields are skipped; embedded structs are flattened.
  - The field-to-column mapping is cached per struct type and column list (`sync.Map`), so the reflection cost is
    paid once per query shape.
  - Nullable columns: NULL into `int` fails (`cannot scan NULL into *int`), into `decimal.Decimal` fails (`use
    NullDecimal or *Decimal`), into `timex.Day` fails (`timex: cannot scan NULL into Day (use *Day)`). NULL into
    `*int`, `*timex.Day` gives nil, into `decimal.NullDecimal` `Valid: false`, into `[]string` a nil slice (`'{}'` an
    empty one). `bigint` (`sum(int)`) and an integral `numeric` scan into `int`.
- Consequence: `periodtable.dayRow` tags every column; every aggregate is `coalesce(…, 0)::integer|numeric` and the
  chips come from `array(…)` (never NULL), so the struct has no nullable field. A later query with a nullable
  column scans it into a pointer (or `decimal.NullDecimal`), never a value type. The tests that run `ReadDays` /
  `Build` fail on any drift between `daysSQL` and `dayRow`.

### encoding/json `omitzero` on maps, struct pointers and nil map values   status: verified (2026-11-27, M2 T8)
- SPEC: §7.0a (`null` vs omitted), §7.4 (cells `null` = "–"; `more`, `start` always written)
- Library / version: Go 1.27.4 `encoding/json` (the v2-backed v1 API, see M0 above)
- Checked: a throw-away probe marshalling the field kinds below; `TestNullLists`, `TestFarmReadsOnSeededData`
  (every M2 body on the seeded data through `stack.do`'s null walker).
- Finding:
  - `omitzero` omits a **nil** map, slice or pointer; a non-nil empty map is written `{}`, an empty slice `[]`, a
    pointer to a zero struct `{"before":0,…}`. `omitzero` on a `bool` omits `false`.
  - A pointer field without an omit option writes `null` when nil (`"more": null`).
  - `map[string]*num.Num` writes a nil value as `null` (`{"a":null,"b":1.5}`); the map itself, nil and without
    `omitzero`, is written `null`.
- Consequence: `Cells` has no omit option and the builders always make it (`Cells{}`), so a row's cells are an
  object and a missing figure is `"key": null`. `Status`/`Reasons` maps are `omitzero` and the builder leaves them
  nil when empty (an empty non-nil one would be written `{}`). `Loaded`, `Totals`, `Empty` (pointers) and `NoGuide`
  are left out when unset; `More` and `Start` have no omit option, so they are written `null` until they apply. The
  cmd/api contract walker (`nullableFields`) lists every nullable path of the M2 bodies; `"weeks[].cells.*"` allows
  a null under any cell key but not `"cells": null` itself.

### Hen-days on an emptying day (M2 Q1 (b))            status: verified (2026-11-27, M2 T8)
- SPEC: §5.6 (`hen_days`, amended in M2), §11.7 goldens 20 and 24
- Library / version: `internal/periodtable/sql.go` (`daysSQL`, CTE `unit_day`), PostgreSQL 16.16
- Checked: `daysSQL`; the golden figures in the plan (planner's Python on the extracts) and the periodtable tests.
- Finding: the rule lives once, in `unit_day`: `case when egg_rows > 0 or alive_end_f = 0 then alive_start_f else 0
  end as hen_days`, per egg unit and day. H7 on 14.11.2026 (6 915 hens at the start, sold out, no egg row) counts,
  so SAPT 58 is 156 418 hen-days → 30,4 % and the foot SAPT 56–59 545 038 → 41,6 % (§5.6 as first written gave
  31,8 % and 42,1 %). A unit with no hens contributes 0 either way. Weeks without any egg row (SAPT 21–22) keep a
  null lay % (Σ eggs has no entry), and the hen-day guide weighting (59,9 over 56–59) uses the same hen-days.
  Egg parts and egg weight still count only unit-days with a total.
- Consequence: `calc` rollups take the hen-days from the row as summed by SQL; `WeekSeries` and the summary readout
  use the same rows, so goldens 20, 23 and 24 share one definition.

### Where the JSON golden files live                    status: verified (2026-11-27, M2 review)
- SPEC: §11.8 (JSON golden files), plan M2 T9
- Library / version: Go 1.27.4 test runner (one `TestMain` per test binary); `internal/testdb/seed`
- Checked: `go vet` with a probe import of `seed` from an in-package `periodtable` test; `go test ./internal/golden
  -run TestJSONGolden` (with and without `-update`).
- Finding: the files need the seeded data, so the test binary that compares them needs `TestMain` = `seed.Main`.
  `seed` imports `migratehistory` → `golden` → `farms` → `periodtable`, so an in-package test of `farms` or
  `periodtable` importing it is "import cycle not allowed in test"; only an external `_test` package can. A
  package's in-package and external test files build into one binary with one `TestMain`, so moving `TestMain`
  to the external file (as `farms` and `periodtable` now do for their seeded tests) keeps `testdb.New` working for
  the in-package tests.
- Consequence: the JSON golden files are `internal/golden/testdata/periodtable/*.json` and
  `internal/golden/testdata/farms/*.json`, compared by `internal/golden`'s external tests (`json_test.go`), which
  already run under `seed.Main` and hold the id-token normaliser (M2 D9) every file shares. `-update` rewrites
  them. §11.8 names that location.

### cmd/api on the seeded data                         status: verified (2026-11-27, M2 T8)
- SPEC: §11.8 (contract tests), §11.6 (performance targets)
- Library / version: `internal/testdb/seed`, Go 1.27.4 test runner
- Checked: `go vet ./cmd/api`; `go test -race ./cmd/api`.
- Finding: `cmd/api` is `package main`, which nothing imports, so its in-package tests can use `seed.Main` without
  the import cycle that keeps other packages' seeded tests external (M1 above). The seeded template is built lazily,
  only by the tests that call `testdb.NewSeeded` (`TestFarmReadsOnSeededData`, `BenchmarkFarmReads`).
- Consequence: `cmd/api`'s TestMain is `seed.Main(m)`; every other test there keeps `testdb.New`.

### M2 read latency vs §11.6                           status: verified (2026-11-27, M2 T8)
- SPEC: §11.6 (p95: `GET /api/table` page ≤ 300 ms, load-more ≤ 200 ms, summary/home ≤ 400 ms)
- Library / version: `BenchmarkFarmReads` (cmd/api), PostgreSQL 16.16 in testcontainers (Docker Desktop), seeded data
- Checked: `go test -run '^$' -bench FarmReads -benchtime 200x ./cmd/api` (no `-race`), Apple M5 Max, the owner,
  200 requests each over a socket through the whole chain (session check, service, SQL, JSON), 18.11.2026.
- Finding (p50 / p95 per request):
  | read | p50 | p95 | §11.6 target |
  |---|---|---|---|
  | `GET /api/table` Hîrbovăț page 1 (SAPT 56–59, grain day) | 12.9 ms | 15.1 ms | ≤ 300 ms |
  | load-more `before=56` (SAPT 52–55, totals over 52–59) | 14.7 ms | 16.0 ms | ≤ 200 ms |
  | whole cycle, `grain=week&count=40` (SAPT 20–59) | 27.5 ms | 29.9 ms | – |
  | `GET /api/farms/{id}/page` Hîrbovăț | 7.5 ms | 8.0 ms | – |
  | `GET /api/farms/{id}/summary` Hîrbovăț (laying, readout series over 40 weeks) | 65.0 ms | 73.1 ms | ≤ 400 ms |
  | `GET /api/farms/{id}/summary` Telița (growth) | 17.7 ms | 22.1 ms | ≤ 400 ms |
  | `GET /api/bootstrap` | 1.1 ms | 1.8 ms | – |
- Consequence: every M2 read is well inside §11.6 on the seeded data; the laying summary (its many statements and
  the 40-week `WeekSeries`) is the slowest at about a fifth of its budget. §11.6's table target names detail
  "Compartimente", which arrives in M6: rerun the benchmark then (and on production-sized data, ≤ 5 years). Under
  `-race` the times are several times higher and are not a gate.
