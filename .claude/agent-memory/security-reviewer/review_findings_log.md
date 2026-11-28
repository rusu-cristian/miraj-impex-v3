---
name: review-findings-log
description: Security findings raised per milestone review of the Impex Line backend, to avoid repeats and track follow-ups
metadata:
  type: project
---

- M3 (from SPEC §4.11 note): SEC-1 request_replay key reuse across users/routes, fixed (409 ALREADY_ENTERED only to the creator, else STALE).
- M5 round 1 (2026-11-28): approve. Minors: SEC-1 no test that GET /api/entries/{id} of a *priced* delivery hides price from vet/viewer (fixture delivery unpriced); SEC-2 undo/restore authz is "own latest audit row" only, does not re-check the role right the original action needed (demoted user edge).
- M5 round 2 (2026-11-28): SEC-1 and SEC-2 resolved (priced Telița delivery in matrix fixture + TestEntryReadPricedDeliveryHasNoPrice; TestUndoRestoreNeedCurrentRights). Approve, no new findings.
- Accepted by plan (do not re-raise): vet/viewer may GET any served entry on own farms incl. delivery kg/note (M5 Q10); vet past-day Zilnic form returns data with lockedReason (Q2); 403 vs 404 existence oracle on entry ids (UUIDv4).

**Why:** keeps rounds consistent and stops re-raising decided items.
**How to apply:** check here before raising; in M9 (Istoric) verify audit snapshots' delivery prices are stripped for any non-owner/manager reader. See [[trust-boundaries]].
