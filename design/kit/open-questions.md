## 6. Open questions (most important first)

1. **One series per farm:** can a farm hold birds from two series at once, or must it be emptied first? The default lets the manager start a new cycle manually.
2. **Seria 2026 history:** how is it imported (cycles included), and from which date is the app the source of truth?
3. **Real guide tables:** which guides apply (one growth and one laying), and is the Hîrbovăț productivity guide available?
4. **Egg total formula:** is "Jos" excluded?
5. **Alert thresholds:** confirm mortality 3× the 7-day average, weight ±5%, lay ±3 points, feed ±10% and under 5 days of stock, missing entries from 12:00.
6. **Feed types:** the list of feed types, and which are for hens and which for roosters.
7. **Tone:** informal "tu" or formal "dumneavoastră"?

**Round 2 (19.11.2026, Raport_NOU_2026.xlsx; setup-check-2 §5, D153, DATA03)**

8. **Incubation batches:** the `Incubare` (batches 1–118, 20.03–22.11.2026: eggs set, set/hatch dates, chicks, hatch % vs guide, egg age per hall) and `BD Incubator` (2022–12.2026) sheets have no table in the data model. Not modelled this round (D153): do they belong in the app, and who enters them?
9. **Egg categories Cat. 1–3:** the form sends counts (`Raspunsuri`), `Masa` gives %; `EGG_PRODUCTION` has cat1–cat3 but no screen shows them this round (D153). Counts or %, and is Total ouă still Cat. 1–3 + Incubabile (D18)?
10. **Two lay % guides:** RealGhid col B and Productivitate col I disagree (SAPT 56: 56,9% vs 61,1%; RealGhid SAPT 53 typed "605", read as 60,5%). The app uses **RealGhid** (DATA03); confirm, and whether eggs/hen, hatching eggs/hen (week and cumulative) and the hatchability guide should join `GUIDE_WEEK`.
11. **Feed type at Hîrbovăț:** deliveries carry kg and lei per ♀ feed / ♂ feed only, no product name or lei/kg; the names "Breeder 2" / "Cocoși" stay illustrative (D30). Which products are they?
12. **Total-row errors in the workbook:** farm ♂ mortality (col D) ≠ the compartments' sum on 22 Tuesdays (e.g. 17.11.2026 typed 424, compartments 144), so 22 weekly ♂ totals are off; the Total-row lay % is a mean of daily % (SAPT 58 33,55%; the app uses eggs ÷ start-of-day hen-days, D159); `Masa` farm egg weight is a plain mean with empty H7 (44 g on 18.11.2026); `Hrana Hirbovat` kg = 460 on 02–04.11.2026 where g × birds ≈ 690 kg. The app recomputes from sections (D156, D159) and keeps the typed kg: confirm which side is right at import.
