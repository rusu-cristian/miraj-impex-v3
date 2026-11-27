---
name: bug-patterns
description: Recurring bug patterns seen in the Impex Line Go backend (M0 review), worth re-checking in later milestones
metadata:
  type: project
---

- Check-then-act on DB state in services (login throttle: count failures, then hash, then record) is not atomic; concurrent bursts all pass the check. Look for the same shape in later milestones (replay store, unique-name checks, cycle overlap).
- Free-text request fields go straight into Postgres `text` parameters; JSON `\u0000` makes Postgres error 22021 and the handler returns 500. Any new public text input (login, names, notes) needs a NUL check or a mapped error.
- Reviewers are read-only (hook blocks cp/mkdir/file creation), so probe tests cannot be run; static reasoning only, confidence capped at about 70.
- M1 review (history import, calc, golden): code was clean and `go test -race ./...` was green with Docker available, so reviewers CAN run existing tests (only file creation and in-place edits are blocked). Remaining risk lives in views that read app_today(): v_section_sex_day's dense window feeds v_egg_day.hens_start and v_feed_consumption, so a day after app_today() silently gives 0 hens / 0 kg instead of an error. Re-check any new calc reader for this.
- Import pattern that held up: advisory xact lock plus exists check inside one tx, credentials printed only after commit. Residual: a failed write to SeedOut after commit loses the passwords and a rerun is refused.
- M2 review (periodtable, farms, calc rollups, JSON goldens): clean, `go test -race ./...` green, lint clean. Rollups are ratio-of-sums; the hen-day rule lives once in periodtable/sql.go. Residual spots to re-check later: (1) farms.BuildSummary does not apply periodtable's legacyCycle start shift, so a legacy (no flock count) past cycle gives alive 0/0 and SAPT 19-based weeks while table/page say null / SAPT 23; (2) calc.Paginate rejects `before` <= period.SaptFrom, so load-more only works if the client widens weeks:n/interval first (SPEC says it does); (3) reads span several statements on the pool, not one tx (accepted in plan); (4) summary tests with period=weeks/interval only assert HTTP 200.
- Tests here are strong on happy paths and exact messages (auth_test, matrix_test, codec_test); gaps tend to be concurrency and hostile-input cases, not missing assertions.
