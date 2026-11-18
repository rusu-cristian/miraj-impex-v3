"""S17 · Istoric modificări (spec E2; nav map: Setări › Istoric modificări, manager+). Desktop-led (an admin review page,
like S14): desktops first, then the phone. Today = 16.11.2026 (D23/D24). Real history only: the extracts hold one change,
sample-hirbovat-2026.json › audit_log (Mortalitate H4.1 ♀ 14.11.2026, 3 → 5, Ion Rusu, 15.11.2026, 10:12, real reason);
event_log has no other update/delete. No D39 rows (the S16S deletion has no time and no owner name, D96).
Boards (row below S16, y 30184):
- default (owner, desktop): Setări expanded in the Sidebar, crumbs Setări › Istoric modificări, FilterBar (Perioadă 4 săpt.,
  Fermă, Utilizator, record-type chips for the types with records in the period, D45), ChangeRow list (SYS09) with its column
  head, the one real row, a foot note saying what the log keeps.
- filtre-fara-rezultate: Utilizator = Dr. Ceban (no change by her in the extracts) → li.ip-list__empty (SYS09).
- se-incarca: FilterBar kept, list aria-busy with ip-ev--skel rows, hidden role=status.
- eroare-la-incarcare: FilterBar kept, fault Banner + Reîncearcă in place of the list (S14S/S12 pattern).
- observator-fara-acces: Observator Sidebar (no Setări), title only, EmptyState "fără acces" (S4 fara-acces pattern).
- telefon (manager Ion Rusu, 390×844): FilterBar folded into one key (S14S phone pattern), stacked ChangeRow, tab bar
  Mai mult current.
Reuses by import: s14s_build (phone filter key, Banner fault icon, phone CSS; imports s14_build, s2_build, sstate_css and
rewrites the S14S/S2/S7 boards byte-identical), s14_build (desktop CSS: S4 shell + S11 farm select), s4s_build (sidebar,
seg, chip, icons, empty_box, skeleton, write), s2_build (phone, write_phone). No new system class.
S17P polish (17.11.2026): wording Corecții/managerilor, meta · rule (S2P), foot note measure, ♀ hidden in the row link name.
python3 -I s17_build.py <repo_root> [s17_heights.json] [--canvas]
Writes design/canvas/project/S17-istoric-modificari[--<state>].dc.html, /tmp/s4s|/tmp/s2/<stem>.html, /tmp/s17/boards.json;
--canvas adds the 6 boards to canvas.json (new row, nothing else changes)."""
import sys, re, json, pathlib, importlib.util, datetime as dt

ARGS = sys.argv[1:]
CANVAS = '--canvas' in ARGS; ARGS = [a for a in ARGS if a != '--canvas']
ROOT = pathlib.Path(ARGS[0]); HERE = pathlib.Path(__file__).resolve().parent
HP = pathlib.Path(ARGS[1]) if len(ARGS) > 1 else HERE / 's17_heights.json'
HEIGHTS = json.loads(HP.read_text()) if HP.exists() else {}

sys.argv = ['s14s_build.py', str(ROOT)]
sp = importlib.util.spec_from_file_location('s14s_build', HERE / 's14s_build.py')
Z = importlib.util.module_from_spec(sp); sp.loader.exec_module(Z)   # rewrites S14S (and S2/S7 via s2_build) byte-identical
X, A, s4, s11 = Z.X, Z.A, Z.s4, Z.s11
icon, lamp, num, n, ddate = s4.icon, s4.lamp, X.num, s4.n, s4.ddate
S = s4.S
NB = ' '
NB2 = '\u00a0'  # S2P meta rule: no-break space on both sides of a · separator that must not dangle

# ================================================================= data (extracts only)
TODAY = dt.date(2026, 11, 16)
P0 = TODAY - dt.timedelta(days=27)                       # "4 săpt." = 28 days to today: 20.10–16.11.2026
PER_TXT = f'{P0:%d.%m}–{TODAY:%d.%m.%Y}'; assert PER_TXT == '20.10–16.11.2026'
C0 = dt.date.fromisoformat(s4.c26['start_date']); assert C0 == dt.date(2026, 2, 10)
def sapt(d): return (d - C0).days // 7 + 1
ROLE = {e['user']: e['role'] for e in S['event_log']}
ROLE_RO = {'vet': 'veterinar', 'manager': 'manager', 'owner': 'proprietar'}
AU = S['audit_log']
assert len(AU) == 1 and [e for e in S['event_log'] if e['action'] != 'create'] and \
       all(e['action'] == 'create' for e in S['event_log'] if e['at'] != AU[0]['at']), 'one real change only'
SEXW = {('mortality', 'F'): 'găini moarte', ('mortality', 'M'): 'cocoși morți'}
RECW = {'mortality': 'Mortalitate'}
def change(a):
    rec, sec, sx, day = [x.strip() for x in a['row'].split('·')]
    d = dt.date.fromisoformat(day); at = dt.datetime.fromisoformat(a['at'])
    vals = [(SEXW[(rec, sx)], n(-a['old_values'][k]), n(-a['new_values'][k])) for k in a['new_values']]  # qty < 0 = birds out
    return dict(rec=rec, sec=sec, sx=sx, day=d, at=at, vals=vals, why=a['reason'], user=a['user'],
                role=ROLE_RO[ROLE[a['user']]], kind={'update': 'corectat', 'delete': 'șters'}[a['action']], farm='Hîrbovăț')
CH = [change(a) for a in AU]
assert (CH[0]['sec'], CH[0]['sx'], CH[0]['day'], CH[0]['vals'], CH[0]['user'], CH[0]['role'], f"{CH[0]['at']:%d.%m.%Y, %H:%M}") == \
       ('H4.1', 'F', dt.date(2026, 11, 14), [('găini moarte', '3', '5')], 'Ion Rusu', 'manager', '15.11.2026, 10:12')
assert sapt(CH[0]['day']) == 40
def in_period(c): return P0 <= c['at'].date() <= TODAY
USERS = [('Dr. Ceban', 'veterinar'), ('Ion Rusu', 'manager')]
assert {u for u, _ in USERS} == set(ROLE) and [ROLE_RO[ROLE[u]] for u, _ in USERS] == [r for _, r in USERS]

# record types with records in the period at the user's farms (D45); Telița's cycle closed 28.02.2026, no transfer since
lo, hi = P0.isoformat(), TODAY.isoformat()
inp = lambda rows, key='date': any(lo <= (r[key] or '') [:10] <= hi for r in rows)
TYPES = [('mortality', 'Mortalitate', any(lo <= r['date'] <= hi and r['mortality'] for r in S['daily'])),
         ('cull', 'Sacrificare', inp(S['culls'])),
         ('sexerr', 'Eroare sexare', inp(S['sexing_errors'])),
         ('transfer', 'Transfer', any(lo <= a['date'] <= hi for a in s4.TR['arrivals'])),
         ('feed', 'Hrană zilnică', any(lo <= r['date'] <= hi and r['g_per_bird'] is not None for r in S['daily'])),
         ('delivery', 'Livrare hrană', inp(S['deliveries'])),
         ('price', 'Preț hrană', any(r['priced_at'] and lo <= r['priced_at'] <= hi for r in S['deliveries'])),
         ('weighing', 'Cântărire', inp(S['weighings'])),
         ('eggs', 'Ouă', inp(S['eggs_daily']))]
TYPES = [(k, w) for k, w, on in TYPES if on]
assert [k for k, _ in TYPES] == ['mortality', 'cull', 'feed', 'delivery', 'price', 'weighing', 'eggs'], TYPES

# ================================================================= ChangeRow (SYS09, components.md › Additions; preview markup)
GO = f'<span class="ip-ev__go">{icon("chev", 16).replace(s4.IC["chev"], "<path d=\"m9 18 6-6-6-6\"/>")}</span>'
SEXSR = {'F': ('hen', '♀', 'găini'), 'M': ('rooster', '♂', 'cocoși')}
def chg_row(c, preview=False):
    """The SYS09 ChangeRow. S17P: the ♀/♂ sign in the record link is aria-hidden, so the link's name says "găini" once.
    Its meta groups (ziua date · SAPT · farm) wrap whole and a · never ends or starts a line (.s17-meta).
    preview=True gives the preview's markup (sign exposed, plain-text separators)."""
    cls, sign, word = SEXSR[c['sx']]; d = f"{c['day']:%d.%m.%Y}"
    vals = ''.join(f'<li class="ip-chg__val"><span class="ip-chg__k">{k}</span><del><span class="ip-sr">înainte </span>{o}</del>'
                   f'<span class="ip-chg__arrow" aria-hidden="true">→</span><ins><span class="ip-sr">, după </span>{v}</ins></li>' for k, o, v in c['vals'])
    href = f"#{c['farm'].lower().replace('î', 'i').replace('ă', 'a').replace('ț', 't')}-{c['sec'].lower().replace('.', '-')}-{c['day']:%d-%m-%Y}"
    return (f'<li class="ip-ev ip-ev--change">{s4.tile(c["rec"])}<span class="ip-ev__main">'
            f'<a class="ip-chg__link" href="{href}">{RECW[c["rec"]]} · {c["sec"]} <span class="ip-sex ip-sex--{cls}"{"" if preview else ' aria-hidden="true"'}>{sign}</span><span class="ip-sr"> {word}, ziua {d}</span></a>'
            + (f'<span class="ip-ev__meta">ziua <span class="ip-num">{d} · SAPT {sapt(c["day"])}</span> · {c["farm"]}</span></span>' if preview else
               f'<span class="ip-ev__meta s17-meta"><span><span>ziua <span class="ip-num">{d}</span></span><span class="ip-num">SAPT{NB2}{sapt(c["day"])}</span>'
               f'<span>{c["farm"]}</span></span></span></span>') +
            f'<ul class="ip-chg__vals">{vals}</ul>'
            f'<p class="ip-chg__why"><span class="ip-chg__whyk">Motiv: </span>{c["why"]}</p>'
            f'<span class="ip-chg__meta"><span class="ip-chg__when">{c["at"]:%d.%m.%Y}<span class="ip-chg__dot">, </span><span>{c["at"]:%H:%M}</span></span>'
            f'<span class="ip-chg__dot" aria-hidden="true">·</span><span class="ip-chg__who"><span>{c["user"]}</span><span class="ip-chg__dot" aria-hidden="true">·</span>'
            f'<span class="ip-note">{c["role"]}</span></span></span><span class="ip-tag ip-tag--neutral">{c["kind"]}</span>{GO}</li>')
COLS = ('<li class="ip-chg__cols" aria-hidden="true"><span class="ip-cap">Când</span><span class="ip-cap">Înregistrare</span>'
        '<span class="ip-cap">Valori · vechi → nou</span><span class="ip-cap">Motiv</span><span class="ip-cap">Cine</span><span class="ip-cap">Tip</span></li>')
SKEL = ('<li class="ip-ev ip-ev--skel" aria-hidden="true"><span class="ip-skel ip-skel--tile"></span><span class="ip-ev__main">'
        '<span class="ip-skel ip-skel--w1"></span><span class="ip-skel ip-skel--w2"></span></span><span class="ip-skel ip-skel--w3"></span></li>')
# the preview's markup, so the row is the system's (only the chevron path differs in source: Lucide chevron-right)
_pv = (ROOT / 'design/system/components/ChangeRow/preview.html').read_text(encoding='utf-8')
assert chg_row(CH[0], preview=True).replace('href="#hirbovat-h4-1-14-11-2026"', 'href="#"') in _pv and COLS in _pv and SKEL in _pv

FOOT = ('Istoricul păstrează corecțiile și ștergerile managerilor și ale proprietarului, cu motiv, și modificările autorului '
        f'în aceeași zi, până la {num("23:59")}, fără motiv. Înregistrările noi nu apar aici. Rândul deschide secțiunea în ziua înregistrării.')
def count_txt(k): return 'nicio modificare' if k == 0 else '1 modificare' if k == 1 else f'{k} modificări'

def block(state='default', user=None):
    rows = [c for c in CH if in_period(c) and (user is None or c['user'] == user)]
    who = f' · {user}' if user else ''
    if state == 'loading':
        head_note = f'{num(PER_TXT)}{who} · cele mai noi întâi'
        lst = (f'<p class="ip-sr" role="status">Se încarcă modificările din {PER_TXT}, toate fermele.</p>'
               f'<ul class="ip-list ip-list--change" aria-labelledby="h-chg" aria-busy="true">{SKEL * 4}</ul>')
    elif state == 'error':
        head_note = f'{num(PER_TXT)}{who}'
        lst = (f'<div class="ip-banner ip-banner--fault ip-banner--act" role="alert"><svg class="ip-icon ip-icon--16" viewBox="0 0 24 24" aria-hidden="true">{Z.FAULT}</svg>'
               f'<div class="ip-banner__body"><p class="ip-m0">Istoricul modificărilor din {num(PER_TXT)}, toate fermele, nu s-a încărcat.</p>'
               '<p class="ip-m0">Verifică conexiunea la internet, apoi reîncearcă. Filtrele rămân cum le-ai ales.</p></div>'
               '<button type="button" class="ip-btn ip-btn--secondary ip-banner__key">Reîncearcă</button></div>')
    else:
        head_note = f'{count_txt(len(rows))}{who} · {num(PER_TXT)}' + (' · cele mai noi întâi' if rows else '')
        if rows:
            lst = f'<ul class="ip-list ip-list--change" aria-labelledby="h-chg">{COLS}{"".join(chg_row(c) for c in sorted(rows, key=lambda c: c["at"], reverse=True))}</ul>'
        else:
            by = f' făcută de {user}' if user else ''
            lst = (f'<ul class="ip-list ip-list--change" aria-labelledby="h-chg"><li class="ip-list__empty"><span class="ip-lamp ip-lamp--off">fără modificări</span>'
                   f'<span>Nicio modificare{by} la Hîrbovăț și Telița între <span class="ip-num">{P0:%d.%m.%Y}</span> și <span class="ip-num">{TODAY:%d.%m.%Y}</span>.</span>'
                   '<span>Schimbă perioada sau filtrele.</span></li></ul>')
    foot = '' if state in ('loading', 'error') else f'<p class="ip-note ip-m0">{FOOT}</p>'
    return (f'<section class="ip-block s17-block" aria-labelledby="h-chg"><div class="ip-list__head"><h2 class="ip-t-heading" id="h-chg">Modificări</h2>'
            f'<p class="ip-note">{head_note}</p></div>{lst}{foot}</section>')
assert in_period(CH[0]) and not [c for c in CH if c['user'] == 'Dr. Ceban']

# ================================================================= FilterBar (desktop) + header + sidebars
PER = ['4 săpt.', '13 săpt.', 'Tot istoricul', 'Interval']
def select(i, label, opts, sel):
    o = ''.join(f'<option{" selected=\"selected\"" if t == sel else ""}>{t}</option>' for t in opts)
    return (f'<span class="s11-farm"><label class="ip-label" for="{i}">{label}</label><span class="ip-input s11-select"><select id="{i}">{o}</select>'
            f'{icon("chev", 16)}</span></span>')
USER_OPTS = ['Toți utilizatorii'] + [f'{u} · {r}' for u, r in USERS] + ['Proprietar (tu)']   # FIN08: the owner boards' own entry, as "Manager (tu)" on the manager board
def chips():
    return ('<div class="s4-chips" role="group" aria-label="Tipuri de înregistrări">'
            + ''.join(s4.chip(k, w) for k, w in TYPES) + '</div>')
def filterbar(user=None):
    u = 'Toți utilizatorii' if user is None else f'{user} · {dict(USERS)[user]}'
    return (f'<div class="ip-filterbar s11-filterbar s4-filterbar" role="group" aria-label="Filtre">{s4.seg("Perioadă", PER, 0)}'
            f'{select("f-farm", "Fermă", ["Toate fermele", "Hîrbovăț", "Telița"], "Toate fermele")}{select("f-user", "Utilizator", USER_OPTS, u)}{chips()}</div>')
CRUMBS = ('<nav class="ip-crumbs" aria-label="Cale"><a href="#setari">Setări</a><span aria-hidden="true">›</span>'
          '<span aria-current="page">Istoric modificări</span></nav>')
SCOPE = f'Hîrbovăț,{NB2}Telița'
def meta(scope=SCOPE):  # S2P: the · never ends or starts a line; "noi · Hîrbovăț, Telița" wraps as one unit (S17P)
    return f'<div class="ip-row ip-pagehead__meta"><span class="ip-note">Corecții, ștergeri și modificări din aceeași zi, cu valorile vechi și noi{NB2}·{NB2}{scope}</span></div>'
META = meta()
def header(bare=False):
    if bare: return '<header class="ip-pagehead"><div><h1>Istoric modificări</h1></div></header>'
    return f'<header class="ip-pagehead">\n<div>\n{CRUMBS}\n<h1>Istoric modificări</h1>\n{META}\n</div>\n</header>'
OWNER_SUBS = [('#serii', 'Serii'), ('#ferme', 'Ferme și secțiuni'), ('#ghiduri', 'Ghiduri'), ('#tipuri-hrana', 'Tipuri de hrană'),
              ('#utilizatori', 'Utilizatori'), ('#istoric-modificari', 'Istoric modificări')]  # screens.md nav map, Setări
def sidebar_owner():
    sb = s4.sidebar('Proprietar').replace('class="ip-nav is-current" href="#serie" aria-current="page"', 'class="ip-nav" href="#serie"')
    par = f'<a class="ip-nav" href="#setari">{icon("settings")}Setări</a>'; assert sb.count(par) == 1
    subs = ''.join(f'<a class="ip-nav ip-nav--sub{" is-current" if h == "#istoric-modificari" else ""}" href="{h}"'
                   f'{" aria-current=\"page\"" if h == "#istoric-modificari" else ""}>{t}</a>' for h, t in OWNER_SUBS)
    sb = sb.replace(par, par + subs, 1)
    assert sb.count('is-current') == 1 and 'href="#istoric-modificari" aria-current="page"' in sb
    return sb
def sidebar_viewer():
    sb = s4.sidebar('Observator', viewer=True).replace('class="ip-nav is-current" href="#serie" aria-current="page"', 'class="ip-nav" href="#serie"')
    assert 'is-current' not in sb and '#setari' not in sb
    return sb

def desk(state='default', user=None):
    if state == 'noaccess':
        step = f'<a class="ip-btn ip-btn--secondary" href="#acasa">{icon("home", 16)}Înapoi la Acasă</a>'
        box = s4.empty_box('fără acces', ['Nu ai acces la Istoric modificări.',
                                          'Istoricul îl văd managerii, pentru fermele lor, și proprietarul. Rolul tău, Observator, vede rapoartele și paginile fermelor.'],
                           step, 'Acces')
        return f'<div class="ip-shell">{sidebar_viewer()}<main class="ip-shell__main">\n{header(True)}\n{box}\n</main></div>'
    return f'<div class="ip-shell">{sidebar_owner()}<main class="ip-shell__main">\n{header()}\n{filterbar(user)}\n{block(state, user)}\n</main></div>'

# ================================================================= phone (manager Ion Rusu, 390×844)
def p_header():
    h = CRUMBS.replace('<a href="#setari">Setări</a>', '<a href="#setari"><span class="ip-crumbs__txt">Setări</span></a>').replace('<nav class="ip-crumbs"', '<nav class="ip-crumbs ip-crumbs--touch"', 1)  # FIN06: system touch crumb (section 52)
    return f'<header class="ip-pagehead s2-phead s14-phead"><div>{h}<h1>Istoric modificări</h1></div></header>'
def p_filter():
    pf = Z.p_filter()
    for a, z in (('Filtrează sau exportă', 'Filtrează'), ('Ciclul curent · Hîrbovăț', '4 săpt. · toate fermele · toți utilizatorii · toate tipurile')):
        assert pf.count(a) == 1, a; pf = pf.replace(a, z)
    return pf
def phone_page():
    body = A.phone([p_header(), p_filter(), block()])
    for a, z in (('class="ip-tabbar__item is-current" href="#acasa" aria-current="page"', 'class="ip-tabbar__item" href="#acasa"'),
                 ('class="ip-tabbar__item" href="#mai-mult"', 'class="ip-tabbar__item is-current" href="#mai-mult" aria-current="page"'),
                 ('<div class="ip-screen">', '<div class="ip-screen s14-phone">')):
        assert body.count(a) == 1, a; body = body.replace(a, z)
    assert body.count('is-current') == 1
    return body

# ================================================================= CSS (screen-only; system classes untouched)
CSS17 = '''
.s17-block > .ip-list__head{margin:0} /* the block's gap spaces the head, as .s2-block */
.s17-meta{overflow:hidden} /* S17P: the row meta breaks only between its groups; the · before a group that wraps is clipped */
.s17-meta > span{display:flex;flex-wrap:wrap;margin-left:calc(-1 * var(--space-4))}
.s17-meta > span > span{white-space:nowrap}
.s17-meta > span > span::before{content:"·";display:inline-block;width:var(--space-4);text-align:center}
.s17-block > p.ip-note{max-width:calc(var(--space-12) * 11)} /* S17P: the foot note keeps a reading measure (528px, about 80 characters) */'''
CSS_D = X.CSS + CSS17
CSS_P = Z.CSS_PHONE + CSS17

STEM = 'S17-istoric-modificari'; T = 'S17 · Istoric modificări'
DESK = [(STEM, T, lambda: desk()),
        (STEM + '--filtre-fara-rezultate', f'{T} · filtre fără rezultate', lambda: desk(user='Dr. Ceban')),
        (STEM + '--se-incarca', f'{T} · se încarcă', lambda: desk('loading')),
        (STEM + '--eroare-la-incarcare', f'{T} · eroare la încărcare', lambda: desk('error')),
        (STEM + '--observator-fara-acces', f'{T} · observator · fără acces', lambda: desk('noaccess'))]
PHONE = [(STEM + '--telefon', f'{T} · manager · telefon', phone_page)]
out = []
for stem, title, fn in DESK:
    h = HEIGHTS.get(stem, 1024); s4.write(stem, title, fn(), '', H=h, css=CSS_D); out.append((stem, title, h, 1440))
for stem, title, fn in PHONE:
    A.HEIGHTS[stem] = HEIGHTS.get(stem, 844); out.append(A.write_phone(stem, title, fn(), css=CSS_P))
pathlib.Path('/tmp/s17').mkdir(exist_ok=True)
pathlib.Path('/tmp/s17/boards.json').write_text(json.dumps([{'stem': s, 'w': w, 'h': h} for s, _, h, w in out]))

if CANVAS:  # S17 row, entries keep x/y, a new row goes below the lowest (canvas_rows.py); 80 px gaps
    sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
    c = CR.load(ROOT)
    x = CR.place(c, STEM, [(s, w, h, t) for s, t, h, w in out])
    CR.save(ROOT, c)
    print('canvas: S17 row y', CR.row_y(c, STEM), 'ends at x', x - 80)
