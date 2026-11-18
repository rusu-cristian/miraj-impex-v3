"""S16 · Corectează (spec E1, screens.md S16): the entry with its record icon, editable fields, the effect preview, the reason
(Motiv, required, 200 characters as a soft limit: no maxlength, the counter stays, over 200 = "Scurtează-l cu n"
and save disabled, D102/S16P), Salvează corecția / Șterge (reason + Dialog). Manager and owner only (spec §1).
Content: the one real change in the extracts (sample-hirbovat-2026.json audit_log, SYS09): Mortalitate Hîrbovăț H4.1 ♀ of
14.11.2026 (Zilnic saved by Dr. Ceban at 08:36), corrected 3 → 5 by Ion Rusu on 15.11.2026 at 10:12, reason "Numărătoare
refăcută: 2 găini găsite moarte sub hrănitoare." The daily rows hold the corrected figure (5), so the values before the
correction are the stored ones + 2 (asserted). Every figure is sample (D27); names only Dr. Ceban / Ion Rusu (D96).
Boards
- phones (390×844, the SidePanel as a full-screen sheet, additions.css §39), replay of the correction on 15.11.2026 before
  10:12 (as S9/S13 replay their records, D73/D83):
  S16-corecteaza                    ♀ 3 → 5 typed, the real reason, effect "H4.1 · găini în viață de pe 14.11.2026 se schimbă
                                    cu −2" (♀ 4 121 → 4 119 on 14.11 and on 15.11), Salvează corecția live
  S16-corecteaza--motiv-lipsa       the same value with no reason: Motiv error, Salvează corecția disabled (D39, as S6M)
  S16-corecteaza--refuzat           ♀ 4 130 typed (D39): 4 124 alive at the start of 14.11 → −6 on 14.11 and every later day;
                                    field error + the preview becomes a fault Banner (spec E1 state), save disabled
  S16-corecteaza--sterge-confirmare Dialog "Șterge Mortalitate H4.1 din 14.11.2026" over the untouched sheet: required Motiv,
                                    effect ♀ 4 121 → 4 124, stays in Istoric modificări, keys stacked, focus on Anulează (D39)
- desktop 1440×1024 (16.11.2026, D23, owner): S4 · Păsări (s4t_build) reflowed beside the system SidePanel, scrolled to the
  Jurnal; the H4.1 14.11 row's Corectează key pressed (controls the panel); the record as it is now (♀ 5, "corectat ·
  modificat de Ion Rusu · 15.11.2026, 10:12"), nothing typed, Salvează corecția disabled with its reason.
python3 -I s16_build.py <repo_root> [s16_heights.json] [--canvas]
Writes design/canvas/project/S16-corecteaza[--<state>].dc.html, /tmp/s16/<stem>.html (renders) and, with --canvas only,
the S16 entries of canvas.json (canvas_rows.py: entries keep x/y; a new row goes below the lowest). Imports s4t_build (which loads
s4s_build; neither writes on import) and sstate_css (s13-/s9-/s7s- screen pieces shared with S9, S10, S13)."""
import sys, re, json, pathlib, importlib, datetime as dt

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
ROOT, HP, CANVAS = CR.args('s16_heights.json'); ROOT = ROOT.resolve()  # heights default: s16_heights.json
HERE = pathlib.Path(__file__).resolve().parent
HS = json.loads(HP.read_text()) if HP and HP.exists() else {}
sys.path.insert(0, str(HERE))
sys.argv = ['s4t_build.py', str(ROOT)]
T = importlib.import_module('s4t_build')   # S4 tabs (Păsări Jurnal with SYS09 keys); loads s4s_build by spec
s4 = T.s4
import sstate_css as SC

n, dd, icon, hen, roo, hen_v, NB = s4.n, s4.ddate, s4.icon, s4.hen, s4.roo, s4.hen_v, s4.NB
s4.IC.update({'x': '<path d="M18 6 6 18"/><path d="m6 6 12 12"/>',                                      # SidePanel close (S9)
              'alert': '<circle cx="12" cy="12" r="10"/><path d="M12 8v4"/><path d="M12 16h.01"/>'})  # fault Banner (S7S)
def num(v): return f'<span class="ip-num">{v}</span>'

# ---------- data ----------
S = s4.S
AUD = S['audit_log'][0]
assert AUD['row'] == 'mortality · H4.1 · F · 2026-11-14' and AUD['action'] == 'update' and AUD['user'] == 'Ion Rusu'
SEC, RDAY = 'H4.1', '2026-11-14'
OLD, NEW = -AUD['old_values']['qty'], -AUD['new_values']['qty']; assert (OLD, NEW) == (3, 5)
REASON = AUD['reason']; assert REASON == 'Numărătoare refăcută: 2 găini găsite moarte sub hrănitoare.' and len(REASON) <= 200
AT_D, AT_T = AUD['at'][:10], AUD['at'][11:]; assert (AT_D, AT_T) == ('2026-11-15', '10:12')
DAILY = {(r['date'], r['sex']): r for r in S['daily'] if r['section'] == SEC}
r14f, r14m = DAILY[(RDAY, 'F')], DAILY[(RDAY, 'M')]
assert r14f['mortality'] == NEW and r14m['mortality'] == 0            # the daily rows hold the corrected value
START14 = r14f['alive_start']; OTHER14 = r14f['culls'] + r14f['sexing_errors']
ENTRY = next(e for e in S['event_log'] if e['record'] == 'mortality' and e['action'] == 'create' and e['at'][:10] == RDAY)
assert (ENTRY['user'], ENTRY['at'][11:]) == ('Dr. Ceban', '08:36')
WK = (dt.date.fromisoformat(RDAY) - dt.date.fromisoformat(S['meta']['cycle_start'])).days // 7 + 1
assert WK == r14f['cycle_week'] == 40

def alive_by_day(until, x):
    """H4.1 ♀ alive at the end of each day from 14.11 to `until` if the 14.11 record held x deaths (stored rows have NEW)."""
    out = {}
    for d in sorted({k[0] for k in DAILY if RDAY <= k[0] <= until}):
        out[d] = DAILY[(d, 'F')]['alive_end'] + NEW - x
    return out
# phone replay: "today" = 15.11.2026 (the correction's day), before 10:12
PDAY = AT_D
PRE, POST = alive_by_day(PDAY, OLD), alive_by_day(PDAY, NEW)
assert list(PRE) == ['2026-11-14', '2026-11-15'] and set(PRE.values()) == {4121} and set(POST.values()) == {4119}
DELTA = NEW - OLD; assert DELTA == 2
TRY = 4130  # D39 state example: a typed value over the birds alive
LOW = alive_by_day(PDAY, TRY); FIRST_NEG = min(d for d, v in LOW.items() if v < 0); assert (FIRST_NEG, LOW[FIRST_NEG]) == (RDAY, -6)
DEL = alive_by_day(PDAY, 0); assert set(DEL.values()) == {4124}       # Șterge: the 3 deaths go back
MOST = min(DEL.values()); assert MOST == START14 - OTHER14 == 4124     # the most deaths that leave no day below 0
# desktop: 16.11.2026 (D23), owner; the record as it is now (corrected), 16.11's own deaths in the later days
DDAY = s4.TODAY; assert DDAY == '2026-11-16'
NOW = alive_by_day(DDAY, NEW); assert NOW[DDAY] == s4.alive[(SEC, 'F')] == 4117   # = counts-latest

# ---------- pieces ----------
def rec(k='mortality', size=14): return f'<span class="ip-rec ip-rec--{k} s13-rec">{icon(k, size)}</span>'
TILE = f'<span class="ip-tile ip-rec ip-rec--mortality" aria-hidden="true">{icon("mortality")}</span>'
TITLE = f'Corectează Mortalitate {SEC}'
CTX = f'{"Hîrbovăț".upper()} · {dd(RDAY)} · SAPT {WK}'   # S16P: the Mono context line in caps, as the S9/S10/S13 panel heads
assert CTX == 'HÎRBOVĂȚ · 14.11.2026 · SAPT 40'
def head(touch):
    t = ' ip-iconbtn--touch' if touch else ''
    return (f'<div class="ip-sidepanel__head"><div class="ip-row ip-sidepanel__title">{TILE}<div><h2 class="ip-t-heading ip-m0" id="cp-t">{TITLE}</h2>'
            f'<p class="ip-note ip-num ip-m0">{CTX}</p></div></div><button type="button" class="ip-iconbtn{t}" aria-label="Închide Corectează">{icon("x")}</button></div>')
BY = f'<p class="ip-note ip-m0">Introdus de {ENTRY["user"]} · {num(dd(RDAY) + ", " + ENTRY["at"][11:])} · Zilnic</p>'
MOD = (f'<div class="ip-ev__mod"><span class="ip-tag ip-tag--neutral">corectat</span><span>{hen(n(-OLD))} → {hen_v(n(-NEW))}</span>'
       f'<a class="ip-link" href="#istoric-H4.1-14.11">modificat de {AUD["user"]} · {num(dd(AT_D) + ", " + AT_T)}<span class="ip-sr">, vezi istoricul</span></a></div>')

def counts(f, m, saved_f, saved_m, touch, err=False, focus=False):
    """Mortalitate ♀ / ♂ of the record: two CountInputs side by side, the saved value under each (hint).
    focus: the sheet / panel just opened, focus on the first CountInput (S16P; demo is-focus + autofocus)."""
    md = '' if touch else ' ip-count--md'
    fa = ' aria-invalid="true"' if err else ''
    fd = 'cf-fh cf-e' if err else 'cf-fh'
    fe = ' is-error' if err else ' is-focus' if focus else ''
    if focus: fa += ' autofocus="autofocus"'
    def fld(i, sx, v, saved, extra='', cls='', dsc=''):
        c, s, w = ('hen', '♀', 'găini') if sx == 'F' else ('rooster', '♂', 'cocoși')
        return (f'<div class="ip-field"><label class="ip-label" for="{i}">Mortalitate <span class="ip-sex ip-sex--{c}" aria-hidden="true">{s}</span><span class="ip-sr">{w}</span></label>'
                f'<span class="ip-input ip-count ip-count--{c}{md}{cls}"><input id="{i}" inputmode="numeric" placeholder="–" autocomplete="off" value="{v}"{extra} aria-describedby="{dsc}"/></span>'
                f'<span class="ip-hint" id="{i}h">salvat: {num(saved)}</span></div>')
    out = f'<div class="s16-pair">{fld("cf-f", "F", f, saved_f, fa, fe, fd)}{fld("cf-m", "M", m, saved_m, dsc="cf-mh")}</div>'
    if err:
        out += (f'<p class="ip-err ip-fieldmsg" id="cf-e">{icon("alert", 14)}<span>{SEC} {hen(n(TRY))}: peste cele {num(n(START14 - OTHER14))} găini în viață '
                f'la începutul zilei de {num(dd(RDAY))}. Scrie cel mult {num(n(MOST))}.</span></p>')
    return out

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
    msg = (f'<p class="ip-err ip-fieldmsg" id="cw-e">{icon("alert", 14)}<span>Motiv lipsă. Scrie de ce se schimbă {SEC} {s4.F} pe {num(dd(RDAY))} din {num(OLD)} în {num(NEW)}; '
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
def fxline(label, a, b):
    return f'<span class="s9-fxline"><b class="s9-strong">{label}</b><span class="s9-fxval">{hen(n(a))}{ARROW}{hen(n(b))}</span></span>'
def effect(state='set', today=PDAY):
    """EffectPreview of the correction: the spec sentence, H4.1 ♀ alive before → after on the record's day and every later
    day up to today (the never-below-0 check runs over each). empty: nothing changed yet. refuz: the fault Banner."""
    if state == 'empty':
        return (f'<div class="ip-effect s13-effect" id="fx" role="status">{rec("mortality", 16)}<div class="s9-fxbody"><span class="ip-cap">Efect</span>'
                f'<span>Schimbă o valoare: aici vezi cum se schimbă păsările în viață din {SEC}, de pe {num(dd(RDAY))} până azi.</span></div></div>')
    if state == 'refuz':
        neg = [d for d, v in LOW.items() if v < 0]; assert neg[0] == FIRST_NEG and len({LOW[d] for d in neg}) == 1
        days = ' și pe '.join(num(dd(d)) for d in neg)
        return (f'<div class="ip-banner ip-banner--fault s7s-banner" role="alert" id="fx">{icon("alert", 16)}<div class="s7s-banner__body"><p class="ip-m0">'
                f'Corecția ar lăsa {SEC} cu {hen_v(n(LOW[FIRST_NEG]))} găini pe {days}. Nicio secțiune nu poate scădea sub{NB}0.</p></div></div>')
    pre, post = PRE, POST
    lines = ''.join(fxline(f'{SEC} · {"azi, " if d == today else ""}{dd(d)}', pre[d], post[d]) for d in pre)
    return (f'<div class="ip-effect s13-effect" id="fx" role="status">{rec("mortality", 16)}<div class="s9-fxbody"><span class="ip-cap">Efect</span>'
            f'<span>{SEC} · găini în viață de pe {num(dd(RDAY))} se schimbă cu {hen_v(n(-DELTA))}.</span>{lines}'
            f'<span class="s9-fxsub">Verificat pe {num(len(pre))} zile, până azi: nicio zi sub{NB}0.</span></div></div>')

SAVE = '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch">Salvează corecția</button>'
def save_off(why): return f'<button type="button" class="ip-btn ip-btn--primary ip-btn--touch" disabled aria-describedby="{why}">Salvează corecția</button>'
DELK = '<button type="button" class="ip-btn ip-btn--destructive ip-btn--touch" aria-haspopup="dialog">Șterge</button>'

def sheet(body, save, inert=False):
    ine = ' inert="inert"' if inert else ''
    return (f'<aside class="ip-sidepanel ip-sidepanel--sheet" role="dialog" aria-modal="true" aria-labelledby="cp-t"{ine}>{head(True)}'
            f'<div class="ip-sidepanel__body ip-sidepanel__body--form">{"".join(body)}</div>'
            f'<div class="ip-sidepanel__foot">{DELK}{save}</div></aside>')
def screen(sh, overlay=''): return f'<div class="s16-screen ip-scrimhost">{sh}{overlay}</div>'

boards = []
# 1 · telefon · reluare 15.11.2026 — Ion Rusu types 5 and the real reason, before saving at 10:12
boards.append(('', 'telefon · reluare 15.11.2026', screen(sheet([BY, counts(n(NEW), '0', OLD, 0, True), motiv(REASON), effect()], SAVE))))
# 2 · motiv lipsă — the same value, Motiv left empty: error on blur, focus on the field, Salvează corecția disabled (as S6M, D79)
boards.append(('motiv-lipsa', 'motiv lipsă · exemplu de stare (D39)', screen(sheet([BY, counts(n(NEW), '0', OLD, 0, True), motiv(err=True), effect()], save_off('cw-e')))))
# 3 · refuzat — ♀ 4 130 typed (D39): below 0 on 14.11 and on every later day; preview → fault Banner, save disabled
boards.append(('refuzat', 'refuzat · sub 0 · exemplu de stare (D39)', screen(sheet([BY, counts(n(TRY), '0', OLD, 0, True, err=True), motiv(), effect('refuz')], save_off('fx')))))
# 4 · Șterge · confirmare — Dialog over the untouched sheet (♀ 3 · ♂ 0, nothing typed); reason required (spec E1), focus on Anulează.
#   Its Motiv takes the sheet's Motiv when one was typed (S16P); here none was, so it opens empty and Șterge waits for it.
DMOT, DKEY = dlg_motiv('')
dlg = (f'<div class="ip-scrim ip-scrim--phone"><div class="ip-dialog" role="alertdialog" aria-modal="true" aria-labelledby="dlg-t" aria-describedby="dlg-n">'
       f'<h2 id="dlg-t">Șterge Mortalitate {SEC} din {num(dd(RDAY))}</h2>'
       f'<p class="ip-dialog__note" id="dlg-n">Se șterge rândul introdus de {ENTRY["user"]} la {num(ENTRY["at"][11:])}: {hen(n(-OLD))} · {roo("0")}. '
       f'Găini în viață în {SEC} de pe {num(dd(RDAY))}: {hen(n(PRE[RDAY]))} → {hen(n(DEL[RDAY]))}. Ștergerea rămâne în Istoric modificări.</p>'
       f'{DMOT}'
       f'<div class="ip-dialog__keys ip-dialog__keys--stack"><button type="button" class="ip-btn ip-btn--secondary ip-btn--touch ip-focus" autofocus="autofocus">Anulează</button>'
       f'<button type="button" class="ip-btn ip-btn--destructive ip-btn--touch"{DKEY}>Șterge</button></div></div></div>')  # fix 1: off until Motiv has text (E1)
assert DKEY == ' disabled aria-describedby="dlg-why-h"'
untouched = sheet([BY, counts(n(OLD), '0', OLD, 0, True), motiv(), effect('empty')], save_off('fx'), inert=True)
boards.append(('sterge-confirmare', 'Șterge · confirmare · exemplu de stare (D39)', screen(untouched, dlg)))

# ---------- desktop: SidePanel beside S4 · Păsări (owner, 16.11.2026), scrolled to the Jurnal ----------
T.EVID[0] = 0
PG = T.page('pasari', T.panel_pasari())
KEY = re.search(r'<a class="ip-btn ip-btn--ghost" href="#corecteaza-(ev\d+)">Corectează<span class="ip-sr"> Mortalitate H4\.1, 14\.11\.2026</span></a>', PG)
assert KEY and PG.count(KEY.group(0)) == 1
# FIN05: the record being corrected is the Jurnal's current row (additions §51): is-current on its EventRow, aria-current on
# its Corectează key (aria-expanded/-controls kept, not drawn pressed), as the S19/S21/S22 lists mark their open record
_i = PG.index(KEY.group(0)); _li = PG.rindex('<li class="ip-ev ip-ev--act">', 0, _i)
assert '</li>' not in PG[_li:_i]
PG = (PG[:_li] + '<li class="ip-ev ip-ev--act is-current">' + PG[_li + len('<li class="ip-ev ip-ev--act">'):_i]
      + KEY.group(0).replace('ip-btn--ghost"', 'ip-btn--ghost" aria-expanded="true" aria-controls="cp" aria-current="true"') + PG[_i + len(KEY.group(0)):])
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
{counts(n(NEW), '0', NEW, 0, False, focus=focus)}
{motiv(touch=False)}
{effect('empty')}
</div>
<div class="ip-sidepanel__foot"><button type="button" class="ip-btn ip-btn--destructive" aria-haspopup="dialog">Șterge</button><div class="ip-row ip-sidepanel__keys"><button type="button" class="ip-btn ip-btn--secondary">Anulează</button>{DSAVE}</div></div>
</aside>'''
PANEL = panel(True)

# ---------- CSS (screen only, tokens only) ----------
PAIR = '''
.s16-pair{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:var(--space-3)}'''
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
for slug, state, body in boards:
    stem = STEM0 + (f'--{slug}' if slug else '')
    out.append(write(stem, f'{TITLE0} · {state}', f'<div class="ip-root" style="padding:0; width:390px; height:844px; overflow:hidden">{body}</div>', CSS_P, 390, 844))
DH = SC.SIDE_H
out.append(write(STEM0 + '--desktop', f'{TITLE0} · desktop', f'<div class="ip-root s4s-rel" style="padding:0; min-height:{DH}px">{DESK}{PANEL}</div>', CSS_D, 1440, DH))
if HP:
    hs = dict(HS); hs.update({b['stem']: b['h'] for b in out}); HP.write_text(json.dumps(hs, indent=1) + '\n')

if CANVAS:  # S16 row, entries keep x/y, a new row goes below the lowest (canvas_rows.py); 80 px gaps
    c = CR.load(ROOT)
    x = CR.place(c, STEM0, [(b['stem'], b['w'], b['h'], b['title']) for b in out])
    CR.save(ROOT, c)
    print('canvas: S16 row y', CR.row_y(c, STEM0), 'ends at x', x - 80)
