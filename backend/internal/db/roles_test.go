package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// TestRoleGrants checks migration 0002_roles.sql (SPEC §4.10, §11.9).
func TestRoleGrants(t *testing.T) {
	t.Parallel()
	d := testdb.New(t)
	ctx := context.Background()
	maint := d.Pool(testdb.RoleMaint)

	// Rows to aim the statements at, written by the owner.
	owner := d.Pool(testdb.RoleOwner)
	if _, err := owner.Exec(ctx, `
		insert into app_user (id, name, login, password_hash, role) values ('00000000-0000-0000-0000-000000000001', 'P', 'proprietar', 'h', 'owner');
		insert into audit_log (action, entity, entity_id) values ('create', 'app_user', '00000000-0000-0000-0000-000000000001');
		insert into live_event (farm_ids, type, payload) values ('{}', 'users', '{}');
		insert into request_replay (client_request_id, user_id, route, status, body)
		  values ('00000000-0000-0000-0000-0000000000aa', '00000000-0000-0000-0000-000000000001', 'POST /x', 201, '\x7b7d');
		insert into login_attempt (login, success) values ('ion', false);`); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name   string
		pool   string
		sql    string
		denied bool
	}{
		{"app: update audit_log", testdb.RoleApp, "update audit_log set summary = 'x'", true},
		{"app: delete audit_log", testdb.RoleApp, "delete from audit_log", true},
		{"app: insert audit_log", testdb.RoleApp, "insert into audit_log (action, entity, entity_id) values ('create', 'farm', gen_random_uuid())", false},
		{"app: delete live_event", testdb.RoleApp, "delete from live_event", true},
		{"app: update live_event", testdb.RoleApp, "update live_event set type = 'x'", true},
		{"app: insert live_event", testdb.RoleApp, "insert into live_event (farm_ids, type, payload) values ('{}', 'x', '{}')", false},
		{"app: update request_replay", testdb.RoleApp, "update request_replay set status = 200", true},
		{"app: delete request_replay", testdb.RoleApp, "delete from request_replay", true},
		{"app: select request_replay", testdb.RoleApp, "select count(*) from request_replay", false},
		{"app: insert login_attempt", testdb.RoleApp, "insert into login_attempt (login, at, success) values ('ion', now(), false)", false},
		{"app: update app_user", testdb.RoleApp, "update app_user set name = 'Proprietar'", false},
		{"app: select a view", testdb.RoleApp, "select count(*) from v_section_sex_day", false},
		{"app: execute closed_series_on (0003)", testdb.RoleApp, "select closed_series_on(gen_random_uuid(), current_date)", false},
		{"app: select goose_db_version", testdb.RoleApp, "select max(version_id) from goose_db_version", false},
		{"app: insert goose_db_version", testdb.RoleApp, "insert into goose_db_version (version_id, is_applied) values (99, true)", true},
		{"maint: delete login_attempt", testdb.RoleMaint, "delete from login_attempt", false},
		{"maint: delete live_event", testdb.RoleMaint, "delete from live_event", false},
		{"maint: delete request_replay", testdb.RoleMaint, "delete from request_replay", false},
		{"maint: insert app_user", testdb.RoleMaint, "insert into app_user (name, login, password_hash, role) values ('X', 'xxx', 'h', 'vet')", true},
		{"maint: select app_user", testdb.RoleMaint, "select count(*) from app_user", true},
		{"maint: update login_attempt", testdb.RoleMaint, "update login_attempt set success = true", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool := d.App
			if tc.pool == testdb.RoleMaint {
				pool = maint
			}
			_, err := pool.Exec(ctx, tc.sql)
			var pgErr *pgconn.PgError
			denied := errors.As(err, &pgErr) && pgErr.Code == "42501"
			if tc.denied != denied {
				t.Fatalf("err = %v, want denied=%v (SQLSTATE 42501)", err, tc.denied)
			}
			if !tc.denied && err != nil {
				t.Fatalf("err = %v", err)
			}
		})
	}
}
