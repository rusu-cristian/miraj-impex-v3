# SYS20 · Equal-width items in every control group (D195), 26.11.2026
- additions.css §74 (both copies): `ip-seg` / `ip-chips` = inline grids of 1fr (every item = the widest), `--fill` shares its container; Tabs n × widest (+space-4 pad, dot hangs so no shift), phone tabs share 358; BottomTabBar equal columns. §1 FV02 gap and §38 fill flex rule moved into §74.
- New system classes: `div.ip-chips` (`--wrap` + `--ip-chips-cols`, `--fill`), `div.ip-filterbar__chips`. Removed screen rules: `.s4-chips`, `.s4ns-phone .ip-tabs{overflow-x}`, `.s4ns-fsheet`/`.s14-fsheet .ip-seg*`. S4 narrow page (<1000) chips 4 columns.
- Short words (full word in ip-sr): phone sheet "Ciclu" (Ciclul curent), S17 "Tot" (Tot istoricul), phone Detaliu "Compart." (Compartimente).
- Rebuilt every row + relayout + Main (2nd run byte-identical); canvas.json unchanged; no board w/h changed. S9/S10 desktop backdrops: Exportă wraps under the 720 FilterBar (+40 inside the clipped page). S17 phone sheet tags 5 rows (was 4, sheet scrolls).
- Checks: literal/sex/head scans exit 0; new eq_scan.py 0 mismatches (324 groups, 341 files); heights files unchanged after re-rendering; 122 shots retaken.
- Previews: SegmentedControl 232, Tabs 156, RecordChip 148, FilterBar 165, FilterInterval 1024 (dsCard), BottomTabBar, PeriodTable (phone short word).
- Published canvas `1795695290-d903` (180 files, patch). To review: desktop tab line now spans the tabs only; phone tabs tight (Greutate/Producție ≈ 8 px apart).
