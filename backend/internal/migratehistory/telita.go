package migratehistory

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/fmtx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// telitaBirds plans Telița's bird ledger (§10.3 "Telița birds"): placement,
// mortality per section-day (0 rows kept), sexing errors, the internal moves
// and the links to Hîrbovăț as transfer pairs, the other corrections and the
// residuals as adjustments.
func (b *builder) telitaBirds() error {
	entry := b.p.migrationEntry[FarmTelita]
	if err := b.checkTelitaCorrections(); err != nil {
		return err
	}
	for _, r := range b.x.placement.Rows {
		if err := b.bird(entry, r.FarmID, r.Section, r.Date, r.Sex, dbq.BirdEventTypePlacement, r.Qty, nil, nil); err != nil {
			return err
		}
	}
	for _, r := range b.x.telitaDaily.Rows {
		if err := b.bird(entry, FarmTelita, r.Section, r.Date, r.Sex, dbq.BirdEventTypeMortality, -r.Mortality, nil, nil); err != nil {
			return err
		}
		// A sexing error is a value only where the register has one; an empty
		// cell is "not entered", never 0 (spec C1).
		if r.SexingError != 0 {
			if err := b.bird(entry, FarmTelita, r.Section, r.Date, r.Sex, dbq.BirdEventTypeSexingError, r.SexingError, nil, nil); err != nil {
				return err
			}
		}
	}
	telita, err := b.farm(FarmTelita)
	if err != nil {
		return err
	}
	// Internal moves: one transfer entry per move (user answer Q3). The
	// 23.02.2026 H6 ♀ → H5-1 ♂ move keeps the sexes as recorded (Q2).
	for _, m := range b.x.transfers.MovesTelita {
		e := b.newEntry(dbq.EntryKindTransfer, telita, &telita, m.Date, nil)
		if err := b.pair(e, end{FarmTelita, m.From, m.SexFrom}, end{FarmTelita, m.To, m.SexTo}, m.Date, m.Qty); err != nil {
			return err
		}
	}
	if err := b.links(); err != nil {
		return err
	}
	for _, c := range b.x.transfers.OtherCorrectionsTelita {
		if err := b.adjust(FarmTelita, c.Section, c.Date, c.Sex, c.Qty); err != nil {
			return err
		}
	}
	return b.telitaResiduals()
}

// checkTelitaCorrections proves the register's Corectie column is exactly the
// internal moves + the transfers out + the other corrections (so importing
// those, and not the column, loses nothing).
func (b *builder) checkTelitaCorrections() error {
	type key struct {
		day     timex.Day
		section string
	}
	want := map[key]int32{}
	t := b.x.transfers
	for _, m := range t.MovesTelita {
		want[key{m.Date, m.From}] -= m.Qty
		want[key{m.Date, m.To}] += m.Qty
	}
	for _, o := range t.TransfersOut {
		if o.FromFarm == FarmTelita {
			want[key{o.Date, o.FromSection}] -= o.Qty
		}
	}
	for _, c := range t.OtherCorrectionsTelita {
		want[key{c.Date, c.Section}] += c.Qty
	}
	for _, r := range b.x.telitaDaily.Rows {
		k := key{r.Date, r.Section}
		if r.Correction != want[k] {
			return fmt.Errorf("migratehistory: Telița %s %s: register correction %d ≠ moves + outs + corrections %d", r.Section, r.Date, r.Correction, want[k])
		}
		delete(want, k)
	}
	for k, q := range want {
		if q != 0 {
			return fmt.Errorf("migratehistory: Telița %s %s: %d moved with no register row that day", k.section, k.day, q)
		}
	}
	return nil
}

// links plans the Telița → Hîrbovăț transfers (§10.3 "Transfer pairs") from
// links_inferred_by_date: one transfer entry per link, dated on arrival with
// the departure day in its note, one pair per source × destination. Destinations are filled in code order from
// the sources in code order; each pair carries the destination's arrived
// quantity. The link difference (arrived − out) is an adjustment at the
// source on the arrival day, so the source's counts stay the register's and
// Hîrbovăț's base is the arrivals (user answer Q1 (a)).
func (b *builder) links() error {
	type outKey struct {
		day     timex.Day
		section string
		sex     dbq.Sex
	}
	outs := map[outKey]int32{}
	outFarm := map[outKey]string{}
	for _, o := range b.x.transfers.TransfersOut {
		k := outKey{o.Date, o.FromSection, o.Sex}
		outs[k] += o.Qty
		outFarm[k] = o.FromFarm
	}
	arrivals := map[outKey]int32{}
	arrFarm := map[outKey]string{}
	for _, a := range b.x.transfers.Arrivals {
		k := outKey{a.Date, a.Section, a.Sex}
		arrivals[k] += a.Qty
		arrFarm[k] = a.FarmID
	}

	for _, l := range b.x.transfers.Links {
		sources := slices.Sorted(slices.Values(l.From))
		remaining := make([]int32, len(sources))
		var srcFarm string
		var outTotal int32
		for i, s := range sources {
			k := outKey{l.OutDate, s, l.Sex}
			q, ok := outs[k]
			if !ok {
				return fmt.Errorf("migratehistory: link %s %s: no transfer out of %s", l.OutDate, l.Sex, s)
			}
			if srcFarm != "" && outFarm[k] != srcFarm {
				return fmt.Errorf("migratehistory: link %s %s: sources on two farms", l.OutDate, l.Sex)
			}
			srcFarm = outFarm[k]
			remaining[i] = q
			outTotal += q
		}
		if outTotal != l.QtyOut || len(sources) == 0 {
			return fmt.Errorf("migratehistory: link %s %s: transfers out sum to %d, the link says %d", l.OutDate, l.Sex, outTotal, l.QtyOut)
		}
		dests := sortedKeys(l.To)
		var inTotal int32
		var dstFarm string
		for _, d := range dests {
			k := outKey{l.ArrivalDate, d, l.Sex}
			if arrivals[k] != l.To[d] {
				return fmt.Errorf("migratehistory: link %s %s: %s arrival %d, the link says %d", l.ArrivalDate, l.Sex, d, arrivals[k], l.To[d])
			}
			if dstFarm != "" && arrFarm[k] != dstFarm {
				return fmt.Errorf("migratehistory: link %s %s: destinations on two farms", l.ArrivalDate, l.Sex)
			}
			dstFarm = arrFarm[k]
			inTotal += l.To[d]
		}
		if inTotal != l.QtyIn {
			return fmt.Errorf("migratehistory: link %s %s: arrivals sum to %d, the link says %d", l.ArrivalDate, l.Sex, inTotal, l.QtyIn)
		}

		from, err := b.farm(srcFarm)
		if err != nil {
			return err
		}
		to, err := b.farm(dstFarm)
		if err != nil {
			return err
		}
		note := messages.TransferDeparture(fmtx.Day(l.OutDate))
		entry := b.newEntry(dbq.EntryKindTransfer, from, &to, l.ArrivalDate, &note)

		src := 0
		for _, d := range dests {
			need := l.To[d]
			for need > 0 {
				if remaining[src] <= 0 && src < len(sources)-1 {
					src++
					continue
				}
				take := min(need, remaining[src])
				if src == len(sources)-1 {
					take = need // the last source carries what is still to arrive
				}
				if err := b.pair(entry, end{srcFarm, sources[src], l.Sex}, end{dstFarm, d, l.Sex}, l.ArrivalDate, take); err != nil {
					return err
				}
				remaining[src] -= take
				need -= take
			}
		}
		for i, r := range remaining {
			if r != 0 {
				if err := b.adjust(srcFarm, sources[i], l.ArrivalDate, l.Sex, -r); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// telitaResiduals closes Telița's register (§10.3, D25/D63): a section × sex
// that would end a day below 0 gets an adjustment up to 0 that day, and every
// count left on the cycle's last day is adjusted to 0 then, so the cycle
// ends emptied.
func (b *builder) telitaResiduals() error {
	last, err := b.cycleEnd(FarmTelita)
	if err != nil {
		return err
	}
	telita := b.p.farmIDs[FarmTelita]
	code := map[uuid.UUID]string{}
	for _, s := range b.p.Sections {
		code[s.ID] = s.Code
	}
	net := dailyNet(b.p.BirdEvents, func(e dbq.BirdEvent) bool { return e.FarmID == telita })
	keys := make([]sexSection, 0, len(net))
	for k := range net {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, func(x, y sexSection) int {
		return cmp.Or(cmp.Compare(code[x.section], code[y.section]), cmp.Compare(x.sex, y.sex))
	})
	for _, k := range keys {
		var run int64
		days := sortedDays(net[k])
		for _, d := range days {
			run += net[k][d]
			if run < 0 {
				if err := b.adjust(FarmTelita, code[k.section], d, k.sex, int32(-run)); err != nil {
					return err
				}
				run = 0
			}
		}
		if run == 0 {
			continue
		}
		if days[len(days)-1].After(last) {
			return fmt.Errorf("migratehistory: Telița %s %s has rows after the cycle end %s", code[k.section], k.sex, last)
		}
		if err := b.adjust(FarmTelita, code[k.section], last, k.sex, int32(-run)); err != nil {
			return err
		}
	}
	return nil
}

// cycleEnd is the last day of the farm's latest closed cycle.
func (b *builder) cycleEnd(farm string) (timex.Day, error) {
	var last timex.Day
	for _, c := range b.x.cycles.Cycles {
		if c.FarmID == farm && c.EndDate != nil && c.EndDate.After(last) {
			last = *c.EndDate
		}
	}
	if last.IsZero() {
		return last, fmt.Errorf("migratehistory: %s has no closed cycle", farm)
	}
	return last, nil
}

// telitaWeighings plans one weighing per row of weighings-weekly.json (sex
// from the row, redated rows as given).
func (b *builder) telitaWeighings() error {
	entry := b.p.migrationEntry[FarmTelita]
	farm, err := b.farm(FarmTelita)
	if err != nil {
		return err
	}
	for _, r := range b.x.weighings.Rows {
		sec, err := b.section(FarmTelita, r.Section)
		if err != nil {
			return err
		}
		b.p.Weighings = append(b.p.Weighings, dbq.Weighing{
			ID: uuid.New(), EntryID: entry, FarmID: farm, SectionID: sec, Date: r.Date, Sex: r.Sex,
			AvgWeightG: r.AvgWeightG, CvPct: r.CvPct, UniformityPct: r.UniformityPct, SampleSize: r.SampleSize,
		})
	}
	return nil
}
