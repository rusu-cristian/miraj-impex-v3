"""S16S · Corectează · state boards (spec E1, SYS09, D70/D71, D101; design/notes/S16.md state list), built on s16_build.py.
python3 -I s16s_build.py <repo_root>
Imports s16_build with s16_heights.json (it rewrites the 5 S16 boards byte-identical and resets the S16 canvas.json
entries), reuses its sheet head, record line, CountInputs, Motiv, EffectPreview, Dialog wording, desktop page + panel and
CSS; Toast + Anulează / Banner-with-action CSS from sstate_css (as S9/S10/S15S); desktop scrim = the S4S Dialog rule.
Boards (phones 390×844, the sheet of additions.css §39; desktops 1440×1024 beside / over S4 · Păsări, 16.11.2026, D23):
- salvat                     15.11.2026, 10:12: the sheet as the record now is (♀ 5, "corectat · modificat de Ion Rusu ·
                             15.11.2026, 10:12"), nothing typed, Toast "Corecție salvată …" + Anulează
- eroare-la-salvare          the default draft kept (♀ 5 + the real reason), fault Banner + Reîncearcă above the preview
- se-incarca                 head known (from the row), static skeletons, keys disabled, hidden status line
- vet-blocat                 Dr. Ceban on 15.11 opens her own 14.11 record: read-only, Tag "blocat" + reason (SYS09, D70)
- ciclu-transfer             the real 28.02.2026 move Telița H2 → Hîrbovăț H4.2 ♀ 4 383 (the S9 replay, D83) typed ♀ 4 367
                             (D39): H2 keeps the workbook's 16 hens (counts-latest residual, D35), so the Telița cycle does not
                             close; Hîrbovăț H4.2 −16 on every day to 16.11; Salvează corecția opens the cycle Dialog
- desktop-scris              (merged with Motiv over 200) ♀ 5 → 3 typed (the value before the 15.11 correction, D39) + a
                             212-character reason: preview +2 on 14–16.11, Motiv error, save disabled
- desktop-refuzat            ♀ 4 130 typed (the phone D39 value): −6 on 14.11 and 15.11, −8 on 16.11 → fault Banner
- desktop-sterge-confirmare  Dialog over page + panel (inert), Motiv typed (D39), Șterge enabled, focus on Anulează
- desktop-dupa-stergere      panel closed, page full width at the Jurnal without the H4.1 14.11 row, Toast + Anulează
Writes design/canvas/project/S16-corecteaza--<state>.dc.html, /tmp/s16/<stem>.html, /tmp/s16/states.json and the canvas.json
entries: new phones after x 1410 (x 1880 … 3760), the S16 desktop moved to x 4230 (only its x), new desktops after it
(step 1520), as S9P/S10P moved their desktop. Run after s16_build.py (which puts the desktop back at x 1880)."""
import sys, re, json, pathlib, importlib, datetime as dt

ROOT = pathlib.Path(sys.argv[1]).resolve()
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
def cnt(k, word):  # Romanian: "de" after numbers whose last two digits are 00 or 20–99 (12 caractere, 212 caractere, 262 de zile)
    return f'{num(k)} {"" if 0 < k % 100 < 20 or k < 20 else "de "}{word}'
assert cnt(212, 'caractere').endswith('</span> caractere') and cnt(262, 'zile').endswith('</span> de zile') and cnt(200, 'x').endswith(' de x')
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
    w = 's16s-toastwrap--desk' if desk else 's7s-toastwrap s16s-toastwrap'
    return (f'<div class="{w}"><div class="ip-toast s7s-toast" role="status">{icon("check", 16)}<span class="s7s-toast__txt">{text}</span>'
            f'<button type="button" class="ip-link s7s-undo">Anulează</button></div></div>')

boards = []   # (slug, state title, kind 'p' | 'd', root markup)
# 1 · salvat — 15.11.2026, 10:12: the sheet shows the record as it now is (as the desktop panel), Toast + Anulează
T_SAVED = f'Corecție salvată · {SEC} ♀ {num(f"{OLD}{NBS}→{NBS}{NEW}")} · în viață ♀{NBS}{num(n(Q.POST[PDAY]))}'
boards.append(('salvat', f'salvat · {dd(Q.AT_D)}, {Q.AT_T}', 'p', Q.screen(
    sheet(Q.head(True), [BYMOD_P, Q.counts(n(NEW), '0', NEW, 0, True), Q.motiv(), Q.effect('empty')], Q.DELK + Q.save_off('fx'), toast=toast(T_SAVED)))))

# 2 · eroare la salvare — online only (spec §1): the draft stays (D60), fault Banner + Reîncearcă right above the preview
#   (S9P), Salvează corecția stays enabled; the whole body fits above the foot (no scroll)
BAN = (f'<div class="ip-banner ip-banner--fault s7s-banner" role="alert">{icon("alert", 16)}<div class="s7s-banner__body">'
       f'<p class="s13-m0">Corecția nu s-a salvat: {SEC} ♀ {num(dd(RDAY))}, {num(f"{OLD}{NBS}→{NBS}{NEW}")}. Verifică conexiunea și apasă Reîncearcă.</p>'
       f'<p class="s13-m0">Ce ai scris rămâne păstrat în browser până la salvare.</p>'
       f'<button type="button" class="ip-btn ip-btn--secondary ip-btn--touch">Reîncearcă</button></div></div>')
boards.append(('eroare-la-salvare', 'eroare la salvare', 'p', Q.screen(
    sheet(Q.head(True), [Q.BY, Q.counts(n(NEW), '0', OLD, 0, True), Q.motiv(Q.REASON), BAN, Q.effect()], FOOT))))

# 3 · se încarcă — the head comes from the row that opened the sheet; the record is unknown: static skeletons, keys off
def skf(box): return f'<div class="s16s-skf"><span class="ip-skel ip-skel--w3"></span><span class="ip-skel {box}"></span></div>'
SKEL = (f'<p class="ip-sr" role="status">Se încarcă Mortalitate {SEC} din {dd(RDAY)}.</p>'
        f'<div class="s16s-sk" aria-hidden="true"><span class="ip-skel ip-skel--w1"></span>'
        f'<div class="s16-pair">{skf("s16s-skbox")}{skf("s16s-skbox")}</div>{skf("s16s-skarea")}<span class="ip-skel s16s-skfx"></span></div>')
KEYS_OFF = ('<button type="button" class="ip-btn ip-btn--destructive ip-btn--touch" disabled>Șterge</button>'
            '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch" disabled>Salvează corecția</button>')
boards.append(('se-incarca', 'se încarcă', 'p', Q.screen(sheet(Q.head(True), [SKEL], KEYS_OFF, busy=True))))

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
         f'<p class="ip-note ip-num s13-m0">Telița → Hîrbovăț · {dd(XDAY)}</p></div></div>'
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
       + Q.fxline(f'Hîrbovăț H4.2 · {dd(XDAY)}', XQ, XNEW)
       + Q.fxline('Hîrbovăț H4.2 · azi', H42[DDAY], H42[DDAY] - LEFT)
       + f'<span class="s9-fxcyc">{cmark("start", "", "Ciclul la Telița rămâne deschis")}<span class="s9-fxsub">nu se mai închide pe {num(dd(XDAY))} · rămân {hen(n(LEFT))} în H2</span></span>'
       + f'<span class="s9-fxcyc">{cmark("start", "", f"Ciclul la Hîrbovăț continuă, sapt. {WK}")}</span>'
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
    == Q.PANEL.replace('\n', '')  # the S16 desktop panel, rebuilt from the same pieces
def deffect(x):
    pre, post = DPRE, Q.alive_by_day(DDAY, x); dlt = x - NEW
    lines = ''.join(Q.fxline(f'{SEC} · {"azi, " if d == DDAY else ""}{dd(d)}', pre[d], post[d]) for d in pre)
    sg = '+' if dlt < 0 else ''
    return (f'<div class="ip-effect s13-effect" id="fx" role="status">{Q.rec("mortality", 16)}<div class="s9-fxbody"><span class="ip-cap">Efect</span>'
            f'<span>{SEC} · găini în viață de pe {num(dd(RDAY))} se schimbă cu {hen_v(sg + n(-dlt))}.</span>{lines}'
            f'<span class="s9-fxsub">Verificat pe {num(len(pre))} zile, până azi: nicio zi sub{NB}0.</span></div></div>')
def motiv_long():  # soft limit: the text stays as typed or pasted (no maxlength), the field says how much to cut
    over = len(LONG) - 200
    return (f'<div class="ip-field"><label class="ip-label" for="cw">Motiv</label><span class="ip-input ip-input--area is-error">'
            f'<textarea id="cw" rows="5" required="required" aria-required="true" aria-invalid="true" aria-describedby="cw-e cw-h">{LONG}</textarea></span>'
            f'<p class="ip-err ip-fieldmsg" id="cw-e">{icon("alert", 14)}<span>Motivul are {cnt(len(LONG), 'caractere')}; cel mult {num(200)}. Scurtează-l cu {num(over)} ca să poți salva.</span></p>'
            f'<span class="ip-hint" id="cw-h">Obligatoriu, cel mult 200 de caractere · {num(f"{len(LONG)}/200")} · apare în Istoric modificări</span></div>')
DLOWN = [d for d in DLOW if DLOW[d] < 0]
DMOST = min(DDEL.values()); assert DMOST == 4122 and min(DLOW, key=DLOW.get) == DDAY  # 16.11 binds: 2 more deaths today
def dcounts_refuz():  # the phone message names 14.11 and 4 124; on 16.11 the limit is today's
    h = Q.counts(n(Q.TRY), '0', NEW, 0, False, err=True)
    msg = (f'<p class="ip-err ip-fieldmsg" id="cf-e">{icon("alert", 14)}<span>{SEC} {hen(n(Q.TRY))}: ar lăsa {hen_v(n(DLOW[DDAY]))} găini azi, '
           f'{num(dd(DDAY))}. Scrie cel mult {num(n(DMOST))}.</span></p>')
    h, k = re.subn(r'<p class="ip-err ip-fieldmsg" id="cf-e">.*?</p>', msg, h, flags=re.S); assert k == 1
    return h
DREFUZ = (f'<div class="ip-banner ip-banner--fault s7s-banner" role="alert" id="fx">{icon("alert", 16)}<div class="s7s-banner__body"><p class="s13-m0">'
          f'Corecția ar lăsa {SEC} cu {hen_v(n(DLOW[DLOWN[0]]))} găini pe {num(dd(DLOWN[0]))} și cu {hen_v(n(min(DLOW.values())))} azi, {num(dd(DDAY))}. '
          f'Nicio secțiune nu poate scădea sub{NB}0.</p></div></div>')
DESK_I = Q.DESK.replace('<div class="s13-vp s16-vp">', '<div class="s13-vp s16-vp" inert="inert" aria-hidden="true">', 1)
assert DESK_I != Q.DESK
boards.append(('desktop-scris', 'desktop · scris · motiv peste 200 de caractere · exemplu de stare (D39)', 'd',
               Q.DESK + dpanel([BYMOD, Q.counts(n(BACK), '0', NEW, 0, False), motiv_long(), deffect(BACK)], DSAVE_OFF('cw-e'))))
boards.append(('desktop-refuzat', 'desktop · refuzat · sub 0 · exemplu de stare (D39)', 'd',
               Q.DESK + dpanel([BYMOD, dcounts_refuz(), Q.motiv(touch=False), DREFUZ], DSAVE_OFF('fx'))))
DLG = (f'<div class="ip-scrim s4s-scrim"><div class="ip-dialog" role="alertdialog" aria-modal="true" aria-labelledby="dlg-t" aria-describedby="dlg-n">'
       f'<h2 id="dlg-t">Șterge Mortalitate {SEC} din {num(dd(RDAY))}</h2>'
       f'<p class="ip-dialog__note" id="dlg-n">Se șterge rândul introdus de {Q.ENTRY["user"]} la {num(Q.ENTRY["at"][11:])} și corectat de {Q.AUD["user"]} pe '
       f'{num(dd(Q.AT_D))}: {hen(n(-NEW))} · {roo("0")}. Găini în viață în {SEC} azi: {hen(n(DPRE[DDAY]))} → {hen(n(DDEL[DDAY]))}. '
       f'Ștergerea rămâne în Istoric modificări.</p>'
       f'<div class="ip-field"><label class="ip-label" for="dlg-why">Motiv</label><span class="ip-input ip-input--area">'
       f'<textarea id="dlg-why" rows="3" maxlength="200" required="required" aria-required="true" aria-describedby="dlg-why-h">{DEL_WHY}</textarea></span>'
       f'<span class="ip-hint" id="dlg-why-h">Obligatoriu, cel mult 200 de caractere · {num(f"{len(DEL_WHY)}/200")}</span></div>'
       f'<div class="ip-dialog__keys"><button type="button" class="ip-btn ip-btn--secondary ip-focus" autofocus="autofocus">Anulează</button>'
       f'<button type="button" class="ip-btn ip-btn--destructive">Șterge</button></div></div></div>')
boards.append(('desktop-sterge-confirmare', 'desktop · Șterge · confirmare · exemplu de stare (D39)', 'd',
               DESK_I + Q.PANEL.replace('<aside class="ip-sidepanel s13-panel" id="cp" aria-labelledby="cp-t">',
                                        '<aside class="ip-sidepanel s13-panel" id="cp" aria-labelledby="cp-t" inert="inert" aria-hidden="true">', 1) + DLG))

# după ștergere — the panel closes; the Jurnal (page 1, 15 rows) no longer has the H4.1 14.11 row (D70 D: deleted records
# leave logs and totals), the 16th record moves up; Toast "Șters · …" + Anulează. Page full width (no panel), scrolled to the
# Jurnal or, if the page ends sooner, to its end (measured: s16s_render.py --measure). The blocks above the Jurnal are out of view (their totals are as before).
def log_rows_del():
    key = (RDAY, SEC); keep = T.SD.pop(key)
    try:
        recs = T.log_records()
    finally:
        T.SD[key] = keep
    out = []
    for d, _, _, p in recs[:T.PAGE]:
        if p[0] == 'mortality':   # as s4t_build.log_rows (no "corectat" mark: that row is the one deleted)
            sec = p[2]; f, m = T.SD[(d, sec)]['F'], T.SD[(d, sec)]['M']
            out.append(T.evrow(T.tile('mortality'), f'Mortalitate · {sec} · {T.mortf(f, m)}',
                               f'<span class="ip-num">{T.ddate(d)} · {T.SAVE[d]["at"][11:]} · SAPT {T.cw(d)}</span> · {T.SAVE[d]["user"]}', f'Mortalitate {sec}, {T.ddate(d)}'))
        else:
            assert p[0] == 'cull', p
            c = p[1]; at = next(e['at'] for e in T.S['event_log'] if e['record'] == 'cull' and e['at'][:10] == c['date'])
            fig = hen(n(-c['qty'])) if c['sex'] == 'F' else roo(n(-c['qty']))
            out.append(T.evrow(T.tile('cull'), f'Sacrificare · {c["section"]} · {fig} · {c["reason"]}',
                               f'<span class="ip-num">{T.ddate(c["date"])} · {at[11:]} · SAPT {T.cw(c["date"])}</span> · {c["created_by"]}', f'Sacrificare {c["section"]}, {T.ddate(c["date"])}'))
    assert len(out) == T.PAGE and '#istoric-H4.1-14.11' not in ''.join(out)
    return out, recs
T.EVID[0] = 0
_lr = T.log_rows; T.log_rows = log_rows_del
try:
    PGD = T.page('pasari', T.panel_pasari())
finally:
    T.log_rows = _lr
assert 'din 1097' in PGD and 'Mortalitate · H4.1 · ' in PGD
DSCROLL = HS.get('S16-corecteaza--desktop-dupa-stergere@scroll', 0)
if DSCROLL: PGD = PGD.replace('<main class="s4-main">', f'<main class="s4-main" style="margin-top:calc(-1 * {DSCROLL}px)">', 1)
T_DEL = f'Șters · Mortalitate {SEC} · {num(dd(RDAY))} · în viață azi ♀ {num(n(DDEL[DDAY]))}'
boards.append(('desktop-dupa-stergere', 'desktop · după ștergere · exemplu de stare (D39)', 'd',
               f'<div class="s13-vp s16s-vp--full">{PGD}</div>' + toast(T_DEL, desk=True)))

# ---------- CSS (screen only, tokens only) ----------
CSS_ADD = '''
.s16s-toastwrap{bottom:calc(var(--control-h-touch) + var(--space-3) * 2 + var(--ip-hair) + var(--space-4))} /* above the sheet foot */
.s16s-ct{container-type:inline-size}
.s16s-sk{display:flex;flex-direction:column;gap:var(--space-4)}
.s16s-skf{display:flex;flex-direction:column;gap:var(--space-2)}
.s16s-sk .ip-skel{display:block}
.s16s-skbox{width:100%;height:var(--count-input-h)}
.s16s-skarea{width:100%;height:calc(var(--space-12) * 2)}
.s16s-skfx{width:100%;height:calc(var(--space-12) * 3)}
.s16s-lock{display:flex;align-items:center;flex-wrap:wrap;gap:0 var(--space-2);margin:var(--space-1) 0 0}'''
SCRIM = ''.join('\n' + l for l in s4.CSS_STATES.split('\n') if l.startswith('.s4s-scrim'))
assert SCRIM.count('\n') == 1
CSS_P = Q.CSS_P + SC.TOAST + CSS_ADD
CSS_D = Q.CSS_D + SC.BANNER + SC.TOAST + SCRIM + '''
.s16s-vp--full{margin-right:0} /* panel closed: the page takes the full width */
.s16s-toastwrap--desk{position:absolute;right:var(--space-6);bottom:var(--space-6);z-index:4} /* bottom right: over the empty foot of the right column, not the Jurnal links */'''

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

# canvas.json: S16 row y 29020 — new phones after x 1410, the S16 desktop moved after them (only its x), new desktops after it
cp = ROOT / 'design/canvas/project/canvas.json'
c = json.loads(cp.read_text(encoding='utf-8'))
D0 = f'{STEM0}--desktop.dc.html'
Y = c['boards'][D0]['y']; assert Y == 29020 and c['boards'][f'{STEM0}--sterge-confirmare.dc.html']['x'] == 1410
phones = [b for b in out if b['w'] == 390]; desks = [b for b in out if b['w'] == 1440]
x = 1880
for b in phones + [None] + desks:
    if b is None:
        c['boards'][D0]['x'] = x; x += 1440 + 80; continue
    k = f'{b["stem"]}.dc.html'
    e = {'x': x, 'y': Y, 'w': b['w'], 'h': b['h'], 'title': b['title']}
    if b['w'] == 1440: e = {'x': x, 'y': Y, 'w': 1440, 'h': b['h'], 'expand': 'fill', 'title': b['title']}
    c['boards'][k] = e; x += b['w'] + 80
    if k not in c['order']: c['order'].append(k)
cp.write_text(json.dumps(c, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
print('canvas S16 row y', Y, 'desktop x', c['boards'][D0]['x'], 'ends at x', x - 80)
