---
name: m5-review-patterns
description: How M5 (Zilnic, entry change routes) matched its brief in round 1, and where builders deviate on this project
metadata:
  type: project
---

M5 round 1 (2026-11-28): every plan task had code plus a test; `make lint` and `go test -race -p 1 ./...` green against one postgres:16.16-bookworm container (TEST_DATABASE_URL superuser DSN, port 55999). TestGolden = 35 pass / 3 skip, as the brief says.

Recurring deviations are small and documented, not missing features:
- Builders replace a "read from the seed" source with a constant when the bare import (Verify inside the import tx) lacks the data (golden 37 old value 1).
- Builders widen infra files slightly (backend.yml: `if:` guards on old jobs and a changed concurrency group, to add the nightly job) and widen M3 reads when a new rule needs it (feed form: unserved kinds get no actions).

**Why:** the plan lists exact tasks and "done when" figures; the builder here is careful and mirrors plan figures in tests (4 284 bound, 12 568, 4 283, Toast strings).
**How to apply:** for this project, check scope creep in shared files (workflow yml, M3/M4 queries, SPEC removals) and the unrunnable gates (`make docker`, `make e2e`, CI) rather than hunting for missing tests; say plainly that docker/e2e were not run.
