---
name: trust-boundaries
description: Trust boundaries of the Impex Line Go backend (proxy/XFF, CSRF/Origin, DB roles, secrets) as mapped at M0 review, 2026-11-27
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

**How to apply:** in later milestones check new routes get RequireSession + RequireRole/CanSeeFarm, new tables get explicit grants, and nothing logs bodies/cookies/tokens (§11.11).
