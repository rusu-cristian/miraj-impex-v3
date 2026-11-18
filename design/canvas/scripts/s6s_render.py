"""python3 s6s_render.py [--measure] [--only <slug>] : render the S6S state boards (390×844, from /tmp/s6s, built by s6s_build.py) to
design/shots/S6S-<i>.png and report sideways overflow, duplicate ids, broken ARIA/label refs and the SaveBar height.
--measure: for boards with a target card (boards.json "target" = the card heading id), measure with the form unscrolled
how far it must scroll so the card ends 16 px above the SaveBar (never past its top), and write the offsets to
s6s_heights.json ("scroll"); rebuild with s6s_build.py afterwards (the offsets are baked into the boards)."""
import asyncio, json, pathlib, sys
from playwright.async_api import async_playwright
R = pathlib.Path(__file__).resolve().parents[3]
HP = R / 'design/canvas/scripts/s6s_heights.json'
MEASURE = '--measure' in sys.argv
ONLY = sys.argv[sys.argv.index('--only') + 1] if '--only' in sys.argv else None  # --only <slug>: render (and measure) one board
BOARDS = json.loads(pathlib.Path('/tmp/s6s/boards.json').read_text())
CHECK = """() => {
 const ids = [...document.querySelectorAll('[id]')].map(e => e.id); const dup = ids.filter((x, i) => ids.indexOf(x) !== i);
 const refs = [...document.querySelectorAll('[aria-describedby],[aria-labelledby],[for],[aria-controls]')].flatMap(e =>
   ['aria-describedby','aria-labelledby','for','aria-controls'].flatMap(a => (e.getAttribute(a) || '').split(/\\s+/).filter(Boolean)))
   .filter(r => !document.getElementById(r));
 const over = [...document.querySelectorAll('.ip-screen__main *')].filter(e => e.getBoundingClientRect().right > 390.5).map(e => e.className).slice(0, 5);
 const sb = document.querySelector('.ip-savebar');
 return {dup, broken: refs, over, barH: Math.round(sb.getBoundingClientRect().height)};
}"""
SCROLL = """(t) => {
 const main = document.querySelector('.ip-screen__main'); main.style.marginTop = '0px';
 const card = document.getElementById(t).closest('section'); const view = document.querySelector('.ip-screen__view').getBoundingClientRect();
 const r = card.getBoundingClientRect(); const pad = 16;
 return Math.max(0, Math.min(Math.ceil(r.bottom + pad - view.bottom), Math.floor(r.top - view.top - pad)));
}"""
async def main():
    hj = json.loads(HP.read_text()) if HP.exists() else {}
    sc = hj.get('scroll', {})
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        pg = await b.new_page(viewport={'width': 390, 'height': 844})
        for i, bd in enumerate(BOARDS, 1):
            if ONLY and bd['slug'] != ONLY: continue
            await pg.goto(f'file:///tmp/s6s/{bd["stem"]}.html'); await pg.wait_for_timeout(600)
            if MEASURE and bd['target']:
                sc[bd['slug']] = await pg.evaluate(SCROLL, bd['target'])
                await pg.evaluate(f"() => document.querySelector('.ip-screen__main').style.marginTop = '-{sc[bd['slug']]}px'")
            print(i, bd['slug'], sc.get(bd['slug'], 0), await pg.evaluate(CHECK))
            await pg.screenshot(path=str(R / f'design/shots/S6S-{i}.png'))
        await b.close()
    if MEASURE:
        hj = {bd['stem']: 844 for bd in BOARDS}; hj['scroll'] = sc
        HP.write_text(json.dumps(hj, ensure_ascii=False, indent=1) + '\n')
asyncio.run(main())
