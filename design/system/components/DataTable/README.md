# DataTable

Compact tables on `raised` with uppercase caption headers on `panel`.

- **Rows:** `row-h`, separated by `rule`. Numbers are right-aligned in Plex Mono.
- **Record-type columns:** the header carries the record icon in its colour, and ♀/♂ columns are coloured.
- **Groups:** `sunken` rows.
- **Totals:** a `line-strong` top border.
- **Inline edit cells** (owner prices) use a medium TextField.
- **Phone:** horizontal scroll with the first column sticky.
- **Current row (S22P/FIN05, additions.css §45/§51):** the record open in a SidePanel beside the table: `tr.is-current` + `aria-current="true"` on the row's key (not drawn pressed); `sunken` with a hair `line-strong` frame inside, steel only, no lamp.
