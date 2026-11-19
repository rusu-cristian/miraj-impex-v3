# FINAL4 · cross-screen check after round 2 (19.11.2026) · review only, nothing rebuilt

Checked: literal_scan.py (full) exit 0, every one of the 27 rows PORTED and clean; sex_scan.py 337 files, 0 plain signs.
Visible text of all 267 boards was extracted and searched for the 18.11 figures. Rendered at canvas width with Plex: every row's
default board, every phone default, S2 role boards, S4 tabs, and the salvat / eroare-la-salvare states of S6–S16.
The sample was compared side by side. Every board was measured in Chrome (overflow, tables, tabs).

## Clean
- **18.11 figures (check 1):** these match on every board that shows them:
  - ♀ 12 561 · ♂ 794, SAPT 59, age 407
  - deaths ♀ 3 · ♂ 0; sold ♀ 1 831 · ♂ 289
  - lay 41,1 % vs guide 53,3 %, week 40,4 %
  - 5 914 eggs; stock 31 665 / 1 781 kg
  - 4 unpriced deliveries, 18 160 kg (17.02 ♀+♂, 14.03 ♂, 12.11 ♀)
  - H7 "goală din 14.11"

  Where a board shows something else, it is a D39 example that follows its own story:
  - S16 după ștergere: 12 568
  - S16 corecție: 12 563 → 12 561
  - S10 before the sale: 14 392

  Not found on any board: 47,1 %, 1 834, Zile 408.
- **Weeks and codes (check 2):** every board counts weeks by age. SAPT 49 appears only on the S18 replay boards dated 07.10.2025
  (Seria 2025). No board shows SAPT 40/41 at Hîrbovăț, a "−18" offset or 408. Hîrbovăț uses only H5–H7; H1/H2/H3/H6/H5-1/H5-2
  appear only at Telița.
- **PeriodTables (check 3):** 53 tables checked (44 on boards, 9 in previews).
  - Every week key is `aria-expanded=true`, with no hidden day rows. The one exception is S4 Rezumat · pas Săpt., which is closed
    on purpose (grain Săpt., D171).
  - Where the groups exist, În viață comes first and Vânzare last.
  - Every typed group carries its `ip-ptable__g--<type>` tint; heads are centred.
  - Phone tables are 356 px wide (≤ 358).
  - The wide S4, S3, S14 and S9/S10 boards show their tables in full (no scroll overflow).
  - S3's % cumulat (7,9 % / 36,7 %, counted from placement) differs from S4's (5,7 % / 34,7 %, counted from arrival). Both tables
    give their basis in a note, so this is correct.
- **canvas.json:** 267 boards = 267 order entries = 267 files. No board overlaps another. Gaps within a row are ≥ 80 px and rows
  are 120 px apart.

## Findings

### Must
1. **Stale shots (check 7).** Fresh renders were compared pixel by pixel with design/shots.
   - **Old data:** S7.png, S7S-1…7 and S7-360 are from before RF02 and still show 16.11.2026, SAPT 40 and H1.1.
   - **Taken before RF12/RF13:**
     - S22-1…25 (except S22, S22-8, S22-16) and every S22-*-1280: last sign-in dates were changed after them.
     - S21-*-1280.
     - S14-*-1280 and S16-13-1280.
   - **Taken before the §72 icons or SX01:**
     - S2*, S6*, S8*, S9*, S10*, S12*, S13*: taken 17:12–19:30.
     - SX01 (22:41) changed boards after these shots: S3.png 0,29 % of pixels differ, S4T-1 0,29 %, S6 0,43 %, S9 0,56 %, S10
       3,6 % (also 1769 vs 1772 px tall), S2 1,5 % (873 vs 876), S11 0,08 %.
   - **Fix (FINAL4):** retake the shots of every row. Some render scripts rewrite their heights file (s3/s4t/s8/s11/s12/s14
     always; the rest with `--heights`), so restore those files with `git checkout`, or take the shots from a /tmp copy.
2. **Start-of-day counts are not labelled (S7, S6; this item is already in FINAL4's queue row).**
   - S7 shows H5 "♀ 8 547 · ♂ 665" and H5.1 "4 264 găini". S6 shows H5.2 "în viață ♀ 4 283".
   - S4 (tiles), S10 and S15 show the end-of-day counts on the same day: 8 544 / 4 263 / 4 281.
   - No board says "la începutul zilei".
   - **Fix (FINAL4, s7_build/s6_build):** add a caption "în viață la începutul zilei" and the same in the sr text of the hall and
     section counts. Then rebuild S7, S6 and S13 (S13 rewrites the S7 boards), and run S2 / S8 / S9 / S10 / S13 through s7s.
3. **S2 and S4 Rezumat alerts disagree (check 4, D169).**
   - S2 owner and viewer list 4 alerts: H5 lay, H6 lay, and "H7.1 / H7.2 hrană ♀ – lipsă" for 13.11.2026. That missing day is
     real: the workbook has no kg.
   - S4 Rezumat (owner) lists 3 alerts: H5, H6 and "4 livrări fără preț". It has no lipsă rows, because `s4s_build.alerts()` only
     takes MISSING for today.
   - **Recommended rule:** build one alert set per farm: today's guide deviations, lipsă rows in the same window as Lipsesc date
     (today after 12:00 plus the last 7 days), stock below its limit, and unpriced deliveries (roles that see prices only).
     - S2 shows that set for all farms. Lipsă rows go in "Lipsesc date" for vet/manager and in "De verificat" for owner/viewer;
       unpriced goes in its own block.
     - S4 Rezumat shows the farm's subset, all in "De verificat" (S4 has no Lipsesc block). On 18.11 that is 5 alerts.
     - The role only changes the link: "blocat · o completează managerul" for the vet, "Deschide secțiunea" otherwise.
   - **Fix (new small job, s4s_build):** move the lipsă window from s2_build into realdata (e.g. `missing(farm, window)`) and use
     it in both builds. Rebuild S4 (Rezumat and states), the S9/S10 desktops and S2 (byte-identical expected). Record the rule as
     a D169 note.

### Should
4. **Wide tables clipped beside a docked SidePanel (D167 4).** The S9/S10 desktops were widened to 1976, but the other boards with a
   panel were not:
   - S16 desktop ×4: S4 Păsări PeriodTable, 1070 px in a 718 region.
   - S13 desktop: S11 Consum, 1090 in 718.
   - S18 serie-nouă, desktop-scris/erori/varsta/refuzat: Serii table, 975–1123 in 718.
   - S18 telefon: two Serii tables, 995 / 899 in 356, cut at "CIC…".

   **Fix:** add these boards to board_widths.json and apply the `.s4s-wide` pattern in s16/s13/s18. The phone needs the s4ns
   phone-wide pattern. Then relayout.
5. **S4 phone tabs overflow.** The tab row is 367 px in 358 on S4 Rezumat telefon, Păsări telefon and Păsări telefon-filtre, so
   "Producție" is cut. S3's phone tabs fit (four tabs). **Fix:** S4NS / s4ns_build, or a system rule for phone Tabs.
6. **Save-error Banner position drifts.** S6, S7, S8, S13 and S16 put the §49 Banner at the end of the form, right above the
   SaveBar (D141/D173). S9 and S10 put it above the EffectPreview, so the Effect sits between the Banner and the keys.
   **Fix:** s9_build/s10_build (or record the exception: Effect last).
7. **The same record is formatted two ways.**
   - Hrană zilnică 18.11 reads "8 din 8 rânduri · 2 457,9 kg" on S2 vet ×4 and S15.
   - It reads "… · 2 458 kg" in S4 Ultimele înregistrări (×6) and the S9/S10 backdrops.

   **Fix:** use the decimal kg (as in the source record) in `s4s_build` event_row, then rebuild S4 and the S9/S10 desktops.
8. **S2 block title differs by role.** The StockReadout block is titled "Zile de hrană" for the manager (phone and desktop) and
   "Stoc hrană" for the owner and viewer (S11 calls it "Stoc"). **Fix (s2_build):** use one title.
9. **canvas.json `order` (check 6), confirmed.** These 9 boards sit at the end of `order` instead of with their rows:
   - S2-acasa--manager-ferma-goala
   - S4-ferma-rezumat--etichete, --pas-saptamana, --telefon
   - S4-ferma--pasari--telefon, --telefon-filtre
   - S18-serii--desktop-varsta, --telefon-varsta-eroare
   - S20-ghiduri--ouat-de-la-sapt-1

   No overlaps were found. **Fix (FINAL4 before publish, or a canvas_rows.py `relayout` step):** rewrite `order` as Main, then the
   boards sorted by row y, then x. This is an index-only change with no x/y moves.
10. **The Main index board is stale.** Main.dc.html still reads "Niciun ecran încă" on a canvas with 266 boards. **Fix (FINAL4
    before publish):** list the rows. Note that the canvas rows are not in screen-number order: S4, S7, S3, S11, S12, S13, S6, S9,
    S10, S2, S14, S8, S15–S22, S1.

### Nice
11. **Lay deviation unit:** "pct" on S2 / S4 / S14 / the S4 Readout, "puncte" on S6 and S15 (`−18,4 puncte`). Pick one (s6, s15).
12. **S6 zi trecută (17.11):** the HallCard reads "Ouat azi 44,0 %" on a past day. Use "Ouat" or "Ouat 17.11" (s6s).
13. **S13 default** is a 12.11.2026 replay (header 12.11, Azi selected, stock 16 140 → 25 140, which agrees with S4 Hrană). Its title
    lacks "reluare 12.11.2026", which S9/S10 have (s13 title).
14. **Mortality sign:** S4 Secțiuni tiles and Total fermă show "Mort. săpt. −10 · ♀ −10" as a removal, while the tables, S14 and
    the Stats show the count "10". Events (S2/S15 "♀ −3") are deltas, so their minus is right. Tiles and stats could show counts.
15. **Seria 2025 tables** (S4 ciclu ales, S3 seria închisă, S14 seria-2025) have no În viață or Vânzare group because the previous
    flock has no counts in the data. A caption line saying so would stop them looking like a D167 break.

## Not changed
No repo file was edited except this note, and nothing was committed. Renders went to /tmp/final4.
