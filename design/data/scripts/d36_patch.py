"""D36 patch for previous-flock egg extracts (text-level, keeps formatting).
Usage: python3 -I d36_patch.py <extracts_dir>
- renames row field week_from_start -> data_week (extra index, 1 = first data week from 2025-04-04)
- rewrites the meta note on week indexes (cycle start inferred 2025-03-07, age wk 19)
- asserts cycle_week == age_week - 18 == guide week from 2025-03-07
"""
import sys, os, re, json, datetime

D = sys.argv[1]
OLD = ('age_week: datefinal puts 2025-04-04 at age wk 23. cycle_week = age_week − 18 (D26, same index as guide-laying.json). '
       'week_from_start = guide week by the data-model formula from the inferred start 2025-04-04 in cycles.json (= age_week − 22). '
       'The two differ by 4; see notes/DATA02.md.')
NEW = ('age_week: datefinal puts 2025-04-04 at age wk 23. cycle_week = age_week − 18 (D26, same index as guide-laying.json) '
       '= guide week by the data-model formula from the inferred cycle start 2025-03-07 (age wk 19) in cycles.json (D36). '
       'Data starts 2025-04-04 = cycle wk 5. data_week is an extra index only (1 = first data week from 2025-04-04, = age_week − 22).')
START = datetime.date(2025, 3, 7)
for name in ('eggs-hirbovat-prev-daily.json', 'eggs-hirbovat-prev-weekly.json'):
    p = os.path.join(D, name)
    s = open(p, encoding='utf-8').read()
    assert s.count(OLD) == 1, name
    s = s.replace(OLD, NEW)
    s, k = re.subn(r'"week_from_start": ', '"data_week": ', s)
    assert 'week_from_start' not in s
    obj = json.loads(s)
    for r in obj['rows']:
        assert r['cycle_week'] == r['age_week'] - 18
        assert r['data_week'] == r['age_week'] - 22
        if 'date' in r:
            d = datetime.date.fromisoformat(r['date'])
            assert (d - START).days // 7 + 1 == r['cycle_week'], r
        else:
            d = datetime.date.fromisoformat(r['from'])
            assert (d - START).days // 7 + 1 == r['cycle_week'], r
    open(p, 'w', encoding='utf-8').write(s)
    print(name, 'rows', len(obj['rows']), 'renamed', k)
