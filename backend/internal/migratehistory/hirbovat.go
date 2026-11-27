package migratehistory

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// hirbovatBirds plans Hîrbovăț's Seria 2026 ledger from hirbovat-daily.json
// (§10.3 "Hîrbovăț birds"): mortality per compartment-day (0 rows kept), the
// sales as one entry per day × hall (DATA04: 7 records) and the 22.10.2026
// move as a transfer pair. The arrivals come from the Telița links.
func (b *builder) hirbovatBirds() error {
	entry := b.p.migrationEntry[FarmHirbovat]
	for _, r := range b.x.hirbovat.sectionDay {
		if err := b.bird(entry, FarmHirbovat, r.Section, r.Date, r.Sex, dbq.BirdEventTypeMortality, -r.Mort, nil, nil); err != nil {
			return err
		}
	}
	if err := b.sales(); err != nil {
		return err
	}
	farm, err := b.farm(FarmHirbovat)
	if err != nil {
		return err
	}
	for _, m := range b.x.transfers.MovesHirbovat {
		e := b.newEntry(dbq.EntryKindTransfer, farm, &farm, m.Date, nil)
		if err := b.pair(e, end{FarmHirbovat, m.From, m.Sex}, end{FarmHirbovat, m.To, m.Sex}, m.Date, m.Qty); err != nil {
			return err
		}
	}
	return nil
}

// sales plans the Vânzare records: the compartments of one hall sold on one
// day are one sale entry (D168), with a `sale` row per compartment × sex that
// sold birds. The sales block must agree with section_day's `sold`.
func (b *builder) sales() error {
	type soldKey struct {
		day     timex.Day
		section string
		sex     dbq.Sex
	}
	sold := map[soldKey]int32{}
	for _, r := range b.x.hirbovat.sectionDay {
		if r.Sold != 0 {
			sold[soldKey{r.Date, r.Section, r.Sex}] += r.Sold
		}
	}
	rows := slices.Clone(b.x.hirbovat.sales)
	for _, r := range rows {
		for _, s := range []struct {
			sex dbq.Sex
			qty int32
		}{{dbq.SexF, r.F}, {dbq.SexM, r.M}} {
			k := soldKey{r.Date, r.Section, s.sex}
			if sold[k] != s.qty {
				return fmt.Errorf("migratehistory: Hîrbovăț sale %s %s %s: %d in sales, %d in section_day", r.Section, r.Date, s.sex, s.qty, sold[k])
			}
			delete(sold, k)
		}
	}
	if len(sold) > 0 {
		return fmt.Errorf("migratehistory: Hîrbovăț: %d compartment-day(s) sold in section_day with no sales row", len(sold))
	}

	hall := func(code string) string { return b.parent[sectionKey{FarmHirbovat, code}] }
	slices.SortStableFunc(rows, func(x, y hirbovatSale) int {
		if c := x.Date.Compare(y.Date); c != 0 {
			return c
		}
		if c := cmp.Compare(hall(x.Section), hall(y.Section)); c != 0 {
			return c
		}
		return cmp.Compare(x.Section, y.Section)
	})
	farm, err := b.farm(FarmHirbovat)
	if err != nil {
		return err
	}
	var entry uuid.UUID
	for i, r := range rows {
		if hall(r.Section) == "" {
			return fmt.Errorf("migratehistory: Hîrbovăț sale %s %s: not a compartment", r.Section, r.Date)
		}
		if i == 0 || rows[i-1].Date != r.Date || hall(rows[i-1].Section) != hall(r.Section) {
			entry = b.newEntry(dbq.EntryKindSale, farm, nil, r.Date, nil)
		}
		for _, s := range []struct {
			sex dbq.Sex
			qty int32
		}{{dbq.SexF, r.F}, {dbq.SexM, r.M}} {
			if s.qty == 0 {
				continue
			}
			if err := b.bird(entry, FarmHirbovat, r.Section, r.Date, s.sex, dbq.BirdEventTypeSale, -s.qty, nil, nil); err != nil {
				return err
			}
		}
	}
	return nil
}
