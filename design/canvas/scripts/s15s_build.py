"""S15S · Introdus azi · state boards (spec C8, SYS09, D70, D98; design/notes/S15.md state list), built on s15_build.py.
python3 -I s15s_build.py <repo_root> [--canvas]
Imports s15_build with s15_heights.json (it rewrites the 3 S15 boards and S2/S7 boards byte-identical,
never canvas.json) and reuses its rows (SYS09 act_row through s2_build), group heads, page head, phone viewport (s2s_build
pscreen) and CSS. Toast + Anulează from additions.css §48 (FIN02); Banner, Dialog, skeleton rows as S2S.
Today 18.11.2026 (RF08, realdata): Dr. Ceban's 3 saves (illustrative times on the real figures), Ion Rusu's 13:16 sale.
Boards (phones 390×844 viewports, D60; the desktop a page at 1440):
- se-incarca            vet · head known (route + user), two skeleton lists, hidden status line
- eroare-la-incarcare   vet · fault Banner naming the day + Reîncearcă (S2S/S12 pattern), no figures
- actiuni-deschis       vet · live · Acțiuni open on Mortalitate of the day's largest loss (the step before the S15 Dialog), focus Editează
- dupa-stergere         vet · that row gone, Mortalitate head without it, Toast "Mortalitate ștearsă · …" + Anulează
- dupa-2359             vet · page of 18.11 still open after 23:59: keys aria-disabled + reason, menu open on the first
                        compartment; merged: its row "modificat la 09:52" (D39: the event log has no same-day author edit;
                        the S2S dupa-2359 time on the Mortalitate save)
- manager               manager Ion Rusu · his own entries of 18.11: the 13:16 Vânzare H6.1 + H6.2 (one form = one record,
                        D168; real figures); merged: several farms (he has Hîrbovăț and Telița, so the farm leads each group
                        head, as MissingRow; Telița has nothing, closed 28.02.2026)
- desktop               vet · 18.11 at 1440: the whole page, inline Editează / Șterge keys (list ≥ 480px, SYS09 C)
Writes design/canvas/project/S15-introdus-azi--<state>.dc.html, /tmp/s15/<stem>.html (phones), /tmp/s4s/<stem>.html
(desktop), /tmp/s15/states.json; --canvas adds these entries to the S15 row after the 3 S15 boards: phones from x 1410
step 470, then the desktop (canvas_rows.py: on a rerun every entry keeps x/y)."""
import sys, json, pathlib, importlib

ARGS = sys.argv[1:]
CANVAS = '--canvas' in ARGS; ARGS = [a for a in ARGS if a != '--canvas']
ROOT = pathlib.Path(ARGS[0]).resolve()
HERE = pathlib.Path(__file__).resolve().parent
HP = HERE / 's15_heights.json'
sys.path.insert(0, str(HERE))
sys.argv = ['s15_build.py', str(ROOT), str(HP)]
Q = importlib.import_module('s15_build')       # writes the S15 boards (and S2/S7, unchanged)
import sstate_css as SC
import sexsign as SXS  # SX01: ♀/♂ outside a figure span = the §72 hook (sexsign.py)
Z, A, s4, B = Q.Z, Q.A, Q.s4, Q.B
n, num, NB, D, R = A.n, A.num, A.NB, A.D, Q.R
hen, roo, pair = A.hen, A.roo, A.pair
HS = json.loads(HP.read_text()) if HP.exists() else {}

# ---------- data checks ----------
DEL, dF = Q.DEL, Q.dF                       # the day's largest ♀ loss (s15_build)
AFTER = {'F': Q.DTH['F'] - dF['mortality'], 'M': Q.DTH['M'] - Q.DAY[(DEL, 'M')]['mort']}
assert AFTER['F'] >= 0 and AFTER['M'] >= 0
EDIT_SEC, EDIT_AT = Q.SECS[0], '09:52'      # D39: no same-day author edit in the event log; the S2S dupa-2359 time, on the first compartment
assert Q.t(Q.LOG['mortality']) < EDIT_AT < '23:59'
# manager · 18.11.2026: Ion Rusu's own entries that day = the 13:16 sale form H6.1 + H6.2 (D168, real figures)
SL = A.MGR_LOG; assert len(SL) == 1 and SL[0]['record'] == 'sale'
SL = SL[0]; assert (SL['F'], SL['M']) == (R.sold(A.TODAY, sx='F'), R.sold(A.TODAY, sx='M'))
N_ALL = len(Q.LOG); N_ROWS = Q.N_REC

# ---------- pieces ----------
row, group, mort_rows, egg_rows, feed_rows = Q.row, Q.group, Q.mort_rows, Q.egg_rows, Q.feed_rows
WIN = f'Poți modifica sau șterge până la {num("23:59")}.'
SHUT = f'Timpul de modificare s-a încheiat la {num("23:59")}. Managerul poate corecta.'
def lead(who, rest=WIN): return f'{who}</p><p class="ip-note ip-m0">{rest}'
def counts(k): return f'Dr. Ceban · {num(N_ALL)} salvări, {num(k)} înregistrări'

class acts:
    """Every row of the board drawn with these act_row defaults (off = the 23:59 window closed), s2_build untouched after."""
    def __init__(self, **kw): self.kw = kw
    def __enter__(self):
        self.orig, self.row = A.act_row, Q.row
        A.act_row = lambda *a, **k: self.orig(*a, **{**self.kw, **k})
        if self.kw.get('off'):  # disabled Editează has no href (role=link, tabindex=0), so s15_build.row's href swap is skipped
            def row_off(k, title, meta, name, href, **kw):
                Q.I[0] += 1
                h = A.act_row(k, Q.I[0], title, meta, name, **kw); assert 'href="#editeaza-' not in h
                return h
            Q.row = row_off
    def __exit__(self, *e): A.act_row, Q.row = self.orig, self.row

def groups(mort_kw=None, skip=None):
    """s15_build.groups() with per-section row options and one section left out (after Șterge)."""
    secs = Q.SECS
    if skip: Q.SECS = [s for s in secs if s != skip]
    try:
        mo = mort_rows(**(mort_kw or {}))
    finally:
        Q.SECS = secs
    d = AFTER if skip else Q.DTH
    return [  # the S15 group heads (s15_build NOTE_*)
        group('g-mo', 'Mortalitate', Q.NOTE_MO(d), mo),
        group('g-ou', 'Ouă', Q.NOTE_OU, egg_rows()),
        group('g-hr', 'Hrană zilnică', Q.NOTE_HR, feed_rows()),
    ]
def fresh(): Q.I[0] = 0
fresh(); _g = groups(); fresh(); assert _g == Q.groups()  # unchanged options → the S15 default's rows exactly
boards = []  # (slug, state title, kind 'p' | 'd', body)

# 1 · se încarcă — head from the route and the signed-in user; what was saved is unknown: two skeleton lists (no captions)
fresh()
def skgroup(k):
    return (f'<section class="s2-block" aria-hidden="true"><div class="ip-list__head"><span class="ip-skel ip-skel--w3"></span>'
            f'<span class="ip-skel ip-skel--w2"></span></div>{Z.sklist(k)}</section>')
LOADING = f'<div class="s2-block s15s-busy" aria-busy="true">{skgroup(4)}{skgroup(2)}</div>'
boards.append(('se-incarca', 'vet · se încarcă', 'p', Z.pscreen([
    Q.pagehead(A.CTX_VET, lead('Dr. Ceban')), Z.STATUS('Introdus azi'), LOADING], vet=True)))

# 2 · eroare la încărcare — online only (spec §1): fault Banner with the day and the fix + Reîncearcă; no figures
ban = (f'<div class="ip-banner ip-banner--fault ip-banner--act" role="alert">{A.icon("fault", 16)}<div class="ip-banner__body">'
       f'<p class="ip-m0">Înregistrările tale din {num(D)} nu s-au încărcat.</p><p class="ip-m0">Verifică conexiunea la internet, apoi apasă Reîncearcă.</p></div>'
       f'<button type="button" class="ip-btn ip-btn--secondary ip-btn--touch ip-banner__key">Reîncearcă</button></div>')
boards.append(('eroare-la-incarcare', 'vet · eroare la încărcare', 'p', Z.pscreen([
    Q.pagehead(A.CTX_VET, lead('Dr. Ceban')), ban], vet=True)))

# 3 · Acțiuni deschis — 18.11 live, menu open on the Dialog's Mortalitate row (then Șterge → the S15 Dialog), focus on Editează
fresh()
boards.append(('actiuni-deschis', 'vet · Acțiuni deschis', 'p', Z.pscreen([
    Q.pagehead(A.CTX_VET, lead(counts(N_ROWS))), *groups({DEL: {'menu': 'open'}})], vet=True)))

# 4 · după ștergere — the deleted row is gone, the group head counts without it, Toast with Anulează (graphite, role=status)
fresh()
toast = SXS.sexify(f'<div class="ip-toastwrap"><div class="ip-toast" role="status">{A.icon("check", 16)}'
         f'<p class="ip-toast__txt">Mortalitate ștearsă · {DEL} · în viață ♀ {num(n(dF["alive_start"]))}</p>'
         f'<button type="button" class="ip-link ip-toast__undo">Anulează</button></div></div>')
boards.append(('dupa-stergere', 'vet · după ștergere', 'p', Z.pscreen([
    Q.pagehead(A.CTX_VET, lead(counts(N_ROWS - 1))), *groups(skip=DEL), toast], vet=True)))

# 5 · după 23:59 — the page of 18.11 still open after midnight: every key aria-disabled with the reason; the menu open on
# the first compartment shows it; that row edited by its author the same day (D39) shows "modificat la 09:52" (merged)
fresh()
with acts(off=True):
    G_OFF = groups({EDIT_SEC: {'menu': 'open', 'edited': EDIT_AT}})
boards.append(('dupa-2359', 'vet · după 23:59 · modificat la · exemplu de stare (D39)', 'p', Z.pscreen([
    Q.pagehead(A.CTX_VET, lead(counts(N_ROWS), SHUT)), *G_OFF], vet=True)))

# 6 · manager · 18.11.2026: his own entries only (the vet's saves are not his): the 13:16 sale form (one record, D168); two
# farms, so the farm leads the group head (merged "mai multe ferme"); Editează → S10 in edit mode, Șterge → Dialog
fresh()
codes = R.sale_codes(SL)
title = f'<span class="ip-sr">Vânzare · </span>{codes} · {s4.sexpair(SL["F"], SL["M"])}'
meta = f'în viață {pair(n(A.ALIVE["F"]), n(A.ALIVE["M"]))}'
SROW = row('sale', title, meta, f'Vânzare {codes}', '#vanzare')
G_MGR = group('g-vz', 'Hîrbovăț · Vânzare', f'salvat la {num(Q.t(SL))} · {pair(n(-SL["F"]), n(-SL["M"]))}', [SROW])
boards.append(('manager', 'manager · vânzare', 'p', Z.pscreen([
    Q.pagehead(A.CTX_MGR, lead(f'Ion Rusu · {num(1)} salvare, {num(1)} înregistrare')), G_MGR], vet=False)))

# 7 · desktop · vet 18.11 at 1440 — the whole page; list ≥ 480px → inline keys (SYS09 C). Vet Sidebar: Acasă ·
# Înregistrează (current); no series selector, Serie, Rapoarte or Setări (spec §1: the vet sees entry screens and Introdus azi)
fresh()
def vet_sidebar():
    import re
    sb = s4.sidebar('Medic veterinar')
    sb, k = re.subn(r'<div class="ip-select ip-series">.*?</ul></div>\n', '', sb, count=1, flags=re.S); assert k == 1
    for h in ('#serie', '#rapoarte', '#setari'):
        sb, k = re.subn(r'<a class="ip-nav[^"]*" href="' + h + r'"[^>]*>.*?</a>', '', sb, count=1); assert k == 1
    # Înregistrează expands (components.md › Sidebar, screens.md nav map): the vet's 3 entry forms, each with its record icon in
    # its rec-* colour (Sidebar preview), then Introdus azi, the current page (is-current + aria-current on the sub-item).
    # Introdus azi has no record type: an empty 14px icon slot keeps its label in the sub-items' label column.
    sub = lambda h, ic, lbl, cur=False: (f'<a class="ip-nav ip-nav--sub{" is-current" if cur else ""}" href="{h}"{" aria-current=\"page\"" if cur else ""}>'
                                         f'{ic}{lbl}</a>')
    rec = lambda k: f'<span class="ip-rec ip-rec--{A.REC[k]} s13-rec">{A.icon(k, 14)}</span>'
    subs = (sub('#zilnic', rec('mortality'), 'Zilnic') + sub('#hrana-zilnica', rec('feed'), 'Hrană zilnică')
            + sub('#cantarire', rec('weighing'), 'Cântărire')
            + sub('#introdus-azi', '<svg class="ip-icon ip-icon--14" viewBox="0 0 24 24" aria-hidden="true"></svg>', 'Introdus azi', True))
    par = f'<a class="ip-nav" href="#inregistreaza">{A.icon("plus")}Înregistrează</a>'
    assert sb.count(par) == 1
    sb = sb.replace(par, par + subs, 1)
    assert sb.count('is-current') == 1 and sb.count('class="ip-nav') == 6 and 'href="#introdus-azi" aria-current="page"' in sb
    return sb
# S15P: breadcrumb above the h1, as the S4 / S11 / S14 desktop heads (parent › this page); FX07/D142: the parent is
# Înregistrează (the Sidebar group this page sits in), not Acasă, though S2's Introdus azi block also links here
HEAD_D = Q.pagehead(A.CTX_VET, Q.LEAD, Q.CRUMBS).replace(' s2-phead', '', 1)
assert HEAD_D.count(f'<div>{Q.CRUMBS}<h1>') == 1
DESK = f'<div class="ip-shell">{vet_sidebar()}<main class="ip-shell__main s15s-desk">\n{HEAD_D}{A.desk_heads("".join(groups()))}\n</main></div>'  # FX02: desktop section heads (s2_build.desk_heads)
boards.append(('desktop', 'vet · desktop', 'd', DESK))

# phones: Înregistrează is the current tab on S15 (s15_build.nav_here)
boards = [(s, t, k, Q.nav_here(b) if k == 'p' else b) for s, t, k, b in boards]

# ---------- CSS ----------
CSS_ADD = '''
.s15s-busy{gap:var(--space-6)} /* the two skeleton lists sit as the groups do (s2-main gap) */'''
CSS_P = Q.CSS_VP + CSS_ADD
CSS_D = A.CSS_D + Q.CSS_ADD + SC.REC  # FW01 (D148): the groups span the page (the 960px reading column is gone)

# ---------- write ----------
STEM0, TITLE0 = Q.STEM0, Q.TITLE0
out = []
for slug, state, kind, body in boards:
    stem, ttl = f'{STEM0}--{slug}', f'{TITLE0} · {state}'
    if kind == 'p':
        b = Q.write(stem, ttl, body, CSS_P, 844, True)
        out.append({**b})
    else:
        h = HS.get(stem, 1024)
        s4.write(stem, ttl, body, H=h, css=CSS_D)
        out.append({'stem': stem, 'title': ttl, 'h': h, 'w': 1440, 'fixed': False})
RENAMED = [(f'{STEM0}--manager-10-11.dc.html', f'{STEM0}--manager.dc.html')]   # RF08: was the 10.11 delivery replay (D73)
for k0, _ in RENAMED: (ROOT / 'design/canvas/project' / k0).unlink(missing_ok=True)
pathlib.Path('/tmp/s15/states.json').write_text(json.dumps(out, ensure_ascii=False))

if CANVAS:  # S15 row: the 3 S15 boards (x 0/470/940), these phones, then the desktop
    sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
    c = CR.load(ROOT)
    for k0, k1 in RENAMED:  # RF08: the stem now says what the board shows (x/y, title slot and order position kept)
        if k0 in c['boards']:
            c['boards'] = {(k1 if k == k0 else k): v for k, v in c['boards'].items()}; c['order'] = [k1 if k == k0 else k for k in c['order']]
    x = CR.place(c, STEM0, [b['stem'] for b in Q.out] + [(b['stem'], b['w'], b['h'], b['title']) for b in out], 0)
    CR.save(ROOT, c)
    print('canvas: S15 row y', CR.row_y(c, STEM0), 'ends at x', x - 80)
