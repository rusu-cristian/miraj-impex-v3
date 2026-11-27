---
name: recurring-spec-gaps
description: Recurring small gaps seen in Impex Line backend milestone reviews (M3 onward) and review-mechanics notes
metadata:
  type: feedback
---

Builders on this project cover plan acceptance criteria thoroughly (tests named per criterion, SPEC edited in the same
tree with *(proposed)* markers, NOTES-verify entries). The gaps that keep showing up are small:

- New read routes are proven on the primary seeded case only (Hîrbovăț open cycle). Other seeded cycles/farms
  (Telița growth cycle, Seria 2025 without flock count) are not run through the new route, so null-list and
  nullable-field contract checks (`TestFarmReadsOnSeededData`, `nullableFields`) are unproven for them.
- NOTES-verify wording goes stale after a later SPEC edit (e.g. "the SPEC sketch calls it timex.Version" after the
  sketch was renamed to `version`).
- Extra copy beyond the user's Q answers (plurals, "sub limită", purged-key sentence) is added; it is marked
  *(proposed)* in SPEC, so only needs the lead to confirm with the user.

**Why:** M3 round 1 review (2026-11-27) found no blocking spec gap; these were the only deviations.
**How to apply:** In future milestones check "every seeded cycle/farm through the new route" and NOTES/SPEC
consistency early; do not spend time re-verifying the per-criterion tests, they were solid.

Mechanics: reviewers are read-only (hook blocks Write/heredoc outside agent-memory), so `go test -overlay` probes
cannot be created; run existing tests only. Full gate (`go build`, `make lint`, `go test -race ./...`) takes ~2 min
with Docker/Postgres running.
