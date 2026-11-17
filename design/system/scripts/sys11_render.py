"""Render the SYS11 preview (GuideImportPreview) at 1440: python3 sys11_render.py <tag>"""
import sys, asyncio, pathlib
from playwright.async_api import async_playwright
ROOT = pathlib.Path('/home/claude/miraj-impex-v3/design/system')
OUT = pathlib.Path('/tmp/sys11'); OUT.mkdir(exist_ok=True)
async def main():
    tag = sys.argv[1] if len(sys.argv) > 1 else 'r'
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        comp = "GuideImportPreview"
        frag = (ROOT/'components'/comp/'preview.html').read_text(encoding='utf-8')
        html = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8">
<link rel="stylesheet" href="file://{ROOT}/tokens.css">
<link rel="stylesheet" href="file://{ROOT}/components/bundle.css">
<link rel="stylesheet" href="file://{ROOT}/components/additions.css">
<style>body{{margin:0;padding:24px;background:var(--ground)}}</style></head><body>{frag}</body></html>'''
        pf = OUT/f'{comp}.html'; pf.write_text(html, encoding='utf-8')
        pg = await b.new_page(viewport={'width': 1440, 'height': 900})
        await pg.goto(f'file://{pf}'); await pg.evaluate('document.fonts.ready.then(() => true)'); await pg.wait_for_timeout(800)
        ov = await pg.evaluate("""() => [...document.querySelectorAll('.ip-gimp,.ip-gimp__src,.ip-gimp__bar,.ip-gimp__foot,.ip-banner')].filter(e => e.scrollWidth > e.clientWidth+1 || e.getBoundingClientRect().width > e.parentElement.getBoundingClientRect().width+1).map(e=>e.className)""")
        hs = await pg.evaluate("""() => [...document.querySelectorAll('.ip-gimp__t tbody tr')].map(e=>Math.round(e.getBoundingClientRect().height))""")
        ids = await pg.evaluate("""() => { const bad=[]; document.querySelectorAll('[aria-controls],[aria-describedby],[aria-labelledby]').forEach(e=>{['aria-controls','aria-describedby','aria-labelledby'].forEach(a=>{const v=e.getAttribute(a); if(v) v.split(' ').forEach(i=>{ if(!document.getElementById(i)) bad.push(a+':'+i)})})}); const all=[...document.querySelectorAll('[id]')].map(e=>e.id); const dup=all.filter((x,i)=>all.indexOf(x)!==i); return {bad, dup}; }""")
        print('overflow:', ov[:6], 'row heights:', sorted(set(hs)), 'refs:', ids)
        await pg.screenshot(path=str(OUT/f'{comp}-{tag}.png'), full_page=True)
        await pg.close()
        await b.close()
asyncio.run(main())
