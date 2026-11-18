# EntryKeys (project addition, SYS08)

The entry shortcuts on Acasă (spec D2, screens S2: vet item 3, manager item 4). Secondary Button keys that open the entry forms. CSS: additions.css section 24. Decided by default (user away), to review.

- **Markup:** `section.ip-entrykeys[aria-labelledby]` (size container) > `h2.ip-cap` "Înregistrează" > `ul.ip-entrykeys__list` > `li` > `a.ip-btn.ip-btn--secondary.ip-entrykey[href]` > RecordTile (`aria-hidden`) + `.ip-entrykey__txt` (label, Button type 14/20 500; hint `span.ip-note` 12/16).
- **Keys and order:** vet: Zilnic (mortality · "Mortalitate, sacrificare, ouă"), Hrană zilnică (wheat · "Grame pe găină și cocoș"), Cântărire (weighing · "Greutate, CV, uniformitate"). Manager adds Transfer ("Între secțiuni sau ferme"), Vânzare (sale · "Păsări vândute pe secțiuni"), Livrare hrană (package · "Kg pe tip de hrană"). Owner: no EntryKeys on Acasă; the Înregistrează tab covers entry (D66).
- **Hint = what the form records**, not today's state (Introdus azi and Lipsesc date already carry today's state).
- **Layout:** grid of ≥ 168px columns (`space-12` × 3.5): 6 keys in one row at 1136, 3 keep the same width; 2 columns on a 358px phone; with 2 columns an odd last key spans the row (vet: Cântărire across). Height ≥ `control-h-touch` (48) everywhere; padding `space-2`/`space-3`.
- **No primary:** Acasă has no single main action, so every key is secondary (one primary per view stays free for the screen).
- **States:** hover `sunken` · pressed `rule`, no shadow, hint → `ink-secondary` (D34a) · focus ring · disabled = `a[role=link][tabindex=0][aria-disabled=true]` without href, still in the Tab order so its reason is read (as SYS09's disabled Editează; FIN13) (`sunken` + `ink-disabled`, tile loses its record colour), with one `p.ip-note[id]` under the grid naming why ("Telița e goală din 28.02.2026. Intrările se deschid cu ciclul nou."); every disabled key carries `aria-describedby` pointing to that note (not the section, where screen readers ignore it).
- **A11y:** links (they navigate); label + hint form the name; icons decorative.
