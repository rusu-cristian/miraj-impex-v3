package periodtable

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// The Consum day query (view=report-feed, M3 D11). It is leaf-grained, unlike
// the unit-grained Rezumat query (sql.go): Consum shows the farm's days and
// each compartment's week, and names the compartment × sex rows with no
// ration ("lipsă"), so it returns one row per leaf section, sex and day and
// the rows are added up in Go (reportfeed.go).

// feedDaysSQL is that query, one statement per page (§7.4 Performance: one
// pass over the days' deltas plus one running-count seed at the first day).
// Parameters: $1 farm, $2–$3 the days, $4–$5 the period (does it have any
// ration at all, D45). The "-- name:" line lets a query tracer name it.
//
// Rules applied here, once:
//   - alive = the running sum of live bird events (§5.3); birds at the start =
//     the end of the day before;
//   - kg of a live allocation = the migrated kg, else g × birds at the start
//     of the day / 1000 (§5.7, DATA03);
//   - moved: the section × sex had birds placed, sold, transferred, re-sexed
//     or adjusted that day (anything but deaths and culls);
//   - hall_emptied_before: the last day before $2 on which the hall (a hall
//     with its compartments, or a section without parent and children) went
//     from birds to none, null when it never did. Reportfeed.go carries it
//     over the page's days to tell an emptied hall (§7.4 "goală din …").
//
// Leaves are ordered as the farm's tree: hall, then compartment.
const feedDaysSQL = `-- name: PeriodTableFeedDays
with leaf as (
  select s.id as section_id, s.code, coalesce(p.id, s.id) as hall_id, coalesce(p.code, s.code) as hall_code,
         row_number() over (order by coalesce(p.sort_order, s.sort_order), coalesce(p.code, s.code),
                                     p.id is not null, s.sort_order, s.code) as tree_order
  from section s
  left join section p on p.id = s.parent_id
  where s.farm_id = $1 and not exists (select 1 from section c where c.parent_id = s.id)
), days as (
  select g::date as date from generate_series($2::date, $3::date, interval '1 day') g
), delta as (
  select x.section_id, x.sex, x.date, x.delta,
         (x.placed <> 0 or x.sexing_error <> 0 or x.transfer_in <> 0 or x.transfer_out <> 0
          or x.sold <> 0 or x.adjustment <> 0) as moved
  from v_section_sex_delta x
  where x.farm_id = $1 and x.date between $2::date and $3::date
), seed as (
  select x.section_id, x.sex, sum(x.delta) as alive
  from v_section_sex_delta x
  where x.farm_id = $1 and x.date < $2::date
  group by x.section_id, x.sex
), leaf_day as (
  select l.section_id, sx.sex, d.date, coalesce(x.delta, 0) as delta, coalesce(x.moved, false) as moved,
         coalesce(sd.alive, 0)
           + sum(coalesce(x.delta, 0)) over (partition by l.section_id, sx.sex order by d.date) as alive_end
  from leaf l
  cross join (values ('F'::sex), ('M'::sex)) sx(sex)
  cross join days d
  left join seed sd on sd.section_id = l.section_id and sd.sex = sx.sex
  left join delta x on x.section_id = l.section_id and x.sex = sx.sex and x.date = d.date
), alloc as (
  select fa.section_id, fa.sex, fa.date, fa.g_per_bird, fa.kg_recorded
  from feed_allocation fa
  join entry n on n.id = fa.entry_id and n.deleted_at is null
  where fa.farm_id = $1 and fa.deleted_at is null and fa.date between $2::date and $3::date
), hall_delta as (
  select l.hall_id, x.date, sum(x.delta) as delta
  from leaf l
  join v_section_sex_delta x on x.section_id = l.section_id
  where x.farm_id = $1 and x.date < $2::date
  group by l.hall_id, x.date
), hall_emptied as (
  select r.hall_id, max(r.date) filter (where r.alive_end = 0 and r.alive_end - r.delta > 0) as emptied
  from (select hall_id, date, delta, sum(delta) over (partition by hall_id order by date) as alive_end
        from hall_delta) r
  group by r.hall_id
)
select l.section_id, l.code, l.hall_id, l.hall_code, ld.date, ld.sex::text as sex,
       (ld.alive_end - ld.delta)::integer as birds_start,
       ld.alive_end::integer as alive_end,
       ld.moved,
       (a.section_id is not null) as allocated,
       a.g_per_bird,
       a.kg_recorded,
       coalesce(a.kg_recorded, a.g_per_bird * (ld.alive_end - ld.delta) / 1000.0, 0)::numeric as kg,
       he.emptied as hall_emptied_before,
       exists (select 1 from feed_allocation fa join entry n on n.id = fa.entry_id and n.deleted_at is null
               where fa.farm_id = $1 and fa.deleted_at is null and fa.date between $4::date and $5::date) as period_has_feed
from leaf_day ld
join leaf l on l.section_id = ld.section_id
left join alloc a on a.section_id = ld.section_id and a.sex = ld.sex and a.date = ld.date
left join hall_emptied he on he.hall_id = l.hall_id
order by ld.date, l.tree_order, ld.sex`

// feedRow is one row of feedDaysSQL: a leaf section × sex × day.
type feedRow struct {
	SectionID  uuid.UUID `db:"section_id"`
	Code       string    `db:"code"`
	HallID     uuid.UUID `db:"hall_id"`
	HallCode   string    `db:"hall_code"`
	Date       timex.Day `db:"date"`
	Sex        dbq.Sex   `db:"sex"`
	BirdsStart int       `db:"birds_start"`
	AliveEnd   int       `db:"alive_end"`
	Moved      bool      `db:"moved"`
	// Allocated is true when the section × sex has a live allocation that day;
	// G and KgRecorded are its typed g and migrated kg, Kg the kg it counts.
	Allocated         bool                `db:"allocated"`
	G                 decimal.NullDecimal `db:"g_per_bird"`
	KgRecorded        decimal.NullDecimal `db:"kg_recorded"`
	Kg                decimal.Decimal     `db:"kg"`
	HallEmptiedBefore *timex.Day          `db:"hall_emptied_before"`
	PeriodHasFeed     bool                `db:"period_has_feed"`
}

// fed reports whether the row is a fed day (§5.7): a live allocation above 0
// kg. A register row of 0 kg on a sell-out day is not one.
func (r feedRow) fed() bool { return r.Allocated && r.Kg.IsPositive() }

// feedDaysQuery asks for the Consum rows of a farm.
type feedDaysQuery struct {
	FarmID               uuid.UUID
	From, To             timex.Day // the days to return, both included
	PeriodFrom, PeriodTo timex.Day // the period, for PeriodHasFeed
}

// feedDays is the result of readFeedDays.
type feedDays struct {
	// PeriodHasFeed is true when the period has a live allocation.
	PeriodHasFeed bool
	// Rows are every leaf section × sex × day of [From, To], by day, then in
	// tree order, ♀ before ♂.
	Rows []feedRow
}

// readFeedDays runs feedDaysSQL. A farm without sections gives no rows.
func readFeedDays(ctx context.Context, db dbq.DBTX, q feedDaysQuery) (feedDays, error) {
	if q.From.After(q.To) {
		return feedDays{}, fmt.Errorf("periodtable: feed days %s–%s start after they end", q.From, q.To)
	}
	rows, err := db.Query(ctx, feedDaysSQL, q.FarmID, q.From, q.To, q.PeriodFrom, q.PeriodTo)
	if err != nil {
		return feedDays{}, fmt.Errorf("periodtable: feed days %s–%s: %w", q.From, q.To, err)
	}
	raw, err := pgx.CollectRows(rows, pgx.RowToStructByName[feedRow])
	if err != nil {
		return feedDays{}, fmt.Errorf("periodtable: feed days %s–%s: %w", q.From, q.To, err)
	}
	out := feedDays{Rows: raw}
	if len(raw) > 0 {
		out.PeriodHasFeed = raw[0].PeriodHasFeed
	}
	return out, nil
}
