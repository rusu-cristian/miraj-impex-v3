## 6. Open questions (most important first)

1. **One series per farm:** can a farm hold birds from two series at once, or must it be emptied first? The default lets the manager start a new cycle manually.
2. **Seria 2026 history:** how is it imported (cycles included), and from which date is the app the source of truth?
3. **Real guide tables:** which guides apply (one growth and one laying), and is the Hîrbovăț productivity guide available?
4. **Egg total formula:** is "Jos" excluded?
5. **Alert thresholds:** confirm mortality 3× the 7-day average, weight ±5%, lay ±3 points, feed ±10% and under 5 days of stock, missing entries from 12:00.
6. **Feed types:** the list of feed types, and which are for hens and which for roosters.
7. **Tone:** informal "tu" or formal "dumneavoastră"?

**Round 2 (19.11.2026, Raport_NOU_2026.xlsx; setup-check-2 §5, D153, DATA03)**

8. **Incubation batches:** the `Incubare` (batches 1–118, 20.03–22.11.2026: eggs set, set/hatch dates, chicks, hatch % vs guide, egg age per hall) and `BD Incubator` (2022–12.2026) sheets have no table in the data model. Not modelled this round (D153): do they belong in the app, and who enters them? **Resolved (D182):** out of scope, no table and no screen.
9. **Egg categories Cat. 1–3:** the form sends counts (`Raspunsuri`), `Masa` gives %; `EGG_PRODUCTION` has cat1–cat3 but no screen shows them this round (D153). Counts or %, and is Total ouă still Cat. 1–3 + Incubabile (D18)? **Resolved (D183):** counts per hall per day on Zilnic, Total ouă = Cat. 1–3 + Incubabile (D18); Producție tables (S4, S14) show % Cat. 1–3 of the eggs; D164's placeholder goes.
10. **Two lay % guides:** RealGhid col B and Productivitate col I disagree (SAPT 56: 56,9% vs 61,1%; RealGhid SAPT 53 typed "605", read as 60,5%). The app uses **RealGhid** (DATA03); confirm, and whether eggs/hen, hatching eggs/hen (week and cumulative) and the hatchability guide should join `GUIDE_WEEK`. **Resolved (D184):** guide lay % = Productivitate col I, SAPT 23–62 (SAPT 56 61,1%); no new guide lines. S20's paste example becomes that guide with week 53 typed "645" (D185).
11. **Feed type at Hîrbovăț:** deliveries carry kg and lei per ♀ feed / ♂ feed only, no product name or lei/kg; the names "Breeder 2" / "Cocoși" stay illustrative (D30). Which products are they?
12. **Total-row errors in the workbook:** farm ♂ mortality (col D) ≠ the compartments' sum on 22 Tuesdays (e.g. 17.11.2026 typed 424, compartments 144), so 22 weekly ♂ totals are off; the Total-row lay % is a mean of daily % (SAPT 58 33,55%; the app uses eggs ÷ start-of-day hen-days, D159); `Masa` farm egg weight is a plain mean with empty H7 (44 g on 18.11.2026); `Hrana Hirbovat` kg = 460 on 02–04.11.2026 where g × birds ≈ 690 kg. The app recomputes from sections (D156, D159) and keeps the typed kg: confirm which side is right at import.

**Round 3 (26.11.2026, voice, D199; spec §8 F1)**

13. **Speech and AI provider:** which transcription service and which model map a Romanian sentence to a catalogue action; where the audio is processed (EU?) and whether the contract guarantees no audio or text is kept or used for training. The design assumes audio is never stored.
14. **Cost:** price per command (transcription + model) × users × commands a day; is a daily cap per user needed, and what does the panel say when it is reached?
15. **Wake word:** not planned (no "Hei Impex", no always-on listening: battery, privacy, fans in the hall). Confirm.
16. **Accuracy of numbers in a hall:** fans and birds are loud; spoken figures ("două sute cincizeci", "o mie opt sute treizeci și unu") will sometimes be misheard. The design never saves a figure from voice (always a prefilled form with "din voce" marks); test in a hall before release and decide whether a headset is advised.
17. **Language mix:** staff in Hîrbovăț and Telița may mix Russian words or speak without diacritics. The interface is Romanian only (PRODUCT); should the recogniser accept Russian farm words, or refuse them as neînțeles?
18. **Transcripts in the audit (notice):** the transcript is kept with a save that still carries a "din voce" mark (`AUDIT_LOG`, `sursa: voce`, D199 note), and the panel shows one line on first use: "Ce spui nu se păstrează. Dacă salvezi un formular completat din voce, textul comenzii rămâne în Istoric modificări." Is that notice enough for the staff (and the law), or is a written consent needed?
19. **Spoken answers:** answers are written only; does anyone want them read aloud (speech synthesis), e.g. a vet walking the hall?
20. **Shortcut:** Ctrl + Shift + Space; ChromeOS uses Ctrl + Shift + Space to switch input methods (a Chromebook in the office would lose it, or switch layout instead), and Windows may switch the keyboard layout (Romanian/Russian) on Ctrl + Shift. Which machines are used, and is a second chord needed there?
21. **iPhone home-screen app:** iOS may ask for the microphone each session in a home-screen app; check before release whether the vets' phones are iPhones.
