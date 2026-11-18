# CycleTrack

The signature: the series' cycles over calendar time, one lane per farm.

- **Lanes:** `sunken`, with uppercase farm labels.
- **Cycle blocks:**
  - closed: `panel` + `line-strong`
  - open: `raised` + a 1.5px `ink` edge, label weight 500
- **Markers:**
  - a placement start mark in `rec-placement`
  - the transfer link and dot in `rec-transfer`
  - a sale end mark in `rec-sale`
- **Today:** the "azi" line in `ok-lamp` across both lanes, with a small Readout per open cycle ("SAPT 41").
- **Track text:** the azi label, the Readout figure and the month labels are 11px (`caption` size; bundle.css drew them at 10, below the craft floor), set by additions.css for every `svg.ip-track` (FIN13). The Readout box is 56 wide for "SAPT 40" at 11px Mono (52 on phone).
- **Interaction:** clicking a block opens that farm at that cycle.
- **Variants:**
  - `full`: series page
  - `compact`: Acasă for the owner, the farm header
  - `phone` (S2V Acasă for the Observator, S18 Serii phone; D91): the full variant drawn 1:1 in the 358px phone content box, not scaled or scrolled: lanes 26 / blocks 20 spaced 48 apart, each block link with a transparent 44px hit rect (touch target), short block labels ("creștere", "ouat · din 10.02.2026"), month ticks every month with labels only on quarter starts, the axis run on past today (to 03.2027) so the 52px "SAPT 40" Readout fits right of the azi line. Track text 11px (`caption`), as every variant (FIN13).
  - `farm` (S4 farm page, one lane): `CycleTrackFarm/`.
- **Accessibility:** the SVG carries a text summary. Blocks are focusable links. Hidden names, titles and the summary are running text: the week reads "azi sapt. 40" (D115, FIN13), while the drawn Readout keeps "SAPT 40".
