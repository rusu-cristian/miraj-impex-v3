# DataTable

Compact tables on `raised` with uppercase caption headers on `panel`.

- **Rows:** `row-h`, separated by `rule`. Numbers are right-aligned in Plex Mono.
- **Record-type columns:** the header carries the record icon in its colour, and ♀/♂ columns are coloured.
- **Groups:** `sunken` rows.
- **Totals:** a `line-strong` top border.
- **Inline edit cells** (owner prices) use a medium TextField.
- **Phone:** horizontal scroll with the first column sticky.
- **Row headers (additions.css §32, FIN14):** `table.ip-table.ip-table--rowheads` for rows named by `th[scope=row]` (group rows `th[scope=rowgroup]`, a totals row in `tbody` or `tfoot`): body text 13/18 `ink` on `raised`, `row-h`; group th `sunken` 500, totals th 500 on `panel` with the `line-strong` edge; `th.ip-table__child` indents a child row `space-6`. Used by S3, S4, S11, S12, S14, S18–S22, CostBlock and the small-curve table.
- **Wrapping heads (§32, FIN14):** `ip-table--wraphead`: figure heads (`thead th.num`) may wrap between words so the table fits before it scrolls; heads sit at the bottom. Keep a label that must not split with a no-break space ("Cat. 1"). Not with measured `--ip-colmin` heads (§54).
- **Scroll region (§36, FIN14):** `div.ip-tablescroll[role=region][tabindex=0][aria-label=the caption]` (+ `--framed` for the table's box): scrolls sideways in its own box, first column sticky with a `rule` edge; a first cell spanning columns does not stick; row states keep their fill on the first column. A loading copy is `aria-hidden`, not a tab stop.
- **Current row (S22P/FIN05, additions.css §45/§51):** the record open in a SidePanel beside the table: `tr.is-current` + `aria-current="true"` on the row's key (not drawn pressed); `sunken` with a hair `line-strong` frame inside, steel only, no lamp.
