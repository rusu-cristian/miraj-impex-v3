"""python3 s8_render.py <heights.json> [stem ...] : render the S8 phone boards (390 wide, from /tmp/s8, written by s8_build.py)
to design/shots/S8[-n].png at the board height (natural content height, at least 844), record that height in heights.json,
and report sideways overflow, duplicate ids, broken ARIA/label refs and unnamed fields. Rebuild with the heights after."""
import asyncio, json, pathlib, sys
from playwright.async_api import async_playwright
R = pathlib.Path(__file__).resolve().parents[3]
ORDER = ['S8-cantarire', 'S8-cantarire--gol', 'S8-cantarire--in-norma', 'S8-cantarire--abatere',
         'S8-cantarire--valoare-refuzata', 'S8-cantarire--salvat', 'S8-cantarire--eroare-la-salvare',
         # S8S (s8s_build.py): shots S8-7 … S8-12; the sheet board is one 844 screen (s7s-vp)
         'S8-cantarire--se-incarca', 'S8-cantarire--alege-sectiunea', 'S8-cantarire--zi-trecuta',
         'S8-cantarire--ferma-fara-pasari', 'S8-cantarire--deja-introdus', 'S8-cantarire--gol-telita']
CHECK = """() => {
 const ids = [...document.querySelectorAll('[id]')].map(e => e.id); const dup = ids.filter((x, i) => ids.indexOf(x) !== i);
 const refs = [...document.querySelectorAll('[aria-describedby],[aria-labelledby],[for],[aria-controls]')].flatMap(e =>
   ['aria-describedby','aria-labelledby','for','aria-controls'].flatMap(a => (e.getAttribute(a) || '').split(/\\s+/).filter(Boolean)))
   .filter(r => !document.getElementById(r));
 const unnamed = [...document.querySelectorAll('input,select,textarea,[role=combobox]')].filter(e => !(e.labels && e.labels.length) && !e.getAttribute('aria-labelledby') && !e.getAttribute('aria-label')).map(e => e.id);
 const over = [...document.querySelectorAll('.s7-screen *')].filter(e => e.getBoundingClientRect().right > 390.5).map(e => e.className).slice(0, 5);
 const r = document.querySelector('.ip-root'); const mh = r.style.minHeight; r.style.minHeight = '0px';
 const natural = Math.ceil(document.querySelector('.s7-screen').getBoundingClientRect().height); r.style.minHeight = mh;
 return {dup, broken: refs, unnamed, over, sw: document.documentElement.scrollWidth, natural};
}"""
async def main():
    hp = pathlib.Path(sys.argv[1]); stems = sys.argv[2:] or ORDER
    hs = json.loads(hp.read_text()) if hp.exists() else {}
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for st in stems:
            pg = await b.new_page(viewport={'width': 390, 'height': 844})
            await pg.goto(f'file:///tmp/s8/{st}.html'); await pg.wait_for_timeout(1000)
            res = await pg.evaluate(CHECK); h = max(844, res['natural']); hs[st] = h
            await pg.set_viewport_size({'width': 390, 'height': h}); await pg.wait_for_timeout(200)
            i = ORDER.index(st)
            await pg.screenshot(path=str(R / 'design/shots' / ('S8' + (f'-{i}' if i else '') + '.png')), full_page=True)
            print(st, h, res); await pg.close()
        await b.close()
    hp.write_text(json.dumps(hs, indent=1) + '\n')
asyncio.run(main())
