"""python3 s14_render.py <out_heights.json> [--w 1280] <stem>... : render /tmp/s4s/<stem>.html (plain wrapper written by
s14_build.py, linking the repo's tokens/bundle/additions) at 1440 (or --w), shot design/shots/S14[-n][-1280].png; at 1440
record the board height (main content + 24, at least 1024). Prints table/region widths, clipped elements, duplicate ids
and broken aria references."""
import sys, json, asyncio, pathlib
from playwright.async_api import async_playwright
R = pathlib.Path(__file__).resolve().parents[3]
ORDER = ['S14-rapoarte-productie', 'S14-rapoarte-productie--se-incarca', 'S14-rapoarte-productie--fara-date-in-perioada',
         'S14-rapoarte-productie--fara-ghid', 'S14-rapoarte-productie--seria-2025',
         # S14S (s14s_build.py; heights in s14s_heights.json): shots S14-5 … S14-8; the phone renders at 390 from /tmp/s2
         'S14-rapoarte-productie--observator-se-exporta', 'S14-rapoarte-productie--eroare-la-incarcare',
         'S14-rapoarte-productie--fara-ferma-de-ouat', 'S14-rapoarte-productie--telefon',
         'S14-rapoarte-productie--telefon-filtre']  # S14P: phone filters open (fixed 390×844, shot S14-9.png)
PHONE = {'S14-rapoarte-productie--telefon', 'S14-rapoarte-productie--telefon-filtre'}
FIXED = {'S14-rapoarte-productie--telefon-filtre'}
async def main():
    args = sys.argv[1:]; hp = pathlib.Path(args.pop(0)); W = 1440
    if args and args[0] == '--w': args.pop(0); W = int(args.pop(0))
    hs = json.loads(hp.read_text()) if hp.exists() else {}
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for st in args:
            ph = st in PHONE; w = 390 if ph else W
            pg = await b.new_page(viewport={'width': w, 'height': hs.get(st, 844) if ph else 1000})
            await pg.goto(f'file:///tmp/{"s2" if ph else "s4s"}/{st}.html'); await pg.wait_for_timeout(1200)
            if ph:  # phone: main content (with its padding) + the tab bar at the foot
                mh = await pg.evaluate("(()=>{const m=document.querySelector('.ip-screen__main');const l=m.lastElementChild.getBoundingClientRect().bottom;return Math.ceil(l+parseFloat(getComputedStyle(m).paddingBottom)+document.querySelector('.ip-tabbar').offsetHeight)})()")
            else:
                mh = await pg.evaluate("Math.ceil(document.querySelector('.ip-shell__main').lastElementChild.getBoundingClientRect().bottom + 24)")
            clip = await pg.evaluate("""[...document.querySelectorAll('.ip-readout *,.ip-stat *,.ip-filterbar *,.ip-pagehead *,.ip-table td,.ip-table th')].filter(e=>e.scrollWidth>e.clientWidth+1 && getComputedStyle(e).overflow!='visible').map(e=>e.className+':'+e.textContent.slice(0,40)).slice(0,8)""")
            tw = await pg.evaluate("[...document.querySelectorAll('.ip-tablescroll')].map(r=>[Math.round(r.clientWidth), r.scrollWidth]).concat([[Math.round(document.querySelector('.ip-shell__main,.ip-screen__main').getBoundingClientRect().width), document.documentElement.scrollWidth]])")
            ids = await pg.evaluate("(()=>{const c={};document.querySelectorAll('[id]').forEach(e=>c[e.id]=(c[e.id]||0)+1);const d=Object.keys(c).filter(k=>c[k]>1);const refs=[];document.querySelectorAll('[aria-describedby],[aria-labelledby],[aria-controls]').forEach(e=>['aria-describedby','aria-labelledby','aria-controls'].forEach(a=>(e.getAttribute(a)||'').split(/\\s+/).filter(Boolean).forEach(i=>{if(!document.getElementById(i))refs.push(a+':'+i)})));return {dup:d,broken:refs}})()")
            print(st, w, 'main h', mh, 'regions [client, scroll]', tw, 'clip', clip, ids)
            i = ORDER.index(st); name = 'S14' + (f'-{i}' if i else '') + ('' if W == 1440 or ph else f'-{W}') + '.png'
            if ph: hs[st] = 844 if st in FIXED else max(mh, 844)
            elif W == 1440: hs[st] = max(mh, 1024)
            await pg.screenshot(path=str(R / 'design/shots' / name), full_page=True)
            await pg.close()
        await b.close()
    if W == 1440: hp.write_text(json.dumps(hs, indent=1))
asyncio.run(main())
