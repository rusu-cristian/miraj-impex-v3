package farms

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

func dayPtr(d timex.Day) *timex.Day { return &d }

// hirbovatLane is Hîrbovăț's farm track: both cycles oldest first, the two
// Telița links (the deleted entry and the move inside the farm are not links).
func (f *fixture) hirbovatLane() TrackLane {
	return TrackLane{
		Farm: FarmRef{ID: f.hirbovat, Name: "Hîrbovăț"},
		Cycles: []TrackCycle{
			{ID: f.cycleH25, Series: "Seria 2025", StartDate: day(2025, time.March, 7), EndDate: dayPtr(day(2026, time.January, 28)),
				StartInferred: true, SaptSpan: [2]int{19, 65}},
			{ID: f.cycleH26, Series: "Seria 2026", StartDate: day(2026, time.February, 22), Open: true, SaptSpan: [2]int{20, 59}},
		},
		Transfers: f.links(),
	}
}

func (f *fixture) links() []TrackTransfer {
	return []TrackTransfer{
		{Date: day(2026, time.February, 22), FromFarm: "Telița", ToFarm: "Hîrbovăț", F: 60, M: 10},
		{Date: day(2026, time.February, 28), FromFarm: "Telița", ToFarm: "Hîrbovăț", F: 40, M: 10},
	}
}

func (f *fixture) hirbovatCycles() []CycleItem {
	return []CycleItem{
		{ID: f.cycleH26, Series: "Seria 2026", Label: "22.02.2026–azi · SAPT 20–59", Open: true},
		{ID: f.cycleH25, Series: "Seria 2025", Label: "07.03.2025–28.01.2026 · SAPT 19–65", StartInferred: true},
	}
}

func TestBuildPageLayingFarmOpenCycle(t *testing.T) {
	f := newFixture(t)
	got, err := BuildPage(t.Context(), f.d.App, f.input(principal(auth.Manager, f.hirbovat), f.hirbovat, nil))
	if err != nil {
		t.Fatal(err)
	}
	want := PageResponse{
		Farm: PageFarm{ID: f.hirbovat, Name: "Hîrbovăț", Type: auth.FarmLaying, Guide: &GuideRef{ID: f.guideLay, Name: "Ross 308 ouat"}},
		Cycle: &PageCycle{ID: f.cycleH26, Series: "Seria 2026", StartDate: day(2026, time.February, 22),
			SaptStart: 20, SaptToday: 59, AgeDaysToday: 407, Weeks: 40},
		Alive:   &httpx.SexPair[int]{F: 93, M: 20},
		Cycles:  f.hirbovatCycles(),
		Track:   Track{Lanes: []TrackLane{f.hirbovatLane()}, Today: today},
		Actions: Actions{Move: true, Sell: true, NewCycle: true, ChangeGuide: true},
		Tabs:    []string{"rezumat", "pasari", "hrana", "greutate", "productie"},
		AsOf:    timex.At(testdb.Now()),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("page:\n got %+v\nwant %+v", got, want)
	}

	// The §7.3 example's wire shape.
	b, err := httpx.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{
		`"cycle":{"id":"` + f.cycleH26.String() + `","series":"Seria 2026","startDate":"2026-02-22","endDate":null,"saptStart":20,"saptToday":59,"ageDaysToday":407,"weeks":40}`,
		`"alive":{"F":93,"M":20}`,
		`"label":"22.02.2026–azi · SAPT 20–59","open":true}`,
		`"label":"07.03.2025–28.01.2026 · SAPT 19–65","open":false,"startInferred":true}`,
		`"actions":{"move":true,"sell":true,"newCycle":true,"changeGuide":true}`,
		`"asOf":"2026-11-18T12:00:00+02:00"`,
		`{"date":"2026-02-22","fromFarm":"Telița","toFarm":"Hîrbovăț","F":60,"M":10}`,
	} {
		if !strings.Contains(string(b), s) {
			t.Errorf("JSON lacks %s\n%s", s, b)
		}
	}
	if strings.Contains(string(b), `"empty"`) {
		t.Errorf("a farm with an open cycle has no empty state: %s", b)
	}

	// A viewer sees the same page with no action.
	got, err = BuildPage(t.Context(), f.d.App, f.input(principal(auth.Viewer, f.hirbovat), f.hirbovat, nil))
	if err != nil {
		t.Fatal(err)
	}
	if got.Actions != (Actions{}) {
		t.Errorf("viewer actions = %+v, want none", got.Actions)
	}
}

func TestBuildPagePastCycle(t *testing.T) {
	f := newFixture(t)
	got, err := BuildPage(t.Context(), f.d.App, f.input(owner(), f.hirbovat, &f.cycleH25))
	if err != nil {
		t.Fatal(err)
	}
	want := PageResponse{
		// The Seria 2025 cycle has no guide: the header shows "fără ghid".
		Farm: PageFarm{ID: f.hirbovat, Name: "Hîrbovăț", Type: auth.FarmLaying},
		Cycle: &PageCycle{ID: f.cycleH25, Series: "Seria 2025", StartDate: day(2025, time.March, 7),
			EndDate: dayPtr(day(2026, time.January, 28)), SaptStart: 19, SaptToday: 65, AgeDaysToday: 453, Weeks: 47},
		Alive:   nil, // no flock count
		Cycles:  f.hirbovatCycles(),
		Track:   Track{Lanes: []TrackLane{f.hirbovatLane()}, Today: today},
		Actions: Actions{Move: true, Sell: true, NewCycle: true, ChangeGuide: true}, // farm actions, as today
		Tabs:    []string{"rezumat", "pasari", "hrana", "greutate", "productie"},
		AsOf:    timex.At(testdb.Now()),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("page:\n got %+v\nwant %+v", got, want)
	}
	b, _ := json.Marshal(got)
	if !strings.Contains(string(b), `"alive":null`) || !strings.Contains(string(b), `"guide":null`) {
		t.Errorf("alive and guide must be null: %s", b)
	}
}

func TestBuildPageEmptyFarm(t *testing.T) {
	f := newFixture(t)
	got, err := BuildPage(t.Context(), f.d.App, f.input(principal(auth.Manager, f.telita), f.telita, nil))
	if err != nil {
		t.Fatal(err)
	}
	want := PageResponse{
		Farm: PageFarm{ID: f.telita, Name: "Telița", Type: auth.FarmGrowth, Guide: &GuideRef{ID: f.guideGrow, Name: "Ross 308 creștere"}},
		Cycle: &PageCycle{ID: f.cycT, Series: "Seria 2026", StartDate: day(2025, time.October, 7),
			EndDate: dayPtr(day(2026, time.February, 28)), SaptStart: 1, SaptToday: 21, AgeDaysToday: 144, Weeks: 21},
		Alive:  &httpx.SexPair[int]{F: 0, M: 0},
		Cycles: []CycleItem{{ID: f.cycT, Series: "Seria 2026", Label: "07.10.2025–28.02.2026 · SAPT 1–21"}},
		Track: Track{Lanes: []TrackLane{{
			Farm: FarmRef{ID: f.telita, Name: "Telița"},
			Cycles: []TrackCycle{{ID: f.cycT, Series: "Seria 2026", StartDate: day(2025, time.October, 7),
				EndDate: dayPtr(day(2026, time.February, 28)), SaptSpan: [2]int{1, 21}}},
			Transfers: f.links(),
		}}, Today: today},
		Actions: Actions{Move: true, NewCycle: true}, // no birds today, no open cycle
		Tabs:    []string{"rezumat", "pasari", "hrana", "greutate"},
		Empty: &EmptyState{
			Since: dayPtr(day(2026, time.February, 28)),
			Label: "Ferma e goală din 28.02.2026. Ultimul ciclu: Seria 2026, SAPT 1–21.",
			Hint:  "Ciclul pornește la prima populare sau transfer.",
		},
		AsOf: timex.At(testdb.Now()),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("page:\n got %+v\nwant %+v", got, want)
	}
	b, _ := json.Marshal(got.Empty)
	if string(b) != `{"since":"2026-02-28","label":"Ferma e goală din 28.02.2026. Ultimul ciclu: Seria 2026, SAPT 1–21.","hint":"Ciclul pornește la prima populare sau transfer."}` {
		t.Errorf("empty = %s", b)
	}
}

func TestBuildPageFarmWithoutCycle(t *testing.T) {
	f := newFixture(t)
	got, err := BuildPage(t.Context(), f.d.App, f.input(owner(), f.noua, nil))
	if err != nil {
		t.Fatal(err)
	}
	b, err := httpx.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"farm":{"id":"` + f.noua.String() + `","name":"Ferma Nouă","type":"laying","guide":null},` +
		`"cycle":null,"alive":null,"cycles":[],` +
		`"track":{"lanes":[{"farm":{"id":"` + f.noua.String() + `","name":"Ferma Nouă"},"cycles":[],"transfers":[]}],"today":"2026-11-18"},` +
		`"actions":{"move":true,"sell":false,"newCycle":true,"changeGuide":false},"tabs":[],` +
		`"empty":{"label":"Ferma nu are încă niciun ciclu.","hint":"Ciclul pornește la prima populare sau transfer."},` +
		`"asOf":"2026-11-18T12:00:00+02:00"}` + "\n"
	if string(b) != want {
		t.Errorf("page JSON:\n got %s\nwant %s", b, want)
	}
}

func TestBuildPageNotFound(t *testing.T) {
	f := newFixture(t)
	for _, tc := range []struct {
		name  string
		farm  uuid.UUID
		cycle *uuid.UUID
		want  error
	}{
		{"unknown farm", uuid.New(), nil, ErrFarmNotFound},
		{"another farm's cycle", f.hirbovat, &f.cycT, ErrCycleNotFound},
		{"unknown cycle", f.hirbovat, new(uuid.New()), ErrCycleNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := BuildPage(t.Context(), f.d.App, f.input(owner(), tc.farm, tc.cycle)); !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestActionsFor(t *testing.T) {
	all := Actions{Move: true, Sell: true, NewCycle: true, ChangeGuide: true}
	for _, tc := range []struct {
		name             string
		role             auth.Role
		birds, openCycle bool
		want             Actions
	}{
		{"owner, birds, open cycle", auth.Owner, true, true, all},
		{"manager, birds, open cycle", auth.Manager, true, true, all},
		{"manager, sold out today", auth.Manager, false, true, Actions{Move: true, NewCycle: true, ChangeGuide: true}},
		{"manager, empty farm", auth.Manager, false, false, Actions{Move: true, NewCycle: true}},
		{"viewer", auth.Viewer, true, true, Actions{}},
		{"vet", auth.Vet, true, true, Actions{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := actionsFor(tc.role, tc.birds, tc.openCycle); got != tc.want {
				t.Errorf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}
