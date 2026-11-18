"""Screen CSS shared by the phone state boards (S7S, S6S; S13 gets it through s7s_build.CSS): the 390×844 viewport, the
Banner with an action, and the skeleton columns (the Toast with Anulează and its placement moved to additions.css §48, FIN02:
.ip-toastwrap[--page], .ip-toast__txt, .ip-toast__undo). Class names stay `s7s-` (as S6 uses s7-screen,
s7-main, s7-savebar). Strings are concatenated in this order by s7s_build (its boards are byte-identical to before).
MAIN: the phone form main gap (space-5; s7-main is space-6). SIDE / M0 / REC / EFFECT: the desktop entry form in a non-modal SidePanel beside the page (D73) and the form's record icon
and EffectPreview line-up, shared by s13_build and s9_build. Class names stay `s13-` (as S6 reuses s7s-); s13_build
concatenates them in its own order (its boards are byte-identical to before).
FIN: move the Banner-action layout into additions.css as system classes."""

VIEW = '''
.s7s-vp{height:844px;min-height:0;overflow:hidden}
.s7s-view{flex:1;min-height:0;overflow:hidden;position:relative;display:flex;flex-direction:column}
.s7s-view > .s7-main{flex:none}
.s7s-view--end{justify-content:flex-end}'''

BANNER = '''
.s7s-banner .ip-icon{flex:none;margin-top:var(--ip-hair)}
.s7s-banner__body{display:flex;flex-direction:column;align-items:flex-start;gap:var(--space-2)}
.s7s-banner__body > .ip-btn{margin-top:var(--space-1)}'''

SKEL = '''
.s7s-skblock{display:flex;flex-direction:column;gap:var(--space-6)}
.s7s-skrow{align-items:start}
.s7s-skcol{display:flex;flex-direction:column;align-items:flex-start;gap:var(--space-1);flex:1;min-width:0}
.s7s-skcol--end{align-items:flex-end}'''


# the page reflows beside the panel (margin = panel width, nothing under it); board height 1024 (desktop board h)
SIDE_H = 1024
SIDE = '''
.s13-vp{position:relative;height:''' + str(SIDE_H) + '''px;overflow:hidden;margin-right:var(--panel-w)}
.s13-panel{position:absolute;top:0;right:0;bottom:0;width:var(--panel-w)}
.s13-phead{gap:var(--space-2);flex-wrap:nowrap}
.s13-pbody{flex:1;min-height:0;overflow:auto;gap:var(--space-4)}
.s13-pkeys{gap:var(--space-2);flex-wrap:nowrap}'''

MAIN = '\n.s13-main{gap:var(--space-5)}'  # the phone form main gap (s7-main is space-6)

M0 = '''
.s13-m0{margin:0}'''

OPT = '''
.s13-opt{font-weight:400;color:var(--ink-muted)}'''  # "(opțional)" after a field label (craft §5), shared by s13_build and s10_build

REC = '''
.s13-rec{color:var(--c);display:inline-flex;flex:none}'''

EFFECT = '''
.s13-effect{align-items:flex-start}
.s13-effect > .s13-rec{margin-top:var(--ip-hair)}'''

# EffectPreview body: one line per scope (name left, before → after right), cycle-mark lines with a sub-line; and the phone
# screen as the anchor of a modal over it (BottomSheet scrim, Dialog scrim). Shared by s9_build and s10_build; class names
# stay `s9-` (s9_build concatenates them where its own rules stood, so its boards are byte-identical to before).
FX = '''
.s9-fxbody{display:flex;flex-direction:column;gap:var(--space-1);min-width:0;flex:1}
.s9-fxline{display:flex;flex-wrap:wrap;justify-content:space-between;gap:0 var(--space-2)}
.s9-fxval{display:inline-flex;align-items:baseline;gap:var(--space-1);white-space:nowrap}
.s9-fxval .ip-pair{gap:var(--space-1)}
.s9-to{color:var(--ink-secondary)}
.s9-strong{font-weight:500;color:var(--ink)}
.s9-fxcyc{display:flex;flex-wrap:wrap;align-items:center;gap:0 var(--space-2);margin-top:var(--space-1)}
.s9-fxsub{color:var(--ink-secondary);font-size:var(--ip-fs-small);line-height:var(--ip-lh-small)}'''

MODAL = '''
.s9-modal{position:relative} /* the BottomSheet scrim fills the phone screen */'''
