# GuideChart

Actual vs guide, drawn on a Readout.

- **Grid:** `readout-line`.
- **Guide:** dashed `readout-guide`, with its tolerance band in `readout-band`.
- **Actual:** `readout-ok`.
- **Extra sections:** `readout-2` … `readout-4`, each with its own dash.
- **Today:** a faint `readout-ok` line, with the last point marked.
- **X axis:** guide weeks of the cycle.
- **Accessibility:** keyboard readout, a hidden data table, no animation.
- **Variants:** weight, lay, feed g, mortality %, egg weight.
- **Preview (repo copy, SYS06):** Telița Seria 2026 ♀ weight (real, hen-weighted) vs the real growth guide, SAPT 1–21, warn "ABATERE" (+6,8% at SAPT 20); markup `role="group"` + `aria-roledescription="grafic"`, live-region week readout, hidden table with `th scope="col"` (D42/D44). Variants: `GuideChartVariants/`.
- **Phone drawing (SYS06 fix 1):** a 600-wide viewBox shrinks its 10px axis text below 10px under a 600px box, so every chart is `div.ip-chartbox` (container query) holding `svg.ip-chart.ip-chart--wide` (viewBox 600×150, plot x 34–592) and `svg.ip-chart.ip-chart--phone` (viewBox 320×150, plot x 40–308, axis text 11px = caption size, fewer ticks: X first · middle · last, ≤3–4 Y ticks). Below 600px the phone one shows (≥11px text at 390), above it the wide one; the hidden one is `display:none`, so only one is focusable and read. Both carry the same ARIA and point at the same help line, readout and hidden table. The hidden `table.ip-sr` is `display:block` so it never widens the page. CSS: additions.css › 16.
- **Legend:** `div.ip-chart__legend` swatches under the readout line (act line, dashed guide, band) + words; the sub-line keeps only context (dates, missing weeks), no text legend.
