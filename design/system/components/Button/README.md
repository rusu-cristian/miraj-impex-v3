# Button

Panel keys. `primary` (graphite `control`, one per view), `secondary`, `ghost` and `destructive`.

- **Primary:** `control` with `on-control` text and `shadow-key`.
  - hover: `control-hover`
  - pressed: `control-pressed` (the key sinks; no shadow)
- **Secondary:** `raised` with 1px `line-strong` and `shadow-key`.
- **Destructive:** `raised`, with the border and text in `fault`.
- **Disabled:** `sunken` + `ink-disabled`.
- **Loading:** a spinner, the label stays. `span.ip-spin` before the label is a static quarter arc in the label colour on a full quiet track, turned to the top right (track `line` on light keys, `ink-muted` on the graphite primary; same 14px box, 2px stroke, no animation; additions.css section 46, S1P, D121), so it never reads as a "C" before the label. While busy the key carries `aria-busy="true"` (with `aria-disabled="true"` when presses are ignored) and keeps its variant's colours, not the disabled grey (section 37, S14P).
- **Height:** `control-h` on desktop, `control-h-touch` on phone. Radius `radius-sm`.
- **Labels:** verbs: Salvează, Mută, Înregistrează vânzarea, Salvează prețurile.
