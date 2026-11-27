package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// passwordLine is a generated password alone on its line (SPEC §6.10).
var passwordLine = regexp.MustCompile(`^[abcdefghjkmnpqrstuvwxyz23456789]{12}$`)

func TestNewOwnerInput(t *testing.T) {
	t.Parallel()
	sixty := strings.Repeat("ă", 60)
	for _, tc := range []struct {
		name, login string
		want        ownerInput
		err         string
	}{
		{"Ion Rusu", "ion.rusu", ownerInput{"Ion Rusu", "ion.rusu"}, ""},
		{"  Proprietar  ", " Ion.Rusu ", ownerInput{"Proprietar", "ion.rusu"}, ""},
		{sixty, "abc", ownerInput{sixty, "abc"}, ""},
		{sixty + "x", "abc", ownerInput{}, "U1"},
		{"", "ion", ownerInput{}, "U1"},
		{"   ", "ion", ownerInput{}, "U1"},
		{"Ion", "", ownerInput{}, "U2"},
		{"Ion", "io", ownerInput{}, "U2"},
		{"Ion", strings.Repeat("a", 41), ownerInput{}, "U2"},
		{"Ion", strings.Repeat("a", 40), ownerInput{"Ion", strings.Repeat("a", 40)}, ""},
		{"Ion", "ion_rusu", ownerInput{}, "U2"},
		{"Ion", "ion rusu", ownerInput{}, "U2"},
		{"Ion", "ionuț", ownerInput{}, "U2"},
	} {
		got, err := newOwnerInput(tc.name, tc.login)
		if tc.err == "" {
			if err != nil || got != tc.want {
				t.Errorf("newOwnerInput(%q, %q) = %+v, %v; want %+v", tc.name, tc.login, got, err, tc.want)
			}
			continue
		}
		if err == nil || !strings.Contains(err.Error(), tc.err) {
			t.Errorf("newOwnerInput(%q, %q): err %v, want %s", tc.name, tc.login, err, tc.err)
		}
	}
}

func TestCreateOwner(t *testing.T) {
	t.Parallel()
	d := testdb.New(t)
	ctx := context.Background()
	store := db.NewStore(d.App) // role il_app, as DATABASE_URL

	created, err := createOwner(ctx, store, d.Clock, ownerInput{Name: "Proprietar", Login: "proprietar"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !passwordLine.MatchString(created.Password) {
		t.Fatalf("password %q", created.Password)
	}

	// The owner can sign in with the printed password.
	res, err := auth.NewService(store, d.Clock).Login(ctx, auth.LoginInput{Login: "proprietar", Password: created.Password})
	if err != nil {
		t.Fatalf("login with the generated password: %v", err)
	}
	if res.User.ID != created.ID || res.User.Role != "owner" || res.User.Name != "Proprietar" {
		t.Fatalf("user %+v", res.User)
	}

	// One audit row: user_change by the system, at the clock's now, without the hash.
	var (
		userID    *uuid.UUID
		action    string
		entity    string
		entityID  uuid.UUID
		at        time.Time
		newValues []byte
		n         int
	)
	if err := d.App.QueryRow(ctx, `select count(*) over (), user_id, action::text, entity, entity_id, at, new_values
		from audit_log`).Scan(&n, &userID, &action, &entity, &entityID, &at, &newValues); err != nil {
		t.Fatal(err)
	}
	if n != 1 || userID != nil || action != "user_change" || entity != "app_user" || entityID != created.ID ||
		!at.Equal(testdb.Now()) {
		t.Fatalf("audit row: n=%d user=%v action=%s entity=%s id=%s at=%s", n, userID, action, entity, entityID, at)
	}
	var snap map[string]any
	if err := json.Unmarshal(newValues, &snap); err != nil {
		t.Fatal(err)
	}
	if snap["login"] != "proprietar" || snap["role"] != "owner" || snap["active"] != true || len(snap) != 5 {
		t.Fatalf("new_values %s", newValues)
	}
	if bytes.Contains(newValues, []byte("argon2")) || bytes.Contains(newValues, []byte(created.Password)) {
		t.Fatalf("new_values carries the password: %s", newValues)
	}

	// An active owner exists: refused without --force, nothing written.
	if _, err := createOwner(ctx, store, d.Clock, ownerInput{Name: "Al doilea", Login: "doilea"}, false); !errors.Is(err, errOwnerExists) {
		t.Fatalf("second owner without force: %v", err)
	}
	if got := count(t, d, "select count(*) from app_user"); got != 1 {
		t.Fatalf("%d users after a refusal", got)
	}
	// --force adds a second owner.
	if _, err := createOwner(ctx, store, d.Clock, ownerInput{Name: "Al doilea", Login: "doilea"}, true); err != nil {
		t.Fatalf("second owner with force: %v", err)
	}
	// A taken login is refused even with --force.
	if _, err := createOwner(ctx, store, d.Clock, ownerInput{Name: "Altul", Login: "doilea"}, true); !errors.Is(err, errLoginTaken) {
		t.Fatalf("taken login: %v", err)
	}
	if got := count(t, d, "select count(*) from audit_log"); got != 2 {
		t.Fatalf("%d audit rows, want 2", got)
	}
}

func TestCreateOwnerDeactivatedOwnerOnly(t *testing.T) {
	t.Parallel()
	d := testdb.New(t)
	ctx := context.Background()
	store := db.NewStore(d.App)
	if _, err := createOwner(ctx, store, d.Clock, ownerInput{Name: "Vechi", Login: "vechi"}, false); err != nil {
		t.Fatal(err)
	}
	if _, err := d.App.Exec(ctx, "update app_user set active = false"); err != nil {
		t.Fatal(err)
	}
	// No active owner: allowed without --force; the deactivated login stays taken (U3).
	if _, err := createOwner(ctx, store, d.Clock, ownerInput{Name: "Vechi", Login: "vechi"}, false); !errors.Is(err, errLoginTaken) {
		t.Fatalf("deactivated login: %v", err)
	}
	if _, err := createOwner(ctx, store, d.Clock, ownerInput{Name: "Nou", Login: "nou"}, false); err != nil {
		t.Fatalf("with only a deactivated owner: %v", err)
	}
}

func TestCreateOwnerConcurrent(t *testing.T) {
	t.Parallel()
	d := testdb.New(t)
	store := db.NewStore(d.App)
	errs := make([]error, 2)
	var wg sync.WaitGroup
	for i := range errs {
		wg.Go(func() {
			login := []string{"unu", "doi"}[i]
			_, errs[i] = createOwner(context.Background(), store, d.Clock, ownerInput{Name: login, Login: login}, false)
		})
	}
	wg.Wait()
	ok, refused := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			ok++
		case errors.Is(err, errOwnerExists):
			refused++
		default:
			t.Fatal(err)
		}
	}
	if ok != 1 || refused != 1 {
		t.Fatalf("errors %v: want one owner and one refusal", errs)
	}
}

func count(t *testing.T, d *testdb.DB, sql string) int {
	t.Helper()
	var n int
	if err := d.App.QueryRow(context.Background(), sql).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func runCmd(t *testing.T, env map[string]string, args ...string) (int, string, string) {
	t.Helper()
	var out, errOut bytes.Buffer
	lookup := func(k string) (string, bool) { v, ok := env[k]; return v, ok }
	code := run(context.Background(), args, lookup, timex.FixedClock(testdb.Now()), &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestRun(t *testing.T) {
	t.Parallel()
	d := testdb.New(t)
	// Only DATABASE_URL (Q6), like the image's api service environment minus the rest.
	env := map[string]string{"DATABASE_URL": testdb.DSN(d.Name)}

	code, out, errOut := runCmd(t, env, "--name", "Proprietar E2E", "--login", "E2E.Owner", "--force")
	if code != exitOK {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	password := lines[len(lines)-1]
	if !passwordLine.MatchString(password) {
		t.Fatalf("last stdout line %q is not a password; stdout %q", password, out)
	}
	if strings.Count(out, password) != 1 || strings.Contains(errOut, password) {
		t.Fatalf("password printed more than once: stdout %q stderr %q", out, errOut)
	}
	if !strings.Contains(errOut, "login e2e.owner") {
		t.Fatalf("stderr %q", errOut)
	}
	// The e2e script's extraction: the last 12-character token of the alphabet.
	var last string
	for tok := range strings.FieldsFuncSeq(out, func(r rune) bool { return !strings.ContainsRune("abcdefghjkmnpqrstuvwxyz23456789", r) }) {
		if len(tok) == 12 {
			last = tok
		}
	}
	if last != password {
		t.Fatalf("e2e extraction gives %q, want %q", last, password)
	}

	code, out, errOut = runCmd(t, env, "--name", "Altul", "--login", "altul")
	if code != exitError || out != "" || !strings.Contains(errOut, "--force") {
		t.Fatalf("second owner: exit %d stdout %q stderr %q", code, out, errOut)
	}
	code, _, errOut = runCmd(t, env, "--name", "Altul", "--login", "e2e.owner", "--force")
	if code != exitError || !strings.Contains(errOut, `"e2e.owner" is already used`) {
		t.Fatalf("taken login: exit %d stderr %q", code, errOut)
	}
}

func TestRunErrors(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		env  map[string]string
		args []string
		code int
		want string
	}{
		{"no flags", nil, nil, exitUsage, "U1"},
		{"unknown flag", nil, []string{"--role", "vet"}, exitUsage, "usage: create-owner"},
		{"extra argument", nil, []string{"--name", "A", "--login", "abc", "extra"}, exitUsage, "unexpected argument"},
		{"bad login", nil, []string{"--name", "A", "--login", "a"}, exitUsage, "U2"},
		{"no DATABASE_URL", map[string]string{"MIGRATE_DATABASE_URL": "postgres://x@127.0.0.1:1/x"},
			[]string{"--name", "A", "--login", "abc"}, exitError, "DATABASE_URL"},
		{"unreachable database", map[string]string{
			"DATABASE_URL": "postgres://il_app:hunter2@127.0.0.1:1/impex?sslmode=disable&connect_timeout=2"},
			[]string{"--name", "A", "--login", "abc"}, exitError, "create-owner:"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			code, out, errOut := runCmd(t, tc.env, tc.args...)
			if code != tc.code || out != "" || !strings.Contains(errOut, tc.want) || strings.Contains(errOut, "hunter2") {
				t.Fatalf("exit %d, stdout %q, stderr %q", code, out, errOut)
			}
		})
	}
}
