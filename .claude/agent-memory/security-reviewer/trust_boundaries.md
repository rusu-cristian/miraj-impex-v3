---
name: trust-boundaries
description: Trust boundaries of the Impex Line Go backend (proxy/XFF, CSRF/Origin + per-route Content-Type, DB roles, secrets, migrate-history guards) as mapped at the M0 and M1 reviews, 2026-11-27
metadata:
  type: project
---

Trust boundaries mapped in the M0 review (2026-11-27, branch team/backend):

- **Edge**: reverse proxy (Caddy in dev, `deploy/Caddyfile`) -> Go API. X-Forwarded-For/-Proto believed only from `TRUSTED_PROXIES` (`internal/platform/httpx/clientip.go`, right-most untrusted hop). Caddy without trusted_proxies overwrites client XFF. Compose pins the proxy at a fixed IP and trusts only that /32.
- **CSRF**: global middleware, Origin must equal APP_ORIGIN on non-GET/HEAD/OPTIONS, else Sec-Fetch-Site same-origin; JSON content type required. http.CrossOriginProtection deliberately not used.
- **Session**: il_session cookie, sha256 in user_session; lookup joins active app_user; RequireSession is per-route inside the mux (modules must add it themselves; the §3.3 matrix test is the guard).
- **Login throttle**: counted in login_attempt (DB). Since M0 round 2, reserveAttempt runs count+insert-as-failure in one READ COMMITTED tx under pg_advisory_xact_lock('login:'||login) then ('ip:'||host(ip)); argon2 runs after commit; success flips the row. Keep READ COMMITTED: under REPEATABLE READ the snapshot predates the lock wait and the throttle breaks. Advisory keys share one 64-bit space with farm locks (hashtextextended(farm_id)), kept apart by the 'login:'/'ip:' prefixes.
- **Login input**: normalized (TrimSpace+ToLower) before the throttle key; storedText maps control chars/invalid UTF-8 to U+FFFD and caps length before storing or locking.
- **DB roles**: il_owner (migrate), il_app (serve; no UPDATE/DELETE on audit_log, live_event, request_replay; read-only goose_db_version), il_maint (purge). Roles created by infra, grants in 0002_roles.sql. New tables need explicit grants (no default privileges).
- **Secrets**: APP_SECRET (HKDF root, tokens pkg), three DSNs; config.Config redacts in LogValue/String/GoString. `.env` git-ignored; `.env.example` holds dev placeholders including a fixed dev APP_SECRET, which config refuses outside APP_ENV=dev (default env is production). Each command reads only its own DSNs.

Added in M1 (review 2026-11-27):

- **CSRF split**: `httpx.CSRF` (Origin rule) is global; the Content-Type rule is per route via `httpx.ContentType`, carried by the `public`/`authed` chains in `cmd/api/routes.go`. A route registered on any other chain has no Content-Type check, so look at the chain of every new write route (multipart/audio routes build their own).
- **Session**: GetActiveSession returns farm_ids in the same query; owner FarmIDs is forced to nil in Go (principalFarms).
- **MuxFallback**: wraps the mux inside the global chain. Unmatched /api paths get the 404/405 envelope; it runs the mux's unmatched handler against a throw-away probe writer.
- **499**: WriteError writes nothing when the error is context.Canceled and AccessLog's request ctx is done. No BaseContext on the server, so this only happens when the client is gone.
- **migrate-history**: ships in the prod image. `--replace` (TRUNCATE of every public table except goose_db_version/app_setting, needs owner) and `--seed-dev` (4 users incl. an owner, crypto/rand passwords printed once to stdout) run only with APP_ENV=dev (allowlist since M1 round 2; staging/production/unset refused). Still no DB-side guard: the check is the operator's env, not the target DB. `testdb.DevPassword` is a fixed test password that stays out of the binaries (go list -deps checked). ErrCredentialsLost wraps only the stdout write error, never the passwords.
- **Route-walk guard**: `TestEveryWriteRouteChecksContentType` (cmd/api/contract_test.go) walks every registered /api write pattern with an owner session + valid Origin + text/plain and wants 403; non-JSON routes must be listed in `otherMediaRoutes`. A new write route that skips ContentType fails it.
- **Hashing**: all password hashes come from `auth.NewHasher()` (Service, create-owner, seed-dev); each Hasher has its own maxHashers=4 semaphore, so only the Service's instance bounds request-path argon2.

**How to apply:** in later milestones check new routes get RequireSession + RequireRole/CanSeeFarm, new tables get explicit grants, and nothing logs bodies/cookies/tokens (§11.11).
