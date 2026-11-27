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
-- Stock at the end of the day per feed type of a farm (§5.8): the running stock of the latest movement day on or
-- before it. Feed types with no movement up to that day are absent.
select t.feed_type_id, t.name, t.date, t.stock_end
from (
  select distinct on (v.feed_type_id) v.feed_type_id, ft.name, ft.name_key, v.date, v.stock_end
  from v_feed_stock v
  join feed_type ft on ft.id = v.feed_type_id
  where v.farm_id = sqlc.arg(farm_id)::uuid
    and v.date <= sqlc.arg(day)::date
  order by v.feed_type_id, v.date desc
) t
order by t.name_key, t.feed_type_id;

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
