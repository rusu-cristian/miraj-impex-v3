package periodtable

import (
	"errors"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
)

var (
	qFarm  = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	qCycle = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	qOther = uuid.MustParse("33333333-3333-4333-8333-333333333333")
)

const farmQ = "view=farm&farmId=11111111-1111-4111-8111-111111111111"

func mustValues(t *testing.T, raw string) url.Values {
	t.Helper()
	v, err := url.ParseQuery(raw)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestParseQueryAccepts(t *testing.T) {
	farm, cycle := qFarm, qCycle
	base := TableQuery{View: "farm", FarmID: &farm}
	with := func(f func(q *TableQuery)) TableQuery {
		q := base
		f(&q)
		return q
	}
	for _, tc := range []struct {
		name string
		raw  string
		want TableQuery
	}{
		{"defaults", farmQ, base},
		{"cycleId", farmQ + "&cycleId=" + qCycle.String(), with(func(q *TableQuery) { q.CycleID = &cycle })},
		{"tab rezumat", farmQ + "&tab=rezumat", with(func(q *TableQuery) { q.Tab = "rezumat" })},
		{"period cycle", farmQ + "&period=cycle", with(func(q *TableQuery) { q.Period = Period{Kind: "cycle"} })},
		{"period weeks:4", farmQ + "&period=weeks:4", with(func(q *TableQuery) { q.Period = Period{Kind: "weeks", Weeks: 4} })},
		{"period weeks:12 (load-more widens it)", farmQ + "&period=weeks:12",
			with(func(q *TableQuery) { q.Period = Period{Kind: "weeks", Weeks: 12} })},
		{"period weeks:1", farmQ + "&period=weeks:1", with(func(q *TableQuery) { q.Period = Period{Kind: "weeks", Weeks: 1} })},
		{"period interval", farmQ + "&period=interval:2026-11-01,2026-11-15", with(func(q *TableQuery) {
			q.Period = Period{Kind: "interval", From: day(time.November, 1), To: day(time.November, 15)}
		})},
		{"one-day interval", farmQ + "&period=interval:2026-11-15,2026-11-15", with(func(q *TableQuery) {
			q.Period = Period{Kind: "interval", From: day(time.November, 15), To: day(time.November, 15)}
		})},
		{"detail farm", farmQ + "&detail=farm", with(func(q *TableQuery) { q.Detail = "farm" })},
		{"types, one chip (phone)", farmQ + "&types=oua", with(func(q *TableQuery) { q.Types = []string{"oua"} })},
		{"types, every chip, a repeat dropped", farmQ +
			"&types=mortalitate,sacrificare,eroare_sexare,transfer,populare,vanzare,hrana,cantarire,oua,oua",
			with(func(q *TableQuery) { q.Types = Chips })},
		{"grain day", farmQ + "&grain=day", with(func(q *TableQuery) { q.Grain = "day" })},
		{"grain week", farmQ + "&grain=week", with(func(q *TableQuery) { q.Grain = "week" })},
		{"before and count", farmQ + "&before=56&count=20", with(func(q *TableQuery) { q.Before, q.Count = intp(56), 20 })},
		{"sapt", farmQ + "&sapt=58", with(func(q *TableQuery) { q.Sapt = intp(58) })},
		{"fromSapt with before", farmQ + "&before=56&fromSapt=20", with(func(q *TableQuery) { q.Before, q.FromSapt = intp(56), intp(20) })},
	} {
		t.Run(tc.name, func(t *testing.T) {
			q, err := ParseQuery(mustValues(t, tc.raw))
			if err != nil {
				t.Fatalf("ParseQuery: %v", err)
			}
			if err := httpx.Validate(q); err != nil {
				t.Fatalf("Validate: %v", err)
			}
			diff(t, "query", tc.want, q)
		})
	}
}

func TestParseQueryRefuses(t *testing.T) {
	type issue struct{ Path, Code string }
	for _, tc := range []struct {
		name string
		raw  string
		want []issue
	}{
		// ParseQuery: one parameter at a time.
		{"unknown parameters, in name order", farmQ + "&zz=1&page=2", []issue{{"page", "unknown_field"}, {"zz", "unknown_field"}}},
		{"repeated view", farmQ + "&view=farm", []issue{{"view", "bad_type"}}},
		{"repeated period", farmQ + "&period=cycle&period=weeks:4", []issue{{"period", "bad_type"}}},
		{"malformed farmId", "view=farm&farmId=hirbovat", []issue{{"farmId", "bad_type"}}},
		{"empty farmId", "view=farm&farmId=", []issue{{"farmId", "bad_type"}}},
		{"malformed cycleId", farmQ + "&cycleId=x", []issue{{"cycleId", "bad_type"}}},
		{"malformed sectionId", farmQ + "&sectionId=x", []issue{{"sectionId", "bad_type"}}},
		{"malformed seriesId", farmQ + "&seriesId=x", []issue{{"seriesId", "bad_type"}}},
		{"period kind", farmQ + "&period=month", []issue{{"period", "out_of_range"}}},
		{"period empty", farmQ + "&period=", []issue{{"period", "out_of_range"}}},
		{"period cycle with an argument", farmQ + "&period=cycle:2", []issue{{"period", "bad_type"}}},
		{"weeks without n", farmQ + "&period=weeks", []issue{{"period", "bad_type"}}},
		{"weeks not an integer", farmQ + "&period=weeks:four", []issue{{"period", "bad_type"}}},
		{"weeks 0", farmQ + "&period=weeks:0", []issue{{"period", "out_of_range"}}},
		{"weeks negative", farmQ + "&period=weeks:-4", []issue{{"period", "out_of_range"}}},
		{"weeks beyond int32", farmQ + "&period=weeks:99999999999", []issue{{"period", "out_of_range"}}},
		{"interval without days", farmQ + "&period=interval", []issue{{"period", "bad_type"}}},
		{"interval with one day", farmQ + "&period=interval:2026-11-01", []issue{{"period", "bad_type"}}},
		{"interval, not a date", farmQ + "&period=interval:2026-02-30,2026-03-01", []issue{{"period", "bad_date"}}},
		{"interval, end not a date", farmQ + "&period=interval:2026-11-01,azi", []issue{{"period", "bad_date"}}},
		{"interval backwards", farmQ + "&period=interval:2026-11-15,2026-11-01", []issue{{"period", "out_of_range"}}},
		{"types, unknown chip", farmQ + "&types=oua,ouat", []issue{{"types", "out_of_range"}}},
		{"types, empty", farmQ + "&types=", []issue{{"types", "out_of_range"}}},
		{"before not an integer", farmQ + "&before=x", []issue{{"before", "bad_type"}}},
		{"before 0", farmQ + "&before=0", []issue{{"before", "out_of_range"}}},
		{"count 0", farmQ + "&count=0", []issue{{"count", "out_of_range"}}},
		{"count decimal", farmQ + "&count=2.5", []issue{{"count", "bad_type"}}},
		{"sapt not an integer", farmQ + "&sapt=SAPT58", []issue{{"sapt", "bad_type"}}},
		{"fromSapt negative", farmQ + "&fromSapt=-1", []issue{{"fromSapt", "out_of_range"}}},
		{"several, unknown first then §7.4 order", farmQ + "&x=1&count=0&before=y",
			[]issue{{"x", "unknown_field"}, {"before", "bad_type"}, {"count", "out_of_range"}}},

		// Validate: what M2 serves (M2 D2) and how parameters combine.
		{"no view", "farmId=" + qFarm.String(), []issue{{"view", "required"}}},
		{"view section (M6)", "view=section&sectionId=" + qOther.String(), []issue{{"view", "out_of_range"}}},
		{"view series (M6)", "view=series&seriesId=" + qOther.String(), []issue{{"view", "out_of_range"}}},
		{"view report-production (M6)", "view=report-production&farmId=" + qFarm.String(), []issue{{"view", "out_of_range"}}},
		{"unknown view", "view=ferma", []issue{{"view", "out_of_range"}}},
		{"no farmId", "view=farm", []issue{{"farmId", "required"}}},
		{"sectionId on the farm view", farmQ + "&sectionId=" + qOther.String(), []issue{{"sectionId", "out_of_range"}}},
		{"seriesId on the farm view", farmQ + "&seriesId=" + qOther.String(), []issue{{"seriesId", "out_of_range"}}},
		{"tab pasari (M6)", farmQ + "&tab=pasari", []issue{{"tab", "out_of_range"}}},
		{"unknown tab", farmQ + "&tab=x", []issue{{"tab", "out_of_range"}}},
		{"detail halls on rezumat", farmQ + "&detail=halls", []issue{{"detail", "out_of_range"}}},
		{"detail sections on rezumat", farmQ + "&detail=sections", []issue{{"detail", "out_of_range"}}},
		{"unknown grain", farmQ + "&grain=month", []issue{{"grain", "out_of_range"}}},
		{"sapt with before", farmQ + "&sapt=55&before=56", []issue{{"sapt", "out_of_range"}}},
		{"sapt with fromSapt", farmQ + "&sapt=55&fromSapt=20", []issue{{"sapt", "out_of_range"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			q, err := ParseQuery(mustValues(t, tc.raw))
			if err == nil {
				err = httpx.Validate(q)
			}
			var he *httpx.Error
			if !errors.As(err, &he) {
				t.Fatalf("err = %v, want a VALIDATION error", err)
			}
			if he.Status != 400 || he.Code != httpx.CodeValidation || he.Message != "Datele trimise nu sunt valide." {
				t.Errorf("error = %d %s %q", he.Status, he.Code, he.Message)
			}
			got := make([]issue, 0, len(he.Issues))
			for _, i := range he.Issues {
				got = append(got, issue{i.Path, i.Code})
				if i.Message != "Datele trimise nu sunt valide." {
					t.Errorf("issue %s message %q", i.Path, i.Message)
				}
			}
			diff(t, "issues", tc.want, got)
		})
	}
}

func TestBuildError(t *testing.T) {
	q := TableQuery{Period: Period{Kind: "interval", From: day(time.November, 1), To: day(time.November, 20)}}
	for _, tc := range []struct {
		name                  string
		err                   error
		status                int
		code, path, issueCode string
		message               string
	}{
		{"cycle of another farm", fmt.Errorf("x: %w", calc.ErrNoCycle), 404, "NOT_FOUND", "", "", "Adresa cerută nu există."},
		{"interval in the future", calc.ErrIntervalFuture, 400, "VALIDATION", "period", "out_of_range",
			"Până la 20.11.2026 e în viitor. Alege o dată până la 18.11.2026."},
		{"period", fmt.Errorf("%w: weeks:0", calc.ErrPeriod), 400, "VALIDATION", "period", "out_of_range", "Datele trimise nu sunt valide."},
		{"paging", &calc.PageError{Param: "fromSapt"}, 400, "VALIDATION", "fromSapt", "out_of_range", "Datele trimise nu sunt valide."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var he *httpx.Error
			if !errors.As(buildError(tc.err, q, day(time.November, 18)), &he) {
				t.Fatalf("not an envelope error")
			}
			if he.Status != tc.status || he.Code != tc.code || he.Message != tc.message {
				t.Errorf("error = %d %s %q", he.Status, he.Code, he.Message)
			}
			if tc.path != "" && (len(he.Issues) != 1 || he.Issues[0].Path != tc.path || he.Issues[0].Code != tc.issueCode) {
				t.Errorf("issues = %+v", he.Issues)
			}
		})
	}
	other := errors.New("db down")
	if err := buildError(other, q, day(time.November, 18)); err != other {
		t.Errorf("an unexpected error = %v, want it unchanged (500)", err)
	}
}
