# SYS02 · FeedTotals (S7 Hrană zilnică) · 16.11.2026
- Built `ip-ftot` (phone default + `ip-ftot--wide`) and compact `ip-ftot-sum` for the SaveBar; CSS additions.css §8; spec components.md › Additions › SYS02; folder system/components/FeedTotals/.
- Figures from sample-hirbovat-2026.json: Breeder 2 ♀ 3 332,5 kg (5/6), 31 584 → 28 252 kg, 7,1 zile; Cocoși ♂ 358,8 kg (5/6), 2 009 → 1 650 kg, 3,9 zile (abatere); total 3 691,3 kg, 2 de confirmat (H3.2 ♂, H4.2 ♀).
- Stoc sub 0: warn lamp "stoc sub 0" (D40) + the system's own warn Banner sentence; Salvează enabled (warning, not refusal). −1 240 kg row is a D39 state example.
- Nothing confirmed: Azi "–" "nimic confirmat", stock now and days from it (7,9 / 4,8), Salvează disabled.
- Compact counts rows ("10 din 12 rânduri"), not sections (6 sections = 12 rows); kg one decimal everywhere so ♀+♂ add up.
- Wide variant by modifier class, not a container query (system uses modifiers like `--touch`; no literal breakpoint).
- Rendered 2× at 1440 (headless Chrome); fixed footer wording and date wrapping.
- Canvas: additions.css published to project/additions.css, version 1794848329-7c69.
- SYS04 fix 1 (D50, 16.11.2026): stock below 0 now shows the warn lamp "abatere" (row and SaveBar summary); "stoc sub 0" moves to the detail text (first sub-line of Stoc după azi; SaveBar: "Breeder 2 · stoc sub 0" after the lamp). README + components.md › SYS02 updated. Rendered 1440/390.
