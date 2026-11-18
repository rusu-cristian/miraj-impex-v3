"""Render the SYS13 previews (LoginCard, Wordmark) at 1440: python3 sys13_render.py <tag>"""
import sys, asyncio, pathlib
from playwright.async_api import async_playwright
ROOT = pathlib.Path(__file__).resolve().parents[1]  # design/system from this script's place (design/system/scripts), as sys11_render
OUT = pathlib.Path('/tmp/sys13'); OUT.mkdir(exist_ok=True)
OV = """() => [...document.querySelectorAll('.ip-login,.ip-login__form,.ip-field,.ip-input,.ip-banner,.ip-login-page,.ip-phone,.ip-sidebar,.ip-wordmark')]
  .filter(e => e.scrollWidth > e.clientWidth+1 || e.getBoundingClientRect().width > e.parentElement.getBoundingClientRect().width+1)
  .map(e=>e.className)"""
HS = """() => [...document.querySelectorAll('.ip-input,.ip-login__go,.ip-login__eye')].map(e=>e.className.split(' ').slice(-2).join('.')+':'+Math.round(e.getBoundingClientRect().width)+'x'+Math.round(e.getBoundingClientRect().height))"""
IDS = """() => { const bad=[]; document.querySelectorAll('[aria-controls],[aria-describedby],[aria-labelledby]').forEach(e=>{
  ['aria-controls','aria-describedby','aria-labelledby'].forEach(a=>{const v=e.getAttribute(a); if(v) v.split(' ').forEach(i=>{
  if(!document.getElementById(i)) bad.push(a+':'+i)})})}); const all=[...document.querySelectorAll('[id]')].map(e=>e.id);
  document.querySelectorAll('label[for]').forEach(l=>{const t=document.getElementById(l.htmlFor); if(!t||!t.labels.length) bad.push('for:'+l.htmlFor)});
  return {bad, dup: all.filter((x,i)=>all.indexOf(x)!==i)}; }"""
AC = """() => [...document.querySelectorAll('input')].map(e=>e.id+'='+e.autocomplete+'/'+e.type).filter((x,i,a)=>a.indexOf(x)===i).length + ' inputs ok; bad: ' + [...document.querySelectorAll('input')].filter(e=>!['username','current-password'].includes(e.getAttribute('autocomplete'))).map(e=>e.id).join(',')"""
ERR = """() => [...document.querySelectorAll('.ip-login .ip-err')].map(e=>{const r=e.getBoundingClientRect(),
  i=e.previousElementSibling.getBoundingClientRect(),n=e.closest('.ip-field').nextElementSibling.getBoundingClientRect();
  return e.id+':'+Math.round(r.top-i.bottom)+'/'+Math.round(n.top-r.bottom)})"""

async def main():
    tag = sys.argv[1] if len(sys.argv) > 1 else 'r'
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for comp in ('LoginCard', 'Wordmark'):
            frag = (ROOT / 'components' / comp / 'preview.html').read_text(encoding='utf-8')
            html = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8">
<link rel="stylesheet" href="file://{ROOT}/tokens.css">
<link rel="stylesheet" href="file://{ROOT}/components/bundle.css">
<link rel="stylesheet" href="file://{ROOT}/components/additions.css">
<style>body{{margin:0;padding:24px;background:var(--ground)}}</style></head><body>{frag}</body></html>'''
            pf = OUT / f'{comp}.html'; pf.write_text(html, encoding='utf-8')
            pg = await b.new_page(viewport={'width': 1440, 'height': 900})
            await pg.goto(f'file://{pf}'); await pg.wait_for_timeout(800)
            print(comp, 'overflow:', (await pg.evaluate(OV))[:6], 'sizes:', sorted(set(await pg.evaluate(HS)))[:14],
                  'refs:', await pg.evaluate(IDS), 'height:', await pg.evaluate('document.body.scrollHeight'), 'autocomplete:', await pg.evaluate(AC), 'err gap above/below:', await pg.evaluate(ERR))
            await pg.screenshot(path=str(OUT / f'{comp}-{tag}.png'), full_page=True)
            await pg.close()
        await b.close()
asyncio.run(main())
