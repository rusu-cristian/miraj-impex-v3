# Impex Line Panel: design system index

Copy of the Design System artifact `95f22f4a-b2df-421d-b638-4d5e9d0656e2`, version `1794843774-ade9` (87 files, sha256-verified). Paths are relative to `design/system/`.

## Files
- `README.md`: the brand book (colour jobs, record-type code, content rules, surfaces, type, layout). Read it first.
- `tokens.json`: source tokens (colour, type, spacing, radius, shadow, size). Light theme only.
- `tokens.css`: **CSS custom properties** generated here from tokens.json (the artifact compiles this file but does not store it; bundle.css does NOT define any variables). Adds `.title` … `.figure-sm` type classes.
- `components/bundle.css`: every `ip-` class. HTML/CSS only, no bundle.js, no JS behaviour.
- `components/<Comp>/README.md` (spec) + `preview.html` (reference markup fragment, line 1 is a `@dsCard` marker). `components/Cover/` is the cover art, not a component.

## To render components in a page
```html
<link rel="stylesheet" href="design/system/tokens.css">            <!-- 1st: variables -->
<link rel="stylesheet" href="design/system/components/bundle.css"> <!-- 2nd: ip- classes (it @imports Google Fonts) -->
<body class="ip-root"> ... </body>  <!-- or wrap content in .ip-root: ground, ink, Plex Sans 14/20, box-sizing, focus ring -->
```
Previews are fragments (no `<html>/<head>`); paste their markup inside an `.ip-root`. Variables: `--rec-hen`, `--ground`, `--panel`, `--raised`, `--sunken`, `--ink`, `--control`, `--ok-lamp`, `--readout-ok`, `--space-1..12`, `--radius-xs|sm|md|lg|round`, `--shadow-key`, `--shadow-overlay`, `--control-h`, `--control-h-touch`, `--count-input-h`, `--row-h`, `--record-tile`, `--lamp`, `--header-h`, `--tabbar-h`, `--sidebar-w`, `--panel-w`, `--font-sans`, `--font-mono`.

## Icons
Inline `<svg class="ip-icon" viewBox="0 0 24 24" aria-hidden="true">` with `<path>/<circle>` children; Lucide-style line icons (plus custom skull, wheat, venus, mars, coins), stroke `currentColor`, width 1.75, round caps, no fill (set by `.ip-icon`). Sizes: default 18px, `ip-icon--14|16|22`. No sprite, no icon font, no icon files: copy the path data from the previews (RecordTile/RecordCode hold every record icon; BottomTabBar/Sidebar hold nav icons). Record icons: hen venus, rooster mars, mortality skull, cull ban, placement log-in, transfer arrows, sale truck, feed wheat/package/coins, weighing scale, eggs egg; sexing error = `ip-tile--split` with two rings.

## Fonts
IBM Plex Sans + IBM Plex Mono, weights 400/500, loaded by the `@import` of Google Fonts on line 1 of bundle.css. No font files ship (`type.fonts` is empty). Offline, the stacks fall back to `system-ui` / `ui-monospace`; self-host woff2 files if exact rendering is needed offline.

## Component classes (`ip-`)
**Utilities:** `ip-root` `ip-stack` `ip-row` `ip-num` `ip-cap` `ip-note` `ip-icon` (`--14|16|22`) `ip-focus`
**Record types**
- RecordCode / RecordTile: `ip-tile ip-rec ip-rec--{hen|rooster|mortality|cull|placement|transfer|sale|feed|weighing|eggs}`, `ip-tile--split`
- RecordChip: `ip-chip ip-rec ip-rec--*`, `ip-chip--off`
- SexCount: `ip-sex ip-sex--hen|rooster` (figure in `<b>`), `ip-pair` + `.sep`, `ip-pair--lg`
**Status**
- Lamp: `ip-lamp ip-lamp--ok|warn|fault|info|off`; on readouts `ip-rlamp ip-rlamp--ok|warn|fault`
- Tag: `ip-tag ip-tag--ok|warn|fault|info|neutral`
**Actions**
- Button: `ip-btn ip-btn--primary|secondary|ghost|destructive`, `ip-btn--touch`, `is-hover` `is-pressed` `is-disabled`/`:disabled`, `ip-spin`; link `ip-link`
- IconButton: `ip-iconbtn`, `ip-iconbtn--touch`, `is-hover`
**Inputs**
- TextField: `ip-field` > `ip-label` + `ip-input` (`.suffix`; `is-focus|is-error|is-warn|is-disabled|is-readonly|is-suggested`; `ip-input--touch|--area`) + `ip-hint|ip-err|ip-warnline`
- CountInput: `ip-input ip-count ip-count--hen|rooster`, `ip-count--md`, empty `.ph`
- Select / SectionSelect / CycleSelector: `ip-input` + `ip-menu` > `ip-menu__group`, `ip-menu__item` (`is-active|is-disabled`)
- Checkbox: `ip-check` (`is-checked`) > `ip-check__box`
- SegmentedControl / DateField: `ip-seg` > `ip-seg__item` (`is-selected`), `ip-seg--touch`
- FileDrop: `ip-drop`
**Data**
- Stat: `ip-stats` > `ip-stat` > `ip-cap`, `ip-stat__val` (`small`), `ip-stat__guide`
- DataTable: `ip-table` (`th`, `td.num`, `td.edit`, `tr.is-hover|is-group|is-total`)
- EventRow: `ip-list` > `ip-ev` (`is-hover`) > `ip-ev__main` > `ip-ev__meta`
- SectionTile: `ip-tiles` > `ip-stile` (`is-hover|is-missing`) > `ip-stile__head`, `ip-stile__code`, `ip-stile__kv`, `ip-bar`
**Moment:** Readout `ip-readout` > `ip-readout__cap|__val|__sub`; GuideChart `svg.ip-chart` (`.grid .axis .ax .band .guide .act .s2 .pt .today`)
**Signature:** CycleTrack `svg.ip-track` (`.lane .lanelbl .cyc[.open] .cyctxt .link .linkdot .place .sale .today .todaylbl .ro .rotxt .month .tick`)
**Entry:** EntryCard `ip-panel` (+ `ip-panel__head`); AllocationRow `ip-alloc` > `ip-alloc__who`, `ip-alloc__kg`; EffectPreview `ip-effect`
**Feedback:** Banner `ip-banner ip-banner--info|warn|fault|ok`; Toast `ip-toast`; EmptyState `ip-empty`; Skeleton `ip-skel`
**Containers:** Dialog `ip-scrim` > `ip-dialog`; SidePanel `ip-sidepanel` > `__head|__body|__foot`; BottomSheet `ip-sheet` > `ip-sheet__handle`
**App shell:** Sidebar `ip-sidebar` `ip-wordmark` `ip-nav` (`is-current`, `ip-nav--sub`); BottomTabBar `ip-tabbar` > `ip-tabbar__item` (`is-current`), `ip-tabbar__key`; FormHeader `ip-formhead`; SaveBar `ip-savebar` > `ip-savebar__sum`; PageHeader `ip-pagehead` + `ip-crumbs`; Tabs `ip-tabs` > `ip-tab` (`is-selected`); FilterBar `ip-filterbar`; phone frame `ip-phone`

`is-hover`/`is-pressed`/`is-focus` are static demo states; real pages use `:hover`/`:focus-visible` (only `.ip-root :focus-visible` is wired).
