# SegmentedControl

2–4 exclusive options: period, grain, ♀/♂.

- **Track:** `sunken`.
- **Selected segment:** `raised` with 1px `line-strong`.
- **Sex segments** carry their ♀ / ♂ colour.

## Project note (repo copy only, SYS20, 26.11.2026)
- **Equal segments (D195, additions.css §74):** `span.ip-seg` is an inline grid of `1fr` columns, so every segment is as wide as the widest (n × widest + the track's 2px gaps), label centred on one line; stretched by a flex column it keeps that width (`max-width: max-content`). `.ip-seg--fill` spans its field and shares it equally (phone FilterBar sheet, phone ♀/♂). A label that would not fit shows a short word, `<span aria-hidden="true">Ciclu</span><span class="ip-sr">Ciclul curent</span>` (also "Tot" = Tot istoricul, "Compart." = Compartimente on the phone Detaliu). Spec: components.md › Additions › SYS20. To copy into the system artifact (HANDOFF).
- **Sliding thumb (MO04, D198, additions.css §77):** the track draws one thumb (`.ip-seg::before`, `raised` + 1px `line-strong`) behind the selected segment and slides it to a new one (translateX, 200 ms `--ip-ease`), labels' colour in step; arrow keys and reduced motion move it at once. The selected segment's own fill is transparent.
