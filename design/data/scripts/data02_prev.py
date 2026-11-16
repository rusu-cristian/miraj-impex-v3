"""DATA02 (part 1): previous Hîrbovăț flock eggs, REAL data.
Usage: python3 -I data02_prev.py <workbook.xlsx> <out_dir>
Reads sheets "import date" (eggs per hall + mortality per compartment) and
"datefinal" (weekly roll-up, used only as a cross-check). "DataHirbovat" mirrors
"import date" cell for cell (checked) and adds only one feed row (2025-04-04).
"""
import sys, json, os, datetime
import openpyxl
from openpyxl.utils import column_index_from_string as ci

WB, OUT = sys.argv[1], sys.argv[2]
wb = openpyxl.load_workbook(WB, data_only=True)
LOG = []
def log(*a):
    s = ' '.join(str(x) for x in a); LOG.append(s); print(s)

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
    p = os.path.join(OUT, name)
    with open(p, 'w', encoding='utf-8') as f:
        f.write(enc(obj) + '\n')
    json.load(open(p, encoding='utf-8'))

def n(v):
    if v is None or isinstance(v, str): return 0
    f = float(v); return int(f) if f.is_integer() else round(f, 2)

START = datetime.date(2025, 4, 4)          # first data day = age wk 23 (datefinal row 3)
AGE0 = 23
def age_week(d): return AGE0 + (d - START).days // 7
def cycle_week(d): return age_week(d) - 18  # D26
def week_from_start(d): return (d - START).days // 7 + 1  # formula from cycles.json start

# hall -> (cols Jos, Cat1, Cat2, Cat3, Incub, Total, Masa), compartments (code, ♂ col, ♀ col)
HALLS = {
    'H1': (('B', 'C', 'D', 'E', 'F', 'G', 'H'), [('H1.1', 'I', 'J'), ('H1.2', 'K', 'L')]),
    'H3': (('M', 'N', 'O', 'P', 'Q', 'R', 'S'), [('H3.1', 'T', 'U'), ('H3.2', 'V', 'W')]),
    'H4': (('X', 'Y', 'Z', 'AA', 'AB', 'AC', 'AD'), [('H4.1', 'AE', 'AF'), ('H4.2', 'AG', 'AH')]),
}
ws = wb['import date']
raw = []
for r in range(3, ws.max_row + 1):
    dv = ws.cell(r, 1).value
    if not isinstance(dv, datetime.datetime): continue
    d = dv.date()
    if d > datetime.date(2026, 1, 28): continue   # rows after are #N/A
    raw.append((r, d))
log('import date rows', len(raw), raw[0][1], raw[-1][1])

# last day with eggs per hall; depopulation start = first day with hall hen deaths >= 250
daily, total_mismatch = [], []
hall_last, depop_start = {}, {}
for h, (cols, comps) in HALLS.items():
    for r, d in raw:
        tot = n(ws.cell(r, ci(cols[5])).value)
        if tot > 0: hall_last[h] = d
        hens_dead = sum(n(ws.cell(r, ci(c[2])).value) for c in comps)
        if hens_dead >= 250 and h not in depop_start: depop_start[h] = d
log('hall last egg day', {h: str(v) for h, v in hall_last.items()})
log('depopulation start', {h: str(v) for h, v in depop_start.items()})

for r, d in raw:
    for h, (cols, comps) in HALLS.items():
        if d > hall_last[h]: continue
        jos, c1, c2, c3, inc, tot, masa = (n(ws.cell(r, ci(c)).value) for c in cols)
        calc = c1 + c2 + c3 + inc
        if calc != tot: total_mismatch.append((str(d), h, tot, calc))
        mort = {code: {'F': n(ws.cell(r, ci(fc)).value), 'M': n(ws.cell(r, ci(mc)).value)} for code, mc, fc in comps}
        daily.append({
            'date': d.isoformat(), 'age_week': age_week(d), 'cycle_week': cycle_week(d),
            'week_from_start': week_from_start(d), 'hall': h,
            'floor_eggs': jos, 'cat1': c1, 'cat2': c2, 'cat3': c3, 'hatching_eggs': inc,
            'total_eggs': calc, 'hatching_pct': round(100 * inc / calc, 1) if calc else None,
            'avg_egg_weight_g': masa if masa else None,
            'mortality': mort,
            'mortality_F': sum(m['F'] for m in mort.values()), 'mortality_M': sum(m['M'] for m in mort.values()),
            'hens_alive': None, 'lay_pct': None,
            'depopulation': h in depop_start and d >= depop_start[h],
        })
log('daily rows', len(daily), 'Total ou != Cat1+2+3+Incub:', len(total_mismatch), total_mismatch[:5])

# weekly
def agg(rows):
    o = {k: sum(x[k] for x in rows) for k in ('floor_eggs', 'cat1', 'cat2', 'cat3', 'hatching_eggs', 'total_eggs', 'mortality_F', 'mortality_M')}
    ws_ = [x['avg_egg_weight_g'] for x in rows if x['avg_egg_weight_g']]
    o['avg_egg_weight_g'] = round(sum(ws_) / len(ws_), 2) if ws_ else None
    o['hatching_pct'] = round(100 * o['hatching_eggs'] / o['total_eggs'], 1) if o['total_eggs'] else None
    o['floor_pct_of_total'] = round(100 * o['floor_eggs'] / o['total_eggs'], 1) if o['total_eggs'] else None
    o['eggs_per_day'] = round(o['total_eggs'] / len({x['date'] for x in rows}), 0)
    return o

weekly = []
weeks = sorted({x['age_week'] for x in daily})
for aw in weeks:
    for h in list(HALLS) + ['all']:
        rows = [x for x in daily if x['age_week'] == aw and (h == 'all' or x['hall'] == h)]
        if not rows: continue
        days = sorted({x['date'] for x in rows})
        o = {'age_week': aw, 'cycle_week': aw - 18, 'week_from_start': aw - AGE0 + 1, 'hall': h,
             'from': days[0], 'to': days[-1], 'days': len(days), 'partial': len(days) < 7}
        o.update(agg(rows))
        o['hens_alive'] = None; o['lay_pct'] = None
        o['depopulation'] = any(x['depopulation'] for x in rows)
        weekly.append(o)

# cross-check with datefinal (ouat / incubabil / greutate ou ÷ 7)
df = wb['datefinal']
DFC = {'H1': ('N', 'O', 'P'), 'H3': ('AC', 'AD', 'AE'), 'H4': ('AR', 'AS', 'AT')}
xmis = []
for r in range(3, df.max_row + 1):
    aw = df.cell(r, 1).value
    if not isinstance(aw, (int, float)): continue
    aw = int(aw)
    for h, (co, cinc, cw) in DFC.items():
        mine = next((x for x in weekly if x['age_week'] == aw and x['hall'] == h), None)
        o, i_ = n(df.cell(r, ci(co)).value), n(df.cell(r, ci(cinc)).value)
        if mine is None:
            xmis.append({'age_week': aw, 'hall': h, 'datefinal_eggs': o, 'daily_sum': None}); continue
        if (o, i_) != (mine['total_eggs'], mine['hatching_eggs']):
            xmis.append({'age_week': aw, 'hall': h, 'datefinal_eggs': o, 'daily_sum': mine['total_eggs'],
                         'datefinal_hatching': i_, 'daily_hatching': mine['hatching_eggs']})
log('datefinal mismatches', len(xmis), xmis[:8])

tot_all = {k: sum(x[k] for x in daily) for k in ('floor_eggs', 'cat1', 'cat2', 'cat3', 'hatching_eggs', 'total_eggs', 'mortality_F', 'mortality_M')}
by_hall = {h: {k: sum(x[k] for x in daily if x['hall'] == h) for k in ('total_eggs', 'hatching_eggs', 'floor_eggs', 'mortality_F', 'mortality_M')} for h in HALLS}
log('totals', tot_all); log('by hall', by_hall)

GEN = {'generated_by': 'DATA02', 'script': '/home/claude/scripts/data02_prev.py <xlsx> <out_dir>',
       'source': 'design/data/Seria_2026.xlsx', 'workbook_last_real_date': '2026-02-28',
       'series_id': 'seria-2025', 'cycle_id': 'cyc-hirbovat-2025', 'farm_id': 'hirbovat', 'real': True}
common_notes = [
    'Previous Hîrbovăț flock ("Seria 2025", illustrative name), not Seria 2026.',
    'total_eggs = cat1 + cat2 + cat3 + hatching_eggs; floor_eggs (Jos) kept separate (D18). The sheet\'s "Total ou" equals this sum on every hall-day.',
    'age_week: datefinal puts 2025-04-04 at age wk 23. cycle_week = age_week − 18 (D26, same index as guide-laying.json). week_from_start = guide week by the data-model formula from the inferred start 2025-04-04 in cycles.json (= age_week − 22). The two differ by 4; see notes/DATA02.md.',
    'hens_alive and lay_pct are null: the workbook has no flock count for the previous flock (no placement/transfer row, no alive column), so % lay cannot be derived. Mortality is per compartment only.',
    'avg_egg_weight_g = sheet "Masa" (g); 0 in the sheet (first days of lay) is null.',
    'depopulation = true from the first day the hall records ≥ 250 hen deaths (H1 2026-01-15, H3 2026-01-16, H4 2026-01-20): the "pierit" columns then hold removals of the flock, not ordinary mortality. Do not chart those as mortality.',
    'Halls H1 and H3 have no eggs after 2026-01-22 (rows of zeros omitted); H4 runs to 2026-01-28.',
    'DataHirbovat mirrors "import date" cell for cell (checked); its only extra data is one feed row (2025-04-04: H1.1 98 g ♂ / 113 g ♀) — no grams are extracted.',
    'Single-day spike kept as data: 2026-01-10 H1.1 ♀ 102 dead.',
]
dump('eggs-hirbovat-prev-daily.json', {
    'meta': dict(GEN, sheet='import date: per hall Jos, Cat 1–3, Incubabil, Total ou, Masa; "x.y cocos/gaina pierit" per compartment (labels "2.1/2.2" = H1.2)',
                 range=[daily[0]['date'], daily[-1]['date']], rows=len(daily), notes=common_notes,
                 totals=tot_all, totals_by_hall=by_hall,
                 hall_last_egg_day={h: str(v) for h, v in hall_last.items()},
                 depopulation_start={h: str(v) for h, v in depop_start.items()}),
    'rows': daily})
dump('eggs-hirbovat-prev-weekly.json', {
    'meta': dict(GEN, sheet='derived from eggs-hirbovat-prev-daily.json (import date); cross-checked with datefinal',
                 range=[weekly[0]['from'], weekly[-1]['to']], age_weeks=[weeks[0], weeks[-1]], notes=common_notes + [
                     'hall "all" = the three halls together.',
                     'avg_egg_weight_g = mean of the daily hall averages in the week (datefinal stores their sum, ÷ 7 by hand).',
                     'partial = fewer than 7 days (last week of each hall).'],
                 datefinal_mismatches=xmis),
    'rows': weekly})
