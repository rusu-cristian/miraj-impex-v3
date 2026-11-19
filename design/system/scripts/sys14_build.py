"""Build the SYS14 preview (PeriodTable, spec components.md › Additions › PeriodTable (round 2, spec)): python3 sys14_build.py
Every figure comes from design/canvas/scripts/realdata.py (the DATA03 extracts, real to 18.11.2026, D152): Hîrbovăț,
Seria 2026, SAPT 56–59 newest first (D159), deaths and sales apart (D155), lay % on start-of-day hens (D156/D159).
The Greutate block is illustrative (D152: Hîrbovăț has no real weighings; DATA03's one H5.1 weighing of 10.11.2026)."""
import sys, pathlib, datetime as dt
REPO = pathlib.Path(__file__).resolve().parents[3]  # repo root from design/system/scripts
sys.path.insert(0, str(REPO / 'design/canvas/scripts'))
import realdata as R
import sexsign as SXS  # SX01: every ♀/♂ of the preview = the §72 hook
OUT = REPO / 'design/system/components/PeriodTable/preview.html'

DASH = '–'
WEEKS = [59, 58, 57, 56]                    # D159: newest first, the last 4 weeks
P_FROM, P_TO = R.week_dates(56)[0], R.TODAY
WD = ['Lu', 'Ma', 'Mi', 'Jo', 'Vi', 'Sâ', 'Du']
WDF = ['luni', 'marți', 'miercuri', 'joi', 'vineri', 'sâmbătă', 'duminică']
H7_EMPTY = R.empty_since('H7')              # 14.11.2026 (D158)
GOALA = f'goală din {R.ddm(H7_EMPTY)}'

def svg(paths, size=16):
    return f'<svg class="ip-icon ip-icon--{size}" viewBox="0 0 24 24" aria-hidden="true">{paths}</svg>'
IC = {
    'mortality': '<path d="M12 3a7 7 0 0 0-7 7v4l2 2v3h10v-3l2-2v-4a7 7 0 0 0-7-7z"/><circle cx="9.5" cy="11" r="1.2"/><circle cx="14.5" cy="11" r="1.2"/><path d="M11 19v-2M13 19v-2"/>',
    'sale': '<path d="M3 7h11v9H3zM14 10h4l3 3v3h-7"/><circle cx="7" cy="18" r="2"/><circle cx="17" cy="18" r="2"/>',
    'feed': '<path d="M12 21V9M12 9c-3 0-4-2-4-5 3 0 4 2 4 5zM12 9c3 0 4-2 4-5-3 0-4 2-4 5zM12 15c-3 0-4-2-4-5 3 0 4 2 4 5zM12 15c3 0 4-2 4-5-3 0-4 2-4 5z"/>',
    'eggs': '<path d="M12 22c4 0 7-3 7-7.5C19 9 16 2 12 2S5 9 5 14.5C5 19 8 22 12 22z"/>',
    'hen': '<circle cx="12" cy="9" r="5"/><path d="M12 14v8M9 19h6"/>',
    'weighing': '<path d="M12 3v18"/><path d="M7 21h10"/><path d="M3 7h18"/><path d="m6 7-3 7a3 3 0 0 0 6 0z"/><path d="m18 7-3 7a3 3 0 0 0 6 0z"/>',
}
CHEV = {'open': '<path d="m6 9 6 6 6-6"/>', 'closed': '<path d="m9 18 6-6-6-6"/>'}

def num(s): return f'<span class="ip-num">{s}</span>'
def ico(t): return f'<span class="ip-rec ip-rec--{t} ip-table__ico">{svg(IC[t], 14)}</span>'
SX = {'F': ('is-hen', '♀', 'găini'), 'M': ('is-rooster', '♂', 'cocoși')}
def sx(s, word=''):
    c, g, w = SX[s]
    return f'<span class="{c}">{g}</span><span class="ip-sr"> {w}</span>' + (f' {word}' if word else '')
def nd(v, dec=0): return DASH if v is None else R.n(v, dec)

def days_of(k):
    a, b = R.week_dates(k); out = []
    while a <= min(b, R.TODAY): out.append(a); a += dt.timedelta(days=1)
    return out
def wk_label(k):
    a, b = R.week_dates(k); ds = days_of(k)
    return a, min(b, R.TODAY), len(ds)
class Cal:
    """the flock's calendar a table is laid out on (D151 age weeks): the current flock by default; S4NG (D170) passes the
    previous flock's (Seria 2025, placement inferred) for the past-cycle table"""
    def __init__(self, placed=None, last=None, weeks=None):
        self.placed, self.last, self.weeks = placed, last, weeks
    def age(self, d): return R.age_days(d) if self.placed is None else (d - self.placed).days
    def sapt(self, d): return R.sapt(d) if self.placed is None else self.age(d) // 7 + 1
    def week_dates(self, k):
        if self.placed is None: return R.week_dates(k)
        a = self.placed + dt.timedelta(days=(k - 1) * 7); return a, a + dt.timedelta(days=6)
    def days(self, k):
        if self.placed is None: return days_of(k)
        a, b = self.week_dates(k); out = []
        while a <= min(b, self.last): out.append(a); a += dt.timedelta(days=1)
        return out
    def label(self, k):
        a, b = self.week_dates(k); ds = self.days(k)
        return a, ds[-1] if self.placed is not None else min(b, R.TODAY), len(ds)
    def period(self, weeks):
        a, b = self.week_dates(weeks[-1])[0], (P_TO if self.placed is None else self.label(weeks[0])[1])
        return a, b
CAL = Cal()

# ---------- values: each leaf = (head html, class, value(kind, key) -> str | (DASH, reason)) ----------
# kind: 'w' week k, 'd' day date, 'f' the period
ARR_DAY = R.D(min(r['date'] for r in R.SECTION_DAY))   # 22.02.2026, first arrival at Hîrbovăț
ARR_HALL = {}
for _r in R.SECTION_DAY: ARR_HALL[(_r['section'].split('.')[0], _r['sex'])] = ARR_HALL.get((_r['section'].split('.')[0], _r['sex']), 0) + _r['arr']
def end_of(kind, key):
    if kind == 'w': return min(R.week_dates(key)[1], R.TODAY)
    return key if kind == 'd' else P_TO
def cum(s, h=None):
    """D162: cumulative mortality % = real deaths since arrival ÷ birds arrived (D155: sales not counted); ♀ 5,66 % on 18.11."""
    def f(kind, key):
        e = end_of(kind, key)
        if h == 'H7' and H7_EMPTY and e > H7_EMPTY: return (DASH, GOALA)
        base = ARR_HALL[(h, s)] if h else R.ARRIVED[s]
        return R.pct(R.deaths(ARR_DAY, e, sx=s, hall=h) / base * 100, 1)
    return f
CUM_GROUP = lambda h=None: {'name': f'{ico("mortality")}% cumulat', 'leaves': [(sx('F'), 'is-hen', cum('F', h)), (sx('M'), 'is-rooster', cum('M', h))]}

# SYS17 (D189): phone group heads stay on one line; a group whose full word is wider than its two figure columns (so the tag
# would not fit 358 beside the sticky column, D171) shows the system's short word, the full word hidden (ip-sr), as Rezumat's
# short leaves (s4s_build.rezumat_groups). "Morți" = the record's deaths (Mortalitate's own word on Rezumat, spec D1).
PHONE_SHORT = {'Mortalitate': 'Morți'}
def stacked(main, sub, sub_word='% cumulat'):
    """Phone variant (S4NS fix 2, Claude, to review): a group whose companion rate rides in the same cell, as a second small Mono
    line under the count (Mortalitate ♀ "10" / "5,7%"), so the tag keeps 2 figure columns and the row stays 48. Group head:
    name + note line naming the second figure; hidden text names both values."""
    def wrap(fa, fb):
        def f(kind, key):
            a, b = fa(kind, key), fb(kind, key)
            if isinstance(a, tuple): return a
            bb = f'{DASH}<span class="ip-sr"> {b[1]}</span>' if isinstance(b, tuple) else b
            return f'{a}<span class="ip-ptable__pct"><span class="ip-sr">, {sub_word} </span>{bb}</span>'
        return f
    name = main['name']; icon, word = name.rsplit('</span>', 1) if '</span>' in name else ('', name)
    if word in PHONE_SHORT:   # SYS17 (D189): the head stays on one line; where the full word would not fit 358, the short word shows
        name = (icon + '</span>' if icon else '') + f'<span aria-hidden="true">{PHONE_SHORT[word]}</span><span class="ip-sr">{word}</span>'
    return {'name': f'{name}<span class="ip-ptable__note">cu {sub_word.replace(" ", chr(160))}</span>',
            'leaves': [(lab, c, wrap(fa, fb)) for (lab, c, fa), (_, _, fb) in zip(main['leaves'], sub['leaves'])]}

def farm_pasari():
    def mort(s): return lambda kind, key: R.n(R.week(key)[f'mort_{s}'] if kind == 'w' else R.day(key)[f'mort_{s}'] if kind == 'd' else R.deaths(P_FROM, P_TO, sx=s))
    def sold(s): return lambda kind, key: R.n(R.week(key)[f'sold_{s}'] if kind == 'w' else R.day(key)[f'sold_{s}'] if kind == 'd' else R.sold(P_FROM, P_TO, sx=s))
    def alive(s): return lambda kind, key: R.n(R.week(key)[f'alive_{s}_end'] if kind == 'w' else R.day(key)[f'alive_{s}'] if kind == 'd' else R.alive_farm(s, P_TO))
    return [   # D167: "În viață" first (right after the spine), "Vânzare" always the last group
        {'name': 'În viață', 'leaves': [(sx('F'), 'is-hen', alive('F')), (sx('M'), 'is-rooster', alive('M'))]},
        {'name': f'{ico("mortality")}Mortalitate', 'leaves': [(sx('F'), 'is-hen', mort('F')), (sx('M'), 'is-rooster', mort('M'))]},
        CUM_GROUP(),
        {'name': f'{ico("sale")}Vânzare', 'leaves': [(sx('F'), 'is-hen', sold('F')), (sx('M'), 'is-rooster', sold('M'))]},
    ]

def hall_pasari(h):
    """Păsări for one hall: a day after the hall emptied, or a week/period end after it, is – (goală din 14.11)."""
    def empty_day(d): return R.day(d, h) is None
    def mort(s):
        def f(kind, key):
            if kind == 'w':
                w = R.week(key, h)
                return (DASH, GOALA) if not w['alive_F_start'] and not w['alive_M_start'] else R.n(w[f'mort_{s}'])
            if kind == 'd': return (DASH, GOALA) if empty_day(key) else R.n(R.day(key, h)[f'mort_{s}'])
            return R.n(R.deaths(P_FROM, P_TO, sx=s, hall=h))
        return f
    def sold(s):
        def f(kind, key):
            if kind == 'w':
                w = R.week(key, h)
                return (DASH, GOALA) if not w['alive_F_start'] and not w['alive_M_start'] else R.n(w[f'sold_{s}'])
            if kind == 'd': return (DASH, GOALA) if empty_day(key) else R.n(R.day(key, h)[f'sold_{s}'])
            return R.n(R.sold(P_FROM, P_TO, sx=s, hall=h))
        return f
    def alive(s):
        def f(kind, key):
            if kind == 'w':
                a, b = R.week_dates(key); end = min(b, R.TODAY)
                return (DASH, GOALA) if (H7_EMPTY and h == 'H7' and end > H7_EMPTY) else R.n(R.week(key, h)[f'alive_{s}_end'])
            if kind == 'd': return (DASH, GOALA) if empty_day(key) else R.n(R.day(key, h)[f'alive_{s}'])
            return (DASH, GOALA) if (h == 'H7' and H7_EMPTY) else R.n(R.alive_hall(h, s, P_TO))
        return f
    return [   # D167 order, as farm_pasari
        {'name': 'În viață', 'leaves': [(sx('F'), 'is-hen', alive('F')), (sx('M'), 'is-rooster', alive('M'))]},
        {'name': f'{ico("mortality")}Mortalitate', 'leaves': [(sx('F'), 'is-hen', mort('F')), (sx('M'), 'is-rooster', mort('M'))]},
        CUM_GROUP(h),
        {'name': f'{ico("sale")}Vânzare', 'leaves': [(sx('F'), 'is-hen', sold('F')), (sx('M'), 'is-rooster', sold('M'))]},
    ]

CAT_HEAD = {c: f'%\u00a0Cat.\u00a0{c}' for c in (1, 2, 3)}   # D183 heads "% Cat. 1" (prefix form as % ouat), never split
def productie(guide=True, h=None, weeks=WEEKS):
    """Farm Producție (D3/D1): lay % = eggs ÷ start-of-day hens; week and period = eggs ÷ hen-days (D156, D159).
    h (S14N): one hall's figures (the report's hall week rows, spec D3 rowGroups), on the farm's guide of the age week; a
    hall week (or day) without hens at its start is – (goală din 14.11, D158)."""
    def W(k): return R.week(k) if h is None else R.week(k, h)
    def D_(d): return R.day(d) if h is None else R.day(d, h)
    def gone(kind, key):
        if h is None or kind == 'f': return False
        return (not W(key)['alive_F_start']) if kind == 'w' else (D_(key) is None or not D_(key)['eggs'])
    def guard(fn):
        def f(kind, key): return (DASH, GOALA) if gone(kind, key) else fn(kind, key)
        return f
    hd = sum(W(k)['hen_days'] for k in weeks)
    def hens(kind, key):
        if kind == 'w': return R.n(W(key)['alive_F_start'])
        if kind == 'f': return R.n(W(weeks[-1])['alive_F_start'])
        prev = key - dt.timedelta(days=1)
        return R.n(R.alive_farm('F', prev) if h is None else R.alive_hall(h, 'F', prev))
    def eggs(kind, key):
        return R.n(W(key)['eggs'] if kind == 'w' else R.eggs(key, h) if kind == 'd' else sum(W(k)['eggs'] or 0 for k in weeks))
    def lay_v(kind, key):
        if kind == 'w': return W(key)['eggs'] / W(key)['hen_days'] * 100
        if kind == 'd': return R.lay_day(key, h)
        return sum(W(k)['eggs'] or 0 for k in weeks) / hd * 100
    def gl_v(kind, key):
        if kind == 'w': return R.week(key)['guide']['lay_pct']
        if kind == 'd': return R.day(key)['guide_lay_pct']
        return sum(R.week(k)['guide']['lay_pct'] * W(k)['hen_days'] for k in weeks) / hd   # hen-day weighted
    def lay(kind, key): return R.pct(lay_v(kind, key))
    def gl(kind, key): return R.pct(gl_v(kind, key))
    def dv(kind, key): return R.rnd(lay_v(kind, key), 1) - R.rnd(gl_v(kind, key), 1)
    def vs(kind, key): return R.sgn(dv(kind, key), 1) + ' pct'
    def stare(kind, key):   # ±3 pct band, as S14 (craft: lamp + word)
        return ('<span class="ip-lamp ip-lamp--warn">abatere</span>' if abs(dv(kind, key)) > 3
                else '<span class="ip-lamp ip-lamp--ok">în normă</span>')
    def inc(kind, key):
        return R.n(W(key)['inc'] if kind == 'w' else D_(key)['inc'] if kind == 'd' else sum(W(k)['inc'] for k in weeks))
    def jos(kind, key):
        return R.n(W(key)['jos'] if kind == 'w' else D_(key)['jos'] if kind == 'd' else sum(W(k)['jos'] for k in weeks))
    def share(field):
        def f(kind, key):
            if kind == 'w': a, b = W(key)[field], W(key)['eggs']
            elif kind == 'd': a, b = D_(key)[field], D_(key)['eggs']
            else: a, b = sum(W(k)[field] for k in weeks), sum(W(k)['eggs'] or 0 for k in weeks)
            return R.pct(a / b * 100)
        return f
    def egg_g(kind, key):
        if kind == 'w': return R.n(W(key)['egg_g'], 1)
        if kind == 'd': return R.n(D_(key)['egg_g'], 1)
        return R.n(sum(W(k)['egg_g'] * W(k)['eggs'] for k in weeks if W(k)['eggs']) / sum(W(k)['eggs'] or 0 for k in weeks), 1)
    def cat(c):   # D183 (S4C): Cat. c as % of the row's eggs (form counts, DATA06), one decimal; a row with no eggs is – (never 0)
        fld = f'cat{c}'
        def f(kind, key):
            if kind == 'w': a, b = W(key)[fld], W(key)['eggs']
            elif kind == 'd': r = D_(key); a, b = (r[fld], r['eggs']) if r else (None, None)
            else: a, b = sum(W(k)[fld] or 0 for k in weeks), sum(W(k)['eggs'] or 0 for k in weeks)
            if not b: return (DASH, 'fără ouă')
            if a is None: return (DASH, 'fără categorii în formular')
            return R.pct(a / b * 100)
        return f
    if h is not None:
        hens, eggs, lay, gl, vs, stare, inc, jos, egg_g = (guard(f) for f in (hens, eggs, lay, gl, vs, stare, inc, jos, egg_g))
        share = (lambda s: (lambda fld: guard(s(fld))))(share)
        cat = (lambda c0: (lambda c: guard(c0(c))))(cat)
    lay_leaves = [('ouă', '', eggs), ('% ouat', '', lay)]
    if guide: lay_leaves += [('ghid', '', gl), ('față de ghid', '', vs), ('stare', 'txt', stare)]
    lay_leaves += [(CAT_HEAD[c], '', cat(c)) for c in (1, 2, 3)]   # D183/KIT03: after stare (after % ouat without a guide), before incub.
    return [
        {'name': f'{ico("hen")}Găini', 'single': True, 'sr': ' la început de zi', 'cls': 'is-hen', 'leaves': [('', 'is-hen', hens)]},
        {'name': f'{ico("eggs")}Ouat', 'leaves': lay_leaves},
        {'name': f'{ico("eggs")}Incubabile', 'leaves': [('ouă', '', inc), ('%', '', share('inc'))]},
        {'name': f'{ico("eggs")}Jos', 'leaves': [('ouă', '', jos), ('%', '', share('jos'))]},
        {'name': f'{ico("eggs")}Masă ou', 'single': True, 'sr': ', grame', 'leaves': [('g', '', egg_g)]},
    ]
def _cat_check():   # D183/DATA06: the week cells = the extract's cat*_pct (count ÷ eggs × 100) to one decimal
    ou = [g for g in productie() if g['name'].endswith('Ouat')][0]; lv = {lab: fn for lab, _, fn in ou['leaves']}
    for k in WEEKS:
        for c in (1, 2, 3): assert lv[CAT_HEAD[c]]('w', k) == R.pct(R.week(k)[f'cat{c}_pct']), (k, c)
    assert [lv[CAT_HEAD[c]]('w', 59) for c in (1, 2, 3)] == ['0,3%', '6,2%', '1,1%']   # spec D3 sample (KIT03)
_cat_check()
def group(groups, word):   # a record group by its visible word ("În viață", "Mortalitate" …), so callers never index by position
    g = [r for r in groups if r['name'].rsplit('</span>', 1)[-1] == word]; assert len(g) == 1, word; return g[0]

def hrana(h=None):
    """Hrană g/pasăre (kg ÷ bird-days, D38) vs guide, and kg, per sex; farm (h None) or one hall; days, weeks, period."""
    def wk(k): return R.week(k) if h is None else R.week(k, h)
    def bd(k, s):  # bird-days of week k for sex s, as the week's g was computed (kg ÷ bird-days)
        w = wk(k); return w[f'kg_{s}'] * 1000 / w[f'g_{s}'] if w[f'g_{s}'] else 0
    def gone(k): w = wk(k); return h is not None and not w[f'kg_F'] and not w[f'kg_M']
    def dgone(d): return h is not None and R.day(d, h) is None   # after the hall emptied
    NOFEED = (DASH, 'fără hrană în registru')
    def dg(d, s):   # g/bird of a day: farm = R.feed_g; hall = kg ÷ birds fed over its compartments
        if h is None: return R.feed_g(d, s)
        rs = [r for r in (R.sd(d, c, s) for c in R.HALL_SECS[h]) if r and r['kg']]
        b = sum(r['kg'] * 1000 / r['g'] for r in rs); return sum(r['kg'] for r in rs) * 1000 / b if b else None
    def g(s):
        def f(kind, key):
            if kind == 'w': return (DASH, GOALA) if gone(key) else R.n(wk(key)[f'g_{s}'], 1)
            if kind == 'd': return (DASH, GOALA) if dgone(key) else (NOFEED if dg(key, s) is None else R.n(dg(key, s), 1))
            t = sum(bd(k, s) for k in WEEKS); return R.n(sum(wk(k)[f'kg_{s}'] for k in WEEKS) * 1000 / t, 1)
        return f
    def gd(s):
        def f(kind, key):
            if kind == 'w': return (DASH, GOALA) if gone(key) else R.n(R.week(key)['guide'][f'feed_g_{s}'])
            if kind == 'd': return (DASH, GOALA) if dgone(key) else R.n(R.day(key)[f'guide_g_{s}'])   # the guide holds on a day without feed
            t = sum(bd(k, s) for k in WEEKS); return R.n(sum(R.week(k)['guide'][f'feed_g_{s}'] * bd(k, s) for k in WEEKS) / t)
        return f
    def kg(s):
        def f(kind, key):
            if kind == 'w': return (DASH, GOALA) if gone(key) else R.n(wk(key)[f'kg_{s}'], 1)
            if kind == 'd': return (DASH, GOALA) if dgone(key) else R.n(R.day(key, h)[f'kg_{s}'] if h else R.feed_kg(key, s), 1)
            return R.n(sum(wk(k)[f'kg_{s}'] for k in WEEKS), 1)
        return f
    return [
        {'name': f'{ico("feed")}Hrană · g/pasăre', 'leaves': [(sx('F', 'efectiv'), 'is-hen', g('F')), (sx('F', 'ghid'), 'is-hen', gd('F')),
                                                         (sx('M', 'efectiv'), 'is-rooster', g('M')), (sx('M', 'ghid'), 'is-rooster', gd('M'))]},
        {'name': f'{ico("feed")}Hrană · kg', 'leaves': [(sx('F'), 'is-hen', kg('F')), (sx('M'), 'is-rooster', kg('M'))]},
    ]

# Greutate · H5.1 (illustrative weighing 10.11.2026, DATA03; guide weights stop at SAPT 25: partial guide, D29)
WEIGH = {(w['sex']): w for w in R.WEIGHINGS_HIR}
WEIGH_DAY = R.D(R.WEIGHINGS_HIR[0]['date'])
WEIGH_WK = R.WEIGHINGS_HIR[0]['sapt']
def greutate():
    def v(s, field, dec=0, unit=''):
        def f(kind, key):
            if kind == 'w' and key != WEIGH_WK: return (DASH, 'fără cântărire')
            x = WEIGH[s][field]
            return (DASH, 'fără date în ghid') if x is None else R.n(x, dec) + unit
        return f
    def nog(kind, key):
        return (DASH, 'fără cântărire') if kind == 'w' and key != WEIGH_WK else (DASH, 'fără date în ghid')
    out = []
    for s in 'FM':
        out += [(sx(s, 'medie g'), SX[s][0], v(s, 'avg_weight_g')), (sx(s, 'ghid'), SX[s][0], nog),
                (sx(s, 'față de ghid'), SX[s][0], nog), (sx(s, 'CV %'), SX[s][0], v(s, 'cv_pct', 1)),
                (sx(s, 'unif. %'), SX[s][0], v(s, 'uniformity_pct', 0))]
    return [{'name': f'{ico("weighing")}Cântărire', 'leaves': out}]

# ---------- table ----------
# SYS15 (D175) + SYS16 (D177): every record group's cells (head and body) carry .ip-ptable__g and the group's record type,
# .ip-ptable__g--<type>, tinted in its filter tag (RecordChip) colour; a group with no tag (În viață, Găini) carries
# .ip-ptable__g alone (untinted), the spine and the section heads carry none. Its first column draws the group's start edge
# (a full-height rule, additions.css section 71), which also parts same-type neighbours (Mortalitate · % cumulat, Ouat ·
# Incubabile · Jos · Masă ou). Markup hooks only: the look is in CSS.
import re as _re
GTYPES = ('mortality', 'cull', 'sexerr', 'transfer', 'placement', 'sale', 'feed', 'weighing', 'eggs')   # RecordChip types (hen = Găini: no tag)
def rec_of(g):
    """a group's record type: g['rec'] when given, else its head icon's ip-rec--<type> (thico maps delivery/price to feed),
    else "Eroare sexare" (its split icon has no record colour) → sexerr; '' = no tag (În viață, Găini)."""
    if 'rec' in g: return g['rec'] or ''
    m = _re.search(r'ip-rec--(\w+)', g['name'])
    if m: return m.group(1) if m.group(1) in GTYPES else ''
    return 'sexerr' if 'Eroare sexare' in _re.sub(r'<[^>]+>', '', g['name']) else ''
def _gcls(rec, start, sec=False):
    return ' '.join(x for x in ('ip-ptable__g', f'ip-ptable__g--{rec}' if rec else '', 'ip-ptable__gs' if start else '',
                                'ip-ptable__sec' if sec else '') if x)
def cell(val, cls, sec):
    """sec: a leaf's group classes (leaves_of: group + record type, start edge, section edge); True / False kept for older callers."""
    g = sec if isinstance(sec, str) else ('ip-ptable__sec' if sec else '')
    c = ' '.join(x for x in (('' if cls == 'txt' else 'num'), ('' if cls == 'txt' else cls), g) if x)
    if isinstance(val, tuple):
        return f'<td class="{c}">{DASH}<span class="ip-sr"> {val[1]}</span></td>'
    return f'<td class="{c}">{val}</td>'

def head(sections, spine, show_sec, colw=None, fcol=None):
    """sections: [(name html, note, records)]. Rows: section (if show_sec) > record > leaf. fcol: the S3 "Fermă" text column
    (default: spine 's3'; S3N: also after the phone's sticky first column)."""
    if fcol is None: fcol = spine == 's3'
    nrows = (1 if show_sec else 0) + 2
    # SYS17 (D189): the phone spine head on one line in the sticky column: the short "Data · Vârstă", "(zile)" hidden (ip-sr),
    # so the head is no wider than "Ma 17.11 · 406" / "SAPT 58" and one tag's group still fits 358 (D171)
    data_h = 'Data · Vârstă<span class="ip-sr"> (zile)</span>' if spine == 'phone' else 'Data'
    r0 = f'<th scope="col" class="ip-ptable__data" rowspan="{nrows}">{data_h}</th>'
    cw = list(colw or [])
    def st():
        v = cw.pop(0) if cw else None
        return f' style="--ip-colw:{v if isinstance(v, str) else f"{v}%"}"' if v else ''   # S14N: a str is a length as given (px: a table wider than its page)
    if spine != 'phone':
        r0 += f'<th scope="col" class="num" rowspan="{nrows}"{st()}>Vârstă (zile)</th><th scope="col" class="num" rowspan="{nrows}"{st()}>SAPT</th>'
    if fcol:
        r0 += f'<th scope="col" rowspan="{nrows}">Fermă</th>'
    rows = [r0, '', ''] if show_sec else [r0, '']
    ri = 1 if show_sec else 0
    for si, (sname, note, recs) in enumerate(sections):
        width = sum(len(r['leaves']) for r in recs)
        if show_sec:
            nt = f'<span class="ip-ptable__note">{note}</span>' if note else ''
            rows[0] += f'<th scope="colgroup" colspan="{width}" class="ip-ptable__sec ip-ptable__gs">{sname}{nt}</th>'
        first = True
        for r in recs:
            rt = rec_of(r); g = _gcls(rt, True, first)
            if r.get('single'):
                sr = f'<span class="ip-sr">{r.get("sr", "")}</span>'
                lab = r['leaves'][0][0]
                lab = f' <span>{lab}</span>' if lab else ''
                rows[ri] += f'<th scope="col" class="num {g}" rowspan="2"{st()}>{r["name"]}{lab}{sr}</th>'
            else:
                rows[ri] += f'<th scope="colgroup" colspan="{len(r["leaves"])}" class="{g}">{r["name"]}</th>'
                for li, (lab, cls, _) in enumerate(r['leaves']):
                    nc = '' if cls == 'txt' else 'num'
                    rows[ri + 1] += f'<th scope="col" class="{(nc + " " + _gcls(rt, li == 0, first and li == 0)).strip()}"{st()}>{lab}</th>'
            first = False
    return '<thead>' + ''.join(f'<tr>{r}</tr>' for r in rows if r) + '</thead>'

def leaves_of(sections):
    """(cls, value fn, group classes) per leaf column; the group classes (SYS15/SYS16: group + record type, start edge,
    section edge) are a str."""
    out = []
    for (_, _, recs) in sections:
        first = True
        for r in recs:
            rt = rec_of(r)
            for li, (lab, cls, fn) in enumerate(r['leaves']):
                out.append((cls, fn, _gcls(rt, li == 0, first and li == 0)))
            first = False
    return out

def wd(d):
    return f'<span aria-hidden="true">{WD[d.weekday()]}</span><span class="ip-sr">{WDF[d.weekday()]}</span> {num(R.ddm(d))}'

def ptable(uid, caption, sections, open_weeks=None, spine='farm', show_sec=None, weeks=WEEKS, weighings=False,
           focus=None, dense=False, foot=True, skel=False, colw=None, cal=CAL, farm=None, rowgroups=None):
    """farm (S3N): (kind, key) -> the S3 "Fermă" text ('w' week, 'd' day, 'f' period); on the desktop spines it is the Fermă
    column (spine 's3' without it keeps "Hîrbovăț"); on the phone (D174) there is no column: a day whose farm() is not
    None (the transfer weeks) shows it as the sticky cell's second small line.
    rowgroups (S14N, spec D3/D4 and PeriodTable › rowGroups, the reports S11 Consum and S14): after a week's daily rows, one
    section week row per entry, th.ip-table__child (§32), always shown (week rows: grain Săpt. keeps them; days are farm-level
    only). Entry: {'label': visible row head, 'sr': hidden words after it, 'sections': the same shape as `sections` (one
    leaf per column), 'empty': k -> None | muted text, 'skip': k -> True to leave the row out that week}: a week the section had no birds is one muted line across the figure
    columns (D158d, "goală din 14.11.2026 · H7.1, H7.2")."""
    if show_sec is None: show_sec = len(sections) > 1
    fcol = spine == 's3' or (farm is not None and spine != 'phone')
    farm_of = farm or (lambda kind, key: 'Hîrbovăț')
    if open_weeks is None: open_weeks = tuple(weeks)   # D166: every week of the period open by default; () = grain Săpt.
    L = leaves_of(sections)
    cls = 'ip-table ip-table--rowheads ip-ptable' + (' ip-ptable--phone' if spine == 'phone' else '') + (' ip-ptable--dense' if dense else '')
    tid = f'{uid}-t'
    h = f'<table class="{cls}" id="{tid}"><caption class="ip-sr">{caption}</caption>' + head(sections, spine, show_sec, colw, fcol)
    def spine_cells(kind, key, label):
        if spine == 'phone': return ''
        if kind == 'w':
            a, b, _ = cal.label(key); age = f'{cal.age(a)}{DASH}{cal.age(b)}'
            s = f'<td class="num">{age}</td><td class="num">{key}</td>'
        elif kind == 'd':
            s = f'<td class="num">{cal.age(key)}</td><td class="num">{cal.sapt(key)}</td>'
        else:
            pa, pb = cal.period(weeks)
            s = f'<td class="num">{cal.age(pa)}{DASH}{cal.age(pb)}</td><td class="num">{weeks[-1]}{DASH}{weeks[0]}</td>'
        if fcol: s += f'<td>{farm_of(kind, key)}</td>'
        return s
    def figs(kind, key, L=L):   # S14N: L = another row's leaves of the same shape (rowgroups)
        if skel:
            # FIN11G: each bar as wide as its column's usual figure (ch of the cell's Plex Mono), at the figures' edge
            return ''.join(f'<td class="num {sec}"><span class="ip-skel ip-skel--ch{n} ip-skel--end"></span></td>'
                           for (c, fn, sec), n in zip(L, skel))
        return ''.join(cell(fn(kind, key), c, sec) for (c, fn, sec) in L)
    for k in weeks:
        a, b, n = cal.label(k)
        ds = cal.days(k)
        if weighings: ds = [d for d in ds if d == WEIGH_DAY]
        opened = k in open_weeks
        ids = [f'{uid}-{d.strftime("%m%d")}' for d in ds]
        dr = R.ddm(a) + DASH + R.ddm(b)
        no_w = weighings and not ds
        if no_w: ids = [f'{uid}-w{k}-none']   # D167: the week still opens, to one muted row "fără cântărire în această săptămână"
        if spine == 'phone':   # D171: the dates on a second small line (as the partial week's day count, D162), so one tag's group fits 358
            # a partial week's second line is its day count alone (D162 words; its dates are its open day rows and its hidden name):
            # "17–18.11 · 2 din 7 zile" would set the sticky column 30 px wider than a whole week's "10.11–16.11" (S4NS fix 1)
            sub = num(dr) if n == 7 else f'{n} din 7 zile'
            vis = f'SAPT {num(k)}<span class="ip-ptable__sub">{sub}</span>'
        else:
            vis = (f'SAPT {num(k)} <span class="ip-ptable__dates">· {num(dr)}' + (f' · {n} din 7 zile' if n < 7 else '') + '</span>')
        name = f'SAPT {k}, {dr}.{b.year}, ' + (f'{n} din 7 zile' if n < 7 else '7 zile')
        if weighings: name += ', fără cântărire' if no_w else f', {len(ds)} cântărire'
        lbl = f'<span aria-hidden="true">{vis}</span><span class="ip-sr">{name}</span>'
        if skel:
            th = f'<th scope="rowgroup"><span class="ip-skel ip-skel--w2"></span></th>'
        else:
            fc = ' ip-focus' if focus == k else ''
            th = (f'<th scope="rowgroup"><button type="button" class="ip-ptable__key{fc}" aria-expanded="{"true" if opened else "false"}" '
                  f'aria-controls="{" ".join(ids)}">{svg(CHEV["open" if opened else "closed"])}<span>{lbl}</span></button></th>')
        sk = ' ip-table__skel' if skel else ''
        rows = f'<tr class="is-total ip-ptable__wk{sk}">{th}{spine_cells("w", k, "") if not skel else skel_spine(spine, fcol)}{figs("w", k)}</tr>'
        if no_w and not skel:   # one row across the table, its text at the day labels' start (sticky in a scrolled region)
            ncol = (1 if spine == 'phone' else 3) + fcol + len(L)
            rows += (f'<tr class="ip-ptable__day ip-ptable__none" id="{ids[0]}"{"" if opened else " hidden"}><td colspan="{ncol}">'
                     f'<span>fără cântărire în această săptămână</span></td></tr>')
        for d, i in zip(ds, ids):
            hid = '' if opened else ' hidden'
            if skel:
                if not opened: continue
                rows += f'<tr class="ip-ptable__day{sk}" id="{i}"><th scope="row"><span class="ip-skel ip-skel--w3"></span></th>{skel_spine(spine, fcol)}{figs("d", d)}</tr>'
                continue
            lab = wd(d) + (f' · {num(cal.age(d))}' if spine == 'phone' else '')
            if spine == 'phone' and farm and farm('d', d): lab += f'<span class="ip-ptable__sub">{farm("d", d)}</span>'   # D174: the farm on a second line
            rows += f'<tr class="ip-ptable__day" id="{i}"{hid}><th scope="row">{lab}</th>{spine_cells("d", d, "")}{figs("d", d)}</tr>'
        for g in (rowgroups or ()):
            if g.get('skip', lambda k: False)(k): continue   # a section not drawn that week (S11: an emptied hall's compartments, one hall line instead)
            gl = g['label'] + f'<span class="ip-sr">{g.get("sr", "")}, SAPT {k}</span>'
            sp = spine_cells('w', k, '') if not skel else skel_spine(spine, fcol)
            if skel:
                rows += f'<tr class="ip-table__skel"><th scope="row" class="ip-table__child"><span class="ip-skel ip-skel--w3"></span></th>{sp}{figs("w", k)}</tr>'
                continue
            note = g.get('empty', lambda k: None)(k)
            if note:
                rows += (f'<tr><th scope="row" class="ip-table__child">{gl}</th>{sp}<td class="ip-muted ip-ptable__sec" colspan="{len(L)}">{note}</td></tr>')
                continue
            GL_ = leaves_of(g['sections']); assert len(GL_) == len(L), (g['label'], len(GL_), len(L))
            rows += f'<tr><th scope="row" class="ip-table__child">{gl}</th>{sp}{figs("w", k, GL_)}</tr>'
        h += f'<tbody class="ip-ptable__week">{rows}</tbody>'
    if foot:
        lab = 'Total perioadă' if not skel else '<span class="ip-skel ip-skel--w2"></span>'
        h += f'<tfoot><tr class="is-total{" ip-table__skel" if skel else ""}"><th scope="row">{lab}</th>{spine_cells("f", None, "") if not skel else skel_spine(spine, fcol)}{figs("f", None)}</tr></tfoot>'
    return h + '</table>', tid

def skel_spine(spine='farm', fcol=False):   # Vârstă ("406–407") ch4, SAPT ("59") ch2; S3N: the S3 Fermă text cell ("Hîrbovăț") ch4 at its start
    out = '' if spine == 'phone' else ''.join(f'<td class="num"><span class="ip-skel ip-skel--end ip-skel--ch{n}"></span></td>' for n in (4, 2))
    return out + ('<td><span class="ip-skel ip-skel--ch4"></span></td>' if fcol else '')

def region(table_html, caption, hidden=False):
    if hidden:
        return f'<div class="ip-tablescroll ip-tablescroll--framed" aria-hidden="true">{table_html}</div>'
    return f'<div class="ip-tablescroll ip-tablescroll--framed" role="region" tabindex="0" aria-label="{caption}">{table_html}</div>'

def detaliu(uid, sel='Fermă', touch=False):
    t = ' ip-seg--touch' if touch else ''
    items = ''.join(f'<span class="ip-seg__item{" is-selected" if o == sel else ""}" role="radio" aria-checked="{"true" if o == sel else "false"}" '
                    f'tabindex="{"0" if o == sel else "-1"}">{o}</span>' for o in ('Fermă', 'Hale', 'Compartimente'))
    return (f'<div class="ip-ptable__bar"><span class="ip-cap" id="{uid}-det">Detaliu</span>'
            f'<span class="ip-seg{t}" role="radiogroup" aria-labelledby="{uid}-det">{items}</span></div>')

def block(uid, caption, table_html, bar=True, sel='Fermă', touch=False, after=''):
    out = (detaliu(uid, sel, touch) if bar else '') + f'<p class="ip-cap ip-ptable__cap" aria-hidden="true">{caption}</p>' + region(table_html, caption)
    return out + after

def demo(cap, body, width='100%'):
    return (f'<section class="ip-stack" style="gap:var(--space-2);max-width:{width}">'
            f'<span class="ip-cap" style="color:var(--ink-secondary)">{cap}</span>{body}</section>')

SER = 'Hîrbovăț · Seria 2026 · SAPT 56–59'
# loaded column shares of the Păsări · Fermă table at 1440 (section 54: both copies declare them), px / 1126
# loaded column shares of the Păsări · Fermă table at 1440 (section 54: both copies declare them; px / 1126): only the
# columns whose figures are wider than their head (SAPT "56–59", "34,7%", Vânzare, În viață; Mortalitate ♀ · ♂ as wide as their group head); the others are sized by the head.
COLW = [None, round(63.2 / 1126 * 100, 2)] + [round(x / 1126 * 100, 2) for x in (71, 63.1, 71, 62, 60, 64, 71, 63.1)]   # D167 order: În viață, Mortalitate, % cumulat, Vânzare
html = '<!-- @dsCard group="Data" height=11488 subtitle="PeriodTable (SYS14)" -->\n<div class="ip-root ip-stack" style="gap:var(--space-12);max-width:1160px">\n'

# 1 Păsări · Fermă: default (D166: every week open), SAPT 58 key focused
c1 = f'Păsări · {SER}'
t1, _ = ptable('pa', c1, [('Total fermă', '', farm_pasari())], focus=58, colw=COLW)
html += demo('Păsări · Detaliu Fermă · implicit: toate săptămânile deschise, fiecare cu zilele ei (D166; SAPT 59 are 2 din 7 zile) · tasta SAPT 58 cu focus închide săptămâna',
             block('pa', c1, t1))

# 2 Păsări · Hale: emptying weeks, sales beside real deaths (D155), H7 empty from 14.11 (D158)
c2 = f'Păsări · {SER} · pe hale'
secs2 = [('Total fermă', '', farm_pasari())] + [(f'Hala {h[1:]}', GOALA if h == 'H7' else '', hall_pasari(h)) for h in R.HHALLS]
t2, _ = ptable('pb', c2, secs2)
html += demo('Păsări · Detaliu Hale · toate săptămânile deschise: golire prin vânzare (D155), Hala 7 goală din 14.11 (D158) · tabelul derulează lateral, prima coloană rămâne',
             block('pb', c2, t2, sel='Hale'))

# 3 Producție · Fermă: lay % on start-of-day hens (D156/D159)
c3 = f'Producție · {SER}'
t3, _ = ptable('pc', c3, [('Total fermă', '', productie())])
html += demo('Producție · Detaliu Fermă · % ouat: zi = ouă ÷ găini la început de zi; săptămână și perioadă = ouă ÷ zile-găină (D156, D159) · % Cat. 1–3 = din ouăle rândului (D183)',
             block('pc', c3, t3))

# 4 Hrană · Hale: three head rows, combined leaves, grain Săpt. (week rows only, every week closed)
c4 = f'Hrană · {SER} · pe hale'
secs4 = [('Total fermă', '', hrana())] + [(f'Hala {h[1:]}', GOALA if h == 'H7' else '', hrana(h)) for h in R.HHALLS]
t4, _ = ptable('pd', c4, secs4)
html += demo('Hrană · Detaliu Hale · trei rânduri de cap (secțiune › înregistrare › frunză „♀ efectiv”) · toate săptămânile deschise (D167) · tabelul derulează lateral',
             block('pd', c4, t4, sel='Hale'))

# 5 Greutate · H5.1 (section page, no switch): weighings variant + partial guide
c5 = f'Greutate · Hîrbovăț › H5.1 · Seria 2026 · SAPT 56–59'
t5, _ = ptable('pe', c5, [('H5.1', '', greutate())], weighings=True)
after5 = ('<div class="ip-ptable__after"><p class="ip-note">– fără date în ghid: ghidul are greutăți până la SAPT 25. '
          'Cântărire ilustrativă (Hîrbovăț nu are cântăriri reale, D152).</p></div>')
html += demo('Greutate · pagina secțiunii H5.1 (fără Detaliu) · varianta weighings: săptămâna cu cântărire arată ziua ei, celelalte se deschid la un rând „fără cântărire în această săptămână” (D167) · ghid parțial',
             block('pe', c5, t5, bar=False, after=after5))

# 6 Serie (S3): allWeeks, Fermă column, no FilterBar, every week open (D166)
c6 = f'Păsări · Seria 2026 · Total serie · SAPT 56–59'
t6, tid6 = ptable('pf', c6, [('Total serie', '', farm_pasari())], spine='s3')
after6 = (f'<div class="ip-ptable__after"><button type="button" class="ip-btn ip-btn--secondary" aria-controls="{tid6}">Arată toate săptămânile</button>'
          f'<p class="ip-note">apoi: SAPT 59 → 1 și „Arată ultimele 4 săptămâni”</p></div>')
html += demo('Seria 2026 (S3) · allWeeks · coloana Fermă după SAPT · fără FilterBar și fără Detaliu',
             block('pf', c6, t6, bar=False, after=after6))

# 6b S3 at 1280: dense (figure cells space-2 inline, same row height)
t6b, tid6b = ptable('pj', c6, [('Total serie', '', farm_pasari())], spine='s3', dense=True)
after6b = f'<div class="ip-ptable__after"><button type="button" class="ip-btn ip-btn--secondary" aria-controls="{tid6b}">Arată toate săptămânile</button></div>'
html += demo('Seria 2026 (S3) la 1280 · ip-ptable--dense: celulele cu cifre au un pas de spațiere mai puțin (space-2), aceeași înălțime de rând',
             block('pj', c6 + ' · 1280', t6b, bar=False, after=after6b), '960px')

# 7 States: loading · no data · no guide
c7 = f'Păsări · {SER}'
t7, _ = ptable('pg', c7, [('Total fermă', '', farm_pasari())], skel=[4, 4, 3, 2, 4, 4, 4, 3], colw=COLW)
load = (detaliu('pg') + f'<p class="ip-cap ip-ptable__cap" aria-hidden="true">{c7}</p><div aria-busy="true">'
        f'<p class="ip-sr" role="status">Se încarcă tabelul Păsări, SAPT 56–59.</p>{region(t7, c7, hidden=True)}</div>')
html += demo('Se încarcă · același tabel, bare la geometria încărcată (D132), barele rândurilor de săptămână în rule', load)
empty = ('<div class="ip-emptybox" role="region" aria-label="Păsări, interval 01.11–15.11.2025"><div class="ip-empty">'
         '<span class="ip-lamp ip-lamp--off">fără date</span><p class="ip-empty__line">Ciclul Seria 2026 nu are înregistrări între '
         '01.11.2025 și 15.11.2025; a început pe 22.02.2026.</p><button type="button" class="ip-btn ip-btn--secondary">Alege ciclul curent</button></div></div>')
html += demo('Fără date în perioadă · EmptyState în locul regiunii (D48)', empty)
c8 = f'Producție · {SER} · fără ghid'
t8, _ = ptable('ph', c8, [('Total fermă', '', productie(guide=False))])
html += demo('Fără ghid · coloanele ghid și față de ghid lipsesc (D92); antetul paginii poartă Tag-ul <span class="ip-tag ip-tag--neutral">fără ghid</span> · stare ilustrativă pe cifrele reale',
             block('ph', c8, t8))

# 8 Phone 390
c9 = f'Păsări · {SER} · Mortalitate'
pa = farm_pasari()
t9, _ = ptable('pi', c9, [('Total fermă', '', [group(pa, 'În viață'), stacked(group(pa, 'Mortalitate'), group(pa, '% cumulat'))])], spine='phone')
phone = (f'<div class="ip-phone" style="width:390px"><div style="padding:var(--space-4)">'
         + block('pi', c9, t9, touch=True) + '</div></div>')
html += demo('Telefon 390 · o grupă de coloane (eticheta Mortalitate; În viață rămâne; % cumulat pe al doilea rând al celulei) · rânduri 48 · prima coloană lipită „Ma 17.11 · 406”, datele săptămânii sub SAPT', phone, '390px')

html += (f'<p class="ip-note" style="margin:0">Cifre reale Hîrbovăț, Seria 2026, până la {R.dd(R.TODAY)} (extrase DATA03): '
         f'SAPT 59 = {R.drange(*wk_label(59)[:2])} (2 din 7 zile). Mortalitate = morți reali, golirea din 12.11 e Vânzare (D155). '
         f'Greutatea H5.1 e ilustrativă. Rata pe perioadă a ghidului = media pe zile-găină.</p>\n</div>\n')
if __name__ == '__main__':  # S4N: s4t_build imports the row/column builders above; only a direct run writes the preview
    OUT.parent.mkdir(exist_ok=True)
    OUT.write_text(SXS.sexify(html), encoding='utf-8'); print(len(html))
