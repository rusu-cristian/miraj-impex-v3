"""Spot-check previous-flock eggs against the workbook (read-only).
Usage: python3 -I data02_spot.py <xlsx> <extracts_dir>
"""
import sys, os, json, datetime
import openpyxl
wb = openpyxl.load_workbook(sys.argv[1], data_only=True, read_only=True)
ws = wb['import date']
rows = list(ws.iter_rows(min_row=1, max_row=ws.max_row, values_only=True))
for i in (0, 1):
    print('hdr', i + 1, [(openpyxl.utils.get_column_letter(j + 1), v) for j, v in enumerate(rows[i][:34]) if v is not None])
PD = json.load(open(os.path.join(sys.argv[2], 'eggs-hirbovat-prev-daily.json'), encoding='utf-8'))['rows']
PW = json.load(open(os.path.join(sys.argv[2], 'eggs-hirbovat-prev-weekly.json'), encoding='utf-8'))['rows']
byd = {}
for r in rows[2:]:
    if isinstance(r[0], datetime.datetime): byd[r[0].date()] = r
COLS = {'H1': 1, 'H3': 12, 'H4': 23}  # 0-based start of Jos
def sheet(d, h):
    r = byd[d]; b = COLS[h]
    v = [x or 0 for x in r[b:b + 11]]
    return {'floor_eggs': v[0], 'cat1': v[1], 'cat2': v[2], 'cat3': v[3], 'hatching_eggs': v[4], 'total_sheet': v[5],
            'masa': v[6], 'mort_M': v[7] + v[9], 'mort_F': v[8] + v[10]}
for ds in ('2025-06-15', '2025-11-03'):
    d = datetime.date.fromisoformat(ds)
    for h in COLS:
        s = sheet(d, h); x = next(r for r in PD if r['date'] == ds and r['hall'] == h)
        ok = all(s[k] == x[k] for k in ('floor_eggs', 'cat1', 'cat2', 'cat3', 'hatching_eggs')) and s['total_sheet'] == x['total_eggs'] \
            and s['mort_F'] == x['mortality_F'] and s['mort_M'] == x['mortality_M'] and (s['masa'] or None) == x['avg_egg_weight_g']
        print(ds, h, 'OK' if ok else 'MISMATCH', s, {k: x[k] for k in ('floor_eggs', 'cat1', 'cat2', 'cat3', 'hatching_eggs', 'total_eggs', 'avg_egg_weight_g', 'mortality_F', 'mortality_M', 'age_week', 'cycle_week')})
# one week: age wk 40 (2025-08-01 .. 08-07), all halls
d0 = datetime.date(2025, 8, 1)
tot = {k: 0 for k in ('floor_eggs', 'cat1', 'cat2', 'cat3', 'hatching_eggs', 'total_sheet', 'mort_F', 'mort_M')}
for i in range(7):
    for h in COLS:
        s = sheet(d0 + datetime.timedelta(i), h)
        for k in tot: tot[k] += s[k]
w = next(r for r in PW if r['age_week'] == 40 and r['hall'] == 'all')
print('wk40 sheet', tot)
print('wk40 json ', {k: w[k] for k in ('from', 'to', 'age_week', 'cycle_week', 'data_week', 'floor_eggs', 'cat1', 'cat2', 'cat3', 'hatching_eggs', 'total_eggs', 'mortality_F', 'mortality_M')})
