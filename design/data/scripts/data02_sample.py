"""DATA02 (part 2): ILLUSTRATIVE Seria 2026 sample at Hîrbovăț (2026-03-01 … 2026-11-16)
and counts-latest.json. Deterministic (SEED). Reads only DATA01/DATA02 extracts.
Usage: python3 -I data02_sample.py <extracts_dir>
"""
import sys, os, json, math, random, datetime as dt

EX = sys.argv[1]
SEED = 20261116
R = random.Random(SEED)
D = dt.date
def load(n): return json.load(open(os.path.join(EX, n), encoding='utf-8'))
def iso(d): return d.isoformat()
def dump(name, obj):
    def enc(o, ind=0):
        pad = '  ' * ind
        if isinstance(o, dict):
            if not o: return '{}'
            items = [f'{pad}  {json.dumps(k, ensure_ascii=False)}: {enc(v, ind + 1)}' for k, v in o.items()]
            return '{\n' + ',\n'.join(items) + '\n' + pad + '}'
        if isinstance(o, list) and o and all(isinstance(x, dict) for x in o):
            return '[\n' + ',\n'.join(pad + '  ' + json.dumps(x, ensure_ascii=False) for x in o) + '\n' + pad + ']'
        return json.dumps(o, ensure_ascii=False)
    p = os.path.join(EX, name)
    with open(p, 'w', encoding='utf-8') as f: f.write(enc(obj) + '\n')
    json.load(open(p, encoding='utf-8'))
def poisson(lam):
    L, k, p = math.exp(-lam), 0, 1.0
    while True:
        p *= R.random()
        if p <= L: return k
        k += 1

# ---------- inputs ----------
transfers = load('transfers-seria2026.json')
guide = load('guide-laying.json')['rows']
prevw = load('eggs-hirbovat-prev-weekly.json')['rows']
telita_mort = load('mortality-telita-daily.json')['meta']['totals']
placement = load('placement-seria2026.json')
G = {(r['cycle_week'], r['sex']): r for r in guide}

CYCLE_START, FIRST, TODAY = D(2026, 2, 10), D(2026, 3, 1), D(2026, 11, 16)
def cw(d): return (d - CYCLE_START).days // 7 + 1
def aw(d): return cw(d) + 18
SECS = ['H1.1', 'H1.2', 'H3.1', 'H3.2', 'H4.1', 'H4.2']
HALLS = {'H1': ['H1.1', 'H1.2'], 'H3': ['H3.1', 'H3.2'], 'H4': ['H4.1', 'H4.2']}
START = {a['section']: {'F': a['F'], 'M': a['M']} for a in transfers['arrival_counts']}
assert sum(v['F'] for v in START.values()) == 26765 and sum(v['M'] for v in START.values()) == 3060
DAYS = [FIRST + dt.timedelta(i) for i in range((TODAY - FIRST).days + 1)]
VET, MGR = 'Dr. Ceban', 'Ion Rusu'

# ---------- removals ----------
CULLS = [  # date, section, sex, qty, reason
    (D(2026, 3, 24), 'H4.2', 'F', 4, 'șchiopătare'), (D(2026, 4, 15), 'H1.2', 'M', 2, 'șchiopătare'),
    (D(2026, 5, 6), 'H3.1', 'F', 6, 'boală'), (D(2026, 6, 2), 'H4.1', 'M', 3, 'boală'),
    (D(2026, 6, 23), 'H1.1', 'F', 2, 'altul'), (D(2026, 7, 14), 'H3.2', 'M', 2, 'șchiopătare'),
    (D(2026, 8, 11), 'H4.2', 'F', 5, 'boală'), (D(2026, 9, 1), 'H1.2', 'F', 3, 'șchiopătare'),
    (D(2026, 9, 29), 'H3.2', 'M', 1, 'altul'), (D(2026, 10, 20), 'H4.1', 'F', 4, 'boală'),
    (D(2026, 11, 12), 'H4.2', 'F', 3, 'șchiopătare'), (D(2026, 11, 15), 'H3.1', 'M', 2, 'boală'),
]
CULL_NOTE = {'altul': 'pasăre prinsă în echipament'}
SEXERR = [(D(2026, 3, 18), 'H3.2', 'F', 1, 'cocoș găsit printre găini, scos din efectiv'),
          (D(2026, 4, 9), 'H4.1', 'F', 1, 'cocoș găsit printre găini, scos din efectiv')]
FORCED_MORT = {(D(2026, 11, 14), 'H4.1', 'F'): 5}   # corrected 3 → 5 on 15.11 (audit log)
SEC_MULT = {s: R.uniform(0.85, 1.2) for s in SECS}

alive = {s: dict(START[s]) for s in SECS}
daily = {}  # (date, sec, sex) -> row
for d in DAYS:
    for s in SECS:
        for x in ('F', 'M'):
            a0 = alive[s][x]
            if x == 'F': rate = 0.00018 + 0.00008 * (cw(d) - 3) / 37
            else: rate = 0.00040 + 0.00020 * (cw(d) - 3) / 37
            m = poisson(a0 * rate * SEC_MULT[s])
            if s == 'H1.1' and d <= D(2026, 3, 10): m = 0   # keeps the real 10.03 row (4 500 ♀ · 498 ♂) exact
            if d == TODAY:  # keep today's entry under the D19 mortality alert (3× the 7-day average)
                avg7 = sum(daily[(d - dt.timedelta(k), s, x)]['mortality'] for k in range(1, 8)) / 7
                m = min(m, int(3 * avg7))
            m = FORCED_MORT.get((d, s, x), m)
            c = sum(q for (dd, ss, xx, q, _) in CULLS if (dd, ss, xx) == (d, s, x))
            se = sum(q for (dd, ss, xx, q, _) in SEXERR if (dd, ss, xx) == (d, s, x))
            a1 = a0 - m - c - se
            assert a1 >= 0
            alive[s][x] = a1
            daily[(d, s, x)] = {'date': iso(d), 'cycle_week': cw(d), 'section': s, 'sex': x, 'alive_start': a0,
                                'mortality': m, 'culls': c, 'sexing_errors': se, 'alive_end': a1}

# ---------- feed allocation ----------
FT = {'F': 'ft-breeder2', 'M': 'ft-cocosi'}
MISSING_TODAY = {('H3.2', 'M'), ('H4.2', 'F')}
DEVIATION_TODAY = ('H1.2', 'M', 172)
off = {}
for d in DAYS:
    if d == FIRST or d.weekday() == 1:  # new offsets each cycle week (cycle weeks start on Tuesday)
        for s in SECS:
            off[(s, 'F')] = R.choice([-2, -1, 0, 0, 0, 1, 2])
            off[(s, 'M')] = R.choice([-3, -2, 0, 0, 1, 2, 3])
    for s in SECS:
        for x in ('F', 'M'):
            row = daily[(d, s, x)]
            gg = G[(cw(d), x)]['feed_g']
            g = gg + off[(s, x)]
            note = None
            if d == D(2026, 3, 10) and s == 'H1.1':
                g = 113 if x == 'F' else 98; note = 'real (Hrana Hîrbovăț 10.03.2026: 508,5 kg ♀ / 48,8 kg ♂)'
            if d == TODAY and (s, x) in MISSING_TODAY: g = None
            if d == TODAY and (s, x) == DEVIATION_TODAY[:2]: g = DEVIATION_TODAY[2]
            row.update({'feed_type_id': FT[x], 'g_per_bird': g, 'guide_feed_g': gg,
                        'kg': round(g * row['alive_end'] / 1000, 1) if g is not None else None,
                        'deviation_pct': round(100 * (g - gg) / gg, 1) if g is not None else None})
            if note: row['note'] = note
assert daily[(D(2026, 3, 10), 'H1.1', 'F')]['kg'] == 508.5 and daily[(D(2026, 3, 10), 'H1.1', 'M')]['kg'] == 48.8

cons = {(d, t): 0.0 for d in DAYS for t in FT.values()}
for (d, s, x), row in daily.items():
    if row['kg'] is not None: cons[(d, row['feed_type_id'])] += row['kg']

# ---------- deliveries & stock ----------
deliv = []
def stock_series(t):
    s, out = 0.0, {}
    for d in DAYS:
        s += sum(x['qty_kg'] for x in deliv if x['date'] == d and x['feed_type_id'] == t)
        s -= cons[(d, t)]
        out[d] = s
    return out
def greedy(t, until, size, low_days, rng):
    s = 0.0
    for d in DAYS:
        if d > until: break
        c = cons[(d, t)]
        if s < low_days * c:
            q = int(round(rng.uniform(*size) / 10) * 10)
            deliv.append({'date': d, 'feed_type_id': t, 'qty_kg': q}); s += q
        s -= c
B2, CC = 'ft-breeder2', 'ft-cocosi'
def plan(t, size, low_days, at, lo, hi):
    """greedy deliveries up to `at`; try sub-seeds in order (deterministic) until the stock at the
    end of `at` lies in [lo, hi] and never went below 0"""
    global deliv
    keep = [x for x in deliv if x['feed_type_id'] != t]
    for k in range(200):
        deliv = list(keep)
        greedy(t, at, size, low_days, random.Random(SEED * 1000 + k + (0 if t == B2 else 500)))
        st = stock_series(t)
        if lo <= st[at] <= hi and min(v for d, v in st.items() if d <= at) >= 0:
            return k, round(st[at])
    raise SystemExit('no sub-seed fits for ' + t)
print('B2 plan', plan(B2, (22000, 25000), 1.6, D(2026, 11, 3), 6500, 12500))
print('CC plan', plan(CC, (5500, 8000), 3.2, D(2026, 11, 5), 300, 3000))
def cons_sum(t, a, b): return sum(cons[(d, t)] for d in DAYS if a <= d <= b)
ss = stock_series(B2)
q1 = int(round(7450 - ss[D(2026, 11, 3)] + cons_sum(B2, D(2026, 11, 4), D(2026, 11, 9))))
deliv.append({'date': D(2026, 11, 4), 'feed_type_id': B2, 'qty_kg': q1})
deliv.append({'date': D(2026, 11, 10), 'feed_type_id': B2, 'qty_kg': 24000, 'note': 'exemplul din spec C5'})
deliv.append({'date': D(2026, 11, 14), 'feed_type_id': B2, 'qty_kg': 24000})
sc = stock_series(CC)
qc = int(round(1650 - sc[D(2026, 11, 5)] + cons_sum(CC, D(2026, 11, 6), TODAY)))
deliv.append({'date': D(2026, 11, 6), 'feed_type_id': CC, 'qty_kg': qc})
deliv.sort(key=lambda x: (x['date'], x['feed_type_id']))
UNPRICED_FROM = D(2026, 11, 4)
pw = {B2: 8.95, CC: 8.55}
rows_deliv = []
for x in deliv:
    t = x['feed_type_id']
    pw[t] = min(max(pw[t] + R.uniform(-0.06, 0.07), 8.4 if t == B2 else 8.1), 9.45 if t == B2 else 9.0)
    priced = x['date'] < UNPRICED_FROM
    o = {'id': f"del-hir-{iso(x['date'])}-{'b2' if t == B2 else 'coc'}", 'date': iso(x['date']), 'farm_id': 'hirbovat',
         'feed_type_id': t, 'qty_kg': x['qty_kg'], 'note': x.get('note'), 'created_by': MGR, 'priced': priced,
         'price_per_kg': round(pw[t], 2) if priced else None,
         'total_lei': round(round(pw[t], 2) * x['qty_kg']) if priced else None,
         'priced_by': 'owner' if priced else None,
         'priced_at': iso(x['date'] + dt.timedelta(R.randint(2, 6))) if priced else None}
    rows_deliv.append(o)
ss, sc = stock_series(B2), stock_series(CC)
for t, s_ in ((B2, ss), (CC, sc)):
    assert min(s_.values()) >= 0, (t, min(s_.values()))
feed_by_type = []
for d in DAYS:
    for t, s_ in ((B2, ss), (CC, sc)):
        feed_by_type.append({'date': iso(d), 'feed_type_id': t, 'delivered_kg': sum(x['qty_kg'] for x in deliv if x['date'] == d and x['feed_type_id'] == t),
                             'consumption_kg': round(cons[(d, t)], 1), 'stock_end_kg': round(s_[d], 1)})
def stock_today(t, s_):
    last3 = [cons[(TODAY - dt.timedelta(k), t)] for k in (3, 2, 1)]
    avg = sum(last3) / 3
    st = s_[TODAY]
    dl = st / avg
    return {'farm_id': 'hirbovat', 'feed_type_id': t, 'stock_kg': round(st), 'avg_daily_kg_last3': round(avg),
            'avg_window': ['2026-11-13', '2026-11-15'], 'days_left': round(dl, 1),
            'run_out_date': iso(TODAY + dt.timedelta(int(dl))), 'status': 'ok' if dl >= 5 else 'warn (sub 5 zile, D19)'}
stocks = [stock_today(B2, ss), stock_today(CC, sc)]
b2_before_1011 = ss[D(2026, 11, 9)]
toast_stock = round(b2_before_1011 + 24000)

# ---------- eggs ----------
pw_all = {r['age_week']: r for r in prevw if r['hall'] == 'all'}
pw_h = {(r['age_week'], r['hall']): r for r in prevw}
def prev(aw_, hall=None):
    k = min(max(aw_, 23), 62)
    return pw_h[(k, hall)] if hall else pw_all[k]
HALL_OFF = {'H1': 0.6, 'H3': -0.4, 'H4': 0.3}
ar = {h: 0.0 for h in HALLS}
eggs = []
for d in DAYS:
    c = cw(d); gl = G[(c, 'F')]['lay_pct']
    for h, secs in HALLS.items():
        hens = sum(daily[(d, s, 'F')]['alive_start'] for s in secs)
        ar[h] = max(min(0.85 * ar[h] + R.gauss(0, 0.8), 3.0), -3.0)
        if gl is None:
            base = {3: 0.0, 4: 0.6, 5: 9.0, 6: 42.0}[c]
            lay = max(base + (R.uniform(-0.3, 0.3) if c == 4 else R.uniform(-2, 2) if base else 0), 0)
        elif d >= D(2026, 11, 10):     # cycle wk 40: H3 below guide by > 3 points (alert), H1/H4 above
            lay = gl + {'H1': 1.9, 'H3': -4.0, 'H4': 1.7}[h] + R.uniform(-0.3, 0.3)
        elif d >= D(2026, 11, 3) and h == 'H3':
            lay = gl - 2.0 - 0.15 * (d - D(2026, 11, 3)).days + R.uniform(-0.4, 0.4)
        else:
            lay = gl + HALL_OFF[h] + ar[h]
        total = round(hens * lay / 100)
        if total == 0:
            eggs.append({'date': iso(d), 'cycle_week': c, 'age_week': aw(d), 'hall': h, 'hens_start': hens, 'floor_eggs': 0,
                         'cat1': 0, 'cat2': 0, 'cat3': 0, 'hatching_eggs': 0, 'total_eggs': 0, 'lay_pct': 0.0,
                         'guide_lay_pct': gl, 'hatching_pct': None, 'avg_egg_weight_g': None}); continue
        pa = prev(aw(d)); ph = prev(aw(d), h)
        hp = 0.0 if aw(d) < 23 else min(pa['hatching_pct'] + R.uniform(-0.8, 0.8), 97.5)
        hat = round(total * hp / 100)
        rest = total - hat
        nonh = pa['cat1'] + pa['cat2'] + pa['cat3']
        s1, s2 = (pa['cat1'] / nonh, pa['cat2'] / nonh) if nonh else (0.2, 0.6)
        c1 = round(rest * s1 * R.uniform(0.85, 1.15)); c2 = min(round(rest * s2 * R.uniform(0.9, 1.1)), rest - c1)
        c3 = rest - c1 - c2
        assert min(c1, c2, c3) >= 0
        jr = (ph['floor_pct_of_total'] or 4.0) if aw(d) >= 23 else 30.0
        jos = round(total * jr / 100 * R.uniform(0.85, 1.15))
        ew = (ph['avg_egg_weight_g'] if aw(d) >= 24 else 46.0) + R.gauss(0, 0.3)
        eggs.append({'date': iso(d), 'cycle_week': c, 'age_week': aw(d), 'hall': h, 'hens_start': hens,
                     'floor_eggs': jos, 'cat1': c1, 'cat2': c2, 'cat3': c3, 'hatching_eggs': hat,
                     'total_eggs': total, 'lay_pct': round(100 * total / hens, 1), 'guide_lay_pct': gl,
                     'hatching_pct': round(100 * hat / total, 1), 'avg_egg_weight_g': round(ew, 1)})
eggs_weekly = []
for c in sorted({e['cycle_week'] for e in eggs}):
    for h in list(HALLS) + ['all']:
        rs = [e for e in eggs if e['cycle_week'] == c and (h == 'all' or e['hall'] == h)]
        tot = sum(e['total_eggs'] for e in rs); hens = sum(e['hens_start'] for e in rs)
        hat = sum(e['hatching_eggs'] for e in rs); gl = G[(c, 'F')]['lay_pct']
        lay = round(100 * tot / hens, 1)
        ews = [e['avg_egg_weight_g'] for e in rs if e['avg_egg_weight_g']]
        eggs_weekly.append({'cycle_week': c, 'age_week': c + 18, 'hall': h, 'from': rs[0]['date'], 'to': rs[-1]['date'],
                            'days': len({e['date'] for e in rs}), 'total_eggs': tot, 'floor_eggs': sum(e['floor_eggs'] for e in rs),
                            'hatching_eggs': hat, 'lay_pct': lay, 'guide_lay_pct': gl,
                            'diff_pts': round(lay - gl, 1) if gl is not None else None,
                            'alert_lay': (abs(lay - gl) > 3) if gl is not None else None,
                            'hatching_pct': round(100 * hat / tot, 1) if tot else None,
                            'avg_egg_weight_g': round(sum(ews) / len(ews), 1) if ews else None})

# ---------- weighings (Mondays = last day of each cycle week) ----------
def wcurve(x, a):
    pts = {'F': [(21, 2650), (25, 3080), (30, 3380), (40, 3640), (50, 3820), (58, 3950)],
           'M': [(21, 3520), (25, 3900), (30, 4180), (40, 4450), (50, 4620), (58, 4760)]}[x]
    for (a0, w0), (a1, w1) in zip(pts, pts[1:]):
        if a0 <= a <= a1: return w0 + (w1 - w0) * (a - a0) / (a1 - a0)
    return pts[-1][1]
SEC_W = {(s, x): R.uniform(-40, 40) for s in SECS for x in 'FM'}
weighings = []
for d in DAYS:
    if d.weekday() != 0: continue
    for s in SECS:
        for x in ('F', 'M'):
            age = aw(d) + 6 / 7
            w = wcurve(x, age) + SEC_W[(s, x)] + R.gauss(0, 25)
            cv = R.uniform(6.8, 9.2) if x == 'F' else R.uniform(7.8, 10.8)
            un = max(min(round(100 - cv * 2.45 + R.uniform(-3, 3), 1), 92), 60)
            weighings.append({'date': iso(d), 'cycle_week': cw(d), 'section': s, 'sex': x, 'avg_weight_g': int(round(w)),
                              'cv_pct': round(cv, 1), 'uniformity_pct': un,
                              'sample_size': R.randint(50, 100) if x == 'F' else R.randint(50, 60),
                              'guide_weight_g': None, 'guide_state': 'fără date în ghid', 'created_by': VET})

# ---------- 16.11 allocation state ----------
alloc_today = []
for s in SECS:
    for x in ('F', 'M'):
        t, y = daily[(TODAY, s, x)], daily[(TODAY - dt.timedelta(1), s, x)]
        st = 'missing' if t['g_per_bird'] is None else ('confirmed · abatere > 10%' if abs(t['deviation_pct']) > 10 else 'confirmed')
        alloc_today.append({'section': s, 'sex': x, 'alive': t['alive_end'], 'feed_type_id': t['feed_type_id'],
                            'yesterday_g': y['g_per_bird'], 'today_g': t['g_per_bird'], 'guide_g': t['guide_feed_g'],
                            'kg': t['kg'], 'deviation_pct': t['deviation_pct'], 'state': st})

# ---------- event log (10.11 – 16.11) & audit ----------
log, audit = [], []
def fmt(v): return f'{int(round(v)):,}'.replace(',', '\u202f')  # spec: thin no-break space
def signed(v): return f'−{v}' if v else '0'
def day_tot(d, x): return sum(daily[(d, s, x)]['mortality'] + daily[(d, s, x)]['culls'] + daily[(d, s, x)]['sexing_errors'] for s in SECS)
for i in range(7):
    d = D(2026, 11, 10) + dt.timedelta(i)
    hm = R.randint(0, 25)
    log.append({'at': f'{iso(d)}T08:{20 + hm:02d}', 'user': VET, 'role': 'vet', 'record': 'mortality', 'action': 'create',
                'scope': 'Hîrbovăț · 6 secțiuni', 'summary': f'Zilnic salvat · 6 secțiuni · ♀ {signed(day_tot(d, "F"))} · ♂ {signed(day_tot(d, "M"))}'})
    for (dd, s, x, q, why) in CULLS:
        if dd == d:
            log.append({'at': f'{iso(d)}T08:{20 + hm:02d}', 'user': VET, 'role': 'vet', 'record': 'cull', 'action': 'create',
                        'scope': f'Hîrbovăț · {s} · {"♀" if x == "F" else "♂"}', 'summary': f'Sacrificare {q} · motiv: {why}'})
    te = sum(e['total_eggs'] for e in eggs if e['date'] == iso(d))
    log.append({'at': f'{iso(d)}T08:{21 + hm:02d}', 'user': VET, 'role': 'vet', 'record': 'eggs', 'action': 'create',
                'scope': 'Hîrbovăț · H1, H3, H4', 'summary': f'Ouă · {fmt(te)} total (Cat. 1–3 + Incubabile)'})
    n_ok = sum(1 for s in SECS for x in 'FM' if daily[(d, s, x)]['g_per_bird'] is not None)
    log.append({'at': f'{iso(d)}T09:{R.randint(5, 50):02d}', 'user': VET, 'role': 'vet', 'record': 'feed', 'action': 'create',
                'scope': 'Hîrbovăț', 'summary': f'Hrană zilnică · {n_ok} din 12 rânduri · {fmt(cons[(d, B2)] + cons[(d, CC)])} kg'})
    for x in deliv:
        if x['date'] == d:
            nm = 'Breeder 2' if x['feed_type_id'] == B2 else 'Cocoși'
            log.append({'at': f'{iso(d)}T11:{R.randint(0, 59):02d}', 'user': MGR, 'role': 'manager', 'record': 'feed_delivery',
                        'action': 'create', 'scope': 'Hîrbovăț', 'summary': f'Livrare {fmt(x["qty_kg"])} kg {nm} · fără preț'})
    if d.weekday() == 0:
        log.append({'at': f'{iso(d)}T10:{R.randint(0, 40):02d}', 'user': VET, 'role': 'vet', 'record': 'weighing', 'action': 'create',
                    'scope': 'Hîrbovăț · 12 rânduri', 'summary': 'Cântărire · 6 secțiuni · ♀ și ♂'})
corr = {'at': '2026-11-15T10:12', 'user': MGR, 'role': 'manager', 'record': 'mortality', 'action': 'update',
        'scope': 'Hîrbovăț · H4.1 · ♀ · 14.11.2026', 'summary': 'Mortalitate 3 → 5',
        'reason': 'Numărătoare refăcută: 2 găini găsite moarte sub hrănitoare.'}
log.append(corr); log.sort(key=lambda e: e['at'])
audit.append({'table': 'BIRD_EVENT', 'row': 'mortality · H4.1 · F · 2026-11-14', 'action': 'update', 'old_values': {'qty': -3},
              'new_values': {'qty': -5}, 'reason': corr['reason'], 'user': MGR, 'at': '2026-11-15T10:12'})

# ---------- summaries & checks ----------
def sec_summary(s, x):
    rs = [daily[(d, s, x)] for d in DAYS]
    m, c, se = (sum(r[k] for r in rs) for k in ('mortality', 'culls', 'sexing_errors'))
    st = START[s][x]; end = rs[-1]['alive_end']
    assert st - m - c - se == end and end >= 0
    return {'start': st, 'mortality': m, 'culls': c, 'sexing_errors': se, 'alive': end,
            'mortality_pct': round(100 * m / st, 2), 'removals_pct': round(100 * (m + c + se) / st, 2)}
secsum = {s: {x: sec_summary(s, x) for x in 'FM'} for s in SECS}
def farm_sum(x):
    o = {k: sum(secsum[s][x][k] for s in SECS) for k in ('start', 'mortality', 'culls', 'sexing_errors', 'alive')}
    o['mortality_pct'] = round(100 * o['mortality'] / o['start'], 2)
    o['removals_pct'] = round(100 * (o['mortality'] + o['culls'] + o['sexing_errors']) / o['start'], 2)
    return o
farm = {x: farm_sum(x) for x in 'FM'}
assert all(r['alive_end'] >= 0 and r['alive_start'] >= 0 for r in daily.values())
today_eggs = [e for e in eggs if e['date'] == iso(TODAY)]
lay_today = round(100 * sum(e['total_eggs'] for e in today_eggs) / sum(e['hens_start'] for e in today_eggs), 1)
wk40 = {e['hall']: e for e in eggs_weekly if e['cycle_week'] == 40}
mort_alerts = []
for s in SECS:
    for x in 'FM':
        prev7 = [daily[(TODAY - dt.timedelta(k), s, x)]['mortality'] for k in range(1, 8)]
        avg = sum(prev7) / 7
        if daily[(TODAY, s, x)]['mortality'] > 3 * avg and daily[(TODAY, s, x)]['mortality'] > 0:
            mort_alerts.append({'section': s, 'sex': x, 'today': daily[(TODAY, s, x)]['mortality'], 'avg7': round(avg, 2)})
unpriced_hir = [x for x in rows_deliv if not x['priced']]
tel_deliv = load('feed-deliveries.json')['rows']
unpriced_total = len(unpriced_hir) + sum(1 for x in tel_deliv if not x['priced'])
summary = {
    'date': '2026-11-16', 'cycle_week': cw(TODAY), 'age_week': aw(TODAY),
    'alive': {'F': farm['F']['alive'], 'M': farm['M']['alive'], 'total': farm['F']['alive'] + farm['M']['alive']},
    'lay_pct_today': lay_today, 'guide_lay_pct': G[(40, 'F')]['lay_pct'],
    'lay_pct_week40_by_hall': {h: {'lay_pct': v['lay_pct'], 'diff_pts': v['diff_pts'], 'alert': v['alert_lay']} for h, v in wk40.items()},
    'cumulative_mortality_pct': {'F': farm['F']['mortality_pct'], 'M': farm['M']['mortality_pct']},
    'cumulative_removals_pct': {'F': farm['F']['removals_pct'], 'M': farm['M']['removals_pct']},
    'stocks': stocks, 'unpriced_deliveries_all_farms': unpriced_total,
    'unpriced_kg': sum(x['qty_kg'] for x in unpriced_hir),
    'feed_rows_missing_today': sorted(f'{s} {x}' for s, x in MISSING_TODAY),
    'feed_deviation_today': {'section': DEVIATION_TODAY[0], 'sex': DEVIATION_TODAY[1], 'g': DEVIATION_TODAY[2],
                             'guide_g': G[(40, 'M')]['feed_g'], 'deviation_pct': daily[(TODAY, 'H1.2', 'M')]['deviation_pct']},
    'mortality_alerts_today': mort_alerts,
    'spec_examples': {
        'C5_toast_stock_after_10_11_delivery_kg': toast_stock,
        'C2_warning_minus_1240_true_today': False,
        'C2_note': f'Breeder 2 stock on 16.11 is {round(ss[TODAY])} kg; entering the 2 missing rows uses < 700 kg, so it cannot reach −1 240 kg.'},
}
print(json.dumps(summary, ensure_ascii=False, indent=1))
print('deliveries', len(rows_deliv), 'B2', sum(1 for x in rows_deliv if x['feed_type_id'] == B2), 'CC', sum(1 for x in rows_deliv if x['feed_type_id'] == CC),
      'unpriced', [(x['date'], x['feed_type_id'], x['qty_kg']) for x in unpriced_hir])
print('B2 stock end 09.11', round(b2_before_1011, 1), 'min stocks', round(min(ss.values())), round(min(sc.values())))
print('farm', farm)
print('lay weekly sample', [(e['cycle_week'], e['lay_pct'], e['guide_lay_pct']) for e in eggs_weekly if e['hall'] == 'all'][::4])
assert summary['unpriced_deliveries_all_farms'] == 4
assert stocks[0]['days_left'] > 5 and stocks[1]['days_left'] < 5
assert wk40['H3']['diff_pts'] < -3 and not wk40['H1']['alert_lay'] and not wk40['H4']['alert_lay']

META = {'generated_by': 'DATA02', 'script': '/home/claude/scripts/data02_sample.py <extracts_dir>', 'seed': SEED,
        'illustrative': True,
        'label': 'SAMPLE / ILLUSTRATIVE — not workbook data (D27). The workbook ends 28.02.2026 (feed 10.03.2026).',
        'range': ['2026-03-01', '2026-11-16'], 'farm_id': 'hirbovat', 'series_id': 'seria-2026', 'cycle_id': 'cyc-hirbovat-2026',
        'cycle_start': '2026-02-10', 'week_basis': 'cycle_week = floor((date − 2026-02-10) / 7) + 1; age_week = cycle_week + 18 (D26)',
        'real_inputs': ['start counts = Hîrbovăț arrivals (Sheet9, ♀ 26 765 · ♂ 3 060); nothing recorded at Hîrbovăț to 28.02.2026',
                        'guide feed_g ♀/♂ and lay_pct from guide-laying.json (lay_pct itself illustrative, D29)',
                        'H1.1 grams on 10.03.2026 (♀ 113 g · ♂ 98 g = 508,5 / 48,8 kg) are the one real Hrana Hîrbovăț row; H1.1 has 0 deaths 01–10.03 so the kg match',
                        'egg categories, hatching share, floor share and egg weight are shaped on the previous flock at the same age week (eggs-hirbovat-prev-weekly.json)'],
        'definitions': {
            'alive_start / alive_end': 'birds before / after that day\'s events; alive_end = alive_start − mortality − culls − sexing_errors',
            'kg': 'g_per_bird × alive_end / 1000 (data-model: birds alive up to that date); null g = missing (not carried over)',
            'lay_pct': 'total_eggs (Cat. 1–3 + Incubabile, D18; Jos excluded) ÷ hens alive at the start of the day × 100',
            'hatching_pct': 'hatching_eggs ÷ total_eggs × 100 (no guide, D29)',
            'stock': 'Σ deliveries − Σ consumption per feed type since 01.03.2026 (no Hîrbovăț stock before); days_left = stock on 16.11 ÷ average consumption 13–15.11 (16.11 still being entered)',
            'sexing_error': 'a bird of the other sex found in the count, removed from the recorded sex (qty −1), as in the Telița sheet',
            'cull': 'Sacrificare (spec C1) with reason; BIRD_EVENT has no "cull" type yet (data-model gap, setup-check §1)'},
        'designed_states': [
            'H3 lay in cycle wk 40 below guide by > 3 points (alert); H1, H4 within ±3',
            '16.11 feed: H3.2 ♂ and H4.2 ♀ missing; H1.2 ♂ 172 g vs guide 155 g (+11 %, warning); 15.11 complete (suggestions)',
            'Cocoși stock under 5 days (warn); Breeder 2 above 5 days',
            '4 unpriced deliveries (all at Hîrbovăț, 04.11–14.11); every earlier delivery and every Telița delivery is priced',
            '10.11 "24 000 kg Breeder 2" (spec C5): stock after it = 31 450 kg Breeder 2, as in the C5 toast',
            'spec C2 "−1 240 kg" warning is a state example only; not true today',
            'weighings: no laying guide for weight/CV/uniformity → "fără date în ghid"',
            'correction 15.11 by the manager with a reason (mortality H4.1 ♀ 14.11: 3 → 5)'],
        'people': 'Dr. Ceban (vet) and Ion Rusu (manager) are illustrative names; "owner" prices deliveries'}

dump('sample-hirbovat-2026.json', {
    'meta': META, 'summary_today': summary,
    'start_counts': [{'section': s, 'F': START[s]['F'], 'M': START[s]['M']} for s in SECS],
    'section_totals': [dict(section=s, sex=x, **secsum[s][x]) for s in SECS for x in 'FM'],
    'culls': [{'date': iso(d), 'section': s, 'sex': x, 'qty': q, 'reason': why, 'note': CULL_NOTE.get(why), 'created_by': VET} for d, s, x, q, why in CULLS],
    'sexing_errors': [{'date': iso(d), 'section': s, 'sex': x, 'qty': q, 'note': nt, 'created_by': VET} for d, s, x, q, nt in SEXERR],
    'eggs_weekly': eggs_weekly,
    'deliveries': rows_deliv,
    'stock_today': stocks,
    'allocations_today': alloc_today,
    'event_log': log,
    'audit_log': audit,
    'weighings': weighings,
    'feed_by_type_daily': feed_by_type,
    'eggs_daily': eggs,
    'daily': sorted(daily.values(), key=lambda r: (r['date'], r['section'], r['sex'])),
})

# ---------- counts-latest.json ----------
tel_secs = ['H1', 'H2', 'H3', 'H6', 'H5-1', 'H5-2']
plc = {}
for r in placement.get('rows', []):
    plc[r['section']] = plc.get(r['section'], 0) + r['qty']
dump('counts-latest.json', {
    'meta': {'generated_by': 'DATA02', 'script': '/home/claude/scripts/data02_sample.py <extracts_dir>', 'as_of': '2026-11-16',
             'illustrative': True,
             'label': 'Hîrbovăț counts on 16.11.2026 are SAMPLE (from sample-hirbovat-2026.json, D27); the 28.02.2026 baseline and Telița figures are real.',
             'definitions': 'alive = start − mortality − culls − sexing_errors (end of 16.11.2026, after the morning entry); mortality_pct and removals_pct are of the cycle start count',
             'telita_note': 'Telița cycle closed 28.02.2026 (D25); workbook residuals (23/16/−9/0/454/244) are data errors, not shown (D35).'},
    'farms': [
        {'farm_id': 'hirbovat', 'cycle_id': 'cyc-hirbovat-2026', 'cycle_status': 'open', 'cycle_week': 40, 'illustrative': True,
         'sections': [{'section': s, 'F': secsum[s]['F']['alive'], 'M': secsum[s]['M']['alive'],
                       'detail': secsum[s]} for s in SECS],
         'halls': [{'hall': h, 'F': sum(secsum[s]['F']['alive'] for s in v), 'M': sum(secsum[s]['M']['alive'] for s in v)} for h, v in HALLS.items()],
         'totals': {'F': farm['F']['alive'], 'M': farm['M']['alive'], 'all': farm['F']['alive'] + farm['M']['alive']},
         'cycle': {'start_counts': {'F': farm['F']['start'], 'M': farm['M']['start']},
                   'removals': {x: {k: farm[x][k] for k in ('mortality', 'culls', 'sexing_errors')} for x in 'FM'},
                   'mortality_pct': {'F': farm['F']['mortality_pct'], 'M': farm['M']['mortality_pct']},
                   'removals_pct': {'F': farm['F']['removals_pct'], 'M': farm['M']['removals_pct']}},
         'baseline_real_2026_02_28': {'F': 26765, 'M': 3060, 'source': 'Sheet9 arrivals'}},
        {'farm_id': 'telita', 'cycle_id': 'cyc-telita-2026', 'cycle_status': 'closed', 'cycle_end': '2026-02-28', 'illustrative': False,
         'sections': [{'section': s, 'F': 0, 'M': 0} for s in tel_secs],
         'totals': {'F': 0, 'M': 0, 'all': 0},
         'cycle': {'start_counts': {'F': 28080, 'M': 3510}, 'mortality': {'F': telita_mort['mortality_F'], 'M': telita_mort['mortality_M']},
                   'mortality_pct': {'F': round(100 * telita_mort['mortality_F'] / 28080, 2), 'M': round(100 * telita_mort['mortality_M'] / 3510, 2)},
                   'source': 'mortality-telita-daily.json (real)'}}],
    'series': [{'series_id': 'seria-2026', 'F': farm['F']['alive'], 'M': farm['M']['alive'], 'all': farm['F']['alive'] + farm['M']['alive'],
                'by_farm': {'hirbovat': farm['F']['alive'] + farm['M']['alive'], 'telita': 0}, 'illustrative': True}],
})
print('OK')
