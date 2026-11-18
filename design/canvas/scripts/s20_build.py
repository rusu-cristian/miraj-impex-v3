"""S20 · Ghiduri (spec B2, screens.md S20; nav map: Setări › Ghiduri, owner). Desktop-led like S17–S19 (a Setări page):
desktops first, then the phone. Today = 16.11.2026 (D23/D24). D113: the guide detail and the import are sub-pages of
Ghiduri (crumbs), not SidePanels (SYS11 needs ≥672px, D76); a guide used by a cycle is replaced from its detail page
(re-import into the same guide, Tip locked), never deleted.
Content, from the extracts only: guide-growth.json (Ross 308 creștere, SAPT 1–25, hrană g ♀/♂, greutate g ♀/♂ to SAPT 23)
and guide-laying.json (Ross 308 ouat, vârsta 19–65 = SAPT 1–47, hrană g ♀/♂, ouat % ♀ SAPT 7–45 illustrative, D29);
which cycles use each guide from cycles.json + farms-sections.json (through s18_build). Guide and file names illustrative
(D22). The import preview is SYS11's GuideImportPreview (additions.css §29): its helpers are read from
design/system/scripts/sys11_build.py (the part above its preview blocks, so nothing is written there).
Boards (row below S19: y 34557 + 1024 + 120 = 35701; x step w + 80; desktops first):
  S20-ghiduri                  list: PageHeader (Setări › Ghiduri, "Importă ghid"), DataTable of guides (Tip, Săptămâni,
                               Coloane + fără date în ghid, Folosit de, Deschide)
  --ross-308-ouat              guide detail (laying): Folosit de (Hîrbovăț 2026 open SAPT 40, Hîrbovăț 2025 closed), small
                               curves (ip-sparktable), weekly table SAPT × ♀/♂ with Vârstă; Șterge aria-disabled + reason
  --ross-308-crestere          guide detail (growth): partial weight (SAPT 24–25 "–" fără date în ghid)
  --importa-ghid               new import: Nume (focus), Tip unset, FileDrop (paste or file), Salvează ghidul aria-disabled
  --inlocuieste-fisier         D39 flow: Înlocuiește Ross 308 creștere with its own .xlsx (real 50 rows): all valid
  --inlocuieste-refuzate       D39: the same file with rows 7, 10, 12 refused (SYS11 state examples), save blocked
  --inlocuieste-lipit          D39 flow: Înlocuiește Ross 308 ouat by paste (real 94 rows, vârstă + SAPT)
  --salvat                     D39: back on the Ross 308 creștere detail after the replace, Toast + Anulează
  --telefon                    phone, read-only list ("Ghidurile le imporți pe calculator", D76 desktop only)
python3 -I s20_build.py <repo_root> [s20_heights.json] [--canvas]
Writes design/canvas/project/S20-ghiduri[--<state>].dc.html, /tmp/s20/<stem>.html, /tmp/s20/boards.json; --canvas adds the
boards to canvas.json (new row, nothing else changes). Renders: s20_render.py [--measure] [--w 1280] [stems]."""
import sys, re, json, pathlib, importlib.util, io, contextlib, datetime as dt

ARGS = sys.argv[1:]
CANVAS = '--canvas' in ARGS; ARGS = [a for a in ARGS if a != '--canvas']
ROOT = pathlib.Path(ARGS[0]).resolve(); HERE = pathlib.Path(__file__).resolve().parent
HP = pathlib.Path(ARGS[1]) if len(ARGS) > 1 else HERE / 's20_heights.json'
HEIGHTS = json.loads(HP.read_text()) if HP.exists() else {}
sys.path.insert(0, str(HERE))
import skel as SKL
sys.argv = ['s19_build.py', str(ROOT)]
sp = importlib.util.spec_from_file_location('s19_build', HERE / 's19_build.py')
M = importlib.util.module_from_spec(sp)
with contextlib.redirect_stdout(io.StringIO()):
    sp.loader.exec_module(M)   # rewrites the S19 boards (and through s18_build S18/S17/S14S/S2/S7) byte-identical
T, SC, Q, s4, A = M.T, M.SC, M.Q, M.s4, M.A
icon, lamp, num, dd, NB = M.icon, M.lamp, M.num, M.dd, M.NB
FARM, GUIDE, WK, cH, cT = M.FARM, M.GUIDE, M.WK, M.cH, M.cT
cP = T.cP

# GuideImportPreview helpers (SYS11): the code above the preview blocks only (no file is written)
SRC = (ROOT / 'design/system/scripts/sys11_build.py').read_text(encoding='utf-8')
CUT = '# ------------------------------------------------------------------ blocks'
assert SRC.count(CUT) == 1
G = {'__name__': 'sys11_parts', '__file__': str(ROOT / 'design/system/scripts/sys11_build.py')}  # sys11 finds the repo from its own path
exec(compile(SRC.split(CUT)[0], 'sys11_build.py', 'exec'), G)

# ================================================================= data (extracts only)
EX = ROOT / 'design/data/extracts'
GROW = json.loads((EX / 'guide-growth.json').read_text(encoding='utf-8'))
LAY = json.loads((EX / 'guide-laying.json').read_text(encoding='utf-8'))
assert GUIDE == {'g-growth': 'Ross 308 creștere', 'g-laying': 'Ross 308 ouat'}
assert len(GROW['rows']) == 50 and len(LAY['rows']) == 94
assert cH['guide_id'] == cP['guide_id'] == 'g-laying' and cT['guide_id'] == 'g-growth'
assert FARM['telita']['guide_id'] == 'g-growth' and FARM['hirbovat']['guide_id'] == 'g-laying'
fmt = G['fmt']   # SYS11's fmt: thin no-break thousands, true minus, a percentage keeps one decimal (89,0), as the preview
def wk_rows(gid):
    """{week: {'F': row, 'M': row}} and the (week, age) list, by guide week (laying: cycle week, age beside, D26)."""
    out = {}
    if gid == 'g-growth':
        for r in GROW['rows']: out.setdefault(r['week'], {'age': None})[r['sex']] = r
    else:
        for r in LAY['rows']: out.setdefault(r['cycle_week'], {'age': r['age_week']})[r['sex']] = r
    return out
W = {g: wk_rows(g) for g in GUIDE}
assert sorted(W['g-growth']) == list(range(1, 26)) and sorted(W['g-laying']) == list(range(1, 48))
assert W['g-laying'][40]['age'] == 58 and W['g-laying'][40]['F']['feed_g'] == 158 and W['g-laying'][40]['M']['feed_g'] == 155
TYPE = {'g-growth': 'growth', 'g-laying': 'laying'}
TYPE_RO, TYPE_LO = M.TYPE_RO, M.TYPE_LO
# indicators of B2, in spec order; key in the extracts, label, unit, sexes that carry it
IND = [('feed_g', 'Hrană', 'g', 'FM'), ('body_weight_g', 'Greutate', 'g', 'FM'), ('mortality_pct', 'Mortalitate', '%', ''),
       ('lay_pct', 'Ouat', '%', 'F'), ('hatching_pct', 'Incubabile', '%', ''), ('egg_weight_g', 'Masă ou', 'g', '')]
HAS = {'g-growth': ['feed_g', 'body_weight_g'], 'g-laying': ['feed_g', 'lay_pct']}
for g, ks in HAS.items():   # columns present = indicators with at least one figure
    assert ks == [k for k, *_ in IND if any(W[g][w][s].get(k) is not None for w in W[g] for s in 'FM')], g
def series(g, k, s):
    return [(w, W[g][w][s].get(k)) for w in sorted(W[g])]
def span(g, k, s):
    pts = [(w, v) for w, v in series(g, k, s) if v is not None]
    return pts[0][0], pts[-1][0]
def gaps(g, k, s):
    """Weeks without a figure, as ranges ('24–25')."""
    miss = [w for w, v in series(g, k, s) if v is None]; out = []
    for w in miss:
        if out and out[-1][1] == w - 1: out[-1][1] = w
        else: out.append([w, w])
    return ['–'.join(str(x) for x in sorted(set(r))) for r in out]
assert gaps('g-growth', 'body_weight_g', 'F') == ['24–25'] and gaps('g-laying', 'lay_pct', 'F') == ['1–6', '46–47']
NWK = {g: len(W[g]) for g in GUIDE}
NROWS = {'g-growth': len(GROW['rows']), 'g-laying': len(LAY['rows'])}

# cycles that use each guide (cycles.json), newest first
CYC = {'g-laying': [cH, cP], 'g-growth': [cT]}
SER = {'seria-2026': 'Seria 2026', 'seria-2025': 'Seria 2025'}
def cyc_words(k): return f'{num(k)} {"ciclu" if k == 1 else "cicluri"}'

# ================================================================= shared parts
SEXW = {'F': ('♀', 'hen', 'găini'), 'M': ('♂', 'rooster', 'cocoși')}
def sx(s): return f'<span class="ip-sex ip-sex--{SEXW[s][1]}">{SEXW[s][0]}</span><span class="ip-sr"> {SEXW[s][2]}</span>'
UPLOAD = G['IC']['upload']
s4.IC.setdefault('upload', UPLOAD)
def crumbs(*parts):
    """Setări › … ; the last part is the current page."""
    links = [('#setari', 'Setări')] + list(parts[:-1])
    a = ''.join(f'<a href="{h}">{t}</a><span aria-hidden="true">›</span>' for h, t in links)
    return f'<nav class="ip-crumbs" aria-label="Cale">{a}<span aria-current="page">{parts[-1]}</span></nav>'
GH = ('#ghiduri', 'Ghiduri')
def slug(g): return '#ghid-' + GUIDE[g].lower().replace(' ', '-').replace('ș', 's').replace('ț', 't').replace('ă', 'a').replace('â', 'a').replace('î', 'i')
def sidebar():
    sb = M.sidebar_owner()
    for a, z in (('class="ip-nav ip-nav--sub is-current" href="#ferme" aria-current="page"', 'class="ip-nav ip-nav--sub" href="#ferme"'),
                 ('class="ip-nav ip-nav--sub" href="#ghiduri"', 'class="ip-nav ip-nav--sub is-current" href="#ghiduri" aria-current="page"')):
        assert sb.count(a) == 1, a; sb = sb.replace(a, z)
    assert sb.count('is-current') == 1
    return sb
def header(cr, title, meta, keys='', tid=None):
    t = f' id="{tid}" tabindex="-1"' if tid else ''
    k = f'\n<div class="s20-keys">{keys}</div>' if keys else ''
    return (f'<header class="ip-pagehead">\n<div>\n{cr}\n<h1{t}>{title}</h1>\n'
            f'<div class="ip-row s4-meta"><span class="ip-note">{meta}</span></div>\n</div>{k}\n</header>')
def shell(hdr, main):
    return f'<div class="s4-shell">{sidebar()}<main class="s4-main">\n{hdr}\n{main}\n</main></div>'
def block(hid, title, inner, note='', count=''):
    c = f'<span class="ip-note">{count}</span>' if count else ''
    n = f'<p class="ip-note s4s-m0">{note}</p>' if note else ''
    return (f'<section class="s4-block" aria-labelledby="{hid}"><div class="s4-head"><h2 class="ip-t-heading" id="{hid}">{title}</h2>{c}</div>'
            f'{inner}{n}</section>')
def framed(label_id, table, focus=True):
    t = ' tabindex="0"' if focus else ''
    return f'<div class="ip-tablescroll ip-tablescroll--framed"{t} role="region" aria-labelledby="{label_id}">{table}</div>'

# ---- what a guide holds (list and detail)
def weeks_txt(g):
    if g == 'g-laying':
        return (f'<span class="s19-nw">SAPT 1–{NWK[g]}</span> <span class="ip-note s19-nw">vârsta {W[g][1]["age"]}–{W[g][NWK[g]]["age"]}</span>')
    return f'<span class="s19-nw">SAPT 1–{NWK[g]}</span>'
def cols_txt(g):
    return ' · '.join(f'<span class="s19-nw">{lbl} {u}</span>' for k, lbl, u, _ in IND if k in HAS[g])
def missing(g):
    return [f'{lbl.lower()} {u}' for k, lbl, u, _ in IND if k not in HAS[g]]
def missing_txt(g): return ', '.join(missing(g))
def used_txt(g, short=False):
    out = []
    for c in CYC[g]:
        f = FARM[c['farm_id']]['name']
        if c['status'] == 'open':
            out.append(f'<span class="s20-use"><span class="s19-nw">{lamp("ok", "deschis")} {f}</span> · <span class="s19-nw">{SER[c["series_id"]]}</span>'
                       f' · <span class="ip-note ip-num s19-nw">SAPT {WK}</span></span>')
        else:
            out.append(f'<span class="s20-use"><span class="s19-nw">{lamp("off", "închis")} {f}</span> · <span class="s19-nw">{SER[c["series_id"]]}</span></span>')
    return ''.join(out)

# ================================================================= list
def list_key(pressed=False):
    p = ' is-pressed' if pressed else ''
    return f'<a class="ip-btn ip-btn--secondary{p}" href="#importa-ghid">{icon("upload", 16)}Importă ghid</a>'
LCOLS = ('<thead><tr><th scope="col">Ghid</th><th scope="col">Tip</th><th scope="col">Săptămâni</th><th scope="col">Coloane</th>'
         '<th scope="col">Folosit de</th><th scope="col"><span class="ip-sr">Acțiuni</span></th></tr></thead>')
# FIN11G (§54): the desktop list's column shares at 1440 (as content-sized before) on the head the loading skeleton reuses;
# None = the column that takes the rest
LCOLS = SKL.widths(LCOLS, [11.8, 6.6, 15.4, None, 24.4, 10.4], [111.8, 52.58, 86.7, 64.59, 258.38, 96.78])  # + loaded min-content px (fix 1)
ORDER = ['g-growth', 'g-laying']   # by name
assert ORDER == sorted(GUIDE, key=lambda g: GUIDE[g])
def lrow(g):
    return (f'<tr><th scope="row" class="s19-fname">{GUIDE[g]}</th><td>{TYPE_RO[TYPE[g]]}</td><td class="ip-num">{weeks_txt(g)}</td>'
            f'<td><span class="s19-secs"><span>{cols_txt(g)}</span><span class="ip-note">fără date în ghid: {missing_txt(g)}</span></span></td>'
            f'<td><span class="s20-uses">{used_txt(g)}</span></td>'
            f'<td class="s19-act"><a class="ip-btn ip-btn--ghost" href="{slug(g)}">Deschide<span class="ip-sr"> {GUIDE[g]}</span></a></td></tr>')
def glist():
    tb = (f'<table class="ip-table ip-table--rowheads s20-table"><caption class="ip-sr">Ghidurile, cu tipul, săptămânile, coloanele și ciclurile '
          f'care le folosesc</caption>{LCOLS}<tbody>{"".join(lrow(g) for g in ORDER)}</tbody></table>')
    return block('h-gh', 'Ghiduri', framed('h-gh', tb),
                 'Un ghid folosit de un ciclu se poate înlocui, dar nu șterge. Comparațiile se fac pe săptămâna ghidului și pe sex.',
                 f'{num(len(ORDER))} ghiduri')
LIST_META = f'Ghidurile de creștere și de ouat, pe săptămâni și sexe{NB}·{NB}Ross 308 creștere, Ross 308 ouat'
def list_page():
    return shell(header(crumbs('Ghiduri'), 'Ghiduri', LIST_META, list_key()), glist())

# ================================================================= detail
def used_block(g, name, cyc):
    if not cyc:   # an unused guide: the EmptyState's parts (unlit lamp, one sentence, the next step) on one line, no padded box
        return block('h-use', 'Folosit de', f'<p class="ip-row s4s-m0">{lamp("off", "nefolosit")}<span>Niciun ciclu nu folosește {name}: '
                     f'îl alegi la o fermă de {TYPE_LO[TYPE[g]]} în <a class="ip-link s20-flink" href="#ferme">Ferme și secțiuni</a>.</span></p>')
    rows = ''
    for c in cyc:
        f = FARM[c['farm_id']]['name']
        a, b = dd(c['start_date']), (dd(c['end_date']) if c['end_date'] else 'azi')
        ded = '<span class="ip-note"> · început dedus</span>' if c.get('start_inferred') else ''
        if c['status'] == 'open':
            st, wk = lamp('ok', 'deschis'), f'SAPT {WK}'
        else:
            st, wk = lamp('off', 'închis'), f'{c["weeks"]} săpt.'
        rows += (f'<tr><th scope="row"><a class="ip-link s20-flink" href="#ferma-{c["farm_id"]}">{f}</a></th><td>{SER[c["series_id"]]}</td><td class="ip-num"><span class="s19-nw">{a}–{b}</span>{ded}</td>'
                 f'<td class="num">{wk}</td><td>{st}</td></tr>')
    tb = (f'<table class="ip-table ip-table--rowheads s20-table"><caption class="ip-sr">Ciclurile care folosesc {name}</caption>'
          '<thead><tr><th scope="col">Fermă</th><th scope="col">Serie</th><th scope="col">Perioada</th><th scope="col" class="num">Săptămâni</th>'
          f'<th scope="col">Ciclu</th></tr></thead><tbody>{rows}</tbody></table>')
    farms = sorted({FARM[c['farm_id']]['name'] for c in cyc})
    k = len(cyc)
    note = f'E și ghidul fermei {", ".join(farms)}: ciclul ei următor îl folosește.'
    return block('h-use', 'Folosit de', framed('h-use', tb, focus=False), note, cyc_words(k))

def spark(g, k, s, last_wk):
    """Small curve (SYS06 .ip-spark, 160×48) of the guide's own figure: x = guide week over the whole guide, so curves of one
    guide line up; y = this series' min…max. Drawn as the curve's figure (.act, ♀/♂ colour; D113): the detail has no
    actuals, the guide is the subject."""
    pts = [(w, v) for w, v in series(g, k, s) if v is not None]
    lo, hi = min(v for _, v in pts), max(v for _, v in pts)
    X = lambda w: 4 + (w - 1) * 152 / (NWK[g] - 1)
    Y = lambda v: 44 - (v - lo) * 40 / (hi - lo)
    pl = ' '.join(f'{X(w):.1f},{Y(v):.1f}' for w, v in pts)
    w1, v1 = pts[-1]
    lbl = next(l for kk, l, *_ in IND if kk == k).lower()
    u = next(u for kk, _, u, _ in IND if kk == k)
    pk = [w for w, v in pts if v == hi]
    pkw = f'sapt. {pk[0]}' if len(pk) == 1 else f'sapt. {pk[0]}–{pk[-1]}'
    wu = lambda v: f'{fmt(v, u)}%' if u == '%' else f'{fmt(v, u)} {u}'   # D46: 70,5%, but 158 g
    al = (f'{lbl.capitalize()} {SEXW[s][2]} în ghid, sapt. {pts[0][0]}–{w1}: de la {wu(pts[0][1])} la {wu(v1)}, '
          f'maxim {wu(hi)} în {pkw}')   # a sentence read aloud: "sapt." (craft §6)
    return (f'<svg class="ip-spark ip-spark--{SEXW[s][1]}" viewBox="0 0 160 48" role="img" aria-label="{al}">'
            f'<polyline class="act" points="{pl}"/><circle class="pt" cx="{X(w1):.1f}" cy="{Y(v1):.1f}" r="2"/></svg>')
def curves_block(g, name):
    rows = ''
    for k, lbl, u, sexes in IND:
        if k in HAS[g]:
            for s in sexes:
                a, b = span(g, k, s)
                vs = [v for _, v in series(g, k, s) if v is not None]
                gp = gaps(g, k, s)
                gtxt = (f'<span class="ip-note s20-gap">– fără date în ghid la sapt. {", ".join(gp)}</span>' if gp else '')
                first = next(v for _, v in series(g, k, s) if v is not None)
                rows += (f'<tr><th scope="row"><span class="ip-sparktable__key">{lbl} {u} {sx(s)}</span></th>'
                         f'<td class="ip-num"><span class="s20-wk"><span class="s19-nw">SAPT {a}–{b}</span>{gtxt}</span></td>'
                         f'<td class="num"><span class="ip-sex ip-sex--{SEXW[s][1]}">{fmt(first, u)}</span></td>'
                         f'<td class="num"><span class="ip-sex ip-sex--{SEXW[s][1]}">{fmt(max(vs), u)}</span></td>'
                         f'<td class="num"><span class="ip-sex ip-sex--{SEXW[s][1]}">{fmt(vs[-1], u)}</span></td>'
                         f'<td class="ip-sparktable__curve">{spark(g, k, s, b)}</td></tr>')
        else:
            rows += (f'<tr><th scope="row"><span class="ip-sparktable__key">{lbl} {u}</span></th>'
                     f'<td colspan="5" class="s20-none">– fără date în ghid</td></tr>')
    tb = (f'<table class="ip-table ip-sparktable"><caption class="ip-sr">{name} pe indicatori: săptămânile cu cifre, prima, maximul, '
          f'ultima și curba</caption><thead><tr><th scope="col">Indicator</th><th scope="col">Săptămâni</th><th scope="col" class="num">Prima</th>'
          f'<th scope="col" class="num">Maxim</th><th scope="col" class="num">Ultima</th><th scope="col">Curba · SAPT 1–{NWK[g]}</th></tr></thead>'
          f'<tbody>{rows}</tbody></table>')
    return block('h-cur', 'Pe indicatori', framed('h-cur', tb), '', f'{num(len(HAS[g]))} din {num(len(IND))} indicatori în ghid')

def weekly_block(g, name, cyc):
    lay = g == 'g-laying'
    opn = [c for c in cyc if c['status'] == 'open']
    today = FARM[opn[0]['farm_id']]['name'] if lay and opn else None   # the laying farm's week today, shaded (Hîrbovăț SAPT 40)
    cols = [(k, lbl, u, sexes) for k, lbl, u, sexes in IND if k in HAS[g]]
    h1 = '<th scope="col" rowspan="2">SAPT</th>' + ('<th scope="col" rowspan="2" class="num">Vârstă</th>' if lay else '')
    h1 += ''.join(f'<th scope="colgroup" colspan="{len(sexes)}">{lbl} {u}</th>' for k, lbl, u, sexes in cols)
    h2 = ''.join(f'<th scope="col" class="num">{sx(s)}</th>' for k, lbl, u, sexes in cols for s in sexes)
    body = ''
    for w in sorted(W[g]):
        r = W[g][w]
        now = today and w == WK
        rh = f'<th scope="row" class="ip-num">{w}{f"<span class=\"ip-note s20-now\">{today} azi</span>" if now else ""}</th>'
        cells = f'<td class="num">{r["age"]}</td>' if lay else ''
        for k, lbl, u, sexes in cols:
            for s in sexes:
                v = r[s].get(k)
                cells += (f'<td class="num"><span class="ip-sex ip-sex--{SEXW[s][1]}">{fmt(v, u)}</span></td>' if v is not None else
                          '<td class="num s20-dash">–<span class="ip-sr"> fără date în ghid</span></td>')
        body += f'<tr{" class=\"s20-today\"" if now else ""}>{rh}{cells}</tr>'
    tb = (f'<table class="ip-table ip-table--rowheads s20-table s20-week"><caption class="ip-sr">{name} pe săptămâni și sexe</caption>'
          f'<thead><tr>{h1}</tr><tr>{h2}</tr></thead><tbody>{body}</tbody></table>')
    gl = []
    for k, lbl, u, sexes in cols:
        for s in sexes:
            gp = gaps(g, k, s)
            if gp: gl.append(f'{lbl.lower()} <span class="ip-sex ip-sex--{SEXW[s][1]}" aria-hidden="true">{SEXW[s][0]}</span><span class="ip-sr">{SEXW[s][2]}</span> la sapt. {", ".join(gp)}')
    note = f'„–” = fără date în ghid: {"; ".join(gl)}.' if gl else ''
    if lay:
        note += ' Ghidul e pe vârstă; SAPT = vârstă − 18 (ciclul de ouat începe la vârsta 19).'
    if today:
        note += f' {today} e azi în sapt. {WK}.'
    return f'<div class="s20-wkcol">{block("h-wk", "Pe săptămâni", framed("h-wk", tb), note.strip(), f"{num(NWK[g])} de săptămâni · {num(NROWS[g])} de rânduri")}</div>'

def detail_keys(g, cyc, pressed=False):
    if not cyc:   # unused: Șterge ghidul is live and opens its Dialog (S20S --sterge draws it pressed)
        p = ' is-pressed' if pressed else ''
        return (f'<div class="s20-keyrow"><a class="ip-btn ip-btn--secondary" href="#inlocuieste">{icon("upload", 16)}Înlocuiește ghidul</a>'
                f'<button type="button" class="ip-btn ip-btn--destructive{p}" aria-haspopup="dialog">Șterge ghidul</button></div>')
    k = len(cyc)
    why = f'Folosit de {"un ciclu" if k == 1 else f"{k} cicluri"}: îl poți înlocui, nu șterge.'
    return (f'<div class="s20-keyrow"><a class="ip-btn ip-btn--secondary" href="#inlocuieste">{icon("upload", 16)}Înlocuiește ghidul</a>'
            f'<button type="button" class="ip-btn ip-btn--destructive" aria-disabled="true" aria-describedby="del-why">Șterge ghidul</button></div>'
            f'<p class="ip-note s4s-m0" id="del-why">{why}</p>')
def detail_meta(g):
    ag = f' · vârsta {W[g][1]["age"]}–{W[g][NWK[g]]["age"]}' if g == 'g-laying' else ''
    return f'{TYPE_RO[TYPE[g]]}{NB}·{NB}SAPT 1–{NWK[g]}{ag}{NB}·{NB}{NROWS[g]} de rânduri, câte unul pe săptămână și sex'
def detail_page(g, focus_h1=False, name=None, cycles=None, pressed=False):
    """The guide detail. name/cycles default to the guide's own (GUIDE, CYC); S20S passes a new guide's name and no cycles
    (the unused-guide detail: compact Folosit de, Șterge ghidul live)."""
    name = name or GUIDE[g]; cyc = CYC[g] if cycles is None else cycles
    hdr = header(crumbs(GH, name), name, detail_meta(g), detail_keys(g, cyc, pressed), tid='g-h1' if focus_h1 else None)
    return shell(hdr, used_block(g, name, cyc) + '\n' + curves_block(g, name) + '\n' + weekly_block(g, name, cyc))

# ================================================================= import (SYS11 GuideImportPreview)
OPT = M.OPT
def tip_seg(sel=None):
    f = M.tip_field(sel=sel)
    a = 'aria-required="true">'; assert f.count(a) == 1
    f = f.replace(a, 'aria-required="true" aria-describedby="gi-tip-h">')
    return f[:-6] + ('<span class="ip-hint" id="gi-tip-h">Tipul alege coloanele citite: creștere pe SAPT, ouat pe vârstă '
                     '(SAPT = vârstă − 18).</span></div>')
def tip_ro(g):
    used = ', '.join(sorted({FARM[c['farm_id']]['name'] for c in CYC[g]}))
    return (f'<div class="ip-field"><label class="ip-label" for="gi-tip">Tip</label><span class="ip-input is-readonly">'
            f'<input id="gi-tip" type="text" readonly="readonly" value="{TYPE_RO[TYPE[g]]}" aria-describedby="gi-tip-h"/></span>'
            f'<span class="ip-hint" id="gi-tip-h">Tipul rămâne {TYPE_LO[TYPE[g]]}: ghidul e folosit de {used}.</span></div>')
def fields(name='', tip='', focus=False):
    return f'<div class="s20-fields">{T.tfield("gi-name", "Nume", name, touch=False, focus=focus)}{tip}</div>'
DROP_P = G['DROP']   # the system FileDrop paste target (additions section 43): focusable group, Ctrl+V pastes into it
assert DROP_P.count('<div class="ip-drop ip-drop--paste" tabindex="0" role="group" aria-label=') == 1
def import_page(cr, title, meta, flds, src_note, gimp):
    main = (block('h-in', 'Ghid', flds) + '\n' +
            block('h-tab', 'Tabelul ghidului', (f'<p class="ip-note s4s-m0">{src_note}</p>' if src_note else '') + gimp))
    return shell(header(cr, title, meta), f'<div class="s20-col">{main}</div>')
gr = G['number'](G['growth_rows']())
lr = G['number'](G['laying_rows']())
er = G['error_rows']()
assert len(gr) == 50 and len(lr) == 94
nbad = sum(1 for r in er if G['reason'](r))
GF = 'Ross 308 creștere.xlsx'
gl = [r['line'] for r in gr if r['v']['body_weight_g'] is None]
INFO_G = (f'Greutatea lipsește la sapt. 24–25 (rândurile {gl[0]}–{gl[-1]}), iar fișierul nu are coloanele '
          'Mortalitate %, Ouat %, Incubabile % și Masă ou g. În ghid apar ca „–” fără date în ghid.')
gle = [r['line'] for r in er if r['v']['body_weight_g'] is None]
INFO_E = INFO_G.replace(f'{gl[0]}–{gl[-1]}', f'{gle[0]}–{gle[-1]}')
INFO_L = ('Ouat % lipsește la vârsta 19–24 și 64–65 (sapt. 1–6 și 46–47), iar tabelul nu are coloanele Greutate g, '
          'Mortalitate %, Incubabile % și Masă ou g. În ghid apar ca „–” fără date în ghid. Ouat % se citește doar pe rândurile <span class="ip-sex ip-sex--hen" aria-hidden="true">♀</span><span class="ip-sr">de găini</span>.')
def repl_note(g, n_new):
    cs = ', '.join(f'{FARM[c["farm_id"]]["name"]} · {SER[c["series_id"]]}' for c in CYC[g])
    return (f'Rândurile citite înlocuiesc toate cele {NROWS[g]} de rânduri de acum ale ghidului. '
            f'Se compară apoi cu ele: {cs}.')
def g_new():
    gimp = G['wrap']('gi', DROP_P + G['foot']('gi', False, 'Scrie numele, alege tipul și adaugă tabelul ghidului.'), 'Import ghid')
    return import_page(crumbs(GH, 'Importă ghid'), 'Importă ghid', 'Un ghid nou, din fișier .xlsx sau .csv sau din tabelul lipit',
                       fields('', tip_seg(), focus=True), '', gimp)
def g_repl_file():
    g = 'g-growth'; p = 'gi'
    gimp = G['wrap'](p, G['src_line'](p, GF, 'fișier .xlsx · tip creștere', G['counts'](len(gr), len(gr), 0))
                     + G['banner']('ok', f'<p><b>Toate cele {len(gr)} de rânduri sunt valide.</b> Verifică cifrele și salvează ghidul.</p>')
                     + G['info_banner'](INFO_G) + G['bar'](p) + G['table'](p, 'growth', gr, GF)
                     + G['foot'](p, True, f'Se salvează {len(gr)} de rânduri în ghidul {GUIDE[g]}.')
                     + G['live'](f'Fișier citit: {len(gr)} de rânduri, toate valide.'), 'Import ghid')
    return import_page(crumbs(GH, (slug(g), GUIDE[g]), 'Înlocuiește'), f'Înlocuiește {GUIDE[g]}', detail_meta(g),
                       fields(GUIDE[g], tip_ro(g)), repl_note(g, len(gr)), gimp)
def g_repl_bad():
    g = 'g-growth'; p = 'gi'
    note = f'{nbad} rânduri refuzate: corectează fișierul și încarcă-l din nou.'
    gimp = G['wrap'](p, G['src_line'](p, GF, 'fișier .xlsx · tip creștere', G['counts'](len(er), len(er) - nbad, nbad))
                     + G['err_banner'](p, er) + G['info_banner'](INFO_E) + G['seg'](p, len(er), nbad, 'all') + G['table'](p, 'growth', er, GF)
                     + G['foot'](p, False, note) + G['live'](f'Fișier citit: {len(er)} de rânduri, {nbad} refuzate.'), 'Import ghid')
    return import_page(crumbs(GH, (slug(g), GUIDE[g]), 'Înlocuiește'), f'Înlocuiește {GUIDE[g]}', detail_meta(g),
                       fields(GUIDE[g], tip_ro(g)), repl_note(g, len(er)), gimp)
def g_repl_paste():
    g = 'g-laying'; p = 'gi'
    gimp = G['wrap'](p, G['src_line'](p, 'Tabel lipit', 'din clipboard · tip ouat', G['counts'](len(lr), len(lr), 0), 'Lipește alt tabel')
                     + G['banner']('ok', f'<p><b>Toate cele {len(lr)} de rânduri sunt valide.</b> Verifică cifrele și salvează ghidul.</p>')
                     + G['info_banner'](INFO_L) + G['bar'](p) + G['table'](p, 'laying', lr, 'tabel lipit')
                     + G['foot'](p, True, f'Se salvează {len(lr)} de rânduri în ghidul {GUIDE[g]}.')
                     + G['live'](f'Tabel citit: {len(lr)} de rânduri, toate valide.'), 'Import ghid')
    return import_page(crumbs(GH, (slug(g), GUIDE[g]), 'Înlocuiește'), f'Înlocuiește {GUIDE[g]}', detail_meta(g),
                       fields(GUIDE[g], tip_ro(g)), repl_note(g, len(lr)), gimp)

# ================================================================= saved (back on the detail, Toast)
FOCUS_NOTE = ('<!-- Focus after Salvează ghidul: the guide detail opens with focus on its title (h1 tabindex=-1); the Toast '
              '(role=status) announces the save and keeps Anulează in the tab order. -->')
def saved():
    g = 'g-growth'
    toast = (f'<div class="ip-toastwrap ip-toastwrap--page"><div class="ip-toast" role="status">{icon("check", 16)}<span class="ip-toast__txt">'
             f'{GUIDE[g]} salvat · {NROWS[g]} de rânduri înlocuite · Telița · Seria 2026 se compară cu ele</span>'
             '<button type="button" class="ip-link ip-toast__undo">Anulează</button></div></div>')
    pg = detail_page(g, focus_h1=True)
    a = '<h1 id="g-h1" tabindex="-1">'; assert pg.count(a) == 1
    pg = pg.replace(a, FOCUS_NOTE + '<h1 id="g-h1" tabindex="-1" class="ip-focus">')
    return f'<div class="s20-clip">{pg}</div>' + toast

# ================================================================= phone list (owner, read-only, D113)
def phone_list():
    rows = ''
    for g in ORDER:
        rows += (f'<tbody><tr class="is-group"><th scope="rowgroup" colspan="2" id="pg-{TYPE[g]}">{GUIDE[g]}</th></tr>'
                 f'<tr><th scope="row" class="ip-table__child">Tip</th><td>{TYPE_RO[TYPE[g]]}</td></tr>'
                 f'<tr><th scope="row" class="ip-table__child">Săptămâni</th><td class="ip-num">{weeks_txt(g)}</td></tr>'
                 f'<tr><th scope="row" class="ip-table__child">Coloane</th><td><span class="s19-secs"><span>{cols_txt(g)}</span>'
                 f'<span class="ip-note">fără date în ghid: {missing_txt(g)}</span></span></td></tr>'
                 f'<tr><th scope="row" class="ip-table__child">Folosit de</th><td><span class="s20-uses">{used_txt(g)}</span></td></tr></tbody>')
    blk = (f'<section class="s4-block" aria-labelledby="h-gh"><div class="s4-head"><h2 class="ip-t-heading" id="h-gh">Ghiduri</h2>'
           f'<span class="ip-note">{num(len(ORDER))} ghiduri</span></div>'
           f'<div class="ip-tablescroll ip-tablescroll--framed" role="region" aria-labelledby="h-gh">'
           f'<table class="ip-table ip-table--rowheads s20-table s19-ptable"><caption class="ip-sr">Ghidurile, cu tipul, săptămânile, '
           f'coloanele și ciclurile care le folosesc</caption>{rows}</table></div>'
           f'<p class="ip-note s4s-m0">Ghidurile le imporți și le înlocuiești pe calculator: tabelul citit are nevoie de un ecran lat.</p></section>')
    cr = crumbs('Ghiduri').replace('<a href="#setari">Setări</a>', '<a href="#setari"><span class="ip-crumbs__txt">Setări</span></a>').replace('<nav class="ip-crumbs"', '<nav class="ip-crumbs ip-crumbs--touch"', 1)  # FIN06: system touch crumb (section 52)
    body = A.phone([f'<header class="ip-pagehead s2-phead s14-phead"><div>{cr}<h1>Ghiduri</h1></div></header>', blk])
    for a, z in (('class="ip-tabbar__item is-current" href="#acasa" aria-current="page"', 'class="ip-tabbar__item" href="#acasa"'),
                 ('class="ip-tabbar__item" href="#mai-mult"', 'class="ip-tabbar__item is-current" href="#mai-mult" aria-current="page"')):
        assert body.count(a) == 1, a; body = body.replace(a, z)
    assert body.count('is-current') == 1
    return body

# ================================================================= CSS (screen-only; tokens only; system classes untouched)
CSS20 = '''
.s20-table{width:100%}
.s20-table thead th{white-space:nowrap}
.s20-table td{vertical-align:middle;white-space:normal}
.s20-uses{display:flex;flex-direction:column;align-items:flex-start;gap:0}
.s20-flink{display:inline-flex;align-items:center;min-height:var(--space-6)} /* the farm link in a row: a 24 px target (as S18 .s18-rlink) */
.s20-use{display:block} /* one cycle = one line of text: lamp + farm, series and week stay whole, a wrap falls only at "·" */
.s20-use .ip-lamp{margin-right:var(--space-1)}
.s20-table:not(.s19-ptable) .s20-use{white-space:nowrap} /* desktop list: each cycle on one line (the phone breaks at "·") */
h1[tabindex="-1"]{margin-block:calc((var(--ip-focus-w) + var(--ip-focus-offset)) * 2)} /* the title focused after a save: its ring clears the crumbs and the meta */
.s20-keys{display:flex;flex-direction:column;align-items:flex-end;gap:var(--space-1)}
.s20-keyrow{display:flex;gap:var(--space-2);align-items:center}
.s20-wkcol{max-width:calc(var(--space-12) * 12)}
.s20-col{display:flex;flex-direction:column;gap:var(--space-6);max-width:calc(var(--space-12) * 18);min-width:0}
.s20-fields{display:flex;flex-direction:column;gap:var(--space-4);max-width:calc(var(--space-12) * 8)}
.s20-fields .ip-seg{align-self:flex-start}
.s20-wk{display:flex;flex-direction:column}
.s20-gap{white-space:nowrap;font-family:var(--font-sans)}
.s20-none{color:var(--ink-muted)}
.s20-dash{color:var(--ink-muted)}
.s20-week th[scope=row]{font-family:var(--font-mono);font-variant-numeric:tabular-nums}
.s20-now{margin-left:var(--space-2);font-family:var(--font-sans);white-space:nowrap}
.s20-week .s20-today > th,.s20-week .s20-today > td{background:var(--sunken)} /* the farm's week today (Hîrbovăț SAPT 40); th beats the row-head raised */
.s20-clip{position:relative;height:''' + str(SC.SIDE_H) + '''px;overflow:hidden}'''
CSS_DESK = M.CSS_DESK + CSS20
CSS_PL = M.CSS_PL + CSS20

# ================================================================= write
pathlib.Path('/tmp/s20').mkdir(exist_ok=True)
def write(stem, title, root, css, w, h):
    r = M.write(stem, title, root, css, w, h)   # S19's writer: the .dc.html + /tmp/s19/<stem>.html; mirror into /tmp/s20
    src = pathlib.Path(f'/tmp/s19/{stem}.html'); (pathlib.Path('/tmp/s20') / src.name).write_text(src.read_text(encoding='utf-8'), encoding='utf-8'); src.unlink()
    return r
STEM, TT = 'S20-ghiduri', 'S20 · Ghiduri'
D39 = 'exemplu de stare (D39)'
out = []
def desk(stem, title, root, h=None):
    h = h or HEIGHTS.get(STEM + stem, 1024)
    out.append(write(STEM + stem, f'{TT}{" · " + title if title else ""}', f'<div class="ip-root s4s-rel" style="padding:0; min-height:{h}px">{root}</div>', CSS_DESK, 1440, h))

desk('', '', list_page())
desk('--ross-308-ouat', 'Ross 308 ouat · detaliu', detail_page('g-laying'))
desk('--ross-308-crestere', 'Ross 308 creștere · detaliu · ghid parțial', detail_page('g-growth'))
desk('--importa-ghid', 'Importă ghid', g_new())
desk('--inlocuieste-fisier', f'Înlocuiește Ross 308 creștere · fișier citit, toate valide · {D39}: fișierul', g_repl_file())
desk('--inlocuieste-refuzate', f'Înlocuiește Ross 308 creștere · rânduri refuzate · {D39}: rândurile 7, 10, 12', g_repl_bad())
desk('--inlocuieste-lipit', f'Înlocuiește Ross 308 ouat · tabel lipit · {D39}: lipirea', g_repl_paste())
desk('--salvat', f'salvat · {D39}: Ross 308 creștere înlocuit cu același fișier', saved(), SC.SIDE_H)
h = HEIGHTS.get(STEM + '--telefon', 844)
out.append(write(STEM + '--telefon', f'{TT} · telefon · doar citire', f'<div class="ip-root" style="padding:0; width:390px; min-height:{h}px">{phone_list()}</div>', CSS_PL, 390, h))
pathlib.Path('/tmp/s20/boards.json').write_text(json.dumps([{k: b[k] for k in ('stem', 'w', 'h')} for b in out]))

if CANVAS:  # S20 row, entries keep x/y, a new row goes below the lowest (canvas_rows.py); 80 px gaps
    sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
    c = CR.load(ROOT)
    x = CR.place(c, STEM, [(b['stem'], b['w'], b['h'], b['title']) for b in out])
    CR.save(ROOT, c)
    print('canvas: S20 row y', CR.row_y(c, STEM), 'ends at x', x - 80)
