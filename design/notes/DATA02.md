# DATA02 · Hîrbovăț extracts (16.11.2026)
- Built `eggs-hirbovat-prev-daily.json` / `-weekly.json` (real, 888 hall-days, 4 606 935 eggs), `sample-hirbovat-2026.json` (illustrative, seed 20261116) and `counts-latest.json`; README rows + "Sample (illustrative)" paragraph appended. Scripts: `/home/claude/scripts/data02_prev.py`, `data02_sample.py`, `data02_check.py` (0 errors).
- Previous flock: no flock count anywhere, so hens alive and % lay are null. Depopulation sits in the mortality columns from 15–20.01.2026 (flagged). datefinal wk 63–64 rows don't match the daily data (kept in meta).
- Previous-flock week: `cycle_week` = age − 18 (D26, matches the guide); D36 moves the inferred start to 07.03.2025 (age wk 19), so the data-model formula gives the same index; data starts 04.04.2025 = cycle wk 5. The old `week_from_start` is renamed `data_week` (age − 22, extra index only).
- 16.11.2026 (sample): ♀ 25 252 · ♂ 2 685; cumulative mortality ♀ 5,54 % · ♂ 11,93 %; lay 74,0 % vs guide 74,1 %; H3 70,1 % in sapt. 40 (−4,0 pts, alert).
- Feed: Breeder 2 28 252 kg · 7,1 days; Cocoși 1 650 kg · 3,9 days (warn); 4 unpriced deliveries (04.11, 06.11, 10.11, 14.11 · 76 107 kg); C5 toast stock 31 450 kg matches; the C2 "−1 240 kg" warning is not true today.
- 16.11 grams: H3.2 ♂ and H4.2 ♀ missing; H1.2 ♂ 172 g (+11 %). No mortality alert today (capped under 3× the 7-day average).
- Culls use a "cull" record with a reason. BIRD_EVENT type `cull` + `reason?` added to data-model.md (D36). Sexing errors are recorded as −1 on the recorded sex, as Telița does.
- H1.1 has 0 deaths from 01.03 to 10.03, so the real 10.03 row (508,5 / 48,8 kg) matches exactly.
