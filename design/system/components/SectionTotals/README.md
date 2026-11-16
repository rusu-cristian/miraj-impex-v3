# SectionTotals (project addition, SYS01)

The farm totals strip under a SectionTile grid: farm totals = the sum of the sections (spec D1).

- **Markup:** `div.ip-stotals[role=group][aria-label]` > one `div` per cell; `.ip-stotals__key` (record icon in `.ip-rec.ip-rec--*` + key), `.ip-stotals__val` (Plex Mono 14/20, 500, `ink`, unit in `<small>`), `.ip-stotals__sub` (12px `ink-muted`: guide, what's excluded, reason for `–`).
- **Strip:** `panel`, 1px `line`, top edge `line-strong` (like DataTable's total row), radius `radius-md`; cells separated by 1px `rule`; grid `auto-fit`, min 144px.
- **Order (same as SectionTile):** Total fermă (caption, section count, date · SAPT) · În viață ♀ · ♂ (SexCount, total under it) · Mort. săpt. (skull, `rec-mortality`; total, SexCount ♀ · ♂ under it) · Hrană azi ♀ · ♂ g (wheat, `rec-feed`; SexCount in `rec-hen` / `rec-rooster`; ♀ average weighted by hens, ♂ by roosters, D38; sub-line "ghid 158 / 155 g") · Greutate ♀ vs guide (scale, `rec-weighing`, weighted by hens) · Ouat vs guide (egg, `rec-eggs`, laying farms only).
- **Figures (preview):** recomputed from `counts-latest.json` and `sample-hirbovat-2026.json` for 16.11.2026 (sample, D27): alive ♀ 25 252 · ♂ 2 685; week 40 mortality −64 (♀ −51 · ♂ −13); feed ♀ 158 g (fără H4.2) · ♂ 159 g (fără H3.2); weight ♀ 3 959 g; lay 74,0 % vs guide 74,1 %. Tiles above use the same data (lay per hall). Grouped numbers use U+202F.
- **Totals say what they exclude:** a section without an allocation today is left out of that sex's average and named ("medie fără ♀ H4.2", "fără ♂ H3.2"). Nothing entered: `–` with the reason ("nimic introdus azi", "lipsă azi · ieri 73,9 %").
- **Status:** a Lamp with its word in the sub-line, only where a guide exists (Hrană azi: "în normă" while both sex averages are within ±10 % of the guide, D19; preview ♀ 158 vs 158, ♂ 159 vs 155 g = +2,6 %); no guide figure: "ghid: – fără date în ghid" (D29). Figures stay `ink`; only ♀/♂ use their colours.
- Static (not a link). Growth farms drop the Ouat cell.
