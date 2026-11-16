"""python3 s3_render.py <out_heights.json> <stem>... : render design/canvas/project/<stem>.dc.html at 1440 (repo mirror), shot design/shots/S3[-n].png, record main height."""
import sys, json, asyncio, pathlib
from playwright.async_api import async_playwright
R = pathlib.Path('/home/claude/miraj-impex-v3'); ORDER = ['S3-seria-2026', 'S3-seria-2026--se-incarca', 'S3-seria-2026--seria-inchisa', 'S3-seria-2026--fara-acces']
async def main():
    hp = pathlib.Path(sys.argv[1]); hs = json.loads(hp.read_text()) if hp.exists() else {}
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for st in sys.argv[2:]:
            pg = await b.new_page(viewport={'width': 1440, 'height': 1000})
            await pg.goto(f'file://{R}/design/canvas/project/{st}.dc.html'); await pg.wait_for_timeout(1500)
            mh = await pg.evaluate("Math.ceil(document.querySelector('.s4-main').lastElementChild.getBoundingClientRect().bottom + 24)")
            wide = await pg.evaluate("""[...document.querySelectorAll('.ip-fsum *,.ip-table *,.ip-pagehead *')].filter(e=>e.scrollWidth>e.clientWidth+1 && getComputedStyle(e).overflow!='visible').map(e=>e.className+':'+e.textContent.slice(0,40)).slice(0,8)""")
            tw = await pg.evaluate("[document.querySelector('.ip-table')?.getBoundingClientRect().width, document.querySelector('.s4-main').getBoundingClientRect().width, document.documentElement.scrollWidth]")
            print(st, 'main h', mh, 'widths', tw, 'clip', wide)
            hs[st] = max(mh, 1024)
            n = ORDER.index(st); name = 'S3.png' if n == 0 else f'S3-{n}.png'
            await pg.screenshot(path=str(R / 'design/shots' / name), full_page=True)
            await pg.close()
        await b.close()
    hp.write_text(json.dumps(hs, indent=1))
asyncio.run(main())
