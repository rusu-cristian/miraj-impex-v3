# Canvas generators (design/canvas/scripts)

Each `sNN_build.py` writes one screen row of boards into `design/canvas/project/`. Run from this folder with the repo root
as first argument: `R=$(git rev-parse --show-toplevel); cd $R/design/canvas/scripts; python3 -I sNN_build.py $R [--canvas]`.

## Rules (FIN02G, `canvas_rows.py`)
- 2nd argument = heights file; omitted, it is the screen's own `sNN_heights.json` (s4/s7 take an int height; defaults 2944/2426).
- canvas.json is written only with `--canvas` (scripts with no canvas step accept and ignore it). Imported builds never write it.
- Under `--canvas` every existing entry keeps its x and y (w/h/title refreshed in place). A row with no entry yet goes below the
  lowest row (row y + tallest h + 120). A run that adds a board lays the row out again (80 px gaps). Overlapping boards stop the save.
- Real data (RF00/RF01, D152, D158): rows move to `realdata.py` one RF job at a time. A row not yet ported stops on the DATA03
  extracts (its own sample-era asserts); do not rebuild it before its job. The check per job is `literal_scan.py` (clean on
  the ported rows) + idempotence (run twice → byte-identical), not byte-identity with the old output.
- After a row's `--canvas` rebuild run `python3 -I canvas_rows.py relayout $R --canvas` (re-stacks the rows below when its
  tallest board changed; x kept, y = previous row y + tallest h + 120; overlap check).

## Rebuild per row (in this order; whole canvas: run the rows top to bottom)
| Row | Commands (each `python3 -I <script> $R …`) |
|---|---|
| S4  | `s4s_build.py --canvas`, `s4_build.py --canvas` (Rezumat: h/title refreshed in place), `s4t_build.py --canvas` (the 4 tabs: h/title in place, S4N) |
| S7  | `s7_build.py --canvas`, `s7s_build.py --canvas` (h/title refreshed in place, RF02) |
| S3  | `s3_build.py --canvas` |
| S11 | `s11_build.py --canvas` (h/title refreshed in place, RF03) |
| S12 | `s12_build.py --canvas` (h/title refreshed in place, RF03) |
| S13 | `s13_build.py --canvas` (h/title refreshed in place, RF03; rewrites the S7 boards byte-identical) |
| S6  | `s6_build.py --canvas`, `s6s_build.py --canvas` (h/title refreshed in place, RF04) |
| S9  | `s9_build.py --canvas` (h/title refreshed in place, RF05) |
| S10 | `s10_build.py --canvas` (h/title in place; tall default re-measured by `s10_render.py --heights`, RF05) |
| S2  | `s2_build.py --canvas`, `s2s_build.py --canvas` |
| S14 | `s14_build.py --canvas`, `s14s_build.py --canvas` |
| S8  | `s8_build.py --canvas`, `s8s_build.py --canvas` (h/title in place; heights from `s8_render.py s8_heights.json`, min-height released, RF07) |
| S15 | `s15_build.py --canvas`, `s15s_build.py --canvas` |
| S16 | `s16_build.py --canvas`, `s16s_build.py --canvas` |
| S17 | `s17_build.py --canvas`, `s17s_build.py --canvas` |
| S18 | `s18_build.py --canvas`, `s18s_build.py --canvas` |
| S19 | `s19_build.py --canvas`, `s19s_build.py --canvas` |
| S20 | `s20_build.py --canvas`, `s20s_build.py --canvas` |
| S21 | `s21_build.py --canvas` |
| S22 | `s22_build.py --canvas`, `s22s_build.py --canvas` |
| S1  | `s1_build.py --canvas` |
S4/S7/S11/S12/S13/S6/S9 entries were placed by hand: their x never changes; y only through `canvas_rows.py relayout`, and the
S4 Rezumat entries' h/title through `s4s_build.py`/`s4_build.py --canvas` (RF01), the S7 ones through `s7_build.py`/`s7s_build.py --canvas` (RF02),
the S11/S12/S13 ones through `s11_build.py`/`s12_build.py`/`s13_build.py --canvas` (RF03), the S6 ones through
`s6_build.py`/`s6s_build.py --canvas` (RF04; s6_render.py reports the natural height with the min-height released → s6_heights.json),
the S9 ones through `s9_build.py --canvas` (RF05).

- Heights (RF03 fix): `s4s_build.write` (the desktop board writer every 1440 row uses) pins the board's `min-height` to the stored h
  only in the `.dc.html`; its `/tmp/s4s/<stem>.html` render wrapper keeps just the 1024 viewport floor, so `sNN_render.py` measures
  the content's natural height instead of reading the stored h back. Re-measure → write `sNN_heights.json` → rebuild `--canvas` → relayout.

## Never run for a row
- `sNN_render.py` / `s4s_render.py` (any flag) as part of a rebuild: it re-measures in Chrome (fallback fonts) and rewrites `sNN_heights.json`
  (s3/s4t/s8/s11/s12/s14 always, the rest with `--heights`/`--measure`). Run it only to re-measure, then rebuild with `--canvas`.
- A heights file of another screen as 2nd argument, or a height other than 2944 (s4) / 2426 (s7).
- Scripts from older commits, and `s3_canvas.py` (removed: wrote canvas.json on every run; now `s3_build.py --canvas`).

## Import chain (a change re-runs every row downstream; imported rows are rewritten byte-identical)
- realdata (RF01) ← s4s_build (and every row through it). s4s_build (desktop shell) ← s4, s4t, s3, s11, s2, s9, s10. s7_build ← s7s, s6; s7s ← s2, s8, s9, s10, s13; s6 ← s6s.
- S4N: s4t_build ← s4s_build + design/system/scripts/sys14_build.py (PeriodTable builders, imported; it writes its preview only when run itself); s4_build ← s4t_build (Rezumat table).
- s12 ← s11; s13 ← s7s + s11 (rewrites the S7 boards); s14 ← s4t + s11; s16 ← s4t.
- s2s ← s2; s14s ← s14 + s2; s15 ← s2s; s15s ← s15; s16s ← s16; s8s ← s8.
- s17 ← s14s; s17s ← s17; s18 ← s17 + s3; s18s ← s18; s19 ← s18; s20, s21, s22 ← s19; s19s/s20s/s22s ← their base.
- s20_build, s21_build and s22_build import s19_build, which rewrites /tmp/s19/boards.json with only the 10 base S19 boards (the 12 S19S boards drop out): re-run `s19s_build.py` after them before any `s19_render.py`. Renders only; repo files are unaffected.
- s1, s19, s20, s22 also exec the top of design/system/scripts/sys13/sys10/sys11/sys12_build.py.
- sstate_css.py ← s2s, s6s, s7s, s8, s9, s10, s13, s14s, s15s, s16, s18 (s15 and later get it through these too).
- canvas_rows.py: args/row_y/place/save for every build above.
- skel.py (FIN11G, additions.css §54): skeleton markup and table column shares (readout, track, field, line, widths) ← s2s, s6s, s7s, s8s, s11 (S12/S13/S14 through it), s14, s16s, s18 (+ s18s), s19 (+ s19s, s21, s22), s20 (+ s20s), s21, s22. A change re-runs those rows.

## Real data and checks (RF01)
- `realdata.py`: the one data module (pure data + formats, no HTML, imports no sNN script). Reads design/data/extracts/ of its
  own repo copy on import. Today `TODAY` = 18.11.2026, `WK` = SAPT 59; `age_days(d)`, `sapt(d)` (D151: age // 7 + 1),
  `week_dates(k)` (Tue–Mon); formats `n` (U+202F thousands, decimal comma, true minus), `sgn`, `pct` (D46), `rnd` (D82),
  `dd`/`ddm`/`ddt` ("18.11.2026, 08:15") and `drange` (D129: "17.11–18.11.2026"), `SAPT(k)`/`sapt_txt(k)`. Farms/sections:
  `FARMS`, `SECTIONS[farm]`, `HALLS`, `HALL_SECS`, `OLD2NEW`/`new_code` (H1→H5 …, D152). Hîrbovăț per day/section/sex:
  `sd`, `alive(sec, sx, d)` (end of day), `alive_start`, `alive_hall`, `alive_farm`, `deaths(a, b, sec, sx, hall)` (real deaths)
  vs `sold(…)`/`sales(a, b)` (Vânzare, D155), `empty_since(hall)`, `has_birds`; `day(d, hall)` (farm/hall rows), `eggs`,
  `lay_day(d, hall)` (eggs ÷ hens at the start of the day, D156); weekly `week(k, hall, sec)`, `lay_week(k)` (eggs ÷ start-of-day
  hen-days, D156/D159), `WEEKS`. Guides by age week: `guide_row`, `guide_lay`, `guide_feed`, `guide_mort`, `guide_growth`,
  `weigh_guide(k, sx)` (= week k − 1, D157), `GUIDE_LAY_WEEKS`. Feed: `feed_g(d, sx, sec)` (norm g/bird; farm = kg ÷ birds),
  `feed_kg`, `stock(d)`, `stock_days(sx)` (days left on the 3-day mean, run-out date), `DELIVERIES`, `UNPRICED`, `FEED_TYPE(S)`.
  Records: `COUNTS` (counts-latest), `ARRIVED`, `TR`, `ILL` / `EVENTS` / `WEIGHINGS_HIR` (illustrative; a `sale` event = one form, DATA04/D168: `sections` [{section, F, M}], totals `F`/`M`, `section` = first compartment for old readers; `SALE_EVENTS`, `sale_codes(e)` → "H6.1, H6.2"), cycles `C26`, `C25`,
  `CTEL`, `CYCLE_START`, `EMPTYING_SINCE`; Telița `PLACEMENT_ROWS`, `WEIGHINGS_TELITA`, `MORT_TELITA`, `FEED_TELITA`; the previous
  flock `PREV_DAILY`/`PREV_WEEKLY`/`PREV_META` with halls mapped to H5–H7 (D158c; use `age_week`). Asserts only the stable key
  figures (♀ 12 561 · ♂ 794, 5 914 eggs, SAPT 59, arrivals ♀ 26 883 · ♂ 3 024); figures still moving are computed, never copied.
  `s4s_build` re-exports its names (`TODAY`, `WK`, `SECTIONS`, `HALLS`, `alive`, `tot`, `CL`, `GL`, `TR`, `n`, …); `s4s_build.S`
  (the superseded sample) is a lazy legacy alias for rows not yet ported only.
- `literal_scan.py $R [--row S4R …] [FILE …]`: lists file:line of old literals (H1.x/H3.x/H4.x and Hîrbovăț H1/H3/H4, 16.11.2026,
  SAPT 40/41, 10.02.2026, the sample's figures, sample file, cycle_week, age − 18, "47 săpt."); exit 1 on a hit in a PORTED row
  (or in any row/file named). Allowed: Telița codes, Seria 2025 weeks, script lines marked `literal-ok`, its ALLOW list. Each
  RF job adds its row to PORTED.
- `canvas_rows.py relayout $R [--canvas]`: see Rules. `s4s_render.py [--heights] [--w 1280] [stem …]`: re-measures the S4
  Rezumat boards (shots S4.png, S4S-n.png); `--heights` updates s4s_heights.json and prints the s4_build int (2944).
