---
name: backend-go-conventions
description: Inferred layout and patterns of the Impex Line Go backend that reviews compare new code against (M4 review)
metadata:
  type: project
---

Inferred from reading M3/M4 code (not all written in CLAUDE.md):

- Module shape: `internal/<module>/{handler,service,dto,<builder>}.go`. The builder is an exported `BuildX(ctx, dbq.DBTX, Input{Principal, ..., Today, Now})`; the `Service` is a thin wrapper (store + clock + role check) running it in `store.InReadTx`; writes use `store.InTx` or `entries.Runner.Create`.
- Role-dependent JSON uses separate Go types per role (viewer type has no price fields), not nil pointers: `reports.FeedResponse` vs `FeedResponsePriced`.
- Handler order for creates: decode -> role/farm check -> refs read -> `Validate` (level 1) -> service (level 2 inside tx after farm lock). The role check runs in both handler and service (accepted precedent since M3).
- Batched SQL readers use parallel arrays + `generate_subscripts`; sqlc output in `internal/db/dbq` is never hand-edited.
- Where SQL re-implements a `calc` reader for statement-count reasons (`ReportsCostInputs`), a test (`TestCostInputsMatchCalc`) pins them equal.
- The closed-series predicate ("date inside a cycle of a closed series") is copy-pasted in SQL (prices.sql x2, entries.sql, forms.sql) and Go (`reports.farmScope.closedOn`); known drift risk, reported as minor in M4.
- Verify-at-implementation findings go to `backend/NOTES-verify.md` with a status line; statement budgets are checked with a pgx tracer in tests.
