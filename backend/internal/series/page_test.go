package series

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// fixture is hand-built data for the rules the seed can't show, today
// 18.11.2026:
//
//	Seria S (active, placed 07.10.2025 at age 0); Seria T (closed, placed 01.01.2024); Seria U (active, no cycle).
//	Alfa (growth): cycle S 07.10.2025–28.02.2026 (emptied); placement 07.10.2025 ♀ 100 · ♂ 20.
//	Beta (laying): cycle S from 22.02.2026, open.
//	Gama (laying): cycle S from 20.02.2026, open.
//	Delta (laying): cycle T 01.01.2024–30.06.2024 (ended manually); no cycle of S.
//
//	22.02.2026 transfer Alfa → Beta ♀ 60; 23.02.2026 transfer Alfa → Gama ♀ 30;
//	01.03.2026 transfer Beta → Gama ♀ 5; 01.03.2024 transfer Delta → Beta ♀ 1 (a link of T, not of S).
//
// At the end of 18.11.2026: Beta ♀ 56, Gama ♀ 35.
type fixture struct {
	t                    *testing.T
	d                    *testdb.DB
	user                 uuid.UUID
	s, sT, sU            uuid.UUID
	alfa, beta, gama, de uuid.UUID
	sec                  map[uuid.UUID]uuid.UUID // farm → its one section
}

func day(y int, m time.Month, d int) timex.Day { return timex.MustDay(y, m, d) }

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{t: t, d: testdb.New(t), sec: map[uuid.UUID]uuid.UUID{}}
	f.user = f.id(`insert into app_user (name, login, password_hash, role) values ('Ion', 'ion', 'x', 'owner')`)
	f.s = f.id(`insert into series (name, name_key, year, placement_date) values ('Seria S', 'seria s', 2026, '2025-10-07')`)
	f.sT = f.id(`insert into series (name, name_key, year, placement_date, status, closed_at)
		values ('Seria T', 'seria t', 2024, '2024-01-01', 'closed', $1)`, testdb.Now())
	f.sU = f.id(`insert into series (name, name_key, year, placement_date) values ('Seria U', 'seria u', 2027, '2026-11-01')`)
	f.alfa = f.farm("Alfa", "alfa", "growth")
	f.beta = f.farm("Beta", "beta", "laying")
	f.gama = f.farm("Gama", "gama", "laying")
	f.de = f.farm("Delta", "delta", "laying")

	f.exec(`insert into cycle (series_id, farm_id, start_date, end_date, started_by, ended_by)
		values ($1, $2, '2025-10-07', '2026-02-28', 'auto', 'emptied')`, f.s, f.alfa)
	f.exec(`insert into cycle (series_id, farm_id, start_date, started_by) values ($1, $2, '2026-02-22', 'auto')`, f.s, f.beta)
	f.exec(`insert into cycle (series_id, farm_id, start_date, started_by) values ($1, $2, '2026-02-20', 'auto')`, f.s, f.gama)
	f.exec(`insert into cycle (series_id, farm_id, start_date, end_date, started_by, ended_by)
		values ($1, $2, '2024-01-01', '2024-06-30', 'auto', 'manual')`, f.sT, f.de)

	d := day(2025, time.October, 7)
	e := f.id(`insert into entry (kind, farm_id, date, series_id, created_by) values ('placement', $1, $2, $3, $4)`, f.alfa, d, f.s, f.user)
	f.bird(e, f.alfa, d, "F", "placement", 100)
	f.bird(e, f.alfa, d, "M", "placement", 20)
	f.transfer(f.alfa, f.beta, day(2026, time.February, 22), 60)
	f.transfer(f.alfa, f.gama, day(2026, time.February, 23), 30)
	f.transfer(f.beta, f.gama, day(2026, time.March, 1), 5)
	f.transfer(f.de, f.beta, day(2024, time.March, 1), 1)
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

func (f *fixture) farm(name, key, typ string) uuid.UUID {
	f.t.Helper()
	id := f.id(`insert into farm (name, name_key, type) values ($1, $2, $3)`, name, key, typ)
	f.sec[id] = f.id(`insert into section (farm_id, code) values ($1, 'H1')`, id)
	return id
}

func (f *fixture) bird(entry, farm uuid.UUID, d timex.Day, sex, typ string, qty int) {
	f.t.Helper()
	f.exec(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty) values ($1, $2, $3, $4, $5, $6, $7)`,
		entry, farm, f.sec[farm], d, sex, typ, qty)
}

// transfer moves qty ♀ from one farm's section to another's on d.
func (f *fixture) transfer(from, to uuid.UUID, d timex.Day, qty int) {
	f.t.Helper()
	e := f.id(`insert into entry (kind, farm_id, peer_farm_id, date, created_by) values ('transfer', $1, $2, $3, $4)`, from, to, d, f.user)
	f.exec(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty, peer_section_id)
		values ($1, $2, $3, $4, 'F', 'transfer', $5, $6), ($1, $7, $6, $4, 'F', 'transfer', $8, $3)`,
		e, from, f.sec[from], d, -qty, f.sec[to], to, qty)
}

func (f *fixture) page(p auth.Principal, id uuid.UUID) (PageResponse, error) {
	return BuildPage(f.t.Context(), f.d.App, Input{Principal: p, SeriesID: id, Today: testdb.Today, Now: testdb.Now()})
}

func (f *fixture) mustPage(p auth.Principal, id uuid.UUID) PageResponse {
	f.t.Helper()
	res, err := f.page(p, id)
	if err != nil {
		f.t.Fatal(err)
	}
	return res
}

func owner() auth.Principal { return auth.Principal{UserID: uuid.New(), Role: auth.Owner} }

func viewer(farms ...uuid.UUID) auth.Principal {
	return auth.Principal{UserID: uuid.New(), Role: auth.Viewer, FarmIDs: farms}
}

// link is a lane's transfer link, flattened.
type link struct {
	Lane, From, To string
	F              int
}

func links(res PageResponse) []link {
	var out []link
	for _, l := range res.Track.Lanes {
		for _, tr := range l.Transfers {
			out = append(out, link{l.Farm.Name, tr.FromFarm, tr.ToFarm, tr.F})
		}
	}
	return out
}

func lanes(res PageResponse) []string {
	var out []string
	for _, l := range res.Track.Lanes {
		out = append(out, l.Farm.Name)
	}
	return out
}

func panels(t *testing.T, res PageResponse) []string {
	t.Helper()
	var out []string
	for _, p := range res.Farms {
		switch p := p.(type) {
		case FarmPanel:
			out = append(out, p.Farm.Name)
		case HiddenFarmPanel:
			out = append(out, p.Farm.Name+" (fără acces)")
		default:
			t.Fatalf("panel %T", p)
		}
	}
	return out
}

// TestBuildPageOwner: every farm the owner sees is a lane and a panel,
// growth first, then laying by name; a link is listed once, on its source
// lane, and only when dated inside a cycle of the series; alive sums the
// farms whose open cycle is of the series.
func TestBuildPageOwner(t *testing.T) {
	f := newFixture(t)
	res := f.mustPage(owner(), f.s)

	if want := []string{"Alfa", "Beta", "Delta", "Gama"}; !cmp.Equal(lanes(res), want) {
		t.Errorf("lanes = %v, want %v", lanes(res), want)
	}
	if want := []string{"Alfa", "Beta", "Delta", "Gama"}; !cmp.Equal(panels(t, res), want) {
		t.Errorf("panels = %v, want %v", panels(t, res), want)
	}
	wantLinks := []link{{"Alfa", "Alfa", "Beta", 60}, {"Alfa", "Alfa", "Gama", 30}, {"Beta", "Beta", "Gama", 5}}
	if diff := cmp.Diff(wantLinks, links(res)); diff != "" {
		t.Errorf("links (-want +got):\n%s", diff)
	}
	if res.Alive == nil || *res.Alive != (httpx.SexPair[int]{F: 91, M: 0}) {
		t.Errorf("alive = %v, want Beta 56 + Gama 35 = ♀ 91", res.Alive)
	}
	if res.OwnFarmsOnly || res.SaptToday == nil || *res.SaptToday != 59 || res.SaptSpan != [2]int{1, 59} || res.End != nil {
		t.Errorf("ownFarmsOnly %v, saptToday %v, saptSpan %v, end %v", res.OwnFarmsOnly, res.SaptToday, res.SaptSpan, res.End)
	}
	if p := res.Placement; p == nil || p.Farm != "Alfa" || p.Inferred || p.Date != day(2025, time.October, 7) ||
		p.Placed == nil || *p.Placed != (httpx.SexPair[int]{F: 100, M: 20}) {
		t.Errorf("placement = %+v", p)
	}
	if want := []string{"pasari", "hrana", "greutate", "productie"}; !cmp.Equal(res.Tabs, want) {
		t.Errorf("tabs = %v, want %v", res.Tabs, want)
	}
	delta := res.Farms[2].(FarmPanel)
	if delta.Cycle != nil || delta.Guide != nil || delta.Stats == nil || len(delta.Stats) != 0 || len(res.Track.Lanes[2].Cycles) != 0 {
		t.Errorf("Delta panel = %+v, want no cycle and []", delta)
	}
	b, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"lanes":null`, `"cycles":null`, `"transfers":null`, `"farms":null`, `"stats":null`, `"tabs":null`} {
		if strings.Contains(string(b), key) {
			t.Errorf("%s in %s", key, b)
		}
	}
}

// TestBuildPageOwnFarmsOnly: a viewer sees his farms only; a farm with a
// cycle of the series he can't see is named without access, and a link that
// touches it is not drawn; the placement is on a hidden farm, so omitted.
func TestBuildPageOwnFarmsOnly(t *testing.T) {
	f := newFixture(t)
	res := f.mustPage(viewer(f.beta, f.gama), f.s)

	if want := []string{"Beta", "Gama"}; !cmp.Equal(lanes(res), want) {
		t.Errorf("lanes = %v, want %v", lanes(res), want)
	}
	if want := []string{"Alfa (fără acces)", "Beta", "Gama"}; !cmp.Equal(panels(t, res), want) {
		t.Errorf("panels = %v, want %v", panels(t, res), want)
	}
	if diff := cmp.Diff([]link{{"Beta", "Beta", "Gama", 5}}, links(res)); diff != "" {
		t.Errorf("links (-want +got):\n%s", diff)
	}
	if !res.OwnFarmsOnly || res.Placement != nil || res.SaptSpan != [2]int{20, 59} {
		t.Errorf("ownFarmsOnly %v, placement %+v, saptSpan %v", res.OwnFarmsOnly, res.Placement, res.SaptSpan)
	}
	b, err := json.Marshal(res.Farms[0])
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(b), `{"farm":{"name":"Alfa"},"access":false}`; got != want {
		t.Errorf("hidden panel = %s, want %s", got, want)
	}

	// A viewer of the growth farm only: no Producție, and no alive birds of
	// the series on his farms (its cycle is closed).
	res = f.mustPage(viewer(f.alfa), f.s)
	if want := []string{"pasari", "hrana", "greutate"}; !cmp.Equal(res.Tabs, want) {
		t.Errorf("growth-only tabs = %v, want %v", res.Tabs, want)
	}
	if res.Alive == nil || *res.Alive != (httpx.SexPair[int]{}) || !res.OwnFarmsOnly {
		t.Errorf("alive = %v, ownFarmsOnly %v; want 0 · 0, true", res.Alive, res.OwnFarmsOnly)
	}
}

// TestBuildPagePlacedOnVisibleFarms (M4 COR-1): placement.placed counts only
// the birds placed at farms the caller sees. Beta gets a placement of its
// own in its cycle of S (♀ 9 · ♂ 2), so S has placements at two farms.
func TestBuildPagePlacedOnVisibleFarms(t *testing.T) {
	f := newFixture(t)
	d := day(2026, time.February, 22)
	e := f.id(`insert into entry (kind, farm_id, date, series_id, created_by) values ('placement', $1, $2, $3, $4)`, f.beta, d, f.s, f.user)
	f.bird(e, f.beta, d, "F", "placement", 9)
	f.bird(e, f.beta, d, "M", "placement", 2)

	for _, tc := range []struct {
		name string
		p    auth.Principal
		want httpx.SexPair[int]
	}{
		{"owner: every farm", owner(), httpx.SexPair[int]{F: 109, M: 22}},
		{"a viewer of both farms", viewer(f.alfa, f.beta), httpx.SexPair[int]{F: 109, M: 22}},
		{"a viewer of Alfa only: not Beta's birds", viewer(f.alfa), httpx.SexPair[int]{F: 100, M: 20}},
		{"a viewer of Alfa and Gama: not Beta's birds", viewer(f.alfa, f.gama), httpx.SexPair[int]{F: 100, M: 20}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := f.mustPage(tc.p, f.s).Placement
			if p == nil || p.Farm != "Alfa" || p.Placed == nil || *p.Placed != tc.want {
				t.Errorf("placement = %+v, want Alfa with placed %v", p, tc.want)
			}
		})
	}
	// A viewer of Beta only can't see the placement farm: no placement block,
	// though Beta has placed birds of its own.
	if p := f.mustPage(viewer(f.beta), f.s).Placement; p != nil {
		t.Errorf("viewer of Beta: placement = %+v, want none", p)
	}
}

// TestBuildPageClosedSeries: a closed series has no alive and no SAPT today;
// its end is its last cycle's end and how that cycle ended.
func TestBuildPageClosedSeries(t *testing.T) {
	f := newFixture(t)
	res := f.mustPage(owner(), f.sT)
	if res.Series.Status != "closed" || res.Alive != nil || res.SaptToday != nil {
		t.Errorf("status %s, alive %v, saptToday %v", res.Series.Status, res.Alive, res.SaptToday)
	}
	if res.End == nil || res.End.Date != day(2024, time.June, 30) || res.End.By != EndManual {
		t.Errorf("end = %+v, want 30.06.2024 manual", res.End)
	}
	if res.SaptSpan != [2]int{1, 26} {
		t.Errorf("saptSpan = %v, want [1 26]", res.SaptSpan)
	}
	if p := res.Placement; p == nil || p.Farm != "Delta" || p.Placed != nil {
		t.Errorf("placement = %+v, want Delta without placed birds", p)
	}
	// Delta → Beta leaves Delta's cycle of T: a link of T (dated inside a
	// cycle of the series at either farm), though Beta has no cycle of T.
	if diff := cmp.Diff([]link{{"Delta", "Delta", "Beta", 1}}, links(res)); diff != "" {
		t.Errorf("links (-want +got):\n%s", diff)
	}
}

// TestBuildPageNotFound: an unknown series, a series without a cycle, and a
// series without a cycle on the caller's farms (M4 Q4).
func TestBuildPageNotFound(t *testing.T) {
	f := newFixture(t)
	for _, tc := range []struct {
		name string
		p    auth.Principal
		id   uuid.UUID
	}{
		{"unknown id", owner(), uuid.New()},
		{"no cycle at all", owner(), f.sU},
		{"no cycle on the caller's farms", viewer(f.alfa, f.beta), f.sT},
		{"a caller without farms", viewer(), f.s},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := f.page(tc.p, tc.id); !errors.Is(err, ErrSeriesNotFound) {
				t.Errorf("err = %v, want ErrSeriesNotFound", err)
			}
		})
	}
}
