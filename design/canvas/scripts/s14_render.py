"""python3 s14_render.py <out_heights.json> [--w 1280] [--widths] <stem>... : render /tmp/s4s/<stem>.html (desktops) or
/tmp/s2/<stem>.html (phones), plain wrappers written by s14_build.py / s14s_build.py, at their canvas width (board_widths.json,
else 1440 / 390; or --w for desktops), shot design/shots/S14[-n][-1280].png; at 1440 record the board height (main content + 24,
at least 1024; phones: content + tab bar, at least 844). Prints table/region widths, the width the board needs for its widest
table (D167: 1440 / 390 + overflow, up to the next 8 px), clipped elements, duplicate ids and broken aria references.
--widths (S14N): measure as the 1440 / 390 page (canvas-only wide CSS dropped) and write the needs into board_widths.json."""
import sys, json, asyncio, pathlib
from playwright.async_api import async_playwright
HERE = pathlib.Path(__file__).resolve().parent; R = HERE.parents[2]
sys.path.insert(0, str(HERE)); import canvas_rows as CR
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
    wid = '--widths' in args; args = [a for a in args if a != '--widths']
    hs = json.loads(hp.read_text()) if hp.exists() else {}
    bw = dict(CR.WIDTHS)
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for st in args:
            ph = st in PHONE; base = 390 if ph else 1440
            w = base if wid else CR.width(st, base) if (ph or W == 1440) else W
            pg = await b.new_page(viewport={'width': w, 'height': hs.get(st, 844) if ph else 1000})
            await pg.goto(f'file:///tmp/{"s2" if ph else "s4s"}/{st}.html'); await pg.wait_for_timeout(1200)
            if wid: await pg.evaluate("""document.querySelectorAll('.s4s-wide,.s14-wide').forEach(e => {e.classList.remove('s4s-wide','s14-wide'); if (e.style.width) e.style.width='390px'})""")
            if ph:  # phone: main content (with its padding) + the tab bar at the foot
                mh = await pg.evaluate("(()=>{const m=document.querySelector('.ip-screen__main');const l=m.lastElementChild.getBoundingClientRect().bottom;return Math.ceil(l+parseFloat(getComputedStyle(m).paddingBottom)+document.querySelector('.ip-tabbar').offsetHeight)})()")
            else:
                mh = await pg.evaluate("Math.ceil(document.querySelector('.ip-shell__main').lastElementChild.getBoundingClientRect().bottom + 24)")
            clip = await pg.evaluate("""[...document.querySelectorAll('.ip-readout *,.ip-stat *,.ip-filterbar *,.ip-pagehead *,.ip-table td,.ip-table th')].filter(e=>e.scrollWidth>e.clientWidth+1 && getComputedStyle(e).overflow!='visible').map(e=>e.className+':'+e.textContent.slice(0,40)).slice(0,8)""")
            tw = await pg.evaluate("[...document.querySelectorAll('.ip-tablescroll')].map(r=>[Math.round(r.clientWidth), r.scrollWidth]).concat([[Math.round(document.querySelector('.ip-shell__main,.ip-screen__main').getBoundingClientRect().width), document.documentElement.scrollWidth]])")
            over = await pg.evaluate("Math.max(0, ...[...document.querySelectorAll('.ip-tablescroll')].map(r => r.scrollWidth - r.clientWidth))")
            need = base + (-(-over // 8) * 8 if over else 0)
            ids = await pg.evaluate("(()=>{const c={};document.querySelectorAll('[id]').forEach(e=>c[e.id]=(c[e.id]||0)+1);const d=Object.keys(c).filter(k=>c[k]>1);const refs=[];document.querySelectorAll('[aria-describedby],[aria-labelledby],[aria-controls]').forEach(e=>['aria-describedby','aria-labelledby','aria-controls'].forEach(a=>(e.getAttribute(a)||'').split(/\\s+/).filter(Boolean).forEach(i=>{if(!document.getElementById(i))refs.push(a+':'+i)})));return {dup:d,broken:refs}})()")
            rows = await pg.evaluate("[...new Set([...document.querySelectorAll('.ip-ptable tbody tr:not([hidden]),.ip-ptable tfoot tr')].map(r=>Math.round(r.getBoundingClientRect().height*2)/2))]")
            print(st, w, 'main h', mh, 'regions [client, scroll]', tw, 'needs w', need, 'row h', rows, 'clip', clip, ids)
            if wid:
                if need > base: bw[st] = need
                else: bw.pop(st, None)
                await pg.close(); continue
            i = ORDER.index(st); name = 'S14' + (f'-{i}' if i else '') + ('' if W == 1440 or ph else f'-{W}') + '.png'
            if ph: hs[st] = 844 if st in FIXED else max(mh, 844)
            elif W == 1440: hs[st] = max(mh, 1024)
            await pg.screenshot(path=str(R / 'design/shots' / name), full_page=True)
            await pg.close()
        await b.close()
    if wid:
        (HERE / 'board_widths.json').write_text(json.dumps(dict(sorted(bw.items())), indent=1, ensure_ascii=False) + '\n'); print('board_widths.json', bw)
    elif W == 1440: hp.write_text(json.dumps(hs, indent=1))
asyncio.run(main())
