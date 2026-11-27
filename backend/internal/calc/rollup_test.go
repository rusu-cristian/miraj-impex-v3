package calc

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// hirbovatDay is one row of hirbovatDays.
type hirbovatDay struct {
	sapt                             int
	date                             string
	startF, startM, endF, endM       int
	mortF, mortM, soldF, soldM       int
	eggs, inc, jos, cat1, cat2, cat3 int
	henDays                          int
	eggGrams                         string
	weighedEggs                      int
	kgF, kgM                         string
	fedF, fedM                       int
	guideLay, guideGF, guideGM       string
}

// hirbovatDays are Hîrbovăț's farm figures for SAPT 56–59 (27.10–18.11.2026),
// worked out from design/data/extracts/hirbovat-daily.json the way the day
// query returns them: hen-days by hall under the M2 Q1 rule (H7 counts on
// 14.11, the day it was emptied, with no eggs), Σ eggs × hall egg weight,
// fed bird-days over the section × sex rows with a ration (H7 ♀ has none on
// 13.11, §5.11), and the guide of each day's SAPT.
var hirbovatDays = []hirbovatDay{
	{56, "2026-10-27", 25896, 2053, 25890, 2049, 6, 4, 0, 0, 18923, 18555, 710, 37, 160, 171, 25896, "1279099.60", 18923, "4173.3", "289.3", 25896, 2053, "61.1", "158", "153"},
	{56, "2026-10-28", 25890, 2049, 25885, 2047, 5, 2, 0, 0, 18896, 18532, 730, 39, 149, 176, 25890, "1289974.04", 18896, "4172.3", "288.6", 25890, 2049, "61.1", "158", "153"},
	{56, "2026-10-29", 25885, 2047, 25877, 2044, 8, 3, 0, 0, 18547, 18233, 743, 43, 118, 153, 25885, "1247190.05", 18547, "4171.5", "288.3", 25885, 2047, "61.1", "158", "153"},
	{56, "2026-10-30", 25877, 2044, 25855, 2040, 22, 4, 0, 0, 18728, 18442, 742, 38, 116, 132, 25877, "1256181.26", 18728, "4170.3", "287.9", 25877, 2044, "61.1", "158", "153"},
	{56, "2026-10-31", 25855, 2040, 25827, 2035, 28, 5, 0, 0, 17081, 16778, 739, 29, 80, 194, 25855, "1145342.05", 17081, "4166.8", "287.3", 25855, 2040, "61.1", "158", "153"},
	{56, "2026-11-01", 25827, 2035, 25765, 2027, 62, 8, 0, 0, 14478, 14260, 704, 15, 70, 133, 25827, "980032.38", 14478, "4162.3", "286.7", 25827, 2035, "61.1", "158", "153"},
	{56, "2026-11-02", 25765, 2027, 25668, 2018, 97, 9, 0, 0, 11276, 11120, 547, 13, 44, 99, 25765, "759600.92", 11276, "3691.7", "285.5", 25765, 2027, "61.1", "158", "153"},
	{57, "2026-11-03", 25668, 2018, 25547, 2008, 121, 10, 0, 0, 8685, 8570, 424, 10, 36, 69, 25668, "577552.69", 8685, "2760", "284.2", 25668, 2018, "59.9", "158", "154"},
	{57, "2026-11-04", 25547, 2008, 25477, 2002, 70, 6, 0, 0, 7261, 7172, 282, 8, 41, 40, 25547, "488915.81", 7261, "2760", "282.7", 25547, 2008, "59.9", "158", "154"},
	{57, "2026-11-05", 25477, 2002, 25446, 1998, 31, 4, 0, 0, 6596, 6510, 240, 11, 34, 41, 25477, "439952.26", 6596, "3309.9", "281.9", 25477, 2002, "59.9", "158", "154"},
	{57, "2026-11-06", 25446, 1998, 25434, 1995, 12, 3, 0, 0, 6379, 6313, 220, 15, 21, 30, 25446, "428232.21", 6379, "3394.1", "281.3", 25446, 1998, "59.9", "158", "154"},
	{57, "2026-11-07", 25434, 1995, 25419, 1989, 15, 6, 0, 0, 6577, 6490, 210, 24, 28, 35, 25434, "442235.36", 6577, "3646.1", "280.9", 25434, 1995, "59.9", "158", "154"},
	{57, "2026-11-08", 25419, 1989, 25413, 1984, 6, 5, 0, 0, 6846, 6761, 194, 14, 29, 42, 25419, "460268.54", 6846, "3643.9", "280.1", 25419, 1989, "59.9", "158", "154"},
	{57, "2026-11-09", 25413, 1984, 25408, 1984, 5, 0, 0, 0, 6881, 6790, 180, 12, 38, 41, 25413, "464766.70", 6881, "3770.2", "279.5", 25413, 1984, "59.9", "158", "154"},
	{58, "2026-11-10", 25408, 1984, 25403, 1982, 5, 2, 0, 0, 7250, 7147, 242, 17, 43, 43, 25408, "489051.54", 7250, "3895.8", "281.7", 25408, 1984, "58.8", "158", "155"},
	{58, "2026-11-11", 25403, 1982, 25389, 1981, 14, 1, 0, 0, 7301, 7190, 231, 20, 47, 44, 25403, "490202.21", 7301, "3895.6", "281.5", 25403, 1982, "58.8", "158", "155"},
	{58, "2026-11-12", 25389, 1981, 23941, 1979, 6, 2, 1442, 0, 7578, 7421, 231, 18, 80, 59, 25389, "506947.51", 7578, "4020.4", "281.3", 25389, 1981, "58.8", "158", "155"},
	{58, "2026-11-13", 23941, 1979, 23934, 1978, 7, 1, 0, 0, 7191, 7010, 169, 20, 113, 48, 23941, "486724.73", 7191, "2723.3", "281.1", 17021, 1979, "58.8", "158", "155"},
	{58, "2026-11-14", 23934, 1978, 16710, 1285, 4, 1, 7220, 692, 6262, 5855, 138, 19, 315, 73, 23934, "413771.32", 6262, "2723", "190.8", 17019, 1339, "58.8", "158", "155"},
	{58, "2026-11-15", 16710, 1285, 15633, 1284, 0, 1, 1077, 0, 6082, 5713, 113, 19, 294, 56, 16710, "394558.02", 6082, "2471.5", "183.2", 16710, 1285, "58.8", "158", "155"},
	{58, "2026-11-16", 15633, 1284, 14826, 1227, 0, 0, 807, 57, 5848, 5409, 97, 16, 357, 66, 15633, "383362.40", 5848, "2501.3", "183.1", 15633, 1284, "58.8", "158", "155"},
	{59, "2026-11-17", 14826, 1227, 14395, 1083, 7, 0, 424, 144, 5892, 5610, 153, 12, 201, 69, 14826, "387615.41", 5892, "2372.1", "175", 14826, 1227, "57.6", "157", "156"},
	{59, "2026-11-18", 14395, 1083, 12561, 794, 3, 0, 1831, 289, 5914, 5308, 158, 21, 527, 58, 14395, "392472.24", 5914, "2303.2", "154.7", 14395, 1083, "57.6", "157", "156"},
}

func mustDec(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func mustDecp(s string) *decimal.Decimal {
	d := mustDec(s)
	return &d
}

func (h hirbovatDay) figures(t *testing.T) DayFigures {
	t.Helper()
	d, err := timex.ParseDay(h.date)
	if err != nil {
		t.Fatal(err)
	}
	f := DayFigures{
		Date:        d,
		AliveStart:  PerSex[int]{F: h.startF, M: h.startM},
		AliveEnd:    PerSex[int]{F: h.endF, M: h.endM},
		Mortality:   PerSex[int]{F: h.mortF, M: h.mortM},
		Sold:        PerSex[int]{F: h.soldF, M: h.soldM},
		Eggs:        h.eggs,
		Hatching:    h.inc,
		Floor:       h.jos,
		Cat1:        h.cat1,
		Cat2:        h.cat2,
		Cat3:        h.cat3,
		EggGrams:    mustDec(h.eggGrams),
		WeighedEggs: h.weighedEggs,
		HenDays:     h.henDays,
		FeedKg:      PerSex[decimal.Decimal]{F: mustDec(h.kgF), M: mustDec(h.kgM)},
		FedBirdDays: PerSex[int]{F: h.fedF, M: h.fedM},
		GuideLay:    mustDecp(h.guideLay),
		GuideFeedG:  PerSex[*decimal.Decimal]{F: mustDecp(h.guideGF), M: mustDecp(h.guideGM)},
	}
	// The dev seed carves culls out of the register's deaths (D165); do the
	// same on two days so Deaths() has to add them back.
	switch h.date {
	case "2026-11-18":
		f.Mortality.F, f.Culls.F = 2, 1
	case "2026-11-10":
		f.Mortality.M, f.Culls.M = 1, 1
	}
	return f
}

// weekRollups adds hirbovatDays up per SAPT.
func weekRollups(t *testing.T) map[int]Rollup {
	t.Helper()
	weeks := map[int]Rollup{}
	for _, h := range hirbovatDays {
		r := weeks[h.sapt]
		r.Add(h.figures(t))
		weeks[h.sapt] = r
	}
	return weeks
}

// rnd is p rounded once to places decimals (§5.15), "nil" for nil.
func rnd(p *decimal.Decimal, places int32) string {
	if p == nil {
		return "nil"
	}
	return num.Round(*p, places).String()
}

func r1(p *decimal.Decimal) string { return rnd(p, 1) }

type rollupCheck struct {
	name      string
	got, want any
}

func runChecks(t *testing.T, checks []rollupCheck) {
	t.Helper()
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

func aliveEnd(t *testing.T, r Rollup) PerSex[int] {
	t.Helper()
	a, ok := r.AliveEnd()
	if !ok {
		t.Fatal("AliveEnd: no day")
	}
	return a
}

func TestRollupWeekRows(t *testing.T) {
	weeks := weekRollups(t)
	lay3 := mustDec("3")

	// Golden 18, 19: SAPT 59 (17–18.11, 2 din 7 zile).
	w := weeks[59]
	diff := DiffPts(w.LayPct(), w.GuideLay())
	st, _ := StatusOf(diff, lay3)
	cat := w.CatPct()
	runChecks(t, []rollupCheck{
		{"59 eggs", w.Eggs, 11806}, {"59 hen-days", w.HenDays, 29221},
		{"59 lay", r1(w.LayPct()), "40.4"}, {"59 lay 2 dp", rnd(w.LayPct(), 2), "40.4"},
		{"59 guide", r1(w.GuideLay()), "57.6"}, {"59 diff", r1(diff), "-17.2"}, {"59 status", st, Abatere},
		{"59 inc", r1(w.HatchingPct()), "92.5"}, {"59 jos", r1(w.FloorPct()), "2.6"},
		{"59 cat1", r1(cat[0]), "0.3"}, {"59 cat2", r1(cat[1]), "6.2"}, {"59 cat3", r1(cat[2]), "1.1"},
		{"59 deaths", w.Deaths(), PerSex[int]{F: 10, M: 0}}, {"59 culls", w.Culls, PerSex[int]{F: 1, M: 0}},
		{"59 sold", w.Sold, PerSex[int]{F: 2255, M: 433}},
		{"59 kg F", w.FeedKg.F.String(), "4675.3"}, {"59 kg M", w.FeedKg.M.String(), "329.7"},
		{"59 alive", aliveEnd(t, w), PerSex[int]{F: 12561, M: 794}},
		{"59 feed guide F", r1(w.GuideFeedG().F), "157"}, {"59 feed guide M", r1(w.GuideFeedG().M), "156"},
	})

	// Golden 20, 21, 17: SAPT 58 (H7 emptied on 14.11 counts as a hen-day, Q1).
	w = weeks[58]
	cat = w.CatPct()
	g := w.FeedGPerBird()
	runChecks(t, []rollupCheck{
		{"58 eggs", w.Eggs, 47512}, {"58 hen-days", w.HenDays, 156418},
		{"58 lay", r1(w.LayPct()), "30.4"}, {"58 guide", r1(w.GuideLay()), "58.8"},
		{"58 inc", r1(w.HatchingPct()), "96.3"},
		{"58 cat1", r1(cat[0]), "0.3"}, {"58 cat2", r1(cat[1]), "2.6"}, {"58 cat3", r1(cat[2]), "0.8"},
		{"58 deaths", w.Deaths(), PerSex[int]{F: 36, M: 8}}, {"58 sold", w.Sold, PerSex[int]{F: 10546, M: 749}},
		{"58 alive", aliveEnd(t, w), PerSex[int]{F: 14826, M: 1227}},
		{"58 kg F", w.FeedKg.F.String(), "22230.9"}, {"58 kg M", w.FeedKg.M.String(), "1682.7"},
		{"58 bird-days", w.FedBirdDays, PerSex[int]{F: 142583, M: 11834}},
		{"58 g F", r1(g.F), "155.9"}, {"58 g M", r1(g.M), "142.2"},
		{"58 egg weight", r1(w.EggWeight()), "66.6"},
	})

	// Golden 22: SAPT 57 and 56.
	w57, w56 := weeks[57], weeks[56]
	runChecks(t, []rollupCheck{
		{"57 eggs", w57.Eggs, 49225}, {"57 lay", r1(w57.LayPct()), "27.6"}, {"57 guide", r1(w57.GuideLay()), "59.9"},
		{"56 eggs", w56.Eggs, 117929}, {"56 lay", r1(w56.LayPct()), "65.2"}, {"56 guide", r1(w56.GuideLay()), "61.1"},
		{"56 inc", r1(w56.HatchingPct()), "98.3"},
	})
}

func TestRollupTotals(t *testing.T) {
	weeks := weekRollups(t)
	// Golden 24: foot "Total SAPT 56–59", the week rows added newest first as
	// the table holds them.
	var tot Rollup
	for _, k := range []int{59, 58, 57, 56} {
		tot = tot.Plus(weeks[k])
	}
	runChecks(t, []rollupCheck{
		{"eggs", tot.Eggs, 226472}, {"hen-days", tot.HenDays, 545038},
		{"lay", r1(tot.LayPct()), "41.6"}, {"guide (hen-day weighted)", r1(tot.GuideLay()), "59.9"},
		{"deaths", tot.Deaths(), PerSex[int]{F: 534, M: 77}}, {"sold", tot.Sold, PerSex[int]{F: 12801, M: 1182}},
		{"inc", r1(tot.HatchingPct()), "97.7"},
		{"kg F", tot.FeedKg.F.String(), "78898.6"}, {"kg M", tot.FeedKg.M.String(), "5996.6"},
		// Foot alive = the latest loaded day, though SAPT 56 was added last.
		{"alive", aliveEnd(t, tot), PerSex[int]{F: 12561, M: 794}},
	})
	// Adding every day into one rollup gives the same row as adding weeks.
	var flat Rollup
	for _, h := range hirbovatDays {
		flat.Add(h.figures(t))
	}
	runChecks(t, []rollupCheck{
		{"flat lay", rnd(flat.LayPct(), 4), rnd(tot.LayPct(), 4)},
		{"flat guide", rnd(flat.GuideLay(), 4), rnd(tot.GuideLay(), 4)},
		{"flat g F", rnd(flat.FeedGPerBird().F, 4), rnd(tot.FeedGPerBird().F, 4)},
		{"flat egg weight", rnd(flat.EggWeight(), 4), rnd(tot.EggWeight(), 4)},
	})
}

func TestRollupDayRow(t *testing.T) {
	// Goldens 12, 15, 17 and the §7.4 day row of 18.11.2026.
	var d Rollup
	d.Add(hirbovatDays[len(hirbovatDays)-1].figures(t))
	runChecks(t, []rollupCheck{
		{"lay 2 dp", rnd(d.LayPct(), 2), "41.08"}, {"guide", r1(d.GuideLay()), "57.6"},
		{"diff 2 dp", rnd(DiffPts(d.LayPct(), d.GuideLay()), 2), "-16.52"},
		{"inc", r1(d.HatchingPct()), "89.8"}, {"jos", r1(d.FloorPct()), "2.7"},
		{"egg weight", r1(d.EggWeight()), "66.4"},
		{"deaths", d.Deaths(), PerSex[int]{F: 3, M: 0}},
		{"g F", r1(d.FeedGPerBird().F), "160"},
	})
}

func TestRatesAreRatiosOfSums(t *testing.T) {
	// SAPT 59: 5 892 ÷ 14 826 = 39,74 % and 5 914 ÷ 14 395 = 41,08 %. Their
	// mean is 40,41; the week's rate is 11 806 ÷ 29 221 = 40,40.
	var r Ratio
	r.AddInt(5892, 14826)
	r.AddInt(5914, 14395)
	if got := rnd(r.Pct(), 2); got != "40.4" {
		t.Errorf("Σ eggs ÷ Σ hen-days = %s, want 40.40", got)
	}
	mean := Pct(mustDec("5892"), mustDec("14826")).Add(*Pct(mustDec("5914"), mustDec("14395"))).Div(mustDec("2"))
	if got := num.Round(mean, 2).String(); got != "40.41" {
		t.Errorf("mean of day rates = %s; the test no longer tells the two apart", got)
	}

	// Golden 23: the peak, rounded once.
	var peak Ratio
	peak.AddInt(168122, 186088)
	if got := r1(peak.Pct()); got != "90.3" {
		t.Errorf("168 122 ÷ 186 088 = %s, want 90.3", got)
	}
	// Plus, Value and Times.
	var a, b Ratio
	a.Add(mustDec("2303.2"), mustDec("14395"))
	b.Add(mustDec("2372.1"), mustDec("14826"))
	sum := a.Plus(b)
	if got := r1(sum.Times(mustDec("1000"))); got != "160" {
		t.Errorf("g = %s, want 160", got)
	}
	if got := rnd(sum.Value(), 4); got != "0.16" {
		t.Errorf("Value = %s, want 0.16", got)
	}
	var empty Ratio
	if empty.Pct() != nil || empty.Value() != nil || empty.Times(mustDec("1000")) != nil {
		t.Error("a rate with no base is not nil")
	}
}

func TestWeighted(t *testing.T) {
	// Golden 24's guide from the week hen-days: (61,1 × 180 995 + 59,9 ×
	// 178 404 + 58,8 × 156 418 + 57,6 × 29 221) ÷ 545 038 = 59,86.
	var w Weighted
	w.Add(mustDecp("61.1"), 180995)
	w.Add(mustDecp("59.9"), 178404)
	w.Add(mustDecp("58.8"), 156418)
	w.Add(mustDecp("57.6"), 29221)
	if got := rnd(w.Value(), 2); got != "59.86" {
		t.Errorf("weighted guide = %s, want 59.86", got)
	}
	// A day without weight doesn't count, with or without a guide.
	w.Add(nil, 0)
	w.Add(mustDecp("99"), 0)
	if got := rnd(w.Value(), 2); got != "59.86" {
		t.Errorf("after weightless days = %s, want 59.86", got)
	}
	// A weighted day without a guide row makes the row's guide "–".
	var gap Weighted
	gap.Add(mustDecp("57.6"), 100)
	gap.Add(nil, 50)
	if gap.Value() != nil {
		t.Errorf("partial guide = %s, want nil", gap.Value())
	}
	if w.Plus(gap).Value() != nil {
		t.Error("Plus loses the gap")
	}
	var none Weighted
	if none.Value() != nil {
		t.Error("no weight is not nil")
	}
}

func TestLatest(t *testing.T) {
	var l Latest[int]
	if _, ok := l.Value(); ok {
		t.Error("empty Latest has a value")
	}
	l.Add(day(2026, time.November, 18), 12561)
	l.Add(day(2026, time.November, 17), 14395) // an older day loses
	if v, _ := l.Value(); v != 12561 || l.Day() != day(2026, time.November, 18) {
		t.Errorf("Latest = %d on %s", v, l.Day())
	}
	var older Latest[int]
	older.Add(day(2026, time.November, 16), 14826)
	if v, _ := older.Plus(l).Value(); v != 12561 {
		t.Errorf("older.Plus(newer) = %d", v)
	}
	if v, _ := l.Plus(Latest[int]{}).Value(); v != 12561 {
		t.Errorf("Plus(empty) = %d", v)
	}
}

func TestStatus(t *testing.T) {
	three := mustDec("3")
	for _, tc := range []struct {
		diff *decimal.Decimal
		want Status
	}{
		{mustDecp("-17.2"), Abatere},
		{mustDecp("2.9"), InNorma},
		{mustDecp("-2.9"), InNorma},
		{mustDecp("3"), InNorma}, // |diff| > 3, so 3 itself is in norm
		{mustDecp("-3.0001"), Abatere},
	} {
		if got, ok := StatusOf(tc.diff, three); !ok || got != tc.want {
			t.Errorf("StatusOf(%s, 3) = %q, %v; want %q", tc.diff, got, ok, tc.want)
		}
	}
	if _, ok := StatusOf(nil, three); ok {
		t.Error("StatusOf(nil) is ok")
	}

	// Golden 32: 2 553 g vs 2 370 g = +7,7 %, abatere at ±5 %.
	dev := DeviationPct(mustDecp("2553"), mustDecp("2370"))
	if got := r1(dev); got != "7.7" {
		t.Errorf("weight deviation = %s, want 7.7", got)
	}
	if st, _ := StatusOf(dev, mustDec("5")); st != Abatere {
		t.Errorf("weight status = %q", st)
	}
	// §7.6 feedToday: 160 g vs 157 g = +1,9 %, in norm at ±10 %.
	if st, _ := StatusOf(DeviationPct(mustDecp("160"), mustDecp("157")), mustDec("10")); st != InNorma {
		t.Errorf("feed status = %q", st)
	}
	if DeviationPct(mustDecp("1"), mustDecp("0")) != nil || DeviationPct(nil, mustDecp("1")) != nil || DiffPts(mustDecp("1"), nil) != nil {
		t.Error("a comparison without a guide is not nil")
	}

	// Golden 10 / §7.6: 5,66 % vs guide 7,25 → in norm; above the guide → abatere.
	if st, ok := CumMortalityStatus(CumMortalityPct(1521, 26883), mustDecp("7.25")); !ok || st != InNorma {
		t.Errorf("cum. mortality status = %q, %v", st, ok)
	}
	if st, _ := CumMortalityStatus(mustDecp("7.26"), mustDecp("7.25")); st != Abatere {
		t.Errorf("above the guide = %q", st)
	}
	if _, ok := CumMortalityStatus(mustDecp("34.66"), nil); ok {
		t.Error("no ♂ guide is ok")
	}
}

// TestWorst covers how a stat over both sexes combines their lamps (§7.6
// feedToday).
func TestWorst(t *testing.T) {
	in, ab := Compared(InNorma, true), Compared(Abatere, true)
	none := Compared("", false)
	for _, tc := range []struct {
		name string
		all  []Comparison
		want Status
		ok   bool
	}{
		{"nothing compared", []Comparison{none, none}, "", false},
		{"no comparisons", nil, "", false},
		{"one in norm", []Comparison{none, in}, InNorma, true},
		{"abatere first", []Comparison{ab, in}, Abatere, true},
		{"abatere second", []Comparison{in, ab}, Abatere, true},
		{"abatere beside none", []Comparison{ab, none}, Abatere, true},
	} {
		if got, ok := Worst(tc.all...); got != tc.want || ok != tc.ok {
			t.Errorf("%s: Worst = %q, %v; want %q, %v", tc.name, got, ok, tc.want, tc.ok)
		}
	}
}
