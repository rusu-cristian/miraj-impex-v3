"""Screen CSS shared by the phone state boards (S7S, S6S; S13 gets it through s7s_build.CSS): the 390×844 viewport, the
Toast with Anulează, the Banner with an action, and the skeleton columns. Class names stay `s7s-` (as S6 uses s7-screen,
s7-main, s7-savebar). Strings are concatenated in this order by s7s_build (its boards are byte-identical to before).
FIN: move the Toast-undo and Banner-action layouts into additions.css as system classes."""

VIEW = '''
.s7s-vp{height:844px;min-height:0;overflow:hidden}
.s7s-view{flex:1;min-height:0;overflow:hidden;position:relative;display:flex;flex-direction:column}
.s7s-view > .s7-main{flex:none}
.s7s-view--end{justify-content:flex-end}'''

TOAST = '''
.s7s-toastwrap{position:absolute;left:var(--space-4);right:var(--space-4);bottom:var(--space-4);z-index:3}
.s7s-toast{display:flex;width:100%;align-items:flex-start}
.s7s-toast > .ip-icon{flex:none;margin-top:var(--ip-hair)}
.s7s-toast__txt{flex:1;min-width:0}
.s7s-undo{background:none;border:0;padding:0 var(--space-2);margin:calc((var(--space-5) - var(--control-h-touch)) / 2) calc(-1 * var(--space-2));min-height:var(--control-h-touch);font:inherit;font-weight:500;flex:none;align-self:center}'''

BANNER = '''
.s7s-banner .ip-icon{flex:none;margin-top:var(--ip-hair)}
.s7s-banner__body{display:flex;flex-direction:column;align-items:flex-start;gap:var(--space-2)}
.s7s-banner__body > .ip-btn{margin-top:var(--space-1)}'''

SKEL = '''
.s7s-skblock{display:flex;flex-direction:column;gap:var(--space-6)}
.s7s-skrow{align-items:start}
.s7s-skcol{display:flex;flex-direction:column;align-items:flex-start;gap:var(--space-1);flex:1;min-width:0}
.s7s-skcol--end{align-items:flex-end}'''
