"""python3 s11_render.py <out_heights.json> <stem>... : render design/canvas/project/<stem>.dc.html at 1440 (repo mirror), shot design/shots/S11[-n].png, record main height."""
import sys, json, asyncio, pathlib
from playwright.async_api import async_playwright
R = pathlib.Path(__file__).resolve().parents[3]
ORDER = ['S11-rapoarte-hrana', 'S11-rapoarte-hrana--se-incarca', 'S11-rapoarte-hrana--manager', 'S11-rapoarte-hrana--observator', 'S11-rapoarte-hrana--fara-livrari-in-perioada', 'S11-rapoarte-hrana--modifica-pret', 'S11-rapoarte-hrana--modifica-pret--eroare']
async def main():
    hp = pathlib.Path(sys.argv[1]); hs = json.loads(hp.read_text()) if hp.exists() else {}
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for st in sys.argv[2:]:
            pg = await b.new_page(viewport={'width': 1440, 'height': 1000})
            await pg.goto(f'file:///tmp/s4s/{st}.html'); await pg.wait_for_timeout(1200)
            mh = await pg.evaluate("Math.ceil(document.querySelector('.ip-shell__main').lastElementChild.getBoundingClientRect().bottom + 24)")
            clip = await pg.evaluate("""[...document.querySelectorAll('.ip-table *,.ip-pagehead *,.ip-readout *,.ip-filterbar *')].filter(e=>e.scrollWidth>e.clientWidth+1 && getComputedStyle(e).overflow!='visible').map(e=>e.className+':'+e.textContent.slice(0,40)).slice(0,8)""")
            tw = await pg.evaluate("[...document.querySelectorAll('.ip-table')].map(t=>Math.round(t.getBoundingClientRect().width)).concat([document.querySelector('.ip-shell__main').getBoundingClientRect().width, document.documentElement.scrollWidth])")
            print(st, 'main h', mh, 'widths', tw, 'clip', clip)
            hs[st] = max(mh, 1024)
            n = ORDER.index(st); name = 'S11.png' if n == 0 else f'S11-{n}.png'
            await pg.screenshot(path=str(R / 'design/shots' / name), full_page=True)
            await pg.close()
        await b.close()
    hp.write_text(json.dumps(hs, indent=1))
asyncio.run(main())
