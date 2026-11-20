"""S11 · Rapoarte · Hrană (desktop 1440): Stoc · Livrări · Consum · Cost, owner view on 18.11.2026 (SAPT 59), plus state boards.
One long page with section headings in screens.md order (not Tabs). Reuses the S4 shell (s4s_build.py: sidebar, icons,
Lamp/SexCount helpers, number formats, board writer). RF03: every figure, date and week from realdata (DATA03, D151–D159):
the 69 real Hîrbovăț deliveries (kg, lei; 4 fără preț) + Telița's 13; the Hîrbovăț feed type is only the stock the kg went
into (sheet column ♀ "gaina" → Breeder 2 · ♀, ♂ "cocos" → Cocoși · ♂; product unknown, names illustrative D30). The Consum
table (S14N) is a PeriodTable on the last 4 age weeks: week rows, their days, then the compartments H5.1–H7.2.
python3 -I s11_build.py <repo_root> [heights.json, default s11_heights.json] [--canvas]   (--canvas: refresh the 7 entries' h/title in place)
Writes design/canvas/project/S11-rapoarte-hrana[--<state>].dc.html and /tmp/s4s/<stem>.html (plain wrappers)."""
import sys, re, json, pathlib, importlib.util, datetime as dt
from collections import defaultdict

spec = importlib.util.spec_from_file_location('s4s_build', pathlib.Path(__file__).resolve().with_name('s4s_build.py'))
s4 = importlib.util.module_from_spec(spec); spec.loader.exec_module(s4)  # reads sys.argv[1] (repo root)
sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
import skel as SKL
import sexsign as SXS  # SX01: the §72 sign markup (Cost foot)
import realdata as R
ROOT, HP, CANVAS = CR.args('s11_heights.json')  # heights default: s11_heights.json
HEIGHTS = json.loads(HP.read_text()) if HP.exists() else {}

EX = s4.EX
n, ddate, icon, lamp, F, M = s4.n, s4.ddate, s4.icon, s4.lamp, s4.F, s4.M
FT = R.FEED_TYPES
TODAY = R.TODAY_ISO; WK = R.WK                       # 18.11.2026, SAPT 59 (D151)
FARM = {'hirbovat': 'Hîrbovăț', 'telita': 'Telița'}
NBSP = ' '

def d2(v): return n(v, 2)
def ddm(iso): return R.ddm(iso)
def nsp(v, dec=0): return f'<span class="ip-num">{n(v, dec)}</span>'
def dsp(iso): return f'<span class="ip-num">{ddate(iso)}</span>'

# icons not in the shell: package (delivery, from RecordTile preview = shell 'delivery'), Lucide coins (CostBlock README)
COINS = '<circle cx="8" cy="8" r="6"/><path d="M18.09 10.37A6 6 0 1 1 10.34 18"/><path d="M7 6h1v4"/><path d="m16.71 13.88.7.71-2.82 2.82"/>'
def rec_ico(path, size=16, rec='feed'):
    return f'<span class="ip-rec ip-rec--{rec} s11-ico"><svg class="ip-icon ip-icon--{size}" viewBox="0 0 24 24" aria-hidden="true">{path}</svg></span>'
PKG = rec_ico(s4.IC['delivery']); PKG14 = rec_ico(s4.IC['delivery'], 14); WHEAT14 = rec_ico(s4.IC['feed'], 14); COIN16 = rec_ico(COINS)

# ---------- data (realdata) ----------
DT = R._DEL['meta']['totals']                          # the extract's own totals, cross-checked below
dels = [dict(d, farm_id='telita') for d in R.DELIVERIES_TELITA] + list(R.DELIVERIES)
dels.sort(key=lambda d: (d['date'], d['feed_type_id'] == 'ft-breeder2'), reverse=True)  # newest first; same day: Breeder 2 above Cocoși
N_ALL = len(dels); N_TEL = len(R.DELIVERIES_TELITA); N_HIR = len(R.DELIVERIES)
assert (N_TEL, N_HIR) == (DT['telita']['deliveries'], DT['hirbovat']['deliveries'])
unpriced = [d for d in dels if not d['priced']]; UNP_KG = sum(d['qty_kg'] for d in unpriced)
assert unpriced == sorted(R.UNPRICED, key=lambda d: (d['date'], d['feed_type_id'] == 'ft-breeder2'), reverse=True)
assert sorted((d['date'], d['for_sex'], d['qty_kg']) for d in unpriced) == sorted(tuple(u) for u in DT['hirbovat']['unpriced'])
SHOWN = dels[:12]
EDIT = max(unpriced, key=lambda d: d['date']); EDIT_ID = EDIT['id']  # the owner's inline-edit row (newest unpriced: 12.11.2026)
assert EDIT in SHOWN
# S11M (D69): the owner changes a saved price with a required Motiv. Row = the newest priced Breeder 2 delivery on the
# board (13.11.2026). The new price on the error board is a labelled sample: the last earlier Breeder 2 price that differs
# from it at 2 decimals (the price as typed), total derived from kg (craft §7, as D68).
MOD = next(d for d in SHOWN if d['priced'] and d['feed_type_id'] == 'ft-breeder2')
MOD_PREV = max((d for d in dels if d['priced'] and d['feed_type_id'] == 'ft-breeder2' and d['farm_id'] == 'hirbovat' and d['date'] < MOD['date']
                and d2(d['price_per_kg']) != d2(MOD['price_per_kg'])), key=lambda d: d['date'])
MOD_NEW = R.rnd(MOD_PREV['price_per_kg'], 2); MOD_NEW_TOT = R.rnd(MOD['qty_kg'] * MOD_NEW)
ALERT = '<circle cx="12" cy="12" r="10"/><path d="M12 8v4"/><path d="M12 16h.01"/>'  # Lucide circle-alert (field error line, shared with S12)
def ico(path, size=16): return f'<svg class="ip-icon ip-icon--{size}" viewBox="0 0 24 24" aria-hidden="true">{path}</svg>'
def priced_by(d):  # the workbook has no who/when for a price: real prices read "din registru" (D163)
    return f'proprietar · {dsp(d["priced_at"])}' if d.get('priced_at') else 'din registru'

stocks = s4.stocks                                     # end of 18.11.2026; days left on the 16–18.11 mean (realdata.stock_days)
AVG_WIN = (R.TODAY - dt.timedelta(days=2), R.TODAY)
STOCK_LIMIT = s4.STOCK_LIMIT

# consumption by age week × section × sex (Hîrbovăț, the last 4 weeks), D46: compare over the days that have rows. A
# section-day with birds at the start and no kg in the workbook is "lipsă"; a sale day that empties it (kg 0) is not.
WEEKS = [WK - i for i in range(4)]
SECS = R.HSECS
def wdays(w):
    a, b = R.week_dates(w); b = min(b, R.TODAY)
    return [a + dt.timedelta(days=i) for i in range((b - a).days + 1)]
guide = {(w, sx): R.guide_feed(w, sx) for w in WEEKS for sx in 'FM'}
agg = defaultdict(lambda: {'kg': 0.0, 'g': [], 'bird': 0, 'days': 0, 'miss': []})
wdates = {w: [d.isoformat() for d in wdays(w)] for w in WEEKS}
for w in WEEKS:
    for d in wdays(w):
        for sec in SECS:
            for sx in 'FM':
                r, st = R.sd(d, sec, sx), R.alive_start(sec, sx, d)
                if not st or r is None: continue
                for key in ((w, sec, sx), (w, 'all', sx)):
                    a = agg[key]; a['days'] += 1
                    if r['kg'] is None: a['miss'].append((sec, d.isoformat())); continue
                    if not r['kg']: continue      # emptied by a sale that day: not fed, not missing
                    a['kg'] += r['kg']; a['g'].append(r['g']); a['bird'] += st
def g_avg(a, weighted=False):  # section: mean of the daily grams; farm: kg ÷ bird-days (♀ by hens, ♂ by roosters, D38)
    return a['kg'] * 1000 / a['bird'] if weighted else sum(a['g']) / len(a['g'])
def fed(w, sec): return any(agg[(w, sec, sx)]['g'] or agg[(w, sec, sx)]['miss'] for sx in 'FM')
EMPTY_HALLS = {h: R.empty_since(h) for h in R.HHALLS if R.empty_since(h)}  # H7: 14.11.2026 (D158d)
assert all(round(agg[(w, 'all', sx)]['kg'], 1) == round(R.week(w)[f'kg_{sx}'], 1) for w in WEEKS for sx in 'FM')
# D163: the workbook's kg cells far from g × birds at the start of the day (♀ 460 kg/compartiment, 02–06.11; README anomaly),
# kept as typed; named per shown week in the Consum foot
ODD = defaultdict(lambda: defaultdict(list))   # week → hall → [(date, g × birds kg)]
for w in WEEKS:
    for d in wdays(w):
        for sec in SECS:
            for sx in 'FM':
                r, st = R.sd(d, sec, sx), R.alive_start(sec, sx, d)
                if r and r['kg'] and st and not r['sold'] and abs(r['kg'] / (r['g'] * st / 1000) - 1) > 0.25:  # a sale day feeds fewer birds
                    ODD[w][sec.split('.')[0]].append((d, sx, r['kg'], r['g'] * st / 1000))
def odd_note():
    if not ODD: return ''
    cells = [c for w in ODD for h in ODD[w] for c in ODD[w][h]]
    kgs = {R.rnd(c[2]) for c in cells}; sxs = {c[1] for c in cells}
    assert len(kgs) == 1 and len(sxs) == 1, (kgs, sxs)   # one typed figure, one sex: one sentence
    exp = R.rnd(sum(c[3] for c in cells) / len(cells) / 10) * 10
    def span(ds): ds = sorted(set(ds)); return R.drange(ds[0], ds[-1], year_end=False) if len(ds) > 1 else R.ddm(ds[0])
    parts = []
    for w in sorted(ODD, reverse=True):
        by = defaultdict(list)                   # same days → one group of halls
        for h in sorted(ODD[w]): by[span([c[0] for c in ODD[w][h]])].append(h)
        parts.append(f'<span class="ip-num">{R.SAPT(w)}</span>: ' + ' · '.join(f'{", ".join(hs)} <span class="ip-num">{sp}</span>' for sp, hs in by.items()))
    sx = F if sxs == {'F'} else M
    return (f' {sx} <span class="ip-num">{n(kgs.pop())}{NBSP}kg</span>/compartiment în registru (g × păsări ≈ <span class="ip-num">{n(exp)}{NBSP}kg</span>): '
            + '; '.join(parts) + '.')
def dev(g, gd): return (g / gd - 1) * 100
def pct1(v):  # D46, no space; a value that rounds to 0 is written 0,0% (never −0,0%)
    v = round(v, 1)
    return s4.sgn(v, 1) + '%' if v != 0 else '0,0%'

# cost (CostBlock): priced deliveries; birds at the start of each cycle; eggs of the Hîrbovăț cycle to today
TEL_DEL = R.DELIVERIES_TELITA
tel_kg = sum(d['qty_kg'] for d in TEL_DEL); tel_lei = sum(d['total_lei'] for d in TEL_DEL)
hp = [d for d in R.DELIVERIES if d['priced']]; hir_kg = sum(d['qty_kg'] for d in hp); hir_lei = sum(d['total_lei'] for d in hp)
assert (tel_kg, tel_lei) == (DT['telita']['qty_kg'], DT['telita']['total_lei']) and R.rnd(hir_lei) == R.rnd(DT['hirbovat']['total_lei'])
assert hir_kg + UNP_KG == DT['hirbovat']['qty_kg_F'] + DT['hirbovat']['qty_kg_M']
PL = R.PLACEMENT_ROWS
tel_birds = PL['totals']['F'] + PL['totals']['M']
arr = dict(R.ARRIVED); hir_birds = arr['F'] + arr['M']
ARR_DAYS = sorted({r['date'] for r in R.SECTION_DAY if r['arr'] > 0})  # 22.02–28.02.2026
eggs = sum(r['eggs'] or 0 for r in R.FARM_DAY if r['date'] <= TODAY)
assert eggs == sum((R.week(k) or {}).get('eggs') or 0 for k in R.WEEKS)
_te = R.D(R.CTEL['end_date']); IV_A = dt.date(_te.year + _te.month // 12, _te.month % 12 + 1, 1).isoformat()  # Telița empty board: 01.03.2026
TEL_SPAN = f'{dsp(R.CTEL["start_date"])}–{dsp(R.CTEL["end_date"])}'
HEAD_META = (f'Seria 2026 · Hîrbovăț <span class="ip-num">{R.SAPT(WK)}</span>, ciclu din {dsp(R.CYCLE_START)} · '
             f'Telița fără ciclu deschis din {dsp(R.CTEL["end_date"])}')

# ---------- sidebar (Rapoarte current) ----------
def sidebar(role):
    viewer = role == 'Observator'
    sb = s4.sidebar('Observator', viewer=True) if viewer else s4.sidebar(role)
    sb = sb.replace('class="ip-nav is-current" href="#serie" aria-current="page"', 'class="ip-nav" href="#serie"')
    sb = sb.replace('class="ip-nav" href="#rapoarte"', 'class="ip-nav is-current" href="#rapoarte" aria-current="page"')
    assert sb.count('is-current') == 1
    return sb

# ---------- 1. page header + filter bar ----------
def header(role):
    acts = '' if role == 'Observator' else f'<div class="ip-row"><a class="ip-btn ip-btn--primary" href="#livrare-hrana">{icon("delivery", 16)}Livrare hrană</a></div>'
    return f'''<header class="ip-pagehead">
<div>
<nav class="ip-crumbs" aria-label="Cale"><a href="#rapoarte">Rapoarte</a><span aria-hidden="true">›</span><span aria-current="page">Hrană</span></nav>
<h1>Hrană</h1>
<div class="ip-row ip-pagehead__meta"><span class="ip-note">{HEAD_META}</span></div>
</div>
{acts}
</header>'''

def farm_select(sel):
    opts = ''.join(f'<option{" selected=\"selected\"" if o == sel else ""}>{o}</option>' for o in ['Toate fermele', 'Hîrbovăț', 'Telița'])
    return f'<span class="s11-farm"><label class="ip-label" for="f-farm">Fermă</label><span class="ip-input s11-select"><select id="f-farm">{opts}</select>{icon("chev", 16)}</span></span>'

def export(disabled=False):  # FilterBar end, as on S4 (s4s_build.EXPORT)
    dis = ' is-disabled" disabled="disabled' if disabled else ''
    return f'<button type="button" class="ip-btn ip-btn--ghost ip-filterbar__end{dis}">{icon("download", 16)}Exportă .xlsx</button>'

def filterbar(mode='default'):
    if mode == 'empty':  # Telița, Interval from the 1st of the month after its cycle closed (28.02.2026) to today: no delivery and no feed row
        itv = f'<span class="ip-interval" role="group" aria-label="Interval">{s4.datef("iv-a", "de la", ddate(IV_A))}{s4.datef("iv-b", "până la", ddate(TODAY))}</span>'
        return f'<div class="ip-filterbar s11-filterbar" role="group" aria-label="Filtre">{s4.seg("Perioadă", s4.PERIODS, 3)}{itv}{farm_select("Telița")}{export(True)}</div>'
    return f'<div class="ip-filterbar s11-filterbar" role="group" aria-label="Filtre">{s4.seg("Perioadă", s4.PERIODS, 0)}{farm_select("Toate fermele")}{export()}</div>'

def sec_head(hid, title, meta='', right=''):
    m = f'<span class="ip-note">{meta}</span>' if meta else ''
    return f'<div class="ip-sechead"><div class="s11-headl"><h2 class="ip-t-heading" id="{hid}">{title}</h2>{m}</div>{right}</div>'

def tscroll(cap, inner, fit=False, rowheads=False):
    """S11X: a table that scrolls sideways inside its own box wherever the content area is narrower than the table (Consum
    1113 px; Livrări beside S13's SidePanel), as a focusable region named by its caption (WCAG 1.4.10, 2.1.1). FIN14: the
    system scroll region (additions.css §36, framed §32: sticky first column) and row headers (§32); also used by S13."""
    f = ' s11-tablebox--fit' if fit else ''
    rh = ' ip-table--rowheads' if rowheads else ''
    return (f'<div class="ip-tablescroll ip-tablescroll--framed{f}" role="region" tabindex="0" aria-label="{cap}">'
            f'<table class="ip-table{rh} s11-table"><caption class="ip-sr">{cap}</caption>{inner}</table></div>')

# ---------- 2. Stoc ----------
def stock_ro(ft, sr):
    s = stocks[ft]; name = FT[ft]['name'].replace(' ', NBSP); ok = s['days_left'] >= STOCK_LIMIT
    rl = '<span class="ip-rlamp ip-rlamp--ok">în normă</span>' if ok else '<span class="ip-rlamp ip-rlamp--warn">abatere</span>'
    sub = '' if ok else f'<p class="ip-readout__sub ip-m0">sub pragul de {STOCK_LIMIT} zile</p>'
    w0, w1 = AVG_WIN
    return f'''<li><div class="ip-readout ip-readout--stock" role="group" aria-labelledby="sr-{ft}"><div class="ip-readout__cap"><span id="sr-{ft}">Hîrbovăț · {name}<span class="ip-sr">, {sr}</span></span>{rl}</div><div class="ip-readout__val">{n(s['stock_kg'])}<small> kg</small></div><dl class="ip-readout__kv"><div><dt>zile rămase</dt><dd>{n(s['days_left'], 1)}<small> zile</small></dd></div><div><dt>medie/zi {ddm(w0)}–{ddm(w1)}</dt><dd>{n(s['avg_daily_kg_last3'])}<small> kg</small></dd></div><div><dt>se termină</dt><dd>{ddate(s['run_out_date'])}</dd></div></dl>{sub}</div></li>'''
TELITA_RO = f'<li><div class="ip-readout ip-readout--stock ip-readout--off" role="group" aria-labelledby="sr-telita"><div class="ip-readout__cap"><span id="sr-telita">Telița</span><span class="ip-rlamp ip-rlamp--off">oprit</span></div><div class="ip-readout__val">–</div><p class="ip-readout__sub ip-m0">fără ciclu deschis</p><p class="ip-readout__sub ip-m0">ultimul ciclu: Seria 2026, închis pe <span class="ip-num ip-nowrap">{ddate(R.CTEL["end_date"])}</span></p></div></li>'
STOCK_NOTE = (f'<p class="ip-note ip-m0">Stoc la {dsp(TODAY)} = livrări − consum (rânduri confirmate). Zile rămase = stoc ÷ consumul mediu din ultimele 3 zile încheiate '
              f'(<span class="ip-num">{R.drange(*AVG_WIN)}</span>); sub {STOCK_LIMIT} zile = abatere.</p>')
def sec_stock(farm='all'):
    ros = TELITA_RO if farm == 'telita' else stock_ro('ft-breeder2', 'pentru găini') + stock_ro('ft-cocosi', 'pentru cocoși') + TELITA_RO
    return f'''<section class="ip-block" aria-labelledby="h-stoc">
{sec_head("h-stoc", "Stoc", f'pe fermă și tip de hrană · {dsp(TODAY)}')}
<ul class="ip-readouts" aria-labelledby="h-stoc">{ros}</ul>
{'' if farm == 'telita' else STOCK_NOTE}
</section>'''

# ---------- 3. Livrări ----------
def mod_rows(d, head, lab, err):
    """S11M (D69), one row height since FW03 (D150): "Modifică" turns the delivery row into the edited row
    (tr.ip-table__edit, the table's row height): Preț nou in the Lei/kg and Total lei columns (the s11-in lei/kg ↔ total lei
    inputs, where the saved figures were), who/when kept. Under it one detail row (additions.css §69): the hint naming the
    saved price and how the other figure is calculated, then the required Motiv field + Salvează prețul / Anulează + the journal
    note on one line, then Motiv's error. err: Salvează pressed with Motiv empty."""
    t = FT[d['feed_type_id']]; nm = t['name']; did = d['id']
    old = f'{d2(d["price_per_kg"])}{NBSP}lei/kg'
    kg_v = f' value="{d2(MOD_NEW)}"' if err else ''
    desc = 'mp-help mp-sample' if err else 'mp-help'
    kg = (f'<span class="ip-input s11-in{"" if err else " is-focus"}"><input id="mp-kg" type="text" inputmode="decimal"{kg_v} aria-label="Lei/kg nou, livrarea din {lab}" aria-describedby="{desc}"/><span class="suffix">lei/kg</span></span>')
    tot = (f'<span class="ip-input s11-in s11-in--tot is-suggested"><input id="mp-tot" type="text" inputmode="decimal" value="{n(MOD_NEW_TOT)}" aria-label="Total lei nou, livrarea din {lab}" aria-describedby="{desc}"/><span class="suffix">lei</span></span>' if err
           else f'<span class="ip-input s11-in s11-in--tot"><input id="mp-tot" type="text" inputmode="decimal" aria-label="Total lei nou, livrarea din {lab}" aria-describedby="{desc}"/><span class="suffix">lei</span></span>')
    row = (f'<tr class="ip-table__edit" id="modifica-pret-{did}">{head}<td class="edit">{kg}</td><td class="edit">{tot}</td>'
           f'<td class="ip-muted">{priced_by(d)}<span class="ip-sr">, preț salvat {old}</span></td><td class="s11-act"></td></tr>')
    hint = (f'<p class="ip-hint" id="mp-help">Preț salvat: <span class="ip-num">{old}</span> · <span class="ip-num">{n(d["total_lei"])}{NBSP}lei</span>. '
            f'Scrie prețul nou în lei/kg sau totalul; celălalt se calculează din <span class="ip-num">{n(d["qty_kg"])}{NBSP}kg</span>.</p>')
    jur = f'<p class="ip-hint ip-table__aside" id="mp-jur"><span>Prețul vechi (<span class="ip-num">{old}</span>), cel nou și motivul rămân în jurnal, în Istoric modificări.</span></p>'
    keys = '<span class="ip-table__keys"><button type="button" class="ip-btn ip-btn--secondary">Salvează prețul</button><button type="button" class="ip-btn ip-btn--ghost">Anulează</button></span>'
    if err:
        msg = (f'{nm} <span class="ip-num">{ddate(d["date"])}</span>: scrie motivul ca să schimbi prețul din <span class="ip-num">{d2(d["price_per_kg"])}</span> '
               f'în <span class="ip-num">{d2(MOD_NEW)}{NBSP}lei/kg</span>')
        fld = ('<span class="ip-input is-error is-focus"><input id="mp-mot" type="text" required="required" aria-required="true" aria-invalid="true" aria-errormessage="mp-err" aria-describedby="mp-err mp-jur" placeholder="de ex. preț corectat pe factură"/></span>')
        e = f'<p class="ip-err ip-fieldmsg" id="mp-err">{ico(ALERT, 14)}<span>{msg}</span></p>'
    else:
        fld = '<span class="ip-input"><input id="mp-mot" type="text" required="required" aria-required="true" aria-describedby="mp-jur" placeholder="de ex. preț corectat pe factură"/></span>'
        e = ''
    why = f'<div class="ip-field"><label class="ip-label" for="mp-mot">Motiv</label>{fld}</div>'
    detail = f'<tr class="ip-table__detail"><td colspan="8"><div class="ip-table__detailbody">{hint}{why}{e}{keys}{jur}</div></td></tr>'
    return row + detail

def del_row(d, role, mode='default'):
    t = FT[d['feed_type_id']]; sx = {'F': ' · ' + F, 'M': ' · ' + M}.get(t['for_sex'], '')
    who = f'<td class="num">{n(d["qty_kg"])}</td>'
    lab = f'{ddate(d["date"])}, {n(d["qty_kg"])} kg {t["name"]}'
    head = f'<td class="num">{ddate(d["date"])}</td><td>{FARM[d["farm_id"]]}</td><td><span class="s11-type">{PKG}<span>{t["name"]}{sx}</span></span></td>{who}'
    owner = role == 'Proprietar'
    if role == 'Observator': return f'<tr>{head}</tr>'  # D64: viewers see no prices
    if owner and mode in ('mod', 'moderr') and d['id'] == MOD['id']: return mod_rows(d, head, lab, mode == 'moderr')
    if owner and mode == 'default' and d['id'] == EDIT_ID:  # inline edit (DataTable: medium TextField, lei/kg ↔ total)
        return f'''<tr class="ip-table__edit">{head}<td class="edit"><span class="ip-input is-focus s11-in"><input id="pr-kg" type="text" inputmode="decimal" aria-label="Lei/kg, livrarea din {lab}" aria-describedby="pr-help"/><span class="suffix">lei/kg</span></span></td><td class="edit"><span class="ip-input s11-in s11-in--tot"><input id="pr-tot" type="text" inputmode="decimal" aria-label="Total lei, livrarea din {lab}" aria-describedby="pr-help"/><span class="suffix">lei</span></span></td><td colspan="2" class="edit"><span class="s11-acts"><button type="button" class="ip-btn ip-btn--secondary">Salvează prețul</button><button type="button" class="ip-btn ip-btn--ghost">Anulează</button></span></td></tr>'''
    if d['priced']:
        by = priced_by(d)
        price = f'<td class="num s11-pc">{d2(d["price_per_kg"])}</td><td class="num s11-pc">{n(d["total_lei"])}</td><td class="ip-muted">{by}</td>'
        act = f'<td class="s11-act"><a class="ip-link s11-lk" href="#modifica-pret-{d["id"]}" aria-label="Modifică prețul, livrarea din {lab}">Modifică</a></td>' if owner else ''
    else:
        price = '<td class="num s11-pc"><span class="ip-tag ip-tag--info">fără preț</span></td><td class="num s11-pc">–</td><td class="ip-muted">–</td>'
        act = f'<td class="s11-act"><a class="ip-link s11-lk" href="#pret-{d["id"]}" aria-label="Adaugă prețul, livrarea din {lab}">Adaugă preț</a></td>' if owner else ''
    return f'<tr>{head}{price}{act}</tr>'

def del_ths(role):
    ths = ('<th scope="col" class="num">Data</th><th scope="col">Fermă</th>'
           f'<th scope="col"><span class="ip-rec ip-rec--feed ip-table__ico">{icon("delivery", 14)}</span>Tip hrană</th><th scope="col" class="num">Kg</th>')
    if role != 'Observator': ths += '<th scope="col" class="num s11-thedit">Lei/kg</th><th scope="col" class="num s11-thedit">Total lei</th><th scope="col">Preț adăugat</th>'
    if role == 'Proprietar': ths += '<th scope="col"><span class="ip-sr">Acțiuni</span></th>'
    if role == 'Proprietar': ths = SKL.widths(ths, W_DEL, M_DEL)  # FIN11G: the loaded column shares, shared with the loading table
    if role == 'Observator': ths = SKL.widths(ths, W_DEL_RO)  # FW01 (D148): the 4-column table spans the block, Tip hrană takes the rest
    return ths
# FIN11G (§54): column shares of the owner's tables at 1440 (as content-sized before), declared on the head the loading
# skeleton reuses, so its columns sit where the loaded ones do
W_DEL = [11.0, 8.2, 13.7, 7.7, 14.7, 17.3, None, 8.6]   # None: the column that takes the rest
W_DEL_RO = [11.0, 11.0, None, 9.0]   # FW01: viewer Livrări (no prices) at the block's full width, was fit-content (576)
W_COST = [None, 13.6, 14.7, 13.4, 17.9]
# FIN11G fix 1: each column's loaded min-content (px, without padding; measured with the table at its narrowest), so at
# 1280, where the content outgrows the shares, the loading columns grow as the loaded ones do
M_DEL = [78.28, 52.2, 103.25, 46.97, 116, 140, 150.08, 75.33]
M_COST = [256.53, 70.45, 78.28, 69.09, 100.11]

def sec_deliveries(role, mode='default'):
    owner = role == 'Proprietar'; viewer = role == 'Observator'
    ths = del_ths(role)
    rows = ''.join(del_row(d, role, mode) for d in SHOWN)
    unp_days = sorted({d['date'] for d in unpriced}); unp_dates = ', '.join(ddm(x) for x in unp_days[:-1]) + (', ' if len(unp_days) > 1 else '') + ddate(unp_days[-1])
    if owner:
        right = f'<a class="ip-link s11-s12 s11-lk" href="#preturi-hrana">{rec_ico(COINS, 14)}<span><span class="ip-num">{len(unpriced)}</span> livrări fără preț · Adaugă prețurile</span></a>'
    elif viewer:  # D64: the count only, no lei, no S12 link
        right = f'<span class="ip-note s11-s12"><span class="ip-num">{len(unpriced)}</span> livrări fără preț</span>'
    else:
        right = f'<span class="ip-note s11-s12">{rec_ico(COINS, 14)}<span><span class="ip-num">{len(unpriced)}</span> livrări fără preț · prețurile le adaugă proprietarul</span></span>'
    help_ = f'<p class="ip-sr" id="pr-help">Scrie lei/kg sau totalul în lei; celălalt se calculează din {n(EDIT["qty_kg"])} kg.</p>' if owner and mode == 'default' else ''
    if mode == 'moderr':  # D68-style label for the typed sample price
        help_ = (f'<p class="ip-hint ip-m0" id="mp-sample">Valoare de exemplu: <span class="ip-num">{d2(MOD_NEW)}{NBSP}lei/kg</span> la <span class="ip-num">{ddate(MOD["date"])}</span> '
                 f'(prețul livrării Breeder 2 din <span class="ip-num">{ddate(MOD_PREV["date"])}</span>); totalul <span class="ip-num">{n(MOD_NEW_TOT)}{NBSP}lei</span> e calculat din <span class="ip-num">{n(MOD["qty_kg"])}{NBSP}kg</span>.</p>')
    unp = '' if viewer else f' · fără preț: <span class="ip-num">{unp_dates}</span> · <span class="ip-num">{n(UNP_KG)} kg</span>'
    foot = f'<div class="s11-foot"><span class="ip-note">Afișate <span class="ip-num">{len(SHOWN)}</span> din <span class="ip-num">{N_ALL}</span>{unp}</span><a class="ip-link s11-lk" href="#livrari-toate">Vezi toate (<span class="ip-num">{N_ALL}</span>)</a></div>'
    ro = ' · prețuri doar de citit' if role == 'Manager' else ''
    return f'''<section class="ip-block" aria-labelledby="h-livrari">
{sec_head("h-livrari", "Livrări", f'<span class="ip-num">{N_ALL}</span> livrări în Seria 2026 (Hîrbovăț <span class="ip-num">{N_HIR}</span> · Telița <span class="ip-num">{N_TEL}</span>) · cele mai noi întâi{ro}', right)}
{tscroll(f'Livrări hrană, cele mai noi {len(SHOWN)} din {N_ALL}', f'<thead><tr>{ths}</tr></thead><tbody>{rows}</tbody>')}
{help_}{foot}
</section>'''

# ---------- 4. Consum ----------
# RF03 → S14N (spec D4 › Consum, D154/D166): a PeriodTable (SYS14, imported through s4s_build): per age week the farm's week
# row, its daily farm rows (every week open), then the compartment week rows (rowGroups); g/pasăre on every week row = kg ÷
# bird-days (♀ by hens, ♂ by roosters, D38; compartments too, D163); a day = the register's g (realdata.feed_g), its guide
# the day's; a compartment-day with birds and no kg is "lipsă" (D46: the week compares over the days that have rows)
P = s4.P
def pct1s(v): return pct1(v)
def miss_txt(w, key, day=None):
    miss = [(sx, sec, d) for sx in 'FM' for (sec, d) in agg[(w, key, sx)]['miss'] if day is None or d == day.isoformat()]
    if not miss: return ''
    if key != 'all': return ' · '.join(f'{F if sx == "F" else M} <span class="ip-num">{ddm(d)}</span>' for sx, _, d in miss)
    if day is not None: return ' · '.join(f'{sec} {F if sx == "F" else M}' for sx, sec, _ in miss)
    ds = sorted({d for *_, d in miss}); span = ddm(ds[0]) if len(ds) == 1 else f'{ddm(ds[0])}–{ddm(ds[-1])}'
    return f'<span class="ip-num">{len(miss)}</span> rânduri · <span class="ip-num">{span}</span>'
def cons_groups(key='all'):
    """key 'all' = Total fermă (week, day and period rows); a compartment code = its week rows (rowGroups)"""
    def a(kind, k, sx): return agg[(k, key, sx)]
    def g_v(kind, k, sx):
        if kind == 'w': return g_avg(agg[(k, key, sx)], True)
        if kind == 'd': return R.feed_g(k, sx)
        t = [agg[(w, key, sx)] for w in WEEKS]; return sum(x['kg'] for x in t) * 1000 / sum(x['bird'] for x in t)
    def gd_v(kind, k, sx):
        if kind == 'w': return guide[(k, sx)]
        if kind == 'd': return R.day(k)[f'guide_g_{sx}']
        t = {w: agg[(w, key, sx)]['bird'] for w in WEEKS}; return sum(guide[(w, sx)] * t[w] for w in WEEKS) / sum(t.values())   # bird-day weighted (D162)
    def kg(sx):
        def f(kind, k):
            if kind == 'w': return n(agg[(k, key, sx)]['kg'], 1)
            if kind == 'd': return n(R.feed_kg(k, sx), 1)
            return n(sum(agg[(w, key, sx)]['kg'] for w in WEEKS), 1)
        return f
    def g(sx): return lambda kind, k: n(g_v(kind, k, sx), 1)
    def gd(sx): return lambda kind, k: n(gd_v(kind, k, sx))
    def dv(sx): return lambda kind, k: pct1(dev(g_v(kind, k, sx), gd_v(kind, k, sx)))
    def stare(kind, k):
        what = '' if kind == 'f' else miss_txt(R.sapt(k) if kind == 'd' else k, key, k if kind == 'd' else None)
        if what: return f'<span class="s11-state">{lamp("warn", "lipsă")}<span class="ip-muted">{what}</span></span>'
        if any(abs(dev(g_v(kind, k, sx), gd_v(kind, k, sx))) > 10 for sx in 'FM'): return lamp('warn', 'abatere')
        return lamp('ok', 'în normă')
    kgg = P.group(P.hrana(), 'Hrană · kg')
    sx = P.sx
    return [{'name': kgg['name'], 'leaves': [(sx('F'), 'is-hen', kg('F')), (sx('M'), 'is-rooster', kg('M'))]},
            {'name': P.group(P.hrana(), 'Hrană · g/pasăre')['name'],
             'leaves': [(sx('F', 'efectiv'), 'is-hen', g('F')), (sx('F', 'ghid'), 'is-hen', gd('F')), (sx('F', 'față de ghid'), 'is-hen', dv('F')),
                        (sx('M', 'efectiv'), 'is-rooster', g('M')), (sx('M', 'ghid'), 'is-rooster', gd('M')), (sx('M', 'față de ghid'), 'is-rooster', dv('M')),
                        ('stare', 'txt', stare)]}]
assert [R.n(R.feed_kg(R.TODAY, x), 1) for x in 'FM'] == ['2 303,2', '154,7'] and R.n(R.feed_g(R.TODAY, 'F')) == '160' and R.day(R.TODAY)['guide_g_F'] == 157   # spec D4 sample
EMPTY_HALL_TXT = {h: f'goală din {dsp(EMPTY_HALLS[h])} · {", ".join(R.HALL_SECS[h])}' for h in EMPTY_HALLS}
def cons_rows():
    out = [{'label': sec, 'sr': '', 'sections': [('', '', cons_groups(sec))], 'skip': (lambda k, sec=sec: not fed(k, sec))} for sec in SECS]
    out += [{'label': h, 'sr': '', 'sections': [('', '', cons_groups())], 'empty': (lambda k, h=h: EMPTY_HALL_TXT[h]),
             'skip': (lambda k, h=h: any(fed(k, sec) for sec in R.HALL_SECS[h]))} for h in EMPTY_HALLS]   # an emptied hall: one muted line (D158d)
    return out
CONS_SKEL = [4, 3, 4, 3, 4, 4, 3, 4, 4]
CONS_CAP = f'Consum · Hîrbovăț · Seria 2026 · {R.SAPT(WEEKS[-1])}–{WEEKS[0]} · pe zile și secțiuni'
def cons_table(skel=False):
    t, _ = P.ptable('cs', CONS_CAP, [('Total fermă', '', cons_groups())], weeks=WEEKS, rowgroups=cons_rows(), skel=CONS_SKEL if skel else False,
                    more=P.more())   # SYS18 (D190, D192): the load-more row, 4 din 40 (Ciclul curent, D191)
    if not skel: assert t.count(s4.PART[0]) == 1; t = t.replace(s4.PART[0], s4.PART[1])   # "17–18.11 · 2 din 7 zile" (D165)
    assert t.count('aria-expanded="true"') == (0 if skel else len(WEEKS))   # D166: every week open
    a = 'class="ip-table ip-table--rowheads ip-ptable"'; assert t.count(a) == 1   # dense figure cells; SYS17 (D189): heads on one line, no wraphead
    return t.replace(a, 'class="ip-table ip-table--rowheads ip-ptable ip-ptable--dense"')
def sec_consum():
    span = f'{R.SAPT(WEEKS[-1])}–{WEEKS[0]}'
    foot = (f'<div class="s11-foot"><span class="ip-note">g/pasăre: zi = g din registru; săptămână, secțiune și perioadă = kg ÷ zile-pasăre ({F} după găini, {M} după cocoși). Toleranță <span class="ip-num">±10%</span>. '
            f'Telița: fără consum în ciclul curent (ciclu închis pe {dsp(R.CTEL["end_date"])}).{odd_note()}</span></div>')   # D191: the load-more row replaces "Vezi toate săptămânile"
    return f'''<section class="ip-block" aria-labelledby="h-consum">
{sec_head("h-consum", "Consum", f'Hîrbovăț · pe săptămână, zi și secțiune · <span class="ip-num">{span}</span> · ghid Ross 308 ouat')}
{P.block('cs', CONS_CAP, cons_table(), bar=False)}
{foot}
</section>'''

# ---------- 5. Cost (CostBlock, ip-cost--wide) ----------
def unp_row(cyc, kg, note, tag=True):
    lbl = '<span class="ip-tag ip-tag--info">fără preț</span>' if tag else '<span>fără preț</span>'
    return f'<tr class="ip-cost__unpriced"><th scope="row"><span class="ip-sr">{cyc}, </span><span>{COIN16}{lbl}</span></th><td class="num">{n(kg)}</td><td class="ip-cost__note" colspan="3">{note}</td></tr>'
COST_THS = SKL.widths('<th scope="col">Ciclu</th><th scope="col" class="num">Kg</th><th scope="col" class="num">Lei</th><th scope="col" class="num">Lei/pasăre</th><th scope="col" class="num">Lei/ou</th>', W_COST, M_COST)
def sec_cost():
    d0, d1 = min(d['date'] for d in unpriced), max(d['date'] for d in unpriced)
    t = f'''<thead><tr>{COST_THS}</tr></thead>
<tbody><tr><th scope="row">Telița · Seria 2026<span class="ip-cost__sub">{TEL_SPAN} · <span class="ip-num">{N_TEL}</span> livrări cu preț</span></th><td class="num">{n(tel_kg)}</td><td class="num">{n(tel_lei)}</td><td class="num">{d2(tel_lei / tel_birds)}</td><td class="num">–<small>creștere, fără ouă</small></td></tr>{unp_row("Telița · Seria 2026", 0, "Toate livrările au preț.", False)}</tbody>
<tbody><tr><th scope="row">Hîrbovăț · Seria 2026<span class="ip-cost__sub">din {dsp(R.CYCLE_START)} · <span class="ip-num">{R.SAPT(WK)}</span> · <span class="ip-num">{len(hp)}</span> livrări cu preț</span></th><td class="num">{n(hir_kg)}</td><td class="num">{n(hir_lei)}</td><td class="num">{d2(hir_lei / hir_birds)}</td><td class="num">{d2(hir_lei / eggs)}</td></tr>{unp_row("Hîrbovăț · Seria 2026", UNP_KG, f'<span class="ip-num">{len(unpriced)}</span> livrări, <span class="ip-num">{R.drange(d0, d1)}</span> · neincluse în cost')}</tbody>
<tbody><tr class="is-total"><th scope="row">Total Seria 2026<span class="ip-cost__sub">nu include <span class="ip-num">{n(UNP_KG)} kg</span> fără preț</span></th><td class="num">{n(tel_kg + hir_kg)}</td><td class="num">{n(tel_lei + hir_lei)}</td><td class="num">{d2((tel_lei + hir_lei) / tel_birds)}</td><td class="num">{d2((tel_lei + hir_lei) / eggs)}</td></tr></tbody>'''
    foot = (f'<p class="ip-note ip-cost__foot">Lei = suma livrărilor cu preț; kg fără preț nu intră în cost. Lei/pasăre: la păsările de la începutul ciclului (Telița: populare {dsp(R.PLACEMENT)}, '
            f'{SXS.figure("F", n(PL["totals"]["F"]), num=True)} · {SXS.figure("M", n(PL["totals"]["M"]), num=True)}; Hîrbovăț: sosite <span class="ip-num">{R.drange(ARR_DAYS[0], ARR_DAYS[-1])}</span>, '
            f'{SXS.figure("F", n(arr["F"]), num=True)} · {SXS.figure("M", n(arr["M"]), num=True)}; seria: populare). Lei/ou: ouă totale până pe {dsp(TODAY)}: <span class="ip-num">{n(eggs)}</span>.</p>')
    return f'''<section class="ip-block" aria-labelledby="h-cost">
{sec_head("h-cost", "Cost", 'pe ciclu, pe pasăre și pe ou · din livrările cu preț')}
<section class="ip-cost ip-cost--wide" aria-labelledby="cost-d"><div class="ip-cost__head"><h3 class="ip-cap" id="cost-d">Cost hrană · Seria 2026</h3><span class="ip-note">din livrările cu preț · până pe {dsp(TODAY)}</span></div><div class="ip-cost__scroll" role="region" tabindex="0" aria-label="Cost hrană · Seria 2026"><table class="ip-table ip-table--rowheads ip-cost__table"><caption class="ip-sr">Cost hrană · Seria 2026</caption>{t}</table></div>{foot}</section>
</section>'''

# ---------- states ----------
def sk(c=''): return s4.sk(c)
def sec_loading():
    ro = f'<li>{SKL.readout()}</li>'  # §54 StockReadout skeleton, as tall as the loaded row, as S2 (FIN11G; FIN11 §35 box, was s11-skro)
    def tbl(ths, rows, cost=''):
        assert ths.count('<th scope="col"') == len(re.findall(r'<th scope="col"(?: class="[^"]*")?(?: style="[^"]*")?>', ths)), 'unmatched head cell'  # the real thead; skeleton cells keep each column's alignment
        cols = re.findall(r'<th scope="col"(?: class="([^"]*)")?(?: style="[^"]*")?>', ths)  # FIN11G: heads carry the §54 style  # num → td.num; s11-thedit → td.s11-pc (same right edge)
        cl = lambda c: (' class="' + ' '.join(['num'] * ('num' in c.split()) + ['s11-pc'] * ('s11-thedit' in c.split())) + '"') if 'num' in c.split() else ''
        tr = ''.join('<tr>' + ''.join(f'<td{cl(c)}>{sk("s4s-w70")}</td>' for c in cols) + '</tr>' for _ in range(rows))
        if cost:   # FW04 (D150): CostBlock rows are two lines (§70), so the Ciclu cell's bar has a sub-line bar under it, as the loaded name + sub-line
            sub = f'<span class="ip-cost__sub"><span class="ip-skel-line">{sk("s4s-w40")}</span></span>'
            tr = ''.join('<tr>' + ''.join(f'<td{cl(c)}>{sk("s4s-w70")}{sub if i == 0 else ""}</td>' for i, c in enumerate(cols)) + '</tr>' for _ in range(rows))
            return f'<table class="ip-table s11-table {cost}"><thead><tr>{ths}</tr></thead><tbody>{tr}</tbody></table>'
        return f'<div class="ip-tablescroll ip-tablescroll--framed"><table class="ip-table s11-table"><thead><tr>{ths}</tr></thead><tbody>{tr}</tbody></table></div>'  # aria-hidden: scrolls, not a tab stop
    return f'''<div class="s11-page" aria-busy="true">
<p class="ip-sr" role="status">Se încarcă raportul Hrană pentru Seria 2026.</p>
<section class="ip-block" aria-hidden="true">{sec_head("h-stoc", "Stoc")}<ul class="ip-readouts">{ro * 3}</ul></section>
<section class="ip-block" aria-hidden="true">{sec_head("h-livrari", "Livrări")}{tbl(del_ths("Proprietar"), 6)}</section>
<section class="ip-block" aria-hidden="true">{sec_head("h-consum", "Consum")}<p class="ip-cap ip-ptable__cap">{CONS_CAP}</p>{P.region(cons_table(True), CONS_CAP, hidden=True)}</section>
<section class="ip-block" aria-hidden="true">{sec_head("h-cost", "Cost")}<div class="ip-cost ip-cost--wide"><div class="ip-cost__head">{sk("s4s-w40")}</div><div class="ip-cost__scroll">{tbl(COST_THS, 5, "ip-cost__table")}</div></div></section>
</div>'''

def sec_empty():
    tl = max(TEL_DEL, key=lambda d: d['date'])   # Telița's last real delivery (before its cycle closed)
    assert all(d['date'] < IV_A for d in TEL_DEL) and max(r['date'] for r in R.FEED_TELITA) < IV_A
    def box(hid, lines, step):
        body = ''.join(f'<p class="ip-empty__line">{l}</p>' for l in lines)
        return f'<div class="ip-emptybox"><div class="ip-empty s11-empty">{lamp("off", "fără date")}{body}{step}</div></div>'
    itv = f'{dsp(IV_A)} și {dsp(TODAY)}'
    IV_NOTE = f'Telița · <span class="ip-num">{R.drange(IV_A, TODAY)}</span>'
    step_all = '<button type="button" class="ip-btn ip-btn--secondary">Vezi toate fermele</button>'
    return f'''<div class="s11-page">
{sec_stock('telita')}
<section class="ip-block" aria-labelledby="h-livrari">{sec_head("h-livrari", "Livrări", IV_NOTE)}{box("h-livrari", [f"Nicio livrare la Telița între {itv}.", f'Ultima livrare: <span class="ip-num">{ddate(tl["date"])}</span> · <span class="ip-num">{n(tl["qty_kg"])} kg</span> {FT[tl["feed_type_id"]]["name"]}.'], step_all)}</section>
<section class="ip-block" aria-labelledby="h-consum">{sec_head("h-consum", "Consum", IV_NOTE)}{box("h-consum", [f"Niciun consum la Telița între {itv}.", f'Ciclul Seria 2026 s-a încheiat pe {dsp(R.CTEL["end_date"])}.'], '<a class="ip-link s11-lk" href="#telita-seria-2026">Vezi ciclul Seria 2026 la Telița</a>')}</section>
<section class="ip-block" aria-labelledby="h-cost">{sec_head("h-cost", "Cost", IV_NOTE)}{box("h-cost", [f"Nicio livrare cu preț între {itv}, deci niciun cost în interval."], '<a class="ip-link s11-lk" href="#cost-seria-2026">Vezi costul Seria 2026</a>')}</section>
</div>'''

def page(role='Proprietar', mode='default'):
    if mode == 'loading': body = sec_loading()
    elif mode == 'empty': body = sec_empty()
    else: body = f'<div class="s11-page">{sec_stock()}{sec_deliveries(role, mode)}{sec_consum()}{"" if role == "Observator" else sec_cost()}</div>'
    h = header(role)
    return f'<div class="ip-shell">{sidebar(role)}<main class="ip-shell__main">\n{h}\n{filterbar("empty" if mode == "empty" else "default")}\n{body}\n</main></div>'

CSS = s4.CSS + '\n' + s4.CSS_STATES + '''
.s11-page{display:flex;flex-direction:column;gap:var(--space-8)}
.s11-headl{display:flex;align-items:baseline;gap:var(--space-3);flex-wrap:wrap}
.s11-farm{display:inline-flex;align-items:center;gap:var(--space-2)}
.s11-farm .ip-label{margin:0}
.s11-select{width:calc(var(--space-12) * 4);cursor:pointer}
.s11-filterbar:has(> .ip-interval) .s11-farm{flex-direction:column;align-items:flex-start;gap:var(--space-1)}
.s11-gsub{color:var(--ink-secondary);font-weight:400;font-family:var(--font-mono);font-size:var(--ip-fs-small)}
.s11-ico{color:var(--c);display:inline-flex;flex:none}
.s11-type{display:inline-flex;align-items:center;gap:var(--space-2)}
.s11-hen{color:var(--rec-hen)}
.s11-rooster{color:var(--rec-rooster)}
.s11-state{display:inline-flex;align-items:center;gap:var(--space-2)}
.s11-act{text-align:right;width:1%}
.s11-lk{display:inline-flex;align-items:center;min-height:var(--space-6)}
.s11-in{width:calc(var(--space-12) * 2 + var(--space-6));margin-left:auto;font-family:var(--font-mono)}
.s11-in > input{text-align:right;font-variant-numeric:tabular-nums}
.s11-in--tot{width:calc(var(--space-12) * 3)}
.s11-acts{display:flex;gap:var(--space-2);justify-content:flex-end}
.s11-s12{display:inline-flex;align-items:center;gap:var(--space-2);white-space:nowrap}
.s11-foot > .ip-link{white-space:nowrap;flex:none}
.s11-foot{display:flex;justify-content:space-between;align-items:center;gap:var(--space-4)}
.s11-table td.num > .ip-skel{margin-left:auto}
.s11-empty{padding:var(--space-8) var(--space-4)}
.s11-table{--ip-row-field-h:var(--control-h)} /* FW03 (additions.css §69): rows held at 37 by the 24px links, so in-row fields keep control-h */
.s11-table tr.ip-table__edit > td.edit{padding-block:calc(var(--space-1) / 2)} /* … in td.edit's 2px (bundle.css): 37 as the rows round it */
.s11-table th.s11-thedit,.s11-table td.s11-pc{padding-right:var(--space-2)} /* Lei/kg + Total lei: header, figures and the inputs' right edge on one line (td.edit padding) */'''

STEM = 'S11-rapoarte-hrana'
BOARDS = [
 (STEM, 'S11 · Rapoarte · Hrană', page()),
 (STEM + '--se-incarca', 'S11 · Rapoarte · Hrană · se încarcă', page(mode='loading')),
 (STEM + '--manager', 'S11 · Rapoarte · Hrană · manager (prețuri doar de citit)', page('Manager')),
 (STEM + '--observator', 'S11 · Rapoarte · Hrană · observator', page('Observator')),
 (STEM + '--fara-livrari-in-perioada', 'S11 · Rapoarte · Hrană · fără livrări în perioadă', page(mode='empty')),
 (STEM + '--modifica-pret', 'S11 · Rapoarte · Hrană · modifică preț', page(mode='mod')),
 (STEM + '--modifica-pret--eroare', 'S11 · Rapoarte · Hrană · modifică preț · eroare', page(mode='moderr')),
]
if __name__ == '__main__':
    for stem, title, body in BOARDS:
        s4.write(stem, title, body, '', H=HEIGHTS.get(stem, 1024), css=CSS)
    if CANVAS:  # RF03: refresh the 7 hand-placed entries' h/title in place (x, y kept; canvas_rows.place)
        c = CR.load(ROOT); CR.place(c, STEM, [(stem, 1440, HEIGHTS.get(stem, 1024), title) for stem, title, _ in BOARDS]); CR.save(ROOT, c)
