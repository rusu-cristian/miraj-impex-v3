"""python3 s10_render.py [--heights] : render the S10 boards (phone 390 from /tmp/s10, desktop 1440×1024 + 1280 from /tmp/s4s)
to design/shots/S10.png (the tall default, full page), S10-1 … S10-6.png (canvas order) + S10-6-1280.png, and report sideways
scroll, duplicate ids, broken ARIA references, unnamed form controls, text under 4.5:1, targets under 44 (phone) / 24
(desktop), the tab order, the EffectPreview / Banner / Dialog vs the SaveBar, and page elements under the open SidePanel.
--heights writes the default board's natural height to s10_heights.json (then rebuild with s10_build.py)."""
import asyncio, json, pathlib, sys
from playwright.async_api import async_playwright
R = pathlib.Path('/home/claude/miraj-impex-v3')
HJ = R / 'design/canvas/scripts/s10_heights.json'
ORDER = [('S10-vanzare', 390, 844, 's10'), ('S10-vanzare--confirmare', 390, 844, 's10'), ('S10-vanzare--refuzat', 390, 844, 's10'),
         ('S10-vanzare--salvat', 390, 844, 's10'), ('S10-vanzare--eroare-la-salvare', 390, 844, 's10'),
         ('S10-vanzare--partiala', 390, 844, 's10'),
         ('S10-vanzare--desktop', 1440, 1024, 's4s'), ('S10-vanzare--desktop', 1280, 1024, 's4s')]
DESK = 6
CHECK = r"""() => {
 const ids = [...document.querySelectorAll('[id]')].map(e => e.id); const dup = ids.filter((x, i) => ids.indexOf(x) !== i);
 const refs = [...document.querySelectorAll('[aria-describedby],[aria-labelledby],[for],[aria-controls],[aria-activedescendant]')].flatMap(e =>
   ['aria-describedby','aria-labelledby','for','aria-controls','aria-activedescendant'].flatMap(a => (e.getAttribute(a) || '').split(/\s+/).filter(Boolean)))
   .filter(r => !document.getElementById(r));
 const unnamed = [...document.querySelectorAll('input:not([type=checkbox]),select,textarea,[role=combobox]')].filter(e => !(e.labels && e.labels.length) && !e.getAttribute('aria-labelledby') && !e.getAttribute('aria-label')).map(e => e.id || e.outerHTML.slice(0, 60));
 const main = document.querySelector('.s7-main'); const sb = document.querySelector('.ip-savebar');
 const pnl = document.querySelector('.ip-sidepanel'); let under = null;
 if (pnl) { const px = pnl.getBoundingClientRect().left; under = [...document.querySelectorAll('.s13-vp *')].filter(e => { const r = e.getBoundingClientRect(); return r.width && r.right > px + 1 && r.top < 1024; }).map(e => e.className.baseVal ?? e.className).slice(0, 4); }
 const fx = document.querySelector('#fx, #pfx'); const tall = [...document.querySelectorAll('.s7-main > *')].map(e => Math.round(e.getBoundingClientRect().bottom));
 const lum = c => { const m = c.match(/[\d.]+/g).map(Number); const f = v => { v /= 255; return v <= .03928 ? v / 12.92 : ((v + .055) / 1.055) ** 2.4 }; return [.2126 * f(m[0]) + .7152 * f(m[1]) + .0722 * f(m[2]), m.length > 3 ? m[3] : 1] };
 const bgOf = e => { for (let x = e; x; x = x.parentElement) { const c = getComputedStyle(x).backgroundColor; const [, a] = lum(c); if (a > .5) return c; } return 'rgb(255,255,255)'; };
 const low = []; for (const e of document.querySelectorAll('body *')) { if (e.closest('[inert],.ip-sr,svg,[hidden]') || e.closest(':disabled,[aria-disabled=true]')) continue;
   const t = [...e.childNodes].some(nd => nd.nodeType === 3 && nd.textContent.trim()); if (!t) continue; const r = e.getBoundingClientRect(); if (!r.width) continue;
   const [l1] = lum(getComputedStyle(e).color), [l2] = lum(bgOf(e)); const cr = (Math.max(l1, l2) + .05) / (Math.min(l1, l2) + .05); if (cr < 4.5) low.push(e.textContent.trim().slice(0, 24) + ' ' + cr.toFixed(2)); }
 const touch = innerWidth < 500; const foc = [...document.querySelectorAll('a[href],button,input,[tabindex],label.ip-check')].filter(e => !e.closest('[inert],[hidden]') && !(e.matches('input.ip-sr')) && e.getAttribute('tabindex') !== '-1');
 const small = foc.map(e => { const t = e.matches('input') ? (e.closest('.ip-input') || e) : e; const r = t.getBoundingClientRect(); return [e, r]; })
   .filter(([e, r]) => r.width && Math.min(r.width, r.height) < (touch ? 44 : 24)).map(([e, r]) => (e.id || e.textContent.trim().slice(0, 16) || e.className) + ' ' + Math.round(r.width) + 'x' + Math.round(r.height));
 const order = foc.filter(e => !e.matches('label.ip-check')).map(e => e.getAttribute('aria-label') || e.id || e.textContent.trim().slice(0, 14));
 const barT = sb ? sb.getBoundingClientRect().top : 1e9; const vis = sel => [...document.querySelectorAll(sel)].map(e => { const r = e.getBoundingClientRect(); return Math.round(r.top) + '-' + Math.round(r.bottom) + (r.bottom > barT + 1 ? ' UNDER' : ''); });
 return {low, small, order, fxpos: vis('#fx'), dlg: vis('.ip-dialog'), dlgkeys: vis('.s4s-dlgfoot .ip-btn'), banpos: vis('.s7-main > .ip-banner'), dup, broken: refs, unnamed, sw: document.documentElement.scrollWidth, under,
   mainBottom: main ? Math.round(main.getBoundingClientRect().bottom) : null, parts: tall, barTop: sb ? Math.round(sb.getBoundingClientRect().top) : null,
   barH: sb ? Math.round(sb.getBoundingClientRect().height) : null,
   sumLines: sb ? [...sb.querySelectorAll('.ip-ftot-sum__line')].map(l => Math.round(l.getBoundingClientRect().height)) : null,
   allkey: [...document.querySelectorAll('.ip-sexrows__allkey')].map(k => k.getAttribute('aria-pressed') + ' ' + Math.round(k.getBoundingClientRect().width) + 'x' + Math.round(k.getBoundingClientRect().height)),
   natural: (() => { const m = document.querySelector('.s7-main'); return m && m.lastElementChild ? m.lastElementChild.getBoundingClientRect().bottom + parseFloat(getComputedStyle(m).paddingBottom) + sb.getBoundingClientRect().height : null })(), body: (() => { const b = document.querySelector('.s13-pbody'); return b ? [b.scrollHeight, b.clientHeight] : null })()};
}"""
async def main():
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for k, (st, w, h, d) in enumerate(ORDER):
            vh = json.loads(HJ.read_text()).get(st, h) if (k == 0 and HJ.exists() and '--heights' not in sys.argv) else h  # the tall default: its whole height
            pg = await b.new_page(viewport={'width': w, 'height': vh})
            await pg.goto(f'file:///tmp/{d}/{st}.html'); await pg.wait_for_timeout(1000)
            r = await pg.evaluate(CHECK); print(st, w, r)
            if k == 0 and '--heights' in sys.argv:
                hs = json.loads(HJ.read_text()) if HJ.exists() else {}
                hs[st] = int(-(-r['natural'] // 4) * 4); HJ.write_text(json.dumps(hs, indent=1)); print('height', st, hs[st])
            await pg.screenshot(path=str(R / 'design/shots' / ('S10.png' if k == 0 else f'S10-{DESK}-1280.png' if w == 1280 else f'S10-{k}.png')), full_page=(k == 0))
            await pg.close()
        await b.close()
asyncio.run(main())
