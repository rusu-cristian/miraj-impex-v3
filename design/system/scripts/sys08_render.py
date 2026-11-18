"""Render SYS08 previews at 1440 and 390: python3 sys08_render.py <tag>"""
import sys, asyncio, pathlib
from playwright.async_api import async_playwright
ROOT = pathlib.Path(__file__).resolve().parents[1]  # design/system from this script's place (design/system/scripts), as sys11_render
OUT = pathlib.Path('/tmp/sys08'); OUT.mkdir(exist_ok=True)
async def main():
    tag = sys.argv[1] if len(sys.argv) > 1 else 'r'
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for comp in ('MissingRow', 'EntryKeys', 'UnpricedCount'):
            frag = (ROOT/'components'/comp/'preview.html').read_text(encoding='utf-8')
            html = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8">
<link rel="stylesheet" href="file://{ROOT}/tokens.css">
<link rel="stylesheet" href="file://{ROOT}/components/bundle.css">
<link rel="stylesheet" href="file://{ROOT}/components/additions.css">
<style>body{{margin:0;padding:24px;background:var(--ground)}}</style></head><body>{frag}</body></html>'''
            pf = OUT/f'{comp}.html'; pf.write_text(html, encoding='utf-8')
            pg = await b.new_page(viewport={'width': 1440, 'height': 900})
            await pg.goto(f'file://{pf}'); await pg.wait_for_timeout(1000)
            ov = await pg.evaluate("""() => [...document.querySelectorAll('.ip-ev,.ip-entrykey,.ip-unpriced,.ip-unpriced li,.ip-missing__empty')].filter(e => e.scrollWidth > e.clientWidth+1).map(e=>e.className)""")
            hs = await pg.evaluate("""() => [...document.querySelectorAll('.ip-ev--missing,.ip-entrykey')].map(e=>Math.round(e.getBoundingClientRect().height))""")
            print(comp, 'overflow:', ov[:6], 'heights:', sorted(set(hs)))
            await pg.screenshot(path=str(OUT/f'{comp}-{tag}.png'), full_page=True)
            await pg.close()
        await b.close()
asyncio.run(main())
