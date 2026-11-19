"""S14S · Rapoarte · Producție states (16.11.2026). Appended to the S14 row (y 23061) after its 5 boards, desktops then the phone:
- observator · se exportă (desktop): the default page for the Observator role (Sidebar Acasă · Serie · Rapoarte, account
  "Observator", D91/D92: same content, Producție has no prices) with Exportă .xlsx running (spinner in the key, label kept,
  a hidden role=status line names the file's filters). Merged: Observator parity board + export in progress.
- eroare la încărcare (desktop): crumbs + title, the FilterBar the user chose (export off: nothing to export), fault Banner
  naming what failed + Reîncearcă (online only, spec §1).
- fără fermă de ouat · exemplu de stare (D39) (desktop): a manager assigned only Telița (growth, no eggs) opens Producție:
  EmptyState with the real Telița cycle and a step to Rapoarte › Hrană. "Nu ai acces la această fermă." is the same picture
  (lamp "fără acces", that sentence, Înapoi la Acasă as S4 fara-acces), so it is not drawn twice.
- telefon (390, owner): FilterBar collapsed to one key (FilterBar spec, BottomSheet closed), Readout without the GuideChart
  (its 540-wide drawing would set axis text under 11px at 358), Stats, then the weekly table in its scroll box with the
  sticky first column (DataTable spec · phone), the whole tall page with the BottomTabBar (Rapoarte current) at its foot.
Reuses by import: s14_build (page, header, filterbar, top, tb_weeks, CSS), s11_build (sidebars), s2_build (phone screen,
BottomTabBar, phone CSS, write_phone; importing it rewrites the S2/S7 default boards byte-identical), s4s_build (empty_box, write).
python3 -I s14s_build.py <repo_root> [s14s_heights.json] [--canvas]
Writes design/canvas/project/S14-rapoarte-productie--<state>.dc.html, /tmp/s4s/<stem>.html (desktops), /tmp/s2/<stem>.html (phone);
--canvas appends/updates the 4 boards in canvas.json (S14 row, after the existing 5; nothing else moves)."""
import sys, re, json, pathlib, importlib, importlib.util

ARGS = sys.argv[1:]
CANVAS = '--canvas' in ARGS; ARGS = [a for a in ARGS if a != '--canvas']
ROOT = pathlib.Path(ARGS[0]); HERE = pathlib.Path(__file__).resolve().parent
HP = pathlib.Path(ARGS[1]) if len(ARGS) > 1 else HERE / 's14s_heights.json'
HEIGHTS = json.loads(HP.read_text()) if HP.exists() else {}

sys.argv = ['s14_build.py', str(ROOT), str(HERE / 's14_heights.json')]
sp = importlib.util.spec_from_file_location('s14_build', HERE / 's14_build.py')
X = importlib.util.module_from_spec(sp); sp.loader.exec_module(X)   # writes nothing on import
sys.path.insert(0, str(HERE))
import sstate_css as SS                                              # S14P: phone viewport + modal anchor for the sheet board
sys.argv = ['s2_build.py', str(ROOT), str(HERE / 's2_heights.json')]
A = importlib.import_module('s2_build')                              # rewrites the S2 default boards (and S7) unchanged
s4, s11, t4, num, icon = X.s4, X.s11, X.t4, X.num, X.icon
STEM = X.STEM

# ================================================================= desktop · observator · se exportă
EXP = s11.export()
assert X.filterbar().count(EXP) == 1
EXP_RUN = ('<button type="button" class="ip-btn ip-btn--ghost ip-filterbar__end" aria-busy="true" aria-disabled="true" aria-describedby="exp-st">'  # Button · loading (system preview): spinner, label kept; busy + aria-disabled while it runs (LoginCard busy pattern: repeat presses ignored)
           '<span class="ip-spin" aria-hidden="true"></span>Exportă .xlsx</button>'
           f'<span class="ip-sr" role="status" id="exp-st">Se pregătește fișierul .xlsx: Producție Hîrbovăț, ciclul curent, sapt. {X.WEEKS_ALL[0]}–{X.WK}.</span>')
# S14P: the busy key's look (variant colours kept over the aria-disabled grey, cursor:progress) is additions.css section 37
def page_viewer_export():
    pg = X.page()
    assert pg.count(X.sidebar()) == 1 and pg.count(EXP) == 1
    return pg.replace(X.sidebar(), s11.sidebar('Observator')).replace(EXP, EXP_RUN)

# ================================================================= desktop · eroare la încărcare
def head_bare():  # crumbs + title only: the cycle line comes with the report, which did not load
    h = X.header()
    h2, k = re.subn(r'\n<div class="ip-row ip-pagehead__meta">.*?</div>\n</div>', '\n</div>', h, count=1, flags=re.S); assert k == 1
    return h2
FAULT = '<circle cx="12" cy="12" r="10"/><path d="M12 8v4"/><path d="M12 16h.01"/>'  # Banner fault icon (system preview, as S2S)
def page_error():
    fb = X.filterbar().replace(EXP, s11.export(True))
    ban = (f'<div class="ip-banner ip-banner--fault ip-banner--act" role="alert"><svg class="ip-icon ip-icon--16" viewBox="0 0 24 24" aria-hidden="true">{FAULT}</svg>'
           '<div class="ip-banner__body"><p class="ip-m0">Raportul Producție pentru Hîrbovăț, ciclul curent, nu s-a încărcat.</p>'
           '<p class="ip-m0">Verifică conexiunea la internet, apoi apasă Reîncearcă.</p></div>'
           '<button type="button" class="ip-btn ip-btn--secondary ip-banner__key">Reîncearcă</button></div>')
    return f'<div class="ip-shell">{X.sidebar()}<main class="ip-shell__main">\n{head_bare()}\n{fb}\n<div class="s14-page">{ban}</div>\n</main></div>'

# ================================================================= desktop · fără fermă de ouat (manager with Telița only), D39
cT = A.cT
assert (cT['status'], cT['end_date'], cT['weeks']) == ('closed', '2026-02-28', 21) and A.TL['totals']['all'] == 0
def page_nolay():
    lines = ['Producția (ouă, ouat, incubabile) se raportează pe fermele de ouat; ferma ta, Telița, e de creștere.',
             f'Ultimul ciclu la Telița: Seria 2026, creștere, {num(cT["weeks"])} săpt., încheiat pe {num(s4.ddate(cT["end_date"]))}. Consumul și stocul de hrană sunt în Rapoarte › Hrană.']
    step = '<a class="ip-btn ip-btn--secondary" href="#rapoarte-hrana">Deschide raportul Hrană</a>'
    return (f'<div class="ip-shell">{s11.sidebar("Manager")}<main class="ip-shell__main">\n{head_bare()}\n'
            f'<div class="s14-page">{s4.empty_box("fără ouat", lines, step, "Producție")}</div>\n</main></div>')

# ================================================================= phone · default (owner, 16.11.2026)
SLIDERS = ('<path d="M21 4h-7"/><path d="M10 4H3"/><path d="M21 12h-9"/><path d="M8 12H3"/><path d="M21 20h-5"/><path d="M12 20H3"/>'
           '<path d="M14 2v4"/><path d="M8 10v4"/><path d="M16 18v4"/>')  # Lucide sliders-horizontal
def p_filter():  # FilterBar on phone: one key opening a BottomSheet (period, farm, export); closed here
    return ('<div class="ip-filterbar s14-pfilter" role="group" aria-label="Filtre">'
            '<button type="button" class="ip-btn ip-btn--secondary ip-btn--touch" aria-haspopup="dialog" aria-expanded="false" aria-describedby="pf-sum">'
            f'<svg class="ip-icon ip-icon--16" viewBox="0 0 24 24" aria-hidden="true">{SLIDERS}</svg>Filtrează sau exportă</button>'
            '<span class="ip-note" id="pf-sum">Ciclul curent · Hîrbovăț</span></div>')
def p_readout():  # the Readout without its GuideChart (chart, keyboard help, live read, hidden table, legend)
    ro = t4.ro_productie()
    for pat in (r'<svg class="ip-chart".*?</svg>\n?', r'<span class="ip-sr" id="lay-help">.*?</span>\n?', r'<div class="ip-chart__read".*?</div>\n?',
                r'<table class="ip-sr">.*?</table>\n?', r'<div class="ip-chart__legend">.*?</span></div>\n?'):
        ro, k = re.subn(pat, '', ro, count=1, flags=re.S); assert k == 1, pat
    assert 'ip-chart' not in ro and 'lay-' not in ro
    return ro
CRUMB = '<a href="#rapoarte">Rapoarte</a>'
def p_header():  # S14P/FIN06: system touch crumb (additions.css section 52): the word's span takes the ring, the link the 46px tap area
    h = X.header().replace('<header class="ip-pagehead">', '<header class="ip-pagehead s2-phead s14-phead">', 1)
    assert h.count(CRUMB) == 1
    assert h.count('<nav class="ip-crumbs"') == 1  # FIN06: system touch crumb (section 52)
    return h.replace(CRUMB, '<a href="#rapoarte"><span class="ip-crumbs__txt">Rapoarte</span></a>').replace('<nav class="ip-crumbs"', '<nav class="ip-crumbs ip-crumbs--touch"', 1)
def phone_page(sheet=False):
    top = t4.top('h-key', 'Producție față de ghid, ciclul curent', p_readout(), X.stats_cur())
    pf = p_filter()
    if sheet:  # the key that opened the sheet: expanded, controls it
        k = 'aria-haspopup="dialog" aria-expanded="false"'; assert pf.count(k) == 1
        pf = pf.replace(k, 'aria-haspopup="dialog" aria-expanded="true" aria-controls="pf-sheet"')
    body = A.phone([p_header(), pf, f'<div class="s14-page">{top}{X.tb_weeks(phone=True)}</div>'])
    body = body.replace('class="ip-tabbar__item is-current" href="#acasa" aria-current="page"', 'class="ip-tabbar__item" href="#acasa"', 1)
    body = body.replace('class="ip-tabbar__item" href="#rapoarte"', 'class="ip-tabbar__item is-current" href="#rapoarte" aria-current="page"', 1)
    assert body.count('is-current') == 1 and 'href="#rapoarte" aria-current="page"' in body
    return body.replace('<div class="ip-screen">', '<div class="ip-screen s14-phone">', 1)

# ================================================================= phone · filters open (S14P): the system BottomSheet over the page
X_ICON = '<path d="M18 6 6 18"/><path d="m6 6 12 12"/>'  # Lucide x (as the S9P sheet's Închide)
def p_sheet():
    seg = s4.seg('Perioadă', s4.PERIODS, 0)
    a = '<span class="ip-seg" role="radiogroup" aria-label="Perioadă">'; assert seg.startswith(a)
    seg = ('<span class="ip-seg ip-seg--touch" role="radiogroup" aria-labelledby="pf-per">' + seg[len(a):]).replace(
        '<span class="ip-seg__item is-selected"', '<span class="ip-seg__item is-selected ip-focus"', 1)  # focus starts on the chosen period
    # FIN12 (D130): the section 53 Select on touch: opens its own BottomSheet (Telița listed disabled, X.TEL_WHY), no list in the page
    farm = ('<div class="ip-field"><span class="ip-label" id="pf-farm-l">Fermă</span><div class="ip-select">'
            '<div class="ip-input ip-select__btn ip-input--touch" id="pf-farm" role="combobox" tabindex="0" aria-haspopup="dialog" aria-expanded="false" aria-labelledby="pf-farm-l">'
            f'<span class="ip-select__val">Hîrbovăț</span>{icon("chev", 16)}</div></div></div>')
    exp = ('<div class="s14-fsheet__exp"><button type="button" class="ip-btn ip-btn--secondary ip-btn--touch" aria-describedby="pf-exp">'
           f'{icon("download", 16)}Exportă .xlsx</button>'
           f'<p class="ip-hint" id="pf-exp">Tabelul pe săptămâni și hale cu filtrele de mai sus: Hîrbovăț, ciclul curent, <span class="ip-num">sapt. {X.WEEKS_ALL[0]}–{X.WK}</span>.</p></div>')
    return ('<div class="ip-scrim ip-scrim--sheet">'
            '<div class="ip-sheet s14-fsheet" role="dialog" aria-modal="true" id="pf-sheet" aria-labelledby="pf-t" aria-describedby="pf-d">'
            '<div class="ip-sheet__handle" aria-hidden="true"></div>'
            '<div class="ip-sheet__head"><div><h2 class="ip-t-heading ip-sheet__title" id="pf-t">Filtrează sau exportă</h2>'
            '<p class="ip-note ip-sheet__sub" id="pf-d">Raportul se schimbă la fiecare alegere.</p></div>'
            f'<button type="button" class="ip-iconbtn ip-iconbtn--touch" aria-label="Închide filtrele"><svg class="ip-icon" viewBox="0 0 24 24" aria-hidden="true">{X_ICON}</svg></button></div>'
            f'<div class="s14-fsheet__body"><div class="ip-field"><span class="ip-label" id="pf-per">Perioadă</span>{seg}</div>{farm}{exp}</div>'
            '</div></div>')
def phone_sheet():  # 390×844: the page (scrolled to its top) and tab bar inert under the scrim, the sheet at the foot
    b = phone_page(sheet=True)
    for a, z in (('<div class="ip-screen s14-phone">', '<div class="ip-screen s14-phone ip-screen--fixed ip-scrimhost">'),
                 ('<main class="ip-screen__main s2-main">', '<div class="ip-screen__view" inert="inert"><main class="ip-screen__main s2-main">'),
                 ('</main>', '</main></div>'), ('<nav class="ip-tabbar', '<nav inert="inert" class="ip-tabbar')):
        assert b.count(a) == 1, a; b = b.replace(a, z)
    assert b.endswith('</div>')
    return b[:-len('</div>')] + p_sheet() + '</div>'

CSS_PHONE = A.CSS_P + '\n' + X.CSS + '''
.s14-phone .s4-top,.s14-phone .s4-stats{grid-template-columns:minmax(0,1fr)} /* phone: Readout, then the Stats one per line */
.s14-phone .ip-table--rowheads thead th:first-child{white-space:normal} /* FW04 (D150): the week's dates stay beside SAPT (were under it: week rows 46 beside 32 hall rows); the sticky column is as wide as that line */
.s14-phone .ip-sechead{flex-direction:column;align-items:flex-start;gap:var(--space-1)}
.s14-phone .s4t-foot{flex-direction:column;align-items:flex-start;gap:var(--space-2)}
.s14-phone .s4t-lk{min-height:var(--control-h-touch)} /* touch: 48px target */
.s14-phead .ip-pagehead__meta{margin-top:var(--space-2)}
.s14-pfilter{gap:var(--space-2) var(--space-3)}'''
CSS_SHEET = CSS_PHONE + SS.VIEW + SS.MODAL + '''
.s14-fsheet{display:flex;flex-direction:column;min-height:0;max-height:100%;padding-bottom:0} /* S14P: phone FilterBar sheet, as the section 33 sheet */
.s14-fsheet__body{flex:1;min-height:0;overflow:auto;display:flex;flex-direction:column;gap:var(--space-4);padding:var(--space-4);border-top:var(--ip-hair) solid var(--rule)} /* fields space-4 apart, as every sheet and panel form (section 62), FX03 */
.s14-fsheet .ip-seg{display:flex}
.s14-fsheet .ip-seg__item{flex:1 1 auto;justify-content:center}
.s14-fsheet .s11-select{width:100%}
.s14-fsheet__exp{display:flex;flex-direction:column;gap:var(--space-2);padding-top:var(--space-4);border-top:var(--ip-hair) solid var(--rule)}
.s14-fsheet__exp > .ip-btn{width:100%}
.s14-fsheet__exp > .ip-hint{margin:0}'''

# ================================================================= boards (x after the 5 S14 boards: 7600, desktops then the phone)
DESK = [
 (STEM + '--observator-se-exporta', 'S14 · Rapoarte · Producție · observator · se exportă', page_viewer_export),
 (STEM + '--eroare-la-incarcare', 'S14 · Rapoarte · Producție · eroare la încărcare', page_error),
 (STEM + '--fara-ferma-de-ouat', 'S14 · Rapoarte · Producție · fără fermă de ouat · exemplu de stare (D39)', page_nolay),
]
PHONE = [(STEM + '--telefon', 'S14 · Rapoarte · Producție · telefon', phone_page),
         (STEM + '--telefon-filtre', 'S14 · Rapoarte · Producție · telefon · filtre și export', phone_sheet)]
out = []
for stem, title, fn in DESK:
    h = HEIGHTS.get(stem, 1024)
    s4.write(stem, title, fn(), '', H=h, css=X.CSS); out.append((stem, title, h, 1440))
for stem, title, fn in PHONE:
    A.HEIGHTS[stem] = HEIGHTS.get(stem, 844)
    out.append(A.write_phone(stem, title, fn(), css=CSS_SHEET if fn is phone_sheet else CSS_PHONE, fixed=fn is phone_sheet))

if CANVAS:  # S14 row: these boards after the 5 S14 boards
    sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
    c = CR.load(ROOT)
    X0 = max(c['boards'][f'{s}.dc.html']['x'] + c['boards'][f'{s}.dc.html']['w'] for s, _, _ in X.BOARDS) + 80
    x = CR.place(c, STEM, [(s, w, h, t) for s, t, h, w in out], X0)
    CR.save(ROOT, c)
    print('canvas: S14 row y', CR.row_y(c, STEM), 'ends at x', x - 80)
