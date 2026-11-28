---
name: project-review-patterns
description: Recurring spec-review patterns for the Impex Line backend milestones (what builders do well, where gaps tend to hide)
metadata:
  type: project
---

M4 round 1 (2026-11-28): builders delivered every acceptance criterion with a named test; 0 blocking findings. Plan briefs are very detailed, so most review time is best spent on:
- Scope drift hidden in "thin export" tasks (M4: `farms.BuildCycleStats` was a refactor of `BuildSummary`/`readSummary` signatures, guarded only by the byte-equal summary golden file).
- Doc/ledger drift: `.claude/team/backend/review-at-end.md` and `status.md` can disagree with the code (M4: golden 35 logged as "implemented in M4" but code keeps it in `deferredRows`).
- Gates the reviewer cannot rerun cheaply (`make docker`, `make e2e`): take from the lead's gate line, say so in notes.

**Why:** the plan's acceptance criteria are mirrored almost 1:1 by tests, so gaps show up at the edges (ledger files, refactors outside the named files), not in the core criteria.
**How to apply:** check `git diff --stat` for files outside the plan's task table first; cross-check review-at-end.md claims against golden.go / tests.

Tooling: reviewers' Bash hook blocks redirects into the scratchpad; write test output to /tmp/ instead.
