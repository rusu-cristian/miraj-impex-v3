"""S15S · Introdus azi · state boards (spec C8, SYS09, D70, D98; design/notes/S15.md state list), built on s15_build.py.
python3 -I s15s_build.py <repo_root> [--canvas]
Imports s15_build with s15_heights.json (it rewrites the 3 S15 boards and S2/S7 boards byte-identical,
never canvas.json) and reuses its rows (SYS09 act_row through s2_build), group heads, page head, phone viewport (s2s_build
pscreen) and CSS. Toast + Anulează from additions.css §48 (FIN02); Banner, Dialog, skeleton rows as S2S.
Boards (phones 390×844 viewports, D60; the desktop a page at 1440):
- se-incarca            vet · head known (route + user), two skeleton lists, hidden status line
- eroare-la-incarcare   vet · fault Banner naming the day + Reîncearcă (S2S/S12 pattern), no figures
- actiuni-deschis       vet · 16.11 live · Acțiuni open on Mortalitate H4.2 (the step before the S15 Dialog), focus Editează
- dupa-stergere         vet · H4.2 row gone, Mortalitate head ♀ −5 · ♂ −1, Toast "Șters · …" + Anulează (H4.2 ♀ 4 110 again)
- dupa-2359             vet · page of 16.11 still open after 23:59: keys aria-disabled + reason, menu open on H1.1;
                        merged: H1.1 "modificat la 09:52" (D39, the S2S dupa-2359 time on the Mortalitate save)
- manager-10-11         manager Ion Rusu · 10.11.2026 replay (D73): his own Livrare hrană 24 000 kg Breeder 2 (11:52,
                        stock 7 450 → 31 450 kg); merged: several farms (he has Hîrbovăț and Telița, so the farm leads
                        each group head, as MissingRow; Telița has nothing, closed 28.02.2026)
- desktop               vet · 16.11 at 1440: the whole page, inline Editează / Șterge keys (list ≥ 480px, SYS09 C)
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
Z, A, s4, B = Q.Z, Q.A, Q.s4, Q.B
n, num, NB, D, S = A.n, A.num, A.NB, A.D, A.S
hen, roo, pair = A.hen, A.roo, A.pair
HS = json.loads(HP.read_text()) if HP.exists() else {}

# ---------- data checks ----------
DEL, dF = Q.DEL, Q.dF                       # H4.2 ♀ −3 (4 110 → 4 107), ♂ 0
D16 = A.D16; assert (D16['F'], D16['M']) == (8, 1)
AFTER = {'F': D16['F'] - dF['mortality'], 'M': D16['M'] - Q.DAY[(DEL, 'M')]['mortality']}; assert AFTER == {'F': 5, 'M': 1}
EDIT_SEC, EDIT_AT = 'H1.1', '09:52'         # D39: no same-day edit in the sample; the S2S dupa-2359 time, on the first section
assert Q.t(Q.LOG['mortality']) < EDIT_AT < '23:59'
# manager · 10.11.2026 (D73): Ion Rusu's only save that day, the Breeder 2 delivery; stock before/after from the sample
DL = [e for e in S['event_log'] if e['at'].startswith('2026-11-10') and e['user'] == 'Ion Rusu']
assert len(DL) == 1 and DL[0]['record'] == 'feed_delivery' and DL[0]['at'][11:] == '11:52'
DV = [d for d in S['deliveries'] if d['date'] == '2026-11-10']; assert len(DV) == 1
DV = DV[0]; assert (DV['feed_type_id'], DV['qty_kg'], DV['priced'], DV['created_by']) == ('ft-breeder2', 24000, False, 'Ion Rusu')
ST9 = [r for r in S['feed_by_type_daily'] if r['date'] == '2026-11-09' and r['feed_type_id'] == 'ft-breeder2'][0]['stock_end_kg']
K0, K1 = round(ST9), round(ST9 + DV['qty_kg']); assert (K0, K1) == (7450, 31450)
D10 = '10.11.2026'; WK10 = 40  # 10.11.2026 − 10.02.2026 = 273 days → SAPT 40 (D02 formula)
CTX_MGR10 = f'SERIA 2026 · {D10} · HÎRBOVĂȚ SAPT {WK10}'

# ---------- pieces ----------
row, group, mort_rows, egg_rows, feed_rows, wt_rows = Q.row, Q.group, Q.mort_rows, Q.egg_rows, Q.feed_rows, Q.wt_rows
WIN = f'Poți modifica sau șterge până la {num("23:59")}.'
SHUT = f'Timpul de modificare s-a încheiat la {num("23:59")}. Managerul poate corecta.'
def lead(who, rest=WIN): return f'{who}</p><p class="ip-note s2-m0">{rest}'
def counts(k): return f'Dr. Ceban · {num(4)} salvări, {num(k)} înregistrări'

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
    d = AFTER if skip else D16
    return [  # the S15 group heads (s15_build NOTE_*)
        group('g-mo', 'Mortalitate', Q.NOTE_MO(d), mo),
        group('g-ou', 'Ouă', Q.NOTE_OU, egg_rows()),
        group('g-hr', 'Hrană zilnică', Q.NOTE_HR, feed_rows()),
        group('g-ca', 'Cântărire', Q.NOTE_CA, wt_rows()),
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
       f'<p class="s7-m0">Înregistrările tale din {num(D)} nu s-au încărcat. Verifică conexiunea și reîncearcă.</p></div>'
       f'<button type="button" class="ip-btn ip-btn--secondary ip-btn--touch ip-banner__key">Reîncearcă</button></div>')
boards.append(('eroare-la-incarcare', 'vet · eroare la încărcare', 'p', Z.pscreen([
    Q.pagehead(A.CTX_VET, lead('Dr. Ceban')), ban], vet=True)))

# 3 · Acțiuni deschis — 16.11 live, menu open on Mortalitate H4.2 (then Șterge → the S15 Dialog), focus on Editează
fresh()
boards.append(('actiuni-deschis', 'vet · Acțiuni deschis', 'p', Z.pscreen([
    Q.pagehead(A.CTX_VET, lead(counts(22))), *groups({DEL: {'menu': 'open'}})], vet=True)))

# 4 · după ștergere — the H4.2 row is gone, the group head counts without it, Toast with Anulează (graphite, role=status)
fresh()
toast = (f'<div class="ip-toastwrap"><div class="ip-toast" role="status">{A.icon("check", 16)}'
         f'<p class="ip-toast__txt">Șters · Mortalitate {DEL} · în viață ♀ {num(n(dF["alive_start"]))}</p>'
         f'<button type="button" class="ip-link ip-toast__undo">Anulează</button></div></div>')
boards.append(('dupa-stergere', 'vet · după ștergere', 'p', Z.pscreen([
    Q.pagehead(A.CTX_VET, lead(counts(21))), *groups(skip=DEL), toast], vet=True)))

# 5 · după 23:59 — the page of 16.11 still open after midnight: every key aria-disabled with the reason; the menu open on
# H1.1 shows it; H1.1 edited by its author the same day (D39) shows "modificat la 09:52" (merged)
fresh()
with acts(off=True):
    G_OFF = groups({EDIT_SEC: {'menu': 'open', 'edited': EDIT_AT}})
boards.append(('dupa-2359', 'vet · după 23:59 · modificat la · exemplu de stare (D39)', 'p', Z.pscreen([
    Q.pagehead(A.CTX_VET, lead(counts(22), SHUT)), *G_OFF], vet=True)))

# 6 · manager · 10.11.2026 (D73 replay): his own entries only (the vet's saves of that day are not his); two farms, so the
# farm leads the group head (merged "mai multe ferme"); Editează → S13 in edit mode, Șterge → Dialog
fresh()
title = (f'<span class="ip-sr">Livrare hrană · </span>Breeder 2 · {num(n(DV["qty_kg"]) + NB + "kg")}')
meta = f'{A.nw("fără preț")} · {A.nw("stoc " + num(n(K0)) + " → " + num(n(K1) + NB + "kg"))}'
DROW = row('delivery', title, meta, 'Livrare hrană Breeder 2', '#livrare-hrana')
G_MGR = group('g-lv', 'Hîrbovăț · Livrare hrană', f'salvat la {num(DL[0]["at"][11:])} · {num(n(DV["qty_kg"]) + NB + "kg")}', [DROW])
boards.append(('manager-10-11', 'manager · 10.11.2026 · livrare · mai multe ferme', 'p', Z.pscreen([
    Q.pagehead(CTX_MGR10, lead(f'Ion Rusu · {num(1)} salvare, {num(1)} înregistrare')), G_MGR], vet=False)))

# 7 · desktop · vet 16.11 at 1440 — the whole page; list ≥ 480px → inline keys (SYS09 C). Vet Sidebar: Acasă ·
# Înregistrează (current); no series selector, Serie, Rapoarte or Setări (spec §1: the vet sees entry screens and Introdus azi)
fresh()
def vet_sidebar():
    import re
    sb = s4.sidebar('Medic veterinar')
    sb, k = re.subn(r'<div class="ip-series">.*?</ul></div>\n', '', sb, count=1, flags=re.S); assert k == 1
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
# S15P: breadcrumb above the h1, as the S4 / S11 / S14 desktop heads (parent › this page); S15 opens from Acasă's
# Introdus azi block ("Vezi pe secțiuni", S2), so the parent is Acasă
CRUMBS = ('<nav class="ip-crumbs" aria-label="Cale"><a href="#acasa">Acasă</a><span aria-hidden="true">›</span>'
          '<span aria-current="page">Introdus azi</span></nav>')
HEAD_D = Q.pagehead(A.CTX_VET, Q.LEAD).replace(' s2-phead', '', 1)
assert HEAD_D.count('<div><h1>') == 1
HEAD_D = HEAD_D.replace('<div><h1>', f'<div>{CRUMBS}<h1>', 1)
DESK = f'<div class="s4-shell">{vet_sidebar()}<main class="s4-main s15s-desk">\n{"".join([HEAD_D] + groups())}\n</main></div>'
boards.append(('desktop', 'vet · desktop', 'd', DESK))

# phones: Înregistrează is the current tab on S15 (s15_build.nav_here)
boards = [(s, t, k, Q.nav_here(b) if k == 'p' else b) for s, t, k, b in boards]

# ---------- CSS ----------
CSS_ADD = '''
.s15s-busy{gap:var(--space-6)} /* the two skeleton lists sit as the groups do (s2-main gap) */'''
CSS_P = Q.CSS_VP + CSS_ADD
CSS_D = A.CSS_D + Q.CSS_ADD + SC.REC + '''
.s4-main > .ip-pagehead h1{margin:0}
.s15s-desk > *{max-width:calc(var(--space-12) * 20)} /* one reading column: figures and their keys stay within reach (960px) */'''

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
pathlib.Path('/tmp/s15/states.json').write_text(json.dumps(out, ensure_ascii=False))

if CANVAS:  # S15 row: the 3 S15 boards (x 0/470/940), these phones, then the desktop
    sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
    c = CR.load(ROOT)
    x = CR.place(c, STEM0, [b['stem'] for b in Q.out] + [(b['stem'], b['w'], b['h'], b['title']) for b in out], 0)
    CR.save(ROOT, c)
    print('canvas: S15 row y', CR.row_y(c, STEM0), 'ends at x', x - 80)
