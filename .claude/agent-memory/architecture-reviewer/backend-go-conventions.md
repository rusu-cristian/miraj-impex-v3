---
name: backend-go-conventions
description: Inferred layering conventions of the Go backend (M0 baseline, M1/M2 updates) to compare later milestones against
metadata:
  type: project
---

Inferred from M0/M1 reviews (backend/, SPEC §12.3 is the source):
- Modules expose NewService / NewHandler / (*Handler).Routes(httpx.Router, httpx.Chain). Global middleware wraps the mux in cmd/api/wiring.go; per-route (session, role, content-type) goes through the chains built in cmd/api/routes.go (public, authed).
- Spec rule: dbq types stay inside services. As of M1 auth.Role/Principal no longer leak dbq (M0 leak fixed); httpx no longer imports internal/db (pgerr split). calc readers take dbq.DBTX but return plain structs.
- Session middleware is 1 query (farm ids aggregated in GetActiveSession) as of M1.
- Import direction M1: cmd/migrate-history -> migratehistory -> {auth (seeddev hashing), calc, golden, dbq}; golden -> calc; testdb/seed -> {testdb, migratehistory}. Go forbids in-package tests importing a package that imports the package under test, so any module that golden/migratehistory imports (calc, auth, later periodtable, ledger...) can use testdb.NewSeeded only from an external `_test` package. Recorded only partly in NOTES-verify; flag if later milestones hit it.
- golden.Env.DB is dbq.DBTX so Verify can run rows inside the uncommitted import tx. Later-milestone Got funcs must call DBTX-taking builders, not *db.Store services that open their own InTx.
- migratehistory/load.go write() maps dbq.X -> dbq.CopyXParams by hand per column; a new column is silently dropped unless a test guards it.
- Health, route recorder and readiness SQL live in cmd/api (package main), not internal.

M2 additions (inferred):
- Builders are `Build*(ctx, dbq.DBTX, Input{..., Today timex.Day, Now time.Time})`; the Service wrapper only checks access, calls them on `store.Pool()` (autocommit, no shared snapshot) and maps errors to httpx. golden is imported only by migratehistory (so cmd/migrate-history links farms/periodtable).
- Hand-written SQL is allowed only for the PeriodTable day query (periodtable/sql.go, scope whitelist); everything else is sqlc. Day query aggregates to egg-unit before scope, so a section scope (M6) needs CTE rework, not just a new scope fragment.
- Guide lookups (layPct, feedG, bodyWeight) and cycleType/pickCycle exist in both periodtable and farms; M3+ should move them onto calc.GuideWeeks / calc.
- SPEC §11.8 puts JSON goldens in each module's testdata; M2 put them in internal/golden/testdata (deviation).

**How to apply:** when reviewing M2+ check (1) new golden Got funcs stay DBTX-based, (2) calc/periodtable per-day readers are not looped over day ranges (N+1), (3) new module tests that need the seeded DB are external test packages.
