package migratehistory

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// eggRow is one hall-day of eggs as the register has it.
type eggRow struct {
	hall                         string
	day                          timex.Day
	floor, cat1, cat2, cat3, inc *int32
	total                        *int32 // the register's total (Cat. 1–3 + Incubabile), checked
	weight                       decimal.NullDecimal
}

// addEggs plans one egg_production row on the farm's migration entry. The
// register's total must equal Cat. 1 + 2 + 3 + Incubabile (D18), because the
// database derives total_eggs from them. A weight of 0 is "not weighed".
func (b *builder) addEggs(farm string, r eggRow) error {
	if r.cat1 == nil || r.cat2 == nil || r.cat3 == nil || r.inc == nil || r.total == nil ||
		*r.cat1+*r.cat2+*r.cat3+*r.inc != *r.total {
		return fmt.Errorf("migratehistory: eggs %s %s: total is not Cat. 1–3 + Incubabile", r.hall, r.day)
	}
	farmID, err := b.farm(farm)
	if err != nil {
		return err
	}
	sec, err := b.section(farm, r.hall)
	if err != nil {
		return err
	}
	if r.weight.Valid && r.weight.Decimal.IsZero() {
		r.weight = decimal.NullDecimal{}
	}
	b.p.EggProductions = append(b.p.EggProductions, dbq.EggProduction{
		ID: uuid.New(), EntryID: b.p.migrationEntry[farm], FarmID: farmID, SectionID: sec, Date: r.day,
		FloorEggs: r.floor, Cat1: r.cat1, Cat2: r.cat2, Cat3: r.cat3, HatchingEggs: r.inc, AvgEggWeightG: r.weight,
	})
	return nil
}

// eggs plans Seria 2026's eggs: one row per Hîrbovăț hall-day with eggs
// (hall-days without eggs get no row).
func (b *builder) eggs() error {
	for _, r := range b.x.hirbovat.hallDay {
		if r.Eggs == nil {
			continue
		}
		err := b.addEggs(FarmHirbovat, eggRow{
			hall: r.Hall, day: r.Date, floor: r.Jos, cat1: r.Cat1, cat2: r.Cat2, cat3: r.Cat3, inc: r.Inc,
			total: r.Eggs, weight: r.EggG,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// seria2025 plans the previous Hîrbovăț flock (§10.3 "Seria 2025", D158c):
// eggs per hall-day and legacy_mortality per compartment-day and sex, the old
// codes H1/H3/H4 mapped to H5/H6/H7; depopulation days are `emptying`.
func (b *builder) seria2025() error {
	codes := b.x.farms.Meta.CodeMapOldToNew
	newCode := func(old string) (string, error) {
		c, ok := codes[old]
		if !ok {
			return "", fmt.Errorf("migratehistory: Seria 2025: old code %q has no new code", old)
		}
		return c, nil
	}
	farm, err := b.farm(FarmHirbovat)
	if err != nil {
		return err
	}
	entry := b.p.migrationEntry[FarmHirbovat]
	for _, r := range b.x.prevEggs.Rows {
		hall, err := newCode(r.Hall)
		if err != nil {
			return err
		}
		err = b.addEggs(FarmHirbovat, eggRow{
			hall: hall, day: r.Date, floor: r.FloorEggs, cat1: r.Cat1, cat2: r.Cat2, cat3: r.Cat3,
			inc: r.HatchingEggs, total: r.TotalEggs, weight: r.AvgEggWeightG,
		})
		if err != nil {
			return err
		}
		for _, old := range sortedKeys(r.Mortality) {
			code, err := newCode(old)
			if err != nil {
				return err
			}
			sec, err := b.section(FarmHirbovat, code)
			if err != nil {
				return err
			}
			m := r.Mortality[old]
			for _, s := range []struct {
				sex    dbq.Sex
				deaths int32
			}{{dbq.SexF, m.F}, {dbq.SexM, m.M}} {
				if s.deaths < 0 {
					// Three ♂ cells are negative: birds moved in, not deaths.
					// They are skipped (§10.3, user answer Q-T4-1).
					continue
				}
				b.p.LegacyMortality = append(b.p.LegacyMortality, dbq.LegacyMortality{
					ID: uuid.New(), EntryID: entry, FarmID: farm, SectionID: sec, Date: r.Date, Sex: s.sex,
					Deaths: s.deaths, Emptying: r.Depopulation,
				})
			}
		}
	}
	return nil
}
