"""Build the S4 · Fermă tab boards (S4T): Păsări, Hrană, Greutate, Producție, one desktop board per tab (owner, 16.11.2026,
Hîrbovăț · Seria 2026, period "Ciclul curent", grain "Săpt." as the S4 default board). Rezumat is the S4 default board.
python3 -I s4t_build.py <repo_root>
Imports the S4 shell from s4s_build.py (sidebar, header, cycle selector + track, FilterBar pieces, formats, icons, write());
S4TP: the S4 boards share s4s.CSS_S4, s4s.gline (Stat lines) and s4s.rnd (.5 up, D82) with S4/S4S. Figures come only from the extracts
(sample-hirbovat-2026.json, counts-latest.json, guide-laying.json, transfers-seria2026.json); every derived figure is
asserted against the figure the S4 board already shows. Writes design/canvas/project/S4-ferma--<tab>.dc.html and
/tmp/s4s/S4-ferma--<tab>.html (plain wrappers for the render script)."""
import sys, re, json, pathlib, importlib.util, collections as C, datetime as dt

_spec = importlib.util.spec_from_file_location('s4s_build', pathlib.Path(__file__).resolve().with_name('s4s_build.py'))
s4 = importlib.util.module_from_spec(_spec)
DROP = globals().get('DROP')   # S16P fix 1: a deleted Mortalitate record (S16 after-delete board only); unset by default
if DROP: s4.DROP = DROP
_spec.loader.exec_module(s4)  # reads sys.argv[1] (repo root)

n, sgn, ddate, pct, icon, tile, ico, hen, roo, pair, lamp, sg = s4.n, s4.sgn, s4.ddate, s4.pct, s4.icon, s4.tile, s4.ico, s4.hen, s4.roo, s4.pair, s4.lamp, s4.sg
F, M, NB = s4.F, s4.M, s4.NB
# D82: .5 rounds up (away from zero), never to even; float noise is cut at 1e-9 first so 3 332,4999… still reads 3 333.
from decimal import Decimal, ROUND_HALF_UP
def rnd(v, dec=0):
    d = Decimal(repr(float(v))).quantize(Decimal('1e-9')).quantize(Decimal(1).scaleb(-dec), rounding=ROUND_HALF_UP)
    return float(d) if dec else int(d)
def n(v, dec=0): return s4.n(rnd(v, dec), dec)
def sgn(v, dec=1):
    r = rnd(v, dec); return ('+' if r > 0 else '') + n(r, dec)
def pct(v, dec=1): return n(v, dec) + '%'
S, CL, GL, TR = s4.S, s4.CL, s4.GL, s4.TR
TODAY, WK, SECTIONS, HALLS = s4.TODAY, s4.WK, s4.SECTIONS, ['H1', 'H3', 'H4']
START = dt.date(2026, 2, 10)
def cw(iso): return (dt.date.fromisoformat(iso) - START).days // 7 + 1
def ddm(iso): return ddate(iso)[:5]
def num(v): return f'<span class="ip-num">{v}</span>'
DASH = '–'
NBSP = '\u00a0'  # no-break space: a ♀/♂ sign, its figure and unit stay on one line
NOGD = 'ghid: – fără date în ghid'
HERE = pathlib.Path(__file__).resolve().parent
HEIGHTS = json.loads((HERE / 's4t_heights.json').read_text()) if (HERE / 's4t_heights.json').exists() else {}

# ---------------------------------------------------------------- shell: tabs + filter bar
TABS = [('rezumat', 'Rezumat'), ('pasari', 'Păsări'), ('hrana', 'Hrană'), ('greutate', 'Greutate'), ('productie', 'Producție')]
def tabs(sel):  # same markup as S4 (roving tabindex, one tabpanel #tp-ferma that every tab swaps); only the selection moves
    return '<div class="ip-tabs" role="tablist" aria-label="Vederi Hîrbovăț">' + ''.join(
        f'<a class="ip-tab{" is-selected" if k == sel else ""}" role="tab" id="tab-{k}" aria-controls="tp-ferma" aria-selected="{"true" if k == sel else "false"}" tabindex="{0 if k == sel else -1}" href="#{k}">{t}</a>'
        for k, t in TABS) + '</div>'
# Păsări: the event log is the bird types (D1); the 7 chips of the period stay listed (D45/D47), the other 3 are off.
BIRD = {'mortality', 'cull', 'sexerr', 'transfer'}
def chips(on):
    sx = 'sexerr' in on
    split = f'<button type="button" class="ip-chip ip-rec ip-chip--split{"" if sx else " ip-chip--off"}" role="checkbox" aria-checked="{"true" if sx else "false"}"><span class="ip-tile ip-tile--split" aria-hidden="true">{icon("sexerr", 16)}</span>Eroare sexare</button>'
    c = s4.chip
    return (f'<div class="s4-chips" role="group" aria-label="Tipuri de înregistrări">{c("mortality", "Mortalitate", "mortality" in on)}{c("cull", "Sacrificare", "cull" in on)}{split}'
            f'{c("transfer", "Transfer", "transfer" in on)}{c("feed", "Hrană", "feed" in on)}{c("weighing", "Cântărire", "weighing" in on)}{c("eggs", "Ouă", "eggs" in on)}</div>')
def filterbar(tab):
    ch = chips(BIRD) if tab == 'pasari' else s4.CHIPS
    return f'<div class="ip-filterbar s4-filterbar" role="group" aria-label="Filtre">{s4.seg("Perioadă", s4.PERIODS, 0)}{s4.seg("Pas", ["Zi", "Săpt."], 1)}{s4.EXPORT}{ch}</div>'
def page(tab, panel):
    main = f'''<main class="s4-main">
{s4.header('default')}
{s4.cycle('default')}
<div class="s4-tight">{filterbar(tab)}{tabs(tab)}</div>
<div class="s4-panel" id="tp-ferma" role="tabpanel" aria-labelledby="tab-{tab}">
{panel}
</div>
</main>'''
    return s4.shell(main)

# ---------------------------------------------------------------- Readout + GuideChart (S4 geometry: 540×216, drawn 1:1)
CW, CH = s4.CW, s4.CH
YT, YB = 10, CH - 22
def gchart(cid, label, wmax, series, ylo, yhi, yticks, yfmt, x0=40, guide=None, band=None, xticks=(1, 10, 20, 30, 40)):
    """series: [{'cls': 'act'|'s2', 'pt': 'pt'|'pt2', 'pts': {week: value}}]; guide {week: v}; band (lo, hi) dicts."""
    X1 = CW - 8
    def X(w): return round(x0 + (w - 1) * (X1 - x0) / (wmax - 1), 1)
    def Y(v): return round(YB - (v - ylo) / (yhi - ylo) * (YB - YT), 1)
    g = ''
    for i, t in enumerate(yticks):
        g += f'<line class="{"axis" if i == 0 else "grid"}" x1="{x0}" x2="{X1}" y1="{Y(t)}" y2="{Y(t)}"/><text class="ax" x="{x0 - 4}" y="{Y(t) + 3}" text-anchor="end">{yfmt(t)}</text>'
    if band:
        lo, hi = band; ws = sorted(lo)
        g += '<polygon class="band" points="' + ' '.join(f'{X(w)},{Y(hi[w])}' for w in ws) + ' ' + ' '.join(f'{X(w)},{Y(lo[w])}' for w in reversed(ws)) + '"/>'
    if guide:
        g += '<polyline class="guide" points="' + ' '.join(f'{X(w)},{Y(v)}' for w, v in sorted(guide.items())) + '"/>'
    for s in series:
        g += f'<polyline class="{s["cls"]}" points="' + ' '.join(f'{X(w)},{Y(v)}' for w, v in sorted(s['pts'].items())) + '"/>'
    g += f'<line class="today" x1="{X(WK)}" x2="{X(WK)}" y1="{YT - 2}" y2="{YB}"/><line class="cur" x1="{X(WK)}" x2="{X(WK)}" y1="{YT - 2}" y2="{YB}"/>'
    for s in series:
        g += f'<circle class="{s.get("pt", "pt")}" cx="{X(WK)}" cy="{Y(s["pts"][WK])}" r="3.5"/>'
    g += ''.join(f'<text class="ax" x="{X(w)}" y="{CH - 6}" text-anchor="middle">{w}</text>' for w in xticks)
    return (f'<svg class="ip-chart" viewBox="0 0 {CW} {CH}" role="group" aria-roledescription="grafic" tabindex="0" aria-label="{label}" '
            f'aria-describedby="{cid}-help {cid}-read"><g aria-hidden="true">{g}</g></svg>')
SW = {'act': '<svg viewBox="0 0 24 8" aria-hidden="true"><line class="act" x1="0" x2="24" y1="4" y2="4"/></svg>',
      's2': '<svg viewBox="0 0 24 8" aria-hidden="true"><line class="s2" x1="0" x2="24" y1="4" y2="4"/></svg>',
      'guide': '<svg viewBox="0 0 24 8" aria-hidden="true"><line class="guide" x1="0" x2="24" y1="4" y2="4"/></svg>',
      'band': '<svg viewBox="0 0 24 8" aria-hidden="true"><rect class="band" x="0" y="0" width="24" height="8"/></svg>'}
def legend(items, note=''):
    return '<div class="ip-chart__legend">' + ''.join(f'<span>{SW[k]}{t}</span>' for k, t in items) + (f'<span>{note}</span>' if note else '') + '</div>'
def readout(cid, cap, rl, val, sub, chart, read, cols, rows, caption, leg):
    rlamp = {'ok': '<span class="ip-rlamp ip-rlamp--ok">ÎN NORMĂ</span>', 'warn': '<span class="ip-rlamp ip-rlamp--warn">ABATERE</span>',
             'off': '<span class="ip-rlamp ip-rlamp--off">FĂRĂ DATE ÎN GHID</span>'}[rl]
    tb = ('<table class="ip-sr"><caption>' + caption + '</caption><thead><tr>' + ''.join(f'<th scope="col">{c}</th>' for c in cols) + '</tr></thead><tbody>'
          + ''.join('<tr><th scope="row">' + r[0] + '</th>' + ''.join(f'<td>{c}</td>' for c in r[1:]) + '</tr>' for r in rows) + '</tbody></table>')
    return f'''<div class="ip-readout s4-readout">
<div class="ip-readout__cap"><span>{cap}</span>{rlamp}</div>
<div class="ip-readout__val">{val}</div>
<div class="ip-readout__sub">{sub}</div>
{chart}
<span class="ip-sr" id="{cid}-help">Săgețile stânga și dreapta mută săptămâna; Home și End: prima și ultima săptămână.</span>
<div class="ip-chart__read" id="{cid}-read" role="status" aria-live="polite">{read}</div>
{tb}
{leg}
</div>'''
HB, RB = '<b>♀</b>', '<b>♂</b>'
def rpair(f, m, unit=''):  # ♀ · ♂ figures on the dark Readout: signs in readout-ink (rec-hen/rooster are <3:1 there, D58)
    return f'♀ {f}<small>{unit}</small> · ♂ {m}<small>{unit}</small>'

# ---------------------------------------------------------------- Stat + table helpers
def stat(k, capt, val, *lines):
    c = s4.cap(k, capt) if k != 'sexerr' else f'<span class="ip-cap s4-capi"><span class="s4-ico" aria-hidden="true">{icon("sexerr", 14)}</span>{capt}</span>'
    return f'<div class="ip-stat">{c}<span class="ip-stat__val">{val}</span>' + ''.join(f'<span class="ip-stat__guide">{s4.gline(l)}</span>' for l in lines) + '</div>'
def stats(*items): return '<div class="ip-stats s4-stats">' + ''.join(items) + '</div>'
def top(hid, title, ro, st):
    return f'<section class="s4-top" aria-labelledby="{hid}"><h2 class="ip-sr" id="{hid}">{title}</h2>{ro}{st}</section>'
def block(hid, title, note, body):
    return f'<section class="s4-block" aria-labelledby="{hid}"><div class="s4-head"><h2 class="ip-t-heading" id="{hid}">{title}</h2><span class="ip-note">{note}</span></div>{body}</section>'
def tbox(cap, inner, cls=''):  # DataTable in the system scroll region (SYS06 .ip-tablescroll, framed) with system row headers (§32)
    return (f'<div class="ip-tablescroll ip-tablescroll--framed" role="region" tabindex="0" aria-label="{cap}"><table class="ip-table ip-table--rowheads{(" " + cls) if cls else ""}">'
            f'<caption class="ip-sr">{cap}</caption>{inner}</table></div>')
def foot(note, link=None):
    lk = f'<a class="ip-link s4t-lk" href="{link[1]}">{link[0]}</a>' if link else ''
    return f'<div class="s4t-foot"><span class="ip-note">{note}</span>{lk}</div>'
def thico(k): return f'<span class="ip-rec ip-rec--{s4.REC.get(k, k)} s4t-thico">{icon(k, 14)}</span>'
def thsex(sx):
    return (f'<span class="ip-sex ip-sex--hen">♀</span><span class="ip-sr"> găini</span>' if sx == 'F'
            else f'<span class="ip-sex ip-sex--rooster">♂</span><span class="ip-sr"> cocoși</span>')
SXC = {'F': 's4t-hen', 'M': 's4t-roo'}

# ================================================================= 1. PĂSĂRI
STARTC = {sx: CL['farms'][0]['cycle']['start_counts'][sx] for sx in 'FM'}
assert STARTC == TR['arrival_totals'] == {'F': 26765, 'M': 3060}
mw = C.defaultdict(lambda: {'F': 0, 'M': 0})
for r in S['daily']: mw[r['cycle_week']][r['sex']] += r['mortality']
cum = {'F': 0, 'M': 0}; mF, mM = {}, {}
for w in sorted(mw):
    for sx in 'FM': cum[sx] += mw[w][sx]
    mF[w] = cum['F'] / STARTC['F'] * 100; mM[w] = cum['M'] / STARTC['M'] * 100
assert (round(mF[WK], 2), round(mM[WK], 2)) == (S['summary_today']['cumulative_mortality_pct']['F'], S['summary_today']['cumulative_mortality_pct']['M'])
assert min(mw) == 3 and max(mw) == WK
wkF, wkM = mw[WK]['F'], mw[WK]['M']
assert (wkF, wkM) == (s4.mort_wk_sex['F'], s4.mort_wk_sex['M'])  # S4 totals strip: ♀ −51 · ♂ −13
ST = {(r['section'], r['sex']): r for r in S['section_totals']}
for k, r in ST.items():
    assert r['start'] - r['mortality'] - r['culls'] - r['sexing_errors'] == r['alive'] == s4.alive[k], k
culls = S['culls']; cullsx = {sx: sum(c['qty'] for c in culls if c['sex'] == sx) for sx in 'FM'}
assert cullsx == {sx: sum(ST[(s, sx)]['culls'] for s in SECTIONS) for sx in 'FM'}
reasons = C.Counter(c['reason'] for c in culls)
sexerr = S['sexing_errors']; seF = sum(e['qty'] for e in sexerr if e['sex'] == 'F'); seM = sum(e['qty'] for e in sexerr if e['sex'] == 'M')
assert (seF, seM) == (2, 0)
assert not any(e['record'] == 'sale' for e in S['event_log'])  # no sale at Hîrbovăț in the cycle
tot = s4.tot

def ro_pasari():
    rows = [[str(w), DASH + ' fără date', DASH + ' fără date', DASH + ' fără date în ghid'] for w in (1, 2)]
    rows += [[str(w), pct(mF[w], 2), pct(mM[w], 2), DASH + ' fără date în ghid'] for w in sorted(mF)]
    lbl = f'Mortalitate cumulată găini și cocoși, Hîrbovăț Seria 2026, SAPT 3–{WK}; ghidul nu are mortalitate'
    ch = gchart('mort', lbl, WK, [{'cls': 'act', 'pts': mF}, {'cls': 's2', 'pt': 'pt2', 'pts': mM}], 0, 15, [0, 5, 10, 15], lambda t: f'{t}%')
    return readout('mort', f'MORTALITATE CUMULATĂ · HÎRBOVĂȚ · SAPT {WK}', 'off', rpair(n(mF[WK], 2), n(mM[WK], 2), '%'),
                   f'din sosire ♀ {n(STARTC["F"])} · ♂ {n(STARTC["M"])} · SAPT {WK}: ♀ {n(-wkF)} · ♂ {n(-wkM)}',
                   ch, f'SAPT {WK} · ♀ {pct(mF[WK], 2)} · ♂ {pct(mM[WK], 2)} · ghid {DASH} fără date în ghid',
                   ['SAPT', 'Găini ♀', 'Cocoși ♂', 'Ghid'], rows, 'Mortalitate cumulată pe săptămânile ciclului, % din sosire; înregistrări din 01.03.2026 (SAPT 3)',
                   legend([('act', f'{HB} găini'), ('s2', f'{RB} cocoși')], 'axa: săpt. ciclului, 1–2 fără date'))

def st_pasari():
    out = {sx: sum(ST[(s, sx)]['mortality'] + ST[(s, sx)]['culls'] + ST[(s, sx)]['sexing_errors'] for s in SECTIONS) for sx in 'FM'}
    rs = ' · '.join(f'{r} {num(reasons[r])}' for r in ('boală', 'șchiopătare', 'altul') if reasons[r])
    sed = ', '.join(ddate(e['date']) for e in sorted(sexerr, key=lambda e: e['date']))
    return stats(
        stat(None, 'În viață', pair(n(tot['F']), n(tot['M'])), f'<span class="ip-num">total {n(tot["all"])}</span>', f'ieșiri în ciclu {pair(n(-out["F"]), n(-out["M"]))}'),
        stat('mortality', f'Mortalitate · SAPT {WK}', pair(n(-wkF), n(-wkM)), NOGD, f'<span class="ip-num">10.11–16.11.2026 · 7 zile</span>'),
        stat('cull', 'Sacrificare în ciclu', pair(n(-cullsx['F']), n(-cullsx['M'])), f'{num(len(culls))} înregistrări', f'<span>{rs}</span>'),
        stat('sexerr', 'Eroare sexare în ciclu', pair(n(-seF), n(-seM)), f'{num(len(sexerr))} înregistrări · cocoși printre găini', f'<span class="ip-num">{sed}</span>'),
        stat('transfer', 'Transfer · sosire', pair(n(STARTC['F']), n(STARTC['M'])), 'Telița → Hîrbovăț', f'<span class="ip-num">22.02–28.02.2026 · 6 secțiuni</span>'),
        stat('sale', 'Vânzare', DASH, 'nicio vânzare în ciclu', f'<span class="ip-num">din 10.02.2026</span>'))

def tb_pasari():
    grp = [('start', 'La sosire', 'transfer'), ('mortality', 'Mortalitate', 'mortality'), ('culls', 'Sacrificare', 'cull'),
           ('sexing_errors', 'Eroare sexare', 'sexerr'), ('alive', 'În viață', None), ('mortality_pct', 'Mortalitate %', 'mortality')]
    def ic(k): return '' if k is None else (f'<span class="s4t-thico" aria-hidden="true">{icon("sexerr", 14)}</span>' if k == 'sexerr' else thico(k))
    h1 = '<tr><th scope="col" rowspan="2">Secțiune</th>' + ''.join(f'<th scope="colgroup" colspan="2">{ic(k)}{t}</th>' for _, t, k in grp) + '</tr>'
    h2 = '<tr>' + ''.join(f'<th scope="col" class="num">{thsex(sx)}</th>' for _ in grp for sx in 'FM') + '</tr>'
    def cells(get):
        o = ''
        for key, _, _ in grp:
            for sx in 'FM':
                v = get(key, sx)
                if key == 'mortality_pct': o += f'<td class="num {SXC[sx]}">{pct(v, 2)}</td>'
                elif key in ('start', 'alive'): o += f'<td class="num {SXC[sx]}">{n(v)}</td>'
                else: o += f'<td class="num {SXC[sx]}">{n(-v) if v else "0"}</td>'
        return o
    body = ''
    for h in HALLS:
        secs = [s for s in SECTIONS if s.startswith(h + '.')]
        def hget(key, sx, secs=secs):
            if key == 'mortality_pct': return sum(ST[(s, sx)]['mortality'] for s in secs) / sum(ST[(s, sx)]['start'] for s in secs) * 100
            return sum(ST[(s, sx)][key] for s in secs)
        body += f'<tbody><tr class="is-group"><th scope="rowgroup">Hala {h}</th>{cells(hget)}</tr>'
        for s in secs:
            body += f'<tr><th scope="row" class="ip-table__child">{s}</th>{cells(lambda key, sx, s=s: ST[(s, sx)][key])}</tr>'
        body += '</tbody>'
    def fget(key, sx):
        if key == 'mortality_pct': return sum(ST[(s, sx)]['mortality'] for s in SECTIONS) / STARTC[sx] * 100
        return sum(ST[(s, sx)][key] for s in SECTIONS)
    assert (round(fget('mortality_pct', 'F'), 2), round(fget('mortality_pct', 'M'), 2)) == ((5.54, 11.93) if not DROP else tuple(CL['farms'][0]['cycle']['mortality_pct'][x] for x in 'FM'))
    assert (fget('alive', 'F'), fget('alive', 'M')) == (tot['F'], tot['M'])
    body += f'<tbody><tr class="is-total"><th scope="row">Total fermă</th>{cells(fget)}</tr></tbody>'
    t = tbox(f'Păsări pe secțiuni, Hîrbovăț, ciclul Seria 2026, 10.02–16.11.2026', f'<thead>{h1}{h2}</thead>{body}')
    return block('h-psec', 'Pe secțiuni', f'ciclul curent · {num("10.02–16.11.2026")} · SAPT {WK}', t
                 + foot(f'Mortalitate % din sosire. Fără vânzări în ciclu. Cifrele includ corecția din {num("15.11.2026")} (H4.1 {num("14.11")}: {hen(n(AUD["old_values"]["qty"]))} → {s4.hen_v(n(AUD["new_values"]["qty"]))}).'
                        if not DROP else f'Mortalitate % din sosire. Fără vânzări în ciclu. Cifrele nu mai includ rândul șters azi (H4.1 {num("14.11")}: {hen(n(-s4.DROPQ))}).'))

# --- Pe săptămâni (spec D1: alive and mortality per day or week; Pas = Săpt.), from the daily section rows ---
PW = C.defaultdict(lambda: C.Counter())
for r in S['daily']:
    w = PW[r['cycle_week']]
    for k in ('mortality', 'culls', 'sexing_errors'): w[(k, r['sex'])] += r[k]
LASTD = {}
for r in S['daily']: LASTD[r['cycle_week']] = max(LASTD.get(r['cycle_week'], ''), r['date'])
for r in S['daily']:
    if r['date'] == LASTD[r['cycle_week']]: PW[r['cycle_week']][('alive', r['sex'])] += r['alive_end']
assert (PW[WK][('alive', 'F')], PW[WK][('alive', 'M')]) == (tot['F'], tot['M']) and LASTD[WK] == TODAY
assert all(PW[w][('mortality', sx)] == mw[w][sx] for w in mw for sx in 'FM')
def tb_pweeks():
    W = list(range(WK, WK - 6, -1))
    grp = [('alive', 'În viață', None), ('mortality', 'Mortalitate', 'mortality'), ('culls', 'Sacrificare', 'cull'),
           ('sexing_errors', 'Eroare sexare', 'sexerr'), ('cum', 'Mortalitate cumulată', 'mortality')]
    def ic(k): return '' if k is None else (f'<span class="s4t-thico" aria-hidden="true">{icon("sexerr", 14)}</span>' if k == 'sexerr' else thico(k))
    h1 = '<tr><th scope="col" rowspan="2">Săptămâna</th>' + ''.join(f'<th scope="colgroup" colspan="2">{ic(k)}{t}</th>' for _, t, k in grp) + '</tr>'
    h2 = '<tr>' + ''.join(f'<th scope="col" class="num">{thsex(sx)}</th>' for _ in grp for sx in 'FM') + '</tr>'
    body = '<tbody>'
    for w in W:
        d0 = (START + dt.timedelta(days=7 * (w - 1))).isoformat(); d1 = LASTD[w]
        o = ''
        for key, _, _ in grp:
            for sx in 'FM':
                c = SXC[sx]
                if key == 'cum': o += f'<td class="num {c}">{pct((mF if sx == "F" else mM)[w], 2)}</td>'
                elif key == 'alive': o += f'<td class="num {c}">{n(PW[w][(key, sx)])}</td>'
                else: v = PW[w][(key, sx)]; o += f'<td class="num {c}">{n(-v) if v else "0"}</td>'
        body += f'<tr><th scope="row"><span class="ip-num">SAPT {w}</span> <span class="s4t-sub">{ddm(d0)}–{ddate(d1)}</span></th>{o}</tr>'
    body += '</tbody>'
    t = tbox(f'Păsări pe săptămâni, Hîrbovăț, SAPT {W[-1]}–{WK}: în viață la sfârșitul săptămânii, mortalitate, sacrificare, eroare de sexare, mortalitate cumulată', f'<thead>{h1}{h2}</thead>{body}')
    return block('h-pswk', 'Pe săptămâni', f'{num(f"SAPT {W[-1]}–{WK}")} · cele mai noi · pas săptămână', t
                 + foot(f'În viață = la sfârșitul săptămânii. Mortalitate cumulată = morți din sosire ÷ păsări la sosire ({F} {num(n(STARTC["F"]))} · {M} {num(n(STARTC["M"]))}).',
                        ('Vezi toate săptămânile', '#pasari-saptamani')))

# --- event log (SYS09 owner keys: Corectează + Șterge on every record, D70 B: others' entries of today too) ---
EVID = [0]
def acts(name):
    EVID[0] += 1; i = f'ev{EVID[0]}'
    dots = '<svg class="ip-icon" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="5" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="12" cy="19" r="1"/></svg>'
    return (f'<span class="ip-ev__acts"><a class="ip-btn ip-btn--ghost" href="#corecteaza-{i}">Corectează<span class="ip-sr"> {name}</span></a>'
            f'<button type="button" class="ip-btn ip-btn--ghost" aria-haspopup="dialog">Șterge<span class="ip-sr"> {name}</span></button></span>'
            f'<span class="ip-ev__more"><button type="button" class="ip-iconbtn ip-iconbtn--touch" aria-label="Acțiuni, {name}" aria-haspopup="menu" aria-expanded="false" aria-controls="{i}-m">{dots}</button>'
            f'<div class="ip-menu ip-ev__menu" hidden="hidden"><div class="ip-ev__items" id="{i}-m" role="menu" aria-label="Acțiuni, {name}" aria-describedby="{i}-n">'
            f'<a class="ip-menu__item" role="menuitem" tabindex="-1" href="#corecteaza-{i}">Corectează</a><button type="button" class="ip-menu__item" role="menuitem" tabindex="-1">Șterge</button></div>'
            f'<p class="ip-note" id="{i}-n">Corectarea și ștergerea cer un motiv.</p></div></span>')
def evrow(tl, title, meta, name, mod='', slot=True):
    return (f'<li class="ip-ev ip-ev--act">{tl}<span class="ip-ev__main"><span class="ip-ev__title">{title}</span><span class="ip-ev__meta">{meta}</span>{mod}</span>'
            f'{acts(name) if slot else ""}</li>')
SPLIT = f'<span class="ip-tile ip-tile--split" aria-hidden="true">{icon("sexerr")}</span>'
def mortf(f, m): return f'{hen(n(-f)) if f else hen("0")} · {roo(n(-m)) if m else roo("0")}'

LOGD = sorted({e['at'][:10] for e in S['event_log']})
assert LOGD[0] == '2026-11-10' and LOGD[-1] == TODAY
AUD = S['audit_log'][0]
assert AUD['row'] == 'mortality · H4.1 · F · 2026-11-14' and AUD['old_values']['qty'] == -3 and AUD['new_values']['qty'] == -5
# D82: one Jurnal row per section-day record (deaths ♀ · ♂ of one section on one day; culls and sexing errors are their own
# records), newest day first, sections in farm order inside a day; paged by row count. A section-day with no deaths has no row.
PAGE = 15
SD = C.defaultdict(lambda: {'F': 0, 'M': 0})
for r in S['daily']:
    if r['mortality']: SD[(r['date'], r['section'])][r['sex']] += r['mortality']
assert SD[('2026-11-14', 'H4.1')]['F'] == -AUD['new_values']['qty'] if not DROP else (DROP == ('2026-11-14', 'H4.1', 'F') and ('2026-11-14', 'H4.1') not in SD)
SAVE = {e['at'][:10]: e for e in S['event_log'] if e['record'] == 'mortality' and e['action'] == 'create'}
for d, e in SAVE.items():  # the Zilnic save total = its section-day deaths + that day's culls (− the 15.11 correction on 14.11)
    f = sum(v['F'] for (dd, _), v in SD.items() if dd == d); m = sum(v['M'] for (dd, _), v in SD.items() if dd == d)
    cf = sum(c['qty'] for c in culls if c['date'] == d and c['sex'] == 'F'); cm = sum(c['qty'] for c in culls if c['date'] == d and c['sex'] == 'M')
    fix = (2 if not DROP else 2 - s4.DROPQ) if d == '2026-11-14' else 0  # deleted (S16P): the save's 3 are gone too
    sf, sm = re.search(r'♀ (−?\d+) · ♂ (−?\d+)', e['summary']).groups()
    assert (int(sf.replace('−', '-')), int(sm.replace('−', '-'))) == (-(f + cf - fix), -(m + cm)), (d, e['summary'])
def xfer_days():
    byd = C.defaultdict(list)
    for a in TR['arrivals']: byd[a['date']].append(a)
    return byd
def log_records():  # (date, section order, kind order, payload), every bird record of the cycle in the Jurnal's order
    recs = [(d, SECTIONS.index(sec), 0, ('mortality', d, sec)) for (d, sec) in SD]
    recs += [(c['date'], SECTIONS.index(c['section']), 1, ('cull', c)) for c in culls]
    recs += [(e['date'], SECTIONS.index(e['section']), 2, ('sexerr', e)) for e in sexerr]
    recs += [(d, 0, 3, ('transfer', d)) for d in xfer_days()]
    recs.sort(key=lambda x: (-int(x[0].replace('-', '')), x[1], x[2]))
    return recs
def log_rows():
    recs = log_records(); out = []
    for d, _, _, p in recs[:PAGE]:
        if p[0] == 'mortality':
            sec = p[2]; f, m = SD[(d, sec)]['F'], SD[(d, sec)]['M']; t = SAVE[d]['at'][11:]
            mod = ''
            if (d, sec) == ('2026-11-14', 'H4.1'):
                mod = (f'<span class="ip-ev__mod"><span class="ip-tag ip-tag--neutral">corectat</span><span>{hen(n(AUD["old_values"]["qty"]))} → {s4.hen_v(n(AUD["new_values"]["qty"]))}</span>'
                       f'<a class="ip-link" href="#istoric-H4.1-14.11">modificat de {AUD["user"]} · <span class="ip-num">{ddate(AUD["at"][:10])}, {AUD["at"][11:]}</span><span class="ip-sr">, vezi istoricul</span></a></span>')
            out.append(evrow(tile('mortality'), f'Mortalitate · {sec} · {mortf(f, m)}', f'<span class="ip-num">{ddate(d)}, {t} · SAPT {cw(d)}</span> · {SAVE[d]["user"]}',
                             f'Mortalitate {sec}, {ddate(d)}', mod))
        elif p[0] == 'cull':
            c = p[1]; at = next(e['at'] for e in S['event_log'] if e['record'] == 'cull' and e['at'][:10] == c['date'])
            fig = hen(n(-c['qty'])) if c['sex'] == 'F' else roo(n(-c['qty']))
            out.append(evrow(tile('cull'), f'Sacrificare · {c["section"]} · {fig} · {c["reason"]}', f'<span class="ip-num">{ddate(c["date"])}, {at[11:]} · SAPT {cw(c["date"])}</span> · {c["created_by"]}',
                             f'Sacrificare {c["section"]}, {ddate(c["date"])}'))
        else: raise AssertionError(p)  # page 1 holds only Zilnic records (14–16.11)
    assert len(out) == PAGE and ('#istoric-H4.1-14.11' in ''.join(out)) == (not DROP)
    return out, recs
def moves_rows():
    out = []
    for e in sorted(sexerr, key=lambda e: e['date'], reverse=True):
        out.append(evrow(SPLIT, f'Eroare sexare · {e["section"]} · {hen(n(-e["qty"]))}', f'<span class="ip-num">{ddate(e["date"])} · SAPT {cw(e["date"])}</span> · {e["created_by"]} · cocoș printre găini',
                         f'Eroare sexare {e["section"]}, {ddate(e["date"])}'))
    byd = xfer_days()
    for d in sorted(byd, reverse=True):
        rs = byd[d]; sx = {a['sex'] for a in rs}; assert len(sx) == 1; sx = sx.pop()
        q = sum(a['qty'] for a in rs); ss = sorted({a['section'] for a in rs}); secs = '6 secțiuni' if ss == SECTIONS else ', '.join(ss)
        fig = hen('+' + n(q)) if sx == 'F' else roo('+' + n(q))
        out.append(evrow(tile('transfer'), f'Transfer · Telița → {secs} · {fig}', f'<span class="ip-num">{ddate(d)} · SAPT {cw(d)}</span> · importat din registru',
                         f'Transfer Telița → Hîrbovăț, {ddate(d)}'))
    assert sum(a['qty'] for a in TR['arrivals'] if a['sex'] == 'F') == STARTC['F']
    cm = '<span class="ip-cmark ip-cmark--start xfer"><span class="ip-cmark__g" aria-hidden="true"><i class="ip-cmark__bar"></i><i class="ip-cmark__blk"></i></span>ciclu nou prin transfer</span>'
    ctile = '<span class="ip-tile ip-tile--cycle" aria-hidden="true"><span class="ip-cmark ip-cmark--start xfer"><span class="ip-cmark__g"><i class="ip-cmark__bar"></i><i class="ip-cmark__blk"></i></span></span></span>'
    out.append(evrow(ctile, 'Ciclu nou · Seria 2026 · Hîrbovăț · ouat', f'<span class="ip-num">10.02.2026 · SAPT 1</span> · ghid Ross 308 ouat · {cm}', '', slot=False))
    return out

def panel_pasari():
    (lg, recs), mv = log_rows(), moves_rows()
    pages = -(-len(recs) // PAGE); last = recs[PAGE - 1][0]
    left = block('h-jurnal', 'Jurnal păsări', 'un rând pe secțiune și zi · cele mai noi',
                 f'<ul class="ip-list ip-list--act" aria-label="Jurnal păsări Hîrbovăț, pagina 1 din {pages}">{"".join(lg)}</ul>'
                 + foot(f'<span class="ip-num">1–{PAGE} din {n(len(recs))} · {ddm(last)}–{ddate(TODAY)} · pagina 1 din {pages}</span> · secțiile fără morți în ziua respectivă nu au rând',
                        ('Mai vechi', '#jurnal-pagina-2')))
    right = block('h-misc', 'Transferuri, erori de sexare, cicluri', 'tot ciclul · fără vânzări',
                  f'<ul class="ip-list ip-list--act" aria-label="Transferuri, erori de sexare și cicluri, Hîrbovăț, ciclul curent">{"".join(mv)}</ul>'
                  + foot('Transferurile sunt din registrul Seria 2026; ciclul pornește pe 10.02.2026.'))
    return (top('h-key', 'Păsări, mortalitate', ro_pasari(), st_pasari()) + '\n' + tb_pasari() + '\n' + tb_pweeks() + '\n'
            + f'<div class="s4t-split">{left}{right}</div>')

# ================================================================= 2. HRANĂ
FT = {'ft-breeder2': 'Breeder 2', 'ft-cocosi': 'Cocoși'}
ftw = C.defaultdict(lambda: C.defaultdict(list))
# D82/D46: today's feed (16.11) is still being entered (10 of 12 rows), so the week means and the guide comparison leave it out.
for r in S['feed_by_type_daily']:
    if r['date'] < TODAY: ftw[r['feed_type_id']][cw(r['date'])].append(r['consumption_kg'])
FEED_WK = sorted({r['date'] for r in S['daily'] if r['cycle_week'] == WK and r['date'] < TODAY})
assert len(FEED_WK) == 6 and FEED_WK[0] == '2026-11-10' and FEED_WK[-1] == '2026-11-15' and len(ftw['ft-breeder2'][WK]) == 6
kF = {w: sum(v) / len(v) for w, v in ftw['ft-breeder2'].items()}; kM = {w: sum(v) / len(v) for w, v in ftw['ft-cocosi'].items()}
act40 = C.Counter(); gd40 = C.Counter()
for r in S['daily']:
    if r['cycle_week'] == WK and r['date'] < TODAY and r['g_per_bird'] is not None:
        act40[r['sex']] += r['g_per_bird'] * r['alive_end'] / 1000; gd40[r['sex']] += r['guide_feed_g'] * r['alive_end'] / 1000
devF = (act40['F'] / gd40['F'] - 1) * 100; devM = (act40['M'] / gd40['M'] - 1) * 100
feed_ok = abs(devF) <= 10 and abs(devM) <= 10
AL = s4.alloc
kg_today = sum(a['kg'] for a in AL.values() if a['today_g'] is not None)
rows_today = sum(1 for a in AL.values() if a['today_g'] is not None)
assert round(kg_today) == 3691 and rows_today == 10
kg_type = {ft: sum(a['kg'] for a in AL.values() if a['today_g'] is not None and a['feed_type_id'] == ft) for ft in FT}
assert rnd(kg_type['ft-breeder2'], 1) + rnd(kg_type['ft-cocosi'], 1) == rnd(kg_today, 1) == 3691.3  # parts at 0,1 kg add up to the total
cons = {ft: sum(r['consumption_kg'] for r in S['feed_by_type_daily'] if r['feed_type_id'] == ft) for ft in FT}
dels = [d for d in S['deliveries'] if d['farm_id'] == 'hirbovat']
del_kg = sum(d['qty_kg'] for d in dels)
unp = [d for d in dels if not d['priced']]
assert len(unp) == 4 and sum(d['qty_kg'] for d in unp) == s4.unpriced_kg == 76107
sb2, sco = s4.stocks['ft-breeder2'], s4.stocks['ft-cocosi']
cons_tot = sum(rnd(v) for v in cons.values())  # the total is the sum of the shown parts (D82)
assert cons_tot == rnd(sum(cons.values())) == 1189745

def ro_hrana():
    rows = [[str(w), DASH + ' fără date', DASH + ' fără date'] for w in (1, 2)] + [[str(w), n(kF[w]) + ' kg', n(kM[w]) + ' kg'] for w in sorted(kF)]
    assert max(kF.values()) < 6000
    lbl = f'Consum de hrană pe zi, pe tip de hrană, Hîrbovăț Seria 2026, SAPT 3–{WK}: Breeder 2 pentru găini, Cocoși pentru cocoși'
    ch = gchart('kgzi', lbl, WK, [{'cls': 'act', 'pts': kF}, {'cls': 's2', 'pt': 'pt2', 'pts': kM}], 0, 6000, [0, 2000, 4000, 6000], lambda t: n(t), x0=44)
    return readout('kgzi', f'HRANĂ KG/ZI · BREEDER 2 ♀ · COCOȘI ♂ · SAPT {WK}', 'ok' if feed_ok else 'warn', rpair(n(kF[WK]), n(kM[WK]), 'kg/zi'),
                   f'față de ghid ♀ {sgn(devF)}% · ♂ {sgn(devM)}% · {ddm(FEED_WK[0])}–{ddate(FEED_WK[-1])} · fără {ddm(TODAY)}',
                   ch, f'SAPT {WK} · ♀ {n(kF[WK])} · ♂ {n(kM[WK])} kg/zi · ghid ♀ {sgn(devF)}% · ♂ {sgn(devM)}%',
                   ['SAPT', 'Breeder 2 ♀, kg/zi', 'Cocoși ♂, kg/zi'], rows, f'Consum mediu pe zi al săptămânii, pe tip de hrană; înregistrări din 01.03.2026 (SAPT 3); SAPT {WK} = {ddm(FEED_WK[0])}–{ddate(FEED_WK[-1])}, fără {ddate(TODAY)} (incomplet)',
                   legend([('act', f'{HB} Breeder 2'), ('s2', f'{RB} Cocoși')], f'SAPT {WK} fără {ddm(TODAY)}, incomplet'))

def st_hrana():
    return stats(
        stat('feed', 'Hrană azi pe pasăre', f'{pair(s4.g_f, s4.g_m)}<small>g</small>', f'ghid {pair(s4.gf_g, s4.gm_g)} {lamp("ok", "în normă")}', f'<span>fără {F} H4.2 · {M} H3.2 (lipsă azi)</span>'),
        stat('feed', 'Hrană azi', f'{n(kg_today, 1)}<small>kg</small>', f'{num(rows_today)} din {num(12)} rânduri {lamp("warn", "lipsă")}',
             pair(n(kg_type["ft-breeder2"], 1), n(kg_type["ft-cocosi"], 1))),
        stat('delivery', f'Stoc Breeder 2 {F}', f'{n(sb2["days_left"], 1)}<small>zile</small>', f'<span class="ip-num">până pe {ddate(sb2["run_out_date"])}</span>{lamp("ok", "în normă")}',
             f'<span class="ip-num">{n(sb2["stock_kg"])} kg · {n(sb2["avg_daily_kg_last3"])} kg/zi</span>'),
        stat('delivery', f'Stoc Cocoși {M}', f'{n(sco["days_left"], 1)}<small>zile</small>', f'<span class="ip-num">până pe {ddate(sco["run_out_date"])}</span>{lamp("warn", "abatere")}',
             f'<span class="ip-num">{n(sco["stock_kg"])} kg · {n(sco["avg_daily_kg_last3"])} kg/zi</span>'),
        stat('feed', 'Consum în ciclu', f'{n(cons_tot)}<small>kg</small>', pair(n(cons["ft-breeder2"]), n(cons["ft-cocosi"])),
             f'<span class="ip-num">din 01.03 · {ddm(TODAY)}</span> incomplet'),
        stat('delivery', 'Livrări în ciclu', f'{n(del_kg)}<small>kg</small>', f'{num(len(dels))} livrări · {num(len(unp))} fără preț {lamp("info", "informație")}',
             f'<span class="ip-num">{n(s4.unpriced_kg)} kg · {ddm(unp[0]["date"])}–{ddate(unp[-1]["date"])}</span>'))

def tb_hrana():
    ths = (f'<tr><th scope="col">Secțiune</th><th scope="col">{thico("feed")}Tip hrană</th><th scope="col" class="num">În viață</th>'
           '<th scope="col" class="num">Ieri g</th><th scope="col" class="num">Azi g/pasăre</th><th scope="col" class="num">Ghid g</th>'
           '<th scope="col" class="num">Față de ghid</th><th scope="col" class="num">Kg</th><th scope="col">Stare</th></tr>')
    body = ''
    for h in HALLS:
        secs = [s for s in SECTIONS if s.startswith(h + '.')]
        hk = sum(AL[(s, sx)]['kg'] for s in secs for sx in 'FM' if AL[(s, sx)]['today_g'] is not None)
        hr = sum(1 for s in secs for sx in 'FM' if AL[(s, sx)]['today_g'] is not None)
        body += f'<tbody><tr class="is-group"><th scope="rowgroup">Hala {h}</th><td colspan="6" class="s4t-gnote">{num(hr)} din {num(4)} rânduri</td><td class="num">{n(hk, 1)}</td><td></td></tr>'
        for s in secs:
            for sx in 'FM':
                a = AL[(s, sx)]; c = SXC[sx]
                if a['today_g'] is None:
                    g, dv, kg, stt = f'<span class="s4t-muted">{DASH} lipsă</span>', f'<span class="s4t-muted">{DASH}</span>', f'<span class="s4t-muted">{DASH}</span>', lamp('warn', 'lipsă')
                else:
                    g = a['today_g']; dv = sgn(a['deviation_pct']) + '%'; kg = n(a['kg'], 1)
                    stt = lamp('warn', 'abatere') if abs(a['deviation_pct']) > 10 else lamp('ok', 'în normă')
                assert a['guide_g'] == s4.guide[(WK, sx)]['feed_g']
                body += (f'<tr><th scope="row" class="ip-table__child">{s} {sg(sx)}<span class="ip-sr"> {"găini" if sx == "F" else "cocoși"}</span></th><td>{FT[a["feed_type_id"]]}</td>'
                         f'<td class="num {c}">{n(a["alive"])}</td><td class="num s4t-muted">{a["yesterday_g"]}</td><td class="num {c}">{g}</td>'
                         f'<td class="num s4t-muted">{a["guide_g"]}</td><td class="num">{dv}</td><td class="num">{kg}</td><td>{stt}</td></tr>')
        body += '</tbody>'
    body += (f'<tbody><tr class="is-total"><th scope="row">Total fermă</th><td colspan="6"><span class="ip-note">{num(rows_today)} din {num(12)} rânduri · '
             f'{num(n(kg_type["ft-breeder2"], 1) + " kg")} Breeder 2 · {num(n(kg_type["ft-cocosi"], 1) + " kg")} Cocoși</span></td><td class="num">{n(kg_today, 1)}</td><td>{lamp("warn", "lipsă 2")}</td></tr></tbody>')
    t = tbox('Hrană azi pe secțiuni și sex, Hîrbovăț, 16.11.2026, SAPT 40, față de ghid', f'<thead>{ths}</thead>{body}')
    return block('h-hazi', 'Azi pe secțiuni', f'{num("16.11.2026 · SAPT 40")} · ghid Ross 308 ouat · toleranță {num("±10%")}', t
                 + foot(f'Kg = g/pasăre × păsări în viață. Un rând lipsă nu se completează din ziua anterioară; „Ieri” e sugestia de confirmat.', ('Deschide Hrană zilnică', '#hrana-zilnica')))

def panel_hrana():
    return top('h-key', 'Hrană față de ghid', ro_hrana(), st_hrana()) + '\n' + tb_hrana()

# ================================================================= 3. GREUTATE
alive_on = {(r['date'], r['section'], r['sex']): r['alive_end'] for r in S['daily']}
WG = C.defaultdict(list)
for r in S['weighings']: WG[r['cycle_week']].append(r)
assert all(r['guide_weight_g'] is None and r['guide_state'] == 'fără date în ghid' for r in S['weighings'])
assert all(r['body_weight_g'] is None for r in GL['rows'])  # laying guide has no body weight (D29)
def wfarm(w, sx):
    rs = [r for r in WG[w] if r['sex'] == sx]
    return sum(r['avg_weight_g'] * alive_on[(r['date'], r['section'], sx)] for r in rs) / sum(alive_on[(r['date'], r['section'], sx)] for r in rs)
def sfarm(w, sx, key):
    rs = [r for r in WG[w] if r['sex'] == sx]
    return sum(r[key] * r['sample_size'] for r in rs) / sum(r['sample_size'] for r in rs)
WKS = sorted(WG); assert WKS[0] == 3 and WKS[-1] == WK and len(WKS) == 38
wF = {w: wfarm(w, 'F') for w in WKS}; wM = {w: wfarm(w, 'M') for w in WKS}
assert (rnd(wF[WK]), rnd(wM[WK])) == (s4.w_f, s4.w_m)  # S4: ♀ 3 959 · ♂ 4 777 g
dF = rnd(wF[WK]) - rnd(wF[WK - 1]); dM = rnd(wM[WK]) - rnd(wM[WK - 1])  # shown − shown
W40 = {(r['section'], r['sex']): r for r in WG[WK]}; W39 = {(r['section'], r['sex']): r for r in WG[WK - 1]}
WAT = next(e['at'] for e in S['event_log'] if e['record'] == 'weighing')

def ro_greutate():
    rows = [[str(w), DASH + ' fără cântărire', DASH + ' fără cântărire', DASH + ' fără date în ghid'] for w in (1, 2)]
    rows += [[str(w), n(wF[w]) + ' g', n(wM[w]) + ' g', DASH + ' fără date în ghid'] for w in WKS]
    assert 2000 < min(wF.values()) and max(wM.values()) < 5000
    lbl = f'Greutate medie găini și cocoși, Hîrbovăț Seria 2026, SAPT 3–{WK}; ghidul de ouat nu are greutate'
    ch = gchart('wt', lbl, WK, [{'cls': 'act', 'pts': wF}, {'cls': 's2', 'pt': 'pt2', 'pts': wM}], 2000, 5000, [2000, 3000, 4000, 5000], lambda t: n(t), x0=44)
    return readout('wt', f'GREUTATE · HÎRBOVĂȚ · SAPT {WK}', 'off', rpair(n(wF[WK]), n(wM[WK]), 'g'),
                   f'cântărit {ddate(TODAY)} · față de SAPT {WK - 1}: ♀{NBSP}{sgn(dF, 0)} · ♂{NBSP}{sgn(dM, 0)}{NBSP}g',  # S4TP fix 1: sign, figure, unit never split
                   ch, f'SAPT {WK} · ♀ {n(wF[WK])} g · ♂ {n(wM[WK])} g · ghid {DASH} fără date în ghid',
                   ['SAPT', 'Găini ♀', 'Cocoși ♂', 'Ghid'], rows, 'Greutate medie pe săptămânile ciclului (medie pe secțiuni după păsări în viață); cântăriri din 02.03.2026 (SAPT 3)',
                   legend([('act', f'{HB} găini'), ('s2', f'{RB} cocoși')], 'medie după păsări în viață'))

def st_greutate():
    uF, uM = sfarm(WK, 'F', 'uniformity_pct'), sfarm(WK, 'M', 'uniformity_pct')
    cF, cM = sfarm(WK, 'F', 'cv_pct'), sfarm(WK, 'M', 'cv_pct')
    nF = sum(r['sample_size'] for r in WG[WK] if r['sex'] == 'F'); nM = sum(r['sample_size'] for r in WG[WK] if r['sex'] == 'M')
    return stats(
        stat('weighing', f'Greutate {F} / {M}', f'{pair(n(wF[WK]), n(wM[WK]))}<small>g</small>', NOGD, f'<span class="ip-num">cântărit {ddate(TODAY)}, {WAT[11:]}</span>'),
        stat('weighing', f'Față de SAPT {WK - 1}', f'{pair(sgn(dF, 0), sgn(dM, 0))}<small>g</small>', NOGD, f'<span class="ip-num">SAPT {WK - 1}: ♀ {n(wF[WK - 1])} · ♂ {n(wM[WK - 1])} g</span>'),
        stat('weighing', 'Uniformitate', f'{pair(n(uF, 1), n(uM, 1))}<small>%</small>', NOGD, f'păsări în {num("±10%")} din medie'),
        stat('weighing', 'Coeficient de variație', f'{pair(n(cF, 1), n(cM, 1))}<small>%</small>', NOGD, 'medie pe secțiuni după probă'),
        stat('weighing', 'Probă cântărită', pair(n(nF), n(nM)), f'{num(6)} secțiuni · {num(12)} rânduri', f'din păsări {pair(pct(nF / tot["F"] * 100, 1), pct(nM / tot["M"] * 100, 1))}'),  # S4TP fix 2: "din păsări ♀ 1,7% · ♂ 12,0%" (share of birds alive)
        stat('weighing', 'Cântăriri în ciclu', f'{len(WKS)}<small>săptămâni</small>', f'<span class="ip-num">SAPT 3–{WK} · {ddm(min(r["date"] for r in S["weighings"]))}–{ddate(TODAY)}</span>', 'una pe săptămână, toate secțiunile'))

SPL, SPH = 2600, 5000
def spark(sec, sx):
    pts = {w: next(r['avg_weight_g'] for r in WG[w] if r['section'] == sec and r['sex'] == sx) for w in WKS}
    assert SPL < min(pts.values()) and max(pts.values()) < SPH
    def X(w): return round(4 + (w - 1) * 152 / (WK - 1), 1)
    def Y(v): return round(44 - (v - SPL) / (SPH - SPL) * 40, 1)
    cls = 'ip-spark--hen' if sx == 'F' else 'ip-spark--rooster'
    lbl = f'Greutate {sec} {"găini" if sx == "F" else "cocoși"}, SAPT 3–{WK}: de la {n(pts[3])} g la {n(pts[WK])} g; fără ghid'
    return (f'<svg class="ip-spark {cls}" viewBox="0 0 160 48" role="img" aria-label="{lbl}"><polyline class="act" points="'
            + ' '.join(f'{X(w)},{Y(v)}' for w, v in sorted(pts.items())) + f'"/><circle class="pt" cx="{X(WK)}" cy="{Y(pts[WK])}" r="2"/></svg>')
def tb_greutate():
    ths = (f'<tr><th scope="col">Secțiune</th><th scope="col" class="num">{thico("weighing")}Greutate g</th><th scope="col" class="num">Față de SAPT {WK - 1}</th>'
           '<th scope="col" class="num">Uniformitate</th><th scope="col" class="num">CV</th><th scope="col" class="num">Probă</th><th scope="col">Curba · SAPT 3–40</th></tr>')
    body = ''
    for h in HALLS:
        body += f'<tbody><tr class="is-group"><th scope="rowgroup" colspan="7">Hala {h}</th></tr>'
        for s in [x for x in SECTIONS if x.startswith(h + '.')]:
            for sx in 'FM':
                r = W40[(s, sx)]; d = r['avg_weight_g'] - W39[(s, sx)]['avg_weight_g']; c = SXC[sx]
                body += (f'<tr><th scope="row" class="ip-table__child">{s} {sg(sx)}<span class="ip-sr"> {"găini" if sx == "F" else "cocoși"}</span></th><td class="num {c}">{n(r["avg_weight_g"])}</td>'
                         f'<td class="num">{sgn(d, 0)}</td><td class="num">{pct(r["uniformity_pct"])}</td><td class="num">{pct(r["cv_pct"])}</td><td class="num">{r["sample_size"]}</td>'
                         f'<td class="ip-sparktable__curve">{spark(s, sx)}</td></tr>')
        body += '</tbody>'
    t = tbox(f'Greutate pe secțiuni și sex, Hîrbovăț, cântărirea din 16.11.2026, SAPT 40', f'<thead>{ths}</thead>{body}', 'ip-sparktable')
    return block('h-wsec', 'Pe secțiuni', f'cântărirea din {num("16.11.2026 · SAPT 40")} · {num("12")} rânduri', t
                 + foot(f'Ghid: {DASH} fără date în ghid. Ghidul Ross 308 ouat nu are greutate, uniformitate sau CV. Curbele au aceeași scară ({num(n(2600) + "–" + n(5000) + " g")}).'))

def panel_greutate():
    return top('h-key', 'Greutate', ro_greutate(), st_greutate()) + '\n' + tb_greutate()

# ================================================================= 4. PRODUCȚIE
EW = {(r['hall'], r['cycle_week']): r for r in S['eggs_weekly']}
lay_all = s4.lay_all; gv = s4.gv
cat = C.defaultdict(lambda: C.Counter())
for r in S['eggs_daily']:
    if r['cycle_week'] == WK:
        for k in ('cat1', 'cat2', 'cat3', 'hatching_eggs', 'total_eggs', 'floor_eggs'): cat[r['hall']][k] += r[k]
for h in HALLS:
    assert cat[h]['cat1'] + cat[h]['cat2'] + cat[h]['cat3'] + cat[h]['hatching_eggs'] == cat[h]['total_eggs'] == EW[(h, WK)]['total_eggs']  # D18
    assert cat[h]['floor_eggs'] == EW[(h, WK)]['floor_eggs']
for k in ('cat1', 'cat2', 'cat3', 'hatching_eggs', 'total_eggs', 'floor_eggs'): cat['all'][k] = sum(cat[h][k] for h in HALLS)
assert cat['all']['total_eggs'] == EW[('all', WK)]['total_eggs'] == 130921
eggs_cycle = sum(r['total_eggs'] for r in S['eggs_daily'])
first_egg = min(r['date'] for r in S['eggs_daily'] if r['total_eggs'] > 0)
hall_alive = {h: sum(s4.alive[(s, 'F')] for s in SECTIONS if s.startswith(h + '.')) for h in HALLS}

def ro_productie():
    aw = sorted(w for w in lay_all if w <= WK); a40 = lay_all[WK]['lay_pct']; dev = rnd(a40 - gv[WK], 1)
    band = ({w: gv[w] - 3 for w in gv}, {w: min(gv[w] + 3, 100) for w in gv})
    lbl = f'Ouat Hîrbovăț față de ghid, pe săptămânile ciclului, SAPT 3–{WK}; ghid SAPT 7–45'
    ch = gchart('lay', lbl, 45, [{'cls': 'act', 'pts': {w: lay_all[w]['lay_pct'] for w in aw}}], 0, 100, [0, 50, 100], lambda t: f'{t}%',
                guide=gv, band=band, xticks=(1, 10, 20, 30, 40, 45))
    rows = [[str(w), pct(lay_all[w]['lay_pct']), pct(gv[w]) if w in gv else DASH + ' fără date în ghid'] for w in aw]
    return readout('lay', f'OUAT · HÎRBOVĂȚ · SAPT {WK}', 'ok' if abs(dev) <= 3 else 'warn', f'{n(a40, 1)}<small>%</small>',
                   f'ghid {pct(gv[WK])} · {sgn(dev)} pct · 10.11–16.11.2026', ch, f'SAPT {WK} · efectiv {pct(a40)} · ghid {pct(gv[WK])} · {sgn(dev)} pct',
                   ['SAPT', 'Efectiv', 'Ghid'], rows, 'Ouat Hîrbovăț pe săptămâni, efectiv față de ghid; înregistrări din 01.03.2026 (SAPT 3)',
                   legend([('act', 'efectiv'), ('guide', 'ghid'), ('band', '±3 pct')], 'axa: săpt. ciclului'))

def st_productie():
    a = EW[('all', WK)]; c = cat['all']
    return stats(
        stat('eggs', f'Ouă · SAPT {WK}', n(a['total_eggs']), f'<span class="ip-num">{n(sum(e["total_eggs"] for e in s4.eggs_today.values()))} azi · 3 hale</span>', 'Cat. 1–3 + incubabile'),
        stat('eggs', f'Incubabile · SAPT {WK}', f'{n(a["hatching_pct"], 1)}<small>%</small>', NOGD, f'<span class="ip-num">{n(a["hatching_eggs"])} din {n(a["total_eggs"])} ouă</span>'),
        stat('eggs', f'Masă medie ou · SAPT {WK}', f'{n(a["avg_egg_weight_g"], 1)}<small>g</small>', NOGD, f'<span class="ip-num">SAPT {WK - 1}: {n(EW[("all", WK - 1)]["avg_egg_weight_g"], 1)} g</span>'),
        stat('eggs', f'Categorii · SAPT {WK}', n(c['cat1'] + c['cat2'] + c['cat3']), f'Cat. 1 · 2 · 3: <span class="ip-num">{n(c["cat1"])} · {n(c["cat2"])} · {n(c["cat3"])}</span>', 'neincubabile, intră în total'),
        stat('eggs', f'Ouă jos · SAPT {WK}', n(a['floor_eggs']), 'nu intră în total', f'<span class="ip-num">{pct(a["floor_eggs"] / a["total_eggs"] * 100)}</span> față de total'),
        stat('eggs', 'Ouă în ciclu', n(eggs_cycle), f'<span class="ip-num">primul ou {ddate(first_egg)} · SAPT {cw(first_egg)}</span>', f'<span class="ip-num">01.03–16.11.2026</span>'))

def lay_state(h, w):
    r = EW[(h, w)]
    if r['guide_lay_pct'] is None: return f'<span class="s4t-muted">{DASH} fără date în ghid</span>'
    return lamp('warn', 'abatere') if abs(r['diff_pts']) > 3 else lamp('ok', 'în normă')
def tb_halls():
    ths = (f'<tr><th scope="col">Hala</th><th scope="col" class="num">Găini</th><th scope="col" class="num">{thico("eggs")}Ouă</th><th scope="col" class="num">Ouat</th>'
           '<th scope="col" class="num">Față de ghid</th><th scope="col">Stare</th><th scope="col" class="num">Incubabile</th>'
           '<th scope="col" class="num">Cat. 1</th><th scope="col" class="num">Cat. 2</th><th scope="col" class="num">Cat. 3</th><th scope="col" class="num">Jos</th><th scope="col" class="num">Masă ou</th></tr>')
    def row(h, th):
        r = EW[(h, WK)]; c = cat[h]
        al = sum(hall_alive.values()) if h == 'all' else hall_alive[h]
        stt = lay_state(h, WK)
        if h == 'H3': stt = f'<span class="s4t-state">{stt}<span class="s4t-muted ip-num">din 09.11</span></span>'
        return (f'{th}<td class="num s4t-hen">{n(al)}</td><td class="num">{n(r["total_eggs"])}</td><td class="num">{pct(r["lay_pct"])}</td>'
                f'<td class="num">{sgn(r["diff_pts"])} pct</td><td>{stt}</td><td class="num">{pct(r["hatching_pct"])}</td><td class="num">{n(c["cat1"])}</td><td class="num">{n(c["cat2"])}</td>'
                f'<td class="num">{n(c["cat3"])}</td><td class="num s4t-muted">{n(r["floor_eggs"])}</td><td class="num">{n(r["avg_egg_weight_g"], 1)} g</td></tr>')
    body = '<tbody>' + ''.join(row(h, f'<tr><th scope="row">{h}</th>') for h in HALLS) + '</tbody>'
    body += '<tbody>' + row('all', '<tr class="is-total"><th scope="row">Total fermă</th>') + '</tbody>'
    assert s4.h3_since == '2026-11-09' and sum(hall_alive.values()) == tot['F']
    assert len({EW[(h, WK)]['guide_lay_pct'] for h in HALLS + ['all']}) == 1  # one guide value for all halls → in the head
    t = tbox(f'Producție pe hale, Hîrbovăț, SAPT 40, 10.11–16.11.2026, față de ghid', f'<thead>{ths}</thead>{body}')
    return block('h-phal', 'Pe hale', f'{num("SAPT 40 · 10.11–16.11.2026")} · ghid Ross 308 ouat {num(pct(EW[("all", WK)]["guide_lay_pct"]))} · toleranță {num("±3 pct")}', t
                 + foot(f'Găini = în viață azi. Ouă = Cat. 1–3 + incubabile; ouăle jos nu intră în total. Ouat = ouă ÷ găini la începutul zilei. Incubabile și masa oului: {NOGD}.'))
def tb_weeks():
    W = list(range(WK, WK - 6, -1))
    ths = ('<tr><th scope="col">Săptămâna</th><th scope="col" class="num">Ouă</th><th scope="col" class="num">Ouat fermă</th>'
           + ''.join(f'<th scope="col" class="num">Ouat {h}</th>' for h in HALLS)
           + '<th scope="col" class="num">Ghid</th><th scope="col">Stare</th><th scope="col" class="num">Incubabile</th><th scope="col" class="num">Masă ou</th></tr>')
    body = '<tbody>'
    for w in W:
        a = EW[('all', w)]
        hc = ''.join(f'<td class="num">{pct(EW[(h, w)]["lay_pct"])}</td>' for h in HALLS)
        dev = [h for h in HALLS if abs(EW[(h, w)]['diff_pts']) > 3]
        stt = lay_state('all', w) if not dev else f'<span class="s4t-state">{lamp("warn", "abatere")}<span class="s4t-muted">{", ".join(dev)}</span></span>'
        body += (f'<tr><th scope="row"><span class="ip-num">SAPT {w}</span> <span class="s4t-sub">{ddm(a["from"])}–{ddate(a["to"])}</span></th><td class="num">{n(a["total_eggs"])}</td><td class="num">{pct(a["lay_pct"])}</td>{hc}'
                 f'<td class="num s4t-muted">{pct(a["guide_lay_pct"])}</td><td>{stt}</td><td class="num">{pct(a["hatching_pct"])}</td><td class="num">{n(a["avg_egg_weight_g"], 1)} g</td></tr>')
    body += '</tbody>'
    t = tbox(f'Producție pe săptămâni, Hîrbovăț, SAPT {W[-1]}–{WK}, ouat pe fermă și pe hale față de ghid', f'<thead>{ths}</thead>{body}')
    return block('h-pwk', 'Pe săptămâni', f'{num(f"SAPT {W[-1]}–{WK}")} · cele mai noi', t
                 + foot(f'Stare = ouatul fermei și al fiecărei hale față de ghid ({num("±3 pct")}); se numește hala în abatere.', ('Vezi toate săptămânile', '#productie-saptamani')))

def panel_productie():
    return top('h-key', 'Producție față de ghid', ro_productie(), st_productie()) + '\n' + tb_halls() + '\n' + tb_weeks()

# ================================================================= CSS (screen only, tokens only) + write
CSS = s4.CSS + '\n' + s4.CSS_S4 + '''
.s4t-split{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:var(--space-4);align-items:start}
.s4-panel .ip-table .ip-lamp{white-space:nowrap}
.s4t-thico{display:inline-flex;vertical-align:top;margin-right:var(--space-1);color:var(--c)}
.s4t-hen{color:var(--rec-hen)}
.s4t-roo{color:var(--rec-rooster)}
.s4t-muted{color:var(--ink-muted)}
.s4t-gnote{color:var(--ink-secondary)}
.s4t-sub{color:var(--ink-secondary);font-family:var(--font-mono);font-size:var(--ip-fs-small)}
.s4t-state{display:inline-flex;align-items:center;gap:var(--space-2)}
.s4t-foot{display:flex;justify-content:space-between;align-items:center;gap:var(--space-4)}
.s4t-lk{display:inline-flex;align-items:center;min-height:var(--space-6);white-space:nowrap;flex:none}'''

B = 'S4-ferma--'
BOARDS = [
 (B + 'pasari', 'S4 · Fermă · Păsări', lambda: page('pasari', panel_pasari())),
 (B + 'hrana', 'S4 · Fermă · Hrană', lambda: page('hrana', panel_hrana())),
 (B + 'greutate', 'S4 · Fermă · Greutate', lambda: page('greutate', panel_greutate())),
 (B + 'productie', 'S4 · Fermă · Producție', lambda: page('productie', panel_productie())),
]
if __name__ == '__main__':
    for stem, title, fn in BOARDS:
        EVID[0] = 0
        s4.write(stem, title, fn(), H=HEIGHTS.get(stem, 1874), css=CSS)
