"""Independent review of DATA02 extracts (read-only).
Usage: python3 -I data02_review.py <extracts_dir>
"""
import sys, os, json, datetime, collections
D = sys.argv[1]
L = lambda f: json.load(open(os.path.join(D, f), encoding='utf-8'))
S, C = L('sample-hirbovat-2026.json'), L('counts-latest.json')
TR, FD, GL = L('transfers-seria2026.json'), L('feed-deliveries.json'), L('guide-laying.json')
PD, PW, CY = L('eggs-hirbovat-prev-daily.json'), L('eggs-hirbovat-prev-weekly.json'), L('cycles.json')
ERR = []
def chk(c, *m):
    if not c: ERR.append(' '.join(map(str, m))); print('FAIL', *m)

print('--- (1) counts')
sc = S['start_counts']
print('start sums', sum(x['F'] for x in sc), sum(x['M'] for x in sc))
chk(sum(x['F'] for x in sc) == 26765 and sum(x['M'] for x in sc) == 3060, 'start sums')
print('transfers keys', list(TR.keys()))
for k, v in TR.items():
    if k != 'meta' and isinstance(v, list) and v: print(' ', k, len(v), json.dumps(v[0], ensure_ascii=False)[:200])
st = {(x['section'], x['sex']): x for x in S['section_totals']}
for x in S['section_totals']:
    chk(x['alive'] == x['start'] - x['mortality'] - x['culls'] - x['sexing_errors'], 'sec total', x)
    chk(x['alive'] >= 0, 'neg', x)
    chk(abs(x['mortality_pct'] - round(100 * x['mortality'] / x['start'], 2)) < 0.006, 'mort pct', x)
# daily chain
by = collections.defaultdict(list)
for r in S['daily']: by[(r['section'], r['sex'])].append(r)
for k, rows in by.items():
    rows.sort(key=lambda r: r['date'])
    start = next(x[k[1]] for x in sc if x['section'] == k[0])
    chk(rows[0]['alive_start'] == start, 'chain start', k)
    prev = None
    for r in rows:
        chk(r['alive_end'] == r['alive_start'] - r['mortality'] - r['culls'] - r['sexing_errors'], 'day', r)
        chk(min(r['alive_end'], r['mortality'], r['culls'], r['sexing_errors']) >= 0, 'neg day', r)
        if prev: chk(r['alive_start'] == prev['alive_end'], 'chain', r['date'], k)
        if r['g_per_bird'] is not None:
            chk(abs(r['kg'] - r['g_per_bird'] * r['alive_end'] / 1000) < 0.06, 'kg', r)
        prev = r
    t = st[k]
    chk(sum(r['mortality'] for r in rows) == t['mortality'] and sum(r['culls'] for r in rows) == t['culls']
        and sum(r['sexing_errors'] for r in rows) == t['sexing_errors'] and rows[-1]['alive_end'] == t['alive'], 'totals vs daily', k)
    chk(rows[-1]['date'] == '2026-11-16', 'last day', k, rows[-1]['date'])
# culls / sexing lists vs daily
cd = collections.Counter(); [cd.update({(c['date'], c['section'], c['sex']): c['qty']}) for c in S['culls']]
dd = collections.Counter({(r['date'], r['section'], r['sex']): r['culls'] for r in S['daily'] if r['culls']})
chk(cd == dd, 'culls list vs daily')
chk(all(c['reason'] in ('boală', 'șchiopătare', 'altul') for c in S['culls']), 'cull reasons')
se = collections.Counter(); [se.update({(c['date'], c['section'], c['sex']): c['qty']}) for c in S['sexing_errors']]
sd = collections.Counter({(r['date'], r['section'], r['sex']): r['sexing_errors'] for r in S['daily'] if r['sexing_errors']})
chk(se == sd, 'sexing list vs daily')
F = sum(t['alive'] for (s, x), t in st.items() if x == 'F'); M = sum(t['alive'] for (s, x), t in st.items() if x == 'M')
print('alive', F, M, F + M, 'summary', S['summary_today']['alive'])
chk(S['summary_today']['alive'] == {'F': F, 'M': M, 'total': F + M}, 'summary alive')
h = C['farms'][0]
for s in h['sections']:
    for x in 'FM': chk(s[x] == st[(s['section'], x)]['alive'], 'counts-latest', s['section'], x)
print('counts-latest farms', [(f['farm_id'], {k: v for k, v in f.items() if k not in ('sections',)}) for f in C['farms']])
print('series', C['series'])
chk(C['series'][0]['F'] == F and C['series'][0]['M'] == M, 'series totals')
mF = sum(t['mortality'] for (s, x), t in st.items() if x == 'F'); mM = sum(t['mortality'] for (s, x), t in st.items() if x == 'M')
print('cum mort pct', round(100 * mF / 26765, 2), round(100 * mM / 3060, 2), S['summary_today']['cumulative_mortality_pct'])

print('--- (2) eggs')
for r in S['eggs_daily'] + PD['rows']:
    chk(r['total_eggs'] == r['cat1'] + r['cat2'] + r['cat3'] + r['hatching_eggs'], 'egg total', r.get('date'), r['hall'])
for r in S['eggs_daily']:
    if r['hens_start']: chk(abs(r['lay_pct'] - round(100 * r['total_eggs'] / r['hens_start'], 1)) < 0.06, 'lay', r['date'], r['hall'])
# hens_start = F alive_start of hall sections
hs = collections.Counter()
for r in S['daily']:
    if r['sex'] == 'F': hs[(r['date'], r['section'].split('.')[0])] += r['alive_start']
for r in S['eggs_daily']: chk(hs[(r['date'], r['hall'])] == r['hens_start'], 'hens_start', r['date'], r['hall'], hs[(r['date'], r['hall'])], r['hens_start'])
# weekly vs daily (prev + sample)
for name, daily, weekly, key in (('prev', PD['rows'], PW['rows'], 'age_week'), ('sample', S['eggs_daily'], S['eggs_weekly'], 'cycle_week')):
    for w in weekly:
        rows = [r for r in daily if r[key] == w[key] and (w['hall'] == 'all' or r['hall'] == w['hall'])]
        for f in ('total_eggs', 'floor_eggs', 'hatching_eggs'):
            chk(sum(r[f] for r in rows) == w[f], name, 'weekly', w[key], w['hall'], f)
w40 = [w for w in S['eggs_weekly'] if w['cycle_week'] == 40]
print('wk40', [(w['hall'], w['from'], w['to'], w['days'], w['lay_pct'], w['guide_lay_pct'], w['diff_pts'], w['alert_lay']) for w in w40])
for w in S['eggs_weekly']:
    if w['guide_lay_pct'] is not None and w['diff_pts'] is not None:
        chk(abs(w['diff_pts'] - round(w['lay_pct'] - w['guide_lay_pct'], 1)) < 0.06, 'diff', w['cycle_week'], w['hall'])
        chk(w['alert_lay'] == (abs(w['diff_pts']) > 3) or w['alert_lay'] == (w['diff_pts'] < -3), 'alert flag', w['cycle_week'], w['hall'])
alerts = [(w['cycle_week'], w['hall'], w['diff_pts']) for w in S['eggs_weekly'] if w['alert_lay']]
print('all lay alerts', alerts)

print('--- (4) states')
al = [a for a in S['allocations_today']]
print('alloc states', collections.Counter(a['state'] for a in al))
for a in al: print('  ', a)
miss = [(a['section'], a['sex']) for a in al if a['state'] == 'missing']
d16 = [r for r in S['daily'] if r['date'] == '2026-11-16']
print('16.11 null g', [(r['section'], r['sex']) for r in d16 if r['g_per_bird'] is None])
dev = [(r['date'], r['section'], r['sex'], r['deviation_pct']) for r in S['daily'] if r['deviation_pct'] is not None and abs(r['deviation_pct']) > 10]
print('>10% deviations all days', dev)
for r in S['daily']:
    if r['g_per_bird'] is not None and r['guide_feed_g']:
        chk(abs(r['deviation_pct'] - round(100 * (r['g_per_bird'] - r['guide_feed_g']) / r['guide_feed_g'], 1)) < 0.06, 'devpct', r['date'], r['section'], r['sex'])
print('stock_today', S['stock_today'])
dl = S['deliveries']
unp = [d for d in dl if not d['priced']]
print('unpriced sample', [(d['date'], d['feed_type_id'], d['qty_kg']) for d in unp], sum(d['qty_kg'] for d in unp))
fdl = FD.get('deliveries') or FD.get('rows')
print('feed-deliveries keys', list(FD.keys()), 'unpriced telita', [d for d in fdl if not d.get('priced', True)])
d1011 = [d for d in dl if d['date'] == '2026-11-10']
print('10.11', d1011)
# stock chain per type
fb = collections.defaultdict(list)
for r in S['feed_by_type_daily']: fb[r['feed_type_id']].append(r)
cons = collections.Counter()
for r in S['daily']:
    if r['kg'] is not None: cons[(r['date'], r['feed_type_id'])] += r['kg']
delv = collections.Counter()
for d in dl: delv[(d['date'], d['feed_type_id'])] += d['qty_kg']
for t, rows in fb.items():
    rows.sort(key=lambda r: r['date']); s = 0
    for r in rows:
        chk(r['delivered_kg'] == delv[(r['date'], t)], 'deliv', t, r['date'])
        chk(abs(r['consumption_kg'] - round(cons[(r['date'], t)], 1)) < 0.2, 'cons', t, r['date'], r['consumption_kg'], cons[(r['date'], t)])
        s = s + r['delivered_kg'] - r['consumption_kg']
        chk(abs(s - r['stock_end_kg']) < 1, 'stock', t, r['date'], s, r['stock_end_kg'])
        chk(r['stock_end_kg'] >= 0, 'stock<0', t, r['date'])
        if r['date'] in ('2026-11-09', '2026-11-10', '2026-11-16'): print('  ', r)
    print(t, 'last', rows[-1])
print('event_log')
for e in S['event_log']: print('  ', e)
print('audit', S['audit_log'])

print('--- (5) guide week')
d0 = datetime.date(2026, 2, 10); t = datetime.date(2026, 11, 16)
print('sapt', (t - d0).days // 7 + 1)
g58 = [g for g in (GL.get('rows') or GL.get('weeks')) if g.get('age_week') == 58]
print('guide age 58', g58)
print('cycles prev', CY['cycles'][2])
print('ERRORS', len(ERR)); [print(' ', e) for e in ERR[:30]]
