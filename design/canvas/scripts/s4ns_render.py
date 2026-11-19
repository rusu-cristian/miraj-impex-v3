"""python3 s4ns_render.py [--heights] [stem ...] : render the S4NS boards listed by s4ns_build.py in /tmp/s4s/s4ns.json
(desktops from /tmp/s4s and phones from /tmp/s2, each at its canvas width: board_widths.json, else 1440 / 390) to design/shots/S4NS-<n>.png,
print the natural height (desktop: main content + 24, floor 1024; phone: the page, floor 844; the fixed sheet board 844), the
table overflow (desktop: the width its tables need, as s4s_render; phone: sideways scroll inside the region), duplicate ids,
broken ARIA references and phone targets under 44 px. --heights writes s4ns_heights.json (then rebuild s4ns_build.py --canvas).
Run only to re-measure (README › Never run for a row)."""
import sys, json, asyncio, pathlib
sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from playwright.async_api import async_playwright
HERE = pathlib.Path(__file__).resolve().parent; R = HERE.parents[2]
CHECK = r"""() => {
 const ids = [...document.querySelectorAll('[id]')].map(e => e.id); const dup = [...new Set(ids.filter((x, i) => ids.indexOf(x) !== i))];
 const broken = [...document.querySelectorAll('[aria-describedby],[aria-labelledby],[aria-controls],[for]')].flatMap(e =>
   ['aria-describedby','aria-labelledby','aria-controls','for'].flatMap(a => (e.getAttribute(a) || '').split(/\s+/).filter(Boolean))).filter(r => !document.getElementById(r));
 const touch = innerWidth < 500;
 const foc = [...document.querySelectorAll('a[href],button,input,[tabindex]')].filter(e => !e.closest('[inert],[hidden],.ip-sr') && e.getAttribute('tabindex') !== '-1' && getComputedStyle(e).display !== 'none' && e.getClientRects().length);
 const small = foc.map(e => [e, e.getBoundingClientRect()]).filter(([e, r]) => r.width && Math.min(r.width, r.height) < (touch ? 44 : 24))
   .map(([e, r]) => (e.getAttribute('aria-label') || e.textContent.trim().slice(0, 18) || e.className) + ' ' + Math.round(r.width) + 'x' + Math.round(r.height));
 const regs = [...document.querySelectorAll('.ip-tablescroll')].map(r => r.scrollWidth - r.clientWidth);
 const rows = [...new Set([...document.querySelectorAll('.ip-ptable tbody tr:not([hidden]), .ip-ptable tfoot tr')].map(t => Math.round(t.getBoundingClientRect().height)))];
 const first = [...document.querySelectorAll('.ip-ptable thead th:first-child')].map(t => Math.round(t.getBoundingClientRect().width));
 const main = document.querySelector('.ip-shell__main, .ip-screen');
 const h = document.querySelector('.ip-shell__main') ? Math.ceil(document.querySelector('.ip-shell__main').lastElementChild.getBoundingClientRect().bottom + 24) : Math.ceil(document.documentElement.scrollHeight);
 return {h, dup, broken, small: small.slice(0, 12), regs, rows, first, sw: document.documentElement.scrollWidth};
}"""
async def main():
    a = sys.argv[1:]; save = '--heights' in a; only = [x for x in a if not x.startswith('--')]
    boards = json.loads(pathlib.Path('/tmp/s4s/s4ns.json').read_text())
    import canvas_rows as CR
    hp = HERE / 's4ns_heights.json'; hs = json.loads(hp.read_text()) if hp.exists() else {}
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for i, bd in enumerate(boards, 1):
            st = bd['stem']
            if only and st not in only: continue
            phone = bd['w'] < 1000
            vw = CR.width(st, 390 if phone else 1440); vh = 844 if phone else 1024   # D167: wide boards render at their canvas width
            pg = await b.new_page(viewport={'width': vw, 'height': vh})
            await pg.goto(f'file:///tmp/{"s2" if phone else "s4s"}/{st}.html'); await pg.wait_for_timeout(1200)
            r = await pg.evaluate(CHECK)
            fixed = 'telefon-filtre' in st
            h = 844 if fixed else max(r['h'], 844 if phone else 1024)
            over = max(r['regs'] or [0])
            need = vw + (-(-over // 8) * 8 if over and not phone else 0)
            print(st, vw, 'h', h, 'regions', r['regs'], 'needs w', need, 'rows', r['rows'], 'first col', r['first'], 'sw', r['sw'],
                  'dup', r['dup'], 'broken', r['broken'], 'small', r['small'])
            hs[st] = h
            if phone and not fixed:  # S4NS fix: the shot at the board's real height (the tab bar at its foot, not at an 844 viewport)
                await pg.set_viewport_size({'width': vw, 'height': h}); await pg.wait_for_timeout(300)
            await pg.screenshot(path=str(R / 'design/shots' / f'S4NS-{i}.png'), full_page=not phone)
            await pg.close()
        await b.close()
    if save: hp.write_text(json.dumps(hs, indent=1, ensure_ascii=False) + '\n'); print('s4ns_heights.json', hs)
asyncio.run(main())
