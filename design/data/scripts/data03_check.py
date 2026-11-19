"""DATA03 check: internal consistency of the rebuilt extracts (no workbook access).
Usage: python3 -I data03_check.py <extracts_dir>"""
import sys, json, os
from collections import defaultdict
EX = sys.argv[1]
def L(n): return json.load(open(os.path.join(EX, n), encoding='utf-8'))
fails = []
def ok(cond, msg):
    if not cond: fails.append(msg)

d = L('hirbovat-daily.json')
sc = d['section_day']['cols']; S = [dict(zip(sc, r)) for r in d['section_day']['rows']]
fc = d['farm_day']['cols']; F = {r[0]: dict(zip(fc, r)) for r in d['farm_day']['rows']}
hc = d['hall_day']['cols']; H = [dict(zip(hc, r)) for r in d['hall_day']['rows']]
prev = defaultdict(int)
agg = defaultdict(lambda: defaultdict(int))
for r in S:
    k = (r['section'], r['sex'])
    ok(r['alive'] == prev[k] - r['mort'] - r['sold'] + r['arr'] + r['mv'], f"alive {r['date']} {k}")
    ok(r['mort_sheet'] == r['mort'] + r['sold'] - r['arr'] - r['mv'], f"mort_sheet {r['date']} {k}")
    ok(not r['sold'] or (r['date'] >= '2026-11-12' and r['mort'] == 0), f"sold rule {r['date']} {k}")
    r['start'] = prev[k]
    ok(r['sapt'] == r['age_days'] // 7 + 1, f"sapt {r['date']}")
    prev[k] = r['alive']
    for f in ('alive', 'mort', 'sold', 'arr', 'mv', 'start'): agg[r['date']][f + '_' + r['sex']] += r[f]
for ds, f in F.items():
    for x in 'FM':
        ok(f['alive_' + x] == agg[ds]['alive_' + x], f'farm alive {ds} {x}')
        ok(f['mort_' + x] == agg[ds]['mort_' + x], f'farm mort {ds} {x}')
        ok(f['sold_' + x] == agg[ds]['sold_' + x], f'farm sold {ds} {x}')
        ok(agg[ds]['mv_' + x] == 0, f'net moves {ds} {x}')
    st = agg[ds]['start_F']
    ok(f['lay_pct'] == (round(100 * f['eggs'] / st, 2) if f['eggs'] is not None and st else None), f'D156 lay {ds} {f["lay_pct"]}')
heggs = defaultdict(int)
for h in H: heggs[h['date']] += h['eggs'] or 0
for ds, f in F.items(): ok((f['eggs'] or 0) == heggs[ds], f'eggs {ds}')
w = L('hirbovat-weekly.json')
for wk in w['weeks']:
    days = [F[ds] for ds in F if wk['from'] <= ds <= wk['to']]
    fm = wk['farm']
    for x in 'FM':
        ok(fm['mort_' + x] == sum(r['mort_' + x] for r in days), f"week {wk['sapt']} mort {x}")
        ok(fm['sold_' + x] == sum(r['sold_' + x] for r in days), f"week {wk['sapt']} sold {x}")
    ok(fm['lay_pct'] == (round(100 * fm['eggs'] / fm['hen_days'], 2) if fm['hen_days'] else None), f"week {wk['sapt']} lay")
    ok(fm['eggs'] == sum(r['eggs'] or 0 for r in days), f"week {wk['sapt']} eggs {fm['eggs']}")
    ok(fm['hen_days_prod'] == sum(r['alive_F'] for r in days), f"week {wk['sapt']} hen-days prod")
    ok(fm['hen_days'] == sum(agg[r['date']]['start_F'] for r in days), f"week {wk['sapt']} hen-days start (D159)")
    ok(fm['lay_pct_prod'] == (round(100 * fm['eggs'] / fm['hen_days_prod'], 2) if fm['hen_days_prod'] else None), f"week {wk['sapt']} lay prod")
    ok(abs(fm['kg_F'] - sum(r['kg_F'] or 0 for r in days)) < 1, f"week {wk['sapt']} kg_F")
sc_ = d['sales']['cols']; SL = [dict(zip(sc_, r)) for r in d['sales']['rows']]
for x in 'FM':
    ok(sum(r[x] for r in SL) == sum(r['sold'] for r in S if r['sex'] == x), f'sales list {x}')
c = L('counts-latest.json')['farms'][0]
last = F['2026-11-18']
ok(c['totals']['F'] == last['alive_F'] and c['totals']['M'] == last['alive_M'], 'counts-latest totals')
ok(c['cycle']['today_mortality'] == {'F': 3, 'M': 0}, f"today deaths {c['cycle']['today_mortality']}")
ok(c['cycle_status'] == 'open' and c['totals']['F'] + c['totals']['M'] > 0, 'cycle open with birds')
wg = L('weighings-weekly.json')
for r in wg['rows']:
    ok(r['sapt'] == r['age_days'] // 7 + 1 and r['guide_week'] == r['sapt'] - 1, f"weighing week {r['section']} {r['date']}")
h1 = [r for r in wg['rows'] if r['section'] == 'H1' and r['date'] == '2026-02-24'][0]
ok(h1['sapt'] == 21 and h1['guide_weight_g'] == 2370 and h1['guide_diff_pct'] == 7.7, 'D157 example')
t = L('transfers-seria2026.json')
ok(t['arrival_totals'] == {'F': sum(r['arr'] for r in S if r['sex'] == 'F'), 'M': sum(r['arr'] for r in S if r['sex'] == 'M')}, 'arrivals')
e = L('illustrative-events-hirbovat.json')
smap = {(r['date'], r['section'], r['sex']): r for r in S}
for sv in e['sales']:
    ok(any(r['date'] == sv['date'] and r['section'] == sv['section'] and r['F'] == sv['F'] and r['M'] == sv['M'] for r in SL), f'sale event {sv}')
# DATA04 (D168): one sale form = one event_log record; per day one record per hall sold; record totals = the sales block
SR = [x for x in e['event_log'] if x['record'] == 'sale']
SL_day = defaultdict(list)
for r in SL: SL_day[r['date']].append(r)
for ds, rows in SL_day.items():
    recs = [x for x in SR if x['at'][:10] == ds]
    halls = sorted({r['section'].split('.')[0] for r in rows})
    ok(len(recs) == len(halls), f'sale records {ds}: {len(recs)} for halls {halls}')
    ok([x['at'] for x in recs] == sorted({x['at'] for x in recs}), f'sale record times {ds}')
    ok(sorted(s_['section'] for x in recs for s_ in x['sections']) == sorted(r['section'] for r in rows), f'sale record sections {ds}')
    for x in recs:
        ok(len({s_['section'].split('.')[0] for s_ in x['sections']}) == 1, f"one hall per record {x['at']}")
        for s_ in x['sections']:
            r = next(r for r in rows if r['section'] == s_['section'])
            ok((s_['F'], s_['M']) == (r['F'], r['M']), f"sale record row {x['at']} {s_}")
        ok((x['F'], x['M']) == (sum(s_['F'] for s_ in x['sections']), sum(s_['M'] for s_ in x['sections'])), f"sale record totals {x['at']}")
        ok(x['scope'] == 'Hîrbovăț · ' + ', '.join(s_['section'] for s_ in x['sections']), f"sale record scope {x['at']}")
    for sv in e['sales']:
        if sv['date'] == ds:
            ok(sum(1 for x in recs if x['at'] == sv['created_at'] and any(s_['section'] == sv['section'] for s_ in x['sections'])) == 1, f'sales row → record {sv}')
ok({x['at'][:10] for x in SR} == set(SL_day), 'sale record days')
ok((sum(x['F'] for x in SR), sum(x['M'] for x in SR)) == (sum(r['F'] for r in SL), sum(r['M'] for r in SL)), 'sale records = sales block')
r18 = [x for x in SR if x['at'][:10] == '2026-11-18']
ok(len(r18) == 1 and (r18[0]['F'], r18[0]['M']) == (1831, 289), f'18.11 one record ♀ 1 831 · ♂ 289 {r18}')
r14 = [x for x in SR if x['at'][:10] == '2026-11-14']
ok([[s_['section'] for s_ in x['sections']] for x in r14] == [['H6.2'], ['H7.1', 'H7.2']] and (r14[1]['F'], r14[1]['M']) == (6915, 639), f'14.11 H6.2 then H7 {r14}')
for ev in e['culls'] + e['sexing_errors']:
    r = smap[(ev['date'], ev['section'], ev['sex'])]
    ok(abs(ev['qty']) <= r['mort'] and ev['carved_from_mortality_of_day'] == r['mort'] and not r['sold'], f"carve {ev}")
# RF08 (D158a): the one shared correction = a real compartment-day (stored = the real deaths), a vet Zilnic entry before it
CO = e['corrections']
ok(len(CO) == 1, f'one correction {len(CO)}')
for c in CO:
    rw = c['row']; r = smap[(rw['date'], rw['section'], rw['sex'])]
    ok(c['new_values']['qty'] == r['mort'] and 0 <= c['old_values']['qty'] != r['mort'] and not r['sold'], f'correction figures {c}')
    ok(any(x['record'] == 'mortality' and x['at'] == c['entry_at'] and x['user'] == c['entry_user'] for x in e['event_log']), f'correction entry {c}')
    ok(rw['date'] < c['at'][:10] <= '2026-11-18' and c['entry_at'] < c['at'] and len(c['reason']) <= 200, f'correction time/reason {c}')
    ok(rw['sapt'] == r['sapt'] and not any(x['sold'] for x in S if x['section'] == rw['section']), f'correction section without sales {c}')
for n in ('mortality-telita-daily.json',):
    for r in L(n)['rows']: ok(r['sapt'] == r['age_week'], f'telita sapt {r["date"]}')
fdl = L('feed-deliveries.json')
hir = [r for r in fdl['rows'] if r['farm_id'] == 'hirbovat']
ok(round(sum(r['total_lei'] or 0 for r in hir)) == fdl['meta']['totals']['hirbovat']['sheet_header']['lei'], 'lei total')
print('checks:', 'OK' if not fails else f'{len(fails)} FAIL'); [print(' ', f) for f in fails[:20]]
print('rows: section_day', len(S), 'hall_day', len(H), 'farm_day', len(F), 'weeks', len(w['weeks']))
