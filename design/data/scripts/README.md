# Data scripts
Run with `python3 -I` from the repo root, passing paths (workbooks are untrusted data: openpyxl read-only, data_only).

Rebuild order (extracts dir = `design/data/extracts`):
1. `data01_build.py design/data/Seria_2026.xlsx <extracts>` — Telița side, weighings, guides (old workbook).
2. `data02_prev.py design/data/Seria_2026.xlsx <extracts>` — previous Hîrbovăț flock eggs.
3. `d36_patch.py <extracts>` — re-applies D36 (after data02_prev.py).
4. **`data03_build.py design/data/Raport_NOU_2026.xlsx <extracts>`** — Hîrbovăț from the new workbook (real to 18.11.2026, D151 week rule): rewrites farms-sections, cycles, transfers-seria2026, counts-latest, feed-deliveries, guide-laying; writes hirbovat-daily, hirbovat-weekly, illustrative-events-hirbovat, key-figures-18-11.md; adds age_days/sapt to the Telița daily files, age_days/sapt/guide_week (D157) to weighings-weekly and age_week to guide-growth; splits emptying removals into sales (D155) and computes lay % per D156. It reads the step 1–3 outputs (Telița rows), so run it last. Idempotent (a second run changes nothing). Prints a check log (age/Sapt per row, alive continuity, farm vs compartments, eggs, lay %, kg = g × birds, stock, Total rows, Productivitate/RealGhid).
5. **`data03_check.py <extracts>`** — consistency of the rebuilt files: alive = previous − deaths − sold + arrivals ± moves, mort_sheet = deaths + sold − arr − mv, sales only from 12.11 (D155), daily lay = eggs ÷ hens at day start and weekly = eggs ÷ hen-days (D156), weighing sapt and guide_week = sapt − 1 (D157), farm = Σ compartments, weeks = Σ days, counts, carve-outs and sale events, lei total; DATA04 (D168): one sale record per day × hall (14.11: H6.2, then H7), times rising within a day, record sections/totals = the sales block.

Retired: `data02_sample.py` (sample-hirbovat-2026.json, superseded — do not rerun after data03), `data02_check.py` (checks the old sample; now exits early with a RETIRED message on DATA03 extracts), `data02_review.py`/`data02_spot.py` (DATA02 review aids).
Helper: `data03_probe.py <xlsx> <sheet> <row_from> <row_to> [col_from col_to]` prints a bounded, non-empty window of a sheet.
