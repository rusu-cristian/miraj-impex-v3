"""Render SYS09 previews (EventRowActions, ChangeRow) at 1440: python3 sys09_render.py <tag>"""
import sys, asyncio, pathlib
from playwright.async_api import async_playwright
ROOT = pathlib.Path('/home/claude/miraj-impex-v3/design/system')
OUT = pathlib.Path('/tmp/sys09'); OUT.mkdir(exist_ok=True)
async def main():
    tag = sys.argv[1] if len(sys.argv) > 1 else 'r'
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for comp in ('EventRowActions', 'ChangeRow'):
            frag = (ROOT/'components'/comp/'preview.html').read_text(encoding='utf-8')
            html = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8">
<link rel="stylesheet" href="file://{ROOT}/tokens.css">
<link rel="stylesheet" href="file://{ROOT}/components/bundle.css">
<link rel="stylesheet" href="file://{ROOT}/components/additions.css">
<style>body{{margin:0;padding:24px;background:var(--ground)}}</style></head><body>{frag}</body></html>'''
            pf = OUT/f'{comp}.html'; pf.write_text(html, encoding='utf-8')
            pg = await b.new_page(viewport={'width': 1440, 'height': 900})
            await pg.goto(f'file://{pf}'); await pg.wait_for_timeout(800)
            ov = await pg.evaluate("""() => [...document.querySelectorAll('.ip-ev,.ip-ev__main,.ip-chg__vals,.ip-ev__acts')].filter(e => e.scrollWidth > e.clientWidth+1).map(e=>e.className)""")
            hs = await pg.evaluate("""() => [...document.querySelectorAll('.ip-ev--act,.ip-ev--change')].map(e=>Math.round(e.getBoundingClientRect().height))""")
            tg = await pg.evaluate("""() => [...document.querySelectorAll('.ip-ev__acts .ip-btn,.ip-ev__more .ip-iconbtn,.ip-ev__mod .ip-link,.ip-ev__menu .ip-menu__item')].filter(e=>e.offsetParent).map(e=>{const r=e.getBoundingClientRect();return Math.round(r.height)})""")
            print(comp, 'overflow:', ov[:6], 'row heights:', sorted(set(hs)), 'target heights:', sorted(set(tg)))
            await pg.screenshot(path=str(OUT/f'{comp}-{tag}.png'), full_page=True)
            await pg.close()
        await b.close()
asyncio.run(main())
