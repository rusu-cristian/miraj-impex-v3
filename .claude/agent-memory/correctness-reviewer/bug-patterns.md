---
name: bug-patterns
description: Recurring bug patterns seen in the Impex Line Go backend (M0 review), worth re-checking in later milestones
metadata:
  type: project
---

- Check-then-act on DB state in services (login throttle: count failures, then hash, then record) is not atomic; concurrent bursts all pass the check. Look for the same shape in later milestones (replay store, unique-name checks, cycle overlap).
- Free-text request fields go straight into Postgres `text` parameters; JSON `\u0000` makes Postgres error 22021 and the handler returns 500. Any new public text input (login, names, notes) needs a NUL check or a mapped error.
- Reviewers are read-only (hook blocks cp/mkdir/file creation), so probe tests cannot be run; static reasoning only, confidence capped at about 70.
- Tests here are strong on happy paths and exact messages (auth_test, matrix_test, codec_test); gaps tend to be concurrency and hostile-input cases, not missing assertions.
