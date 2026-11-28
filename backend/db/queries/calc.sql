-- name: CalcSeriesAge :one
-- The age parameters of a series (§5.1).
select placement_date, age_at_placement_days::integer as age_at_placement_days
from series
where id = sqlc.arg(series_id)::uuid;

-- name: CalcCycleAt :one
-- The cycle of a farm whose [start_date, end_date] contains the day (an open cycle has no end), with its series'
-- age parameters (§4.6, §5.1).
select c.id, c.series_id, c.start_date, c.end_date, s.placement_date,
       s.age_at_placement_days::integer as age_at_placement_days
from cycle c
join series s on s.id = c.series_id
where c.farm_id = sqlc.arg(farm_id)::uuid
  and c.start_date <= sqlc.arg(day)::date
  and (c.end_date is null or c.end_date >= sqlc.arg(day)::date);

-- name: CalcFarmSexDay :one
-- Farm × sex on one day (§5.3, §5.4): birds alive at the start and end of the day, and the day's deaths, culls
-- and sales. alive_start = Σ deltas before the day, alive_end = Σ deltas up to and including it.
select coalesce(sum(delta) filter (where sex = 'F' and date < sqlc.arg(day)::date), 0)::integer as alive_start_f,
       coalesce(sum(delta) filter (where sex = 'M' and date < sqlc.arg(day)::date), 0)::integer as alive_start_m,
       coalesce(sum(delta) filter (where sex = 'F'), 0)::integer                              as alive_end_f,
       coalesce(sum(delta) filter (where sex = 'M'), 0)::integer                              as alive_end_m,
       coalesce(sum(deaths) filter (where sex = 'F' and date = sqlc.arg(day)::date), 0)::integer as deaths_f,
       coalesce(sum(deaths) filter (where sex = 'M' and date = sqlc.arg(day)::date), 0)::integer as deaths_m,
       coalesce(sum(culls) filter (where sex = 'F' and date = sqlc.arg(day)::date), 0)::integer  as culls_f,
       coalesce(sum(culls) filter (where sex = 'M' and date = sqlc.arg(day)::date), 0)::integer  as culls_m,
       coalesce(sum(sold) filter (where sex = 'F' and date = sqlc.arg(day)::date), 0)::integer   as sold_f,
       coalesce(sum(sold) filter (where sex = 'M' and date = sqlc.arg(day)::date), 0)::integer   as sold_m
from v_section_sex_delta
where farm_id = sqlc.arg(farm_id)::uuid
  and date <= sqlc.arg(day)::date;

-- name: CalcSectionAliveEnd :many
-- Birds alive at the end of the day per bird-holding section (a section without children) × sex (§5.3), every
-- such section of the farm, 0 when it has no birds. Ordered as the farm's tree: hall, then compartment.
select s.id as section_id, s.code,
       coalesce(sum(x.delta) filter (where x.sex = 'F'), 0)::integer as alive_end_f,
       coalesce(sum(x.delta) filter (where x.sex = 'M'), 0)::integer as alive_end_m
from section s
left join section p on p.id = s.parent_id
left join v_section_sex_delta x on x.section_id = s.id and x.date <= sqlc.arg(day)::date
where s.farm_id = sqlc.arg(farm_id)::uuid
  and not exists (select 1 from section c where c.parent_id = s.id)
group by s.id, s.code, s.sort_order, p.code, p.sort_order
order by coalesce(p.sort_order, s.sort_order), coalesce(p.code, s.code), s.sort_order, s.code;

-- name: CalcFarmRemovals :one
-- Farm × sex over [from_day, to_day]: deaths (mortality), culls and sales (§5.4). Sales are never deaths (D155).
select coalesce(sum(deaths) filter (where sex = 'F'), 0)::integer as deaths_f,
       coalesce(sum(deaths) filter (where sex = 'M'), 0)::integer as deaths_m,
       coalesce(sum(culls) filter (where sex = 'F'), 0)::integer  as culls_f,
       coalesce(sum(culls) filter (where sex = 'M'), 0)::integer  as culls_m,
       coalesce(sum(sold) filter (where sex = 'F'), 0)::integer   as sold_f,
       coalesce(sum(sold) filter (where sex = 'M'), 0)::integer   as sold_m
from v_section_sex_delta
where farm_id = sqlc.arg(farm_id)::uuid
  and date between sqlc.arg(from_day)::date and sqlc.arg(to_day)::date;

-- name: CalcFarmBase :one
-- The base of §5.5 at farm scope over [from_day, to_day]: placements plus transfers in whose peer section is in
-- another farm. A transfer row without a peer section (migration rows may have none) takes the source farm of
-- its transfer entry, and counts as inside the farm in any other entry. Adjustments are never part of the base.
-- ReportsCostInputs (reports.sql, lateral b) repeats this sum per cycle in its batch: change both together
-- (TestCostInputsMatchCalc checks they agree).
select coalesce(sum(e.qty) filter (where e.sex = 'F'), 0)::integer as base_f,
       coalesce(sum(e.qty) filter (where e.sex = 'M'), 0)::integer as base_m
from v_bird_event e
join entry n on n.id = e.entry_id
left join section ps on ps.id = e.peer_section_id
where e.farm_id = sqlc.arg(farm_id)::uuid
  and e.date between sqlc.arg(from_day)::date and sqlc.arg(to_day)::date
  and (e.type = 'placement'
       or (e.type = 'transfer' and e.qty > 0
           and coalesce(ps.farm_id, case when n.kind = 'transfer' then n.farm_id end, e.farm_id)
               <> sqlc.arg(farm_id)::uuid));

-- name: CalcLayDay :one
-- Lay of a farm on one day (§5.6): Σ eggs and Σ hens at the start of the day over the egg units whose total is
-- filled (a unit-day without a total is not a hen-day, D159). Hens at the start of the day = Σ ♀ deltas before the
-- day over the unit's sections (§5.3, D156), a running sum over v_section_sex_delta like the other readers, so it
-- holds for any day: v_egg_day reads the dense v_section_sex_day, which stops at greatest(last event, app_today()).
with eggs as (
  select ep.section_id, ep.total_eggs
  from egg_production ep
  join entry n on n.id = ep.entry_id and n.deleted_at is null
  where ep.farm_id = sqlc.arg(farm_id)::uuid
    and ep.date = sqlc.arg(day)::date
    and ep.deleted_at is null
    and ep.total_eggs is not null
)
select coalesce((select sum(total_eggs) from eggs), 0)::integer as eggs,
       coalesce((select sum(x.delta)
                 from eggs e
                 join v_egg_unit u on u.egg_unit_id = e.section_id
                 join v_section_sex_delta x on x.section_id = u.section_id
                 where x.farm_id = sqlc.arg(farm_id)::uuid
                   and x.sex = 'F'
                   and x.date < sqlc.arg(day)::date), 0)::integer as hens_start;

-- name: CalcFeedKg :one
-- Feed consumed by a farm on one day per sex (§5.7): the migrated kg, else g × birds at the start of the day / 1000
-- (DATA03, as v_feed_consumption). Birds at the start of the day = Σ deltas of the section × sex before the day, a
-- running sum over v_section_sex_delta, so it holds for any day (v_feed_consumption reads the dense
-- v_section_sex_day, which stops at greatest(last event, app_today())).
select coalesce(sum(c.kg) filter (where c.sex = 'F'), 0)::numeric as kg_f,
       coalesce(sum(c.kg) filter (where c.sex = 'M'), 0)::numeric as kg_m
from (
  select fa.sex,
         coalesce(fa.kg_recorded, fa.g_per_bird * coalesce(b.birds_start, 0) / 1000.0) as kg
  from feed_allocation fa
  join entry n on n.id = fa.entry_id and n.deleted_at is null
  left join lateral (
    select sum(x.delta)::integer as birds_start
    from v_section_sex_delta x
    where x.farm_id = fa.farm_id
      and x.section_id = fa.section_id
      and x.sex = fa.sex
      and x.date < fa.date
  ) b on true
  where fa.farm_id = sqlc.arg(farm_id)::uuid
    and fa.date = sqlc.arg(day)::date
    and fa.deleted_at is null
) c;

-- name: CalcStockEnd :many
-- Stock at the end of the day per feed type of a farm (§5.8): Σ deliveries − Σ allocation kg + Σ adjustments dated on
-- or before it, with the latest movement day. Feed types with no movement up to that day are absent. An allocation's
-- kg is the migrated kg, else g × birds at the start of its day / 1000 (DATA03), the birds taken once per allocation
-- from one running sum over the farm's v_section_sex_delta (an allocation sorts before the same day's deltas, so the
-- sum at its row is the deltas before its day). It reads no dense view (those stop at the last event or the app's
-- today, so a later allocation would count 0 kg there).
-- Its leaf/alloc CTEs (birds at the start of each allocation's day) are repeated in CalcStockSeries and
-- CalcConsumption, only the filters differ: change the three together. calc.RowFeedKg is the kg in Go.
with leaf as (
  select x.section_id, x.sex, x.date, x.delta::bigint as delta, 1 as ord,
         null::uuid as feed_type_id, null::numeric as g_per_bird, null::numeric as kg_recorded
  from v_section_sex_delta x
  where x.farm_id = sqlc.arg(farm_id)::uuid
    and x.date < sqlc.arg(day)::date
  union all
  select fa.section_id, fa.sex, fa.date, 0, 0, fa.feed_type_id, fa.g_per_bird, fa.kg_recorded
  from feed_allocation fa
  join entry n on n.id = fa.entry_id and n.deleted_at is null
  where fa.farm_id = sqlc.arg(farm_id)::uuid
    and fa.deleted_at is null
    and fa.date <= sqlc.arg(day)::date
), alloc as (
  select l.feed_type_id, l.date, l.ord, l.g_per_bird, l.kg_recorded,
         sum(l.delta) over (partition by l.section_id, l.sex order by l.date, l.ord
                            rows between unbounded preceding and current row) as birds_start
  from leaf l
), mv as (
  select fd.feed_type_id, fd.date, fd.qty_kg::numeric as kg
  from feed_delivery fd
  join entry n on n.id = fd.entry_id and n.deleted_at is null
  where fd.farm_id = sqlc.arg(farm_id)::uuid
    and fd.deleted_at is null
    and fd.date <= sqlc.arg(day)::date
  union all
  select a.feed_type_id, a.date, -coalesce(a.kg_recorded, a.g_per_bird * a.birds_start / 1000.0)
  from alloc a
  where a.ord = 0
  union all
  select sa.feed_type_id, sa.date, sa.qty_kg::numeric
  from feed_stock_adjustment sa
  where sa.farm_id = sqlc.arg(farm_id)::uuid
    and sa.date <= sqlc.arg(day)::date
)
select ft.id as feed_type_id, ft.name, ft.for_sex, max(mv.date)::date as last_move, sum(mv.kg)::numeric as stock_end
from mv
join feed_type ft on ft.id = mv.feed_type_id
group by ft.id, ft.name, ft.name_key, ft.for_sex
order by ft.name_key, ft.id;

-- name: CalcStockSeries :many
-- The stock of the asked feed types of a farm (§5.8, as CalcStockEnd) at the end of from_day, then at the end of every
-- movement day in (from_day, to_day], ordered by type, then day. Every asked type has its from_day row (0 when it
-- has no movement up to then); H5 takes the lowest value of a type's rows.
-- Its leaf/alloc CTEs are CalcStockEnd's (and CalcConsumption's) with the type filter: change the three together.
with leaf as (
  select x.section_id, x.sex, x.date, x.delta::bigint as delta, 1 as ord,
         null::uuid as feed_type_id, null::numeric as g_per_bird, null::numeric as kg_recorded
  from v_section_sex_delta x
  where x.farm_id = sqlc.arg(farm_id)::uuid
    and x.date < sqlc.arg(to_day)::date
  union all
  select fa.section_id, fa.sex, fa.date, 0, 0, fa.feed_type_id, fa.g_per_bird, fa.kg_recorded
  from feed_allocation fa
  join entry n on n.id = fa.entry_id and n.deleted_at is null
  where fa.farm_id = sqlc.arg(farm_id)::uuid
    and fa.deleted_at is null
    and fa.feed_type_id = any(sqlc.arg(feed_type_ids)::uuid[])
    and fa.date <= sqlc.arg(to_day)::date
), alloc as (
  select l.feed_type_id, l.date, l.ord, l.g_per_bird, l.kg_recorded,
         sum(l.delta) over (partition by l.section_id, l.sex order by l.date, l.ord
                            rows between unbounded preceding and current row) as birds_start
  from leaf l
), mv as (
  select fd.feed_type_id, fd.date, fd.qty_kg::numeric as kg
  from feed_delivery fd
  join entry n on n.id = fd.entry_id and n.deleted_at is null
  where fd.farm_id = sqlc.arg(farm_id)::uuid
    and fd.deleted_at is null
    and fd.feed_type_id = any(sqlc.arg(feed_type_ids)::uuid[])
    and fd.date <= sqlc.arg(to_day)::date
  union all
  select a.feed_type_id, a.date, -coalesce(a.kg_recorded, a.g_per_bird * a.birds_start / 1000.0)
  from alloc a
  where a.ord = 0
  union all
  select sa.feed_type_id, sa.date, sa.qty_kg::numeric
  from feed_stock_adjustment sa
  where sa.farm_id = sqlc.arg(farm_id)::uuid
    and sa.feed_type_id = any(sqlc.arg(feed_type_ids)::uuid[])
    and sa.date <= sqlc.arg(to_day)::date
), daily as (
  select mv.feed_type_id, mv.date, sum(mv.kg) as net
  from mv
  group by mv.feed_type_id, mv.date
), run as (
  select d.feed_type_id, d.date,
         sum(d.net) over (partition by d.feed_type_id order by d.date rows between unbounded preceding and current row) as stock_end
  from daily d
)
select t.id::uuid as feed_type_id, sqlc.arg(from_day)::date as date,
       coalesce((select sum(d.net) from daily d where d.feed_type_id = t.id and d.date <= sqlc.arg(from_day)::date),
                0)::numeric as stock_end
from (select distinct u.id from unnest(sqlc.arg(feed_type_ids)::uuid[]) as u(id)) t
union all
select r.feed_type_id, r.date, r.stock_end
from run r
where r.date > sqlc.arg(from_day)::date
order by 1, 2;

-- name: CalcFeedDayComplete :one
-- Whether every section × sex of a farm with birds on the day has a live allocation that day (§5.8 window end).
-- has_birds = alive_start > 0 and alive_end > 0 (§5.11): the sale day that empties a section is not missing. Running
-- sums over v_section_sex_delta, so it holds for any day. A farm without birds that day is complete.
with s as (
  select x.section_id, x.sex,
         coalesce(sum(x.delta) filter (where x.date < sqlc.arg(day)::date), 0) as alive_start,
         sum(x.delta) as alive_end
  from v_section_sex_delta x
  where x.farm_id = sqlc.arg(farm_id)::uuid
    and x.date <= sqlc.arg(day)::date
  group by x.section_id, x.sex
)
select (not exists (
  select 1
  from s
  where s.alive_start > 0
    and s.alive_end > 0
    and not exists (select 1
                    from feed_allocation fa
                    join entry n on n.id = fa.entry_id and n.deleted_at is null
                    where fa.section_id = s.section_id
                      and fa.sex = s.sex
                      and fa.date = sqlc.arg(day)::date
                      and fa.deleted_at is null)
))::boolean as complete;

-- name: CalcConsumption :many
-- Feed consumed by a farm per feed type over [from_day, to_day] (§5.7, §5.8): Σ allocation kg (the migrated kg, else
-- g × birds at the start of the day / 1000), the birds from one running sum over v_section_sex_delta as in
-- CalcStockEnd. Types without an allocation in the range are absent.
-- Its leaf/alloc CTEs are CalcStockEnd's (and CalcStockSeries') with the date range filter: change the three together.
with leaf as (
  select x.section_id, x.sex, x.date, x.delta::bigint as delta, 1 as ord,
         null::uuid as feed_type_id, null::numeric as g_per_bird, null::numeric as kg_recorded
  from v_section_sex_delta x
  where x.farm_id = sqlc.arg(farm_id)::uuid
    and x.date < sqlc.arg(to_day)::date
  union all
  select fa.section_id, fa.sex, fa.date, 0, 0, fa.feed_type_id, fa.g_per_bird, fa.kg_recorded
  from feed_allocation fa
  join entry n on n.id = fa.entry_id and n.deleted_at is null
  where fa.farm_id = sqlc.arg(farm_id)::uuid
    and fa.deleted_at is null
    and fa.date between sqlc.arg(from_day)::date and sqlc.arg(to_day)::date
), alloc as (
  select l.feed_type_id, l.ord, l.g_per_bird, l.kg_recorded,
         sum(l.delta) over (partition by l.section_id, l.sex order by l.date, l.ord
                            rows between unbounded preceding and current row) as birds_start
  from leaf l
)
select a.feed_type_id::uuid as feed_type_id,
       sum(coalesce(a.kg_recorded, a.g_per_bird * a.birds_start / 1000.0))::numeric as kg
from alloc a
where a.ord = 0
group by a.feed_type_id
order by a.feed_type_id;

-- name: CalcSectionRemovals :many
-- Per bird-holding section (a section without children) × sex over [from_day, to_day]: mortality, culls and sales
-- (§5.4), every such section of the farm, 0 when it has none. Ordered as the farm's tree: hall, then compartment.
select s.id as section_id, s.code,
       coalesce(sum(x.deaths) filter (where x.sex = 'F'), 0)::integer as mortality_f,
       coalesce(sum(x.deaths) filter (where x.sex = 'M'), 0)::integer as mortality_m,
       coalesce(sum(x.culls) filter (where x.sex = 'F'), 0)::integer  as culls_f,
       coalesce(sum(x.culls) filter (where x.sex = 'M'), 0)::integer  as culls_m,
       coalesce(sum(x.sold) filter (where x.sex = 'F'), 0)::integer   as sold_f,
       coalesce(sum(x.sold) filter (where x.sex = 'M'), 0)::integer   as sold_m
from section s
left join section p on p.id = s.parent_id
left join v_section_sex_delta x on x.section_id = s.id
     and x.date between sqlc.arg(from_day)::date and sqlc.arg(to_day)::date
where s.farm_id = sqlc.arg(farm_id)::uuid
  and not exists (select 1 from section c where c.parent_id = s.id)
group by s.id, s.code, s.sort_order, p.code, p.sort_order
order by coalesce(p.sort_order, s.sort_order), coalesce(p.code, s.code), s.sort_order, s.code;

-- name: CalcGuideWeeks :many
-- The rows of a guide over [from_week, to_week] (age weeks SAPT, D151), one per week × sex present. A null cell is
-- "– fără date în ghid" (D29); a week × sex outside the guide span has no row.
select gw.week::integer as week, gw.sex, gw.feed_g, gw.body_weight_g, gw.mortality_cum_pct, gw.lay_pct,
       gw.hatching_pct, gw.egg_weight_g
from guide_week gw
where gw.guide_id = sqlc.arg(guide_id)::uuid
  and gw.week between sqlc.arg(from_week)::integer and sqlc.arg(to_week)::integer
order by gw.week, gw.sex;

-- name: CalcSettings :many
-- The asked app settings (D19 thresholds and the other numeric settings) as their JSON text; keys that don't exist
-- are absent.
select a.key, coalesce(a.value #>> '{}', '')::text as value
from app_setting a
where a.key = any(sqlc.arg(keys)::text[])
order by a.key;

-- name: CalcCycleByID :one
-- One cycle of a farm (§4.6) with its series, its guide and whether it has any live bird event (a past cycle without
-- one has no flock count, D52/D54). No row when the cycle belongs to another farm.
select c.id, c.series_id, s.name as series_name, c.farm_id, c.start_date, c.end_date, c.started_by,
       c.start_inferred, c.ended_by, c.guide_id, g.name as guide_name, g.type as guide_type,
       s.placement_date, s.age_at_placement_days::integer as age_at_placement_days,
       exists (select 1
               from bird_event e
               join entry n on n.id = e.entry_id and n.deleted_at is null
               where e.farm_id = c.farm_id
                 and e.deleted_at is null
                 and e.date between c.start_date and coalesce(c.end_date, 'infinity'::date))::boolean as has_bird_events
from cycle c
join series s on s.id = c.series_id
left join guide g on g.id = c.guide_id
where c.id = sqlc.arg(cycle_id)::uuid
  and c.farm_id = sqlc.arg(farm_id)::uuid;

-- name: CalcFarmCycles :many
-- Every cycle of a farm, newest first (CycleSelector, §7.3), with the fields of CalcCycleByID.
select c.id, c.series_id, s.name as series_name, c.farm_id, c.start_date, c.end_date, c.started_by,
       c.start_inferred, c.ended_by, c.guide_id, g.name as guide_name, g.type as guide_type,
       s.placement_date, s.age_at_placement_days::integer as age_at_placement_days,
       exists (select 1
               from bird_event e
               join entry n on n.id = e.entry_id and n.deleted_at is null
               where e.farm_id = c.farm_id
                 and e.deleted_at is null
                 and e.date between c.start_date and coalesce(c.end_date, 'infinity'::date))::boolean as has_bird_events
from cycle c
join series s on s.id = c.series_id
left join guide g on g.id = c.guide_id
where c.farm_id = sqlc.arg(farm_id)::uuid
order by c.start_date desc;

-- name: CalcLatestWeighings :many
-- The latest live weighing per section × sex of a farm in [from_day, to_day] (§5.10, §5.13), with the birds of the
-- section × sex at the end of the weighing day (for a bird-weighted mean, §5.10). Ordered by sex, then newest
-- first, then the farm's tree (hall, then compartment): the first row of a sex is the farm's latest weighing.
select w.section_id, s.code, w.sex, w.date, w.avg_weight_g,
       coalesce((select sum(x.delta)
                 from v_section_sex_delta x
                 where x.farm_id = w.farm_id
                   and x.section_id = w.section_id
                   and x.sex = w.sex
                   and x.date <= w.date), 0)::integer as alive_end
from (
  select distinct on (w.section_id, w.sex) w.farm_id, w.section_id, w.sex, w.date, w.avg_weight_g
  from weighing w
  join entry n on n.id = w.entry_id and n.deleted_at is null
  where w.farm_id = sqlc.arg(farm_id)::uuid
    and w.deleted_at is null
    and w.date between sqlc.arg(from_day)::date and sqlc.arg(to_day)::date
  order by w.section_id, w.sex, w.date desc
) w
join section s on s.id = w.section_id
left join section p on p.id = s.parent_id
order by w.sex, w.date desc, coalesce(p.sort_order, s.sort_order), coalesce(p.code, s.code), s.sort_order, s.code;

-- name: CalcSectionFeedDay :many
-- Per bird-holding section (a section without children) × sex on one day (§5.3, §5.7, §5.11): birds at the start
-- and end of the day, whether a live mortality row was entered, and the day's live allocation: fed or not, the typed
-- g (null for a migrated kg-only row) and the kg (the migrated kg, else g × birds at the start / 1000, DATA03).
-- Every such section of the farm, in tree order (hall, then compartment), with its hall (itself when parentless).
select s.id as section_id, s.code, coalesce(p.id, s.id)::uuid as hall_id, coalesce(p.code, s.code)::text as hall_code,
       coalesce(b.alive_start_f, 0)::integer as alive_start_f,
       coalesce(b.alive_start_m, 0)::integer as alive_start_m,
       coalesce(b.alive_end_f, 0)::integer as alive_end_f,
       coalesce(b.alive_end_m, 0)::integer as alive_end_m,
       coalesce(b.mortality_entered_f, false)::boolean as mortality_entered_f,
       coalesce(b.mortality_entered_m, false)::boolean as mortality_entered_m,
       (af.id is not null)::boolean as fed_f,
       (am.id is not null)::boolean as fed_m,
       af.g_per_bird as g_f,
       am.g_per_bird as g_m,
       coalesce(af.kg_recorded, af.g_per_bird * coalesce(b.alive_start_f, 0) / 1000.0, 0)::numeric as kg_f,
       coalesce(am.kg_recorded, am.g_per_bird * coalesce(b.alive_start_m, 0) / 1000.0, 0)::numeric as kg_m
from section s
left join section p on p.id = s.parent_id
left join lateral (
  select sum(x.delta) filter (where x.sex = 'F' and x.date < sqlc.arg(day)::date) as alive_start_f,
         sum(x.delta) filter (where x.sex = 'M' and x.date < sqlc.arg(day)::date) as alive_start_m,
         sum(x.delta) filter (where x.sex = 'F') as alive_end_f,
         sum(x.delta) filter (where x.sex = 'M') as alive_end_m,
         bool_or(x.mortality_entered) filter (where x.sex = 'F' and x.date = sqlc.arg(day)::date) as mortality_entered_f,
         bool_or(x.mortality_entered) filter (where x.sex = 'M' and x.date = sqlc.arg(day)::date) as mortality_entered_m
  from v_section_sex_delta x
  where x.farm_id = s.farm_id
    and x.section_id = s.id
    and x.date <= sqlc.arg(day)::date
) b on true
left join lateral (
  select fa.id, fa.g_per_bird, fa.kg_recorded
  from feed_allocation fa
  join entry n on n.id = fa.entry_id and n.deleted_at is null
  where fa.section_id = s.id and fa.sex = 'F' and fa.date = sqlc.arg(day)::date and fa.deleted_at is null
) af on true
left join lateral (
  select fa.id, fa.g_per_bird, fa.kg_recorded
  from feed_allocation fa
  join entry n on n.id = fa.entry_id and n.deleted_at is null
  where fa.section_id = s.id and fa.sex = 'M' and fa.date = sqlc.arg(day)::date and fa.deleted_at is null
) am on true
where s.farm_id = sqlc.arg(farm_id)::uuid
  and not exists (select 1 from section c where c.parent_id = s.id)
order by coalesce(p.sort_order, s.sort_order), coalesce(p.code, s.code), s.sort_order, s.code;

-- name: CalcEmptySince :many
-- empty_since per section of a farm at the end of the day (§5.3): every section, a hall counting its compartments
-- and a section without children itself, both sexes together. alive_end is the scope's count at the end of the day;
-- empty_since is the last day its count went from > 0 to 0, null while it has birds or when it never had any.
-- Ordered as the farm's tree: hall, then compartment.
with scope_leaf as (
  select l.id as leaf_id, l.id as scope_id
  from section l
  where l.farm_id = sqlc.arg(farm_id)::uuid
    and not exists (select 1 from section c where c.parent_id = l.id)
  union all
  select l.id, l.parent_id
  from section l
  where l.farm_id = sqlc.arg(farm_id)::uuid
    and l.parent_id is not null
    and not exists (select 1 from section c where c.parent_id = l.id)
), day_delta as (
  select m.scope_id, x.date, sum(x.delta) as delta
  from scope_leaf m
  join v_section_sex_delta x on x.section_id = m.leaf_id
  where x.farm_id = sqlc.arg(farm_id)::uuid
    and x.date <= sqlc.arg(day)::date
  group by m.scope_id, x.date
), running as (
  select scope_id, date, sum(delta) over w as alive_end,
         coalesce(sum(delta) over w - delta, 0) as alive_before
  from day_delta
  window w as (partition by scope_id order by date rows between unbounded preceding and current row)
), scope as (
  select scope_id,
         (array_agg(alive_end order by date desc))[1] as alive_end,
         max(date) filter (where alive_end = 0 and alive_before > 0) as emptied
  from running
  group by scope_id
), emptied as (
  select scope_id, emptied::date as empty_since
  from scope
  where alive_end = 0
)
select s.id as section_id, s.code,
       exists (select 1 from section c where c.parent_id = s.id)::boolean as is_hall,
       coalesce(sc.alive_end, 0)::integer as alive_end,
       e.empty_since
from section s
left join section p on p.id = s.parent_id
left join scope sc on sc.scope_id = s.id
left join emptied e on e.scope_id = s.id
where s.farm_id = sqlc.arg(farm_id)::uuid
order by coalesce(p.sort_order, s.sort_order), coalesce(p.code, s.code), p.id is not null, s.sort_order, s.code;

-- name: CalcDeliveries :many
-- The live deliveries of each asked farm over its day range (§5.9, §7.12): one row per asked range, in the order
-- asked (ord, from 1), 0 and empty arrays when the range has none. A range runs from from_day (only when has_from;
-- else from the farm's first delivery) to to_day, both included. kg per the feed type's for_sex (F, M, null =
-- ambele sexe); priced = total_lei set (feed_delivery_price_ck sets both prices together); the unpriced deliveries
-- one by one in parallel arrays, oldest first, then by type name, then id.
with r as (
  select g.i as ord,
         (sqlc.arg(farm_ids)::uuid[])[g.i] as farm_id,
         (sqlc.arg(has_from)::boolean[])[g.i] as has_from,
         (sqlc.arg(from_days)::date[])[g.i] as from_day,
         (sqlc.arg(to_days)::date[])[g.i] as to_day
  from generate_subscripts(sqlc.arg(farm_ids)::uuid[], 1) as g(i)
), d as (
  select r.ord, fd.id, fd.date, fd.qty_kg, fd.total_lei, ft.id as feed_type_id, ft.name, ft.name_key, ft.for_sex
  from r
  join feed_delivery fd on fd.farm_id = r.farm_id
  join entry n on n.id = fd.entry_id and n.deleted_at is null
  join feed_type ft on ft.id = fd.feed_type_id
  where fd.deleted_at is null
    and fd.date <= r.to_day
    and (not r.has_from or fd.date >= r.from_day)
)
select r.ord::integer as ord, r.farm_id::uuid as farm_id,
       count(d.id)::integer as deliveries,
       coalesce(sum(d.qty_kg) filter (where d.for_sex = 'F'), 0)::numeric as kg_f,
       coalesce(sum(d.qty_kg) filter (where d.for_sex = 'M'), 0)::numeric as kg_m,
       coalesce(sum(d.qty_kg) filter (where d.id is not null and d.for_sex is null), 0)::numeric as kg_both,
       (count(d.id) filter (where d.total_lei is not null))::integer as priced_count,
       coalesce(sum(d.qty_kg) filter (where d.total_lei is not null), 0)::numeric as priced_kg,
       coalesce(sum(d.total_lei), 0)::numeric as priced_lei,
       coalesce(array_agg(d.id order by d.date, d.name_key, d.id)
                  filter (where d.id is not null and d.total_lei is null), '{}')::uuid[] as unpriced_ids,
       coalesce(array_agg(d.date order by d.date, d.name_key, d.id)
                  filter (where d.id is not null and d.total_lei is null), '{}')::date[] as unpriced_dates,
       coalesce(array_agg(d.qty_kg order by d.date, d.name_key, d.id)
                  filter (where d.id is not null and d.total_lei is null), '{}')::numeric[] as unpriced_kgs,
       coalesce(array_agg(d.feed_type_id order by d.date, d.name_key, d.id)
                  filter (where d.id is not null and d.total_lei is null), '{}')::uuid[] as unpriced_type_ids,
       coalesce(array_agg(d.name order by d.date, d.name_key, d.id)
                  filter (where d.id is not null and d.total_lei is null), '{}')::text[] as unpriced_types
from r
left join d on d.ord = r.ord
group by r.ord, r.farm_id
order by r.ord;

-- name: CalcEggsBetween :one
-- Σ eggs of a farm over [from_day, to_day] (§5.6, §5.9): the live egg rows whose total is filled (Cat. 1 + 2 + 3 +
-- Incubabile; Jos apart). No from_day = from the first egg row. ReportsCostInputs (reports.sql, lateral eg) repeats
-- this sum per cycle in its batch: change both together (TestCostInputsMatchCalc checks they agree).
select coalesce(sum(ep.total_eggs), 0)::bigint as eggs
from egg_production ep
join entry n on n.id = ep.entry_id and n.deleted_at is null
where ep.farm_id = sqlc.arg(farm_id)::uuid
  and ep.deleted_at is null
  and ep.total_eggs is not null
  and (sqlc.narg(from_day)::date is null or ep.date >= sqlc.narg(from_day)::date)
  and ep.date <= sqlc.arg(to_day)::date;

-- name: CalcSeriesPlacement :one
-- The series placement per sex (§5.9 series total, §7.3 series page): Σ live placement rows dated inside a cycle of
-- the series at their farm, over every farm when all_farms, else the listed farms only (the caller's: a series page
-- never counts birds placed at a farm the caller can't see, M4 COR-1). The migration's placements sit in a migration
-- entry without a series, so the cycle is the link; an app placement entry (always of the cycle's series) counts only
-- for its own series. ReportsCostInputs (reports.sql, CTE sp) repeats this sum over every farm in its batch: change
-- both together (TestCostInputsMatchCalc checks they agree).
select coalesce(sum(e.qty) filter (where e.sex = 'F'), 0)::integer as placed_f,
       coalesce(sum(e.qty) filter (where e.sex = 'M'), 0)::integer as placed_m
from cycle c
join bird_event e on e.farm_id = c.farm_id
                 and e.date between c.start_date and coalesce(c.end_date, 'infinity'::date)
join entry n on n.id = e.entry_id and n.deleted_at is null
where c.series_id = sqlc.arg(series_id)::uuid
  and (sqlc.arg(all_farms)::boolean or c.farm_id = any(sqlc.arg(farm_ids)::uuid[]))
  and e.type = 'placement'
  and e.deleted_at is null
  and (n.series_id is null or n.series_id = sqlc.arg(series_id)::uuid);

-- name: CalcDeathsWindow :many
-- Per bird-holding section (a section without children) × sex: the deaths (mortality + culls, D165, plan M5 Q3) of
-- the day before the day and of the 7 days before it (day − 7 … day − 1), the inputs of the Zilnic form's Ieri and
-- media 7 zile and of Z6 (§6.2). Every such section of the farm, 0 when it has none, in tree order (hall, then
-- compartment). A running sum over v_section_sex_delta, not the dense view (ARC-6).
select s.id as section_id, s.code,
       coalesce(sum(x.deaths + x.culls) filter (where x.sex = 'F' and x.date = sqlc.arg(day)::date - 1), 0)::integer
         as yesterday_f,
       coalesce(sum(x.deaths + x.culls) filter (where x.sex = 'M' and x.date = sqlc.arg(day)::date - 1), 0)::integer
         as yesterday_m,
       coalesce(sum(x.deaths + x.culls) filter (where x.sex = 'F'), 0)::integer as week_f,
       coalesce(sum(x.deaths + x.culls) filter (where x.sex = 'M'), 0)::integer as week_m
from section s
left join section p on p.id = s.parent_id
left join v_section_sex_delta x on x.section_id = s.id
     and x.farm_id = sqlc.arg(farm_id)::uuid
     and x.date between sqlc.arg(day)::date - 7 and sqlc.arg(day)::date - 1
where s.farm_id = sqlc.arg(farm_id)::uuid
  and not exists (select 1 from section c where c.parent_id = s.id)
group by s.id, s.code, s.sort_order, p.code, p.sort_order
order by coalesce(p.sort_order, s.sort_order), coalesce(p.code, s.code), s.sort_order, s.code;

-- name: CalcHensStart :many
-- Hens at the start of the day per egg unit of the farm (§5.6, D156): Σ ♀ deltas before the day over the unit's
-- bird-holding sections (v_egg_unit: a hall's compartments, or a section without parent and children itself).
-- Every egg unit of the farm (a section without parent), 0 when it has no hens, in tree order. A running sum over
-- v_section_sex_delta like CalcLayDay, so it holds for any day (ARC-6).
with h as (
  select u.egg_unit_id, sum(x.delta) as hens
  from v_egg_unit u
  join v_section_sex_delta x on x.section_id = u.section_id
  where u.farm_id = sqlc.arg(farm_id)::uuid
    and x.farm_id = sqlc.arg(farm_id)::uuid
    and x.sex = 'F'
    and x.date < sqlc.arg(day)::date
  group by u.egg_unit_id
)
select s.id as egg_unit_id, s.code, coalesce(h.hens, 0)::integer as hens_start
from section s
left join h on h.egg_unit_id = s.id
where s.farm_id = sqlc.arg(farm_id)::uuid
  and s.parent_id is null
order by s.sort_order, s.code;
