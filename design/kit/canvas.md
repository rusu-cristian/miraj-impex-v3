# Design canvas: Impex Line Panel · Ecrane

- Slug `e44dd7fb-2990-41f9-a791-2841f1b1dcd1` · URL https://claude.ai/artifact/VC8kqksPgxxYo9kPpK14yv
- Type: Design (canvas), `bcd2878f-a88d-4094-a616-8e39d8a5fd8a`. Version at setup: `1794844216-3062` (always re-fetch: `get_artifact`).
- Design system installed: Impex Line Panel (`95f22f4a-…`, v `1794843774-ade9`) as folder `impexlinepanel`.
- Setup sub-agent could call publish_files and copy_file: **yes**. Caveat: the server only lets a session change artifacts it made
  (`network_off_limit`); if a helper's publish is refused for that reason, hand the files to the coordinator to publish.

## File layout (only files under `project/` show; local mirror: `design/canvas/project/`)
- `project/canvas.json` — index (content.index). `boards` keyed by file name, `order`, `designSystems` (leave as is).
- `project/Main.dc.html` — artboard "00 · Index" at x 0, y 0, 880×560.
- `project/ds/impexlinepanel/tokens.css` (CSS vars + type classes with `.ip-t-` aliases; published from the repo's `design/system/tokens.css`, the single source), `tokens.json` (Theme menu), `components/bundle.css` (`ip-` classes; @imports Plex fonts).
- `project/additions.css`: the project's additions (D32: interaction states, motion vars, new components), published from `design/system/components/additions.css`. Kept outside `ds/` because the type reserves `project/ds/<folder>/` for the system install.
- One screen = one file `project/<stem>.dc.html`, flat (no subfolders). Stem: ASCII, letters/digits/`_`/`-`/`.`, unique.
  "S4 · Fermă · Rezumat" → `S4-ferma-rezumat.dc.html`; state "S4 · Fermă · Rezumat · se încarcă" → `S4-ferma-rezumat--se-incarca.dc.html`.
  The human name (with diacritics) goes in the board's `title`.

## Sizes
- Desktop (PAGE): `w` 1440, `"expand":"fill"`; root fluid (no px width; `max-width` container, %/fr), `h` = its height at 1440 (e.g. 1024; over-tall beats clipped).
- Phone: 390×844, root FIXED `width:390px;height:844px` (taller `h` only for a long scrolling form). No fake status bar.
- `$preview` in data-props = the board's `w`×`h`. Gaps: 80 px between boards in a row, 120 px between rows.
- Layout: one row per screen, starting y = 800. Default state at x 0, its states to the right (x += w + 80), phone boards after the desktop ones. Next row y = row y + tallest h + 120.

## Turning a self-contained screen HTML into an artboard (`.dc.html`)
Shared CSS is NOT re-published per screen: link the three shared files in this order, tokens.css → bundle.css → additions.css
(relative, flat path, no `../`; exact hrefs `ds/impexlinepanel/tokens.css`, `ds/impexlinepanel/components/bundle.css`, `additions.css`):
```html
<!doctype html>
<html lang="ro">
<head>
<meta charset="utf-8">
<title>S4 · Fermă · Rezumat</title>
<script src="./support.js"></script>
<link rel="stylesheet" href="ds/impexlinepanel/tokens.css">
<link rel="stylesheet" href="ds/impexlinepanel/components/bundle.css">
<link rel="stylesheet" href="additions.css">
</head>
<body>
<x-dc>
<helmet>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500&amp;family=IBM+Plex+Mono:wght@400;500&amp;display=swap">
<style>body{margin:0} /* + screen-only CSS rules from the source <style>, if any */</style>
</helmet>
<div class="ip-root" style="padding:0; min-height:1024px"> …the screen's <body> markup… </div>
</x-dc>
<script type="text/x-dc" data-dc-script data-props='{"$preview":{"width":1440,"height":1024}}'>
class Component extends DCLogic { renderVals() { return {}; } }
</script>
</body>
</html>
```
Rules (each fails silently): keep the `support.js` line exactly; drop the source's inlined tokens/bundle CSS (linked above) and any
`<script>` that builds UI; close every element, quote every attribute; no remote images, iframes, emoji; icons = inline stroke SVG
(Lucide-style, 1.75); real `<button>`/`<a href>`/`<input>`+`<label>`. Static comps: no `is_interactive`.
Type: use the `ip-t-<name>` classes (`ip-t-title`, `ip-t-body`, `ip-t-small`, `ip-t-caption`, `ip-t-readout`, …), not the bare `.title`/`.body`.
States on artboards: real `:hover`/`:active`/`:disabled` come from additions.css; to show a state statically use the demo classes (`is-hover`, `is-pressed`, `is-disabled`, `ip-focus`/`is-focus`).

## Adding one screen (one publish per call, sequential — never two publishes in parallel)
1. `get_artifact` slug → live `version`. `read_file` slug, path `project/canvas.json` → write it to `design/canvas/project/canvas.json`.
2. Write `design/canvas/project/S4-ferma-rezumat.dc.html`. In canvas.json add (keep every other key/entry untouched):
   `"boards": {"S4-ferma-rezumat.dc.html": {"x":0,"y":800,"w":1440,"h":1024,"expand":"fill","title":"S4 · Fermă · Rezumat"}}`
   (phone: `{"x":…,"y":…,"w":390,"h":844,"title":"S4 · Fermă · Rezumat · telefon"}`) and push the path onto `order`.
   Optionally update Main.dc.html's list.
3. publish_files:
```json
{"slug":"e44dd7fb-2990-41f9-a791-2841f1b1dcd1","base_version":"<live version>","mode":"patch",
 "files":[{"as":"project/S4-ferma-rezumat.dc.html","path":"/home/claude/miraj-impex-v3/design/canvas/project/S4-ferma-rezumat.dc.html"},
          {"as":"project/canvas.json","path":"/home/claude/miraj-impex-v3/design/canvas/project/canvas.json"}]}
```
   Keep the returned `version` for the next call. Replacing a screen = same `as`, no canvas.json unless layout changes.
   Remove = `"remove":["project/<stem>.dc.html"]` plus taking it out of `boards` and `order`.
4. Conflict: re-run step 1, redo the canvas.json edit on the fresh copy, retry (stop after 3). Do not render or read back to verify.
- Never use publish_version/mode "replace" on this canvas. Updating shared CSS: edit the repo file first (`design/system/tokens.css` or
  `design/system/components/additions.css`), copy it to the mirror, then publish it at `project/ds/impexlinepanel/tokens.css` or `project/additions.css`.
  bundle.css is the system's and stays unchanged; additions and fixes go in additions.css. (SYS00 published both; canvas version after SYS00 fix 2: `1794845755-a7ea`; after SYS01: `1794846357-5576`; after SYS01 fix 1: `1794847363-9139`; after SYS01 fix 2: `1794847872-f8cd`; after SYS02: `1794848329-7c69`; after SYS03 fix 1: `1794849537-0345`; after SYS04: `1794849991-b4dd`; after SYS04 fix 1: `1794850418-ec52`; after SYS05: `1794850956-bcf8`; after SYS06: `1794852762-3f0c`; after SYS06 fix 1: `1794853378-53b6`; after SYS07 fix 1: `1794854256-1a8a`.)
