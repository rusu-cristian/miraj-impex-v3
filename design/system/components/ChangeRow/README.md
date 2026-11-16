# ChangeRow · Istoric modificări (project addition, SYS09)

One change in the change log (spec E2, screen S17): date, user, table, section, old → new, reason, with a link to the record. CSS: additions.css section 27. Decided by default (user away), to review (SYS09 decisions A–D).

- **Markup:** `ul.ip-list.ip-list--change[aria-label]` (size container) > optional `li.ip-chg__cols[aria-hidden]` (column head: Când · Înregistrare · Valori · vechi → nou · Motiv · Cine · Tip, shown ≥ 720px) > `li.ip-ev.ip-ev--change`:
  - RecordTile of the record changed (the spec's "table" is shown as the record type: tile + word, never a table name).
  - `.ip-ev__main` > `a.ip-chg__link` "Mortalitate · H4.1 ♀" (+ hidden " găini, ziua 14.11.2026"), stretched over the row (`::after`), to the section page at the record's day; + `.ip-ev__meta` "ziua 14.11.2026 · SAPT 40 · Hîrbovăț".
  - `ul.ip-chg__vals` > `li.ip-chg__val` per changed field: key (12px `ink-secondary`) · `<del>` old (Mono `ink-secondary`, struck in `ink-muted`) · arrow `ink-muted` (`aria-hidden`) · `<ins>` new (Mono 500 `ink`, no underline); hidden "înainte" / ", după"; deleted = new "–" read as "șters".
  - `p.ip-chg__why` "Motiv: …" (12px `ink`); none = "– în aceeași zi, fără motiv" (`is-none`, `ink-secondary`).
  - `.ip-chg__meta` > `.ip-chg__when` (Mono "15.11.2026, 10:12") · `.ip-chg__who` (name + role `ip-note`: veterinar / manager / proprietar).
  - Tag `neutral` of the change type · `.ip-ev__go` 16px chevron (Lucide chevron-right).
- **Change types (Tag word):** **corectat** — E1 Corectează, reason required (real: Mortalitate H4.1 ♀ 14.11.2026, 3 → 5, Ion Rusu 15.11.2026, 10:12) · **șters** with reason — E1 Șterge by manager/owner · **editat** — the author's same-day edit (C8, before 23:59), no reason · **șters** without reason — the author's same-day delete. One row may hold several fields (Cântărire: greutate medie, CV, uniformitate). Logging same-day changes is decision A.
- **Layouts:** phone stacked (tile | record + Tag + chevron; values; reason; when · who on one line), whole row the target (≥ 48px). ≥ 720px one line per change: when 84 · tile · record 1.1fr · values 1fr · reason 1.5fr · who 120 · Tag 72 · chevron 16; "Motiv:" and the separators become hidden text (the head names the columns).
- **States:** hover `sunken` (neutral Tag → `rule` so it keeps its fill), pressed `rule` (arrow + chevron → `ink-secondary`, D34a), focus ring on the row (link's own outline moved to the row, 2px offset, over neighbours), empty (`li.ip-list__empty`: off Lamp "fără modificări" + "Nicio modificare la Hîrbovăț între 09.11.2026 și 16.11.2026." + next step "Schimbă perioada sau filtrele."), loading (`ip-ev--skel` rows, `aria-busy`).
- **Filters** (user, farm, period, record type) are S17's FilterBar, not part of the row.
- **Sources:** audit_log in sample-hirbovat-2026.json is the one real change. State examples (D39) built from real figures: deleted cull H3.1 ♂ 2 · boală (15.11.2026), edit H1.2 ♂ 155 → 172 g (kg × 437 cocoși), deleted weighing H1.1 ♀ (3 967 g, 7,3%, 81,9%); their times and the delete reason are not in the data.
