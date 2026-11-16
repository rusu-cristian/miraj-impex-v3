"""Build S4 · Fermă · Rezumat (the default board) from the extracts.
python3 -I s4_build.py <repo_root> [<height>]
Since S4P the default board is drawn by the same functions as its 8 state boards (s4s_build.py), so the shell,
the figures and the formats (D46: no space before %) can't drift between them. Writes
design/canvas/project/S4-ferma-rezumat.dc.html and /tmp/s4s/S4-ferma-rezumat.html (plain wrapper)."""
import sys, pathlib, importlib.util

H = int(sys.argv[2]) if len(sys.argv) > 2 else 1874
spec = importlib.util.spec_from_file_location('s4s_build', pathlib.Path(__file__).resolve().with_name('s4s_build.py'))
s4s = importlib.util.module_from_spec(spec); spec.loader.exec_module(s4s)  # reads sys.argv[1] (repo root)

body = s4s.shell(s4s.main_page(panel=s4s.panel_default()))
s4s.write('S4-ferma-rezumat', 'S4 · Fermă · Rezumat', body, H=H, css=s4s.CSS)
