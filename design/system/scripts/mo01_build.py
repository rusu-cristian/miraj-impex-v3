"""Build the MO01 motion demos (D196, additions.css §75): python3 -I mo01_build.py <repo_root> [--canvas]
Writes design/system/components/Motion/preview.html (full page, plain JS, not on the system card list) and the canvas board
design/canvas/project/M1-miscare.dc.html ("M1 · Mișcare", interactive: the same demos driven by the board's DCLogic state).
One markup for both: bound attributes are `{{ key }}` holes on the board and initial values + data-b="attr:key;…" in the
preview; one JS core (class MoCore: init / act / vals) computes every bound value for both. Figures: the PeriodTable comes
from sys14_build (realdata, Hîrbovăț Seria 2026, SAPT 52–59); the Readout is the S4 Rezumat one (18.11.2026).
--canvas: adds or refreshes the board's canvas.json entry as its own row right under Main (y 800), then run
`canvas_rows.py relayout <root> --canvas` and `main_build.py <root> --canvas`."""
import sys, re, json, pathlib
HERE = pathlib.Path(__file__).resolve().parent
REPO = pathlib.Path(sys.argv[1]) if len(sys.argv) > 1 and not sys.argv[1].startswith('--') else HERE.parents[2]
CANVAS = '--canvas' in sys.argv
sys.path.insert(0, str(HERE)); sys.path.insert(0, str(REPO / 'design/canvas/scripts'))
import sys14_build as P   # PeriodTable builders (its module-level preview html is built, not written)
import sexsign as SXS

W, H = 1440, 3472   # board frame; H = the preview's height at 1440 (3462, fallback fonts) + slack
STEM = 'M1-miscare'
TITLE = 'M1 · Mișcare'

def svg(paths, size=None):
    c = f'ip-icon ip-icon--{size}' if size else 'ip-icon'
    return f'<svg class="{c}" viewBox="0 0 24 24" aria-hidden="true">{paths}</svg>'
I = {
    'x': '<path d="M18 6 6 18"/><path d="m6 6 12 12"/>',
    'check': '<path d="M20 6 9 17l-5-5"/>',
    'chev': '<path d="m6 9 6 6 6-6"/>',
    'chevup': '<path d="m18 15-6-6-6 6"/>',
    'fault': '<circle cx="12" cy="12" r="10"/><path d="M12 8v4"/><path d="M12 16h.01"/>',
    'mortality': '<path d="M12 3a7 7 0 0 0-7 7v4l2 2v3h10v-3l2-2v-4a7 7 0 0 0-7-7z"/><circle cx="9.5" cy="11" r="1.2"/><circle cx="14.5" cy="11" r="1.2"/><path d="M11 19v-2M13 19v-2"/>',
    'feed': '<path d="M12 21V9M12 9c-3 0-4-2-4-5 3 0 4 2 4 5zM12 9c3 0 4-2 4-5-3 0-4 2-4 5zM12 15c-3 0-4-2-4-5 3 0 4 2 4 5zM12 15c3 0 4-2 4-5-3 0-4 2-4 5z"/>',
    'weighing': '<path d="M12 3v18"/><path d="M7 21h10"/><path d="M3 7h18"/><path d="m6 7-3 7a3 3 0 0 0 6 0z"/><path d="m18 7-3 7a3 3 0 0 0 6 0z"/>',
    'transfer': '<path d="m16 3 4 4-4 4"/><path d="M20 7H4"/><path d="m8 21-4-4 4-4"/><path d="M4 17h16"/>',
    'sale': '<path d="M3 7h11v9H3zM14 10h4l3 3v3h-7"/><circle cx="7" cy="18" r="2"/><circle cx="17" cy="18" r="2"/>',
    'pencil': '<path d="M21.2 6.8a2.8 2.8 0 0 0-4-4L3.8 16.2 3 21l4.8-.8z"/><path d="m15 5 4 4"/>',
    'trash': '<path d="M3 6h18"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6"/><path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>',
}
OPEN_D, CLOSED_D = 'm6 9 6 6 6-6', 'm9 18 6-6-6-6'


class Out:
    """mode 'dc' (board: {{ holes }}, onClick) or 'html' (preview: initial values, data-b / data-mo for the JS)"""
    def __init__(self, mode): self.mode = mode
    def b(self, pairs):
        """bound attributes: [(attr, key, initial)]; initial None/False = absent, True = boolean attribute"""
        if self.mode == 'dc':
            return ''.join(f' {a}="{{{{ {k} }}}}"' for a, k, _ in pairs)
        out = ''
        for a, k, v in pairs:
            if v is True: out += f' {a}=""'
            elif v not in (None, False): out += f' {a}="{v}"'
        return out + ' data-b="' + ';'.join(f'{a}:{k}' for a, k, _ in pairs) + '"'
    def on(self, key, ev='onClick'):
        return f' {ev}="{{{{ on.{key} }}}}"' if self.mode == 'dc' else f' data-mo="{key}"'
    def t(self, key, init):
        """bound text: a hole on the board, a span the preview's JS rewrites"""
        return f'{{{{ {key} }}}}' if self.mode == 'dc' else f'<span data-t="{key}">{init}</span>'


# ---------- PeriodTable: SAPT 59–56 loaded, SAPT 55–52 appended by the load row (D190–D192) ----------
W8 = [59 - i for i in range(8)]
UID = 'mo'
CAP4 = 'Păsări · Hîrbovăț · Seria 2026 · SAPT 56–59'
t8, TID = P.ptable(UID, CAP4, [('Total fermă', '', P.farm_pasari())], weeks=W8, colw=P.COLW, more=P.more())
t4, _ = P.ptable(UID, CAP4, [('Total fermă', '', P.farm_pasari())], weeks=W8[:4], colw=P.COLW, more=P.more())
WEEK_RE = re.compile(r'<tbody class="ip-ptable__week">.*?</tbody>', re.S)
weeks8 = WEEK_RE.findall(t8); assert len(weeks8) == 8
head8 = t8[:t8.index('<tbody class="ip-ptable__week">')]
assert head8 == t4[:t4.index('<tbody class="ip-ptable__week">')]
more4 = re.search(r'<tbody class="ip-ptable__morebody">.*?</tbody>', t4, re.S).group(0)
foot4 = re.search(r'<tfoot>.*?</tfoot>', t4, re.S).group(0)
foot8 = re.search(r'<tfoot>.*?</tfoot>', t8, re.S).group(0)


def ptable(o):
    body = ''
    for k, w in zip(W8, weeks8):
        new = k < 56
        w = w.replace('<tbody class="ip-ptable__week">',
                      f'<tbody{o.b([("class", f"w{k}c", f"ip-ptable__week is-new mo-i{55 - k}" if new else "ip-ptable__week"), ("data-state", f"w{k}s", None)] + ([("hidden", f"w{k}h", True)] if new else []))}>')
        w = w.replace('<tr class="ip-ptable__day"', f'<tr class="ip-ptable__day"{o.b([("hidden", f"w{k}d", False)])}')
        w = re.sub(r'(<button type="button" class="ip-ptable__key)( ip-focus)?(" )aria-expanded="true"',
                   lambda m: f'{m.group(1)}{m.group(3)}{o.b([("aria-expanded", f"w{k}e", "true")])}{o.on(f"wk{k}")}', w)
        w = w.replace(f'<path d="{OPEN_D}"/>', f'<path{o.b([("d", f"w{k}v", OPEN_D)])}/>', 1)
        body += w
    m = more4.replace('<tbody class="ip-ptable__morebody">', f'<tbody class="ip-ptable__morebody">')
    m = m.replace('<button type="button" class="ip-ptable__load"', f'<button type="button" class="ip-ptable__load" id="mo-load"{o.on("more")}')
    m = m.replace('<span class="ip-num">56–59</span>', f'<span class="ip-num">{o.t("moreSpan", "56–59")}</span>')
    m = m.replace('<span class="ip-num">4</span> din', f'<span class="ip-num">{o.t("moreN", "4")}</span> din')
    m = m.replace('aria-live="polite"></span>', f'aria-live="polite">{o.t("moreMsg", "")}</span>')
    assert o.mode == 'html' or '{{ moreSpan }}' in m and '{{ moreN }}' in m and '{{ moreMsg }}' in m, m
    f4 = foot4.replace('<tfoot>', f'<tfoot{o.b([("hidden", "f4h", False)])}>')
    f8 = foot8.replace('<tfoot>', f'<tfoot{o.b([("hidden", "f8h", True)])}>')
    cap = f'<p class="ip-cap ip-ptable__cap" aria-hidden="true">Păsări · Hîrbovăț · Seria 2026 · SAPT {o.t("capSpan", "56–59")}</p>'
    return cap + P.region(head8 + body + m + f4 + f8 + '</table>', 'Păsări · Hîrbovăț · Seria 2026')


# ---------- the demos ----------
def card(o, n, title, spec, keys, stage, wide=False):
    return (f'<section class="ip-panel mo-card{" mo-card--wide" if wide else ""}" aria-labelledby="mo-h{n}">'
            f'<div class="mo-card__head"><h2 class="ip-t-heading ip-m0" id="mo-h{n}">{title}</h2>'
            f'<p class="ip-note ip-m0">{spec}</p></div>'
            + (f'<div class="ip-row mo-keys">{keys}</div>' if keys else '') + stage + '</section>')


def btn(o, label, key, var='secondary', extra=''):
    return f'<button type="button" class="ip-btn ip-btn--{var}"{o.on(key)}{extra}>{label}</button>'


def demos(o):
    out = []
    # 1 keys
    keys = (f'<button type="button" class="ip-btn ip-btn--primary">Salvează</button>'
            f'<button type="button" class="ip-btn ip-btn--secondary">Anulează</button>'
            f'<button type="button" class="ip-btn ip-btn--ghost">{svg(I["pencil"], 16)}Editează</button>'
            f'<button type="button" class="ip-btn ip-btn--destructive">{svg(I["trash"], 16)}Șterge</button>'
            f'<button type="button" class="ip-iconbtn" aria-label="Închide">{svg(I["x"])}</button>'
            f'<button type="button" class="ip-btn ip-btn--primary" disabled="disabled">Salvează</button>')
    out.append(card(o, 1, 'Taste · apăsare',
                    'Feedback, de zeci de ori pe zi · tranziție CSS · transform scale 1 → 0,97 la :active · --ip-ease · 120 ms · tasta dezactivată nu se mișcă',
                    '', f'<div class="ip-row mo-stage mo-stage--flat">{keys}</div>'))
    # 2 toggles: tabs, segment, chips, checkbox (colour only)
    tabs = ''.join(f'<button type="button" role="tab" class="ip-tab"{o.b([("class", f"tab{i}", "ip-tab is-selected" if i == 0 else "ip-tab"), ("aria-selected", f"tab{i}a", "true" if i == 0 else "false")])}{o.on(f"tab{i}")}>{t}</button>'
                   for i, t in enumerate(['Rezumat', 'Păsări', 'Hrană', 'Greutate', 'Producție']))
    seg = ''.join(f'<span role="radio" tabindex="0"{o.b([("class", f"seg{i}", "ip-seg__item is-selected" if i == 0 else "ip-seg__item"), ("aria-checked", f"seg{i}a", "true" if i == 0 else "false")])}{o.on(f"seg{i}")}>{t}</span>'
                  for i, t in enumerate(['Zi', 'Săpt.']))
    chips = ''.join(f'<button type="button" role="checkbox"{o.b([("class", f"chip{i}", f"ip-chip ip-rec ip-rec--{r}" + ("" if i == 0 else " ip-chip--off")), ("aria-checked", f"chip{i}a", "true" if i == 0 else "false")])}{o.on(f"chip{i}")}>{svg(I[r], 16)}{t}</button>'
                    for i, (r, t) in enumerate([('mortality', 'Mortalitate'), ('sale', 'Vânzare'), ('transfer', 'Transfer')]))
    chk = (f'<label class="ip-check"><input type="checkbox" class="ip-sr" id="mo-chk"{o.b([("checked", "chk", False)])}{o.on("chk", "onChange")}/>'
           f'<span class="ip-check__box">{svg(I["check"], 14)}</span>Fără pierderi azi</label>')
    out.append(card(o, 2, 'Comutatoare · doar culoare',
                    'Starea aleasă, de zeci de ori pe zi (și din tastatură) · fără alunecare: fundal, margine și culoare în 120 ms · Tabs, SegmentedControl, RecordChip, Checkbox',
                    '', f'<div class="ip-stack mo-stage mo-stage--flat"><div class="ip-tabs" role="tablist" aria-label="Fermă">{tabs}</div>'
                        f'<div class="ip-row"><span class="ip-cap" id="mo-pas">Pas</span><span class="ip-seg" role="radiogroup" aria-labelledby="mo-pas">{seg}</span></div>'
                        f'<div class="ip-chips" role="group" aria-label="Etichete">{chips}</div>{chk}</div>'))
    # 3 dialog
    dlg = (f'<div class="ip-scrim ip-scrim--page"{o.b([("data-state", "dlg", "closed")])}>'
           f'<div class="ip-dialog" role="alertdialog" aria-modal="true" aria-labelledby="mo-dt" aria-describedby="mo-dn">'
           f'<h2 id="mo-dt">Retrage secțiunea H5-2 de la Telița</h2><p class="ip-dialog__note" id="mo-dn">H5-2 nu mai apare în formulare și pe pagina fermei. '
           f'Înregistrările ei din Seria 2026 rămân în rapoarte.</p><div class="ip-dialog__keys">{btn(o, "Anulează", "dlgClose")}'
           f'{btn(o, "Retrage secțiunea", "dlgClose", "destructive")}</div></div></div>')
    page = '<div class="mo-page"><p class="ip-cap ip-m0">FERME ȘI SECȚIUNI · TELIȚA</p><p class="ip-t-body ip-m0">H5-1 · H5-2 · H6-1 · H6-2</p></div>'
    out.append(card(o, 3, 'Dialog',
                    'Continuitate spațială, ocazional · @starting-style + data-state · fundalul: opacity; dialogul: scale 0,96 → 1, centrat · 200 ms, iese pe același drum',
                    btn(o, 'Retrage secțiunea', 'dlgOpen', 'destructive'), f'<div class="mo-stage mo-frame">{page}{dlg}</div>'))
    # 4 bottom sheet (phone)
    items = ''.join(f'<button type="button" class="ip-ev"{o.on("sheetClose")}><span class="ip-tile ip-rec ip-rec--{r}" aria-hidden="true">{svg(I[r])}</span>'
                    f'<span class="ip-ev__main">{t}<span class="ip-ev__meta">{m}</span></span></button>'
                    for r, t, m in [('mortality', 'Zilnic', 'Mortalitate, sacrificare, ouă'), ('feed', 'Hrană zilnică', 'Grame pe găină și cocoș'),
                                    ('weighing', 'Cântărire', 'Greutatea pe secțiune'), ('transfer', 'Transfer', 'Între secțiuni sau ferme')])
    sheet = (f'<div class="ip-scrim ip-scrim--sheet"{o.b([("data-state", "sheet", "closed")])}>'
             f'<div class="ip-sheet ip-sheet--select" role="dialog" aria-modal="true" aria-labelledby="mo-st"><div class="ip-sheet__handle"></div>'
             f'<div class="ip-sheet__head"><h2 class="ip-sheet__title ip-t-heading" id="mo-st">Înregistrează</h2>'
             f'<button type="button" class="ip-iconbtn ip-iconbtn--touch" aria-label="Închide lista"{o.on("sheetClose")}>{svg(I["x"])}</button></div>'
             f'<div class="ip-list mo-sheetlist">{items}</div></div></div>')
    phone = ('<div class="mo-phone ip-scrimhost"><div class="mo-page"><p class="ip-cap ip-m0">ACASĂ · HÎRBOVĂȚ</p>'
             '<p class="ip-t-body ip-m0">18.11.2026 · SAPT 59</p></div>' + sheet + '</div>')
    out.append(card(o, 4, 'BottomSheet · telefon',
                    'Continuitate spațială, ocazional · urcă cât propria înălțime (translateY 100% → 0) cu fundalul, 280 ms (traversează ecranul), coboară la fel',
                    btn(o, 'Înregistrează', 'sheetOpen', 'primary'), f'<div class="mo-stage mo-stage--phone">{phone}</div>'))
    # 5 docked side panel
    panel = (f'<aside class="ip-sidepanel ip-sidepanel--dock" aria-labelledby="mo-pt"{o.b([("data-state", "panel", "closed")])}>'
             f'<div class="ip-sidepanel__head"><div class="ip-row ip-sidepanel__title"><span class="ip-tile ip-rec ip-rec--mortality" aria-hidden="true">{svg(I["mortality"])}</span>'
             f'<div><h2 class="ip-t-heading ip-m0" id="mo-pt">Corectează Mortalitate</h2><p class="ip-note ip-num ip-m0">HÎRBOVĂȚ · 17.11.2026 · SAPT 59</p></div></div>'
             f'<button type="button" class="ip-iconbtn" aria-label="Închide Corectează"{o.on("panelClose")}>{svg(I["x"])}</button></div>'
             f'<div class="ip-sidepanel__body"><label class="ip-field"><span class="ip-label">Motiv</span><span class="ip-input"><input type="text" placeholder="De ce corectezi"/></span></label></div>'
             f'<div class="ip-sidepanel__foot"><span></span><div class="ip-row ip-sidepanel__keys">{btn(o, "Anulează", "panelClose")}{btn(o, "Salvează corecția", "panelClose", "primary")}</div></div></aside>')
    out.append(card(o, 5, 'SidePanel · lângă pagină',
                    'Continuitate spațială, ocazional · intră din marginea dreaptă (translateX 100% → 0, opacity) · 200 ms · pagina își păstrează marginea, fără animație de lățime',
                    btn(o, 'Corectează', 'panelOpen'), f'<div class="mo-stage mo-frame mo-frame--panel">{page.replace("FERME ȘI SECȚIUNI · TELIȚA", "INTRODUS AZI · HÎRBOVĂȚ").replace("H5-1 · H5-2 · H6-1 · H6-2", "Mortalitate · 17.11.2026 · H5.1")}{panel}</div>'))
    # 6 toast
    toast = (f'<div class="ip-toastwrap ip-toastwrap--page"><div class="ip-toast" role="status"{o.b([("data-state", "toast", "closed")])}>{svg(I["check"], 16)}'
             f'<span class="ip-toast__txt">Livrare adăugată · <span class="ip-num">25 640</span> kg · <span class="ip-num">13.11.2026</span></span>'
             f'<button type="button" class="ip-link ip-toast__undo"{o.on("toastClose")}>Anulează</button></div></div>')
    out.append(card(o, 6, 'Toast',
                    'Feedback, ocazional · tranziție (nu keyframes: o a doua salvare o întoarce din drum) · translateY 100% → 0 + opacity · 200 ms · pleacă pe jos după 4 s',
                    btn(o, 'Salvează livrarea', 'toastOpen', 'primary'), f'<div class="mo-stage mo-frame">{toast}</div>'))
    # 7 select
    opts = ''.join(f'<li class="ip-menu__item ip-select__opt" role="option"{o.b([("aria-selected", f"opt{i}", "true" if i == 1 else "false")])}{o.on(f"opt{i}")}>'
                   f'<span class="ip-select__chk" aria-hidden="true"{o.b([("data-chk", f"opt{i}k", "on" if i == 1 else None)])}>{svg(I["check"], 16)}</span><span class="ip-select__txt">{t}</span></li>'
                   for i, t in enumerate(['Breeder 1', 'Breeder 2', 'Cocoși ouat']))
    sel = (f'<div class="ip-field mo-sel"><span class="ip-label" id="mo-sl">Tip hrană</span><div class="ip-select">'
           f'<div class="ip-input ip-select__btn" role="combobox" tabindex="0" aria-haspopup="listbox" aria-controls="mo-sm" aria-labelledby="mo-sl"'
           f'{o.b([("aria-expanded", "menuE", "false")])}{o.on("menuToggle")}><span class="ip-select__val">{o.t("selVal", "Breeder 2")}</span>{svg(I["chev"], 16)}</div>'
           f'<ul class="ip-menu ip-select__menu" role="listbox" id="mo-sm" aria-labelledby="mo-sl"{o.b([("data-state", "menu", "closed")])}>{opts}</ul></div></div>')
    out.append(card(o, 7, 'Select · listă',
                    'Continuitate spațială, de zeci de ori pe zi · scale 0,95 → 1 + opacity, din marginea de sus a câmpului (transform-origin la declanșator) · 120 ms',
                    '', f'<div class="mo-stage mo-stage--flat mo-stage--menu">{sel}</div>'))
    # 8 banner
    ban = (f'<div class="ip-banner ip-banner--fault ip-banner--act" role="alert"{o.b([("data-state", "banner", "closed")])}>{svg(I["fault"], 16)}'
           f'<div class="ip-banner__body"><p class="ip-m0">Tipurile de hrană nu s-au încărcat.</p><p class="ip-m0">Verifică conexiunea la internet, apoi apasă Reîncearcă.</p></div>'
           f'<button type="button" class="ip-btn ip-btn--secondary ip-banner__key"{o.on("bannerClose")}>Reîncearcă</button></div>')
    out.append(card(o, 8, 'Banner',
                    'Evită o schimbare bruscă, ocazional · opacity + translateY −8 px → 0 · 200 ms · conținutul de sub el se mută imediat (layout-ul nu se animă)',
                    btn(o, 'Arată eroarea', 'bannerOpen'), f'<div class="ip-stack mo-stage mo-stage--flat">{ban}<p class="ip-t-body ip-m0">Tipuri de hrană · Breeder 1 · Breeder 2 · Cocoși ouat</p></div>'))
    # 9 skeleton -> content
    skel = ('<div class="ip-readout ip-readout--skel" aria-hidden="true"' + o.b([("hidden", "skelH", True)]) + '><div class="ip-readout__cap"><span><span class="ip-skel ip-skel--w2"></span></span></div>'
            '<div class="ip-readout__val"><span class="ip-skel ip-skel--w3"></span><small><span class="ip-skel"></span></small></div>'
            '<div class="ip-readout__sub"><span class="ip-skel ip-skel--w1"></span></div></div>')
    ro = ('<div class="ip-readout" data-state="loaded"' + o.b([("hidden", "roH", False)]) + '><div class="ip-readout__cap"><span>OUAT · HÎRBOVĂȚ · SAPT 59</span><span class="ip-rlamp ip-rlamp--warn">ABATERE</span></div>'
          '<div class="ip-readout__val">40,4<small>%</small></div><div class="ip-readout__sub">ghid 57,6% · −17,2 pct · 17.11–18.11.2026</div></div>')
    out.append(card(o, 9, 'Skeleton → conținut',
                    'Evită o schimbare bruscă, la fiecare încărcare · scheletul (aceeași geometrie, fără shimmer) pleacă imediat, blocul încărcat apare cu opacity în 120 ms · cifrele nu numără',
                    btn(o, 'Încarcă din nou', 'reload'), f'<div class="mo-stage mo-stage--flat">{skel}{ro}</div>'))
    # 10 PeriodTable
    out.append(card(o, 10, 'PeriodTable · săptămâni',
                    'Închide / deschide o săptămână (tasta SAPT): rândurile zilelor își iau locul imediat și apar cu opacity în 120 ms, fără animație de înălțime · '
                    '„Încarcă încă 4 săptămâni”: fiecare săptămână nouă urcă 8 px și apare în 200 ms, la 40 ms una după alta; rândul păstrează focusul, nimic nu așteaptă animația · '
                    'a doua apăsare revine la 4 săptămâni (doar în demonstrație)',
                    '', f'<div class="mo-stage mo-stage--flat">{ptable(o)}</div>', wide=True))
    return ''.join(out)


CSS = """
.mo-wrap{box-sizing:border-box;max-width:1440px;margin:0 auto;padding:var(--space-8);display:flex;flex-direction:column;gap:var(--space-6)}
.mo-head{display:flex;flex-direction:column;gap:var(--space-1)}
.mo-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:var(--space-6);align-items:start}
.mo-card{display:flex;flex-direction:column;gap:var(--space-3);min-width:0}
.mo-card--wide{grid-column:1 / -1}
.mo-card__head{display:flex;flex-direction:column;gap:var(--space-1)}
.mo-keys{gap:var(--space-2)}
.mo-stage{min-width:0}
.mo-stage--flat{gap:var(--space-4);flex-wrap:wrap}
.mo-stage--menu{min-height:200px}
.mo-sel{width:240px}
.mo-frame{position:relative;height:280px;overflow:hidden;border:var(--ip-hair) solid var(--line);border-radius:var(--radius-md);background:var(--ground)}
.mo-frame--panel{height:300px}
.mo-page{padding:var(--space-4);display:flex;flex-direction:column;gap:var(--space-1)}
.mo-stage--phone{display:flex}
.mo-phone{width:360px;height:420px;overflow:hidden;border:var(--ip-hair) solid var(--line);border-radius:var(--radius-md);background:var(--ground)}
.mo-sheetlist{border-radius:0;border-inline:0;border-bottom:0}
.mo-i1{--ip-i:1}.mo-i2{--ip-i:2}.mo-i3{--ip-i:3}
.ip-select__chk:not([data-chk]) > .ip-icon{visibility:hidden}
.mo-stage > .ip-readout[hidden]{display:none} /* bundle.css .ip-readout sets display over the UA [hidden] */
"""

# One JS core for both: the board's DCLogic and the preview's script. vals() returns every bound value by key.
CORE = r"""
  moInit() {
    return { reduce: false, dlg: 'closed', sheet: 'closed', panel: 'closed', toast: 'closed', menu: 'closed',
      banner: 'closed', loading: false, tab: 0, seg: 0, chips: [true, false, false], chk: false, sel: 1,
      weeks: { 59: '', 58: '', 57: '', 56: '', 55: '', 54: '', 53: '', 52: '' }, more: 0 };
  }
  moAct(s, k) {
    const wk = /^wk(\d+)$/.exec(k);
    if (wk) { const w = Object.assign({}, s.weeks); w[wk[1]] = (w[wk[1]] === 'closed') ? 'open' : 'closed'; return { weeks: w }; }
    let m;
    if ((m = /^tab(\d)$/.exec(k))) return { tab: +m[1] };
    if ((m = /^seg(\d)$/.exec(k))) return { seg: +m[1] };
    if ((m = /^chip(\d)$/.exec(k))) { const c = s.chips.slice(); c[+m[1]] = !c[+m[1]]; return { chips: c }; }
    if ((m = /^opt(\d)$/.exec(k))) return { sel: +m[1], menu: 'closed' };
    switch (k) {
      case 'reduce': return { reduce: !s.reduce };
      case 'chk': return { chk: !s.chk };
      case 'dlgOpen': return { dlg: 'open' };
      case 'dlgClose': return { dlg: 'closed' };
      case 'sheetOpen': return { sheet: 'open' };
      case 'sheetClose': return { sheet: 'closed' };
      case 'panelOpen': return { panel: s.panel === 'open' ? 'closed' : 'open' };
      case 'panelClose': return { panel: 'closed' };
      case 'toastOpen': return { toast: 'open', __after: [4000, 'toastClose'] };
      case 'toastClose': return { toast: 'closed' };
      case 'menuToggle': return { menu: s.menu === 'open' ? 'closed' : 'open' };
      case 'bannerOpen': return { banner: s.banner === 'open' ? 'closed' : 'open' };
      case 'bannerClose': return { banner: 'closed' };
      case 'reload': return { loading: true, __after: [900, 'loaded'] };
      case 'loaded': return { loading: false };
      case 'more': return { more: s.more ? 0 : 1 };
    }
    return {};
  }
  moVals(s) {
    const v = { rootCls: s.reduce ? 'mo-root ip-motion--reduce' : 'mo-root', reduce: s.reduce,
      dlg: s.dlg, sheet: s.sheet, panel: s.panel, toast: s.toast, menu: s.menu, banner: s.banner,
      menuE: s.menu === 'open' ? 'true' : 'false', chk: s.chk,
      selVal: ['Breeder 1', 'Breeder 2', 'Cocoși ouat'][s.sel],
      skelH: !s.loading, roH: s.loading,
      moreSpan: s.more ? '52–59' : '56–59', moreN: s.more ? '8' : '4', capSpan: s.more ? '52–59' : '56–59',
      moreMsg: s.more ? 'Au fost adăugate SAPT 52–55.' : '', f4h: !!s.more, f8h: !s.more };
    for (let i = 0; i < 5; i++) { v['tab' + i] = 'ip-tab' + (s.tab === i ? ' is-selected' : ''); v['tab' + i + 'a'] = s.tab === i ? 'true' : 'false'; }
    for (let i = 0; i < 2; i++) { v['seg' + i] = 'ip-seg__item' + (s.seg === i ? ' is-selected' : ''); v['seg' + i + 'a'] = s.seg === i ? 'true' : 'false'; }
    const rec = ['mortality', 'sale', 'transfer'];
    for (let i = 0; i < 3; i++) { v['chip' + i] = 'ip-chip ip-rec ip-rec--' + rec[i] + (s.chips[i] ? '' : ' ip-chip--off'); v['chip' + i + 'a'] = s.chips[i] ? 'true' : 'false'; }
    for (let i = 0; i < 3; i++) { v['opt' + i] = s.sel === i ? 'true' : 'false'; v['opt' + i + 'k'] = s.sel === i ? 'on' : undefined; }
    for (let k = 52; k <= 59; k++) {
      const st = s.weeks[k] || undefined, closed = st === 'closed', isNew = k < 56;
      v['w' + k + 'c'] = 'ip-ptable__week' + (isNew ? ' is-new mo-i' + (55 - k) : '');
      v['w' + k + 's'] = st; v['w' + k + 'h'] = isNew && !s.more; v['w' + k + 'd'] = closed;
      v['w' + k + 'e'] = closed ? 'false' : 'true'; v['w' + k + 'v'] = closed ? '""" + CLOSED_D + r"""' : '""" + OPEN_D + r"""';
    }
    return v;
  }
"""
ACTS = (['reduce', 'chk', 'dlgOpen', 'dlgClose', 'sheetOpen', 'sheetClose', 'panelOpen', 'panelClose', 'toastOpen', 'toastClose',
         'menuToggle', 'bannerOpen', 'bannerClose', 'reload', 'loaded', 'more'] + [f'tab{i}' for i in range(5)] + [f'seg{i}' for i in range(2)]
        + [f'chip{i}' for i in range(3)] + [f'opt{i}' for i in range(3)] + [f'wk{k}' for k in W8])


def header(o):
    return ('<div class="mo-head"><span class="ip-cap">SISTEM · MIȘCARE · D196</span><h1 class="ip-t-title ip-m0">Mișcare</h1>'
            '<p class="ip-t-body ip-m0" style="color:var(--ink-secondary)">O singură curbă (cubic-bezier(0.2, 0, 0, 1)) și trei durate: 120 ms taste, liste și rânduri · '
            '200 ms dialog, panou, toast, banner · 280 ms foi pe telefon. Doar transform și opacity. Cifrele, graficele, lămpile, CycleTrack și '
            'navigarea nu se mișcă. Apasă tastele fiecărui exemplu.</p>'
            f'<label class="ip-check"><input type="checkbox" class="ip-sr" id="mo-rm"{o.b([("checked", "reduce", False)])}{o.on("reduce", "onChange")}/>'
            f'<span class="ip-check__box">{svg(I["check"], 14)}</span>Simulează mișcare redusă (prefers-reduced-motion): doar opacitate, aceeași durată, nimic nu se deplasează</label></div>')


def page(o):
    return f'<div class="mo-wrap">{header(o)}<div class="mo-grid">{demos(o)}</div></div>'


FONTS = ('<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500&amp;'
         'family=IBM+Plex+Mono:wght@400;500&amp;display=swap">')

PREVIEW_JS = r"""
class MoCore {""" + CORE + r"""}
const M = new MoCore();
let S = M.moInit();
function apply() {
  const v = M.moVals(S);
  document.querySelectorAll('[data-b]').forEach(el => {
    el.dataset.b.split(';').forEach(p => {
      const [a, k] = p.split(':'); const x = v[k];
      if (a === 'checked') { el.checked = !!x; return; }
      if (x === undefined || x === null || x === false) el.removeAttribute(a);
      else el.setAttribute(a, x === true ? '' : String(x));
    });
  });
  document.querySelectorAll('[data-t]').forEach(el => { el.textContent = v[el.dataset.t]; });
  document.getElementById('mo-root').className = 'ip-root ' + v.rootCls;
}
function fire(k) {
  const p = M.moAct(S, k); const after = p.__after; delete p.__after;
  S = Object.assign({}, S, p); apply();
  if (after) setTimeout(() => fire(after[1]), after[0]);
}
document.addEventListener('click', e => {
  const t = e.target.closest('[data-mo]'); if (!t || t.tagName === 'INPUT') return; fire(t.dataset.mo);
});
document.addEventListener('change', e => { const t = e.target.closest('input[data-mo]'); if (t) fire(t.dataset.mo); });
window.moFire = fire;
apply();
"""


def preview():
    o = Out('html')
    body = SXS.sexify(page(o))
    return f'''<!doctype html>
<html lang="ro">
<head>
<meta charset="utf-8">
<title>Mișcare · Impex Line Panel (additions §75)</title>
<!-- Project preview for components/additions.css §75 (MO01, D196): every motion of the system, triggered by its keys. Not on the system card list; on the canvas as M1 · Mișcare. Built by design/system/scripts/mo01_build.py. -->
<link rel="stylesheet" href="../../tokens.css">
<link rel="stylesheet" href="../bundle.css">
<link rel="stylesheet" href="../additions.css">
{FONTS}
<style>body{{margin:0}}{CSS}</style>
</head>
<body>
<div class="ip-root mo-root" id="mo-root">{body}</div>
<script>{PREVIEW_JS}</script>
</body>
</html>
'''


def board():
    o = Out('dc')
    body = SXS.sexify(page(o))
    on = ', '.join(f'{k}: () => this.moFire(\'{k}\')' for k in ACTS)
    js = ('class Component extends DCLogic {\nconstructor(props) {\nsuper(props);\nthis.state = this.moInit();\n}\n'
          'moFire(k) {\nconst p = this.moAct(this.state, k);\nconst after = p.__after;\ndelete p.__after;\nthis.setState(p);\n'
          'if (after) setTimeout(() => this.moFire(after[1]), after[0]);\n}\n' + CORE +
          'renderVals() {\nconst v = this.moVals(this.state);\nv.on = { ' + on + ' };\nreturn v;\n}\n}\n')
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
{FONTS}
<style>
body{{margin:0}}{CSS}</style>
</helmet>
<div class="ip-root {{{{ rootCls }}}}" style="padding:0;min-height:{H}px">{body}</div>
</x-dc>
<script type="text/x-dc" data-dc-script data-props='{{"$preview":{{"width":{W},"height":{H}}}}}'>
{js}</script>
</body>
</html>
'''


if __name__ == '__main__':
    out = REPO / 'design/system/components/Motion/preview.html'
    out.parent.mkdir(exist_ok=True)
    pv = preview()
    if not out.exists() or out.read_text(encoding='utf-8') != pv: out.write_text(pv, encoding='utf-8')
    bd = board()
    bo = REPO / f'design/canvas/project/{STEM}.dc.html'
    if not bo.exists() or bo.read_text(encoding='utf-8') != bd: bo.write_text(bd, encoding='utf-8')
    if CANVAS:
        import canvas_rows as CR
        c = CR.load(REPO); B = c['boards']; f = f'{STEM}.dc.html'
        if f in B: B[f].update(w=W, h=H, title=TITLE, is_interactive=True)
        else:
            # its own row right under Main (y = ROW0): every screen row moves down by the board's height + 120 (relayout re-stacks)
            for k, v in B.items():
                if v['y'] >= CR.ROW0: v['y'] += H + CR.GAP_Y
            B[f] = {'x': 0, 'y': CR.ROW0, 'w': W, 'h': H, 'expand': 'fill', 'title': TITLE, 'is_interactive': True}
            c['order'].append(f)
        CR.sort_order(c); CR.save(REPO, c)
    print('ok', len(pv), len(bd))
