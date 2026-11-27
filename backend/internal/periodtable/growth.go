package periodtable

import (
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
)

// The growth Rezumat (§7.4): g/pasăre ♀ · ♂ against the guide and the ♀
// weighings against the guide of SAPT − 1, beside the common groups.

// feedGFigures is a row's g/pasăre for one sex and its guide (§5.7, D38):
// g = 1000 × Σ kg ÷ Σ bird-days on fed days. The guide of a day is its
// SAPT's; of a week, weighted by fed bird-days, or its SAPT's when it has
// none; of the totals, weighted by fed bird-days. gap reports a missing
// guide value.
func (g guide) feedGFigures(a acc, k int, kind rowKind, sex dbq.Sex) (actual, guidePct *decimal.Decimal, gap bool) {
	all, guides := a.r.FeedGPerBird(), a.r.GuideFeedG()
	fed := a.r.FedBirdDays.M
	actual, guidePct = all.M, guides.M
	if sex == dbq.SexF {
		fed, actual, guidePct = a.r.FedBirdDays.F, all.F, guides.F
	}
	if g.weeks == nil {
		return actual, nil, false
	}
	if kind == rowDay || (kind == rowWeek && fed == 0) {
		guidePct = g.weeks.FeedG(k, sex)
		return actual, guidePct, guidePct == nil
	}
	return actual, guidePct, guidePct == nil && fed > 0
}

// growthCells writes the g/pasăre and Greutate ♀ cells of a row.
//   - g/pasăre: null and "missing" when the row has birds and no ration; the
//     status compares the deviation from the guide with tolerance.feed_pct.
//   - Greutate ♀: the bird-weighted mean of the row's ♀ weighings (A8); a
//     week without one is null with "no_weighing", the totals are "–"
//     (§5.14). The guide is the weight at the end of SAPT − 1 (§5.10), shown
//     with the weighing; the status compares the deviation with
//     tolerance.weight_pct.
func (l layout) growthCells(w *cellWriter, a acc, k int, kind rowKind) {
	for _, sex := range []struct {
		sex         dbq.Sex
		key, guided string
	}{{dbq.SexF, KeyFeedGF, KeyFeedGGuideF}, {dbq.SexM, KeyFeedGM, KeyFeedGGuideM}} {
		actual, guidePct, gap := l.guide.feedGFigures(a, k, kind, sex.sex)
		w.set(sex.key, numDec(actual))
		if actual == nil && a.feedRows == 0 && a.birdDays > 0 {
			w.reason(sex.key, ReasonMissing)
		}
		if l.guide.weeks != nil {
			w.set(sex.guided, numDec(guidePct))
			if gap {
				w.reason(sex.guided, ReasonNoGuide)
			}
			w.compare(sex.key, calc.DeviationPct(actual, guidePct), l.tolFeed)
		}
	}

	var mean, guidePct *decimal.Decimal
	if kind != rowTotals {
		mean = a.weight.F.Value()
	}
	w.set(KeyWeightF, numDec(mean))
	if mean == nil && kind == rowWeek {
		w.reason(KeyWeightF, ReasonNoWeighing)
	}
	if mean != nil && l.guide.weeks != nil {
		guidePct = l.guide.weeks.BodyWeightG(k-1, dbq.SexF)
		if guidePct == nil {
			w.reason(KeyWeightGuideF, ReasonNoGuide)
		}
	}
	diff := calc.DeviationPct(mean, guidePct)
	w.set(KeyWeightGuideF, numDec(guidePct))
	w.set(KeyWeightDiffF, numDec(diff))
	w.compare(KeyWeightF, diff, l.tolWeight)
}
