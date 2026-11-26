"""S16 · Corectează (spec E1, screens.md S16): the entry with its record icon, editable fields, the effect preview, the reason
(Motiv, required, 200 characters as a soft limit: no maxlength, the counter stays, over 200 = "Scurtează-l cu n"
and save disabled, D102/S16P), Salvează corecția / Șterge (reason + Dialog). Manager and owner only (spec §1).
Content (RF08, D158a): the one illustrative correction S15/S16/S17 share (illustrative-events-hirbovat.json `corrections`, via
realdata): Ion Rusu on 18.11.2026 at 10:12 corrects Dr. Ceban's Zilnic row of 17.11.2026 (08:07) for one compartment ♀ from
the value she typed (old) to the real deaths of that compartment-day (new; the daily rows hold it, asserted), reason
"Numărătoare refăcută: …". Every figure from realdata.py; author, time, old value and reason illustrative (D152); names only
Dr. Ceban / Ion Rusu (D96). Today = 18.11.2026 (D152), SAPT by age (D151).
Boards
- phones (390×844, the SidePanel as a full-screen sheet, additions.css §39), 18.11.2026 before 10:12:
  S16-corecteaza                    the real figure typed over the old one, the real reason, effect on the compartment's hens
                                    from the record's day to today, Salvează corecția live
  S16-corecteaza--motiv-lipsa       the same value with no reason: Motiv error, Salvează corecția disabled (D39, as S6M)
  S16-corecteaza--refuzat           a value over the hens alive (D39): below 0 on the record's day and on every later day;
                                    field error + the preview becomes a fault Banner (spec E1 state), save disabled
  S16-corecteaza--sterge-confirmare Dialog "Șterge Mortalitate … din 17.11.2026" over the untouched sheet: required Motiv,
                                    effect, stays in Istoric modificări, keys stacked, focus on Anulează (D39)
- desktop 1440×1024 (18.11.2026 after 10:12, owner): S4 · Păsări (s4t_build, the table-first page) beside the system
  SidePanel, scrolled to the Jurnal; the 17.11 Mortalitate row is the current one (its Corectează key controls the panel)
  and carries the correction mark; the record as it is now (corrected, "corectat · modificat de Ion Rusu · 18.11.2026,
  10:12"), nothing typed, Salvează corecția disabled with its reason.
python3 -I s16_build.py <repo_root> [s16_heights.json] [--canvas]
Writes design/canvas/project/S16-corecteaza[--<state>].dc.html, /tmp/s16/<stem>.html (renders) and, with --canvas only,
the S16 entries of canvas.json (canvas_rows.py: entries keep x/y; a new row goes below the lowest). Imports s4t_build (which loads
s4s_build; neither writes on import) and sstate_css (s13-/s9-/s7s- screen pieces shared with S9, S10, S13)."""
import sys, re, json, pathlib, importlib, datetime as dt

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
import vkey as V   # VO02: the voice key on every signed-in board (spec §8, D199; vkey.add in the writer)
ROOT, HP, CANVAS = CR.args('s16_heights.json'); ROOT = ROOT.resolve()  # heights default: s16_heights.json
HERE = pathlib.Path(__file__).resolve().parent
HS = json.loads(HP.read_text()) if HP and HP.exists() else {}
sys.path.insert(0, str(HERE))
sys.argv = ['s4t_build.py', str(ROOT)]
T = importlib.import_module('s4t_build')   # S4 tabs (Păsări Jurnal with SYS09 keys); loads s4s_build by spec
s4 = T.s4
import sstate_css as SC
sys.argv = ['s6_build.py', str(ROOT)]
Z6 = importlib.import_module('s6_build')   # D172: the save's EntryCards (SYS05) as the Zilnic form builds them; rewrites S6/S7 byte-identical

n, dd, icon, hen, roo, hen_v, NB = s4.n, s4.ddate, s4.icon, s4.hen, s4.roo, s4.hen_v, s4.NB
s4.IC.update({'x': '<path d="M18 6 6 18"/><path d="m6 6 12 12"/>',                                      # SidePanel close (S9)
              'alert': '<circle cx="12" cy="12" r="10"/><path d="M12 8v4"/><path d="M12 16h.01"/>'})  # fault Banner (S7S)
def num(v): return f'<span class="ip-num">{v}</span>'

# ---------- data (RF08: realdata, the shared illustrative correction D158a) ----------
R = s4.R
COR = R.ILL['corrections'][0]
assert COR['record'] == 'mortality' and COR['action'] == 'update' and COR['user'] == 'Ion Rusu' and COR['row']['sex'] == 'F'
SEC, RDAY = COR['row']['section'], COR['row']['date']
HALL = SEC.split('.')[0]
OLD, NEW = COR['old_values']['qty'], COR['new_values']['qty']
REASON = COR['reason']; assert len(REASON) <= 200
AT_D, AT_T = COR['at'][:10], COR['at'][11:16]
assert R.sd(RDAY, SEC, 'F')['mort'] == NEW and R.sd(RDAY, SEC, 'M')['mort'] == 0   # the daily rows hold the corrected value
OTHER = sum(abs(c['qty']) for c in R.ILL['culls'] + R.ILL['sexing_errors'] if (c['date'], c['section'], c['sex']) == (RDAY, SEC, 'F'))
START = R.alive_start(SEC, 'F', RDAY)
ENTRY = next(e for e in R.EVENTS if e['record'] == 'mortality' and e['at'] == COR['entry_at'])
assert ENTRY['user'] == COR['entry_user'] == 'Dr. Ceban'
ENTRY_T = ENTRY['at'][11:16]
WK = R.sapt(RDAY)
# D172: the record is the whole Zilnic save of that day (every compartment × sex in it, one save = one record, D168)
SAVE_SECS = [x for x in R.HSECS if R.sd(RDAY, x, 'F') or R.sd(RDAY, x, 'M')]
assert SAVE_SECS == [x for h in R.HHALLS for x in Z6.secs_with_birds(h, R.D(RDAY))] and SEC in SAVE_SECS   # H7 sold out 14.11: not in it
KEYS = [(x, sx) for x in SAVE_SECS for sx in 'FM']
STORED = {k: Z6.day(k[0], k[1], R.D(RDAY))['mortality'] for k in KEYS}              # the save as it is now (corrected, real)
assert STORED[(SEC, 'F')] == NEW and all(sum(STORED[(x, sx)] for x in SAVE_SECS) == R.deaths(RDAY, sx=sx) for sx in 'FM')
BEFORE = {**STORED, (SEC, 'F'): OLD}                                                # as Dr. Ceban saved it at 08:07
assert f"{len(SAVE_SECS)} secțiuni · mortalitate ♀ {sum(STORED[(x, 'F')] for x in SAVE_SECS)}" in ENTRY['summary']

def days_to(until): return [d for d in (R.D(RDAY) + dt.timedelta(days=i) for i in range((R.D(until) - R.D(RDAY)).days + 1))]
def alive_by_day(until, x):
    """SEC ♀ alive at the end of each day from the record's day to `until` if the record held x deaths (stored rows have NEW)."""
    return {d.isoformat(): R.alive(SEC, 'F', d) + NEW - x for d in days_to(until)}
# phones: today 18.11.2026 (the correction's day), before 10:12
PDAY = AT_D; assert R.D(PDAY) == R.TODAY
PRE, POST = alive_by_day(PDAY, OLD), alive_by_day(PDAY, NEW)
assert list(PRE) == [RDAY, PDAY] and POST[PDAY] == R.alive(SEC, 'F')
DELTA = NEW - OLD; assert DELTA > 0
TRY = START - OTHER + 6  # D39 state example: a typed value 6 over the hens alive at the start of the record's day
LOW = alive_by_day(PDAY, TRY); FIRST_NEG = min(d for d, v in LOW.items() if v < 0); assert FIRST_NEG == RDAY
DEL = alive_by_day(PDAY, 0)                                          # Șterge: the record's deaths go back
MOST = min(DEL.values()); BIND = min(DEL, key=DEL.get)               # the most deaths that leave no day below 0, and the day that binds
# desktop: 18.11.2026 after 10:12 (owner); the record as it is now (corrected)
DDAY = PDAY
NOW = alive_by_day(DDAY, NEW); assert NOW[DDAY] == s4.R.alive(SEC, 'F')
def save_alive(vals, until=None):
    """{(sec, sx): {day: alive at the end of day}} from the record's day to `until` (today) if the save held `vals` deaths."""
    return {k: {d.isoformat(): R.alive(k[0], k[1], d) + STORED[k] - vals[k] for d in days_to(until or R.TODAY)} for k in KEYS}
def farm_today(vals, sx): return R.alive_farm(sx) + sum(STORED[k] - vals[k] for k in KEYS if k[1] == sx)
CORR = dict(STORED)                                                                 # the phone's typed save = the real figures
TRYV = {**STORED, (SEC, 'F'): TRY}
def refuz_runs(low, today):
    """the days below 0 grouped by figure: [(first, last, figure)], today kept on its own ("azi")."""
    runs = []
    for d, v in low.items():
        if v >= 0: continue
        if runs and runs[-1][2] == v and runs[-1][1] != today and d != today: runs[-1][1] = d
        else: runs.append([d, d, v])
    return runs
def refuz_sentence(low, today, sec=SEC):
    parts = []
    for a, b, v in refuz_runs(low, today):
        when = (f'azi, {num(dd(b))}' if b == today else f'pe {num(dd(a))}' if a == b else f'pe {num(R.drange(a, b))}')
        parts.append(f'{hen_v(n(v))} găini {when}')
    return f'Corecția ar lăsa {sec} cu ' + ' și cu '.join(parts) + f'. Nicio secțiune nu poate scădea sub{NB}0.'

# ---------- pieces ----------
def rec(k='mortality', size=14): return f'<span class="ip-rec ip-rec--{k} s13-rec">{icon(k, size)}</span>'
TILE = f'<span class="ip-tile ip-rec ip-rec--mortality" aria-hidden="true">{icon("mortality")}</span>'
TITLE = 'Corectează Mortalitate'   # D172: the whole save (its compartments are the cards)
CTX = f'{"Hîrbovăț".upper()} · {dd(RDAY)} · {R.SAPT(WK)}'   # S16P: the Mono context line in caps, as the S9/S10/S13 panel heads
def head(touch):
    t = ' ip-iconbtn--touch' if touch else ''
    return (f'<div class="ip-sidepanel__head"><div class="ip-row ip-sidepanel__title">{TILE}<div><h2 class="ip-t-heading ip-m0" id="cp-t">{TITLE}</h2>'
            f'<p class="ip-note ip-num ip-m0">{CTX}</p></div></div><button type="button" class="ip-iconbtn{t}" aria-label="Închide Corectează">{icon("x")}</button></div>')
BY = f'<p class="ip-note ip-m0">Introdus de {ENTRY["user"]} · {num(dd(RDAY) + ", " + ENTRY_T)} · Zilnic · {num(len(SAVE_SECS))} secțiuni</p>'
MOD = (f'<div class="ip-ev__mod"><span class="ip-tag ip-tag--neutral">corectat</span><span>{SEC} {hen(n(-OLD))} → {hen_v(n(-NEW))}</span>'
       f'<a class="ip-link" href="#istoric-{SEC}-{R.ddm(RDAY)}">modificat de {COR["user"]} · {num(dd(AT_D) + ", " + AT_T)}<span class="ip-sr">, vezi istoricul</span></a></div>')

HEN_I = {'F': hen, 'M': roo}
def cards(typed, saved, meta=None, focus=False, err=None):
    """D172: one Zilnic EntryCard (s6_build.entry, SYS05: Mortalitate / Sacrificare ♀ ♂, Mai mult · Eroare sexare, Fără pierderi)
    per compartment of the save, holding `typed` Mortalitate. A figure that differs from the saved one is marked: Tag
    "modificat" in its card head and the hint "salvat: n" under its CountInput. meta: read-only (saved meta foot, s6s).
    focus: just opened, focus on the first CountInput. err: (sec, sx, message) replaces the card's own refusal line (the
    correction checks every later day too)."""
    out = []
    for x in SAVE_SECS:
        uid = x.replace('.', '')
        h, t, _ = Z6.entry(x, R.D(RDAY), ov={('mortality', sx): typed[(x, sx)] for sx in 'FM'}, saved=meta)
        h = h.replace('>Fără pierderi azi<', '>Fără pierderi<')   # the save's day, not today
        changed = [sx for sx in 'FM' if typed[(x, sx)] != saved[(x, sx)]]
        for sx in changed:
            hid = f'h{Z6.CLS[sx][0]}-{uid}'; i = h.index(f'<span class="ip-entry__hint" id="{hid}">'); j = h.index('</span></span></span>', i) + 21
            h = h[:i] + f'<span class="ip-entry__hint" id="{hid}"><span>salvat: <span class="ip-num">{n(saved[(x, sx)])}</span></span></span>' + h[j:]
        code = f'id="t-{uid}">{x}</h3>'; assert h.count(code) == 1
        i = h.index(code) + len(code); j = h.index('<span class="ip-entry__alive">', i); tag = h[i:j]
        tags = ('<span class="ip-tag ip-tag--info">modificat</span>' if changed else '') + tag
        if tags: h = h[:i] + f'<span class="ip-entry__tags">{tags}</span>' + h[j:]   # D173: the head's tags in one row with a gap
        if err and err[0] == x:
            mid = f'e-{uid}-{Z6.CLS[err[1]][0]}'; i = h.index(f'id="{mid}">'); i = h.index('<span>', h.index('</svg>', i)); j = h.index('</p>', i)
            h = h[:i] + f'<span>{err[2]}</span>' + h[j:]
        out.append(h)
    if focus:   # D173: the first editable CountInput (a "Fără pierderi" card stays locked until unticked, as on Zilnic)
        k = next(i for i, c in enumerate(out) if ' readonly' not in c[c.index('<input id="m0-'):c.index('>', c.index('<input id="m0-'))])
        c = out[k]; i = c.index('<input id="m0-'); a_ = c.rindex('<span class="ip-input ip-count ip-count--hen', 0, i)
        c = c[:a_] + c[a_:i].replace('ip-count--hen', 'ip-count--hen is-focus', 1) + c[i:].replace('<input ', '<input autofocus="autofocus" ', 1)
        out[k] = c
    return out
WHY = (f'peste cele {num(n(START - OTHER))} găini în viață la începutul zilei de {num(dd(RDAY))}' if BIND == RDAY else
       f'ar lăsa {hen_v(n(LOW[BIND]))} găini {"azi, " if BIND == PDAY else "pe "}{num(dd(BIND))}')
ERR = (SEC, 'F', f'{SEC} {hen(n(TRY))}: {WHY}. Scrie cel mult {num(n(MOST))}.')

LIM = 200   # Motiv soft limit (spec E1 "≤ 200 characters", D102): no maxlength, so a paste is never cut silently
def cnt(k, word):  # Romanian: "de" after numbers whose last two digits are 00 or 20–99 (12 caractere, 212 caractere, 262 de zile)
    return f'{num(k)} {"" if 0 < k % 100 < 20 or k < 20 else "de "}{word}'
assert cnt(212, 'caractere').endswith('</span> caractere') and cnt(262, 'zile').endswith('</span> de zile') and cnt(200, 'x').endswith(' de x')
def over_msg(i, text, verb='salva'):  # the soft-limit error: the length, the limit and how much to cut
    return (f'<p class="ip-err ip-fieldmsg" id="{i}">{icon("alert", 14)}<span>Motivul are {cnt(len(text), "caractere")}; cel mult {num(LIM)}. '
            f'Scurtează-l cu {num(len(text) - LIM)} ca să poți {verb}.</span></p>')
def motiv(text='', touch=True, err=False, rows=3):
    """Motiv of the sheet / panel. err: missing (on blur / save). Over LIM characters: the soft-limit error (save disabled
    by the caller, aria-describedby="cw-e")."""
    over = len(text) > LIM; assert not (err and over)
    t = ' ip-input--touch' if touch else ''
    cls = ' is-error is-focus' if err else ' is-error' if over else ''
    dsc = 'cw-e cw-h' if err or over else 'cw-h'
    inv = ' aria-invalid="true"' if err or over else ''
    msg = (f'<p class="ip-err ip-fieldmsg" id="cw-e">{icon("alert", 14)}<span>Motiv lipsă. Scrie de ce se schimbă {SEC} {s4.F} din {num(OLD)} în {num(NEW)}; '
           f'fără motiv corecția nu se salvează.</span></p>') if err else over_msg('cw-e', text) if over else ''
    return (f'<div class="ip-field"><label class="ip-label" for="cw">Motiv</label><span class="ip-input ip-input--area{t}{cls}">'
            f'<textarea id="cw" rows="{rows}" required="required" aria-required="true"{inv} aria-describedby="{dsc}">{text}</textarea></span>{msg}'
            f'<span class="ip-hint" id="cw-h">Obligatoriu, cel mult {LIM} de caractere · {num(f"{len(text)}/{LIM}")} · apare în Istoric modificări</span></div>')
def dlg_motiv(text='', touch=True):
    """Motiv of the Șterge Dialog: prefilled from the sheet's Motiv when one was typed (S16P), else empty; the same soft
    limit and counter. Returns (field markup, Șterge key attributes): disabled with the reason until Motiv has text, or
    while it is over the limit."""
    t = ' ip-input--touch' if touch else ''
    over = len(text) > LIM
    why = (f'Obligatoriu, cel mult {LIM} de caractere · {num(f"{len(text)}/{LIM}")}'
           + ('' if text else ' · scrie motivul ca să poți șterge'))
    err = over_msg('dlg-why-e', text, 'șterge') if over else ''
    dsc = 'dlg-why-e dlg-why-h' if over else 'dlg-why-h'
    inv = ' aria-invalid="true"' if over else ''
    f = (f'<div class="ip-field"><label class="ip-label" for="dlg-why">Motiv</label><span class="ip-input ip-input--area{t}{" is-error" if over else ""}">'
         f'<textarea id="dlg-why" rows="3" required="required" aria-required="true"{inv} aria-describedby="{dsc}">{text}</textarea></span>'
         f'{err}<span class="ip-hint" id="dlg-why-h">{why}</span></div>')
    key = f' disabled aria-describedby="{dsc}"' if (not text or over) else ''
    return f, key

ARROW = '<span class="s9-to" aria-hidden="true">→</span><span class="ip-sr">după corecție</span>'
def fxline(label, a, b, sx='F'):
    return f'<span class="s9-fxline"><b class="s9-strong">{label}</b><span class="s9-fxval">{HEN_I[sx](n(a))}{ARROW}{HEN_I[sx](n(b))}</span></span>'
def fxbox(inner): return (f'<div class="ip-effect s13-effect" id="fx" role="status">{rec("mortality", 16)}<div class="s9-fxbody">'
                          f'<span class="ip-cap">Efect</span>{inner}</div></div>')
def effect(state='set', typed=CORR, saved=BEFORE, today=PDAY):
    """EffectPreview of the correction (D172: the whole save): for each changed compartment × sex the spec sentence and its
    birds alive before → after on the save's day and every later day up to today (the never-below-0 check runs over each),
    then the totals that follow: the save's Mortalitate and the farm's birds alive today. empty: nothing changed yet.
    refuz: the fault Banner naming the first compartment that would go below 0."""
    if state == 'empty':
        return fxbox(f'<span>Schimbă o valoare: aici vezi cum se schimbă păsările în viață din secțiunile salvării, de pe {num(dd(RDAY))} până azi.</span>')
    post = save_alive(typed, today)
    if state == 'refuz':
        k = next(k for k in KEYS if min(post[k].values()) < 0)
        return (f'<div class="ip-banner ip-banner--fault s7s-banner" role="alert" id="fx">{icon("alert", 16)}<div class="s7s-banner__body"><p class="ip-m0">'
                f'{refuz_sentence(post[k], today, k[0])}</p></div></div>')
    pre = save_alive(saved, today); ch = [k for k in KEYS if typed[k] != saved[k]]; assert ch
    W = {'F': 'găini', 'M': 'cocoși'}
    body = ''.join(f'<span>{x} · {W[sx]} în viață de pe {num(dd(RDAY))} se schimbă cu {HEN_I[sx](n(saved[(x, sx)] - typed[(x, sx)]))}.</span>'
                   + ''.join(fxline(f'{x} · {"azi, " if d == today else ""}{dd(d)}', pre[(x, sx)][d], post[(x, sx)][d], sx) for d in pre[(x, sx)])
                   for x, sx in ch)
    for sx in sorted({k[1] for k in ch}, key='FM'.index):
        body += fxline(f'Salvarea · morți', -sum(saved[k] for k in KEYS if k[1] == sx), -sum(typed[k] for k in KEYS if k[1] == sx), sx)
        body += fxline(f'Hîrbovăț · azi', farm_today(saved, sx), farm_today(typed, sx), sx)
    nd = len(pre[ch[0]])
    return fxbox(body + f'<span class="s9-fxsub">Verificat pe {cnt(nd, "zile" if nd > 1 else "zi")}, până azi: nicio zi sub{NB}0.</span>')

SAVE = '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch">Salvează corecția</button>'
def save_off(why): return f'<button type="button" class="ip-btn ip-btn--primary ip-btn--touch" disabled aria-describedby="{why}">Salvează corecția</button>'
DELK = '<button type="button" class="ip-btn ip-btn--destructive ip-btn--touch" aria-haspopup="dialog">Șterge</button>'

def sheet(body, save, inert=False):
    ine = ' inert="inert"' if inert else ''
    return (f'<aside class="ip-sidepanel ip-sidepanel--sheet" role="dialog" aria-modal="true" aria-labelledby="cp-t"{ine}>{head(True)}'
            f'<div class="ip-sidepanel__body ip-sidepanel__body--form">{"".join(body)}</div>'
            f'<div class="ip-sidepanel__foot">{DELK}{save}</div></aside>')
def screen(sh, overlay=''): return f'<div class="s16-screen ip-scrimhost">{sh}{overlay}</div>'

boards = []   # (slug, state, body, tall): tall = the sheet at its natural height (D172: the save's cards), else one 844 viewport
# 1 · telefon · 18.11.2026 before 10:12 — Ion Rusu types the real figure into the save and the reason, before saving at 10:12
boards.append(('', f'telefon · înainte de {AT_T}', screen(sheet([BY, *cards(CORR, BEFORE), motiv(REASON), effect()], SAVE)), True))
# 2 · motiv lipsă — the same value, Motiv left empty: error on blur, focus on the field, Salvează corecția disabled (as S6M, D79)
boards.append(('motiv-lipsa', 'motiv lipsă · exemplu de stare (D39)', screen(sheet([BY, *cards(CORR, BEFORE), motiv(err=True), effect()], save_off('cw-e'))), True))
# 3 · refuzat — TRY typed (D39): below 0 on the save's day and on every later day; preview → fault Banner, save disabled
boards.append(('refuzat', 'refuzat · sub 0 · exemplu de stare (D39)', screen(sheet([BY, *cards({**BEFORE, (SEC, 'F'): TRY}, BEFORE, err=ERR), motiv(), effect('refuz', TRYV, BEFORE)], save_off('fx'))), True))
# 4 · Șterge · confirmare — Dialog over the untouched sheet (the save as saved at 08:07, nothing typed); Șterge deletes the whole
#   save (one record, D168/D172); reason required (spec E1), focus on Anulează. Its Motiv takes the sheet's Motiv when one was
#   typed (S16P); here none was, so it opens empty and Șterge waits for it.
ZERO = {k: 0 for k in KEYS}
def del_note(vals, corrected=False):
    tf, tm = (sum(vals[k] for k in KEYS if k[1] == sx) for sx in 'FM')
    who = f'introdusă de {ENTRY["user"]} la {num(ENTRY_T)}' + (f' și corectată de {COR["user"]} pe {num(dd(AT_D))}' if corrected else '')
    return (f'Se șterge salvarea Zilnic {who}: {num(len(SAVE_SECS))} secțiuni · {hen(n(-tf))} · {roo(n(-tm))}. '
            f'Găini în viață azi în Hîrbovăț: {hen(n(farm_today(vals, "F")))} → {hen(n(farm_today(ZERO, "F")))}. Ștergerea rămâne în Istoric modificări.')
DMOT, DKEY = dlg_motiv('')
dlg = (f'<div class="ip-scrim ip-scrim--phone"><div class="ip-dialog" role="alertdialog" aria-modal="true" aria-labelledby="dlg-t" aria-describedby="dlg-n">'
       f'<h2 id="dlg-t">Șterge Mortalitate din {num(dd(RDAY))}</h2>'
       f'<p class="ip-dialog__note" id="dlg-n">{del_note(BEFORE)}</p>'
       f'{DMOT}'
       f'<div class="ip-dialog__keys ip-dialog__keys--stack"><button type="button" class="ip-btn ip-btn--secondary ip-btn--touch ip-focus" autofocus="autofocus">Anulează</button>'
       f'<button type="button" class="ip-btn ip-btn--destructive ip-btn--touch"{DKEY}>Șterge</button></div></div></div>')  # fix 1: off until Motiv has text (E1)
assert DKEY == ' disabled aria-describedby="dlg-why-h"'
untouched = sheet([BY, *cards(BEFORE, BEFORE), motiv(), effect('empty')], save_off('fx'), inert=True)
boards.append(('sterge-confirmare', 'Șterge · confirmare · exemplu de stare (D39)', screen(untouched, dlg), False))

# ---------- desktop: SidePanel beside S4 · Păsări (owner, 18.11.2026 after 10:12, the table-first page), scrolled to the Jurnal ----------
# The S4 Jurnal lists one Mortalitate row per Zilnic save (s4t_build.log_row) and Corectează opens that save (D172), so the 17.11 row
# is the current one. Its "corectat" mark on the S4 page itself waits for a later S4 job (s4t log_row; S4NS owns s4t now).
JNAME = f'Mortalitate, {dd(RDAY)}'                                   # s4t_build.log_row's accessible name of that save
def jurnal_key(pg, name=JNAME):
    k = re.search(r'<a class="ip-btn ip-btn--ghost" href="#corecteaza-(ev\d+)">Corectează<span class="ip-sr"> ' + re.escape(name) + '</span></a>', pg)
    assert k and pg.count(k.group(0)) == 1, name
    return k.group(0)
def mark_row(pg, mod='', current=True, name=JNAME):
    """the Jurnal row of the corrected save: the correction mark after its meta (components.md › EventRow marks) and, while the
    panel is open, is-current + aria-current on its Corectează key (FIN05, additions §51; aria-expanded/-controls kept)."""
    key = jurnal_key(pg, name); i = pg.index(key); li = pg.rindex('<li class="ip-ev ip-ev--act">', 0, i)
    assert '</li>' not in pg[li:i]
    row = pg[li:i]; j = row.rindex('</span><span class="ip-ev__acts">'); row = row[:j] + mod + row[j:]
    if current:
        row = row.replace('<li class="ip-ev ip-ev--act">', '<li class="ip-ev ip-ev--act is-current">', 1)
        key = key.replace('ip-btn--ghost"', 'ip-btn--ghost" aria-expanded="true" aria-controls="cp" aria-current="true"')
    return pg[:li] + row + key + pg[i + len(jurnal_key(pg, name)):]
T.EVID[0] = 0
PG = mark_row(T.page('pasari', T.panel_pasari()))
SCROLL = HS.get('S16-corecteaza--desktop@scroll', 0)
assert PG.count('<main class="ip-shell__main s4-main">') == 1
if SCROLL: PG = PG.replace('<main class="ip-shell__main s4-main">', f'<main class="ip-shell__main s4-main" style="margin-top:calc(-1 * {SCROLL}px)">')
DESK = f'<div class="ip-dockpage s16-vp">{PG}</div>'
DSAVE = '<button type="button" class="ip-btn ip-btn--primary" disabled aria-describedby="fx">Salvează corecția</button>'
def panel(focus):  # focus: just opened from the Jurnal key, focus on the first CountInput (S16P); Esc / Închide return it to that key
    return f'''<aside class="ip-sidepanel ip-sidepanel--dock" id="cp" aria-labelledby="cp-t">
{head(False)}
<div class="ip-sidepanel__body ip-sidepanel__body--form">
<div>{BY}{MOD}</div>
{"".join(cards(STORED, STORED, focus=focus))}
{motiv(touch=False)}
{effect('empty')}
</div>
<div class="ip-sidepanel__foot"><button type="button" class="ip-btn ip-btn--destructive" aria-haspopup="dialog">Șterge</button><div class="ip-row ip-sidepanel__keys"><button type="button" class="ip-btn ip-btn--secondary">Anulează</button>{DSAVE}</div></div>
</aside>'''
PANEL = panel(True)

# ---------- CSS (screen only, tokens only) ----------
PAIR = '''
.s16-pair{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:var(--space-3)}
.ip-entry__tags{display:flex;align-items:center;gap:var(--space-2)} /* D173: several head tags (modificat, refuzat, abatere) in one row */'''
SHEET = ''  # FIN15B: the sheet's scrolling form body is additions.css §62 ip-sidepanel__body--form
CSS_P = '''body{margin:0}
.s16-screen{height:100%;display:flex}''' + SHEET + SC.M0 + SC.REC + SC.EFFECT + SC.FX + SC.BANNER + SC.MODAL + PAIR
CSS_D = T.CSS + SC.SIDE + SC.M0 + SC.REC + SC.EFFECT + SC.FX + PAIR

# ---------- write ----------
STEM0, TITLE0 = 'S16-corecteaza', 'S16 · Corectează'
SYS = ROOT / 'design/system'
FONTS = s4.FONTS
pathlib.Path('/tmp/s16').mkdir(exist_ok=True)
def write(stem, title, root, css, w, h):
    root = V.add(root)   # VO02
    dc = f'''<!doctype html>
<html lang="ro">
<head>
<meta charset="utf-8">
<title>{title}</title>
<script src="./support.js"></script>
<link rel="stylesheet" href="ds/impexlinepanel/tokens.css">
<link rel="stylesheet" href="ds/impexlinepanel/components/bundle.css">
<link rel="stylesheet" href="additions.css">
</head>
<body>
<x-dc>
<helmet>
{FONTS}
<style>{css}</style>
</helmet>
{root}
</x-dc>
<script type="text/x-dc" data-dc-script data-props='{{"$preview":{{"width":{w},"height":{h}}}}}'>
class Component extends DCLogic {{ renderVals() {{ return {{}}; }} }}
</script>
</body>
</html>
'''
    (ROOT / f'design/canvas/project/{stem}.dc.html').write_text(dc, encoding='utf-8')
    wrap = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8"><title>{title}</title>
<link rel="stylesheet" href="file://{SYS}/tokens.css"><link rel="stylesheet" href="file://{SYS}/components/bundle.css"><link rel="stylesheet" href="file://{SYS}/components/additions.css">
{FONTS}<style>{css}</style></head><body>{root}</body></html>'''
    pathlib.Path(f'/tmp/s16/{stem}.html').write_text(wrap, encoding='utf-8')
    print('ok', stem, w, h)
    return {'stem': stem, 'title': title, 'w': w, 'h': h}

out = []
def phone_h(stem, tall): return max(844, HS.get(stem, 844)) if tall else 844   # D172: a tall sheet board at its natural height (s16_render.py --heights)
def phone_root(body, h): return f'<div class="ip-root" style="padding:0; width:390px; height:{h}px; overflow:hidden">{body}</div>'
def desk_h(stem): return max(SC.SIDE_H, HS.get(stem, SC.SIDE_H))                    # the panel at its natural height (min the 1024 viewport)
for slug, state, body, tall in boards:
    stem = STEM0 + (f'--{slug}' if slug else ''); h = phone_h(stem, tall)
    out.append(write(stem, f'{TITLE0} · {state}', phone_root(body, h), CSS_P, 390, h))
DH = desk_h(STEM0 + '--desktop')
out.append(write(STEM0 + '--desktop', f'{TITLE0} · desktop', f'<div class="ip-root s4s-rel" style="padding:0; min-height:{DH}px">{DESK}{PANEL}</div>', CSS_D, 1440, DH))
if HP:
    hs = dict(HS); hs.update({b['stem']: b['h'] for b in out}); HP.write_text(json.dumps(hs, indent=1) + '\n')

if CANVAS:  # S16 row, entries keep x/y, a new row goes below the lowest (canvas_rows.py); 80 px gaps
    c = CR.load(ROOT)
    x = CR.place(c, STEM0, [(b['stem'], b['w'], b['h'], b['title']) for b in out])
    CR.save(ROOT, c)
    print('canvas: S16 row y', CR.row_y(c, STEM0), 'ends at x', x - 80)
