# UnpricedCount · Livrări fără preț (project addition, SYS08)

Owner only (spec D2, screens S2 owner item 3: count + link to S11). How many feed deliveries wait for a price. CSS: additions.css section 25. Decided by default (user away), to review.

- **Markup:** `section.ip-unpriced[aria-labelledby]` (`panel`, 1px `line`, `radius-md`, 16px padding, gaps 12) >
  - head: RecordTile coins (price, `rec-feed`, `aria-hidden`) + `h2.ip-cap` "Livrări fără preț";
  - `p.ip-unpriced__fig`: count `.ip-unpriced__val` (readout style, Plex Mono 20/24 500 `ink`, unit after a real space: `4<small> livrări</small>`, as Readout, so it reads "4 livrări") + total `.ip-unpriced__kg` (Mono 14/20 `ink-secondary`);
  - `ul.ip-unpriced__list`: one line per farm × feed type with unpriced deliveries, split by `rule`: "Hîrbovăț · Breeder 2" | "3 livrări · 72 068 kg" (count `span.ip-num` Mono 400 `ink-secondary`, kg `span.ip-num` Mono 500 `ink`, right);
  - `p.ip-note`: "Cea mai veche din 04.11.2026. Kg fără preț nu intră în cost.";
  - `a.ip-btn.ip-btn--secondary` "Adaugă prețurile" → S11 › Livrări filtered to "fără preț" (`ip-btn--touch` on phone).
- **States:** default · none (state example: "0 livrări", "Toate livrările au preț.", key "Vezi livrările") · loading (`div.ip-unpriced__skel` stack of static `ip-skel`, `aria-busy`) · key hover/pressed/focus from Button.
- **Not a KPI card:** small tile, one Mono figure, no big icon, no lamp (the title says "fără preț"; no threshold exists for a warning).
- **Roles:** owner only; managers (read-only prices) and viewers/vets never get it (D06, D64).
- **Sources:** `sample-hirbovat-2026.json` › `deliveries` (sample, D27): 04.11, 10.11, 14.11 Breeder 2 (24 068 + 24 000 + 24 000 kg) and 06.11 Cocoși 4 039 kg = 4 livrări, 76 107 kg; every Telița delivery is priced.
