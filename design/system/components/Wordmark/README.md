# Wordmark · Impex Line (project addition, SYS13)

The product name as the only brand mark (PRODUCT.md › Brand Commitments: "No logo exists. The name is used as a wordmark."). Text only: no icon, no mark, no artwork. CSS: bundle.css `.ip-wordmark` (unchanged) + additions.css section 31. Decided by default (user away), to review.

- **Markup:** the text is written `Impex Line` (A1 copy: title "Impex Line"); bundle.css sets it uppercase, Plex Sans 500, tracked `.08em`, in `ink`. Because the capitals come from CSS, a screen reader says the name, not the letters I-M-P-E-X.
  - **Sidebar:** `div.ip-wordmark` at the top of `ip-sidebar`, 15/20 with the Sidebar's own padding (as components.md › Sidebar draws it; nothing changed). Not a heading and not a link (the page's h1 is its PageHeader title).
  - **Login card:** `h1.ip-wordmark.ip-wordmark--login`, the title style 24/30 (`--ip-fs-title` / `--ip-lh-title`, named copies of tokens.css `.ip-t-title`), no padding, no margin. It is the S1 page's only heading and names the sign-in form (`aria-labelledby`). Same size on desktop and phone.
- **Rules:** one wordmark per view. Never coloured, never on a dark panel, never with a tagline or slogan (craft §4: no filler copy). Uppercase here is one of the named exceptions (components.md › Additions › Uppercase exceptions).
- **Keyboard / a11y:** not interactive. Login: the h1 is the page heading (the document `<title>` is "Autentificare · Impex Line").
- **States:** none.
- **Sources:** name from PRODUCT.md / spec A1. Builder/render: `design/system/scripts/sys13_build.py`, `sys13_render.py`.
