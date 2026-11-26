# RecordChip

A labelled record type, used in legends and as filter chips in the FilterBar.

- **Style:** background `rec-*-bg`, text and icon `rec-*`, 1px border at 30% of the colour, radius `radius-sm`, height 28px.
- **Filter chips:**
  - on: the full chip
  - off: `raised` with `ink-muted` text, keeping the coloured icon
- **Behaviour:** `role="checkbox"` with `aria-checked`.

## Project note (repo copy only, S4P, 16.11.2026)
- **Split variant `ip-chip--split`** for Eroare sexare (no single record colour, D04): neutral steel chip (`sunken` + `ink`; off: `raised` + `ink-muted`) led by the split RecordTile at 20px with the two rings. Markup: `<button type="button" class="ip-chip ip-rec ip-chip--split" role="checkbox" aria-checked="true"><span class="ip-tile ip-tile--split" aria-hidden="true">…two rings…</span>Eroare sexare</button>`. CSS: additions.css › 14; spec: components.md › Additions › S4P. To copy into the system artifact (HANDOFF).

## Project note (repo copy only, SYS20, 26.11.2026)
- **Chip group, equal chips (D195, additions.css §74):** `div.ip-chips[role=group]` (checkbox chips; a phone's single-choice tags `role=radiogroup` of `role=radio` chips) is an inline grid: every chip as wide as the widest, icon/tile + word centred, `space-2` apart, one row. `.ip-chips--wrap` breaks the row into `--ip-chips-cols` equal columns (default 2); `.ip-chips--fill` spans its container (phone FilterBar sheet: 2 × 175 in 358). In a FilterBar the group sits on its own line in `div.ip-filterbar__chips`. Spec: components.md › Additions › SYS20. To copy into the system artifact (HANDOFF).
