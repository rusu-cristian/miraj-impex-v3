# EventRow

One record in a log: RecordTile, a title with section and figures (♀/♂ coloured), meta (date · SAPT · author), then a status lamp or tag.

- **Row:** height ≥ 48px, separated by `rule`; hover `sunken`.
- **Filtering:** by RecordChips in the FilterBar.
- **Locked rows** show a "blocat" tag.
- **Current row (FIN05, additions.css §51):** the record open in a SidePanel beside the log: `li.ip-ev.is-current` + `aria-current="true"` on its key (not drawn pressed); the DataTable current-row look (`sunken`, hair `line-strong` frame), steel only.
