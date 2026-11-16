"""Build S4 · Fermă · Rezumat artboard from the extracts.
python3 -I s4_build.py <repo_root> <height>
Writes design/canvas/project/S4-ferma-rezumat.dc.html and /tmp/s4/render.html (plain wrapper)."""
import sys, json, pathlib, datetime as dt

ROOT = pathlib.Path(sys.argv[1]); H = int(sys.argv[2]) if len(sys.argv) > 2 else 1700
EX = ROOT / 'design/data/extracts'
S = json.load(open(EX / 'sample-hirbovat-2026.json', encoding='utf-8'))
CL = json.load(open(EX / 'counts-latest.json', encoding='utf-8'))
GL = json.load(open(EX / 'guide-laying.json', encoding='utf-8'))
CY = json.load(open(EX / 'cycles.json', encoding='utf-8'))

NB = ' '; MINUS = '−'

def n(v, dec=0):
    s = f'{abs(v):,.{dec}f}'.replace(',', 'X').replace('.', ',').replace('X', NB)
    return (MINUS if v < 0 else '') + s

def sgn(v, dec=1):
    return ('+' if v > 0 else '') + n(v, dec)

def ddate(iso):
    y, m, d = iso.split('-'); return f'{d}.{m}.{y}'

# ---------- icons (copied path data from RecordTile / Sidebar / PageHeader previews) ----------
IC = {
 'mortality': '<path d="M12 3a7 7 0 0 0-7 7v4l2 2v3h10v-3l2-2v-4a7 7 0 0 0-7-7z"/><circle cx="9.5" cy="11" r="1.2"/><circle cx="14.5" cy="11" r="1.2"/><path d="M11 19v-2M13 19v-2"/>',
 'cull': '<circle cx="12" cy="12" r="9"/><path d="M5.7 5.7l12.6 12.6"/>',
 'transfer': '<path d="m16 3 4 4-4 4"/><path d="M20 7H4"/><path d="m8 21-4-4 4-4"/><path d="M4 17h16"/>',
 'sale': '<path d="M3 7h11v9H3zM14 10h4l3 3v3h-7"/><circle cx="7" cy="18" r="2"/><circle cx="17" cy="18" r="2"/>',
 'feed': '<path d="M12 21V9M12 9c-3 0-4-2-4-5 3 0 4 2 4 5zM12 9c3 0 4-2 4-5-3 0-4 2-4 5zM12 15c-3 0-4-2-4-5 3 0 4 2 4 5zM12 15c3 0 4-2 4-5-3 0-4 2-4 5z"/>',
 'delivery': '<path d="M12 3 3 7.5v9L12 21l9-4.5v-9zM3 7.5 12 12l9-4.5M12 12v9"/>',
 'price': '<ellipse cx="9" cy="7" rx="6" ry="3"/><path d="M3 7v4c0 1.7 2.7 3 6 3s6-1.3 6-3V7"/><path d="M9 14v3c0 1.7 2.7 3 6 3s6-1.3 6-3v-4c0-1.7-2.7-3-6-3"/>',
 'weighing': '<path d="M12 3v18"/><path d="M7 21h10"/><path d="M3 7h18"/><path d="m6 7-3 7a3 3 0 0 0 6 0z"/><path d="m18 7-3 7a3 3 0 0 0 6 0z"/>',
 'eggs': '<path d="M12 22c4 0 7-3 7-7.5C19 9 16 2 12 2S5 9 5 14.5C5 19 8 22 12 22z"/>',
 'sexerr': '<circle cx="8" cy="12" r="4" stroke="var(--rec-hen)"/><circle cx="16" cy="12" r="4" stroke="var(--rec-rooster)"/>',
 'chev': '<path d="m6 9 6 6 6-6"/>',
 'cycle': '<path d="M21 12a9 9 0 0 1-15.5 6.2L3 16M3 12a9 9 0 0 1 15.5-6.2L21 8"/><path d="M21 3v5h-5M3 21v-5h5"/>',
 'download': '<path d="M12 3v12"/><path d="m7 10 5 5 5-5"/><path d="M5 21h14"/>',
 'home': '<path d="M3 10.5 12 3l9 7.5V21H3z"/><path d="M9 21v-6h6v6"/>',
 'plus': '<path d="M5 12h14"/><path d="M12 5v14"/>',
 'layers': '<path d="m12 2 10 5-10 5L2 7z"/><path d="m2 12 10 5 10-5"/><path d="m2 17 10 5 10-5"/>',
 'chart': '<path d="M3 3v18h18"/><path d="m7 15 4-4 3 3 5-6"/>',
 'settings': '<path d="M4 6h10"/><path d="M18 6h2"/><circle cx="16" cy="6" r="2"/><path d="M4 18h4"/><path d="M12 18h8"/><circle cx="10" cy="18" r="2"/>',
}
REC = {'mortality': 'mortality', 'cull': 'cull', 'transfer': 'transfer', 'sale': 'sale', 'feed': 'feed',
       'delivery': 'feed', 'price': 'feed', 'weighing': 'weighing', 'eggs': 'eggs'}

def icon(k, size=None):
    cls = 'ip-icon' + (f' ip-icon--{size}' if size else '')
    return f'<svg class="{cls}" viewBox="0 0 24 24" aria-hidden="true">{IC[k]}</svg>'

def tile(k):
    return f'<span class="ip-tile ip-rec ip-rec--{REC[k]}" aria-hidden="true">{icon(k)}</span>'

def ico(k):  # small record icon in its colour, for captions and key/value lines
    return f'<span class="ip-rec ip-rec--{REC[k]} s4-ico">{icon(k, 14)}</span>'

def hen_v(v):  # hen figure without the sign (the label already carries ♀)
    return f'<span class="ip-sex ip-sex--hen"><b>{v}</b></span>'

def hen(v, unit=''):
    return f'<span class="ip-sex ip-sex--hen">♀ <b>{v}{unit}</b></span>'

def roo(v, unit=''):
    return f'<span class="ip-sex ip-sex--rooster">♂ <b>{v}{unit}</b></span>'

def sg(sx):  # bare ♀/♂ sign in its record colour
    return '<span class="ip-sex ip-sex--hen">♀</span>' if sx == 'F' else '<span class="ip-sex ip-sex--rooster">♂</span>'
F, M = sg('F'), sg('M')

def pair(f, m, unit=''):
    return f'<span class="ip-pair">{hen(f, unit)}<span class="sep">·</span>{roo(m, unit)}</span>'

def lamp(kind, word):
    return f'<span class="ip-lamp ip-lamp--{kind}">{word}</span>'

# ---------- data ----------
farm = CL['farms'][0]
alive = {(x['section'], sx): x[sx] for x in farm['sections'] for sx in 'FM'}
SECTIONS = ['H1.1', 'H1.2', 'H3.1', 'H3.2', 'H4.1', 'H4.2']
TODAY = '2026-11-16'; WK = 40
wk_rows = [r for r in S['daily'] if '2026-11-10' <= r['date'] <= TODAY]
mort_wk = {s: sum(r['mortality'] for r in wk_rows if r['section'] == s) for s in SECTIONS}
mort_wk_sex = {sx: sum(r['mortality'] for r in wk_rows if r['sex'] == sx) for sx in 'FM'}
alloc = {(a['section'], a['sex']): a for a in S['allocations_today']}
wt = {(w['section'], w['sex']): w['avg_weight_g'] for w in S['weighings'] if w['date'] == TODAY}
def wavg(d, sx):
    keys = [k for k in d if k[1] == sx and d[k] is not None]
    return sum(d[k] * alive[k] for k in keys) / sum(alive[k] for k in keys)
w_f = round(wavg(wt, 'F')); w_m = round(wavg(wt, 'M'))
g_today = {k: a['today_g'] for k, a in alloc.items()}
g_f = round(wavg(g_today, 'F')); g_m = round(wavg(g_today, 'M'))
guide = {(r['cycle_week'], r['sex']): r for r in GL['rows']}
gf_g = guide[(WK, 'F')]['feed_g']; gm_g = guide[(WK, 'M')]['feed_g']
eggs_today = {r['hall']: r for r in S['eggs_daily'] if r['date'] == TODAY}
lay_all = {r['cycle_week']: r for r in S['eggs_weekly'] if r['hall'] == 'all'}
st = S['summary_today']
lay_today = st['lay_pct_today']; lay_guide = st['guide_lay_pct']
stocks = {s['feed_type_id']: s for s in S['stock_today']}
unpriced = [d for d in S['deliveries'] if not d['priced']]
unpriced_kg = sum(d['qty_kg'] for d in unpriced)
cum = farm['cycle']
tot = farm['totals']

# ---------- sidebar ----------
sidebar = f'''<nav class="ip-sidebar s4-side" aria-label="Navigare">
<div class="ip-wordmark">Impex Line</div>
<div class="ip-series"><span class="ip-sr" id="ser-lbl">Serie</span><div class="ip-input ip-series__btn" id="ser-combo" role="combobox" tabindex="0" aria-haspopup="listbox" aria-expanded="false" aria-controls="ser-list" aria-labelledby="ser-lbl"><span class="name">Seria 2026</span>{lamp('ok', 'deschisă')}{icon('chev', 16)}</div>
<ul class="ip-menu ip-series__menu" role="listbox" id="ser-list" aria-labelledby="ser-lbl" tabindex="-1" hidden="hidden"><li class="ip-menu__group" role="presentation">Serii</li><li class="ip-menu__item ip-series__opt" role="option" id="ser-2026" aria-selected="true"><span class="chk"></span><span class="txt"><span class="top"><b>Seria 2026</b>{lamp('ok', 'deschisă')}</span><span class="ip-num">din 07.10.2025</span></span></li><li class="ip-menu__item ip-series__opt" role="option" id="ser-2025" aria-selected="false"><span class="chk"></span><span class="txt"><span class="top"><b>Seria 2025</b>{lamp('off', 'închisă')}</span><span class="ip-num">până la 28.01.2026</span></span></li></ul></div>
<a class="ip-nav" href="#acasa">{icon('home')}Acasă</a>
<a class="ip-nav" href="#inregistreaza">{icon('plus')}Înregistrează</a>
<a class="ip-nav is-current" href="#serie" aria-current="page">{icon('layers')}Serie</a>
<a class="ip-nav" href="#rapoarte">{icon('chart')}Rapoarte</a>
<a class="ip-nav" href="#setari">{icon('settings')}Setări</a>
<span class="s4-grow"></span>
<div class="ip-note s4-account"><b>Proprietar</b><a class="ip-link" href="#iesi">Ieși</a></div>
</nav>'''

# ---------- 1. page header ----------
header = f'''<header class="ip-pagehead">
<div>
<nav class="ip-crumbs" aria-label="Cale"><a href="#seria-2026">Seria 2026</a><span aria-hidden="true">›</span><span aria-current="page">Hîrbovăț</span></nav>
<h1>Hîrbovăț</h1>
<div class="ip-row s4-meta"><span class="ip-tag ip-tag--neutral">Ouat</span><span class="ip-tag ip-tag--neutral">Ross 308 ouat</span>{pair(n(tot['F']), n(tot['M']))}<span class="ip-note ip-num">SAPT {WK} · ciclu din 10.02.2026</span></div>
</div>
<div class="ip-row">
<button type="button" class="ip-btn ip-btn--secondary">{icon('transfer', 16)}Mută păsări</button>
<button type="button" class="ip-btn ip-btn--secondary">{icon('sale', 16)}Vânzare</button>
<button type="button" class="ip-btn ip-btn--ghost" aria-haspopup="menu" aria-expanded="false" aria-label="Acțiuni ciclu">{icon('cycle', 16)}Ciclu{icon('chev', 14)}</button>
</div>
</header>'''

# ---------- 2. cycle selector + farm track ----------
TW = 800  # track drawn 1:1 at the 1440 layout (content 1160 − selector 336 − gap 24)
d0 = dt.date(2025, 3, 1); d1 = dt.date(2027, 1, 31)
ppd = TW / (d1 - d0).days
def tx(iso):
    return round(((dt.date.fromisoformat(iso) - d0).days) * ppd, 1)
c25 = next(c for c in CY['cycles'] if c['id'] == 'cyc-hirbovat-2025')
c26 = next(c for c in CY['cycles'] if c['id'] == 'cyc-hirbovat-2026')
x25a, x25b = tx(c25['start_date']), tx(c25['end_date'])
x26a, xt = tx(c26['start_date']), tx(TODAY)
ticks = ''
for iso, lbl in [('2025-04-01', '04.25'), ('2025-07-01', '07.25'), ('2025-10-01', '10.25'), ('2026-01-01', '01.26'),
                 ('2026-04-01', '04.26'), ('2026-07-01', '07.26'), ('2026-10-01', '10.26'), ('2027-01-01', '01.27')]:
    x = tx(iso); ticks += f'<line class="tick" x1="{x}" x2="{x}" y1="40" y2="45"/><text class="month" x="{x}" y="56" text-anchor="middle">{lbl}</text>'
assert c25['start_inferred'] and c25['start_date'] == '2025-03-07' and c25['end_date'] == '2026-01-28'  # D54: no start event, ended by emptying (no sale)
lab25 = 'Seria 2025, ouat, început dedus 07.03.2025 · încheiat prin golire 28.01.2026, 47 săptămâni, închis. Deschide ciclul.'
lab26 = f'Seria 2026, ouat, din 10.02.2026, deschis, azi SAPT {WK}. Început prin transfer. Ciclul afișat.'
w25 = round(x25b - x25a, 1); w26 = round(xt - x26a, 1)
track = f'''<svg class="ip-track ip-track--farm" viewBox="0 0 {TW} 60" role="group" aria-label="Ciclurile fermei Hîrbovăț"><desc>Hîrbovăț: Seria 2025 ouat, început dedus 07.03.2025 · încheiat prin golire 28.01.2026, 47 săpt., închis; Seria 2026 ouat din 10.02.2026 (transfer), deschis, azi 16.11.2026, SAPT {WK}.</desc><rect class="lane" x="0" y="14" width="{TW}" height="26"/><a href="#ciclu-seria-2025" aria-label="{lab25}"><title>{lab25}</title><rect class="ring" x="{round(x25a-4,1)}" y="12" width="{round(w25+8,1)}" height="30" rx="3"/><rect class="cyc" x="{x25a}" y="16" width="{w25}" height="18" rx="2"/><text class="cyctxt" x="{round(x25a+7,1)}" y="29">Seria 2025 · ouat · 47 săpt.</text><line class="s4-ded" x1="{round(x25a-3,1)}" x2="{round(x25a-3,1)}" y1="14" y2="40"/><rect class="sel" x="{x25a}" y="37" width="{w25}" height="2"/></a><a href="#ciclu-seria-2026" aria-current="true" aria-label="{lab26}"><title>{lab26}</title><rect class="ring" x="{round(x26a-4,1)}" y="12" width="{round(w26+8,1)}" height="30" rx="3"/><rect class="cyc open" x="{x26a}" y="16" width="{w26}" height="18" rx="2"/><text class="cyctxt open" x="{round(x26a+7,1)}" y="29">Seria 2026 · ouat · din 10.02.2026</text><rect class="xfer" x="{round(x26a-1,1)}" y="16" width="3" height="18"/><rect class="sel" x="{x26a}" y="37" width="{w26}" height="2"/></a><line class="today" x1="{xt}" x2="{xt}" y1="12" y2="42"/><text class="todaylbl" x="{xt}" y="9" text-anchor="middle">azi 16.11</text><rect class="ro" x="{round(xt+4,1)}" y="17" width="48" height="16" rx="2"/><text class="rotxt" x="{round(xt+28,1)}" y="28.5" text-anchor="middle">SAPT {WK}</text>{ticks}</svg>'''
cycle = f'''<section class="s4-cycle" aria-label="Ciclu">
<div class="ip-field"><span class="ip-label" id="cyc-lbl">Ciclu</span><div class="ip-input s4-cycsel" role="combobox" tabindex="0" aria-haspopup="listbox" aria-expanded="false" aria-labelledby="cyc-lbl"><span class="s4-grow">Seria 2026 · din 10.02.2026</span>{lamp('ok', 'deschis')}{icon('chev', 16)}</div><span class="ip-hint">2 cicluri la Hîrbovăț · deschis din 10.02.2026</span></div>
<div class="ip-stack s4-tight">{track}<div class="ip-tracklegend" aria-hidden="true"><span><i class="ip-rec ip-rec--transfer"></i>transfer</span><span><svg class="s4-dedkey" viewBox="0 0 3 12"><line class="s4-ded" x1="1.5" x2="1.5" y1="0" y2="12"/></svg>început dedus</span><span><i class="today"></i>azi</span></div></div>
</section>'''

# ---------- 3. filter bar + 4. tabs ----------
def seg(label, items, sel):
    its = ''.join(f'<span class="ip-seg__item{" is-selected" if i == sel else ""}" role="radio" aria-checked="{"true" if i == sel else "false"}" tabindex="{0 if i == sel else -1}">{t}</span>' for i, t in enumerate(items))
    return f'<span class="ip-seg" role="radiogroup" aria-label="{label}">{its}</span>'
def chip(k, word, on=True):
    return f'<button type="button" class="ip-chip ip-rec ip-rec--{REC[k]}{"" if on else " ip-chip--off"}" role="checkbox" aria-checked="{"true" if on else "false"}">{icon(k, 16)}{word}</button>'
# Eroare sexare: neutral chip led by the split ♀/♂ tile (two rings), as the RecordTile code
chip_sexerr = f'<button type="button" class="ip-chip ip-rec s4-chipsplit" role="checkbox" aria-checked="true"><span class="ip-tile ip-tile--split" aria-hidden="true">{icon("sexerr", 16)}</span>Eroare sexare</button>'
# D45: every record type with records in "Ciclul curent" (transfer in 10.02.2026, sexing errors 18.03 and 09.04.2026)
assert S['sexing_errors'] and all(e['date'] >= '2026-02-10' for e in S['sexing_errors'])
filterbar = f'''<div class="ip-filterbar s4-filterbar" role="group" aria-label="Filtre">{seg('Perioadă', ['Ciclul curent', '4 săpt.', '13 săpt.', 'Interval'], 0)}{seg('Pas', ['Zi', 'Săpt.'], 1)}<button type="button" class="ip-btn ip-btn--ghost ip-filterbar__end">{icon('download', 16)}Exportă .xlsx</button><div class="s4-chips" role="group" aria-label="Tipuri de înregistrări">{chip('mortality', 'Mortalitate')}{chip('cull', 'Sacrificare')}{chip_sexerr}{chip('transfer', 'Transfer')}{chip('feed', 'Hrană')}{chip('weighing', 'Cântărire')}{chip('eggs', 'Ouă')}</div></div>'''
TABS = [('rezumat', 'Rezumat'), ('pasari', 'Păsări'), ('hrana', 'Hrană'), ('greutate', 'Greutate'), ('productie', 'Producție')]
# one tabpanel (#tp-ferma) whose content each tab swaps; roving tabindex
tabs = '<div class="ip-tabs" role="tablist" aria-label="Vederi Hîrbovăț">' + ''.join(
    f'<a class="ip-tab{" is-selected" if i == 0 else ""}" role="tab" id="tab-{k}" aria-controls="tp-ferma" aria-selected="{"true" if i == 0 else "false"}" tabindex="{0 if i == 0 else -1}" href="#{k}">{t}</a>'
    for i, (k, t) in enumerate(TABS)) + '</div>'

# ---------- 5a. readout + guide chart ----------
CW, CH = 548, 220  # readout height = the 3 Stat rows (no stretched band)
X0, X1, YT, YB = 40, CW - 8, 10, CH - 22
WMAX = 45
def cx(w): return round(X0 + (w - 1) * (X1 - X0) / (WMAX - 1), 1)
def cy(p): return round(YB - p / 100 * (YB - YT), 1)
gweeks = [w for w in range(1, WMAX + 1) if guide[(w, 'F')]['lay_pct'] is not None]
gv = {w: guide[(w, 'F')]['lay_pct'] for w in gweeks}
band = ' '.join(f'{cx(w)},{cy(min(gv[w]+3,100))}' for w in gweeks) + ' ' + ' '.join(f'{cx(w)},{cy(gv[w]-3)}' for w in reversed(gweeks))
gline = ' '.join(f'{cx(w)},{cy(gv[w])}' for w in gweeks)
aweeks = sorted(w for w in lay_all if w <= WK)
aline = ' '.join(f'{cx(w)},{cy(lay_all[w]["lay_pct"])}' for w in aweeks)
grid = ''
for p, lbl in [(0, '0 %'), (50, '50 %'), (100, '100 %')]:
    cls = 'axis' if p == 0 else 'grid'
    grid += f'<line class="{cls}" x1="{X0}" x2="{X1}" y1="{cy(p)}" y2="{cy(p)}"/><text class="ax" x="{X0-4}" y="{cy(p)+3}" text-anchor="end">{lbl}</text>'
xl = ''.join(f'<text class="ax" x="{cx(w)}" y="{CH-6}" text-anchor="middle">{w}</text>' for w in [1, 10, 20, 30, 40, 45])
a40 = lay_all[WK]['lay_pct']; dev = round(a40 - gv[WK], 1)
chart = f'''<svg class="ip-chart" viewBox="0 0 {CW} {CH}" role="group" aria-roledescription="grafic" tabindex="0" aria-label="Ouat Hîrbovăț față de ghid, pe săptămânile ciclului, SAPT 3–{WK}; ghid SAPT 7–45" aria-describedby="lay-help lay-read">{grid}<polygon class="band" points="{band}"/><polyline class="guide" points="{gline}"/><polyline class="act" points="{aline}"/><line class="today" x1="{cx(WK)}" x2="{cx(WK)}" y1="{YT-2}" y2="{YB}"/><line class="cur" x1="{cx(WK)}" x2="{cx(WK)}" y1="{YT-2}" y2="{YB}"/><circle class="pt" cx="{cx(WK)}" cy="{cy(a40)}" r="3.5"/>{xl}</svg>'''
rows = ''.join(f'<tr><td>{w}</td><td>{n(lay_all[w]["lay_pct"],1)} %</td><td>{(n(gv[w],1)+" %") if w in gv else "– fără date în ghid"}</td></tr>' for w in aweeks)
readout = f'''<div class="ip-readout s4-readout">
<div class="ip-readout__cap"><span>OUAT · HÎRBOVĂȚ · SAPT {WK}</span><span class="ip-rlamp ip-rlamp--ok">ÎN NORMĂ</span></div>
<div class="ip-readout__val">{n(a40,1)}<small> %</small></div>
<div class="ip-readout__sub">ghid {n(gv[WK],1)} % · {sgn(dev)} pct · 10–16.11.2026</div>
{chart}
<span class="ip-sr" id="lay-help">Săgețile stânga și dreapta mută săptămâna; Home și End: prima și ultima săptămână.</span>
<div class="ip-chart__read" id="lay-read" role="status" aria-live="polite">SAPT {WK} · efectiv {n(a40,1)} % · ghid {n(gv[WK],1)} % · {sgn(dev)} pct</div>
<table class="ip-sr"><caption>Ouat Hîrbovăț pe săptămâni, efectiv față de ghid; înregistrări din 01.03.2026 (SAPT 3)</caption><tr><th scope="col">SAPT</th><th scope="col">Efectiv</th><th scope="col">Ghid</th></tr>{rows}</table>
<div class="ip-readout__sub">— efectiv · - - ghid · bandă ±3 pct · axa: săptămâna ciclului</div>
</div>'''

# ---------- 5b. stats ----------
def cap(k, text):
    return f'<span class="ip-cap s4-capi">{ico(k) if k else ""}{text}</span>'
s_b2 = stocks['ft-breeder2']; s_co = stocks['ft-cocosi']
stats = f'''<div class="ip-stats s4-stats">
<div class="ip-stat">{cap(None, 'În viață')}<span class="ip-stat__val">{pair(n(tot['F']), n(tot['M']))}</span><span class="ip-stat__guide ip-num">total {n(tot['all'])}</span><span class="ip-stat__guide">la sosire {pair(n(cum['start_counts']['F']), n(cum['start_counts']['M']))}</span></div>
<div class="ip-stat">{cap('mortality', 'Mortalitate cumulată')}<span class="ip-stat__val">{pair(n(cum['mortality_pct']['F'],2), n(cum['mortality_pct']['M'],2))}<small>%</small></span><span class="ip-stat__guide">ghid: – fără date în ghid</span><span class="ip-stat__guide">din 10.02.2026 {pair(n(-cum['removals']['F']['mortality']), n(-cum['removals']['M']['mortality']))}</span></div>
<div class="ip-stat">{cap('weighing', f'Greutate {F} / {M}')}<span class="ip-stat__val">{pair(n(w_f), n(w_m))}<small>g</small></span><span class="ip-stat__guide">ghid: – fără date în ghid</span><span class="ip-stat__guide ip-num">cântărit 16.11.2026 · 6 secțiuni</span></div>
<div class="ip-stat">{cap('feed', 'Hrană azi pe pasăre')}<span class="ip-stat__val">{pair(g_f, g_m)}<small>g</small></span><span class="ip-stat__guide">ghid {pair(gf_g, gm_g)} {lamp('ok', 'în normă')}</span><span class="ip-stat__guide"><span>fără {F} H4.2 · {M} H3.2 (lipsă azi)</span></span></div>
<div class="ip-stat">{cap('feed', 'Stoc hrană · zile rămase')}<span class="ip-stat__val">{n(s_co['days_left'],1)}<small>zile · Cocoși</small></span><span class="ip-stat__guide"><span class="ip-num">până pe {ddate(s_co['run_out_date'])}</span>{lamp('warn', 'abatere')}</span><span class="ip-stat__guide"><span>Breeder 2 · <span class="ip-num">{n(s_b2['days_left'],1)} zile · {n(s_b2['stock_kg'])} kg</span></span></span></div>

<div class="ip-stat">{cap('eggs', 'Ouat azi')}<span class="ip-stat__val">{n(lay_today,1)}<small>%</small></span><span class="ip-stat__guide">ghid {n(lay_guide,1)} % · {sgn(round(lay_today-lay_guide,1))} pct {lamp('ok', 'în normă')}</span><span class="ip-stat__guide ip-num">{n(sum(e['total_eggs'] for e in eggs_today.values()))} ouă · 3 hale</span></div>
</div>'''

# ---------- 5c. section tiles grouped by hall + totals ----------
# tile status from the figures (Lamp words only, D42): lipsă = a feed row missing today;
# abatere = feed ±10 % or the hall's lay ±3 pct (D19); lay is per hall, so both sections of the hall carry it
def tstate(sec):
    hall = sec.split('.')[0]; out = []
    if g_today[(sec, 'F')] is None or g_today[(sec, 'M')] is None: out.append(('warn', 'lipsă'))
    feed_dev = any(g_today[(sec, sx)] is not None and abs(g_today[(sec, sx)] / guide[(WK, sx)]['feed_g'] - 1) > 0.10 for sx in 'FM')
    lay_dev = abs(eggs_today[hall]['lay_pct'] - eggs_today[hall]['guide_lay_pct']) > 3
    if feed_dev or lay_dev: out.append(('warn', 'abatere'))
    return out or [('ok', 'în normă')]
def gpair(sec):
    f = g_today[(sec, 'F')]; m = g_today[(sec, 'M')]
    if f is None or m is None:  # missing value: "– lipsă", no unit on it
        miss = '<span class="s4-miss">– lipsă</span>'
        fs = F + miss if f is None else hen(f) + '<span class="s4-unit">g</span>'
        ms = M + miss if m is None else roo(m) + '<span class="s4-unit">g</span>'
        return f'<span class="ip-pair s4-kvpair">{fs}<span class="sep">·</span>{ms}</span>'
    return f'<span class="ip-pair s4-kvpair">{hen(f)}<span class="sep">·</span>{roo(m)}<span class="s4-unit">g</span></span>'
def stile(sec):
    hall = sec.split('.')[0]; f, m = alive[(sec, 'F')], alive[(sec, 'M')]
    pf = round(f / (f + m) * 100)
    missing = g_today[(sec, 'F')] is None or g_today[(sec, 'M')] is None
    lamps = ''.join(lamp(k, w) for k, w in tstate(sec))
    return f'''<a href="#sectiune-{sec}" class="ip-stile{" is-missing" if missing else ""}"><div class="ip-stile__head"><span class="ip-stile__code">{sec}</span><span class="s4-lamps">{lamps}</span></div>{pair(n(f), n(m))}<div class="ip-bar" aria-hidden="true"><span style="width:{pf}%"></span><span style="width:{100-pf}%"></span></div><div class="ip-stile__kv"><span class="s4-capi">{ico('mortality')}Mort. săpt.</span><b>{n(-mort_wk[sec])}</b></div><div class="ip-stile__kv"><span class="s4-capi">{ico('feed')}Hrană azi</span>{gpair(sec)}</div><div class="ip-stile__kv"><span class="s4-capi">{ico('weighing')}Greutate {F}</span><span class="ip-pair s4-kvpair">{hen_v(n(wt[(sec,'F')]))}<span class="s4-unit">g</span></span></div><div class="ip-stile__kv"><span class="s4-capi">{ico('eggs')}Ouat {hall} · ghid {n(eggs_today[hall]['guide_lay_pct'],1)} %</span><b>{n(eggs_today[hall]['lay_pct'],1)} %</b></div></a>'''
halls = ''
for hall in ['H1', 'H3', 'H4']:
    secs = [s for s in SECTIONS if s.startswith(hall + '.')]
    hf = sum(alive[(s, 'F')] for s in secs); hm = sum(alive[(s, 'M')] for s in secs)
    halls += f'<div class="s4-hall" role="group" aria-label="Hala {hall}"><div class="s4-hallhead"><span class="ip-cap">Hala {hall}</span>{pair(n(hf), n(hm))}</div><div class="ip-tiles">{"".join(stile(s) for s in secs)}</div></div>'
totals = f'''<div class="ip-stotals" role="group" aria-label="Total fermă Hîrbovăț, 16.11.2026"><div><span class="ip-cap">Total fermă</span><span class="ip-stotals__sub">6 secțiuni</span><span class="ip-stotals__sub ip-num">16.11.2026 · SAPT {WK}</span></div><div><span class="ip-stotals__key">În viață</span><span class="ip-stotals__val">{pair(n(tot['F']), n(tot['M']))}</span><span class="ip-stotals__sub ip-num">total {n(tot['all'])}</span></div><div><span class="ip-stotals__key">{ico('mortality')}Mort. săpt.</span><span class="ip-stotals__val">{n(-(mort_wk_sex['F']+mort_wk_sex['M']))}</span><span class="ip-stotals__sub">{pair(n(-mort_wk_sex['F']), n(-mort_wk_sex['M']))}</span></div><div><span class="ip-stotals__key">{ico('feed')}Hrană azi</span><span class="ip-stotals__val">{pair(g_f, g_m)}<small>g</small></span><span class="ip-stotals__sub s4-nowrap"><span>ghid <span class="ip-num">{gf_g}/{gm_g} g</span></span>{lamp('ok', 'în normă')}</span><span class="ip-stotals__sub s4-nowrap"><span>fără {F} H4.2 · {M} H3.2</span></span></div><div><span class="ip-stotals__key">{ico('weighing')}Greutate {F}</span><span class="ip-stotals__val">{hen_v(n(w_f))}<small>g</small></span><span class="ip-stotals__sub">ghid: – fără date în ghid</span><span class="ip-stotals__sub">medie pe găini</span></div><div><span class="ip-stotals__key">{ico('eggs')}Ouat</span><span class="ip-stotals__val">{n(lay_today,1)}<small>%</small></span><span class="ip-stotals__sub">ghid {n(lay_guide,1)} % {lamp('ok', 'în normă')}</span></div></div>'''

# ---------- 5d. alerts ----------
h3 = st['lay_pct_week40_by_hall']['H3']; fd = st['feed_deviation_today']
# first day of the current run of H3 days more than 3 pct under the guide (D19): 09.11 (72,1 vs 75,4)
h3_days = sorted((r for r in S['eggs_daily'] if r['hall'] == 'H3' and r['date'] <= TODAY and r['guide_lay_pct'] is not None), key=lambda r: r['date'])
h3_since = None
for r in reversed(h3_days):
    if r['lay_pct'] - r['guide_lay_pct'] < -3: h3_since = r['date']
    else: break
assert h3_since == '2026-11-09', h3_since
def alert(k, title, meta, status, act=None):
    a = f'<a class="ip-link ip-ev__act" href="{act[1]}">{act[0]}</a>' if act else ''
    return f'<div class="ip-ev ip-ev--alert">{tile(k)}<div class="ip-ev__main">{title}<div class="ip-ev__meta ip-num">{meta}</div></div>{a}<span class="ip-ev__status">{status}</span></div>'
alerts = [
 alert('eggs', f'H3 · ouat <span class="ip-num">{n(h3["lay_pct"],1)} %</span> · ghid <span class="ip-num">{n(lay_guide,1)} %</span> · <span class="ip-num">{n(h3["diff_pts"],1)} pct</span>',
       f'16.11.2026 · SAPT {WK} · sub ghid din {ddate(h3_since)}', lamp('warn', 'abatere'), ('Vezi producția', '#productie')),
 alert('feed', f'H1.2 · hrană {roo(fd["g"], " g")} · ghid {roo(fd["guide_g"], " g")} · <span class="ip-num">+{n(fd["deviation_pct"])} %</span>',
       f'16.11.2026 · SAPT {WK} · Dr. Ceban · ieri {roo(alloc[("H1.2","M")]["yesterday_g"], " g")}', lamp('warn', 'abatere'), ('Deschide secțiunea', '#sectiune-H1.2')),
 alert('feed', f'H3.2 · hrană {M} azi – lipsă · ghid {roo(gm_g, " g")}',
       f'16.11.2026 · SAPT {WK} · de la 12:00 · ieri {roo(alloc[("H3.2","M")]["yesterday_g"], " g")}', lamp('warn', 'lipsă'), ('Deschide secțiunea', '#sectiune-H3.2')),
 alert('feed', f'H4.2 · hrană {F} azi – lipsă · ghid {hen(gf_g, " g")}',
       f'16.11.2026 · SAPT {WK} · de la 12:00 · ieri {hen(alloc[("H4.2","F")]["yesterday_g"], " g")}', lamp('warn', 'lipsă'), ('Deschide secțiunea', '#sectiune-H4.2')),
 alert('delivery', f'Cocoși · stoc <span class="ip-num">{n(s_co["stock_kg"])} kg</span> · <span class="ip-num">{n(s_co["days_left"],1)} zile</span> · limită 5 zile',
       f'16.11.2026 · se termină pe {ddate(s_co["run_out_date"])} · {s_co["avg_daily_kg_last3"]} kg/zi', lamp('warn', 'abatere'), ('Deschide stocul', '#hrana-stoc')),
 alert('delivery', f'Hrană · {len(unpriced)} livrări fără preț · <span class="ip-num">{n(unpriced_kg)} kg</span>',
       f'{ddate(unpriced[0]["date"])[:5]}–{ddate(unpriced[-1]["date"])} · Ion Rusu · așteaptă prețul', lamp('info', 'informație'), ('Adaugă prețurile', '#preturi-hrana')),
]

# ---------- 5e. latest records (event log, 16.11 and 15.11) ----------
def ev(k, title, meta, status=''):
    return f'<a class="ip-ev" href="#inregistrare">{tile(k)}<div class="ip-ev__main">{title}<div class="ip-ev__meta ip-num">{meta}</div></div>{status}</a>'
log = {(e['at'], e['record']): e for e in S['event_log']}
events = [
 ev('weighing', f'Cântărire · 6 secțiuni · {pair(n(w_f), n(w_m), " g")}', f'16.11.2026 · 10:10 · SAPT {WK} · Dr. Ceban'),
 ev('feed', f'Hrană zilnică · 10 din 12 rânduri · <span class="ip-num">3{NB}691 kg</span>', f'16.11.2026 · 09:47 · SAPT {WK} · Dr. Ceban', lamp('warn', 'lipsă 2')),
 ev('eggs', f'Ouă · H1, H3, H4 · <span class="ip-num">18{NB}688</span> ouă', f'16.11.2026 · 08:35 · SAPT {WK} · Dr. Ceban'),
 ev('mortality', f'Mortalitate · 6 secțiuni · {hen(n(-8))} · {roo(n(-1))}', f'16.11.2026 · 08:34 · SAPT {WK} · Dr. Ceban'),
 ev('mortality', f'Corecție mortalitate · H4.1 · {F} · <span class="ip-num">14.11</span> · {hen_v(n(-3))} → {hen_v(n(-5))}', f'15.11.2026 · 10:12 · SAPT {WK} · Ion Rusu', '<span class="ip-tag ip-tag--neutral">corectat</span>'),
 ev('feed', f'Hrană zilnică · 12 din 12 rânduri · <span class="ip-num">4{NB}393 kg</span>', f'15.11.2026 · 09:47 · SAPT {WK} · Dr. Ceban'),
 ev('eggs', f'Ouă · H1, H3, H4 · <span class="ip-num">18{NB}679</span> ouă', f'15.11.2026 · 08:21 · SAPT {WK} · Dr. Ceban'),
 ev('cull', f'Sacrificare · H3.1 · {roo(n(-2))} · boală', f'15.11.2026 · 08:20 · SAPT {WK} · Dr. Ceban'),
 ev('mortality', f'Mortalitate · 6 secțiuni · {hen(n(-7))} · {roo(n(-4))}', f'15.11.2026 · 08:20 · SAPT {WK} · Dr. Ceban'),
]
# sanity: every event above exists in the extract log
for at in ['2026-11-16T10:10', '2026-11-16T09:47', '2026-11-16T08:35', '2026-11-16T08:34', '2026-11-15T10:12', '2026-11-15T09:47', '2026-11-15T08:21', '2026-11-15T08:20']:
    assert any(e['at'] == at for e in S['event_log']), at

main = f'''<main class="s4-main">
{header}
{cycle}
<div class="s4-tight">{filterbar}{tabs}</div>
<div class="s4-panel" id="tp-ferma" role="tabpanel" aria-labelledby="tab-rezumat">
<section class="s4-top" aria-labelledby="h-key"><h2 class="ip-sr" id="h-key">Indicatori față de ghid</h2>{readout}{stats}</section>
<section class="s4-block" aria-labelledby="h-sec"><div class="s4-head"><h2 class="ip-t-heading" id="h-sec">Secțiuni</h2><span class="ip-note ip-num">6 secțiuni · 3 hale · 16.11.2026</span></div><div class="s4-halls">{halls}</div>{totals}</section>
<div class="s4-split">
<section class="s4-block" aria-labelledby="h-chk"><div class="s4-head"><h2 class="ip-t-heading" id="h-chk">De verificat</h2><span class="ip-note"><span class="ip-num">{len(alerts)}</span> alerte · <span class="ip-num">16.11.2026</span></span></div><div class="ip-list">{"".join(alerts)}</div></section>
<section class="s4-block" aria-labelledby="h-log"><div class="s4-head"><h2 class="ip-t-heading" id="h-log">Ultimele înregistrări</h2><a class="ip-link ip-note" href="#pasari">Tot jurnalul</a></div><div class="ip-list">{"".join(events)}</div></section>
</div>
</div>
</main>'''

CSS = '''body{margin:0}
.s4-shell{display:flex;align-items:stretch}
.s4-side{flex:none;min-height:0}
.s4-grow{flex:1}
.s4-account{display:flex;justify-content:space-between;align-items:center;padding:var(--space-2);border-top:var(--ip-hair) solid var(--line)}
.s4-account b{color:var(--ink);font-weight:500}
.s4-main{flex:1;min-width:0;padding:var(--space-6);display:flex;flex-direction:column;gap:var(--space-6)}
.s4-tight{display:flex;flex-direction:column;gap:var(--space-2)}
.s4-panel{display:flex;flex-direction:column;gap:var(--space-6)}
.s4-lamps{display:flex;gap:var(--space-3)}
.s4-meta{gap:var(--space-2) var(--space-3)}
.s4-cycle{display:grid;grid-template-columns:calc(var(--space-12) * 7) minmax(0,1fr);gap:var(--space-6);align-items:start}
.s4-cycsel{cursor:pointer}
.s4-top{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:var(--space-4);align-items:stretch}
.s4-readout{justify-content:flex-start}
.s4-stats{grid-template-columns:repeat(2,minmax(0,1fr))}
.s4-block{display:flex;flex-direction:column;gap:var(--space-3);min-width:0}
.s4-head{display:flex;justify-content:space-between;align-items:baseline;gap:var(--space-3)}
.s4-head h2{margin:0}
.s4-halls{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:var(--space-4)}
.s4-hall{display:flex;flex-direction:column;gap:var(--space-2);min-width:0}
.s4-hallhead{display:flex;justify-content:space-between;align-items:baseline;gap:var(--space-2);font-size:var(--ip-fs-small);line-height:var(--ip-lh-small)}
.s4-split{display:grid;grid-template-columns:minmax(0,7fr) minmax(0,5fr);gap:var(--space-4);align-items:start}
.s4-capi{display:flex;gap:var(--space-1);align-items:center}
.s4-ico{color:var(--c);display:inline-flex}
.s4-kvpair{gap:var(--space-1);font-size:inherit;line-height:inherit}
.s4-kvpair .ip-sex b{font:inherit}
.s4-unit{color:var(--ink-muted);font-family:var(--font-sans)}
.s4-miss{color:var(--ink-muted);font-family:var(--font-sans);margin-left:var(--space-1)}
.s4-nowrap{flex-wrap:nowrap;white-space:nowrap}
.s4-filterbar{row-gap:var(--space-2)}
.s4-chips{flex-basis:100%;display:flex;flex-wrap:wrap;gap:var(--space-2);align-items:center}
.s4-chipsplit{color:var(--ink);padding-left:var(--space-1)}
.s4-chipsplit .ip-tile{width:var(--space-5);height:var(--space-5);border-radius:var(--radius-xs)}
/* D54: an inferred cycle start (no recorded placement/transfer) is drawn as a dashed steel line across the lane, not a record mark */
.s4-ded{stroke:var(--ink-secondary);stroke-width:var(--ip-edge-strong);stroke-dasharray:var(--ip-mark-w) var(--ip-mark-w)}
.s4-dedkey{width:var(--ip-mark-w);height:var(--space-3);flex:none;overflow:visible}'''

body = f'<div class="s4-shell">{sidebar}{main}</div>'
FONTS = '<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500&amp;family=IBM+Plex+Mono:wght@400;500&amp;display=swap">'
dc = f'''<!doctype html>
<html lang="ro">
<head>
<meta charset="utf-8">
<title>S4 · Fermă · Rezumat</title>
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
<div class="ip-root" style="padding:0; min-height:{H}px">{body}</div>
</x-dc>
<script type="text/x-dc" data-dc-script data-props='{{"$preview":{{"width":1440,"height":{H}}}}}'>
class Component extends DCLogic {{ renderVals() {{ return {{}}; }} }}
</script>
</body>
</html>
'''
out = ROOT / 'design/canvas/project/S4-ferma-rezumat.dc.html'
out.write_text(dc, encoding='utf-8')
SYS = ROOT / 'design/system'
wrap = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8"><title>S4</title>
<link rel="stylesheet" href="file://{SYS}/tokens.css"><link rel="stylesheet" href="file://{SYS}/components/bundle.css"><link rel="stylesheet" href="file://{SYS}/components/additions.css">
{FONTS}<style>{CSS}</style></head><body><div class="ip-root" style="padding:0; min-height:{H}px">{body}</div></body></html>'''
pathlib.Path('/tmp/s4').mkdir(exist_ok=True)
pathlib.Path('/tmp/s4/render.html').write_text(wrap, encoding='utf-8')
print('ok', out, len(dc))
