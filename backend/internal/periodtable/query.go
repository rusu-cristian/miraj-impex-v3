package periodtable

import (
	"errors"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Query parameter names of GET /api/table, in §7.4 order.
const (
	paramView      = "view"
	paramFarmID    = "farmId"
	paramSectionID = "sectionId"
	paramSeriesID  = "seriesId"
	paramCycleID   = "cycleId"
	paramTab       = "tab"
	paramPeriod    = "period"
	paramDetail    = "detail"
	paramTypes     = "types"
	paramGrain     = "grain"
	paramBefore    = "before"
	paramCount     = "count"
	paramSapt      = "sapt"
	paramFromSapt  = "fromSapt"
)

var params = []string{
	paramView, paramFarmID, paramSectionID, paramSeriesID, paramCycleID, paramTab, paramPeriod,
	paramDetail, paramTypes, paramGrain, paramBefore, paramCount, paramSapt, paramFromSapt,
}

// ParseQuery reads the query string of GET /api/table (§7.4) into a
// TableQuery. It checks each parameter on its own; Validate then checks what
// M2 serves and how the parameters go together. Every problem is a 400
// VALIDATION issue named by its parameter (M2 D2, the §7.0a codes):
//   - an unknown parameter → unknown_field (in name order, first);
//   - a repeated parameter, a malformed uuid, a non-integer or a malformed
//     period → bad_type; an interval day that is not a date → bad_date;
//   - a chip outside §7.4's list, a period kind other than cycle, weeks or
//     interval, weeks < 1, an interval that starts after it ends, and
//     before, count, sapt or fromSapt < 1 → out_of_range.
//
// Integers are read as int32 (§7.0a); a larger one is out_of_range. An
// absent parameter keeps its zero value (Count 0 = calc.DefaultPageWeeks).
func ParseQuery(values url.Values) (TableQuery, error) {
	var v httpx.Validator
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		v.Check(slices.Contains(params, name), name, httpx.IssueUnknownField, messages.InvalidData())
	}

	p := queryParser{v: &v, values: values}
	var q TableQuery
	q.View, _ = p.single(paramView)
	q.FarmID = p.uuid(paramFarmID)
	q.SectionID = p.uuid(paramSectionID)
	q.SeriesID = p.uuid(paramSeriesID)
	q.CycleID = p.uuid(paramCycleID)
	q.Tab, _ = p.single(paramTab)
	if s, ok := p.single(paramPeriod); ok {
		q.Period = p.period(s)
	}
	q.Detail, _ = p.single(paramDetail)
	if s, ok := p.single(paramTypes); ok {
		q.Types = p.types(s)
	}
	q.Grain, _ = p.single(paramGrain)
	q.Before = p.positive(paramBefore)
	if c := p.positive(paramCount); c != nil {
		q.Count = *c
	}
	q.Sapt = p.positive(paramSapt)
	q.FromSapt = p.positive(paramFromSapt)
	return q, v.Err()
}

// Validate checks the parsed query against what M2 serves (M2 D2) and how
// the parameters combine. Everything outside that is out_of_range on its
// parameter; a missing view or farmId is required:
//   - view must be farm (the other views come in M6), with a farmId and no
//     sectionId or seriesId;
//   - tab must be rezumat (or absent), detail farm (or absent): Rezumat has
//     no detail groups;
//   - grain must be day or week;
//   - sapt opens one week, so it goes with neither before nor fromSapt.
func (q TableQuery) Validate(v *httpx.Validator) {
	invalid := messages.InvalidData()
	if q.View == "" {
		v.Add(paramView, httpx.IssueRequired, invalid)
	} else {
		v.Check(q.View == viewFarm, paramView, httpx.IssueOutOfRange, invalid)
	}
	if q.View == viewFarm {
		v.Check(q.FarmID != nil, paramFarmID, httpx.IssueRequired, invalid)
		v.Check(q.SectionID == nil, paramSectionID, httpx.IssueOutOfRange, invalid)
		v.Check(q.SeriesID == nil, paramSeriesID, httpx.IssueOutOfRange, invalid)
	}
	v.Check(q.Tab == "" || q.Tab == tabRezumat, paramTab, httpx.IssueOutOfRange, invalid)
	v.Check(q.Detail == "" || q.Detail == detailFarm, paramDetail, httpx.IssueOutOfRange, invalid)
	v.Check(q.Grain == "" || q.Grain == grainDay || q.Grain == grainWeek, paramGrain, httpx.IssueOutOfRange, invalid)
	v.Check(q.Sapt == nil || (q.Before == nil && q.FromSapt == nil), paramSapt, httpx.IssueOutOfRange, invalid)
}

// queryParser reads single parameters, recording issues on v.
type queryParser struct {
	v      *httpx.Validator
	values url.Values
}

// single is the one value of parameter name, false when it is absent or
// repeated (a bad_type issue).
func (p queryParser) single(name string) (string, bool) {
	vals, ok := p.values[name]
	if !ok {
		return "", false
	}
	if !p.v.Check(len(vals) == 1, name, httpx.IssueBadType, messages.InvalidData()) {
		return "", false
	}
	return vals[0], true
}

func (p queryParser) uuid(name string) *uuid.UUID {
	s, ok := p.single(name)
	if !ok {
		return nil
	}
	id, err := uuid.Parse(s)
	if !p.v.Check(err == nil, name, httpx.IssueBadType, messages.InvalidData()) {
		return nil
	}
	return &id
}

// positive is an integer parameter ≥ 1, nil when absent or invalid.
func (p queryParser) positive(name string) *int {
	s, ok := p.single(name)
	if !ok {
		return nil
	}
	n, ok := p.int(name, s)
	if !ok {
		return nil
	}
	return &n
}

// int reads s as an int32 ≥ 1: bad_type when it is not an integer,
// out_of_range when it is < 1 or beyond int32.
func (p queryParser) int(name, s string) (int, bool) {
	n, err := strconv.ParseInt(s, 10, 32)
	if errors.Is(err, strconv.ErrRange) {
		p.v.Add(name, httpx.IssueOutOfRange, messages.InvalidData())
		return 0, false
	}
	if !p.v.Check(err == nil, name, httpx.IssueBadType, messages.InvalidData()) {
		return 0, false
	}
	if !p.v.Check(n >= 1, name, httpx.IssueOutOfRange, messages.InvalidData()) {
		return 0, false
	}
	return int(n), true
}

// period reads "cycle", "weeks:n" (any n ≥ 1; the FilterBar offers 4, 8
// and 13, load-more widens it) or "interval:from,to" (§5.2). The future end
// of an interval needs today, so Build reports it (calc.ErrIntervalFuture).
func (p queryParser) period(s string) Period {
	invalid := messages.InvalidData()
	kind, arg, hasArg := strings.Cut(s, ":")
	switch kind {
	case calc.PeriodCycle:
		if p.v.Check(!hasArg, paramPeriod, httpx.IssueBadType, invalid) {
			return Period{Kind: calc.PeriodCycle}
		}
	case calc.PeriodWeeks:
		if !p.v.Check(hasArg, paramPeriod, httpx.IssueBadType, invalid) {
			return Period{}
		}
		if n, ok := p.int(paramPeriod, arg); ok {
			return Period{Kind: calc.PeriodWeeks, Weeks: n}
		}
	case calc.PeriodInterval:
		a, b, ok := strings.Cut(arg, ",")
		if !p.v.Check(hasArg && ok, paramPeriod, httpx.IssueBadType, invalid) {
			return Period{}
		}
		from, errFrom := timex.ParseDay(a)
		to, errTo := timex.ParseDay(b)
		if !p.v.Check(errFrom == nil && errTo == nil, paramPeriod, httpx.IssueBadDate, invalid) {
			return Period{}
		}
		if p.v.Check(!from.After(to), paramPeriod, httpx.IssueOutOfRange, invalid) {
			return Period{Kind: calc.PeriodInterval, From: from, To: to}
		}
	default:
		p.v.Add(paramPeriod, httpx.IssueOutOfRange, invalid)
	}
	return Period{}
}

// types reads the comma list of chips; each must be a §7.4 chip.
func (p queryParser) types(s string) []string {
	out := []string{}
	for chip := range strings.SplitSeq(s, ",") {
		if !p.v.Check(slices.Contains(Chips, chip), paramTypes, httpx.IssueOutOfRange, messages.InvalidData()) {
			return nil
		}
		if !slices.Contains(out, chip) {
			out = append(out, chip)
		}
	}
	return out
}
