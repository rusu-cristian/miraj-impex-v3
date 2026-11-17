"""python3 s16_render.py [--measure] [stems…] : render the S16 boards from /tmp/s16 (phones 390×844, desktop 1440×1024 and
1280×1024) to design/shots/S16.png, S16-1 … S16-4.png (canvas order) + S16-4-1280.png, and report sideways scroll,
duplicate ids, broken ARIA references, unnamed form controls, text under 4.5:1, targets under 44 (phone) / 24 (desktop),
the tab order, the sheet body overflow, and page elements under the open SidePanel.
--measure: on the desktop board (built with scroll 0) read the top of the Jurnal block, write scroll = top − space-6 into
s16_heights.json ("S16-corecteaza--desktop@scroll"); rebuild with s16_build.py <repo> s16_heights.json afterwards."""
import asyncio, json, pathlib, sys
from playwright.async_api import async_playwright
R = pathlib.Path('/home/claude/miraj-impex-v3')
HJ = R / 'design/canvas/scripts/s16_heights.json'
ORDER = [('S16-corecteaza', 390, 844), ('S16-corecteaza--motiv-lipsa', 390, 844), ('S16-corecteaza--refuzat', 390, 844),
         ('S16-corecteaza--sterge-confirmare', 390, 844), ('S16-corecteaza--desktop', 1440, 1024), ('S16-corecteaza--desktop', 1280, 1024)]
CHECK = r"""() => {
 const ids = [...document.querySelectorAll('[id]')].map(e => e.id); const dup = ids.filter((x, i) => ids.indexOf(x) !== i);
 const refs = [...document.querySelectorAll('[aria-describedby],[aria-labelledby],[for],[aria-controls],[aria-activedescendant]')].flatMap(e =>
   ['aria-describedby','aria-labelledby','for','aria-controls','aria-activedescendant'].flatMap(a => (e.getAttribute(a) || '').split(/\s+/).filter(Boolean)))
   .filter(r => !document.getElementById(r));
 const unnamed = [...document.querySelectorAll('input:not([type=checkbox]),select,textarea,[role=combobox]')].filter(e => !(e.labels && e.labels.length) && !e.getAttribute('aria-labelledby') && !e.getAttribute('aria-label')).map(e => e.id || e.outerHTML.slice(0, 60));
 const pnl = document.querySelector('.s13-panel'); let under = null;
 if (pnl) { const px = pnl.getBoundingClientRect().left; under = [...document.querySelectorAll('.s13-vp *')].filter(e => { const r = e.getBoundingClientRect(); return r.width && r.right > px + 1 && r.top < 1024 && r.bottom > 0; }).map(e => e.className.baseVal ?? e.className).slice(0, 4); }
 const lum = c => { const m = c.match(/[\d.]+/g).map(Number); const f = v => { v /= 255; return v <= .03928 ? v / 12.92 : ((v + .055) / 1.055) ** 2.4 }; return [.2126 * f(m[0]) + .7152 * f(m[1]) + .0722 * f(m[2]), m.length > 3 ? m[3] : 1] };
 const bgOf = e => { for (let x = e; x; x = x.parentElement) { const c = getComputedStyle(x).backgroundColor; const [, a] = lum(c); if (a > .5) return c; } return 'rgb(255,255,255)'; };
 const low = []; for (const e of document.querySelectorAll('body *')) { if (e.closest('[inert],.ip-sr,svg,[hidden]') || e.closest(':disabled,[aria-disabled=true]')) continue;
   const t = [...e.childNodes].some(nd => nd.nodeType === 3 && nd.textContent.trim()); if (!t) continue; const r = e.getBoundingClientRect(); if (!r.width || r.bottom < 0 || r.top > innerHeight) continue;
   const [l1] = lum(getComputedStyle(e).color), [l2] = lum(bgOf(e)); const cr = (Math.max(l1, l2) + .05) / (Math.min(l1, l2) + .05); if (cr < 4.5) low.push(e.textContent.trim().slice(0, 24) + ' ' + cr.toFixed(2)); }
 const touch = innerWidth < 500;
 const scope = document.querySelector('.ip-scrim') || document.querySelector('.s13-panel, .ip-sidepanel--sheet') || document.body;
 const foc = [...scope.querySelectorAll('a[href],button,input,textarea,[tabindex]')].filter(e => !e.closest('[inert],[hidden]') && e.getAttribute('tabindex') !== '-1');
 const small = foc.map(e => { const t = e.matches('input,textarea') ? (e.closest('.ip-input') || e) : e; const r = t.getBoundingClientRect(); return [e, r]; })
   .filter(([e, r]) => r.width && Math.min(r.width, r.height) < (touch ? 44 : 24)).map(([e, r]) => (e.id || e.textContent.trim().slice(0, 16) || e.className) + ' ' + Math.round(r.width) + 'x' + Math.round(r.height));
 const order = foc.map(e => e.getAttribute('aria-label') || e.id || e.textContent.trim().slice(0, 18));
 const b = document.querySelector('.s13-pbody'); const jr = document.getElementById('h-jurnal');
 const key = document.querySelector('[aria-controls="cp"]'); const kr = key ? key.getBoundingClientRect() : null;
 return {low, small, order, dup, broken: refs, unnamed, sw: document.documentElement.scrollWidth, under,
   body: b ? [b.scrollHeight, b.clientHeight] : null, jurnal: jr ? Math.round(jr.closest('section').getBoundingClientRect().top) : null,
   key: kr ? [Math.round(kr.left), Math.round(kr.top), Math.round(kr.right), Math.round(kr.bottom)] : null};
}"""
async def main():
    measure = '--measure' in sys.argv; only = [a for a in sys.argv[1:] if not a.startswith('--')]
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for k, (st, w, h) in enumerate(ORDER):
            if only and st not in only: continue
            if measure and st != 'S16-corecteaza--desktop': continue
            pg = await b.new_page(viewport={'width': w, 'height': h})
            await pg.goto(f'file:///tmp/s16/{st}.html'); await pg.wait_for_timeout(1000)
            res = await pg.evaluate(CHECK); print(st, w, res)
            if measure:
                hs = json.loads(HJ.read_text()) if HJ.exists() else {}
                hs['S16-corecteaza--desktop@scroll'] = res['jurnal'] - 24; HJ.write_text(json.dumps(hs, indent=1) + '\n')
                print('scroll', hs['S16-corecteaza--desktop@scroll']); await pg.close(); break
            name = 'S16.png' if k == 0 else 'S16-4-1280.png' if w == 1280 else f'S16-{k}.png'
            await pg.screenshot(path=str(R / 'design/shots' / name))
            await pg.close()
        await b.close()
asyncio.run(main())
