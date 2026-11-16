"""Build the S7 · Hrană zilnică phone artboard (default state) from the extracts.
Hîrbovăț, 16.11.2026, SAPT 40, Dr. Ceban mid-entry: 10 of 12 rows confirmed (allocations_today), H3.2 ♂ and H4.2 ♀
still yesterday's suggestion, H1.2 ♂ 172 g (+11% vs guide 155 g). Composed from system classes only (FormHeader,
DateField, Button, AllocationRow (ip-alloc--phone, S7P), Select, TextField, Lamp, SexCount, HallCard hall head, FeedTotals
+ compact SaveBar); screen-only CSS (s7-*) is page layout with tokens. One tall phone board (canvas.md: taller h for a long
scrolling form). savebar() is the one SaveBar shell shared with the state boards (s7s_build.py).
python3 -I s7_build.py <repo_root> [height]
Writes design/canvas/project/S7-hrana-zilnica.dc.html and /tmp/s7/S7-hrana-zilnica.html (plain wrapper for renders)."""
import sys, json, pathlib, datetime as dt

ROOT = pathlib.Path(sys.argv[1])
H = int(sys.argv[2]) if len(sys.argv) > 2 else 2400
EX = ROOT / 'design/data/extracts'
S = json.load(open(EX / 'sample-hirbovat-2026.json', encoding='utf-8'))
GL = json.load(open(EX / 'guide-laying.json', encoding='utf-8'))
FT = json.load(open(EX / 'feed-types.json', encoding='utf-8'))
CY = json.load(open(EX / 'cycles.json', encoding='utf-8'))

NB = ' '; MINUS = '−'
TODAY = dt.date(2026, 11, 16); YDAY = TODAY - dt.timedelta(days=1)
START = dt.date.fromisoformat(S['meta']['cycle_start'])
WEEK = (TODAY - START).days // 7 + 1
assert WEEK == 40 == S['summary_today']['cycle_week']

def n(v, dec=0):
    s = f'{abs(v):,.{dec}f}'.replace(',', 'X').replace('.', ',').replace('X', NB)
    return (MINUS if v < 0 else '') + s
def dd(d): return d.strftime('%d.%m.%Y')

GUIDE = {r['sex']: r['feed_g'] for r in GL['rows'] if r['cycle_week'] == WEEK}
assert GUIDE == {'F': 158, 'M': 155}
FTN = {r['id']: r['name'] for r in FT['rows']}
FTS = {r['id']: r['for_sex'] for r in FT['rows']}
HB_TYPES = [r['id'] for r in FT['rows'] if r['farm_id'] == 'hirbovat' and r['active']]
ROWS = S['allocations_today']; assert len(ROWS) == 12

# ---------- icons (path data copied from the FormHeader / DateField / Select / FeedTotals previews; copy = Lucide "copy") ----------
IC = {
 'back': '<path d="m12 19-7-7 7-7"/><path d="M19 12H5"/>',
 'feed': '<path d="M12 21V9M12 9c-3 0-4-2-4-5 3 0 4 2 4 5zM12 9c3 0 4-2 4-5-3 0-4 2-4 5zM12 15c-3 0-4-2-4-5 3 0 4 2 4 5zM12 15c3 0 4-2 4-5-3 0-4 2-4 5z"/>',
 'cal': '<rect x="3" y="4" width="18" height="18" rx="2"/><path d="M16 2v4"/><path d="M8 2v4"/><path d="M3 10h18"/>',
 'chev': '<path d="m6 9 6 6 6-6"/>',
 'copy': '<rect width="14" height="14" x="8" y="8" rx="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/>',
 'warn': '<path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"/><path d="M12 9v4"/><path d="M12 17h.01"/>',
}
def icon(k, size=None):
    return f'<svg class="ip-icon{f" ip-icon--{size}" if size else ""}" viewBox="0 0 24 24" aria-hidden="true">{IC[k]}</svg>'
def sexv(sx, v, sign=True):
    c = 'hen' if sx == 'F' else 'rooster'
    return f'<span class="ip-sex ip-sex--{c}">{("♀ " if sx == "F" else "♂ ") if sign else ""}<b>{v}</b></span>'
def pair(f, m, sr=''):
    return f'<span class="ip-pair">{f'<span class="ip-sr">{sr}</span>' if sr else ''}{sexv("F", f)}<span class="sep">·</span>{sexv("M", m)}</span>'
def lamp(kind, word): return f'<span class="ip-lamp ip-lamp--{kind}">{word}</span>'

# ---------- header + date + copy key ----------
FARM = 'HÎRBOVĂȚ'
HEAD = f'''<header class="ip-formhead">
<a class="ip-iconbtn ip-iconbtn--touch" href="#acasa" aria-label="Înapoi la Acasă">{icon("back")}</a>
<span class="ip-tile ip-rec ip-rec--feed" aria-hidden="true">{icon("feed")}</span>
<div class="s7-title"><h1 class="ip-t-heading s7-m0">Hrană zilnică</h1><p class="ip-note ip-num s7-m0">{FARM} · {dd(TODAY)} · SAPT {WEEK}</p></div>
</header>'''

DATE = f'''<div class="ip-field">
<span class="ip-label" id="d-lbl">Data</span>
<div class="s7-date">
<span class="ip-seg ip-seg--touch" role="radiogroup" aria-labelledby="d-lbl"><span class="ip-seg__item is-selected" role="radio" aria-checked="true" tabindex="0">Azi</span><span class="ip-seg__item" role="radio" aria-checked="false" tabindex="-1">Ieri</span></span>
<span class="ip-input ip-input--touch ip-num">{icon("cal", 16)}<input id="d-val" type="text" inputmode="numeric" value="{dd(TODAY)}" aria-labelledby="d-lbl" aria-describedby="d-wk"/></span>
</div>
<span class="ip-note ip-num" id="d-wk">SAPT {WEEK} · ciclu din {dd(START)}</span>
</div>'''

COPY = f'''<div class="s7-copy">
<button type="button" class="ip-btn ip-btn--secondary ip-btn--touch">{icon("copy", 16)}Copiază ieri</button>
<p class="ip-note s7-m0">ghid <span class="ip-num">SAPT {WEEK}</span>: {pair(f"{GUIDE['F']}{NB}g", f"{GUIDE['M']}{NB}g")}</p>
</div>'''

# ---------- allocation rows ----------
def row(r):
    sx, sec, alive = r['sex'], r['section'], r['alive']
    c = 'hen' if sx == 'F' else 'rooster'
    word = 'găini' if sx == 'F' else 'cocoși'
    gid = f'{sec.replace(".", "-")}-{sx}'
    guide = r['guide_g']; assert guide == GUIDE[sx]
    confirmed = r['today_g'] is not None
    g = r['today_g'] if confirmed else r['yesterday_g']
    kg = round(g * alive / 1000, 1)
    if confirmed: assert abs(kg - r['kg']) < 0.051, (sec, sx, kg, r['kg'])
    dev = (g - guide) / guide * 100
    opts = ''.join(f'<option{" selected" if t == r["feed_type_id"] else ""}>{FTN[t]}</option>' for t in HB_TYPES if FTS[t] == sx)
    sel = f'<label class="ip-field"><span class="ip-label">Tip hrană<span class="ip-sr">, {sec} {word}</span></span><span class="ip-input ip-input--touch"><select>{opts}</select>{icon("chev", 16)}</span></label>'
    msg = ''
    if not confirmed:
        inp = f'<span class="ip-input ip-input--touch ip-num is-suggested"><input id="g-{gid}" type="text" inputmode="decimal" value="{g}" aria-describedby="s-{gid}"/><span class="suffix">ieri</span></span>'
        right = f'<div class="ip-alloc__kg is-suggested">{n(kg, 1)}{NB}kg<small>ghid {guide}{NB}g</small><small id="s-{gid}">de confirmat</small></div>'
    elif abs(dev) > 10:  # D19 feed ±10%: warning, saveable
        inp = f'<span class="ip-input ip-input--touch ip-num is-warn"><input id="g-{gid}" type="text" inputmode="decimal" value="{g}" aria-describedby="w-{gid}"/><span class="suffix">g</span></span>'
        right = f'<div class="ip-alloc__kg">{n(kg, 1)}{NB}kg<small>ghid {guide}{NB}g · <span class="ip-num">{"+" if dev > 0 else MINUS}{abs(dev):.0f}%</span></small>{lamp("warn", "abatere")}</div>'
        msg = f'<p class="ip-warnline ip-alloc__msg" id="w-{gid}">{icon("warn", 14)}<span>{sec} <span class="ip-sex ip-sex--{c}">{"♀" if sx == "F" else "♂"}</span>: <span class="ip-num">{g}{NB}g</span>, cu <span class="ip-num">{abs(dev):.0f}%</span> {"peste" if dev > 0 else "sub"} ghidul de <span class="ip-num">{guide}{NB}g</span>. Verifică gramajul; se poate salva.</span></p>'
    else:
        inp = f'<span class="ip-input ip-input--touch ip-num"><input id="g-{gid}" type="text" inputmode="decimal" value="{g}"/><span class="suffix">g</span></span>'
        right = f'<div class="ip-alloc__kg">{n(kg, 1)}{NB}kg<small>ghid {guide}{NB}g</small>{lamp("ok", "în normă")}</div>'
    ginp = f'<label class="ip-field" for="g-{gid}"><span class="ip-label">g/pasăre<span class="ip-sr">, {sec} {word}</span></span>{inp}</label>'
    return f'''<div class="ip-alloc ip-alloc--phone" role="group" aria-label="{sec}, {word}, {n(alive)} în viață">
<div class="ip-alloc__who"><span class="ip-sex ip-sex--{c} ip-t-heading" aria-hidden="true">{"♀" if sx == "F" else "♂"}</span><div><div class="ip-t-body-strong">{sec}</div><div class="ip-note"><span class="ip-num">{n(alive)}</span> {word}</div></div></div>
{right}
<div class="ip-alloc__ctl">{sel}{ginp}</div>
{msg}
</div>''', confirmed, kg, r['feed_type_id'], sx, sec

built = [row(r) for r in ROWS]
halls = {}
for r, b in zip(ROWS, built):
    halls.setdefault(r['section'].split('.')[0], []).append((r, b[0]))
HALLS = '<div class="ip-halls">' + ''.join(
    f'''<section class="ip-hall" aria-labelledby="h-{h}"><header class="ip-hall__head"><span class="ip-cap">Hală</span><h2 class="ip-hall__code" id="h-{h}">{h}</h2><span class="ip-note">secțiuni <span class="ip-num">{" · ".join(dict.fromkeys(r["section"] for r, _ in rs))}</span></span>{pair(n(sum(r["alive"] for r, _ in rs if r["sex"] == "F")), n(sum(r["alive"] for r, _ in rs if r["sex"] == "M")), sr="în viață ")}</header>
<div class="ip-hall__sections">{"".join(b for _, b in rs)}</div></section>''' for h, rs in halls.items()) + '</div>'

# ---------- feed totals (FeedTotals, phone) ----------
fbt = {(f['date'], f['feed_type_id']): f for f in S['feed_by_type_daily']}
STK = {s['feed_type_id']: s for s in S['stock_today']}
def ftrow(t):
    sx = FTS[t]; rs = [b for b in built if b[3] == t]
    conf = [b for b in rs if b[1]]; miss = [b for b in rs if not b[1]]
    azi = round(sum(b[2] for b in conf), 1)
    start = fbt[(YDAY.isoformat(), t)]['stock_end_kg']
    after = start - azi
    st = STK[t]; assert abs(round(after) - st['stock_kg']) <= 1 and abs(azi - fbt[(TODAY.isoformat(), t)]['consumption_kg']) < 0.05
    days = after / st['avg_daily_kg_last3']; assert round(days, 1) == st['days_left']
    run = dt.date.fromisoformat(st['run_out_date'])
    ok = days >= 5
    sign = '♀' if sx == 'F' else '♂'; c = 'hen' if sx == 'F' else 'rooster'
    left = f'fără {sign} ' + ', '.join(b[5] for b in miss) if miss else 'toate confirmate'
    return {'t': t, 'days': days, 'ok': ok, 'html': f'''<div class="ip-ftot__row" role="group" aria-label="{FTN[t]}, pentru {"găini" if sx == "F" else "cocoși"}"><div class="ip-ftot__type"><span class="ip-tile ip-rec ip-rec--feed" aria-hidden="true">{icon("feed")}</span><div><span class="ip-ftot__name">{FTN[t]} <span class="ip-sex ip-sex--{c}" aria-hidden="true">{sign}</span></span><span class="ip-ftot__sub">{len(conf)} din {len(rs)} rânduri confirmate</span></div></div><div class="ip-ftot__cell ip-ftot__cell--kg"><span class="ip-ftot__key">Azi</span><span class="ip-ftot__val">{n(azi, 1)}<small>kg</small></span><span class="ip-ftot__sub">{left}</span></div><div class="ip-ftot__cell ip-ftot__cell--stock"><span class="ip-ftot__key">Stoc după azi</span><span class="ip-ftot__val">{n(round(after))}<small>kg</small></span><span class="ip-ftot__sub">din <span class="ip-num">{n(round(start))}</span> kg</span></div><div class="ip-ftot__cell ip-ftot__cell--days"><span class="ip-ftot__key">Zile rămase</span><span class="ip-ftot__val">{n(days, 1)}<small>zile</small></span><span class="ip-ftot__sub">până pe <span class="ip-num">{dd(run)}</span></span></div><div class="ip-ftot__status">{lamp("ok", "în normă") if ok else lamp("warn", "abatere")}</div></div>'''}
FTR = [ftrow(t) for t in HB_TYPES]
w0, w1 = STK[HB_TYPES[0]]['avg_window']
avg = ', '.join(f'{FTN[t]}: <span class="ip-num">{n(STK[t]["avg_daily_kg_last3"])}</span> kg/zi' for t in HB_TYPES)
FTOT = f'''<section class="ip-ftot" aria-labelledby="ft-h"><div class="ip-ftot__head"><h2 class="ip-cap s7-m0" id="ft-h">Total pe tip de hrană</h2><span class="ip-note"><span class="ip-num">{dd(TODAY)}</span> · doar rânduri confirmate</span></div>
{"".join(f["html"] for f in FTR)}
<p class="ip-note ip-ftot__foot">Zile rămase = stocul după azi ÷ consumul mediu <span class="ip-num">{dt.date.fromisoformat(w0).strftime("%d")}–{dd(dt.date.fromisoformat(w1))}</span> ({avg}). Abatere sub 5 zile.</p></section>'''

# ---------- save bar (compact FeedTotals summary; lamp word "abatere" per D42/D50, detail names type + days) ----------
# One SaveBar shell for the default and the state boards (s7s_build.py imports savebar/sum_rows). Two lines at 390 (S7P):
# line 1 = wheat · "n din m rânduri ·" · kg; line 2 = one status (Lamp + word + detail) or the saved meta. Each segment
# is one no-wrap flex item (additions.css §20), so a narrower bar wraps between segments, never inside a figure.
# "2 de confirmat" is not repeated on line 2: it is 12 − 10 from line 1, and each such row says "de confirmat" (S7P).
FEED14 = f'<span class="ip-rec ip-rec--feed">{icon("feed", 14)}</span>'
def sum_rows(nconf, nrows, kg, lead=''):
    kgs = f'<span><span class="ip-num">{n(kg, 1)}</span> kg</span>' if kg is not None else '<span>– kg</span>'
    rows = f'<span><span class="ip-num">{nconf} din {nrows}</span> rânduri ·</span>' if lead == '' else f'<span>{lead} · <span class="ip-num">{nrows}</span> rânduri ·</span>'
    return f'{FEED14}{rows}{kgs}'
def savebar(line1, line2, key):
    return f'''<div class="ip-savebar s7-savebar"><span class="ip-savebar__sum ip-ftot-sum" id="sb-sum"><span class="ip-ftot-sum__line">{line1}</span><span class="ip-ftot-sum__line">{line2}</span></span>{key}</div>'''
nconf = sum(1 for b in built if b[1]); total = round(sum(b[2] for b in built if b[1]), 1)
worst = min(FTR, key=lambda f: f['days'])
slamp = (lamp('warn', 'abatere') + f'<span>{FTN[worst["t"]]} <span class="ip-num">{n(worst["days"], 1)}</span> zile</span>') if not worst['ok'] else lamp('ok', 'în normă') + '<span>stoc</span>'
SAVE = savebar(sum_rows(nconf, len(built), total), slamp, '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch">Salvează</button>')
assert (nconf, n(total, 1)) == (10, f'3{NB}691,3')

BODY = f'''<div class="s7-screen">
{HEAD}
<main class="s7-main">
{DATE}
{COPY}
{HALLS}
{FTOT}
</main>
{SAVE}
</div>'''

CSS = '''body{margin:0}
html{scroll-padding-bottom:calc(var(--control-h-touch) + var(--space-6) + var(--ip-hair) + var(--ip-hair) + var(--ip-focus-offset) + var(--ip-focus-w))} /* WCAG 2.4.11: SaveBar 73 px at 390 (48 key + 2x12 padding + 1 top line; the 2-line summary, 36 px, sits inside the key height) + the field border and focus ring (1 + 2 + 2), so a focused field scrolls clear of the sticky SaveBar: 78 px */
@media (max-width: 389.98px){html{scroll-padding-bottom:calc(var(--ip-lh-small) * 4 + var(--space-1) * 3 + var(--space-6) + var(--ip-hair) + var(--ip-hair) + var(--ip-focus-offset) + var(--ip-focus-w))}} /* under 390 a summary line can wrap between segments: up to 4 lines at 360 (measured, refusal state: 4x16 + 3x4 + 2x12 + 1 = 101 px) + 5 = 106 px */
.s7-m0{margin:0}
.s7-screen{width:390px;min-height:inherit;display:flex;flex-direction:column;background:var(--ground)}
.s7-title{display:flex;flex-direction:column;min-width:0}
.s7-main{flex:1;display:flex;flex-direction:column;gap:var(--space-6);padding:var(--space-4) var(--space-4) var(--space-8)}
.s7-date{display:flex;gap:var(--space-2);align-items:center}
.s7-date > .ip-input{flex:1;min-width:0}
.s7-date > .ip-input > .ip-icon{color:var(--ink-secondary);flex:none}
.s7-copy{display:flex;flex-direction:column;align-items:flex-start;gap:var(--space-2)}
.s7-copy .ip-pair{font-size:inherit;line-height:inherit;gap:var(--space-1)}
.s7-savebar{position:sticky;bottom:0;z-index:2}'''

FONTS = '<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500&amp;family=IBM+Plex+Mono:wght@400;500&amp;display=swap">'
SYS = ROOT / 'design/system'
STEM, TITLE = 'S7-hrana-zilnica', 'S7 · Hrană zilnică'
root = f'<div class="ip-root" style="padding:0; width:390px; min-height:{H}px">{BODY}</div>'
dc = f'''<!doctype html>
<html lang="ro">
<head>
<meta charset="utf-8">
<title>{TITLE}</title>
<script src="./support.js"></script>
<link rel="stylesheet" href="ds/impexlinepanel/tokens.css">
<link rel="stylesheet" href="ds/impexlinepanel/components/bundle.css">
<link rel="stylesheet" href="additions.css">
</head>
<body>
<x-dc>
<helmet>
{FONTS}
<style>{CSS}</style>
</helmet>
{root}
</x-dc>
<script type="text/x-dc" data-dc-script data-props='{{"$preview":{{"width":390,"height":{H}}}}}'>
class Component extends DCLogic {{ renderVals() {{ return {{}}; }} }}
</script>
</body>
</html>
'''
(ROOT / f'design/canvas/project/{STEM}.dc.html').write_text(dc, encoding='utf-8')
wrap = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8"><title>{TITLE}</title>
<link rel="stylesheet" href="file://{SYS}/tokens.css"><link rel="stylesheet" href="file://{SYS}/components/bundle.css"><link rel="stylesheet" href="file://{SYS}/components/additions.css">
{FONTS}<style>{CSS}</style></head><body>{root}</body></html>'''
pathlib.Path('/tmp/s7').mkdir(exist_ok=True)
pathlib.Path(f'/tmp/s7/{STEM}.html').write_text(wrap, encoding='utf-8')
print('ok', STEM, H, len(dc))
