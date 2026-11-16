# Data extracts (Seria 2026 workbook)

Source: `design/data/Seria_2026.xlsx`. Built by DATA01 with a script kept outside the repo (`/home/claude/scripts/data01_build.py <xlsx> <out_dir>`).

**Last real date in the workbook: 28.02.2026** (one Hîrbovăț feed row on 10.03.2026). Anything later is sample/illustrative (D27). Guide and template rows that run to 2027 are not actuals.

## Conventions
- Dates are ISO (`2026-02-10`). Numbers are JSON numbers. Sex is `"F"` (♀) / `"M"` (♂).
- Section codes are Romanian, as in the workbook: Telița `H1 H2 H3 H6 H5-1 H5-2`; Hîrbovăț halls `H1 H3 H4`, compartments `H1.1 … H4.2` (`parent_id` = hall).
- Each file has a `meta` block: source sheet/columns, range, what is real vs illustrative, and anomalies. Anomalies are reported, not fixed.
- Growth: cycle week = age week. Laying: cycle week = age week − 18 (D26). Guide week = `floor((date − start) / 7) + 1`.
- `null` = not in the workbook. Blank daily mortality cells = 0.
- Illustrative names: guides "Ross 308 creștere" / "Ross 308 ouat" (D22); feed types (D30); series name "Seria 2025".

## Files
| File | What | Source | Range | Real / illustrative |
|---|---|---|---|---|
| `farms-sections.json` | farms, guides, sections (hall → compartment) | Telita r2, Hrana Hîrbovăț r1, Sheet9 r2, H-sheet names | static | real codes; guide names illustrative |
| `placement-seria2026.json` | placement per section and sex (♀ 28 080 · ♂ 3 510) | Telita r3 (H-sheets r4 for reference) | 07.10.2025 | real |
| `transfers-seria2026.json` | Telița → Hîrbovăț outs, Hîrbovăț arrivals per section (♀ 26 765 · ♂ 3 060), date-matched links, internal moves, unexplained corrections | Telita "Corectie" R–W; Sheet9 | 05.11.2025 – 28.02.2026 | real; links inferred by date |
| `mortality-telita-daily.json` | daily mortality, sexing errors, corrections, alive (sheet and running) per section | Telita D–AC | 07.10.2025 – 28.02.2026 | real |
| `feed-deliveries.json` | 13 deliveries, 263 740 kg, 2 031 768 lei, lei/kg, priced flag | Hrana Telita L, X, Y | 07.10.2025 – 10.02.2026 | real kg/lei; feed type illustrative (D30) |
| `feed-daily-telita.json` | g/bird and kg per section, kg per sex, stock | Hrana Telita F–K, N–V | 07.10.2025 – 26.02.2026 (g from 09.10) | real |
| `feed-types.json` | Starter / Grower / Pre-ouat (Telița), Breeder 2 ♀ / Cocoși ♂ (Hîrbovăț), with the workbook note each maps from | Hrana Telita col Y | notes 24.10.2025 – 20.01.2026 | illustrative names |
| `weighings-weekly.json` | weekly avg g, CV %, uniformity %, standard g, feed standard, weekly mortality per section | H1, H2, H3, H6, H51., H52 | age wk 1–20 (14.10.2025 – 24.02.2026) | real; H3/H6 re-dated by age week |
| `guide-growth.json` | growth guide wk 1–25: feed_g ♀/♂, body_weight_g ♀/♂ (wk 1–23) | Hrana Telita "Ghid"; H-sheet "Standard" | age wk 1–25 | real figures, illustrative name |
| `guide-laying.json` | laying guide by age and cycle week: feed_g ♀/♂, lay_pct (♀) | Hrana Hîrbovăț "Ghid"; Sheet9 col P | age wk 19–65 (lay 25–63) | feed real; lay_pct illustrative (D29) |
| `cycles.json` | series and cycles: Telița 2026 (closed, 21 wk), Hîrbovăț 2026 (open from 10.02.2026, sapt. 40 on 16.11.2026), Hîrbovăț previous flock (07.03.2025 – 28.01.2026, 47 wk, start inferred at age wk 19 per D36; data from 04.04.2025 = cycle wk 5) | derived | – | dates real except Hîrbovăț 2026 start (D21/D25); "Seria 2025" illustrative |

<!-- DATA02 appends below: eggs-hirbovat-prev-daily.json, eggs-hirbovat-prev-weekly.json, counts-latest.json, sample-hirbovat-2026.json -->
| `eggs-hirbovat-prev-daily.json` | previous Hîrbovăț flock per hall and day: Jos, Cat. 1–3, Incubabile, total (D18), egg weight, hatching %, mortality per compartment ♀/♂, depopulation flag; hens alive and % lay null (no flock count in the workbook) | import date (DataHirbovat mirrors it) | 04.04.2025 – 28.01.2026 (H1, H3 to 22.01) | real |
| `eggs-hirbovat-prev-weekly.json` | the same by age week and cycle week (age − 18), per hall and all halls; datefinal cross-check (wk 63–64 rows of datefinal are wrong) | derived from the daily file | age wk 23–65 | real |
| `sample-hirbovat-2026.json` | Seria 2026 at Hîrbovăț: daily mortality, culls, sexing errors, alive, grams and kg per section × sex; eggs per hall; weekly lay vs guide; weighings; deliveries, stock, days left; 16.11 allocation state; 7-day event log and one audit entry | generated, seed 20261116 (`/home/claude/scripts/data02_sample.py <extracts_dir>`) | 01.03.2026 – 16.11.2026 | **illustrative** (`meta.illustrative: true`) |
| `counts-latest.json` | birds alive per section and sex on 16.11.2026 (Hîrbovăț sample; Telița 0, cycle closed 28.02.2026), hall/farm/Seria totals, cycle mortality % | sample + real baseline (arrivals 28.02.2026) | 16.11.2026 | Hîrbovăț illustrative; Telița real |

### Sample (illustrative) after 28.02.2026
Every Seria 2026 figure at Hîrbovăț dated 01.03.2026 – 16.11.2026 is sample (D27), except H1.1's grams on 10.03.2026 (♀ 113 g · ♂ 98 g, real). On screen that means: birds alive (♀ 25 252 · ♂ 2 685 = 27 937 on 16.11.2026) and every count, total or % built on them; cumulative mortality (♀ 5,54 % · ♂ 11,93 %); daily mortality, culls and their reasons, the 2 sexing errors; eggs by category, total eggs, % lay (74,0 % farm, H3 70,1 % vs guide 74,1 % in sapt. 40), hatching %, egg weight; g/bird, kg per section, the 16.11 allocation states (H3.2 ♂ and H4.2 ♀ missing, H1.2 ♂ 172 g +11 %); weighings (avg g, CV %, uniformity %, sample size); all Hîrbovăț deliveries, their prices and the "4 livrări fără preț" (76 107 kg); stock and days left (Breeder 2 28 252 kg · 7,1 zile; Cocoși 1 650 kg · 3,9 zile); the C5 toast stock 31 450 kg; the event log, the 15.11 correction and the names Dr. Ceban and Ion Rusu; every alert built on these. Real: guide feed grams (the % lay guide is itself illustrative, D29), arrivals ♀ 26 765 · ♂ 3 060, all Telița data and the previous flock's eggs.
