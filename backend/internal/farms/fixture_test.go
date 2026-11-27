package farms

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// fixture is a hand-built copy of the real farms' shape (the §7.3 example
// dates), today 18.11.2026:
//
//	Seria 2026 placed 07.10.2025 at age 0 (active); Seria 2025 placed 01.11.2024 at age 0 (closed).
//	Telița (growth, guide "Ross 308 creștere"): T1, retired T9; cycle Seria 2026 07.10.2025–28.02.2026 (emptied).
//	Hîrbovăț (laying, guide "Ross 308 ouat"): H5 (H5.1, H5.2), H6 (H6.1), parentless H7;
//	  cycle Seria 2025 07.03.2025–28.01.2026 (migration, start inferred, no guide, no bird event);
//	  cycle Seria 2026 from 22.02.2026, open, guide "Ross 308 ouat".
//	Ferma Nouă (laying, no guide): N1, never a cycle.
//
//	07.10.2025 placement Telița T1 ♀ 100 · ♂ 20
//	22.02.2026 transfer Telița → Hîrbovăț: T1 → H5.1 ♀ 60 · ♂ 10
//	28.02.2026 transfer Telița → Hîrbovăț: T1 → H5.2 ♀ 40, T1 → H6.1 ♂ 10 (Telița empty)
//	01.03.2026 a deleted transfer entry Telița → Hîrbovăț ♀ 7 (not a link)
//	02.03.2026 move inside Hîrbovăț H5.1 → H6.1 ♀ 5 (not a link)
//	10.11.2026 mortality H5.1 ♀ 2; 12.11.2026 sale H5.2 ♀ 5
//
// Hîrbovăț at the end of 18.11.2026: ♀ 93 · ♂ 20.
type fixture struct {
	t                        *testing.T
	d                        *testdb.DB
	user                     uuid.UUID
	series26, series25       uuid.UUID
	guideLay, guideGrow      uuid.UUID
	hirbovat, telita, noua   uuid.UUID
	cycleH26, cycleH25, cycT uuid.UUID
	sec                      map[string]uuid.UUID
}

func day(y int, m time.Month, d int) timex.Day { return timex.MustDay(y, m, d) }

var today = day(2026, time.November, 18)

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{t: t, d: testdb.New(t), sec: map[string]uuid.UUID{}}
	f.user = f.id(`insert into app_user (name, login, password_hash, role) values ('Ion', 'ion', 'x', 'owner')`)
	f.series26 = f.id(`insert into series (name, name_key, year, placement_date) values ('Seria 2026', 'seria 2026', 2026, '2025-10-07')`)
	f.series25 = f.id(`insert into series (name, name_key, year, placement_date, status, closed_at)
		values ('Seria 2025', 'seria 2025', 2025, '2024-11-01', 'closed', $1)`, testdb.Now())
	f.guideLay = f.id(`insert into guide (name, name_key, type) values ('Ross 308 ouat', 'ross 308 ouat', 'laying')`)
	f.guideGrow = f.id(`insert into guide (name, name_key, type) values ('Ross 308 creștere', 'ross 308 crestere', 'growth')`)
	f.telita = f.id(`insert into farm (name, name_key, type, guide_id) values ('Telița', 'telita', 'growth', $1)`, f.guideGrow)
	f.hirbovat = f.id(`insert into farm (name, name_key, type, guide_id) values ('Hîrbovăț', 'hirbovat', 'laying', $1)`, f.guideLay)
	f.noua = f.id(`insert into farm (name, name_key, type) values ('Ferma Nouă', 'ferma noua', 'laying')`)

	f.section(f.telita, "T1", nil, 1)
	t9 := f.section(f.telita, "T9", nil, 9)
	f.exec(`update section set retired_at = $1 where id = $2`, testdb.Now(), t9)
	h6 := f.section(f.hirbovat, "H6", nil, 2)
	h5 := f.section(f.hirbovat, "H5", nil, 1)
	f.section(f.hirbovat, "H7", nil, 3)
	f.section(f.hirbovat, "H5.2", &h5, 2) // before H5.1: order comes from sort_order
	f.section(f.hirbovat, "H5.1", &h5, 1)
	f.section(f.hirbovat, "H6.1", &h6, 1)
	f.section(f.noua, "N1", nil, 1)

	f.cycT = f.id(`insert into cycle (series_id, farm_id, guide_id, start_date, end_date, started_by, ended_by)
		values ($1, $2, $3, '2025-10-07', '2026-02-28', 'auto', 'emptied')`, f.series26, f.telita, f.guideGrow)
	f.cycleH25 = f.id(`insert into cycle (series_id, farm_id, start_date, end_date, started_by, start_inferred, ended_by)
		values ($1, $2, '2025-03-07', '2026-01-28', 'migration', true, 'migration')`, f.series25, f.hirbovat)
	f.cycleH26 = f.id(`insert into cycle (series_id, farm_id, guide_id, start_date, started_by)
		values ($1, $2, $3, '2026-02-22', 'auto')`, f.series26, f.hirbovat, f.guideLay)

	T, H := f.telita, f.hirbovat
	e := f.entry("placement", T, nil, day(2025, time.October, 7))
	f.bird(e, T, "T1", day(2025, time.October, 7), "F", "placement", 100, "")
	f.bird(e, T, "T1", day(2025, time.October, 7), "M", "placement", 20, "")

	d := day(2026, time.February, 22)
	e = f.entry("transfer", T, &H, d)
	f.bird(e, T, "T1", d, "F", "transfer", -60, "H5.1")
	f.bird(e, H, "H5.1", d, "F", "transfer", 60, "T1")
	f.bird(e, T, "T1", d, "M", "transfer", -10, "H5.1")
	f.bird(e, H, "H5.1", d, "M", "transfer", 10, "T1")
	d = day(2026, time.February, 28)
	e = f.entry("transfer", T, &H, d)
	f.bird(e, T, "T1", d, "F", "transfer", -40, "H5.2")
	f.bird(e, H, "H5.2", d, "F", "transfer", 40, "T1")
	e = f.entry("transfer", T, &H, d)
	f.bird(e, T, "T1", d, "M", "transfer", -10, "H6.1")
	f.bird(e, H, "H6.1", d, "M", "transfer", 10, "T1")

	d = day(2026, time.March, 1)
	e = f.entry("transfer", T, &H, d)
	f.bird(e, T, "T1", d, "F", "transfer", -7, "H5.1")
	f.bird(e, H, "H5.1", d, "F", "transfer", 7, "T1")
	f.exec(`update bird_event set deleted_at = $1 where entry_id = $2`, testdb.Now(), e)
	f.exec(`update entry set deleted_at = $1, deleted_by = $2 where id = $3`, testdb.Now(), f.user, e)

	d = day(2026, time.March, 2)
	e = f.entry("transfer", H, &H, d)
	f.bird(e, H, "H5.1", d, "F", "transfer", -5, "H6.1")
	f.bird(e, H, "H6.1", d, "F", "transfer", 5, "H5.1")

	d = day(2026, time.November, 10)
	e = f.entry("daily", H, nil, d)
	f.bird(e, H, "H5.1", d, "F", "mortality", -2, "")
	d = day(2026, time.November, 12)
	e = f.entry("sale", H, nil, d)
	f.bird(e, H, "H5.2", d, "F", "sale", -5, "")
	return f
}

func (f *fixture) id(sql string, args ...any) uuid.UUID {
	f.t.Helper()
	var id uuid.UUID
	if err := f.d.App.QueryRow(context.Background(), sql+" returning id", args...).Scan(&id); err != nil {
		f.t.Fatalf("%s: %v", sql, err)
	}
	return id
}

func (f *fixture) exec(sql string, args ...any) {
	f.t.Helper()
	if _, err := f.d.App.Exec(context.Background(), sql, args...); err != nil {
		f.t.Fatalf("%s: %v", sql, err)
	}
}

func (f *fixture) section(farm uuid.UUID, code string, parent *uuid.UUID, sort int) uuid.UUID {
	f.t.Helper()
	id := f.id(`insert into section (farm_id, code, parent_id, sort_order) values ($1, $2, $3, $4)`, farm, code, parent, sort)
	f.sec[code] = id
	return id
}

func (f *fixture) entry(kind string, farm uuid.UUID, peerFarm *uuid.UUID, d timex.Day) uuid.UUID {
	f.t.Helper()
	var series *uuid.UUID
	if kind == "placement" {
		series = &f.series26
	}
	return f.id(`insert into entry (kind, farm_id, peer_farm_id, date, series_id, created_by)
		values ($1, $2, $3, $4, $5, $6)`, kind, farm, peerFarm, d, series, f.user)
}

// bird inserts a live bird_event row; peer is a section code or "".
func (f *fixture) bird(entry, farm uuid.UUID, code string, d timex.Day, sex, typ string, qty int, peer string) {
	f.t.Helper()
	var peerID *uuid.UUID
	if peer != "" {
		p := f.sec[peer]
		peerID = &p
	}
	f.id(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty, peer_section_id)
		values ($1, $2, $3, $4, $5, $6, $7, $8)`, entry, farm, f.sec[code], d, sex, typ, qty, peerID)
}

// input is a builder Input for p on farm at the frozen clock.
func (f *fixture) input(p auth.Principal, farm uuid.UUID, cycle *uuid.UUID) Input {
	return Input{Principal: p, FarmID: farm, CycleID: cycle, Today: today, Now: testdb.Now()}
}

func owner() auth.Principal { return auth.Principal{UserID: uuid.New(), Role: auth.Owner} }

func principal(role auth.Role, farms ...uuid.UUID) auth.Principal {
	if farms == nil {
		farms = []uuid.UUID{}
	}
	return auth.Principal{UserID: uuid.New(), Role: role, FarmIDs: farms}
}
