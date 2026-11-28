---
name: trust-boundaries
description: Impex Line backend trust boundaries mapped in reviews (price visibility by role, farm scoping, undo tokens, audit/live data that later read routes must filter)
metadata:
  type: project
---

Trust boundaries of the Impex Line Go backend, as mapped in the M4 security review (2026-11-28).

- **Prices by role.** Owner and manager see prices; viewer and vet never get a price key (pricePerKg, totalLei, lei*, cost, pricedBy, "fără preț" tag). `reports` uses role-typed DTOs (FeedResponse vs FeedResponsePriced); the viewer's unpriced block is `{count}` only. Price writes (`/api/deliveries/*`) are owner-only: manager gets 403 PRICE_FORBIDDEN, others FORBIDDEN, checked in middleware before the body is decoded.
- **Price data at rest that later routes must filter.** `audit_log` rows `price_set` / `price_change` / `undo` (record_types `{pret}`) hold lei in `changes`, `old_values`, `new_values`. `live_event` price payloads carry ids/count/kg only (no lei). The future Jurnal (S17) read and the M11 live fan-out must strip `pret` data for vet and viewer.
- **Undo tokens.** HKDF purpose "il undo v1" is shared by the price undo (first field "price") and M10's setup undo. M10's verifier must reject a "price" kind token. Price tokens bind kind, audit row id and user id; they expire at local midnight. Replay is blocked by state: rows must still have priced_at = audit.at and priced_by = caller.
- **Farm scoping.** Non-owners are limited to `Principal.FarmIDs` (CanSeeFarm). An owner's unknown farm gets 404; for others it is FARM_FORBIDDEN. A series with no cycle on the caller's farms gets 404. Showing hidden farm *names* on the series page is a decision the user accepted (M2 SEC-1, M4 Q5), so do not re-flag it.
- **Livrări keyset cursor** is unsigned on purpose: the scope comes from the principal, so a forged cursor only moves within the caller's own farms.

- **SQL function `closed_series_on` (0003).** It is the single gate for "closed series = read-only" on price set/change/undo, the Livrare save and the Livrări setPrice flag. It is SECURITY INVOKER with no `set search_path`. That is fine because il_maint has no SELECT on cycle/series and functions are PUBLIC-executable. If a later migration makes it or any new function SECURITY DEFINER, it must pin search_path and REVOKE EXECUTE FROM PUBLIC. calc.FarmFilter is the pattern for farm-scoped aggregate reads.

**How to apply:** in later milestones (M5 delivery edit/delete, M8 Acasă unpriced, M9 Jurnal/export, M10 setup undo, M11 live), check each new response and read path against these lines first. See also [[past-findings]].
