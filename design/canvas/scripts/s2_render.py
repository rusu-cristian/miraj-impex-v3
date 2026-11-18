"""python3 s2_render.py [--heights] : render the S2 boards (phones 390 from /tmp/s2 at their full height, desktops 1440 from
/tmp/s4s, plus 1280) to design/shots/S2.png (vet phone), S2-1.png (manager phone), S2-2.png (manager desktop),
S2-3.png (owner desktop), S2-2-1280.png, S2-3-1280.png, and report sideways scroll, duplicate ids, broken ARIA references,
unnamed controls, text under 4.5:1, targets under 44 (phone) / 24 (desktop) and the tab order.
--states (S2S) also renders the state boards listed by s2s_build.py in /tmp/s2/states.json (phones 390×844 from /tmp/s2,
desktops 1440 from /tmp/s4s) to design/shots/S2-<state>.png; --only-states skips the 6 default shots.
--viewer (S2V) renders only the Observator phone (390, full page) and desktop (1440) to design/shots/S2-observator.png and
S2-observator-desktop.png at their board heights.
--heights writes each board's natural height (max of 844 / 1024 and the content, rounded up to 4) to s2_heights.json
(then rebuild with s2_build.py)."""
import asyncio, json, pathlib, sys
from playwright.async_api import async_playwright
R = pathlib.Path(__file__).resolve().parents[3]
HJ = R / 'design/canvas/scripts/s2_heights.json'
ORDER = [('S2-acasa', 390, 844, 's2', 'S2.png'), ('S2-acasa--manager', 390, 844, 's2', 'S2-1.png'),
         ('S2-acasa--manager-desktop', 1440, 1024, 's4s', 'S2-2.png'), ('S2-acasa--proprietar-desktop', 1440, 1024, 's4s', 'S2-3.png'),
         ('S2-acasa--manager-desktop', 1280, 1024, 's4s', 'S2-2-1280.png'), ('S2-acasa--proprietar-desktop', 1280, 1024, 's4s', 'S2-3-1280.png')]
CHECK = r"""() => {
 const ids = [...document.querySelectorAll('[id]')].map(e => e.id); const dup = ids.filter((x, i) => ids.indexOf(x) !== i);
 const refs = [...document.querySelectorAll('[aria-describedby],[aria-labelledby],[for],[aria-controls]')].flatMap(e =>
   ['aria-describedby','aria-labelledby','for','aria-controls'].flatMap(a => (e.getAttribute(a) || '').split(/\s+/).filter(Boolean))).filter(r => !document.getElementById(r));
 const unnamed = [...document.querySelectorAll('input,select,textarea,[role=combobox]')].filter(e => !(e.labels && e.labels.length) && !e.getAttribute('aria-labelledby') && !e.getAttribute('aria-label')).map(e => e.id || e.outerHTML.slice(0, 60));
 const lum = c => { const m = c.match(/[\d.]+/g).map(Number); const f = v => { v /= 255; return v <= .03928 ? v / 12.92 : ((v + .055) / 1.055) ** 2.4 }; return [.2126 * f(m[0]) + .7152 * f(m[1]) + .0722 * f(m[2]), m.length > 3 ? m[3] : 1] };
 const bgOf = e => { for (let x = e; x; x = x.parentElement) { const c = getComputedStyle(x).backgroundColor; const [, a] = lum(c); if (a > .5) return c; } return 'rgb(255,255,255)'; };
 const low = []; for (const e of document.querySelectorAll('body *')) { if (e.closest('[inert],.ip-sr,svg,[hidden]') || e.closest(':disabled,[aria-disabled=true]')) continue;
   const t = [...e.childNodes].some(nd => nd.nodeType === 3 && nd.textContent.trim()); if (!t) continue; const r = e.getBoundingClientRect(); if (!r.width) continue;
   const [l1] = lum(getComputedStyle(e).color), [l2] = lum(bgOf(e)); const cr = (Math.max(l1, l2) + .05) / (Math.min(l1, l2) + .05); if (cr < 4.5) low.push(e.textContent.trim().slice(0, 24) + ' ' + cr.toFixed(2)); }
 const touch = innerWidth < 500;
 const foc = [...document.querySelectorAll('a[href],button,input,[tabindex]')].filter(e => !e.closest('[inert],[hidden]') && e.getAttribute('tabindex') !== '-1' && getComputedStyle(e).display !== 'none' && e.getClientRects().length);
 const small = foc.map(e => [e, e.getBoundingClientRect()]).filter(([e, r]) => r.width && Math.min(r.width, r.height) < (touch ? 44 : 24))
   .map(([e, r]) => (e.getAttribute('aria-label') || e.textContent.trim().slice(0, 18) || e.className) + ' ' + Math.round(r.width) + 'x' + Math.round(r.height));
 const order = foc.map(e => (e.getAttribute('aria-label') || e.textContent.trim()).slice(0, 18));
 const over = [...document.querySelectorAll('.ip-ev__main, .ip-ev__title, .ip-entrykey__txt > *, .ip-readout__cap > *, .ip-fsum__head *')].filter(e => e.scrollWidth > e.clientWidth + 1).map(e => e.textContent.trim().slice(0, 30));
 const root = document.querySelector('.ip-root');
 return {low, small, dup, broken: refs, unnamed, sw: document.documentElement.scrollWidth, over, natural: Math.ceil(document.documentElement.scrollHeight), order};
}"""
if '--states' in sys.argv or '--only-states' in sys.argv:
    if '--only-states' in sys.argv: ORDER = []
    for b in json.loads(pathlib.Path('/tmp/s2/states.json').read_text()):
        ORDER.append((b['stem'], b['w'], b['h'], 's2' if b['w'] == 390 else 's4s', 'S2-' + b['stem'].split('--', 1)[1] + '.png'))
if '--viewer' in sys.argv:  # S2V: only the 2 Observator boards (s2s_build.py writes /tmp/s2/viewer.json), shot at their board heights
    ORDER = [(v['stem'], v['w'], v['h'], 's2' if v['w'] == 390 else 's4s', 'S2-' + v['stem'].split('--', 1)[1] + '.png')
             for v in json.loads(pathlib.Path('/tmp/s2/viewer.json').read_text())]
async def main():
    hs = json.loads(HJ.read_text()) if HJ.exists() else {}
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        new = {}
        for st, w, h, d, shot in ORDER:
            vh = h if '--heights' in sys.argv else hs.get(st, h)
            pg = await b.new_page(viewport={'width': w, 'height': vh})
            await pg.goto(f'file:///tmp/{d}/{st}.html'); await pg.wait_for_timeout(800)
            r = await pg.evaluate(CHECK); print(st, w, json.dumps(r, ensure_ascii=False))
            if w in (390, 1440): new[st] = max(h, int(-(-r['natural'] // 4) * 4))
            await pg.screenshot(path=str(R / 'design/shots' / shot), full_page=True)
            await pg.close()
        await b.close()
    if '--heights' in sys.argv:
        HJ.write_text(json.dumps({**hs, **new}, indent=1)); print('heights', new)
asyncio.run(main())
