---
name: trust-boundaries
description: Impex Line backend trust boundaries and where each authz/CSRF/price control lives (mapped at M5 review, 2026-11-28)
metadata:
  type: project
---

Trust boundaries mapped in the M5 review (2026-11-28). Verify paths still exist before citing.

- CSRF: global `httpx.CSRF` (Origin == APP_ORIGIN, else Sec-Fetch-Site same-origin) + per-chain `httpx.ContentType(application/json)` on every non-GET/HEAD/OPTIONS method, DELETE included. Contract test `TestEveryWriteRouteChecksContentType` must list each new write route.
- Role/farm: `checkWriter` (role + CanSeeFarm) in handlers before any read; entry routes re-check under farm advisory locks in `entries.Runner.Change` (farm check, closed series 409 after Allow so 403 wins).
- Edit vs correction is server-side: `Actions()` (actions.go) Editable = own && created today; Correctable = manager/owner. Undo/restore = `checkOwnLatest` on `audit_log` latest row by entry_id (user_id + local day), not on client input.
- M5 round 2 (T11): entry kind dispatch is the `kinds` map in internal/entries/kinds.go (`adapterOf`/`Served`); unknown kind → 404 in `checkEntryHead` (change.go, shared by PUT pre-read and Runner.Change: 404 missing → 403 farm → 404 unserved → 404 deleted) and in `BuildEntryRead`. Per-kind update plans re-check `e.Kind` under the lock. Undo/restore rule = `ownLatestError` + `originalAllowed` (delete.go): latest audit row must be caller's own, today, AND current Actions allow it (create/same_day_author → Editable, else Correctable); entry read `readActions` uses the same function. A new kind's adapter must keep `values` price-free.
- Restore/undo read row ids from `audit_log.old_values` (server data, not client); all child-row UPDATE SQL is scoped `entry_id = @entry_id`.
- Prices (D200): vet/viewer never; manager read-only. Entry read `DeliveryValues` has no price fields by type; audit snapshots DO hold prices (only Istoric M9 could expose them; vet/viewer have no Istoric). Matrix test runs `noPrices` on vet/viewer 200s, but the fixture delivery is unpriced.
- Live events (`live.EntryPayload`) carry summary/author, never Motiv or prices.
- CI: `.github/workflows/backend.yml` contents:read, persist-credentials false, SHA-pinned actions, no secrets; nightly `concurrency-nightly` job on schedule/workflow_dispatch.

**Why:** lets the next review jump straight to the controls instead of re-mapping them.
**How to apply:** for a new route, check it is in the matrix + content-type tests, uses checkWriter/Runner, and that any price-bearing struct is not reachable by vet/viewer. Past findings: [[review-findings-log]].
