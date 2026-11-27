package periodtable

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Scopes the day query can sum over (a whitelist: the scope fragment is
// chosen from it, never built from input).
//
// The query is unit-grained (plan M2-D1): every figure is first summed per
// egg unit and day (CTE unit_day: a hall with its compartments, or a section
// without parent and children), and a scope is a set of whole units. So a
// scope can only be the farm or one egg unit. The compartment and section
// scopes of M6 (detail=sections, section pages) don't fit as another entry
// here: a compartment is part of a unit, while the hen-days and the egg
// figures exist only per unit. They need a leaf-first restructure (bird, feed
// and weighing sums per leaf section, egg figures per unit, then each rolled
// up to its scope), not just another whitelist entry.
const (
	ScopeFarm = "farm" // one row per day for the whole farm (all its units)
	ScopeHall = "hall" // one row per egg unit and day: a hall, or a section without parent and children
)

// DaysQuery asks for the day sums of a farm.
type DaysQuery struct {
	FarmID   uuid.UUID
	Scope    string    // ScopeFarm | ScopeHall
	From, To timex.Day // the days to return, both included
	// WithChips asks for Days.Chips, the record types of the period (D45).
	// Off by default: only Build shows chips, so only it pays for their CTE.
	WithChips bool
	// PeriodFrom and PeriodTo bound the record types of Days.Chips (D45: the
	// whole period, not only the returned days). Zero = From and To. Unused
	// without WithChips.
	PeriodFrom, PeriodTo timex.Day
}

// Days is the result of ReadDays.
type Days struct {
	FarmType string   // growth | laying
	Chips    []string // with WithChips, the record types with a live record in the period, in §7.4 order; else []
	Rows     []ScopeDay
}

// ScopeDay is one scope's sums for one day. The guide fields of DayFigures
// are left nil: the caller fills them from the cycle's guide.
type ScopeDay struct {
	ScopeID uuid.UUID // the farm, or the egg unit for ScopeHall
	Code    string    // "" for the farm, the unit's code for ScopeHall
	calc.DayFigures
	// MortalityRows counts the section × sex with a live mortality row that
	// day (a 0 row counts: "entered 0"). EggRows counts the egg units with a
	// filled total. FeedRows counts the live allocations.
	MortalityRows, EggRows, FeedRows int
	// Weighings (§5.10): WeightGrams is Σ avg weight × birds at the end of the
	// weighing day and WeighedBirds Σ those birds, over the day's weighings
	// (Weighings of them), so a row's mean is bird-weighted (A8).
	WeightGrams  calc.PerSex[decimal.Decimal]
	WeighedBirds calc.PerSex[int]
	Weighings    calc.PerSex[int]
	// Legacy mortality of a cycle without a flock count (D52, D54): deaths
	// and rows that are not emptying, and the emptying rows (never deaths).
	LegacyDeaths calc.PerSex[int]
	LegacyRows   calc.PerSex[int]
	EmptyingRows calc.PerSex[int]
}

// ErrScope is returned for a scope outside the whitelist.
var ErrScope = errors.New("periodtable: unknown scope")

// scopeSQL maps each scope to the CTEs that assign whole egg units to scope
// keys: unit_scope(unit_id, scope_id) and scopes(scope_id, scope_code), one
// row per scope even when it has no data. A scope smaller than a unit can't
// be written here (see ScopeHall).
var scopeSQL = map[string]string{
	ScopeFarm: `unit_scope as (
  select distinct l.unit_id, $1::uuid as scope_id from leaf l
), scopes as (
  select $1::uuid as scope_id, ''::text as scope_code
)`,
	ScopeHall: `unit_scope as (
  select distinct l.unit_id, l.unit_id as scope_id from leaf l
), scopes as (
  select us.scope_id, s.code as scope_code from unit_scope us join section s on s.id = us.scope_id
)`,
}

// chipsSQL is the record types (chips, D45) with a live record in the period
// $4–$5, in §7.4 order; noChipsSQL stands for it when they aren't asked for.
const (
	chipsSQL = `array(
           select t.chip from (
             select c.ord, c.chip
             from (values (1, 'mortalitate', 'mortality'::bird_event_type), (2, 'sacrificare', 'cull'),
                          (3, 'eroare_sexare', 'sexing_error'), (4, 'transfer', 'transfer'),
                          (5, 'populare', 'placement'), (6, 'vanzare', 'sale')) c(ord, chip, typ)
             where exists (select 1 from bird_event e join entry n on n.id = e.entry_id and n.deleted_at is null
                           where e.farm_id = $1 and e.deleted_at is null and e.type = c.typ
                             and e.date between $4::date and $5::date)
                or (c.typ = 'mortality' and exists (
                      select 1 from legacy_mortality lm join entry n on n.id = lm.entry_id and n.deleted_at is null
                      where lm.farm_id = $1 and lm.date between $4::date and $5::date))
             union all
             select 7, 'hrana' where exists (
               select 1 from feed_allocation fa join entry n on n.id = fa.entry_id and n.deleted_at is null
               where fa.farm_id = $1 and fa.deleted_at is null and fa.date between $4::date and $5::date)
             union all
             select 8, 'cantarire' where exists (
               select 1 from weighing w join entry n on n.id = w.entry_id and n.deleted_at is null
               where w.farm_id = $1 and w.deleted_at is null and w.date between $4::date and $5::date)
             union all
             select 9, 'oua' where exists (
               select 1 from egg_production ep join entry n on n.id = ep.entry_id and n.deleted_at is null
               where ep.farm_id = $1 and ep.deleted_at is null and ep.date between $4::date and $5::date)
           ) t order by t.ord)::text[]`
	noChipsSQL = `'{}'::text[]`
)

// daysSQL is the PeriodTable day query (§7.4 Performance: one pass over the
// days' deltas plus one running-count seed at the first day). Parameters: $1
// farm, $2–$3 the days, and with chips $4–$5 the period of the chips. The
// placeholders {{chips}} and {{scope}} take chipsSQL (or noChipsSQL) and a
// scopeSQL entry, constants only. The "-- name:" line lets a query tracer
// name it like the sqlc queries.
//
// Rules applied here, once:
//   - alive = the running sum of live bird events (§5.3); alive_start = the end
//     of the day before;
//   - hen-days (§5.6 as amended by M2 Q1): an egg unit's hens at the start of
//     the day count when the unit has a filled egg total that day, or when its
//     hens are gone at the end of the day (the sale or emptying day, H7 on
//     14.11.2026). Egg parts and egg weight count only unit-days with a total;
//   - feed kg = the migrated kg, else g × birds at the start of the day / 1000
//     (DATA03); fed bird-days count the allocations whose kg is above 0, so a
//     register row of 0 kg on a sell-out day (H7 14.11.2026) is not a fed day
//     (golden 21: 142 583 ♀ bird-days in SAPT 58).
const daysSQL = `-- name: PeriodTableDays
with meta as (
  select f.type::text as farm_type,
         {{chips}} as chips
  from farm f
  where f.id = $1
), leaf as (
  select s.id as section_id, coalesce(s.parent_id, s.id) as unit_id
  from section s
  where s.farm_id = $1 and not exists (select 1 from section c where c.parent_id = s.id)
), days as (
  select g::date as date from generate_series($2::date, $3::date, interval '1 day') g
), delta as (
  select x.section_id, x.sex, x.date, x.delta, x.deaths, x.culls, x.sold, x.mortality_entered
  from v_section_sex_delta x
  where x.farm_id = $1 and x.date between $2::date and $3::date
), seed as (
  select x.section_id, x.sex, sum(x.delta) as alive
  from v_section_sex_delta x
  where x.farm_id = $1 and x.date < $2::date
  group by x.section_id, x.sex
), leaf_day as (
  select l.section_id, l.unit_id, sx.sex, d.date,
         coalesce(x.delta, 0) as delta, coalesce(x.deaths, 0) as mortality, coalesce(x.culls, 0) as culls,
         coalesce(x.sold, 0) as sold, coalesce(x.mortality_entered, false) as mortality_entered,
         coalesce(sd.alive, 0)
           + sum(coalesce(x.delta, 0)) over (partition by l.section_id, sx.sex order by d.date) as alive_end
  from leaf l
  cross join (values ('F'::sex), ('M'::sex)) sx(sex)
  cross join days d
  left join seed sd on sd.section_id = l.section_id and sd.sex = sx.sex
  left join delta x on x.section_id = l.section_id and x.sex = sx.sex and x.date = d.date
), feed as (
  select ld.unit_id, ld.date, ld.sex, ld.alive_end - ld.delta as birds_start,
         coalesce(fa.kg_recorded, fa.g_per_bird * (ld.alive_end - ld.delta) / 1000.0) as kg
  from feed_allocation fa
  join entry n on n.id = fa.entry_id and n.deleted_at is null
  join leaf_day ld on ld.section_id = fa.section_id and ld.sex = fa.sex and ld.date = fa.date
  where fa.farm_id = $1 and fa.deleted_at is null and fa.date between $2::date and $3::date
), unit_bird as (
  select unit_id, date,
         sum(alive_end - delta) filter (where sex = 'F') as alive_start_f,
         sum(alive_end - delta) filter (where sex = 'M') as alive_start_m,
         sum(alive_end) filter (where sex = 'F') as alive_end_f,
         sum(alive_end) filter (where sex = 'M') as alive_end_m,
         sum(mortality) filter (where sex = 'F') as mortality_f,
         sum(mortality) filter (where sex = 'M') as mortality_m,
         sum(culls) filter (where sex = 'F') as culls_f,
         sum(culls) filter (where sex = 'M') as culls_m,
         sum(sold) filter (where sex = 'F') as sold_f,
         sum(sold) filter (where sex = 'M') as sold_m,
         count(*) filter (where mortality_entered) as mortality_rows
  from leaf_day
  group by unit_id, date
), unit_feed as (
  select unit_id, date,
         sum(kg) filter (where sex = 'F') as kg_f,
         sum(kg) filter (where sex = 'M') as kg_m,
         sum(birds_start) filter (where sex = 'F' and kg > 0) as fed_f,
         sum(birds_start) filter (where sex = 'M' and kg > 0) as fed_m,
         count(*) as feed_rows
  from feed
  group by unit_id, date
), unit_egg as (
  select ep.section_id as unit_id, ep.date,
         sum(ep.total_eggs) as eggs,
         count(ep.total_eggs) as egg_rows,
         sum(ep.hatching_eggs) filter (where ep.total_eggs is not null) as hatching,
         sum(ep.floor_eggs) filter (where ep.total_eggs is not null) as floor_eggs,
         sum(ep.cat1) filter (where ep.total_eggs is not null) as cat1,
         sum(ep.cat2) filter (where ep.total_eggs is not null) as cat2,
         sum(ep.cat3) filter (where ep.total_eggs is not null) as cat3,
         sum(ep.total_eggs * ep.avg_egg_weight_g) as egg_grams,
         sum(ep.total_eggs) filter (where ep.avg_egg_weight_g is not null) as weighed_eggs
  from egg_production ep
  join entry n on n.id = ep.entry_id and n.deleted_at is null
  where ep.farm_id = $1 and ep.deleted_at is null and ep.date between $2::date and $3::date
  group by ep.section_id, ep.date
), unit_weigh as (
  select ld.unit_id, ld.date,
         sum(w.avg_weight_g * ld.alive_end) filter (where w.sex = 'F') as weight_grams_f,
         sum(w.avg_weight_g * ld.alive_end) filter (where w.sex = 'M') as weight_grams_m,
         sum(ld.alive_end) filter (where w.sex = 'F') as weighed_birds_f,
         sum(ld.alive_end) filter (where w.sex = 'M') as weighed_birds_m,
         count(*) filter (where w.sex = 'F') as weighings_f,
         count(*) filter (where w.sex = 'M') as weighings_m
  from weighing w
  join entry n on n.id = w.entry_id and n.deleted_at is null
  join leaf_day ld on ld.section_id = w.section_id and ld.sex = w.sex and ld.date = w.date
  where w.farm_id = $1 and w.deleted_at is null and w.date between $2::date and $3::date
  group by ld.unit_id, ld.date
), unit_legacy as (
  select l.unit_id, lm.date,
         sum(lm.deaths) filter (where lm.sex = 'F' and not lm.emptying) as legacy_deaths_f,
         sum(lm.deaths) filter (where lm.sex = 'M' and not lm.emptying) as legacy_deaths_m,
         count(*) filter (where lm.sex = 'F' and not lm.emptying) as legacy_rows_f,
         count(*) filter (where lm.sex = 'M' and not lm.emptying) as legacy_rows_m,
         count(*) filter (where lm.sex = 'F' and lm.emptying) as emptying_rows_f,
         count(*) filter (where lm.sex = 'M' and lm.emptying) as emptying_rows_m
  from legacy_mortality lm
  join entry n on n.id = lm.entry_id and n.deleted_at is null
  join leaf l on l.section_id = lm.section_id
  where lm.farm_id = $1 and lm.date between $2::date and $3::date
  group by l.unit_id, lm.date
), unit_day as (
  select b.unit_id, b.date, b.alive_start_f, b.alive_start_m, b.alive_end_f, b.alive_end_m,
         b.mortality_f, b.mortality_m, b.culls_f, b.culls_m, b.sold_f, b.sold_m, b.mortality_rows,
         e.eggs, e.egg_rows, e.hatching, e.floor_eggs, e.cat1, e.cat2, e.cat3, e.egg_grams, e.weighed_eggs,
         f.kg_f, f.kg_m, f.fed_f, f.fed_m, f.feed_rows,
         wg.weight_grams_f, wg.weight_grams_m, wg.weighed_birds_f, wg.weighed_birds_m, wg.weighings_f, wg.weighings_m,
         lg.legacy_deaths_f, lg.legacy_deaths_m, lg.legacy_rows_f, lg.legacy_rows_m, lg.emptying_rows_f, lg.emptying_rows_m,
         case when coalesce(e.egg_rows, 0) > 0 or b.alive_end_f = 0 then b.alive_start_f else 0 end as hen_days
  from unit_bird b
  left join unit_egg e on e.unit_id = b.unit_id and e.date = b.date
  left join unit_feed f on f.unit_id = b.unit_id and f.date = b.date
  left join unit_weigh wg on wg.unit_id = b.unit_id and wg.date = b.date
  left join unit_legacy lg on lg.unit_id = b.unit_id and lg.date = b.date
), {{scope}}
select sc.scope_id, sc.scope_code, d.date,
       coalesce(sum(u.alive_start_f), 0)::integer as alive_start_f,
       coalesce(sum(u.alive_start_m), 0)::integer as alive_start_m,
       coalesce(sum(u.alive_end_f), 0)::integer as alive_end_f,
       coalesce(sum(u.alive_end_m), 0)::integer as alive_end_m,
       coalesce(sum(u.mortality_f), 0)::integer as mortality_f,
       coalesce(sum(u.mortality_m), 0)::integer as mortality_m,
       coalesce(sum(u.culls_f), 0)::integer as culls_f,
       coalesce(sum(u.culls_m), 0)::integer as culls_m,
       coalesce(sum(u.sold_f), 0)::integer as sold_f,
       coalesce(sum(u.sold_m), 0)::integer as sold_m,
       coalesce(sum(u.mortality_rows), 0)::integer as mortality_rows,
       coalesce(sum(u.eggs), 0)::integer as eggs,
       coalesce(sum(u.egg_rows), 0)::integer as egg_rows,
       coalesce(sum(u.hatching), 0)::integer as hatching,
       coalesce(sum(u.floor_eggs), 0)::integer as floor_eggs,
       coalesce(sum(u.cat1), 0)::integer as cat1,
       coalesce(sum(u.cat2), 0)::integer as cat2,
       coalesce(sum(u.cat3), 0)::integer as cat3,
       coalesce(sum(u.egg_grams), 0)::numeric as egg_grams,
       coalesce(sum(u.weighed_eggs), 0)::integer as weighed_eggs,
       coalesce(sum(u.hen_days), 0)::integer as hen_days,
       coalesce(sum(u.kg_f), 0)::numeric as kg_f,
       coalesce(sum(u.kg_m), 0)::numeric as kg_m,
       coalesce(sum(u.fed_f), 0)::integer as fed_f,
       coalesce(sum(u.fed_m), 0)::integer as fed_m,
       coalesce(sum(u.feed_rows), 0)::integer as feed_rows,
       coalesce(sum(u.weight_grams_f), 0)::numeric as weight_grams_f,
       coalesce(sum(u.weight_grams_m), 0)::numeric as weight_grams_m,
       coalesce(sum(u.weighed_birds_f), 0)::integer as weighed_birds_f,
       coalesce(sum(u.weighed_birds_m), 0)::integer as weighed_birds_m,
       coalesce(sum(u.weighings_f), 0)::integer as weighings_f,
       coalesce(sum(u.weighings_m), 0)::integer as weighings_m,
       coalesce(sum(u.legacy_deaths_f), 0)::integer as legacy_deaths_f,
       coalesce(sum(u.legacy_deaths_m), 0)::integer as legacy_deaths_m,
       coalesce(sum(u.legacy_rows_f), 0)::integer as legacy_rows_f,
       coalesce(sum(u.legacy_rows_m), 0)::integer as legacy_rows_m,
       coalesce(sum(u.emptying_rows_f), 0)::integer as emptying_rows_f,
       coalesce(sum(u.emptying_rows_m), 0)::integer as emptying_rows_m,
       meta.farm_type, meta.chips
from meta
cross join scopes sc
cross join days d
left join unit_scope us on us.scope_id = sc.scope_id
left join unit_day u on u.unit_id = us.unit_id and u.date = d.date
group by sc.scope_id, sc.scope_code, d.date, meta.farm_type, meta.chips
order by sc.scope_code, d.date`

// dayRow is one row of daysSQL.
type dayRow struct {
	ScopeID       uuid.UUID       `db:"scope_id"`
	ScopeCode     string          `db:"scope_code"`
	Date          timex.Day       `db:"date"`
	AliveStartF   int             `db:"alive_start_f"`
	AliveStartM   int             `db:"alive_start_m"`
	AliveEndF     int             `db:"alive_end_f"`
	AliveEndM     int             `db:"alive_end_m"`
	MortalityF    int             `db:"mortality_f"`
	MortalityM    int             `db:"mortality_m"`
	CullsF        int             `db:"culls_f"`
	CullsM        int             `db:"culls_m"`
	SoldF         int             `db:"sold_f"`
	SoldM         int             `db:"sold_m"`
	MortalityRows int             `db:"mortality_rows"`
	Eggs          int             `db:"eggs"`
	EggRows       int             `db:"egg_rows"`
	Hatching      int             `db:"hatching"`
	FloorEggs     int             `db:"floor_eggs"`
	Cat1          int             `db:"cat1"`
	Cat2          int             `db:"cat2"`
	Cat3          int             `db:"cat3"`
	EggGrams      decimal.Decimal `db:"egg_grams"`
	WeighedEggs   int             `db:"weighed_eggs"`
	HenDays       int             `db:"hen_days"`
	KgF           decimal.Decimal `db:"kg_f"`
	KgM           decimal.Decimal `db:"kg_m"`
	FedF          int             `db:"fed_f"`
	FedM          int             `db:"fed_m"`
	FeedRows      int             `db:"feed_rows"`
	WeightGramsF  decimal.Decimal `db:"weight_grams_f"`
	WeightGramsM  decimal.Decimal `db:"weight_grams_m"`
	WeighedBirdsF int             `db:"weighed_birds_f"`
	WeighedBirdsM int             `db:"weighed_birds_m"`
	WeighingsF    int             `db:"weighings_f"`
	WeighingsM    int             `db:"weighings_m"`
	LegacyDeathsF int             `db:"legacy_deaths_f"`
	LegacyDeathsM int             `db:"legacy_deaths_m"`
	LegacyRowsF   int             `db:"legacy_rows_f"`
	LegacyRowsM   int             `db:"legacy_rows_m"`
	EmptyingRowsF int             `db:"emptying_rows_f"`
	EmptyingRowsM int             `db:"emptying_rows_m"`
	FarmType      string          `db:"farm_type"`
	Chips         []string        `db:"chips"`
}

// ReadDays runs the day query: one statement, one row per scope and day of
// [From, To], every day present (zeros when nothing happened), and the
// period's chips when q.WithChips. An unknown farm gives no rows.
func ReadDays(ctx context.Context, db dbq.DBTX, q DaysQuery) (Days, error) {
	scope, ok := scopeSQL[q.Scope]
	if !ok {
		return Days{}, fmt.Errorf("%w: %q", ErrScope, q.Scope)
	}
	if q.From.After(q.To) {
		return Days{}, fmt.Errorf("periodtable: days %s–%s start after they end", q.From, q.To)
	}
	chips, args := noChipsSQL, []any{q.FarmID, q.From, q.To}
	if q.WithChips {
		pFrom, pTo := q.PeriodFrom, q.PeriodTo
		if pFrom.IsZero() || pTo.IsZero() {
			pFrom, pTo = q.From, q.To
		}
		chips, args = chipsSQL, append(args, pFrom, pTo)
	}
	sql := strings.NewReplacer("{{chips}}", chips, "{{scope}}", scope).Replace(daysSQL)
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return Days{}, fmt.Errorf("periodtable: days %s–%s: %w", q.From, q.To, err)
	}
	raw, err := pgx.CollectRows(rows, pgx.RowToStructByName[dayRow])
	if err != nil {
		return Days{}, fmt.Errorf("periodtable: days %s–%s: %w", q.From, q.To, err)
	}
	out := Days{Chips: []string{}, Rows: make([]ScopeDay, 0, len(raw))}
	for _, r := range raw {
		out.FarmType, out.Chips = r.FarmType, r.Chips
		out.Rows = append(out.Rows, r.scopeDay())
	}
	return out, nil
}

func (r dayRow) scopeDay() ScopeDay {
	ps := func(f, m int) calc.PerSex[int] { return calc.PerSex[int]{F: f, M: m} }
	return ScopeDay{
		ScopeID: r.ScopeID,
		Code:    r.ScopeCode,
		DayFigures: calc.DayFigures{
			Date:        r.Date,
			AliveStart:  ps(r.AliveStartF, r.AliveStartM),
			AliveEnd:    ps(r.AliveEndF, r.AliveEndM),
			Mortality:   ps(r.MortalityF, r.MortalityM),
			Culls:       ps(r.CullsF, r.CullsM),
			Sold:        ps(r.SoldF, r.SoldM),
			Eggs:        r.Eggs,
			Hatching:    r.Hatching,
			Floor:       r.FloorEggs,
			Cat1:        r.Cat1,
			Cat2:        r.Cat2,
			Cat3:        r.Cat3,
			EggGrams:    r.EggGrams,
			WeighedEggs: r.WeighedEggs,
			HenDays:     r.HenDays,
			FeedKg:      calc.PerSex[decimal.Decimal]{F: r.KgF, M: r.KgM},
			FedBirdDays: ps(r.FedF, r.FedM),
		},
		MortalityRows: r.MortalityRows,
		EggRows:       r.EggRows,
		FeedRows:      r.FeedRows,
		WeightGrams:   calc.PerSex[decimal.Decimal]{F: r.WeightGramsF, M: r.WeightGramsM},
		WeighedBirds:  ps(r.WeighedBirdsF, r.WeighedBirdsM),
		Weighings:     ps(r.WeighingsF, r.WeighingsM),
		LegacyDeaths:  ps(r.LegacyDeathsF, r.LegacyDeathsM),
		LegacyRows:    ps(r.LegacyRowsF, r.LegacyRowsM),
		EmptyingRows:  ps(r.EmptyingRowsF, r.EmptyingRowsM),
	}
}

// registerStartSQL is the first day with a register record (legacy mortality
// or eggs) of a farm in [$2, $3]: where a cycle without a flock count starts
// its table (M2 Q7: Seria 2025's register starts 04.04.2025, SAPT 23).
const registerStartSQL = `-- name: PeriodTableRegisterStart
select least(
  (select min(lm.date) from legacy_mortality lm join entry n on n.id = lm.entry_id and n.deleted_at is null
   where lm.farm_id = $1 and lm.date between $2::date and $3::date),
  (select min(ep.date) from egg_production ep join entry n on n.id = ep.entry_id and n.deleted_at is null
   where ep.farm_id = $1 and ep.deleted_at is null and ep.date between $2::date and $3::date))`

// registerStart returns the first register day of the farm in [from, to],
// or nil when it has none.
func registerStart(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, from, to timex.Day) (*timex.Day, error) {
	var d *timex.Day
	if err := db.QueryRow(ctx, registerStartSQL, farmID, from, to).Scan(&d); err != nil {
		return nil, fmt.Errorf("periodtable: register start: %w", err)
	}
	return d, nil
}
