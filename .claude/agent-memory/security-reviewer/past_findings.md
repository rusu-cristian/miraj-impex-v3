---
name: past-findings
description: Security findings raised in earlier Impex Line milestone reviews and how they ended, so they are not re-raised or forgotten
metadata:
  type: project
---

- **M4 round 1 (2026-11-28): approved.** One minor, SEC-1: the series page's `placement.placed` (SeriesGet) adds up placement events across every farm of the series, including farms hidden from the caller. In practice the placement farm is the only farm with placements, so nothing leaks today. The fix is to limit the sum to visible farms. Price isolation, owner-only price writes, undo token binding/expiry/replay, cursor and SQL all checked out clean.
- **M4 round 2 (2026-11-28): approved, SEC-1 resolved.** CalcSeriesPlacement now takes all_farms/farm_ids (calc.FarmFilter), and the series page passes the caller's farms. TestBuildPagePlacedOnVisibleFarms covers it. The T9 cleanup added nothing new.
- **Env note:** testcontainers can time out ("ready … matched 1 times, expected 2") when other reviewers start Postgres containers in parallel. Say so in the notes instead of claiming the tests failed.

**How to apply:** at a re-review, check SEC-1's status first. Don't re-raise the decisions the user accepted that are listed in [[trust-boundaries]].
