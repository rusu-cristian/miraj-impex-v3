---
name: m5-review-patterns
description: Recurring correctness hotspots in the Impex Line backend (change runner, audit/version columns, undo/restore) seen in the M5 round 1 review
metadata:
  type: project
---

M5 round 1 (base 55d7d40): all packages green against real PG16 (ledger, entries, golden, calc, migratehistory, cmd/api); `-count=8` of TestConcurrent* stable. No blocking bug found; review was by code reading since reviewers cannot write scratch test files (readonly-guard blocks Write/cp outside the agent-memory folder, so `-overlay` probes are not possible).

Hotspots worth re-checking in later milestones:
- `entry.updated_by/updated_at` are overloaded: `UpdateEntryHeader` sets them on every change (author edit, delete, restore, undo), but GET /api/entries/{id} (`corrected {by, at}`) and the delete Dialog read them as "who corrected / when". A later same-day author edit or a section delete overwrites the corrector.
- Price writes (M4 `price_set` / `price_change`) write audit rows with no `entry_id`, so `EntryLatestAudit` (by entry_id) never sees them: an undo of a create still passes after someone priced the delivery.
- Undo/restore/PUT dispatch on the entry's latest audit row; seeded audit rows (migratehistory/seeddev.go) have no old_values snapshot, so `actions.undo` can be true while POST undo answers 409 STALE (seed only).
- Z5 gives way to CheckCounts' stricter bound in `planDailyUpdate` (dayBounds); refusal order in the response is CheckCounts first, then leftover Z5 bounds, not document order.

**How to apply:** when a later milestone adds adapters (transfer, sale, weighing, placement) re-check that they go through `Runner.Change`, take `ledger.LockFarms` before `EntryChangeLock`, and decide what `updated_by` means for their reads.
