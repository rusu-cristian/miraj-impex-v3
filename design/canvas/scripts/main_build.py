"""main_build.py · the 00 · Index board (Main.dc.html, FV03 / FINAL4 10): lists the canvas rows top to bottom (row, screen,
board count) from canvas.json, with the round-2 summary. Run last, after `canvas_rows.py relayout`:
python3 -I main_build.py <repo_root> [--canvas]   (--canvas: refresh Main's w/h in canvas.json in place; x/y stay 0/0)
Also writes /tmp/main/Main.html (plain wrapper) for a render. Uses only system tokens and components (ip-table, ip-cap, ip-t-*)."""
import pathlib, sys
sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR

ROOT, _, CANVAS = CR.args()
W, H = 880, 640   # H < ROW0 − GAP_Y (680): the first row stays where it is
NAMES = {  # screen → its name on the index (the default board's title without the state)
    'S1': 'Autentificare', 'S2': 'Acasă', 'S3': 'Seria 2026', 'S4': 'Fermă · Rezumat, Păsări, Hrană, Producție',
    'S6': 'Zilnic', 'S7': 'Hrană zilnică', 'S8': 'Cântărire', 'S9': 'Transfer', 'S10': 'Vânzare',
    'S11': 'Rapoarte · Hrană', 'S12': 'Prețuri hrană', 'S13': 'Livrare hrană', 'S14': 'Rapoarte · Producție',
    'S15': 'Introdus azi', 'S16': 'Corectează', 'S17': 'Istoric modificări', 'S18': 'Serii', 'S19': 'Ferme și secțiuni',
    'S20': 'Ghiduri', 'S21': 'Tipuri de hrană', 'S22': 'Utilizatori', 'M1': 'Mișcare (sistem, interactiv)', 'V1': 'Voce (sistem: tastă, panou, marcaj)',
    'VO': 'Comenzi vocale (VO02: peste ecrane reale)'}
SUMMARY = ('Runda 2 (19.11.2026): date reale până la 18.11.2026 · săptămâni după vârsta păsărilor (azi SAPT 59) · '
           'tabelul întâi: fiecare raport pornește de la tabelul pe perioade.')


def de(n):
    """Romanian: 20 or more (by the last two digits) takes "de" before the noun"""
    return ' de' if n % 100 >= 20 or (n >= 100 and n % 100 == 0) else ''


def n(x):
    return f'{x:,}'.replace(',', ' ')


c = CR.load(ROOT)
rows = []
for y, keys in CR.rows(c):
    if y < CR.ROW0: continue
    first = min(keys, key=lambda k: c['boards'][k]['x'])
    scr = first.split('-')[0]
    rows.append((scr, NAMES[scr], len(keys)))
total = sum(r[2] for r in rows)
half = (len(rows) + 1) // 2


def table(part, start):
    body = ''.join(f'<tr><td class="num">{start + i}</td><th scope="row">{s} · {nm}</th><td class="num">{k}</td></tr>'
                   for i, (s, nm, k) in enumerate(part))
    return ('<table class="ip-table ip-table--rowheads" style="width:auto;flex:1 1 0;min-width:0">'
            '<thead><tr><th scope="col" class="num">Rând</th><th scope="col">Ecran</th><th scope="col" class="num">Planșe</th></tr></thead>'
            f'<tbody>{body}</tbody></table>')


root = f'''<div class="ip-root" style="width:{W}px;height:{H}px;box-sizing:border-box;padding:var(--space-8);display:flex;flex-direction:column;gap:var(--space-4);background:var(--ground);color:var(--ink);font-family:var(--font-sans)">
<div style="display:flex;flex-direction:column;gap:var(--space-1)">
<span class="ip-cap">IMPEX LINE · ECRANE</span>
<h1 class="ip-t-title" style="margin:0">Index</h1>
<p class="ip-t-body" style="margin:0;color:var(--ink-secondary)">{SUMMARY}</p>
</div>
<div style="display:flex;flex-direction:column;gap:var(--space-2)">
<p class="ip-t-small" style="margin:0;color:var(--ink-secondary)"><span class="ip-num">{len(rows)}</span>{de(len(rows))} rânduri · <span class="ip-num">{n(total)}</span>{de(total)} planșe, de sus în jos în ordinea de pe canvas; în fiecare rând starea implicită e prima, stările și telefonul la dreapta ei.</p>
<div style="display:flex;gap:var(--space-4);align-items:flex-start">
{table(rows[:half], 1)}
{table(rows[half:], half + 1)}
</div>
</div>
</div>'''
FONTS = ('<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500&amp;'
         'family=IBM+Plex+Mono:wght@400;500&amp;display=swap">')
dc = f'''<!doctype html>
<html lang="ro">
<head>
<meta charset="utf-8">
<title>00 · Index</title>
<script src="./support.js"></script>
<link rel="stylesheet" href="ds/impexlinepanel/tokens.css">
<link rel="stylesheet" href="ds/impexlinepanel/components/bundle.css">
<link rel="stylesheet" href="additions.css">
</head>
<body>
<x-dc>
<helmet>
{FONTS}
<style>
body{{margin:0}}
</style>
</helmet>
{root}
</x-dc>
<script type="text/x-dc" data-dc-script data-props='{{"$preview":{{"width":{W},"height":{H}}}}}'>
class Component extends DCLogic {{
renderVals() {{
return {{}};
}}
}}
</script>
</body>
</html>
'''
out = ROOT / 'design/canvas/project/Main.dc.html'
if not out.exists() or out.read_text(encoding='utf-8') != dc: out.write_text(dc, encoding='utf-8')
SYS = ROOT / 'design/system'
pathlib.Path('/tmp/main').mkdir(exist_ok=True)
pathlib.Path('/tmp/main/Main.html').write_text(
    f'''<!doctype html><html lang="ro"><head><meta charset="utf-8"><title>00 · Index</title>
<link rel="stylesheet" href="file://{SYS}/tokens.css"><link rel="stylesheet" href="file://{SYS}/components/bundle.css"><link rel="stylesheet" href="file://{SYS}/components/additions.css">
{FONTS}<style>body{{margin:0}}</style></head><body>{root}</body></html>''', encoding='utf-8')
if CANVAS:
    c['boards']['Main.dc.html'].update(w=W, h=H)
    CR.save(ROOT, c)
print('ok Main', len(rows), 'rows', total, 'boards', W, H)
