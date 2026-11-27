package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

func TestMain(m *testing.M) { testdb.Main(m) }

var extracts = filepath.Join("..", "..", "..", "design", "data", "extracts")

// unreachable is a DSN nothing listens on: a run that gets as far as the
// database fails with a connection error instead of a usage one.
const unreachable = "postgres://nobody@127.0.0.1:1/none?connect_timeout=1"

type result struct {
	code           int
	stdout, stderr string
}

func runCmd(t *testing.T, env map[string]string, args ...string) result {
	t.Helper()
	var out, errOut bytes.Buffer
	lookup := func(k string) (string, bool) { v, ok := env[k]; return v, ok }
	code := run(context.Background(), args, lookup, timex.FixedClock(testdb.Now()), &out, &errOut)
	return result{code, out.String(), errOut.String()}
}

// asRole is the test server's DSN for database name, acting as role from the
// start of the session (the startup option role=…).
func asRole(name, role string) string {
	u, err := url.Parse(testdb.DSN(name))
	if err != nil {
		panic(err)
	}
	q := u.Query()
	q.Set("options", "--role="+role)
	u.RawQuery = q.Encode()
	return u.String()
}

func TestHelp(t *testing.T) {
	r := runCmd(t, nil, "--help")
	if r.code != exitOK || !strings.Contains(r.stderr, "--extracts") || strings.Count(r.stderr, "dev only (APP_ENV=dev)") != 2 {
		t.Errorf("--help: code %d, stderr %q", r.code, r.stderr)
	}
}

func TestUsageErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		env  map[string]string
		args []string
		want string
	}{
		{"no extracts", nil, []string{"--db", unreachable}, "--extracts is required"},
		{"stray argument", nil, []string{"--extracts", extracts, "now"}, "unexpected argument"},
		{"no DSN", nil, []string{"--extracts", extracts}, "set DATABASE_URL"},
		{"replace without DSN", map[string]string{"APP_ENV": "dev"}, []string{"--extracts", extracts, "--replace"}, "set MIGRATE_DATABASE_URL"},
		// --replace and --seed-dev run only with APP_ENV=dev (an allowlist).
		{"replace in production", map[string]string{"APP_ENV": "production"}, []string{"--extracts", extracts, "--replace", "--db", unreachable}, "--replace is dev only: it runs only with APP_ENV=dev (APP_ENV is production)"},
		{"replace in staging", map[string]string{"APP_ENV": "staging"}, []string{"--extracts", extracts, "--replace", "--db", unreachable}, "--replace is dev only: it runs only with APP_ENV=dev (APP_ENV is staging)"},
		{"replace with APP_ENV unset", nil, []string{"--extracts", extracts, "--replace", "--db", unreachable}, "--replace is dev only: it runs only with APP_ENV=dev (APP_ENV is unset)"},
		{"replace with APP_ENV blank", map[string]string{"APP_ENV": " "}, []string{"--extracts", extracts, "--replace", "--db", unreachable}, "(APP_ENV is unset)"},
		{"seed-dev in production", map[string]string{"APP_ENV": "production"}, []string{"--extracts", extracts, "--seed-dev", "--db", unreachable}, "--seed-dev is dev only: it runs only with APP_ENV=dev (APP_ENV is production)"},
		{"seed-dev in staging", map[string]string{"APP_ENV": "staging"}, []string{"--extracts", extracts, "--seed-dev", "--db", unreachable}, "--seed-dev is dev only: it runs only with APP_ENV=dev (APP_ENV is staging)"},
		{"seed-dev with APP_ENV unset", nil, []string{"--extracts", extracts, "--seed-dev", "--db", unreachable}, "--seed-dev is dev only: it runs only with APP_ENV=dev (APP_ENV is unset)"},
		{"seed-dev with a bad APP_ENV", map[string]string{"APP_ENV": "Dev"}, []string{"--extracts", extracts, "--seed-dev", "--db", unreachable}, "APP_ENV must be"},
		{"bad APP_ENV", map[string]string{"APP_ENV": "prod"}, []string{"--extracts", extracts, "--db", unreachable}, "APP_ENV must be"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := runCmd(t, tc.env, tc.args...)
			if r.code != exitUsage || !strings.Contains(r.stderr, tc.want) {
				t.Errorf("code %d, stderr %q; want %d and %q", r.code, r.stderr, exitUsage, tc.want)
			}
		})
	}
}

// In dev the two flags pass the APP_ENV check and the run goes on to the
// database (unreachable here): the refusal is an allowlist, not a denylist.
func TestDevFlagsAllowedInDev(t *testing.T) {
	for _, flag := range []string{"--replace", "--seed-dev"} {
		r := runCmd(t, map[string]string{"APP_ENV": "dev"}, "--extracts", extracts, flag, "--db", unreachable)
		if r.code != exitError || strings.Contains(r.stderr, "dev only") {
			t.Errorf("%s with APP_ENV=dev: code %d, stderr %q; want it past the APP_ENV check (exit %d at the database)",
				flag, r.code, r.stderr, exitError)
		}
	}
}

func TestBadExtractsBeforeTheDatabase(t *testing.T) {
	// The DSN is unreachable: the error must be about the extracts, so the
	// command stopped before connecting.
	r := runCmd(t, nil, "--db", unreachable, "--extracts", filepath.Join(t.TempDir(), "missing"))
	if r.code != exitError || !strings.Contains(r.stderr, "extracts directory") || strings.Contains(r.stderr, "connect") {
		t.Errorf("code %d, stderr %q", r.code, r.stderr)
	}
}

func TestImportOnceThenRefuse(t *testing.T) {
	d := testdb.New(t)
	dsn := asRole(d.Name, testdb.RoleApp)
	r := runCmd(t, nil, "--db", dsn, "--extracts", extracts)
	if r.code != exitOK {
		t.Fatalf("first run: code %d, stderr:\n%s", r.code, r.stderr)
	}
	for _, line := range []string{"entry  31", "bird_event  4041", "audit_log  2"} {
		if !strings.Contains(strings.Join(strings.Fields(r.stdout), "  "), line) {
			t.Errorf("stdout lacks %q:\n%s", line, r.stdout)
		}
	}
	if strings.Contains(r.stdout+r.stderr, "postgres://") {
		t.Error("the DSN was printed")
	}

	r = runCmd(t, map[string]string{"DATABASE_URL": dsn}, "--extracts", extracts)
	if r.code != exitError || !strings.Contains(r.stderr, "already holds the history migration") {
		t.Errorf("second run: code %d, stderr %q", r.code, r.stderr)
	}

	// --replace as il_app is refused (it can't empty audit_log).
	r = runCmd(t, map[string]string{"APP_ENV": "dev"}, "--extracts", extracts, "--replace", "--db", dsn)
	if r.code != exitError || !strings.Contains(r.stderr, "needs the owner role") {
		t.Errorf("--replace as il_app: code %d, stderr %q", r.code, r.stderr)
	}

	// --replace in dev with the owner DSN from MIGRATE_DATABASE_URL.
	r = runCmd(t, map[string]string{"APP_ENV": "dev", "MIGRATE_DATABASE_URL": asRole(d.Name, testdb.RoleOwner)},
		"--extracts", extracts, "--replace")
	if r.code != exitOK {
		t.Fatalf("--replace: code %d, stderr:\n%s", r.code, r.stderr)
	}
	var entries int
	if err := d.App.QueryRow(t.Context(), "select count(*) from entry").Scan(&entries); err != nil || entries != 31 {
		t.Errorf("entries after --replace = %d (%v), want 31", entries, err)
	}
}

func TestSeedDevPrintsPasswordsOnceAndNeverLogs(t *testing.T) {
	var logs bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(prev) })

	d := testdb.New(t)
	r := runCmd(t, map[string]string{"APP_ENV": "dev", "DATABASE_URL": asRole(d.Name, testdb.RoleApp)}, "--extracts", extracts, "--seed-dev")
	if r.code != exitOK {
		t.Fatalf("code %d, stderr:\n%s", r.code, r.stderr)
	}
	for _, login := range []string{"proprietar", "ion.rusu", "dr.ceban", "observator"} {
		var password string
		for _, line := range strings.Split(r.stdout, "\n") {
			if f := strings.Split(line, "\t"); len(f) == 3 && f[0] == login {
				password = f[1]
			}
		}
		if len(password) != 12 {
			t.Errorf("no 12-character password for %s in stdout:\n%s", login, r.stdout)
			continue
		}
		if n := strings.Count(r.stdout, login+"\t"); n != 1 {
			t.Errorf("%s printed %d times", login, n)
		}
		if n := strings.Count(r.stdout, password); n != 1 {
			t.Errorf("the password of %s printed %d times", login, n)
		}
		if strings.Contains(r.stderr, password) || strings.Contains(logs.String(), password) {
			t.Errorf("the password of %s reached stderr or the log", login)
		}
	}
	if !strings.Contains(r.stdout, "app_user") {
		t.Errorf("stdout lacks the app_user count:\n%s", r.stdout)
	}
}

// failingWriter fails every write, as a closed stdout would.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write /dev/stdout: broken pipe")
}

// When the dev passwords can't be printed after the commit, the run fails
// and says the import was committed, the passwords are lost, and --replace
// (dev, owner DSN) is the way back. The import itself stays.
func TestSeedDevPasswordsLostAfterCommit(t *testing.T) {
	d := testdb.New(t)
	var errOut bytes.Buffer
	lookup := func(k string) (string, bool) {
		v, ok := map[string]string{"APP_ENV": "dev", "DATABASE_URL": asRole(d.Name, testdb.RoleApp)}[k]
		return v, ok
	}
	code := run(context.Background(), []string{"--extracts", extracts, "--seed-dev"}, lookup,
		timex.FixedClock(testdb.Now()), failingWriter{}, &errOut)
	if code != exitError {
		t.Errorf("code %d, want %d", code, exitError)
	}
	for _, want := range []string{"the import was committed", "passwords failed", "lost", "--replace", "APP_ENV=dev",
		"owner DSN", "MIGRATE_DATABASE_URL", "broken pipe"} {
		if !strings.Contains(errOut.String(), want) {
			t.Errorf("stderr lacks %q:\n%s", want, errOut.String())
		}
	}
	var entries, users int
	if err := d.App.QueryRow(t.Context(), "select (select count(*) from entry), (select count(*) from app_user)").
		Scan(&entries, &users); err != nil {
		t.Fatal(err)
	}
	if entries == 0 || users != 4 {
		t.Errorf("after the failed print: %d entries, %d users; want the committed import with the 4 dev users", entries, users)
	}
}
