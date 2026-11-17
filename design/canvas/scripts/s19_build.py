"""S19 · Ferme și secțiuni (spec B1, screens.md S19; nav map: Setări › Ferme și secțiuni, owner). Desktop-led like S17/S18
(a Setări page): desktops first, then the phone. Today = 16.11.2026 (D23/D24).
Content, from the extracts only (farms-sections, cycles through s18_build): Telița creștere, guide Ross 308 creștere,
sections H1 H2 H3 H6 H5-1 H5-2, empty since its cycle closed 28.02.2026 (D25); Hîrbovăț ouat, guide Ross 308 ouat, halls
H1 H3 H4 with compartments H1.1 … H4.2, laying cycle open from 10.02.2026 (SAPT 40). Guide names illustrative (D22).
The farm panel's section tree is the SYS10 SectionTree (additions.css §28): its markup helpers are read from
design/system/scripts/sys10_build.py (the part above its preview blocks, so nothing is written there).
Boards (row below S18: y 32813 + 1624 + 120 = 34557; x step w + 80; D110):
  S19-ferme-sectiuni                         desktop list: PageHeader (Setări › Ferme și secțiuni, "Fermă nouă"), DataTable
                                             of farms (Tip, Ghid Tag, Secțiuni, Ciclu, Editează)
  --hirbovat                                 list beside the non-modal farm SidePanel (Hîrbovăț): Nume (focus), Tip read-only
                                             (cycle open), Ghid Select + next-cycle rule, SectionTree, Salvează ferma
  --telita                                   Telița panel (empty farm): Tip SegmentedControl, H5-2 menu open (Retrage live)
  --retragere-blocata                        Hîrbovăț panel, H1.1 menu open: Retrage aria-disabled with the B1 reason
  --retrage-sectiunea                        D39: Retrage H5-2 Dialog over the inert page + panel, focus on Anulează
  --fara-ghid                                D39/D51: Telița without a guide: Tag "fără ghid" in the list and the tree, Ghid
                                             "Fără ghid" with the actuals-only line (the open list is left to S19S)
  --ghid-nepotrivit                          D39: Telița Tip back to Creștere with Ross 308 ouat kept → the B1 sentence,
                                             Salvează ferma disabled
  --fara-ferme                               D39: no farm yet: EmptyState "fără ferme" + "Adaugă prima fermă"
  --ferma-noua                               the list beside an empty "Fermă nouă" panel (Tip unset, Ghid waits for it,
                                             sections after the first save)
  --telefon                                  phone, read-only list (D110: editing is desktop only, D75)
python3 -I s19_build.py <repo_root> [s19_heights.json] [--canvas]
Writes design/canvas/project/S19-ferme-sectiuni[--<state>].dc.html, /tmp/s19/<stem>.html, /tmp/s19/boards.json; --canvas
adds the boards to canvas.json (new row, nothing else changes). Renders: s19_render.py [--measure] [stems]."""
import sys, re, json, pathlib, importlib.util, io, contextlib

ARGS = sys.argv[1:]
CANVAS = '--canvas' in ARGS; ARGS = [a for a in ARGS if a != '--canvas']
ROOT = pathlib.Path(ARGS[0]).resolve(); HERE = pathlib.Path(__file__).resolve().parent
HP = pathlib.Path(ARGS[1]) if len(ARGS) > 1 else HERE / 's19_heights.json'
HEIGHTS = json.loads(HP.read_text()) if HP.exists() else {}
sys.path.insert(0, str(HERE))
sys.argv = ['s18_build.py', str(ROOT)]
sp = importlib.util.spec_from_file_location('s18_build', HERE / 's18_build.py')
T = importlib.util.module_from_spec(sp)
with contextlib.redirect_stdout(io.StringIO()):
    sp.loader.exec_module(T)   # rewrites the S18 (and through s17_build S17/S14S/S2/S7) boards byte-identical
SC, Q, s4, A = T.SC, T.Q, T.s4, T.A
icon, lamp, num, dd, NB = T.icon, T.lamp, T.num, T.dd, T.NB
cT, cH, WK, GUIDE, FARM, FS = T.cT, T.cH, T.WK, T.GUIDE, T.FARM, T.FS

# SectionTree helpers (SYS10): the code above the preview blocks only (no file is written)
SRC = (ROOT / 'design/system/scripts/sys10_build.py').read_text(encoding='utf-8')
CUT = '# ------------------------------------------------------------------ blocks'
assert SRC.count(CUT) == 1
ST = {'__name__': 'sys10_parts'}
exec(compile(SRC.split(CUT)[0], 'sys10_build.py', 'exec'), ST)
assert ST['DATA'] == FS and ST['BIRDS'] == {'hirbovat': True, 'telita': False}

# ================================================================= data (extracts only)
TYPE_RO = {'growth': 'Creștere', 'laying': 'Ouat'}
TYPE_LO = {'growth': 'creștere', 'laying': 'ouat'}
assert [f['id'] for f in FS['farms']] == ['telita', 'hirbovat']
assert FARM['telita']['type'] == 'growth' and FARM['hirbovat']['type'] == 'laying'
assert GUIDE == {'g-growth': 'Ross 308 creștere', 'g-laying': 'Ross 308 ouat'}
assert cH['start_date'] == '2026-02-10' and cH['end_date'] is None and WK == 40
assert cT['end_date'] == '2026-02-28'
ORDER = ['hirbovat', 'telita']   # list order: the farm with an open cycle first, then by name
def tree_of(fid): return ST['farm_tree'](fid)
assert tree_of('telita') == [(c, []) for c in ['H1', 'H2', 'H3', 'H6', 'H5-1', 'H5-2']]
assert tree_of('hirbovat') == [('H1', ['H1.1', 'H1.2']), ('H3', ['H3.1', 'H3.2']), ('H4', ['H4.1', 'H4.2'])]
def pl(k, one, many): return f'<span class="s19-nw">{num(k)} {one if k == 1 else many}</span>'
def sec_count(fid):
    t = tree_of(fid); halls = [x for x in t if x[1]]
    if halls:
        return f'{pl(len(halls), "hală", "hale")} · {pl(sum(len(c) for _, c in halls), "compartiment", "compartimente")}'
    return pl(len(t), 'secțiune', 'secțiuni')
def sec_codes(fid):
    t = tree_of(fid)
    if any(k for _, k in t):
        return ' · '.join(f'<span class="s19-nw">{" ".join(k)}</span>' for _, k in t)
    return ' '.join(f'<span class="s19-nw">{c}</span>' for c, _ in t)
def cyc_cell(fid):
    if fid == 'hirbovat':
        return f'<span class="s19-cyc">{lamp("ok", "deschis")}<span class="ip-note ip-num"><span class="s19-nw">SAPT {WK}</span> · <span class="s19-nw">din {dd(cH["start_date"])}</span></span></span>'
    return f'<span class="s19-cyc">{lamp("off", "goală")}<span class="ip-note ip-num"><span class="s19-nw">din {dd(cT["end_date"])}</span></span></span>'
def gtag(fid, noguide=False):
    return f'<span class="ip-tag ip-tag--neutral">{"fără ghid" if noguide else GUIDE[FARM[fid]["guide_id"]]}</span>'

# ================================================================= list (desktop)
CRUMBS = ('<nav class="ip-crumbs" aria-label="Cale"><a href="#setari">Setări</a><span aria-hidden="true">›</span>'
          '<span aria-current="page">Ferme și secțiuni</span></nav>')
def meta(farms=True):
    tail = f'{NB}·{NB}Hîrbovăț, Telița' if farms else ''
    return f'<div class="ip-row s4-meta"><span class="ip-note">Tipul, ghidul și secțiunile fiecărei ferme{tail}</span></div>'
def newkey(pressed=False):
    p = ' is-pressed" aria-expanded="true" aria-controls="fp' if pressed else ''
    return f'<button type="button" class="ip-btn ip-btn--secondary{p}">{icon("plus", 16)}Fermă nouă</button>'
def header(pressed=False, farms=True):
    key = f'\n<div class="ip-row">{newkey(pressed)}</div>' if farms else ''   # no farms: the EmptyState holds the one next step
    return f'<header class="ip-pagehead">\n<div>\n{CRUMBS}\n<h1>Ferme și secțiuni</h1>\n{meta(farms)}\n</div>{key}\n</header>'
def sidebar_owner(no_series=False):
    sb = s4.sidebar('Proprietar').replace('class="ip-nav is-current" href="#serie" aria-current="page"', 'class="ip-nav" href="#serie"')
    par = f'<a class="ip-nav" href="#setari">{icon("settings")}Setări</a>'; assert sb.count(par) == 1
    subs = ''.join(f'<a class="ip-nav ip-nav--sub{" is-current" if h == "#ferme" else ""}" href="{h}"'
                   f'{" aria-current=\"page\"" if h == "#ferme" else ""}>{t}</a>' for h, t in Q.OWNER_SUBS)
    sb = sb.replace(par, par + subs, 1)
    assert sb.count('is-current') == 1 and 'href="#ferme" aria-current="page"' in sb
    if no_series:  # D39 first-use board: no farm, so no series either
        cur = f'<span class="name">Seria 2026</span>{lamp("ok", "deschisă")}'; assert sb.count(cur) == 1
        sb = sb.replace(cur, '<span class="name">Nicio serie</span>')
        sb, k = re.subn(r'(<ul class="ip-menu ip-series__menu"[^>]*>).*?(</ul>)', r'\1<li class="ip-menu__group" role="presentation">Serii</li>\2', sb, flags=re.S)
        assert k == 1
    return sb
def editkey(fid, pressed=False):
    nm = FARM[fid]['name']
    p = ' is-pressed" aria-expanded="true" aria-controls="fp' if pressed else ''
    return f'<button type="button" class="ip-btn ip-btn--ghost{p}">Editează<span class="ip-sr"> {nm}</span></button>'
COLS = ('<thead><tr><th scope="col">Fermă</th><th scope="col">Tip</th><th scope="col">Ghid</th><th scope="col">Secțiuni</th>'
        '<th scope="col">Ciclu</th><th scope="col"><span class="ip-sr">Acțiuni</span></th></tr></thead>')
def frow(fid, open_=None, noguide=()):
    f = FARM[fid]
    return (f'<tr><th scope="row" class="s19-fname">{f["name"]}</th><td>{TYPE_RO[f["type"]]}</td><td>{gtag(fid, fid in noguide)}</td>'
            f'<td><span class="s19-secs"><span>{sec_count(fid)}</span><span class="ip-note ip-num">{sec_codes(fid)}</span></span></td>'
            f'<td>{cyc_cell(fid)}</td><td class="s19-act">{editkey(fid, open_ == fid)}</td></tr>')
def flist(open_=None, noguide=()):
    rows = ''.join(frow(fid, open_, noguide) for fid in ORDER)
    note = ('<p class="ip-note s4s-m0">Un ghid nou se aplică de la ciclul următor al fermei. Secțiunile cu păsări nu se pot retrage.</p>')
    return (f'<section class="s4-block" aria-labelledby="h-fer"><div class="s4-head"><h2 class="ip-t-heading" id="h-fer">Ferme</h2>'
            f'<span class="ip-note">{pl(len(ORDER), "fermă", "ferme")}</span></div>'
            f'<div class="ip-tablescroll ip-tablescroll--framed" tabindex="0" role="region" aria-labelledby="h-fer">'
            f'<table class="ip-table ip-table--rowheads s19-table"><caption class="ip-sr">Fermele, cu tipul, ghidul, secțiunile și ciclul lor</caption>'
            f'{COLS}<tbody>{rows}</tbody></table></div>{note}</section>')
def page(body, open_=None, pressed_new=False, farms=True, inert=False, noguide=()):
    main = body if body is not None else flist(open_, noguide)
    i = ' inert="inert"' if inert else ''
    return (f'<div class="s4-shell"{i}>{sidebar_owner(no_series=not farms)}<main class="s4-main">\n{header(pressed_new, farms)}\n{main}\n</main></div>')

# ================================================================= farm panel (SidePanel beside the page, D73)
OPT = '<span class="s13-opt">(opțional)</span>'
ERRI = icon('alert', 14)
def name_field(val, focus=False):
    return T.tfield('fp-name', 'Nume', val, touch=False, focus=focus)
def tip_field(fid=None, sel=None, focus=False, readonly=False):
    if readonly:
        return (f'<div class="ip-field"><label class="ip-label" for="fp-tip">Tip</label><span class="ip-input is-readonly">'
                f'<input id="fp-tip" type="text" readonly="readonly" value="{TYPE_RO[FARM[fid]["type"]]}" aria-describedby="fp-tip-h"/></span>'
                f'<span class="ip-hint" id="fp-tip-h">Tipul se schimbă doar cât ferma e goală: ciclul de ouat e deschis din {num(dd(cH["start_date"]))}.</span></div>')
    its = ''
    for i, k in enumerate(('growth', 'laying')):
        on = sel == k; first = (sel is None and i == 0) or on
        f = ' ip-focus' if (focus and on) else ''
        its += (f'<span class="ip-seg__item{" is-selected" if on else ""}{f}" role="radio" aria-checked="{"true" if on else "false"}" '
                f'tabindex="{0 if first else -1}">{TYPE_RO[k]}</span>')
    return (f'<div class="ip-field"><span class="ip-label" id="fp-tip-l">Tip</span>'
            f'<span class="ip-seg" role="radiogroup" aria-labelledby="fp-tip-l" aria-required="true">{its}</span></div>')
def gsel(text, hint, state='', desc='fp-ghid-h', placeholder=False):
    st = f' {state}' if state else ''
    dis = ' disabled="disabled"' if state == 'is-disabled' else ''
    inv = ' aria-invalid="true"' if state == 'is-error' else ''
    opt = f'<option value=""{" selected=\"selected\"" if placeholder else ""}>{text}</option>' if placeholder else f'<option selected="selected">{text}</option>'
    req = ''   # S19S fix 1: Ghid is optional, so no required (the placeholder option stays)
    return (f'<span class="ip-input{st}"><select id="fp-ghid" aria-describedby="{desc}"{dis}{inv}{req}>{opt}</select>{icon("chev", 16)}</span>{hint}')
def ghid_field(fid, mode='set'):
    lbl = f'<label class="ip-label" for="fp-ghid">Ghid {OPT}</label>'
    if mode == 'set' and fid == 'hirbovat':   # open cycle: the B1 next-cycle rule, the current cycle changes from its panel (B4)
        h = (f'<span class="ip-hint" id="fp-ghid-h">Doar ghiduri de ouat. Un ghid nou se aplică de la ciclul următor; pe cel curent '
             f'(SAPT {WK}) îl schimbi din panoul ciclului, pe pagina fermei.</span>')
        return f'<div class="ip-field">{lbl}{gsel(GUIDE["g-laying"], h)}</div>'
    if mode == 'set':                          # empty farm: the next cycle is the first one to use it
        h = '<span class="ip-hint" id="fp-ghid-h">Doar ghiduri de creștere. Se aplică de la următorul ciclu al fermei.</span>'
        return f'<div class="ip-field">{lbl}{gsel(GUIDE["g-growth"], h)}</div>'
    if mode == 'wrong':                        # D39: Tip switched back to Creștere, the ouat guide kept and flagged
        e = (f'<p class="ip-err ip-fieldmsg" id="fp-ghid-e">{ERRI}<span>Ghidul „{GUIDE["g-laying"]}” este pentru ouat. '
             f'Ferma {FARM[fid]["name"]} este de creștere. Alege un ghid de creștere sau schimbă tipul.</span></p>')
        return f'<div class="ip-field">{lbl}{gsel(GUIDE["g-laying"], e, "is-error", "fp-ghid-e")}</div>'
    if mode == 'new':                          # no type yet: the guide list waits for it
        h = '<span class="ip-hint" id="fp-ghid-h">Alege întâi tipul: lista arată doar ghidurile de tipul fermei.</span>'
        return f'<div class="ip-field">{lbl}{gsel("Alege un ghid", h, "is-disabled", placeholder=True)}</div>'
    if mode == 'open':                         # D39 fără ghid: the open list, filtered by the farm's type
        items = (f'<li class="ip-menu__item is-active" role="option" id="fp-g0" aria-selected="true"><span>Fără ghid</span>{icon("check", 16)}</li>'
                 f'<li class="ip-menu__group" role="presentation">Ghiduri de creștere</li>'
                 f'<li class="ip-menu__item" role="option" id="fp-g1" aria-selected="false"><span>{GUIDE["g-growth"]}</span></li>')
        combo = (f'<div class="ip-input is-focus s19-combo" id="fp-ghid" role="combobox" tabindex="0" aria-haspopup="listbox" aria-expanded="true" '
                 f'aria-controls="fp-glist" aria-activedescendant="fp-g0" aria-labelledby="fp-ghid-l" aria-describedby="fp-ghid-h">'
                 f'<span class="s4-grow">Fără ghid</span>{icon("chev", 16)}</div>')
        h = '<span class="ip-hint" id="fp-ghid-h">Fără ghid, graficele fermei arată doar valorile reale.</span>'
        return (f'<div class="ip-field"><span class="ip-label" id="fp-ghid-l">Ghid {OPT}</span><div class="s19-anchor">{combo}'
                f'<ul class="ip-menu s19-pop" role="listbox" id="fp-glist" aria-labelledby="fp-ghid-l">{items}</ul></div>{h}</div>')
    if mode == 'none':                         # D39 fără ghid: closed, so the actuals-only line stays in view
        h = ('<span class="ip-hint" id="fp-ghid-h">Fără ghid, graficele fermei arată doar valorile reale. '
             'Lista are doar ghiduri de creștere.</span>')
        return f'<div class="ip-field">{lbl}{gsel("Fără ghid", h)}</div>'
    raise ValueError(mode)
def tree_farm(fid, noguide=False, **kw):
    fr = ST['farm_row'](fid, tag=gtag(fid, True) if noguide else None)
    return ST['tree']('st', f'Secțiuni, {FARM[fid]["name"]}', [fr] + ST['section_rows']('st', fid, **kw) + [ST['add_row'](fid)])
NEWTREE = ('<div class="ip-stree"><div class="ip-list__head"><h3 class="ip-cap">Secțiuni</h3></div><div class="ip-stree__box">'
           '<table class="ip-table ip-stree__t" aria-label="Secțiuni, fermă nouă"><tbody><tr class="ip-stree__empty ip-stree--l2"><td colspan="3">'
           '<span class="ip-stree__cell"><span class="ip-lamp ip-lamp--off">fără secțiuni</span>'
           '<span>Adaugi secțiunile după ce salvezi ferma.</span></span></td></tr></tbody></table></div></div>')
def ctx(fid):
    if fid == 'hirbovat': return f'OUAT · CICLU DESCHIS · SAPT {WK}'
    return f'CREȘTERE · GOALĂ DIN {dd(cT["end_date"])}'
def panel(title, note, body, key, inert=False):
    i = ' inert="inert"' if inert else ''
    nt = f'<p class="ip-note ip-num s13-m0">{note}</p>' if note else ''
    return (f'<aside class="ip-sidepanel s13-panel" id="fp" aria-labelledby="fp-t"{i}>\n<div class="ip-sidepanel__head"><div>'
            f'<h2 class="ip-t-heading s13-m0" id="fp-t">{title}</h2>{nt}</div>'
            f'<button type="button" class="ip-iconbtn" aria-label="Închide panoul fermei">{icon("x")}</button></div>\n'
            f'<div class="ip-sidepanel__body s13-pbody s19-pbody">\n{"".join(body)}\n</div>\n'
            f'<div class="ip-sidepanel__foot s18-foot"><div class="ip-row s13-pkeys"><button type="button" class="ip-btn ip-btn--secondary">Anulează</button>{key}</div></div>\n</aside>')
SAVE = '<button type="button" class="ip-btn ip-btn--primary">Salvează ferma</button>'
SAVE_NO = '<button type="button" class="ip-btn ip-btn--primary" aria-disabled="true" aria-describedby="fp-ghid-e">Salvează ferma</button>'
ADD = '<button type="button" class="ip-btn ip-btn--primary">Adaugă ferma</button>'
def fpanel(fid, tip, ghid, tree, key=SAVE, focus_name=False, inert=False):
    return panel(FARM[fid]['name'], ctx(fid), [name_field(FARM[fid]['name'], focus_name), tip, ghid, tree], key, inert)
def vp(open_=None, pressed_new=False, noguide=(), inert=False):
    return f'<div class="s13-vp s18-vp">{page(None, open_, pressed_new, noguide=noguide, inert=inert)}</div>'

# ================================================================= no farms (EmptyState, B1)
NOFARMS = s4.empty_box('fără ferme', ['Nu există nicio fermă.', 'Adaugă prima fermă, apoi secțiunile și ghidul ei.'],
                       '<button type="button" class="ip-btn ip-btn--primary">Adaugă prima fermă</button>', 'Ferme')

# ================================================================= phone list (owner, read-only, D110)
def p_header():
    h = CRUMBS.replace('<a href="#setari">Setări</a>', '<a href="#setari"><span class="s14-crumb">Setări</span></a>')
    return f'<header class="ip-pagehead s2-phead s14-phead"><div>{h}<h1>Ferme și secțiuni</h1></div></header>'
def phone_list():
    """Phone (D110): one framed table, a group row per farm and one row per field (fits 358 px; no sideways scroll)."""
    rows = ''
    for f in ORDER:
        nm = FARM[f]['name']
        rows += (f'<tbody><tr class="is-group"><th scope="rowgroup" colspan="2" id="pf-{f}">{nm}</th></tr>'
                 f'<tr><th scope="row" class="ip-table__child">Tip</th><td>{TYPE_RO[FARM[f]["type"]]}</td></tr>'
                 f'<tr><th scope="row" class="ip-table__child">Ghid</th><td>{gtag(f)}</td></tr>'
                 f'<tr><th scope="row" class="ip-table__child">Secțiuni</th><td><span class="s19-secs"><span>{sec_count(f)}</span>'
                 f'<span class="ip-note ip-num">{sec_codes(f)}</span></span></td></tr>'
                 f'<tr><th scope="row" class="ip-table__child">Ciclu</th><td>{cyc_cell(f)}</td></tr></tbody>')
    blk = (f'<section class="s4-block" aria-labelledby="h-fer"><div class="s4-head"><h2 class="ip-t-heading" id="h-fer">Ferme</h2>'
           f'<span class="ip-note">{pl(len(ORDER), "fermă", "ferme")}</span></div>'
           f'<div class="ip-tablescroll ip-tablescroll--framed" role="region" aria-labelledby="h-fer">'
           f'<table class="ip-table ip-table--rowheads s19-table s19-ptable"><caption class="ip-sr">Fermele, cu tipul, ghidul, secțiunile și ciclul lor</caption>'
           f'{rows}</table></div>'
           f'<p class="ip-note s4s-m0">Fermele și secțiunile le editezi pe calculator.</p></section>')
    body = A.phone([p_header(), blk])
    for a, z in (('class="ip-tabbar__item is-current" href="#acasa" aria-current="page"', 'class="ip-tabbar__item" href="#acasa"'),
                 ('class="ip-tabbar__item" href="#mai-mult"', 'class="ip-tabbar__item is-current" href="#mai-mult" aria-current="page"')):
        assert body.count(a) == 1, a; body = body.replace(a, z)
    assert body.count('is-current') == 1
    return body

# ================================================================= CSS (screen-only; tokens only; system classes untouched)
CSS19 = '''
.s19-table{width:100%}
.s19-table thead th{white-space:nowrap}
.s19-table td{vertical-align:middle;white-space:normal}
.s19-nw{white-space:nowrap}
.s19-fname{font-weight:500}
.s19-secs{display:flex;flex-direction:column;gap:0}
.s19-cyc{display:inline-flex;align-items:center;flex-wrap:wrap;gap:0 var(--space-2)}
.s19-act{text-align:right;white-space:nowrap}
.s19-anchor{position:relative}
.s19-combo{cursor:pointer}
.s19-pop{position:absolute;top:calc(100% + var(--space-1));left:0;right:0;z-index:4;list-style:none;margin:0}
.s19-pbody{position:relative}
.s19-scrim{position:absolute;inset:0;z-index:5;border-radius:0;align-items:center} /* the Dialog over the whole desktop page (as S16S/S18S) */'''
CSS_DESK = T.CSS_DESK + SC.OPT + CSS19
CSS_PL = T.CSS_PL + CSS19 + '''
.s19-ptable td{height:auto;padding-block:var(--space-2)}
.s19-ptable .ip-table__child{width:calc(var(--space-12) * 2);vertical-align:middle}'''

# ================================================================= write
FONTS = s4.FONTS; SYS = ROOT / 'design/system'
pathlib.Path('/tmp/s19').mkdir(exist_ok=True)
def write(stem, title, root, css, w, h):
    dc = f'''<!doctype html>
<html lang="ro">
<head>
<meta charset="utf-8">
<title>{title}</title>
<script src="./support.js"></script>
<link rel="stylesheet" href="ds/impexlinepanel/tokens.css">
<link rel="stylesheet" href="ds/impexlinepanel/components/bundle.css">
<link rel="stylesheet" href="additions.css">
</head>
<body>
<x-dc>
<helmet>
{FONTS}
<style>{css}</style>
</helmet>
{root}
</x-dc>
<script type="text/x-dc" data-dc-script data-props='{{"$preview":{{"width":{w},"height":{h}}}}}'>
class Component extends DCLogic {{ renderVals() {{ return {{}}; }} }}
</script>
</body>
</html>
'''
    (ROOT / f'design/canvas/project/{stem}.dc.html').write_text(dc, encoding='utf-8')
    wrap = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8"><title>{title}</title>
<link rel="stylesheet" href="file://{SYS}/tokens.css"><link rel="stylesheet" href="file://{SYS}/components/bundle.css"><link rel="stylesheet" href="file://{SYS}/components/additions.css">
{FONTS}<style>{css}</style></head><body>{root}</body></html>'''
    pathlib.Path(f'/tmp/s19/{stem}.html').write_text(wrap, encoding='utf-8')
    print('ok', stem, w, h)
    return {'stem': stem, 'title': title, 'w': w, 'h': h}

STEM, TT = 'S19-ferme-sectiuni', 'S19 · Ferme și secțiuni'
out = []
def desk(stem, title, root, h=None):
    h = h or HEIGHTS.get(STEM + stem, 1024)
    out.append(write(STEM + stem, f'{TT}{" · " + title if title else ""}', f'<div class="ip-root s4s-rel" style="padding:0; min-height:{h}px">{root}</div>', CSS_DESK, 1440, h))
PH = SC.SIDE_H   # panel boards: the page clipped beside the panel, 1024

# 1 · the list (default)
desk('', '', page(None))
# 2 · Hîrbovăț panel (open cycle): Tip read-only, guide rule, the tree with birds; focus on Nume (first field)
desk('--hirbovat', 'Hîrbovăț · panoul fermei',
     vp('hirbovat') + fpanel('hirbovat', tip_field('hirbovat', readonly=True), ghid_field('hirbovat'), tree_farm('hirbovat'), focus_name=True), PH)
# 3 · Telița panel (empty farm): Tip editable, H5-2 menu open with Retrage live
desk('--telita', 'Telița · panoul fermei · meniul H5-2',
     vp('telita') + fpanel('telita', tip_field(sel='growth'), ghid_field('telita'), tree_farm('telita', open_code='H5-2', istate={4: 'is-active'})), PH)
# 4 · Retrage blocked: Hîrbovăț H1.1 has birds (real, cycle open)
desk('--retragere-blocata', 'Hîrbovăț · retragere blocată: H1.1 are păsări',
     vp('hirbovat') + fpanel('hirbovat', tip_field('hirbovat', readonly=True), ghid_field('hirbovat'),
                             tree_farm('hirbovat', open_code='H1.1', istate={4: 'is-active'})), PH)
# 5 · Retrage H5-2 Dialog (D39), focus on Anulează, page and panel inert
DLG = (f'<div class="ip-scrim s19-scrim"><div class="ip-dialog" role="alertdialog" aria-modal="true" aria-labelledby="dlg-t" aria-describedby="dlg-n">'
       f'<h2 id="dlg-t">Retrage secțiunea H5-2 de la Telița</h2>'
       f'<p class="ip-dialog__note" id="dlg-n">H5-2 nu mai apare în formulare și pe pagina fermei. Înregistrările ei din Seria 2026 rămân în rapoarte. '
       f'O secțiune retrasă nu se mai poate reactiva.</p>'
       f'<div class="ip-dialog__keys"><button type="button" class="ip-btn ip-btn--secondary ip-focus" autofocus="autofocus">Anulează</button>'
       f'<button type="button" class="ip-btn ip-btn--destructive">Retrage secțiunea</button></div></div></div>')
desk('--retrage-sectiunea', 'Telița · Retrage H5-2 · exemplu de stare (D39)',
     vp('telita', inert=True) + fpanel('telita', tip_field(sel='growth'), ghid_field('telita'), tree_farm('telita'), inert=True) + DLG, PH)
# 6 · fără ghid (D39/D51): Telița without a guide, the Ghid list open (filtered to creștere)
desk('--fara-ghid', 'Telița fără ghid · exemplu de stare (D39)',
     vp('telita', noguide=('telita',)) + fpanel('telita', tip_field(sel='growth'), ghid_field('telita', 'none'), tree_farm('telita', noguide=True)), PH)
# 7 · ghid nepotrivit (D39): Tip switched back to Creștere with Ross 308 ouat kept → the B1 sentence, save refused
desk('--ghid-nepotrivit', 'Telița · ghid de alt tip: refuzat · exemplu de stare (D39)',
     vp('telita') + fpanel('telita', tip_field(sel='growth', focus=True), ghid_field('telita', 'wrong'), tree_farm('telita'), key=SAVE_NO), PH)
# 8 · no farms (D39 first use): EmptyState with the one next step
desk('--fara-ferme', 'nicio fermă · exemplu de stare (D39)', page(NOFARMS, farms=False))
# 9 · Fermă nouă: the list beside an empty panel; Tip unset, Ghid waits, sections after the first save; focus on Nume
desk('--ferma-noua', 'Fermă nouă',
     f'<div class="s13-vp s18-vp">{page(None, pressed_new=True)}</div>'
     + panel('Fermă nouă', '', [name_field('', True), tip_field(), ghid_field(None, 'new'), NEWTREE], ADD), PH)
# 10 · phone, read-only list (D110)
h = HEIGHTS.get(STEM + '--telefon', 844)
out.append(write(STEM + '--telefon', f'{TT} · telefon · doar citire', f'<div class="ip-root" style="padding:0; width:390px; min-height:{h}px">{phone_list()}</div>', CSS_PL, 390, h))
pathlib.Path('/tmp/s19/boards.json').write_text(json.dumps([{k: b[k] for k in ('stem', 'w', 'h')} for b in out]))

if CANVAS:  # new row below the S18 row (row y + tallest h + 120); desktops first (a Setări page), 80 px gaps
    cp = ROOT / 'design/canvas/project/canvas.json'; src = cp.read_text(encoding='utf-8'); c = json.loads(src)
    others = {k: v for k, v in c['boards'].items() if not k.startswith('S19-')}
    ly = max(v['y'] for v in others.values()); y = ly + max(v['h'] for v in others.values() if v['y'] == ly) + 120
    assert (ly, y) == (32813, 34557), (ly, y)
    x = 0
    for b in out:
        k = f'{b["stem"]}.dc.html'
        e = {'x': x, 'y': y, 'w': b['w'], 'h': b['h'], 'expand': 'fill', 'title': b['title']} if b['w'] == 1440 else {'x': x, 'y': y, 'w': b['w'], 'h': b['h'], 'title': b['title']}
        c['boards'][k] = e; x += b['w'] + 80
        if k not in c['order']: c['order'].append(k)
    txt = json.dumps(c, indent=2, ensure_ascii=False) + ('\n' if src.endswith('\n') else '')
    assert json.loads(txt) == c; cp.write_text(txt, encoding='utf-8')
    print('canvas: S19 row y', y, 'ends at x', x - 80)
