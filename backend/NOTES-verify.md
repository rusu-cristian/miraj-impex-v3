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

## M3

### pg_notify inside a transaction                     status: verified (2026-11-27, M3 T3)
- SPEC: §4.11, §9, §12.5 (M3 D1)
- Library / version: PostgreSQL 16.16 (testcontainers), pgx v5.13.0 `(*pgx.Conn).WaitForNotification`
- Checked: `internal/live` `TestEmitCommit` / `TestEmitRollback`: a separate connection runs `LISTEN live_event`;
  `live.Emit` in an open il_app transaction; `WaitForNotification` with a 300 ms timeout before the commit, 5 s after.
  pgx source (`pgconn.peekMessage`): a timeout ends the wait without closing the connection, so the same listener
  waits again.
- Finding: no notification (and no visible row) before the commit; one notification whose payload is the new
  `live_event.id` right after it; none after a rollback. `insert … returning id` and `pg_notify` run as one statement
  (`EmitLiveEvent`: a CTE with `cross join lateral pg_notify(…)`), which sqlc v1.31.1 accepts (`:one`, `int64`).
- Consequence: `live.Emit` writes one statement per event; the M11 Listener can rely on "notified ⇒ committed".

### A 23505 inside pgx.BeginTxFunc                     status: verified (2026-11-27, M3 T3)
- SPEC: §12.5 (M3 D2)
- Library / version: PostgreSQL 16.16, pgx v5.13.0, `internal/db.InTx`
- Checked: psql on PostgreSQL 16.16 (a duplicate insert, then `select 1` in the same transaction);
  `internal/entries` `TestCreateOnConflict` and `TestCreateKeyConstraint`.
- Finding: after a unique violation PostgreSQL answers every further statement of the transaction with
  `25P02 current transaction is aborted`, so no lookup can run inside it. `BeginTxFunc` rolls back and returns the
  callback's error unwrapped (`*pgconn.PgError`); `InTx` turns it into `*pgerr.ConstraintError` (`Constraint` = the
  index name, e.g. `entry_client_request_uq`, `farm_name_uq`), still `errors.As`-able to `*pgconn.PgError`.
- Consequence: `entries.Runner.Create` maps the key constraints (`entry_client_request_uq`, `request_replay_pkey`)
  and the form's own (`Create.OnConflict`, e.g. `feed_allocation_uq`) after `InTx` returns, on the pool.

### hashtextextended for the farm locks                status: verified (2026-11-27, M3 T3)
- SPEC: §12.1, §12.5 (`LockFarm`)
- Library / version: PostgreSQL 16.16
- Checked: `select hashtextextended(…, 0)` in psql and in `internal/ledger` `TestLockFarms` (asserts the two test
  ids differ); `pg_locks` for the advisory locks held.
- Finding: `hashtextextended(text, bigint) → bigint` exists in PostgreSQL 16; `LockFarms(b, a, b)` holds exactly two
  advisory locks, taken a then b; a second transaction locking `a` waits until the first commits.
- Consequence: the 64-bit key per farm id stays. Two farm ids whose hashes collide would only serialise more than
  needed (never a missed lock).

### request_replay.body and entry.created_at round trips   status: verified (2026-11-27, M3 T3)
- SPEC: §4.11, §7.0, §12.5 (M3 D3)
- Library / version: sqlc v1.31.1 (pgx/v5), pgx v5.13.0
- Checked: the generated `dbq` types; `internal/entries` `TestCreateSaves`, `TestCreateReplay`,
  `TestCreateConcurrentSameKey`.
- Finding: `bytea` maps to `[]byte` and comes back byte-exact (the stored body equals the 201 body, trailing newline
  of `httpx.Marshal` included, and a replay one minute later on the clock is the same bytes). `entry.created_at`
  from the frozen clock reads back equal; the version string is `2026-11-18T10:00:00.000000Z`, the same from the
  row's `coalesce(updated_at, created_at)`. The new queries' types: `uuid[]` → `[]uuid.UUID`, `jsonb` params →
  `[]byte` (nil = SQL null), `client_request_id` → `*uuid.UUID` in `InsertEntry`, `uuid.UUID` where cast `::uuid`,
  nullable `entry_kind` → `*dbq.EntryKind`, `status smallint` → `int16`.
- Consequence: the replay stores `httpx.Marshal` output and `WriteRaw` sends it unchanged. `entries.version`
  formats the version (fixed width, microseconds, UTC), as the SPEC §12.5 sketch and its "Version" note name it.

### numeric(6,1) rounds g silently (Q9)               status: verified (2026-11-27, M3 T4)
- SPEC: §5.15, §6.3 H1; M3 Q9
- Library / version: PostgreSQL 16.16 (`postgres:16.16-bookworm`)
- Checked: `insert into t values (160.25), (160.35), (300.04)` into a `numeric(6,1) check (g between 1 and 300)`
  column, then `300.05`.
- Finding: the cast rounds half away from zero before the check: 160.25 → 160.3, 160.35 → 160.4, 300.04 → 300.0
  (accepted), 300.05 → 300.1 (refused by the check). Nothing tells the client its value changed.
- Consequence: `FeedDailyCreateRequest.Validate` refuses a g with more than one decimal (`rows[i].g`, `out_of_range`,
  "Datele trimise nu sunt valide.") and checks 1–300 on the value as sent, so the stored g always equals the typed
  one (`TestFeedDailyValidate`).

### Embedded WriteMeta in a create body under encoding/json/v2   status: verified (2026-11-27, M3 T4)
- SPEC: §7.0a (WriteMeta "fields are promoted in JSON"), §7.8
- Library / version: Go 1.27.4 `encoding/json/v2` with `json.DefaultOptionsV1()` (`httpx.Decode`)
- Checked: `TestFeedDailyHandler` and `TestCreateFeedDailySeeded` decode `{"farmId", "date", "clientRequestId",
  "rows"}` into `FeedDailyCreateRequest` (which embeds `httpx.WriteMeta` without a tag) and replay on the key;
  `{"rows":[{"grams":160}]}` gives `unknown_field` on `rows[0].grams`.
- Finding: an untagged embedded struct is inlined (its members are top-level JSON members), and
  `RejectUnknownMembers` still applies inside slices of structs with the full path.
- Consequence: every create request embeds `httpx.WriteMeta` as §7.8 shows; no `inline` tag is needed.

### Stock reader plan on the seeded data (D7 shape)       status: verified (2026-11-27, M3 T1; recorded in T8)
- SPEC: §5.8, §11.6; plan M3 D7, ARC-6
- Library / version: PostgreSQL 16.16 (testcontainers), seeded template (history import + dev seed)
- Checked (T1): `EXPLAIN ANALYZE` of `CalcStockEnd` (db/queries/calc.sql) for Hîrbovăț on 18.11.2026; the timings of
  `calc.ReadStockDays` (stock + window + consumption + days left) and `calc.StockSeries` over 01.01–18.11.2026.
- Finding: the D7 shape (Σ deliveries − Σ allocation kg + Σ adjustments with an explicit day; alive_start computed
  once per allocation as a window over the leaf deltas, not a per-row lateral sum; no `v_feed_stock`,
  `v_feed_consumption`, `v_section_sex_day`, no `app_today()`) runs over the farm's whole history in about 28.8 ms
  execution for `CalcStockEnd`; `ReadStockDays` takes about 32–40 ms; `StockSeries` 01.01–18.11 about 14 ms.
  `grep -nE "app_today|v_section_sex_day|v_farm_sex_day|v_egg_day|v_feed_consumption|v_feed_stock"` over the M3
  query files (`entries`, `forms`, `ledger`, `audit`, `live`, `replay`, `reports`) and `periodtable/feedsql.go`
  prints nothing.
- Consequence: H5 (the lowest stock from the save's date to today) and the stock rows of the form state, the summary
  and `reports/feed` read the whole history on every call and stay well inside §11.6 on the seeded data (latency
  below). A per-type `from` bound is the next step if production data makes them slow.

### Consum (view=report-feed) on the seeded data          status: verified (2026-11-27, M3 T7; recorded in T8)
- SPEC: §7.4 report-feed, §7.12, §5.7, §11.6; plan M3 D11, Q2, Q8
- Library / version: PostgreSQL 16.16, pgx v5.13.0 `QueryTracer`
- Checked: `internal/periodtable` `TestConsumStatements` (statement count per page), `TestConsumSeeded` (the board
  figures of SAPT 56–59), `BenchmarkFarmReads/consum-*` (below).
- Finding: a page is the cycle read, one Consum day statement (`feedsql.go`: per farm-day and per leaf section ×
  day the kg ♀ · ♂, the bird-days on fed days, the birds at the start and the missing keys), the guide weeks (only with a guide) and the settings: at most 4 statements
  (`CalcFarmCycles`, `PeriodTableFeedDays`, `CalcGuideWeeks`, `CalcSettings`). The farm day g is the register's g (Σ kg ÷ Σ kg/g, user answer to T7's question, §5.7), week,
  section and period g are kg ÷ bird-days. "lipsă" is computed on every day of the period, today's only from
  `missing.from_hour`; the D163 line is derived from the loaded days only, so it appears on the pages whose weeks
  hold those days (SAPT 56–57 for Hîrbovăț).
- Consequence: Consum costs about what the Rezumat page costs (below). `notes[]` varies with the loaded weeks, so the
  client shows the notes of the latest response.

### M3 latency vs §11.6                                   status: verified (2026-11-27, M3 T8)
- SPEC: §11.6 (p95: table page ≤ 300 ms, load-more ≤ 200 ms, summary/home ≤ 400 ms, any save ≤ 300 ms)
- Library / version: `BenchmarkFarmReads` (cmd/api), PostgreSQL 16.16 in testcontainers (Docker Desktop), seeded data
- Checked: `go test -run '^$' -bench FarmReads -benchtime 200x ./cmd/api` (no `-race`), Apple M5 Max, the owner,
  200 requests each over a socket through the whole chain, 18.11.2026. The save is a real save of Hîrbovăț's 8 rows
  of 18.11 (§7.8's example); before each request, outside the timed part, the day's live entry is soft-deleted (so
  soft-deleted rows pile up over the run, a slight pessimisation). The dry run is the same save rolled back.
  Other test binaries were running on the machine at the same time, so the p95s are noisy.
- Finding (p50 / p95 per request):
  | request | p50 | p95 | §11.6 target |
  |---|---|---|---|
  | `GET /api/forms/feed-daily` Hîrbovăț 18.11 (8 rows, stock) | 62.7 ms | 75.6 ms | – |
  | `POST /api/entries/feed-daily` 8 rows (lock, H2–H5, inserts, audit, live, replay) | 32.5 ms | 35.8 ms | ≤ 300 ms |
  | the same as a dry run (rolled back) | 31.3 ms | 37.5 ms | ≤ 300 ms |
  | `GET /api/reports/feed` owner, every farm | 40.4 ms | 61.5 ms | – |
  | `GET /api/table?view=report-feed` Hîrbovăț page 1 (SAPT 56–59) | 34.1 ms | 56.8 ms | ≤ 300 ms |
  | Consum load-more `before=56` | 34.5 ms | 53.1 ms | ≤ 200 ms |
  | `GET /api/farms/{id}/summary` Hîrbovăț (now with `feedDaysLeft`) | 105.0 ms | 130.4 ms | ≤ 400 ms |
  | `GET /api/table?view=farm` page 1 (M2, for comparison) | 20.6 ms | 32.3 ms | ≤ 300 ms |
- Consequence: every M3 route is well inside §11.6 on the seeded data. The laying summary grew from 65 ms (M2) to
  about 105 ms p50 with the stock and days-left reads; the form state is the heaviest M3 read (the section × sex
  rows plus the stock over the whole history). Rerun on production-sized data (≤ 5 years) before M6.

## M4

### `json.RawMessage` request fields: null vs absent        status: verified (2026-11-28, M4 T5)
- SPEC: §6.7 (a manager's request *containing* price fields → 403 `PRICE_FORBIDDEN`), M4 D5, Q9
- Library / version: Go 1.27.4 `encoding/json/v2` with `json.DefaultOptionsV1()` (the `httpx.Decode` options)
- Checked: a throwaway test decoding `{"b":1}`, `{"a":null}`, `{"a":0}`, `{"a":"x"}`, `{"a":{…}}`, `{"a":[1]}` into a
  struct with `A json.RawMessage` through `httpx.Decode`; kept as `entries.TestCreateDeliveryPriceFields` (manager
  with `"pricePerKg": null` → 403, owner with both null → 400 on both paths).
- Finding: an absent member leaves the field nil; any present member, `null` included, gives its raw bytes (`null` →
  `"null"`, length 4). No value is refused at decode time, so a price field is never an unknown-field 400 for a manager.
- Consequence: `DeliveryCreateRequest.PricePerKg` / `TotalLei` are `json.RawMessage`; "present" is `!= nil`.

### numeric(12,4) / numeric(14,2) on a price update        status: verified (2026-11-28, M4 T4)
- SPEC: §6.8, §7.13 (M4 D3: level-1 bounds before the update)
- Library / version: PostgreSQL 16.16 (`postgres:16.16-bookworm`)
- Checked: a throwaway container: `insert` of `8.12345` into `numeric(12,4)`, `8000.125` into `numeric(14,2)`, `10^8`
  and `10^12` into each; `prices.TestSetValidation` (5 / 3 decimals, a value at 10^8 → 400 before any write).
- Finding: the column's typmod rounds **silently** (half away from zero): 8.12345 → 8.1235, 8000.125 → 8000.13. A value
  that rounds to 10^8 / 10^12 or more fails with `22003 numeric field overflow` ("A field with precision 12, scale 4
  must round to an absolute value less than 10^8").
- Consequence: `calc.DerivePrice` refuses more than 4 / 2 decimals and values outside the columns (typed or derived),
  so `PricesSet` / `PricesChange` never round or overflow; a `22003` reaching `WriteError` would be a bug (500).

### shopspring/decimal division and rounding for price derivation   status: verified (2026-11-28, M4 T2)
- SPEC: §6.8 (`price = total ÷ kg`, 4 decimals; `total = round(kg × price, 2)`), §5.15 (half away from zero)
- Library / version: `github.com/shopspring/decimal` v1.6.0
- Checked: `go doc decimal.Decimal.DivRound`, `Decimal.Round`, `Decimal.Truncate`; `TestDerivePrice` half cases.
- Finding: `DivRound(d2, 4)` divides exactly (quotient + remainder) and rounds the 5th decimal half away from zero,
  so it never double-rounds through `DivisionPrecision` (16 digits) the way `Div(…).Round(4)` could: 0,50 ÷ 10 000 →
  0,0001; 2 ÷ 3 → 0,6667; 199 000 ÷ 25 640 → 7,7613. `Mul` is exact, then `Round(2)` (half away from zero): 0,5 × 0,01
  → 0,01; 25 640 × 7,74 → 198 453,60. `v.Equal(v.Truncate(n))` tells "more than n decimals" apart from trailing
  zeros ("7.740000" has 2).
- Consequence: `calc.DerivePrice` uses `DivRound` for the price per kg and `Mul(…).Round(2)` for the total; a typed or
  derived value that rounds to 0 or reaches 10^8 lei/kg (`numeric(12,4)`) / 10^12 lei (`numeric(14,2)`) is
  `ErrPriceRange`, so the update never meets the column's overflow or `> 0` check.

### priced_at = audit_log.at after a round trip            status: verified (2026-11-28, M4 T4)
- SPEC: §7.13 price undo (M4 D3, D4: the undo matches "still the prices of this save" on the instant)
- Library / version: pgx/v5 v5.13.0, PostgreSQL 16.16
- Checked: the throwaway container above (`'… 12:00:00.123456789+02'` stored as `.123457`: PostgreSQL **rounds** to
  microseconds); `prices.TestSet` with a clock at 12:00:00.123456789: the two rows' `priced_at` equal the
  `price_set` row's `at` in SQL and the undo of that batch succeeds.
- Finding: both columns are written from the same Go `time.Time` in one transaction, so they are equal either way; the
  service truncates the clock to microseconds first (`Service.now`) so the Go value is also the stored one.
- Consequence: `PricesUndo` matches `priced_at = audit.at and priced_by = caller`; the Go check uses `time.Equal`.

### Price undo token expiry across DST                     status: verified (2026-11-28, M4 T4)
- SPEC: §8.5, §7.13 (valid until 23:59 of the save day), M4 D4
- Library / version: Go 1.27.4 `time.Date` with the embedded `Europe/Chisinau` zone (`timex.Chisinau`)
- Checked: `prices.TestUndoExpiry`: saves at 18.11 00:00 / 12:00 / 23:59:59, at 25.10.2026 01:30 +03:00 (the day
  clocks go back), at 25.10 12:00 +02:00 and at 29.03.2026 01:00 (clocks go forward).
- Finding: `timex.DayOf(now).AddDays(1).Start()` is 00:00 local of the next day in every case; 25.10.2026 lasts 25 h.
  `tokens.Verify` accepts while `now < exp`, so 23:59:59 is valid and 00:00 is `ErrExpired`.
- Consequence: `prices.undoExpiry`; an expired token is `409 STALE` "Nu se mai poate anula: ziua salvării s-a încheiat."
  (tested with the clock at 19.11 00:00).

### audit_log.id through sqlc `returning id`              status: verified (2026-11-28, M4 T1)
- SPEC: §4.10, §8.5 (the price undo token names its audit row, M4 D4)
- Library / version: sqlc v1.31.1, pgx/v5 v5.13.0
- Checked: the generated `dbq.InsertAuditLogID` (`:one … returning id`); `internal/audit` `TestWriteID` writes rows as
  `il_app` and reads them back by the returned id.
- Finding: `bigserial` is generated as `int64`; the id is the row's, later rows get larger ids, and `il_app` needs no
  sequence grant beyond `0002_roles.sql`'s (the insert already worked with the default). The generated
  `InsertAuditLogIDParams` has the same fields as `InsertAuditLogParams`, so `audit.WriteID` converts one into the other.
- Consequence: `audit.WriteID(ctx, q, Row) (int64, error)`; `audit.Write` keeps `InsertAuditLog` (so the M3 save
  tracers keep their statement names).

### select … for update on feed_delivery under il_app       status: verified (2026-11-28, M4 T4)
- SPEC: §12.1 (farm advisory locks), M4 D3
- Library / version: PostgreSQL 16.16, `0002_roles.sql` grants
- Checked: every `internal/prices` test runs as `il_app` (`testdb`); `TestSetConcurrent` sends two batches on one
  delivery through a barrier; `TestSetStatements` records the order.
- Finding: `select … for update of fd` needs `UPDATE` on `feed_delivery`, which `il_app` has; `of fd` keeps the lock off
  the joined `entry` and `feed_type` rows. The second batch waits on the farm advisory lock, then reads the row priced
  → `409 STALE`; exactly one 200, no deadlock, no 500. A batch is 9 statements: begin, `PricesDeliveryFarms`,
  `LockFarm` per farm, `PricesLockDeliveries`, one `PricesSet` for every row, `InsertAuditLogID`, `EmitLiveEvent`,
  commit (plus `PricesDeliveryRefs` before the transaction for the level-1 messages).
- Consequence: no migration; the farm lock comes first so the order is the same as every other save's.

### Series placement of migrated history                status: verified (2026-11-28, M4 T2)
- SPEC: §5.9 (series total ÷ series placement ♀ + ♂)
- Library / version: seeded test DB (`seed.Main`, extracts to 18.11.2026)
- Checked: `calc.SeriesPlacement`, `calc.ReadDeliveries`, `calc.EggsBetween` and `calc.FarmBase` on the seed (a
  throwaway seeded test, not kept; T8's goldens 29 and 31 keep the figures).
- Finding: the migration puts Telița's 07.10.2025 placement in its **migration** entry (`entry.series_id` null), so
  "Σ placement rows of the series' placement entries" reads 0. Linking by cycle (placement rows dated inside a cycle
  of the series at their farm; an app placement entry counts only for its own series) gives Seria 2026 ♀ 28 080 ·
  ♂ 3 510 = 31 590. On the same seed: Hîrbovăț's Seria 2026 range (29.01.2026–18.11.2026) has 69 deliveries,
  ♀ 1 114 110 · ♂ 78 530 kg, 65 priced 1 174 480 kg · 9 562 437 lei, 4 unpriced 18 160 kg (17.02 Breeder 2 2 000,
  17.02 Cocoși 1 000, 14.03 Cocoși 6 160, 12.11 Breeder 2 9 000); Telița (…–28.02.2026) 13 deliveries, 263 740 kg
  (type "ambele sexe"), all priced, 2 031 768 lei; eggs Hîrbovăț 5 073 155; `FarmBase` Telița 31 590 (28 080 · 3 510),
  Hîrbovăț 29 907 (26 883 · 3 024).
- Consequence: `CalcSeriesPlacement` joins `cycle` → `bird_event` (type `placement`) by farm and date. Golden 29 and
  the golden 31 inputs reproduce.

### calc.FarmBase for the Seria 2026 cycles              status: verified (2026-11-28, M4 T2; rechecked in T7)
- SPEC: §5.9 (lei/pasăre ÷ the birds at the cycle start = the §5.5 farm base), golden 31
- Library / version: seeded test DB (`seed.Main`), PostgreSQL 16.16
- Checked: T2's throwaway seeded test (entry above); in T7 a throwaway `cmd/api` test calling `calc.FarmBase` for
  Telița from 07.10.2025 and Hîrbovăț from 22.02.2026 to 18.11.2026; `TestDeliveryPriceChainSeeded`'s Cost lines.
- Finding: Telița ♀ 28 080 · ♂ 3 510 = 31 590 (the A11 register adjustments of its closing are not inflows);
  Hîrbovăț ♀ 26 883 · ♂ 3 024 = 29 907 (the transfers from Telița, 22–28.02). The Cost rows over HTTP give
  64,32 and 319,74 lei/pasăre, golden 31.
- Consequence: `reports` divides by `FarmBase` (batched in `ReportsCostInputs`); no change.

### calc.ReadStockDays on a farm without birds            status: verified (2026-11-28, M4 T7)
- SPEC: §5.8, §7.8 delivery form state (M4 Q14)
- Library / version: seeded test DB, PostgreSQL 16.16
- Checked: a throwaway `cmd/api` test calling `calc.ReadStockDays(Telița, 18.11.2026)` with the seeded settings;
  `entries.TestBuildDeliveryFormSeeded` and `TestFarmReadsOnSeededData` (`GET /api/forms/delivery` for Telița, 200).
- Finding: no error. With no section holding birds the window ends today (16–18.11) and every type's average is 0:
  Grower and Starter `days: null`, lamp oprit (`no_consumption`); Pre-ouat (below 0) `days 0`, run-out 18.11, lamp
  abatere (`stock_negative`). The delivery form never calls it for Telița (no open cycle → `stock: []`, `cycle: null`).
- Consequence: no guard needed; the form keeps `stock: []` for a farm without an open cycle (§5.8).

### Telița's computed stock now (the workbook's −716 kg)   status: verified (2026-11-28, M4 T7)
- SPEC: §5.8 (an empty farm's leftover is never shown), §6.7 (a delivery Toast shows the stock at the end of its date)
- Library / version: seeded test DB, PostgreSQL 16.16
- Checked: the throwaway test above, `calc.StockAt(Telița, 18.11.2026)`.
- Finding: Grower 14 637 kg (last move 19.01.2026), Pre-ouat −22 462 kg (26.02.2026), Starter 7 109 kg
  (13.11.2025): Σ −716 kg, the workbook's figure. No screen shows it, but a new delivery to Telița puts this
  leftover into its Toast (e.g. 9 000 kg Pre-ouat → "stoc Telița: −13 462 kg").
- Consequence: left as is (user answer 28.11.2026: logged for the end review, plan Risks).

### Livrări keyset page and the report's statements on the seed   status: verified (2026-11-28, M4 T3)
- SPEC: §7.12 (Livrări rows, `nextCursor`), M4 D6, Q7; T3 criterion 3 (≤ 8 statements for the first page, 2 for a
  cursor page)
- Library / version: PostgreSQL 16 (testcontainers `postgres:16`), seeded test DB (`seed.Main`), pgx/v5 tracer
- Checked: `EXPLAIN ANALYZE` of `ReportsDeliveryRows` on the seed (owner, both farms, Seria 2026 ranges, limit 13) in a
  throwaway test (not kept); `reports.TestFeedStatementsSeeded` (statement names through a `pgx.QueryTracer`);
  `reports.TestFeedPagesSeeded` (82 rows over 7 pages, no repeat); `reports.TestFeedPagesLimit` (three deliveries
  tied on date, farm and type, pages of one row).
- Finding: the plan is a bitmap scan of `feed_delivery_farm_type_date_idx` (82 live rows) hashed against the
  `unnest` ranges, then a top-N heapsort on `date desc, farm name_key, type name_key, id`: 0,7 ms execution. The
  keyset predicate `(f.name_key, ft.name_key, fd.id) > (…)` on the cursor day compares uuids bytewise (= their
  lowercase text order), so ties on (date, farm, type) are ordered by id and no page boundary drops or repeats a
  row. The owner's first page runs 8 statements (`ReportsFeedCycles`, `CalcSettings`, `CalcStockEnd`,
  `CalcFeedDayComplete`, `CalcConsumption`, `CalcDeliveries`, `ReportsDeliveryRows`, `ReportsCostInputs`), about
  48 ms per build on a laptop without `-race`; a cursor page 2 (`ReportsFeedCycles`, `ReportsDeliveryRows`).
- Consequence: no new index. `ReportsCostInputs` batches `calc.FarmBase`, `calc.EggsBetween` and
  `calc.SeriesPlacement` into one statement to stay within 8; `reports.TestCostInputsMatchCalc` keeps it equal to
  those readers on the seed (Telița 28 080 · 3 510, Hîrbovăț 26 883 · 3 024 and 5 073 155 eggs, series 31 590).

### Delivery stockAfter on the seed (the planner's 21 119,3)   status: verified (2026-11-28, M4 T5; over HTTP in T7)
- SPEC: §6.7 (stock at the end of the delivery date, M4 Q13)
- Library / version: seeded test DB, PostgreSQL 16.16
- Checked: `entries.TestCreateDeliverySeeded` (oracle `calc.StockAt(12.11)` after the save) and `cmd/api`
  `TestDeliveryPriceChainSeeded` (the same re-save as Ion Rusu through the whole server).
- Finding: with the seeded 12.11.2026 delivery soft-deleted, the form's Breeder 2 stock now is 22 664,9 kg
  (31 664,9 − 9 000); the re-save answers `stockAfter` 21 119,3 and "Livrare adăugată · 9 000 kg Breeder 2 · stoc
  Hîrbovăț: 21 119 kg · fără preț"; the stock now is golden 27's 31 664,9 again. The boards' 16 140 / 25 140 kg are
  D39/D73 state examples, not reproducible on the seed.
- Consequence: the tests use `calc.StockAt` as the oracle; SPEC §6.7 quotes 21 119,3 as the seed's figure.

### M4 latency vs §11.6                                   status: verified (2026-11-28, M4 T7)
- SPEC: §11.6 (p95: summary/home ≤ 400 ms, any save ≤ 300 ms; the series page and Rapoarte › Hrană are page reads,
  held to the summary's 400 ms, plan Architecture notes)
- Library / version: `BenchmarkFarmReads` (cmd/api), PostgreSQL 16.16 in testcontainers (Docker Desktop), seeded data
- Checked: `go test -run '^$' -bench FarmReads -benchtime 200x ./cmd/api` (no `-race`), Apple M5 Max, the owner, 200
  requests each over a socket through the whole chain, 18.11.2026; three runs for the series page. The delivery save
  is a real save (a new 9 000 kg Breeder 2 delivery on 18.11 each time); the price save prices the seeded 12.11
  delivery, its price cleared before each request outside the timed part. The writes run after every read.
- Finding (p50 / p95 per request, the full run):
  | request | p50 | p95 | §11.6 target |
  |---|---|---|---|
  | `GET /api/series/{id}/page` Seria 2026, owner (two lanes, two panels with `BuildCycleStats`) | 104.9 ms | 139.9 ms | ≤ 400 ms |
  | `GET /api/reports/feed` owner, every farm, first page (Stoc, Livrări, Cost; 8 statements) | 41.3 ms | 53.7 ms | ≤ 400 ms |
  | the same with `cursor` (page 2: 2 statements) | 0.9 ms | 1.1 ms | ≤ 200 ms |
  | `GET /api/forms/delivery` Hîrbovăț | 11.8 ms | 16.6 ms | – |
  | `GET /api/deliveries?priced=false` | 0.5 ms | 0.7 ms | – |
  | `POST /api/entries/delivery` (11 statements + begin/commit) | 6.4 ms | 7.1 ms | ≤ 300 ms |
  | `POST /api/deliveries/prices`, one delivery | 1.4 ms | 2.7 ms | ≤ 300 ms |
  | `GET /api/farms/{id}/summary` Hîrbovăț (M3, for comparison) | 97.8 ms | 108.0 ms | ≤ 400 ms |
  The series page varied between runs (p50 / p95: 41.9 / 130.6, 84.7 / 107.8, 104.9 / 139.9 ms); it is the summary's
  stats once per panel, so it costs about one laying summary plus one closed growth cycle's stats.
- Consequence: every M4 route is inside §11.6 on the seeded data, the series page with about 3× headroom at p95.
  If production-sized data (≤ 5 years) brings it near 400 ms, the plan's next step is a per-panel snapshot or a slimmer
  stats read (plan Risks). Rerun before M6.

## M5

### Delta-based `CheckCounts` vs the dense view, and its plan      status: verified (2026-11-28, M5 T2)
- SPEC: §6.1 steps 2–3, plan M5-D4 (ARC-6: no `v_section_sex_day`, explicit `today`)
- Library / version: PostgreSQL 16.15 (`postgres:16`), sqlc v1.31.1, the seeded test database
- Checked: `TestCheckCountsSeeded` and `TestCheckCountsMatchesView` (`internal/ledger/counts_seeded_test.go`): every
  Hîrbovăț bird-holding section × sex, `from` = 22.02, 01.06, 03.06, 05.06, 17.11 and 18.11.2026, against
  `v_section_sex_day.alive_end < 0` grouped per pair (first day and its count, last day, lowest count, today's count)
  with `app.today = 2026-11-18`; negative days added to the seed: a 4 289 sale on H5.2 ♀ 17.11 (§6.1's 4 292
  correction) and an H6.1 ♂ −100 000 / +100 000 adjustment pair on 01.06 / 05.06. `EXPLAIN (ANALYZE, BUFFERS)` of
  `NegativeCounts` for H5.2 ♀ from 22.02.2026 (a temporary test, not kept).
- Finding: on the plain seed both report no negative day. With the added rows both agree on every pair and `from`;
  H5.2 ♀ from 17.11 = first 17.11 at −6, last 18.11, lowest −8, today −8, so the bound for 4 292 is 4 284. sqlc does
  not resolve the multi-argument `unnest(a, b) with ordinality`, so the query indexes two parallel arrays with
  `generate_subscripts`, and the sex array is `text[]` cast to `sex` (pgx has no encode plan for an unregistered
  enum array). Plan: `section_id = any($ids)` is pushed into the view (index scan on
  `bird_event_section_sex_date_idx`, 541 rows for H5.2 ♀), 537 event days, one sort and one window; execution
  3.4 ms, planning 1.0 ms (≤ 30 ms asked). `TestCheckCountsSeededTiming` logs about 2 ms per call through pgx.
- Consequence: `ledger.CheckCounts(ctx, db dbq.DBTX, affected, from, today)` is the §6.1 check of every M5 write and
  of golden row 37; it never reads `app_today()`, so it holds under the frozen clock and inside the import tx.

### `audit_log.changes` with number / string / null values      status: verified (2026-11-28, M5 T1)
- SPEC: §4.10, §7.8 (`Change`), plan M5-D9, Q12
- Library / version: PostgreSQL 16 (`postgres:16`), `encoding/json` (Go 1.27.4), pgx v5 (`[]byte` ↔ `jsonb`)
- Checked: `TestChangesRoundTripJSONB` (`internal/audit/change_test.go`): `audit.Write` of four `audit.Change` rows
  (int 1 → 3, text "boală" → "șchiopătare", null → 67.04, −2 → null) as il_app, read back as `jsonb` bytes and decoded
  into `[]audit.Change`; `TestChangeSeedShape` against the M1 seed's private `auditChange`.
- Finding: jsonb keeps numbers as numbers (67.04, −2), strings as strings and `null` as `null`; key order changes
  (jsonb sorts keys) but decoding is by name, so every `Change` comes back equal. A `Value` holds a number, a string or
  null; a bool, an object or an array fails to decode. Numbers go out as `num.Num` (at most 4 decimals). The §4.10
  example encodes byte for byte as the seed's `auditChange` (`section`, `sex`, `field` present there, `omitzero`
  only drops them when empty).
- Consequence: corrections (T6, T8), deletes and restores (T7) write `[]audit.Change`; T8 can swap the seed's
  `auditChange` for `audit.Change` without changing the seeded JSON.

### Monotonic version bump and `ifVersion` round trip      status: verified (2026-11-28, M5 T3)
- SPEC: §7.0 optimistic concurrency, §12.5 Version, plan M5-D3
- Library / version: PostgreSQL 16 (`postgres:16`), pgx v5 (`time.Time` ↔ `timestamptz`), Go 1.27.4 `time.Parse`
- Checked: `TestUpdateHeaderVersions` (`internal/entries/change_test.go`): an entry created at the frozen 12:00, two
  `UpdateEntryHeader` calls at the same frozen instant, each with the version string the previous answer returned,
  parsed back with `entries.ParseVersion` (RFC 3339, up to ns); then the first version again.
- Finding: `greatest($now, coalesce(updated_at, created_at) + interval '1 microsecond')` keeps µs precision through
  pgx: the versions are `…10:00:00.000001Z` and `…10:00:00.000002Z`; a version string parses back to the exact stored
  instant and compares equal in SQL; the stale one updates 0 rows (409 STALE, nothing written).
- Consequence: every M5 change moves the version through `ChangeTx.UpdateHeader`; no query sets `updated_at` with a
  plain `$now`.

### Entry row lock after the farm advisory locks, as il_app      status: verified (2026-11-28, M5 T3)
- SPEC: §12.1, plan M5-D2
- Library / version: PostgreSQL 16 (`postgres:16`), the `il_app` role (`0002_roles.sql`: select, insert, update on
  `entry`)
- Checked: `TestChangeLocksThenReadsThenClock`: a transaction holds farm A's advisory lock; `Runner.Change` on an
  entry of farm A waits (one ungranted lock in `pg_locks`) while another session takes `select … for update nowait`
  on the entry row successfully; after the holder commits the change runs with the clock read after the wait.
- Finding: `select … for update of e` works for `il_app` inside `InTx` (UPDATE privilege suffices); the row lock is
  taken only after the advisory locks, so the order is always farms (ascending) → entry row, the same direction as
  the M4 price writes (farms → `feed_delivery` rows): no lock-order inversion.
- Consequence: `Runner.Change` reads the entry's farms without a lock, locks them, then locks and re-checks the row.

### Zilnic rows in one statement each (`jsonb_to_recordset`), `total_eggs` on insert      status: verified (2026-11-28, M5 T5)
- SPEC: §4.7, §4.8, §12.5 (child inserts), plan M5 T5 criterion 6 (≤ 14 statements per save)
- Library / version: PostgreSQL 16.16 (`postgres:16.16-bookworm`), sqlc v1.31.1, pgx v5, `shopspring/decimal`
- Checked: `DailyInsertBirdEvents` / `DailyInsertEggProduction` (`db/queries/daily.sql`) through
  `TestCreateDailySaves`, `TestCreateDailyEggsOnly`, `TestCreateDailyStatements` and `TestDailyInsertConstraintNames`
  (`internal/entries/daily_test.go`) as `il_app`.
- Finding: sqlc accepts `jsonb_to_recordset(sqlc.arg(rows)::jsonb) as r(… sex sex, type bird_event_type …)` and
  generates a `[]byte` parameter; enum and uuid columns cast from their JSON strings, JSON `null` gives SQL NULL (an
  empty egg field, a missing cull reason), and a `decimal.Decimal` (encoded as a quoted string, `"67.04"`) lands in
  `numeric(5,2)` exactly. `total_eggs` is generated on insert (H5: 13 + 417 + 27 + 3 415 = 3 872). A save with a key,
  bird values and eggs is 14 statements besides begin and commit; the bird_event_daily_uq / egg_production_uq
  violation of a batch insert still carries the constraint name for the after-transaction mapping.
- Consequence: the Zilnic save inserts all its bird rows and all its egg rows in two statements, whatever the number
  of sections.

### Soft delete and restore of child rows by id; restore against the partial unique indexes      status: verified (2026-11-28, M5 T7)
- SPEC: §4.7 (soft delete), §7.8 (delete, restore), plan M5-D7
- Library / version: PostgreSQL 16.16 (`postgres:16.16-bookworm`), sqlc v1.31.1, pgx v5, the `il_app` role
- Checked: `EntrySoftDeleteRows` / `EntryRestoreRows` (`db/queries/entrydelete.sql`: four data-modifying CTEs, one per
  child table, in one statement) through `TestDeleteByAuthor`, `TestDeleteSection`, `TestRestoreKeyRetaken` and
  `TestRestoreKeepsSupersededRows` (`internal/entries/delete_test.go`) as `il_app`.
- Finding: `update … returning 1` in four CTEs runs in one statement under `il_app` (UPDATE grants only) and the
  outer count is every row changed. Un-deleting a `bird_event` row whose (section, date, sex, type) another live row
  holds fails `bird_event_daily_uq` with 23505 and the constraint name, as an insert would; a row soft-deleted by an
  earlier edit at the same frozen instant stays deleted because the restore names row ids from the delete's
  `old_values`, never `deleted_at`.
- Consequence: a delete and a restore write their child rows in one statement each; a retaken key is mapped after
  the transaction to 409 ALREADY_ENTERED with the holder (`EntryRestoreHolder`).

### Child rows changed in place (`UPDATE … FROM jsonb_to_recordset`), `total_eggs` after an update      status: verified (2026-11-28, M5 T6)
- SPEC: §4.7, §4.8, §6.15, §7.8 (PUT), plan M5-D6, Verify "Partial unique indexes" and "`egg_production.total_eggs`"
- Library / version: PostgreSQL 16.16 (`postgres:16.16-bookworm`), sqlc v1.31.1, pgx v5, the `il_app` role
- Checked: `EntryUpdateBirdRows` / `EntryUpdateEggRows` (`db/queries/entryupdate.sql`, `:execrows`) through
  `TestUpdateDailyAuthorEdit`, `TestUpdateDailyRows`, `TestUpdateDailyCullReason`, `TestUpdateDailyEggs` and the undo
  tests (`TestUndoEditRows`, `TestUndoUpdateConflicts`) in `internal/entries`, as `il_app`.
- Finding: sqlc accepts `update … from jsonb_to_recordset(sqlc.arg(rows)::jsonb) as r(id uuid, qty integer,
  cull_reason cull_reason)` and reports the rows changed; the enum casts from its JSON string, JSON `null` clears a
  column. A value changed in place keeps its row id and its `bird_event_daily_uq` / `egg_production_uq` entry (no
  delete + insert of the same key is ever needed: a key present before and after is always an in-place change).
  `total_eggs` is recomputed by the UPDATE of one category (13 → 15 gives 15 with the other three at 0) and is never in
  the SET list. An undo un-deletes a row whose key another live row took meanwhile → 23505 with the constraint name
  (mapped to 409 ALREADY_ENTERED by `EntryRestoreHolder`).
- Consequence: a Zilnic PUT writes soft deletes, in-place changes and inserts in at most one statement each; a
  correction is 15 statements besides begin and commit (`TestUpdateDailyStatements`: 14 + the one per-row write).

### Z5 against §6.1 on a correction (the S16 4 292 refusal)      status: verified (2026-11-28, M5 T6)
- SPEC: §6.1 (Corectează "removal raised"), §6.2 Z5, §11.8 "Rules", plan M5-D6, Assumptions (bound 4 284)
- Library / version: the seeded test database (`TestUpdateDailySeeded`, `internal/entries/update_seeded_test.go`)
- Checked: H5.2 ♀ 17.11.2026 corrected to 1 by the owner, then Ion Rusu's dry run 1 → 3 and a typed 4 292.
- Finding: 4 292 also breaks Z5 on 17.11 (4 286 birds at the start of the day: "… cel mult 4 286"), which would hide
  the §6.1 answer the spec and the S16 board give. The dry run reproduces the §7.8 effects ("… se schimbă cu −2.",
  "Mortalitate · morți ♀ −5 → −7", "Hîrbovăț · ♀ azi 12 563 → 12 561"), the alive rows 4 285 → 4 283 / 4 283 → 4 281,
  `checkedDays` 2 and golden 33's single warning; the save's Toast is "Corecție salvată · H5.2 ♀ 1 → 3 · în viață ♀
  4 281".
- Consequence: on a PUT, Z5 refusals alone don't stop the plan: the rows are written and §6.1 runs; a pair §6.1 finds
  below 0 answers §6.1's field sentence and bound ("H5.2 ♀ 4 292: ar lăsa −8 găini azi, 18.11.2026. Scrie cel mult
  4 284.", banner "Corecția ar lăsa …"), a Z5 refusal stays only for a pair §6.1 lets through (a positive sexing error
  the same day). Z3, Z4, Z7 and Z9 still stop before any write.

### Feed and delivery rows changed in place; `closed_series_on` in the feed reads      status: verified (2026-11-28, M5 T8)
- SPEC: §4.8, §6.3, §6.7, §6.8, §7.8 (PUT), plan M5-D6, D12, Q11, Verify "Partial unique indexes"
- Library / version: PostgreSQL 16.16 (`postgres:16.16-bookworm`), sqlc v1.31.1, pgx v5, shopspring/decimal, the
  `il_app` role
- Checked: `EntryUpdateFeedRows`, `EntryInsertFeedRows`, `EntryUpdateDeliveryRow` (`db/queries/entryupdate.sql`)
  through `TestUpdateFeedDaily*`, `TestUndoFeedUpdateRows`, `TestUpdateDelivery*`, `TestUndoDeliveryUpdateKeepsNewPrice`
  and the seeded `TestUpdateFeedDailySeeded` (`internal/entries`); `EntriesFeedSaveContext` and `FormsFeedHead` with
  `closed_series_on` through the unchanged M3 feed tests (`TestCreateFeedDaily*`, `TestBuildFeedForm*`).
- Finding: a g or type changed in place keeps the row id and its `feed_allocation_uq` entry; `g` from a JSON number
  (`num.Num`) lands in `numeric(6,1)` unchanged; a JSON `null` `kg_recorded` clears it. A Hrană zilnică correction of
  the seed's 18.11 save (H5.1 ♀ −2 g) moves the save from 2 457,9 kg to 2 449,4 kg (2 g × 4 264 hens) and its undo
  back. A priced delivery re-derived by `calc.DerivePrice` satisfies `feed_delivery`'s price check (both values set,
  `priced_at` kept): 400 kg at 7,74 lei/kg typed → 3 096 lei; 3 870 lei typed → 9,675 lei/kg. `closed_series_on(farm,
  day)` gives the same answer as the former cycle join with `series.status = 'closed'` (the closed Seria 2025 of the
  M3 fixtures: SERIES_CLOSED and the form's `lockedReason` unchanged).
- Consequence: a feed PUT writes soft deletes, in-place changes and inserts in one statement each; a corrected g
  charges g × birds (an imported row's `kg_recorded` is cleared only when its g changes). A delivery's undo keeps the
  typed price the row has now (a price changed in Prețuri after the correction stays; the other value is derived again
  for the kg put back). A register price (no `price_input`) keeps its total lei on a kg change (assumption, C6
  "invoice total").

### A DELETE with a JSON body: decoder, middleware, Go client, Caddy      status: verified (2026-11-28, M5 T9)
- SPEC: §7.8 (`DELETE /api/entries/:id` `{ reason?, sectionId?, dryRun? }`), §7.0a, §11.4, plan M5 Verify "DELETE with
  a JSON body"
- Library / version: Go 1.27.4 `net/http` (client and ServeMux), `httpx.Decode`, `httpx.ContentType`, `httpx.CSRF`;
  Caddy 2 `reverse_proxy` (compose `proxy`)
- Checked: `TestM5WritesContract` and `TestEveryWriteRouteChecksContentType` (`cmd/api`): `http.NewRequest("DELETE",
  …, body)` through the whole chain, an unknown member, a 65 KiB body, the Origin / `Sec-Fetch-Site` / Content-Type
  refusals, a valid dry run; `scripts/e2e-login.sh` step 3b (`make e2e`): a DELETE with `{"dryRun":true,"e2e":1}`
  through `https://localhost:8443`.
- Finding: Go's client sends a DELETE body (Content-Length set from the reader); `httpx.Decode` reads it as for POST
  (unknown member → 400 naming it, over 64 KiB → 413); the Content-Type rule and the CSRF Origin rule cover DELETE
  (they key on the method, not on POST); Caddy's `reverse_proxy` forwards the body unchanged.
- Consequence: no special case for DELETE anywhere; the route sits on the same `authed` chain as the other writes.

### The seed's 18.11 Zilnic entries (Q13 keys, S15 times)      status: verified (2026-11-28, M5 T9)
- SPEC: §11.8 concurrency (M5 Q13), §6.15 (S15 "Se șterge rândul salvat la 08:15"), plan M5 Verify "The seed's
  18.11 Zilnic entries"
- Library / version: the seeded test database (`testdb.NewSeeded`), a temporary query test (not kept)
- Checked: every live `daily` entry of Hîrbovăț on 18.11.2026 with its author, local creation time, live bird rows,
  egg rows and audit rows; live `cull` / `sexing_error` rows of H6.2 ♂ on 18.11.
- Finding: two entries by Dr. Ceban: 08:15 mortality of all four compartments and both sexes (H5.1 ♀ −1, H5.2 ♀ −2,
  every other value 0, so H6.2 ♂ holds a 0 mortality row), and 08:27 the eggs of H5 and H6 (two `egg_production`
  rows, no bird row). H6.2 ♂ has no cull and no sexing-error row on 18.11, so the concurrency test's two keys are
  free. Neither entry has an audit row (the seed writes audit rows only for the corrected 17.11 save), so Dr. Ceban's
  undo of his own seeded 18.11 save answers 409 STALE and the entry read shows `undo: false`; his own creates through
  the API are undoable (`TestZilnicChainSeeded`).
- Consequence: the S15 08:15 figures and the Q13 keys hold as planned; tests that need an undoable seeded save
  create it through the API first.

### `entries.Actions` under the frozen clock      status: verified (2026-11-28, M5 T9)
- SPEC: §3.2 windows (D70, D71), plan M5 Verify "`entries.Actions` with `created_at` from the frozen clock"
- Library / version: `timex.FixedClock(2026-11-18T12:00 Europe/Chisinau)`, the seeded database
- Checked: `TestZilnicChainSeeded` (`cmd/api`): the entry read of Dr. Ceban's seeded 08:15 save as Dr. Ceban, of the
  seeded 17.11 save as the owner and Ion Rusu; Ion Rusu's form state of 18.11 (`existing`); `TestAuthMatrix`: the
  17.11 entry of another vet.
- Finding: an entry whose `created_at` local day is the clock's day (the seed's 08:15 save, or one created by a test at
  12:00) is `edit: true, delete: true` for its author and `correct` only for a manager or the owner; the seeded 17.11
  save is `edit: false, correct: true, delete: true` for the owner and nothing for Dr. Ceban (a past day).
- Consequence: the windows read only `created_at` and the injected clock, never `now()` or `app_today()`.

### READ COMMITTED and the missing-lock test (§11.8 (3))      status: verified (2026-11-28, M5 T10)
- SPEC: §11.8 concurrency (1) and (3), plan M5-D8, Q13, Verify "READ COMMITTED visibility proving test (3)"
- Library / version: PostgreSQL 16.16 (`postgres:16.16-bookworm`, one container via `TEST_DATABASE_URL`), pgx/v5 pool
  (4 connections per test database), the seeded database
- Checked: `internal/entries/concurrency_test.go`: two Zilnic saves of Ion Rusu on H6.2 ♂ 18.11.2026 started together
  (A Sacrificare 4 boală, B Eroare sexare −4; the fixture first asserts H6.2 ♂ 93 at the start and 5 at the end of
  18.11 and no live cull / sexing-error row there). `TestConcurrentZilnicSavesHoldTheCount`: the first save to reach
  `Tx.CheckCounts` (via `Runner.SetBeforeCheck`) waits until `pg_locks` shows an ungranted advisory lock in the test's
  database, i.e. the other save is blocked on the farm lock while the first sits between its insert and its §6.1
  check. `TestConcurrentZilnicSavesWithoutTheLockBreakTheCount`: `Runner.SetLock` no-op and a two-party barrier in
  `beforeCheck` (both saves inserted, neither committed). `go test -race -count=20 -run 'TestConcurrentZilnic'
  ./internal/entries` (the nightly job's form): 40 runs, all pass.
- Finding: with the lock, the second save waits on `pg_advisory_xact_lock` until the first commits, then its
  `CheckCounts` (a new statement, so a new READ COMMITTED snapshot) sees the committed cull: one 201, one 422 whose
  refusal is H6.2 ♂ "… Scrie cel mult 1." with `max 1`; H6.2 ♂ ends 18.11 at 1 and `v_section_sex_day` has no row
  below 0. Without the lock, each `CheckCounts` sees only its own uncommitted rows (5 − 4 = 1), both pass, both answer
  201, and H6.2 ♂ ends 18.11 at −3 (one negative section × sex day). The two keys are different (cull, sexing_error),
  so `bird_event_daily_uq` never serialises them.
- Consequence: the farm advisory lock is what holds §6.1 under concurrency; READ COMMITTED is enough behind it (no
  SERIALIZABLE). The nightly job runs both tests 20 times; the opposite-direction transfer test (2) comes with M7.

### Goldens 33 and 37 on the bare import and on the dev seed      status: verified (2026-11-28, M5 T10)
- SPEC: §11.7 rows 33, 37, §10.3 (Verify inside the import transaction), plan M5-D11
- Library / version: `go run ./cmd/migrate-history` against a fresh compose-style PostgreSQL 16.16 (`deploy/initdb`
  roles, `api migrate up`), then `APP_ENV=dev … --replace --seed-dev`; `TestLoad` (Verify on the committed bare
  import); `TestGolden` (the seed)
- Checked: where the 17.11.2026 Zilnic values live: on the seed in Dr. Ceban's corrected `daily` entry (latest audit
  row: Ion Rusu's `update`, changes `[{key "H5.2 · găini moarte", old 1, new 3}]`), on the bare import in the farm's
  `migration` entry (no author, no update audit row). The Zilnic form state (`entries.BuildDailyForm`) shows both as
  the day's saved values.
- Finding: the bare import holds the register's 3 for H5.2 ♀ and 4 for H6.1 ♀ on 17.11, the same as the seed, so
  golden 33 (one Z6 warning, H6.1 4 vs mean 0,857) holds on both. The old value 1 of golden 37 is illustrative
  (§11.7, `illustrative-events-hirbovat.json` corrections[0].old_values) and exists only in the seed's audit row.
  Both imports exit 0 ("imported and verified"); Verify runs 38 checks (35 goldens).
- Consequence: both rows read the saved values through the form state; golden 37 takes the old value from the
  latest `update` audit row of the entry holding H5.2's values, or the illustrative 1 when that entry has no author
  (the register). An authored entry without an update row is an error, so a seed that lost its correction fails.

### M5 latency vs §11.6                                   status: verified (2026-11-28, M5 T9)
- SPEC: §11.6 (p95: any save ≤ 300 ms incl. §6.1 + audit + live; the form state is held to the summary's 400 ms)
- Library / version: `BenchmarkFarmReads` (cmd/api), PostgreSQL 16.16 (`postgres:16.16-bookworm`, one container via
  `TEST_DATABASE_URL`, Docker Desktop), seeded data
- Checked: `go test -run '^$' -bench 'FarmReads/(summary$|daily|correction|delete|restore)' -benchtime 200x ./cmd/api`
  (no `-race`), Apple M5 Max, the owner, 200 requests each over a socket through the whole chain, two runs. The Zilnic
  save re-saves Hîrbovăț's seeded 18.11 mortality (4 sections) after its live rows are soft-deleted outside the timed
  part; the correction moves H5.2 ♀ of the corrected 17.11 save 3 → 4 → 3 (body with the version just read, built
  outside the timed part); the delete dry run is the whole delete path rolled back; the delete and the restore of that
  save alternate (the other one done outside the timed part).
- Finding (p50 / p95 per request, two runs):
  | request | p50 | p95 | §11.6 target |
  |---|---|---|---|
  | `GET /api/forms/daily` Hîrbovăț 18.11 | 17.7 ms | 18.6–19.0 ms | ≤ 400 ms |
  | `POST /api/entries/daily` (4 sections, §6.1, audit, 2 live events) | 18.1–18.2 ms | 19.0 ms | ≤ 300 ms |
  | `PUT /api/entries/{id}` Corectează 17.11 (§6.1 over 2 days) | 10.4–21.8 ms | 22.8 ms | ≤ 300 ms |
  | `DELETE /api/entries/{id}` dry run (rolled back) | 11.5–17.9 ms | 11.9–18.5 ms | ≤ 300 ms |
  | `DELETE /api/entries/{id}` with a Motiv | 8.4–14.0 ms | 8.9–14.8 ms | ≤ 300 ms |
  | `POST /api/entries/{id}/restore` | 9.2–15.7 ms | 9.7–18.0 ms | ≤ 300 ms |
  | `GET /api/farms/{id}/summary` Hîrbovăț (M2, for comparison) | 74.5 ms | 76.9 ms | ≤ 400 ms |
- Consequence: every M5 route is an order of magnitude inside §11.6 on the seeded data. Rerun before M6.
