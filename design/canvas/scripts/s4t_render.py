"""python3 s4t_render.py <out_heights.json> <stem>... : render /tmp/s4s/<stem>.html (plain wrapper written by s4t_build.py,
linking the repo's tokens/bundle/additions) at its canvas width (D167: board_widths.json, else 1440), shot
design/shots/S4T-<n>.png, record the board height (main content + 24) and print the width its tables need
(s4s_render.py --widths writes board_widths.json)."""
import sys, json, asyncio, pathlib
sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from playwright.async_api import async_playwright
R = pathlib.Path(__file__).resolve().parents[3]
ORDER = ['S4-ferma--pasari', 'S4-ferma--hrana', 'S4-ferma--greutate', 'S4-ferma--productie', 'S4-ferma--pasari--pe-hale']
async def main():
    hp = pathlib.Path(sys.argv[1]); hs = json.loads(hp.read_text()) if hp.exists() else {}
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        import canvas_rows as CR
        for st in sys.argv[2:]:
            vw = CR.width(st, 1440)
            pg = await b.new_page(viewport={'width': vw, 'height': 1000})
            await pg.goto(f'file:///tmp/s4s/{st}.html'); await pg.wait_for_timeout(1200)
            mh = await pg.evaluate("Math.ceil(document.querySelector('.ip-shell__main').lastElementChild.getBoundingClientRect().bottom + 24)")
            clip = await pg.evaluate("""[...document.querySelectorAll('.ip-tablescroll,.ip-readout *,.ip-stat *,.ip-ev *,.ip-filterbar *')].filter(e=>e.scrollWidth>e.clientWidth+1 && getComputedStyle(e).overflow!='visible').map(e=>e.className+':'+e.textContent.slice(0,40)).slice(0,8)""")
            lists = await pg.evaluate("[...document.querySelectorAll('.ip-list--act')].map(l=>Math.round(l.getBoundingClientRect().width))")
            ids = await pg.evaluate("(()=>{const c={};document.querySelectorAll('[id]').forEach(e=>c[e.id]=(c[e.id]||0)+1);const d=Object.keys(c).filter(k=>c[k]>1);const refs=[];document.querySelectorAll('[aria-describedby],[aria-labelledby],[aria-controls]').forEach(e=>['aria-describedby','aria-labelledby','aria-controls'].forEach(a=>(e.getAttribute(a)||'').split(/\\s+/).filter(Boolean).forEach(i=>{if(!document.getElementById(i))refs.push(a+':'+i)})));return {dup:d,broken:refs}})()")
            over = await pg.evaluate("Math.max(0, ...[...document.querySelectorAll('.ip-tablescroll')].map(r => r.scrollWidth - r.clientWidth))")
            print(st, vw, 'main h', mh, 'table overflow', over, 'lists', lists, 'clip', clip, ids)
            hs[st] = max(mh, 1024)
            await pg.screenshot(path=str(R / 'design/shots' / f'S4T-{ORDER.index(st) + 1}.png'), full_page=True)
            await pg.close()
        await b.close()
    hp.write_text(json.dumps(hs, indent=1))
asyncio.run(main())
