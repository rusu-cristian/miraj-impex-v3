"""DATA03: rebuild the Hîrbovăț / current extracts from Raport_NOU_2026.xlsx (real data to 18.11.2026).
Usage: python3 -I data03_build.py <Raport_NOU_2026.xlsx> <extracts_dir>
Run after data01_build.py, data02_prev.py and d36_patch.py (it reads their Telița-side outputs:
placement-seria2026.json, transfers-seria2026.json Telița part, mortality-telita-daily.json,
feed-daily-telita.json, feed-deliveries.json Telița rows, guide-laying.json wk 19 feed, guide-growth.json).
Week rule D151: age_days = date − placement (07.10.2025, age 0) + age_at_placement_days (0);
sapt = floor(age_days / 7) + 1. Prints a check log; writes the JSON files and key-figures-18-11.md.
"""
import sys, json, os, datetime
import openpyxl
from openpyxl.utils import column_index_from_string as ci

WB, OUT = sys.argv[1], sys.argv[2]
wb = openpyxl.load_workbook(WB, read_only=True, data_only=True)
LOG = []
def log(*a):
    s = ' '.join(str(x) for x in a); LOG.append(s); print(s)

PLACEMENT = datetime.date(2025, 10, 7)
AGE_AT_PLACEMENT = 0
TODAY = '2026-11-18'
FIRST = '2026-02-17'
SRC = 'design/data/Raport_NOU_2026.xlsx'
GEN = {'generated_by': 'DATA03', 'script': 'design/data/scripts/data03_build.py <Raport_NOU_2026.xlsx> <extracts_dir>',
       'source': SRC, 'last_real_date': TODAY,
       'week_rule': 'D151: age_days = date − 2025-10-07 (+ age_at_placement_days 0); sapt = floor(age_days / 7) + 1'}

def iso(d): return d.isoformat()
def todate(v):
    if isinstance(v, datetime.datetime): return v.date()
    if isinstance(v, datetime.date): return v
    return None
def num(v):
    if v is None: return None
    if isinstance(v, str):
        v = v.strip().replace(',', '.')
        if v == '': return None
        try: v = float(v)
        except ValueError: return None
    if isinstance(v, bool): return None
    if isinstance(v, (int, float)):
        f = float(v)
        return int(f) if f.is_integer() else round(f, 4)
    return None
def z(v): return 0 if v is None else v
def age_days(ds): return (datetime.date.fromisoformat(ds) - PLACEMENT).days + AGE_AT_PLACEMENT
def sapt(ds): return age_days(ds) // 7 + 1
def r1(x): return None if x is None else round(x, 1)
def r2(x): return None if x is None else round(x, 2)
def pct(a, b, nd=2): return None if not b else round(100.0 * a / b, nd)

def dump(name, obj, compact_lists=()):
    """JSON with one row per line inside lists of dicts / lists of lists."""
    def enc(o, ind=0):
        pad = '  ' * ind
        if isinstance(o, dict):
            if not o: return '{}'
            items = [f'{pad}  {json.dumps(k, ensure_ascii=False)}: {enc(v, ind + 1)}' for k, v in o.items()]
            return '{\n' + ',\n'.join(items) + '\n' + pad + '}'
        if isinstance(o, list) and o and all(isinstance(x, (dict, list)) for x in o):
            return '[\n' + ',\n'.join(pad + '  ' + json.dumps(x, ensure_ascii=False) for x in o) + '\n' + pad + ']'
        return json.dumps(o, ensure_ascii=False)
    p = os.path.join(OUT, name)
    with open(p, 'w', encoding='utf-8') as f:
        f.write(enc(obj) + '\n')
    json.load(open(p, encoding='utf-8'))
    log(f'wrote {name} ({os.path.getsize(p)} B)')

def load(name): return json.load(open(os.path.join(OUT, name), encoding='utf-8'))

def sheet_rows(name, max_col):
    ws = wb[name]
    return [list(r) for r in ws.iter_rows(min_row=1, max_col=max_col, values_only=True)]

def by_date_and_totals(rows):
    """daily[iso] = row; totals[sapt] = Total row (sapt of the last daily row above it)."""
    daily, totals, last = {}, {}, None
    for r in rows[2:]:
        a = r[0]
        d = todate(a)
        if d:
            daily[iso(d)] = r; last = iso(d)
        elif isinstance(a, str) and a.strip().lower() == 'total' and last:
            totals[sapt(last)] = r
    return daily, totals

def C(row, col): return num(row[ci(col) - 1])

SECTIONS = ['H5.1', 'H5.2', 'H6.1', 'H6.2', 'H7.1', 'H7.2']
HALLS = ['H5', 'H6', 'H7']
HALL_OF = {s: s.split('.')[0] for s in SECTIONS}
OLD_CODE = {'H5': 'H1', 'H6': 'H3', 'H7': 'H4', 'H5.1': 'H1.1', 'H5.2': 'H1.2', 'H6.1': 'H3.1',
            'H6.2': 'H3.2', 'H7.1': 'H4.1', 'H7.2': 'H4.2'}
# Hirbovat: per section [♂ mort, ♂ nr, ♀ mort, ♀ nr]; per hall [Ouă, %, inc, %, jos, %]
HSEC = {'H5.1': ('O', 'P', 'Q', 'R'), 'H5.2': ('S', 'T', 'U', 'V'), 'H6.1': ('AC', 'AD', 'AE', 'AF'),
        'H6.2': ('AG', 'AH', 'AI', 'AJ'), 'H7.1': ('AQ', 'AR', 'AS', 'AT'), 'H7.2': ('AU', 'AV', 'AW', 'AX')}
HHALL = {'H5': ('W', 'X', 'Y', 'Z', 'AA', 'AB'), 'H6': ('AK', 'AL', 'AM', 'AN', 'AO', 'AP'),
         'H7': ('AY', 'AZ', 'BA', 'BB', 'BC', 'BD')}
HFARM = dict(mM='D', aM='E', mF='F', aF='G', eggs='H', lay='I', inc='J', incp='K', jos='L', josp='M')
# Hrana Hirbovat
NORM = {'H5.1': ('F', 'G'), 'H5.2': ('H', 'I'), 'H6.1': ('J', 'K'), 'H6.2': ('L', 'M'), 'H7.1': ('N', 'O'), 'H7.2': ('P', 'Q')}
KG = {'H5.1': ('T', 'U'), 'H5.2': ('V', 'W'), 'H6.1': ('X', 'Y'), 'H6.2': ('Z', 'AA'), 'H7.1': ('AB', 'AC'), 'H7.2': ('AD', 'AE')}
MASA_HALL = {'H5': ('I', 'M'), 'H6': ('N', 'R'), 'H7': ('S', 'W')}
PCT_HALL = {'H5': ('G', 'H', 'I'), 'H6': ('J', 'K', 'L'), 'H7': ('M', 'N', 'O')}

H_rows = sheet_rows('Hirbovat', 56)
F_rows = sheet_rows('Hrana Hirbovat', 38)
M_rows = sheet_rows('Masa', 30)
P_rows = sheet_rows('%%', 15)
G_rows = sheet_rows('RealGhid', 35)
PR_rows = sheet_rows('Productivitate', 13)
Hd, Ht = by_date_and_totals(H_rows)
Fd, _ = by_date_and_totals(F_rows)
Md, Mt = by_date_and_totals(M_rows)
Pd, Pt = by_date_and_totals(P_rows)

# ---------------------------------------------------------------- age check (sheet Zile/Sapt vs D151)
bad = [d for d, r in Hd.items() if num(r[1]) != age_days(d) + 1 or num(r[2]) != sapt(d)]
log('age check Hirbovat: Zile = age_days+1 and Sapt = D151 on all', len(Hd), 'rows; mismatches', len(bad))
badf = [d for d, r in Fd.items() if num(r[2]) != sapt(d)]
log('age check Hrana Hirbovat Sapt mismatches', len(badf))

dates = []
d = datetime.date.fromisoformat(FIRST)
while iso(d) <= TODAY:
    dates.append(iso(d)); d += datetime.timedelta(days=1)
assert all(x in Hd for x in dates), 'missing Hirbovat dates'

# ---------------------------------------------------------------- guides
feed_guide = {}  # sapt -> (F, M) from Hrana D/E (weekly rows; whole template)
for ds, r in sorted(Fd.items()):
    gF, gM = C(r, 'D'), C(r, 'E')
    if gF is not None or gM is not None:
        w = sapt(ds)
        if w not in feed_guide: feed_guide[w] = (gF, gM)
real_ghid = {}
for r in G_rows[2:44]:
    a = num(r[0])
    if a is None or a < 19: continue
    real_ghid[int(a)] = dict(lay=num(r[1]), mort=num(r[2]), egg_cum=num(r[3]), inc_cum=num(r[4]),
                             egg_wk=num(r[5]), inc_wk=num(r[6]))
typo_53 = real_ghid.get(53, {}).get('lay')
if typo_53 == 605:
    real_ghid[53]['lay'] = 60.5
    log('RealGhid wk 53 lay "605" read as 60.5 (typo)')
plan = {}
for r in G_rows[3:44]:
    k = num(r[ci('K') - 1])
    if k is None: continue
    plan[int(k)] = dict(date=iso(todate(r[ci('J') - 1])) if todate(r[ci('J') - 1]) else None,
                        hens_plan=C(r, 'L'), eggs_plan=C(r, 'M'), eggs_real=C(r, 'N'), inc_plan=C(r, 'S'), inc_real=C(r, 'T'))
hatch_guide = {}
for r in G_rows[3:46]:
    w, g = C(r, 'AF'), C(r, 'AG')
    if w is not None and g is not None: hatch_guide[int(w)] = g
prod = {}
for r in PR_rows[2:50]:
    a = num(r[0])
    if a is None: continue
    a = int(a)
    if a in prod: continue  # duplicated template labels 64–66
    prod[a] = dict(hen_days=num(r[1]), avg_hens=num(r[2]), eggs=num(r[3]), eggs_cum=num(r[4]), inc=num(r[5]),
                   inc_cum=num(r[6]), lay_real=num(r[7]), lay_guide=num(r[8]), inc_hen_cum=num(r[10]), inc_hen_cum_guide=num(r[11]))
lay_disagree = [(w, real_ghid[w]['lay'], prod[w]['lay_guide']) for w in sorted(real_ghid)
                if w in prod and prod[w]['lay_guide'] is not None and real_ghid[w]['lay'] is not None
                and abs(real_ghid[w]['lay'] - prod[w]['lay_guide']) >= 1.0]
log('lay guide RealGhid B vs Productivitate I differ by ≥1 pt in', len(lay_disagree), 'weeks, e.g.', lay_disagree[-6:])

# ---------------------------------------------------------------- daily
sec_rows, hall_rows, farm_rows = [], [], []
alive_prev = {(s, x): 0 for s in SECTIONS for x in 'FM'}
checks = dict(alive_continuity=[], farm_vs_sections=[], hall_eggs_vs_farm=[], lay_pct=[], kg_vs_norm=[], stock=[],
              pct_sheet=[], total_rows=[])
state = {}  # per date: dict
stock_prev = None
deliv_rows = []  # raw (date, sex, kg, rowidx) from Hrana
lei_cells = []
notes_cells = []
F_index = {iso(todate(r[0])): i for i, r in enumerate(F_rows) if todate(r[0])}
max_normal_mort = {}
internal_moves = []
for ds in dates:
    r = Hd[ds]
    fr = Fd.get(ds)
    mr = Md.get(ds)
    pr = Pd.get(ds)
    age, w = age_days(ds), sapt(ds)
    day = dict(sec={}, hall={}, farm={})
    # sections: raw sheet cell (negative = inflow); inflows after the arrival week are moves between compartments
    raws = {}
    for s in SECTIONS:
        cmM, caM, cmF, caF = HSEC[s]
        raws[(s, 'F')] = (z(C(r, cmF)), C(r, caF)); raws[(s, 'M')] = (z(C(r, cmM)), C(r, caM))
    mv = {k: 0 for k in raws}
    if ds > '2026-03-01':
        for (s, sex), (raw, _) in raws.items():
            if raw < 0:
                q = -raw
                src = max(((raws[(t, sex)][0], t) for t in SECTIONS if t != s and raws[(t, sex)][0] >= q), default=None)
                if src:
                    mv[(s, sex)] += q; mv[(src[1], sex)] -= q
                    internal_moves.append(dict(date=ds, sapt=w, **{'from': src[1]}, to=s, sex=sex, qty=q, inferred=True,
                                               note=f'sheet: {s} {"♀" if sex == "F" else "♂"} mortality −{q} (inflow), {src[1]} mortality {src[0]} the same day; read as {q} moved + {src[0] - q} dead'))
    for s in SECTIONS:
        for sex in ('F', 'M'):
            raw, alive = raws[(s, sex)]
            m_ = raw + mv[(s, sex)]          # deaths (+) / arrivals (−) after removing moves
            arr = -m_ if m_ < 0 else 0
            mort = m_ if m_ > 0 else 0
            if alive is None: alive = alive_prev[(s, sex)] - raw
            exp = alive_prev[(s, sex)] - raw
            if alive != exp:
                checks['alive_continuity'].append((ds, s, sex, alive_prev[(s, sex)], raw, alive))
            g = kg = None
            if fr:
                g = C(fr, NORM[s][0 if sex == 'F' else 1]); kg = C(fr, KG[s][0 if sex == 'F' else 1])
                start = alive_prev[(s, sex)]
                if g is not None and kg is not None and abs(g * start / 1000 - kg) > 0.15:
                    checks['kg_vs_norm'].append((ds, s, sex, g, start, kg))
            sold = mort if ds >= '2026-11-12' and mort >= 40 else 0   # D155: emptying removals = Vânzare
            mort -= sold
            if '2026-03-02' < ds < '2026-11-12':
                k = (s, sex); max_normal_mort[k] = max(max_normal_mort.get(k, (0, '')), (mort, ds))
            start = alive_prev[(s, sex)]
            day['sec'][(s, sex)] = dict(mort=mort, sold=sold, arr=arr, mv=mv[(s, sex)], raw=raw, alive=alive, start=start, g=g, kg=kg)
            if start or alive or mort or sold or arr or (kg or 0) > 0:
                sec_rows.append([ds, age, w, s, sex, mort, sold, arr, mv[(s, sex)], alive, g, kg, raw])
            alive_prev[(s, sex)] = alive
    # halls
    for h in HALLS:
        ce, cl, cinc, cincp, cj, cjp = HHALL[h]
        secs = [s for s in SECTIONS if HALL_OF[s] == h]
        aF = sum(day['sec'][(s, 'F')]['alive'] for s in secs); aM = sum(day['sec'][(s, 'M')]['alive'] for s in secs)
        sF = sum(day['sec'][(s, 'F')]['start'] for s in secs); sM = sum(day['sec'][(s, 'M')]['start'] for s in secs)
        mF = sum(day['sec'][(s, 'F')]['mort'] for s in secs); mM = sum(day['sec'][(s, 'M')]['mort'] for s in secs)
        soF = sum(day['sec'][(s, 'F')]['sold'] for s in secs); soM = sum(day['sec'][(s, 'M')]['sold'] for s in secs)
        eggs, lay, inc, incp, jos, josp = (C(r, c) for c in (ce, cl, cinc, cincp, cj, cjp))
        ew = C(mr, MASA_HALL[h][0]) if mr else None
        if ew == 0: ew = None
        if eggs is not None and aF and lay is not None and abs(100 * eggs / aF - lay) > 0.05:
            checks['lay_pct'].append((ds, h, eggs, aF, lay))
        if pr and lay is not None:
            pl = C(pr, PCT_HALL[h][0])
            if pl is not None and abs(pl - lay) > 0.01: checks['pct_sheet'].append((ds, h, lay, pl))
        kgF = sum(z(day['sec'][(s, 'F')]['kg']) for s in secs); kgM = sum(z(day['sec'][(s, 'M')]['kg']) for s in secs)
        lay_d = pct(eggs, sF) if eggs is not None else None   # D156: eggs ÷ hens at the start of the day
        day['hall'][h] = dict(aF=aF, aM=aM, sF=sF, sM=sM, mF=mF, mM=mM, soF=soF, soM=soM, eggs=eggs, inc=inc, jos=jos,
                              lay=lay_d, lay_sheet=lay, incp=incp, josp=josp, ew=ew, kgF=kgF, kgM=kgM)
        if aF or aM or sF or sM or eggs:
            hall_rows.append([ds, age, w, h, aF, aM, mF, mM, soF, soM, eggs, inc, jos, lay_d, r2(lay), r2(incp), r2(josp), r2(ew),
                              round(kgF, 1), round(kgM, 1)])
    # farm
    f = {k: C(r, c) for k, c in HFARM.items()}
    sums = dict(mF=sum(v['raw'] for (s, x), v in day['sec'].items() if x == 'F'),
                mM=sum(v['raw'] for (s, x), v in day['sec'].items() if x == 'M'),
                aF=sum(v['alive'] for (s, x), v in day['sec'].items() if x == 'F'),
                aM=sum(v['alive'] for (s, x), v in day['sec'].items() if x == 'M'))
    for k in ('mF', 'mM', 'aF', 'aM'):
        if z(f[k]) != sums[k]: checks['farm_vs_sections'].append((ds, k, f[k], sums[k]))
    he = [day['hall'][h]['eggs'] for h in HALLS]
    if f['eggs'] is not None and sum(z(x) for x in he) != f['eggs']:
        checks['hall_eggs_vs_farm'].append((ds, f['eggs'], he))
    aF, aM = sums['aF'], sums['aM']
    sF = sum(v['start'] for (s, x), v in day['sec'].items() if x == 'F')
    sM = sum(v['start'] for (s, x), v in day['sec'].items() if x == 'M')
    mF = sum(v['mort'] for (s, x), v in day['sec'].items() if x == 'F')
    mM = sum(v['mort'] for (s, x), v in day['sec'].items() if x == 'M')
    arF = sum(v['arr'] for (s, x), v in day['sec'].items() if x == 'F')
    arM = sum(v['arr'] for (s, x), v in day['sec'].items() if x == 'M')
    dpF = sum(v['sold'] for (s, x), v in day['sec'].items() if x == 'F')
    dpM = sum(v['sold'] for (s, x), v in day['sec'].items() if x == 'M')
    # egg weight: egg-weighted mean of halls (sheet D is a plain mean of the three halls, incl. empty H7 = 0)
    ews = [(day['hall'][h]['ew'], day['hall'][h]['eggs']) for h in HALLS if day['hall'][h]['ew'] and day['hall'][h]['eggs']]
    ew = round(sum(a * b for a, b in ews) / sum(b for a, b in ews), 2) if ews else None
    ew_sheet = C(mr, 'D') if mr else None
    kgF = r1(C(fr, 'AF')) if fr else None; kgM = r1(C(fr, 'AG')) if fr else None
    dF = C(fr, 'R') if fr else None; dM = C(fr, 'S') if fr else None
    lei = C(fr, 'AJ') if fr else None
    stF = C(fr, 'AH') if fr else None; stM = C(fr, 'AI') if fr else None
    gw = feed_guide.get(w, (None, None))
    gl = real_ghid.get(w, {}).get('lay')
    if fr:
        i = F_index[ds]
        if dF or dM:
            for sex, q in (('F', dF), ('M', dM)):
                if q: deliv_rows.append(dict(date=ds, sex=sex, qty=q, row=i))
        if lei: lei_cells.append(dict(date=ds, lei=lei, row=i))
        for c in ('AK', 'AL'):
            v = fr[ci(c) - 1]
            if isinstance(v, str) and v.strip(): notes_cells.append((ds, c, v.strip()))
        # stock recompute
        if stock_prev is not None:
            for sex, st, dl, kg_ in (('F', stF, dF, C(fr, 'AF')), ('M', stM, dM, C(fr, 'AG'))):
                exp = stock_prev[sex] + z(dl) - z(kg_)
                if st is not None and abs(exp - st) > 0.6: checks['stock'].append((ds, sex, round(exp, 1), st))
        stock_prev = {'F': z(stF), 'M': z(stM)}
    lay_start = pct(f['eggs'], sF) if f['eggs'] is not None else None
    farm_rows.append([ds, age, w, aF, aM, mF, mM, dpF, dpM, arF, arM, sums['mF'], sums['mM'], f['eggs'], f['inc'], f['jos'],
                      lay_start, r2(f['lay']), r2(f['incp']), r2(f['josp']), ew, r2(ew_sheet), kgF, kgM, dF, dM, lei,
                      r1(stF), r1(stM), gw[0], gw[1], gl])
    day['farm'] = dict(aF=aF, aM=aM, sF=sF, sM=sM, mF=mF, mM=mM, arF=arF, arM=arM, dpF=dpF, dpM=dpM, eggs=f['eggs'],
                       inc=f['inc'], jos=f['jos'], lay=lay_start, lay_sheet=f['lay'], lay_start=lay_start, incp=f['incp'], josp=f['josp'],
                       ew=ew, ew_sheet=ew_sheet, kgF=kgF, kgM=kgM, dF=dF, dM=dM, lei=lei, stF=stF, stM=stM,
                       gF=gw[0], gM=gw[1], gl=gl, mF_sheet=f['mF'], mM_sheet=f['mM'])
    state[ds] = day

for k, v in checks.items():
    if k != 'total_rows': log(f'check {k}: {len(v)} issue(s)', v[:8])
log('max single-day mortality per section × sex before 12.11:', {f'{s}{x}': v for (s, x), v in sorted(max_normal_mort.items())})
inflows_after = [(row[0], row[3], row[4], row[7], row[8]) for row in sec_rows if row[0] > '2026-03-01' and (row[7] or row[8])]
log('arrivals/inflows after 01.03 (moves between halls):', inflows_after)

# ---------------------------------------------------------------- weekly + Total-row check
weeks = []
mismatch = []
for w in sorted(Ht):
    tr = Ht[w]
    wd = [ds for ds in dates if sapt(ds) == w]
    if not wd: continue
    full = len(wd) == 7
    # sheet-style raw sums (mortality incl. negative arrivals)
    def raw_sum(sec, sex): return sum(state[ds]['sec'][(sec, sex)]['raw'] for ds in wd)
    comp = []
    comp.append(('farm ♂ mort', sum(state[ds]['farm']['mM_sheet'] or 0 for ds in wd), C(tr, 'D')))
    comp.append(('farm ♀ mort', sum(state[ds]['farm']['mF_sheet'] or 0 for ds in wd), C(tr, 'F')))
    comp.append(('farm ♂ mort (from sections)', sum(raw_sum(s, 'M') for s in SECTIONS), C(tr, 'D')))
    comp.append(('farm ♀ mort (from sections)', sum(raw_sum(s, 'F') for s in SECTIONS), C(tr, 'F')))
    for k, c in (('eggs', 'H'), ('inc', 'J'), ('jos', 'L')):
        comp.append((f'farm {k}', sum(z(state[ds]['farm'][k]) for ds in wd), C(tr, c)))
    for s in SECTIONS:
        cmM, _, cmF, _ = HSEC[s]
        comp.append((f'{s} ♂ mort', raw_sum(s, 'M'), C(tr, cmM)))
        comp.append((f'{s} ♀ mort', raw_sum(s, 'F'), C(tr, cmF)))
    for h in HALLS:
        ce, cl, cinc, cincp, cj, cjp = HHALL[h]
        for k, c in (('eggs', ce), ('inc', cinc), ('jos', cj)):
            comp.append((f'{h} {k}', sum(z(state[ds]['hall'][h][k]) for ds in wd), C(tr, c)))
    for name, mine, sheet in comp:
        if z(sheet) != mine: mismatch.append((w, name, mine, sheet))
    # sheet lay% on Total = plain mean of daily % (check)
    lays = [state[ds]['farm']['lay_sheet'] for ds in wd if state[ds]['farm']['lay_sheet'] is not None]
    lay_sheet = C(tr, 'I')
    if lays and lay_sheet is not None and abs(sum(lays) / len(lays) - lay_sheet) > 0.02:
        mismatch.append((w, 'farm lay% (mean of daily)', round(sum(lays) / len(lays), 2), lay_sheet))

    def scope(get_secs, hall=None):
        secs = get_secs
        mF = sum(state[ds]['sec'][(s, 'F')]['mort'] for ds in wd for s in secs)
        mM = sum(state[ds]['sec'][(s, 'M')]['mort'] for ds in wd for s in secs)
        dpF = sum(state[ds]['sec'][(s, 'F')]['sold'] for ds in wd for s in secs)
        dpM = sum(state[ds]['sec'][(s, 'M')]['sold'] for ds in wd for s in secs)
        arF = sum(state[ds]['sec'][(s, 'F')]['arr'] for ds in wd for s in secs)
        arM = sum(state[ds]['sec'][(s, 'M')]['arr'] for ds in wd for s in secs)
        mvF = sum(state[ds]['sec'][(s, 'F')]['mv'] for ds in wd for s in secs)
        mvM = sum(state[ds]['sec'][(s, 'M')]['mv'] for ds in wd for s in secs)
        startF = sum(state[wd[0]]['sec'][(s, 'F')]['start'] for s in secs)
        startM = sum(state[wd[0]]['sec'][(s, 'M')]['start'] for s in secs)
        endF = sum(state[wd[-1]]['sec'][(s, 'F')]['alive'] for s in secs)
        endM = sum(state[wd[-1]]['sec'][(s, 'M')]['alive'] for s in secs)
        hen_days = sum(state[ds]['sec'][(s, 'F')]['start'] for ds in wd for s in secs)        # D159: start-of-day hens
        hen_days_prod = sum(state[ds]['sec'][(s, 'F')]['alive'] for ds in wd for s in secs)  # Productivitate: end of day
        kgF = sum(z(state[ds]['sec'][(s, 'F')]['kg']) for ds in wd for s in secs)
        kgM = sum(z(state[ds]['sec'][(s, 'M')]['kg']) for ds in wd for s in secs)
        bird_days_F = sum(state[ds]['sec'][(s, 'F')]['start'] for ds in wd for s in secs if state[ds]['sec'][(s, 'F')]['kg'])
        bird_days_M = sum(state[ds]['sec'][(s, 'M')]['start'] for ds in wd for s in secs if state[ds]['sec'][(s, 'M')]['kg'])
        o = dict(mort_F=mF, mort_M=mM, sold_F=dpF, sold_M=dpM, mort_F_sheet=mF + dpF, mort_M_sheet=mM + dpM, arr_F=arF, arr_M=arM, mv_F=mvF, mv_M=mvM, alive_F_start=startF,
                 alive_M_start=startM, alive_F_end=endF, alive_M_end=endM, hen_days=hen_days, hen_days_prod=hen_days_prod,
                 kg_F=round(kgF, 1), kg_M=round(kgM, 1),
                 g_F=r1(1000 * kgF / bird_days_F) if bird_days_F else None,
                 g_M=r1(1000 * kgM / bird_days_M) if bird_days_M else None)
        return o
    farm = scope(SECTIONS)
    farm.update(eggs=C(tr, 'H'), inc=C(tr, 'J'), jos=C(tr, 'L'),
                lay_pct=pct(z(C(tr, 'H')), farm['hen_days']), lay_pct_prod=pct(z(C(tr, 'H')), farm['hen_days_prod']), lay_pct_sheet=r2(lay_sheet),
                inc_pct=pct(z(C(tr, 'J')), C(tr, 'H')), jos_pct_sheet=r2(C(tr, 'M')),
                egg_g=r2(C(Mt[w], 'D')) if w in Mt else None,
                cum_mort_pct_F_sheet=r2(C(tr, 'G')), cum_mort_pct_M_sheet=r2(C(tr, 'E')),
                deliv_F_kg=sum(z(state[ds]['farm']['dF']) for ds in wd), deliv_M_kg=sum(z(state[ds]['farm']['dM']) for ds in wd),
                lei=sum(z(state[ds]['farm']['lei']) for ds in wd),
                stock_F_end=r1(state[wd[-1]]['farm']['stF']), stock_M_end=r1(state[wd[-1]]['farm']['stM']))
    if farm['egg_g'] is not None and farm['egg_g'] == 0: farm['egg_g'] = None
    halls = {}
    for h in HALLS:
        secs = [s for s in SECTIONS if HALL_OF[s] == h]
        o = scope(secs)
        ce, cl, cinc, cincp, cj, cjp = HHALL[h]
        o.update(eggs=C(tr, ce), inc=C(tr, cinc), jos=C(tr, cj), lay_pct=pct(z(C(tr, ce)), o['hen_days']), lay_pct_prod=pct(z(C(tr, ce)), o['hen_days_prod']),
                 lay_pct_sheet=r2(C(tr, cl)), inc_pct=pct(z(C(tr, cinc)), C(tr, ce)), jos_pct_sheet=r2(C(tr, cjp)),
                 egg_g=r2(C(Mt[w], MASA_HALL[h][0])) if w in Mt and C(Mt[w], MASA_HALL[h][0]) else None)
        o['sections'] = {}
        for s in secs:
            cmM, caM, cmF, caF = HSEC[s]
            so = scope([s])
            o['sections'][s] = dict(mort_F=so['mort_F'], mort_M=so['mort_M'], sold_F=so['sold_F'], sold_M=so['sold_M'],
                                    arr_F=so['arr_F'], arr_M=so['arr_M'], mv_F=so['mv_F'], mv_M=so['mv_M'], alive_F_end=so['alive_F_end'], alive_M_end=so['alive_M_end'],
                                    kg_F=so['kg_F'], kg_M=so['kg_M'], g_F=so['g_F'], g_M=so['g_M'],
                                    cum_mort_pct_F_sheet=r2(C(tr, caF)), cum_mort_pct_M_sheet=r2(C(tr, caM)))
        halls[h] = o
    hw = [(halls[h]['egg_g'], halls[h]['eggs']) for h in HALLS if halls[h]['egg_g'] and halls[h]['eggs']]
    farm['egg_g_sheet'] = farm['egg_g']
    farm['egg_g'] = round(sum(a * b for a, b in hw) / sum(b for a, b in hw), 2) if hw else None
    g = real_ghid.get(w, {})
    fg = feed_guide.get(w, (None, None))
    guide = dict(lay_pct=g.get('lay'), mort_cum_pct=g.get('mort'), eggs_hen_week=g.get('egg_wk'),
                 hatch_eggs_hen_week=g.get('inc_wk'), feed_g_F=fg[0], feed_g_M=fg[1],
                 lay_pct_productivitate=prod.get(w, {}).get('lay_guide'),
                 eggs_plan_week=plan.get(w, {}).get('eggs_plan'))
    if guide['lay_pct'] is not None and farm['lay_pct'] is not None and farm['eggs']:
        farm['lay_vs_guide_pts'] = round(farm['lay_pct'] - guide['lay_pct'], 2)
        for h in HALLS:
            if halls[h]['lay_pct'] is not None and halls[h]['eggs']:
                halls[h]['lay_vs_guide_pts'] = round(halls[h]['lay_pct'] - guide['lay_pct'], 2)
    # cross-check with Productivitate / RealGhid
    if w in prod and prod[w]['hen_days'] is not None:
        if prod[w]['hen_days'] != farm['hen_days_prod']: mismatch.append((w, 'Productivitate hen-days', farm['hen_days_prod'], prod[w]['hen_days']))
        if prod[w]['eggs'] != farm['eggs']: mismatch.append((w, 'Productivitate eggs', farm['eggs'], prod[w]['eggs']))
    if w in plan and plan[w]['eggs_real'] is not None and plan[w]['eggs_real'] != farm['eggs']:
        mismatch.append((w, 'RealGhid real eggs', farm['eggs'], plan[w]['eggs_real']))
    weeks.append(dict(sapt=w, **{'from': wd[0], 'to': wd[-1]}, days=len(wd), complete=full and wd[-1] <= TODAY,
                      age_days_from=age_days(wd[0]), farm=farm, halls=halls, guide=guide))
log('Total-row check: weeks', len(weeks), 'mismatches', len(mismatch))
for m in mismatch: log('  mismatch', m)

# cumulative-mortality % base test (sheet Total E/G)
w0 = 27
tr = Ht[w0]
cumF = sum(state[ds]['farm']['mF'] for ds in dates if sapt(ds) <= w0)
log('cum mort test wk27: sheet ♀', C(tr, 'G'), '♂', C(tr, 'E'), '| mine vs arrivals ♀',
    pct(cumF, 26883), '♂', pct(sum(state[ds]['farm']['mM'] for ds in dates if sapt(ds) <= w0), 3024))

# ---------------------------------------------------------------- sales (D155)
sales_rows = []
for ds in dates:
    for s_ in SECTIONS:
        sF, sM = state[ds]['sec'][(s_, 'F')]['sold'], state[ds]['sec'][(s_, 'M')]['sold']
        if sF or sM: sales_rows.append([ds, age_days(ds), sapt(ds), s_, sF, sM])
log('sales', len(sales_rows), 'rows; ♀', sum(r[4] for r in sales_rows), '♂', sum(r[5] for r in sales_rows))

# ---------------------------------------------------------------- hirbovat-daily.json
anom = {
    'farm_column_vs_sections': [dict(date=a, col=b, sheet=c, sections=d) for a, b, c, d in checks['farm_vs_sections']],
    'alive_continuity': [dict(date=a, section=b, sex=c, prev=d, mort_raw=e, alive=f) for a, b, c, d, e, f in checks['alive_continuity']],
    'kg_not_norm_x_alive_start': [dict(date=a, section=b, sex=c, g=d, alive_start=e, kg=f) for a, b, c, d, e, f in checks['kg_vs_norm']],
    'stock_recompute': [dict(date=a, sex=b, expected=c, sheet=d) for a, b, c, d in checks['stock']],
    'hall_eggs_vs_farm': checks['hall_eggs_vs_farm'],
    'lay_pct_vs_eggs_over_hens': checks['lay_pct'],
}
dump('hirbovat-daily.json', {
    'meta': dict(GEN, range=[FIRST, TODAY], farm_id='hirbovat', series_id='seria-2026', cycle_id='cyc-hirbovat-2026', real=True,
                 sheets={'Hirbovat': 'mortality + alive per compartment × sex (cols O–BD), eggs per hall and farm (D–M)',
                         'Hrana Hirbovat': 'D/E guide g, F–Q g/bird, R/S deliveries kg ♀/♂ feed, T–AE kg per compartment × sex, AF/AG kg ♀/♂, AH/AI stock ♀/♂, AJ lei',
                         'Masa': 'egg weight per hall (I, N, S) and farm (D)', '%%': 'lay/inc/jos % (cross-check only; identical to Hirbovat)',
                         'RealGhid': 'guide lay % by age week (col B)'},
                 conventions=[
                     'Rows are arrays; column names in each block\'s "cols". Sex F = ♀, M = ♂. Dates ISO. null = not in the sheet.',
                     'age_days is the UI column "Vârstă (zile)" (D159): bird age per D151, 407 on 18.11.2026 (the workbook\'s Zile = age_days + 1 = 408 is not used).',
                     'age_days = date − 2025-10-07 (age 0 on placement day); sapt = floor(age_days/7)+1 (D151). Sheet "Zile" = age_days + 1, "Sapt" = sapt.',
                     'mort = real deaths that day; sold = emptying removals that the sheet books as mortality, which are Vânzare (sale) events in the app (D155); mort_sheet = the sheet\'s raw cell (deaths + sales, arrivals and moves negative) = mort + sold − arr − mv. arr = arrivals (negative cells 22–28.02).',
                     'alive = birds at the end of the day as typed in the sheet (= previous alive − mort − sold + arr + mv, checked on every row).',
                     'sold rule (D155): a compartment-day from 2026-11-12 with ≥ 40 birds booked is a sale (all of it; no real deaths assumed in that cell); 03.03–11.11 a compartment never lost more than %d ♀ / %d ♂ in a day. sales = the same per date × compartment (♀, ♂).' % (max(v[0] for (a, b), v in max_normal_mort.items() if b == 'F'), max(v[0] for (a, b), v in max_normal_mort.items() if b == 'M')),
                     'mv = birds moved between compartments (signed, + in / − out). The sheet has one: 22.10.2026 H7.1 ♂ −19 "mortality" (inflow) with H7.2 ♂ 20 the same day, read as 19 moved H7.2 → H7.1 + 1 dead (transfers-seria2026.json internal_moves_hirbovat). mort excludes it; sheet cell = mort − arr − mv.',
                     'g = norm g/bird (Hrana "Norma de consum"); kg = sheet kg = g × birds alive at the START of the day (previous day end) / 1000 (checked; differs from data-model "alive up to that date").',
                     'section_day rows exist from the first arrival in that compartment to the day it was emptied (pre-arrival and post-emptying zero rows dropped).',
                     'lay_pct (hall_day, farm_day) = eggs ÷ hens alive at the START of the day × 100 (D156; 18.11 farm 41,1 %). lay_pct_sheet = the sheet\'s % = eggs ÷ hens at the end of the day (47,08 % on 18.11, inflated by sales).',
                     'eggs = total collected = incubabile + Cat 1–3 (Jos counted separately, not in eggs; as D18). Cat 1–3 not extracted (D153).',
                     'egg_g farm = egg-weighted mean of hall weights (the sheet\'s Masa col D is a plain mean of three halls incl. an empty H7 = 0 after 13.11; kept as egg_g_sheet).',
                     'deliv_F_kg / deliv_M_kg = Hrana R ("gaina") / S ("cocos") — feed for hens / roosters; negative = stock adjustment (−300 on 21.02; −3 000 ♀ / +3 000 ♂ on 26.02 = feed moved between the two stocks). lei = Hrana AJ on that row (see feed-deliveries.json for the invoice pairing).',
                     'guide_g_F/M = Hrana D/E by sapt; guide_lay_pct = RealGhid col B by age week (see guide-laying.json).',
                 ],
                 not_extracted=['Hirbovat col N "Total incubabil" (row 2 = cumulative incubabile 4 902 522; daily cells = day-on-day change of incubabile)',
                                'Masa Cat 1–3 % and Raspunsuri counts (D153)', 'Hrana rows after 18.11 (19.11 partly filled)'],
                 anomalies=anom),
    'sections': SECTIONS, 'halls': {h: [s for s in SECTIONS if HALL_OF[s] == h] for h in HALLS},
    'section_day': {'cols': ['date', 'age_days', 'sapt', 'section', 'sex', 'mort', 'sold', 'arr', 'mv', 'alive', 'g', 'kg', 'mort_sheet'], 'rows': sec_rows},
    'sales': {'cols': ['date', 'age_days', 'sapt', 'section', 'F', 'M'], 'rows': sales_rows},
    'hall_day': {'cols': ['date', 'age_days', 'sapt', 'hall', 'alive_F', 'alive_M', 'mort_F', 'mort_M', 'sold_F', 'sold_M', 'eggs', 'inc', 'jos',
                          'lay_pct', 'lay_pct_sheet', 'inc_pct', 'jos_pct', 'egg_g', 'kg_F', 'kg_M'], 'rows': hall_rows},
    'farm_day': {'cols': ['date', 'age_days', 'sapt', 'alive_F', 'alive_M', 'mort_F', 'mort_M', 'sold_F', 'sold_M', 'arr_F', 'arr_M',
                          'mort_F_sheet', 'mort_M_sheet', 'eggs', 'inc', 'jos', 'lay_pct', 'lay_pct_sheet', 'inc_pct', 'jos_pct', 'egg_g', 'egg_g_sheet', 'kg_F', 'kg_M',
                          'deliv_F_kg', 'deliv_M_kg', 'lei', 'stock_F_kg', 'stock_M_kg', 'guide_g_F', 'guide_g_M', 'guide_lay_pct'],
                 'rows': farm_rows},
})

# ---------------------------------------------------------------- hirbovat-weekly.json
dump('hirbovat-weekly.json', {
    'meta': dict(GEN, farm_id='hirbovat', series_id='seria-2026', real=True,
                 grain='one entry per age week (Tue–Mon) = one sheet "Total" row, SAPT 20–59. SAPT 59 (17–23.11) is partial: its Total row is a template whose formulas sum the two filled days 17–18.11 (complete = false, days = 2).',
                 definitions=[
                     'mort_* = real deaths that week (D155: sales excluded); sold_* = birds sold (emptying removals the sheet books as mortality); mort_*_sheet = mort + sold (what the sheet books); arrivals split out as arr_*, moves as mv_*.',
                     'Farm totals use compartment sums (D156): the sheet\'s farm ♂ mortality is wrong on 22 Tuesdays (see total_row_check).',
                     'hen_days = Σ hens alive at the START of each day (D159, same basis as the daily lay %); lay_pct = eggs / hen_days × 100 (D156/D159). hen_days_prod = Σ hens at the end of each day (= Productivitate col B, checked) and lay_pct_prod = eggs / hen_days_prod (= Productivitate "de facto", wk 58 32,58), kept for reference.',
                     'lay_pct_sheet = the sheet Total row = plain mean of the daily % (wk 58 33,55).',
                     'inc_pct = inc / eggs × 100; jos_pct_sheet as in the Total row; hall egg_g = Masa Total row; farm egg_g = egg-weighted mean of the halls (egg_g_sheet = Masa col D Total, a plain mean of three halls that counts an empty H7 as 0 from 14.11).',
                     'cum_mort_pct_*_sheet = the Total row\'s % under the "în viață" columns = cumulative deaths since 10.03.2026 (start of SAPT 24, lay start) ÷ birds alive at the end of 09.03.2026 (♀ 26 845 = RealGhid "intrat în ouat"; ♂ 2 931) — reproduced to ±0,05 pt; from 12.11 it includes the sales (D155). Compare with guide.mort_cum_pct.',
                     'kg_* = Σ daily kg; g_* = kg / bird-days on fed days (bird-weighted g/bird).',
                     'guide: lay_pct, mort_cum_pct, eggs_hen_week, hatch_eggs_hen_week = RealGhid by age week (app guide); feed_g = Hrana Hirbovat D/E; lay_pct_productivitate = the second lay guide (Productivitate col I), for reference; eggs_plan_week = RealGhid col M (26 845 × eggs/hen/week guide).',
                     'lay_vs_guide_pts = lay_pct − guide.lay_pct (percentage points).',
                 ],
                 total_row_check=dict(compared='farm and per-compartment mortality (sheet-style sums, arrivals negative), farm and hall eggs/inc/jos, farm lay% (mean of daily); plus Productivitate hen-days/eggs and RealGhid real eggs',
                                      weeks=len(weeks), mismatches=[dict(sapt=a, item=b, from_daily=c, sheet=d) for a, b, c, d in mismatch])),
    'weeks': weeks,
})

# ---------------------------------------------------------------- guide-laying.json
old_gl = load('guide-laying.json')
old_feed = {(r['age_week'], r['sex']): r['feed_g'] for r in old_gl['rows']}
feed_diff = []
rows = []
for w in range(19, 66):
    fg = feed_guide.get(w)
    g = real_ghid.get(w, {})
    for sex in ('F', 'M'):
        if fg and fg[0 if sex == 'F' else 1] is not None:
            fv, fsrc = fg[0 if sex == 'F' else 1], 'Hrana Hirbovat'
        else:
            fv, fsrc = old_feed.get((w, sex)), 'Seria_2026.xlsx Hrana Hîrbovăț (old)'
        if (w, sex) in old_feed and old_feed[(w, sex)] != fv: feed_diff.append((w, sex, old_feed[(w, sex)], fv))
        row = dict(guide_id='g-laying', age_week=w, sex=sex, feed_g=fv, feed_source=fsrc,
                   lay_pct=g.get('lay') if sex == 'F' else None,
                   mortality_cum_pct=g.get('mort') if sex == 'F' else None,
                   eggs_hen_cum=g.get('egg_cum') if sex == 'F' else None,
                   hatch_eggs_hen_cum=g.get('inc_cum') if sex == 'F' else None,
                   eggs_hen_week=g.get('egg_wk') if sex == 'F' else None,
                   hatch_eggs_hen_week=g.get('inc_wk') if sex == 'F' else None,
                   hatchability_pct=hatch_guide.get(w) if sex == 'F' else None,
                   body_weight_g=None, egg_weight_g=None, hatching_pct=None)
        rows.append(row)
log('feed guide new vs old differences:', feed_diff)
alt = [dict(age_week=w, lay_pct=prod[w]['lay_guide'], hatch_eggs_hen_cum=prod[w]['inc_hen_cum_guide'],
            lay_pct_realghid=real_ghid.get(w, {}).get('lay')) for w in sorted(prod) if prod[w]['lay_guide'] is not None]
dump('guide-laying.json', {
    'meta': dict(GEN, guide_name='Ross 308 ouat (illustrative name, D22)',
                 week_basis='age week (D151); rows age wk 19–65. The old cycle_week field (age − 18, D26) is dropped.',
                 app_guide='RealGhid (lay_pct, mortality_cum_pct, eggs/hatching eggs per hen) — recommended: it is the sheet the farm uses for plan vs real (RealGhid M/N, from 10.03.2026), and its eggs/hen/week column reproduces the plan figures.',
                 sources={'feed_g': 'Hrana Hirbovat cols D (G, ♀) / E (C, ♂), weekly, by the row\'s Sapt (wk 20–70); wk 19 from the old workbook',
                          'lay_pct, mortality_cum_pct, eggs_hen_cum, hatch_eggs_hen_cum, eggs_hen_week, hatch_eggs_hen_week': 'RealGhid cols B, C, D, E, F, G, age wk 23–62 (on ♀ rows)',
                          'hatchability_pct': 'RealGhid cols AF/AG (wk 24–65); not used this round (D153)'},
                 anomalies=['RealGhid wk 53 lay % typed "605" → 60.5 (between 61.7 and 59.3).',
                            'Two lay guides disagree: RealGhid B vs Productivitate I differ by ≥ 1 pt in %d weeks; e.g. wk 56 56,9 vs 61,1, wk 57 55,7 vs 59,9, wk 58 54,5 vs 58,8. Productivitate\'s curve is kept in alt_productivitate for reference only.' % len(lay_disagree),
                            'The previous lay_pct (Sheet9 col P, D29) was the hatchability guide shifted; replaced.',
                            'feed_g differences vs the old file: %s' % (feed_diff or 'none')],
                 missing='body weight, egg weight, laying-flock ♂ mortality: no guide (D29)'),
    'rows': rows,
    'alt_productivitate': alt,
})

# guide-growth: add age_week (= week)
gg = load('guide-growth.json')
for r in gg['rows']: r['age_week'] = r['week']
gg['meta']['week_basis'] = 'age week (D151): week = age_week = floor(age_days/7)+1, age 0 on placement day; growth rows were already by age week'
gg['meta']['updated_by'] = 'DATA03 (added age_week = week)'
dump('guide-growth.json', gg)

# weighings-weekly: D151 sapt + D157 guide week
ww = load('weighings-weekly.json')
bw = {(r['week'], r['sex']): r.get('body_weight_g') for r in gg['rows']}
wbad = []
for r in ww['rows']:
    r['age_days'] = age_days(r['date']); r['sapt'] = sapt(r['date']); r['guide_week'] = r['sapt'] - 1
    gwt = bw.get((r['guide_week'], r['sex']))
    r['guide_weight_g'] = gwt
    r['guide_diff_pct'] = round(100 * (r['avg_weight_g'] - gwt) / gwt, 1) if gwt and r['avg_weight_g'] is not None else None
    if r['age_week'] != r['guide_week'] or (r['standard_g'] is not None and gwt is not None and r['standard_g'] != gwt): wbad.append((r['section'], r['date']))
ww['meta']['week_basis'] = ('D151/D157: weighed on age day 7N, so sapt = N + 1 (e.g. 24.02.2026 = age 140 = SAPT 21); the row\'s age_week (= N, the sheet\'s week) = guide_week = '
                            'sapt − 1 = the week just ended; guide_weight_g = guide-growth body_weight_g of guide_week (= the sheet\'s standard_g); guide_diff_pct = avg vs that guide.')
ww['meta']['updated_by'] = 'DATA03 (added age_days, sapt, guide_week, guide_weight_g, guide_diff_pct)'
log('weighings: rows', len(ww['rows']), 'age_week ≠ guide_week or standard ≠ guide:', wbad[:5])
dump('weighings-weekly.json', ww)

# ---------------------------------------------------------------- Telița files: sapt
for name, keys in (('mortality-telita-daily.json', ('daily_totals', 'rows')), ('feed-daily-telita.json', ('days',))):
    obj = load(name)
    diff = 0
    for k in keys:
        for r in obj[k]:
            s = sapt(r['date']); r['age_days'] = age_days(r['date']); r['sapt'] = s
            if r.get('age_week') != s: diff += 1
    obj['meta']['week_basis'] = 'D151: age_days = date − 2025-10-07; sapt = floor(age_days/7)+1 (= age_week; day = age_days + 1)'
    obj['meta']['updated_by'] = 'DATA03 (added age_days, sapt)'
    log(name, 'sapt differs from age_week in', diff, 'rows')
    dump(name, obj)

# ---------------------------------------------------------------- farms-sections.json
fs = load('farms-sections.json')
to_new = {'H6': ['H5.1', 'H5.2'], 'H3': ['H5.2', 'H6.1'], 'H1': ['H6.1', 'H6.2', 'H7.1'], 'H2': ['H7.1', 'H7.2'],
          'H5-1': SECTIONS, 'H5-2': SECTIONS}
secs = []
for s in fs['sections']:
    if s['farm_id'] == 'telita':
        s.pop('to_hirbovat_by_sheet_name', None)
        s['to_hirbovat_2026'] = to_new[s['code']]
        secs.append(s)
for h in HALLS:
    secs.append(dict(id=f'hir-{h}', farm_id='hirbovat', code=h, parent_id=None, active=True, is_hall=True,
                     workbook_labels=[f'Hala {h[1]}'], old_code=OLD_CODE[h]))
    for s in SECTIONS:
        if HALL_OF[s] != h: continue
        active = state[TODAY]['sec'][(s, 'F')]['alive'] + state[TODAY]['sec'][(s, 'M')]['alive'] > 0
        secs.append(dict(id=f'hir-{s}', farm_id='hirbovat', code=s, parent_id=f'hir-{h}', active=True, sexes=['F', 'M'],
                         workbook_labels=[s, f'{s}G', f'{s}C'], old_code=OLD_CODE[s], birds_on_2026_11_18=active))
fs['sections'] = secs
fs['meta'] = dict(GEN, real='farm names, section codes and hall grouping from the workbooks (Telița: Seria_2026.xlsx Telita row 2; Hîrbovăț: Raport_NOU_2026.xlsx Hirbovat rows 1–2, Hrana Hirbovat row 2)',
                  illustrative='guide names (D22) and ids',
                  code_map_old_to_new={OLD_CODE[k]: k for k in ['H5', 'H6', 'H7'] + SECTIONS},
                  notes=['Telița sections hold one sex each in Seria 2026 (H1, H2, H3, H6 hens; H5-1, H5-2 roosters).',
                         'Hîrbovăț compartments hold both sexes; eggs are counted per hall (parent).',
                         'Hîrbovăț codes renamed H1/H3/H4 → H5/H6/H7 (D152); the old Hrana Hîrbovăț labels H5.1G … H7.2C were right. The previous-flock files (eggs-hirbovat-prev-*) still use the old hall codes H1/H3/H4 = H5/H6/H7.',
                         'to_hirbovat_2026 = where each Telița section\'s birds went on 22–28.02.2026 (dated movements, transfers-seria2026.json).',
                         'active = the section exists; birds_on_2026_11_18 = still holds birds (H7 emptied 14.11; H6.2 5 ♂ + 475 ♀ left).'])
dump('farms-sections.json', fs)

# ---------------------------------------------------------------- transfers-seria2026.json
tr_old = load('transfers-seria2026.json')
arrivals = []
for ds in dates:
    for s in SECTIONS:
        for sex in ('M', 'F'):
            a = state[ds]['sec'][(s, sex)]['arr']
            if a: arrivals.append(dict(date=ds, farm_id='hirbovat', section=s, sex=sex, qty=a))
arr_tot = {'F': sum(a['qty'] for a in arrivals if a['sex'] == 'F'), 'M': sum(a['qty'] for a in arrivals if a['sex'] == 'M')}
arr_counts = [dict(section=s, F=sum(a['qty'] for a in arrivals if a['section'] == s and a['sex'] == 'F'),
                   M=sum(a['qty'] for a in arrivals if a['section'] == s and a['sex'] == 'M')) for s in SECTIONS]
outs = tr_old['transfers_out']
out_tot = {x: sum(o['qty'] for o in outs if o['sex'] == x) for x in 'FM'}
def arr_on(ds, sex): return {a['section']: a['qty'] for a in arrivals if a['date'] == ds and a['sex'] == sex}
links = []
links.append(dict(out_date='2026-02-22', **{'from': ['H5-1', 'H5-2']}, sex='M', qty_out=1387 + 1607, arrival_date='2026-02-22', to=arr_on('2026-02-22', 'M')))
for od, frm, ad in (('2026-02-25', 'H6', '2026-02-25'), ('2026-02-26', 'H3', '2026-02-26'), ('2026-02-27', 'H1', '2026-02-27'), ('2026-02-27', 'H2', '2026-02-28')):
    q = next(o['qty'] for o in outs if o['date'] == od and o['from_section'] == frm)
    links.append(dict(out_date=od, **{'from': [frm]}, sex='F', qty_out=q, arrival_date=ad, to=arr_on(ad, 'F')))
for l in links:
    l['qty_in'] = sum(l['to'].values()); l['difference_in_minus_out'] = l['qty_in'] - l['qty_out']
transfers = dict(meta=dict(GEN, sheets='Telița side: Seria_2026.xlsx Telita "Corectie" R–W (unchanged from DATA01); Hîrbovăț side: Raport_NOU_2026.xlsx Hirbovat (arrivals = negative "mortality" cells 22–28.02.2026)',
                           range=['2025-11-05', '2026-02-28'], real=True,
                           conventions='qty positive; sex F/M; section codes new (H5.1 … H7.2, D152)',
                           design_dates='D151/D152: the Hîrbovăț cycle starts 2026-02-22 (first birds). D21/D25 (start 10.02) are superseded by the real dates.',
                           reconciliation=dict(out_of_telita=out_tot, arrived_hirbovat=arr_tot,
                                               difference_arrived_minus_out={x: arr_tot[x] - out_tot[x] for x in 'FM'},
                                               superseded_arrivals_old_workbook={'F': 26765, 'M': 3060, 'source': 'Seria_2026.xlsx Sheet9 (D28)'},
                                               notes=['Every ♀ link matches exactly except H2: 6 999 out on 27.02, 7 001 in on 28.02 (+2).',
                                                      '♂: 2 994 left Telița on 22.02, 3 024 arrived (+30); Telița then still shows ♂ residuals (D25).',
                                                      'Pairing by date: H6 → H5.1/H5.2, H3 → H5.2/H6.1, H1 → H6.1/H6.2/H7.1, H2 → H7.1/H7.2; ♂ H5-1+H5-2 → all six compartments (498 each, H7.2 534).',
                                                      'One move inside Hîrbovăț after 28.02: 22.10.2026 H7.2 → H7.1 ♂ 19 (inferred from an inflow cell, see internal_moves_hirbovat). No move between halls.'])),
                 transfers_out=outs, arrivals=arrivals, arrival_counts=arr_counts, arrival_totals=arr_tot,
                 links_inferred_by_date=links, internal_moves_hirbovat=internal_moves, internal_moves_telita=tr_old['internal_moves_telita'],
                 other_corrections_telita=tr_old['other_corrections_telita'])
dump('transfers-seria2026.json', transfers)
log('arrivals', arr_tot, 'outs', out_tot)

# ---------------------------------------------------------------- cycles.json
cy = load('cycles.json')
cy['meta'] = dict(GEN, derived=True, week='sapt = age week (D151), not counted from the cycle start; cycle dates only bound the stay')
for s in cy['series']:
    if s['id'] == 'seria-2026':
        s.update(placement_date='2025-10-07', age_at_placement_days=0)
    if s['id'] == 'seria-2025':
        s.update(placement_date='2024-11-01', placement_inferred=True, age_at_placement_days=0,
                 placement_note='inferred: previous-flock data label 2025-04-04 as age wk 23 and 2025-03-07 as wk 19 (D36) → age 0 on 2024-11-01 if placed as day-old chicks')
for c in cy['cycles']:
    if c['id'] == 'cyc-telita-2026':
        c.update(age_days_start=0, sapt_start=1, age_days_end=age_days('2026-02-28'), sapt_end=sapt('2026-02-28'))
        c.pop('weeks', None); c['weeks'] = sapt('2026-02-28')
    if c['id'] == 'cyc-hirbovat-2026':
        for k in list(c): c.pop(k) if k not in ('id', 'series_id', 'farm_id', 'guide_id') else None
        c.update(start_date='2026-02-22', end_date=None, status='open', started_by='auto',
                 age_days_start=age_days('2026-02-22'), sapt_start=sapt('2026-02-22'),
                 today=TODAY, age_days_today=age_days(TODAY), sapt_today=sapt(TODAY),
                 emptying_since='2026-11-12', real=True,
                 notes='Start = first day birds are on the farm: ♂ 3 024 arrived 22.02.2026 (♀ 26 883 on 25–28.02), the data-model auto start (first placement on an empty farm). The sheets start 17.02 (feed stocked 2 000 ♀ + 1 000 ♂ kg, no birds). Supersedes 10.02 (D21/D25). Still open on 18.11 (♀ 12 561 · ♂ 794 alive). Emptying by sale since 12.11 (D155; ♀ ' + f"{sum(r[4] for r in sales_rows):,}".replace(',', ' ') + ' · ♂ ' + f"{sum(r[5] for r in sales_rows):,}".replace(',', ' ') + ' sold to 18.11): H7 empty 14.11, H6.2 nearly empty, H6.1 started 18.11. The cycle closes automatically when the last sale takes the farm to 0.')
    if c['id'] == 'cyc-hirbovat-2025':
        c.update(age_days_start=126, sapt_start=19, age_days_end=(datetime.date(2026, 1, 28) - datetime.date(2024, 11, 1)).days,
                 sapt_end=(datetime.date(2026, 1, 28) - datetime.date(2024, 11, 1)).days // 7 + 1,
                 week_note='D151: weeks are age weeks from the inferred placement 2024-11-01; the old cycle_week (age − 18) is dropped')
        for k in ('data_start_cycle_week',): c.pop(k, None)
        c['notes'] = 'Start inferred (D36): age wk 19 = 2025-03-07. Data from 2025-04-04 (SAPT 23, earliest egg/feed date in "import date"/DataHirbovat) to 2026-01-28 (SAPT 65). Weeks are age weeks (D151); the eggs-hirbovat-prev-* files still carry cycle_week = age − 18 and the old hall codes H1/H3/H4 (= H5/H6/H7).'
        c['data_start_sapt'] = sapt_prev = (datetime.date(2025, 4, 4) - datetime.date(2024, 11, 1)).days // 7 + 1
dump('cycles.json', cy)

# ---------------------------------------------------------------- counts-latest.json
start_counts = {(a['section'], a['sex']): 0 for a in arrivals}
for a in arrivals: start_counts[(a['section'], a['sex'])] += a['qty']
sections_out, halls_out = [], []
tot = {'F': {}, 'M': {}}
for s in SECTIONS:
    det = {}
    for sex in 'FM':
        st = start_counts.get((s, sex), 0)
        mort = sum(state[ds]['sec'][(s, sex)]['mort'] for ds in dates)
        sold = sum(state[ds]['sec'][(s, sex)]['sold'] for ds in dates)
        alive = state[TODAY]['sec'][(s, sex)]['alive']
        det[sex] = dict(start=st, mortality=mort, sold=sold, mortality_sheet=mort + sold, alive=alive,
                        mortality_pct=pct(mort, st), today_mortality=state[TODAY]['sec'][(s, sex)]['mort'],
                        today_sold=state[TODAY]['sec'][(s, sex)]['sold'], culls=None, sexing_errors=None)
        for k in ('start', 'mortality', 'sold', 'mortality_sheet', 'alive', 'today_mortality', 'today_sold'):
            tot[sex][k] = tot[sex].get(k, 0) + det[sex][k]
    sections_out.append(dict(section=s, hall=HALL_OF[s], F=det['F']['alive'], M=det['M']['alive'], detail=det))
for h in HALLS:
    halls_out.append(dict(hall=h, F=sum(x['F'] for x in sections_out if x['hall'] == h), M=sum(x['M'] for x in sections_out if x['hall'] == h)))
fh = state[TODAY]['farm']
counts = dict(
    meta=dict(GEN, as_of=TODAY, illustrative=False,
              label='Real (Hirbovat sheet, end of 18.11.2026). Telița cycle closed 28.02.2026 (real, unchanged).',
              definitions='alive = sheet count at the end of 18.11.2026 (= start − mortality − sold ± moves); mortality = real deaths since arrival; sold = Vânzare (D155: emptying removals from 12.11 that the sheet books as mortality, ≥ 40 birds in a compartment-day); mortality_sheet = mortality + sold; mortality_pct = mortality / start × 100; culls/sexing_errors = null (not in the workbook; illustrative-events-hirbovat.json carves a few out of mortality)',
              schema_changes='cycle_week now = sapt (age week, D151); added hall per section, sold, mortality_sheet, today_mortality, today_sold, sapt, age_days (mortality now excludes sales, D155); removals_pct dropped; culls/sexing_errors null'),
    farms=[dict(farm_id='hirbovat', cycle_id='cyc-hirbovat-2026', cycle_status='open', cycle_week=sapt(TODAY), sapt=sapt(TODAY),
                age_days=age_days(TODAY), illustrative=False, halls=halls_out,
                totals=dict(F=fh['aF'], M=fh['aM'], all=fh['aF'] + fh['aM']), sections=sections_out,
                cycle=dict(start_date='2026-02-22', start_counts=arr_tot,
                           removals={x: dict(mortality=tot[x]['mortality'], sold=tot[x]['sold'], mortality_sheet=tot[x]['mortality_sheet']) for x in 'FM'},
                           mortality_pct={x: pct(tot[x]['mortality'], arr_tot[x]) for x in 'FM'},
                           today_mortality={x: tot[x]['today_mortality'] for x in 'FM'},
                           today_sold={x: tot[x]['today_sold'] for x in 'FM'},
                           close_rule='the cycle closes automatically when the last sale takes the farm to 0; on 18.11.2026 H5 and H6 still hold birds, so it stays open')),
           next(f for f in load('counts-latest.json')['farms'] if f['farm_id'] == 'telita')],
    series=[dict(series_id='seria-2026', F=fh['aF'], M=fh['aM'], all=fh['aF'] + fh['aM'],
                 by_farm={'hirbovat': fh['aF'] + fh['aM'], 'telita': 0}, illustrative=False)])
dump('counts-latest.json', counts)

# ---------------------------------------------------------------- feed-deliveries.json
fd_old = load('feed-deliveries.json')
tel_rows = [r for r in fd_old['rows'] if r['farm_id'] == 'telita']
# pair lei cells with delivery rows
for d_ in deliv_rows: d_['lei'] = None; d_['basis'] = None
def cands(row): return [d_ for d_ in deliv_rows if d_['row'] == row and d_['qty'] > 0 and d_['lei'] is None]
unpaired = []
for lc in lei_cells:
    cs = cands(lc['row']); basis = 'same row'
    if not cs:
        cs = cands(lc['row'] - 1); basis = 'lei on the next day\'s row'
    if not cs:
        unpaired.append(lc); continue
    if len(cs) == 1:
        cs[0]['lei'] = lc['lei']; cs[0]['basis'] = basis
    else:
        tot_kg = sum(c['qty'] for c in cs)
        if 7.0 <= lc['lei'] / tot_kg <= 9.5:
            for c in cs:
                c['lei'] = round(lc['lei'] * c['qty'] / tot_kg, 2); c['basis'] = 'one lei figure for ♀+♂ on this row, split pro rata by kg'
        else:
            fF = [c for c in cs if c['sex'] == 'F'][0]
            fF['lei'] = lc['lei']; fF['basis'] = basis + ' (♀ only: lei / ♀+♂ kg out of range)'
log('lei cells unpaired:', unpaired)
hir_rows = []
for d_ in deliv_rows:
    q = d_['qty']
    if q <= 0 or (d_['date'] == '2026-02-26' and d_['sex'] == 'M' and q == 3000):
        continue
    hir_rows.append(dict(id=f"del-hir-{d_['date']}-{'f' if d_['sex'] == 'F' else 'm'}", date=d_['date'], farm_id='hirbovat',
                         qty_kg=q, for_sex=d_['sex'], workbook_column='R "gaina"' if d_['sex'] == 'F' else 'S "cocos"',
                         feed_type_id='ft-breeder2' if d_['sex'] == 'F' else 'ft-cocosi', feed_type_basis='sex of the sheet column; product unknown',
                         priced=d_['lei'] is not None, total_lei=d_['lei'],
                         price_per_kg=round(d_['lei'] / q, 4) if d_['lei'] else None, lei_basis=d_['basis']))
adjust = [dict(date=d_['date'], for_sex=d_['sex'], qty_kg=d_['qty'], note=('stock moved from ♀ to ♂ feed' if d_['date'] == '2026-02-26' else 'stock adjustment (no explanation in the sheet)'))
          for d_ in deliv_rows if d_['qty'] < 0 or (d_['date'] == '2026-02-26' and d_['sex'] == 'M' and d_['qty'] == 3000)]
for r in tel_rows: r.setdefault('for_sex', None)
sumR = sum(z(C(r, 'R')) for r in F_rows[2:] if todate(r[0])); sumS = sum(z(C(r, 'S')) for r in F_rows[2:] if todate(r[0]))
sumAJ = sum(z(C(r, 'AJ')) for r in F_rows[2:] if todate(r[0]))
log('Hrana header totals R2/S2/AJ2', C(F_rows[1], 'R'), C(F_rows[1], 'S'), C(F_rows[1], 'AJ'), 'vs column sums', sumR, sumS, sumAJ)
hk = {x: sum(r['qty_kg'] for r in hir_rows if r['for_sex'] == x) for x in 'FM'}
price_out = [(r['date'], r['for_sex'], r['qty_kg'], r['total_lei'], r['price_per_kg']) for r in hir_rows if r['price_per_kg'] and not 7.0 <= r['price_per_kg'] <= 9.5]
log('price outliers', price_out)
hl = round(sum(r['total_lei'] or 0 for r in hir_rows), 2)
unpriced = [r for r in hir_rows if not r['priced']]
fdel = dict(meta=dict(GEN, sources={'telita': 'Seria_2026.xlsx Hrana Telita L, X, Y (DATA01, unchanged)',
                                    'hirbovat': 'Raport_NOU_2026.xlsx Hrana Hirbovat R (kg feed for hens, "gaina"), S (kg for roosters, "cocos"), AJ (lei), AK/AL text notes'},
                      range=['2025-10-07', '2026-11-13'],
                      real='date, kg, lei real; feed_type_id: Telița illustrative per D30; Hîrbovăț mapped from the sheet column only (♀ feed → ft-breeder2, ♂ feed → ft-cocosi; names illustrative, D30) — the product/phase is not in the sheet (unknown)',
                      conventions='price_per_kg = total_lei / qty_kg; priced = the sheet has lei for the delivery; for_sex = which stock the kg went into (Hîrbovăț)',
                      totals=dict(telita=fd_old['meta']['totals'].get('telita', fd_old['meta']['totals']),
                                  hirbovat=dict(deliveries=len(hir_rows), qty_kg_F=hk['F'], qty_kg_M=hk['M'], total_lei=hl,
                                                unpriced=[(r['date'], r['for_sex'], r['qty_kg']) for r in unpriced],
                                                sheet_header=dict(kg_F=C(F_rows[1], 'R'), kg_M=C(F_rows[1], 'S'), lei=C(F_rows[1], 'AJ')))),
                      hirbovat_price_outliers=[dict(date=a, for_sex=b, qty_kg=c, total_lei=d, price_per_kg=e, note='lei/kg outside 7–9,5: the lei on this row may belong to another load; kept as typed') for a, b, c, d, e in price_out],
                      hirbovat_adjustments=adjust,
                      hirbovat_text_notes=[dict(date=a, col=b, text=c) for a, b, c in notes_cells],
                      notes=['Lei is typed once per invoice, usually on the delivery row (on the next day\'s row for %s); where ♀ and ♂ kg arrive on one row with one lei figure it is split pro rata by kg, unless lei ÷ (♀+♂ kg) is outside 7–9,5 lei/kg, then it is the ♀ delivery\'s (lei_basis says which).' % ', '.join(sorted({d_['date'] for d_ in deliv_rows if d_['basis'] and 'next' in d_['basis']})),
                             '17.02.2026: 2 000 kg ♀ + 1 000 kg ♂ before the birds arrived, no lei (unpriced, "fără preț").',
                             'Text notes in AK/AL ("norp", "intr", "ng", "fara") are not explained; "norp" may be norponin as at Telița. Not mapped to a feed type.']),
            rows=tel_rows + hir_rows)
dump('feed-deliveries.json', fdel)
log('Hîrbovăț deliveries', len(hir_rows), hk, 'lei', hl, 'unpriced', [(r['date'], r['qty_kg']) for r in unpriced])

# ---------------------------------------------------------------- illustrative-events-hirbovat.json
def secday(ds, s, x): return state[ds]['sec'][(s, x)]
cull_plan = [('2026-04', 'boală'), ('2026-06', 'șchiopătare'), ('2026-08', 'boală'), ('2026-10', 'altul')]
culls = []
for month, reason in cull_plan:
    for ds in [x for x in dates if x.startswith(month) and x >= month + '-10']:
        best = max(((secday(ds, s, x)['mort'], s, x) for s in SECTIONS for x in 'FM' if not secday(ds, s, x)['sold']), default=None)
        if best and best[0] >= 3:
            m, s, x = best; q = 2 if m >= 4 else 1
            culls.append(dict(date=ds, sapt=sapt(ds), section=s, sex=x, qty=q, reason=reason,
                              note='pasăre cu prolaps, scoasă din efectiv' if reason == 'altul' else None,
                              carved_from_mortality_of_day=m, mortality_left=m - q, created_by='Dr. Ceban', created_at=f'{ds}T08:{20 + len(culls)}'))
            break
# a recent ♂ cull (no sale that compartment-day)
for ds in ('2026-11-17', '2026-11-16', '2026-11-11'):
    c = [(secday(ds, s, 'M')['mort'], s) for s in SECTIONS if not secday(ds, s, 'M')['sold'] and secday(ds, s, 'M')['mort'] >= 1]
    if c:
        m, s = max(c)
        culls.append(dict(date=ds, sapt=sapt(ds), section=s, sex='M', qty=1, reason='șchiopătare', note=None,
                          carved_from_mortality_of_day=m, mortality_left=m - 1, created_by='Dr. Ceban', created_at=f'{ds}T08:31'))
        break
sex_err = None
for ds in ('2026-05-19', '2026-05-20', '2026-05-21'):
    c = [(secday(ds, s, 'F')['mort'], s) for s in SECTIONS if secday(ds, s, 'F')['mort'] >= 2]
    if c:
        m, s = max(c)
        sex_err = dict(date=ds, sapt=sapt(ds), section=s, sex='F', qty=-1, note='cocoș găsit printre găini, scos din efectiv',
                       carved_from_mortality_of_day=m, mortality_left=m - 1, created_by='Dr. Ceban', created_at=f'{ds}T08:40')
        break
log_days = ['2026-11-16', '2026-11-17', '2026-11-18']
events = []
times = {'2026-11-16': ('08:12', '08:25', '08:41', '09:05'), '2026-11-17': ('08:07', '08:19', '08:36', '08:58'), '2026-11-18': ('08:15', '08:27', '08:44', '09:10')}
for ds in log_days:
    f = state[ds]['farm']; t = times[ds]
    secs_with = [s for s in SECTIONS if secday(ds, s, 'F')['alive'] + secday(ds, s, 'M')['alive'] + secday(ds, s, 'F')['mort'] + secday(ds, s, 'M')['mort'] > 0]
    events.append(dict(at=f'{ds}T{t[0]}', user='Dr. Ceban', role='vet', record='mortality', action='create',
                       scope=f'Hîrbovăț · {len(secs_with)} secțiuni',
                       summary=f"Zilnic salvat · {len(secs_with)} secțiuni · mortalitate ♀ {f['mF']} · ♂ {f['mM']}"))
    halls_w = [h for h in HALLS if state[ds]['hall'][h]['eggs']]
    events.append(dict(at=f'{ds}T{t[1]}', user='Dr. Ceban', role='vet', record='eggs', action='create',
                       scope='Hîrbovăț · ' + ', '.join(halls_w), summary=f"Ouă · {f['eggs']} total · {f['inc']} incubabile · {f['jos']} jos"))
    events.append(dict(at=f'{ds}T{t[2]}', user='Dr. Ceban', role='vet', record='feed', action='create', scope='Hîrbovăț · alocare zilnică',
                       summary=f"Hrană · ♀ {f['kgF']} kg · ♂ {f['kgM']} kg"))
# DATA04 (D168): one sale form = one record (one farm, one time). A day's sales in one hall are one form (18.11 H6.1 + H6.2);
# a day with sales in two halls has two forms in hall order (14.11: H6.2, then H7.1 + H7.2 — S10 parțială's "before"
# ♀ 23 625 · ♂ 1 924 already lacks H6.2's sale). Quantities real (D155); the time (illustrative) = the latest old per-row
# slot of its rows (14.11 H7 13:34 after H6.2 12:17; 18.11 13:16), so times rise within a day.
def _slot(i): return f"{11 + i % 3:02d}:{(i * 17) % 60:02d}"
sale_forms = []
for i_, (ds, ad, sp, sec, sF, sM) in enumerate(sales_rows):
    if sale_forms and sale_forms[-1]['date'] == ds and sale_forms[-1]['hall'] == HALL_OF[sec]:
        fm_ = sale_forms[-1]
    else:
        fm_ = dict(date=ds, hall=HALL_OF[sec], rows=[]); sale_forms.append(fm_)
    fm_['rows'].append(dict(section=sec, F=sF, M=sM)); fm_['time'] = max(fm_.get('time', ''), _slot(i_))
for d_ in sorted({f_['date'] for f_ in sale_forms}):
    ts = [f_['time'] for f_ in sale_forms if f_['date'] == d_]
    assert ts == sorted(ts) and len(set(ts)) == len(ts), (d_, ts)   # forms of one day: hall order = time order
assert [[r['section'] for r in f_['rows']] for f_ in sale_forms if f_['date'] in ('2026-11-14', '2026-11-18')] == [['H6.2'], ['H7.1', 'H7.2'], ['H6.1', 'H6.2']]
for f_ in sale_forms:
    secs_ = [r['section'] for r in f_['rows']]; tF = sum(r['F'] for r in f_['rows']); tM = sum(r['M'] for r in f_['rows'])
    events.append(dict(at=f"{f_['date']}T{f_['time']}", user='Ion Rusu', role='manager', record='sale', action='create',
                       scope='Hîrbovăț · ' + ', '.join(secs_), summary='Vânzare · ' + ', '.join(secs_) + (f" · ♀ {tF}" if tF else '') + (f" · ♂ {tM}" if tM else ''),
                       real_figures=True, sections=f_['rows'], F=tF, M=tM))
log('sale records (D168)', len(sale_forms), '·', '; '.join(f"{f_['date'][5:]} {f_['time']} " + '+'.join(r['section'] for r in f_['rows']) for f_ in sale_forms))
dl = [r for r in hir_rows if r['date'] >= '2026-11-12']
for r in dl:
    events.append(dict(at=f"{r['date']}T10:30", user='Ion Rusu', role='manager', record='feed_delivery', action='create',
                       scope='Hîrbovăț', summary=f"Livrare · {r['qty_kg']} kg · {'găini' if r['for_sex'] == 'F' else 'cocoși'}" + (f" · {r['total_lei']} lei" if r['total_lei'] else ' · fără preț')))
for m in internal_moves:
    events.append(dict(at=f"{m['date']}T09:20", user='Ion Rusu', role='manager', record='transfer', action='create',
                       scope=f"Hîrbovăț · {m['from']} → {m['to']}", summary=f"Transfer · {m['qty']} {'găini' if m['sex'] == 'F' else 'cocoși'} · {m['from']} → {m['to']}"))
events.sort(key=lambda e: e['at'])
_form_at = {(f_['date'], r['section']): f"{f_['date']}T{f_['time']}" for f_ in sale_forms for r in f_['rows']}
sales_ev = [dict(date=r[0], sapt=r[2], section=r[3], F=r[4], M=r[5], real=True, created_by='Ion Rusu', created_at=_form_at[(r[0], r[3])]) for r in sales_rows]  # created_at = its record's `at` (D168)
weighings = [dict(date='2026-11-10', sapt=sapt('2026-11-10'), section='H5.1', sex='F', avg_weight_g=3960, cv_pct=8.1, uniformity_pct=79.0,
                  sample_size=60, guide_weight_g=None, guide_state='fără date în ghid', created_by='Dr. Ceban', created_at='2026-11-10T09:15'),
             dict(date='2026-11-10', sapt=sapt('2026-11-10'), section='H5.1', sex='M', avg_weight_g=4870, cv_pct=9.4, uniformity_pct=74.0,
                  sample_size=40, guide_weight_g=None, guide_state='fără date în ghid', created_by='Dr. Ceban', created_at='2026-11-10T09:32')]
dump('illustrative-events-hirbovat.json', dict(
    meta=dict(generated_by='DATA03', script=GEN['script'], illustrative=True,
              label='ILLUSTRATIVE — records the workbook does not hold (D152): culls with reasons, a sexing error, event authors/times, weighings. Not workbook data.',
              consistency='Every cull / sexing error is carved out of the real deaths of that compartment-day (carved_from_mortality_of_day); daily totals and alive counts stay the real ones. Event-log figures are the real daily totals of hirbovat-daily.json. sales = the real Vânzare quantities (D155, hirbovat-daily sales) per compartment; only their author/time are illustrative. One sale form = one event_log record (D168, DATA04): record=sale lists its compartments in `sections` [{section, F, M}] with totals F/M (= Σ its sales rows); the compartments of one hall sold on one day are one form, created_at of each sales row = its record\'s at; 14.11 has two forms (H6.2, then H7.1 + H7.2).',
              not_included='No transfer between Hîrbovăț halls: the sheet has none (not invented); the one real move between compartments (22.10.2026 H7.2 → H7.1 ♂ 19) is in transfers-seria2026.json, its event_log entry below gives it an illustrative author/time. Sales 12–18.11 are real quantities (D155), listed under sales with illustrative author/time.',
              names='Dr. Ceban (vet), Ion Rusu (manager) — illustrative, as before (D27 sample).',
              replaces='sample-hirbovat-2026.json (superseded; kept for reference)'),
    culls=culls, sexing_errors=[sex_err] if sex_err else [], sales=sales_ev, weighings=weighings, event_log=events))

# ---------------------------------------------------------------- key-figures-18-11.md
f = state[TODAY]['farm']; fy = state['2026-11-17']['farm']
def gr(x):
    gs = sorted({secday(TODAY, s, x)['g'] for s in SECTIONS if secday(TODAY, s, x)['g'] and secday(TODAY, s, x)['start']})
    return f'{gs[0]}' if len(gs) == 1 else f'{gs[0]}–{gs[-1]}'
def n(x):
    if x is None: return '–'
    if isinstance(x, float):
        s = f'{x:,.1f}' if abs(x - round(x)) > 1e-9 else f'{int(round(x)):,}'
    else:
        s = f'{x:,}'
    return s.replace(',', ' ').replace('.', ',')
lines = ['# Key figures · Hîrbovăț Seria 2026 · today 18.11.2026 (real, Raport_NOU_2026.xlsx)', '',
         f"- Age {age_days(TODAY)} days (UI \"Vârstă (zile)\" = age_days; workbook Zile 408) · **SAPT {sapt(TODAY)}** (D151; placed 07.10.2025, age 0). Week 59 = 17–23.11; weeks run Tue–Mon.",
         '- Cycles: Telița 07.10.2025–28.02.2026 (SAPT 1–21, closed) · Hîrbovăț from 22.02.2026 (SAPT 20; ♂ 22.02, ♀ 25–28.02), open, emptying since 12.11.',
         f"- Arrived: ♀ {n(arr_tot['F'])} · ♂ {n(arr_tot['M'])}.",
         f"- **Alive end 18.11: ♀ {n(f['aF'])} · ♂ {n(f['aM'])} · total {n(f['aF'] + f['aM'])}** (17.11: ♀ {n(fy['aF'])} · ♂ {n(fy['aM'])}).",
         '- Per hall / section (♀ · ♂):']
for h in HALLS:
    hh = state[TODAY]['hall'][h]
    parts = ' · '.join(f"{s} {n(secday(TODAY, s, 'F')['alive'])} · {n(secday(TODAY, s, 'M')['alive'])}" for s in SECTIONS if HALL_OF[s] == h)
    lines.append(f"  - {h}: ♀ {n(hh['aF'])} · ♂ {n(hh['aM'])} ({parts})" + (' — empty since 14.11' if h == 'H7' else ''))
sale_today = ' · '.join(f"{r[3]} ♀ {n(r[4])} · ♂ {n(r[5])}" for r in sales_rows if r[0] == TODAY)
lines += [f"- Mortality 18.11 (real deaths, D155): ♀ {n(f['mF'])} · ♂ {n(f['mM'])}. Vânzare 18.11: ♀ {n(f['dpF'])} · ♂ {n(f['dpM'])} ({sale_today}). Sheet books ♀ {n(f['mF'] + f['dpF'])} · ♂ {n(f['mM'] + f['dpM'])} as mortality.",
          f"- Cycle: deaths ♀ {n(tot['F']['mortality'])} ({n(pct(tot['F']['mortality'], arr_tot['F']))} %) · ♂ {n(tot['M']['mortality'])} ({n(pct(tot['M']['mortality'], arr_tot['M']))} %); sold 12–18.11 ♀ {n(tot['F']['sold'])} · ♂ {n(tot['M']['sold'])}. Cycle open (closes when sales reach 0).",
          f"- Eggs 18.11: {n(f['eggs'])} total · {n(f['inc'])} incubabile ({n(r1(f['incp']))} %) · {n(f['jos'])} jos · H5 {n(state[TODAY]['hall']['H5']['eggs'])} · H6 {n(state[TODAY]['hall']['H6']['eggs'])} · H7 –.",
          f"- **Lay 18.11: {n(r1(f['lay']))} %** (eggs ÷ ♀ at start of day, D156) vs guide SAPT 59 {n(real_ghid[59]['lay'])} % (RealGhid). Sheet shows {n(r1(f['lay_sheet']))} % (end-of-day count).",
          f"- Egg weight 18.11: H5 {n(r1(state[TODAY]['hall']['H5']['ew']))} g · H6 {n(r1(state[TODAY]['hall']['H6']['ew']))} g · farm {n(r1(f['ew']))} g (sheet farm {n(r1(f['ew_sheet']))} = plain mean incl. empty H7).",
          f"- Feed 18.11: ♀ {n(f['kgF'])} kg · ♂ {n(f['kgM'])} kg; norm ♀ {gr('F')} g · ♂ {gr('M')} g vs guide SAPT 59 ♀ {feed_guide[59][0]} · ♂ {feed_guide[59][1]} g.",
          f"- Stock end 18.11: ♀ feed {n(f['stF'])} kg · ♂ feed {n(f['stM'])} kg. Last deliveries 12.11 9 000 kg ♀ (fără preț) · 13.11 25 640 kg ♀ ({n([r for r in hir_rows if r['date'] == '2026-11-13'][0]['total_lei'])} lei).",
          f"- Deliveries Hîrbovăț 17.02–13.11: {len(hir_rows)} · ♀ {n(hk['F'])} kg · ♂ {n(hk['M'])} kg · {n(hl)} lei · {len(unpriced)} unpriced.",
          '- Last weeks (farm; lay = eggs ÷ start-of-day hen-days, D159; Productivitate end-of-day basis in brackets; SAPT 59 = 17–18.11 only):',
          '  | SAPT | dates | eggs | inc % | lay % (prod) | guide | deaths ♀/♂ | sold ♀/♂ | alive end ♀/♂ | feed kg ♀/♂ |',
          '  |---|---|---|---|---|---|---|---|---|---|']
for wk in [x for x in weeks if x['sapt'] in (56, 57, 58, 59)]:
    fm = wk['farm']; g = wk['guide']
    lines.append(f"  | {wk['sapt']} | {wk['from'][8:10]}.{wk['from'][5:7]}–{wk['to'][8:10]}.{wk['to'][5:7]} | {n(fm['eggs'])} | {n(pct(fm['inc'], fm['eggs'], 1))} | {n(r1(fm['lay_pct']))} ({n(r1(fm.get('lay_pct_prod')))}) | {n(g['lay_pct'])} | {n(fm['mort_F'])}/{n(fm['mort_M'])} | {n(fm['sold_F'])}/{n(fm['sold_M'])} | {n(fm['alive_F_end'])}/{n(fm['alive_M_end'])} | {n(fm['kg_F'])}/{n(fm['kg_M'])} |")
pk = max((x for x in weeks if x['farm']['lay_pct']), key=lambda x: x['farm']['lay_pct'])
wl = {x['sapt']: x['farm']['lay_pct'] for x in weeks}
lines += [f"- Peak lay: SAPT {pk['sapt']} ({n(r1(pk['farm']['lay_pct']))} %, {pk['from'][8:10]}.{pk['from'][5:7]}–{pk['to'][8:10]}.{pk['to'][5:7]}); then SAPT 55 {n(r1(wl[55]))} % → 56 {n(r1(wl[56]))} % → 57 {n(r1(wl[57]))} % → 58 {n(r1(wl[58]))} %.",
          '- Not in the workbook (illustrative only, illustrative-events-hirbovat.json): culls + reasons, sexing errors, weighings, event authors/times.',
          '- Weighings (D157): label = age week, compare with guide week SAPT − 1 (e.g. 24.02.2026 SAPT 21 vs guide wk 20 2 370 g).',
          '- Sources: hirbovat-daily.json · hirbovat-weekly.json · counts-latest.json · feed-deliveries.json · guide-laying.json.']
open(os.path.join(OUT, 'key-figures-18-11.md'), 'w', encoding='utf-8').write('\n'.join(lines) + '\n')
log('key-figures lines', len(lines))
