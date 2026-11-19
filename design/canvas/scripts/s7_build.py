"""Build the S7 · Hrană zilnică phone artboard (default state) from realdata (RF02; D151–D158, today 18.11.2026).
Hîrbovăț, 18.11.2026, SAPT 59, Dr. Ceban (illustrative) has entered every row: the 8 rows of the 4 compartments with birds
(H5.1 H5.2 H6.1 H6.2 × ♀ · ♂, the workbook's g/bird of the day), birds = alive at the start of the day (the fed count, C2:
♀ 14 395 × 160 g = 2 303,2 kg); H7 empty since 14.11 = one muted line (D158d). Composed from system classes only
(FormHeader, DateField, Button, AllocationRow (ip-alloc--phone, S7P), Select, TextField, Lamp, SexCount, HallCard hall head,
FeedTotals + compact SaveBar); screen-only CSS (s7-*) is page layout with tokens. One tall phone board (canvas.md: taller h
for a long scrolling form). savebar()/halls_html() are shared with the state boards (s7s_build.py).
python3 -I s7_build.py <repo_root> [height] [--canvas]   (--canvas: refresh the entry's h/title in place, x and y kept)
Writes design/canvas/project/S7-hrana-zilnica.dc.html and /tmp/s7/S7-hrana-zilnica.html (plain wrapper for renders)."""
import sys, json, pathlib, datetime as dt

ROOT = pathlib.Path(sys.argv[1])
A = [a for a in sys.argv[1:] if a != '--canvas']
H = int(A[1]) if len(A) > 1 else 2466  # FV01 height: + the start-of-day caption (40 px; RF02 2426, was 3162, S7P); s7s/s6 take this default (no int passed by importers)
EX = ROOT / 'design/data/extracts'
sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
import realdata as R  # RF02: every figure, date, week and stock from realdata (DATA03 extracts)
from realdata import NB, MINUS, n, dd, rnd

def __getattr__(name):  # PEP 562: `S` is a legacy alias for rows not yet ported (s6, s13; RF00 §3); the S7 boards never read it
    if name == 'S':
        globals()['S'] = json.load(open(EX / 'sample-hirbovat-2026.json', encoding='utf-8'))  # literal-ok: superseded file, legacy alias only
        return globals()['S']
    raise AttributeError(name)

TODAY = R.TODAY; YDAY = TODAY - dt.timedelta(days=1)
START = R.CYCLE_START                   # Hîrbovăț cycle from 22.02.2026 (SAPT 20)
WEEK = R.WK                             # D151 age week: SAPT 59
assert WEEK == R.sapt(TODAY) == R.C26['sapt_today']

GUIDE = {sx: R.guide_feed(WEEK, sx) for sx in 'FM'}  # guide feed g/bird of the age week (♀ 157 · ♂ 156 at SAPT 59)
assert all(GUIDE.values())
FTN = {t: r['name'] for t, r in R.FEED_TYPES.items()}
FTS = {t: r['for_sex'] for t, r in R.FEED_TYPES.items()}
HB_TYPES = [t for t, r in R.FEED_TYPES.items() if r['farm_id'] == 'hirbovat' and r['active']]
assert [FTS[t] for t in HB_TYPES] == ['F', 'M'] and all(R.FEED_TYPE[FTS[t]] == t for t in HB_TYPES)

def _row(sec, sx):
    """one allocation row of today: birds at the start of the day (the fed count), the day's g (today) and the day before's
    (yesterday = the Copiază ieri suggestion), the workbook's kg."""
    t, y = R.sd(TODAY, sec, sx), R.sd(YDAY, sec, sx)
    return {'section': sec, 'sex': sx, 'alive': R.alive_start(sec, sx), 'guide_g': GUIDE[sx], 'feed_type_id': R.FEED_TYPE[sx],
            'today_g': t['g'] if t and t['kg'] else None, 'yesterday_g': y['g'] if y and y['kg'] else None, 'kg': t['kg'] if t else None}
ROWS = [_row(sec, sx) for sec in R.HSECS for sx in 'FM' if R.alive_start(sec, sx)]
EMPTY = {h: R.empty_since(h) for h in R.HHALLS if not any(R.alive_start(s, sx) for s in R.HALL_SECS[h] for sx in 'FM')}
assert len(ROWS) == 2 * len([s for s in R.HSECS if R.has_birds(s, YDAY)]) and all(EMPTY.values())
assert sum(r['alive'] for r in ROWS if r['sex'] == 'F') == R.alive_farm('F', YDAY) == R.day(TODAY)['alive_F'] + R.deaths(TODAY, sx='F') + R.sold(TODAY, sx='F')
# stock at the start of the day and the 3-day mean consumption (days left, run-out date: realdata.stock_days)
STK0 = {t: R.stock(YDAY)[FTS[t]] for t in HB_TYPES}
AVGK = {t: R.stock_days(FTS[t])[1] for t in HB_TYPES}
AVG_WIN = (TODAY - dt.timedelta(days=2), TODAY)
FOOT = (f'Zile rămase = stocul după azi ÷ consumul mediu <span class="ip-num">{R.drange(*AVG_WIN)}</span> ('
        + ', '.join(f'{FTN[t]}: <span class="ip-num">{n(AVGK[t])}</span> kg/zi' for t in HB_TYPES) + '). Abatere sub 5 zile.')

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
# FX07/D142: every entry form's back link goes to Înregistrează (the Sidebar item / tab bar key, #inregistreaza)
FARM = 'HÎRBOVĂȚ'
HEAD = f'''<header class="ip-formhead">
<a class="ip-iconbtn ip-iconbtn--touch" href="#inregistreaza" aria-label="Înapoi la Înregistrează">{icon("back")}</a>
<span class="ip-tile ip-rec ip-rec--feed" aria-hidden="true">{icon("feed")}</span>
<div class="s7-title"><h1 class="ip-t-heading ip-m0">Hrană zilnică</h1><p class="ip-note ip-num ip-m0">{FARM} · {dd(TODAY)} · SAPT {WEEK}</p></div>
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
<p class="ip-note ip-m0">ghid <span class="ip-num">SAPT {WEEK}</span>: {pair(f"{GUIDE['F']}{NB}g", f"{GUIDE['M']}{NB}g")}</p>
</div>'''

# ---------- allocation rows ----------
# D181: the counts on the form (hall heads, rows) are the birds alive at the START of the day (the day's g × birds = kg; S4/S10/S15
# show end-of-day counts): one caption above the halls says so, and every count's text alternative carries the same words.
START_SR = 'în viață la începutul zilei'
def start_caption(d):
    return f'<p class="ip-note ip-m0">Păsări {START_SR}, <span class="ip-num">{dd(d)}</span></p>'
def row(r):
    sx, sec, alive = r['sex'], r['section'], r['alive']
    c = 'hen' if sx == 'F' else 'rooster'
    word = 'găini' if sx == 'F' else 'cocoși'
    gid = f'{sec.replace(".", "-")}-{sx}'
    guide = r['guide_g']; assert guide == GUIDE[sx]
    confirmed = r['today_g'] is not None
    g = r['today_g'] if confirmed else r['yesterday_g']
    kg = rnd(g * alive / 1000, 1)
    if confirmed: assert kg == r['kg'], (sec, sx, kg, r['kg'])
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
    return f'''<div class="ip-alloc ip-alloc--phone" role="group" aria-label="{sec}, {word}, {n(alive)} {START_SR}">
<div class="ip-alloc__who"><span class="ip-sex ip-sex--{c} ip-t-heading" aria-hidden="true">{"♀" if sx == "F" else "♂"}</span><div><div class="ip-t-body-strong">{sec}</div><div class="ip-note"><span class="ip-num">{n(alive)}</span> {word}<span class="ip-sr"> {START_SR}</span></div></div></div>
{right}
<div class="ip-alloc__ctl">{sel}{ginp}</div>
{msg}
</div>''', confirmed, kg, r['feed_type_id'], sx, sec

built = [row(r) for r in ROWS]
def halls_html(pairs):
    """HallCards for [(row, AllocationRow html)] in hall order; a hall without birds today (H7 since 14.11) keeps its head and
    shows one muted line "goală din <date>" (D158d, as S4's D161 line), no rows."""
    hs = {}
    for r, h in pairs: hs.setdefault(r['section'].split('.')[0], []).append((r, h))
    out = ''
    for h in R.HHALLS:
        if h in EMPTY:
            secs, f, m = R.HALL_SECS[h], 0, 0
            body = f'<p class="ip-note ip-m0">{lamp("off", "goală")} din <span class="ip-num">{dd(EMPTY[h])}</span></p>'
        else:
            rs = hs[h]; secs = list(dict.fromkeys(r['section'] for r, _ in rs))
            f = sum(r['alive'] for r, _ in rs if r['sex'] == 'F'); m = sum(r['alive'] for r, _ in rs if r['sex'] == 'M')
            body = f'<div class="ip-hall__sections">{"".join(x for _, x in rs)}</div>'
        out += f'''<section class="ip-hall" aria-labelledby="h-{h}"><header class="ip-hall__head"><span class="ip-cap">Hală</span><h2 class="ip-hall__code" id="h-{h}">{h}</h2><span class="ip-note">secțiuni <span class="ip-num">{" · ".join(secs)}</span></span>{pair(n(f), n(m), sr=START_SR + " ")}</header>
{body}</section>'''
    assert set(hs) | set(EMPTY) == set(R.HHALLS)
    return start_caption(TODAY) + '<div class="ip-halls">' + out + '</div>'
HALLS = halls_html([(r, b[0]) for r, b in zip(ROWS, built)])

# ---------- feed totals (FeedTotals, phone) ----------
def ftrow(t):
    sx = FTS[t]; rs = [b for b in built if b[3] == t]
    conf = [b for b in rs if b[1]]; miss = [b for b in rs if not b[1]]
    azi = rnd(sum(b[2] for b in conf), 1)
    start = STK0[t]
    after = start - azi
    assert rnd(after, 1) == R.stock(TODAY)[sx] and azi == R.feed_kg(TODAY, sx)  # the day's real consumption and end stock
    days, _, run = R.stock_days(sx); assert abs(days - after / AVGK[t]) < 1e-9
    ok = days >= 5
    sign = '♀' if sx == 'F' else '♂'; c = 'hen' if sx == 'F' else 'rooster'
    left = f'fără <span class="ip-sex ip-sex--{c}" aria-hidden="true">{sign}</span><span class="ip-sr">{"găini" if sx == "F" else "cocoși"}</span> ' + ', '.join(b[5] for b in miss) if miss else 'toate confirmate'  # FIN08: coloured sign
    return {'t': t, 'days': days, 'ok': ok, 'html': f'''<div class="ip-ftot__row" role="group" aria-label="{FTN[t]}, pentru {"găini" if sx == "F" else "cocoși"}"><div class="ip-ftot__type"><span class="ip-tile ip-rec ip-rec--feed" aria-hidden="true">{icon("feed")}</span><div><span class="ip-ftot__name">{FTN[t]} <span class="ip-sex ip-sex--{c}" aria-hidden="true">{sign}</span></span><span class="ip-ftot__sub">{len(conf)} din {len(rs)} rânduri confirmate</span></div></div><div class="ip-ftot__cell ip-ftot__cell--kg"><span class="ip-ftot__key">Azi</span><span class="ip-ftot__val">{n(azi, 1)}<small>kg</small></span><span class="ip-ftot__sub">{left}</span></div><div class="ip-ftot__cell ip-ftot__cell--stock"><span class="ip-ftot__key">Stoc după azi</span><span class="ip-ftot__val">{n(after)}<small>kg</small></span><span class="ip-ftot__sub">din <span class="ip-num">{n(start)}</span> kg</span></div><div class="ip-ftot__cell ip-ftot__cell--days"><span class="ip-ftot__key">Zile rămase</span><span class="ip-ftot__val">{n(days, 1)}<small>zile</small></span><span class="ip-ftot__sub">până pe <span class="ip-num">{dd(run)}</span></span></div><div class="ip-ftot__status">{lamp("ok", "în normă") if ok else lamp("warn", "abatere")}</div></div>'''}
FTR = [ftrow(t) for t in HB_TYPES]
FTOT = f'''<section class="ip-ftot" aria-labelledby="ft-h"><div class="ip-ftot__head"><h2 class="ip-cap ip-m0" id="ft-h">Total pe tip de hrană</h2><span class="ip-note"><span class="ip-num">{dd(TODAY)}</span> · doar rânduri confirmate</span></div>
{"".join(f["html"] for f in FTR)}
<p class="ip-note ip-ftot__foot">{FOOT}</p></section>'''

# ---------- save bar (compact FeedTotals summary; lamp word "abatere" per D42/D50, detail names type + days) ----------
# One SaveBar shell for the default and the state boards (s7s_build.py imports savebar/sum_rows). Two lines at 390 (S7P):
# line 1 = wheat · "n din m rânduri ·" · kg; line 2 = one status (Lamp + word + detail) or the saved meta. Each segment
# is one no-wrap flex item (additions.css §20), so a narrower bar wraps between segments, never inside a figure.
# An unconfirmed count is not repeated on line 2: it is m − n from line 1, and each such row says "de confirmat" (S7P).
FEED14 = f'<span class="ip-rec ip-rec--feed">{icon("feed", 14)}</span>'
def sum_rows(nconf, nrows, kg, lead=''):
    kgs = f'<span><span class="ip-num">{n(kg, 1)}</span> kg</span>' if kg is not None else '<span>– kg</span>'
    rows = f'<span><span class="ip-num">{nconf} din {nrows}</span> rânduri ·</span>' if lead == '' else f'<span>{lead} · <span class="ip-num">{nrows}</span> rânduri ·</span>'
    return f'{FEED14}{rows}{kgs}'
def savebar(line1, line2, key):
    return f'''<div class="ip-savebar ip-screen__savebar"><span class="ip-savebar__sum ip-savebar__sum--2l" id="sb-sum"><span class="ip-savebar__line">{line1}</span><span class="ip-savebar__line">{line2}</span></span>{key}</div>'''
nconf = sum(1 for b in built if b[1]); total = rnd(sum(b[2] for b in built if b[1]), 1)
worst = min(FTR, key=lambda f: f['days'])
slamp = (lamp('warn', 'abatere') + f'<span>{FTN[worst["t"]]} <span class="ip-num">{n(worst["days"], 1)}</span> zile</span>') if not worst['ok'] else lamp('ok', 'în normă') + '<span>stoc</span>'
SAVE = savebar(sum_rows(nconf, len(built), total), slamp, '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch">Salvează</button>')
assert (nconf, total) == (len(ROWS), rnd(R.feed_kg(TODAY, 'F') + R.feed_kg(TODAY, 'M'), 1))  # screens.md #s7: 8 rânduri · 2 457,9 kg

BODY = f'''<div class="ip-screen">
{HEAD}
<main class="ip-screen__main">
{DATE}
{COPY}
{HALLS}
{FTOT}
</main>
{SAVE}
</div>'''

CSS = '''body{margin:0}
.s7-title{display:flex;flex-direction:column;min-width:0}
.s7-date{display:flex;gap:var(--space-2);align-items:center}
.s7-date > .ip-input{flex:1;min-width:0}
.s7-date > .ip-input > .ip-icon{color:var(--ink-secondary);flex:none}
.s7-copy{display:flex;flex-direction:column;align-items:flex-start;gap:var(--space-2)}
.s7-copy .ip-pair{font-size:inherit;line-height:inherit;gap:var(--space-1)}'''  # FIN15B: the sticky SaveBar is additions.css §63 ip-screen__savebar; FX05: its scroll-padding too

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
if '--canvas' in sys.argv[2:] and __name__ == '__main__':  # RF02: refresh the hand-placed entry's h/title in place (x, y kept; canvas_rows.place)
    import canvas_rows as CR
    c = CR.load(ROOT); CR.place(c, STEM, [(STEM, 390, H, TITLE)]); CR.save(ROOT, c)
print('ok', STEM, H, len(dc))
