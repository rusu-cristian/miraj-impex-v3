---
name: backend-build-spec-drift
description: Recurring spec-review gap in the Go backend milestones - build-time decisions not written back to SPEC.md, and CI/actionlint criteria that cannot be proven pre-push
metadata:
  type: project
---

M0 review (2026-11-27): the builders were strong on tests and criteria coverage (every plan criterion had a test; gates, TEST_DATABASE_URL run and `make e2e` all passed). The recurring gap is SPEC.md drift:
- Decisions made during the build (new `Issue.code` values, using encoding/json/v2 instead of the spec's `encoding/json` text, not using http.CrossOriginProtection, reusing a copy string for a second case) were recorded in NOTES-verify.md or the status log but SPEC.md was not updated, although backend/CLAUDE.md says to update SPEC in the same commit once agreed.
- New enum-like JSON values (issue codes) were added without a user answer.
- CI criteria ("actionlint passes", "run green on team/backend") cannot be proven before the push; flag them as open instead of approving silently.

**How to apply:** in later milestones, diff SPEC.md against the answers list, grep new `Code*`/`Issue*` constants against the SPEC code lists, and check each "verify at implementation" outcome that contradicts SPEC text has a matching SPEC edit or a question to the user.
