package migratehistory

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/textkey"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Feed type keys (feed-types.json ids) the §10.3 rules name.
const (
	FeedStarter         = "ft-starter"
	FeedGrower          = "ft-grower"
	FeedPreOuat         = "ft-preouat"
	FeedBreeder2        = "ft-breeder2"
	FeedCocosi          = "ft-cocosi"
	hirbovatHenFeed     = FeedBreeder2
	hirbovatRoosterFeed = FeedCocosi
)

// telitaFeedWindows is §10.3's Telița feed type by date: Starter to
// 13.11.2025, Grower 14.11.2025–19.01.2026, Pre-ouat from 20.01.2026
// (assumption from the delivery notes). Each window starts on its day.
var telitaFeedWindows = []struct {
	from timex.Day
	key  string
}{
	{timex.Day{}, FeedStarter},
	{timex.MustDay(2025, 11, 14), FeedGrower},
	{timex.MustDay(2026, 1, 20), FeedPreOuat},
}

func telitaFeedType(day timex.Day) string {
	key := ""
	for _, w := range telitaFeedWindows {
		if !day.Before(w.from) {
			key = w.key
		}
	}
	return key
}

func (b *builder) feedTypes() error {
	for _, f := range b.x.feedTypes.Rows {
		row := dbq.FeedType{
			ID: uuid.New(), Name: f.Name, NameKey: textkey.Key(f.Name), ForSex: f.ForSex, Active: f.Active, CreatedAt: b.now,
		}
		b.p.feedTypeIDs[f.ID] = row.ID
		b.p.FeedTypes = append(b.p.FeedTypes, row)
	}
	return nil
}

func (b *builder) feedType(key string) (uuid.UUID, error) {
	id, ok := b.p.feedTypeIDs[key]
	if !ok {
		return uuid.Nil, fmt.Errorf("migratehistory: unknown feed type %q", key)
	}
	return id, nil
}

// noKg reports a register row with g but no kg: it is not imported, so the
// day stays "– fără hrană în registru" / lipsă (§10.3: user answer Q6 for the
// 14 Hîrbovăț rows, user answer 27.11.2026 for Telița H5-2 ♂ 22–23.02.2026).
func noKg(g, kg decimal.NullDecimal) bool {
	return g.Valid && (!kg.Valid || kg.Decimal.IsZero())
}

// feedAllocations plans feed_allocation rows: g_per_bird as typed (null on
// Telița's first days) and kg_recorded = the register's kg (D163).
func (b *builder) feedAllocations() error {
	add := func(farm, code string, day timex.Day, sex dbq.Sex, feedKey string, g, kg decimal.NullDecimal) error {
		if noKg(g, kg) || (!g.Valid && !kg.Valid) {
			return nil
		}
		farmID, err := b.farm(farm)
		if err != nil {
			return err
		}
		sec, err := b.section(farm, code)
		if err != nil {
			return err
		}
		ft, err := b.feedType(feedKey)
		if err != nil {
			return err
		}
		b.p.FeedAllocations = append(b.p.FeedAllocations, dbq.FeedAllocation{
			ID: uuid.New(), EntryID: b.p.migrationEntry[farm], FarmID: farmID, SectionID: sec, Date: day,
			Sex: sex, FeedTypeID: ft, GPerBird: g, KgRecorded: kg,
		})
		return nil
	}
	for _, d := range b.x.telitaFeed.Days {
		for _, code := range sortedKeys(d.Sections) {
			s := d.Sections[code]
			if err := add(FarmTelita, code, d.Date, s.Sex, telitaFeedType(d.Date), s.GPerBird, s.Kg); err != nil {
				return err
			}
		}
	}
	for _, r := range b.x.hirbovat.sectionDay {
		key := hirbovatHenFeed
		if r.Sex == dbq.SexM {
			key = hirbovatRoosterFeed
		}
		if err := add(FarmHirbovat, r.Section, r.Date, r.Sex, key, r.G, r.Kg); err != nil {
			return err
		}
	}
	return nil
}

// deliveries plans feed_delivery rows (priced ones "din registru", D163; the
// unpriced stay "fără preț") and Hîrbovăț's stock adjustments: "mutare ♀ → ♂"
// on both rows of a stock move, "diferență registru" on the others.
func (b *builder) deliveries() error {
	for _, r := range b.x.deliveries.Rows {
		farm, err := b.farm(r.FarmID)
		if err != nil {
			return err
		}
		ft, err := b.feedType(r.FeedTypeID)
		if err != nil {
			return err
		}
		row := dbq.FeedDelivery{
			ID: uuid.New(), EntryID: b.p.migrationEntry[r.FarmID], FarmID: farm, FeedTypeID: ft,
			Date: r.Date, QtyKg: r.QtyKg, Note: r.WorkbookNote,
		}
		if r.Priced {
			if !r.PricePerKg.Valid || !r.TotalLei.Valid {
				return fmt.Errorf("migratehistory: delivery %s %s kg: priced without lei/kg and total", r.Date, r.QtyKg)
			}
			row.PricePerKg, row.TotalLei, row.PriceFromRegister = r.PricePerKg, r.TotalLei, true
		}
		b.p.FeedDeliveries = append(b.p.FeedDeliveries, row)
	}
	farm, err := b.farm(FarmHirbovat)
	if err != nil {
		return err
	}
	adjustments := b.x.deliveries.Meta.HirbovatAdjustments
	// A same-day ♀ −q / ♂ +q pair is feed moved from one stock to the other;
	// any other adjustment is a register difference (§10.3).
	moved := map[timex.Day]bool{}
	for _, f := range adjustments {
		for _, m := range adjustments {
			if f.Date == m.Date && f.ForSex == dbq.SexF && m.ForSex == dbq.SexM && f.QtyKg.Add(m.QtyKg).IsZero() && f.QtyKg.IsNegative() {
				moved[f.Date] = true
			}
		}
	}
	for _, a := range adjustments {
		note := messages.RegisterDifference()
		if moved[a.Date] {
			note = messages.FeedStockMove()
		}
		key := hirbovatHenFeed
		if a.ForSex == dbq.SexM {
			key = hirbovatRoosterFeed
		}
		ft, err := b.feedType(key)
		if err != nil {
			return err
		}
		b.p.FeedStockAdjustments = append(b.p.FeedStockAdjustments, dbq.FeedStockAdjustment{
			ID: uuid.New(), EntryID: b.p.migrationEntry[FarmHirbovat], FarmID: farm, FeedTypeID: ft,
			Date: a.Date, QtyKg: a.QtyKg, Note: note,
		})
	}
	return nil
}
