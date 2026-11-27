package farms

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
)

// feedTypes adds two active feed types and a retired one.
func (f *fixture) feedTypes() (breeder, universal uuid.UUID) {
	breeder = f.id(`insert into feed_type (name, name_key, for_sex) values ('Breeder 2', 'breeder 2', 'F')`)
	universal = f.id(`insert into feed_type (name, name_key) values ('Universal', 'universal')`)
	f.id(`insert into feed_type (name, name_key, active) values ('Vechi', 'vechi', false)`)
	return breeder, universal
}

func (f *fixture) hirbovatBootstrap() BootstrapFarm {
	return BootstrapFarm{
		ID: f.hirbovat, Name: "Hîrbovăț", Type: auth.FarmLaying, Guide: &GuideRef{ID: f.guideLay, Name: "Ross 308 ouat"},
		OpenCycle: &OpenCycle{ID: f.cycleH26, SeriesID: f.series26, SeriesName: "Seria 2026",
			StartDate: day(2026, time.February, 22), SaptStart: 20, SaptToday: 59, AgeDaysToday: 407},
		Sections: []SectionNode{
			{ID: f.sec["H5"], Code: "H5", IsHall: true, Children: []SectionLeaf{{ID: f.sec["H5.1"], Code: "H5.1"}, {ID: f.sec["H5.2"], Code: "H5.2"}}},
			{ID: f.sec["H6"], Code: "H6", IsHall: true, Children: []SectionLeaf{{ID: f.sec["H6.1"], Code: "H6.1"}}},
			{ID: f.sec["H7"], Code: "H7", Children: []SectionLeaf{}},
		},
	}
}

func TestBuildBootstrapOwner(t *testing.T) {
	f := newFixture(t)
	breeder, universal := f.feedTypes()
	got, err := BuildBootstrap(t.Context(), f.d.App, f.input(owner(), uuid.Nil, nil))
	if err != nil {
		t.Fatal(err)
	}
	wantFarms := []BootstrapFarm{
		{ID: f.noua, Name: "Ferma Nouă", Type: auth.FarmLaying,
			Sections: []SectionNode{{ID: f.sec["N1"], Code: "N1", Children: []SectionLeaf{}}}},
		f.hirbovatBootstrap(),
		{ID: f.telita, Name: "Telița", Type: auth.FarmGrowth, Guide: &GuideRef{ID: f.guideGrow, Name: "Ross 308 creștere"},
			Sections: []SectionNode{{ID: f.sec["T1"], Code: "T1", Children: []SectionLeaf{}}}}, // T9 is retired
	}
	if !reflect.DeepEqual(got.Farms, wantFarms) {
		t.Errorf("farms:\n got %+v\nwant %+v", got.Farms, wantFarms)
	}
	wantSeries := []SeriesItem{
		{ID: f.series26, Name: "Seria 2026", Status: "active", SaptSpan: [2]int{1, 59}},
		{ID: f.series25, Name: "Seria 2025", Status: "closed", SaptSpan: [2]int{19, 65}},
	}
	if !reflect.DeepEqual(got.Series, wantSeries) {
		t.Errorf("series:\n got %+v\nwant %+v", got.Series, wantSeries)
	}
	f2 := "F"
	wantFeed := []FeedTypeItem{
		{ID: breeder, Name: "Breeder 2", ForSex: &f2, Active: true},
		{ID: universal, Name: "Universal", Active: true},
	}
	if !reflect.DeepEqual(got.FeedTypes, wantFeed) {
		t.Errorf("feed types:\n got %+v\nwant %+v", got.FeedTypes, wantFeed)
	}

	b, err := httpx.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{
		`"settings":{"toleranceWeightPct":5,"toleranceLayPts":3,"toleranceFeedPct":10,"stockDaysMin":5}`,
		`"openCycle":{"id":"` + f.cycleH26.String() + `","seriesId":"` + f.series26.String() + `","seriesName":"Seria 2026","startDate":"2026-02-22","saptStart":20,"saptToday":59,"ageDaysToday":407}`,
		`{"id":"` + f.sec["H7"].String() + `","code":"H7","isHall":false,"children":[]}`,
		`"name":"Ferma Nouă","type":"laying","guide":null,"openCycle":null`,
		`"saptSpan":[1,59]`,
		`"forSex":null`,
	} {
		if !strings.Contains(string(b), s) {
			t.Errorf("JSON lacks %s\n%s", s, b)
		}
	}
}

func TestBuildBootstrapScopedToTheCaller(t *testing.T) {
	f := newFixture(t)
	f.feedTypes()

	// A vet with one farm: that farm only; series by its cycles, span over them.
	got, err := BuildBootstrap(t.Context(), f.d.App, f.input(principal(auth.Vet, f.hirbovat), uuid.Nil, nil))
	if err != nil {
		t.Fatal(err)
	}
	if want := []BootstrapFarm{f.hirbovatBootstrap()}; !reflect.DeepEqual(got.Farms, want) {
		t.Errorf("vet farms:\n got %+v\nwant %+v", got.Farms, want)
	}
	wantSeries := []SeriesItem{
		{ID: f.series26, Name: "Seria 2026", Status: "active", SaptSpan: [2]int{20, 59}}, // not Telița's SAPT 1
		{ID: f.series25, Name: "Seria 2025", Status: "closed", SaptSpan: [2]int{19, 65}},
	}
	if !reflect.DeepEqual(got.Series, wantSeries) {
		t.Errorf("vet series:\n got %+v\nwant %+v", got.Series, wantSeries)
	}
	if len(got.FeedTypes) != 2 {
		t.Errorf("feed types = %+v, want the 2 active ones", got.FeedTypes)
	}

	// A viewer of the empty farm: no open cycle; only the series that was there.
	got, err = BuildBootstrap(t.Context(), f.d.App, f.input(principal(auth.Viewer, f.telita), uuid.Nil, nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Farms) != 1 || got.Farms[0].ID != f.telita || got.Farms[0].OpenCycle != nil {
		t.Errorf("viewer farms = %+v, want Telița without an open cycle", got.Farms)
	}
	if want := []SeriesItem{{ID: f.series26, Name: "Seria 2026", Status: "active", SaptSpan: [2]int{1, 21}}}; !reflect.DeepEqual(got.Series, want) {
		t.Errorf("viewer series:\n got %+v\nwant %+v", got.Series, want)
	}

	// A user without farms gets empty lists, never null.
	got, err = BuildBootstrap(t.Context(), f.d.App, f.input(principal(auth.Manager), uuid.Nil, nil))
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(got)
	if !strings.HasPrefix(string(b), `{"farms":[],"series":[],"feedTypes":[{`) {
		t.Errorf("no farms: %s", b)
	}
}

func TestBuildBootstrapBadSetting(t *testing.T) {
	f := newFixture(t)
	f.exec(`update app_setting set value = '"cinci"' where key = 'tolerance.weight_pct'`)
	if _, err := BuildBootstrap(t.Context(), f.d.App, f.input(owner(), uuid.Nil, nil)); err == nil {
		t.Error("a setting that is not a number must fail the bootstrap")
	}
}
