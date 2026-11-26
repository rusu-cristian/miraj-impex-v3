"""Build the VO01 voice previews (D199 + note, spec §8 F1 Comenzi vocale, additions.css §78):
python3 -I vo01_build.py [<repo_root>] [--canvas]
Writes design/system/components/{VoiceButton,VoicePanel,VoiceMark}/preview.html (fragments, line 1 = @dsCard marker) and the
static canvas board design/canvas/project/V1-voce.dc.html ("V1 · Voce": the three previews one under another, without the
VoicePanel's step-through demo, which needs its script).
- VoiceButton: the key's states 1:1 (phone 56 round, desktop header key) and where it sits: above the BottomTabBar, above a
  SaveBar, a Toast stacked above it (executat), on the scrim over a BottomSheet, hidden while the soft keyboard is open,
  after a cancel (focus back on the key, "Comanda a fost anulată."); desktop PageHeader + the executat / export Toasts.
- VoicePanel: a step-through demo (plain JS: the key toggles, the meter is written from a level value ≤ 15 times a second,
  auto-stop after a pause, thinking, a result) and every panel state at 390 (BottomSheet over the phone screen) and at 1440
  (popover under the PageHeader key): permisiune, ascultă (start, live), auto-stop, se gândește, transcript corectat,
  răspuns, clarificare (+ "Alta…"), neînțeles, nepermis, fără microfon, fără conexiune.
- VoiceMark: the mark on each field type ("din voce", "din voce · înainte 2") and the formular precompletat at 390 (S6 Zilnic,
  H5.1) and 1440 (S13 Livrare hrană).
Copy: spec §8 word for word; phrases from its 81-intent catalogue; figures real to 18.11.2026 (H5.1 ♀ 4 264 · ♂ 342 at the
start of the day, Hîrbovăț ♀ 12 561 · ♂ 794, SAPT 59 ouat 40,4% / ghid 57,6%, 13.11.2026 25 640 kg Breeder 2); the spoken
figures of a prefilled form (♀ 3 · ♂ 1, "înainte 2") are labelled sample. ♀/♂ through sexsign.sexify (§72).
--canvas: adds or refreshes the board's canvas.json entry as its own row right under M1 · Mișcare, then re-stacks the rows
below (canvas_rows), and main_build lists it. Heights: the dsCard height of each preview = its content at 1440 + 48 (D135),
measured by vo01_render.py (CARD_H below)."""
import sys, pathlib
HERE = pathlib.Path(__file__).resolve().parent
REPO = pathlib.Path(sys.argv[1]) if len(sys.argv) > 1 and not sys.argv[1].startswith('--') else HERE.parents[2]
CANVAS = '--canvas' in sys.argv
sys.path.insert(0, str(REPO / 'design/canvas/scripts'))
import sexsign as SXS

ROOT = REPO / 'design/system/components'
CARD_H = {'VoiceButton': 2392, 'VoicePanel': 14014, 'VoiceMark': 1384}   # vo01_render.py: content height at 1440 + 48
STEM, TITLE = 'V1-voce', 'V1 · Voce'
BW = 1440

# ---------------------------------------------------------------------------------------------------------- icons (Lucide)
I = {
    'mic': '<path d="M12 2a3 3 0 0 0-3 3v7a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3Z"/><path d="M19 10v2a7 7 0 0 1-14 0v-2"/><path d="M12 19v3"/>',
    'x': '<path d="M18 6 6 18"/><path d="m6 6 12 12"/>',
    'enter': '<path d="M20 4v7a4 4 0 0 1-4 4H4"/><path d="m9 10-5 5 5 5"/>',   # Lucide corner-down-left
    'pencil': '<path d="M21.174 6.812a1 1 0 0 0-3.986-3.987L3.842 16.174a2 2 0 0 0-.5.83l-1.321 4.352a.5.5 0 0 0 .623.622l4.353-1.32a2 2 0 0 0 .83-.497z"/><path d="m15 5 4 4"/>',
    'info': '<circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/>',
    'warn': '<path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"/><path d="M12 9v4"/><path d="M12 17h.01"/>',
    'fault': '<circle cx="12" cy="12" r="10"/><path d="M12 8v4"/><path d="M12 16h.01"/>',
    'back': '<path d="m12 19-7-7 7-7"/><path d="M19 12H5"/>',
    'home': '<path d="M3 10.5 12 3l9 7.5V21H3z"/><path d="M9 21v-6h6v6"/>',
    'layers': '<path d="m12 2 10 5-10 5L2 7z"/><path d="m2 12 10 5 10-5"/><path d="m2 17 10 5 10-5"/>',
    'plus': '<path d="M5 12h14"/><path d="M12 5v14"/>',
    'chart': '<path d="M3 3v18h18"/><path d="m7 15 4-4 3 3 5-6"/>',
    'dots': '<circle cx="5" cy="12" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="19" cy="12" r="1"/>',
    'user': '<circle cx="12" cy="8" r="4"/><path d="M4 21a8 8 0 0 1 16 0"/>',
    'cal': '<rect x="3" y="4" width="18" height="18" rx="2"/><path d="M16 2v4"/><path d="M8 2v4"/><path d="M3 10h18"/>',
    'chev': '<path d="m6 9 6 6 6-6"/>',
    'download': '<path d="M12 3v12"/><path d="m7 10 5 5 5-5"/><path d="M5 21h14"/>',
    'mortality': '<path d="M12 3a7 7 0 0 0-7 7v4l2 2v3h10v-3l2-2v-4a7 7 0 0 0-7-7z"/><circle cx="9.5" cy="11" r="1.2"/><circle cx="14.5" cy="11" r="1.2"/><path d="M11 19v-2M13 19v-2"/>',
    'cull': '<circle cx="12" cy="12" r="9"/><path d="M5.7 5.7l12.6 12.6"/>',
    'feed': '<path d="M12 21V9M12 9c-3 0-4-2-4-5 3 0 4 2 4 5zM12 9c3 0 4-2 4-5-3 0-4 2-4 5zM12 15c-3 0-4-2-4-5 3 0 4 2 4 5zM12 15c3 0 4-2 4-5-3 0-4 2-4 5z"/>',
    'eggs': '<path d="M12 22c4 0 7-3 7-7.5C19 9 16 2 12 2S5 9 5 14.5C5 19 8 22 12 22z"/>',
    'check': '<path d="M20 6 9 17l-5-5"/>',
}


def svg(name, size=None):
    c = f'ip-icon ip-icon--{size}' if size else 'ip-icon'
    return f'<svg class="{c}" viewBox="0 0 24 24" aria-hidden="true">{I[name]}</svg>'


def num(x):
    return f'<span class="ip-num">{x}</span>'


def lamp(kind, word):
    return f'<span class="ip-lamp ip-lamp--{kind}">{word}</span>'


def cap(t, tag='p'):
    return f'<{tag} class="ip-cap vo-cap">{t}</{tag}>'


def q(t):
    """a spoken phrase in Romanian quotes"""
    return f'„{t}”'


# ---------------------------------------------------------------------------------------------------------- VoiceButton
def key(form, st='rest', uid='', float_=False, focus=False, hidden=False, demo=False):
    """form 'fab' (phone, 56 round) or 'head' (desktop, control-h). st: rest · hover · listen · open (panel on a result:
    latched) · think · off (aria-disabled, reason "Fără conexiune")."""
    fab = form == 'fab'
    cls = f'ip-voicekey ip-voicekey--{"fab" if fab else "head"}'
    cls += ' ip-voicekey--float' if float_ else ''
    cls += ' is-hover' if st == 'hover' else ''
    cls += ' ip-focus' if focus else ''
    pressed = 'true' if st in ('listen', 'open', 'think') else 'false'
    a = f' aria-label="Comandă vocală" aria-pressed="{pressed}" aria-keyshortcuts="Control+Shift+Space"'
    if st == 'think':
        a += ' aria-busy="true"'
    if st == 'off':
        a += ' aria-disabled="true"' + ('' if fab else f' aria-describedby="{uid}-why"')
    if not fab:
        a += ' title="Comandă vocală · Ctrl+Shift+Space"'
    else:   # the phone key has no tooltip: the same words are its description
        a += f' aria-describedby="{uid}-d' + (f' {uid}-why"' if st == 'off' else '"')
    if hidden:
        a += ' hidden'
    if demo:
        a += ' id="vo-demo-key" data-vo="key"'
    inner = svg('mic', 22 if fab else 16) + '<span class="ip-spin" aria-hidden="true"></span>'
    why = f'<span class="ip-sr" id="{uid}-why">Fără conexiune</span>' if st == 'off' else ''
    if fab:
        why = f'<span class="ip-sr" id="{uid}-d">Comandă vocală · Ctrl+Shift+Space</span>' + why
    return f'<button type="button" class="{cls}"{a}>{inner}</button>{why}'


# ---------------------------------------------------------------------------------------------------------- shells
TABS_VET = [('home', 'Acasă', True), ('plus', 'Înregistrează', False), ('user', 'Cont', False)]
TABS_5 = [('home', 'Acasă', False), ('layers', 'Serie', True), ('plus', 'Înregistrează', False), ('chart', 'Rapoarte', False), ('dots', 'Mai mult', False)]


def tabbar(items):
    out = ''
    for ic, word, cur in items:
        if ic == 'plus':
            out += (f'<a class="ip-tabbar__item" href="#inregistreaza"><span class="ip-tabbar__key">{svg("plus", 22)}</span>{word}</a>')
        else:
            out += (f'<a class="ip-tabbar__item{" is-current" if cur else ""}" href="#"{" aria-current=\"page\"" if cur else ""}>'
                    f'{svg(ic, 22)}{word}</a>')
    return f'<nav class="ip-tabbar ip-screen__tabbar" aria-label="Navigare">{out}</nav>'


def phone(inner, h=844, label=''):
    lab = f' aria-label="{label}"' if label else ''
    return f'<div class="ip-phone vo-phone" style="height:{h}px"{lab}>{inner}</div>'


def acasa_vet():
    return ('<main class="ip-screen__main"><div class="ip-stack" style="gap:var(--space-1)"><h1 class="ip-t-title ip-m0">Acasă</h1>'
            '<p class="ip-note ip-num ip-m0">Dr. Ceban · Hîrbovăț · 18.11.2026 · SAPT 59</p></div>'
            '<section class="ip-stack" style="gap:var(--space-2)"><h2 class="ip-t-heading ip-m0">Lipsesc date</h2>'
            '<p class="ip-note ip-m0">Zilnic, Hrană zilnică: H5.1 · H5.2 · H6.1 · H6.2 · H7.1 · H7.2</p></section></main>')


def ferma_mgr():
    return ('<main class="ip-screen__main"><div class="ip-stack" style="gap:var(--space-1)"><h1 class="ip-t-title ip-m0">Hîrbovăț</h1>'
            '<p class="ip-note ip-m0"><span class="ip-num">SAPT 59</span> · ♀ <span class="ip-num">12 561</span> · ♂ <span class="ip-num">794</span></p></div>'
            '<section class="ip-stack" style="gap:var(--space-2)"><h2 class="ip-t-heading ip-m0">Rezumat</h2>'
            '<p class="ip-note ip-m0">Seria 2026 · ciclu din 22.02.2026 · Ross 308 ouat</p></section></main>')


# ---------------------------------------------------------------------------------------------------------- VoicePanel
def foot(uid, phone_, value='', focus=False, edit=False, key_st='open', data=False):
    """the panel foot: the one command field (textarea, 1–3 lines, grows) + Execută (desktop Button; phone 48 square key with
    corner-down-left) + on phone the docked key. edit: an edited transcript in the field: Execută primary, Anulează beside
    (desktop) or under it (phone)."""
    t = ' ip-input--touch' if phone_ else ''
    fo = ' is-focus' if focus else ''
    dv = ' data-vo="field"' if data else ''
    fld = (f'<div class="ip-field"><label class="ip-label" for="{uid}-in">Scrie comanda</label>'
           f'<span class="ip-input{t} ip-voicepanel__text{fo}"><textarea id="{uid}-in" rows="1" autocomplete="off" spellcheck="false" '
           f'enterkeyhint="send"{dv}>{value}</textarea></span></div>')
    kind = 'primary' if edit else 'secondary'
    da = ' data-vo="run"' if data else ''
    if phone_:
        run = (f'<button type="button" class="ip-btn ip-btn--{kind} ip-btn--touch ip-voicepanel__run" aria-label="Execută"{da}>'
               f'{svg("enter")}</button>')
    else:
        run = f'<button type="button" class="ip-btn ip-btn--{kind}"{da}>Execută</button>'
    cancel = ''
    if edit or data:
        hid = '' if edit else ' hidden'
        cancel = f'<button type="button" class="ip-btn ip-btn--secondary{" ip-btn--touch" if phone_ else ""}" data-vo="undo"{hid}>Anulează</button>'
    if phone_:
        rest = run + key('fab', key_st, uid) + (f'<div class="ip-voicepanel__footkeys">{cancel}</div>' if cancel else '')
    else:
        rest = cancel + run
    return f'<div class="ip-voicepanel__foot">{fld}{rest}</div>'


def panel(uid, phone_, *, state=None, time=None, level=None, hint=False, said=None, partial=False, said_btn=False, cmd=None,
          edit=None, result='', ex=None, note='', sr='', field='', field_focus=False, key_st='open', stopped=''):
    """one VoicePanel (phone: the BottomSheet, its key docked in the foot; desktop: the popover). Returns the section."""
    touch = ' ip-iconbtn--touch' if phone_ else ''
    st = lamp(*state) if state else ''
    srs = f'<span class="ip-sr">{sr}</span>' if sr else ''
    head = (f'<div class="ip-voicepanel__head"><h2 class="ip-t-heading" id="{uid}-t">Comandă vocală</h2>'
            f'<p class="ip-voicepanel__state" role="status" aria-live="polite">{st}{srs}</p>'
            + (f'<span class="ip-voicepanel__time">{time}</span>' if time else '')
            + f'<button type="button" class="ip-iconbtn{touch} ip-voicepanel__close" aria-label="Închide">{svg("x")}</button></div>')
    body = ''
    if level is not None:
        segs = '<i></i>' * 12
        body += (f'<div class="ip-voicepanel__level"><div class="ip-voicemeter" aria-hidden="true" data-level="{level}">{segs}</div>'
                 + ('<p class="ip-voicepanel__hint">Se oprește singur după o pauză. Apasă din nou ca să termini.</p>' if hint else '')
                 + (f'<p class="ip-voicepanel__hint">{stopped}</p>' if stopped else '') + '</div>')
    if note:
        body += note
    if said is not None:
        if said_btn:
            body += (f'<button type="button" class="ip-voicepanel__said" aria-describedby="{uid}-fix"><span>{q(said)}</span>{svg("pencil", 14)}</button>'
                     f'<span class="ip-sr" id="{uid}-fix">Apasă pe text ca să-l corectezi</span>')
        else:
            body += f'<p class="ip-voicepanel__said{" is-partial" if partial else ""}">{q(said)}</p>'
        if cmd:
            body += f'<p class="ip-voicepanel__cmd">{cmd}</p>'
    if result:
        body += f'<div class="ip-voicepanel__result">{result}</div>'
    if ex:
        bt = ' ip-btn--touch' if phone_ else ''
        body += ('<div class="ip-voicepanel__ex"><p>Poți spune:</p>'
                 + ''.join(f'<button type="button" class="ip-btn ip-btn--ghost{bt} ip-voiceex">{q(e)}</button>' for e in ex) + '</div>')
    ft = foot(uid, phone_, edit if edit is not None else field, focus=field_focus or edit is not None, edit=edit is not None, key_st=key_st)
    bd = f'<div class="ip-voicepanel__body">{body}</div>' if body else ''
    if phone_:
        return (f'<section class="ip-sheet ip-voicepanel" aria-labelledby="{uid}-t"><div class="ip-sheet__handle"></div>'
                f'{head}{bd}{ft}</section>')
    return (f'<section class="ip-voicepanel ip-voicepanel--pop" aria-labelledby="{uid}-t" style="position:static;margin:var(--space-2) 0 0 auto">'
            f'{head}{bd}{ft}</section>')


def on_phone(pnl, who='vet'):
    """the VoicePanel sheet over a phone screen (S2 vet, 3 tabs; S4 manager, 5 tabs); the floating key is docked in it"""
    page, tabs = (acasa_vet(), TABS_VET) if who == 'vet' else (ferma_mgr(), TABS_5)
    return phone(f'<div class="ip-screen ip-screen--fixed ip-scrimhost"><div class="ip-screen__view">{page}</div>{tabbar(tabs)}'
                 f'<div class="ip-scrim ip-scrim--sheet">{pnl}</div></div>')


def on_desk(pnl, title, crumbs, actions='', key_st='open', uid=''):
    """the popover under the PageHeader key (1440 layout, the page's right end)"""
    cr = ''.join(f'<a href="#">{c}</a><span aria-hidden="true">›</span>' for c in crumbs)
    return (f'<div class="vo-desk"><div class="ip-pagehead"><div><nav class="ip-crumbs" aria-label="Unde ești">{cr}<span aria-current="page">{title}</span></nav>'
            f'<h1 class="ip-t-title ip-m0">{title}</h1></div><div class="ip-row" style="gap:var(--space-2)">{actions}'
            f'<span class="ip-voiceanchor">{key("head", key_st, uid)}</span></div></div>{pnl}</div>')


def btn(word, kind='secondary', icon=None, touch=False, extra=''):
    ic = svg(icon, 16) if icon else ''
    return f'<button type="button" class="ip-btn ip-btn--{kind}{" ip-btn--touch" if touch else ""}{extra}">{ic}{word}</button>'


def banner(kind, paras, key_='', role='alert', touch=False, focus_key=False, uid=''):
    ico = {'info': 'info', 'warn': 'warn', 'fault': 'fault'}[kind]
    body = ''.join(f'<p>{p}</p>' for p in paras)
    r = f' role="{role}"' if role else ''
    k = ''
    if key_:
        k = (f'<button type="button" class="ip-btn ip-btn--secondary{" ip-btn--touch" if touch else ""} ip-banner__key'
             f'{" ip-focus" if focus_key else ""}">{key_}</button>')
    return (f'<div class="ip-banner ip-banner--{kind} ip-banner--act"{r}>{svg(ico, 16)}<div class="ip-banner__body">{body}</div>{k}</div>')


def chips(uid, qtext, options, alta=False, touch=False, cols=None, fill=False):
    t = ' ip-chip--touch' if touch else ''
    wrap = (' ip-chips--wrap' if cols else '') + (' ip-chips--fill' if fill else '')
    sty = f' style="--ip-chips-cols:{cols}"' if cols else ''
    ch = ''.join(f'<button type="button" class="ip-chip ip-voicechip{t}" role="radio" aria-checked="false"'
                 f'{" tabindex=\"0\"" if i == 0 else " tabindex=\"-1\""}>{o}</button>' for i, o in enumerate(options))
    if alta:
        ch += f'<button type="button" class="ip-chip ip-voicechip{t}" aria-haspopup="dialog" tabindex="-1">Alta…</button>'
    return (f'<p class="ip-voicepanel__q" id="{uid}-q">{qtext}</p>'
            f'<div class="ip-chips ip-voicechips{wrap}" role="radiogroup" aria-labelledby="{uid}-q"{sty}>{ch}</div>')


PRIV = ('<p class="ip-voicepanel__note">Ce spui nu se păstrează. Dacă salvezi un formular completat din voce, textul comenzii '
        'rămâne în Istoric modificări.</p>')
SRNOTE = 'Cu cititorul de ecran pornit, folosește căștile sau scrie comanda.'
EX_VET = ['pune 3 găini moarte la ha cinci punct unu', 'dă 160 de grame la toate găinile', 'ce lipsește azi?', 'deschide Zilnic']
EX_S5 = ['arată mortalitatea pe ultimele 8 săptămâni', 'pe zile', 'deschide SAPT 58', 'exportă tabelul']
EX_S14 = ['care e procentul de ouat săptămâna asta față de ghid?', 'producția de la Hîrbovăț', 'ultimele 8 săptămâni', 'exportă tabelul']
S5_ACT = btn('Exportă .xlsx', 'ghost', 'download')
S13_ACT = ''
S14_ACT = btn('Exportă .xlsx', 'ghost', 'download')
CMD_P01 = 'Zilnic · Hîrbovăț · 18.11.2026 · H5.1 · ♀ Mortalitate 3'
CMD_P12 = 'Livrare hrană · Hîrbovăț · 13.11.2026 · Breeder 2 · 25 640 kg'
SAY_P01 = 'pune trei găini moarte la ha cinci punct unu'
SAY_P12 = 'am primit douăzeci și cinci de mii șase sute patruzeci de kilograme pentru găini pe treisprezece noiembrie'
# transcripts are the words as heard (figures as words); the parsed line under them and the results write figures
SAY_N = 'arată-mi mortalitatea la ha cinci pe ultimele opt săptămâni'


def panel_states():
    """[(name, caption 390, phone drawing, caption 1440, desktop drawing)]; desktop extras appended as (name, None, None, cap, d)"""
    S = []
    D5 = dict(title='H5', crumbs=['Seria 2026', 'Hîrbovăț'], actions=S5_ACT)
    D13 = dict(title='Hrană', crumbs=['Rapoarte'], actions=S13_ACT)
    D14 = dict(title='Producție', crumbs=['Rapoarte'], actions=S14_ACT)

    def row(name, pc, pp, dc, dp, who='vet', dctx=D5, dkey='open'):
        S.append((name, pc, on_phone(pp, who), dc, on_desk(dp, key_st=dkey, uid=f'k{len(S)}', **dctx)))

    # 1 permisiune (first use on the device)
    perm = ('<p>Impex Line are nevoie de microfon ca să-ți audă comanda. Apasă Permite în fereastra browserului.</p>')
    row('Cerere permisiune microfon',
        'Telefon · vet · S2 · prima folosire: panoul spune de ce, înainte de fereastra browserului; câmpul merge deja',
        panel('p1', True, state=('info', 'aștept permisiunea'), note=perm + PRIV, sr=SRNOTE),
        'Desktop · manager · S5 Hîrbovăț › H5',
        panel('d1', False, state=('info', 'aștept permisiunea'), note=perm + PRIV, sr=SRNOTE))
    # 2 ascultă · start (examples until the first word arrives)
    row('Ascultă · început',
        'Ascult… · nivelul live · 0:01 · exemple pentru ecran și rol până la primul cuvânt; tasta blocată (aria-pressed) în piciorul panoului',
        panel('p2', True, state=('ok', 'Ascult…'), time='0:01', level=2, hint=True, ex=EX_VET, key_st='listen'),
        'Desktop · tasta din PageHeader blocată (grafit), popover sub ea',
        panel('d2', False, state=('ok', 'Ascult…'), time='0:01', level=3, hint=True, ex=EX_S5[:1] + EX_S5[1:3], key_st='listen'),
        dkey='listen')
    # 3 ascultă · live transcript
    row('Ascultă · transcriere live',
        'Transcrierea live în ink-secondary cât se mai poate schimba; contorul în Mono',
        panel('p3', True, state=('ok', 'Ascult…'), time='0:04', level=8, hint=True, said=SAY_P01, partial=True, key_st='listen'),
        'Desktop · „arată-mi mortalitatea la ha cinci pe ultimele…” încă parțial',
        panel('d3', False, state=('ok', 'Ascult…'), time='0:03', level=6, hint=True, said='arată-mi mortalitatea la ha cinci pe ultimele', partial=True, key_st='listen'),
        dkey='listen')
    # 4 auto-stop
    row('Oprire automată',
        'Proprietar · 20 s cel mult: contorul se oprește la 0:20, nivelul se stinge, tasta arată spinnerul',
        panel('p4', True, state=('info', 'Mă gândesc…'), time='0:20', level=0, stopped='Am oprit după 20 de secunde.',
              said='serie nouă seria două mii douăzeci și șapte, la Telița pe șase octombrie două mii douăzeci și șapte, vârstă zero zile, ha unu găini paisprezece mii optzeci, ha doi găini paisprezece mii, ha cinci liniuță unu cocoși trei mii cinci sute zece',
              cmd='Serie nouă · Seria 2027 · Telița · 06.10.2027 · vârstă 0 zile · H1 ♀ 14 080 · H2 ♀ 14 000 · H5-1 ♂ 3 510',
              key_st='think'),
        'Desktop · liniște ~1,5 s după vorbire: oprit singur, transcrierea finală (ink)',
        panel('d4', False, state=('info', 'Mă gândesc…'), time='0:05', level=0, said=SAY_N, key_st='think'),
        who='mgr', dkey='think')
    # 5 se gândește
    row('Transcriere · se gândește',
        'Mă gândesc… · transcrierea finală e o tastă (creion): apăsată, oprește rularea și devine câmp; dedesubt comanda normalizată',
        panel('p5', True, state=('info', 'Mă gândesc…'), said=SAY_P01, said_btn=True, cmd=CMD_P01, key_st='think'),
        'Desktop · manager · S13 · aria-busy pe tastă, numele rămâne',
        panel('d5', False, state=('info', 'Mă gândesc…'), said=SAY_P12, said_btn=True, cmd=CMD_P12, key_st='think'),
        dctx=D13, dkey='think')
    # 6 transcript corectat
    row('Transcriere corectată',
        'Apăsat pe text: rularea se oprește, textul intră într-un câmp; Execută (primară) îl rulează ca pe unul spus',
        panel('p6', True, state=('off', 'oprit'), edit='pune trei găini moarte la ha cinci punct doi'),
        'Desktop · VO-9 după apăsare: același text, gata de corectat',
        panel('d6', False, state=('off', 'oprit'), edit=SAY_P12),
        dctx=D13)
    # 7 răspuns
    ans_p = ('<p>La Hîrbovăț sunt azi ♀ <span class="ip-num">12 561</span> găini și ♂ <span class="ip-num">794</span> cocoși '
             '(<span class="ip-num">18.11.2026</span>, sapt. <span class="ip-num">59</span>).</p>'
             '<p class="ip-voicepanel__opened">Am deschis Hîrbovăț · Rezumat<button type="button" class="ip-link">Anulează</button></p>')
    ans_d = ('<p>Sapt. <span class="ip-num">59</span> (<span class="ip-num">2</span> din <span class="ip-num">7</span> zile): ouat '
             '<span class="ip-num">40,4%</span>, ghid <span class="ip-num">57,6%</span>, sub ghid cu <span class="ip-num">17,2</span> puncte.</p>'
             + lamp('warn', 'abatere') +
             '<p class="ip-voicepanel__opened">Am deschis Rapoarte · Producție<button type="button" class="ip-link">Anulează</button></p>')
    row('Răspuns',
        'Manager · S4 · pagina care arată cifra se deschide în spate; răspunsul e scris (fără voce), cifre în Mono',
        panel('p7', True, said='câte găini avem la Hîrbovăț?', result=ans_p,
              sr='La Hîrbovăț sunt azi 12 561 găini și 794 cocoși (18.11.2026, sapt. 59).'),
        'Desktop · proprietar · S14 · Lamp + cuvânt pentru stare',
        panel('d7', False, said='care e procentul de ouat săptămâna asta față de ghid?', result=ans_d,
              sr='Sapt. 59 (2 din 7 zile): ouat 40,4%, ghid 57,6%, sub ghid cu 17,2 puncte. Abatere.'),
        who='mgr', dctx=D14)
    # 8 clarificare
    row('Clarificare',
        'Vet · S6 · întrebarea numește grupul (radiogroup); jetoane la înălțime tactilă; Spune din nou',
        panel('p8', True, said='pune două găini moarte la hala cinci', sr='La care secțiune din H5?',
              result=chips('p8', 'La care secțiune din H5?', ['H5.1', 'H5.2'], touch=True, fill=True)
              + f'<div class="ip-voicepanel__keys">{btn("Spune din nou", "secondary", touch=True)}</div>'),
        'Desktop · motiv lipsă la sacrificare: trei alegeri',
        panel('d8', False, said='am sacrificat două găini la ha cinci punct doi', sr='Ce motiv are sacrificarea?',
              result=chips('d8', 'Ce motiv are sacrificarea?', ['Boală', 'Șchiopătare', 'Altul'])
              + f'<div class="ip-voicepanel__keys">{btn("Spune din nou")}</div>'))
    # 9 neînțeles
    neo = lambda ex1, ex2, lead: (f'<div role="alert" class="ip-stack" style="gap:var(--space-1)"><p>{lead}</p>'
                                  f'<p>Încearcă așa: {q(ex1)} sau {q(ex2)}.</p></div>')
    row('Neînțeles',
        'Vet · S2 · două exemple pentru ecran și rol; Spune din nou (primară) · Scrie comanda (focus în câmp, textul păstrat)',
        panel('p9', True, said='pune trei la hala', result=neo('pune 3 găini moarte la H5.1', 'ce lipsește azi?', 'Nu am înțeles „pune trei la hala”.')
              + f'<div class="ip-voicepanel__keys">{btn("Spune din nou", "primary", touch=True)}{btn("Scrie comanda", touch=True)}</div>'),
        'Desktop · 5 s fără vorbire',
        panel('d9', False, result=neo('deschide raportul de producție', 'arată mortalitatea pe ultimele 8 săptămâni', 'Nu am auzit nimic.')
              + f'<div class="ip-voicepanel__keys">{btn("Spune din nou", "primary")}{btn("Scrie comanda")}</div>'))
    # 10 nepermis
    row('Nepermis',
        'Vet · ziua de ieri: același refuz ca ecranul, în același tip de Banner (info, D77) + jetonul spre locul util',
        panel('p10', True, said='pune mortalitatea de ieri la ha șase punct unu',
              result=banner('info', ['Pentru zilele trecute cere managerului.'])
              + f'<div class="ip-chips ip-voicechips" role="group" aria-label="Unde poți merge">'
                f'<button type="button" class="ip-chip ip-voicechip ip-chip--touch">Deschide Introdus azi</button></div>'),
        'Desktop · lista „niciodată”: vocea nu apasă Salvează',
        panel('d10', False, said='salvează zilnicul', result=banner('info', ['Vocea nu salvează. Verifică formularul și apasă Salvează.'])))
    # 11 fără microfon
    row('Fără microfon',
        'Permisiune refuzată · Chrome pe telefon: pașii pentru dispozitiv; câmpul primește focusul',
        panel('p11', True, result=banner('warn', ['Microfonul e blocat pentru Impex Line.',
                                                  'Apasă lacătul de lângă adresă › Permisiuni › Microfon › Permite.', 'Poți scrie comanda mai jos.']),
              field_focus=True),
        'Desktop · pictograma din bara de adrese',
        panel('d11', False, result=banner('warn', ['Microfonul e blocat pentru Impex Line.',
                                                   'Apasă pictograma microfonului din bara de adrese › Permite.', 'Poți scrie comanda mai jos.']),
              field_focus=True))
    # 12 fără conexiune / eroare
    row('Fără conexiune · eroare',
        'Banner de eșec §49 (role=alert), Reîncearcă primește focusul; ce ai spus rămâne în câmp',
        panel('p12', True, result=banner('fault', ['Comanda nu s-a trimis: „pune 3 găini moarte la H5.1”.',
                                                   'Verifică conexiunea la internet, apoi apasă Reîncearcă. Ce ai spus rămâne scris mai jos.'],
                                         key_='Reîncearcă', touch=True, focus_key=True), field='pune 3 găini moarte la H5.1'),
        'Desktop · transcrierea sau AI-ul a eșuat (sau 10 s fără rezultat)',
        panel('d12', False, result=banner('fault', [f'Comanda nu s-a trimis: {q(SAY_P12)}.',
                                                    'Verifică conexiunea la internet, apoi apasă Reîncearcă. Ce ai spus rămâne scris mai jos.'],
                                          key_='Reîncearcă', focus_key=True), field=SAY_P12),
        dctx=D13)
    # desktop extras
    S.append(('Clarificare · mai mult de 3', None, None,
              'Desktop · cele 3 mai probabile + „Alta…” (deschide selectorul ecranului, SectionSelect)',
              on_desk(panel('d13', False, said='deschide secțiunea de la Hîrbovăț', sr='Ce secțiune de la Hîrbovăț?',
                            result=chips('d13', 'Ce secțiune de la Hîrbovăț?', ['H5.1', 'H5.2', 'H6.1'], alta=True)
                            + f'<div class="ip-voicepanel__keys">{btn("Spune din nou")}</div>'),
                      title='Hîrbovăț', crumbs=['Seria 2026'], actions=btn('Mută păsări', 'secondary') + btn('Vânzare', 'secondary'))))
    S.append(('Fără microfon · dispozitiv', None, None, 'Desktop · niciun microfon sau browser fără captură',
              on_desk(panel('d14', False, result=banner('warn', ['Dispozitivul nu are microfon disponibil. Scrie comanda mai jos.']), field_focus=True),
                      **D14)))
    S.append(('Nepermis · fermă', None, None, 'Desktop · vet doar la Hîrbovăț: același refuz ca ecranul',
              on_desk(panel('d15', False, said='deschide Telița', result=banner('info', ['Nu ai acces la această fermă.'])), **D5)))
    S.append(('Nepermis · parole', None, None, 'Desktop · proprietar · lista „niciodată”: parolele nu trec prin voce',
              on_desk(panel('d16', False, said='spune-mi parola lui Ion Rusu', result=banner('info', ['Parolele nu se spun cu voce.'])),
                      title='Utilizatori', crumbs=['Setări'], actions=btn('Adaugă utilizator', 'secondary'))))
    # phone extras (pc == 'PHONE')
    S.append(('Fără microfon', 'PHONE', None, 'iPhone · Safari',
              on_phone(panel('p13', True, result=banner('warn', ['Microfonul e blocat pentru Impex Line.', 'Setări › Safari › Microfon › Permite.',
                                                                 'Poți scrie comanda mai jos.']), field_focus=True))))
    S.append(('Fără microfon', 'PHONE', None, 'iPhone · Impex Line de pe ecranul principal (când e în listă)',
              on_phone(panel('p14', True, result=banner('warn', ['Microfonul e blocat pentru Impex Line.', 'Setări › Impex Line › Microfon › Permite.',
                                                                 'Poți scrie comanda mai jos.']), field_focus=True))))
    S.append(('Fără conexiune înainte de pornire', 'PHONE', None,
              'Tasta era dezactivată („Fără conexiune”); apăsată, panoul se deschide pe Banner, Reîncearcă în focus',
              on_phone(panel('p15', True, result=banner('fault', ['Comanda vocală nu a pornit: fără conexiune.',
                                                                  'Verifică conexiunea la internet, apoi apasă Reîncearcă.'],
                                                        key_='Reîncearcă', touch=True, focus_key=True), key_st='off'))))
    return S


# ---------------------------------------------------------------------------------------------------------- JS demo (VoicePanel)
DEMO_JS = r'''(function(){
var SAY="__SAY__";
var root=document.currentScript.closest(".ip-root"),demo=root.querySelector("[data-vo=demo]");if(!demo)return;
var key=demo.querySelector("[data-vo=key]"),pop=demo.querySelector("[data-vo=pop]"),body=pop.querySelector("[data-vo=body]"),
 st=pop.querySelector("[data-vo=state]"),tm=pop.querySelector("[data-vo=time]"),app=demo.querySelector("[data-vo=app]"),
 sel=demo.querySelector("[data-vo=pick]"),field=pop.querySelector("[data-vo=field]"),run=pop.querySelector("[data-vo=run]"),
 undo=pop.querySelector("[data-vo=undo]"),toast=demo.querySelector("[data-vo=toast]"),
 tpl={};demo.querySelectorAll("template[data-st]").forEach(function(t){tpl[t.dataset.st]=t.innerHTML;});
var timers=[],ret=null,cur="idle";
function isOpen(){return pop.dataset.state==="open";}
function clear(){timers.forEach(clearTimeout);timers.forEach(clearInterval);timers=[];}
function later(f,ms){timers.push(setTimeout(f,ms));}
function setKey(p,busy){key.setAttribute("aria-pressed",p?"true":"false");if(busy)key.setAttribute("aria-busy","true");else key.removeAttribute("aria-busy");}
/* the panel's one polite region: written only when its text changes (stop → think says "Mă gândesc…" once) */
function status(k,w,sr){var h=(k?'<span class="ip-lamp ip-lamp--'+k+'">'+w+'</span>':"")+(sr?'<span class="ip-sr">'+sr+'</span>':"");if(st.innerHTML!==h)st.innerHTML=h;}
function editMode(on,text){run.classList.toggle("ip-btn--primary",on);run.classList.toggle("ip-btn--secondary",!on);undo.hidden=!on;
 if(on){field.value=text;field.focus();field.setSelectionRange(text.length,text.length);}}
function show(s){clear();cur=s;sel.value=s;var d=STATES[s]||{};app.textContent="";if(s!=="edit")editMode(false);
 if(s==="idle"||s==="cancel"||s==="done"){pop.dataset.state="closed";setKey(false,false);st.innerHTML="";
  if(s==="cancel"){app.textContent="Comanda a fost anulată.";(ret||key).focus();}
  if(s==="done"){toast.dataset.state="open";later(function(){toast.dataset.state="closed";},6000);}
  return;}
 pop.dataset.state="open";setKey(true,d.busy);status(d.lamp,d.word,d.sr);tm.textContent=d.time||"";body.innerHTML=tpl[s]||"";
 if(s==="listen")meter();
 if(s==="stop")later(function(){show("think");},700);   /* auto-stop: the key shows the spinner, then thinking */
 if(s==="think")later(function(){show("done");},1600);  /* a show command runs at once: the panel closes, the Toast says it */
 if(s==="edit")editMode(true,SAY);
 if(d.focus){var f=pop.querySelector(d.focus);if(f)f.focus();}}
/* the level meter is a readout: the demo writes a level 0–12 ten times a second from a fixed sample (in the app: the
   microphone's level, at most 15 a second); segments change at once, no easing. Listening stops by itself after the phrase. */
var LV=[2,5,8,7,9,6,8,10,7,5,8,9,6,4,7,8,6,3,2,1,0];
function meter(){var m=pop.querySelector(".ip-voicemeter"),said=pop.querySelector(".ip-voicepanel__said"),i=0,t0=Date.now(),words=SAY.split(" ");
 timers.push(setInterval(function(){var k=i++;if(k<LV.length){m.dataset.level=LV[k];
  var n=Math.min(words.length,Math.floor(k/2));said.textContent=n?"„"+words.slice(0,n).join(" ")+"”":"";
  tm.textContent="0:0"+Math.min(9,Math.floor((Date.now()-t0)/1000));}
  else show("stop");},100));}
var STATES={
 perm:{lamp:"info",word:"aștept permisiunea"},
 listen:{lamp:"ok",word:"Ascult…",time:"0:00"},
 stop:{lamp:"info",word:"Mă gândesc…",busy:true},
 think:{lamp:"info",word:"Mă gândesc…",busy:true},
 edit:{lamp:"off",word:"oprit"},
 answer:{sr:"Sapt. 59 (2 din 7 zile): ouat 40,4%, ghid 57,6%, sub ghid cu 17,2 puncte. Abatere.",focus:".ip-voicepanel__result .ip-link"},
 clarify:{sr:"La care secțiune din H5?",focus:"[role=radio]"},
 nope:{sr:"Nu am înțeles „pune trei la hala”.",focus:".ip-btn--primary"},
 deny:{},nomic:{focus:"[data-vo=field]"},fail:{focus:".ip-banner__key"}};
key.addEventListener("click",function(){
 if(key.getAttribute("aria-disabled")==="true")return;
 if(!isOpen()){ret=key;show("listen");}
 else if(cur==="listen")show("stop");             /* a second tap while listening = stop and run */
 else show("cancel");});                          /* a tap while thinking (or on a result) cancels */
pop.addEventListener("click",function(e){var t=e.target.closest("button");if(!t)return;
 if(t.matches(".ip-voicepanel__close")||t===undo||t.matches(".ip-voicepanel__opened .ip-link"))show("cancel");
 else if(t.matches(".ip-voicepanel__said"))show("edit");
 else if(t.matches("[role=radio]")){t.setAttribute("aria-checked","true");later(function(){show("done");},300);}
 else if(t.textContent==="Spune din nou")show("listen");
 else if(t.textContent==="Scrie comanda")field.focus();
 else if(t.matches(".ip-voiceex")){field.value=t.textContent.replace(/[„”]/g,"");field.focus();}
 else if(t===run||t.textContent==="Reîncearcă")show("think");});
field.addEventListener("keydown",function(e){if(e.key==="Enter"&&!e.shiftKey){e.preventDefault();if(field.value.trim())show("think");}});
/* Esc: one document listener in the capture phase while the panel is open, so voice closes before a SidePanel, sheet or menu */
document.addEventListener("keydown",function(e){
 if(e.key==="Escape"&&isOpen()){e.preventDefault();e.stopPropagation();show("cancel");return;}
 if(e.ctrlKey&&e.shiftKey&&(e.code==="Space"||e.key===" ")){e.preventDefault();   /* the shortcut never moves focus */
  if(cur==="listen")show("stop");else if(!isOpen()){ret=document.activeElement;show("listen");}}},true);
document.addEventListener("click",function(e){if(isOpen()&&!e.target.closest(".ip-voiceanchor")&&!e.target.closest(".vo-demo__bar"))show("cancel");});   /* a click outside cancels */
sel.addEventListener("change",function(){show(sel.value);});
demo.querySelector("[data-vo=reduce]").addEventListener("change",function(e){demo.classList.toggle("ip-motion--reduce",e.target.checked);});
})();'''


def demo():
    """the step-through demo: a desktop page head with the real key and the real absolute popover (§78 motion), a state
    picker, the app-level status line shown; templates hold each state's body (rendered from the same builders)."""
    T = {
        'perm': '<p>Impex Line are nevoie de microfon ca să-ți audă comanda. Apasă Permite în fereastra browserului.</p>' + PRIV,
        'listen': ('<div class="ip-voicepanel__level"><div class="ip-voicemeter" aria-hidden="true" data-level="0">' + '<i></i>' * 12 + '</div>'
                   '<p class="ip-voicepanel__hint">Se oprește singur după o pauză. Apasă din nou ca să termini.</p></div>'
                   '<p class="ip-voicepanel__said is-partial"></p>'),
        'stop': f'<div class="ip-voicepanel__level"><div class="ip-voicemeter" aria-hidden="true" data-level="0">' + '<i></i>' * 12 + f'</div></div><p class="ip-voicepanel__said">{q(SAY_N)}</p>',
        'think': (f'<button type="button" class="ip-voicepanel__said" aria-describedby="dm-fix"><span>{q(SAY_N)}</span>{svg("pencil", 14)}</button>'
                  '<span class="ip-sr" id="dm-fix">Apasă pe text ca să-l corectezi</span>'
                  '<p class="ip-voicepanel__cmd">Hîrbovăț › H5 · Păsări · Mortalitate · 8 săpt. · SAPT 52–59</p>'),
        'edit': '',   # the transcript is edited in the foot field (one field)
        'answer': ('<p class="ip-voicepanel__said">„care e procentul de ouat săptămâna asta față de ghid?”</p><div class="ip-voicepanel__result">'
                   '<p>Sapt. <span class="ip-num">59</span> (<span class="ip-num">2</span> din <span class="ip-num">7</span> zile): ouat '
                   '<span class="ip-num">40,4%</span>, ghid <span class="ip-num">57,6%</span>, sub ghid cu <span class="ip-num">17,2</span> puncte.</p>'
                   + lamp('warn', 'abatere') + '<p class="ip-voicepanel__opened">Am deschis Rapoarte · Producție<button type="button" class="ip-link">Anulează</button></p></div>'),
        'clarify': ('<p class="ip-voicepanel__said">„pune două găini moarte la hala cinci”</p><div class="ip-voicepanel__result">'
                    + chips('dm', 'La care secțiune din H5?', ['H5.1', 'H5.2']) + f'<div class="ip-voicepanel__keys">{btn("Spune din nou")}</div></div>'),
        'nope': ('<p class="ip-voicepanel__said">„pune trei la hala”</p><div class="ip-voicepanel__result"><div role="alert" class="ip-stack" style="gap:var(--space-1)">'
                 '<p>Nu am înțeles „pune trei la hala”.</p><p>Încearcă așa: „deschide raportul de producție” sau „pune 3 găini moarte la H5.1”.</p></div>'
                 f'<div class="ip-voicepanel__keys">{btn("Spune din nou", "primary")}{btn("Scrie comanda")}</div></div>'),
        'deny': ('<p class="ip-voicepanel__said">„șterge vânzarea de azi”</p><div class="ip-voicepanel__result">'
                 + banner('info', ['Vocea nu șterge. Deschide înregistrarea și apasă Șterge.']) + '</div>'),
        'nomic': '<div class="ip-voicepanel__result">' + banner('warn', ['Microfonul e blocat pentru Impex Line.', 'Apasă pictograma microfonului din bara de adrese › Permite.', 'Poți scrie comanda mai jos.']) + '</div>',
        'fail': ('<div class="ip-voicepanel__result">' + banner('fault', [f'Comanda nu s-a trimis: {q(SAY_N)}.', 'Verifică conexiunea la internet, apoi apasă Reîncearcă. Ce ai spus rămâne scris mai jos.'], key_='Reîncearcă') + '</div>'),
    }
    names = [('idle', 'repaus (panou închis)'), ('perm', 'cerere permisiune'), ('listen', 'ascultă'), ('stop', 'oprire automată'),
             ('think', 'se gândește'), ('edit', 'transcriere corectată'), ('answer', 'răspuns'), ('clarify', 'clarificare'),
             ('nope', 'neînțeles'), ('deny', 'nepermis'), ('nomic', 'fără microfon'), ('fail', 'fără conexiune'),
             ('done', 'executat (Toast)'), ('cancel', 'anulare')]
    opts = ''.join(f'<option value="{k}">{w}</option>' for k, w in names)
    tpls = ''.join(f'<template data-st="{k}">{v}</template>' for k, v in T.items())
    pop = ('<section class="ip-voicepanel ip-voicepanel--pop" data-vo="pop" data-state="closed" aria-labelledby="dm-t">'
           '<div class="ip-voicepanel__head"><h2 class="ip-t-heading" id="dm-t">Comandă vocală</h2>'
           '<p class="ip-voicepanel__state" role="status" aria-live="polite" data-vo="state"></p>'
           '<span class="ip-voicepanel__time" data-vo="time"></span>'
           f'<button type="button" class="ip-iconbtn ip-voicepanel__close" aria-label="Închide">{svg("x")}</button></div>'
           '<div class="ip-voicepanel__body" data-vo="body"></div>'
           + foot('dm', False, data=True) + '</section>')
    toast = ('<div class="ip-toastwrap ip-toastwrap--page"><div class="ip-toast" role="status" data-vo="toast" data-state="closed">'
             f'{svg("mic", 16)}<span class="ip-toast__txt">Am deschis Hîrbovăț › H5 · Păsări: Mortalitate, 8 săpt.</span>'
             '<button type="button" class="ip-link ip-toast__undo">Anulează</button></div></div>')
    return ('<section class="vo-demo" data-vo="demo" aria-label="Demo: comandă vocală pas cu pas">'
            '<div class="vo-demo__bar"><label class="ip-field" style="width:calc(var(--space-12) * 6)"><span class="ip-label">Stare</span>'
            f'<span class="ip-input"><select data-vo="pick" style="all:unset;flex:1;cursor:pointer">{opts}</select></span></label>'
            '<label class="ip-check" style="align-self:flex-end;min-height:var(--control-h);display:inline-flex;align-items:center;gap:var(--space-2)">'
            '<input type="checkbox" data-vo="reduce"> Mișcare redusă</label>'
            '<p class="ip-note ip-m0" style="align-self:flex-end;max-width:calc(var(--space-12) * 11)">Apasă tasta (sau Ctrl+Shift+Space): ascultă, '
            'nivelul se aprinde din valori, se oprește singur după fraza de probă, apoi se gândește și răspunde. A doua apăsare oprește; '
            'Esc, Închide sau o apăsare în timp ce se gândește anulează.</p></div>'
            '<div class="vo-desk vo-demo__page"><div class="ip-pagehead"><div><nav class="ip-crumbs" aria-label="Unde ești"><a href="#">Seria 2026</a>'
            '<span aria-hidden="true">›</span><a href="#">Hîrbovăț</a><span aria-hidden="true">›</span><span aria-current="page">H5</span></nav>'
            '<h1 class="ip-t-title ip-m0">H5</h1></div><div class="ip-row" style="gap:var(--space-2)">' + S5_ACT
            + '<span class="ip-voiceanchor">' + key('head', 'rest', 'dm', demo=True) + pop + '</span></div></div>'
            '<p class="ip-note" style="margin:var(--space-6) 0 0">Stare aplicație (role=status, în shell): <span data-vo="app" role="status" aria-live="polite"></span></p>'
            + toast + '</div>' + tpls + '</section>')


# ---------------------------------------------------------------------------------------------------------- VoiceButton preview
def strip():
    """the key's states 1:1"""
    sts = [('rest', 'Repaus'), ('hover', 'Hover · sunken'), ('focus', 'Focus · inel 2 px'), ('listen', 'Ascultă · aria-pressed'),
           ('think', 'Se gândește · aria-busy'), ('off', 'Dezactivat · „Fără conexiune”')]
    def cell(form, s, w):
        k = key(form, 'rest' if s == 'focus' else s, f'k{form}{s}', focus=(s == 'focus'))
        return f'<div class="vo-keycell">{k}<span class="ip-note">{w}</span></div>'
    return ('<div class="vo-strip">' + cap('Telefon · 56 rotundă · raised + line-strong + shadow-overlay (D199) · mic 22')
            + '<div class="vo-keys">' + ''.join(cell('fab', s, w) for s, w in sts) + '</div>'
            + cap('Desktop · cheia din PageHeader · control-h 32 · aspect IconButton secundar · mic 16 · tooltip „Comandă vocală · Ctrl+Shift+Space”')
            + '<div class="vo-keys">' + ''.join(cell('head', s, w) for s, w in sts) + '</div></div>')


def place_tabbar():
    return phone('<div class="ip-screen ip-screen--fixed"><div class="ip-screen__view">' + ferma_mgr()
                 + '</div>' + tabbar(TABS_5) + key('fab', 'rest', 'pt', float_=True) + '</div>', h=420)


def place_savebar():
    sb = ('<div class="ip-savebar ip-screen__savebar"><span class="ip-savebar__sum">'
          f'<span class="ip-rec ip-rec--feed">{svg("feed", 14)}</span><span><span class="ip-num">4</span> rânduri · '
          '<span class="ip-num">2 303,2 kg</span></span></span>'
          '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch">Salvează</button></div>')
    head = (f'<header class="ip-formhead"><a class="ip-iconbtn ip-iconbtn--touch" href="#inregistreaza" aria-label="Înapoi la Înregistrează">{svg("back")}</a>'
            f'<span class="ip-tile ip-rec ip-rec--feed" aria-hidden="true">{svg("feed")}</span><div><h1 class="ip-t-heading ip-m0">Hrană zilnică</h1>'
            '<p class="ip-note ip-num ip-m0">HÎRBOVĂȚ · 18.11.2026 · SAPT 59</p></div></header>')
    return phone('<div class="ip-screen ip-screen--fixed">' + head + '<div class="ip-screen__view"><main class="ip-screen__main">'
                 '<p class="ip-note ip-m0">Grame pe pasăre · ♀ ghid <span class="ip-num">157 g</span></p></main></div>'
                 + sb + key('fab', 'rest', 'ps', float_=True) + '</div>', h=420)


def place_toast(cancel=False):
    toast = ('<div class="ip-toastwrap"><div class="ip-toast" role="status">' + svg('mic', 16)
             + '<span class="ip-toast__txt">Am deschis Hîrbovăț · Producție</span><button type="button" class="ip-link ip-toast__undo">Anulează</button></div></div>')
    status = '<p class="ip-sr" role="status" aria-live="polite">Comanda a fost anulată.</p>' if cancel else ''
    return phone('<div class="ip-screen ip-screen--fixed"><div class="ip-screen__view">' + ferma_mgr() + ('' if cancel else toast) + '</div>'
                 + tabbar(TABS_5) + key('fab', 'rest', 'pc' if cancel else 'px', float_=True, focus=cancel) + status + '</div>', h=420)


def place_sheet():
    items = [('mortality', 'Zilnic', 'Mortalitate, sacrificare, ouă'), ('feed', 'Hrană zilnică', 'Grame pe găină și cocoș')]
    rows = ''.join(f'<a class="ip-ev" href="#"><span class="ip-tile ip-rec ip-rec--{ic}" aria-hidden="true">{svg(ic)}</span>'
                   f'<div class="ip-ev__main">{t}<div class="ip-ev__meta">{m}</div></div></a>' for ic, t, m in items)
    sheet = (f'<div class="ip-scrim ip-scrim--sheet">{key("fab", "rest", "pb")}<div class="ip-sheet" role="dialog" aria-modal="true" aria-label="Înregistrează">'
             f'<div class="ip-sheet__handle"></div><p class="ip-t-heading" style="margin:0;padding:var(--space-1) var(--space-4) var(--space-2)">Înregistrează</p>'
             f'<div class="ip-list" style="border-radius:0;border-left:0;border-right:0">{rows}</div></div></div>')
    return phone('<div class="ip-screen ip-screen--fixed ip-scrimhost"><div class="ip-screen__view">' + acasa_vet() + '</div>'
                 + tabbar(TABS_VET) + sheet + '</div>', h=420)


def place_keyboard():
    return phone('<div class="ip-screen ip-screen--fixed"><div class="ip-screen__view"><main class="ip-screen__main">'
                 '<div class="ip-field"><label class="ip-label" for="kb-n">Notă (opțional)</label><span class="ip-input ip-input--touch is-focus">'
                 '<input id="kb-n" type="text" value="Livrare de dimineață"></span></div>'
                 '<p class="ip-note ip-m0">Tastatura de pe ecran e deschisă (citit din visual viewport): tasta vocală are <code>hidden</code> și '
                 'iese din ordinea de citire; Ctrl+Shift+Space merge de pe o tastatură fizică.</p></main></div>'
                 + tabbar(TABS_5) + key('fab', 'rest', 'pk', float_=True, hidden=True)
                 + '<div class="vo-kbd" aria-hidden="true">tastatura telefonului</div></div>', h=420)


def place_s16():
    """the S16 Corectează full-screen sheet: the key space-4 above the sheet foot (P15, illustrative record)"""
    return phone('<div class="ip-screen ip-screen--fixed ip-scrimhost"><aside class="ip-sidepanel ip-sidepanel--sheet" role="dialog" aria-modal="true" aria-labelledby="s16-h">'
                 '<div class="ip-sidepanel__head" style="display:flex;justify-content:space-between;align-items:center;padding:var(--space-2) var(--space-2) var(--space-2) var(--space-4)">'
                 f'<div class="ip-row ip-sidepanel__title"><span class="ip-tile ip-rec ip-rec--mortality" aria-hidden="true">{svg("mortality")}</span>'
                 '<div><h2 class="ip-t-heading ip-m0" id="s16-h">Corectează · Mortalitate H6.1</h2><p class="ip-note ip-num ip-m0">14.11.2026 · Dr. Ceban</p></div></div>'
                 f'<button type="button" class="ip-iconbtn ip-iconbtn--touch" aria-label="Închide">{svg("x")}</button></div>'
                 '<div class="ip-sidepanel__body"><div class="ip-field"><label class="ip-label" for="s16-m">Motiv</label>'
                 '<span class="ip-input ip-input--touch"><input id="s16-m" type="text"></span></div></div>'
                 '<div class="ip-sidepanel__foot"><button type="button" class="ip-btn ip-btn--secondary ip-btn--touch">Anulează</button>'
                 '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch">Salvează corecția</button></div>'
                 + key('fab', 'rest', 'p16', float_=True) + '</aside></div>', h=420)


def desk_toasts():
    t1 = ('<div class="ip-toast" role="status">' + svg('mic', 16) + '<span class="ip-toast__txt">Am deschis Hîrbovăț › H5 · Păsări: Mortalitate, 8 săpt.</span>'
          '<button type="button" class="ip-link ip-toast__undo">Anulează</button></div>')
    t0 = ('<div class="ip-toast" role="status">' + svg('mic', 16) + '<span class="ip-toast__txt">Am deschis Rapoarte · Producție, ultimele 8 săptămâni</span>'
          '<button type="button" class="ip-link ip-toast__undo">Anulează</button></div>')
    t2 = ('<div class="ip-toast" role="status">' + svg('download', 16) + '<span class="ip-toast__txt">Export .xlsx descărcat · Hîrbovăț · Păsări · sapt. 20–59</span></div>')
    return ('<div class="vo-desk vo-desk--page">' + '<div class="ip-pagehead"><div><nav class="ip-crumbs" aria-label="Unde ești"><a href="#">Seria 2026</a>'
            '<span aria-hidden="true">›</span><a href="#">Hîrbovăț</a><span aria-hidden="true">›</span><span aria-current="page">H5</span></nav>'
            '<h1 class="ip-t-title ip-m0">H5</h1></div><div class="ip-row" style="gap:var(--space-2)">' + S5_ACT
            + f'<span class="ip-voiceanchor">{key("head", "rest", "dt")}</span></div></div>'
            '<div class="ip-toastwrap ip-toastwrap--page" style="display:flex;flex-direction:column;align-items:flex-end;gap:var(--space-2)">' + t0 + t1 + t2 + '</div></div>')


def vb_preview():
    pl = [('Deasupra BottomTabBar · manager, 5 itemi · space-4 de la margine și de bară', place_tabbar()),
          ('Deasupra SaveBar (formular) · niciodată peste bară sau tasta ei', place_savebar()),
          ('Executat · Toast-ul se așază space-2 deasupra tastei', place_toast()),
          ('Peste un BottomSheet · pe scrim, deasupra marginii sheet-ului', place_sheet()),
          ('Tastatura de pe ecran deschisă · tasta ascunsă', place_keyboard()),
          ('Anulare · focusul revine pe tastă, „Comanda a fost anulată.” (status)', place_toast(cancel=True)),
          ('S16 Corectează · foaia pe tot ecranul: space-4 deasupra piciorului foii', place_s16())]
    grid = '<div class="vo-grid3">' + ''.join(f'<div>{cap(c)}{h}</div>' for c, h in pl) + '</div>'
    return ('<div class="ip-root ip-stack vo-root" style="gap:var(--space-8)">' + CSS_TAG + strip()
            + f'<div>{cap("Unde stă tasta · telefon 390 (piciorul ecranului)")}{grid}</div>'
            + f'<div>{cap("Desktop 1440 · ultima tastă din acțiunile PageHeader · executat: Toast cu Anulează; export: fără Anulează")}{desk_toasts()}</div>'
            + '<p class="ip-note ip-m0" style="max-width:calc(var(--panel-w) * 2)">Toggle: aria-pressed (adevărat cât ascultă și cât panoul e deschis pe un '
              'rezultat), aria-busy cât se gândește, aria-keyshortcuts="Control+Shift+Space"; fără aria-expanded. Ținte: 56 pe telefon, 32 pe desktop. '
              'Blocarea și deblocarea sunt doar culoare (120 ms); apăsarea 0,97; sub mișcare redusă rămâne culoarea.</p></div>')


# ---------------------------------------------------------------------------------------------------------- VoicePanel preview
def vp_preview(with_demo=True):
    S = panel_states()
    rows = ''
    for name, pc, pd, dc, dd in S:
        if pd is None:
            continue
        rows += (f'<div class="vo-row"><div class="vo-rowhead"><h3 class="ip-t-heading ip-m0">{name}</h3></div>'
                 f'<div>{cap("390 · " + pc)}{pd}</div><div>{cap("1440 · " + dc)}{dd}</div></div>')
    extra = ''.join(f'<div>{cap(name + " · " + dc)}{dd}</div>' for name, pc, pd, dc, dd in S if pd is None and pc != 'PHONE')
    pext = ''.join(f'<div>{cap(name + " · " + dc)}{dd}</div>' for name, pc, pd, dc, dd in S if pc == 'PHONE')
    dm = (f'<div>{cap("Demo · pas cu pas (script mic; pe canvas desenat static mai jos)")}{demo()}</div>' if with_demo else '')
    js = f'<script>{DEMO_JS.replace("__SAY__", SAY_N)}</script>' if with_demo else ''
    return ('<div class="ip-root ip-stack vo-root" style="gap:var(--space-8)">' + CSS_TAG + dm + rows
            + f'<div class="vo-ext"><h3 class="ip-t-heading ip-m0">Variante telefon · 390</h3><div class="vo-grid3">{pext}</div></div>'
            + f'<div class="vo-ext"><h3 class="ip-t-heading ip-m0">Variante desktop · 1440</h3><div class="vo-grid2">{extra}</div></div>'
            + '<p class="ip-note ip-m0" style="max-width:calc(var(--panel-w) * 2)">Telefon: BottomSheet (până la 60% din ecran, corpul derulează), '
              'tasta plutitoare andocată în picior; desktop: popover sub tasta din PageHeader, marginile din dreapta aliniate, panel-w. Nu e capcană '
              'de focus. O singură regiune role=status în cap (stările, un răspuns, o întrebare); erorile și refuzurile role=alert. Nivelul are '
              '12 segmente, aria-hidden, scris din valoarea microfonului (≤ 15/s), fără easing, și sub mișcare redusă. Mișcare: sheet 280 ms, '
              'popover 120 ms (scale 0,95 din colțul de sus-dreapta); sub mișcare redusă doar fade.</p>' + js + '</div>')


# ---------------------------------------------------------------------------------------------------------- VoiceMark preview
def mark(uid, before=None):
    b = f' · înainte <span class="ip-num">{before}</span>' if before is not None else ''
    return (f'<span class="ip-voicemark" id="{uid}" aria-hidden="true">{svg("mic", 14)}<span class="ip-sr">completat </span>din voce{b}</span>')


def count_field(uid, sx, rec, word, value, sec='t-H51', hint='', vm=True, before=None, cls=''):
    tone = 'hen' if sx == '♀' else 'rooster'
    m = mark(f'{uid}-vm', before) if vm else ''
    d = ' '.join(x for x in ((f'{uid}-vm' if vm else ''), (f'{uid}-h' if hint else '')) if x)
    return (f'<div class="ip-field is-{tone}"><label class="ip-label" for="{uid}" id="{uid}-l"><span class="ip-rec ip-rec--{rec}">{svg(rec, 14)}</span>'
            f'{word} {SXS.sign(sx)}{m}</label><span class="ip-input ip-count ip-count--{tone}{cls}"><input id="{uid}" inputmode="numeric" pattern="[0-9]*" '
            f'placeholder="–" autocomplete="off" value="{value}" aria-labelledby="{sec} {uid}-l"' + (f' aria-describedby="{d}"' if d else '') + '></span>'
            + (f'<span class="ip-entry__hint" id="{uid}-h">{hint}</span>' if hint else '') + '</div>')


def text_field(uid, label, value, vm=True, focus=False, suffix='', before=None, mono=False, select=False, hint=''):
    m = mark(f'{uid}-vm', before) if vm else ''
    cls = ' is-focus' if focus else ''
    ids = ' '.join(x for x in ((f'{uid}-vm' if vm else ''), (f'{uid}-h' if hint else '')) if x)
    d = f' aria-describedby="{ids}"' if ids else ''
    if select:
        ctl = f'<select id="{uid}"{d} style="all:unset;flex:1"><option>{value}</option></select>{svg("chev", 16)}'
    else:
        ctl = (f'<input id="{uid}" type="text" value="{value}"{" inputmode=\"decimal\"" if suffix else ""}{d}'
               f'{" style=\"font-family:var(--font-mono)\"" if mono else ""}>' + (f'<span class="suffix">{suffix}</span>' if suffix else ''))
    h = f'<span class="ip-hint" id="{uid}-h">{hint}</span>' if hint else ''
    return f'<div class="ip-field"><label class="ip-label" for="{uid}">{label}{m}</label><span class="ip-input{cls}">{ctl}</span>{h}</div>'


def marks_gallery():
    cells = [
        ('CountInput · S6 Zilnic · ♀', count_field('g1', '♀', 'mortality', 'Mortalitate', '3', sec='g1-l')),
        ('CountInput · o valoare scrisă înlocuită (exemplu)', count_field('g2', '♀', 'mortality', 'Mortalitate', '3', sec='g2-l', before=2)),
        ('TextField · S13 Kg', text_field('g3', 'Kg', '25 640', suffix='kg', mono=True)),
        ('Select · S13 Tip hrană', text_field('g4', 'Tip hrană', 'Breeder 2', select=True, hint='♀ pentru găini')),
        ('Checkbox · S6 Fără pierderi azi (P05)',
         '<div class="ip-field"><label class="ip-check is-checked">'
         f'<input type="checkbox" checked class="ip-sr" aria-describedby="g5-vm"><span class="ip-check__box">{svg("check", 14)}</span>Fără pierderi azi'
         + mark('g5-vm') + '</label></div>'),
        ('Fără marcaj · un câmp pe care comanda nu l-a numit (gol = null, nu 0)', count_field('g6', '♂', 'cull', 'Sacrificare', '', sec='g6-l', vm=False)),
    ]
    return '<div class="vo-marks">' + ''.join(f'<div class="vo-markcell">{cap(c)}{h}</div>' for c, h in cells) + '</div>'


def prefilled_phone():
    """VO-3: S6 Zilnic, 18.11.2026 08:20 (before the day's saves), „la H5.1 au murit trei găini și un cocoș”"""
    head = (f'<header class="ip-formhead"><a class="ip-iconbtn ip-iconbtn--touch" href="#inregistreaza" aria-label="Înapoi la Înregistrează">{svg("back")}</a>'
            f'<span class="ip-tile ip-rec ip-rec--mortality" aria-hidden="true">{svg("mortality")}</span><div><h1 class="ip-t-heading ip-m0">Zilnic</h1>'
            '<p class="ip-note ip-num ip-m0">HÎRBOVĂȚ · 18.11.2026 · SAPT 59</p></div></header>')
    bnr = banner('info', ['Am completat din ce ai spus. Verifică și salvează.',
                          '2 câmpuri: <span class="ip-num">H5.1</span> · ♀ Mortalitate <span class="ip-num">3</span> · ♂ Mortalitate <span class="ip-num">1</span>',
                          '<button type="button" class="ip-chip ip-voicechip ip-chip--touch">Continuă cu Hrană zilnică</button>'],
                 key_='Anulează completarea', role='', touch=True)
    card = ('<section class="ip-panel ip-entry" aria-labelledby="pf-t"><div class="ip-entry__head"><h3 class="ip-entry__code" id="pf-t">H5.1</h3>'
            '<span class="ip-entry__alive">în viață<span class="ip-sr"> la începutul zilei</span> <span class="ip-pair"><span class="ip-sex ip-sex--hen">♀ <b>4 264</b></span>'
            '<span class="sep">·</span><span class="ip-sex ip-sex--rooster">♂ <b>342</b></span></span></span></div><div class="ip-entry__grid">'
            + count_field('pf-m0', '♀', 'mortality', 'Mortalitate', '3', sec='pf-t', cls=' is-focus',
                          hint=f'<span>ieri {SXS.sign("♀")} <span class="ip-num">0</span></span><span aria-hidden="true">·</span><span>media 7 zile <span class="ip-num">0,6</span></span>')
            + count_field('pf-m1', '♂', 'mortality', 'Mortalitate', '1', sec='pf-t',
                          hint=f'<span>ieri {SXS.sign("♂")} <span class="ip-num">0</span></span><span aria-hidden="true">·</span><span>media 7 zile <span class="ip-num">0,0</span></span>')
            + count_field('pf-c0', '♀', 'cull', 'Sacrificare', '', sec='pf-t', vm=False)
            + count_field('pf-c1', '♂', 'cull', 'Sacrificare', '', sec='pf-t', vm=False) + '</div></section>')
    sb = ('<div class="ip-savebar ip-screen__savebar"><span class="ip-savebar__sum">'
          f'<span class="ip-rec ip-rec--mortality">{svg("mortality", 14)}</span><span><span class="ip-num">1</span> secțiune · '
          '♀ <span class="ip-num">−3</span> · ♂ <span class="ip-num">−1</span></span></span>'
          '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch">Salvează</button></div>')
    status = '<p class="ip-sr" role="status" aria-live="polite">Am completat 2 câmpuri din ce ai spus. Verifică și salvează.</p>'
    return phone('<div class="ip-screen ip-screen--fixed">' + head + '<div class="ip-screen__view"><main class="ip-screen__main">'
                 + bnr + card + '</main></div>' + sb + key('fab', 'rest', 'pf', float_=True) + status + '</div>', h=844)


def prefilled_desk():
    """VO-10: S13 Livrare hrană, manager, „am primit 25 640 de kilograme pentru găini la Hîrbovăț pe 13 noiembrie”"""
    bnr = banner('info', ['Am completat din ce ai spus. Verifică și salvează.',
                          '4 câmpuri: Hîrbovăț · <span class="ip-num">13.11.2026</span> · Breeder 2 · <span class="ip-num ip-nowrap">25 640 kg</span>'],
                 key_='Anulează completarea', role='')
    body = (bnr + text_field('pd-f', 'Fermă', 'Hîrbovăț', select=True, focus=True)
            + text_field('pd-d', 'Data', '13.11.2026', mono=True)
            + text_field('pd-t', 'Tip hrană', 'Breeder 2', select=True, hint='♀ pentru găini')
            + text_field('pd-k', 'Kg', '25 640', suffix='kg', mono=True)
            + text_field('pd-n', 'Notă (opțional)', '', vm=False))
    side = (f'<aside class="ip-sidepanel" aria-labelledby="pd-h" style="height:100%"><div class="ip-sidepanel__head" style="display:flex;justify-content:space-between;align-items:center;padding:var(--space-3) var(--space-4)">'
            f'<div class="ip-row ip-sidepanel__title"><span class="ip-tile ip-rec ip-rec--feed" aria-hidden="true">{svg("feed")}</span>'
            '<div><h2 class="ip-t-heading ip-m0" id="pd-h">Livrare hrană</h2><p class="ip-note ip-m0">Manager</p></div></div>'
            f'<button type="button" class="ip-iconbtn" aria-label="Închide">{svg("x")}</button></div>'
            f'<div class="ip-sidepanel__body">{body}</div><div class="ip-sidepanel__foot"><span></span><div class="ip-row ip-sidepanel__keys">'
            '<button type="button" class="ip-btn ip-btn--secondary">Anulează</button><button type="button" class="ip-btn ip-btn--primary">Adaugă livrarea</button></div></div></aside>')
    return ('<div class="vo-desk vo-desk--panel"><div class="ip-pagehead"><div><nav class="ip-crumbs" aria-label="Unde ești"><a href="#">Rapoarte</a>'
            '<span aria-hidden="true">›</span><span aria-current="page">Hrană</span></nav><h1 class="ip-t-title ip-m0">Hrană</h1></div>'
            f'<div class="ip-row" style="gap:var(--space-2)"><span class="ip-voiceanchor">{key("head", "rest", "pd")}</span></div></div>'
            f'<div class="vo-dockcol">{side}</div>'
            '<p class="ip-sr" role="status" aria-live="polite">Am completat 4 câmpuri din ce ai spus. Verifică și salvează.</p></div>')


def vm_preview():
    return ('<div class="ip-root ip-stack vo-root" style="gap:var(--space-8)">' + CSS_TAG
            + f'<div>{cap("VoiceMark · în linia etichetei, după etichetă · mic 14 + „din voce”, small 12/16, info · descrierea câmpului: „completat din voce”")}{marks_gallery()}</div>'
            + '<div class="vo-pre">'
            + f'<div>{cap("390 · S6 Zilnic · formular precompletat (18.11.2026, 08:20; cifrele spuse sunt exemplu; comanda a numit și hrana: jetonul „Continuă cu …”) · focus pe primul câmp completat · tasta deasupra SaveBar")}{prefilled_phone()}</div>'
            + f'<div style="flex:1;min-width:0">{cap("1440 · S13 Livrare hrană · manager · Banner cu acțiune (FIN03) primul bloc · focus pe primul câmp completat (Fermă)")}{prefilled_desk()}</div></div>'
            + '<p class="ip-note ip-m0" style="max-width:calc(var(--panel-w) * 2)">Marcajul rămâne până când câmpul e editat de mână, până la '
              'Anulează completarea sau până la salvare. Chenarul câmpului, bara de sex, avertismentul și refuzul rămân ale câmpului. O salvare care '
              'mai poartă un marcaj „din voce” scrie în Istoric modificări sursa: voce și textul comenzii.</p></div>')


# ---------------------------------------------------------------------------------------------------------- preview-only layout
CSS = ('.vo-root{max-width:none}'
       '.vo-cap{display:block;margin:0 0 var(--space-2)}'
       '.vo-phone{width:390px;max-width:none;position:relative;box-sizing:border-box}'
       '.vo-phone > .ip-screen{height:100%}'
       '.vo-strip{display:flex;flex-direction:column;gap:var(--space-3)}'
       '.vo-keys{display:flex;gap:var(--space-6);align-items:flex-start;flex-wrap:wrap;padding:var(--space-6);background:var(--ground);border:var(--ip-hair) solid var(--line);border-radius:var(--radius-md)}'
       '.vo-keycell{display:flex;flex-direction:column;align-items:center;gap:var(--space-3);width:calc(var(--space-12) * 3);text-align:center}'
       '.vo-grid3{display:grid;grid-template-columns:repeat(3,390px);gap:var(--space-8) var(--space-6)}'
       '.vo-desk{position:relative;background:var(--ground);border:var(--ip-hair) solid var(--line);border-radius:var(--radius-md);padding:var(--space-6) var(--space-8);box-sizing:border-box;min-width:0}'
       '.vo-desk--page{min-height:calc(var(--space-12) * 7)}'
       '.vo-desk--panel{padding-right:0;padding-bottom:0;display:flex;flex-direction:column;gap:var(--space-4)}'
       '.vo-desk--panel > .ip-pagehead{padding-right:var(--space-8)}'
       '.vo-dockcol{display:flex;justify-content:flex-end;min-height:0}'
       '.vo-row{display:grid;grid-template-columns:calc(var(--space-12) * 3) 390px minmax(0,1fr);gap:var(--space-6);align-items:start;padding-top:var(--space-6);border-top:var(--ip-hair) solid var(--rule)}'
       '.vo-ext{display:flex;flex-direction:column;gap:var(--space-4);padding-top:var(--space-6);border-top:var(--ip-hair) solid var(--rule)}'
       '.vo-grid2{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:var(--space-8) var(--space-6)}'
       '.vo-kbd{position:absolute;left:0;right:0;bottom:0;height:calc(var(--space-12) * 5);z-index:4;display:flex;align-items:center;justify-content:center;background:var(--sunken);border-top:var(--ip-hair) solid var(--line);color:var(--ink-muted);font-size:var(--ip-fs-small)}'
       '.vo-marks{display:grid;grid-template-columns:repeat(3,calc(var(--space-12) * 5));gap:var(--space-6) var(--space-8);align-items:start}'
       '.vo-markcell .ip-field{max-width:calc(var(--space-12) * 4)}'
       '.vo-pre{display:flex;gap:var(--space-8);align-items:flex-start}.vo-pre > :first-child{flex:none;width:390px}'
       '.vo-demo{display:flex;flex-direction:column;gap:var(--space-4)}'
       '.vo-demo__bar{display:flex;gap:var(--space-6);flex-wrap:wrap;align-items:flex-end}'
       '.vo-demo__page{min-height:calc(var(--space-12) * 11)}')
CSS_TAG = f'<style>{CSS}</style>'


def card(group, comp, html, sub='VO01'):
    return f'<!-- @dsCard group="{group}" height={CARD_H[comp]} subtitle="{sub}" -->\n' + SXS.sexify(html) + '\n'


def board(parts):
    body = SXS.sexify(''.join(parts))
    return f'''<!doctype html>
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
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500&amp;family=IBM+Plex+Mono:wght@400;500&amp;display=swap">
<style>body{{margin:0}}</style>
</helmet>
<div class="ip-root" style="padding:var(--space-6);min-height:{BOARD_H}px;box-sizing:border-box">
<h1 class="ip-t-title" style="margin:0 0 var(--space-2)">V1 · Voce</h1>
<p class="ip-note" style="margin:0 0 var(--space-8)">Componentele sistemului pentru comenzi vocale (spec §8 F1, D199): VoiceButton, VoicePanel, VoiceMark · additions.css §78 · VO01. Static: demo-ul pas cu pas e în preview-ul VoicePanel.</p>
{body}
</div>
</x-dc>
<script type="text/x-dc" data-dc-script data-props='{{"$preview":{{"width":{BW},"height":{BOARD_H}}}}}'>
class Component extends DCLogic {{ renderVals() {{ return {{}}; }} }}
</script>
</body>
</html>
'''


BOARD_H = 17280   # vo01_render.py --board: its height at 1440 (17 226, fallback fonts) + slack

if __name__ == '__main__':
    vb, vp, vm = vb_preview(), vp_preview(), vm_preview()
    out = {'VoiceButton': card('App shell', 'VoiceButton', vb), 'VoicePanel': card('Containers', 'VoicePanel', vp),
           'VoiceMark': card('Inputs', 'VoiceMark', vm)}
    for comp, html in out.items():
        (ROOT / comp).mkdir(exist_ok=True)
        p = ROOT / comp / 'preview.html'
        if not p.exists() or p.read_text(encoding='utf-8') != html:
            p.write_text(html, encoding='utf-8')
    bd = board([vb, '<div style="height:var(--space-12)"></div>', vp_preview(with_demo=False), '<div style="height:var(--space-12)"></div>', vm])
    bo = REPO / f'design/canvas/project/{STEM}.dc.html'
    if not bo.exists() or bo.read_text(encoding='utf-8') != bd:
        bo.write_text(bd, encoding='utf-8')
    if CANVAS:
        import canvas_rows as CR
        c = CR.load(REPO); B = c['boards']; f = f'{STEM}.dc.html'
        if f in B:
            B[f].update(w=BW, h=BOARD_H, title=TITLE)
        else:
            m1 = B['M1-miscare.dc.html']; y = m1['y'] + m1['h'] + CR.GAP_Y
            for k, v in B.items():   # its own row right under M1: every later row moves down by the board + 120
                if v['y'] >= y: v['y'] += BOARD_H + CR.GAP_Y
            B[f] = {'x': 0, 'y': y, 'w': BW, 'h': BOARD_H, 'expand': 'fill', 'title': TITLE}
            c['order'].append(f)
        if CR.overlaps(c):
            for mv in CR.relayout(c): print('relayout', *mv)
        CR.sort_order(c); CR.save(REPO, c)
    print('ok', {k: len(v) for k, v in out.items()}, len(bd))
