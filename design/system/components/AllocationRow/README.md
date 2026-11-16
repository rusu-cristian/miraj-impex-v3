# AllocationRow

One section × sex row in Hrană zilnică.

- **Left:** the ♀/♂ sign and the section code, with birds alive.
- **Middle:** the feed type select and the g/pasăre input. Yesterday's grams appear as a *suggestion* (dashed `sunken` field, "ieri") until confirmed with a tap or by typing.
- **Right:** the live kg, plus the guide figure and a deviation lamp.
- **Unconfirmed rows** aren't saved.
- **Card:** `raised`, 1px `line`, radius `radius-md`.

## Phone layout (project addition, S7P · `ip-alloc--phone`)

Repo copy only (D16/D32): CSS in `additions.css` §19, spec in `components.md` › Additions › S7P. Used by S7 Hrană zilnică at 390.

- **Markup:** `div.ip-alloc.ip-alloc--phone[role=group][aria-label="H1.1, găini, 4 280 în viață"]` > `.ip-alloc__who` (`span.ip-sex` sign, `aria-hidden`; code + "4 280 găini") > `.ip-alloc__kg` (kg, `<small>` "ghid 158 g", Lamp + word or `<small>` "de confirmat") > `.ip-alloc__ctl` (two `label.ip-field`: Tip hrană `ip-input ip-input--touch` with a native `<select>` + chevron; g/pasăre `ip-input ip-input--touch ip-num` with `<input inputmode=decimal>` + suffix "g" / "ieri") > optional `p.ip-warnline|ip-err.ip-alloc__msg` (referenced by the input's `aria-describedby`).
- **Layout:** line 1 who | kg on one baseline; line 2 Select (1fr) + g field (120px) at 48px, labels above; message across the card. No empty grid track without a message (166px card).
- **States:** în normă (ok Lamp) · abatere > ±10% (`is-warn` field, warn Lamp, "+11%" after the guide, warnline; saveable) · suggestion (`is-suggested` field with "ieri", `.ip-alloc__kg.is-suggested` muted kg, "de confirmat"; not saved) · empty (placeholder = guide g, kg "–", "de completat") · out of range (`is-error` + `aria-invalid`, kg "–", fault Lamp "refuzat", ip-err) · saved (`is-readonly` fields, Lamp kept).
