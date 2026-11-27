package calc

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

var (
	today18 = day(2026, time.November, 18)
	// Hîrbovăț, Seria 2026: open since 22.02.2026 (SAPT 20).
	hirbovat2026 = Cycle{StartDate: day(2026, time.February, 22), Age: seria2026}
	// Telița, Seria 2026: 07.10.2025–28.02.2026 (SAPT 1–21), closed.
	telitaEnd  = day(2026, time.February, 28)
	telita2026 = Cycle{StartDate: day(2025, time.October, 7), EndDate: &telitaEnd, Age: seria2026}
)

func intp(i int) *int { return &i }

func resolve(t *testing.T, c Cycle, p Period) ResolvedPeriod {
	t.Helper()
	r, err := ResolvePeriod(c, p, today18)
	if err != nil {
		t.Fatalf("ResolvePeriod(%+v): %v", p, err)
	}
	return r
}

// span is a WeekSpan as the tests compare it.
type span struct {
	Sapt      int
	From, To  timex.Day
	DaysInRow int
	Partial   bool
	Label     string
}

func spans(ws []WeekSpan) []span {
	out := make([]span, 0, len(ws))
	for _, w := range ws {
		out = append(out, span{w.Sapt, w.From, w.To, w.DaysInRow(), w.Partial(), w.Label()})
	}
	return out
}

func TestResolvePeriod(t *testing.T) {
	nov := func(d int) timex.Day { return day(2026, time.November, d) }
	for _, tc := range []struct {
		name                       string
		c                          Cycle
		p                          Period
		label                      string
		from, to                   timex.Day
		saptFrom, saptTo, total    int
		cycleFrom, cycleTo, cycleN int
	}{
		{"cycle on the open cycle", hirbovat2026, Period{Kind: PeriodCycle}, "Ciclul curent",
			day(2026, time.February, 22), today18, 20, 59, 40, 20, 59, 40},
		{"weeks:4 = SAPT 56–59", hirbovat2026, Period{Kind: PeriodWeeks, Weeks: 4}, "4 săpt.",
			day(2026, time.October, 27), today18, 56, 59, 4, 20, 59, 40},
		{"weeks:13", hirbovat2026, Period{Kind: PeriodWeeks, Weeks: 13}, "13 săpt.",
			day(2026, time.August, 25), today18, 47, 59, 13, 20, 59, 40},
		// 01.11 is in SAPT 56 (27.10–02.11) and 15.11 in SAPT 58, so the
		// interval touches SAPT 56–58 (§5.2 "weeks touched").
		{"interval 01.11–15.11", hirbovat2026, Period{Kind: PeriodInterval, From: nov(1), To: nov(15)}, "01.11–15.11.2026",
			nov(1), nov(15), 56, 58, 3, 20, 59, 40},
		{"interval clipped to the cycle start", hirbovat2026,
			Period{Kind: PeriodInterval, From: day(2026, time.February, 1), To: day(2026, time.February, 24)}, "01.02–24.02.2026",
			day(2026, time.February, 22), day(2026, time.February, 24), 20, 21, 2, 20, 59, 40},
		{"past cycle: Tot ciclul", telita2026, Period{Kind: PeriodCycle}, "Tot ciclul",
			day(2025, time.October, 7), telitaEnd, 1, 21, 21, 1, 21, 21},
		{"weeks:4 on a past cycle ends at its last day", telita2026, Period{Kind: PeriodWeeks, Weeks: 4}, "4 săpt.",
			day(2026, time.February, 3), telitaEnd, 18, 21, 4, 1, 21, 21},
		{"weeks:13 longer than the cycle", Cycle{StartDate: nov(10), Age: seria2026}, Period{Kind: PeriodWeeks, Weeks: 13}, "13 săpt.",
			nov(10), today18, 58, 59, 2, 58, 59, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := resolve(t, tc.c, tc.p)
			if r.Empty {
				t.Fatal("Empty = true")
			}
			got := []any{r.Label, r.From, r.To, r.SaptFrom, r.SaptTo, r.WeeksTotal(), r.CycleSaptFrom, r.CycleSaptTo, r.CycleWeeks()}
			want := []any{tc.label, tc.from, tc.to, tc.saptFrom, tc.saptTo, tc.total, tc.cycleFrom, tc.cycleTo, tc.cycleN}
			if diff := cmp.Diff(want, got, cmp.AllowUnexported(timex.Day{})); diff != "" {
				t.Errorf("label, from, to, saptFrom, saptTo, weeks, cycle span (-want +got):\n%s", diff)
			}
			if r.Kind != tc.p.Kind {
				t.Errorf("Kind = %q, want %q", r.Kind, tc.p.Kind)
			}
		})
	}
}

func TestPartialWeeks(t *testing.T) {
	r := resolve(t, hirbovat2026, Period{Kind: PeriodCycle})
	nov := func(d int) timex.Day { return day(2026, time.November, d) }
	feb := func(d int) timex.Day { return day(2026, time.February, d) }
	want := []span{
		{59, nov(17), nov(18), 2, true, "SAPT 59 · 17.11–18.11 · 2 din 7 zile"},
		{58, nov(10), nov(16), 7, false, "SAPT 58 · 10.11–16.11"},
	}
	if diff := cmp.Diff(want, spans(r.WeeksNewestFirst(58, 59)), cmp.AllowUnexported(timex.Day{})); diff != "" {
		t.Errorf("SAPT 58–59 (-want +got):\n%s", diff)
	}
	want = []span{
		{21, feb(24), day(2026, time.March, 2), 7, false, "SAPT 21 · 24.02–02.03"},
		{20, feb(22), feb(23), 2, true, "SAPT 20 · 22.02–23.02 · 2 din 7 zile"},
	}
	if diff := cmp.Diff(want, spans(r.WeeksNewestFirst(20, 21)), cmp.AllowUnexported(timex.Day{})); diff != "" {
		t.Errorf("SAPT 20–21 (-want +got):\n%s", diff)
	}

	// The interval clips both of its end weeks.
	ri := resolve(t, hirbovat2026, Period{Kind: PeriodInterval, From: nov(1), To: nov(15)})
	want = []span{
		{58, nov(10), nov(15), 6, true, "SAPT 58 · 10.11–15.11 · 6 din 7 zile"},
		{57, nov(3), nov(9), 7, false, "SAPT 57 · 03.11–09.11"},
		{56, nov(1), nov(2), 2, true, "SAPT 56 · 01.11–02.11 · 2 din 7 zile"},
	}
	if diff := cmp.Diff(want, spans(ri.WeeksNewestFirst(ri.SaptFrom, ri.SaptTo)), cmp.AllowUnexported(timex.Day{})); diff != "" {
		t.Errorf("interval weeks (-want +got):\n%s", diff)
	}

	// A one-day row and the closed cycle's last week.
	r1 := resolve(t, Cycle{StartDate: nov(17), Age: seria2026}, Period{Kind: PeriodInterval, From: nov(17), To: nov(17)})
	if got := spans(r1.WeeksNewestFirst(59, 59)); len(got) != 1 || got[0].Label != "SAPT 59 · 17.11 · 1 din 7 zile" || got[0].DaysInRow != 1 {
		t.Errorf("one-day row = %+v", got)
	}
	rt := resolve(t, telita2026, Period{Kind: PeriodCycle})
	if w, _ := rt.Week(21); w.DaysInRow() != 5 || w.Label() != "SAPT 21 · 24.02–28.02 · 5 din 7 zile" {
		t.Errorf("Telița SAPT 21 = %d days, %q", w.DaysInRow(), w.Label())
	}
	if _, ok := rt.Week(22); ok {
		t.Error("Week(22) of a cycle ending in SAPT 21 is ok")
	}
	if w, _ := r.Week(40); !w.Contains(w.From) || !w.Contains(w.To) || w.Contains(w.To.AddDays(1)) {
		t.Error("Contains is wrong at the row's edges")
	}
}

func TestResolvePeriodEmptyAndErrors(t *testing.T) {
	// An interval before the cycle (S4 "fără date în perioadă" board).
	from, to := day(2025, time.November, 1), day(2025, time.November, 15)
	r := resolve(t, hirbovat2026, Period{Kind: PeriodInterval, From: from, To: to})
	if !r.Empty || r.WeeksTotal() != 0 || r.From != from || r.To != to || r.Label != "01.11–15.11.2025" {
		t.Errorf("interval before the cycle = %+v, want Empty with the asked days", r)
	}
	if ws := r.WeeksNewestFirst(r.SaptFrom, r.SaptTo); len(ws) != 0 {
		t.Errorf("Empty period has weeks %v", ws)
	}
	// A cycle that starts after today.
	if r := resolve(t, Cycle{StartDate: today18.AddDays(1), Age: seria2026}, Period{Kind: PeriodCycle}); !r.Empty {
		t.Errorf("cycle starting tomorrow = %+v, want Empty", r)
	}

	for _, tc := range []struct {
		name string
		p    Period
		want error
	}{
		{"interval ends tomorrow", Period{Kind: PeriodInterval, From: today18, To: today18.AddDays(1)}, ErrIntervalFuture},
		{"interval ends today", Period{Kind: PeriodInterval, From: today18, To: today18}, nil},
		{"interval backwards", Period{Kind: PeriodInterval, From: today18, To: today18.AddDays(-1)}, ErrPeriod},
		{"interval without days", Period{Kind: PeriodInterval}, ErrPeriod},
		{"weeks:0", Period{Kind: PeriodWeeks}, ErrPeriod},
		{"unknown kind", Period{Kind: "month"}, ErrPeriod},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ResolvePeriod(hirbovat2026, tc.p, today18); !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

// page is a Page as the tests compare it.
type page struct {
	Weeks       []int
	Loaded      *PageLoaded
	More        *PageMore
	Start       string
	TotalsLabel string
}

func paginate(t *testing.T, r ResolvedPeriod, req PageRequest) page {
	t.Helper()
	p, err := Paginate(r, req)
	if err != nil {
		t.Fatalf("Paginate(%+v): %v", req, err)
	}
	weeks := []int{}
	for _, w := range p.Weeks {
		weeks = append(weeks, w.Sapt)
	}
	return page{weeks, p.Loaded, p.More, p.Start, p.TotalsLabel}
}

func TestPaginate(t *testing.T) {
	cycle := resolve(t, hirbovat2026, Period{Kind: PeriodCycle})
	weeks4 := resolve(t, hirbovat2026, Period{Kind: PeriodWeeks, Weeks: 4})
	weeks8 := resolve(t, hirbovat2026, Period{Kind: PeriodWeeks, Weeks: 8})
	loaded := func(from int) *PageLoaded {
		w, _ := cycle.Week(from)
		return &PageLoaded{SaptFrom: from, SaptTo: 59, Weeks: 60 - from, From: w.From, To: today18}
	}
	for _, tc := range []struct {
		name string
		r    ResolvedPeriod
		req  PageRequest
		want page
	}{
		{"default: the newest 4 weeks", cycle, PageRequest{}, page{
			Weeks: []int{59, 58, 57, 56}, Loaded: loaded(56),
			More:        &PageMore{Before: 56, Remaining: 36, Label: "Încarcă încă 4 săptămâni · SAPT 56–59 · 4 din 40"},
			TotalsLabel: "Total SAPT 56–59",
		}},
		{"before=56: SAPT 52–55, totals over 52–59", cycle, PageRequest{Before: intp(56)}, page{
			Weeks: []int{55, 54, 53, 52}, Loaded: loaded(52),
			More:        &PageMore{Before: 52, Remaining: 32, Label: "Încarcă încă 4 săptămâni · SAPT 52–59 · 8 din 40"},
			TotalsLabel: "Total SAPT 52–59",
		}},
		{"before=24: the first weeks, then the start note", cycle, PageRequest{Before: intp(24)}, page{
			Weeks: []int{23, 22, 21, 20}, Loaded: loaded(20),
			Start: "Începutul ciclului · SAPT 20", TotalsLabel: "Total SAPT 20–59",
		}},
		{"before=23: fewer than 4 left on this page", cycle, PageRequest{Before: intp(23)}, page{
			Weeks: []int{22, 21, 20}, Loaded: loaded(20),
			Start: "Începutul ciclului · SAPT 20", TotalsLabel: "Total SAPT 20–59",
		}},
		{"3 weeks left", cycle, PageRequest{Before: intp(27)}, page{
			Weeks: []int{26, 25, 24, 23}, Loaded: loaded(23),
			More:        &PageMore{Before: 23, Remaining: 3, Label: "Încarcă încă 3 săptămâni · SAPT 23–59 · 37 din 40"},
			TotalsLabel: "Total SAPT 23–59",
		}},
		{"1 week left", cycle, PageRequest{Before: intp(25)}, page{
			Weeks: []int{24, 23, 22, 21}, Loaded: loaded(21),
			More:        &PageMore{Before: 21, Remaining: 1, Label: "Încarcă încă o săptămână · SAPT 21–59 · 39 din 40"},
			TotalsLabel: "Total SAPT 21–59",
		}},
		{"count=20", cycle, PageRequest{Count: 20}, page{
			Weeks: []int{59, 58, 57, 56, 55, 54, 53, 52, 51, 50, 49, 48, 47, 46, 45, 44, 43, 42, 41, 40}, Loaded: loaded(40),
			More:        &PageMore{Before: 40, Remaining: 20, Label: "Încarcă încă 20 de săptămâni · SAPT 40–59 · 20 din 40"},
			TotalsLabel: "Total SAPT 40–59",
		}},
		{"sapt=58: that week only", cycle, PageRequest{Sapt: intp(58)}, page{
			Weeks:       []int{58},
			Loaded:      &PageLoaded{SaptFrom: 58, SaptTo: 58, Weeks: 1, From: day(2026, time.November, 10), To: day(2026, time.November, 16)},
			TotalsLabel: "Total SAPT 58",
		}},
		{"fromSapt=50&before=56", cycle, PageRequest{FromSapt: intp(50), Before: intp(56)}, page{
			Weeks: []int{55, 54, 53, 52, 51, 50}, Loaded: loaded(50),
			More:        &PageMore{Before: 50, Remaining: 30, Label: "Încarcă încă 4 săptămâni · SAPT 50–59 · 10 din 40"},
			TotalsLabel: "Total SAPT 50–59",
		}},
		{"fromSapt before the cycle is clipped to its start", cycle, PageRequest{FromSapt: intp(1), Before: intp(22)}, page{
			Weeks: []int{21, 20}, Loaded: loaded(20),
			Start: "Începutul ciclului · SAPT 20", TotalsLabel: "Total SAPT 20–59",
		}},
		// A set-length period still offers load-more back to the cycle start
		// (D192): the client widens the period and sends before.
		{"weeks:4 still counts the cycle's weeks", weeks4, PageRequest{}, page{
			Weeks: []int{59, 58, 57, 56}, Loaded: loaded(56),
			More:        &PageMore{Before: 56, Remaining: 36, Label: "Încarcă încă 4 săptămâni · SAPT 56–59 · 4 din 40"},
			TotalsLabel: "Total SAPT 56–59",
		}},
		{"weeks:8 with before=56", weeks8, PageRequest{Before: intp(56)}, page{
			Weeks: []int{55, 54, 53, 52}, Loaded: loaded(52),
			More:        &PageMore{Before: 52, Remaining: 32, Label: "Încarcă încă 4 săptămâni · SAPT 52–59 · 8 din 40"},
			TotalsLabel: "Total SAPT 52–59",
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if diff := cmp.Diff(tc.want, paginate(t, tc.r, tc.req), cmp.AllowUnexported(timex.Day{})); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestPaginateClipsDays(t *testing.T) {
	r := resolve(t, hirbovat2026, Period{Kind: PeriodCycle})
	p, err := Paginate(r, PageRequest{Before: intp(24)})
	if err != nil {
		t.Fatal(err)
	}
	first := p.Weeks[len(p.Weeks)-1]
	if first.Sapt != 20 || first.From != day(2026, time.February, 22) || first.DaysInRow() != 2 {
		t.Errorf("oldest week = %+v, want SAPT 20 from 22.02, 2 days", first)
	}
	if p.Loaded.From != day(2026, time.February, 22) || p.Loaded.To != today18 {
		t.Errorf("loaded days = %s–%s", p.Loaded.From, p.Loaded.To)
	}
}

func TestPaginateErrors(t *testing.T) {
	cycle := resolve(t, hirbovat2026, Period{Kind: PeriodCycle})
	weeks4 := resolve(t, hirbovat2026, Period{Kind: PeriodWeeks, Weeks: 4})
	for _, tc := range []struct {
		name  string
		r     ResolvedPeriod
		req   PageRequest
		param string
	}{
		{"before at the first week", cycle, PageRequest{Before: intp(20)}, "before"},
		{"before after the newest + 1", cycle, PageRequest{Before: intp(61)}, "before"},
		{"before outside a weeks:4 period", weeks4, PageRequest{Before: intp(56)}, "before"},
		{"sapt before the period", cycle, PageRequest{Sapt: intp(19)}, "sapt"},
		{"sapt after the period", cycle, PageRequest{Sapt: intp(60)}, "sapt"},
		{"sapt with before", cycle, PageRequest{Sapt: intp(58), Before: intp(59)}, "sapt"},
		{"sapt with fromSapt", cycle, PageRequest{Sapt: intp(58), FromSapt: intp(50)}, "sapt"},
		{"fromSapt not before before", cycle, PageRequest{FromSapt: intp(56), Before: intp(56)}, "fromSapt"},
		{"fromSapt after the newest", cycle, PageRequest{FromSapt: intp(60)}, "fromSapt"},
		{"negative count", cycle, PageRequest{Count: -1}, "count"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Paginate(tc.r, tc.req)
			var pe *PageError
			if !errors.As(err, &pe) || pe.Param != tc.param {
				t.Errorf("err = %v, want a PageError on %q", err, tc.param)
			}
		})
	}
	// before = the newest + 1 is the same as no cursor.
	if p := paginate(t, cycle, PageRequest{Before: intp(60)}); p.Weeks[0] != 59 {
		t.Errorf("before=60 starts at SAPT %d", p.Weeks[0])
	}
}

func TestPaginateEmptyPeriod(t *testing.T) {
	r := resolve(t, hirbovat2026, Period{Kind: PeriodInterval, From: day(2025, time.November, 1), To: day(2025, time.November, 15)})
	p, err := Paginate(r, PageRequest{Before: intp(5)})
	if err != nil {
		t.Fatal(err)
	}
	if p.Weeks == nil || len(p.Weeks) != 0 || p.Loaded != nil || p.More != nil || p.Start != "" || p.TotalsLabel != "" {
		t.Errorf("empty period page = %+v, want weeks [] and nothing else", p)
	}
}
