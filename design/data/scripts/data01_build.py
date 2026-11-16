"""DATA01: build JSON extracts from Seria_2026.xlsx.
Usage: python3 -I data01_build.py <workbook.xlsx> <out_dir>
Prints a check log (totals, anomalies) to stdout.
"""
import sys, json, datetime, os
import openpyxl
from openpyxl.utils import column_index_from_string as ci

WB, OUT = sys.argv[1], sys.argv[2]
os.makedirs(OUT, exist_ok=True)
wb = openpyxl.load_workbook(WB, data_only=True)
LOG = []
def log(*a):
    s = ' '.join(str(x) for x in a); LOG.append(s); print(s)

def iso(d): return d.strftime('%Y-%m-%d')
def num(v):
    if v is None: return None
    if isinstance(v, str):
        v = v.strip()
        if v == '': return None
        try: v = float(v.replace(',', '.'))
        except ValueError: return None
    if isinstance(v, (int, float)):
        f = float(v)
        return int(f) if f.is_integer() else round(f, 4)
    return None
def cell(ws, r, col): return ws.cell(r, ci(col)).value

PLACEMENT = datetime.datetime(2025, 10, 7)
LAST_REAL = '2026-02-28'
def age_day(d): return (d - PLACEMENT).days + 1
def age_week(d): return (d - PLACEMENT).days // 7 + 1

def dump(name, obj):
    """JSON with one row per line inside top-level lists of dicts."""
    def enc(o, ind=0):
        pad = '  ' * ind
        if isinstance(o, dict):
            if not o: return '{}'
            items = [f'{pad}  {json.dumps(k, ensure_ascii=False)}: {enc(v, ind + 1)}' for k, v in o.items()]
            return '{\n' + ',\n'.join(items) + '\n' + pad + '}'
        if isinstance(o, list) and o and all(isinstance(x, dict) for x in o):
            return '[\n' + ',\n'.join(pad + '  ' + json.dumps(x, ensure_ascii=False) for x in o) + '\n' + pad + ']'
        return json.dumps(o, ensure_ascii=False)
    with open(os.path.join(OUT, name), 'w', encoding='utf-8') as f:
        f.write(enc(obj) + '\n')
    json.load(open(os.path.join(OUT, name), encoding='utf-8'))  # validate

GEN = {'generated_by': 'DATA01', 'source': 'design/data/Seria_2026.xlsx', 'workbook_last_real_date': LAST_REAL}

# ---------------------------------------------------------------- sections
TEL = ['H1', 'H2', 'H3', 'H6', 'H5-1', 'H5-2']
TEL_SEX = {'H1': 'F', 'H2': 'F', 'H3': 'F', 'H6': 'F', 'H5-1': 'M', 'H5-2': 'M'}
HIR = ['H1.1', 'H1.2', 'H3.1', 'H3.2', 'H4.1', 'H4.2']
SHEET9_COLS = {'H1.1': ('D', 'E'), 'H1.2': ('F', 'G'), 'H3.1': ('H', 'I'), 'H3.2': ('J', 'K'), 'H4.1': ('L', 'M'), 'H4.2': ('N', 'O')}
SHEET_NAME_LINK = {'H1': ['H4.1', 'H4.2'], 'H3': ['H1.1', 'H1.2'], 'H6': ['H3.1', 'H3.2']}

sections = []
for c in TEL:
    s = {'id': f'tel-{c}', 'farm_id': 'telita', 'code': c, 'parent_id': None, 'active': True,
         'sex_seria2026': TEL_SEX[c], 'workbook_labels': [c, c.replace('-', '/')]}
    if c in SHEET_NAME_LINK: s['to_hirbovat_by_sheet_name'] = SHEET_NAME_LINK[c]
    sections.append(s)
for hall in ['H1', 'H3', 'H4']:
    n = hall[1]
    sections.append({'id': f'hir-{hall}', 'farm_id': 'hirbovat', 'code': hall, 'parent_id': None, 'active': True,
                     'is_hall': True, 'workbook_labels': [f'Hala {n}']})
    for k in ('1', '2'):
        code = f'{hall}.{k}'
        sections.append({'id': f'hir-{code}', 'farm_id': 'hirbovat', 'code': code, 'parent_id': f'hir-{hall}',
                         'active': True, 'sexes': ['F', 'M'],
                         'workbook_labels': [code, f'G{n}{k}', f'C{n}{k}']})
dump('farms-sections.json', {
    'meta': {**GEN, 'real': 'farm names, section codes and hall grouping are from the workbook (Telita row 2, Hrana Hîrbovăț row 1, Sheet9 row 2, import date row 1)',
             'illustrative': 'guide names (D22) and ids',
             'notes': ['Telița sections hold one sex each in Seria 2026 (H1, H2, H3, H6 hens; H5-1, H5-2 roosters).',
                       'Hîrbovăț compartments hold both sexes; eggs are counted per hall (parent).',
                       'to_hirbovat_by_sheet_name comes from the H-sheet names; the dated movements pair differently (see transfers-seria2026.json).',
                       'Hrana Hîrbovăț row 2 labels its norm columns H5.1G … H7.2C (wrong; they sit under H1.1 … H4.2).']},
    'guides': [{'id': 'g-growth', 'name': 'Ross 308 creștere', 'type': 'growth', 'illustrative_name': True},
               {'id': 'g-laying', 'name': 'Ross 308 ouat', 'type': 'laying', 'illustrative_name': True}],
    'farms': [{'id': 'telita', 'name': 'Telița', 'type': 'growth', 'guide_id': 'g-growth'},
              {'id': 'hirbovat', 'name': 'Hîrbovăț', 'type': 'laying', 'guide_id': 'g-laying'}],
    'sections': sections})

# ---------------------------------------------------------------- Telita daily
ws = wb['Telita']
MORT = dict(zip(TEL, 'DEFGHI')); ALIVE = dict(zip(TEL, 'JKLMNO'))
CORR = dict(zip(TEL, 'RSTUVW')); SEXERR = dict(zip(TEL, ['X', 'Y', 'Z', 'AA', 'AB', 'AC']))
place = {c: num(cell(ws, 3, MORT[c])) or 0 for c in TEL}
log('placement Telita row3:', place, 'P/Q totals:', cell(ws, 3, 'P'), cell(ws, 3, 'Q'))

days = []
for r in range(4, ws.max_row + 1):
    d = cell(ws, r, 'C')
    if not isinstance(d, datetime.datetime): continue
    if iso(d) > LAST_REAL: break
    days.append((r, d))

running = dict(place)
mort_rows, day_tot, anomalies, corr_events = [], [], [], []
for r, d in days:
    tot = {'F': {'mortality': 0, 'alive': 0}, 'M': {'mortality': 0, 'alive': 0}}
    for c in TEL:
        m = num(cell(ws, r, MORT[c])) or 0
        a = num(cell(ws, r, ALIVE[c]))
        k = num(cell(ws, r, CORR[c])) or 0
        e = num(cell(ws, r, SEXERR[c])) or 0
        running[c] = running[c] - m + k + e
        if k: corr_events.append({'date': iso(d), 'section': c, 'qty': k})
        if a != running[c]:
            anomalies.append({'date': iso(d), 'section': c, 'alive_sheet': a, 'alive_calc': running[c]})
        populated = (a or 0) != 0 or m or k or e or running[c] != 0
        if populated:
            row = {'date': iso(d), 'day': age_day(d), 'age_week': age_week(d), 'section': c, 'sex': TEL_SEX[c],
                   'mortality': m, 'sexing_error': e, 'correction': k, 'alive_sheet': a, 'alive_calc': running[c]}
            mort_rows.append(row)
        tot[TEL_SEX[c]]['mortality'] += m
        tot[TEL_SEX[c]]['alive'] += (a or 0)
    day_tot.append({'date': iso(d), 'day': age_day(d), 'age_week': age_week(d),
                    'mortality_F': tot['F']['mortality'], 'mortality_M': tot['M']['mortality'],
                    'alive_F_sheet': num(cell(ws, r, 'Q')), 'alive_M_sheet': num(cell(ws, r, 'P'))})
log('Telita alive sheet vs calc mismatches:', anomalies)
neg = [x for x in mort_rows if (x['alive_sheet'] or 0) < 0 or x['alive_calc'] < 0]
neg_first = {}
for x in neg: neg_first.setdefault(x['section'], x)
log('negative counts (first day per section):', list(neg_first.values()))
end = {c: running[c] for c in TEL}
log('Telita calc end 28.02:', end)
tm = sum(x['mortality_F'] for x in day_tot); tmm = sum(x['mortality_M'] for x in day_tot)
se = sum(x['sexing_error'] for x in mort_rows)
log('Telita mortality total F', tm, 'M', tmm, 'sexing errors', se)

dump('mortality-telita-daily.json', {
    'meta': {**GEN, 'sheet': 'Telita', 'columns': 'D–I mortality, J–O alive ("Pui la zi"), R–W Corectie, X–AC Erori sexare, P/Q totals',
             'range': ['2025-10-07', LAST_REAL], 'real': True,
             'conventions': 'one row per populated section-day; mortality and sexing_error are counts (blank cell = 0); correction is the signed Corectie value (moves between sections and transfers out); alive_calc = placement − mortality + correction + sexing_error, running; alive_sheet as typed in the workbook',
             'totals': {'mortality_F': tm, 'mortality_M': tmm, 'sexing_errors': se},
             'anomalies': {
                 'alive_sheet_differs_from_calc': anomalies,
                 'negative_or_residual_end_counts_2026_02_28': end,
                 'notes': ['Only one sexing error in the sheet (H5-2, 2026-02-23, −1).',
                           'H6 on 2026-02-25: 5 771 − 5 783 transferred = −12, the sheet types 0.',
                           'H5-2 on 2026-02-23: 1 bird left, 2 dead + 1 sexing error recorded → −2; the sheet shows 0, so H5-2 stays 2 above the running sum afterwards (244 vs 242).',
                           'sexing_error is signed as typed (−1 = one bird removed).',
                           'H3 goes to −9 on 2026-02-26 and stays there (setup-check; D25 treats leftovers as data errors).',
                           'On 2026-02-23 277 birds move from H6 (hens) into H5-1 (roosters) via Corectie; sex as recorded per section.']}},
    'daily_totals': day_tot,
    'rows': mort_rows})

# ---------------------------------------------------------------- placement
hs = {'H1': 'H1 ( h41,h42)', 'H2': 'H2', 'H3': 'H3 ( h11,h12)', 'H6': 'H6( h31,h32)', 'H5-1': 'H51.', 'H5-2': 'H52'}
hrow4 = {}
for c, sname in hs.items():
    w = wb[sname]; dd = cell(w, 4, 'D')
    hrow4[c] = {'sheet': sname, 'date': iso(dd) if isinstance(dd, datetime.datetime) else None, 'qty': num(cell(w, 4, 'F'))}
prow = [{'date': '2025-10-07', 'farm_id': 'telita', 'section': c, 'sex': TEL_SEX[c], 'qty': place[c]} for c in TEL if place[c]]
ptot = {'F': sum(x['qty'] for x in prow if x['sex'] == 'F'), 'M': sum(x['qty'] for x in prow if x['sex'] == 'M')}
log('placement rows', prow, ptot)
dump('placement-seria2026.json', {
    'meta': {**GEN, 'sheet': 'Telita row 3 (confirmed by H1, H2, H51. row 4)', 'real': True,
             'series': 'Seria 2026', 'date': '2025-10-07',
             'notes': ['H3, H6 and H5-2 had no placement; they were filled by moves between sections (Corectie) from 2025-11-05, 2025-12-16 and 2025-11-11.',
                       'H-sheet row 4 for H3 (8 264, dated 2024-10-07), H6 (1, dated 2024-11-01) and H52 (1 855) is not a placement; kept below for reference only.']},
    'totals': ptot,
    'rows': prow,
    'h_sheet_row4_reference': [{'section': c, **v} for c, v in hrow4.items()]})

# ---------------------------------------------------------------- transfers
s9 = wb['Sheet9']
prev = {(s, x): 0 for s in HIR for x in 'FM'}
arrivals = []
for r in range(3, s9.max_row + 1):
    d = cell(s9, r, 'A')
    if not isinstance(d, datetime.datetime): continue
    for s in HIR:
        for x, col in zip('FM', SHEET9_COLS[s]):
            v = num(cell(s9, r, col)) or 0
            if v != prev[(s, x)]:
                if iso(d) > LAST_REAL or v < prev[(s, x)]:
                    anomalies.append({'sheet9_change': iso(d), 'section': s, 'sex': x, 'from': prev[(s, x)], 'to': v})
                else:
                    arrivals.append({'date': iso(d), 'farm_id': 'hirbovat', 'section': s, 'sex': x, 'qty': v - prev[(s, x)]})
                prev[(s, x)] = v
arr_counts = {s: {'F': prev[(s, 'F')], 'M': prev[(s, 'M')]} for s in HIR}
arr_tot = {'F': sum(v['F'] for v in arr_counts.values()), 'M': sum(v['M'] for v in arr_counts.values())}
log('Sheet9 arrivals', arrivals); log('arrival counts', arr_counts, arr_tot)
if arr_tot != {'F': 26765, 'M': 3060}: log('!! arrival totals differ from setup-check 26765/3060')

# classify Telita corrections per date
by_date = {}
for e in corr_events: by_date.setdefault(e['date'], []).append(e)
internal, outs, unmatched = [], [], []
arr_by_date = {}
for a in arrivals: arr_by_date.setdefault((a['date'], a['sex']), 0); arr_by_date[(a['date'], a['sex'])] += a['qty']
for dt, evs in sorted(by_date.items()):
    pos = [e for e in evs if e['qty'] > 0]; negs = [e for e in evs if e['qty'] < 0]
    used = set()
    # internal moves: a negative and a positive of equal size the same day
    for n_ in negs:
        for p in pos:
            if id(p) in used: continue
            if p['qty'] == -n_['qty']:
                internal.append({'date': dt, 'from': n_['section'], 'to': p['section'], 'sex_from': TEL_SEX[n_['section']], 'sex_to': TEL_SEX[p['section']], 'qty': p['qty']})
                used.add(id(p)); used.add(id(n_)); break
    # one negative split over several positives (e.g. 2025-11-24 H1 −100, H2 −200 → H3 +300)
    rest_n = [e for e in negs if id(e) not in used]; rest_p = [e for e in pos if id(e) not in used]
    if rest_n and rest_p and sum(e['qty'] for e in rest_n) == -sum(e['qty'] for e in rest_p) and len(rest_p) == 1:
        for n_ in rest_n:
            internal.append({'date': dt, 'from': n_['section'], 'to': rest_p[0]['section'], 'sex_from': TEL_SEX[n_['section']], 'sex_to': TEL_SEX[rest_p[0]['section']], 'qty': -n_['qty']})
            used.add(id(n_))
        used.add(id(rest_p[0]))
    for e in evs:
        if id(e) in used: continue
        if e['qty'] < 0 and dt >= '2026-02-22' and abs(e['qty']) > 1000:
            outs.append({'date': dt, 'from_farm': 'telita', 'from_section': e['section'], 'sex': TEL_SEX[e['section']], 'qty': -e['qty']})
        else:
            unmatched.append({'date': dt, 'section': e['section'], 'sex': TEL_SEX[e['section']], 'qty': e['qty'],
                              'kind': 'outflow_not_at_hirbovat' if e['qty'] < 0 else 'inflow_unexplained'})
out_tot = {'F': sum(o['qty'] for o in outs if o['sex'] == 'F'), 'M': sum(o['qty'] for o in outs if o['sex'] == 'M')}
log('internal moves', internal); log('transfers out', outs, out_tot); log('unmatched corrections', unmatched)

# date-matched links (inferred)
links = [
    {'out_date': '2026-02-22', 'from': ['H5-1', 'H5-2'], 'sex': 'M', 'qty_out': 2994, 'arrival_date': '2026-02-22',
     'to': {s: arr_counts[s]['M'] for s in HIR}, 'qty_in': arr_tot['M'], 'difference_in_minus_out': arr_tot['M'] - 2994},
]
def arrived(date, sex): return {a['section']: a['qty'] for a in arrivals if a['date'] == date and a['sex'] == sex}
for o_date, sec, a_date in [('2026-02-25', 'H6', '2026-02-25'), ('2026-02-26', 'H3', '2026-02-26'),
                            ('2026-02-27', 'H1', '2026-02-27'), ('2026-02-27', 'H2', '2026-02-28')]:
    q = sum(o['qty'] for o in outs if o['date'] == o_date and o['from_section'] == sec)
    to = arrived(a_date, 'F'); qi = sum(to.values())
    links.append({'out_date': o_date, 'from': [sec], 'sex': 'F', 'qty_out': q, 'arrival_date': a_date, 'to': to, 'qty_in': qi,
                  'difference_in_minus_out': qi - q})
for l in links: log('link', l)

dump('transfers-seria2026.json', {
    'meta': {**GEN, 'sheets': 'Telita "Corectie" cols R–W (out side); Sheet9 G11…C42 first change per section (in side)',
             'range': ['2025-11-05', LAST_REAL], 'real': True,
             'conventions': 'qty positive; sex F/M by Telița section; arrivals are the day-on-day increase of the forward-filled Sheet9 grid',
             'design_dates': 'D21/D25: the screens start the Hîrbovăț cycle on 2026-02-10 and show one transfer link that day; the real movements below are 2026-02-22 … 2026-02-28.',
             'reconciliation': {
                 'out_of_telita': out_tot, 'arrived_hirbovat': arr_tot,
                 'difference_arrived_minus_out': {'F': arr_tot['F'] - out_tot['F'], 'M': arr_tot['M'] - out_tot['M']},
                 'notes': ['The kit sample ♀ 26 881 · ♂ 2 994 equals the Telița-side transfer outs; Hîrbovăț recorded ♀ 26 765 · ♂ 3 060 (D28 uses the arrivals).',
                           '♀ gap of 116 is all on H2 (6 999 out on 2026-02-27, 6 883 in on 2026-02-28).',
                           '♂: 66 more arrived than left on 2026-02-22; Telița then still shows 698 ♂ (H5-1 454, H5-2 244), partly refilled by unexplained +11/+277/+144/+16/+14 corrections.',
                           'Pairing by date (links) contradicts the H-sheet names (H1→H4.x, H3→H1.x, H6→H3.x): dated movements put H6→H1.1/H1.2, H3→H1.2/H3.1, H1→H3.1/H3.2/H4.1, H2→H4.1/H4.2.']}},
    'transfers_out': outs,
    'arrivals': arrivals,
    'arrival_counts': [{'section': s, 'F': v['F'], 'M': v['M']} for s, v in arr_counts.items()],
    'arrival_totals': arr_tot,
    'links_inferred_by_date': links,
    'internal_moves_telita': internal,
    'other_corrections_telita': unmatched})

# ---------------------------------------------------------------- feed (Hrana Telita)
ht = wb['Hrana Telita']
NORM = dict(zip(TEL, 'FGHIJK')); KG = dict(zip(TEL, 'NOPQRS'))
alive_by = {(x['date'], x['section']): x['alive_sheet'] or 0 for x in mort_rows}
# feed types per D30
TYPES = [
    {'id': 'ft-starter', 'name': 'Starter', 'for_sex': None, 'farm_id': 'telita', 'active': False, 'workbook_note': 'rost', 'note_date': '2025-10-24', 'illustrative_name': True},
    {'id': 'ft-grower', 'name': 'Grower', 'for_sex': None, 'farm_id': 'telita', 'active': False, 'workbook_note': 'razvitie', 'note_date': '2025-11-14', 'illustrative_name': True,
     'also_covers_note': 'norponin (2025-11-29; not a phase name, not in D30; the delivery is kept on Grower)'},
    {'id': 'ft-preouat', 'name': 'Pre-ouat', 'for_sex': None, 'farm_id': 'telita', 'active': False, 'workbook_note': 'preouat', 'note_date': '2026-01-20', 'illustrative_name': True},
    {'id': 'ft-breeder2', 'name': 'Breeder 2', 'for_sex': 'F', 'farm_id': 'hirbovat', 'active': True, 'workbook_note': None, 'note_date': None, 'illustrative_name': True},
    {'id': 'ft-cocosi', 'name': 'Cocoși', 'for_sex': 'M', 'farm_id': 'hirbovat', 'active': True, 'workbook_note': None, 'note_date': None, 'illustrative_name': True},
]
def type_for(date):
    if date >= '2026-01-20': return 'ft-preouat'
    if date >= '2025-11-14': return 'ft-grower'
    return 'ft-starter'
deliv, feed_days, guide_tel = [], [], {}
for r in range(3, ht.max_row + 1):
    d = cell(ht, r, 'C')
    if not isinstance(d, datetime.datetime): continue
    ds = iso(d)
    wk = num(cell(ht, r, 'A'))
    if wk is not None: guide_tel[int(wk)] = {'F': num(cell(ht, r, 'D')), 'M': num(cell(ht, r, 'E')), 'date': ds}
    kg_del = num(cell(ht, r, 'L')); lei = num(cell(ht, r, 'X')); note = cell(ht, r, 'Y')
    if num(cell(ht, r, 'M')): log('!! rooster delivery', ds)
    if kg_del:
        deliv.append({'id': f'del-{ds}', 'date': ds, 'farm_id': 'telita', 'qty_kg': kg_del,
                      'workbook_note': note.strip() if isinstance(note, str) else None,
                      'feed_type_id': type_for(ds), 'priced': lei is not None,
                      'price_per_kg': round(lei / kg_del, 4) if lei else None, 'total_lei': lei})
    if ds > '2026-02-26': continue
    secs = {}
    for c in TEL:
        g = num(cell(ht, r, NORM[c])); kg = num(cell(ht, r, KG[c]))
        if g is None and not kg: continue
        secs[c] = {'sex': TEL_SEX[c], 'g_per_bird': g, 'kg': kg}
    tot = {}
    for x in 'FM':
        kgs = sum((v['kg'] or 0) for v in secs.values() if v['sex'] == x)
        w = [(v['g_per_bird'], alive_by.get((ds, c), 0)) for c, v in secs.items() if v['sex'] == x and v['g_per_bird'] is not None]
        birds = sum(b for _, b in w)
        tot[x] = {'kg': kgs, 'g_per_bird_avg': round(sum(g * b for g, b in w) / birds, 1) if birds else None}
    feed_days.append({'date': ds, 'day': age_day(d), 'age_week': age_week(d), 'sections': secs,
                      'kg_F': tot['F']['kg'], 'kg_M': tot['M']['kg'],
                      'g_per_bird_F': tot['F']['g_per_bird_avg'], 'g_per_bird_M': tot['M']['g_per_bird_avg'],
                      'kg_F_sheet': num(cell(ht, r, 'T')), 'kg_M_sheet': num(cell(ht, r, 'U')),
                      'delivery_kg': kg_del, 'stock_kg': num(cell(ht, r, 'V'))})
dk = sum(x['qty_kg'] for x in deliv); dl = sum(x['total_lei'] for x in deliv)
log('deliveries', len(deliv), dk, dl, 'sheet L2/X2:', cell(ht, 2, 'L'), cell(ht, 2, 'X'))
for x in deliv: log('  ', x)
kgsum_mismatch = [x['date'] for x in feed_days if x['kg_F'] != (x['kg_F_sheet'] or 0) or x['kg_M'] != (x['kg_M_sheet'] or 0)]
log('kg total mismatches vs T/U:', kgsum_mismatch)
negstock = [(x['date'], x['stock_kg']) for x in feed_days if (x['stock_kg'] or 0) < 0]
log('negative stock:', negstock)
cons = sum(x['kg_F'] + x['kg_M'] for x in feed_days)
log('consumption kg 07.10–26.02:', cons, 'delivered', dk, 'delivered-consumed', dk - cons)
# stock check: running delivered - consumed vs V
stock = 0; st_mis = []
for x in feed_days:
    stock += (x['delivery_kg'] or 0) - x['kg_F'] - x['kg_M']
    if x['stock_kg'] is not None and abs(stock - x['stock_kg']) > 0.5: st_mis.append((x['date'], stock, x['stock_kg']))
log('stock recompute mismatches (first 5):', st_mis[:5], len(st_mis))

ptypes = {}
for x in deliv: ptypes.setdefault(x['feed_type_id'], 0); ptypes[x['feed_type_id']] += x['qty_kg']
dump('feed-deliveries.json', {
    'meta': {**GEN, 'sheet': 'Hrana Telita: C date, L kg ("Gaina"), X lei, Y note', 'range': ['2025-10-07', '2026-02-10'],
             'real': 'date, kg, lei and the note are real; feed_type_id is illustrative per D30 (type carried forward from the latest note; the two deliveries before the first note are put on Starter)',
             'conventions': 'price_per_kg = total_lei / qty_kg, rounded to 4 decimals; priced = the workbook has lei for the row',
             'totals': {'deliveries': len(deliv), 'qty_kg': dk, 'total_lei': dl, 'priced_kg': sum(x['qty_kg'] for x in deliv if x['priced']),
                        'unpriced_kg': sum(x['qty_kg'] for x in deliv if not x['priced']), 'kg_by_type': ptypes},
             'notes': ['No deliveries to Hîrbovăț in the workbook (Hrana Hîrbovăț "Cantitate primită" is empty); D30: Hîrbovăț deliveries are sample (DATA02).',
                       'No rooster-only deliveries (col M "Cocos" empty).',
                       'Every Telița delivery carries a price; "fără preț" rows would be illustrative.']},
    'rows': deliv})

dump('feed-types.json', {
    'meta': {**GEN, 'decision': 'D30', 'illustrative': True,
             'notes': ['Workbook notes in Hrana Telita col Y: rost 2025-10-24, razvitie 2025-11-14, norponin 2025-11-29, preouat 2026-01-20 (norponin also in Telita col AD on 2025-11-27).',
                       'Telița types are fed to both sexes (for_sex null). Hîrbovăț types have no workbook source.']},
    'rows': TYPES})

dump('feed-daily-telita.json', {
    'meta': {**GEN, 'sheet': 'Hrana Telita: F–K g/bird ("Norma de consum"), N–S kg per section, T/U kg ♀/♂, L delivery, V stock ("Sold")',
             'range': ['2025-10-07', '2026-02-26'], 'grams_from': '2025-10-09', 'real': True,
             'conventions': 'g_per_bird_F/M = bird-weighted mean of the section norms (alive from the Telita sheet); kg_F/kg_M = sum of section kg; *_sheet = totals as typed',
             'anomalies': {'kg_totals_differ_from_sheet': kgsum_mismatch, 'negative_stock': negstock, 'negative_stock_note': 'consumption 2026-02-25/26 exceeds stock; "Sold" stays −716 kg for the rest of the template',
                           'stock_recompute_mismatches': len(st_mis)},
             'totals': {'kg_consumed': cons, 'kg_delivered': dk},
             'not_extracted': 'Hrana Hîrbovăț has one real row (2026-03-10, H1.1: 113 g ♀ / 98 g ♂, 508,5 kg / 48,8 kg); left for the Hîrbovăț sample (DATA02).'},
    'days': feed_days})

# ---------------------------------------------------------------- weighings + growth guide
weigh, std = [], {'F': {}, 'M': {}}
for c, sname in hs.items():
    w = wb[sname]; x = TEL_SEX[c]
    redated = c in ('H3', 'H6')
    for r in range(9, 40):
        wk = num(cell(w, r, 'A'))
        if not wk: continue
        wk = int(wk)
        k = num(cell(w, r, 'K')); fs = num(cell(w, r, 'I'))
        std[x].setdefault(wk, set()).add((k, fs))
        L = num(cell(w, r, 'L'))
        if L is None: continue
        d = PLACEMENT + datetime.timedelta(days=7 * wk)
        sheet_d = cell(w, r, 'B')
        weigh.append({'section': c, 'sex': x, 'age_week': wk, 'date': iso(d),
                      'date_in_sheet': iso(sheet_d) if isinstance(sheet_d, datetime.datetime) else None, 'redated': redated,
                      'avg_weight_g': L, 'cv_pct': num(cell(w, r, 'N')), 'uniformity_pct': num(cell(w, r, 'O')), 'sample_size': None,
                      'standard_g': k, 'diff_g': num(cell(w, r, 'M')), 'feed_standard_g': fs,
                      'mortality_week': num(cell(w, r, 'C')),
                      'mortality_cum_pct': round(num(cell(w, r, 'E')), 2) if c in ('H1', 'H2', 'H5-1') and num(cell(w, r, 'E')) is not None else None,
                      'source_sheet': sname})
# cross-check weekly mortality against the daily sheet for each weighing row
wk_m = {}
for x in mort_rows: wk_m[(x['section'], x['age_week'])] = wk_m.get((x['section'], x['age_week']), 0) + x['mortality']
wm_mis = [(w_['section'], w_['age_week'], w_['mortality_week'], wk_m.get((w_['section'], w_['age_week']))) for w_ in weigh if w_['mortality_week'] != wk_m.get((w_['section'], w_['age_week']))]
log('weighing rows', len(weigh), 'weekly mortality mismatches vs daily:', wm_mis)
std_conflicts = {x: {wk: v for wk, v in s.items() if len(v) > 1} for x, s in std.items()}
log('standard conflicts between H-sheets:', std_conflicts)
lastw = {}
for w_ in weigh: lastw[w_['section']] = max(lastw.get(w_['section'], ''), w_['date'])
log('last weighing per section:', lastw)
dump('weighings-weekly.json', {
    'meta': {**GEN, 'sheets': 'H1 ( h41,h42), H2, H3 ( h11,h12), H6( h31,h32) (hens); H51., H52 (roosters): rows 9–31',
             'range': ['2025-10-14', max(lastw.values())], 'real': True,
             'conventions': 'one row per section and age week with an actual weight (col L "Efectiv"); date = 2025-10-07 + 7 × age_week (the sheet\'s "data w/c", end of week); cv_pct = col N; uniformity_pct = col O "(+/-) 10%"; standard_g = col K; feed_standard_g = col I; sample_size not in the workbook (null)',
             'notes': ['H3 and H6 sheets carry dates from an earlier year; their weekly mortality matches the Seria 2026 daily sheet, so they are re-dated by age week (redated true, date_in_sheet kept).',
                       'The second blocks in H3/H6 (rows 70+, labelled H 1.1, H1.2, dated 2023–2024) are an older flock and are not extracted.',
                       'mortality_cum_pct only where the sheet\'s base is the real placement (H1, H2, H5-1).',
                       '"Pasari in hala" (col H) is not extracted: it ignores most moves between sections and goes negative in H6.',
                       'Actual weights end at age wk 20 for hens (2026-02-24) and wk 19 for roosters (2026-02-17).'],
             'weekly_mortality_mismatches_vs_daily': [{'section': a, 'age_week': b, 'h_sheet': c_, 'daily_sum': d_} for a, b, c_, d_ in wm_mis]},
    'rows': weigh})

growth = []
for wk in range(1, 26):
    for x in 'FM':
        g = guide_tel.get(wk, {}).get(x)
        bw = None
        if wk in std[x]:
            vals = {k for k, _ in std[x][wk] if k is not None}
            bw = sorted(vals)[0] if vals else None
            fs = {f for _, f in std[x][wk] if f is not None}
            if g is not None and fs and fs != {g}: log('!! growth feed guide differs', wk, x, g, fs)
        growth.append({'guide_id': 'g-growth', 'week': wk, 'sex': x, 'feed_g': g, 'body_weight_g': bw})
dump('guide-growth.json', {
    'meta': {**GEN, 'guide_name': 'Ross 308 creștere (illustrative name, D22)', 'week_basis': 'age week = cycle week for growth (D26)',
             'sources': 'feed_g: Hrana Telita "Ghid" cols D (♀) / E (♂), first row of each week, wk 1–25 (matches the H-sheet "Furaj standard" wk 1–23); body_weight_g: H-sheet col K "Standard" (♀ H1/H2/H3/H6, ♂ H51./H52 identical), wk 1–23',
             'real': 'figures are real workbook guide values; the guide name is illustrative',
             'missing': 'body_weight_g wk 24–25; no mortality, CV or uniformity targets (D29)'},
    'rows': growth})

# ---------------------------------------------------------------- laying guide
hh = wb['Hrana Hîrbovăț']
lay_feed, label_issues = {}, []
for r in range(3, hh.max_row + 1):
    d = cell(hh, r, 'C'); lab = num(cell(hh, r, 'A'))
    if not isinstance(d, datetime.datetime) or lab is None: continue
    wk = age_week(d)
    if wk != lab: label_issues.append({'date': iso(d), 'label': lab, 'age_week_by_date': wk})
    lay_feed[wk] = {'F': num(cell(hh, r, 'D')), 'M': num(cell(hh, r, 'E')), 'date': iso(d)}
log('laying guide weeks by date:', min(lay_feed), max(lay_feed), 'label issues:', label_issues)
lay_pct = {}
for r in range(3, s9.max_row + 1):
    d = cell(s9, r, 'A'); p = num(cell(s9, r, 'P'))
    if not isinstance(d, datetime.datetime) or p is None: continue
    lay_pct.setdefault(age_week(d), set()).add(p)
multi = {k: v for k, v in lay_pct.items() if len(v) > 1}
log('lay pct weeks', min(lay_pct), max(lay_pct), 'weeks with >1 value:', multi)
laying = []
for wk in range(19, 66):
    f = lay_feed.get(wk, {})
    lp = sorted(lay_pct[wk])[0] if wk in lay_pct else None
    for x in 'FM':
        laying.append({'guide_id': 'g-laying', 'age_week': wk, 'cycle_week': wk - 18, 'sex': x, 'feed_g': f.get(x),
                       'lay_pct': lp if x == 'F' else None, 'body_weight_g': None, 'mortality_pct': None,
                       'hatching_pct': None, 'egg_weight_g': None})
c40 = [x for x in laying if x['cycle_week'] == 40]
log('cycle wk 40:', c40)
dump('guide-laying.json', {
    'meta': {**GEN, 'guide_name': 'Ross 308 ouat (illustrative name, D22)',
             'week_basis': 'cycle_week = age_week − 18 (D26); Hîrbovăț Seria 2026 cycle from 2026-02-10 = age wk 19 = cycle wk 1; 2026-11-16 = cycle wk 40 = age wk 58',
             'sources': 'feed_g: Hrana Hîrbovăț "Ghid" cols D (G, ♀) / E (C, ♂), weekly rows 2026-02-10 … 2026-12-29 (age wk 19–65); lay_pct: Sheet9 col P (unlabelled), age wk 25–63, set on ♀ rows',
             'real': 'feed_g real workbook guide values; lay_pct ILLUSTRATIVE (inferred from an unlabelled column, D29); guide name illustrative',
             'missing': 'body weight, mortality, hatching, egg weight: no laying guide in the workbook (null, D29)',
             'anomalies': ['Col A week labels repeat at the end (63, 63, 64, 64, 65): by date the rows are age wk 63–67. Kept by date to wk 65; the wk 66–67 rows (2027-01-05, 2027-01-12; 156 g / 158 g, same as wk 62–65) are not extracted.'] + [f"label {i['label']} on {i['date']} = age wk {i['age_week_by_date']} by date" for i in label_issues]},
    'rows': laying})

# ---------------------------------------------------------------- cycles
def gw(start, d): return (datetime.datetime.fromisoformat(d) - datetime.datetime.fromisoformat(start)).days // 7 + 1
cycles = [
    {'id': 'cyc-telita-2026', 'series_id': 'seria-2026', 'farm_id': 'telita', 'guide_id': 'g-growth',
     'start_date': '2025-10-07', 'end_date': '2026-02-28', 'status': 'closed', 'started_by': 'auto',
     'weeks': gw('2025-10-07', '2026-02-28'), 'real': True,
     'notes': 'Placement 2025-10-07; transfers out 2026-02-22 … 2026-02-28. Telița never reaches 0 in the workbook (residuals 23/16/−9/0/454/244); closed 2026-02-28 per D25.'},
    {'id': 'cyc-hirbovat-2026', 'series_id': 'seria-2026', 'farm_id': 'hirbovat', 'guide_id': 'g-laying',
     'start_date': '2026-02-10', 'end_date': None, 'status': 'open', 'started_by': 'manual',
     'today': '2026-11-16', 'week_today': gw('2026-02-10', '2026-11-16'), 'age_week_today': gw('2026-02-10', '2026-11-16') + 18,
     'real': 'start date is a design decision (D21/D25); real data ends 2026-02-28 (feed row 2026-03-10); later figures are sample (D27)',
     'notes': 'An auto start would fall on 2026-02-22 (first arrival, Sheet9). 2026-02-10 = first Hrana Hîrbovăț row (age wk 19).'},
    {'id': 'cyc-hirbovat-2025', 'series_id': 'seria-2025', 'farm_id': 'hirbovat', 'guide_id': 'g-laying',
     'start_date': '2025-04-04', 'start_inferred': True, 'end_date': '2026-01-28', 'status': 'closed', 'started_by': 'auto',
     'weeks': gw('2025-04-04', '2026-01-28'), 'age_week_at_start': 23, 'real': True,
     'notes': 'Start = earliest egg/feed date in "import date" and DataHirbovat (2025-04-04), which datefinal places at age wk 23 (cycle wk 5 by D26); a start at age wk 19 would be about 2025-03-07. End = last data date. 300 data days.'},
]
for c in cycles: log('cycle', c['id'], c.get('weeks'), c.get('week_today'))
dump('cycles.json', {
    'meta': {**GEN, 'derived': True, 'guide_week': 'floor((date − start_date) / 7) + 1 (data-model)'},
    'series': [{'id': 'seria-2026', 'name': 'Seria 2026', 'year': 2026, 'status': 'active', 'illustrative_name': False},
               {'id': 'seria-2025', 'name': 'Seria 2025', 'year': 2025, 'status': 'closed', 'illustrative_name': True,
                'note': 'previous Hîrbovăț flock; the workbook does not name it'}],
    'cycles': cycles})

with open('/home/claude/scripts/data01_log.txt', 'w') as f: f.write('\n'.join(LOG))
