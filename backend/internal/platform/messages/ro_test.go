package messages

import (
	"os"
	"strings"
	"testing"
)

// cedillas are the wrong Romanian letters, written as escapes so this file
// stays free of them.
const cedillas = "\u015f\u0163\u015e\u0162"

// specPath is backend/SPEC.md seen from this package's directory.
const specPath = "../../../SPEC.md"

func TestStrings(t *testing.T) {
	spec, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		got  string
		want string
	}{
		{"LoginRequired", LoginRequired(), "Scrie loginul."},
		{"PasswordRequired", PasswordRequired(), "Scrie parola."},
		{"LoginFailed", LoginFailed(), "Login sau parolă greșită."},
		{"LoginRateLimited", LoginRateLimited(), "Prea multe încercări. Încearcă din nou peste 5 minute."},
		{"AccountDisabled", AccountDisabled(), "Contul tău este dezactivat. Vorbește cu administratorul."},
		{"FarmForbidden", FarmForbidden(), "Nu ai acces la această fermă."},
		{"PageForbidden", PageForbidden(), "Nu ai acces la această pagină."},
		{"Internal", Internal(), "A apărut o eroare. Încearcă din nou."},
		{"PayloadTooLarge", PayloadTooLarge(), "Datele trimise sunt prea mari."},
		{"Unauthenticated", Unauthenticated(), "Sesiunea a expirat. Intră din nou în cont."},
		{"Forbidden", Forbidden(), "Cererea a fost refuzată. Reîncarcă pagina și încearcă din nou."},
		{"InvalidData", InvalidData(), "Datele trimise nu sunt valide."},
		{"Unavailable", Unavailable(), "Serviciul nu răspunde acum. Încearcă din nou."},
		{"NotFound", NotFound(), "Adresa cerută nu există."},
		{"MethodNotAllowed", MethodNotAllowed(), "Operația nu este permisă la această adresă."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("got %q, want %q", tc.got, tc.want)
			}
			// The expected text must be the SPEC's own text, byte for byte.
			if !strings.Contains(string(spec), `"`+tc.want+`"`) {
				t.Fatalf("%q is not quoted in SPEC.md", tc.want)
			}
		})
	}
}

// TestNoCedilla fails when ro.go or this file uses s/t with a cedilla
// (U+015F, U+0163, U+015E, U+0162) instead of the comma-below ș/ț/Ș/Ț.
func TestNoCedilla(t *testing.T) {
	for _, file := range []string{"ro.go", "ro_test.go"} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(string(src), "\n") {
			if strings.ContainsAny(line, cedillas) {
				t.Errorf("%s:%d uses a cedilla s/t (want U+0219/U+021B): %s", file, i+1, line)
			}
		}
	}
}

func TestMigrationStrings(t *testing.T) {
	spec, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := RegisterDifference(), "diferență registru"; got != want || !strings.Contains(string(spec), `"`+want+`"`) {
		t.Errorf("RegisterDifference() = %q, want %q quoted in SPEC.md", got, want)
	}
	if got, want := FeedStockMove(), "mutare ♀ → ♂"; got != want || !strings.Contains(string(spec), `"`+want+`"`) {
		t.Errorf("FeedStockMove() = %q, want %q quoted in SPEC.md", got, want)
	}
	// SPEC §10.3 gives the template `"plecare {out_date}"`.
	if !strings.Contains(string(spec), `"plecare {out_date}"`) {
		t.Error(`"plecare {out_date}" is not quoted in SPEC.md`)
	}
	if got, want := TransferDeparture("27.02.2026"), "plecare 27.02.2026"; got != want {
		t.Errorf("TransferDeparture = %q, want %q", got, want)
	}
}

func TestAuditKeys(t *testing.T) {
	spec, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := AuditKeyHensDead("H5.2"), "H5.2 · găini moarte"; got != want || !strings.Contains(string(spec), `"`+want+`"`) {
		t.Errorf("AuditKeyHensDead(H5.2) = %q, want %q quoted in SPEC.md", got, want)
	}
}
