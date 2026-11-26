"""Build the S6 · Zilnic phone artboard (default state) from realdata (RF04; D151–D158, today 18.11.2026).
Hîrbovăț, 18.11.2026, SAPT 59, Dr. Ceban (illustrative) after the morning round, everything typed, before Salvează: the real
18.11 deaths per compartment × sex (Mortalitate = real deaths only, D155; the day's sales are Vânzare, not Zilnic), Sacrificare 0
(no illustrative cull that day), the Ouă group of each hall with birds (eggs, incubabile, jos, egg weight of 18.11). Cards = the
compartments with birds at the start of the day (H5.1 H5.2 H6.1 H6.2); H7, empty since 14.11, keeps its hall head and one muted
line "goală din 14.11.2026" (D158d, as S7). Alive on a card = start of the day (D53); yesterday hint = 17.11 and the mean of the
7 days before (SYS05); illustrative culls / sexing errors (illustrative-events) are carved out of the real deaths, as in S15/S16.
Compartments with no death that day carry "Fără pierderi azi" checked. Lay % = eggs ÷ hens at the start of the day (D156) vs the
guide of the age week; Cat. 1, Cat. 2, Cat. 3 = the hall's real counts of the day (D183, DATA06 `cat1–cat3`; 18.11 H5 13 · 417 · 27,
H6 8 · 110 · 31), with Incubabile; Total ouă = Cat. 1–3 + Incubabile = the day's real eggs (D18; S6C supersedes D164's placeholder).
Composed from system classes only: FormHeader, DateField (s7_build shell), HallCard, EntryCard + SYS05 additions,
CountInput, Checkbox, Lamp, SexCount, SaveBar (s7_build.savebar, two-line summary). No screen-only CSS beyond S7's: "din ♀ …"
stays on one line through HallCard's span.ip-eggs__base (S6P, system since FIN15). Field names = section heading + label via aria-labelledby (S6P).
One tall phone board (canvas.md: taller h for a long scrolling form), SaveBar sticky at the foot, as S7.
python3 -I s6_build.py <repo_root> [heights.json, default s6_heights.json] [--canvas]   (--canvas: refresh the entry's h/title in place)
Imports s7_build (it rewrites the S7 default board at its default height, byte-identical). Writes
design/canvas/project/S6-zilnic.dc.html and /tmp/s6/S6-zilnic.html (plain wrapper for renders; the render measures the natural
height with the min-height released, s6_render.py)."""
import sys, json, pathlib, importlib, datetime as dt

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
import vkey as V   # VO02: the voice key on every signed-in board (spec §8, D199; vkey.add in the writer)
ROOT, HP, CANVAS = CR.args('s6_heights.json')  # heights default: s6_heights.json
HERE = pathlib.Path(__file__).resolve().parent
STEM, TITLE = 'S6-zilnic', 'S6 · Zilnic'
H = json.loads(HP.read_text())[STEM] if HP and HP.exists() else 3400

sys.argv = ['s7_build.py', str(ROOT)]  # RF02: s7_build's own default height
sys.path.insert(0, str(HERE))
B = importlib.import_module('s7_build')
import sexsign as SXS  # SX01: the §72 sign markup
import realdata as R  # RF04: every figure, date, week and record from realdata (DATA03 extracts)
n, dd, sexv, lamp, NB, MINUS = B.n, B.dd, B.sexv, B.lamp, B.NB, B.MINUS
TODAY, YDAY, WEEK, START = B.TODAY, B.YDAY, B.WEEK, B.START

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

# ---------- data (realdata, 18.11.2026) ----------
HALLS = R.HHALLS                                     # H5 H6 H7
LAYG = R.guide_lay(WEEK)                             # guide lay % of the age week (Productivitate col I, D184; 57,6 at SAPT 59)
assert LAYG is not None and WEEK == R.sapt(TODAY)
def secs_with_birds(h, d=TODAY):
    """the compartments of hall h with birds at the start of day d (the cards of that day)."""
    return [s for s in R.HALL_SECS[h] if any(R.alive_start(s, sx, d) for sx in 'FM')]
def _ill(kind, d, sec, sx):  # illustrative culls / sexing errors of a compartment-day (carved out of its real deaths)
    return sum(abs(c['qty']) for c in R.ILL[kind] if c['date'] == R.iso(d) and c['section'] == sec and c['sex'] == sx)
def cull_reason(d, sec, sx):
    return next(c['reason'] for c in R.ILL['culls'] if c['date'] == R.iso(d) and c['section'] == sec and c['sex'] == sx)

def day(sec, sx, d=TODAY):
    """what the vet types for a compartment × sex on day d: Mortalitate (real deaths − illustrative culls / sexing errors),
    Sacrificare, Eroare sexare, and the birds alive at the start of the day (D53)."""
    c, x = _ill('culls', d, sec, sx), _ill('sexing_errors', d, sec, sx)
    m = R.deaths(d, sec=sec, sx=sx) - c - x; assert m >= 0, (d, sec, sx)
    return {'mortality': m, 'culls': c, 'sexing_errors': x, 'alive_start': R.alive_start(sec, sx, d)}
def hint(sec, sx, d=TODAY):
    w = [day(sec, sx, d - dt.timedelta(days=k))['mortality'] for k in range(1, 8)]
    return w[0], sum(w) / 7

# ---------- EntryCard (SYS05) ----------
def removed(r): return r['mortality'] + r['culls'] + r['sexing_errors']  # D53: per sex, refused above alive_start
def most(r): return max(0, r['alive_start'] - r['culls'] - r['sexing_errors'])  # the most Mortalitate still accepted
SIGN = {'F': '♀', 'M': '♂'}; CLS = {'F': 'hen', 'M': 'rooster'}; WORD = {'F': 'găini', 'M': 'cocoși'}
def dead(sx, v):  # VO03: number + gender agreement ("1 găină moartă", "3 găini moarte", "1 cocoș mort", "3 cocoși morți"; 20 or more take "de")
    w = {('F', 1): 'găină moartă', ('F', 2): 'găini moarte', ('M', 1): 'cocoș mort', ('M', 2): 'cocoși morți'}[(sx, 1 if v == 1 else 2)]
    return ('de ' if v % 100 >= 20 or (v >= 100 and v % 100 == 0) else '') + w
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
    # accessible name = the card's section heading + the field label ("H5.1 Mortalitate găini"); <label for> keeps the tap target
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
            assert t[sx]['culls'] == 0 and t[sx]['sexing_errors'] == 0  # 18.11: no illustrative cull / sexing error (no Motiv Select today)
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
            msgs += (f'<p class="ip-err ip-entry__msg ip-m0" id="{mid[sx]}">{icon("alert", 14)}<span>{sec} <span class="ip-sex ip-sex--{CLS[sx]}">{SIGN[sx]}</span>: Mortalitate + Sacrificare + Eroare sexare =&#160;<span class="ip-num">{n(removed(t[sx]))}</span>, '
                     f'peste cei <span class="ip-num">{n(a)}</span> {WORD[sx]} în viață pe <span class="ip-num">{dd(d)}</span>. La Mortalitate scrie cel mult <span class="ip-num">{n(most(t[sx]))}</span>.</span></p>')
        elif st[sx] == 'warn':
            msgs += (f'<p class="ip-warnline ip-entry__msg ip-m0" id="{mid[sx]}" role="status">{icon("warn", 14)}<span>{sec}: <span class="ip-num">{n(v)}</span> {dead(sx, v)}, '
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
                motiv += (f'<p class="ip-err ip-entry__msg ip-m0" id="{em}">{icon("alert", 14)}<span>{sec} <span class="ip-sex ip-sex--{CLS[sx]}">{SIGN[sx]}</span>: Motiv lipsă pentru <span class="ip-num">{n(t[sx]["culls"])}</span> {WORD[sx]} sacrificate pe <span class="ip-num">{dd(d)}</span>. '
                          f'Alege motivul: boală, șchiopătare sau altul.</span></p>')
        elif t[sx]['culls'] > 0:
            rsn = cull_reason(d, sec, sx)
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
    html = f'''<section class="ip-panel ip-entry" aria-labelledby="t-{uid}"><div class="ip-entry__head"><h3 class="ip-entry__code" id="t-{uid}">{sec}</h3>{tag}<span class="ip-entry__alive">în viață<span class="ip-sr"> la începutul zilei</span> <span class="ip-pair"><span class="ip-sex ip-sex--hen">♀ <b>{n(t["F"]["alive_start"])}</b></span><span class="sep">·</span><span class="ip-sex ip-sex--rooster">♂ <b>{n(t["M"]["alive_start"])}</b></span></span></span></div>
<div class="ip-entry__grid">{grid}</div>
<button type="button" class="ip-btn ip-btn--touch ip-btn--ghost ip-entry__more" id="mm-{uid}" aria-labelledby="t-{uid} mm-{uid}" aria-expanded="{"true" if more else "false"}" aria-controls="more-{uid}">{icon("chevup" if more else "chev", 16)}Mai mult · Eroare sexare</button><div class="ip-entry__panel" id="more-{uid}"{panel}</div>
<div class="ip-entry__foot">{chk if saved is None else saved}</div></section>'''
    return html, t, nolosses

# ---------- Ouă group (HallCard) ----------
def eggs(h, secs, hens, d=TODAY, ro=False):
    """the hall's Ouă group of day d (realdata hall row): Jos, Cat. 1–3, Incubabile, Masa medie; Total ouă = Cat. 1 + 2 + 3 +
    Incubabile (D18, Jos separate). Cat. 1–3 = the hall's real counts of day d (D183, realdata day()['cat1'..'cat3']; S6C)."""
    e = R.day(d, h); u = h
    assert hens == R.alive_hall(h, 'F', d - dt.timedelta(days=1))  # hens at the start of the day (D156)
    cat = {'ec1': e['cat1'], 'ec2': e['cat2'], 'ec3': e['cat3']}  # D183: the day's real counts (DATA06), no placeholder
    assert None not in cat.values(), (d, h)  # a board's day with eggs has its Cat. 1–3 (Raspunsuri, 10.03–18.11.2026)
    tot = sum(cat.values()) + e['inc']; assert tot == e['eggs'], (d, h)  # D18: Total ouă = Cat. 1–3 + Incubabile, Jos separate
    lay = R.rnd(tot / hens * 100, 1); assert lay == R.rnd(R.lay_day(d, h), 1)
    diff = R.rnd(lay - LAYG, 1); ok = abs(diff) <= 3  # D19 ±3 points
    def f(fid, lbl, v):
        pat = '[0-9 &#8239;]*' if (v or 0) >= 1000 else '[0-9]*'
        val = '' if v is None else n(v)  # empty is null, not 0 (spec C1)
        return f'<div class="ip-field"><label class="ip-label" for="{fid}-{u}" id="l-{fid}-{u}">{lbl}</label><span class="ip-input ip-count"><input id="{fid}-{u}" inputmode="numeric" pattern="{pat}" placeholder="–" autocomplete="off" value="{val}" aria-labelledby="t-eggs-{u} l-{fid}-{u}"{" readonly" if ro else ""}></span></div>'
    grid = (f('jos', 'Jos', e['jos']) + f('ec1', 'Cat. 1', cat['ec1']) + f('ec2', 'Cat. 2', cat['ec2']) + f('ec3', 'Cat. 3', cat['ec3'])
            + f('inc', 'Incubabile', e['inc'])
            + f'<div class="ip-field"><label class="ip-label" for="w-{u}" id="l-w-{u}">Masa medie<span class="ip-sr"> grame</span></label><span class="ip-input ip-eggs__weight ip-input--touch"><input id="w-{u}" inputmode="decimal" placeholder="–" autocomplete="off" value="{n(e["egg_g"], 1)}" aria-labelledby="t-eggs-{u} l-w-{u}"{" readonly" if ro else ""}><span class="suffix">g</span></span></div>')
    sgn = '+' if diff > 0 else (MINUS if diff < 0 else '')
    sub = f'ghid <span class="ip-num">{n(LAYG, 1)}%</span> · <span class="ip-num">{sgn}{n(abs(diff), 1)}</span> puncte{"" if ok else ' (prag <span class="ip-num">±3</span>)'} · <span class="ip-eggs__base">din {SXS.figure("F", n(hens), num=True)}</span>'  # SX01: §72 sign + hidden word (was a ~5 px glyph)
    calc = (f'<dl class="ip-eggs__calc" aria-label="Calculat din câmpuri, doar citire"><div><dt>Total ouă</dt><dd class="ip-eggs__val">{n(tot)}</dd>'
            f'<dd class="ip-eggs__sub">= Cat. 1 + 2 + 3 + Incubabile · Jos separat (<span class="ip-num">{n(e["jos"])}</span>)</dd></div>'
            f'<div><dt>Ouat azi</dt><dd class="ip-eggs__val">{n(lay, 1)}%</dd><dd class="ip-eggs__sub">{sub}</dd>'
            f'<dd class="ip-eggs__lamp">{lamp("ok", "în normă") if ok else lamp("warn", "abatere")}</dd></div></dl>')
    html = f'''<section class="ip-panel ip-eggs" aria-labelledby="t-eggs-{u}"><div class="ip-eggs__head"><span class="ip-tile ip-rec ip-rec--eggs" aria-hidden="true">{icon("eggs")}</span><h3 class="ip-entry__code" id="t-eggs-{u}">Ouă · {h}</h3><span class="ip-note">pe hală, pentru {" + ".join(secs)}</span></div>
<div class="ip-eggs__grid">{grid}</div>
{calc}</section>'''
    return html, tot, lay, ok

def hall_head(h, secs, f, m):
    return (f'<header class="ip-hall__head"><span class="ip-cap">Hală</span><h2 class="ip-hall__code" id="h-{h}">{h}</h2><span class="ip-note">secțiuni <span class="ip-num">{" · ".join(secs)}</span></span>'
            f'<span class="ip-pair"><span class="ip-sr">{B.START_SR} </span>{sexv("F", n(f))}<span class="sep">·</span>{sexv("M", n(m))}</span></header>')
def halls(d=TODAY, card=None, ro=False):
    """HallCards of day d: per hall with birds its compartment cards (card(sec) → entry(), default entry(sec, d)) and its Ouă
    group; a hall without birds at the start of the day keeps its head and one muted line "goală din <date>" (D158d, as S7).
    Returns (html, [(section, entry())], [(hall, total eggs, lay %, ok)])."""
    parts, ALL, EG = [], [], []
    for h in HALLS:
        secs = secs_with_birds(h, d)
        if not secs:
            since = R.empty_since(h); assert since and since < d
            parts.append(f'''<section class="ip-hall" aria-labelledby="h-{h}">{hall_head(h, R.HALL_SECS[h], 0, 0)}
<p class="ip-note ip-m0">{lamp("off", "goală")} din <span class="ip-num">{dd(since)}</span></p></section>''')
            continue
        cards = [(card or (lambda s: entry(s, d)))(s) for s in secs]; ALL += list(zip(secs, cards))
        af = sum(c[1]['F']['alive_start'] for c in cards); am = sum(c[1]['M']['alive_start'] for c in cards)
        eh, tot, lay, ok = eggs(h, secs, af, d, ro); EG.append((h, tot, lay, ok))
        parts.append(f'''<section class="ip-hall" aria-labelledby="h-{h}">{hall_head(h, secs, af, am)}
<div class="ip-hall__sections">{"".join(c[0] for c in cards)}</div>
{eh}</section>''')
    return B.start_caption(d) + '<div class="ip-halls">' + ''.join(parts) + '</div>', ALL, EG  # D181: start-of-day counts, captioned

HALLS_HTML, ALL, EGGS = halls()
assert [s for s, _ in ALL] == [s for s in R.HSECS if R.has_birds(s, YDAY)]  # 18.11: H5.1 H5.2 H6.1 H6.2 (H7 empty since 14.11)
assert sum(c[1][sx]['alive_start'] for _, c in ALL for sx in 'F') == R.alive_farm('F', YDAY)

# ---------- header + date (S7 shell, Zilnic title and the mortality RecordTile) ----------
HEAD = B.HEAD.replace(f'<span class="ip-tile ip-rec ip-rec--feed" aria-hidden="true">{icon("feed")}</span>',
                      f'<span class="ip-tile ip-rec ip-rec--mortality" aria-hidden="true">{icon("mortality")}</span>').replace('>Hrană zilnică<', '>Zilnic<')
assert '>Zilnic<' in HEAD and 'ip-rec--mortality' in HEAD and f'HÎRBOVĂȚ · {dd(TODAY)} · SAPT {WEEK}' in HEAD
DATE = B.DATE

# ---------- SaveBar (s7_build.savebar: two lines; segments never break inside, §20) ----------
# Line 1: removals of the day (spec C1 toast form "6 secțiuni · ♀ −23 · ♂ −5"; 0 stays "0"); line 2: eggs of the day + the one status.
remF = sum(c[1]['F']['mortality'] + c[1]['F']['culls'] + c[1]['F']['sexing_errors'] for _, c in ALL)
remM = sum(c[1]['M']['mortality'] + c[1]['M']['culls'] + c[1]['M']['sexing_errors'] for _, c in ALL)
assert (remF, remM) == (R.deaths(TODAY, sx='F'), R.deaths(TODAY, sx='M'))  # Mortalitate = the real deaths of the day (D155)
def ev(d, rec):  # the illustrative event-log entry of a day (author, time) on the real figures
    return next(e for e in R.EVENTS if e['at'].startswith(R.iso(d)) and e['record'] == rec and e['action'] == 'create')
EV, EVE = ev(TODAY, 'mortality'), ev(TODAY, 'eggs')
egg_tot = sum(x[1] for x in EGGS)
assert EV['summary'] == f'Zilnic salvat · {len(ALL)} secțiuni · mortalitate ♀ {remF} · ♂ {remM}'  # the 08:15 save of these values
assert EVE['summary'].startswith(f'Ouă · {egg_tot} total') and egg_tot == R.eggs()
def signed(v): return f'{MINUS}{v}' if v else '0'
def rec14(k): return f'<span class="ip-rec ip-rec--{k}">{icon(k, 14)}</span>'
line1 = (f'{rec14("mortality")}<span><span class="ip-num">{len(ALL)}</span> secțiuni ·</span>'
         f'<span>{sexv("F", signed(remF))} · {sexv("M", signed(remM))}</span>')
bad = [x for x in EGGS if not x[3]]
line2 = f'{rec14("eggs")}<span><span class="ip-num">{n(egg_tot)}</span> ouă ·</span>' + (
    lamp('warn', 'abatere') + '<span>' + ', '.join(h for h, _, _, _ in bad) + '</span>' if bad else lamp('ok', 'în normă'))
# line 2 names only the halls ("abatere · H5, H6"): with their % it wraps to a third line at 390; each Ouă group
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

CSS = B.CSS  # FIN15: "din ♀ …" on one line is HallCard's span.ip-eggs__base (additions.css)
root = f'<div class="ip-root" style="padding:0; width:390px; min-height:{H}px">{V.add(BODY)}</div>'
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
if CANVAS and __name__ == '__main__':  # RF04: refresh the hand-placed entry's h/title in place (x, y kept; canvas_rows.place)
    c = CR.load(ROOT); CR.place(c, STEM, [(STEM, 390, H, TITLE)]); CR.save(ROOT, c)
print('ok', STEM, H, len(dc), 'removals', remF, remM, 'eggs', egg_tot, [(h, l, ok) for h, _, l, ok in EGGS])
