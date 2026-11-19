"""S15 · Introdus azi (spec C8, screens.md S15): the user's entries of today, grouped by record type with their icons,
each with Editează / Șterge until 23:59 (D07, SYS09 EventRow actions). Phone, today 18.11.2026 (D152), Hîrbovăț SAPT 59.
Reached from Acasă › Introdus azi › "Vezi pe secțiuni" (S2, D86), so it agrees with S2's block: the same 3 saves of
Dr. Ceban (illustrative event log on the real figures, RF08: 08:15 Zilnic, 08:27 Ouă, 08:44 Hrană zilnică; no weighing at
Hîrbovăț today, S8) split into one row per record (D82): Mortalitate per compartment of today's Zilnic (real deaths, D155),
Ouă per hall with eggs (C1; H7 empty since 14.11), Hrană zilnică one row per form (D60/D77, SYS09: 8 din 8 rânduri, as S7/S2).
Every figure from realdata.py (DATA03 extracts); authors/times illustrative; names only Dr. Ceban / Ion Rusu (D96).
Boards (phones, 390):
- S15-introdus-azi            vet · natural height (the whole page)
- S15-introdus-azi--gol       manager Ion Rusu before 13:16 (his first save of 18.11 is the 13:16 sale; S2 manager block, D87/D88)
- S15-introdus-azi--sterge-confirmare  vet · Dialog over the page, Șterge Mortalitate of the compartment with the day's
                              largest loss (no reason, D70 A; focus Anulează)
Reuses s2s_build (imports s2_build, which rewrites S2/S7 boards byte-identical): act_row (SYS09), head, pagehead,
tabbar, phone shell, pscreen (390×844 viewport), CSS_PS (S2 phone + sstate_css VIEW/BANNER/MODAL + EmptyState box).
python3 -I s15_build.py <repo_root> [s15_heights.json] [--canvas]
Writes design/canvas/project/S15-introdus-azi[--<state>].dc.html, /tmp/s15/<stem>.html, /tmp/s15/boards.json, and, with
--canvas only, the S15 entries of canvas.json (canvas_rows.py: entries keep x/y; a new row goes below the lowest)."""
import sys, json, pathlib, importlib, datetime as dt

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
ROOT, HP, CANVAS = CR.args('s15_heights.json'); ROOT = ROOT.resolve()  # heights default: s15_heights.json
HERE = pathlib.Path(__file__).resolve().parent
HS = json.loads(HP.read_text()) if HP and HP.exists() else {}
sys.path.insert(0, str(HERE))
sys.argv = ['s2s_build.py', str(ROOT)]
Z = importlib.import_module('s2s_build')       # writes S2 (+S7) boards, unchanged
A, s4, B = Z.A, Z.s4, Z.B
R = A.R                                         # realdata (DATA03 extracts, RF08)
n, num, NB, D, TODAY = A.n, A.num, A.NB, A.D, A.TODAY
hen, roo, pair, F, M = A.hen, A.roo, A.pair, A.F, A.M
NB2 = A.NB2
T = TODAY.isoformat()

# ---------- data (asserted against S2's block) ----------
LOG = A.LOG; t = A.t_of
assert sorted(LOG) == ['eggs', 'feed', 'mortality'] and A.INTRO == ['mortality', 'eggs', 'feed']                     # 3 saves; no weighing at Hîrbovăț today (S8)
SECS = list(A.ZSEC)                                                                 # compartments in today's Zilnic (H7 empty since 14.11)
DAY = {(s, sx): R.sd(TODAY, s, sx) for s in SECS for sx in 'FM'}
assert all(DAY.values()) and not [c for c in R.ILL['culls'] + R.ILL['sexing_errors'] if c['date'] == T]   # no culls / sexing errors today
DTH = A.DTH; assert {sx: sum(DAY[(s, sx)]['mort'] for s in SECS) for sx in 'FM'} == DTH
EGG_HALLS = list(A.EGG_HALLS); assert sum(R.eggs(TODAY, h) for h in EGG_HALLS) == A.EGG_T
EGGS = {h: dict(total=R.eggs(TODAY, h), inc=R.day(TODAY, h)['inc'], lay=R.lay_day(TODAY, h)) for h in EGG_HALLS}
LAYG = R.guide_lay(A.WK)
FT = {R.FEED_TYPE[sx]: A.FEED_KGS[sx] for sx in 'FM'}; assert round(sum(FT.values()), 1) == A.FEED_KG
GS = {(s, sx): R.feed_g(TODAY, sx, s) for s, sx in A.FEED_ROWS}                    # norm g/bird typed per row (S7)
GG = {sx: R.guide_feed(A.WK, sx) for sx in 'FM'}
DEVS = [(s, sx, g, GG[sx], (g / GG[sx] - 1) * 100) for (s, sx), g in GS.items() if abs(g / GG[sx] - 1) > 0.10]   # ±10 % (C2, S7)
assert all(LOG[k]['user'] == 'Dr. Ceban' for k in LOG)
assert [e['record'] for e in A.MGR_LOG] == ['sale']                                # manager: the 13:16 sale (one form, D168)
DEL = max(SECS, key=lambda s: (DAY[(s, 'F')]['mort'], -SECS.index(s)))             # the Dialog's row: the day's largest ♀ loss
dF = dict(mortality=DAY[(DEL, 'F')]['mort'], alive_start=R.alive_start(DEL, 'F'), alive_end=DAY[(DEL, 'F')]['alive'])
assert dF['alive_start'] - dF['mortality'] - DAY[(DEL, 'F')]['sold'] == dF['alive_end'] and DAY[(DEL, 'M')]['mort'] == 0 and dF['mortality'] > 0
MINUS = A.MINUS
SIGN = {'F': '♀', 'M': '♂'}; WORD = {'F': 'găini', 'M': 'cocoși'}

def mloss(v): return n(-v) if v else '0'
def sexv(sx, v, unit=''): return (hen if sx == 'F' else roo)(v, unit)
def pct(v): return f'{n(v, 1)}%'

# ---------- rows (SYS09 EventRow actions; the type word is in the row text for screen readers, the tile is aria-hidden) ----------
ROWS = []  # (record key, index, title, meta, accessible name, Editează href)
I = [0]
def row(k, title, meta, name, href, **kw):
    I[0] += 1
    h = A.act_row(k, I[0], title, meta, name, **kw)
    old = f'href="#editeaza-{k}"'; assert h.count(old) == 2
    return h.replace(old, f'href="{href}"')

def mort_rows(**kw):
    out = []
    for s in SECS:
        f, m = DAY[(s, 'F')], DAY[(s, 'M')]
        title = f'{s}<span class="ip-sr"> · Mortalitate</span> · {hen(mloss(f["mort"]))} · {roo(mloss(m["mort"]))}'
        meta = f'în viață {pair(n(f["alive"]), n(m["alive"]))}'
        out.append(row('mortality', title, meta, f'Mortalitate {s}', f'#zilnic-{s}', **kw.get(s, {})))
    return out
def egg_rows():
    out = []
    for h, r in EGGS.items():
        title = f'{h}<span class="ip-sr"> · Ouă</span> · {num(n(r["total"]))} ouă'
        diff = r['lay'] - LAYG
        dev = '' if abs(diff) <= 3 else (f' · {A.nw("ghid " + num(pct(LAYG)))} · '   # D19 ±3 points, S6 wording
                                         f'{A.nw(num(("+" if diff > 0 else MINUS) + n(abs(diff), 1)) + " puncte " + A.lamp("warn", "abatere"))}')
        # S15P: incubabile first, then ouat · ghid · deviation (the comparison reads in one run, the lamp ends the meta)
        meta = f'{A.nw("incubabile " + num(n(r["inc"])))} · {A.nw("ouat " + num(pct(r["lay"])))}{dev}'
        out.append(row('eggs', title, meta, f'Ouă {h}', f'#zilnic-oua-{h}'))
    return out
def grams(sx):
    gs = sorted({R.rnd(g) for (s, x), g in GS.items() if x == sx})
    return f'{F if sx == "F" else M} ' + num(n(gs[0]) + NB + 'g' if len(gs) == 1 else f'{n(gs[0])}–{n(gs[-1])}' + NB + 'g')
def feed_rows():
    title = (f'<span class="ip-sr">Hrană zilnică · </span>{num(len(A.FEED_ROWS))} din {num(len(A.DUE))} rânduri · '
             f'{num(n(A.FEED_KG, 1) + NB + "kg")}')
    lack = [f'{s} {F if sx == "F" else M}' for s, sx in A.DUE if (s, sx) not in A.FEED_ROWS]
    parts = [A.nw('fără ' + ', '.join(lack))] if lack else []
    if DEVS:   # the rows off the guide (S7 abatere): section, g, guide, deviation + lamp
        for s, sx, g, gg, dv in DEVS:
            parts += [A.nw(f'{s} {F if sx == "F" else M} ' + num(n(g) + NB + 'g')), A.nw('ghid ' + num(n(gg) + NB + 'g')),
                      A.nw(num(A.s4.sgn(dv, 0) + '%') + ' ' + A.lamp('warn', 'abatere'))]
    else:      # every row within ±10 % of the guide: the grams typed (S7 g/pasăre)
        parts += [A.nw(grams('F') + ' · ' + grams('M'))]
    return [row('feed', title, ' · '.join(parts), 'Hrană zilnică de azi', '#hrana-zilnica')]

# S15P: group-head notes, shared with s15s_build (each head carries its group's totals; Hrană zilnică: kg per feed type)
NOTE_MO = lambda d: f'Zilnic · salvat la {num(t(LOG["mortality"]))} · {pair(n(-d["F"]), n(-d["M"]))}'
NOTE_OU = f'Zilnic · salvat la {num(t(LOG["eggs"]))} · {num(n(A.EGG_T))} ouă · ouat {num(s4.pct(A.LAY))}'
NOTE_HR = (f'salvat la {num(t(LOG["feed"]))} · {A.nw(A.FTN[R.FEED_TYPE["F"]] + ": " + num(n(FT[R.FEED_TYPE["F"]], 1) + NB + "kg"))} · '   # S7 wording: name + colon
           f'{A.nw(A.FTN[R.FEED_TYPE["M"]] + ": " + num(n(FT[R.FEED_TYPE["M"]], 1) + NB + "kg"))}')
def group(gid, cap, note, rows):
    return (f'<section class="s2-block" aria-labelledby="{gid}">{A.head(cap, note, gid)}'
            f'<ul class="ip-list ip-list--act" aria-labelledby="{gid}">{"".join(rows)}</ul></section>')
def groups():
    return [
        group('g-mo', 'Mortalitate', NOTE_MO(DTH), mort_rows()),
        group('g-ou', 'Ouă', NOTE_OU, egg_rows()),
        group('g-hr', 'Hrană zilnică', NOTE_HR, feed_rows()),
    ]

# FX07/D142: Înregistrează is the parent (Sidebar item / tab bar key #inregistreaza, as the entry forms' back links); phones
# take the system touch crumb (additions.css section 52), as the S14/S17–S22 phones; s15s_build uses CRUMBS on the desktop
CRUMBS = ('<nav class="ip-crumbs" aria-label="Cale"><a href="#inregistreaza">Înregistrează</a><span aria-hidden="true">›</span>'
          '<span aria-current="page">Introdus azi</span></nav>')
CRUMBS_P = CRUMBS.replace('<a href="#inregistreaza">Înregistrează</a>', '<a href="#inregistreaza"><span class="ip-crumbs__txt">Înregistrează</span></a>', 1) \
    .replace('<nav class="ip-crumbs"', '<nav class="ip-crumbs ip-crumbs--touch"', 1)
assert CRUMBS_P.count('ip-crumbs__txt') == 1 and 'ip-crumbs--touch' in CRUMBS_P
def pagehead(ctx, note, crumbs=CRUMBS_P):
    return (f'<header class="ip-pagehead s2-phead"><div>{crumbs}<h1>Introdus azi</h1><p class="ip-note ip-num ip-m0">{ctx}</p>'
            f'<p class="ip-note ip-m0 s15-lead">{note}</p></div></header>')
N_REC = len(SECS) + len(EGG_HALLS) + 1   # one row per record (D82)
LEAD = f'Dr. Ceban · {num(len(LOG))} salvări, {num(N_REC)} înregistrări</p><p class="ip-note ip-m0">Poți modifica sau șterge până la {num("23:59")}.'
HEAD_V = pagehead(A.CTX_VET, LEAD)

# ---------- boards ----------
# 1 · vet · default: the whole page (natural height), tab bar at its foot (as S2 phones)
I[0] = 0
VET = A.phone([HEAD_V] + groups(), vet=True)

# 2 · gol · manager before 13:16 (Ion Rusu's first save of 18.11 is the 13:16 sale, D168; his 10:12 correction is not an
#   entry of his, it shows in Istoric modificări): EmptyState box (as S2S fără acces) with a next step
I[0] = 0
EMPTY_1 = 'Nu ai salvat nimic azi.'
EMPTY_2 = f'Ce salvezi azi apare aici, pe tipuri; poți modifica sau șterge până la {num("23:59")}.'
step = '<a class="ip-btn ip-btn--secondary ip-btn--touch" href="#inregistreaza">Înregistrează</a>'
assert A.INTRO_MGR == f'{EMPTY_1} {EMPTY_2}'  # S15P: S2's manager Introdus azi empty block says the same (D98)
box = s4.empty_box('nimic introdus', [EMPTY_1, EMPTY_2], step, 'Introdus azi')
HEAD_M = pagehead(A.CTX_MGR, 'Ion Rusu')
GOL = Z.pscreen([HEAD_M, box], vet=False)

# 3 · Șterge · confirmare — Dialog over the default page, Mortalitate of the day's largest loss (08:15 Zilnic); the author's same-day
# delete has no reason (D70 A); effect on birds alive; stays in Istoric modificări; focus on Anulează (stacked keys, S2P)
I[0] = 0
dlg = (f'<div class="ip-scrim ip-scrim--phone"><div class="ip-dialog" role="alertdialog" aria-modal="true" aria-labelledby="dlg-t" aria-describedby="dlg-n">'
       f'<h2 id="dlg-t">Șterge Mortalitate {DEL} din <span class="ip-num">{D}</span></h2>'
       f'<p class="ip-dialog__note" id="dlg-n">Se șterge rândul salvat la {num(t(LOG["mortality"]))}: {pair(n(-dF["mortality"]), "0")}. '
       f'În viață în {DEL}: {hen(n(dF["alive_end"]))} → {s4.hen_v(n(dF["alive_start"]))}. Ștergerea rămâne în Istoric modificări.</p>'
       f'<div class="ip-dialog__keys ip-dialog__keys--stack"><button type="button" class="ip-btn ip-btn--secondary ip-btn--touch ip-focus" autofocus="autofocus">Anulează</button>'
       f'<button type="button" class="ip-btn ip-btn--destructive ip-btn--touch">Șterge</button></div></div></div>')
DLG = Z.pscreen([HEAD_V] + groups(), vet=True, overlay=dlg, modal=True)

def nav_here(html):
    cur = '<a class="ip-tabbar__item is-current" href="#acasa" aria-current="page">'
    key = '<a class="ip-tabbar__item" href="#inregistreaza">'
    assert html.count(cur) == 1 and html.count(key) == 1
    return html.replace(cur, '<a class="ip-tabbar__item" href="#acasa">').replace(key, '<a class="ip-tabbar__item is-current" href="#inregistreaza" aria-current="page">')
VET, GOL, DLG = nav_here(VET), nav_here(GOL), nav_here(DLG)

# ---------- CSS ----------
CSS_ADD = '''
.s15-lead{margin-top:var(--space-1)}'''
CSS_FULL = A.CSS_P + CSS_ADD
CSS_VP = Z.CSS_PS + CSS_ADD

# ---------- write ----------
STEM0, TITLE0 = 'S15-introdus-azi', 'S15 · Introdus azi'
SYS = ROOT / 'design/system'
pathlib.Path('/tmp/s15').mkdir(exist_ok=True)
def write(stem, title, body, css, h, fixed):
    style = f'padding:0; width:390px; height:{h}px; overflow:hidden' if fixed else f'padding:0; width:390px; min-height:{h}px'
    root = f'<div class="ip-root" style="{style}">{body}</div>'
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
{B.FONTS}
<style>{css}</style>
</helmet>
{root}
</x-dc>
<script type="text/x-dc" data-dc-script data-props='{{"$preview":{{"width":390,"height":{h}}}}}'>
class Component extends DCLogic {{ renderVals() {{ return {{}}; }} }}
</script>
</body>
</html>
'''
    (ROOT / f'design/canvas/project/{stem}.dc.html').write_text(dc, encoding='utf-8')
    wrap = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8"><title>{title}</title>
<link rel="stylesheet" href="file://{SYS}/tokens.css"><link rel="stylesheet" href="file://{SYS}/components/bundle.css"><link rel="stylesheet" href="file://{SYS}/components/additions.css">
{B.FONTS}<style>{css}</style></head><body>{root}</body></html>'''
    pathlib.Path(f'/tmp/s15/{stem}.html').write_text(wrap if fixed else wrap.replace(f'min-height:{h}px', 'min-height:844px', 1), encoding='utf-8')  # RF08: the render wrapper keeps only the 844 floor (natural height)
    print('ok', stem, h)
    return {'stem': stem, 'title': title, 'h': h, 'w': 390, 'fixed': fixed}

out = [write(STEM0, f'{TITLE0} · vet', VET, CSS_FULL, HS.get(STEM0, 844), False),
       write(f'{STEM0}--gol', f'{TITLE0} · gol · manager · înainte de {t(A.MGR_LOG[0])}', GOL, CSS_VP, 844, True),
       write(f'{STEM0}--sterge-confirmare', f'{TITLE0} · vet · Șterge · confirmare', DLG, CSS_VP, 844, True)]
pathlib.Path('/tmp/s15/boards.json').write_text(json.dumps(out, ensure_ascii=False))

if CANVAS:  # S15 row, entries keep x/y, a new row goes below the lowest (canvas_rows.py); 80 px gaps
    c = CR.load(ROOT)
    x = CR.place(c, STEM0, [(b['stem'], b['w'], b['h'], b['title']) for b in out])
    CR.save(ROOT, c)
    print('canvas: S15 row y', CR.row_y(c, STEM0), 'ends at x', x - 80)
