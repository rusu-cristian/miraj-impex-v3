package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// migrateEnv gives the migrate command only MIGRATE_DATABASE_URL, plus a
// DATABASE_URL that would fail if it were used.
func migrateEnv(dsn string) func(string) (string, bool) {
	env := map[string]string{
		"MIGRATE_DATABASE_URL": dsn,
		"DATABASE_URL":         "postgres://nobody@127.0.0.1:1/unused?sslmode=disable",
	}
	return func(k string) (string, bool) { v, ok := env[k]; return v, ok }
}

func runCmd(t *testing.T, lookup func(string) (string, bool), args ...string) (int, string, string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code := run(context.Background(), args, lookup, &out, &errOut)
	return code, out.String(), errOut.String()
}

// doneLine returns the "migrate up done" log record.
func doneLine(t *testing.T, out string) map[string]any {
	t.Helper()
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		var rec map[string]any
		if json.Unmarshal(sc.Bytes(), &rec) == nil && rec["msg"] == "migrate up done" {
			return rec
		}
	}
	t.Fatalf("no 'migrate up done' line in:\n%s", out)
	return nil
}

func appliedVersions(t *testing.T, dsn string) []int64 {
	t.Helper()
	sqlDB, err := db.OpenSQL(dsn, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	rows, err := sqlDB.Query("select version_id from goose_db_version where version_id > 0 order by id")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var vs []int64
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		vs = append(vs, v)
	}
	return vs
}

func TestMigrateUpStatusIdempotent(t *testing.T) {
	t.Parallel()
	dsn := testdb.Empty(t)
	env := migrateEnv(dsn)

	code, out, errOut := runCmd(t, env, "migrate", "up")
	if code != exitOK {
		t.Fatalf("up: exit %d\n%s%s", code, out, errOut)
	}
	if got := doneLine(t, out); got["applied"] != float64(3) || got["version"] != float64(3) {
		t.Fatalf("first up: %v", got)
	}

	code, out, errOut = runCmd(t, env, "migrate", "status")
	if code != exitOK {
		t.Fatalf("status: exit %d\n%s%s", code, out, errOut)
	}
	for _, file := range []string{"0001_baseline.sql", "0002_roles.sql", "0003_closed_series_on.sql"} {
		line := lineWith(out, file)
		if line == "" || !strings.Contains(line, "applied") {
			t.Fatalf("status has no applied line for %s:\n%s", file, out)
		}
	}

	code, out, errOut = runCmd(t, env, "migrate", "up")
	if code != exitOK {
		t.Fatalf("second up: exit %d\n%s%s", code, out, errOut)
	}
	if got := doneLine(t, out); got["applied"] != float64(0) || got["version"] != float64(3) {
		t.Fatalf("second up changed something: %v", got)
	}
	if vs := appliedVersions(t, dsn); len(vs) != 3 || vs[0] != 1 || vs[1] != 2 || vs[2] != 3 {
		t.Fatalf("goose_db_version = %v", vs)
	}
}

func TestMigrateConcurrentRuns(t *testing.T) {
	t.Parallel()
	dsn := testdb.Empty(t)
	env := migrateEnv(dsn)
	var wg sync.WaitGroup
	codes := make([]int, 2)
	outs := make([]string, 2)
	for i := range codes {
		wg.Go(func() {
			var out, errOut bytes.Buffer
			codes[i] = run(context.Background(), []string{"migrate", "up"}, env, &out, &errOut)
			outs[i] = out.String() + errOut.String()
		})
	}
	wg.Wait()
	for i, c := range codes {
		if c != exitOK {
			t.Fatalf("run %d: exit %d\n%s", i, c, outs[i])
		}
	}
	// Each migration was applied exactly once.
	if vs := appliedVersions(t, dsn); len(vs) != 3 {
		t.Fatalf("goose_db_version = %v", vs)
	}
}

func TestMigrateErrors(t *testing.T) {
	t.Parallel()
	none := func(string) (string, bool) { return "", false }
	for _, tc := range []struct {
		name string
		args []string
		code int
		want string
	}{
		{"no command", nil, exitUsage, "usage"},
		{"unknown command", []string{"seed"}, exitUsage, "usage"},
		{"no migrate verb", []string{"migrate"}, exitUsage, "usage: api migrate"},
		{"bad migrate verb", []string{"migrate", "down"}, exitUsage, "usage: api migrate"},
		{"missing DSN", []string{"migrate", "up"}, exitError, "MIGRATE_DATABASE_URL"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			code, out, errOut := runCmd(t, none, tc.args...)
			if code != tc.code || !strings.Contains(out+errOut, tc.want) {
				t.Fatalf("exit %d, output %q", code, out+errOut)
			}
		})
	}
	// An unreachable server fails without printing the DSN's password.
	env := func(k string) (string, bool) {
		if k == "MIGRATE_DATABASE_URL" {
			return "postgres://il_owner:hunter2@127.0.0.1:1/impex?sslmode=disable&connect_timeout=2", true
		}
		return "", false
	}
	code, out, errOut := runCmd(t, env, "migrate", "up")
	if code != exitError || strings.Contains(out+errOut, "hunter2") {
		t.Fatalf("exit %d, output %q", code, out+errOut)
	}
}

func lineWith(out, s string) string {
	for line := range strings.SplitSeq(out, "\n") {
		if strings.Contains(line, s) {
			return line
		}
	}
	return ""
}
