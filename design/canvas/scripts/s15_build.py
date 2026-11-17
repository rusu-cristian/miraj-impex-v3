"""S15 · Introdus azi (spec C8, screens.md S15): the user's entries of today, grouped by record type with their icons,
each with Editează / Șterge until 23:59 (D07, SYS09 EventRow actions). Phone, 16.11.2026 (D23), Hîrbovăț SAPT 40.
Reached from Acasă › Introdus azi › "Vezi pe secțiuni" (S2, D86), so it agrees with S2's block: the same 4 saves of
Dr. Ceban (event log 08:34 Zilnic, 08:35 Ouă, 09:47 Hrană zilnică, 10:10 Cântărire) split into one row per record
(D82): Mortalitate per section (S6 round, ♀ −8 · ♂ −1), Ouă per hall (C1, 18 688), Hrană zilnică one row per form
(D60/D77, SYS09: 10 din 12 rânduri · 3 691,3 kg, as S7/S2), Cântărire per section × sex (C7: one record each, S8's
H1.1 ♀ 3 967 g). Every figure from sample-hirbovat-2026.json (illustrative, D27); names only Dr. Ceban / Ion Rusu (D96).
Boards (phones, 390):
- S15-introdus-azi            vet · natural height (the whole page)
- S15-introdus-azi--gol       manager Ion Rusu, nothing saved on 16.11 (true in the sample; S2 manager block, D87/D88)
- S15-introdus-azi--sterge-confirmare  vet · Dialog over the page, Șterge Mortalitate H4.2 (no reason, D70 A; focus Anulează)
Reuses s2s_build (imports s2_build, which rewrites S2/S7 boards byte-identical): act_row (SYS09), head, pagehead,
tabbar, phone shell, pscreen (390×844 viewport), CSS_PS (S2 phone + sstate_css VIEW/BANNER/MODAL + EmptyState box).
python3 -I s15_build.py <repo_root> [s15_heights.json]
Writes design/canvas/project/S15-introdus-azi[--<state>].dc.html, /tmp/s15/<stem>.html, /tmp/s15/boards.json, and the
S15 entries of canvas.json (new row below the lowest row: y = row y + tallest h + 120; purely additive)."""
import sys, json, pathlib, importlib, datetime as dt

ROOT = pathlib.Path(sys.argv[1]).resolve()
HP = pathlib.Path(sys.argv[2]) if len(sys.argv) > 2 else None
HERE = pathlib.Path(__file__).resolve().parent
HS = json.loads(HP.read_text()) if HP and HP.exists() else {}
sys.path.insert(0, str(HERE))
sys.argv = ['s2s_build.py', str(ROOT)]
Z = importlib.import_module('s2s_build')       # writes S2 (+S7) boards, unchanged
A, s4, B = Z.A, Z.s4, Z.B
n, num, NB, D, TODAY, S = A.n, A.num, A.NB, A.D, A.TODAY, A.S
hen, roo, pair, F, M = A.hen, A.roo, A.pair, A.F, A.M
NB2 = A.NB2
T = TODAY.isoformat()

# ---------- data (asserted against S2's block) ----------
LOG = A.LOG; t = A.t_of
SECS = ['H1.1', 'H1.2', 'H3.1', 'H3.2', 'H4.1', 'H4.2']
DAY = {(r['section'], r['sex']): r for r in S['daily'] if r['date'] == T}
assert sorted({s for s, _ in DAY}) == SECS
assert all(DAY[k]['culls'] == 0 and DAY[k]['sexing_errors'] == 0 for k in DAY)   # no culls / sexing errors on 16.11
assert {sx: sum(DAY[(s, sx)]['mortality'] for s in SECS) for sx in 'FM'} == A.D16 == {'F': 8, 'M': 1}
EGGS = {r['hall']: r for r in A.EGGS16}; assert list(EGGS) == ['H1', 'H3', 'H4'] and sum(r['total_eggs'] for r in EGGS.values()) == A.EGG_T
WT = [w for w in S['weighings'] if w['date'] == T]; assert len(WT) == 12 and all(w['created_by'] == 'Dr. Ceban' for w in WT)
WT.sort(key=lambda w: (SECS.index(w['section']), w['sex']))
assert (WT[0]['section'], WT[0]['sex'], WT[0]['avg_weight_g']) == ('H1.1', 'F', 3967)          # S8 default board
FT = {}
for a in A.FEED_ROWS: FT[a['feed_type_id']] = FT.get(a['feed_type_id'], 0) + round(a['today_g'] * a['alive'] / 1000, 1)
assert round(sum(FT.values()), 1) == A.FEED_KG == 3691.3
assert all(LOG[k]['user'] == 'Dr. Ceban' for k in LOG) and not A.MGR_LOG                       # manager: nothing on 16.11 (D88)
DEL = 'H4.2'  # the Dialog's row: the largest loss of the day (♀ −3), real figures
dF = DAY[(DEL, 'F')]; assert (dF['mortality'], dF['alive_start'], dF['alive_end'], DAY[(DEL, 'M')]['mortality']) == (3, 4110, 4107, 0)
DEV = S['summary_today']['feed_deviation_today']   # the one feed row off the guide today (S7 abatere board)
assert (DEV['section'], DEV['sex'], DEV['g'], DEV['guide_g'], DEV['deviation_pct']) == ('H1.2', 'M', 172, 155, 11.0)
assert [h for h, r in EGGS.items() if abs(r['lay_pct'] - r['guide_lay_pct']) > 3] == ['H3']                 # S6: H3 70,1% vs 74,1%
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
        title = f'{s}<span class="ip-sr"> · Mortalitate</span> · {hen(mloss(f["mortality"]))} · {roo(mloss(m["mortality"]))}'
        meta = f'în viață {pair(n(f["alive_end"]), n(m["alive_end"]))}'
        out.append(row('mortality', title, meta, f'Mortalitate {s}', f'#zilnic-{s}', **kw.get(s, {})))
    return out
def egg_rows():
    out = []
    for h, r in EGGS.items():
        title = f'{h}<span class="ip-sr"> · Ouă</span> · {num(n(r["total_eggs"]))} ouă'
        diff = r['lay_pct'] - r['guide_lay_pct']
        dev = '' if abs(diff) <= 3 else (f' · {A.nw("ghid " + num(pct(r["guide_lay_pct"])))} · '   # D19 ±3 points, S6 wording
                                         f'{A.nw(num(("+" if diff > 0 else MINUS) + n(abs(diff), 1)) + " puncte " + A.lamp("warn", "abatere"))}')
        # S15P: incubabile first, then ouat · ghid · deviation (the comparison reads in one run, the lamp ends the meta)
        meta = f'{A.nw("incubabile " + num(n(r["hatching_eggs"])))} · {A.nw("ouat " + num(pct(r["lay_pct"])))}{dev}'
        out.append(row('eggs', title, meta, f'Ouă {h}', f'#zilnic-oua-{h}'))
    return out
def feed_rows():
    title = f'<span class="ip-sr">Hrană zilnică · </span>{num(len(A.FEED_ROWS))} din {num(12)} rânduri · {num(n(A.FEED_KG, 1) + NB + "kg")}'
    dv = DEV
    meta = (f'{A.nw("fără H3.2 " + M + ", H4.2 " + F)} · '
            f'{A.nw(dv["section"] + " " + M + " " + num(str(dv["g"]) + NB + "g"))} · {A.nw("ghid " + num(str(dv["guide_g"]) + NB + "g"))} · '
            f'{A.nw(num("+" + n(dv["deviation_pct"]) + "%") + " " + A.lamp("warn", "abatere"))}')
    return [row('feed', title, meta, 'Hrană zilnică de azi', '#hrana-zilnica')]
def wt_rows():
    out = []
    for w in WT:
        s, sx = w['section'], w['sex']
        title = f'{s}<span class="ip-sr"> · Cântărire</span> · {sexv(sx, n(w["avg_weight_g"]), NB + "g")}'
        meta = (f'{A.nw("CV " + num(pct(w["cv_pct"])))} · {A.nw("unif. " + num(pct(w["uniformity_pct"])))} · '
                f'{A.nw("probă " + num(w["sample_size"]))}')
        out.append(row('weighing', title, meta, f'Cântărire {s} {WORD[sx]}', f'#cantarire-{s}-{sx}'))
    return out

# S15P: group-head notes, shared with s15s_build (each head carries its group's totals; Hrană zilnică: kg per feed type)
NOTE_MO = lambda d: f'Zilnic · salvat la {num(t(LOG["mortality"]))} · {pair(n(-d["F"]), n(-d["M"]))}'
NOTE_OU = f'Zilnic · salvat la {num(t(LOG["eggs"]))} · {num(n(A.EGG_T))} ouă · ouat {num(s4.pct(A.LAY))}'
NOTE_HR = (f'salvat la {num(t(LOG["feed"]))} · {A.nw("Breeder 2: " + num(n(FT["ft-breeder2"], 1) + NB + "kg"))} · '   # S7 wording: name + colon
           f'{A.nw("Cocoși: " + num(n(FT["ft-cocosi"], 1) + NB + "kg"))}')
NOTE_CA = f'salvat la {num(t(LOG["weighing"]))} · {num(6)} secțiuni · {num(12)} rânduri'
def group(gid, cap, note, rows):
    return (f'<section class="s2-block" aria-labelledby="{gid}">{A.head(cap, note, gid)}'
            f'<ul class="ip-list ip-list--act" aria-labelledby="{gid}">{"".join(rows)}</ul></section>')
def groups():
    return [
        group('g-mo', 'Mortalitate', NOTE_MO(A.D16), mort_rows()),
        group('g-ou', 'Ouă', NOTE_OU, egg_rows()),
        group('g-hr', 'Hrană zilnică', NOTE_HR, feed_rows()),
        group('g-ca', 'Cântărire', NOTE_CA, wt_rows()),
    ]

def pagehead(ctx, note):
    return (f'<header class="ip-pagehead s2-phead"><div><h1>Introdus azi</h1><p class="ip-note ip-num s2-m0">{ctx}</p>'
            f'<p class="ip-note s2-m0 s15-lead">{note}</p></div></header>')
N_REC = 6 + 3 + 1 + 12
LEAD = f'Dr. Ceban · {num(4)} salvări, {num(N_REC)} înregistrări</p><p class="ip-note s2-m0">Poți modifica sau șterge până la {num("23:59")}.'
HEAD_V = pagehead(A.CTX_VET, LEAD)

# ---------- boards ----------
# 1 · vet · default: the whole page (natural height), tab bar at its foot (as S2 phones)
I[0] = 0
VET = A.phone([HEAD_V] + groups(), vet=True)

# 2 · gol · manager (Ion Rusu has no record on 16.11 in the sample): EmptyState box (as S2S fără acces) with a next step
I[0] = 0
EMPTY_1 = 'Nu ai salvat nimic azi.'
EMPTY_2 = f'Ce salvezi azi apare aici, pe tipuri; poți modifica sau șterge până la {num("23:59")}.'
step = '<a class="ip-btn ip-btn--secondary ip-btn--touch" href="#inregistreaza">Înregistrează</a>'
assert A.INTRO_MGR == f'{EMPTY_1} {EMPTY_2}'  # S15P: S2's manager Introdus azi empty block says the same (D98)
box = s4.empty_box('nimic introdus', [EMPTY_1, EMPTY_2], step, 'Introdus azi')
HEAD_M = pagehead(A.CTX_MGR, 'Ion Rusu')
GOL = Z.pscreen([HEAD_M, box], vet=False)

# 3 · Șterge · confirmare — Dialog over the default page, Mortalitate H4.2 (♀ −3 · ♂ 0, 08:34); the author's same-day
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
    pathlib.Path(f'/tmp/s15/{stem}.html').write_text(wrap, encoding='utf-8')
    print('ok', stem, h)
    return {'stem': stem, 'title': title, 'h': h, 'w': 390, 'fixed': fixed}

out = [write(STEM0, f'{TITLE0} · vet', VET, CSS_FULL, HS.get(STEM0, 844), False),
       write(f'{STEM0}--gol', f'{TITLE0} · gol · manager', GOL, CSS_VP, 844, True),
       write(f'{STEM0}--sterge-confirmare', f'{TITLE0} · vet · Șterge · confirmare', DLG, CSS_VP, 844, True)]
pathlib.Path('/tmp/s15/boards.json').write_text(json.dumps(out, ensure_ascii=False))

# canvas.json: new row below the lowest row (S8: y 26059 + tallest 1113 + 120), 80 px gaps, purely additive
cp = ROOT / 'design/canvas/project/canvas.json'
c = json.loads(cp.read_text(encoding='utf-8'))
others = {k: v for k, v in c['boards'].items() if not k.startswith('S15-')}
ly = max(v['y'] for v in others.values()); Y = ly + max(v['h'] for v in others.values() if v['y'] == ly) + 120
assert Y == 26059 + 1113 + 120 == 27292, Y
x = 0
for b in out:
    k = f'{b["stem"]}.dc.html'
    c['boards'][k] = {'x': x, 'y': Y, 'w': 390, 'h': b['h'], 'title': b['title']}; x += 390 + 80
    if k not in c['order']: c['order'].append(k)
cp.write_text(json.dumps(c, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
print('canvas S15 row y', Y)
