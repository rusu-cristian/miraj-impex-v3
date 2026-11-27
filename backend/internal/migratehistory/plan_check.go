package migratehistory

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// validate checks the plan against the rules the database would enforce on
// COPY (references, signs, ranges, one live value per key) and the rule it
// can't (counts never below 0, §6.1), so a broken extract fails here with the
// row named instead of as a constraint error inside the import transaction.
func (p *Plan) validate() error {
	var errs []error
	fail := func(format string, args ...any) {
		if len(errs) < 20 {
			errs = append(errs, fmt.Errorf(format, args...))
		}
	}

	users := map[uuid.UUID]bool{}
	for _, u := range p.Users {
		users[u.ID] = true
	}
	entries := map[uuid.UUID]dbq.Entry{}
	for _, e := range p.Entries {
		entries[e.ID] = e
		for _, by := range []*uuid.UUID{e.CreatedBy, e.UpdatedBy} {
			if by != nil && !users[*by] {
				fail("entry %s %s: author is not a planned user", e.Kind, e.Date)
			}
		}
		if (e.Kind == dbq.EntryKindTransfer) != (e.PeerFarmID != nil) {
			fail("entry %s %s: peer farm set iff transfer", e.Kind, e.Date)
		}
	}
	p.eachRow(func(entry uuid.UUID, day timex.Day, table string) {
		if _, ok := entries[entry]; !ok {
			fail("%s row on %s: entry %s is not planned", table, day, entry)
		}
		if day.IsZero() {
			fail("%s row without a date", table)
		}
	})

	sections := map[uuid.UUID]dbq.Section{}
	isHall := map[uuid.UUID]bool{}
	for _, s := range p.Sections {
		sections[s.ID] = s
		if s.ParentID != nil {
			isHall[*s.ParentID] = true
		}
	}
	feedTypes := map[uuid.UUID]bool{}
	for _, f := range p.FeedTypes {
		feedTypes[f.ID] = true
	}
	label := func(id uuid.UUID) string { return sections[id].Code }
	// birdSection checks a row that must sit on a section holding birds.
	birdSection := func(table string, farm, sec uuid.UUID, day timex.Day) {
		s, ok := sections[sec]
		switch {
		case !ok:
			fail("%s %s: unknown section", table, day)
		case s.FarmID != farm:
			fail("%s %s %s: section of another farm", table, s.Code, day)
		case isHall[sec]:
			fail("%s %s %s: a hall holds no birds", table, s.Code, day)
		}
	}
	type dayKey struct {
		section uuid.UUID
		day     timex.Day
		sex     dbq.Sex
		kind    string
	}
	seen := map[dayKey]bool{}
	unique := func(table string, k dayKey) {
		if seen[k] {
			fail("%s %s %s %s: two live values for one key", table, label(k.section), k.day, k.sex)
		}
		seen[k] = true
	}

	for _, e := range p.BirdEvents {
		birdSection("bird_event", e.FarmID, e.SectionID, e.Date)
		ok := false
		switch e.Type {
		case dbq.BirdEventTypePlacement:
			ok = e.Qty >= 0
		case dbq.BirdEventTypeMortality, dbq.BirdEventTypeCull, dbq.BirdEventTypeSale:
			ok = e.Qty <= 0
		case dbq.BirdEventTypeSexingError, dbq.BirdEventTypeTransfer, dbq.BirdEventTypeAdjustment:
			ok = true
		}
		if !ok {
			fail("bird_event %s %s %s %s: qty %d has the wrong sign", e.Type, label(e.SectionID), e.Date, e.Sex, e.Qty)
		}
		if (e.PeerSectionID != nil) != (e.Type == dbq.BirdEventTypeTransfer) {
			fail("bird_event %s %s %s: peer section set iff transfer", e.Type, label(e.SectionID), e.Date)
		} else if e.PeerSectionID != nil {
			if _, ok := sections[*e.PeerSectionID]; !ok {
				fail("bird_event transfer %s %s: unknown peer section", label(e.SectionID), e.Date)
			}
		}
		reasonOK := e.CullReason == nil
		if e.Type == dbq.BirdEventTypeCull {
			reasonOK = e.Qty == 0 || e.CullReason != nil
		}
		if !reasonOK {
			fail("bird_event %s %s %s: a cull needs a reason, other rows none", e.Type, label(e.SectionID), e.Date)
		}
		switch e.Type {
		case dbq.BirdEventTypeMortality, dbq.BirdEventTypeCull, dbq.BirdEventTypeSexingError:
			unique("bird_event", dayKey{e.SectionID, e.Date, e.Sex, string(e.Type)})
		}
	}

	for _, r := range p.EggProductions {
		s, ok := sections[r.SectionID]
		switch {
		case !ok || s.FarmID != r.FarmID:
			fail("egg_production %s: unknown section of the farm", r.Date)
		case !isHall[r.SectionID] && s.ParentID != nil:
			fail("egg_production %s %s: not an egg unit", s.Code, r.Date)
		}
		if r.TotalEggs != nil {
			fail("egg_production %s %s: total_eggs is generated", s.Code, r.Date)
		}
		if r.AvgEggWeightG.Valid && !between(r.AvgEggWeightG.Decimal, 20, 120) {
			fail("egg_production %s %s: egg weight %s outside 20–120", s.Code, r.Date, r.AvgEggWeightG.Decimal)
		}
		for _, n := range []*int32{r.FloorEggs, r.Cat1, r.Cat2, r.Cat3, r.HatchingEggs} {
			if n != nil && *n < 0 {
				fail("egg_production %s %s: negative count", s.Code, r.Date)
			}
		}
		unique("egg_production", dayKey{r.SectionID, r.Date, "", "egg"})
	}

	for _, r := range p.FeedAllocations {
		birdSection("feed_allocation", r.FarmID, r.SectionID, r.Date)
		if !feedTypes[r.FeedTypeID] {
			fail("feed_allocation %s %s: unknown feed type", label(r.SectionID), r.Date)
		}
		if !r.GPerBird.Valid && !r.KgRecorded.Valid {
			fail("feed_allocation %s %s %s: neither g nor kg", label(r.SectionID), r.Date, r.Sex)
		}
		if r.GPerBird.Valid && !between(r.GPerBird.Decimal, 1, 300) {
			fail("feed_allocation %s %s %s: g %s outside 1–300", label(r.SectionID), r.Date, r.Sex, r.GPerBird.Decimal)
		}
		if r.KgRecorded.Valid && r.KgRecorded.Decimal.IsNegative() {
			fail("feed_allocation %s %s %s: negative kg", label(r.SectionID), r.Date, r.Sex)
		}
		unique("feed_allocation", dayKey{r.SectionID, r.Date, r.Sex, "feed"})
	}

	for _, r := range p.FeedDeliveries {
		if !feedTypes[r.FeedTypeID] {
			fail("feed_delivery %s: unknown feed type", r.Date)
		}
		if !r.QtyKg.IsPositive() {
			fail("feed_delivery %s: qty %s not > 0", r.Date, r.QtyKg)
		}
		priced := r.PricePerKg.Valid && r.TotalLei.Valid && r.PriceFromRegister &&
			r.PricePerKg.Decimal.IsPositive() && r.TotalLei.Decimal.IsPositive()
		unpriced := !r.PricePerKg.Valid && !r.TotalLei.Valid && r.PricedAt == nil && r.PriceInput == nil
		if !priced && !unpriced {
			fail("feed_delivery %s %s kg: price is neither complete nor empty", r.Date, r.QtyKg)
		}
	}

	for _, r := range p.FeedStockAdjustments {
		if !feedTypes[r.FeedTypeID] || r.QtyKg.IsZero() || r.Note == "" {
			fail("feed_stock_adjustment %s: needs a feed type, a non-zero qty and a note", r.Date)
		}
	}

	for _, r := range p.Weighings {
		birdSection("weighing", r.FarmID, r.SectionID, r.Date)
		if !between(r.AvgWeightG, 1, 10000) {
			fail("weighing %s %s: weight %s outside 1–10000", label(r.SectionID), r.Date, r.AvgWeightG)
		}
		unique("weighing", dayKey{r.SectionID, r.Date, r.Sex, "weighing"})
	}

	for _, r := range p.LegacyMortality {
		birdSection("legacy_mortality", r.FarmID, r.SectionID, r.Date)
		if r.Deaths < 0 {
			fail("legacy_mortality %s %s: negative deaths", label(r.SectionID), r.Date)
		}
		unique("legacy_mortality", dayKey{r.SectionID, r.Date, r.Sex, "legacy"})
	}

	// §6.1: no section × sex ever ends a day below 0.
	net := dailyNet(p.BirdEvents, func(dbq.BirdEvent) bool { return true })
	keys := make([]sexSection, 0, len(net))
	for k := range net {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, func(a, b sexSection) int {
		return cmp.Or(cmp.Compare(label(a.section), label(b.section)), cmp.Compare(a.sex, b.sex))
	})
	for _, k := range keys {
		var run int64
		for _, d := range sortedDays(net[k]) {
			run += net[k][d]
			if run < 0 {
				fail("%s %s goes below 0 on %s (%d)", label(k.section), k.sex, d, run)
				break
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("migratehistory: the plan breaks %d rule(s): %w", len(errs), errors.Join(errs...))
	}
	return nil
}

func between(d decimal.Decimal, lo, hi int64) bool {
	return d.GreaterThanOrEqual(decimal.NewFromInt(lo)) && d.LessThanOrEqual(decimal.NewFromInt(hi))
}
