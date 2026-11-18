"""Build the S6 · Zilnic phone artboard (default state) from the extracts (spec C1, screens.md S6, D53).
Hîrbovăț, 16.11.2026, SAPT 40 (D23/D24), Dr. Ceban after the morning round, everything typed, before Salvează: the
16.11 sample day (sample-hirbovat-2026.json, D27) per section × sex (Mortalitate, Sacrificare 0), H3.1 with
"Fără pierderi azi" checked (no losses that day), the Ouă group of each hall (eggs_daily 16.11). Alive on a card =
start of the day (D53); yesterday hint = 15.11 and the mean of 09–15.11 (SYS05). No 3× warning and no refusal holds on
this day (asserted), so no card carries a Tag except H3.1 "fără pierderi"; H3's lay is 70,1% vs ghid 74,1% (abatere).
Composed from system classes only: FormHeader, DateField (s7_build shell), HallCard, EntryCard + SYS05 additions,
CountInput, Checkbox, Lamp, SexCount, SaveBar (s7_build.savebar, two-line summary). No screen-only CSS beyond S7's: "din ♀ 8 475"
stays on one line through HallCard's span.ip-eggs__base (S6P, system since FIN15). Field names = section heading + label via aria-labelledby (S6P).
One tall phone board (canvas.md: taller h for a long scrolling form), SaveBar sticky at the foot, as S7.
python3 -I s6_build.py <repo_root> [heights.json, default s6_heights.json]
Imports s7_build (it rewrites the S7 default board at 3162, byte-identical). Writes
design/canvas/project/S6-zilnic.dc.html and /tmp/s6/S6-zilnic.html (plain wrapper for renders)."""
import sys, json, pathlib, importlib, datetime as dt

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
ROOT, HP, CANVAS = CR.args('s6_heights.json')  # heights default: s6_heights.json
HERE = pathlib.Path(__file__).resolve().parent
STEM, TITLE = 'S6-zilnic', 'S6 · Zilnic'
H = json.loads(HP.read_text())[STEM] if HP and HP.exists() else 3400

sys.argv = ['s7_build.py', str(ROOT), '3162']
sys.path.insert(0, str(HERE))
B = importlib.import_module('s7_build')
n, dd, sexv, lamp, NB, MINUS = B.n, B.dd, B.sexv, B.lamp, B.NB, B.MINUS
TODAY, YDAY, WEEK, START, S = B.TODAY, B.YDAY, B.WEEK, B.START, B.S
EX = ROOT / 'design/data/extracts'
GL = json.load(open(EX / 'guide-laying.json', encoding='utf-8'))
FS = json.load(open(EX / 'farms-sections.json', encoding='utf-8'))

# record icons (path data from the RecordTile / EntryCardPlus / HallCard previews) + Checkbox check
B.IC.update({
 'mortality': '<path d="M12 3a7 7 0 0 0-7 7v4l2 2v3h10v-3l2-2v-4a7 7 0 0 0-7-7z"/><circle cx="9.5" cy="11" r="1.2"/><circle cx="14.5" cy="11" r="1.2"/><path d="M11 19v-2M13 19v-2"/>',
 'cull': '<circle cx="12" cy="12" r="9"/><path d="M5.7 5.7l12.6 12.6"/>',
 'eggs': '<path d="M12 22c4 0 7-3 7-7.5C19 9 16 2 12 2S5 9 5 14.5C5 19 8 22 12 22z"/>',
 'check': '<path d="M20 6 9 17l-5-5"/>',
 'alert': '<circle cx="12" cy="12" r="10"/><path d="M12 8v4"/><path d="M12 16h.01"/>',   # Banner fault / ip-err (S6S states)
 'sexerr': '<circle cx="8" cy="12" r="4" stroke="var(--rec-hen)"/><circle cx="16" cy="12" r="4" stroke="var(--rec-rooster)"/>',  # S6M: Eroare sexare (split tile rings)
 'chevup': '<path d="m18 15-6-6-6 6"/>',  # S6M: Mai mult open
})
icon = B.icon

# ---------- data (16.11.2026, sample D27) ----------
HALLS = [s for s in FS['sections'] if s['farm_id'] == 'hirbovat' and s.get('is_hall') and s['active']]
SECS = {h['id']: [s for s in FS['sections'] if s.get('parent_id') == h['id'] and s['active']] for h in HALLS}
D = {(d['date'], d['section'], d['sex']): d for d in S['daily']}
EGG = {e['hall']: e for e in S['eggs_daily'] if e['date'] == TODAY.isoformat()}
LAYG = next(r['lay_pct'] for r in GL['rows'] if r['cycle_week'] == WEEK and r['sex'] == 'F')
assert LAYG == 74.1 and [h['code'] for h in HALLS] == ['H1', 'H3', 'H4']

def day(sec, sx, d=TODAY): return D[(d.isoformat(), sec, sx)]
def hint(sec, sx, d=TODAY):
    w = [day(sec, sx, d - dt.timedelta(days=k))['mortality'] for k in range(1, 8)]
    return w[0], sum(w) / 7

# ---------- EntryCard (SYS05) ----------
def removed(r): return r['mortality'] + r['culls'] + r['sexing_errors']  # D53: per sex, refused above alive_start
def most(r): return max(0, r['alive_start'] - r['culls'] - r['sexing_errors'])  # the most Mortalitate still accepted
SIGN = {'F': '♀', 'M': '♂'}; CLS = {'F': 'hen', 'M': 'rooster'}; WORD = {'F': 'găini', 'M': 'cocoși'}
def cnt(v):
    return f'{n(v)}'
def field(kind, sec, sx, uid, val, ro, hint_html='', st=None, msg_id=None):
    # st (S6S state boards only): 'warn' / 'error' on the CountInput, msg_id = the row message it is described by
    # kind 'x' (S6M only): Eroare sexare under Mai mult, led by the split-tile rings (no single record colour, D04)
    rk, lbl, pre = {'m': ('mortality', 'Mortalitate', 'm'), 'c': ('cull', 'Sacrificare', 'c'), 'x': ('sexerr', 'Eroare sexare', 'x')}[kind]
    rec = f'<span class="ip-rec ip-rec--{rk}">{icon(rk, 14)}</span>' if kind != 'x' else f'<span class="ip-rec">{icon(rk, 14)}</span>'
    fid = f'{pre}{"0" if sx == "F" else "1"}-{uid}'
    ids = ([f'h{CLS[sx][0]}-{uid}'] if hint_html else []) + ([msg_id] if msg_id else [])
    desc = f' aria-describedby="{" ".join(ids)}"' if ids else ''
    pat = '[0-9 &#8239;]*' if val >= 1000 else '[0-9]*'
    stc = f' is-{st}' if st else ''
    inv = ' aria-invalid="true"' if st == 'error' else ''
    # accessible name = the card's section heading + the field label ("H1.1 Mortalitate găini"); <label for> keeps the tap target
    return (f'<div class="ip-field is-{CLS[sx]}"><label class="ip-label" for="{fid}" id="l-{fid}">{rec}{lbl} '
            f'<span class="ip-sex ip-sex--{CLS[sx]}" aria-hidden="true">{SIGN[sx]}</span><span class="ip-sr">{WORD[sx]}</span></label>'
            f'<span class="ip-input ip-count ip-count--{CLS[sx]}{stc}"><input id="{fid}" inputmode="numeric" pattern="{pat}" placeholder="–" autocomplete="off" value="{cnt(val)}" aria-labelledby="t-{uid} l-{fid}"{" readonly" if ro else ""}{inv}{desc}></span>{hint_html}</div>')

def entry(sec, d=TODAY, ov=None, saved=None, reasons=None, more=False):
    # Default S6 board: entry(sec). S6S state boards: d = the chosen day, ov = {('mortality', 'F'): 4, ...} typed values
    # (D39 state examples), saved = the saved meta html (read-only card, Tag "salvat", foot = meta; Editează is form-level).
    # S6M (D79): reasons = {sx: '' | reason} draws the editable Motiv Select (empty = "Motiv lipsă", a refusal, D53);
    # more = Mai mult open with the Eroare sexare CountInputs.
    uid = sec.replace('.', '')
    t = {sx: dict(day(sec, sx, d)) for sx in 'FM'}
    for (k, sx), v in (ov or {}).items(): t[sx][k] = v
    refused = {sx: removed(t[sx]) > t[sx]['alive_start'] for sx in 'FM'}
    warned = {sx: t[sx]['mortality'] > 3 * hint(sec, sx, d)[1] and saved is None for sx in 'FM'}  # a saved day is not re-validated
    missing = {sx: reasons is not None and t[sx]['culls'] > 0 and not reasons.get(sx) for sx in 'FM'}  # D53: no cull reason = refusal
    if ov is None and saved is None and d == TODAY:
        for sx in 'FM':  # D53: refusal = removals per sex above alive at the start of the day; warning = mortality > 3 × 7-day mean
            assert not refused[sx]
            assert not warned[sx], (sec, sx)
            assert t[sx]['culls'] == 0 and t[sx]['sexing_errors'] == 0  # 16.11: no culls, no sexing errors (no Motiv Select today)
    nolosses = all(t[sx]['mortality'] == 0 and t[sx]['culls'] == 0 for sx in 'FM')
    hints = {}
    for sx in 'FM':
        y, m = hint(sec, sx, d)
        hints[sx] = (f'<span class="ip-entry__hint" id="h{CLS[sx][0]}-{uid}"><span>ieri <span class="ip-sex ip-sex--{CLS[sx]}">{SIGN[sx]}</span> <span class="ip-num">{y}</span></span>'
                     f'<span aria-hidden="true">·</span><span>media 7 zile <span class="ip-num">{n(m, 1)}</span></span></span>')
    ro = nolosses or saved is not None
    st = {sx: 'error' if refused[sx] else ('warn' if warned[sx] else None) for sx in 'FM'}
    mid = {sx: (f'e-{uid}-{CLS[sx][0]}' if st[sx] == 'error' else f'w-{uid}-{CLS[sx][0]}') if st[sx] else None for sx in 'FM'}
    msgs = ''
    for sx in 'FM':
        v = t[sx]['mortality']; a = t[sx]['alive_start']
        if st[sx] == 'error':  # D53: the refusal is on the sum; the fix names the most Mortalitate still accepted
            msgs += (f'<p class="ip-err ip-entry__msg s7-m0" id="{mid[sx]}">{icon("alert", 14)}<span>{sec} <span class="ip-sex ip-sex--{CLS[sx]}">{SIGN[sx]}</span>: Mortalitate + Sacrificare + Eroare sexare =&#160;<span class="ip-num">{n(removed(t[sx]))}</span>, '
                     f'peste cei <span class="ip-num">{n(a)}</span> {WORD[sx]} în viață pe <span class="ip-num">{dd(d)}</span>. La Mortalitate scrie cel mult <span class="ip-num">{n(most(t[sx]))}</span>.</span></p>')
        elif st[sx] == 'warn':
            msgs += (f'<p class="ip-warnline ip-entry__msg s7-m0" id="{mid[sx]}" role="status">{icon("warn", 14)}<span>{sec}: <span class="ip-num">{n(v)}</span> {WORD[sx]} moarte, '
                     f'de peste 3 ori media pe 7 zile (<span class="ip-num">{n(hint(sec, sx, d)[1], 1)}</span>). Verifică numărătoarea; poți salva.</span></p>')
    motiv = ''
    for sx in 'FM':  # SYS05: Sacrificare > 0 shows its Motiv under it (read-only on a saved day: label + readonly input)
        if t[sx]['culls'] > 0 and reasons is not None:  # S6M: the editable Select (boală, șchiopătare, altul; D36), placeholder "Alege"
            r, rv = f'r{"0" if sx == "F" else "1"}-{uid}', reasons.get(sx) or ''
            em = f'e-{uid}-r{"0" if sx == "F" else "1"}'
            opts = '<option value=""' + (' selected' if not rv else '') + ' disabled hidden>Alege</option>' + ''.join(
                f'<option{" selected" if o == rv else ""}>{o}</option>' for o in ('boală', 'șchiopătare', 'altul'))
            aria = f' aria-invalid="true" aria-describedby="{em}"' if missing[sx] else ''
            motiv += (f'<div class="ip-field is-{CLS[sx]}"><label class="ip-label" for="{r}" id="l-{r}">Motiv <span class="ip-sex ip-sex--{CLS[sx]}" aria-hidden="true">{SIGN[sx]}</span><span class="ip-sr">{WORD[sx]}</span></label>'
                      f'<span class="ip-input ip-input--touch{" is-error" if missing[sx] else ""}"><select id="{r}" required aria-labelledby="t-{uid} l-{r}"{aria}>{opts}</select>{icon("chev", 16)}</span></div>')
            if missing[sx]:  # spans the row under the Select (SYS05); refusals name the section, the number, the day and the fix
                motiv += (f'<p class="ip-err ip-entry__msg s7-m0" id="{em}">{icon("alert", 14)}<span>{sec} <span class="ip-sex ip-sex--{CLS[sx]}">{SIGN[sx]}</span>: Motiv lipsă pentru <span class="ip-num">{n(t[sx]["culls"])}</span> {WORD[sx]} sacrificate pe <span class="ip-num">{dd(d)}</span>. '
                          f'Alege motivul: boală, șchiopătare sau altul.</span></p>')
        elif t[sx]['culls'] > 0:
            rsn = next(c['reason'] for c in S['culls'] if c['date'] == d.isoformat() and c['section'] == sec and c['sex'] == sx)
            motiv += (f'<div class="ip-field is-{CLS[sx]}"><label class="ip-label" for="r{"0" if sx == "F" else "1"}-{uid}" id="l-r{"0" if sx == "F" else "1"}-{uid}">Motiv <span class="ip-sex ip-sex--{CLS[sx]}" aria-hidden="true">{SIGN[sx]}</span><span class="ip-sr">{WORD[sx]}</span></label>'
                      f'<span class="ip-input ip-input--touch"><input id="r{"0" if sx == "F" else "1"}-{uid}" type="text" value="{rsn}" aria-labelledby="t-{uid} l-r{"0" if sx == "F" else "1"}-{uid}" readonly></span></div>')
    grid = (field('m', sec, 'F', uid, t['F']['mortality'], ro, hints['F'], st['F'], mid['F']) + field('m', sec, 'M', uid, t['M']['mortality'], ro, hints['M'], st['M'], mid['M'])
            + msgs + field('c', sec, 'F', uid, t['F']['culls'], ro) + field('c', sec, 'M', uid, t['M']['culls'], ro) + motiv)
    panel = ' hidden>' if not more else ('><div class="ip-entry__grid">' + field('x', sec, 'F', uid, t['F']['sexing_errors'], ro)
                                         + field('x', sec, 'M', uid, t['M']['sexing_errors'], ro) + '</div>')
    if any(refused.values()) or any(missing.values()): tag = '<span class="ip-tag ip-tag--fault">refuzat</span>'
    elif any(warned.values()): tag = '<span class="ip-tag ip-tag--warn">abatere</span>'
    elif saved is not None: tag = '<span class="ip-tag ip-tag--neutral">salvat</span>'
    else: tag = '<span class="ip-tag ip-tag--neutral">fără pierderi</span>' if nolosses else ''
    box = f'<span class="ip-check__box">{icon("check", 14)}</span>' if nolosses else '<span class="ip-check__box"></span>'
    chk = f'<label class="ip-check{" is-checked" if nolosses else ""}"><input type="checkbox" class="ip-sr" aria-labelledby="t-{uid} fp-{uid}"{" checked" if nolosses else ""}>{box}<span id="fp-{uid}">Fără pierderi azi</span></label>'
    html = f'''<section class="ip-panel ip-entry" aria-labelledby="t-{uid}"><div class="ip-entry__head"><h3 class="ip-entry__code" id="t-{uid}">{sec}</h3>{tag}<span class="ip-entry__alive">în viață <span class="ip-pair"><span class="ip-sex ip-sex--hen">♀ <b>{n(t["F"]["alive_start"])}</b></span><span class="sep">·</span><span class="ip-sex ip-sex--rooster">♂ <b>{n(t["M"]["alive_start"])}</b></span></span></span></div>
<div class="ip-entry__grid">{grid}</div>
<button type="button" class="ip-btn ip-btn--touch ip-btn--ghost ip-entry__more" id="mm-{uid}" aria-labelledby="t-{uid} mm-{uid}" aria-expanded="{"true" if more else "false"}" aria-controls="more-{uid}">{icon("chevup" if more else "chev", 16)}Mai mult · Eroare sexare</button><div class="ip-entry__panel" id="more-{uid}"{panel}</div>
<div class="ip-entry__foot">{chk if saved is None else saved}</div></section>'''
    return html, t, nolosses

# ---------- Ouă group (HallCard) ----------
def eggs(h, secs, hens, d=TODAY, ro=False):
    e = EGG[h] if d == TODAY else next(x for x in S['eggs_daily'] if x['date'] == d.isoformat() and x['hall'] == h); u = h
    assert e['hens_start'] == hens and e['guide_lay_pct'] == LAYG
    tot = e['cat1'] + e['cat2'] + e['cat3'] + e['hatching_eggs']; assert tot == e['total_eggs']  # D18, Jos separate
    lay = round(tot / hens * 100, 1); assert lay == e['lay_pct']
    diff = round(lay - LAYG, 1); ok = abs(diff) <= 3  # D19 ±3 points
    def f(fid, lbl, v):
        pat = '[0-9 &#8239;]*' if v >= 1000 else '[0-9]*'
        return f'<div class="ip-field"><label class="ip-label" for="{fid}-{u}" id="l-{fid}-{u}">{lbl}</label><span class="ip-input ip-count"><input id="{fid}-{u}" inputmode="numeric" pattern="{pat}" placeholder="–" autocomplete="off" value="{n(v)}" aria-labelledby="t-eggs-{u} l-{fid}-{u}"{" readonly" if ro else ""}></span></div>'
    grid = (f('jos', 'Jos', e['floor_eggs']) + f('ec1', 'Cat. 1', e['cat1']) + f('ec2', 'Cat. 2', e['cat2']) + f('ec3', 'Cat. 3', e['cat3'])
            + f('inc', 'Incubabile', e['hatching_eggs'])
            + f'<div class="ip-field"><label class="ip-label" for="w-{u}" id="l-w-{u}">Masa medie<span class="ip-sr"> grame</span></label><span class="ip-input ip-eggs__weight ip-input--touch"><input id="w-{u}" inputmode="decimal" placeholder="–" autocomplete="off" value="{n(e["avg_egg_weight_g"], 1)}" aria-labelledby="t-eggs-{u} l-w-{u}"{" readonly" if ro else ""}><span class="suffix">g</span></span></div>')
    sgn = '+' if diff > 0 else (MINUS if diff < 0 else '')
    sub = f'ghid <span class="ip-num">{n(LAYG, 1)}%</span> · <span class="ip-num">{sgn}{n(abs(diff), 1)}</span> puncte{"" if ok else ' (prag <span class="ip-num">±3</span>)'} · <span class="ip-eggs__base">din <span class="ip-sex ip-sex--hen">♀ {n(hens)}</span></span>'
    calc = (f'<dl class="ip-eggs__calc" aria-label="Calculat din câmpuri, doar citire"><div><dt>Total ouă</dt><dd class="ip-eggs__val">{n(tot)}</dd>'
            f'<dd class="ip-eggs__sub">= Cat. 1 + 2 + 3 + Incubabile · Jos separat (<span class="ip-num">{n(e["floor_eggs"])}</span>)</dd></div>'
            f'<div><dt>Ouat azi</dt><dd class="ip-eggs__val">{n(lay, 1)}%</dd><dd class="ip-eggs__sub">{sub}</dd>'
            f'<dd class="ip-eggs__lamp">{lamp("ok", "în normă") if ok else lamp("warn", "abatere")}</dd></div></dl>')
    html = f'''<section class="ip-panel ip-eggs" aria-labelledby="t-eggs-{u}"><div class="ip-eggs__head"><span class="ip-tile ip-rec ip-rec--eggs" aria-hidden="true">{icon("eggs")}</span><h3 class="ip-entry__code" id="t-eggs-{u}">Ouă · {h}</h3><span class="ip-note">pe hală, pentru {" + ".join(secs)}</span></div>
<div class="ip-eggs__grid">{grid}</div>
{calc}</section>'''
    return html, tot, lay, ok

parts, ALL, EGGS = [], [], []
for hl in HALLS:
    h = hl['code']; secs = [s['code'] for s in SECS[hl['id']]]
    cards = [entry(s) for s in secs]; ALL += [(s, c) for s, c in zip(secs, cards)]
    af = sum(c[1]['F']['alive_start'] for c in cards); am = sum(c[1]['M']['alive_start'] for c in cards)
    eh, tot, lay, ok = eggs(h, secs, af); EGGS.append((h, tot, lay, ok))
    parts.append(f'''<section class="ip-hall" aria-labelledby="h-{h}"><header class="ip-hall__head"><span class="ip-cap">Hală</span><h2 class="ip-hall__code" id="h-{h}">{h}</h2><span class="ip-note">secțiuni <span class="ip-num">{" · ".join(secs)}</span></span><span class="ip-pair"><span class="ip-sr">în viață </span>{sexv("F", n(af))}<span class="sep">·</span>{sexv("M", n(am))}</span></header>
<div class="ip-hall__sections">{"".join(c[0] for c in cards)}</div>
{eh}</section>''')
HALLS_HTML = '<div class="ip-halls">' + ''.join(parts) + '</div>'

# ---------- header + date (S7 shell, Zilnic title and the mortality RecordTile) ----------
HEAD = B.HEAD.replace(f'<span class="ip-tile ip-rec ip-rec--feed" aria-hidden="true">{icon("feed")}</span>',
                      f'<span class="ip-tile ip-rec ip-rec--mortality" aria-hidden="true">{icon("mortality")}</span>').replace('>Hrană zilnică<', '>Zilnic<')
assert '>Zilnic<' in HEAD and 'ip-rec--mortality' in HEAD and f'HÎRBOVĂȚ · {dd(TODAY)} · SAPT {WEEK}' in HEAD
DATE = B.DATE

# ---------- SaveBar (s7_build.savebar: two lines; segments never break inside, §20) ----------
# Line 1: removals of the day (spec C1 toast form "6 secțiuni · ♀ −23 · ♂ −5"); line 2: eggs of the day + the one status.
remF = sum(c[1]['F']['mortality'] + c[1]['F']['culls'] + c[1]['F']['sexing_errors'] for _, c in ALL)
remM = sum(c[1]['M']['mortality'] + c[1]['M']['culls'] + c[1]['M']['sexing_errors'] for _, c in ALL)
EV = next(e for e in S['event_log'] if e['at'].startswith(TODAY.isoformat()) and e['record'] == 'mortality')
EVE = next(e for e in S['event_log'] if e['at'].startswith(TODAY.isoformat()) and e['record'] == 'eggs')
egg_tot = sum(x[1] for x in EGGS)
assert EV['summary'] == f'Zilnic salvat · {len(ALL)} secțiuni · ♀ {MINUS}{remF} · ♂ {MINUS}{remM}'  # the 08:34 save of these values
assert f'{n(egg_tot)} total' in EVE['summary'] and n(egg_tot) == f'18{NB}688'
def rec14(k): return f'<span class="ip-rec ip-rec--{k}">{icon(k, 14)}</span>'
line1 = (f'{rec14("mortality")}<span><span class="ip-num">{len(ALL)}</span> secțiuni ·</span>'
         f'<span>{sexv("F", f"{MINUS}{remF}")} · {sexv("M", f"{MINUS}{remM}")}</span>')
bad = [x for x in EGGS if not x[3]]
line2 = f'{rec14("eggs")}<span><span class="ip-num">{n(egg_tot)}</span> ouă ·</span>' + (
    lamp('warn', 'abatere') + '<span>' + ', '.join(h for h, _, _, _ in bad) + '</span>' if bad else lamp('ok', 'în normă'))
# line 2 names only the hall ("abatere · H3"): with its % (70,1%) it wraps to a third line at 390; the Ouă group of H3
# carries the figure, the guide and the ±3 threshold.
SAVE = B.savebar(line1, line2, '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch">Salvează</button>')

BODY = f'''<div class="ip-screen">
{HEAD}
<main class="ip-screen__main">
{DATE}
{HALLS_HTML}
</main>
{SAVE}
</div>'''

CSS = B.CSS  # FIN15: "din ♀ 8 475" on one line is HallCard's span.ip-eggs__base (additions.css)
root = f'<div class="ip-root" style="padding:0; width:390px; min-height:{H}px">{BODY}</div>'
dc = f'''<!doctype html>
<html lang="ro">
<head>
<meta charset="utf-8">
<title>{TITLE}</title>
<script src="./support.js"></script>
<link rel="stylesheet" href="ds/impexlinepanel/tokens.css">
<link rel="stylesheet" href="ds/impexlinepanel/components/bundle.css">
<link rel="stylesheet" href="additions.css">
</head>
<body>
<x-dc>
<helmet>
{B.FONTS}
<style>{CSS}</style>
</helmet>
{root}
</x-dc>
<script type="text/x-dc" data-dc-script data-props='{{"$preview":{{"width":390,"height":{H}}}}}'>
class Component extends DCLogic {{ renderVals() {{ return {{}}; }} }}
</script>
</body>
</html>
'''
(ROOT / f'design/canvas/project/{STEM}.dc.html').write_text(dc, encoding='utf-8')
SYS = ROOT / 'design/system'
wrap = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8"><title>{TITLE}</title>
<link rel="stylesheet" href="file://{SYS}/tokens.css"><link rel="stylesheet" href="file://{SYS}/components/bundle.css"><link rel="stylesheet" href="file://{SYS}/components/additions.css">
{B.FONTS}<style>{CSS}</style></head><body>{root}</body></html>'''
pathlib.Path('/tmp/s6').mkdir(exist_ok=True)
pathlib.Path(f'/tmp/s6/{STEM}.html').write_text(wrap, encoding='utf-8')
print('ok', STEM, H, len(dc), 'removals', remF, remM, 'eggs', egg_tot, [(h, l, ok) for h, _, l, ok in EGGS])
