# Canvas generators (design/canvas/scripts)

Each `sNN_build.py` writes one screen row of boards into `design/canvas/project/`. Run from this folder with the repo root
as first argument: `R=$(git rev-parse --show-toplevel); cd $R/design/canvas/scripts; python3 -I sNN_build.py $R [--canvas]`.

## Rules (FIN02G, `canvas_rows.py`)
- 2nd argument = heights file; omitted, it is the screen's own `sNN_heights.json` (s4/s7 take an int height; defaults 1874/3162).
- canvas.json is written only with `--canvas` (scripts with no canvas step accept and ignore it). Imported builds never write it.
- Under `--canvas` every existing entry keeps its x and y (w/h/title refreshed in place). A row with no entry yet goes below the
  lowest row (row y + tallest h + 120). A run that adds a board lays the row out again (80 px gaps). Overlapping boards stop the save.
- At HEAD every command below rebuilds its row, canvas.json and the rows it rewrites through imports byte-identical (checked
  row by row in a copy, FIN02G). The HANDOFF caveat on S3/S10/S11/S14 applies only to old script versions: use these scripts.

## Rebuild per row (in this order; whole canvas: run the rows top to bottom)
| Row | Commands (each `python3 -I <script> $R …`) |
|---|---|
| S4  | `s4s_build.py`, `s4_build.py`, `s4t_build.py` (no canvas step) |
| S7  | `s7_build.py`, `s7s_build.py` (no canvas step) |
| S3  | `s3_build.py --canvas` |
| S11 | `s11_build.py` (no canvas step) |
| S12 | `s12_build.py` (no canvas step) |
| S13 | `s13_build.py` (no canvas step) |
| S6  | `s6_build.py`, `s6s_build.py` (no canvas step) |
| S9  | `s9_build.py` (no canvas step) |
| S10 | `s10_build.py --canvas` |
| S2  | `s2_build.py --canvas`, `s2s_build.py --canvas` |
| S14 | `s14_build.py --canvas`, `s14s_build.py --canvas` |
| S8  | `s8_build.py --canvas`, `s8s_build.py --canvas` |
| S15 | `s15_build.py --canvas`, `s15s_build.py --canvas` |
| S16 | `s16_build.py --canvas`, `s16s_build.py --canvas` |
| S17 | `s17_build.py --canvas`, `s17s_build.py --canvas` |
| S18 | `s18_build.py --canvas`, `s18s_build.py --canvas` |
| S19 | `s19_build.py --canvas`, `s19s_build.py --canvas` |
| S20 | `s20_build.py --canvas`, `s20s_build.py --canvas` |
| S21 | `s21_build.py --canvas` |
| S22 | `s22_build.py --canvas`, `s22s_build.py --canvas` |
| S1  | `s1_build.py --canvas` |
S4/S7/S11/S12/S13/S6/S9 entries were placed by hand: they never change unless edited in canvas.json.

## Never run for a row
- `sNN_render.py` (any flag) as part of a rebuild: it re-measures in Chrome (fallback fonts) and rewrites `sNN_heights.json`
  (s3/s4t/s8/s11/s12/s14 always, the rest with `--heights`/`--measure`). Run it only to re-measure, then rebuild with `--canvas`.
- A heights file of another screen as 2nd argument, or a height other than 1874 (s4) / 3162 (s7).
- Scripts from older commits, and `s3_canvas.py` (removed: wrote canvas.json on every run; now `s3_build.py --canvas`).

## Import chain (a change re-runs every row downstream; imported rows are rewritten byte-identical)
- s4s_build (desktop shell) ← s4, s4t, s3, s11, s2, s9, s10. s7_build ← s7s, s6; s7s ← s2, s8, s9, s10, s13; s6 ← s6s.
- s12 ← s11; s13 ← s7s + s11 (rewrites the S7 boards); s14 ← s4t + s11; s16 ← s4t.
- s2s ← s2; s14s ← s14 + s2; s15 ← s2s; s15s ← s15; s16s ← s16; s8s ← s8.
- s17 ← s14s; s17s ← s17; s18 ← s17 + s3; s18s ← s18; s19 ← s18; s20, s21, s22 ← s19; s19s/s20s/s22s ← their base.
- s20_build, s21_build and s22_build import s19_build, which rewrites /tmp/s19/boards.json with only the 10 base S19 boards (the 12 S19S boards drop out): re-run `s19s_build.py` after them before any `s19_render.py`. Renders only; repo files are unaffected.
- s1, s19, s20, s22 also exec the top of design/system/scripts/sys13/sys10/sys11/sys12_build.py.
- sstate_css.py ← s2s, s6s, s7s, s8, s9, s10, s13, s14s, s15s, s16, s18 (s15 and later get it through these too).
- canvas_rows.py: args/row_y/place/save for every build above.
- skel.py (FIN11G, additions.css §54): skeleton markup and table column shares (readout, track, field, line, widths) ← s2s, s6s, s7s, s8s, s11 (S12/S13/S14 through it), s14, s16s, s18 (+ s18s), s19 (+ s19s, s21, s22), s20 (+ s20s), s21, s22. A change re-runs those rows.
