"""S16S · Corectează · state boards (spec E1, SYS09, D70/D71, D101; design/notes/S16.md state list), built on s16_build.py.
python3 -I s16s_build.py <repo_root> [--canvas]
Imports s16_build with s16_heights.json (it rewrites the 5 S16 boards byte-identical, never canvas.json), reuses its sheet head, record line, CountInputs, Motiv, EffectPreview, Dialog wording, desktop page + panel and
CSS; Toast + Anulează / Banner-with-action CSS from sstate_css (as S9/S10/S15S); desktop Dialog scrim = additions.css §50 (ip-scrim--page, FIN04).
RF08: realdata + the shared illustrative correction (D158a): Ion Rusu, 18.11.2026 10:12, Dr. Ceban's 17.11 Zilnic row of one
compartment ♀ old → new (new = the real deaths). Phones 18.11 before 10:12, desktops 18.11 after it, beside / over S4 · Păsări
(the table-first page, s4t_build).
Boards (phones 390×844, the sheet of additions.css §39; desktops 1440×1024):
D172: the record is Dr. Ceban's whole 17.11 Zilnic save (one EntryCard per compartment × sex, s6_build.entry); the changed
figure is marked (Tag "modificat", hint "salvat: n"); Șterge deletes the save. Tall sheet / panel boards at natural height.
- salvat                     18.11.2026, 10:12: the sheet as the save now is ("corectat · H5.2 … · modificat de Ion Rusu ·
                             18.11.2026, 10:12"), nothing typed, Toast "Corecție salvată …" + Anulează
- eroare-la-salvare          the default draft kept (new + the reason), fault Banner + Reîncearcă (focused) above the preview
- se-incarca                 head known (from the row), static skeletons, keys disabled with the visible status line as reason
- vet-blocat                 Dr. Ceban on 18.11 opens her own 17.11 save: read-only cards, Tag "blocat" + reason (SYS09, D70)
- ciclu-transfer             the real 28.02.2026 move Telița H2 → Hîrbovăț H7.2 ♀ (the S9 replay) typed lower by the real
                             H2 link difference (D39): Telița H2 would keep hens (its cycle would not close on 28.02), but the
                             14.11.2026 sale emptied H7.2, so H7.2 would be below 0 from that day to today: refused (fault Banner)
- desktop-scris              (merged with Motiv over 200) new → old typed (the value before the correction, D39) + a
                             212-character reason: preview +2 on 17–18.11, Motiv error, save disabled
- desktop-refuzat            the phone D39 value: below 0 on 17.11 and today → fault Banner
- desktop-sterge-confirmare  Dialog over page + panel (inert), Motiv typed (D39), Șterge enabled, focus on Anulează
- desktop-dupa-stergere      panel closed, page full width at the Jurnal without the 17.11 Mortalitate save (its deaths back
                             in the birds alive from 17.11), Toast + Anulează; focus on the next row's Corectează;
                             page regenerated from a copy of realdata with the record dropped (s4t/s4s/sys14 run on it with
                             their real-data asserts stripped: they check today's real totals, which the deletion changes)
Writes design/canvas/project/S16-corecteaza--<state>.dc.html, /tmp/s16/<stem>.html, /tmp/s16/states.json and (--canvas) the canvas.json
entries: new phones after x 1410 (x 1880 … 3760), the S16 desktop moved to x 4230 (only its x), new desktops after it
(step 1520), as S9P/S10P moved their desktop; only when it adds a board (canvas_rows.py: a rerun keeps every x/y)."""
import sys, re, json, pathlib, importlib, importlib.util, datetime as dt

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
ROOT, _, CANVAS = CR.args(); ROOT = ROOT.resolve()
HERE = pathlib.Path(__file__).resolve().parent
HP = HERE / 's16_heights.json'
sys.path.insert(0, str(HERE))
import skel as SKL
import sexsign as SXS  # SX01: ♀/♂ outside a figure span = the §72 hook (sexsign.py)
sys.argv = ['s16_build.py', str(ROOT), str(HP)]
Q = importlib.import_module('s16_build')        # writes the 5 S16 boards (unchanged) and the S16 canvas entries
T, s4, SC = Q.T, Q.s4, Q.SC
HS = json.loads(HP.read_text())
n, dd, icon, hen, roo, hen_v, NB, num = Q.n, Q.dd, Q.icon, Q.hen, Q.roo, Q.hen_v, Q.NB, Q.num
SEC, RDAY, OLD, NEW, PDAY, DDAY, WK = Q.SEC, Q.RDAY, Q.OLD, Q.NEW, Q.PDAY, Q.DDAY, Q.WK
s4.IC.update({'transfer': s4.IC['transfer']})
R = Q.R
sys.argv = ['s6s_build.py', str(ROOT)]
Z6S = importlib.import_module('s6s_build')   # S6 state helpers (EntryCard skeleton, saved meta); rewrites S6 boards byte-identical

# ---------- data ----------
# desktop (18.11.2026 after 10:12): the record holds NEW (corrected); typed examples (D39) run the check over 17.11 … 18.11
BACK = OLD                                             # new → old: the value before the 10:12 correction
STORED, BEFORE, CORR, KEYS = Q.STORED, Q.BEFORE, Q.CORR, Q.KEYS
BACKV = {**STORED, (SEC, 'F'): BACK}
DPRE = Q.alive_by_day(DDAY, NEW); assert list(DPRE) == [RDAY, DDAY]
DDEL_F = Q.farm_today(Q.ZERO, 'F')                      # Șterge: the save's deaths go back (farm ♀ today)
LONG = (f'Revin la {OLD}, cât a numărat {Q.ENTRY["user"]} pe {dd(RDAY)} la {Q.ENTRY_T}. Cele {NEW - OLD} găini găsite sub hrănitoare pe {R.ddm(Q.AT_D)} '
        f'nu se văd în fișa halei {Q.HALL} din {R.ddm(RDAY)}, deci le trec în ziua în care au fost găsite, nu în ziua de {dd(RDAY)}.')
assert len(LONG) == 212, len(LONG)
DEL_WHY = f'Rândul se reintroduce din Zilnic, după numărătoarea refăcută în {SEC}.'   # D39 reason, Dialog example

# the cycle board: the real move that closed Telița (transfers-seria2026.json, the S9 replay): H2 → H7.2 on 28.02.2026. Typed
# lower by the real H2 link difference (Hîrbovăț booked 2 more than Telița sent, D39 use of it): H2 would keep those hens, so the
# Telița cycle would stay open, but the 14.11 sale took H7.2 to 0 (D155), so H7.2 would be below 0 from 14.11 to today: refused.
LINK = next(l for l in R.TR['links_inferred_by_date'] if l['from'] == ['H2'] and l['sex'] == 'F')
XDAY, XSEC = LINK['arrival_date'], max(LINK['to'], key=lambda k: LINK['to'][k]); XQ = LINK['to'][XSEC]
LEFT = LINK['qty_in'] - LINK['qty_out']; XNEW = XQ - LEFT; assert LEFT > 0
assert XDAY == R.CTEL['end_date'] and XSEC in R.HSECS
XLOW = {d.isoformat(): R.alive(XSEC, 'F', d) - LEFT for d in (R.D(XDAY) + dt.timedelta(days=i) for i in range((R.TODAY - R.D(XDAY)).days + 1))}
XNEG = [d for d, v in XLOW.items() if v < 0]; XEMPTY = R.empty_since(XSEC)
assert XNEG and XNEG[0] == XEMPTY.isoformat() and XNEG[-1] == R.TODAY_ISO and len({XLOW[d] for d in XNEG}) == 1
XSALE = next(e for e in R.SALE_EVENTS if e['at'][:10] == XNEG[0] and XSEC in R.sale_codes(e))
TWK = R.sapt(XDAY)                                      # both farms in the birds' age week (D151): SAPT 21 on 28.02.2026
XDAYS = len(XLOW)

# ---------- pieces ----------
cnt = Q.cnt   # Romanian number + noun ("212 caractere", "262 de zile"), s16_build
NBS = '\u00a0'  # no-break space: "1 → 3", "♀ 4 281" stay on one line
BYMOD = f'<div>{Q.BY}{Q.MOD}</div>'
BYMOD_P = f'<div class="s16s-ct">{Q.BY}{Q.MOD}</div>'  # phone: a size container, so SYS09's < 480px rule gives the mark link its 48px hit area
ROTITLE = 'Mortalitate'
def ro_head():  # read-only for the vet: the title names the record, not the action she can't take
    h = Q.head(True).replace(f'>{Q.TITLE}<', f'>{ROTITLE}<', 1).replace('aria-label="Închide Corectează"', f'aria-label="Închide {ROTITLE}"', 1)
    assert ROTITLE in h and 'Corectează' not in h
    return h
def sheet(head, body, foot, inert=False, busy=False, toast=''):  # toast: inside the dialog (focus trap, a11y tree), above the foot
    ine = ' inert="inert"' if inert else ''
    bz = ' aria-busy="true"' if busy else ''
    return (f'<aside class="ip-sidepanel ip-sidepanel--sheet" role="dialog" aria-modal="true" aria-labelledby="cp-t"{ine}>{head}'
            f'<div class="ip-sidepanel__body ip-sidepanel__body--form"{bz}>{"".join(body)}</div>'
            f'{toast}<div class="ip-sidepanel__foot">{foot}</div></aside>')
FOOT = Q.DELK + Q.SAVE
def toast(text, desk=False):
    w = 'ip-toastwrap ip-toastwrap--page' if desk else 'ip-toastwrap'  # §48 (FIN02): phone above the sheet foot, desktop bottom right
    return (f'<div class="{w}"><div class="ip-toast" role="status">{icon("check", 16)}<span class="ip-toast__txt">{SXS.sexify(text)}</span>'
            f'<button type="button" class="ip-link ip-toast__undo">Anulează</button></div></div>')

boards = []   # (slug, state title, kind 'p' (tall sheet) | 'v' (844 viewport) | 'd' (natural) | 'dv' (1024 viewport), root markup)
# 1 · salvat — 15.11.2026, 10:12: the sheet shows the record as it now is (as the desktop panel), Toast + Anulează
T_SAVED = f'Corecție salvată · {SEC} ♀ {num(f"{OLD}{NBS}→{NBS}{NEW}")} · în viață ♀{NBS}{num(n(R.alive(SEC, "F")))}'
boards.append(('salvat', f'salvat · {dd(Q.AT_D)}, {Q.AT_T}', 'p', Q.screen(
    sheet(Q.head(True), [BYMOD_P, *Q.cards(STORED, STORED), Q.motiv(), Q.effect('empty')], Q.DELK + Q.save_off('fx'), toast=toast(T_SAVED)))))

# 2 · eroare la salvare — online only (spec §1): the draft stays (D60), fault Banner + Reîncearcă at the end of the form,
#   right before the keys (D141, D173); Salvează corecția stays enabled
BAN = (f'<div class="ip-banner ip-banner--fault ip-banner--act" role="alert">{icon("alert", 16)}<div class="ip-banner__body">'
       f'<p class="ip-m0">Corecția nu s-a salvat: Mortalitate · {num(dd(RDAY))} · {SEC} {hen(f"{OLD}{NBS}→{NBS}{NEW}")}.</p>'
       f'<p class="ip-m0">Verifică conexiunea la internet, apoi apasă Reîncearcă. Ce ai scris rămâne păstrat în browser până la salvare.</p></div>'
       f'<button type="button" class="ip-btn ip-btn--secondary ip-btn--touch ip-banner__key ip-focus" autofocus="autofocus">Reîncearcă</button></div>')  # S16P: focus after a failed save
boards.append(('eroare-la-salvare', 'eroare la salvare', 'p', Q.screen(
    sheet(Q.head(True), [Q.BY, *Q.cards(CORR, BEFORE), Q.motiv(Q.REASON), Q.effect(), BAN], FOOT))))

# 3 · se încarcă — the head comes from the row that opened the sheet; the record is unknown: static skeletons, keys off
# FIN11G: §54 field skeletons (label line, the loaded box, its hint lines) and the EffectPreview's own lines, so the
# fields and the effect keep the loaded sizes (were the s16s-sk* footprints, 20px short per field)
FX = Q.effect(); REC = re.search(r'<span class="ip-rec [^"]*s13-rec">.*?</svg></span>', FX).group(0)  # record type: known from the row
SKFX = (f'<div class="ip-effect s13-effect" aria-hidden="true">{REC}<div class="s9-fxbody">{SKL.line("ip-skel--w3", cls="ip-cap")}'
        f'{SKL.line("ip-skel--w1")}{SKL.line("ip-skel--w2")}{SKL.line("ip-skel--w1")}{SKL.line("ip-skel--w1")}{SKL.line("ip-skel--w2", cls="s9-fxsub")}</div></div>')
# S16P: the status line is visible (in place of the "Introdus de …" line) and gives the disabled keys their reason
# D172: one EntryCard skeleton per compartment of the save (s6s_build.skcard, the loaded card's geometry; the count is known from the row)
SKEL = (f'<p class="ip-note ip-m0" role="status" id="ld">Se încarcă salvarea din {num(dd(RDAY))} · Șterge și Salvează corecția se activează după încărcare.</p>'
        + Z6S.skcard * len(Q.SAVE_SECS) + f'{SKL.field("area", 2)}{SKFX}')
KEYS_OFF = ('<button type="button" class="ip-btn ip-btn--destructive ip-btn--touch" disabled aria-describedby="ld">Șterge</button>'
            '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch" disabled aria-describedby="ld">Salvează corecția</button>')
LHEAD = Q.head(True).replace('<button type="button" class="ip-iconbtn ip-iconbtn--touch"', '<button type="button" class="ip-iconbtn ip-iconbtn--touch ip-focus" autofocus="autofocus"', 1)
assert 'ip-focus' in LHEAD   # fix 1: focus stays on Închide while the record loads (components.md › S16)
boards.append(('se-incarca', 'se încarcă', 'p', Q.screen(sheet(LHEAD, [SKEL], KEYS_OFF, busy=True))))

# 4 · vet · zi trecută · blocat — Dr. Ceban (vet) on 18.11 opens her own Zilnic row of 17.11: past day, so read-only, Tag
#   "blocat" + the SYS09 reason, no Motiv, preview or Șterge (spec §1: vets change their own entries the same day only)
LOCK = (f'<p class="ip-note s16s-lock" id="vb-why"><span class="ip-tag ip-tag--neutral">blocat</span>'
        f'<span>zi trecută · o corectează managerul</span></p>')
VB_NOTE = (f'<p class="ip-m0">Poți modifica sau șterge ce ai introdus doar în aceeași zi, până la {num("23:59")}. '
           f'Pentru o corecție, spune-i managerului; el o salvează cu motiv.</p>')
boards.append(('vet-blocat', 'vet · zi trecută · blocat', 'p', Q.screen(sheet(ro_head(), [
    f'<div>{Q.BY}{LOCK}</div>', *Q.cards(BEFORE, BEFORE, meta=Z6S.saved_meta(Q.ENTRY_T, Q.ENTRY['user'], 'o corectează managerul')), VB_NOTE],
    '<button type="button" class="ip-btn ip-btn--secondary ip-btn--touch">Închide</button>'))))

# 5 · ciclu · transfer — correcting the move that closed the Telița cycle (D39 value, real record): refused by the later sale
XTILE = f'<span class="ip-tile ip-rec ip-rec--transfer" aria-hidden="true">{icon("transfer")}</span>'
XTITLE = f'Corectează Transfer H2 → {XSEC}'
XHEAD = (f'<div class="ip-sidepanel__head"><div class="ip-row ip-sidepanel__title">{XTILE}<div><h2 class="ip-t-heading ip-m0" id="cp-t">{XTITLE}</h2>'
         f'<p class="ip-note ip-num ip-m0">{"Telița → Hîrbovăț".upper()} · {dd(XDAY)}</p></div></div>'
         f'<button type="button" class="ip-iconbtn ip-iconbtn--touch" aria-label="Închide Corectează">{icon("x")}</button></div>')
XBY = f'<p class="ip-note ip-m0">Importat din registru · {num(R.SAPT(TWK))} la Telița și la Hîrbovăț</p>'
def xfld(i, sx, v, saved, err=False):
    c, s, w = ('hen', '♀', 'găini') if sx == 'F' else ('rooster', '♂', 'cocoși')
    e, inv, dsc = (' is-error', ' aria-invalid="true"', f'{i}h {i}-e') if err else ('', '', f'{i}h')
    return (f'<div class="ip-field"><label class="ip-label" for="{i}">Mutate <span class="ip-sex ip-sex--{c}" aria-hidden="true">{s}</span><span class="ip-sr">{w}</span></label>'
            f'<span class="ip-input ip-count ip-count--{c}{e}"><input id="{i}" inputmode="numeric" placeholder="–" autocomplete="off" value="{v}"{inv} aria-describedby="{dsc}"/></span>'
            f'<span class="ip-hint" id="{i}h">salvat: {num(n(saved))}</span></div>')
XERR = (f'<p class="ip-err ip-fieldmsg" id="cf-f-e">{icon("alert", 14)}<span>{XSEC} {hen(n(XNEW))}: vânzarea din {num(dd(XNEG[0]))} '
        f'a scos {hen_v(n(next(x["F"] for x in XSALE["sections"] if x["section"] == XSEC)))} găini din {XSEC}. Scrie cel puțin {num(n(XQ))}.</span></p>')
XWHY = f'Corectez câte găini s-au mutat din H2 în {XSEC}.'  # D39 example reason, neutral
XFX = (f'<div class="ip-banner ip-banner--fault s7s-banner" role="alert" id="fx">{icon("alert", 16)}<div class="s7s-banner__body"><p class="ip-m0">'
       f'Corecția ar lăsa Hîrbovăț {XSEC} cu {hen_v(n(XLOW[XNEG[0]]))} găini de pe {num(dd(XNEG[0]))} până azi: vânzarea din {num(dd(XNEG[0]))} a golit secțiunea. '
       f'Nicio secțiune nu poate scădea sub{NB}0.</p></div></div>')
boards.append(('ciclu-transfer', 'transfer · refuzat · secțiune golită de o vânzare · exemplu de stare (D39)', 'v', Q.screen(sheet(XHEAD, [
    XBY, f'<div class="s16-pair">{xfld("cf-f", "F", n(XNEW), XQ, err=True)}{xfld("cf-m", "M", "0", 0)}</div>{XERR}', Q.motiv(XWHY), XFX],
    Q.DELK + '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch" disabled aria-describedby="fx">Salvează corecția</button>'))))

# ---------- desktop (18.11.2026 after 10:12, owner, beside / over S4 · Păsări) ----------
DSAVE_OFF = lambda why: f'<button type="button" class="ip-btn ip-btn--primary" disabled aria-describedby="{why}">Salvează corecția</button>'
def dpanel(body, save, inert=False):
    ine = ' inert="inert" aria-hidden="true"' if inert else ''
    return (f'<aside class="ip-sidepanel ip-sidepanel--dock" id="cp" aria-labelledby="cp-t"{ine}>\n{Q.head(False)}\n<div class="ip-sidepanel__body ip-sidepanel__body--form">\n'
            + '\n'.join(body) + '\n</div>\n<div class="ip-sidepanel__foot"><button type="button" class="ip-btn ip-btn--destructive" aria-haspopup="dialog">Șterge</button>'
            f'<div class="ip-row ip-sidepanel__keys"><button type="button" class="ip-btn ip-btn--secondary">Anulează</button>{save}</div></div>\n</aside>')
assert dpanel([BYMOD, *Q.cards(STORED, STORED), Q.motiv(touch=False), Q.effect('empty')], Q.DSAVE).replace('\n', '') \
    == Q.panel(False).replace('\n', '')  # the S16 desktop panel (without its open focus), rebuilt from the same pieces
def motiv_long():  # soft limit (Q.motiv): the text stays as typed or pasted (no maxlength), the field says how much to cut
    h = Q.motiv(LONG, touch=False, rows=5); assert 'Scurtează-l cu <span class="ip-num">12</span>' in h
    return h
# the phone and the desktop are the same day now (18.11): the refusal reads the same (s16_build counts / effect)
DESK_I = Q.DESK.replace('<div class="ip-dockpage s16-vp">', '<div class="ip-dockpage s16-vp" inert="inert" aria-hidden="true">', 1)
assert DESK_I != Q.DESK
boards.append(('desktop-scris', 'desktop · scris · motiv peste 200 de caractere · exemplu de stare (D39)', 'd',
               Q.DESK + dpanel([BYMOD, *Q.cards(BACKV, STORED), motiv_long(), Q.effect('set', BACKV, STORED, DDAY)], DSAVE_OFF('cw-e'))))
boards.append(('desktop-refuzat', 'desktop · refuzat · sub 0 · exemplu de stare (D39)', 'd',
               Q.DESK + dpanel([BYMOD, *Q.cards(Q.TRYV, STORED, err=Q.ERR), Q.motiv(touch=False), Q.effect('refuz', Q.TRYV, STORED, DDAY)], DSAVE_OFF('fx'))))
# the panel had no Motiv typed, so the Dialog opened empty (S16P prefill rule) and the owner typed DEL_WHY in it
DMOT, DKEY = Q.dlg_motiv(DEL_WHY, touch=False); assert DKEY == ''
DLG = (f'<div class="ip-scrim ip-scrim--page"><div class="ip-dialog" role="alertdialog" aria-modal="true" aria-labelledby="dlg-t" aria-describedby="dlg-n">'
       f'<h2 id="dlg-t">Șterge Mortalitate din {num(dd(RDAY))}</h2>'
       f'<p class="ip-dialog__note" id="dlg-n">{Q.del_note(STORED, corrected=True)}</p>'
       f'{DMOT}'
       f'<div class="ip-dialog__keys"><button type="button" class="ip-btn ip-btn--secondary ip-focus" autofocus="autofocus">Anulează</button>'
       f'<button type="button" class="ip-btn ip-btn--destructive"{DKEY}>Șterge</button></div></div></div>')
boards.append(('desktop-sterge-confirmare', 'desktop · Șterge · confirmare · exemplu de stare (D39)', 'dv',
               DESK_I + Q.panel(False).replace('<aside class="ip-sidepanel ip-sidepanel--dock" id="cp" aria-labelledby="cp-t">',
                                        '<aside class="ip-sidepanel ip-sidepanel--dock" id="cp" aria-labelledby="cp-t" inert="inert" aria-hidden="true">', 1) + DLG))

# după ștergere — the panel closes; Șterge deleted the whole 17.11 Mortalitate save (D172, one record). The page is S4 · Păsări
# generated again (s4t/s4s/sys14, method A, S16P fix 1) on a copy of realdata with the save dropped in memory: its deaths go back
# to the birds alive from that day (section, hall and farm rows, the week, counts-latest), the Jurnal loses its row, every total
# counts the deletion. s4s refuses its old DROP hook (RF01); the modules run unchanged with their asserts stripped (they pin
# today's real totals). Toast "Mortalitate ștearsă · …" (object first, D143) + Anulează. Page full width (no panel), scrolled to
# its end at every board width (FX06, CSS below). Focus: Corectează of the row now in its place (the next, older one, S16P).
import importlib.machinery
class _NoAssert(importlib.machinery.SourceFileLoader):
    def get_code(self, fullname):  # always from source (no cached bytecode), assert statements stripped
        return compile(self.get_data(self.path), self.path, 'exec', dont_inherit=True, optimize=1)
def drop_save(day, secs):
    """a fresh realdata module with the Mortalitate save of `day` (every compartment × sex in `secs`) deleted"""
    spec = importlib.util.spec_from_file_location('realdata', HERE / 'realdata.py'); R2 = importlib.util.module_from_spec(spec); spec.loader.exec_module(R2)
    q = {(x, sx): R2.sd(day, x, sx)['mort'] for x in secs for sx in 'FM'}
    qh = {(h, sx): sum(v for (x, y), v in q.items() if y == sx and x.startswith(h + '.')) for h in R2.HHALLS for sx in 'FM'}
    qf = {sx: sum(v for (x, y), v in q.items() if y == sx) for sx in 'FM'}
    for r in R2.SECTION_DAY:
        k = (r['section'], r['sex'])
        if k in q and r['date'] >= day:
            r['alive'] += q[k]
            if r['date'] == day: r['mort'] -= q[k]; r['mort_sheet'] -= q[k]
    def fix_rows(rows, qq, sheet):
        prev = {}
        for r in sorted(rows, key=lambda r: r['date']):
            if r['date'] >= day:
                for sx in 'FM':
                    r[f'alive_{sx}'] += qq[sx]
                    if r['date'] == day:
                        r[f'mort_{sx}'] -= qq[sx]
                        if sheet: r[f'mort_{sx}_sheet'] -= qq[sx]
                if r['date'] > day and r['eggs'] and prev.get('alive_F'): r['lay_pct'] = round(r['eggs'] / prev['alive_F'] * 100, 2)
            prev = r
    for h in R2.HHALLS: fix_rows([r for r in R2.HALL_DAY if r['hall'] == h], {sx: qh[(h, sx)] for sx in 'FM'}, False)
    fix_rows(R2.FARM_DAY, qf, True)
    k = R2.sapt(day); w = R2.WEEKS[k]; b = w['to']
    after = sum(1 for r in R2.FARM_DAY if day < r['date'] <= b); onw = sum(1 for r in R2.FARM_DAY if day <= r['date'] <= b)
    for blk, qq in [(w['farm'], qf)] + [(w['halls'][h], {sx: qh[(h, sx)] for sx in 'FM'}) for h in R2.HHALLS]:
        for sx in 'FM':
            blk[f'mort_{sx}'] -= qq[sx]; blk[f'alive_{sx}_end'] += qq[sx]
            if f'mort_{sx}_sheet' in blk: blk[f'mort_{sx}_sheet'] -= qq[sx]
        if qq['F'] and blk.get('hen_days'):
            blk['hen_days'] += qq['F'] * after; blk['hen_days_prod'] += qq['F'] * onw
            blk['lay_pct'] = round(blk['eggs'] / blk['hen_days'] * 100, 2); blk['lay_pct_prod'] = round(blk['eggs'] / blk['hen_days_prod'] * 100, 2)
            if 'lay_vs_guide_pts' in blk: blk['lay_vs_guide_pts'] = round(blk['lay_pct'] - w['guide']['lay_pct'], 2)
    for (x, sx), v in q.items():
        sw = w['halls'][x.split('.')[0]]['sections'][x]; sw[f'mort_{sx}'] -= v; sw[f'alive_{sx}_end'] += v
    assert max(R2.WEEKS) == k                                                         # the save's week is the last one
    C = R2.COUNTS
    for (x, sx), v in q.items():
        cs = next(c for c in C['sections'] if c['section'] == x); ch = next(c for c in C['halls'] if c['hall'] == x.split('.')[0])
        cs[sx] += v; ch[sx] += v; C['totals'][sx] += v; C['totals']['all'] += v
        dsx = cs['detail'][sx]; dsx['mortality'] -= v; dsx['mortality_sheet'] -= v; dsx['alive'] += v
        dsx['mortality_pct'] = round(dsx['mortality'] / dsx['start'] * 100, 2)
    for sx in 'FM':
        cr = C['cycle']['removals'][sx]; cr['mortality'] -= qf[sx]; cr['mortality_sheet'] -= qf[sx]
        C['cycle']['mortality_pct'][sx] = round(cr['mortality'] / C['cycle']['start_counts'][sx] * 100, 2)
    ev = next(e for e in R2.EVENTS if e['record'] == 'mortality' and e['at'][:10] == day)
    R2.EVENTS.remove(ev); R2.ILL['event_log'] = [e for e in R2.ILL['event_log'] if e is not ev and e != ev]
    assert (R2.alive_farm('F'), R2.deaths(day, sx='F')) == (R.alive_farm('F') + qf['F'], 0)
    assert R2.week(k)['alive_F_end'] == R2.alive_farm('F') == C['totals']['F']
    return R2, qf
def load_dropped(day, secs):
    """s4t_build (and the s4s_build / sys14_build it loads) on the dropped realdata, asserts stripped; sys.modules restored."""
    R2, q = drop_save(day, secs)
    keep, spec0 = sys.modules.get('realdata'), importlib.util.spec_from_file_location
    def spec_noassert(name, location, *a, **k):
        sp_ = importlib.util.spec_from_loader(name, _NoAssert(name, str(location)), origin=str(location)); sp_.has_location = True
        return sp_
    sys.modules['realdata'] = R2; importlib.util.spec_from_file_location = spec_noassert
    try:
        sp = spec_noassert('s4t_build_drop', HERE / 's4t_build.py'); T2 = importlib.util.module_from_spec(sp)
        sp.loader.exec_module(T2)
    finally:
        sys.modules['realdata'] = keep; importlib.util.spec_from_file_location = spec0
    assert T2.R is R2 and T2.s4.R is R2 and T2.P.R is R2
    return T2, q
T2, DQ = load_dropped(RDAY, Q.SAVE_SECS)
assert DQ['F'] == R.deaths(RDAY, sx='F') and T2.R.alive_farm('F') == DDEL_F
FT = R.alive_farm('F')
T2.EVID[0] = 0
PGD = T2.page('pasari', T2.panel_pasari())
assert hen(n(DDEL_F)) in PGD and hen(n(FT)) not in PGD and Q.ENTRY['at'][:10] not in [e['at'][:10] for e in T2.LOGE if e['record'] == 'mortality']
# focus: the deleted save's key is gone, so Corectează of the row now in its place (the next, older row of the Jurnal)
NEXT = next(e for e in T2.LOGE if e['at'] < Q.ENTRY['at'])
NNAME = (f'Mortalitate, {dd(NEXT["at"][:10])}' if NEXT['record'] == 'mortality' else f'Vânzare {R.sale_codes(NEXT)}, {dd(NEXT["at"][:10])}')
_key = Q.jurnal_key(PGD, NNAME)
PGD = PGD.replace(_key, _key.replace('ip-btn--ghost"', 'ip-btn--ghost ip-focus"'))
T_DEL = f'Mortalitate ștearsă · {num(dd(RDAY))} · {num(len(Q.SAVE_SECS))} secțiuni · în viață azi ♀ {num(n(DDEL_F))}'
boards.append(('desktop-dupa-stergere', 'desktop · după ștergere · exemplu de stare (D39)', 'd',
               f'<div class="ip-dockpage s16s-vp--full">{PGD}</div>' + toast(T_DEL, desk=True)))

# ---------- CSS (screen only, tokens only) ----------
CSS_ADD = '''
.s16s-ct{container-type:inline-size}
.s16s-lock{display:flex;align-items:center;flex-wrap:wrap;gap:0 var(--space-2);margin:var(--space-1) 0 0}'''
CSS_P = Q.CSS_P + CSS_ADD
CSS_D = Q.CSS_D + SC.BANNER + '''
.s16s-vp--full{right:0} /* panel closed: the page takes the full width */
.s16s-vp--full > .ip-shell{height:100%} /* FX06: the page scrolled to its end at any width: the Sidebar fills the board, */
.s16s-vp--full > .ip-shell > .ip-shell__main{align-self:flex-end} /* the page's foot on the board's foot, its top above */'''

# ---------- write ----------
STEM0, TITLE0 = Q.STEM0, Q.TITLE0
out = []
for slug, state, kind, body in boards:   # D172: tall sheets / panels at their natural height (s16s_render.py --heights)
    stem, ttl = f'{STEM0}--{slug}', f'{TITLE0} · {state}'
    if kind in ('p', 'v'):
        h = Q.phone_h(stem, kind == 'p'); out.append(Q.write(stem, ttl, Q.phone_root(body, h), CSS_P, 390, h))
    else:
        h = Q.desk_h(stem) if kind == 'd' else SC.SIDE_H
        out.append(Q.write(stem, ttl, f'<div class="ip-root s4s-rel" style="padding:0; min-height:{h}px">{body}</div>', CSS_D, 1440, h))
pathlib.Path('/tmp/s16/states.json').write_text(json.dumps(out, ensure_ascii=False))
hs = json.loads(HP.read_text()); hs.update({b['stem']: b['h'] for b in out}); HP.write_text(json.dumps(hs, indent=1) + '\n')

if CANVAS:  # S16 row: S16 phones, these phones, the S16 desktop, these desktops
    c = CR.load(ROOT)
    old = lambda w: [b['stem'] for b in Q.out if b['w'] == w]
    new = lambda w: [(b['stem'], w, b['h'], b['title']) for b in out if b['w'] == w]
    x = CR.place(c, STEM0, old(390) + new(390) + old(1440) + new(1440), 0)
    CR.save(ROOT, c)
    print('canvas: S16 row y', CR.row_y(c, STEM0), 'ends at x', x - 80)
