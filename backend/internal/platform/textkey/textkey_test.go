package textkey

import "testing"

func TestKey(t *testing.T) {
	for _, tc := range []struct {
		name, in, want string
	}{
		{"comma below", "Hîrbovăț", "hirbovat"},
		{"plain ASCII", "Hirbovat", "hirbovat"},
		// Cedilla forms written as escapes: Ţ U+0162, Ş U+015E.
		{"upper case with cedilla", "HÎRBOVAŢ", "hirbovat"},
		{"s cedilla and s comma", "Cocoşi / Cocoși", "cocosi/cocosi"},
		{"decomposed input", "Telița", "telita"},
		{"spaces dropped", "  Ross 308   creștere ", "ross308crestere"},
		{"series", "Seria 2026", "seria2026"},
		{"unit words dropped", "Hrană g / pasăre", "hrana/pasare"},
		{"percent dropped", "Ouat %", "ouat"},
		{"unit letter inside a word kept", "Greutate gaini", "greutategaini"},
		{"tabs and newlines are spaces", "Pre-\touat\n", "pre-ouat"},
		{"empty", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Key(tc.in); got != tc.want {
				t.Errorf("Key(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestKeyCollisions(t *testing.T) {
	// D111: "Hirbovat" collides with "Hîrbovăț", whatever the case or the
	// s/t variant typed.
	want := Key("Hîrbovăț")
	for _, in := range []string{"Hirbovat", "HÎRBOVAŢ", "hîrbovăţ", "HIRBOVAT", "Hîrbovăț "} {
		if got := Key(in); got != want {
			t.Errorf("Key(%q) = %q, want %q (same as Hîrbovăț)", in, got, want)
		}
	}
	if Key("Telița") == want {
		t.Error("different names share a key")
	}
}
