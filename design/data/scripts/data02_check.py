"""DATA02 checks (RETIRED after DATA03; exits early on DATA03 extracts): python3 -I data02_check.py <extracts_dir>"""
import sys, os, json, collections
EX = sys.argv[1]
L = lambda n: json.load(open(os.path.join(EX, n), encoding='utf-8'))
s, c, fd = L('sample-hirbovat-2026.json'), L('counts-latest.json'), L('feed-deliveries.json')
pd_, pw = L('eggs-hirbovat-prev-daily.json'), L('eggs-hirbovat-prev-weekly.json')
if c.get('meta', {}).get('generated_by') == 'DATA03':
    print('data02_check.py RETIRED: it checks sample-hirbovat-2026.json (superseded by DATA03). Use data03_check.py.')
    sys.exit(0)
err = []
last = {}
for r in s['daily']:
    k = (r['section'], r['sex'])
    if r['alive_start'] < 0 or r['alive_end'] < 0: err.append(('neg', r))
    if k in last and last[k] != r['alive_start']: err.append(('chain', r))
    if r['alive_start'] - r['mortality'] - r['culls'] - r['sexing_errors'] != r['alive_end']: err.append(('rec', r))
    if r['g_per_bird'] is not None and abs(r['g_per_bird'] * r['alive_end'] / 1000 - r['kg']) > 0.06: err.append(('kg', r))
    last[k] = r['alive_end']
st = {(x['section']): x for x in s['start_counts']}
for (sec, sx), v in last.items():
    cl = next(x for x in c['farms'][0]['sections'] if x['section'] == sec)
    if cl[sx] != v: err.append(('counts-latest', sec, sx))
tot = {x: sum(v for (sec, sx), v in last.items() if sx == x) for x in 'FM'}
assert tot == {'F': c['farms'][0]['totals']['F'], 'M': c['farms'][0]['totals']['M']}
rem = collections.Counter()
for r in s['daily']: rem[r['sex']] += r['mortality'] + r['culls'] + r['sexing_errors']
assert 26765 - rem['F'] == tot['F'] and 3060 - rem['M'] == tot['M']
for e in s['eggs_daily'] + pd_['rows']:
    if e['cat1'] + e['cat2'] + e['cat3'] + e['hatching_eggs'] != e['total_eggs']: err.append(('eggs', e))
    if min(e['cat1'], e['cat2'], e['cat3'], e['hatching_eggs'], e['floor_eggs']) < 0: err.append(('eggneg', e))
# stock recompute
cons = collections.Counter(); 
for r in s['daily']:
    if r['kg'] is not None: cons[r['feed_type_id']] += r['kg']
dl = collections.Counter()
for x in s['deliveries']: dl[x['feed_type_id']] += x['qty_kg']
for x in s['stock_today']:
    if abs(dl[x['feed_type_id']] - cons[x['feed_type_id']] - x['stock_kg']) > 2: err.append(('stock', x))
run = collections.Counter()
for r in s['feed_by_type_daily']:
    run[r['feed_type_id']] += r['delivered_kg'] - r['consumption_kg']
    if r['stock_end_kg'] < 0: err.append(('stockneg', r))
unpriced = sum(1 for x in s['deliveries'] + fd['rows'] if not x['priced'])
weekly_check = sum(r['total_eggs'] for r in pw['rows'] if r['hall'] != 'all') == sum(r['total_eggs'] for r in pd_['rows'])
print('errors', len(err), err[:3]); print('alive', tot, 'unpriced', unpriced, 'prev weekly=daily', weekly_check)
print('wk40', [(w['hall'], w['lay_pct'], w['diff_pts']) for w in s['eggs_weekly'] if w['cycle_week'] == 40])
print('stock', [(x['feed_type_id'], x['stock_kg'], x['days_left']) for x in s['stock_today']])
