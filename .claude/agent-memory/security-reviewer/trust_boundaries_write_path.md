---
name: trust-boundaries-write-path
description: Impex Line backend write-path trust boundaries mapped at the M3 review (2026-11-27): entry create access order, idempotency/replay scoping, stock/report access, append-only tables
metadata:
  type: project
---

The base trust map (proxy/XFF, CSRF, session, throttle, DB roles, M2 farm-scoped reads) is in the repo-root memory `/Users/cristian/Documents/GitHub/miraj-impex-v3/.claude/agent-memory/security-reviewer/trust_boundaries.md` ([[trust-boundaries]]).

Added in M3 (first write path, POST /api/entries/feed-daily):

- **Create access order**: decode (64 KiB) → `entries.checkWriter` (vet/manager/owner + CanSeeFarm) → FeedRefs DB read → Validate → service (checkWriter again) → runner. A nil farmId skips the farm check, but Validate then 400s. Every future create must call checkWriter before its first read.
- **Section ownership**: level 1 checks sectionId against the farm's sections. Under the farm lock the plan re-reads slots by farm_id and only inserts rows whose (section, sex) slot exists. Children of new forms need the same farm-scoped slot lookup.
- **Vet past day**: H6 is enforced server-side inside the tx (403 PAST_DAY_VET). The form state only shows a lockedReason.
- **Idempotency**: request_replay is keyed by clientRequestId. replay() compares user_id + route before returning bytes (else 409 STALE). SEC-1 (M3 round 1, minor; resolved in round 2): after the 7-day purge, alreadyEntered() now gives ALREADY_ENTERED + entryId only when entry.created_by = caller and entry.kind = route kind, else STALE (null author → STALE). Tests: TestCreateKeyPurgedOtherUserOrRoute. Every new route must pass a correct Create.Kind, or the purge fallback leaks across routes. Dry runs never use the key. Only 201 bodies are stored.
- **Append-only**: audit_log, live_event and request_replay are INSERT-only for il_app (0002_roles). Entries tests run on d.App = il_app. pg_notify carries only the live_event id.
- **Reports**: /api/reports/feed = RequireRole(viewer, manager, owner) + CanSeeFarm. farmId=all → all_farms only for the owner, else = any(FarmIDs). M3 Stoc has no prices. Watch M4: deliveries/cost must drop price keys for the viewer (PRICE_FORBIDDEN).
- **voiceTranscript**: ignored until M12 voiceToken (voiceSource returns manual/nil). Re-check when M12 lands.
- **Clock**: since T10 (COR-3), Runner.Create reads Now/Today after LockFarms, so the vet H6 and future-date checks use the instant after the lock wait.
- **period param**: one parser, calc.ParsePeriod (weeks n is int32, ≥1), shared by /api/table, the farm routes and /api/reports/feed.
- **Numbers**: num.Num JSON caps length and exponent (±18), so request decimals can't trigger shopspring rescale DoS.

**How to apply:** for later create/edit/delete routes, confirm the checkWriter-before-read order, farm-scoped child validation under the lock, and the replay user+route check on every key path (including purge fallbacks).
