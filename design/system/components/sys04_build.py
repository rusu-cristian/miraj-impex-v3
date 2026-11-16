"""SYS04: build StockReadout and CostBlock preview.html fragments. Figures from design/data/extracts (see READMEs)."""
import pathlib
ROOT = pathlib.Path('/home/claude/miraj-impex-v3/design/system/components')
N = ' '   # thousands
M = '−'   # true minus
def g(s): return s.replace(' ', N)  # write grouped numbers with a plain space in source, emit U+202F
COINS = '<svg class="ip-icon ip-icon--16" viewBox="0 0 24 24" aria-hidden="true"><circle cx="8" cy="8" r="6"/><path d="M18.09 10.37A6 6 0 1 1 10.34 18"/><path d="M7 6h1v4"/><path d="m16.71 13.88.7.71-2.82 2.82"/></svg>'

def readout(id_, cap, lamp, word, val, unit, kv, sub=None, off=False, sr=''):
    cls = 'ip-readout ip-readout--stock' + (' ip-readout--off' if off else '')
    u = f'<small> {unit}</small>' if unit else ''
    rows = ''.join(f'<div><dt>{k}</dt><dd>{v}</dd></div>' for k, v in kv)
    kvh = f'<dl class="ip-readout__kv">{rows}</dl>' if kv else ''
    subh = ''.join(f'<p class="ip-readout__sub" style="margin:0">{s}</p>' for s in (sub or []))
    return (f'<li><div class="{cls}" role="group" aria-labelledby="{id_}">'
            f'<div class="ip-readout__cap"><span id="{id_}">{cap}{f'<span class="ip-sr">{sr}</span>' if sr else ''}</span><span class="ip-rlamp ip-rlamp--{lamp}">{word}</span></div>'
            f'<div class="ip-readout__val">{val}{u}</div>{kvh}{subh}</div></li>')

def stock_row(sfx):
    return ('<ul class="ip-readouts" aria-label="Stoc pe fermă și tip de hrană">'
        + readout(f'sr-b2-{sfx}', 'Hîrbovăț · Breeder\u00a02', 'ok', 'în normă', g('28 252'), 'kg',
                  [('zile rămase', '7,1<small> zile</small>'), ('medie/zi 13–15.11', g('3 976') + '<small> kg</small>'), ('se termină', '23.11.2026')],
                  sr=', pentru găini')
        + readout(f'sr-co-{sfx}', 'Hîrbovăț · Cocoși', 'warn', 'abatere', g('1 650'), 'kg',
                  [('zile rămase', '3,9<small> zile</small>'), ('medie/zi 13–15.11', '419<small> kg</small>'), ('se termină', '19.11.2026')],
                  sub=['sub pragul de 5 zile'], sr=', pentru cocoși')
        + readout(f'sr-te-{sfx}', 'Telița', 'off', 'oprit', '–', '', [],
                  sub=['fără ciclu deschis', 'ultimul ciclu: Seria 2026, închis pe <span class="ip-num" style="white-space:nowrap">28.02.2026</span>'], off=True)
        + '</ul>')

def state_row(sfx):
    return ('<ul class="ip-readouts" aria-label="Exemple de stare">'
        + readout(f'sr-neg-{sfx}', 'Hîrbovăț · Breeder\u00a02', 'warn', 'abatere', M + g('1 240'), 'kg',
                  [('zile rămase', '0,0<small> zile</small>'), ('medie/zi 13–15.11', g('3 976') + '<small> kg</small>'), ('epuizat pe', '16.11.2026')],
                  sub=['stoc sub 0 · verifică livrările', 'consumul depășește livrările'], sr=', pentru găini')
        + readout(f'sr-idle-{sfx}', 'Fermă exemplu · Starter', 'off', 'oprit', g('8 000'), 'kg',
                  [('zile rămase', '–'), ('medie/zi 13–15.11', '0<small> kg</small>'), ('se termină', '–')],
                  sub=['– fără consum în ultimele 3 zile'])
        + '</ul>')

FOOT = ('<p class="ip-note" style="margin:0">Stoc la <span class="ip-num">16.11.2026</span> = livrări − consum (rânduri confirmate). '
        'Zile rămase = stoc ÷ consumul mediu din ultimele 3 zile încheiate (<span class="ip-num">13–15.11.2026</span>); '
        'sub 5 zile = abatere.</p>')

def cost_block(sfx, extra=False, wide=True):
    hdr = ('<thead><tr><th scope="col">Ciclu</th><th scope="col" class="num">Kg</th><th scope="col" class="num">Lei</th>'
           '<th scope="col" class="num">Lei/pasăre</th><th scope="col" class="num">Lei/ou</th></tr></thead>')
    def unpriced(cyc, kg, note, tag=True):
        t = '<span class="ip-tag ip-tag--info">fără preț</span>' if tag else '<span>fără preț</span>'
        return (f'<tr class="ip-cost__unpriced"><th scope="row"><span class="ip-sr">{cyc}, </span><span><span class="ip-rec ip-rec--feed">{COINS}</span>{t}</span></th>'
                f'<td class="num">{kg}</td><td class="ip-cost__note" colspan="3">{note}</td></tr>')
    if extra:
        body = ('<tbody><tr><th scope="row">Fermă exemplu · Seria 2027<span class="ip-cost__sub">ciclu nou · 0 livrări cu preț</span></th>'
                '<td class="num">–</td><td class="num">–<small>nicio livrare cu preț</small></td><td class="num">–</td><td class="num">–</td></tr>'
                + unpriced('Fermă exemplu · Seria 2027', g('8 000'), '1 livrare · neinclusă în cost') + '</tbody>')
        title, note = 'Cost hrană · exemplu de stare', 'nicio livrare cu preț'
    else:
        body = ('<tbody><tr><th scope="row">Telița · Seria 2026<span class="ip-cost__sub"><span class="ip-num">07.10.2025</span> – <span class="ip-num">28.02.2026</span> · 13 livrări cu preț</span></th>'
                f'<td class="num">{g("263 740")}</td><td class="num">{g("2 031 768")}</td><td class="num">64,32</td><td class="num">–<small>creștere, fără ouă</small></td></tr>'
                '<tr class="ip-cost__unpriced"><th scope="row"><span class="ip-sr">Telița · Seria 2026, </span><span><span class="ip-rec ip-rec--feed">' + COINS + '</span><span>fără preț</span></span></th>'
                '<td class="num">0</td><td class="ip-cost__note" colspan="3">Toate livrările au preț.</td></tr></tbody>'
                '<tbody><tr><th scope="row">Hîrbovăț · Seria 2026<span class="ip-cost__sub">din <span class="ip-num">10.02.2026</span> · <span class="ip-num">SAPT 40</span> · 59 livrări cu preț</span></th>'
                f'<td class="num">{g("1 143 540")}</td><td class="num">{g("10 417 057")}</td><td class="num">349,27</td><td class="num">1,97</td></tr>'
                + unpriced('Hîrbovăț · Seria 2026', g('76 107'), '4 livrări, <span class="ip-num">04.11–14.11.2026</span> · neincluse în cost') + '</tbody>'
                '<tbody><tr class="is-total"><th scope="row">Total Seria 2026<span class="ip-cost__sub">nu include <span class="ip-num">' + g('76 107') + ' kg</span> fără preț</span></th>'
                f'<td class="num">{g("1 407 280")}</td><td class="num">{g("12 448 825")}</td><td class="num">394,07</td><td class="num">2,35</td></tr></tbody>')
        title, note = 'Cost hrană · Seria 2026', 'din livrările cu preț · până pe <span class="ip-num">16.11.2026</span>'
    foot = ('' if extra else
        '<p class="ip-note ip-cost__foot">Lei = suma livrărilor cu preț; kg fără preț nu intră în cost. '
        'Lei/pasăre: la păsările de la începutul ciclului (Telița: populare <span class="ip-num">07.10.2025</span>, '
        f'<span class="ip-sex ip-sex--hen">♀ {g("28 080")}</span> · <span class="ip-sex ip-sex--rooster">♂ {g("3 510")}</span>; '
        f'Hîrbovăț: sosite <span class="ip-num">22–28.02.2026</span>, <span class="ip-sex ip-sex--hen">♀ {g("26 765")}</span> · '
        f'<span class="ip-sex ip-sex--rooster">♂ {g("3 060")}</span>; seria: populare). '
        f'Lei/ou: ouă totale (Cat. 1–3 + Incubabile) până pe <span class="ip-num">16.11.2026</span>: <span class="ip-num">{g("5 301 053")}</span>.</p>')
    return (f'<section class="ip-cost{" ip-cost--wide" if wide else ""}" aria-labelledby="cost-{sfx}"><div class="ip-cost__head"><h3 class="ip-cap" id="cost-{sfx}">{title}</h3>'
            f'<span class="ip-note">{note}</span></div><div class="ip-cost__scroll"{"" if wide else f' role="region" tabindex="0" aria-labelledby="cost-{sfx}"'}>'
            f'<table class="ip-table ip-cost__table"><caption class="ip-sr">{title}</caption>{hdr}{body}</table></div>{foot}</section>')

def cap(t): return f'<span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">{t}</span>'
def h(t, i): return f'<h2 class="ip-cap" id="{i}" style="margin:0">{t}</h2>'

STOCK = ('<!-- @dsCard group="Moment" height=1500 subtitle="SYS04" -->\n'
 '<div class="ip-root ip-stack" style="gap:var(--space-6)">\n'
 + '<div>' + cap('S11 · Hrană · Stoc · 1440 (conținut 1136 px lângă bara laterală) · până la 4 pe rând') +
 '<div class="ip-stack" style="width:1136px;gap:var(--space-2)">' + h('Stoc · 16.11.2026', 'st-d') + stock_row('d') + FOOT + '</div></div>\n'
 + '<div>' + cap('Exemple de stare (D39): stoc sub 0 (abatere, D50) · fără consum recent') +
 '<div class="ip-stack" style="width:1136px;gap:var(--space-2)">' + state_row('d') + '</div></div>\n'
 + '<div>' + cap('390 · telefon · readout-urile se așază unul sub altul') +
 '<div class="ip-phone" style="width:390px"><div class="ip-stack" style="padding:var(--space-4);gap:var(--space-2)">' + h('Stoc · 16.11.2026', 'st-m')
 + stock_row('m') + state_row('m') + FOOT + '</div></div></div>\n'
 + '<p class="ip-note" style="margin:0">Hîrbovăț: sample (D27), sample-hirbovat-2026.json › stock_today, feed_by_type_daily. Telița: real, ciclu închis 28.02.2026 (cycles.json). "Stoc sub 0" (' + M + g('1 240') + ' kg, exemplul spec C2) și "Fermă exemplu" sunt exemple de stare (D39).</p>\n'
 '</div>\n')

COST = ('<!-- @dsCard group="Data" height=1100 subtitle="SYS04" -->\n'
 '<div class="ip-root ip-stack" style="gap:var(--space-6)">\n'
 + '<div>' + cap('S11 · Hrană · Cost · 1440 (conținut 1136 px) · owner, manager, viewer (același markup; ascuns pentru vet)') +
 '<div style="width:1136px">' + cost_block('d') + '</div></div>\n'
 + '<div>' + cap('Exemplu de stare (D39): nicio livrare cu preț') + '<div style="width:1136px">' + cost_block('x', extra=True) + '</div></div>\n'
 + '<div>' + cap('390 · telefon · derulare orizontală, prima coloană fixă') +
 '<div class="ip-phone" style="width:390px"><div style="padding:var(--space-4)">' + cost_block('m', wide=False) + '</div></div></div>\n'
 + '<p class="ip-note" style="margin:0">Telița: real (feed-deliveries.json, 13 livrări, toate cu preț). Hîrbovăț: sample (D27), sample-hirbovat-2026.json › deliveries, eggs_daily.</p>\n'
 '</div>\n')

for name, html in [('StockReadout', STOCK), ('CostBlock', COST)]:
    d = ROOT / name; d.mkdir(exist_ok=True)
    (d / 'preview.html').write_text(html, encoding='utf-8')
print('ok')
