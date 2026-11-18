"""Render the SYS12 previews (UserRow, OneTimePassword) at 1440: python3 sys12_render.py <tag>"""
import sys, asyncio, pathlib
from playwright.async_api import async_playwright
ROOT = pathlib.Path(__file__).resolve().parents[1]  # design/system from this script's place (design/system/scripts), as sys11_render
OUT = pathlib.Path('/tmp/sys12'); OUT.mkdir(exist_ok=True)
OV = """() => [...document.querySelectorAll('.ip-users,.ip-users__box,.ip-list__head,.ip-dialog,.ip-otp,.ip-otp__row,.ip-banner')]
  .filter(e => e.scrollWidth > e.clientWidth+1 || e.getBoundingClientRect().width > e.parentElement.getBoundingClientRect().width+1)
  .map(e=>e.className)"""
HS = """() => [...document.querySelectorAll('.ip-users__t tbody tr')].map(e=>Math.round(e.getBoundingClientRect().height))"""
IDS = """() => { const bad=[]; document.querySelectorAll('[aria-controls],[aria-describedby],[aria-labelledby]').forEach(e=>{
  ['aria-controls','aria-describedby','aria-labelledby'].forEach(a=>{const v=e.getAttribute(a); if(v) v.split(' ').forEach(i=>{
  if(!document.getElementById(i)) bad.push(a+':'+i)})})}); const all=[...document.querySelectorAll('[id]')].map(e=>e.id);
  document.querySelectorAll('label[for]').forEach(l=>{const t=document.getElementById(l.htmlFor); if(!t||!t.labels.length) bad.push('for:'+l.htmlFor)});
  return {bad, dup: all.filter((x,i)=>all.indexOf(x)!==i)}; }"""
SEL = """() => { const L=c=>{const v=c.match(/[\\d.]+/g).slice(0,3).map(x=>{x/=255;return x<=.03928?x/12.92:((x+.055)/1.055)**2.4});
  return .2126*v[0]+.7152*v[1]+.0722*v[2]}; return [...document.querySelectorAll('.ip-otp__pw')].map(e=>{const s=getComputedStyle(e);
  const a=L(s.color),b=L(s.backgroundColor); return e.id+':'+((Math.max(a,b)+.05)/(Math.min(a,b)+.05)).toFixed(2)+(e.scrollWidth>e.clientWidth?' OVERFLOW':'')}) }"""

async def main():
    tag = sys.argv[1] if len(sys.argv) > 1 else 'r'
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for comp in ('UserRow', 'OneTimePassword'):
            frag = (ROOT / 'components' / comp / 'preview.html').read_text(encoding='utf-8')
            html = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8">
<link rel="stylesheet" href="file://{ROOT}/tokens.css">
<link rel="stylesheet" href="file://{ROOT}/components/bundle.css">
<link rel="stylesheet" href="file://{ROOT}/components/additions.css">
<style>body{{margin:0;padding:24px;background:var(--ground)}}</style></head><body>{frag}</body></html>'''
            pf = OUT / f'{comp}.html'; pf.write_text(html, encoding='utf-8')
            pg = await b.new_page(viewport={'width': 1440, 'height': 900})
            await pg.goto(f'file://{pf}'); await pg.wait_for_timeout(800)
            print(comp, 'overflow:', (await pg.evaluate(OV))[:6], 'row heights:', sorted(set(await pg.evaluate(HS))),
                  'refs:', await pg.evaluate(IDS), 'height:', await pg.evaluate('document.body.scrollHeight'), 'pw:', await pg.evaluate(SEL))
            await pg.screenshot(path=str(OUT / f'{comp}-{tag}.png'), full_page=True)
            await pg.close()
        await b.close()
asyncio.run(main())
