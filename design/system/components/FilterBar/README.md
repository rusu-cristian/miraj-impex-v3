# FilterBar

Period, grain, record-type chips and export. On phone it collapses into one button that opens a BottomSheet.

## Project note (repo copy only, SYS20, 26.11.2026)
- **Equal controls (D195, additions.css §74):** period and grain SegmentedControls have equal segments; the record-type chips are one `div.ip-chips[role=group]` of equal chips on a line of their own, `div.ip-filterbar__chips` (D47); Exportă is pushed right by `ip-filterbar__end`. On the phone sheet: period and grain `ip-seg--fill` (short word "Ciclu" for Ciclul curent), tags `ip-chips--wrap ip-chips--fill` (2 columns). Spec: components.md › Additions › SYS20. To copy into the system artifact (HANDOFF).
