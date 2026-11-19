"""Render the SYS14 preview (PeriodTable) at 1440 and check it: python3 sys14_render.py <tag>
Checks: page overflow, id references, duplicate ids, body row heights per table, loading vs loaded column widths,
keyboard (Tab to an expand key, ring drawn, Enter/Space toggle through a test handler), table roles in the AX tree,
text contrast of the week-row dates, head notes and ♀/♂ figures."""
import sys, asyncio, pathlib
from playwright.async_api import async_playwright
ROOT = pathlib.Path(__file__).resolve().parents[1]  # design/system
OUT = pathlib.Path('/tmp/sys14'); OUT.mkdir(exist_ok=True)

IDS = """() => { const bad=[]; document.querySelectorAll('[aria-controls],[aria-labelledby]').forEach(e=>{
  ['aria-controls','aria-labelledby'].forEach(a=>{const v=e.getAttribute(a); if(v) v.split(' ').forEach(i=>{
  if(!document.getElementById(i)) bad.push(a+':'+i)})})}); const all=[...document.querySelectorAll('[id]')].map(e=>e.id);
  return {bad, dup: all.filter((x,i)=>all.indexOf(x)!==i)}; }"""
ROWS = """() => [...document.querySelectorAll('table.ip-ptable')].map(t => {
  const hs=[...t.querySelectorAll('tbody > tr:not([hidden]), tfoot > tr')].map(r=>Math.round(r.getBoundingClientRect().height*10)/10);
  const heads=[...t.querySelectorAll('thead tr')].map(r=>Math.round(r.getBoundingClientRect().height));
  const reg=t.parentElement.getBoundingClientRect();
  return t.id+': rows '+[...new Set(hs)].join('/')+' n='+hs.length+' head '+heads.join('/')+' table '+Math.round(t.getBoundingClientRect().width)+' region '+Math.round(reg.width)+'x'+Math.round(reg.height); })"""
COLS = """() => ['pa-t','pg-t'].map(id => [...document.getElementById(id).querySelectorAll('tbody:first-of-type > tr:first-child > *')].map(c=>Math.round(c.getBoundingClientRect().width*10)/10).join(' '))"""
OVER = """() => ({page: document.documentElement.scrollWidth, phone: [...document.querySelectorAll('.ip-phone')].map(p=>p.scrollWidth+'/'+p.clientWidth)})"""
CON = """() => { function L(c){const m=c.match(/[\\d.]+/g).map(Number).slice(0,3).map(v=>{v/=255;return v<=0.03928?v/12.92:Math.pow((v+0.055)/1.055,2.4)});return 0.2126*m[0]+0.7152*m[1]+0.0722*m[2];}
  function bg(e){while(e){const b=getComputedStyle(e).backgroundColor; if(b && !b.startsWith('rgba(0, 0, 0, 0)') && b!=='transparent') return b; e=e.parentElement;} return 'rgb(255,255,255)';}
  function r(e){const a=L(getComputedStyle(e).color),b=L(bg(e));return Math.round(((Math.max(a,b)+0.05)/(Math.min(a,b)+0.05))*100)/100;}
  const sel={dates:'.ip-ptable__dates', note:'.ip-ptable__note', hen:'.ip-ptable tbody td.is-hen', rooster:'.ip-ptable tbody td.is-rooster',
    headhen:'.ip-ptable thead .is-hen', head:'.ip-ptable thead th', chev:'.ip-ptable__key .ip-icon', wkhen:'.ip-ptable tr.ip-ptable__wk td.is-hen'};
  const o={}; for(const k in sel){const es=[...document.querySelectorAll(sel[k])]; if(es.length) o[k]=Math.min(...es.map(r));} return o; }"""

async def main():
    tag = sys.argv[1] if len(sys.argv) > 1 else 'r'
    frag = (ROOT / 'components/PeriodTable/preview.html').read_text(encoding='utf-8')
    html = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8">
<link rel="stylesheet" href="file://{ROOT}/tokens.css">
<link rel="stylesheet" href="file://{ROOT}/components/bundle.css">
<link rel="stylesheet" href="file://{ROOT}/components/additions.css">
<style>body{{margin:0;padding:24px;background:var(--ground)}}</style></head><body>{frag}
<script>/* test handler only (not in the preview): what the product does on the key */
document.addEventListener('click', e => {{ const k = e.target.closest('.ip-ptable__key'); if (!k) return;
  const open = k.getAttribute('aria-expanded') !== 'true'; k.setAttribute('aria-expanded', open);
  k.getAttribute('aria-controls').split(' ').forEach(i => document.getElementById(i).hidden = !open); }});</script></body></html>'''
    pf = OUT / 'PeriodTable.html'; pf.write_text(html, encoding='utf-8')
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        pg = await b.new_page(viewport={'width': 1440, 'height': 900})
        await pg.goto(f'file://{pf}'); await pg.wait_for_timeout(1200)
        print('overflow', await pg.evaluate(OVER))
        print('refs', await pg.evaluate(IDS))
        for r in await pg.evaluate(ROWS): print(' ', r)
        print('cols loaded/loading', *await pg.evaluate(COLS), sep='\n  ')
        print('contrast', await pg.evaluate(CON))
        await pg.screenshot(path=str(OUT / f'PeriodTable-{tag}.png'), full_page=True)
        # keyboard: focus the region of table 1, Tab to the first key, check the ring, toggle with Enter and Space
        await pg.focus('[aria-label="Păsări · Hîrbovăț · Seria 2026 · SAPT 56–59"]')
        await pg.keyboard.press('Tab')
        info = await pg.evaluate("""() => {const e=document.activeElement, s=getComputedStyle(e); return e.className+' | '+e.textContent.trim()+' | outline '+s.outlineStyle+' '+s.outlineWidth+' '+s.outlineColor+' off '+s.outlineOffset+' | expanded '+e.getAttribute('aria-expanded')}""")
        print('tab ->', info)
        await pg.keyboard.press('Enter'); a1 = await pg.evaluate("() => document.activeElement.getAttribute('aria-expanded')+' hidden:'+document.getElementById('pa-1117').hidden")
        await pg.keyboard.press(' '); a2 = await pg.evaluate("() => document.activeElement.getAttribute('aria-expanded')+' hidden:'+document.getElementById('pa-1117').hidden")
        await pg.keyboard.press('Tab'); a3 = await pg.evaluate("() => document.activeElement.textContent.trim().slice(0,40)")
        print('enter ->', a1, '| space ->', a2, '| next tab ->', a3)
        await pg.focus('#pa-t button[aria-controls^="pa-1110"]')
        clip = await pg.evaluate("() => {const r=document.querySelector('#pa-t').closest('.ip-tablescroll').getBoundingClientRect(); return {x:r.x, y:r.y, width:r.width, height:r.height}}")
        await pg.keyboard.press('Shift+Tab'); await pg.keyboard.press('Tab')
        await pg.screenshot(path=str(OUT / f'focus-{tag}.png'), clip=clip)
        # AX tree of table 1
        cdp = await pg.context.new_cdp_session(pg)
        ax = (await cdp.send('Accessibility.getFullAXTree'))['nodes']
        roles = {}
        for n in ax:
            r = n.get('role', {}).get('value'); roles[r] = roles.get(r, 0) + 1
        print('ax roles', {k: v for k, v in roles.items() if k in ('table', 'rowgroup', 'row', 'rowheader', 'columnheader', 'cell', 'region', 'radiogroup', 'radio', 'button', 'caption')})
        names = [n.get('name', {}).get('value') for n in ax if n.get('role', {}).get('value') == 'button'][:5]
        print('button names', names)
        rh = [n.get('name', {}).get('value') for n in ax if n.get('role', {}).get('value') == 'rowheader'][:6]
        print('rowheaders', rh)
        await b.close()
asyncio.run(main())
