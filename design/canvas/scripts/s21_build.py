"""S21 · Tipuri de hrană (screens.md S21: list (name, for sex, active), add or edit; data-model FEED_TYPE; nav map: Setări ›
Tipuri de hrană, owner). Desktop-led like S18–S20 (a Setări page): desktops first, then the phone. Today = 16.11.2026.
Content, from the extracts only: feed-types.json (Breeder 2 ♀ and Cocoși ♂ active at Hîrbovăț; Starter, Grower, Pre-ouat
for both sexes, retired, Telița; names illustrative, D30); "in use" from feed-deliveries.json (every Telița type has
deliveries) and the extracts README (Hîrbovăț deliveries and stock are sample: Breeder 2 28 252 kg, Cocoși 1 650 kg).
D116: a used type can be retired but never deleted; retiring is refused while a farm still has it in stock (the vets feed
it daily); "Pentru" is fixed once a type is used; names unique ignoring case and diacritics (as D111/D114).
Boards (row below S20: y 35552 + 2380 + 120 = 38052; x step w + 80; desktops first, then the phone):
  S21-tipuri-hrana          list: PageHeader (Setări › Tipuri de hrană, "Adaugă tip"), DataTable (Tip hrană, Pentru, Stare
                            lamp, Folosit la, Editează)
  --breeder-2               list beside the non-modal SidePanel: Nume (focus), Pentru read-only (used), Activ ticked with the
                            stock rule, Folosit la + Șterge tipul aria-disabled with its reason
  --starter                 a retired Telița type: Activ unticked (tick to reactivate), Șterge aria-disabled
  --retragere-refuzata      D39: Activ unticked on Breeder 2 → refused (28 252 kg in stock), Salvează tipul aria-disabled
  --tip-nou                 empty "Tip nou" panel: Nume (focus), Pentru unset, Activ ticked, nefolosit
  --nume-gol                D39: Adaugă tipul with nothing filled: Nume (focus) + Pentru errors (§41)
  --nume-folosit            D39: "cocosi" = Cocoși ignoring case and diacritics
  --salvat                  D39: Breeder 3 (♀) added: panel closed, the list with it, Toast + Anulează, focus on its Editează
  --eroare-la-salvare       D39: the same new type, save failed: fault Banner + Reîncearcă (focus), draft kept
  --sterge                  D39: Breeder 3 unused → Șterge tipul live, its Dialog (focus on Anulează)
  --se-incarca              header kept (meta without counts), §42 skeleton rows
  --eroare-la-incarcare     fault Banner + Reîncearcă in place of the table, Adaugă tip kept
  --fara-acces              Observator (a manager or a vet: same picture, own nav and role sentence)
  --niciun-tip              D39 first use: EmptyState + "Adaugă primul tip" (no series yet)
  --retras-stoc-0           S21P, D39: Activ unticked on Breeder 2 with Hîrbovăț stock at 0 kg (Breeder 3 added): accepted
  --reactivat               S21P, D39: Starter active again: the list re-sorted, Toast + Anulează, focus on its Editează
  --telefon                 phone, read-only list ("le editezi pe calculator", as S19/S20)
The list region is a size container: under 760px (beside the panel: 558 at 1280, 718 at 1440) "Folosit la" folds under the
type name so Editează stays in view (S19P rule).
python3 -I s21_build.py <repo_root> [s21_heights.json] [--canvas]
Imports s19_build (rewrites the S19 boards and, through s18_build, S18/S17/S14S/S2/S7, byte-identical). Writes
design/canvas/project/S21-tipuri-hrana[--<state>].dc.html, /tmp/s21/<stem>.html, /tmp/s21/boards.json; --canvas adds the
row to canvas.json (nothing else changes). Renders: s21_render.py [--measure] [--w 1280] [stems]."""
import sys, re, json, pathlib, importlib.util, io, contextlib

ARGS = sys.argv[1:]
CANVAS = '--canvas' in ARGS; ARGS = [a for a in ARGS if a != '--canvas']
ROOT = pathlib.Path(ARGS[0]).resolve(); HERE = pathlib.Path(__file__).resolve().parent
HP = pathlib.Path(ARGS[1]) if len(ARGS) > 1 else HERE / 's21_heights.json'
HEIGHTS = json.loads(HP.read_text()) if HP.exists() else {}
sys.path.insert(0, str(HERE))
import skel as SKL
sys.argv = ['s19_build.py', str(ROOT)]
sp = importlib.util.spec_from_file_location('s19_build', HERE / 's19_build.py')
M = importlib.util.module_from_spec(sp)
with contextlib.redirect_stdout(io.StringIO()):
    sp.loader.exec_module(M)   # rewrites the S19 boards (and through s18_build S18/S17/S14S/S2/S7) byte-identical
T, SC, Q, s4, A = M.T, M.SC, M.Q, M.s4, M.A
icon, lamp, num, NB = M.icon, M.lamp, M.num, M.NB

# ================================================================= data (extracts only)
EX = ROOT / 'design/data/extracts'
FT = json.loads((EX / 'feed-types.json').read_text(encoding='utf-8'))['rows']
DEL = json.loads((EX / 'feed-deliveries.json').read_text(encoding='utf-8'))['rows']
README = (EX / 'README.md').read_text(encoding='utf-8')
TY = {r['id']: r for r in FT}
assert [r['name'] for r in FT] == ['Starter', 'Grower', 'Pre-ouat', 'Breeder 2', 'Cocoși']
assert {r['id'] for r in FT if r['active']} == {'ft-breeder2', 'ft-cocosi'}
assert TY['ft-breeder2']['for_sex'] == 'F' and TY['ft-cocosi']['for_sex'] == 'M' and all(TY[k]['for_sex'] is None for k in ('ft-starter', 'ft-grower', 'ft-preouat'))
USED_TEL = {d['feed_type_id'] for d in DEL if d['farm_id'] == 'telita'}
assert USED_TEL == {'ft-starter', 'ft-grower', 'ft-preouat'}   # real deliveries
# Hîrbovăț deliveries and stock are sample (README, D30): both active types are in use, stock on 16.11.2026
assert 'Breeder 2 28 252 kg' in README and 'Cocoși 1 650 kg' in README
STOCK = {'ft-breeder2': 28252, 'ft-cocosi': 1650}
FARM_OF = {r['id']: M.FARM[r['farm_id']]['name'] for r in FT}
assert FARM_OF['ft-breeder2'] == 'Hîrbovăț' and FARM_OF['ft-starter'] == 'Telița'
def n(v): return s4.n(v)
SEXW = {'F': ('♀', 'hen', 'găini', 'Găini'), 'M': ('♂', 'rooster', 'cocoși', 'Cocoși')}
def sexw(s, cap=False):
    if s is None: return 'Ambele sexe' if cap else 'ambele sexe'
    return f'<span class="ip-sex ip-sex--{SEXW[s][1]}">{SEXW[s][0]}</span> {SEXW[s][3] if cap else SEXW[s][2]}'
NEW = {'id': 'ft-new', 'name': 'Breeder 3', 'for_sex': 'F', 'active': True}   # D39: the new type (salvat, eroare, șterge)
def order(rows):   # active first, then retired; by name inside each group
    return sorted(rows, key=lambda r: (not r['active'], r['name']))
ORDER = order(FT)
assert [r['name'] for r in ORDER] == ['Breeder 2', 'Cocoși', 'Grower', 'Pre-ouat', 'Starter']

# ================================================================= list (desktop)
CRUMBS = ('<nav class="ip-crumbs" aria-label="Cale"><a href="#setari">Setări</a><span aria-hidden="true">›</span>'
          '<span aria-current="page">Tipuri de hrană</span></nav>')
def counts(rows):
    a = sum(1 for r in rows if r['active']); z = len(rows) - a
    return f'{num(a)} {"activ" if a == 1 else "active"}, {num(z)} {"retras" if z == 1 else "retrase"}'
def meta(rows=FT):
    tail = f'{NB}·{NB}{counts(rows)}' if rows else ''
    return f'<div class="ip-row ip-pagehead__meta"><span class="ip-note">Tipurile de hrană din livrări și din hrana zilnică{tail}</span></div>'
def newkey(pressed=False):
    p = ' is-pressed" aria-expanded="true" aria-controls="tp' if pressed else ''
    return f'<button type="button" class="ip-btn ip-btn--secondary{p}">{icon("plus", 16)}Adaugă tip</button>'   # FIN07 (D127): keys are verbs (craft §6); the panel keeps its title
def header(pressed=False, rows=FT, key=True):
    k = f'\n<div class="ip-row">{newkey(pressed)}</div>' if key else ''
    return f'<header class="ip-pagehead">\n<div>\n{CRUMBS}\n<h1>Tipuri de hrană</h1>\n{meta(rows)}\n</div>{k}\n</header>'
def sidebar(no_series=False):
    sb = M.sidebar_owner(no_series)
    for a, z in (('class="ip-nav ip-nav--sub is-current" href="#ferme" aria-current="page"', 'class="ip-nav ip-nav--sub" href="#ferme"'),
                 ('class="ip-nav ip-nav--sub" href="#tipuri-hrana"', 'class="ip-nav ip-nav--sub is-current" href="#tipuri-hrana" aria-current="page"')):
        assert sb.count(a) == 1, a; sb = sb.replace(a, z)
    assert sb.count('is-current') == 1
    return sb
def editkey(r, pressed=False, focus=False):
    # FIN05: the type open in the panel is the current row (additions §45/§51): aria-current on this key, is-current on
    # the row; the key keeps aria-expanded/-controls but is not drawn pressed (as S16/S19/S22)
    p = '" aria-expanded="true" aria-controls="tp" aria-current="true' if pressed else ''
    f = ' ip-focus' if focus else ''
    return f'<button type="button" class="ip-btn ip-btn--ghost{f}{p}">Editează<span class="ip-sr"> {r["name"]}</span></button>'
def state_cell(r): return lamp('ok', 'activ') if r['active'] else lamp('off', 'retras')
def use_txt(r): return FARM_OF.get(r['id']) or '<span class="ip-note">nefolosit</span>'
COLS = ('<thead><tr><th scope="col">Tip hrană</th><th scope="col">Pentru</th><th scope="col">Stare</th>'
        '<th scope="col" class="s21-cuse">Folosit la</th><th scope="col"><span class="ip-sr">Acțiuni</span></th></tr></thead>')
# FIN11G (§54): the desktop list's column shares at 1440 (as content-sized before) on the head the loading skeleton reuses;
# None = the column that takes the rest
COLS = SKL.widths(COLS, [None, 21.4, 15.1, 19.9, 24.5], [65.36, 48.42, 47.42, 70.09, 91.34])  # + loaded min-content px (fix 1)
def trow(r, open_=None, focus=None):
    return (f'<tr{" class=\"is-current\"" if open_ == r["id"] else ""}><th scope="row" class="s19-fname"><span class="s21-fn">{r["name"]}</span><span class="s21-fold">{use_txt(r)}</span></th>'
            f'<td class="ip-nowrap">{sexw(r["for_sex"])}</td><td class="ip-nowrap">{state_cell(r)}</td><td class="s21-cuse">{use_txt(r)}</td>'
            f'<td class="s19-act">{editkey(r, open_ == r["id"], focus == r["id"])}</td></tr>')
CAPTION = '<caption class="ip-sr">Tipurile de hrană, pentru cine sunt, starea lor și fermele care le folosesc</caption>'
NOTE = ('Un tip retras nu mai apare în Livrare hrană și Hrană zilnică; livrările lui rămân în rapoarte. '
        'Un tip folosit nu se poate șterge, doar retrage.')
def tlist(rows=ORDER, open_=None, focus=None, between=''):
    body = ''.join(trow(r, open_, focus) for r in rows)
    return (f'<section class="ip-block" aria-labelledby="h-tip"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-tip">Tipuri</h2>'
            f'<span class="ip-note">{num(len(rows))} tipuri</span></div>'
            f'<div class="ip-tablescroll ip-tablescroll--framed s21-region" tabindex="0" role="region" aria-labelledby="h-tip">'
            f'<table class="ip-table ip-table--rowheads s19-table s21-table">{CAPTION}{COLS}<tbody>{body}</tbody></table></div>'
            f'<p class="ip-note ip-m0">{NOTE}</p></section>')
def shell(main, hdr=None, sb=None, inert=False):
    i = ' inert="inert"' if inert else ''
    return f'<div class="ip-shell"{i}>{sb or sidebar()}<main class="ip-shell__main">\n{hdr or header()}\n{main}\n</main></div>'
def vp(open_=None, pressed_new=False, inert=False, rows=ORDER):
    return f'<div class="ip-dockpage s18-vp">{shell(tlist(rows, open_), header(pressed_new, rows), inert=inert)}</div>'

# ================================================================= type panel (SidePanel beside the page, as S19)
OPT = '<span class="s13-opt">(opțional)</span>'
ERRI = icon('alert', 14)
def err(i, text): return f'<p class="ip-err ip-fieldmsg" id="{i}">{ERRI}<span>{text}</span></p>'
def name_field(val, focus=False, e=None, used=False):
    h = 'Numele nou apare și în livrările și rapoartele de până acum.' if used and not e else ''
    f = T.tfield('tp-name', 'Nume', val, hint=h, touch=False, focus=focus)
    if e:   # the first error keeps the focus, the typed value stays (craft §5)
        a = '<input id="tp-name" type="text" autocomplete="off"'; assert f.count(a) == 1
        f = f.replace('class="ip-input is-focus">', 'class="ip-input is-focus is-error">').replace(a, a + ' aria-invalid="true" aria-describedby="tp-name-e"')
        f = f[:-6] + err('tp-name-e', e) + '</div>'
    return f
def for_ro(r):
    """Pentru once the type is used: the read-only TextField (as S8 Sex, S19 Tip): native label + input, the ♀/♂ glyph
    beside it in its colour and aria-hidden; the reason and the fix in the hint (D116)."""
    s = r['for_sex']
    g = f'<span class="ip-sex ip-sex--{SEXW[s][1]}" aria-hidden="true">{SEXW[s][0]}</span>' if s else ''
    v = SEXW[s][3] if s else 'Ambele sexe'
    fix = 'Pentru un singur sex, adaugă un tip nou.' if s is None else 'Pentru alt sex, adaugă un tip nou.'
    return (f'<div class="ip-field"><label class="ip-label" for="tp-for">Pentru</label>'
            f'<span class="ip-input is-readonly">{g}<input id="tp-for" type="text" readonly="readonly" value="{v}" aria-describedby="tp-for-h"/></span>'
            f'<span class="ip-hint" id="tp-for-h">Nu se mai schimbă: {r["name"]} are livrări la {FARM_OF[r["id"]]}. {fix}</span></div>')
def for_seg(sel='-', e=False):
    """Pentru on a new or unused type: SegmentedControl, required (sel '-' = unset)."""
    its = ''
    for i, k in enumerate((None, 'F', 'M')):
        on = sel == k; first = (sel == '-' and i == 0) or on
        its += (f'<span class="ip-seg__item{" is-selected" if on else ""}" role="radio" aria-checked="{"true" if on else "false"}" '
                f'tabindex="{0 if first else -1}">{sexw(k, True)}</span>')
    inv = ' aria-invalid="true" aria-describedby="tp-for-e"' if e else ' aria-describedby="tp-for-h"'
    msg = (err('tp-for-e', 'Alege pentru cine e hrana: ambele sexe, găini sau cocoși. În Hrană zilnică apare doar la sexul ales.') if e
           else '<span class="ip-hint" id="tp-for-h">În Hrană zilnică apare doar la sexul ales.</span>')
    return (f'<div class="ip-field"><span class="ip-label" id="tp-for-l">Pentru</span>'
            f'<span class="ip-seg s19-seg" role="radiogroup" aria-labelledby="tp-for-l" aria-required="true"{inv}>{its}</span>{msg}</div>')
def act_field(on, hint, e=None, focus=False):
    box = f'<span class="ip-check__box{" ip-focus" if focus else ""}">{icon("check", 14) if on else ""}</span>'
    d = 'tp-act-e' if e else 'tp-act-h'
    inv = ' aria-invalid="true"' if e else ''
    msg = err('tp-act-e', e) if e else f'<span class="ip-hint" id="tp-act-h">{hint}</span>'
    return (f'<div class="ip-field"><label class="ip-check{" is-checked" if on else ""}">'
            f'<input type="checkbox" class="ip-sr" id="tp-act"{" checked=\"checked\"" if on else ""} aria-describedby="{d}"{inv}/>{box}Activ</label>{msg}</div>')
ACT_ON = 'Apare în Livrare hrană și Hrană zilnică.'
def act_used(r):   # an active type with stock: the retire rule in the hint (D116)
    return f'{ACT_ON} Îl poți retrage când stocul de la {FARM_OF[r["id"]]} ajunge la 0 kg.'
ACT_OFF = 'Retras: nu apare în Livrare hrană și Hrană zilnică. Bifează Activ ca să-l folosești din nou.'
def use_block(r, del_state='used', pressed=False, stock=None):
    """Folosit la + Șterge tipul. used: aria-disabled with the reason; free: live (opens the Dialog)."""
    if del_state == 'free':
        line = f'Nefolosit: {r["name"]} nu are livrări și nici hrană zilnică.'
        p = ' is-pressed' if pressed else ''
        key = f'<button type="button" class="ip-btn ip-btn--destructive{p}" aria-haspopup="dialog">Șterge tipul</button>'
        why = ''
    else:
        f = FARM_OF[r['id']]
        if r['id'] in STOCK:
            kg = STOCK[r['id']] if stock is None else stock
            line = f'{f}: livrări și hrană zilnică în Seria 2026. Stoc acum: <span class="ip-num ip-nowrap">{n(kg)} kg</span>.'
        else:
            line = f'{f}: livrări în Seria 2026.'
        key = '<button type="button" class="ip-btn ip-btn--destructive" aria-disabled="true" aria-describedby="tp-del-why">Șterge tipul</button>'
        why = f'<p class="ip-note ip-m0" id="tp-del-why">Are livrări: rămâne în rapoarte, așa că nu se poate șterge, doar retrage.</p>'
    return (f'<section class="s21-use" aria-labelledby="tp-use-h"><h3 class="ip-cap ip-m0" id="tp-use-h">Folosit la</h3>'
            f'<p class="ip-m0">{line}</p><div class="s21-delkey">{key}</div>{why}</section>')
NEWUSE = ('<section class="s21-use" aria-labelledby="tp-use-h"><h3 class="ip-cap ip-m0" id="tp-use-h">Folosit la</h3>'
          '<p class="ip-m0">Nefolosit. Îl alegi în Livrare hrană și Hrană zilnică după ce îl adaugi.</p></section>')
def ctx(r):
    s = 'AMBELE SEXE' if r['for_sex'] is None else f'PENTRU {SEXW[r["for_sex"]][2].upper()}'
    u = FARM_OF.get(r['id'], 'nefolosit').upper()
    return f'{s} · {"ACTIV" if r["active"] else "RETRAS"} · {u}'
SAVE = '<button type="button" class="ip-btn ip-btn--primary">Salvează tipul</button>'
ADD = '<button type="button" class="ip-btn ip-btn--primary">Adaugă tipul</button>'
def panel(title, note, body, key=SAVE, inert=False):
    i = ' inert="inert"' if inert else ''
    nt = f'<p class="ip-note ip-num ip-m0">{note}</p>' if note else ''
    return (f'<aside class="ip-sidepanel ip-sidepanel--dock" id="tp" aria-labelledby="tp-t"{i}>\n<div class="ip-sidepanel__head"><div>'
            f'<h2 class="ip-t-heading ip-m0" id="tp-t">{title}</h2>{nt}</div>'
            f'<button type="button" class="ip-iconbtn" aria-label="Închide panoul tipului de hrană">{icon("x")}</button></div>\n'
            f'<div class="ip-sidepanel__body ip-sidepanel__body--form s19-pbody">\n{"".join(body)}\n</div>\n'
            f'<div class="ip-sidepanel__foot s18-foot"><div class="ip-row ip-sidepanel__keys"><button type="button" class="ip-btn ip-btn--secondary">Anulează</button>{key}</div></div>\n</aside>')
def tpanel(r, act, key=SAVE, focus_name=False, inert=False, pressed_del=False, stock=None):
    used = r['id'] in FARM_OF
    return panel(r['name'], ctx(r), [name_field(r['name'], focus_name, used=used), for_ro(r) if used else for_seg(r['for_sex']), act,
                                     use_block(r, 'used' if used else 'free', pressed_del, stock)], key, inert)
def save_banner(*paras):   # D147: the lead, then the retry line + what is kept
    body = ''.join(f'<p class="ip-m0">{t}</p>' for t in paras)
    return (f'<div class="ip-banner ip-banner--fault ip-banner--act" role="alert" id="tp-err">{icon("alert", 16)}<div class="ip-banner__body">'
            f'{body}</div><button type="button" class="ip-btn ip-btn--secondary ip-banner__key ip-focus">Reîncearcă</button></div>')

# ================================================================= states (list side)
def skel(w): return f'<span class="ip-skel ip-skel--{w}"></span>'
SK_R = ('<tr class="ip-table__skel"><th scope="row">' + skel('w2') + '</th>' + ''.join(f'<td class="{c}">{skel(w)}</td>' for c, w in (('', 'w2'), ('', 'w1'), ('s21-cuse', 'w2')))
        + '<td class="s19-act"><span class="ip-skel ip-skel--w3 ip-skel--end"></span></td></tr>')
LOADING = (f'<p class="ip-sr" role="status">Se încarcă tipurile de hrană, cu starea lor.</p>'
           f'<section class="ip-block" aria-labelledby="h-tip" aria-busy="true"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-tip">Tipuri</h2></div>'
           f'<div class="ip-tablescroll ip-tablescroll--framed s21-region" aria-hidden="true"><table class="ip-table ip-table--rowheads s19-table s21-table">'
           f'{COLS}<tbody>{SK_R}{SK_R}</tbody></table></div></section>')
LOADERR = (f'<section class="ip-block" aria-labelledby="h-tip"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-tip">Tipuri</h2></div>'
           f'<div class="ip-banner ip-banner--fault ip-banner--act" role="alert">{icon("alert", 16)}'
           f'<div class="ip-banner__body"><p class="ip-m0">Tipurile de hrană nu s-au încărcat.</p>'
           '<p class="ip-m0">Verifică conexiunea la internet, apoi apasă Reîncearcă. Poți adăuga un tip nou și acum.</p></div>'
           '<button type="button" class="ip-btn ip-btn--secondary ip-banner__key">Reîncearcă</button></div></section>')
def noaccess():
    step = f'<a class="ip-btn ip-btn--secondary" href="#acasa">{icon("home", 16)}Înapoi la Acasă</a>'
    box = s4.empty_box('fără acces', ['Nu ai acces la Tipuri de hrană.',
                                      'Tipurile de hrană le setează proprietarul. Rolul tău, Observator, vede paginile seriilor '
                                      'și ale fermelor, și rapoartele.'], step, 'Acces')
    return f'<div class="ip-shell">{Q.sidebar_viewer()}<main class="ip-shell__main">\n<header class="ip-pagehead"><div><h1>Tipuri de hrană</h1></div></header>\n{box}\n</main></div>'
NOTYPES = s4.empty_box('fără tipuri', ['Nu există niciun tip de hrană.', 'Adaugă primul tip, ca să poți înregistra livrările și hrana zilnică.'],
                       '<button type="button" class="ip-btn ip-btn--primary">Adaugă primul tip</button>', 'Tipuri')

# ================================================================= phone list (owner, read-only, as S19/S20)
def phone_list():
    rows = ''.join(f'<tr><th scope="row" class="s19-fname"><span class="s21-fn">{r["name"]}</span><span class="ip-note s21-pfarm">{use_txt(r)}</span></th>'
                   f'<td>{sexw(r["for_sex"])}</td><td class="ip-nowrap">{state_cell(r)}</td></tr>' for r in ORDER)
    blk = (f'<section class="ip-block" aria-labelledby="h-tip"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-tip">Tipuri</h2>'
           f'<span class="ip-note">{counts(FT)}</span></div>'
           f'<div class="ip-tablescroll ip-tablescroll--framed" tabindex="0" role="region" aria-labelledby="h-tip">'
           f'<table class="ip-table ip-table--rowheads s19-table s19-ptable s21-table">{CAPTION}'
           f'<thead><tr><th scope="col">Tip hrană</th><th scope="col">Pentru</th><th scope="col">Stare</th></tr></thead><tbody>{rows}</tbody></table></div>'
           f'<p class="ip-note ip-m0">Un tip retras nu mai apare în formulare. Tipurile de hrană le adaugi și le editezi pe calculator.</p></section>')
    cr = CRUMBS.replace('<a href="#setari">Setări</a>', '<a href="#setari"><span class="ip-crumbs__txt">Setări</span></a>').replace('<nav class="ip-crumbs"', '<nav class="ip-crumbs ip-crumbs--touch"', 1)  # FIN06: system touch crumb (section 52)
    body = A.phone([f'<header class="ip-pagehead s2-phead s14-phead"><div>{cr}<h1>Tipuri de hrană</h1></div></header>', blk])
    for a, z in (('class="ip-tabbar__item is-current" href="#acasa" aria-current="page"', 'class="ip-tabbar__item" href="#acasa"'),
                 ('class="ip-tabbar__item" href="#mai-mult"', 'class="ip-tabbar__item is-current" href="#mai-mult" aria-current="page"')):
        assert body.count(a) == 1, a; body = body.replace(a, z)
    assert body.count('is-current') == 1
    return body

# ================================================================= CSS (screen-only; tokens only; system classes untouched)
CSS21 = '''
.s21-table td{white-space:normal}
.s21-region{container:s21-tipuri / inline-size}
.s21-fold{display:none}
/* S19P rule: beside the panel the region is under 760px (558 at 1280, 718 at 1440), so "Folosit la" folds under the type
 * name and Editează stays in view; the list alone keeps all five columns. */
@container s21-tipuri (max-width: 759.98px){
  .s21-table th{width:auto} .s21-table th::after{content:none} /* FIN11G: beside the panel the columns are content-sized again (the §54 shares are the list alone's) */
  .s21-table .s21-cuse{display:none}
  .s21-fn{display:block}
  .s21-fold{display:block;font-weight:400;color:var(--ink-secondary)}
}
.s21-use{display:flex;flex-direction:column;align-items:flex-start;gap:var(--space-2);padding-top:var(--space-4);border-top:var(--ip-hair) solid var(--rule)}
.s21-delkey{margin-top:var(--space-1)}
.s21-pfarm{display:block;font-weight:400}'''
CSS_DESK = M.CSS_DESK + CSS21
CSS_PL = M.CSS_PL + CSS21

# ================================================================= write
pathlib.Path('/tmp/s21').mkdir(exist_ok=True)
def write(stem, title, root, css, w, h):
    r = M.write(stem, title, root, css, w, h)   # S19's writer: the .dc.html + /tmp/s19/<stem>.html; mirror into /tmp/s21
    src = pathlib.Path(f'/tmp/s19/{stem}.html'); (pathlib.Path('/tmp/s21') / src.name).write_text(src.read_text(encoding='utf-8'), encoding='utf-8'); src.unlink()
    return r
STEM, TT = 'S21-tipuri-hrana', 'S21 · Tipuri de hrană'
D39 = 'exemplu de stare (D39)'
PH = SC.SIDE_H
out = []
def desk(stem, title, root, h=None):
    h = h or HEIGHTS.get(STEM + stem, 1024)
    out.append(write(STEM + stem, f'{TT}{" · " + title if title else ""}', f'<div class="ip-root s4s-rel" style="padding:0; min-height:{h}px">{root}</div>', CSS_DESK, 1440, h))

B2, ST_ = TY['ft-breeder2'], TY['ft-starter']
# 1 · the list
desk('', '', shell(tlist()))
# 2 · Breeder 2 (active, used, in stock): Pentru read-only, retire rule, Șterge aria-disabled; focus on Nume
desk('--breeder-2', 'Breeder 2 · panoul tipului', vp('ft-breeder2') + tpanel(B2, act_field(True, act_used(B2)), focus_name=True), PH)
# 3 · Starter (retired, used at Telița): tick Activ to reactivate; Șterge aria-disabled
desk('--starter', 'Starter · tip retras', vp('ft-starter') + tpanel(ST_, act_field(False, ACT_OFF)), PH)
# 4 · retire refused (D39 action): Activ unticked on Breeder 2 while 28 252 kg are in stock
RET_E = (f'Breeder 2 nu se poate retrage: Hîrbovăț mai are <span class="ip-num ip-nowrap">{n(STOCK["ft-breeder2"])} kg</span> în stoc. '
         f'Bifează Activ la loc; îl retragi când stocul ajunge la 0 kg.')
SAVE_NO = '<button type="button" class="ip-btn ip-btn--primary" aria-disabled="true" aria-describedby="tp-act-e">Salvează tipul</button>'
desk('--retragere-refuzata', f'Breeder 2 · retragere refuzată: stoc la Hîrbovăț · {D39}: Activ debifat',
     vp('ft-breeder2') + tpanel(B2, act_field(False, '', RET_E, focus=True), key=SAVE_NO), PH)
# 5 · Tip nou: empty panel, Pentru unset, Activ ticked; focus on Nume
NEWVP = vp(pressed_new=True)
desk('--tip-nou', 'Tip nou', NEWVP + panel('Tip nou', '', [name_field('', True), for_seg(), act_field(True, ACT_ON + ' Îl poți retrage oricând.'), NEWUSE], ADD), PH)
# 6 · Nume gol + Pentru lipsă (D39): Adaugă tipul with nothing filled
desk('--nume-gol', f'Tip nou · nume gol și pentru lipsă · {D39}',
     NEWVP + panel('Tip nou', '', [name_field('', True, 'Scrie numele tipului, de exemplu Breeder 3.'), for_seg(e=True),
                                   act_field(True, ACT_ON + ' Îl poți retrage oricând.'), NEWUSE], ADD), PH)
# 7 · Nume deja folosit (D39): "cocosi" = Cocoși without case and diacritics (D116, as D111/D114)
desk('--nume-folosit', f'Tip nou · nume deja folosit · {D39}: „cocosi”',
     NEWVP + panel('Tip nou', '', [name_field('cocosi', True, 'Numele e deja folosit de tipul Cocoși (literele mari și diacriticele nu contează). '
                                                     'Alege alt nume, de exemplu Cocoși 2, sau editează Cocoși din listă.'),
                                   for_seg('M'), act_field(True, ACT_ON + ' Îl poți retrage oricând.'), NEWUSE], ADD), PH)
# 8 · salvat (D39): Breeder 3 (♀) added: panel closed, focus on its Editează, Toast with Anulează
ROWS3 = order(FT + [NEW])
FOCUS_NOTE = ('<!-- Focus after Adaugă tipul: the panel closes and focus goes to the new row\'s Editează key; the Toast (role=status) '
              'announces the save and keeps Anulează in the tab order. -->')
def toast(text, note=FOCUS_NOTE):
    return (f'{note}<div class="ip-toastwrap ip-toastwrap--page"><div class="ip-toast" role="status">{icon("check", 16)}<span class="ip-toast__txt">'
            f'{text}</span><button type="button" class="ip-link ip-toast__undo">Anulează</button></div></div>')
TOAST = toast('Breeder 3 adăugat · pentru găini, apare în Livrare hrană și Hrană zilnică')
desk('--salvat', f'salvat · {D39}: Breeder 3 adăugat', shell(tlist(ROWS3, focus='ft-new'), header(rows=ROWS3)) + TOAST, PH)
# 9 · eroare la salvare (D39): the same new type, draft kept, Banner + Reîncearcă (focus) under the fields
desk('--eroare-la-salvare', f'Tip nou · eroare la salvare · {D39}: Breeder 3',
     NEWVP + panel('Tip nou', '', [name_field('Breeder 3'), for_seg('F'), act_field(True, ACT_ON + ' Îl poți retrage oricând.'),
                                   save_banner('Breeder 3 nu s-a salvat: conexiunea s-a întrerupt.',
                                               'Verifică conexiunea la internet, apoi apasă Reîncearcă. '
                                               'Ce ai scris rămâne în panou.'), NEWUSE], ADD), PH)
# 10 · Șterge (D39): Breeder 3 unused → its Dialog over the inert page and panel, focus on Anulează
DLG = ('<div class="ip-scrim ip-scrim--page"><div class="ip-dialog" role="alertdialog" aria-modal="true" aria-labelledby="dlg-t" aria-describedby="dlg-n">'
       '<h2 id="dlg-t">Șterge tipul de hrană Breeder 3</h2>'
       '<p class="ip-dialog__note" id="dlg-n">Breeder 3 nu are livrări și nici hrană zilnică, așa că dispare din listă și din formulare. '
       'Dacă îl vrei înapoi, îl adaugi din nou cu Adaugă tip.</p>'
       '<div class="ip-dialog__keys"><button type="button" class="ip-btn ip-btn--secondary ip-focus" autofocus="autofocus">Anulează</button>'
       '<button type="button" class="ip-btn ip-btn--destructive">Șterge tipul</button></div></div></div>')
desk('--sterge', f'Breeder 3 · Șterge tipul · {D39}: tip nou, nefolosit',
     vp('ft-new', inert=True, rows=ROWS3) + tpanel(NEW, act_field(True, ACT_ON + ' Îl poți retrage oricând.'), inert=True, pressed_del=True) + DLG, PH)
# 11 · se încarcă
desk('--se-incarca', 'se încarcă', shell(LOADING, header(rows=None)))
# 12 · eroare la încărcare
desk('--eroare-la-incarcare', 'eroare la încărcare', shell(LOADERR, header(rows=None)))
# 13 · fără acces (Observator; a manager or a vet: same picture)
desk('--fara-acces', 'observator · fără acces', noaccess())
# 14 · niciun tip (D39 first use, no series yet)
desk('--niciun-tip', f'niciun tip · {D39}', shell(NOTYPES, header(rows=None, key=False), sidebar(no_series=True)))
# 15 · S21P, D39: retiring is accepted once the stock is 0 kg (Breeder 2 at Hîrbovăț used up, Breeder 3 added for the hens):
#      the counterpart of --retragere-refuzata, Activ unticked (focus) with no error, Salvează tipul live
RET_OK = ('Stoc <span class="ip-num ip-nowrap">0 kg</span> la Hîrbovăț, deci îl poți retrage. După ce salvezi nu mai apare în Livrare hrană și Hrană zilnică; '
          'livrările lui rămân în rapoarte.')
desk('--retras-stoc-0', f'Breeder 2 · retragere cu stoc 0 kg · {D39}: stoc Hîrbovăț 0 kg, Breeder 3 adăugat',
     vp('ft-breeder2', rows=ROWS3) + tpanel(B2, act_field(False, RET_OK, focus=True), stock=0), PH)
# 16 · S21P, D39: a retired type made active again (Starter): panel closed, the list re-sorted, focus on its Editează, Toast
ROWS_RE = order([{**r, 'active': True} if r['id'] == 'ft-starter' else r for r in FT])
assert [r['name'] for r in ROWS_RE] == ['Breeder 2', 'Cocoși', 'Starter', 'Grower', 'Pre-ouat']
RE_NOTE = FOCUS_NOTE.replace('Adaugă tipul', 'Salvează tipul').replace("the new row's", "the type's")
desk('--reactivat', f'Starter activ din nou · salvat · {D39}',
     shell(tlist(ROWS_RE, focus='ft-starter'), header(rows=ROWS_RE))
     + toast('Starter activ din nou · pentru ambele sexe, apare în Livrare hrană și Hrană zilnică', RE_NOTE), PH)
# 17 · phone, read-only list
h = HEIGHTS.get(STEM + '--telefon', 844)
out.append(write(STEM + '--telefon', f'{TT} · telefon · doar citire', f'<div class="ip-root" style="padding:0; width:390px; min-height:{h}px">{phone_list()}</div>', CSS_PL, 390, h))
pathlib.Path('/tmp/s21/boards.json').write_text(json.dumps([{k: b[k] for k in ('stem', 'w', 'h')} for b in out]))

if CANVAS:  # S21 row, entries keep x/y, a new row goes below the lowest (canvas_rows.py); 80 px gaps
    sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
    c = CR.load(ROOT)
    x = CR.place(c, STEM, [(b['stem'], b['w'], b['h'], b['title']) for b in out])
    CR.save(ROOT, c)
    print('canvas: S21 row y', CR.row_y(c, STEM), 'ends at x', x - 80)
