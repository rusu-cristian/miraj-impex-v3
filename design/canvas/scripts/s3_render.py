"""python3 s3_render.py [--heights] [--widths] [--w 1280] [stem ...] : render the S3 boards listed by s3_build.py in /tmp/s4s/s3.json
(desktops from /tmp/s4s, the phone from /tmp/s2: plain wrappers that keep only the 1024 / 844 floor) at their canvas width
(board_widths.json, else 1440 / 390; --w renders the desktops at another width, shots get a -<w> suffix), shots
design/shots/S3.png, S3-<n>.png (s3_build order). Prints the natural height (desktop: main content + 24, floor 1024; phone: the
page, floor 844), the width the tables need (D167: base width + the widest region's overflow, up to the next 8 px; measured
with the wide CSS dropped), body-row heights, the first column's width, duplicate ids, broken ARIA references, small targets.
--heights writes s3_heights.json, --widths the S3 entries of board_widths.json; then rebuild s3_build.py --canvas.
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
 const foc = [...document.querySelectorAll('a[href],button,input,[tabindex]')].filter(e => !e.closest('[inert],[hidden],.ip-sr,[aria-hidden="true"]') && e.getAttribute('tabindex') !== '-1' && getComputedStyle(e).display !== 'none' && e.getClientRects().length);
 const small = foc.map(e => [e, e.getBoundingClientRect()]).filter(([e, r]) => r.width && Math.min(r.width, r.height) < (touch ? 44 : 24))
   .map(([e, r]) => (e.getAttribute('aria-label') || e.textContent.trim().slice(0, 18) || e.className) + ' ' + Math.round(r.width) + 'x' + Math.round(r.height));
 const regs = [...document.querySelectorAll('.ip-tablescroll')].map(r => r.scrollWidth - r.clientWidth);
 const rows = [...new Set([...document.querySelectorAll('.ip-ptable tbody tr:not([hidden]), .ip-ptable tfoot tr')].map(t => Math.round(t.getBoundingClientRect().height * 2) / 2))];
 const first = [...document.querySelectorAll('.ip-ptable thead th:first-child')].map(t => Math.round(t.getBoundingClientRect().width));
 const cols = [...document.querySelectorAll('.ip-ptable thead tr:last-child th')].map(t => Math.round(t.getBoundingClientRect().width * 10) / 10);
 const spine = [...document.querySelectorAll('.ip-ptable thead tr:first-child th[rowspan]')].map(t => Math.round(t.getBoundingClientRect().width * 10) / 10);
 const tw = [...document.querySelectorAll('.ip-ptable')].map(t => Math.round(t.getBoundingClientRect().width));
 const clip = [...document.querySelectorAll('.ip-fsum *,.ip-pagehead *,.ip-readout *')].filter(e => e.scrollWidth > e.clientWidth + 1 && getComputedStyle(e).overflow != 'visible').map(e => e.className + ':' + e.textContent.slice(0, 30)).slice(0, 6);
 const h = document.querySelector('.ip-shell__main') ? Math.ceil(document.querySelector('.ip-shell__main').lastElementChild.getBoundingClientRect().bottom + 24) : Math.ceil(document.documentElement.scrollHeight);
 return {h, dup, broken, small: small.slice(0, 12), regs, rows, first, cols, spine, tw, clip, sw: document.documentElement.scrollWidth};
}"""
async def main():
    a = sys.argv[1:]; W = None
    if '--w' in a: i = a.index('--w'); W = int(a[i + 1]); del a[i:i + 2]
    save, wid = '--heights' in a, '--widths' in a; only = [x for x in a if not x.startswith('--')]
    boards = json.loads(pathlib.Path('/tmp/s4s/s3.json').read_text())
    import canvas_rows as CR
    hp = HERE / 's3_heights.json'; hs = json.loads(hp.read_text()) if hp.exists() else {}
    bw = dict(CR.WIDTHS)
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for i, bd in enumerate(boards):
            st = bd['stem']
            if only and st not in only: continue
            phone = 'telefon' in st; base = 390 if phone else 1440
            vw = base if wid else CR.width(st, base) if phone else (W or CR.width(st, base))
            pg = await b.new_page(viewport={'width': vw, 'height': 844 if phone else 1024})
            await pg.goto(f'file:///tmp/{"s2" if phone else "s4s"}/{st}.html'); await pg.wait_for_timeout(1200)
            if wid: await pg.evaluate("document.querySelectorAll('.s4s-wide,.s4ns-wide').forEach(e => { e.classList.remove('s4s-wide', 's4ns-wide'); if (e.style.width) e.style.width = '390px'; })")
            r = await pg.evaluate(CHECK)
            h = max(r['h'], 844 if phone else 1024)
            over = max(r['regs'] or [0]); need = base + (-(-over // 8) * 8 if over else 0)
            print(st, vw, 'h', h, 'regions', r['regs'], 'needs w', need, 'tables', r['tw'], 'rows', r['rows'], 'first', r['first'], 'spine', r['spine'],
                  'sw', r['sw'], 'dup', r['dup'], 'broken', r['broken'], 'small', r['small'], 'clip', r['clip'])
            print('  leaf cols', r['cols'])
            if wid:
                if need > base: bw[st] = need
                else: bw.pop(st, None)
            elif not W: hs[st] = h
            if not wid:
                if phone: await pg.set_viewport_size({'width': vw, 'height': h}); await pg.wait_for_timeout(300)
                name = ('S3' if i == 0 else f'S3-{i}') + (f'-{W}' if W else '') + '.png'
                await pg.screenshot(path=str(R / 'design/shots' / name), full_page=not phone)
            await pg.close()
        await b.close()
    if wid:
        (HERE / 'board_widths.json').write_text(json.dumps(dict(sorted(bw.items())), indent=1, ensure_ascii=False) + '\n'); print('board_widths.json', {k: v for k, v in bw.items() if k.startswith('S3')})
    if save and not W and not wid:
        hp.write_text(json.dumps(hs, indent=1, ensure_ascii=False) + '\n'); print('s3_heights.json', hs)
asyncio.run(main())
