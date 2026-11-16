"""Render SYS06 previews at 1440 and 390 and check sideways scroll: python3 sys06_render.py <tag>"""
import sys, asyncio, pathlib
from playwright.async_api import async_playwright
ROOT = pathlib.Path('/home/claude/miraj-impex-v3/design/system')
OUT = pathlib.Path('/tmp/sys06'); OUT.mkdir(exist_ok=True)
COMPS = ('GuideChart', 'GuideMissing', 'GuideChartVariants', 'RecordCodePlus')
async def main():
    tag = sys.argv[1] if len(sys.argv) > 1 else 'r'
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for comp in COMPS:
            frag = (ROOT/'components'/comp/'preview.html').read_text(encoding='utf-8')
            html = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width">
<link rel="stylesheet" href="file://{ROOT}/tokens.css">
<link rel="stylesheet" href="file://{ROOT}/components/bundle.css">
<link rel="stylesheet" href="file://{ROOT}/components/additions.css">
<style>body{{margin:0;background:var(--ground)}}</style></head><body>{frag}</body></html>'''
            pf = OUT/f'{comp}.html'; pf.write_text(html, encoding='utf-8')
            for w in (1440, 390):
                pg = await b.new_page(viewport={'width': w, 'height': 800})
                await pg.goto(f'file://{pf}'); await pg.wait_for_timeout(800)
                r = await pg.evaluate("""() => {
                  const de = document.documentElement;
                  const shown = [...document.querySelectorAll('svg.ip-chart')].filter(s => s.getBoundingClientRect().width > 0);
                  const ax = shown.map(s => { const t = s.querySelector('.ax'); return t ? +t.getBoundingClientRect().height.toFixed(1) : null; });
                  const fs = shown.map(s => { const t = s.querySelector('.ax'); if (!t) return null;
                     const m = s.getScreenCTM(); return +(parseFloat(getComputedStyle(t).fontSize) * m.a).toFixed(1); });
                  return {hscroll: de.scrollWidth - de.clientWidth, shown: shown.map(s => s.classList.contains('ip-chart--phone') ? 'phone' : 'wide'), axpx: fs};
                }""")
                print(comp, w, r)
                await pg.screenshot(path=str(OUT/f'{comp}-{w}-{tag}.png'), full_page=True)
                await pg.close()
        await b.close()
asyncio.run(main())
