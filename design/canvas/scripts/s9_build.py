"""S9 · Transfer (manager, spec C3): Data, Din (SectionSelect with alive counts), În, ♀/♂ counts with Toate, the effect
preview including the cycle that opens or closes, Mută.
Phone boards (390×844, D60) reuse the S7 phone form shell by import (s7s_build → s7_build: FormHeader, DateField, icons,
SaveBar shell + 2-line summary, viewport/Toast/Banner CSS from sstate_css, and S13's SidePanel/rec/effect CSS from it
with the s13- class names); the Din row is the system SexCountRow
`ip-sexrows--pick` (SYS07), the preview the system EffectPreview with the CycleTrack cycle marks (`ip-cmark`, SYS06).
The desktop board reuses the S4 Hîrbovăț farm page (s4s_build shell, sidebar, writer) with the system SidePanel beside it.
Content (RF05, realdata; D151–D158): the phone boards replay the real last Telița → Hîrbovăț move (transfers-seria2026.json,
the H2 link, arrivals as recorded at Hîrbovăț on 28.02.2026, D28): H2 → H7.2 ♀ 4 383, the H2 → H7.1 ♀ 2 618 move entered
first that day, so H2 holds ♀ 4 383 and Toate moves them all; H7.2 holds the ♂ 534 that came on 22.02. Telița then has 0
birds (workbook residuals are data errors, D25/D35): its cycle closes on 28.02.2026, sapt. 21 (cycles.json); Hîrbovăț's open
cycle (from 22.02.2026) continues, sapt. 21 (the birds' age week, D151: both farms share it). "Today" on the phone boards =
28.02.2026 (as S13 replays its delivery, D73). The refusal value ♀ 4 500 is a D39 state example. The desktop board is
18.11.2026 (D152) over the S4 page: the Din menu open, alive at the end of the day (realdata; H7 empty since 14.11).
S9P: two more phone boards on 18.11.2026, a D39 example on the real counts: Toate from Hîrbovăț H6.2 (♀ 475 · ♂ 5) into
Telița (empty since its cycle closed on 28.02.2026): "alege secțiunea" (the phone SectionSelect BottomSheet, additions.css §33)
and "ciclu nou" (EffectPreview "Ciclu nou la Telița, sapt. 59", the age week (D151), with the CycleTrack open mark).
python3 -I s9_build.py <repo_root> [heights.json, default s9_heights.json] [--canvas]
Writes design/canvas/project/S9-transfer[--<state>].dc.html, /tmp/s9/<stem>.html and /tmp/s4s/<stem>.html (renders);
--canvas refreshes the S9 entries' h/title in canvas.json in place (hand-placed row: x and y kept, RF05)."""
import sys, re, json, pathlib, importlib, datetime as dt

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
import vkey as V   # VO02: the voice key on every signed-in board (spec §8, D199; vkey.add in the writer)
ROOT, HP, CANVAS = CR.args('s9_heights.json')  # heights default: s9_heights.json
HERE = pathlib.Path(__file__).resolve().parent

# phone shell: s7s_build imports s7_build (both rewrite their S7 boards, unchanged)
sys.argv = ['s7s_build.py', str(ROOT)]
sys.path.insert(0, str(HERE))
P = importlib.import_module('s7s_build'); B = P.B
import sstate_css as SC  # SidePanel-beside-the-page, record icon, EffectPreview CSS (s13- names, shared with s13_build)
import sexsign as SXS  # SX01: ♀/♂ outside a figure span = the §72 hook (sexsign.py)
import realdata as R  # RF05: every figure, date, week and code from realdata (DATA03 extracts)
# desktop shell: the S4 farm page (s4s_build writes nothing on import)
sys.argv = ['s4s_build.py', str(ROOT)]
s4 = importlib.import_module('s4s_build')

n, dd, lamp, NB, MINUS = B.n, B.dd, B.lamp, B.NB, B.MINUS
B.IC.update({'transfer': s4.IC['transfer'], 'x': '<path d="M18 6 6 18"/><path d="m6 6 12 12"/>', 'chevup': '<path d="m18 15-6-6-6 6"/>'})
icon = B.icon

# ---------- data (realdata, RF05) ----------
# the last Telița → Hîrbovăț link: H2's hens, recorded at Hîrbovăț on 28.02.2026 in two compartments (D28 arrivals)
LINK = R.TR['links_inferred_by_date'][-1]
assert LINK['arrival_date'] == max(l['arrival_date'] for l in R.TR['links_inferred_by_date']) and len(LINK['from']) == 1 and LINK['sex'] == 'F'
DAY = R.D(LINK['arrival_date'])
(FIRST_SEC, FIRST), (DST_SEC, QTY) = list(LINK['to'].items())  # the first move entered that day, then the replayed one (Toate)
SRC_F = LINK['qty_in'] - FIRST           # H2 after the first move the same day (arrivals, D28)
assert (dd(DAY), FIRST_SEC, FIRST, DST_SEC, QTY, SRC_F) == ('28.02.2026', 'H7.1', 2618, 'H7.2', 4383, 4383)
# destination before the move: what arrived there before 28.02 (♂ 534 on 22.02, no ♀ yet)
DST = {sx: sum(a['qty'] for a in R.TR['arrivals'] if a['section'] == DST_SEC and a['sex'] == sx and a['date'] < LINK['arrival_date']) for sx in 'FM'}
assert DST == {'F': 0, 'M': 534}
TEL, HIR = R.CTEL, R.C26
HSTART = R.CYCLE_START                   # 22.02.2026 (SAPT 20)
assert TEL['end_date'] == LINK['arrival_date'] and TEL['status'] == 'closed'  # D25: Telița closes 28.02.2026
WK1 = R.sapt(DAY)                        # D151: the birds' age week, the same at both farms (transfer keeps age)
assert WK1 == TEL['sapt_end'] == 21 and HIR['status'] == 'open'
TRY = 4500  # D39 state example (typed value over the alive count)
OVER = SRC_F - TRY; assert OVER == -117
USER = 'Ion Rusu'  # illustrative manager name (craft §7)
# desktop + S9P: 18.11.2026 (D152), alive per compartment at the end of the day (Telița 0, cycle closed)
HSEC = [(s, R.alive(s, 'F'), R.alive(s, 'M')) for s in R.HSECS]
TSEC = R.SECTIONS['telita']
assert R.COUNTS_TELITA['totals']['all'] == 0 and len(TSEC) == 6
DT = R.TODAY; DWK = R.WK; assert dd(DT) == s4.ddate(s4.TODAY) and DWK == 59

def pair(f, m, sr=''): return B.pair(n(f) if isinstance(f, int) else f, n(m) if isinstance(m, int) else m, sr)
def hen(v): return f'<span class="ip-sex ip-sex--hen">♀ <b>{n(v) if isinstance(v, int) else v}</b></span>'
def rec(k='transfer', size=14): return f'<span class="ip-rec ip-rec--{k} s13-rec">{icon(k, size)}</span>'
def cmark(kind, cls, text):  # SYS06 cycle mark: bar (record that caused it) + mini block (open/closed)
    bar = '<i class="ip-cmark__bar"></i>' if cls else ''
    return f'<span class="ip-cmark ip-cmark--{kind}{(" " + cls) if cls else ""}"><span class="ip-cmark__g" aria-hidden="true">{bar}<i class="ip-cmark__blk"></i></span>{text}</span>'
CHK = '<svg class="ip-icon ip-icon--14" viewBox="0 0 24 24" aria-hidden="true"><path d="M20 6 9 17l-5-5"/></svg>'

# ---------- phone form (S7 shell) ----------
assert 'href="#inregistreaza" aria-label="Înapoi la Înregistrează"' in B.HEAD  # FX07/D142: the S7 shell's back link
HEAD = B.HEAD.replace('ip-rec--feed', 'ip-rec--transfer').replace(icon('feed'), icon('transfer')).replace('Hrană zilnică', 'Transfer') \
    .replace(f'{B.FARM} · {dd(B.TODAY)} · SAPT {B.WEEK}', f'SERIA 2026 · {dd(DAY)}')
assert HEAD.count('Transfer') == 1 and dd(DAY) in HEAD and 'ip-rec--transfer' in HEAD and 'SAPT' not in HEAD
WKLINE = f'SAPT {WK1} · ciclu Hîrbovăț din {dd(HSTART)}'  # D151: one age week for both farms
DATE = B.DATE.replace(dd(B.TODAY), dd(DAY)).replace(f'SAPT {B.WEEK} · ciclu din {dd(B.START)}', WKLINE)
assert DATE.count(dd(DAY)) == 1 and WKLINE in DATE
DATE_RO = f'''<div class="ip-field"><label class="ip-label" for="d-val">Data</label><span class="ip-input ip-input--touch ip-num is-readonly"><input id="d-val" type="text" value="{dd(DAY)}" readonly aria-describedby="d-wk"/></span><span class="ip-note ip-num" id="d-wk">{WKLINE}</span></div>'''

SRC_SEC = LINK['from'][0]  # H2

class Mv:  # one move: day, Din farm/section + alive, În farm/section + alive, counts moved
    def __init__(s, day, sfarm, ssec, sf, sm, dfarm, dsec, df, dm, qf, qm, sid, did):
        s.day, s.sfarm, s.ssec, s.sf, s.sm, s.dfarm, s.dsec, s.df, s.dm, s.qf, s.qm = day, sfarm, ssec, sf, sm, dfarm, dsec, df, dm, qf, qm
        s.src, s.dst = f'{sfarm} › {ssec}', f'{dfarm} › {dsec}'
        s.sid, s.did = sid, did  # farm ids (counts-latest), for the farm totals after the move
# 1 · the real 28.02.2026 move (replay, D83)
M1 = Mv(DAY, 'Telița', SRC_SEC, SRC_F, 0, 'Hîrbovăț', DST_SEC, DST['F'], DST['M'], QTY, 0, 'telita', 'hirbovat')
# 2 · ciclu nou (D39 state example, D83): 18.11.2026, Toate from Hîrbovăț H6.2 (the nearly emptied compartment, end-of-day
#   counts) into Telița's first section; Telița is empty since its cycle closed on 28.02.2026 (cycles.json), so a new cycle
#   opens there in the birds' age week, sapt. 59 (D151: a transfer keeps age)
XSRC = 'H6.2'; X = next(x for x in HSEC if x[0] == XSRC); assert X[1:] == (475, 5)
M2 = Mv(DT, 'Hîrbovăț', XSRC, X[1], X[2], 'Telița', TSEC[0], 0, 0, X[1], X[2], 'hirbovat', 'telita')
HLEFT = (R.alive_farm('F') - M2.qf, R.alive_farm('M') - M2.qm); assert HLEFT == (12086, 789)
HEMPTY = {s for s, f, m in HSEC if not f + m}; assert HEMPTY == {'H7.1', 'H7.2'}  # H7 empty since 14.11.2026
SRC, DSTN = M1.src, M1.dst

def combo(i, label, name, touch=True, expanded=False, menu='', active='', desc=''):
    """SectionSelect combobox. Desktop: an APG select-only combobox over its listbox menu. Phone: the same combobox opens
    the BottomSheet (SectionSelect spec), so aria-haspopup="dialog" and aria-controls only while it is open (APG)."""
    t = ' ip-input--touch' if touch else ''
    nm = f'<span class="ip-select__val"><b>{name}</b></span>' if name else '<span class="ip-select__val is-empty">Alege secțiunea</span>'
    ad = f' aria-activedescendant="{active}"' if active else ''
    ds = f' aria-describedby="{desc}"' if desc else ''
    ex = "true" if expanded else "false"
    if touch:
        ctl = f' aria-controls="sh-{i}"' if expanded else ''
        return (f'<div class="ip-field ip-sexrow__sel"><span class="ip-label ip-sexrow__lbl" id="sl-{i}">{label}</span>'
                f'<div class="ip-select"><div class="ip-input ip-select__btn{t}" id="sc-{i}" role="combobox" tabindex="0" aria-haspopup="dialog" aria-expanded="{ex}"{ctl} aria-labelledby="sl-{i}"{ds}>{nm}{icon("chevup" if expanded else "chev", 16)}</div></div></div>')
    return (f'<div class="ip-field ip-sexrow__sel"><span class="ip-label ip-sexrow__lbl" id="sl-{i}">{label}</span>'
            f'<div class="ip-select"><div class="ip-input ip-select__btn{t}{" is-focus" if expanded else ""}" id="sc-{i}" role="combobox" tabindex="0" aria-haspopup="listbox" aria-expanded="{ex}" aria-controls="sm-{i}" aria-labelledby="sl-{i}"{ad}{ds}>{nm}{icon("chevup" if expanded else "chev", 16)}</div>'
            f'<ul class="ip-menu ip-select__menu" role="listbox" id="sm-{i}" aria-labelledby="sl-{i}" tabindex="-1"{"" if expanded else " hidden"}>{menu}</ul></div></div>')

def pick(state='toate', touch=True, date=None, mv=M1):
    """SexCountRow --pick: the Din row. state: toate (all alive, read-only counts) · refuz (♀ over alive) · ro (saved) · empty (desktop, menu open)."""
    md = '' if touch else ' ip-count--md'
    date = date or mv.day
    sname = f'{mv.sfarm} {mv.ssec}'  # accessible names put the section first (S6P)
    if state == 'empty':
        # Din lists every section; an empty one (Hîrbovăț H7 since 14.11, all of Telița) is disabled "goală" (nothing to move)
        def dopt(farm, s, f, m, act):
            dis = ' aria-disabled="true"' if not f + m else ''
            body = pair(f, m) if f + m else '<span class="ip-note">goală</span>'
            return (f'<li class="ip-menu__item ip-select__opt{" is-active" if act else ""}" role="option" id="so-din-{"t" if farm == "Telița" else ""}{s.replace(".", "-")}" aria-selected="false"{dis}>'
                    f'<span class="ip-select__chk" aria-hidden="true"></span><span class="ip-select__txt"><span class="ip-select__top"><span><span class="ip-sr">{farm} </span>{s}</span> {body}</span></span></li>')
        first = next(s for s, f, m in HSEC if f + m)
        menu = '<li class="ip-menu__group" role="presentation">Hîrbovăț · ouat</li>' + ''.join(dopt('Hîrbovăț', s, f, m, s == first) for s, f, m in HSEC) \
            + '<li class="ip-menu__group" role="presentation">Telița · creștere</li>' + ''.join(dopt('Telița', s, 0, 0, False) for s in TSEC)
        sel = combo('din', 'Din', '', touch, True, menu, f'so-din-{first.replace(".", "-")}')
        alive = '<span class="ip-sexrow__alive" id="a-din"><span class="ip-sexrow__k">în viață</span> <span class="ip-num">–</span> alege secțiunea</span>'
    elif state == 'ro':
        sel = f'<div class="ip-field ip-sexrow__sel"><label class="ip-label ip-sexrow__lbl" for="sc-din">Din</label><span class="ip-input ip-input--touch is-readonly"><input id="sc-din" type="text" value="{mv.src}" readonly/></span></div>'
        alive = f'<span class="ip-sexrow__alive" id="a-din"><span class="ip-sexrow__k">în viață înainte</span> {pair(mv.sf, mv.sm)}</span>'
    else:
        sel = combo('din', 'Din', mv.src, touch)
        alive = f'<span class="ip-sexrow__alive" id="a-din"><span class="ip-sexrow__k">în viață</span> {pair(mv.sf, mv.sm)}</span>'
    def lbl(i, sx):
        c, s, w = ('hen', '♀', 'găini') if sx == 'F' else ('rooster', '♂', 'cocoși')
        sec = 'Din' if state == 'empty' else sname
        return f'<label class="ip-label ip-sexrow__lbl" for="{i}"><span class="ip-sr">{sec} </span><span class="ip-sex ip-sex--{c}" aria-hidden="true">{s}</span><span class="ip-sr">{w}</span></label>'
    msg = ''
    if state == 'toate' or state == 'ro':
        ro = ' is-readonly' if state == 'ro' else ''
        dsc = 'a-din' if state == 'ro' else 'a-din all-din'
        f_in = f'<span class="ip-input ip-count ip-count--hen{md}{ro}"><input id="f-din" inputmode="numeric" placeholder="–" autocomplete="off" value="{n(mv.qf)}" readonly aria-describedby="{dsc}"/></span>'
        m_in = f'<span class="ip-input ip-count ip-count--rooster{md}{ro}"><input id="m-din" inputmode="numeric" placeholder="–" autocomplete="off" value="{n(mv.qm)}" readonly aria-describedby="{dsc}"/></span>'
    elif state == 'refuz':
        f_in = f'<span class="ip-input ip-count ip-count--hen{md} is-error"><input id="f-din" inputmode="numeric" placeholder="–" autocomplete="off" value="{n(TRY)}" aria-invalid="true" aria-describedby="a-din e-din"/></span>'
        m_in = f'<span class="ip-input ip-count ip-count--rooster{md}"><input id="m-din" inputmode="numeric" placeholder="–" autocomplete="off" aria-describedby="a-din"/></span>'
        msg = (f'<p class="ip-err ip-sexrow__msg" id="e-din">{icon("alert", 14)}<span>{mv.ssec}: <span class="ip-num">{n(TRY)}</span> depășește cele '
               f'<span class="ip-num">{n(mv.sf)}</span> găini în viață la {mv.sfarm} pe <span class="ip-num">{dd(date)}</span>. Introdu cel mult <span class="ip-num">{n(mv.sf)}</span>.</span></p>')
    else:
        f_in = f'<span class="ip-input ip-count ip-count--hen{md}"><input id="f-din" inputmode="numeric" placeholder="–" autocomplete="off" aria-describedby="a-din"/></span>'
        m_in = f'<span class="ip-input ip-count ip-count--rooster{md}"><input id="m-din" inputmode="numeric" placeholder="–" autocomplete="off" aria-describedby="a-din"/></span>'
    sr = '' if state == 'empty' else f'<span class="ip-sr">{sname} </span>'
    if state == 'ro': allk = ''
    elif state == 'toate': allk = f'<label class="ip-check ip-sexrow__all is-checked" id="all-din"><input type="checkbox" class="ip-sr" checked="checked"/><span class="ip-check__box">{CHK}</span>{sr}Toate</label>'
    else: allk = f'<label class="ip-check ip-sexrow__all" id="all-din"><input type="checkbox" class="ip-sr"/><span class="ip-check__box"></span>{sr}Toate</label>'
    return f'''<section class="ip-panel ip-sexrows ip-sexrows--pick" aria-labelledby="t-px"><div class="ip-sexrows__head"><h2 class="ip-cap" id="t-px">Păsări transferate</h2><span class="ip-note"><span class="ip-num">{dd(date)}</span></span></div>
<ol class="ip-sexrows__list"><li class="ip-sexrow"><div class="ip-sexrow__head">{sel}{alive}</div><div class="ip-field is-hen">{lbl("f-din", "F")}{f_in}</div><div class="ip-field is-rooster">{lbl("m-din", "M")}{m_in}</div>{msg}{allk}</li></ol></section>'''

TEMPTY = f'{dd(dt.date.fromisoformat(TEL["end_date"]))}'  # Telița empty since its cycle closed (28.02.2026)
def dest(state='set', touch=True, mv=M1):
    if state == 'ro':
        return f'<div class="ip-field s9-in"><label class="ip-label" for="sc-in">În</label><span class="ip-input ip-input--touch is-readonly"><input id="sc-in" type="text" value="{mv.dst}" readonly aria-describedby="a-in"/></span><span class="ip-note s9-alive" id="a-in">în viață înainte {pair(mv.df, mv.dm)}</span></div>'
    if state == 'empty':
        return f'<div class="ip-field s9-in">{combo("in", "În", "", touch, desc="a-in")}<span class="ip-note s9-alive" id="a-in">orice fermă; o fermă goală începe un ciclu nou</span></div>'
    if state == 'open':  # phone: the În combobox with its BottomSheet open (nothing chosen yet)
        return f'<div class="ip-field s9-in">{combo("in", "În", "", touch, expanded=True, desc="a-in")}<span class="ip-note s9-alive" id="a-in">orice fermă; o fermă goală începe un ciclu nou</span></div>'
    return f'<div class="ip-field s9-in">{combo("in", "În", mv.dst, touch, desc="a-in")}<span class="ip-note s9-alive" id="a-in">în viață {pair(mv.df, mv.dm)}</span></div>'

def effect(state='set', mv=M1):
    if state == 'empty':
        return (f'<div class="ip-effect s13-effect" id="fx" role="status">{rec("transfer", 16)}<div class="s9-fxbody"><span class="ip-cap">Efect</span>'
                '<span>Alege Din, În și câte păsări muți. Aici vezi ce rămâne în fiecare secțiune și ciclul care se deschide sau se închide.</span></div></div>')
    if state == 'refuz':  # EffectPreview: refused becomes a fault Banner
        return (f'<div class="ip-banner ip-banner--fault s7s-banner" role="alert" id="fx">{icon("alert", 16)}<div class="s7s-banner__body"><p class="ip-m0">'
                f'Transfer refuzat: {mv.src} ar ajunge la {hen(n(OVER))} pe <span class="ip-num">{dd(mv.day)}</span>. Nicio secțiune nu poate scădea sub{NB}0.</p></div></div>')
    arrow = '<span class="s9-to" aria-hidden="true">→</span><span class="ip-sr">după transfer</span>'
    if mv is M1:
        cyc = (f'<span class="s9-fxcyc">{cmark("end", "xfer", f"Ciclu închis la {mv.sfarm}, sapt. {WK1}")}<span class="s9-fxsub">ferma ajunge la 0 păsări</span></span>\n'
               f'<span class="s9-fxcyc">{cmark("start", "", f"Ciclul la {mv.dfarm} continuă, sapt. {WK1}")}</span>')
    else:  # M2: the source farm keeps birds (its cycle continues), the empty destination farm opens a new cycle in the age week (D151)
        cyc = (f'<span class="s9-fxcyc">{cmark("start", "", f"Ciclul la {mv.sfarm} continuă, sapt. {DWK}")}<span class="s9-fxsub">rămân {pair(*HLEFT)}</span></span>\n'
               f'<span class="s9-fxcyc">{cmark("start", "xfer", f"Ciclu nou la {mv.dfarm}, sapt. {DWK}")}<span class="s9-fxsub">ferma e goală</span></span>')
    return f'''<div class="ip-effect s13-effect" id="fx" role="status">{rec("transfer", 16)}<div class="s9-fxbody"><span class="ip-cap">Efect</span>
<span class="s9-fxline"><b class="s9-strong">{mv.src}</b><span class="s9-fxval">{pair(mv.sf, mv.sm)}{arrow}{pair(mv.sf - mv.qf, mv.sm - mv.qm)}</span></span>
<span class="s9-fxline"><b class="s9-strong">{mv.dst}</b><span class="s9-fxval">{pair(mv.df, mv.dm)}{arrow}{pair(mv.df + mv.qf, mv.dm + mv.qm)}</span></span>
{cyc}
</div></div>'''

TR14 = rec()
def sum1(f=QTY, m=0, lead='', mv=M1):
    p = pair(f, m)
    if lead: return f'{TR14}<span>{lead} · {p}</span>'
    return f'{TR14}<span>{p} ·</span><span>{mv.sfarm} → {mv.dfarm}</span>'
SUM2 = f'<span>{cmark("end", "xfer", "ciclul Telița se închide")}</span>'
SUM2_NEW = f'<span>{cmark("start", "xfer", "ciclu nou la Telița")}</span>'
MUTA = '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch">Mută</button>'
MUTA_OFF = '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch" disabled aria-describedby="sb-sum">Mută</button>'

def screen(parts, save, overlay='', head=None, view='', sheet=''):
    """s7s_build.screen with this form's header. view: ' ip-screen__view--end' scrolls the form to its end. sheet: a modal
    BottomSheet over the screen; the header, form and SaveBar behind it are inert (out of the tab order and the a11y tree)."""
    ine = ' inert' if sheet else ''
    h = head or HEAD
    if sheet: h = h.replace('<header class="ip-formhead">', '<header class="ip-formhead" inert>', 1); save = save.replace('<div class="ip-savebar', '<div inert class="ip-savebar', 1)
    return f'''<div class="ip-screen ip-screen--fixed{" ip-scrimhost" if sheet else ""}">
{h}
<div class="ip-screen__view{view}"{ine}>
<main class="ip-screen__main">
{"".join(parts)}
</main>
{overlay}
</div>
{save}{sheet}
</div>'''

FORM = lambda: [DATE, pick('toate'), dest(), effect()]
boards = []
# 1 · telefon — the real 28.02.2026 move typed with Toate, before Mută
boards.append(('', 'telefon · reluare 28.02.2026', screen(FORM(), B.savebar(sum1(), SUM2, MUTA))))
# 2 · salvat — read-only form, Toast with Anulează (undo the same day), Transfer nou
toast = SXS.sexify(f'<div class="ip-toastwrap"><div class="ip-toast" role="status">{icon("check", 16)}<span class="ip-toast__txt">Transfer salvat · '
         f'♀ <span class="ip-num">{n(QTY)}</span> · Telița {SRC_SEC} → Hîrbovăț {DST_SEC} · ciclul Telița închis</span>'
         f'<button type="button" class="ip-link ip-toast__undo">Anulează</button></div></div>')
saved = [DATE_RO, pick('ro'), dest('ro'), effect()]
boards.append(('salvat', 'salvat', screen(saved, B.savebar(sum1(lead='mutat'), f'<span>{USER} ·</span><span>poți anula până la <span class="ip-num">23:59</span></span>', '<button type="button" class="ip-btn ip-btn--secondary ip-btn--touch">Transfer nou</button>'), overlay=toast)))
# 3 · refuzat — ♀ 4 500 typed over the ♀ 4 383 alive (D39 state example): row error, preview = fault Banner, Mută disabled;
#   the refused value is left out of the SaveBar totals (D77, as S6S/S7S: ♀ – · ♂ –), its figure only on line 2
boards.append(('refuzat', 'refuzat · peste cei în viață · exemplu de stare (D39)', screen([DATE, pick('refuz'), dest(), effect('refuz')],
              B.savebar(sum1('–', '–'), f'{lamp("fault", "refuzat")}<span>{SRC_SEC} {hen(TRY)} ·</span><span>cel mult <span class="ip-num">{n(SRC_F)}</span></span>', MUTA_OFF), view=' ip-screen__view--end')))   # VO02: at its end (47 px lower) the voice key clears the refusal Banner
# 4 · eroare la salvare — online only (spec §1): fault Banner + Reîncearcă, typed values kept as a draft (D60). The Banner is
#   the form's last part, right above the SaveBar (D141/D181, as S6/S7/S8/S13/S16; was above the EffectPreview, S9P), and the
#   form is scrolled to its end, so the EffectPreview and the Banner both show above the SaveBar
ban = SXS.sexify(f'<div class="ip-banner ip-banner--fault ip-banner--act" role="alert">{icon("alert", 16)}<div class="ip-banner__body"><p class="ip-m0">Transferul nu s-a salvat: '
       f'♀ <span class="ip-num">{n(QTY)}</span>, Telița {SRC_SEC} → Hîrbovăț {DST_SEC}, <span class="ip-num">{dd(DAY)}</span>.</p>'
       f'<p class="ip-m0">Verifică conexiunea la internet, apoi apasă Reîncearcă. Ce ai scris rămâne păstrat în browser până la salvare.</p></div><button type="button" class="ip-btn ip-btn--secondary ip-btn--touch ip-banner__key ip-focus" autofocus="autofocus">Reîncearcă</button></div>')
boards.append(('eroare-la-salvare', 'eroare la salvare', screen([DATE, pick('toate'), dest(), effect(), ban], B.savebar(sum1(), SUM2, MUTA), view=' ip-screen__view--end')))

# ---------- 18.11.2026 example flow (D39, real end-of-day counts): Hîrbovăț H6.2 → Telița (empty since 28.02.2026) ----------
HEAD2 = HEAD.replace(f'SERIA 2026 · {dd(DAY)}', f'SERIA 2026 · {dd(DT)}'); assert dd(DT) in HEAD2
WK2 = f'SAPT {DWK} · Telița goală din {TEMPTY}'
DATE2 = B.DATE.replace(f'SAPT {B.WEEK} · ciclu din {dd(B.START)}', WK2); assert B.TODAY == DT and WK2 in DATE2
# 5 · alege secțiunea — the phone SectionSelect as a BottomSheet (modal dialog + listbox), opened from În with Din chosen:
#   Hîrbovăț sections with ♀ · ♂ alive (end of 18.11), the Din section disabled with its reason "sursa", empty sections
#   (Hîrbovăț H7.1/H7.2, all of Telița) say "goală" and stay enabled (a move onto an empty farm opens a cycle, C3)
def oid(i, farm, code): return f'so-{i}-{"t" if farm == "Telița" else ""}{code.replace(".", "-")}'
def opt(i, farm, code, body, dis=False, act=False):
    a = ' aria-disabled="true"' if dis else ''
    return f'<div class="ip-menu__item{" is-active" if act else ""}" role="option" id="{oid(i, farm, code)}" aria-selected="false"{a}><span class="ip-sheet__code"><span class="ip-sr">{farm} </span>{code}</span>{body}</div>'
def sheet(i='in', src=M2):
    # the listbox has focus on open (ip-focus = the ring on the board) and its first option is the active one
    # (aria-activedescendant + is-active, as the desktop Din menu): Hîrbovăț's first section
    act = oid(i, 'Hîrbovăț', HSEC[0][0]); assert HSEC[0][0] != src.ssec
    def body(s, f, m): return '<span class="ip-note">sursa</span>' if s == src.ssec else pair(f, m) if f + m else '<span class="ip-note">goală</span>'
    hg = ''.join(opt(i, 'Hîrbovăț', s, body(s, f, m), dis=(s == src.ssec), act=(k == 0)) for k, (s, f, m) in enumerate(HSEC))
    tg = ''.join(opt(i, 'Telița', s, '<span class="ip-note">goală</span>') for s in TSEC)
    return f'''
<!-- BottomSheet (modal): focus moves to the listbox on open, its first enabled option active (aria-activedescendant; arrow keys move it, Enter or a tap picks it and
     closes the sheet), Tab cycles between Închide and the listbox only (focus trap), Esc or Închide closes it and focus returns
     to the În combobox. The form behind is inert. -->
<div class="ip-scrim ip-scrim--sheet">
<div class="ip-sheet ip-sheet--select" role="dialog" aria-modal="true" id="sh-{i}" aria-labelledby="sh-{i}-t" aria-describedby="sh-{i}-d">
<div class="ip-sheet__handle" aria-hidden="true"></div>
<div class="ip-sheet__head"><div><h2 class="ip-t-heading ip-sheet__title" id="sh-{i}-t">În: alege secțiunea</h2><p class="ip-note ip-sheet__sub" id="sh-{i}-d">O fermă goală începe un ciclu nou.</p></div><button type="button" class="ip-iconbtn ip-iconbtn--touch" aria-label="Închide lista">{icon("x")}</button></div>
<div class="ip-sheet__list ip-focus" role="listbox" id="sl-{i}-lb" aria-labelledby="sh-{i}-t" tabindex="0" aria-activedescendant="{act}">
<div role="group" aria-labelledby="sg-{i}-h"><div class="ip-menu__group" id="sg-{i}-h">Hîrbovăț · ouat</div>{hg}</div>
<div role="group" aria-labelledby="sg-{i}-t"><div class="ip-menu__group" id="sg-{i}-t">Telița · creștere</div>{tg}</div>
</div>
</div>
</div>'''
boards.append(('alege-sectiunea', 'alege secțiunea · exemplu de stare (D39)', screen([DATE2, pick('toate', mv=M2), dest('open', mv=M2), effect('empty')],
              B.savebar(f'{TR14}<span>{pair(M2.qf, M2.qm)} ·</span><span>{M2.sfarm} → –</span>', '<span>alege secțiunea În</span>', MUTA_OFF), head=HEAD2, sheet=sheet())))
# 6 · ciclu nou — the same move with Telița's first section chosen: the EffectPreview opens a cycle at Telița in the birds'
#   age week, sapt. 59 (D151; CycleTrack open mark with the transfer bar), Hîrbovăț keeps birds and its cycle continues, sapt. 59
boards.append(('ciclu-nou', 'ciclu nou · exemplu de stare (D39)', screen([DATE2, pick('toate', mv=M2), dest(mv=M2), effect(mv=M2)],
              B.savebar(sum1(M2.qf, M2.qm, mv=M2), SUM2_NEW, MUTA), head=HEAD2, view=' ip-screen__view--end')))   # VO02: at its end (89 px lower) the voice key sits on the 88 px clearance, not on the Efect

SCSS = '''
.s9-in > .ip-sexrow__sel{flex:none}
.s9-alive .ip-pair{font-size:var(--ip-fs-figure);line-height:var(--ip-lh-small)}''' + SC.FX + SC.MODAL  # EffectPreview lines + modal anchor: shared with s10_build (s9- names kept)
CSS = P.CSS + SC.MAIN + SC.REC + SC.EFFECT + SCSS

# ---------- desktop: SidePanel beside S4 · Hîrbovăț (manager), 18.11.2026 ----------
DDATE = f'''<div class="ip-field"><label class="ip-label" for="p-date">Data</label><div class="ip-row s4s-daterow"><span class="ip-seg" role="radiogroup" aria-label="Zi rapidă"><span class="ip-seg__item is-selected" role="radio" aria-checked="true" tabindex="0">Azi</span><span class="ip-seg__item" role="radio" aria-checked="false" tabindex="-1">Ieri</span></span><span class="ip-input s4s-date"><input id="p-date" type="text" inputmode="numeric" value="{dd(DT)}" placeholder="zz.ll.aaaa" aria-describedby="p-date-h"/><button type="button" class="ip-iconbtn" aria-label="Alege data transferului din calendar" aria-haspopup="dialog" aria-expanded="false">{icon("cal", 16)}</button></span></div><span class="ip-hint ip-num" id="p-date-h">Hîrbovăț SAPT {DWK} · ciclu din {dd(HSTART)}</span></div>'''
PANEL = f'''<aside class="ip-sidepanel ip-sidepanel--dock" id="tp" aria-labelledby="tp-t">
<div class="ip-sidepanel__head"><div class="ip-row ip-sidepanel__title"><span class="ip-tile ip-rec ip-rec--transfer" aria-hidden="true">{icon("transfer")}</span><div><h2 class="ip-t-heading ip-m0" id="tp-t">Transfer</h2><p class="ip-note ip-num ip-m0">SERIA 2026 · {dd(DT)}</p></div></div><button type="button" class="ip-iconbtn" aria-label="Închide Transfer">{icon("x")}</button></div>
<div class="ip-sidepanel__body ip-sidepanel__body--form">
{DDATE}
{pick('empty', touch=False, date=DT)}
{dest('empty', touch=False)}
{effect('empty')}
</div>
<div class="ip-sidepanel__foot"><span class="ip-note" id="tp-why">Alege secțiunea din care muți</span><div class="ip-row ip-sidepanel__keys"><button type="button" class="ip-btn ip-btn--secondary">Anulează</button><button type="button" class="ip-btn ip-btn--primary" disabled aria-describedby="tp-why">Mută</button></div></div>
</aside>'''
PANEL = re.sub(r'((?:id|for|aria-describedby|aria-labelledby|aria-controls|aria-activedescendant)="[^"]*")', lambda m: re.sub(r'\b(sl|sc|sm|so|a|f|m|all|t|fx)-', r'p\1-', m.group(1)), PANEL)
DH = SC.SIDE_H
# the page reflows beside the panel (margin = panel-w, nothing under it, D73); its "Mută păsări" key shows pressed and
# controls the open panel, so the panel's Mută stays the one primary key in view
KEY = f'<button type="button" class="ip-btn ip-btn--secondary">{s4.icon("transfer", 16)}Mută păsări</button>'
SPG = s4.shell(s4.main_page(panel=s4.panel_default())); assert SPG.count(KEY) == 1
SPG = SPG.replace(KEY, KEY.replace('ip-btn--secondary"', 'ip-btn--secondary is-pressed" aria-expanded="true" aria-controls="tp"'))
DESK = f'<div class="ip-dockpage">{SPG}</div>'
CSS_D = s4.CSS + '\n' + s4.CSS_STATES + SC.SIDE + SC.M0 + SC.REC + SC.EFFECT + SCSS

# ---------- write ----------
STEM0, TITLE0 = 'S9-transfer', 'S9 · Transfer'
VH = P.VH
def write_phone(slug, state, body):
    stem = STEM0 + (f'--{slug}' if slug else ''); title = f'{TITLE0} · {state}'
    root = f'<div class="ip-root" style="padding:0; width:390px; height:{VH}px; overflow:hidden">{V.add(body)}</div>'
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
{B.FONTS}
<style>{CSS}</style>
</helmet>
{root}
</x-dc>
<script type="text/x-dc" data-dc-script data-props='{{"$preview":{{"width":390,"height":{VH}}}}}'>
class Component extends DCLogic {{ renderVals() {{ return {{}}; }} }}
</script>
</body>
</html>
'''
    (ROOT / f'design/canvas/project/{stem}.dc.html').write_text(dc, encoding='utf-8')
    SYS = ROOT / 'design/system'
    wrap = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8"><title>{title}</title>
<link rel="stylesheet" href="file://{SYS}/tokens.css"><link rel="stylesheet" href="file://{SYS}/components/bundle.css"><link rel="stylesheet" href="file://{SYS}/components/additions.css">
{B.FONTS}<style>{CSS}</style></head><body>{root}</body></html>'''
    pathlib.Path('/tmp/s9').mkdir(exist_ok=True)
    pathlib.Path(f'/tmp/s9/{stem}.html').write_text(wrap, encoding='utf-8')
    print('ok', stem, title, VH)
    return stem, title, VH

out = [write_phone(*b) for b in boards]
s4.write(STEM0 + '--desktop', TITLE0 + ' · desktop', DESK, PANEL, H=DH, css=CSS_D)
out.append((STEM0 + '--desktop', TITLE0 + ' · desktop', DH))
if HP:
    HP.write_text(json.dumps({s: h for s, _, h in out}, indent=1))
if CANVAS:  # RF05: refresh the S9 entries' h/title in place (hand-placed row: x and y kept, canvas_rows.place)
    c = CR.load(ROOT)
    CR.place(c, STEM0, [(s, 1440 if s.endswith('--desktop') else 390, h, t) for s, t, h in out])
    CR.save(ROOT, c)
