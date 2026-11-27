package calc

import (
	"errors"
	"fmt"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Period kinds of §5.2 (the `period` query value before its ":").
const (
	PeriodCycle    = "cycle"
	PeriodWeeks    = "weeks"
	PeriodInterval = "interval"
)

// DefaultPageWeeks is how many weeks a PeriodTable page loads (§5.2, D190).
const DefaultPageWeeks = 4

// Period is the FilterBar period as asked (§5.2, §7.4). Its fields are those
// of periodtable.Period, so either converts to the other.
type Period struct {
	Kind     string    // PeriodCycle | PeriodWeeks | PeriodInterval
	Weeks    int       // weeks:n
	From, To timex.Day // interval, both days included
}

var (
	// ErrPeriod is returned for an unknown kind, weeks < 1, or an interval
	// that is missing a day or starts after it ends. The query parser refuses
	// these first; this guards the builders.
	ErrPeriod = errors.New("calc: not a valid period")
	// ErrIntervalFuture is returned for an interval that ends after today
	// (messages.IntervalInFuture).
	ErrIntervalFuture = errors.New("calc: the interval ends after today")
)

// ResolvedPeriod is a period applied to one cycle and today (§5.2): the days
// it covers are inside the period, the cycle and ≤ today.
type ResolvedPeriod struct {
	Kind  string
	Label string // "Ciclul curent", "Tot ciclul", "4 săpt.", "01.11–15.11.2026"
	// From and To are the days covered. For an Empty period they are the days
	// asked for (the interval, or the cycle start and today).
	From, To timex.Day
	// SaptFrom and SaptTo are the weeks of From and To.
	SaptFrom, SaptTo int
	// CycleSaptFrom and CycleSaptTo span the cycle's weeks up to today: how far
	// load-more can reach and the "din N" of its label. With period `cycle`
	// they equal SaptFrom and SaptTo.
	CycleSaptFrom, CycleSaptTo int
	// Empty is true when no day of the period lies in the cycle up to today
	// (§7.4 States: weeks [], empty.next).
	Empty bool
	Age   SeriesAge
}

// WeeksTotal is the number of weeks the period covers, 0 when Empty.
func (r ResolvedPeriod) WeeksTotal() int {
	if r.Empty {
		return 0
	}
	return r.SaptTo - r.SaptFrom + 1
}

// CycleWeeks is the number of the cycle's weeks up to today ("4 din 40").
func (r ResolvedPeriod) CycleWeeks() int {
	if r.Empty {
		return 0
	}
	return r.CycleSaptTo - r.CycleSaptFrom + 1
}

// Week returns week k of the period with its days clipped to the period
// (partial weeks, §5.1), or false when k is outside it.
func (r ResolvedPeriod) Week(k int) (WeekSpan, bool) {
	if r.Empty || k < r.SaptFrom || k > r.SaptTo {
		return WeekSpan{}, false
	}
	return WeekSpan{
		Sapt: k,
		From: laterDay(WeekFirstDay(r.Age, k), r.From),
		To:   earlierDay(WeekLastDay(r.Age, k), r.To),
	}, true
}

// WeeksNewestFirst returns the period's weeks from SAPT to down to SAPT from,
// both included and clipped to the period.
func (r ResolvedPeriod) WeeksNewestFirst(from, to int) []WeekSpan {
	out := []WeekSpan{}
	for k := to; k >= from; k-- {
		if w, ok := r.Week(k); ok {
			out = append(out, w)
		}
	}
	return out
}

// WeekSpan is one week row: SAPT k with the days of it the row covers.
type WeekSpan struct {
	Sapt     int
	From, To timex.Day
}

// DaysInRow is the number of days the row covers (§5.1 "2 din 7 zile").
func (w WeekSpan) DaysInRow() int { return w.To.DaysSince(w.From) + 1 }

// Partial reports whether the row covers fewer than 7 days.
func (w WeekSpan) Partial() bool { return w.DaysInRow() < 7 }

// Label is the row's label: "SAPT 59 · 17.11–18.11 · 2 din 7 zile".
func (w WeekSpan) Label() string { return messages.WeekRow(w.Sapt, w.From, w.To) }

// Contains reports whether d is one of the row's days.
func (w WeekSpan) Contains(d timex.Day) bool { return !d.Before(w.From) && !d.After(w.To) }

// ResolvePeriod applies p to cycle c on today (§5.2):
//   - cycle: the cycle's start to min(end, today); "Ciclul curent" on the open
//     cycle, "Tot ciclul" on a past one (D52);
//   - weeks:n: SAPT(last day) and the n − 1 weeks before it, clipped to the cycle;
//   - interval: its days clipped to the cycle; the weeks it touches. An end
//     after today is ErrIntervalFuture.
//
// c.StartDate is the first day the table counts. For a cycle without a flock
// count that is the first register day (M2 Q7), so the caller passes a copy of
// the cycle with that start.
func ResolvePeriod(c Cycle, p Period, today timex.Day) (ResolvedPeriod, error) {
	last := today
	if c.EndDate != nil && c.EndDate.Before(today) {
		last = *c.EndDate
	}
	r := ResolvedPeriod{Kind: p.Kind, Age: c.Age}
	from, to := c.StartDate, last
	switch p.Kind {
	case PeriodCycle:
		r.Label = messages.PeriodWholeCycle()
		if c.EndDate == nil {
			r.Label = messages.PeriodCurrentCycle()
		}
	case PeriodWeeks:
		if p.Weeks < 1 {
			return ResolvedPeriod{}, fmt.Errorf("%w: weeks:%d", ErrPeriod, p.Weeks)
		}
		r.Label = messages.PeriodWeeks(p.Weeks)
		from = laterDay(WeekFirstDay(c.Age, SAPT(c.Age, last)-p.Weeks+1), c.StartDate)
	case PeriodInterval:
		if p.From.IsZero() || p.To.IsZero() || p.From.After(p.To) {
			return ResolvedPeriod{}, fmt.Errorf("%w: interval %s–%s", ErrPeriod, p.From, p.To)
		}
		if p.To.After(today) {
			return ResolvedPeriod{}, ErrIntervalFuture
		}
		r.Label = messages.PeriodInterval(p.From, p.To)
		from, to = laterDay(p.From, c.StartDate), earlierDay(p.To, last)
		if from.After(to) {
			from, to = p.From, p.To
			r.Empty = true
		}
	default:
		return ResolvedPeriod{}, fmt.Errorf("%w: kind %q", ErrPeriod, p.Kind)
	}
	if c.StartDate.After(last) { // the cycle starts after today
		r.Empty = true
	}
	r.From, r.To = from, to
	r.SaptFrom, r.SaptTo = SAPT(c.Age, from), SAPT(c.Age, to)
	r.CycleSaptFrom, r.CycleSaptTo = SAPT(c.Age, c.StartDate), SAPT(c.Age, last)
	return r, nil
}

func laterDay(a, b timex.Day) timex.Day {
	if a.After(b) {
		return a
	}
	return b
}

func earlierDay(a, b timex.Day) timex.Day {
	if a.Before(b) {
		return a
	}
	return b
}

// PageRequest is the paging part of a table query (§7.4).
type PageRequest struct {
	Before   *int // the oldest SAPT already loaded (exclusive); nil = the newest weeks
	Count    int  // weeks per page; 0 = DefaultPageWeeks
	Sapt     *int // one week only
	FromSapt *int // every week from here up to the newest not yet loaded
}

// PageError names the paging parameter that is out of range for the period,
// so the handler can report its issue path.
type PageError struct {
	Param string // "before" | "count" | "sapt" | "fromSapt"
}

func (e *PageError) Error() string { return "calc: " + e.Param + " is out of range for the period" }

// Page is one PeriodTable page (§7.4, D190–D192).
type Page struct {
	Weeks []WeekSpan // the weeks to return, newest first; empty for an Empty period
	// Loaded spans every week loaded so far: this page's oldest to the
	// period's newest (the totals run over it). Nil for an Empty period.
	Loaded *PageLoaded
	// More is the load-more row, nil once the cycle's first week is loaded.
	// It reaches back to the cycle start even when the period is shorter: the
	// client widens a weeks:n or interval period as it loads (§7.4).
	More *PageMore
	// Start is "Începutul ciclului · SAPT 20" once the cycle's first week is
	// loaded, "" before.
	Start string
	// TotalsLabel is "Total SAPT 56–59" over Loaded ("" when Loaded is nil).
	TotalsLabel string
}

// PageLoaded is the span of the loaded weeks and their days.
type PageLoaded struct {
	SaptFrom, SaptTo, Weeks int
	From, To                timex.Day
}

// PageMore is the load-more row: the next call sends before = Before.
type PageMore struct {
	Before    int
	Remaining int // the cycle's weeks before Before
	Label     string
}

// Paginate picks the weeks of one page of r (§7.4):
//   - default: the newest Count weeks;
//   - Before = b: the Count weeks before b; Loaded runs to the period's newest;
//   - FromSapt = f (with or without Before): every week from f (clipped to the
//     period) up to Before − 1 or the newest;
//   - Sapt = k: that week only, with no More and no Start (it opens one week
//     of a table already on screen).
//
// A parameter out of range for the period is a *PageError. An Empty period
// gives an empty page whatever the parameters.
func Paginate(r ResolvedPeriod, req PageRequest) (Page, error) {
	if r.Empty {
		return Page{Weeks: []WeekSpan{}}, nil
	}
	count := req.Count
	if count == 0 {
		count = DefaultPageWeeks
	}
	if count < 0 {
		return Page{}, &PageError{Param: "count"}
	}
	if req.Sapt != nil {
		k := *req.Sapt
		if req.Before != nil || req.FromSapt != nil || k < r.SaptFrom || k > r.SaptTo {
			return Page{}, &PageError{Param: "sapt"}
		}
		w, _ := r.Week(k)
		return Page{
			Weeks:       []WeekSpan{w},
			Loaded:      &PageLoaded{SaptFrom: k, SaptTo: k, Weeks: 1, From: w.From, To: w.To},
			TotalsLabel: messages.TotalsLabel(k, k),
		}, nil
	}
	upper := r.SaptTo // the newest week of this page
	if req.Before != nil {
		b := *req.Before
		if b <= r.SaptFrom || b > r.SaptTo+1 {
			return Page{}, &PageError{Param: "before"}
		}
		upper = b - 1
	}
	lower := max(upper-count+1, r.SaptFrom)
	if req.FromSapt != nil {
		if *req.FromSapt > upper {
			return Page{}, &PageError{Param: "fromSapt"}
		}
		lower = max(*req.FromSapt, r.SaptFrom)
	}
	oldest, _ := r.Week(lower)
	newest, _ := r.Week(r.SaptTo)
	p := Page{
		Weeks: r.WeeksNewestFirst(lower, upper),
		Loaded: &PageLoaded{
			SaptFrom: lower, SaptTo: r.SaptTo, Weeks: r.SaptTo - lower + 1,
			From: oldest.From, To: newest.To,
		},
		TotalsLabel: messages.TotalsLabel(lower, r.SaptTo),
	}
	if remaining := lower - r.CycleSaptFrom; remaining > 0 {
		p.More = &PageMore{
			Before:    lower,
			Remaining: remaining,
			Label:     messages.LoadMore(min(count, remaining), lower, r.SaptTo, p.Loaded.Weeks, r.CycleWeeks()),
		}
	} else {
		p.Start = messages.CycleStartNote(r.CycleSaptFrom)
	}
	return p, nil
}
