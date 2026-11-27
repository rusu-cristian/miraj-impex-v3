---
name: backend-architecture-conventions
description: Inferred layering and pattern conventions of the Go backend (builders, services, write runner, SQL) used when reviewing milestones
metadata:
  type: project
---

Conventions inferred from M0-M3 code (not all spelled out in CLAUDE.md):
- Module shape: `Build*(ctx, dbq.DBTX, Input{Principal, ..., Today, Now})` builders (no app_today(), no dense views, ARC-6) + thin `Service` (store, clock, access check, `InReadTx`) + `Handler.Routes(mux, authed)`; DTOs in `*_dto.go` / `dto.go`.
- Write path: `entries.Runner.Create(p, Create{Route,Kind,FarmIDs,Meta,OnConflict}, Plan)`; a form plugs `planX` only. Locks via `ledger.LockFarms`, `audit.Write`, `live.Emit` in the same tx.
- Import direction: cmd -> entries/reports/farms/periodtable -> calc/ledger/audit/live -> db/platform. `calc` already imports `platform/messages` (period.go), so shared Romanian-copy selection for calc enums can live in calc.
- Raw SQL const in periodtable (sql.go, feedsql.go) has precedent from M2; everything else is sqlc in `db/queries/*.sql`.
- Recurring smell to check each milestone: the same wire-mapping helper copied per module (M3: `stockReason` x3, forSex *dbq.Sex -> *string x3).
- Milestone-suffixed file names (`calc/readers_m3.go`, `golden/m3.go`) appeared in M3; prefer domain names.
- Perf: stock/consumption readers scan the farm's whole v_section_sex_delta history; measured 30-40 ms (NOTES-verify), budgets in SPEC §11.6 hold.
