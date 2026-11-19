"""realdata.py · the real data every generator reads (RF01; D151–D158). Pure data + formats, no HTML; imports no sNN script.
Reads design/data/extracts/ (DATA03) next to this script's repo once, on import, so a copy of the repo reads its own copy.
"Today" = 18.11.2026 (D152). Weeks are bird age weeks (D151): age_days = date − placement (+ age at placement), SAPT =
age_days // 7 + 1, weeks run Tue–Mon. Hîrbovăț halls H5–H7 (OLD2NEW maps the old H1/H3/H4 codes of the previous-flock
files). Removals from 12.11 are sales, mortality = real deaths (D155). Lay % = eggs ÷ hens at the start of the day, weekly
eggs ÷ hen-days (D156). A weighing is compared with the guide of the week just ended (D157). See README.md › realdata.

Sex keys are 'F' (♀) / 'M' (♂); dates are datetime.date or ISO strings (both accepted); figures moving under DATA (D155
split, D156 lay) are computed here from the daily file, never copied from key-figures-18-11.md."""
import json, pathlib, datetime as dt
from decimal import Decimal, ROUND_HALF_UP

HERE = pathlib.Path(__file__).resolve().parent
ROOT = HERE.parents[2]
EX = ROOT / 'design/data/extracts'
def J(name): return json.loads((EX / name).read_text(encoding='utf-8'))

# ---------- formats (craft §6, D46, D82, D129/FIN09) ----------
NB = ' '; MINUS = '−'; DASH = '–'
def rnd(v, dec=0):
    """.5 rounds away from zero (D82); float noise cut at 1e-9 first."""
    d = Decimal(repr(float(v))).quantize(Decimal('1e-9')).quantize(Decimal(1).scaleb(-dec), rounding=ROUND_HALF_UP)
    return float(d) if dec else int(d)
def n(v, dec=0):
    """Thin no-break thousands (U+202F), decimal comma, true minus: n(-12561) → '−12 561'."""
    v = rnd(v, dec)
    s = f'{abs(v):,.{dec}f}'.replace(',', 'X').replace('.', ',').replace('X', NB)
    return (MINUS if v < 0 else '') + s
def sgn(v, dec=1):
    v = rnd(v, dec); return ('+' if v > 0 else '') + n(v, dec)
def pct(v, dec=1):  # D46: no space before %
    return n(v, dec) + '%'
def D(d):
    """date from a date or an ISO string."""
    return d if isinstance(d, dt.date) else dt.date.fromisoformat(d[:10])
def iso(d): return D(d).isoformat()
def dd(d): return D(d).strftime('%d.%m.%Y')          # 18.11.2026
def ddm(d): return D(d).strftime('%d.%m')            # 18.11
def ddt(at):                                          # D129: '18.11.2026, 08:15'
    return f'{dd(at[:10])}, {at[11:16]}'
def drange(a, b, year_end=True):
    """D129 range: en dash, no spaces; the start drops only a year it shares with the end ('17.11–18.11.2026',
    '07.10.2025–28.02.2026'); year_end=False drops the end's year too (an in-context day window of the current week)."""
    a, b = D(a), D(b)
    if not year_end: return f'{ddm(a)}{DASH}{ddm(b)}'
    return f'{ddm(a) if a.year == b.year else dd(a)}{DASH}{dd(b)}'
def SAPT(k): return f'SAPT {k}'                       # Mono context lines, Readouts
def sapt_txt(k): return f'sapt. {k}'                  # running text, toasts

# ---------- farms, sections, series, cycles ----------
FS = J('farms-sections.json'); CY = J('cycles.json')
FARMS = {f['id']: f for f in FS['farms']}
GUIDES = {g['id']: g for g in FS['guides']}
OLD2NEW = dict(FS['meta']['code_map_old_to_new'])     # previous-flock / superseded-sample codes → H5–H7 (D152, D158c)
def new_code(c): return OLD2NEW.get(c, c)
SECTIONS = {f: [s['code'] for s in FS['sections'] if s['farm_id'] == f and not s.get('is_hall')] for f in FARMS}
HALLS = {'hirbovat': [s['code'] for s in FS['sections'] if s['farm_id'] == 'hirbovat' and s.get('is_hall')]}
HALL_SECS = {h: [s for s in SECTIONS['hirbovat'] if s.startswith(h + '.')] for h in HALLS['hirbovat']}
HSECS, HHALLS = SECTIONS['hirbovat'], HALLS['hirbovat']
SERIES = {s['id']: s for s in CY['series']}
CYCLES = {c['id']: c for c in CY['cycles']}
C26 = CYCLES['cyc-hirbovat-2026']; C25 = CYCLES['cyc-hirbovat-2025']; CTEL = CYCLES['cyc-telita-2026']
PLACEMENT = D(SERIES['seria-2026']['placement_date'])  # 07.10.2025, age 0 that day

TODAY = D(C26['today']); TODAY_ISO = TODAY.isoformat()
def age_days(d, series='seria-2026'):
    s = SERIES[series]; return (D(d) - D(s['placement_date'])).days + s['age_at_placement_days']
def sapt(d, series='seria-2026'):
    """D151: SAPT = floor(age_days / 7) + 1 (13 days = 2, 14 days = 3)."""
    return age_days(d, series) // 7 + 1
def week_dates(k, series='seria-2026'):
    """(first, last) day of age week k (Tue–Mon for Seria 2026)."""
    s = SERIES[series]; a = D(s['placement_date']) + dt.timedelta(days=7 * (k - 1) - s['age_at_placement_days'])
    return a, a + dt.timedelta(days=6)
WK = sapt(TODAY)
CYCLE_START = D(C26['start_date'])                    # 22.02.2026 (♂ 22.02, ♀ 25–28.02)
EMPTYING_SINCE = D(C26['emptying_since'])             # 12.11.2026, by sale (D155)

# ---------- Hîrbovăț daily (section × sex, hall, farm) ----------
_HD = J('hirbovat-daily.json')
def _rows(block): cols = _HD[block]['cols']; return [dict(zip(cols, r)) for r in _HD[block]['rows']]
SECTION_DAY = _rows('section_day'); HALL_DAY = _rows('hall_day'); FARM_DAY = _rows('farm_day'); SALE_ROWS = _rows('sales')
_SD = {(r['date'], r['section'], r['sex']): r for r in SECTION_DAY}
_HDY = {(r['date'], r['hall']): r for r in HALL_DAY}
_FD = {r['date']: r for r in FARM_DAY}
_ARR = {}  # first day a compartment × sex holds birds
for r in SECTION_DAY: _ARR.setdefault((r['section'], r['sex']), r['date'])
_LAST = {}
for r in SECTION_DAY: _LAST[(r['section'], r['sex'])] = r['date']

def sd(d, sec, sx):
    """the section-day row (mort, sold, arr, mv, alive, g, kg, …) or None (before arrival / after emptying)."""
    return _SD.get((iso(d), sec, sx))
def alive(sec, sx, d=TODAY):
    """birds alive at the END of day d (0 before arrival and after emptying)."""
    k = iso(d)
    if k < _ARR.get((sec, sx), '9999'): return 0
    r = _SD.get((k, sec, sx))
    return r['alive'] if r else 0
def alive_start(sec, sx, d=TODAY):
    return alive(sec, sx, D(d) - dt.timedelta(days=1))
def alive_hall(hall, sx, d=TODAY): return sum(alive(s, sx, d) for s in HALL_SECS[hall])
def alive_farm(sx, d=TODAY): return sum(alive(s, sx, d) for s in HSECS)
def _sum(field, a, b, sec=None, sx=None, hall=None):
    a, b = iso(a), iso(b)
    return sum(r[field] for r in SECTION_DAY if a <= r['date'] <= b and (sec is None or r['section'] == sec)
               and (sx is None or r['sex'] == sx) and (hall is None or r['section'].startswith(hall + '.')))
def deaths(a, b=None, sec=None, sx=None, hall=None):
    """real deaths (D155) in [a, b] (one day if b is None)."""
    return _sum('mort', a, b or a, sec, sx, hall)
def sold(a, b=None, sec=None, sx=None, hall=None):
    """birds sold (Vânzare, D155) in [a, b]."""
    return _sum('sold', a, b or a, sec, sx, hall)
def sales(a=None, b=None):
    """real sales [{date, sapt, section, F, M}] in [a, b] (default: all, 12–18.11.2026)."""
    a, b = iso(a) if a else '0001-01-01', iso(b) if b else '9999-12-31'
    return [r for r in SALE_ROWS if a <= r['date'] <= b]
def empty_since(sec_or_hall):
    """the day a compartment (or every compartment of a hall) reached 0 for good, else None."""
    secs = HALL_SECS.get(sec_or_hall, [sec_or_hall])
    if any(alive(s, sx) for s in secs for sx in 'FM'): return None
    return D(max(_LAST[(s, sx)] for s in secs for sx in 'FM'))
def has_birds(sec_or_hall, d=TODAY):
    return any(alive(s, sx, d) for s in HALL_SECS.get(sec_or_hall, [sec_or_hall]) for sx in 'FM')

def day(d, hall=None):
    """farm row (hall None) or hall row of a day: alive_*, mort_*, sold_*, eggs, inc, jos, lay_pct (D156), egg_g, kg_*,
    and on the farm row stock_*_kg, deliv_*_kg, guide_g_*, guide_lay_pct. None if no row."""
    return _FD.get(iso(d)) if hall is None else _HDY.get((iso(d), hall))
def eggs(d=TODAY, hall=None):
    r = day(d, hall); return r['eggs'] if r else 0
def lay_day(d=TODAY, hall=None):
    """D156: eggs ÷ hens alive at the start of the day × 100 (computed; equals the file's lay_pct to 0,01)."""
    hens = alive_hall(hall, 'F', D(d) - dt.timedelta(days=1)) if hall else alive_farm('F', D(d) - dt.timedelta(days=1))
    return eggs(d, hall) / hens * 100 if hens else None

# ---------- weekly rows (SAPT 20–59; SAPT 59 = 17–18.11, partial) ----------
WEEKS = {w['sapt']: w for w in J('hirbovat-weekly.json')['weeks']}
def week(k, hall=None, sec=None):
    """weekly farm / hall / compartment figures of age week k (lay = eggs ÷ start-of-day hen-days, D156/D159); + 'from', 'to', 'days',
    'complete' on the farm row."""
    w = WEEKS.get(k)
    if w is None: return None
    if sec: return w['halls'][sec.split('.')[0]]['sections'].get(sec)
    if hall: return w['halls'].get(hall)
    return {**w['farm'], 'from': w['from'], 'to': w['to'], 'days': w['days'], 'complete': w['complete'], 'guide': w['guide']}
def lay_week(k, hall=None):
    r = week(k, hall); return r['lay_pct'] if r else None

# ---------- guides (by age week, D151) ----------
GL = J('guide-laying.json'); GG = J('guide-growth.json')
_GL = {(r['age_week'], r['sex']): r for r in GL['rows']}
_GG = {(r['age_week'], r['sex']): r for r in GG['rows']}
def guide_row(k, sx='F'): return _GL.get((k, sx))
def guide_lay(k):
    r = _GL.get((k, 'F')); return r['lay_pct'] if r else None
def guide_feed(k, sx):
    r = _GL.get((k, sx)) or _GG.get((k, sx)); return r['feed_g'] if r else None
def guide_mort(k):
    r = _GL.get((k, 'F')); return r['mortality_cum_pct'] if r else None
def guide_growth(k, sx):
    """guide body weight at the END of age week k (growth guide wk 1–25)."""
    r = _GG.get((k, sx)); return r['body_weight_g'] if r else None
def weigh_guide(k, sx):
    """D157: a weighing in SAPT k is compared with the guide of the week just ended, k − 1."""
    return guide_growth(k - 1, sx)
GUIDE_LAY_WEEKS = sorted(k for (k, sx), r in _GL.items() if sx == 'F' and r['lay_pct'] is not None)

# ---------- feed: norm g/bird, kg, stock, deliveries ----------
def feed_g(d=TODAY, sx='F', sec=None):
    """norm g/bird of a compartment, or the bird-weighted farm g (kg ÷ birds at the start of the day) over the
    compartments fed that day; None if nothing was fed."""
    if sec:
        r = sd(d, sec, sx); return r['g'] if r and r['kg'] else None
    rs = [sd(d, s, sx) for s in HSECS]; rs = [r for r in rs if r and r['kg']]
    birds = sum(r['kg'] * 1000 / r['g'] for r in rs)
    return sum(r['kg'] for r in rs) * 1000 / birds if birds else None
def feed_kg(d=TODAY, sx='F'):
    r = day(d); return r[f'kg_{sx}'] if r else 0
def stock(d=TODAY):
    """feed stock at the end of day d: {'F': kg Breeder 2 ♀, 'M': kg Cocoși ♂}."""
    r = day(d); return {'F': r['stock_F_kg'], 'M': r['stock_M_kg']}
FEED_TYPE = {'F': 'ft-breeder2', 'M': 'ft-cocosi'}
FEED_TYPES = {t['id']: t for t in J('feed-types.json')['rows']}
def stock_days(sx, d=TODAY, days=3):
    """(days left, average daily kg of the last `days` days, run-out date) for one stock at the end of day d."""
    ds = [D(d) - dt.timedelta(days=i) for i in range(days)]
    avg = sum(feed_kg(x, sx) for x in ds) / days
    left = stock(d)[sx] / avg
    return left, avg, D(d) + dt.timedelta(days=int(left))
_DEL = J('feed-deliveries.json')
DELIVERIES = [r for r in _DEL['rows'] if r['farm_id'] == 'hirbovat']
DELIVERIES_TELITA = [r for r in _DEL['rows'] if r['farm_id'] == 'telita']
UNPRICED = [r for r in DELIVERIES if not r['priced']]

# ---------- counts, arrivals, illustrative records ----------
_CL = J('counts-latest.json')
COUNTS = next(f for f in _CL['farms'] if f['farm_id'] == 'hirbovat')
COUNTS_TELITA = next(f for f in _CL['farms'] if f['farm_id'] == 'telita')
TR = J('transfers-seria2026.json')
ARRIVED = dict(TR['arrival_totals'])                  # ♀ 26 883 · ♂ 3 024
ILL = J('illustrative-events-hirbovat.json')          # culls, sexing_errors, sales (authors), weighings, event_log
EVENTS = sorted(ILL['event_log'], key=lambda e: e['at'], reverse=True)
# DATA04 (D168): one sale form = one record. record 'sale' carries `sections` [{section, F, M}] and totals F/M; an older
# extract (one record per compartment) is read into the same shape. `section` = the first compartment (backward compatible).
for _e in EVENTS:
    if _e['record'] == 'sale':
        if 'sections' not in _e:
            _s = _e['scope'].split(' · ')[1]; _r = next(r for r in SALE_ROWS if r['date'] == _e['at'][:10] and r['section'] == _s)
            _e.update(sections=[dict(section=_s, F=_r['F'], M=_r['M'])], F=_r['F'], M=_r['M'])
        _e['section'] = _e['sections'][0]['section']
        assert all(next(r for r in SALE_ROWS if r['date'] == _e['at'][:10] and r['section'] == x['section'])[k] == x[k]
                   for x in _e['sections'] for k in 'FM') and (_e['F'], _e['M']) == tuple(sum(x[k] for x in _e['sections']) for k in 'FM'), _e
SALE_EVENTS = [e for e in EVENTS if e['record'] == 'sale']   # newest first; 7 forms 12–18.11 (18.11: H6.1 + H6.2 ♀ 1 831 · ♂ 289)
assert sorted((s['date'], s['section']) for s in SALE_ROWS) == sorted((e['at'][:10], x['section']) for e in SALE_EVENTS for x in e['sections'])
def sale_codes(e): return ', '.join(x['section'] for x in e['sections'])   # "H6.1, H6.2"
WEIGHINGS_HIR = ILL['weighings']                      # illustrative (D152): H5.1 10.11.2026

# ---------- Telița and the previous flock (Seria 2025, codes mapped to H5–H7) ----------
PLACEMENT_ROWS = J('placement-seria2026.json')
WEIGHINGS_TELITA = J('weighings-weekly.json')['rows']
MORT_TELITA = J('mortality-telita-daily.json')
FEED_TELITA = J('feed-daily-telita.json')['days']
def _prev(r):
    r = dict(r); r['hall'] = new_code(r['hall'])
    if isinstance(r.get('mortality'), dict): r['mortality'] = {new_code(k): v for k, v in r['mortality'].items()}
    return r
_PD = J('eggs-hirbovat-prev-daily.json'); _PW = J('eggs-hirbovat-prev-weekly.json')
PREV_DAILY = [_prev(r) for r in _PD['rows']]          # 'age_week' is the week (D151); the files' old cycle index is not used
PREV_WEEKLY = [_prev(r) for r in _PW['rows']]
PREV_META = {'totals': _PD['meta']['totals'],
             'totals_by_hall': {new_code(h): v for h, v in _PD['meta']['totals_by_hall'].items()},
             'hall_last_egg_day': {new_code(h): v for h, v in _PD['meta']['hall_last_egg_day'].items()},
             'depopulation_start': {new_code(h): v for h, v in _PD['meta']['depopulation_start'].items()},
             'cycle_id': _PD['meta']['cycle_id']}
assert PREV_META['cycle_id'] == _PW['meta']['cycle_id'] == C25['id']

# ---------- stable key figures (key-figures-18-11.md); D155/D156 figures are computed above, not asserted ----------
assert (TODAY, WK, age_days(TODAY)) == (dt.date(2026, 11, 18), 59, 407)
assert week_dates(59) == (dt.date(2026, 11, 17), dt.date(2026, 11, 23))
assert (alive_farm('F'), alive_farm('M')) == (12561, 794) == (COUNTS['totals']['F'], COUNTS['totals']['M'])
assert eggs() == 5914 and ARRIVED == {'F': 26883, 'M': 3024}
assert abs(lay_day() - day(TODAY)['lay_pct']) < 0.01
assert HHALLS == ['H5', 'H6', 'H7'] and HSECS == ['H5.1', 'H5.2', 'H6.1', 'H6.2', 'H7.1', 'H7.2']
