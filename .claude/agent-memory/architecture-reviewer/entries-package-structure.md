---
name: entries-package-structure
description: Inferred conventions of backend/internal/entries (runner, kind dispatch, file layout) as of M5, for judging M7/M9 kinds
metadata:
  type: project
---

- `Runner.Create` (save.go) and `Runner.Change` (change.go) are the only write skeletons; each kind supplies a plan callback (`ChangePlan`) and `Allow`/`OnConflict`. Plan functions live in `<kind>_update.go`, creates in `daily.go`/`feeddaily.go`/`delivery.go`.
- Kind dispatch is by `switch kind` in several places (kinds.go `Served`/`LoadStored`/`Values`, update.go handler, undo.go, delete.go message switches, restore.go toast, entryread.go) plus content-based `len(rows.X) > 0` checks; `Snapshot` is a closed struct with one slice per child table. A new kind (transfer M7, sale, weighing M9) means editing all of these; defaults silently fall to the Zilnic wording.
- delete.go holds helpers shared by undo/restore/update (rowFigures, negativeRefusals, readStockAfter, Snapshot id getters): file name misleads.
- ARC-6 grep (`app_today|v_section_sex_day|...` over db/queries) must print only calc.sql comment lines; it did at M5.
- Review approach that worked: read change.go + kinds.go first, then grep for `EntryKind|case dbq.EntryKind|Snapshot fields` across non-test files to find dispatch sites.
