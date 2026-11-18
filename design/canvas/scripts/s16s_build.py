"""S16S · Corectează · state boards (spec E1, SYS09, D70/D71, D101; design/notes/S16.md state list), built on s16_build.py.
python3 -I s16s_build.py <repo_root> [--canvas]
Imports s16_build with s16_heights.json (it rewrites the 5 S16 boards byte-identical, never canvas.json), reuses its sheet head, record line, CountInputs, Motiv, EffectPreview, Dialog wording, desktop page + panel and
CSS; Toast + Anulează / Banner-with-action CSS from sstate_css (as S9/S10/S15S); desktop Dialog scrim = additions.css §50 (ip-scrim--page, FIN04).
Boards (phones 390×844, the sheet of additions.css §39; desktops 1440×1024 beside / over S4 · Păsări, 16.11.2026, D23):
- salvat                     15.11.2026, 10:12: the sheet as the record now is (♀ 5, "corectat · modificat de Ion Rusu ·
                             15.11.2026, 10:12"), nothing typed, Toast "Corecție salvată …" + Anulează
- eroare-la-salvare          the default draft kept (♀ 5 + the real reason), fault Banner + Reîncearcă (focused) above the preview
- se-incarca                 head known (from the row), static skeletons, keys disabled with the visible status line as reason
- vet-blocat                 Dr. Ceban on 15.11 opens her own 14.11 record: read-only, Tag "blocat" + reason (SYS09, D70)
- ciclu-transfer             the real 28.02.2026 move Telița H2 → Hîrbovăț H4.2 ♀ 4 383 (the S9 replay, D83) typed ♀ 4 367
                             (D39): H2 keeps the workbook's 16 hens (counts-latest residual, D35), so the Telița cycle does not
                             close; Hîrbovăț H4.2 −16 on every day to 16.11; Salvează corecția opens the cycle Dialog
- desktop-scris              (merged with Motiv over 200) ♀ 5 → 3 typed (the value before the 15.11 correction, D39) + a
                             212-character reason: preview +2 on 14–16.11, Motiv error, save disabled
- desktop-refuzat            ♀ 4 130 typed (the phone D39 value): −6 on 14.11 and 15.11, −8 on 16.11 → fault Banner
- desktop-sterge-confirmare  Dialog over page + panel (inert), Motiv typed (D39), Șterge enabled, focus on Anulează
- desktop-dupa-stergere      panel closed, page full width at the Jurnal without the H4.1 14.11 row, Toast + Anulează;
                             focus on the next row's Corectează; page regenerated with the record dropped (s4s DROP)
Writes design/canvas/project/S16-corecteaza--<state>.dc.html, /tmp/s16/<stem>.html, /tmp/s16/states.json and (--canvas) the canvas.json
entries: new phones after x 1410 (x 1880 … 3760), the S16 desktop moved to x 4230 (only its x), new desktops after it
(step 1520), as S9P/S10P moved their desktop; only when it adds a board (canvas_rows.py: a rerun keeps every x/y)."""
import sys, re, json, pathlib, importlib, importlib.util, datetime as dt

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
ROOT, _, CANVAS = CR.args(); ROOT = ROOT.resolve()
HERE = pathlib.Path(__file__).resolve().parent
HP = HERE / 's16_heights.json'
sys.path.insert(0, str(HERE))
sys.argv = ['s16_build.py', str(ROOT), str(HP)]
Q = importlib.import_module('s16_build')        # writes the 5 S16 boards (unchanged) and the S16 canvas entries
T, s4, SC = Q.T, Q.s4, Q.SC
HS = json.loads(HP.read_text())
n, dd, icon, hen, roo, hen_v, NB, num = Q.n, Q.dd, Q.icon, Q.hen, Q.roo, Q.hen_v, Q.NB, Q.num
SEC, RDAY, OLD, NEW, PDAY, DDAY, WK = Q.SEC, Q.RDAY, Q.OLD, Q.NEW, Q.PDAY, Q.DDAY, Q.WK
s4.IC.update({'transfer': s4.IC['transfer']})
EX = ROOT / 'design/data/extracts'

# ---------- data ----------
# desktop (16.11.2026): the record holds 5 (corrected); typed examples (D39) run the check over 14.11 … 16.11
BACK = OLD                                             # ♀ 5 → 3: the value before the 15.11 correction
DPRE, DBACK = Q.alive_by_day(DDAY, NEW), Q.alive_by_day(DDAY, BACK)
assert list(DPRE) == ['2026-11-14', '2026-11-15', '2026-11-16'] and DPRE[DDAY] == 4117 and DBACK[DDAY] == 4119
DLOW = Q.alive_by_day(DDAY, Q.TRY); assert [DLOW[d] for d in DLOW] == [-6, -6, -8]
DDEL = Q.alive_by_day(DDAY, 0); assert DDEL[DDAY] == 4122  # Șterge on 16.11: the 5 deaths go back
LONG = ('Revin la 3, cât a numărat Dr. Ceban pe 14.11.2026 la 08:36. Cele 2 găini găsite sub hrănitoare pe 15.11 '
        'nu se văd în fișa halei H4 din 14.11, deci le trec în ziua în care au fost găsite, nu în ziua de 14.11.2026.')
assert len(LONG) > 200, len(LONG)
DEL_WHY = 'Rândul se reintroduce din Zilnic, după numărătoarea refăcută în H4.1.'   # D39 reason, Dialog example

# the cycle board: the real last move (transfers-seria2026.json, D83 replay as S9): H2 → H4.2 ♀ 4 383 on 28.02.2026,
# which emptied Telița (its cycle closed, sapt. 21). counts-latest names the workbook residuals per Telița section; H2's 16
# hens are the typed example (D39): the move becomes ♀ 4 367, H2 keeps 16, the Telița cycle stays open.
TR = json.load(open(EX / 'transfers-seria2026.json', encoding='utf-8'))
CL = json.load(open(EX / 'counts-latest.json', encoding='utf-8'))
FS = json.load(open(EX / 'farms-sections.json', encoding='utf-8'))
LINK = next(l for l in TR['links_inferred_by_date'] if l['from'] == ['H2'])
XDAY, XQ = LINK['arrival_date'], LINK['to']['H4.2']; assert (XDAY, XQ) == ('2026-02-28', 4383)
TSECS = [s['code'] for s in FS['sections'] if s['farm_id'] == 'telita']
RES = re.search(r'residuals \(([^)]*)\)', CL['meta']['telita_note']).group(1).replace('−', '-').split('/')
RES = dict(zip(TSECS, map(int, RES))); assert RES['H2'] == 16
LEFT = RES['H2']; XNEW = XQ - LEFT; assert XNEW == 4367
H42 = {r['date']: r['alive_end'] for r in s4.S['daily'] if r['section'] == 'H4.2' and r['sex'] == 'F'}
H42[XDAY] = XQ                                         # no ♀ in H4.2 before this move (S9: DST ♀ 0)
assert min(H42) == XDAY and H42[DDAY] == 4107 and min(H42.values()) - LEFT > 0
XDAYS = len(H42); assert XDAYS == 262                   # 28.02 … 16.11.2026: every day checked
TWK = (dt.date.fromisoformat(XDAY) - dt.date(2025, 10, 7)).days // 7 + 1   # Telița placement 07.10.2025 (cycles.json, S9 TWK)
HWK = (dt.date.fromisoformat(XDAY) - dt.date.fromisoformat(s4.S['meta']['cycle_start'])).days // 7 + 1
assert (TWK, HWK) == (21, 3)

# ---------- pieces ----------
cnt = Q.cnt   # Romanian number + noun ("212 caractere", "262 de zile"), s16_build
NBS = '\u00a0'  # no-break space: "3 → 5", "♀ 4 119" stay on one line
BYMOD = f'<div>{Q.BY}{Q.MOD}</div>'
BYMOD_P = f'<div class="s16s-ct">{Q.BY}{Q.MOD}</div>'  # phone: a size container, so SYS09's < 480px rule gives the mark link its 48px hit area
ROTITLE = f'Mortalitate {SEC}'
def ro_head():  # read-only for the vet: the title names the record, not the action she can't take
    h = Q.head(True).replace(f'>{Q.TITLE}<', f'>{ROTITLE}<', 1).replace('aria-label="Închide Corectează"', f'aria-label="Închide {ROTITLE}"', 1)
    assert ROTITLE in h and 'Corectează' not in h
    return h
def sheet(head, body, foot, inert=False, busy=False, toast=''):  # toast: inside the dialog (focus trap, a11y tree), above the foot
    ine = ' inert="inert"' if inert else ''
    bz = ' aria-busy="true"' if busy else ''
    return (f'<aside class="ip-sidepanel ip-sidepanel--sheet" role="dialog" aria-modal="true" aria-labelledby="cp-t"{ine}>{head}'
            f'<div class="ip-sidepanel__body s13-pbody"{bz}>{"".join(body)}</div>'
            f'{toast}<div class="ip-sidepanel__foot">{foot}</div></aside>')
FOOT = Q.DELK + Q.SAVE
def toast(text, desk=False):
    w = 'ip-toastwrap ip-toastwrap--page' if desk else 'ip-toastwrap'  # §48 (FIN02): phone above the sheet foot, desktop bottom right
    return (f'<div class="{w}"><div class="ip-toast" role="status">{icon("check", 16)}<span class="ip-toast__txt">{text}</span>'
            f'<button type="button" class="ip-link ip-toast__undo">Anulează</button></div></div>')

boards = []   # (slug, state title, kind 'p' | 'd', root markup)
# 1 · salvat — 15.11.2026, 10:12: the sheet shows the record as it now is (as the desktop panel), Toast + Anulează
T_SAVED = f'Corecție salvată · {SEC} ♀ {num(f"{OLD}{NBS}→{NBS}{NEW}")} · în viață ♀{NBS}{num(n(Q.POST[PDAY]))}'
boards.append(('salvat', f'salvat · {dd(Q.AT_D)}, {Q.AT_T}', 'p', Q.screen(
    sheet(Q.head(True), [BYMOD_P, Q.counts(n(NEW), '0', NEW, 0, True), Q.motiv(), Q.effect('empty')], Q.DELK + Q.save_off('fx'), toast=toast(T_SAVED)))))

# 2 · eroare la salvare — online only (spec §1): the draft stays (D60), fault Banner + Reîncearcă right above the preview
#   (S9P), Salvează corecția stays enabled; the whole body fits above the foot (no scroll)
BAN = (f'<div class="ip-banner ip-banner--fault ip-banner--act" role="alert">{icon("alert", 16)}<div class="ip-banner__body">'
       f'<p class="s13-m0">Corecția nu s-a salvat: {SEC} · {num(dd(RDAY))} · {hen(f"{OLD}{NBS}→{NBS}{NEW}")}. Verifică conexiunea și apasă Reîncearcă.</p>'
       f'<p class="s13-m0">Ce ai scris rămâne păstrat în browser până la salvare.</p></div>'
       f'<button type="button" class="ip-btn ip-btn--secondary ip-btn--touch ip-banner__key ip-focus" autofocus="autofocus">Reîncearcă</button></div>')  # S16P: focus after a failed save
boards.append(('eroare-la-salvare', 'eroare la salvare', 'p', Q.screen(
    sheet(Q.head(True), [Q.BY, Q.counts(n(NEW), '0', OLD, 0, True), Q.motiv(Q.REASON), BAN, Q.effect()], FOOT))))

# 3 · se încarcă — the head comes from the row that opened the sheet; the record is unknown: static skeletons, keys off
def skf(box): return f'<div class="s16s-skf"><span class="ip-skel ip-skel--w3"></span><span class="ip-skel {box}"></span></div>'
# S16P: the status line is visible (in place of the "Introdus de …" line) and gives the disabled keys their reason
SKEL = (f'<p class="ip-note s13-m0" role="status" id="ld">Se încarcă rândul din {num(dd(RDAY))} · Șterge și Salvează corecția se activează după încărcare.</p>'
        f'<div class="s16s-sk" aria-hidden="true">'
        f'<div class="s16-pair">{skf("s16s-skbox")}{skf("s16s-skbox")}</div>{skf("s16s-skarea")}<span class="ip-skel s16s-skfx"></span></div>')
KEYS_OFF = ('<button type="button" class="ip-btn ip-btn--destructive ip-btn--touch" disabled aria-describedby="ld">Șterge</button>'
            '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch" disabled aria-describedby="ld">Salvează corecția</button>')
LHEAD = Q.head(True).replace('<button type="button" class="ip-iconbtn ip-iconbtn--touch"', '<button type="button" class="ip-iconbtn ip-iconbtn--touch ip-focus" autofocus="autofocus"', 1)
assert 'ip-focus' in LHEAD   # fix 1: focus stays on Închide while the record loads (components.md › S16)
boards.append(('se-incarca', 'se încarcă', 'p', Q.screen(sheet(LHEAD, [SKEL], KEYS_OFF, busy=True))))

# 4 · vet · zi trecută · blocat — Dr. Ceban (vet) on 15.11 opens her own Zilnic row of 14.11: past day, so read-only, Tag
#   "blocat" + the SYS09 reason, no Motiv, preview or Șterge (spec §1: vets change their own entries the same day only)
def ro_field(i, sx, v):
    c, s, w = ('hen', '♀', 'găini') if sx == 'F' else ('rooster', '♂', 'cocoși')
    return (f'<div class="ip-field"><label class="ip-label" for="{i}">Mortalitate <span class="ip-sex ip-sex--{c}" aria-hidden="true">{s}</span><span class="ip-sr">{w}</span></label>'
            f'<span class="ip-input ip-count ip-count--{c} is-readonly"><input id="{i}" inputmode="numeric" value="{v}" readonly="readonly" aria-describedby="vb-why"/></span></div>')
LOCK = (f'<p class="ip-note s16s-lock" id="vb-why"><span class="ip-tag ip-tag--neutral">blocat</span>'
        f'<span>zi trecută · o corectează managerul</span></p>')
VB_NOTE = (f'<p class="s13-m0">Poți modifica sau șterge ce ai introdus doar în aceeași zi, până la {num("23:59")}. '
           f'Pentru o corecție, spune-i managerului; el o salvează cu motiv.</p>')
boards.append(('vet-blocat', 'vet · zi trecută · blocat', 'p', Q.screen(sheet(ro_head(), [
    f'<div>{Q.BY}{LOCK}</div>', f'<div class="s16-pair">{ro_field("cf-f", "F", OLD)}{ro_field("cf-m", "M", 0)}</div>', VB_NOTE],
    '<button type="button" class="ip-btn ip-btn--secondary ip-btn--touch">Închide</button>'))))

# 5 · ciclu · transfer — correcting the move that closed the Telița cycle (D39 value, real record)
def cmark(kind, cls, text):  # SYS06 cycle mark (as s9_build): record bar (if it caused it) + mini block
    bar = '<i class="ip-cmark__bar"></i>' if cls else ''
    return f'<span class="ip-cmark ip-cmark--{kind}{(" " + cls) if cls else ""}"><span class="ip-cmark__g" aria-hidden="true">{bar}<i class="ip-cmark__blk"></i></span>{text}</span>'
XTILE = f'<span class="ip-tile ip-rec ip-rec--transfer" aria-hidden="true">{icon("transfer")}</span>'
XTITLE = 'Corectează Transfer H2 → H4.2'
XHEAD = (f'<div class="ip-sidepanel__head"><div class="ip-row s13-phead">{XTILE}<div><h2 class="ip-t-heading s13-m0" id="cp-t">{XTITLE}</h2>'
         f'<p class="ip-note ip-num s13-m0">{"Telița → Hîrbovăț".upper()} · {dd(XDAY)}</p></div></div>'
         f'<button type="button" class="ip-iconbtn ip-iconbtn--touch" aria-label="Închide Corectează">{icon("x")}</button></div>')
XBY = (f'<p class="ip-note s13-m0">Importat din registru · {num(f"SAPT {TWK}")} la Telița · {num(f"SAPT {HWK}")} la Hîrbovăț</p>')
def xfld(i, sx, v, saved):
    c, s, w = ('hen', '♀', 'găini') if sx == 'F' else ('rooster', '♂', 'cocoși')
    return (f'<div class="ip-field"><label class="ip-label" for="{i}">Mutate <span class="ip-sex ip-sex--{c}" aria-hidden="true">{s}</span><span class="ip-sr">{w}</span></label>'
            f'<span class="ip-input ip-count ip-count--{c}"><input id="{i}" inputmode="numeric" placeholder="–" autocomplete="off" value="{v}" aria-describedby="{i}h"/></span>'
            f'<span class="ip-hint" id="{i}h">salvat: {num(n(saved))}</span></div>')
XWHY = 'Corectez câte găini s-au mutat din H2 în H4.2.'  # D39 example reason, neutral (the residual is a data error, D35)
XFX = (f'<div class="ip-effect s13-effect" id="fx" role="status">{Q.rec("transfer", 16)}<div class="s9-fxbody"><span class="ip-cap">Efect</span>'
       f'<span>Găini în viață de pe {num(dd(XDAY))}: Telița H2 {hen_v("+" + n(LEFT))}, Hîrbovăț H4.2 {hen_v(n(-LEFT))}.</span>'
       + Q.fxline(f'Telița H2 · {dd(XDAY)}', 0, LEFT)
       + Q.fxline('Telița H2 · azi', 0, LEFT)          # S16P: Telița stays empty after the move (S9), so H2 keeps the 16 to today
       + Q.fxline(f'Hîrbovăț H4.2 · {dd(XDAY)}', XQ, XNEW)
       + Q.fxline('Hîrbovăț H4.2 · azi', H42[DDAY], H42[DDAY] - LEFT)
       + f'<span class="s9-fxcyc">{cmark("start", "", "Ciclul la Telița rămâne deschis")}<span class="s9-fxsub">nu se mai închide pe {num(dd(XDAY))}, sapt. {TWK} · rămân {hen(n(LEFT))} în H2</span></span>'
       + f'<span class="s9-fxcyc">{cmark("start", "", f"Ciclul la Hîrbovăț continuă, sapt. {HWK}")}</span>'   # the move's week, as S9
       + f'<span class="s9-fxsub">Verificat pe {cnt(XDAYS, 'zile')}, până azi ({num(dd(DDAY))}): nicio zi sub{NB}0.</span></div></div>')
XSAVE = '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch" aria-haspopup="dialog">Salvează corecția</button>'  # cycle-changing → Dialog
boards.append(('ciclu-transfer', 'transfer · ciclul Telița rămâne deschis · exemplu de stare (D39)', 'p', Q.screen(sheet(XHEAD, [
    XBY, f'<div class="s16-pair">{xfld("cf-f", "F", n(XNEW), XQ)}{xfld("cf-m", "M", "0", 0)}</div>', Q.motiv(XWHY), XFX], Q.DELK + XSAVE))))

# ---------- desktop (16.11.2026, owner, beside / over S4 · Păsări) ----------
DSAVE_OFF = lambda why: f'<button type="button" class="ip-btn ip-btn--primary" disabled aria-describedby="{why}">Salvează corecția</button>'
def dpanel(body, save, inert=False):
    ine = ' inert="inert" aria-hidden="true"' if inert else ''
    return (f'<aside class="ip-sidepanel s13-panel" id="cp" aria-labelledby="cp-t"{ine}>\n{Q.head(False)}\n<div class="ip-sidepanel__body s13-pbody">\n'
            + '\n'.join(body) + '\n</div>\n<div class="ip-sidepanel__foot"><button type="button" class="ip-btn ip-btn--destructive" aria-haspopup="dialog">Șterge</button>'
            f'<div class="ip-row s13-pkeys"><button type="button" class="ip-btn ip-btn--secondary">Anulează</button>{save}</div></div>\n</aside>')
assert dpanel([BYMOD, Q.counts(n(NEW), '0', NEW, 0, False), Q.motiv(touch=False), Q.effect('empty')], Q.DSAVE).replace('\n', '') \
    == Q.panel(False).replace('\n', '')  # the S16 desktop panel (without its open focus), rebuilt from the same pieces
def deffect(x):
    pre, post = DPRE, Q.alive_by_day(DDAY, x); dlt = x - NEW
    lines = ''.join(Q.fxline(f'{SEC} · {"azi, " if d == DDAY else ""}{dd(d)}', pre[d], post[d]) for d in pre)
    sg = '+' if dlt < 0 else ''
    return (f'<div class="ip-effect s13-effect" id="fx" role="status">{Q.rec("mortality", 16)}<div class="s9-fxbody"><span class="ip-cap">Efect</span>'
            f'<span>{SEC} · găini în viață de pe {num(dd(RDAY))} se schimbă cu {hen_v(sg + n(-dlt))}.</span>{lines}'
            f'<span class="s9-fxsub">Verificat pe {num(len(pre))} zile, până azi: nicio zi sub{NB}0.</span></div></div>')
def motiv_long():  # soft limit (Q.motiv): the text stays as typed or pasted (no maxlength), the field says how much to cut
    h = Q.motiv(LONG, touch=False, rows=5); assert 'Scurtează-l cu <span class="ip-num">12</span>' in h
    return h
DLOWN = [d for d in DLOW if DLOW[d] < 0]
DMOST = min(DDEL.values()); assert DMOST == 4122 and min(DLOW, key=DLOW.get) == DDAY  # 16.11 binds: 2 more deaths today
def dcounts_refuz():  # the phone message names 14.11 and 4 124; on 16.11 the limit is today's
    h = Q.counts(n(Q.TRY), '0', NEW, 0, False, err=True)
    msg = (f'<p class="ip-err ip-fieldmsg" id="cf-e">{icon("alert", 14)}<span>{SEC} {hen(n(Q.TRY))}: ar lăsa {hen_v(n(DLOW[DDAY]))} găini azi, '
           f'{num(dd(DDAY))}. Scrie cel mult {num(n(DMOST))}.</span></p>')
    h, k = re.subn(r'<p class="ip-err ip-fieldmsg" id="cf-e">.*?</p>', msg, h, flags=re.S); assert k == 1
    return h
# fix 1: the days with the first low figure as one range (14–15.11.2026: −6), then today (−8)
DRUN = [d for d in DLOWN if DLOW[d] == DLOW[DLOWN[0]] and d != DDAY]; assert DRUN == ['2026-11-14', '2026-11-15'] and DLOW[DDAY] == -8
DSPAN = dd(DRUN[0])[:2] + '–' + dd(DRUN[-1]) if len(DRUN) > 1 else dd(DRUN[0])
DREFUZ = (f'<div class="ip-banner ip-banner--fault s7s-banner" role="alert" id="fx">{icon("alert", 16)}<div class="s7s-banner__body"><p class="s13-m0">'
          f'Corecția ar lăsa {SEC} cu {hen_v(n(DLOW[DLOWN[0]]))} găini pe {num(DSPAN)} și cu {hen_v(n(min(DLOW.values())))} azi, {num(dd(DDAY))}. '
          f'Nicio secțiune nu poate scădea sub{NB}0.</p></div></div>')
DESK_I = Q.DESK.replace('<div class="s13-vp s16-vp">', '<div class="s13-vp s16-vp" inert="inert" aria-hidden="true">', 1)
assert DESK_I != Q.DESK
boards.append(('desktop-scris', 'desktop · scris · motiv peste 200 de caractere · exemplu de stare (D39)', 'd',
               Q.DESK + dpanel([BYMOD, Q.counts(n(BACK), '0', NEW, 0, False), motiv_long(), deffect(BACK)], DSAVE_OFF('cw-e'))))
boards.append(('desktop-refuzat', 'desktop · refuzat · sub 0 · exemplu de stare (D39)', 'd',
               Q.DESK + dpanel([BYMOD, dcounts_refuz(), Q.motiv(touch=False), DREFUZ], DSAVE_OFF('fx'))))
# the panel had no Motiv typed, so the Dialog opened empty (S16P prefill rule) and the owner typed DEL_WHY in it
DMOT, DKEY = Q.dlg_motiv(DEL_WHY, touch=False); assert DKEY == ''
DLG = (f'<div class="ip-scrim ip-scrim--page"><div class="ip-dialog" role="alertdialog" aria-modal="true" aria-labelledby="dlg-t" aria-describedby="dlg-n">'
       f'<h2 id="dlg-t">Șterge Mortalitate {SEC} din {num(dd(RDAY))}</h2>'
       f'<p class="ip-dialog__note" id="dlg-n">Se șterge rândul introdus de {Q.ENTRY["user"]} la {num(Q.ENTRY["at"][11:])} și corectat de {Q.AUD["user"]} pe '
       f'{num(dd(Q.AT_D))}: {hen(n(-NEW))} · {roo("0")}. Găini în viață în {SEC} azi: {hen(n(DPRE[DDAY]))} → {hen(n(DDEL[DDAY]))}. '
       f'Ștergerea rămâne în Istoric modificări.</p>'
       f'{DMOT}'
       f'<div class="ip-dialog__keys"><button type="button" class="ip-btn ip-btn--secondary ip-focus" autofocus="autofocus">Anulează</button>'
       f'<button type="button" class="ip-btn ip-btn--destructive"{DKEY}>Șterge</button></div></div></div>')
boards.append(('desktop-sterge-confirmare', 'desktop · Șterge · confirmare · exemplu de stare (D39)', 'd',
               DESK_I + Q.panel(False).replace('<aside class="ip-sidepanel s13-panel" id="cp" aria-labelledby="cp-t">',
                                        '<aside class="ip-sidepanel s13-panel" id="cp" aria-labelledby="cp-t" inert="inert" aria-hidden="true">', 1) + DLG))

# după ștergere — the panel closes; the page is S4 · Păsări generated again (s4s/s4t, method A, S16P fix 1) with the
# H4.1 ♀ 14.11.2026 record removed in memory (s4s_build DROP): its 5 deaths go back to the birds alive from 14.11, so the
# Jurnal (page 1, 15 rows) loses that row and the 16th moves up, and every total counts the deletion (head ♀ 25 257, Readout
# mortality, Pe secțiuni H4.1 ♀ 4 122, SAPT 40). Toast "Șters · …" + Anulează. Page full width (no panel), scrolled to the
# Jurnal or, if the page ends sooner, to its end (measured: s16s_render.py --measure).
_sp = importlib.util.spec_from_file_location('s4t_build_drop', HERE / 's4t_build.py')
T2 = importlib.util.module_from_spec(_sp); T2.DROP = (RDAY, SEC, 'F'); _sp.loader.exec_module(T2)
assert T2.s4.DROPQ == NEW and T2.s4.alive[(SEC, 'F')] == DDEL[DDAY] == 4122
FT = sum(v for (sec, sx), v in s4.alive.items() if sx == 'F'); assert (FT, T2.tot['F']) == (25252, 25252 + NEW)
T2.EVID[0] = 0
PGD = T2.page('pasari', T2.panel_pasari())
assert 'din 1097' in PGD and '#istoric-H4.1-14.11' not in PGD and hen(n(FT + NEW)) in PGD and hen(n(FT)) not in PGD
# focus after the deletion (S16P): the deleted row's key is gone, so focus goes to Corectează of the row now in its place
# (Mortalitate H4.2 · 14.11.2026); the Toast (role=status) announces the deletion and its Anulează stays in the tab order
_key = re.search(r'<a class="ip-btn ip-btn--ghost" href="#corecteaza-ev\d+">Corectează<span class="ip-sr"> Mortalitate H4\.2, 14\.11\.2026</span></a>', PGD)
assert _key and PGD.count(_key.group(0)) == 1
PGD = PGD.replace(_key.group(0), _key.group(0).replace('ip-btn--ghost"', 'ip-btn--ghost ip-focus"'))
DSCROLL = HS.get('S16-corecteaza--desktop-dupa-stergere@scroll', 0)
if DSCROLL: PGD = PGD.replace('<main class="s4-main">', f'<main class="s4-main" style="margin-top:calc(-1 * {DSCROLL}px)">', 1)
T_DEL = f'Șters · Mortalitate {SEC} · {num(dd(RDAY))} · în viață azi ♀ {num(n(DDEL[DDAY]))}'
boards.append(('desktop-dupa-stergere', 'desktop · după ștergere · exemplu de stare (D39)', 'd',
               f'<div class="s13-vp s16s-vp--full">{PGD}</div>' + toast(T_DEL, desk=True)))

# ---------- CSS (screen only, tokens only) ----------
CSS_ADD = '''
.s16s-ct{container-type:inline-size}
.s16s-sk{display:flex;flex-direction:column;gap:var(--space-4)}
.s16s-skf{display:flex;flex-direction:column;gap:var(--space-2)}
.s16s-sk .ip-skel{display:block}
.s16s-skbox{width:100%;height:var(--count-input-h)}
.s16s-skarea{width:100%;height:calc(var(--space-12) * 2)}
.s16s-skfx{width:100%;height:calc(var(--space-12) * 3)}
.s16s-lock{display:flex;align-items:center;flex-wrap:wrap;gap:0 var(--space-2);margin:var(--space-1) 0 0}'''
CSS_P = Q.CSS_P + CSS_ADD
CSS_D = Q.CSS_D + SC.BANNER + '''
.s16s-vp--full{margin-right:0} /* panel closed: the page takes the full width */'''

# ---------- write ----------
STEM0, TITLE0 = Q.STEM0, Q.TITLE0
out = []
for slug, state, kind, body in boards:
    stem, ttl = f'{STEM0}--{slug}', f'{TITLE0} · {state}'
    if kind == 'p':
        out.append(Q.write(stem, ttl, f'<div class="ip-root" style="padding:0; width:390px; height:844px; overflow:hidden">{body}</div>', CSS_P, 390, 844))
    else:
        out.append(Q.write(stem, ttl, f'<div class="ip-root s4s-rel" style="padding:0; min-height:{SC.SIDE_H}px">{body}</div>', CSS_D, 1440, SC.SIDE_H))
pathlib.Path('/tmp/s16/states.json').write_text(json.dumps(out, ensure_ascii=False))
hs = json.loads(HP.read_text()); hs.update({b['stem']: b['h'] for b in out}); HP.write_text(json.dumps(hs, indent=1) + '\n')

if CANVAS:  # S16 row: S16 phones, these phones, the S16 desktop, these desktops
    c = CR.load(ROOT)
    old = lambda w: [b['stem'] for b in Q.out if b['w'] == w]
    new = lambda w: [(b['stem'], w, b['h'], b['title']) for b in out if b['w'] == w]
    x = CR.place(c, STEM0, old(390) + new(390) + old(1440) + new(1440), 0)
    CR.save(ROOT, c)
    print('canvas: S16 row y', CR.row_y(c, STEM0), 'ends at x', x - 80)
