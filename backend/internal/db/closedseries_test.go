package db_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// TestClosedSeriesOn checks closed_series_on (migration 0003, M4 ARC-2): the
// closed series whose cycle at the farm holds the day, both ends included,
// else null. Farm A: Seria V (closed) 01.01–31.03.2025, Seria W (active)
// from 01.05.2025, open.
func TestClosedSeriesOn(t *testing.T) {
	t.Parallel()
	d := testdb.New(t)
	ctx := t.Context()
	id := func(sql string, args ...any) uuid.UUID {
		t.Helper()
		var id uuid.UUID
		if err := d.App.QueryRow(ctx, sql+" returning id", args...).Scan(&id); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
		return id
	}
	farm := id(`insert into farm (name, name_key, type) values ('A', 'a', 'laying')`)
	other := id(`insert into farm (name, name_key, type) values ('B', 'b', 'laying')`)
	v := id(`insert into series (name, name_key, year, placement_date, status, closed_at)
		values ('Seria V', 'seria v', 2025, '2025-01-01', 'closed', $1)`, testdb.Now())
	w := id(`insert into series (name, name_key, year, placement_date) values ('Seria W', 'seria w', 2026, '2025-05-01')`)
	id(`insert into cycle (series_id, farm_id, start_date, end_date, started_by, ended_by)
		values ($1, $2, '2025-01-01', '2025-03-31', 'auto', 'manual')`, v, farm)
	id(`insert into cycle (series_id, farm_id, start_date, started_by) values ($1, $2, '2025-05-01', 'auto')`, w, farm)

	for _, tc := range []struct {
		name string
		farm uuid.UUID
		day  string
		want *string
	}{
		{"first day of the closed series' cycle", farm, "2025-01-01", ptr("Seria V")},
		{"its last day", farm, "2025-03-31", ptr("Seria V")},
		{"before any cycle", farm, "2024-12-31", nil},
		{"between cycles", farm, "2025-04-15", nil},
		{"inside the open cycle of an active series", farm, "2026-11-18", nil},
		{"another farm on a closed day", other, "2025-02-01", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got *string
			if err := d.App.QueryRow(ctx, `select closed_series_on($1, $2::date)`, tc.farm, tc.day).Scan(&got); err != nil {
				t.Fatal(err)
			}
			if (got == nil) != (tc.want == nil) || (got != nil && *got != *tc.want) {
				t.Errorf("closed_series_on = %v, want %v", deref(got), deref(tc.want))
			}
		})
	}
}

func ptr(s string) *string { return &s }

func deref(s *string) string {
	if s == nil {
		return "null"
	}
	return *s
}
