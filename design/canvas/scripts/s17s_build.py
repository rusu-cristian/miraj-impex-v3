"""S17S · Istoric modificări · state boards (17.11.2026). Appended to the S17 row (y 30164): new desktops after the 5 S17
desktops (x 7600 …), the S17 phone moved right after them (only its x, as S16S did), the new phone after it.
- din-modificat-de (owner, desktop): Istoric opened from the H4.1 14.11.2026 row's "modificat de Ion Rusu · 15.11.2026, 10:12"
  link (S4 · Păsări Jurnal, S16): one record, every period. No FilterBar (user, farm, period and type don't apply to one
  record); a static info Banner (no live role) names the record, with "Arată toate modificările" (back to the
  default filters, 4 săpt.).
- manager-interval-gresit (D39): a manager assigned only Hîrbovăț (unnamed, as the S14S Telița-only manager): manager Sidebar,
  Setări holds only Istoric modificări, no Fermă select (one farm, as D57), meta "ferma ta: Hîrbovăț"; he is in his own
  Utilizator list as "Manager (tu)" (no name, D96; the Sidebar account says "Manager"). Merged with Interval
  with a date error (FilterInterval preview markup and wording: "Până la 20.11.2026 e în viitor. Alege o dată până la
  16.11.2026."); the list keeps the last valid period.
- exemple-randuri (D39, owner): the S15S same-day edit (Mortalitate H1.1, 16.11.2026, 09:52, Dr. Ceban; D39 old values,
  new values = the sample ♀ 1 · ♂ 0) is the "editat în aceeași zi" and the "mai multe câmpuri" row (merged); pressed on it,
  focus on the real row. Not drawn: "șters cu motiv" (the S16S owner deletion of 16.11 has no time and the owner no name)
  and "șters în aceeași zi" (the S15S vet deletion of H4.2 has no time) (D96).
- observator-marcaj-text: S4 · Fermă · Păsări as the Observator (viewer Sidebar, no header actions, no row keys); the H4.1
  14.11.2026 mark "modificat de Ion Rusu · 15.11.2026, 10:12" is plain text (no access to Istoric, D104). The vet sees the
  same: not drawn. The vet's "Nu ai acces la Istoric modificări." = the S17 observator-fara-acces picture: not drawn.
- telefon-filtre (manager Ion Rusu, 390×844): the S17 phone with the system BottomSheet open (S14P pattern): Perioadă
  (touch SegmentedControl, focus on the chosen period), Fermă, Utilizator (touch Selects), the 7 record-type chips at
  touch height (new system modifier .ip-chip--touch, additions.css section 40) in the screen's chip group (.s4-chips,
  space-2 apart); choices apply at once, Închide / Esc closes.
S17P polish (17.11.2026): Banner one line + "Arată toate modificările", "Manager (tu)", corecție wording, no .s17-fchips.
python3 -I s17s_build.py <repo_root> [--canvas]
Imports s17_build (which rewrites the S17 boards, and through s14s_build the S14S/S2/S7 boards, byte-identical).
Writes design/canvas/project/S17-istoric-modificari--<state>.dc.html, /tmp/s4s|/tmp/s2/<stem>.html and /tmp/s17/boards.json
(S17 boards then these, so s17_render.py names the shots S17-6 …); --canvas adds the 5 boards and moves the S17 phone."""
import sys, re, json, pathlib, importlib.util, datetime as dt

ARGS = sys.argv[1:]
CANVAS = '--canvas' in ARGS; ARGS = [a for a in ARGS if a != '--canvas']
ROOT = pathlib.Path(ARGS[0]); HERE = pathlib.Path(__file__).resolve().parent
HP = HERE / 's17_heights.json'
HEIGHTS = json.loads(HP.read_text()) if HP.exists() else {}

sys.argv = ['s17_build.py', str(ROOT)]
sp = importlib.util.spec_from_file_location('s17_build', HERE / 's17_build.py')
Y = importlib.util.module_from_spec(sp); sp.loader.exec_module(Y)   # rewrites the 6 S17 boards byte-identical
Z, X, A, s4 = Y.Z, Y.X, Y.A, Y.s4
t4 = X.t4
icon, num, n, S = Y.icon, Y.num, Y.n, Y.S
STEM, T = Y.STEM, Y.T
REAL = Y.CH[0]

# ================================================================= D39 row: the S15S same-day edit, several fields
ED_DAY, ED_SEC, ED_AT = dt.date(2026, 11, 16), 'H1.1', dt.datetime(2026, 11, 16, 9, 52)   # S15S "modificat la 09:52" (D39 time from S2S)
NOW = {r['sex']: r['mortality'] for r in S['daily'] if r['date'] == ED_DAY.isoformat() and r['section'] == ED_SEC}
assert NOW == {'F': 1, 'M': 0}, NOW                                   # the record as it is now (sample)
OLD = {'F': 2, 'M': 1}                                                 # D39: the values before the edit (example)
SAVE = next(e for e in S['event_log'] if e['record'] == 'mortality' and e['at'][:10] == ED_DAY.isoformat())
assert SAVE['user'] == 'Dr. Ceban' and SAVE['at'] < ED_AT.isoformat(timespec='minutes')  # 08:34 save, then the 09:52 edit
ED = dict(rec='mortality', sec=ED_SEC, sx='F', day=ED_DAY, at=ED_AT, why=None, user='Dr. Ceban', role='veterinar', kind='editat', farm='Hîrbovăț',
          vals=[('găini moarte', n(OLD['F']), n(NOW['F'])), ('cocoși morți', n(OLD['M']), n(NOW['M']))])
def chg_row(c, state=''):
    """Y.chg_row (the SYS09 preview markup) + a both-sexes record link, the 'no reason' line and a demo state class."""
    r = Y.chg_row(dict(c, why=c['why'] or '–'))
    if c['why'] is None:
        a = '<p class="ip-chg__why"><span class="ip-chg__whyk">Motiv: </span>–</p>'; assert r.count(a) == 1
        r = r.replace(a, '<p class="ip-chg__why is-none"><span class="ip-chg__whyk">Motiv: </span>– în aceeași zi, fără motiv</p>')
    if len(c['vals']) == 2:
        d = f"{c['day']:%d.%m.%Y}"
        a = f'<span class="ip-sex ip-sex--hen" aria-hidden="true">♀</span><span class="ip-sr"> găini, ziua {d}</span>'; assert r.count(a) == 1
        r = r.replace(a, f'<span class="ip-ev__sexes" aria-hidden="true"><span class="ip-sex ip-sex--hen">♀</span> și <span class="ip-sex ip-sex--rooster">♂</span></span>'
                         f'<span class="ip-sr">găini și cocoși, ziua {d}</span>')
    if state:
        r = r.replace('<li class="ip-ev ip-ev--change">', f'<li class="ip-ev ip-ev--change {state}">', 1)
    return r
assert chg_row(REAL) == Y.chg_row(REAL)

def block(rows, note, foot=Y.FOOT):
    lst = f'<ul class="ip-list ip-list--change" aria-labelledby="h-chg">{Y.COLS}{"".join(rows)}</ul>'
    return (f'<section class="s4-block s17-block" aria-labelledby="h-chg"><div class="ip-list__head"><h2 class="ip-t-heading" id="h-chg">Modificări</h2>'
            f'<p class="ip-note">{note}</p></div>{lst}<p class="ip-note s4s-m0">{foot}</p></section>')
def shell(sb, parts): return f'<div class="s4-shell">{sb}<main class="s4-main">\n' + '\n'.join(parts) + '\n</main></div>'

# ================================================================= 1 · din „modificat de” (owner): one record, all periods
REC_TXT = f'Mortalitate · {REAL["sec"]} · ziua {REAL["day"]:%d.%m.%Y} · {REAL["farm"]}'
def page_record():
    ban = (f'<div class="ip-banner ip-banner--info ip-banner--act ip-banner--1l">{icon("info", 16)}'   # static: present on load, no live role (S17P)
           f'<div class="ip-banner__body"><p class="s4s-m0">Vezi doar modificările înregistrării Mortalitate · {REAL["sec"]} · ziua <span class="ip-num">{REAL["day"]:%d.%m.%Y}</span> · {REAL["farm"]}, din tot istoricul.</p></div>'
           '<a class="ip-btn ip-btn--secondary ip-banner__key" href="#istoric-modificari">Arată toate modificările</a></div>')   # back to the 4 săpt. default
    note = f'1 modificare · Mortalitate {REAL["sec"]} · ziua {num(format(REAL["day"], "%d.%m.%Y"))} · tot istoricul'
    foot = ('Istoricul unei înregistrări arată fiecare corecție sau ștergere a ei, cu valorile vechi și noi. '
            'Rândul deschide secțiunea în ziua înregistrării.')
    return shell(Y.sidebar_owner(), [Y.header(), ban, block([chg_row(REAL)], note, foot)])

# ================================================================= 2 · manager, doar Hîrbovăț (D39) + Interval cu dată greșită
def sidebar_manager():
    sb = s4.sidebar('Manager').replace('class="ip-nav is-current" href="#serie" aria-current="page"', 'class="ip-nav" href="#serie"')
    par = f'<a class="ip-nav" href="#setari">{icon("settings")}Setări</a>'; assert sb.count(par) == 1
    sb = sb.replace(par, par + '<a class="ip-nav ip-nav--sub is-current" href="#istoric-modificari" aria-current="page">Istoric modificări</a>', 1)
    assert sb.count('is-current') == 1
    return sb
IV_A, IV_B = f'{Y.P0:%d.%m.%Y}', '20.11.2026'
IV_ERR = f'Până la {IV_B} e în viitor. Alege o dată până la {Y.TODAY:%d.%m.%Y}.'          # FilterInterval preview wording (D37)
def interval():
    b = s4.datef('iv-b', 'până la', IV_B)
    for a, z in (('<span class="ip-input">', '<span class="ip-input is-error">'),
                 ('aria-describedby="iv-b-fmt"/>', 'aria-describedby="iv-b-fmt iv-b-err" aria-invalid="true"/>')):
        assert b.count(a) == 1, a; b = b.replace(a, z)
    return f'<span class="ip-interval" role="group" aria-label="Interval">{s4.datef("iv-a", "de la", IV_A)}{b}</span>'
def page_manager():
    users = list(Y.USER_OPTS); users.insert(users.index('Ion Rusu · manager') + 1, 'Manager (tu)')   # the D39 manager himself, unnamed (D96)
    fb = (f'<div class="ip-filterbar s11-filterbar s4-filterbar" role="group" aria-label="Filtre">{s4.seg("Perioadă", Y.PER, 3)}{interval()}'
          f'{Y.select("f-user", "Utilizator", users, users[0])}{Y.chips()}<p class="ip-err ip-filterbar__err" id="iv-b-err">{IV_ERR}</p></div>')
    h = Y.header()
    a = f'·{Y.NB2}{Y.SCOPE}</span>'; assert h.count(a) == 1
    h = h.replace(a, f'·{Y.NB2}ferma ta:{Y.NB2}Hîrbovăț</span>')
    note = f'1 modificare · {num(Y.PER_TXT)} · perioada de dinainte, până corectezi data · cele mai noi întâi'
    return shell(sidebar_manager(), [h, fb, block([chg_row(REAL)], note, Y.FOOT + ' Vezi doar fermele tale.')])

# ================================================================= 3 · rânduri exemplu (D39, owner): editat în aceeași zi, mai multe câmpuri; apăsat, focus
def page_examples():
    note = f'2 modificări · {num(Y.PER_TXT)} · cele mai noi întâi'
    rows = sorted([(ED, 'is-pressed'), (REAL, 'ip-focus')], key=lambda x: x[0]['at'], reverse=True)
    return shell(Y.sidebar_owner(), [Y.header(), Y.filterbar(), block([chg_row(c, st) for c, st in rows], note)])

# ================================================================= 4 · Observator: S4 · Fermă · Păsări, the mark as plain text
MARK_LINK = re.compile(r'<a class="ip-link" href="#istoric-H4\.1-14\.11">(modificat de .*?)<span class="ip-sr">, vezi istoricul</span></a>')
ACTS = re.compile(r'<span class="ip-ev__acts">.*?<p class="ip-note" id="ev\d+-n">[^<]*</p></div></span>')
def page_observer():
    t4.EVID[0] = 0
    pg = t4.page('pasari', t4.panel_pasari())
    sb = s4.sidebar()
    assert pg.count(sb) == 1; pg = pg.replace(sb, s4.sidebar('Observator', viewer=True))
    pg, k = re.subn(r'\n<div class="ip-row">\n<button type="button" class="ip-btn ip-btn--secondary">.*?</div>\n</header>', '\n</header>', pg, count=1, flags=re.S); assert k == 1
    pg, k = ACTS.subn('', pg); assert k == t4.PAGE + len(t4.sexerr) + len(t4.xfer_days()), k   # viewer: no keys on any row (SYS09)
    pg, k = MARK_LINK.subn(r'<span>\1</span>', pg); assert k == 1
    assert 'ip-ev__acts' not in pg and 'ip-ev__more' not in pg and '#istoric' not in pg and 'Mută păsări' not in pg and '#setari' not in pg
    return pg

# ================================================================= 5 · telefon · filtre (manager Ion Rusu): the BottomSheet open
def p_sheet():
    seg = s4.seg('Perioadă', Y.PER, 0)
    a = '<span class="ip-seg" role="radiogroup" aria-label="Perioadă">'; assert seg.startswith(a)
    seg = ('<span class="ip-seg ip-seg--touch" role="radiogroup" aria-labelledby="pf-per">' + seg[len(a):]).replace(
        '<span class="ip-seg__item is-selected"', '<span class="ip-seg__item is-selected ip-focus"', 1)   # focus starts on the chosen period
    def sel(i, label, opts):
        o = ''.join(f'<option{" selected=\"selected\"" if j == 0 else ""}>{t}</option>' for j, t in enumerate(opts))
        return (f'<div class="ip-field"><label class="ip-label" for="{i}">{label}</label><span class="ip-input ip-input--touch s11-select"><select id="{i}">{o}</select>'
                f'{icon("chev", 16)}</span></div>')
    chips = ''.join(s4.chip(k, w).replace('class="ip-chip ', 'class="ip-chip ip-chip--touch ', 1) for k, w in Y.TYPES)
    assert chips.count('ip-chip--touch') == len(Y.TYPES) == 7
    typ = (f'<div class="ip-field"><span class="ip-label" id="pf-typ">Tipuri de înregistrări</span>'
           f'<div class="s4-chips" role="group" aria-labelledby="pf-typ">{chips}</div></div>')   # .s4-chips: wrapping, space-2 apart
    return ('<div class="ip-scrim ip-scrim--sheet">'
            '<div class="ip-sheet s14-fsheet" role="dialog" aria-modal="true" id="pf-sheet" aria-labelledby="pf-t" aria-describedby="pf-d">'
            '<div class="ip-sheet__handle" aria-hidden="true"></div>'
            '<div class="ip-sheet__head"><div><h2 class="ip-t-heading ip-sheet__title" id="pf-t">Filtrează</h2>'
            '<p class="ip-note ip-sheet__sub" id="pf-d">Lista se schimbă la fiecare alegere.</p></div>'
            f'<button type="button" class="ip-iconbtn ip-iconbtn--touch" aria-label="Închide filtrele"><svg class="ip-icon" viewBox="0 0 24 24" aria-hidden="true">{Z.X_ICON}</svg></button></div>'
            f'<div class="s14-fsheet__body"><div class="ip-field"><span class="ip-label" id="pf-per">Perioadă</span>{seg}</div>'
            f'{sel("pf-farm", "Fermă", ["Toate fermele", "Hîrbovăț", "Telița"])}{sel("pf-user", "Utilizator", Y.USER_OPTS)}{typ}</div>'
            '</div></div>')
def phone_sheet():  # as s14s_build.phone_sheet: the page (at its top) and the tab bar inert under the scrim, the sheet at the foot
    b = Y.phone_page()
    for a, z in (('aria-haspopup="dialog" aria-expanded="false"', 'aria-haspopup="dialog" aria-expanded="true" aria-controls="pf-sheet"'),
                 ('<div class="s7-screen s2-phone s14-phone">', '<div class="s7-screen s2-phone s14-phone s7s-vp s9-modal">'),
                 ('<main class="s7-main s2-main">', '<div class="s7s-view" inert="inert"><main class="s7-main s2-main">'),
                 ('</main>', '</main></div>'), ('<nav class="ip-tabbar', '<nav inert="inert" class="ip-tabbar')):
        assert b.count(a) == 1, a; b = b.replace(a, z)
    assert b.endswith('</div>')
    return b[:-len('</div>')] + p_sheet() + '</div>'

CSS_SHEET = Z.CSS_SHEET + Y.CSS17

# ================================================================= boards
DESK = [(STEM + '--din-modificat-de', f'{T} · din „modificat de” · o înregistrare', page_record, Y.CSS_D, 1024),
        (STEM + '--manager-interval-gresit', f'{T} · manager, doar Hîrbovăț · interval cu dată greșită · exemplu de stare (D39)', page_manager, Y.CSS_D, 1024),
        (STEM + '--exemple-randuri', f'{T} · rânduri exemplu: editat în aceeași zi, mai multe câmpuri · apăsat · focus · exemplu de stare (D39)', page_examples, Y.CSS_D, 1024),
        (STEM + '--observator-marcaj-text', f'{T} · observator · „modificat de” ca text (S4 · Fermă · Păsări)', page_observer, t4.CSS, t4.HEIGHTS['S4-ferma--pasari'])]
PHONE = [(STEM + '--telefon-filtre', f'{T} · manager · telefon · filtre', phone_sheet)]
out = []
for stem, title, fn, css, h0 in DESK:
    h = HEIGHTS.get(stem, h0); s4.write(stem, title, fn(), '', H=h, css=css); out.append((stem, title, h, 1440))
for stem, title, fn in PHONE:
    A.HEIGHTS[stem] = HEIGHTS.get(stem, 844); out.append(A.write_phone(stem, title, fn(), css=CSS_SHEET))
prev = json.loads(pathlib.Path('/tmp/s17/boards.json').read_text())
assert [b['stem'] for b in prev] == [s for s, _, _ in Y.DESK] + [s for s, _, _ in Y.PHONE]
pathlib.Path('/tmp/s17/boards.json').write_text(json.dumps(prev + [{'stem': s, 'w': w, 'h': h} for s, _, h, w in out]))

if CANVAS:  # S17 row: S17 desktops, these desktops, the S17 phone, these phones
    sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
    c = CR.load(ROOT)
    new = lambda W: [(s, w, h, t) for s, t, h, w in out if w == W]
    x = CR.place(c, STEM, [s for s, _, _ in Y.DESK] + new(1440) + [s for s, _, _ in Y.PHONE] + new(390), 0)
    CR.save(ROOT, c)
    print('canvas: S17 row y', CR.row_y(c, STEM), 'ends at x', x - 80)
