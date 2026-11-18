"""S11 · Rapoarte · Hrană (desktop 1440): Stoc · Livrări · Consum · Cost, owner view on 16.11.2026, plus state boards.
One long page with section headings in screens.md order (not Tabs). Reuses the S4 shell (s4s_build.py: sidebar, icons,
Lamp/SexCount helpers, number formats, board writer). Every figure comes from the extracts.
python3 -I s11_build.py <repo_root> [heights.json, default s11_heights.json]
Writes design/canvas/project/S11-rapoarte-hrana[--<state>].dc.html and /tmp/s4s/<stem>.html (plain wrappers)."""
import sys, re, json, pathlib, importlib.util, datetime as dt
from collections import defaultdict

spec = importlib.util.spec_from_file_location('s4s_build', pathlib.Path(__file__).resolve().with_name('s4s_build.py'))
s4 = importlib.util.module_from_spec(spec); spec.loader.exec_module(s4)  # reads sys.argv[1] (repo root)
sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
import skel as SKL
ROOT, HP, CANVAS = CR.args('s11_heights.json')  # heights default: s11_heights.json
HEIGHTS = json.loads(HP.read_text()) if HP.exists() else {}

EX = s4.EX; S = s4.S; GL = s4.GL; CY = s4.CY
n, ddate, icon, lamp, F, M = s4.n, s4.ddate, s4.icon, s4.lamp, s4.F, s4.M
FD = json.load(open(EX / 'feed-deliveries.json', encoding='utf-8'))
FT = {t['id']: t for t in json.load(open(EX / 'feed-types.json', encoding='utf-8'))['rows']}
PL = json.load(open(EX / 'placement-seria2026.json', encoding='utf-8'))
TR = json.load(open(EX / 'transfers-seria2026.json', encoding='utf-8'))
TODAY = '2026-11-16'
FARM = {'hirbovat': 'Hîrbovăț', 'telita': 'Telița'}
NBSP = ' '

def d2(v): return n(v, 2)
def ddm(iso): y, m, d = iso.split('-'); return f'{d}.{m}'

# icons not in the shell: package (delivery, from RecordTile preview = shell 'delivery'), Lucide coins (CostBlock README)
COINS = '<circle cx="8" cy="8" r="6"/><path d="M18.09 10.37A6 6 0 1 1 10.34 18"/><path d="M7 6h1v4"/><path d="m16.71 13.88.7.71-2.82 2.82"/>'
def rec_ico(path, size=16, rec='feed'):
    return f'<span class="ip-rec ip-rec--{rec} s11-ico"><svg class="ip-icon ip-icon--{size}" viewBox="0 0 24 24" aria-hidden="true">{path}</svg></span>'
PKG = rec_ico(s4.IC['delivery']); PKG14 = rec_ico(s4.IC['delivery'], 14); WHEAT14 = rec_ico(s4.IC['feed'], 14); COIN16 = rec_ico(COINS)

# ---------- data ----------
dels = [dict(d, farm_id='telita', priced_by=None, priced_at=None) for d in FD['rows']] + list(S['deliveries'])
dels.sort(key=lambda d: (d['date'], d['feed_type_id'] == 'ft-breeder2'), reverse=True)  # newest first; same day: Breeder 2 above Cocoși
N_ALL = len(dels); N_TEL = len(FD['rows']); N_HIR = len(S['deliveries'])
assert (N_ALL, N_TEL, N_HIR) == (76, 13, 63)
unpriced = [d for d in dels if not d['priced']]; UNP_KG = sum(d['qty_kg'] for d in unpriced)
assert len(unpriced) == 4 and UNP_KG == 76107
SHOWN = dels[:12]
EDIT_ID = 'del-hir-2026-11-14-b2'  # the owner's inline-edit row (newest unpriced)
# S11M (D69): the owner changes a saved price with a required Motiv. Row = the newest priced Breeder 2 delivery on the
# board (30.10.2026, 22 730 kg, 9,22 lei/kg, saved 03.11.2026). The new price on the error board is a labelled sample:
# the price of the Breeder 2 delivery before it (24.10.2026, 9,24 lei/kg), total derived from kg (craft §7, as D68).
MOD = next(d for d in SHOWN if d['priced'] and d['feed_type_id'] == 'ft-breeder2')
MOD_PREV = max((d for d in dels if d['priced'] and d['feed_type_id'] == 'ft-breeder2' and d['farm_id'] == 'hirbovat' and d['date'] < MOD['date']), key=lambda d: d['date'])
MOD_NEW = MOD_PREV['price_per_kg']; MOD_NEW_TOT = round(MOD['qty_kg'] * MOD_NEW)
assert (MOD['date'], MOD['qty_kg'], MOD['price_per_kg'], MOD['total_lei'], MOD['priced_at']) == ('2026-10-30', 22730, 9.22, 209571, '2026-11-03')
assert (MOD_PREV['date'], MOD_NEW, MOD_NEW_TOT) == ('2026-10-24', 9.24, 210025)
ALERT = '<circle cx="12" cy="12" r="10"/><path d="M12 8v4"/><path d="M12 16h.01"/>'  # Lucide circle-alert (field error line, shared with S12)
def ico(path, size=16): return f'<svg class="ip-icon ip-icon--{size}" viewBox="0 0 24 24" aria-hidden="true">{path}</svg>'

stocks = {s['feed_type_id']: s for s in S['stock_today']}
assert stocks['ft-breeder2']['stock_kg'] == 28252 and stocks['ft-cocosi']['days_left'] == 3.9

# consumption by cycle week × section × sex (Hîrbovăț, SAPT 37–40), D46: compare over the days that have rows
WEEKS = [40, 39, 38, 37]
SECS = ['H1.1', 'H1.2', 'H3.1', 'H3.2', 'H4.1', 'H4.2']
guide = {(r['cycle_week'], r['sex']): r['feed_g'] for r in GL['rows']}
agg = defaultdict(lambda: {'kg': 0.0, 'g': [], 'bird': 0, 'days': 0, 'miss': []})
wdates = defaultdict(set)
for r in S['daily']:
    if r['cycle_week'] not in WEEKS: continue
    wdates[r['cycle_week']].add(r['date'])
    for key in ((r['cycle_week'], r['section'], r['sex']), (r['cycle_week'], 'all', r['sex'])):
        a = agg[key]; a['days'] += 1
        if r['g_per_bird'] is None: a['miss'].append((r['section'], r['date'])); continue
        a['kg'] += r['kg']; a['g'].append(r['g_per_bird']); a['bird'] += r['alive_end']
def g_avg(a, weighted=False):  # section: mean of the daily grams; farm: kg ÷ bird-days (♀ by hens, ♂ by roosters, D38)
    return a['kg'] * 1000 / a['bird'] if weighted else sum(a['g']) / len(a['g'])
def dev(g, gd): return (g / gd - 1) * 100
def pct1(v):  # D46, no space; a value that rounds to 0 is written 0,0% (never −0,0%)
    v = round(v, 1)
    return s4.sgn(v, 1) + '%' if v != 0 else '0,0%'

# cost (CostBlock figures, recomputed and checked against the SYS04 preview)
tel_kg = sum(d['qty_kg'] for d in FD['rows']); tel_lei = sum(d['total_lei'] for d in FD['rows'])
hp = [d for d in S['deliveries'] if d['priced']]; hir_kg = sum(d['qty_kg'] for d in hp); hir_lei = sum(d['total_lei'] for d in hp)
tel_birds = PL['totals']['F'] + PL['totals']['M']
arr = {sx: sum(a[sx] for a in S['start_counts']) for sx in 'FM'}; hir_birds = arr['F'] + arr['M']
eggs = sum(r['total_eggs'] for r in S['eggs_daily'])
assert (tel_kg, tel_lei, hir_kg, hir_lei, len(hp), tel_birds, hir_birds, eggs) == (263740, 2031768, 1143540, 10417057, 59, 31590, 29825, 5301053)
assert (arr['F'], arr['M']) == (26765, 3060)

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
<div class="ip-row s4-meta"><span class="ip-note">Seria 2026 · Hîrbovăț <span class="ip-num">SAPT 40</span>, ciclu din <span class="ip-num">10.02.2026</span> · Telița fără ciclu deschis din <span class="ip-num">28.02.2026</span></span></div>
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
    if mode == 'empty':  # Telița, Interval 01.03–16.11.2026: after its cycle closed (28.02.2026), no delivery and no feed row
        itv = f'<span class="ip-interval" role="group" aria-label="Interval">{s4.datef("iv-a", "de la", "01.03.2026")}{s4.datef("iv-b", "până la", "16.11.2026")}</span>'
        return f'<div class="ip-filterbar s11-filterbar" role="group" aria-label="Filtre">{s4.seg("Perioadă", s4.PERIODS, 3)}{itv}{farm_select("Telița")}{export(True)}</div>'
    return f'<div class="ip-filterbar s11-filterbar" role="group" aria-label="Filtre">{s4.seg("Perioadă", s4.PERIODS, 0)}{farm_select("Toate fermele")}{export()}</div>'

def sec_head(hid, title, meta='', right=''):
    m = f'<span class="ip-note">{meta}</span>' if meta else ''
    return f'<div class="s4-head"><div class="s11-headl"><h2 class="ip-t-heading" id="{hid}">{title}</h2>{m}</div>{right}</div>'

def tscroll(cap, inner, fit=False):
    """S11X: a table that scrolls sideways inside its own box wherever the content area is narrower than the table (Consum
    1113 px; Livrări beside S13's SidePanel), as a focusable region named by its caption (WCAG 1.4.10, 2.1.1). One
    implementation, also used by S13 (was S13P's .s13-tscroll)."""
    f = ' s11-tablebox--fit' if fit else ''
    return (f'<div class="s11-tablebox s11-tscroll{f}" role="region" tabindex="0" aria-label="{cap}">'
            f'<table class="ip-table s11-table"><caption class="ip-sr">{cap}</caption>{inner}</table></div>')

# ---------- 2. Stoc ----------
def stock_ro(ft, sr):
    s = stocks[ft]; name = FT[ft]['name'].replace(' ', NBSP); ok = s['days_left'] >= 5
    rl = '<span class="ip-rlamp ip-rlamp--ok">în normă</span>' if ok else '<span class="ip-rlamp ip-rlamp--warn">abatere</span>'
    sub = '' if ok else '<p class="ip-readout__sub s11-m0">sub pragul de 5 zile</p>'
    w0, w1 = s['avg_window']
    return f'''<li><div class="ip-readout ip-readout--stock" role="group" aria-labelledby="sr-{ft}"><div class="ip-readout__cap"><span id="sr-{ft}">Hîrbovăț · {name}<span class="ip-sr">, {sr}</span></span>{rl}</div><div class="ip-readout__val">{n(s['stock_kg'])}<small> kg</small></div><dl class="ip-readout__kv"><div><dt>zile rămase</dt><dd>{n(s['days_left'], 1)}<small> zile</small></dd></div><div><dt>medie/zi {ddm(w0)}–{ddm(w1)}</dt><dd>{n(s['avg_daily_kg_last3'])}<small> kg</small></dd></div><div><dt>se termină</dt><dd>{ddate(s['run_out_date'])}</dd></div></dl>{sub}</div></li>'''
TELITA_RO = '<li><div class="ip-readout ip-readout--stock ip-readout--off" role="group" aria-labelledby="sr-telita"><div class="ip-readout__cap"><span id="sr-telita">Telița</span><span class="ip-rlamp ip-rlamp--off">oprit</span></div><div class="ip-readout__val">–</div><p class="ip-readout__sub s11-m0">fără ciclu deschis</p><p class="ip-readout__sub s11-m0">ultimul ciclu: Seria 2026, închis pe <span class="ip-num s11-nw">28.02.2026</span></p></div></li>'
STOCK_NOTE = '<p class="ip-note s11-m0">Stoc la <span class="ip-num">16.11.2026</span> = livrări − consum (rânduri confirmate). Zile rămase = stoc ÷ consumul mediu din ultimele 3 zile încheiate (<span class="ip-num">13.11–15.11.2026</span>); sub 5 zile = abatere.</p>'
def sec_stock(farm='all'):
    ros = TELITA_RO if farm == 'telita' else stock_ro('ft-breeder2', 'pentru găini') + stock_ro('ft-cocosi', 'pentru cocoși') + TELITA_RO
    return f'''<section class="s4-block" aria-labelledby="h-stoc">
{sec_head("h-stoc", "Stoc", 'pe fermă și tip de hrană · <span class="ip-num">16.11.2026</span>')}
<ul class="ip-readouts" aria-labelledby="h-stoc">{ros}</ul>
{'' if farm == 'telita' else STOCK_NOTE}
</section>'''

# ---------- 3. Livrări ----------
def mod_rows(d, head, lab, err):
    """S11M (D69): "Modifică" opens the S11 inline edit row under the saved price: the delivery row keeps its saved
    lei/kg and total (current price), then Preț nou (the same s11-in lei/kg ↔ total lei inputs), then the required
    Motiv field with the journal line, and Salvează prețul / Anulează. err: Salvează pressed with Motiv empty."""
    t = FT[d['feed_type_id']]; nm = t['name']; did = d['id']
    old = f'{d2(d["price_per_kg"])}{NBSP}lei/kg'
    cur = (f'<tr class="s11-editrow s11-modrow" id="modifica-pret-{did}">{head}<td class="num s11-pc">{d2(d["price_per_kg"])}</td><td class="num s11-pc">{n(d["total_lei"])}</td>'
           f'<td class="s11-muted">proprietar · <span class="ip-num">{ddate(d["priced_at"])}</span><span class="ip-sr">, preț salvat</span></td><td class="s11-act"></td></tr>')
    kg_v = f' value="{d2(MOD_NEW)}"' if err else ''
    tot = (f'<span class="ip-input s11-in s11-in--tot is-suggested"><input id="mp-tot" type="text" inputmode="decimal" value="{n(MOD_NEW_TOT)}" aria-label="Total lei nou, livrarea din {lab}" aria-describedby="mp-help mp-sample"/><span class="suffix">lei</span></span>' if err
           else f'<span class="ip-input s11-in s11-in--tot"><input id="mp-tot" type="text" inputmode="decimal" aria-label="Total lei nou, livrarea din {lab}" aria-describedby="mp-help"/><span class="suffix">lei</span></span>')
    new = (f'<tr class="s11-editrow s11-modrow"><td colspan="4"><div class="s11-modlbl"><span class="ip-label">Preț nou</span>'
           f'<span class="ip-hint" id="mp-help">scrie lei/kg sau totalul; celălalt se calculează din <span class="ip-num">{n(d["qty_kg"])}{NBSP}kg</span></span></div></td>'
           f'<td class="edit"><span class="ip-input s11-in{"" if err else " is-focus"}"><input id="mp-kg" type="text" inputmode="decimal"{kg_v} aria-label="Lei/kg nou, livrarea din {lab}" aria-describedby="mp-help{" mp-sample" if err else ""}"/><span class="suffix">lei/kg</span></span></td>'
           f'<td class="edit">{tot}</td><td colspan="2"></td></tr>')
    jur = f'Prețul vechi (<span class="ip-num">{old}</span>), cel nou și motivul rămân în jurnal, în Istoric modificări.'
    if err:
        msg = (f'{nm} <span class="ip-num">{ddate(d["date"])}</span>: scrie motivul ca să schimbi prețul din <span class="ip-num">{d2(d["price_per_kg"])}</span> '
               f'în <span class="ip-num">{d2(MOD_NEW)}{NBSP}lei/kg</span>')
        fld = (f'<span class="ip-input is-error is-focus"><input id="mp-mot" type="text" required="required" aria-required="true" aria-invalid="true" aria-describedby="mp-err mp-jur" placeholder="de ex. preț corectat pe factură"/></span>'
               f'<p class="ip-err s11-err" id="mp-err">{ico(ALERT, 14)}<span>{msg}</span></p>')
    else:
        fld = '<span class="ip-input"><input id="mp-mot" type="text" required="required" aria-required="true" aria-describedby="mp-jur" placeholder="de ex. preț corectat pe factură"/></span>'
    why = (f'<tr class="s11-editrow s11-modwhy"><td colspan="6"><div class="ip-field s11-motiv"><label class="ip-label" for="mp-mot">Motiv</label>{fld}'
           f'<p class="ip-hint s11-m0" id="mp-jur">{jur}</p></div></td>'
           f'<td colspan="2" class="s11-modact"><span class="s11-acts"><button type="button" class="ip-btn ip-btn--secondary">Salvează prețul</button><button type="button" class="ip-btn ip-btn--ghost">Anulează</button></span></td></tr>')
    return cur + new + why

def del_row(d, role, mode='default'):
    t = FT[d['feed_type_id']]; sx = {'F': ' · ' + F, 'M': ' · ' + M}.get(t['for_sex'], '')
    who = f'<td class="num">{n(d["qty_kg"])}</td>'
    lab = f'{ddate(d["date"])}, {n(d["qty_kg"])} kg {t["name"]}'
    head = f'<td class="num">{ddate(d["date"])}</td><td>{FARM[d["farm_id"]]}</td><td><span class="s11-type">{PKG}<span>{t["name"]}{sx}</span></span></td>{who}'
    owner = role == 'Proprietar'
    if role == 'Observator': return f'<tr>{head}</tr>'  # D64: viewers see no prices
    if owner and mode in ('mod', 'moderr') and d['id'] == MOD['id']: return mod_rows(d, head, lab, mode == 'moderr')
    if owner and mode == 'default' and d['id'] == EDIT_ID:  # inline edit (DataTable: medium TextField, lei/kg ↔ total)
        return f'''<tr class="s11-editrow">{head}<td class="edit"><span class="ip-input is-focus s11-in"><input id="pr-kg" type="text" inputmode="decimal" aria-label="Lei/kg, livrarea din {lab}" aria-describedby="pr-help"/><span class="suffix">lei/kg</span></span></td><td class="edit"><span class="ip-input s11-in s11-in--tot"><input id="pr-tot" type="text" inputmode="decimal" aria-label="Total lei, livrarea din {lab}" aria-describedby="pr-help"/><span class="suffix">lei</span></span></td><td colspan="2" class="edit"><span class="s11-acts"><button type="button" class="ip-btn ip-btn--secondary">Salvează prețul</button><button type="button" class="ip-btn ip-btn--ghost">Anulează</button></span></td></tr>'''
    if d['priced']:
        by = f'proprietar · <span class="ip-num">{ddate(d["priced_at"])}</span>' if d.get('priced_at') else '– din registru'
        price = f'<td class="num s11-pc">{d2(d["price_per_kg"])}</td><td class="num s11-pc">{n(d["total_lei"])}</td><td class="s11-muted">{by}</td>'
        act = f'<td class="s11-act"><a class="ip-link s11-lk" href="#modifica-pret-{d["id"]}" aria-label="Modifică prețul, livrarea din {lab}">Modifică</a></td>' if owner else ''
    else:
        price = '<td class="num s11-pc"><span class="ip-tag ip-tag--info">fără preț</span></td><td class="num s11-pc">–</td><td class="s11-muted">–</td>'
        act = f'<td class="s11-act"><a class="ip-link s11-lk" href="#pret-{d["id"]}" aria-label="Adaugă prețul, livrarea din {lab}">Adaugă preț</a></td>' if owner else ''
    return f'<tr>{head}{price}{act}</tr>'

def del_ths(role):
    ths = ('<th scope="col" class="num">Data</th><th scope="col">Fermă</th>'
           f'<th scope="col"><span class="ip-rec ip-rec--feed s11-thico">{icon("delivery", 14)}</span>Tip hrană</th><th scope="col" class="num">Kg</th>')
    if role != 'Observator': ths += '<th scope="col" class="num s11-thedit">Lei/kg</th><th scope="col" class="num s11-thedit">Total lei</th><th scope="col">Preț adăugat</th>'
    if role == 'Proprietar': ths += '<th scope="col"><span class="ip-sr">Acțiuni</span></th>'
    if role == 'Proprietar': ths = SKL.widths(ths, W_DEL, M_DEL)  # FIN11G: the loaded column shares, shared with the loading table
    return ths
# FIN11G (§54): column shares of the owner's tables at 1440 (as content-sized before), declared on the head the loading
# skeleton reuses, so its columns sit where the loaded ones do
W_DEL = [11.0, 8.2, 13.7, 7.7, 14.7, 17.3, None, 8.6]   # None: the column that takes the rest
W_CONS = [None, 6.4, 8.5, 5.8, 10.5, 8.5, 5.8, 10.5, 19.6]
W_COST = [None, 13.6, 14.7, 13.4, 17.9]
# FIN11G fix 1: each column's loaded min-content (px, without padding; measured with the table at its narrowest), so at
# 1280, where the content outgrows the shares, the loading columns grow as the loaded ones do
M_DEL = [78.28, 52.2, 103.25, 46.97, 116, 140, 150.08, 75.33]
M_CONS = [246.11, 46.97, 71.06, 40.91, 93.03, 71.06, 40.91, 93.03, 193.7]
M_COST = [256.53, 70.45, 78.28, 69.09, 100.11]

def sec_deliveries(role, mode='default'):
    owner = role == 'Proprietar'; viewer = role == 'Observator'
    ths = del_ths(role)
    rows = ''.join(del_row(d, role, mode) for d in SHOWN)
    unp_dates = ', '.join(ddm(d['date']) for d in sorted(unpriced, key=lambda d: d['date']))
    if owner:
        right = f'<a class="ip-link s11-s12 s11-lk" href="#preturi-hrana">{rec_ico(COINS, 14)}<span><span class="ip-num">{len(unpriced)}</span> livrări fără preț · Adaugă prețurile</span></a>'
    elif viewer:  # D64: the count only, no lei, no S12 link
        right = f'<span class="ip-note s11-s12"><span class="ip-num">{len(unpriced)}</span> livrări fără preț</span>'
    else:
        right = f'<span class="ip-note s11-s12">{rec_ico(COINS, 14)}<span><span class="ip-num">{len(unpriced)}</span> livrări fără preț · prețurile le adaugă proprietarul</span></span>'
    help_ = f'<p class="ip-sr" id="pr-help">Scrie lei/kg sau totalul în lei; celălalt se calculează din {n(24000)} kg.</p>' if owner and mode == 'default' else ''
    if mode == 'moderr':  # D68-style label for the typed sample price
        help_ = (f'<p class="ip-hint s11-m0" id="mp-sample">Valoare de exemplu: <span class="ip-num">{d2(MOD_NEW)}{NBSP}lei/kg</span> la <span class="ip-num">{ddate(MOD["date"])}</span> '
                 f'(prețul livrării Breeder 2 din <span class="ip-num">{ddate(MOD_PREV["date"])}</span>); totalul <span class="ip-num">{n(MOD_NEW_TOT)}{NBSP}lei</span> e calculat din <span class="ip-num">{n(MOD["qty_kg"])}{NBSP}kg</span>.</p>')
    unp = '' if viewer else f' · fără preț: <span class="ip-num">{unp_dates}.2026</span> · <span class="ip-num">{n(UNP_KG)} kg</span>'
    foot = f'<div class="s11-foot{' s11-tablebox--fit' if viewer else ''}"><span class="ip-note">Afișate <span class="ip-num">{len(SHOWN)}</span> din <span class="ip-num">{N_ALL}</span>{unp}</span><a class="ip-link s11-lk" href="#livrari-toate">Vezi toate (<span class="ip-num">{N_ALL}</span>)</a></div>'
    ro = ' · prețuri doar de citit' if role == 'Manager' else ''
    return f'''<section class="s4-block" aria-labelledby="h-livrari">
{sec_head("h-livrari", "Livrări", f'<span class="ip-num">{N_ALL}</span> livrări în Seria 2026 (Hîrbovăț <span class="ip-num">{N_HIR}</span> · Telița <span class="ip-num">{N_TEL}</span>) · cele mai noi întâi{ro}', right)}
{tscroll(f'Livrări hrană, cele mai noi {len(SHOWN)} din {N_ALL}', f'<thead><tr>{ths}</tr></thead><tbody>{rows}</tbody>', viewer)}
{help_}{foot}
</section>'''

# ---------- 4. Consum ----------
def cons_cells(w, key, weighted):
    out = ''; dvs = []
    for sx in 'FM':
        a = agg[(w, key, sx)]; g = g_avg(a, weighted); gd = guide[(w, sx)]; dv = dev(g, gd); dvs.append(dv)
        cls = 'hen' if sx == 'F' else 'rooster'
        out += f'<td class="num s11-{cls}">{n(g, 1)}</td><td class="num s11-muted">{gd}</td><td class="num s11-muted">{pct1(dv)}</td>'
    return out, dvs
def cons_state(w, key):
    miss = [(sx, d) for sx in 'FM' for (_, d) in agg[(w, key, sx)]['miss']]
    _, dvs = cons_cells(w, key, key == 'all')
    if miss:
        if key != 'all': what = ' · '.join(f'{F if sx == "F" else M} <span class="ip-num">{ddate(d)}</span>' for sx, d in miss)
        else:
            ds = sorted({d for _, d in miss}); span = ddate(ds[0]) if len(ds) == 1 else f'{ddm(ds[0])}–{ddate(ds[-1])}'
            what = f'<span class="ip-num">{len(miss)}</span> rânduri · <span class="ip-num">{span}</span>'
        return f'<span class="s11-state">{lamp("warn", "lipsă")}<span class="s11-muted">{what}</span></span>'
    if any(abs(v) > 10 for v in dvs): return lamp('warn', 'abatere')
    return lamp('ok', 'în normă')
CONS_THS = ('<th scope="col">Secțiune</th>'
           f'<th scope="col" class="num"><span class="ip-rec ip-rec--feed s11-thico">{icon("feed", 14)}</span>Kg</th>'
           f'<th scope="col" class="num">{F} g/pasăre</th><th scope="col" class="num">{F} ghid</th><th scope="col" class="num">{F} față de ghid</th>'
           f'<th scope="col" class="num">{M} g/pasăre</th><th scope="col" class="num">{M} ghid</th><th scope="col" class="num">{M} față de ghid</th><th scope="col">Stare</th>')
CONS_THS = SKL.widths(CONS_THS, W_CONS, M_CONS)
def sec_consum():
    rows = ''
    for w in WEEKS:
        ds = sorted(wdates[w]); kg = agg[(w, 'all', 'F')]['kg'] + agg[(w, 'all', 'M')]['kg']
        cur = ' · în curs' if TODAY in ds else ''
        cells, _ = cons_cells(w, 'all', True)
        rows += f'<tbody><tr class="is-group"><th scope="rowgroup" class="s11-gl"><span class="ip-num">SAPT {w}</span> <span class="s11-gsub">{ddm(ds[0])}–{ddate(ds[-1])}{cur}</span></th><td class="num">{n(kg)}</td>{cells}<td>{cons_state(w, "all")}</td></tr>'
        for sec in SECS:
            kg = agg[(w, sec, 'F')]['kg'] + agg[(w, sec, 'M')]['kg']; cells, _ = cons_cells(w, sec, False)
            rows += f'<tr><th scope="row" class="s11-rh">{sec}<span class="ip-sr">, SAPT {w}</span></th><td class="num">{n(kg)}</td>{cells}<td>{cons_state(w, sec)}</td></tr>'
        rows += '</tbody>'
    ths = CONS_THS
    foot = '<div class="s11-foot"><span class="ip-note">g/pasăre = media zilelor cu rânduri; rândul săptămânii cântărește ♀ după găini și ♂ după cocoși. Toleranță <span class="ip-num">±10%</span>. Telița: fără consum în ciclul curent (ciclu închis pe <span class="ip-num">28.02.2026</span>).</span><a class="ip-link s11-lk" href="#consum-toate">Vezi toate săptămânile</a></div>'
    return f'''<section class="s4-block" aria-labelledby="h-consum">
{sec_head("h-consum", "Consum", 'Hîrbovăț · pe săptămână și secțiune · <span class="ip-num">SAPT 37–40</span> · ghid Ross 308 ouat')}
{tscroll('Consum Hîrbovăț pe săptămână și secțiune, SAPT 37–40, față de ghid', f'<thead><tr>{ths}</tr></thead>{rows}')}
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
<tbody><tr><th scope="row">Telița · Seria 2026<span class="ip-cost__sub"><span class="ip-num">07.10.2025</span>–<span class="ip-num">28.02.2026</span> · <span class="ip-num">{N_TEL}</span> livrări cu preț</span></th><td class="num">{n(tel_kg)}</td><td class="num">{n(tel_lei)}</td><td class="num">{d2(tel_lei / tel_birds)}</td><td class="num">–<small>creștere, fără ouă</small></td></tr>{unp_row("Telița · Seria 2026", 0, "Toate livrările au preț.", False)}</tbody>
<tbody><tr><th scope="row">Hîrbovăț · Seria 2026<span class="ip-cost__sub">din <span class="ip-num">10.02.2026</span> · <span class="ip-num">SAPT 40</span> · <span class="ip-num">{len(hp)}</span> livrări cu preț</span></th><td class="num">{n(hir_kg)}</td><td class="num">{n(hir_lei)}</td><td class="num">{d2(hir_lei / hir_birds)}</td><td class="num">{d2(hir_lei / eggs)}</td></tr>{unp_row("Hîrbovăț · Seria 2026", UNP_KG, f'<span class="ip-num">{len(unpriced)}</span> livrări, <span class="ip-num">{ddm(d0)}–{ddate(d1)}</span> · neincluse în cost')}</tbody>
<tbody><tr class="is-total"><th scope="row">Total Seria 2026<span class="ip-cost__sub">nu include <span class="ip-num">{n(UNP_KG)} kg</span> fără preț</span></th><td class="num">{n(tel_kg + hir_kg)}</td><td class="num">{n(tel_lei + hir_lei)}</td><td class="num">{d2((tel_lei + hir_lei) / tel_birds)}</td><td class="num">{d2((tel_lei + hir_lei) / eggs)}</td></tr></tbody>'''
    foot = (f'<p class="ip-note ip-cost__foot">Lei = suma livrărilor cu preț; kg fără preț nu intră în cost. Lei/pasăre: la păsările de la începutul ciclului (Telița: populare <span class="ip-num">07.10.2025</span>, '
            f'<span class="ip-sex ip-sex--hen">♀ {n(PL["totals"]["F"])}</span> · <span class="ip-sex ip-sex--rooster">♂ {n(PL["totals"]["M"])}</span>; Hîrbovăț: sosite <span class="ip-num">22.02–28.02.2026</span>, '
            f'<span class="ip-sex ip-sex--hen">♀ {n(arr["F"])}</span> · <span class="ip-sex ip-sex--rooster">♂ {n(arr["M"])}</span>; seria: populare). Lei/ou: ouă totale (Cat. 1–3 + Incubabile) până pe <span class="ip-num">16.11.2026</span>: <span class="ip-num">{n(eggs)}</span>.</p>')
    return f'''<section class="s4-block" aria-labelledby="h-cost">
{sec_head("h-cost", "Cost", 'pe ciclu, pe pasăre și pe ou · din livrările cu preț')}
<section class="ip-cost ip-cost--wide" aria-labelledby="cost-d"><div class="ip-cost__head"><h3 class="ip-cap" id="cost-d">Cost hrană · Seria 2026</h3><span class="ip-note">din livrările cu preț · până pe <span class="ip-num">16.11.2026</span></span></div><div class="ip-cost__scroll"><table class="ip-table ip-cost__table"><caption class="ip-sr">Cost hrană · Seria 2026</caption>{t}</table></div>{foot}</section>
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
        if cost: return f'<table class="ip-table s11-table {cost}"><thead><tr>{ths}</tr></thead><tbody>{tr}</tbody></table>'
        return f'<div class="s11-tablebox s11-tscroll"><table class="ip-table s11-table"><thead><tr>{ths}</tr></thead><tbody>{tr}</tbody></table></div>'  # aria-hidden: scrolls, not a tab stop
    return f'''<div class="s11-page" aria-busy="true">
<p class="ip-sr" role="status">Se încarcă raportul Hrană pentru Seria 2026.</p>
<section class="s4-block" aria-hidden="true">{sec_head("h-stoc", "Stoc")}<ul class="ip-readouts">{ro * 3}</ul></section>
<section class="s4-block" aria-hidden="true">{sec_head("h-livrari", "Livrări")}{tbl(del_ths("Proprietar"), 6)}</section>
<section class="s4-block" aria-hidden="true">{sec_head("h-consum", "Consum")}{tbl(CONS_THS, 4)}</section>
<section class="s4-block" aria-hidden="true">{sec_head("h-cost", "Cost")}<div class="ip-cost ip-cost--wide"><div class="ip-cost__head">{sk("s4s-w40")}</div><div class="ip-cost__scroll">{tbl(COST_THS, 5, "ip-cost__table")}</div></div></section>
</div>'''

def sec_empty():
    tl = [d for d in FD['rows']][-1]
    assert tl['date'] == '2026-02-10' and all(d['date'] < '2026-03-01' for d in FD['rows'])
    TD = json.load(open(EX / 'feed-daily-telita.json', encoding='utf-8'))
    last_feed = max(r['date'] for r in TD['rows']) if 'rows' in TD else TD['meta']['range'][1]
    assert last_feed < '2026-03-01', last_feed
    def box(hid, lines, step):
        body = ''.join(f'<p class="s4s-empty-p">{l}</p>' for l in lines)
        return f'<div class="s4s-emptybox"><div class="ip-empty s11-empty">{lamp("off", "fără date")}{body}{step}</div></div>'
    itv = '<span class="ip-num">01.03.2026</span> și <span class="ip-num">16.11.2026</span>'
    IV_NOTE = 'Telița · <span class="ip-num">01.03–16.11.2026</span>'
    step_all = '<button type="button" class="ip-btn ip-btn--secondary">Vezi toate fermele</button>'
    return f'''<div class="s11-page">
{sec_stock('telita')}
<section class="s4-block" aria-labelledby="h-livrari">{sec_head("h-livrari", "Livrări", IV_NOTE)}{box("h-livrari", [f"Nicio livrare la Telița între {itv}.", f'Ultima livrare: <span class="ip-num">{ddate(tl["date"])}</span> · <span class="ip-num">{n(tl["qty_kg"])} kg</span> {FT[tl["feed_type_id"]]["name"]}.'], step_all)}</section>
<section class="s4-block" aria-labelledby="h-consum">{sec_head("h-consum", "Consum", IV_NOTE)}{box("h-consum", [f"Niciun consum la Telița între {itv}.", 'Ciclul Seria 2026 s-a încheiat pe <span class="ip-num">28.02.2026</span>.'], '<a class="ip-link s11-lk" href="#telita-seria-2026">Vezi ciclul Seria 2026 la Telița</a>')}</section>
<section class="s4-block" aria-labelledby="h-cost">{sec_head("h-cost", "Cost", IV_NOTE)}{box("h-cost", [f"Nicio livrare cu preț între {itv}, deci niciun cost în interval."], '<a class="ip-link s11-lk" href="#cost-seria-2026">Vezi costul Seria 2026</a>')}</section>
</div>'''

def page(role='Proprietar', mode='default'):
    if mode == 'loading': body = sec_loading()
    elif mode == 'empty': body = sec_empty()
    else: body = f'<div class="s11-page">{sec_stock()}{sec_deliveries(role, mode)}{sec_consum()}{"" if role == "Observator" else sec_cost()}</div>'
    h = header(role)
    return f'<div class="s4-shell">{sidebar(role)}<main class="s4-main">\n{h}\n{filterbar("empty" if mode == "empty" else "default")}\n{body}\n</main></div>'

CSS = s4.CSS + '\n' + s4.CSS_STATES + '''
.s11-page{display:flex;flex-direction:column;gap:var(--space-8)}
.s11-headl{display:flex;align-items:baseline;gap:var(--space-3);flex-wrap:wrap}
.s11-m0{margin:0}
.s11-nw{white-space:nowrap}
.s11-farm{display:inline-flex;align-items:center;gap:var(--space-2)}
.s11-farm .ip-label{margin:0}
.s11-select{width:calc(var(--space-12) * 4);cursor:pointer}
.s11-filterbar:has(> .ip-interval) .s11-farm{flex-direction:column;align-items:flex-start;gap:var(--space-1)}
.s11-tablebox{border:var(--ip-hair) solid var(--line);border-radius:var(--radius-md);overflow:hidden}
.s11-tscroll{overflow-x:auto} /* S11X: scrolls inside its box where the content is narrower than the table (role=region, tabindex=0) */
.s11-tscroll:focus-visible{outline:var(--ip-focus-w) solid var(--focus);outline-offset:var(--ip-focus-offset)}
.s11-tablebox--fit{width:fit-content;min-width:calc(var(--space-12) * 12);max-width:100%} /* viewer Livrări: 4 columns stay together (craft §5) */
.s11-tablebox--fit > .ip-table{width:100%}
.s11-table > tbody:last-child > tr:last-child > *{border-bottom:0}
.s11-table th[scope=row],.s11-table th[scope=rowgroup]{font:400 var(--ip-fs-table)/var(--ip-lh-table) var(--font-sans);text-transform:none;letter-spacing:0;color:var(--ink);background:var(--raised);border-bottom:var(--ip-hair) solid var(--rule);padding:var(--space-1) var(--space-3);height:var(--row-h)}
.s11-table tr.is-group > th{background:var(--sunken);font-weight:500}
.s11-gsub{color:var(--ink-secondary);font-weight:400;font-family:var(--font-mono);font-size:var(--ip-fs-small)}
.s11-table th.s11-rh{padding-left:var(--space-6)}
.s11-ico{color:var(--c);display:inline-flex;flex:none}
.s11-thico{display:inline-flex;vertical-align:top;margin-right:var(--space-1);color:var(--c)} /* 14px icon on the 14px caption line (S3P) */
.s11-type{display:inline-flex;align-items:center;gap:var(--space-2)}
.s11-hen{color:var(--rec-hen)}
.s11-rooster{color:var(--rec-rooster)}
.s11-muted{color:var(--ink-muted)}
.s11-state{display:inline-flex;align-items:center;gap:var(--space-2)}
.s11-act{text-align:right;width:1%}
.s11-lk{display:inline-flex;align-items:center;min-height:var(--space-6)}
.s11-in{width:calc(var(--space-12) * 2 + var(--space-6));margin-left:auto;font-family:var(--font-mono)}
.s11-in > input{text-align:right;font-variant-numeric:tabular-nums}
.s11-in--tot{width:calc(var(--space-12) * 3)}
.s11-acts{display:flex;gap:var(--space-2);justify-content:flex-end}
.s11-editrow > td{background:var(--sunken)}
.s11-s12{display:inline-flex;align-items:center;gap:var(--space-2);white-space:nowrap}
.s11-foot > .ip-link{white-space:nowrap;flex:none}
.s11-foot{display:flex;justify-content:space-between;align-items:center;gap:var(--space-4)}
.s11-table td.num > .ip-skel{margin-left:auto}
.s11-empty{padding:var(--space-8) var(--space-4)}
.s11-table th.s11-thedit,.s11-table td.s11-pc{padding-right:var(--space-2)} /* Lei/kg + Total lei: header, figures and the inputs' right edge on one line (td.edit padding) */
.s11-err{margin:0;white-space:normal;align-items:flex-start} /* field error line (S11M Motiv, S12 prices) */
.s11-err > .ip-icon{flex:none;margin-top:var(--ip-hair)}
.s11-modrow > td{border-bottom:0} /* S11M: saved price · Preț nou · Motiv read as one open row */
.s11-modlbl{display:flex;justify-content:flex-end;align-items:baseline;gap:var(--space-2)}
.s11-modwhy > td{vertical-align:top;height:auto;white-space:normal;padding-top:var(--space-1);padding-bottom:var(--space-3)}
.s11-motiv{max-width:calc(var(--space-12) * 12)}
.s11-modwhy > td.s11-modact{padding-top:calc(var(--space-1) + var(--space-4) + var(--space-1));padding-right:var(--space-2)} /* keys on the Motiv field's line (label 16 + gap 4) */'''

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
