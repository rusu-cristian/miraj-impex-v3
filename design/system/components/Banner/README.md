# Banner

An inline message (info, warn, fault, ok): the status tint with a 1px status border and icon. The text stays `ink`.

Icon: 16px (`ip-icon--16`), centred on the first 18px text line by a hair above and below (additions.css section 47, FIN01).

With action (FIN03, additions.css section 49): `ip-banner--act` > icon, `ip-banner__body` (text or `p`s, space-2 apart), then the key `.ip-banner__key` (Reîncearcă, a Button link or an `ip-link`). Icon on the first line, key centred on the Banner; under space-12 × 6 (288) of text width the key wraps under the text, at its start. Text that stays one line beside a Button key adds `ip-banner--1l` (body as tall as the key, line centred).

Failure copy (FX04, D143): the object leads. Save error "<Obiect> nu s-a salvat: <what was typed>, <day>." (the action's own verb where it is not a save: "nu s-a creat", "nu s-a resetat"); load error "<Obiecte> nu s-au încărcat."; then one retry line on every failure Banner, "Verifică conexiunea la internet, apoi apasă Reîncearcă.", which names the Banner's own key, so every failure Banner is `ip-banner--act` with Reîncearcă, `role="alert"`; then what is kept ("Ce ai scris rămâne …") or what still works ("Poți … și acum."). Placement (D141): a save-error Banner is the form's last part, above the SaveBar or panel keys, and Reîncearcă takes the focus after the failed save; a load error stays where the content would be. The preview's fault Banner is the S6 Zilnic save error.
