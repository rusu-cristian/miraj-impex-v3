package audit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

func TestMain(m *testing.M) { testdb.Main(m) }

func ptr[T any](v T) *T { return &v }

func id(t *testing.T, d *testdb.DB, sql string, args ...any) uuid.UUID {
	t.Helper()
	var v uuid.UUID
	if err := d.App.QueryRow(context.Background(), sql, args...).Scan(&v); err != nil {
		t.Fatal(err)
	}
	return v
}

// TestWrite: every §4.10 column is written as given, by il_app.
func TestWrite(t *testing.T) {
	ctx := context.Background()
	d := testdb.New(t)
	user := id(t, d, `insert into app_user (name, login, password_hash, role) values ('Ion', 'ion', 'x', 'manager') returning id`)
	farm := id(t, d, `insert into farm (name, name_key, type) values ('Hîrbovăț', 'hirbovat', 'laying') returning id`)
	entry := id(t, d, `insert into entry (kind, farm_id, date, created_by) values ('feed_daily', $1, '2026-11-18', $2) returning id`, farm, user)

	at := time.Date(2026, time.November, 18, 8, 44, 0, 123456000, timex.Chisinau)
	kind := dbq.EntryKindFeedDaily
	row := Row{
		At: at, UserID: &user, Action: dbq.AuditActionUpdate, Entity: "entry", EntityID: entry, EntryID: &entry,
		EntryKind: &kind, RecordTypes: []string{"hrana"}, FarmIDs: []uuid.UUID{farm}, RecordDate: ptr(testdb.Today),
		Reason: ptr("greșeală de tastare"), SameDayAuthor: true, Summary: ptr("Hrană zilnică · 8 din 8 rânduri · 2 457,9 kg"),
		Changes:   []map[string]any{{"key": "H5.1 · ♀ g", "old": 158, "new": 160}},
		OldValues: map[string]any{"rows": []int{158}}, NewValues: map[string]any{"rows": []int{160}},
		Source: dbq.RecordSourceVoice, VoiceTranscript: ptr("pune 160 de grame"),
	}
	if err := Write(ctx, dbq.New(d.App), row); err != nil {
		t.Fatal(err)
	}

	var got dbq.AuditLog
	err := d.App.QueryRow(ctx, `select at, user_id, action, entity, entity_id, entry_id, entry_kind, record_types,
		farm_ids, record_date, reason, same_day_author, summary, changes::text, old_values::text, new_values::text,
		source, voice_transcript from audit_log`).Scan(&got.At, &got.UserID, &got.Action, &got.Entity, &got.EntityID,
		&got.EntryID, &got.EntryKind, &got.RecordTypes, &got.FarmIds, &got.RecordDate, &got.Reason, &got.SameDayAuthor,
		&got.Summary, &got.Changes, &got.OldValues, &got.NewValues, &got.Source, &got.VoiceTranscript)
	if err != nil {
		t.Fatal(err)
	}
	checks := []struct {
		name string
		ok   bool
	}{
		{"at", got.At.Equal(at)},
		{"user_id", got.UserID != nil && *got.UserID == user},
		{"action", got.Action == dbq.AuditActionUpdate},
		{"entity", got.Entity == "entry" && got.EntityID == entry},
		{"entry", got.EntryID != nil && *got.EntryID == entry && got.EntryKind != nil && *got.EntryKind == kind},
		{"record_types", len(got.RecordTypes) == 1 && got.RecordTypes[0] == "hrana"},
		{"farm_ids", len(got.FarmIds) == 1 && got.FarmIds[0] == farm},
		{"record_date", got.RecordDate != nil && *got.RecordDate == testdb.Today},
		{"reason", got.Reason != nil && *got.Reason == "greșeală de tastare"},
		{"same_day_author", got.SameDayAuthor},
		{"summary", got.Summary != nil && *got.Summary == "Hrană zilnică · 8 din 8 rânduri · 2 457,9 kg"},
		{"changes", string(got.Changes) == `[{"key": "H5.1 · ♀ g", "new": 160, "old": 158}]`},
		{"old_values", string(got.OldValues) == `{"rows": [158]}`},
		{"new_values", string(got.NewValues) == `{"rows": [160]}`},
		{"source", got.Source == dbq.RecordSourceVoice},
		{"voice_transcript", got.VoiceTranscript != nil && *got.VoiceTranscript == "pune 160 de grame"},
	}
	for _, c := range checks {
		if !c.ok {
			t.Errorf("column %s: got %+v", c.name, got)
		}
	}
}

// TestWriteDefaults: nil Changes (or a nil slice) is [], nil snapshots are
// null, nil lists are empty arrays, an empty Source is manual.
func TestWriteDefaults(t *testing.T) {
	ctx := context.Background()
	d := testdb.New(t)
	farm := id(t, d, `insert into farm (name, name_key, type) values ('Telița', 'telita', 'growth') returning id`)
	var noChanges []map[string]any
	for _, changes := range []any{nil, noChanges} {
		if err := Write(ctx, dbq.New(d.App), Row{
			At: testdb.Now(), Action: dbq.AuditActionSetupChange, Entity: "farm", EntityID: farm, Changes: changes,
		}); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := d.App.Query(ctx, `select changes::text, old_values is null, new_values is null, record_types::text,
		farm_ids::text, user_id is null, source from audit_log`)
	if err != nil {
		t.Fatal(err)
	}
	type result struct {
		changes              string
		oldNull, newNull     bool
		recordTypes, farmIDs string
		system               bool
		source               dbq.RecordSource
	}
	var got []result
	for rows.Next() {
		var r result
		if err := rows.Scan(&r.changes, &r.oldNull, &r.newNull, &r.recordTypes, &r.farmIDs, &r.system, &r.source); err != nil {
			t.Fatal(err)
		}
		got = append(got, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	want := result{"[]", true, true, "{}", "{}", true, dbq.RecordSourceManual}
	if len(got) != 2 || got[0] != want || got[1] != want {
		t.Fatalf("rows = %+v, want 2 × %+v", got, want)
	}
}
