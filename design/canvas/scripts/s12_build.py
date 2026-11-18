"""S12 · Prețuri hrană (owner, desktop 1440): the unpriced deliveries, oldest first, with inline lei/kg ↔ total lei inputs
(the S11 inline price inputs, s11-in) and one Salvează prețurile for every edited row; plus state boards.
Reuses the S4 shell through s11_build.py (sidebar with Rapoarte current, icons, formats, board writer, S11 table CSS).
No price is invented: the inputs open empty (D64), so Salvează prețurile is disabled; the error board is the load failure.
The "în lucru" board (D68) types the last real Breeder 2 price (9,22 lei/kg, 30.10.2026, labelled sample) into the 04.11 row
with its total derived from kg, and shows the field error on the 06.11 row (0 lei/kg) with focus on it.
python3 -I s12_build.py <repo_root> [heights.json, default s12_heights.json]
Writes design/canvas/project/S12-preturi-hrana[--<state>].dc.html and /tmp/s4s/<stem>.html (plain wrappers)."""
import sys, json, pathlib, importlib.util

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
ROOT, HP, CANVAS = CR.args('s12_heights.json')  # heights default: s12_heights.json
sys.argv = sys.argv[:1] + [str(ROOT)]  # s11_build reads argv[2] as its own heights file
spec = importlib.util.spec_from_file_location('s11_build', pathlib.Path(__file__).resolve().with_name('s11_build.py'))
s11 = importlib.util.module_from_spec(spec); spec.loader.exec_module(s11)
HEIGHTS = json.loads(HP.read_text()) if HP and HP.exists() else {}

s4 = s11.s4
n, ddate, icon, lamp, F, M = s4.n, s4.ddate, s4.icon, s4.lamp, s4.F, s4.M
FT, FARM, NB = s11.FT, s11.FARM, s11.NBSP

# ---------- data (all from the extracts, via s11: Telița feed-deliveries.json + Hîrbovăț sample deliveries) ----------
UNP = sorted(s11.unpriced, key=lambda d: (d['date'], d['feed_type_id'] != 'ft-breeder2'))  # C6: oldest first
UNP_KG = sum(d['qty_kg'] for d in UNP)
assert [d['date'] for d in UNP] == ['2026-11-04', '2026-11-06', '2026-11-10', '2026-11-14'] and UNP_KG == 76107
ALL_KG = sum(d['qty_kg'] for d in s11.dels)
assert (s11.N_ALL, ALL_KG) == (76, 1483387)
B2_LAST = max((d for d in s11.dels if d['priced'] and d['feed_type_id'] == 'ft-breeder2' and d['farm_id'] == 'hirbovat'), key=lambda d: d['date'])
assert (B2_LAST['date'], B2_LAST['price_per_kg']) == ('2026-10-30', 9.22)
W_ROW, E_ROW = UNP[0], UNP[1]  # în lucru: 04.11 Breeder 2 priced (sample), 06.11 Cocoși with the field error
W_TOT = round(W_ROW['qty_kg'] * B2_LAST['price_per_kg'])
assert (W_ROW['date'], W_ROW['qty_kg'], W_TOT, E_ROW['date'], E_ROW['feed_type_id']) == ('2026-11-04', 24068, 221907, '2026-11-06', 'ft-cocosi')
ALERT, ico = s11.ALERT, s11.ico  # shared with S11 (Modifică preț, S11M)

# ---------- header ----------
def header(mode):
    if mode in ('default', 'work'):
        meta = f'<span class="ip-num">{len(UNP)}</span> livrări fără preț · <span class="ip-num">{n(UNP_KG)} kg</span> · toate la Hîrbovăț · cele mai vechi întâi'
    elif mode == 'done':
        meta = '<span class="ip-num">0</span> livrări fără preț'
    elif mode == 'loading':  # the count arrives with the list
        meta = '<span class="ip-num">–</span> livrări fără preț · se încarcă'
    else:  # error: the count comes from the same request, so it is unknown
        meta = 'livrări fără preț: – nu s-au încărcat'
    return f'''<header class="ip-pagehead">
<div>
<nav class="ip-crumbs" aria-label="Cale"><a href="#rapoarte">Rapoarte</a><span aria-hidden="true">›</span><a href="#rapoarte-hrana">Hrană</a><span aria-hidden="true">›</span><span aria-current="page">Prețuri</span></nav>
<h1>Prețuri hrană</h1>
<div class="ip-row ip-pagehead__meta s12-meta">{s11.rec_ico(s11.COINS, 14)}<span class="ip-note">{meta}</span></div>
</div>
</header>'''

# ---------- default: the price table ----------
def lab(d): return f'livrarea din {ddate(d["date"])}, {n(d["qty_kg"])} kg {FT[d["feed_type_id"]]["name"]}'
def err_text(d): return f'{FT[d["feed_type_id"]]["name"]} <span class="ip-num">{ddate(d["date"])}</span>: scrie un preț mai mare de <span class="ip-num">0{NB}lei/kg</span>'
def row(d, first, mode='default'):
    t = FT[d['feed_type_id']]; sx = {'F': ' · ' + F, 'M': ' · ' + M}.get(t['for_sex'], '')
    kg_cls, kg_val, kg_aria, tot_cls, tot_val, tr, after = '', '', '', '', '', '', ''
    if mode == 'default' and first: kg_cls = ' is-focus'
    if mode == 'work' and d is W_ROW:  # typed lei/kg (sample), total calculated from kg: suggested until edited
        kg_val = f' value="{n(B2_LAST["price_per_kg"], 2)}"'; tot_cls = ' is-suggested'; tot_val = f' value="{n(W_TOT)}"'
    if mode == 'work' and d is E_ROW:  # refused on save: 0 lei/kg; the first error takes focus, total stays empty
        kg_cls = ' is-error is-focus'; kg_val = ' value="0"'; kg_aria = f' aria-invalid="true"'; tr = ' class="s12-haserr"'
        after = (f'<tr class="s12-errrow"><td colspan="4"><span class="ip-sr">{lab(d)}</span></td><td colspan="2" class="edit">'
                 f'<p class="ip-err s11-err" id="err-{d["id"]}">{ico(ALERT, 14)}<span>{err_text(d)}</span></p></td></tr>')
    desc = f'err-{d["id"]} pr-help' if kg_aria else 'pr-help'
    return (f'<tr{tr}><td class="num">{ddate(d["date"])}</td><td>{FARM[d["farm_id"]]}</td>'
            f'<td><span class="s11-type">{s11.PKG}<span>{t["name"]}{sx}</span></span></td><td class="num">{n(d["qty_kg"])}</td>'
            f'<td class="edit"><span class="ip-input s11-in{kg_cls}"><input id="kg-{d["id"]}" type="text" inputmode="decimal"{kg_val} aria-label="Lei/kg, {lab(d)}"{kg_aria} aria-describedby="{desc}"/><span class="suffix">lei/kg</span></span></td>'
            f'<td class="edit"><span class="ip-input s11-in s11-in--tot{tot_cls}"><input id="tot-{d["id"]}" type="text" inputmode="decimal"{tot_val} aria-label="Total lei, {lab(d)}" aria-describedby="pr-help"/><span class="suffix">lei</span></span></td></tr>{after}')

THS = ('<th scope="col" class="num">Data</th><th scope="col">Fermă</th>'
       f'<th scope="col"><span class="ip-rec ip-rec--feed ip-table__ico">{icon("delivery", 14)}</span>Tip hrană</th><th scope="col" class="num">Kg</th>'
       f'<th scope="col" class="num s11-thedit"><span class="ip-rec ip-rec--feed ip-table__ico">{ico(s11.COINS, 14)}</span>Lei/kg</th><th scope="col" class="num s11-thedit">Total lei</th>')

# FIN11G (§54): the four content columns declare their loaded shares (1440) next to the two fixed edit columns, so the
# loading table (same colgroup) keeps every column where the loaded one has it
COLS = ('<colgroup><col style="width:16.7%"/><col style="width:12.4%"/><col style="width:20.7%"/><col style="width:11.6%"/>'
        '<col class="s12-col-pk"/><col class="s12-col-tot"/></colgroup>')

def form(mode='default'):
    work = mode == 'work'
    rows = ''.join(row(d, i == 0, mode) for i, d in enumerate(UNP))
    lei = (f'<td class="num" colspan="2">{n(W_TOT)} <span class="s12-why ip-muted">· <span class="ip-num">{n(UNP_KG - W_ROW["qty_kg"])} kg</span> fără preț</span></td>' if work
           else '<td class="num ip-muted" colspan="2">– <span class="s12-why">niciun preț scris</span></td>')
    total = (f'<tfoot><tr class="is-total"><th scope="row" colspan="3">Total · <span class="ip-num">{len(UNP)}</span> livrări</th>'
             f'<td class="num">{n(UNP_KG)}</td>{lei}</tr></tfoot>')
    if work:
        foot = (f'<span class="ip-note" id="save-why"><span class="ip-num">1</span> din <span class="ip-num">{len(UNP)}</span> livrări cu preț scris · <span class="ip-num">1</span> preț de corectat</span>'
                '<button type="submit" class="ip-btn ip-btn--primary" aria-describedby="save-why">Salvează prețurile</button>')
        sample = (f'<p class="ip-hint s12-help">Valoare de exemplu: <span class="ip-num">{n(B2_LAST["price_per_kg"], 2)}{NB}lei/kg</span> la <span class="ip-num">{ddate(W_ROW["date"])}</span> '
                  f'(ultimul preț Breeder 2, din <span class="ip-num">{ddate(B2_LAST["date"])}</span>); totalul <span class="ip-num">{n(W_TOT)}{NB}lei</span> e calculat din <span class="ip-num">{n(W_ROW["qty_kg"])}{NB}kg</span>.</p>')
    else:
        foot = (f'<span class="ip-note" id="save-why"><span class="ip-num">0</span> din <span class="ip-num">{len(UNP)}</span> livrări cu preț scris · scrie un preț ca să salvezi</span>'
                '<button type="submit" class="ip-btn ip-btn--primary" disabled="disabled" aria-describedby="save-why">Salvează prețurile</button>')
        sample = ''
    return f'''<form class="s12-form" aria-labelledby="h-unp" novalidate="novalidate">
<div class="ip-sechead"><div class="s11-headl"><h2 class="ip-t-heading" id="h-unp">Livrări fără preț</h2></div></div>
<p class="ip-note s12-help" id="pr-help">Scrie prețul pe kg sau totalul de pe factură; celălalt se calculează din kg. Poți completa doar o parte din rânduri. După salvare, un preț se modifică doar cu un motiv și rămâne în jurnal.</p>
<div class="ip-tablescroll ip-tablescroll--framed" role="region" tabindex="0" aria-label="Livrări fără preț, cele mai vechi întâi"><table class="ip-table ip-table--rowheads s11-table s12-table"><caption class="ip-sr">Livrări fără preț, cele mai vechi întâi</caption>{COLS}<thead><tr>{THS}</tr></thead><tbody>{rows}</tbody>{total}</table></div>
<div class="s12-foot">{foot}</div>{sample}
</form>'''

# ---------- state: loading (static sunken skeletons in the same 6 columns; nothing to save yet) ----------
def loading():
    sk = s11.sk
    cells = (f'<td class="num">{sk("s4s-w70")}</td><td>{sk("s4s-w70")}</td><td>{sk("s4s-w70")}</td><td class="num">{sk("s4s-w70")}</td>'
             f'<td class="edit">{sk()}</td><td class="edit">{sk()}</td>')  # input-sized bars: rows as tall as the default's
    rows = f'<tr>{cells}</tr>' * len(UNP)
    return f'''<form class="s12-form" aria-labelledby="h-unp" aria-busy="true" novalidate="novalidate">
<div class="ip-sechead"><div class="s11-headl"><h2 class="ip-t-heading" id="h-unp">Livrări fără preț</h2></div></div>
<p class="ip-note s12-help" id="pr-help">Scrie prețul pe kg sau totalul de pe factură; celălalt se calculează din kg. Poți completa doar o parte din rânduri. După salvare, un preț se modifică doar cu un motiv și rămâne în jurnal.</p>
<p class="ip-sr" role="status">Se încarcă livrările fără preț.</p>
<div class="ip-tablescroll ip-tablescroll--framed" aria-hidden="true"><table class="ip-table ip-table--rowheads s11-table s12-table">{COLS}<thead><tr>{THS}</tr></thead><tbody>{rows}</tbody></table></div>
<div class="s12-foot"><span class="ip-note" id="save-why">Se încarcă livrările fără preț · poți scrie prețuri după încărcare</span><button type="submit" class="ip-btn ip-btn--primary" disabled="disabled" aria-describedby="save-why">Salvează prețurile</button></div>
</form>'''

# ---------- state: none unpriced (reached after saving the 4: toast) ----------
def done():
    return f'''<section class="s12-form" aria-labelledby="h-unp">
<div class="ip-sechead"><div class="s11-headl"><h2 class="ip-t-heading" id="h-unp">Livrări fără preț</h2></div></div>
<div class="ip-emptybox"><div class="ip-empty">{lamp("off", "nimic fără preț")}<p class="ip-empty__line">Toate livrările au preț.</p><p class="ip-empty__line">Costul Seria 2026 include toate cele <span class="ip-num">{s11.N_ALL}</span> livrări · <span class="ip-num">{n(ALL_KG)} kg</span>.</p><a class="ip-btn ip-btn--secondary" href="#rapoarte-hrana-cost">Vezi costul hranei</a></div></div>
</section>'''
# FIN02 (D122): bottom right like every desktop Toast (additions.css §48), with Anulează (owner, same day) as every save Toast
TOAST = (f'<div class="ip-toastwrap ip-toastwrap--page"><div class="ip-toast" role="status">{ico(s4.IC["check"])}<span class="ip-toast__txt">Prețuri salvate · <span class="ip-num">{len(UNP)}</span> livrări · '
         f'<span class="ip-num">{n(UNP_KG)} kg</span> intră în cost</span><button type="button" class="ip-link ip-toast__undo">Anulează</button></div></div>')

# ---------- state: error (the list did not load; nothing typed yet, so nothing is lost) ----------
def error():
    return f'''<section class="s12-form" aria-labelledby="h-unp">
<div class="ip-sechead"><div class="s11-headl"><h2 class="ip-t-heading" id="h-unp">Livrări fără preț</h2></div></div>
<div class="ip-banner ip-banner--fault ip-banner--act" role="alert">{ico(ALERT)}<div class="ip-banner__body"><p class="ip-m0">Nu am putut încărca livrările fără preț.</p><p class="ip-m0">Verifică conexiunea la internet, apoi reîncearcă. Prețurile salvate până acum nu s-au schimbat.</p></div><button type="button" class="ip-btn ip-btn--secondary ip-banner__key">Reîncearcă</button></div>
</section>'''

def page(mode='default'):
    body = form(mode) if mode in ('default', 'work') else {'done': done, 'error': error, 'loading': loading}[mode]()
    return f'<div class="ip-shell">{s11.sidebar("Proprietar")}<main class="ip-shell__main">\n{header(mode)}\n{body}\n</main></div>'

CSS = s11.CSS + '''
.s12-meta{align-items:center;gap:var(--space-2)}
.s12-table td.edit > .ip-skel{height:var(--control-h);border-radius:var(--radius-sm)} /* loading: the inputs' footprint, no row jump */
.s12-form{display:flex;flex-direction:column;gap:var(--space-3);width:fit-content;min-width:calc(var(--space-12) * 16);max-width:100%;margin:0}
.s12-help{margin:0;max-width:calc(var(--space-12) * 14)}
.s12-why{font-family:var(--font-sans);font-weight:400}
.s12-foot{display:flex;justify-content:space-between;align-items:center;gap:var(--space-4)}
.s12-foot > .ip-btn{flex:none}
.s12-table tr.s12-haserr > td{border-bottom:0} /* the field error row belongs to the row above */
.s12-table tr.s12-errrow > td{height:auto;padding-top:0;padding-bottom:var(--space-2)}
.s12-errrow .ip-err{contain:inline-size} /* s11-err line; wraps inside Lei/kg + Total lei, never widens them */
.s12-col-pk{width:calc(var(--space-12) * 2 + var(--space-6) + var(--space-4))} /* s11-in + td.edit padding: input fills the cell */
.s12-col-tot{width:calc(var(--space-12) * 3 + var(--space-4))}'''

STEM = 'S12-preturi-hrana'
BOARDS = [
 (STEM, 'S12 · Prețuri hrană', page(), ''),
 (STEM + '--se-incarca', 'S12 · Prețuri hrană · se încarcă', page('loading'), ''),
 (STEM + '--in-lucru', 'S12 · Prețuri hrană · în lucru', page('work'), ''),
 (STEM + '--toate-livrarile-au-pret', 'S12 · Prețuri hrană · toate livrările au preț', page('done'), TOAST),
 (STEM + '--eroare', 'S12 · Prețuri hrană · eroare la încărcare', page('error'), ''),
]
if __name__ == '__main__':
    for stem, title, body, ov in BOARDS:
        s4.write(stem, title, body, ov, H=HEIGHTS.get(stem, 1024), css=CSS)
