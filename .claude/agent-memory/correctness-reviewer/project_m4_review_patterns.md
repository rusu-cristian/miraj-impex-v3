---
name: m4-review-patterns
description: What held up and where to look when reviewing Impex Line backend milestones (prices, deliveries, reports, series page)
metadata:
  type: project
---

M4 round 1 (base 433a1ae): code was clean; `go test -race ./...`, gofmt, staticcheck, `sqlc diff` and the
time.Now / ARC-6 greps all passed. Only one minor data-scope issue found.

Patterns worth re-checking in later milestones:
- Aggregates that span farms (series placement `SeriesGet.placed_f/m`, series totals) are summed over ALL farms,
  then shown to a caller who may see only some farms. Check every "whole-series" figure against role scope.
- Price undo matches on `priced_at = audit_log.at` (same truncated-to-microsecond clock instant). Safe in prod, but under
  `FixedClock` a PATCH in the same instant keeps the match; tests that mix Set, Change and Undo must advance the clock.
- Cost total is omitted unless the rows cover every farm of the series (`meta.SeriesFarms`); a farm with only unpriced
  deliveries has no row, so no total and its kg is not in `excludedKg`. By design (Q6), not a bug.
- Level-1 checks read refs outside the tx (`store.Q()`); level-2 re-reads after `LockFarms` + `for update`. Pattern is
  sound because the locked columns (kg, farm) are never edited in M4; revisit when M5 adds delivery edit.
- Wire numbers go through `num.Num` (4 dp half away from zero), so "full precision" values are double-rounded when the
  client rounds again to 2 dp. Pre-existing design (§5.15), not new per milestone.
