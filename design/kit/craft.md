# Impex Line Panel: craft rules (read before building any screen)

Tokens and specs live in DESIGN.md and components.md; this file is the rulebook. Where it is silent, those files win.

## 1. Direction
- Impex Line is an internal daily record of breeder-flock series across farm cycles, compared with each farm's guide. Character: alert, exact, mechanical.
- The world is the hall's climate controller: light steel panels, engraved uppercase captions, graphite keys, status lamps with words, dark Readouts, IBM Plex Sans + Mono. Light theme only, mixed density (32px desktop, 48/56px phone).
- Moments: one dark Readout per section, the GuideChart, the CycleTrack (signature). Everything else stays quiet steel.
- Motion is mechanical: short, one curve, from where a thing comes; data never moves (D196).
- **Template test:** swap in a generic farm-SaaS name. If the screen still fits, it fails. It must show farm lanes with cycles, lamp-plus-word status or the fixed record-type code.

## 2. Do's and Don'ts
Do:
- Lead every record (event row, entry card, alert, sheet entry) with its RecordTile.
- Colour every ♀ figure `rec-hen` and every ♂ figure `rec-rooster`: rows, inputs, chart lines, legends.
- Pair every lamp with a word. Put every number, week and readout in IBM Plex Mono.
- Write the week as the birds' age week (`SAPT 59` on 18.11.2026; D151: age 0 on the placement day, never restarted by a cycle or a transfer). In tables the day's age goes in the "Vârstă (zile)" column (407 on 18.11.2026, D159). Show unpriced feed as "fără preț".
Don't:
- Use a record colour for status, or a status colour for a record type. Use green for primary buttons.
- Use more than one Readout per page section (feed stock readouts excepted). Use dark panels outside Readouts.
- Use gradients, glows, neon, glassmorphism or blur. Build KPI-card grids with big icons. Use emoji or illustrations.
- Use exclamation marks or greeting headlines. Make lamps blink.
- Borrow anything from the ledger look: serif titles, barn red, paper.

## 3. Colour jobs (hard rules)
- **Record type = `rec-*`**, fixed icon + colour, identical on every screen (table in components.md › RecordCode). Marks *what* a record is. Always with its icon and a word or ♀/♂; never colour alone.
- **Status = lamp + word** in `ok` / `warn` / `fault` / `info` (+ `-bg` tints; `readout-ok/warn/fault` on Readouts). Marks *how* it is. Never a `rec-*` colour.
- **Steel = everything else:** surfaces, text, lines, keys. Primary key is graphite `control`, one per view. Stat values stay `ink`.
- Close pairs (cocoși blue / cântărire cerulean, sacrificare plum / transfer violet) are told apart by icon and word, never by hue.
- Readout series beyond actual: `readout-2…4`, each with its own dash. Guide is dashed `readout-guide` with `readout-band`.

## 4. Unslop checks
- [ ] No gradient, glow, coloured shadow, neon border, blur or glass anywhere. Shadows only `shadow-key` (button lower edge) and `shadow-overlay` (menus, dialogs, toasts, side panel).
- [ ] No emoji, no sparkle, no illustration, no generated image, no logo mark (wordmark only).
- [ ] No row of equal KPI cards with big icons. Stats are a caption, a Mono value, the guide line and a status; the one key figure goes in a Readout.
- [ ] No generic SaaS tells: welcome banner, avatar-heavy header, "Overview" hero, donut charts for decoration, sparklines in stat cards, pill-shaped everything, centred hero layouts. Layout follows DESIGN.md › Layout.
- [ ] Every icon carries meaning: record icons only for their type, Lucide-style line icons at 1.75 stroke, 14/16/18/22px. Remove decorative icons.
- [ ] Spacing on the 4px grid, no one-off values; equal things equal (rows, card padding, gaps); tighter inside groups than between them.
- [ ] Radii only 2/3/4/6 per DESIGN.md › Shapes; round only for lamps and the phone centre key.
- [ ] No filler copy, hype, slogans or unverified claims. Every line names a farm, section, figure, date or action.
- [ ] Weights 400/500 only; uppercase only in captions (and the elements the specs name: wordmark, breadcrumbs, lane labels, Readout caption).
- [ ] Motion: no loops, shimmer, pulse, bounce, `transition-all`, hover scale, `scale(0)`, `ease-in` or keyframes (D196).

## 5. Craft floor
- **Hierarchy:** one title (24px), clear steps through heading 16, body 14, small 12, caption 11. Hierarchy from size and ink role, not colour.
- **Alignment:** everything on the 4px grid; icons in one column, labels on one line, numbers in one right-aligned column.
- **Contrast (WCAG 2.2 AA):** text ≥4.5:1 (use `ink`, `ink-secondary`, `ink-muted`; `ink-disabled` only on disabled); controls, borders that identify a control, and focus ≥3:1 (`line-strong`).
- **Focus:** `:focus-visible` = 2px `focus` ring, 2px offset, on every interactive element, never removed; input-type fields draw it on their own edge, inside the box (offset −2px, additions §68, D149).
- **Targets:** ≥44px hit area on touch; phone controls `control-h-touch` (48) and CountInput 56. Desktop 32px controls keep ≥24px and are spaced.
- **Keyboard:** everything reachable in order; Enter moves to the next CountInput; Dialog focus starts on Anulează; GuideChart has a keyboard readout and a hidden data table; CycleTrack blocks are focusable links with a text summary.
- **Labels and roles:** visible label above every field (inline table edits, D150: the column head plus the field's `aria-label` naming the row, or a `label` in the row's detail row, e.g. the SectionTree code field); IconButton has `aria-label`; RecordTile is `aria-hidden` with the row text naming the type; RecordChip is `role="checkbox"` + `aria-checked`.
- **States, all designed:** hover, focus, pressed, disabled (`sunken` + `ink-disabled`), loading (static `sunken` skeletons; spinner only inside a button, label kept), empty (EmptyState: unlit lamp, one sentence, next step), error (fault border + line naming the fix), warning (saveable) vs refusal (save disabled), no permission ("Nu ai acces la această fermă.").
- **Numbers:** Plex Mono, `tabular-nums`, right-aligned in tables; one format per screen.
- **Forms:** one column, label above, "(opțional)" for optional; validate on blur and save; focus the first error; keep typed values and the draft.
- **Motion (D196, §75):** only `cubic-bezier(0.2, 0, 0, 1)` at 120ms (keys, menus, toggles, rows), 200ms (dialog, panel, toast, banner), 280ms (phone sheets, page); transform + opacity; occasional things move, tens-a-day controls change colour, data and navigation never move. Under `prefers-reduced-motion`, keep the fades, drop all movement.
- **Live updates (D197, §76):** the one exception: data a reader did not change may move once (a new row rolls in 8 px + fade, a value crossfades 4 px, a transfer mark grows from 0.96), and the record tint fades **linear** over 1.6 s (a timer of recency, not a movement); never counting, never looping.

## 6. Copy
- Romanian with full diacritics, ș ț with the comma below (never ş ţ cedilla). Address the user as "tu".
- Buttons are verbs: Salvează, Mută, Înregistrează vânzarea, Salvează prețurile, Începe ciclu nou. Dialog titles name the action and the object.
- Errors name the section, the number and the fix (refusals also the day). No exclamation marks, emoji, hype or greetings; no period on labels and buttons.
- Counts: ♀ before ♂, joined by `·`; thin no-break space for thousands (`14 080`); true minus (`−82`); decimal comma (`2,41 kg`, `−1,4%`); units as suffixes (g, kg, lei/kg, %).
- Dates `dd.mm.yyyy`. Week: `SAPT 59` in Mono context lines and Readouts; `sapt. 59` inside running text and toasts.
- Unknown is `–` with its reason (never 0, never blank). Unpriced feed is "fără preț"; totals say what they exclude ("8 000 kg fără preț").

## 7. Content honesty
- Sample figures come only from the kit and the workbook extracts (design/data/extracts; real to 18.11.2026 = "today" on every board, D152), e.g. placement 07.10.2025: H1 ♀ 14 080, H2 ♀ 14 000, H5-1 ♂ 3 510; Hîrbovăț halls H5, H6, H7 with sections H5.1, H5.2, H6.1, H6.2, H7.1, H7.2; alive 18.11.2026 ♀ 12 561 · ♂ 794; delivery 13.11.2026 25 640 kg (see key-figures-18-11.md). The old codes H1.1…H4.2, "16.11.2026" as today, "SAPT 40/41" and ♀ 25 252 are stale. Anything else is labelled sample.
- Names of people (Dr. Ceban, Ion Rusu), guide names ("Ross 308 ouat") and figures dated after the workbook are illustrative.
- Never invent missing data. Show the waiting or missing state instead: `–` with a reason, the "lipsă" lamp, an EmptyState, a pending guide.
