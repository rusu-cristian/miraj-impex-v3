#!/usr/bin/env python3
"""SYS06 generator: GuideChart variants + base GuideChart preview from the extracts.
Usage: python3 -I sys06_gen.py <repo_root>"""
import json, sys, collections as C, datetime as D, re, os

ROOT = sys.argv[1]
EX = os.path.join(ROOT, 'design/data/extracts')
COMP = os.path.join(ROOT, 'design/system/components')
NB = ' '; MINUS = '−'
def J(n): return json.load(open(os.path.join(EX, n)))

def fint(n):
    s = f'{abs(round(n)):,}'.replace(',', NB)
    return (MINUS if round(n) < 0 else '') + s
def fdec(x, k=1, sign=False):
    s = f'{abs(x):,.{k}f}'.replace(',', 'X').replace('.', ',').replace('X', NB)
    if x < 0 and float(f'{abs(x):.{k}f}') != 0: s = MINUS + s
    elif sign and x > 0: s = '+' + s
    return s
def pct(x, k=1, sign=False): return fdec(x, k, sign) + '%'

# ---------------------------------------------------------------- chart engine
# Two drawings per chart (GuideChart phone pattern, fix round 1): the wide one (viewBox 600×150, 10px axis text, shown
# while the chart box is ≥600px, i.e. ≥1:1) and the phone one (viewBox 320×150, fewer ticks, 11px axis text), switched
# by a container query on div.ip-chartbox (additions.css § 16). Same ARIA on both; only the shown one is focusable/read.
WIDE = dict(W=600, X0=34, X1=592, YB=130, XL=144, LX=30, mod='wide')
PHONE = dict(W=320, X0=40, X1=308, YB=128, XL=144, LX=36, mod='phone')
def draw(G, yt, n_weeks, series, yticks, yfmt, ymin, ymax, guide, band, today, cur, xticks, gend_note):
    X0, X1, YB = G['X0'], G['X1'], G['YB']
    def X(w): return X0 + (w - 1) * (X1 - X0) / (n_weeks - 1)
    def Y(v): return YB - (v - ymin) / (ymax - ymin) * (YB - yt)
    g = []
    for i, t in enumerate(yticks):
        y = Y(t)
        g.append(f'<line class="{"axis" if i == 0 else "grid"}" x1="{X0}" x2="{X1}" y1="{y:.1f}" y2="{y:.1f}"/>')
        g.append(f'<text class="ax" x="{G["LX"]}" y="{y + 3:.1f}" text-anchor="end">{yfmt(t)}</text>')
    if band:
        lo, hi = band
        ws = sorted(lo)
        pts = [f'{X(w):.1f},{Y(hi[w]):.1f}' for w in ws] + [f'{X(w):.1f},{Y(lo[w]):.1f}' for w in reversed(ws)]
        g.append(f'<polygon class="band" points="{" ".join(pts)}"/>')
    if guide:
        g.append('<polyline class="guide" points="' + ' '.join(f'{X(w):.1f},{Y(v):.1f}' for w, v in sorted(guide.items())) + '"/>')
        if gend_note:  # partial guide: end cap + axis-style note at the guide's last week
            lw = max(guide); y1 = Y(band[1][lw]) - 4 if band else Y(guide[lw]) - 9; y2 = Y(band[0][lw]) + 4 if band else Y(guide[lw]) + 9
            g.append(f'<line class="gend" x1="{X(lw):.1f}" x2="{X(lw):.1f}" y1="{y1:.1f}" y2="{y2:.1f}"/>')
            g.append(f'<text class="ax" x="{X(lw) - 6:.1f}" y="{y1 - 4:.1f}" text-anchor="end">{gend_note}</text>')
    for s in series:
        g.append(f'<polyline class="{s["cls"]}" points="' + ' '.join(f'{X(w):.1f},{Y(v):.1f}' for w, v in sorted(s['pts'].items())) + '"/>')
    if today:
        g.append(f'<line class="today" x1="{X(today):.1f}" x2="{X(today):.1f}" y1="8" y2="{YB}"/>')
    if cur:
        g.append(f'<line class="cur" x1="{X(cur):.1f}" x2="{X(cur):.1f}" y1="8" y2="{YB}"/>')
        for s in series:
            if cur in s['pts']:
                g.append(f'<circle class="{s.get("pt", "pt")}" cx="{X(cur):.1f}" cy="{Y(s["pts"][cur]):.1f}" r="3.5"/>')
    for w in (xticks or []):
        g.append(f'<text class="ax" x="{X(w):.1f}" y="{G["XL"]}" text-anchor="middle">{w}</text>')
    return ''.join(g)

def chart(cid, label, n_weeks, series, yticks, yfmt, ymin, ymax, guide=None, band=None, today=None, cur=None,
          help_txt=None, read_txt='', focus=False, table_cols=None, table_rows=None, caption=None, xticks=None,
          pyticks=None, pxticks=None, yt=20, gend_note=None):
    out = '<div class="ip-chartbox">'
    for G, yk, xk in ((WIDE, yticks, xticks), (PHONE, pyticks or yticks, pxticks or xticks)):
        out += (f'<svg class="ip-chart ip-chart--{G["mod"]}{" ip-focus" if focus else ""}" viewBox="0 0 {G["W"]} 150" role="group" '
                f'aria-roledescription="grafic" tabindex="0" aria-label="{label}" aria-describedby="{cid}-help {cid}-read"><g aria-hidden="true">'
                + draw(G, yt, n_weeks, series, yk, yfmt, ymin, ymax, guide, band, today, cur, xk, gend_note) + '</g></svg>')
    out += '</div>'
    help_txt = help_txt or 'Săgețile stânga și dreapta mută săptămâna; Home și End: prima și ultima săptămână.'
    out += f'<span class="ip-sr" id="{cid}-help">{help_txt}</span>'
    out += f'<div class="ip-chart__read" id="{cid}-read" role="status" aria-live="polite">{read_txt}</div>'
    out += f'<table class="ip-sr"><caption>{caption or label}</caption><thead><tr>' + ''.join(f'<th scope="col">{c}</th>' for c in table_cols) + '</tr></thead><tbody>'
    for r in table_rows:
        out += f'<tr><th scope="row">{r[0]}</th>' + ''.join(f'<td>{c}</td>' for c in r[1:]) + '</tr>'
    out += '</tbody></table>'
    return out

SW = {  # legend swatches (24×8)
    'act': '<svg viewBox="0 0 24 8" aria-hidden="true"><line class="act" x1="0" x2="24" y1="4" y2="4"/></svg>',
    's2': '<svg viewBox="0 0 24 8" aria-hidden="true"><line class="s2" x1="0" x2="24" y1="4" y2="4"/></svg>',
    'guide': '<svg viewBox="0 0 24 8" aria-hidden="true"><line class="guide" x1="0" x2="24" y1="4" y2="4"/></svg>',
    'band': '<svg viewBox="0 0 24 8" aria-hidden="true"><rect class="band" x="0" y="0" width="24" height="8"/></svg>',
}
def legend(items, note=''):
    s = '<div class="ip-chart__legend">' + ''.join(f'<span>{SW[k]}{t}</span>' for k, t in items)
    if note: s += f'<span>{note}</span>'
    return s + '</div>'
def readout(cap, lamp_cls, lamp_word, body, sub_html, extra=''):
    return (f'<div class="ip-readout"><div class="ip-readout__cap"><span>{cap}</span><span class="ip-rlamp ip-rlamp--{lamp_cls}">{lamp_word}</span></div>'
            f'{extra}{body}{sub_html}</div>')
HEN = '<b>♀</b>'; ROO = '<b>♂</b>'
DASH = '–'

S = J('sample-hirbovat-2026.json')
START = D.date(2026, 2, 10)
def cw(ds): return (D.date.fromisoformat(ds) - START).days // 7 + 1
NW = 40; XT = [1, 10, 20, 30, 40]; PXT = [1, 20, 40]
HIR_NOTE = 'Hîrbovăț · Seria 2026, date exemplu (D27); SAPT 1–2 fără date înregistrate la Hîrbovăț.'

# 1. mortality cumulative %
startF, startM = 26765, 3060
mw = C.defaultdict(lambda: [0, 0])
for r in S['daily']:
    mw[r['cycle_week']][0 if r['sex'] == 'F' else 1] += r['mortality']
cf = cm = 0; mF = {}; mM = {}
for w in sorted(mw):
    cf += mw[w][0]; cm += mw[w][1]
    mF[w] = cf / startF * 100; mM[w] = cm / startM * 100
rows = [[str(w), DASH + ' fără date', DASH + ' fără date', DASH + ' fără date în ghid'] for w in (1, 2)]
rows += [[str(w), pct(mF[w], 2), pct(mM[w], 2), DASH + ' fără date în ghid'] for w in sorted(mF)]
v1 = readout('MORTALITATE CUMULATĂ ♀ · ♂ · HÎRBOVĂȚ · SAPT 40', 'off', 'FĂRĂ DATE ÎN GHID',
    chart('gv1', 'Mortalitate cumulată găini și cocoși, Hîrbovăț Seria 2026, săptămânile 1–40; ghidul nu are mortalitate', NW,
          [{'cls': 'act', 'pts': mF}, {'cls': 's2', 'pt': 'pt2', 'pts': mM}], [0, 5, 10, 15], lambda t: f'{t}%', 0, 15,
          today=40, cur=40, xticks=XT, pxticks=PXT, focus=True,
          read_txt=f'SAPT 40 · ♀ {pct(mF[40], 2)} · ♂ {pct(mM[40], 2)} · ghid {DASH} fără date în ghid',
          table_cols=['SAPT', 'Găini ♀', 'Cocoși ♂', 'Ghid'], table_rows=rows),
    legend([('act', f'{HEN} găini'), ('s2', f'{ROO} cocoși')], f'% din populare (♀ 26{NB}765 · ♂ 3{NB}060) · ghid: {DASH} fără date în ghid'))

# 2. feed kg/day by feed type
ft = C.defaultdict(lambda: C.defaultdict(list))
for r in S['feed_by_type_daily']:
    ft[r['feed_type_id']][cw(r['date'])].append(r['consumption_kg'])
kF = {w: sum(v) / len(v) for w, v in ft['ft-breeder2'].items()}
kM = {w: sum(v) / len(v) for w, v in ft['ft-cocosi'].items()}
# like-with-like guide kg for week 40 (rows with grams only, D46)
act40 = C.Counter(); gd40 = C.Counter()
for r in S['daily']:
    if r['cycle_week'] == 40 and r['g_per_bird'] is not None:
        act40[r['sex']] += r['g_per_bird'] * r['alive_end'] / 1000
        gd40[r['sex']] += r['guide_feed_g'] * r['alive_end'] / 1000
devF = (act40['F'] / gd40['F'] - 1) * 100; devM = (act40['M'] / gd40['M'] - 1) * 100
ok = abs(devF) <= 10 and abs(devM) <= 10
rows = [[str(w), DASH + ' fără date', DASH + ' fără date'] for w in (1, 2)]
rows += [[str(w), fint(kF[w]) + ' kg', fint(kM[w]) + ' kg'] for w in sorted(kF)]
v2 = readout('HRANĂ KG/ZI · BREEDER 2 ♀ · COCOȘI ♂ · HÎRBOVĂȚ · SAPT 40', 'ok' if ok else 'warn', 'ÎN NORMĂ' if ok else 'ABATERE',
    chart('gv2', 'Consum de hrană pe zi, pe tip de hrană, Hîrbovăț Seria 2026, săptămânile 1–40: Breeder 2 pentru găini, Cocoși pentru cocoși', NW,
          [{'cls': 'act', 'pts': kF}, {'cls': 's2', 'pt': 'pt2', 'pts': kM}], [0, 2000, 4000, 6000], lambda t: fint(t), 0, 6000, pyticks=[0, 3000, 6000],
          today=40, cur=40, xticks=XT, pxticks=PXT,
          read_txt=f'SAPT 40 · Breeder 2 ♀ {fint(kF[40])} kg/zi · Cocoși ♂ {fint(kM[40])} kg/zi · față de ghid ♀ {pct(devF, 1, True)} · ♂ {pct(devM, 1, True)}',
          table_cols=['SAPT', 'Breeder 2 ♀, kg/zi', 'Cocoși ♂, kg/zi'], table_rows=rows),
    legend([('act', f'{HEN} Breeder 2'), ('s2', f'{ROO} Cocoși')], 'medie kg/zi a săptămânii · ghidul e în g/pasăre: abatere = kg față de ghid × păsări, aceleași rânduri (±10%)'))

# 3. CV / uniformity ♀
cv = C.defaultdict(lambda: [0, 0, 0])
for r in S['weighings']:
    if r['sex'] == 'F':
        a = cv[r['cycle_week']]; n = r['sample_size']
        a[0] += r['cv_pct'] * n; a[1] += r['uniformity_pct'] * n; a[2] += n
cvp = {w: a[0] / a[2] for w, a in cv.items()}; un = {w: a[1] / a[2] for w, a in cv.items()}
rows = [[str(w), DASH + ' fără cântărire', DASH + ' fără cântărire', DASH + ' fără date în ghid'] for w in (1, 2)]
rows += [[str(w), pct(un[w]), pct(cvp[w]), DASH + ' fără date în ghid'] for w in sorted(cv)]
v3 = readout('UNIFORMITATE · CV · ♀ · HÎRBOVĂȚ · SAPT 40', 'off', 'FĂRĂ DATE ÎN GHID',
    chart('gv3', 'Uniformitate și coeficient de variație la găini, Hîrbovăț Seria 2026, săptămânile 1–40; ghidul nu are uniformitate sau CV', NW,
          [{'cls': 'act', 'pts': un}, {'cls': 's2', 'pt': 'pt2', 'pts': cvp}], [0, 25, 50, 75, 100], lambda t: f'{t}%', 0, 100, pyticks=[0, 50, 100],
          today=40, cur=40, xticks=XT, pxticks=PXT,
          read_txt=f'SAPT 40 · uniformitate {pct(un[40])} · CV {pct(cvp[40])} · ghid {DASH} fără date în ghid',
          table_cols=['SAPT', 'Uniformitate ♀', 'CV ♀', 'Ghid'], table_rows=rows),
    legend([('act', 'uniformitate · păsări în ±10% din medie'), ('s2', 'CV')], f'♀, medie pe secțiuni după probă · ghid: {DASH} fără date în ghid'))

# 4. hatching % and 5. egg weight (hall = all)
EW = [x for x in S['eggs_weekly'] if x['hall'] == 'all']
hp = {x['cycle_week']: x['hatching_pct'] for x in EW if x['hatching_pct'] is not None}
eg = {x['cycle_week']: x['avg_egg_weight_g'] for x in EW if x['avg_egg_weight_g'] is not None}
inc = {x['cycle_week']: x for x in EW}
rows = [[str(w), DASH + ' fără date', DASH + ' fără date', DASH + ' fără date în ghid'] for w in (1, 2)]
rows += [[str(3), DASH + ' fără ouă', DASH + ' fără ouă', DASH + ' fără date în ghid']]
rows += [[str(w), pct(hp[w]), f'{fint(inc[w]["hatching_eggs"])} din {fint(inc[w]["total_eggs"])}', DASH + ' fără date în ghid'] for w in sorted(hp)]
v4 = readout('INCUBABILE · % DIN TOTAL OUĂ · HÎRBOVĂȚ · SAPT 40', 'off', 'FĂRĂ DATE ÎN GHID',
    chart('gv4', 'Ouă incubabile din totalul ouălor, Hîrbovăț Seria 2026, săptămânile 1–40; ghidul nu are incubabile', NW,
          [{'cls': 'act', 'pts': hp}], [0, 25, 50, 75, 100], lambda t: f'{t}%', 0, 100, pyticks=[0, 50, 100],
          today=40, cur=40, xticks=XT, pxticks=PXT,
          read_txt=f'SAPT 40 · incubabile {pct(hp[40])} · {fint(inc[40]["hatching_eggs"])} din {fint(inc[40]["total_eggs"])} ouă · ghid {DASH} fără date în ghid',
          table_cols=['SAPT', 'Incubabile, % din total ouă', 'Ouă incubabile din total ouă', 'Ghid'], table_rows=rows),
    legend([('act', 'incubabile ÷ total ouă (Cat. 1–3 + incubabile)')], f'SAPT 3 fără ouă · ghid: {DASH} fără date în ghid'))
rows = [[str(w), DASH + ' fără date', DASH + ' fără date în ghid'] for w in (1, 2)] + [['3', DASH + ' fără ouă', DASH + ' fără date în ghid']]
rows += [[str(w), fdec(eg[w]) + ' g', DASH + ' fără date în ghid'] for w in sorted(eg)]
v5 = readout('MASA MEDIE A OULUI · HÎRBOVĂȚ · SAPT 40', 'off', 'FĂRĂ DATE ÎN GHID',
    chart('gv5', 'Masa medie a oului, Hîrbovăț Seria 2026, săptămânile 1–40; ghidul nu are masa oului', NW,
          [{'cls': 'act', 'pts': eg}], [40, 50, 60, 70], lambda t: f'{t} g', 40, 70, pyticks=[40, 55, 70],
          today=40, cur=40, xticks=XT, pxticks=PXT,
          read_txt=f'SAPT 40 · {fdec(eg[40])} g · ghid {DASH} fără date în ghid',
          table_cols=['SAPT', 'Masa medie', 'Ghid'], table_rows=rows),
    legend([('act', 'masa medie, g (medie după numărul de ouă)')], f'SAPT 3 fără ouă · ghid: {DASH} fără date în ghid'))

# 6. small curve (sparkline) 160×48: lay %, feed g ♀ / ♂ vs guide
GL = J('guide-laying.json')['rows']
glay = {r['cycle_week']: r['lay_pct'] for r in GL if r['sex'] == 'F' and r['lay_pct'] is not None and r['cycle_week'] <= 40}
gfeed = {s: {r['cycle_week']: r['feed_g'] for r in GL if r['sex'] == s and r['cycle_week'] <= 40} for s in 'FM'}
lay = {x['cycle_week']: x['lay_pct'] for x in EW}
fg = {s: C.defaultdict(lambda: [0, 0]) for s in 'FM'}
for r in S['daily']:
    if r['g_per_bird'] is not None:
        a = fg[r['sex']][r['cycle_week']]; a[0] += r['g_per_bird'] * r['alive_end']; a[1] += r['alive_end']
fgw = {s: {w: a[0] / a[1] for w, a in fg[s].items()} for s in 'FM'}
def spark(act, guide, mod, label):
    allv = list(act.values()) + list(guide.values())
    lo, hi = min(allv), max(allv)
    def X(w): return 4 + (w - 1) * 152 / 39
    def Y(v): return 44 - (v - lo) / (hi - lo) * 40
    gp = ' '.join(f'{X(w):.1f},{Y(v):.1f}' for w, v in sorted(guide.items()))
    ap = ' '.join(f'{X(w):.1f},{Y(v):.1f}' for w, v in sorted(act.items()))
    lw = max(act)
    return (f'<svg class="ip-spark{mod}" viewBox="0 0 160 48" role="img" aria-label="{label}"><polyline class="act" points="{ap}"/><polyline class="guide" points="{gp}"/>'
            f'<circle class="pt" cx="{X(lw):.1f}" cy="{Y(act[lw]):.1f}" r="2"/></svg>')
fw = min(fgw['F'])
spk = [
    ('<svg class="ip-icon ip-icon--14" viewBox="0 0 24 24" aria-hidden="true" style="color:var(--rec-eggs)"><path d="M12 22c4 0 7-3 7-7.5C19 9 16 2 12 2S5 9 5 14.5C5 19 8 22 12 22z"/></svg>Ouat',
     pct(lay[40]), pct(glay[40]), fdec(lay[40] - glay[40], 1, True) + ' puncte',
     spark({w: v for w, v in lay.items() if w >= 3}, glay, '',
           f'Ouat sapt. 3–40 față de ghid: efectiv {pct(lay[40])} în sapt. 40, maxim {pct(max(lay.values()))}; ghid de la sapt. 7, {pct(glay[40])} în sapt. 40')),
    ('<svg class="ip-icon ip-icon--14" viewBox="0 0 24 24" aria-hidden="true" style="color:var(--rec-feed)"><path d="M12 21V9M12 9c-3 0-4-2-4-5 3 0 4 2 4 5zM12 9c3 0 4-2 4-5-3 0-4 2-4 5zM12 15c-3 0-4-2-4-5 3 0 4 2 4 5zM12 15c3 0 4-2 4-5-3 0-4 2-4 5z"/></svg>Hrană <span class="ip-sex ip-sex--hen">♀</span>',
     f'{fint(fgw["F"][40])} g', f'{gfeed["F"][40]} g', pct((fgw['F'][40] / gfeed['F'][40] - 1) * 100, 1, True),
     spark(fgw['F'], gfeed['F'], ' ip-spark--hen', f'Hrană găini sapt. {fw}–40 față de ghid, g pe pasăre: {fint(fgw["F"][40])} g în sapt. 40, ghid {gfeed["F"][40]} g')),
    ('<svg class="ip-icon ip-icon--14" viewBox="0 0 24 24" aria-hidden="true" style="color:var(--rec-feed)"><path d="M12 21V9M12 9c-3 0-4-2-4-5 3 0 4 2 4 5zM12 9c3 0 4-2 4-5-3 0-4 2-4 5zM12 15c-3 0-4-2-4-5 3 0 4 2 4 5zM12 15c3 0 4-2 4-5-3 0-4 2-4 5z"/></svg>Hrană <span class="ip-sex ip-sex--rooster">♂</span>',
     f'{fint(fgw["M"][40])} g', f'{gfeed["M"][40]} g', pct((fgw['M'][40] / gfeed['M'][40] - 1) * 100, 1, True),
     spark(fgw['M'], gfeed['M'], ' ip-spark--rooster', f'Hrană cocoși sapt. {fw}–40 față de ghid, g pe pasăre: {fint(fgw["M"][40])} g în sapt. 40, ghid {gfeed["M"][40]} g')),
]
v6 = ('<div class="ip-tablescroll" tabindex="0" role="region" aria-label="Curbe mici față de ghid, Hîrbovăț SAPT 40"><table class="ip-table ip-sparktable"><caption class="ip-sr">Hîrbovăț · Seria 2026 față de ghidul Ross 308 ouat, SAPT 40</caption><thead><tr>'
      '<th scope="col">Indicator</th><th scope="col" class="num">SAPT 40</th><th scope="col" class="num">Ghid</th><th scope="col" class="num">Abatere</th><th scope="col">Curba · SAPT 1–40</th></tr></thead><tbody>'
      + ''.join(f'<tr><th scope="row"><span class="ip-sparktable__key">{a}</span></th><td class="num">{b}</td><td class="num">{c}</td><td class="num">{d}</td><td class="ip-sparktable__curve">{e}</td></tr>' for a, b, c, d, e in spk)
      + '</tbody></table></div>')
# the same small curve on a Readout (S20 guide detail)
v6b = readout('OUAT · HÎRBOVĂȚ · SAPT 40', 'ok', 'ÎN NORMĂ',
              '<div class="ip-row" style="gap:var(--space-4);align-items:center"><div class="ip-readout__val">' + fdec(lay[40]) + '<small>%</small></div>' + spk[0][4] + '</div>',
              f'<div class="ip-readout__sub">ghid {pct(glay[40])} · {fdec(lay[40] - glay[40], 1, True)} puncte</div>')

def sect(title, body):
    return f'<div class="ip-stack" style="gap:var(--space-2)"><span class="ip-cap">{title}</span>{body}</div>'
html = ('<!-- @dsCard group="Moment" height=2076 subtitle="GuideChart variants (SYS06)" -->\n'
        '<div class="ip-root ip-stack" style="gap:var(--space-6);max-width:720px">'
        + sect('Mortalitate % · cumulată ♀ și ♂ · focus, cursor pe SAPT 40', v1)
        + sect('Hrană kg/zi · pe tip de hrană', v2)
        + sect('Greutate · uniformitate și CV ♀ · fără date în ghid', v3)
        + sect('Producție · incubabile % · fără date în ghid', v4)
        + sect('Producție · masa oului, g · fără date în ghid', v5)
        + sect('Curbă mică · 160×48 · în tabel (S20, rânduri de tabel) · efectiv + ghid, fără axe', v6)
        + sect('Curbă mică · pe un Readout', v6b)
        + f'<p class="ip-note">{HIR_NOTE} Ghidul de ouat este exemplu (D29); grame ghid reale (guide-laying.json). Mortalitate, uniformitate, CV, incubabile și masa oului nu au ghid (D29).</p></div>\n')
os.makedirs(os.path.join(COMP, 'GuideChartVariants'), exist_ok=True)
open(os.path.join(COMP, 'GuideChartVariants/preview.html'), 'w').write(html)

# ---------------------------------------------------------------- base GuideChart preview: Telița ♀ weight, real
W = J('weighings-weekly.json'); M = J('mortality-telita-daily.json'); GG = J('guide-growth.json')
alive = {(r['date'], r['section']): r['alive_calc'] for r in M['rows'] if r['sex'] == 'F'}
byw = C.defaultdict(lambda: [0, 0])
for r in W['rows']:
    if r['sex'] == 'F':
        n = alive[(r['date'], r['section'])]; byw[r['age_week']][0] += r['avg_weight_g'] * n; byw[r['age_week']][1] += n
tw = {w: a[0] / a[1] for w, a in byw.items()}
gw = {r['week']: r['body_weight_g'] for r in GG['rows'] if r['sex'] == 'F' and r['body_weight_g'] and r['week'] <= 21}
lo = {w: v * 0.95 for w, v in gw.items()}; hi = {w: v * 1.05 for w, v in gw.items()}
d20 = (tw[20] / gw[20] - 1) * 100
rows = [[str(w), fint(tw[w]) + ' g' if w in tw else DASH + ' fără cântărire', fint(gw[w]) + ' g',
         pct((tw[w] / gw[w] - 1) * 100, 1, True) if w in tw else DASH] for w in range(1, 22)]
base = readout(f'GREUTATE ♀ · TELIȚA · SERIA 2026 · SAPT 20', 'warn', 'ABATERE',
    chart('gc1', 'Greutate găini Telița Seria 2026 față de ghid, săptămânile 1–21', 21,
          [{'cls': 'act', 'pts': tw}], [0, 1000, 2000, 3000], lambda t: '0' if t == 0 else f'{t // 1000} kg', 0, 3000,
          guide=gw, band=(lo, hi), cur=20, xticks=[1, 5, 10, 15, 21], pxticks=[1, 10, 21],
          read_txt=f'SAPT 20 · efectiv {fint(tw[20])} g · ghid {fint(gw[20])} g · {pct(d20, 1, True)}',
          table_cols=['SAPT', 'Efectiv', 'Ghid', 'Abatere'], table_rows=rows),
    legend([('act', f'{HEN} efectiv (medie pe hale după găini)'), ('guide', 'ghid'), ('band', 'bandă ±5%')])
    + '<div class="ip-readout__sub">ciclu închis 28.02.2026, ultima cântărire SAPT 20, SAPT 21 fără cântărire.</div>')
open(os.path.join(COMP, 'GuideChart/preview.html'), 'w').write(
    '<!-- @dsCard group="Moment" height=402 -->\n<div class="ip-root" style="max-width:720px">' + base +
    '<p class="ip-note">Telița · Seria 2026: greutăți reale (weighings-weekly.json, medie ponderată cu găinile din hală), ghid de creștere real (guide-growth.json).</p></div>\n')

# ---------------------------------------------------------------- GuideMissing: the two GuideChart states, regenerated
# (fix round 1: phone drawing, legend swatches, today line). The Stat part above them stays as written in SYS01.
# Demo curves are sample (D38a); the growth guide is real (guide-growth.json, ♀ body weight SAPT 1–23).
GM_ACT = [125, 244, 370, 493, 626, 758, 885, 998, 1109, 1201, 1293, 1395, 1488, 1589, 1678, 1783, 1904, 2016, 2145, 2270,
          2427, 2586, 2726, 2851, 2967]
GM_LAY = [0.0, 0.0, 0.0, 18.0, 29.3, 39.0, 47.6, 55.6, 63.3, 70.5, 77.4, 84.0, 83.7, 83.5, 83.3, 83.0, 82.7, 82.5, 82.2, 82.0,
          81.7, 81.5, 81.2, 81.0, 80.7, 80.5, 80.2, 80.0, 79.7, 79.5, 79.2, 79.0, 78.7, 78.5, 78.2, 78.0, 77.7, 77.5, 77.2, 77.0]
pa = {i + 1: v for i, v in enumerate(GM_ACT)}
pg = {r['week']: r['body_weight_g'] for r in GG['rows'] if r['sex'] == 'F' and r['body_weight_g'] and r['week'] <= 23}
plo = {w: v * 0.95 for w, v in pg.items()}; phi = {w: v * 1.05 for w, v in pg.items()}
NOG = DASH + ' fără date în ghid'
rows = [[str(w), fint(pa[w]) + ' g', fint(pg[w]) + ' g' if w in pg else NOG] for w in range(1, 26)]
d12 = (pa[12] / pg[12] - 1) * 100
gm1 = readout('GREUTATE ♀ · FERMĂ EXEMPLU · CREȘTERE · SAPT 25', 'off', 'FĂRĂ DATE ÎN GHID',
    chart('gm1', 'Greutate ♀ Fermă exemplu · creștere, față de ghid, săptămânile 1–25; ghidul are greutate până în săptămâna 23', 25,
          [{'cls': 'act', 'pts': pa}], [0, 1000, 2000, 3000], lambda t: '0' if t == 0 else f'{t // 1000} kg', 0, 3000,
          guide=pg, band=(plo, phi), today=25, cur=12, focus=True, yt=30, gend_note='fără date în ghid după SAPT 23',
          xticks=[1, 5, 10, 15, 20, 25], pxticks=[1, 13, 25],
          read_txt=f'SAPT 12 · efectiv {fint(pa[12])} g · ghid {fint(pg[12])} g · {pct(d12, 1, True)}',
          caption='Greutate ♀ Fermă exemplu · creștere, față de ghid; ghidul are greutate până în săptămâna 23',
          table_cols=['SAPT', 'Efectiv', 'Ghid'], table_rows=rows),
    legend([('act', f'{HEN} efectiv'), ('guide', 'ghid, până la SAPT 23'), ('band', 'bandă ±5%')])
    + '<div class="ip-readout__sub">Fermă exemplu · creștere: ghidul de creștere are greutate până la SAPT 23; după, fără date în ghid. Ghid real, efectiv exemplu.</div>')
la = {i + 1: v for i, v in enumerate(GM_LAY)}
gm2 = readout('OUAT · FERMĂ EXEMPLU · SAPT 40', 'off', 'FĂRĂ GHID',
    chart('gm2', 'Ouat efectiv, Fermă exemplu, săptămânile 1–40; ciclul nu are ghid', 40,
          [{'cls': 'act', 'pts': la}], [0, 50, 100], lambda t: f'{t}%', 0, 100,
          today=40, cur=40, xticks=XT, pxticks=PXT,
          read_txt=f'SAPT 40 · efectiv {pct(la[40])} · ghid {DASH} fără ghid',
          table_cols=['SAPT', 'Efectiv', 'Ghid'], table_rows=[[str(w), pct(la[w]), DASH + ' fără ghid'] for w in range(1, 41)]),
    legend([('act', 'efectiv')], f'ghid: {DASH} fără ghid')
    + '<div class="ip-readout__sub">ciclul nu are ghid, nu se compară · date exemplu</div>')
p = os.path.join(COMP, 'GuideMissing/preview.html'); t = open(p).read()
cut = t.index('<div class="ip-stack" style="gap:var(--space-2)"><span class="ip-cap">GuideChart · ghid parțial')
t = (t[:cut] + sect('GuideChart · ghid parțial · focus, ← până la SAPT 12 · azi SAPT 25', gm1) + sect('GuideChart · fără ghid · Fermă exemplu', gm2)
     + '<p class="ip-note">Hîrbovăț: date exemplu la 16.11.2026 (D27), ghid ouat 74,1% (D29). Fermă exemplu · creștere: ghidul de creștere '
       '(greutate ♀, SAPT 1–23) e real; curba efectivă e exemplu. Fermă exemplu: date exemplu.</p></div>\n')
open(p, 'w').write(t)
print('ok', round(devF, 1), round(devM, 1), pct(mF[40], 2), pct(mM[40], 2), fint(kF[40]), fint(kM[40]), pct(un[40]), pct(cvp[40]), round(d20, 1))
