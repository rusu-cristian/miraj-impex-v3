"""Render SexCountRow preview at 1440 and 390: python3 sys07_render.py <tag>"""
import sys, asyncio, pathlib
from playwright.async_api import async_playwright
ROOT = pathlib.Path(__file__).resolve().parents[1]  # design/system from this script's place (design/system/components)
OUT = pathlib.Path('/tmp/sys07'); OUT.mkdir(exist_ok=True)
async def main():
    tag = sys.argv[1] if len(sys.argv) > 1 else 'r'
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        frag = (ROOT/'components'/'SexCountRow'/'preview.html').read_text(encoding='utf-8')
        html = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8">
<link rel="stylesheet" href="file://{ROOT}/tokens.css">
<link rel="stylesheet" href="file://{ROOT}/components/bundle.css">
<link rel="stylesheet" href="file://{ROOT}/components/additions.css">
<style>body{{margin:0;padding:24px;background:var(--ground)}}</style></head><body>{frag}</body></html>'''
        pf = OUT/'SexCountRow.html'; pf.write_text(html, encoding='utf-8')
        for w in (1440, 390):
            pg = await b.new_page(viewport={'width': w, 'height': 900})
            await pg.goto(f'file://{pf}'); await pg.wait_for_timeout(1200)
            if w == 390:
                await pg.add_style_tag(content='body{padding:0} .ip-root > .ip-row{flex-direction:column;gap:24px}')
            ov = await pg.evaluate("""() => [...document.querySelectorAll('.ip-sexrows,.ip-sexrow,.ip-sexrow__head,.ip-sexrows__total,.ip-field')].filter(e => e.scrollWidth > e.clientWidth+1).map(e=>e.className)""")
            dw = await pg.evaluate("document.documentElement.scrollWidth")
            inv = await pg.evaluate("[...document.querySelectorAll('input')].filter(i=>!i.checkValidity()).map(i=>i.id+':'+i.value)")
            print(w, 'overflow:', ov[:6], 'docw', dw, 'invalid:', inv)
            await pg.screenshot(path=str(OUT/f'SexCountRow-{w}-{tag}.png'), full_page=True)
            await pg.close()
        await b.close()
asyncio.run(main())
