-- name: ReportsFeedCycles :many
-- The farms of Rapoarte › Hrană (§7.12, M3 D9, M4 D6): every farm when all_farms, else the listed ids; ordered by
-- name_key, then each farm's cycles newest first (as CalcFarmCycles orders them). A farm without any cycle is one row
-- with a null cycle. Per cycle: its series (name, status), age parameters and guide type (the cycle's type, else the
-- farm's), and series_farms = how many farms have a cycle of that series (every farm, also those the caller can't
-- see: the Cost total needs them all, M4 Q6).
select f.id as farm_id, f.name as farm_name, f.type as farm_type,
       c.id as cycle_id, c.series_id, s.name as series_name, s.status as series_status,
       c.start_date, c.end_date, s.placement_date, coalesce(s.age_at_placement_days, 0)::integer as age_at_placement_days,
       g.type as guide_type,
       (select count(distinct c2.farm_id) from cycle c2 where c2.series_id = c.series_id)::integer as series_farms
from farm f
left join cycle c on c.farm_id = f.id
left join series s on s.id = c.series_id
left join guide g on g.id = c.guide_id
where sqlc.arg(all_farms)::boolean or f.id = any(sqlc.arg(farm_ids)::uuid[])
order by f.name_key, c.start_date desc;

-- name: ReportsDeliveryRows :many
-- One page of the Livrări rows (§7.12, M4 D6, Q7): the live deliveries of each asked farm over its day range (from_day
-- only when has_from, to_day; both included), ordered date desc, farm name_key, type name_key, id. A keyset cursor
-- (the last row of the previous page: all four cursor_* set) starts after that row. row_limit rows at most. With the
-- price columns and the pricing user's name (the service drops them for a viewer, M4 D2), and series_closed: the
-- delivery is dated inside a cycle of a closed series at its farm (closed_series_on, 0003; no price is set or changed
-- there, §3.1).
with r as (
  select (sqlc.arg(farm_ids)::uuid[])[g.i] as farm_id,
         (sqlc.arg(has_from)::boolean[])[g.i] as has_from,
         (sqlc.arg(from_days)::date[])[g.i] as from_day,
         (sqlc.arg(to_days)::date[])[g.i] as to_day
  from generate_subscripts(sqlc.arg(farm_ids)::uuid[], 1) as g(i)
)
select fd.id, fd.date, fd.farm_id, f.name as farm_name, f.name_key as farm_key, ft.name as type_name,
       ft.name_key as type_key, ft.for_sex, fd.qty_kg, fd.price_per_kg, fd.total_lei, fd.price_from_register,
       fd.priced_at, u.name as priced_by_name,
       (closed_series_on(fd.farm_id, fd.date) is not null)::boolean as series_closed
from r
join feed_delivery fd on fd.farm_id = r.farm_id
join entry n on n.id = fd.entry_id and n.deleted_at is null
join farm f on f.id = fd.farm_id
join feed_type ft on ft.id = fd.feed_type_id
left join app_user u on u.id = fd.priced_by
where fd.deleted_at is null
  and fd.date <= r.to_day
  and (not r.has_from or fd.date >= r.from_day)
  and (sqlc.narg(cursor_date)::date is null
       or fd.date < sqlc.narg(cursor_date)::date
       or (fd.date = sqlc.narg(cursor_date)::date
           and (f.name_key, ft.name_key, fd.id)
               > (sqlc.narg(cursor_farm_key)::text, sqlc.narg(cursor_type_key)::text, sqlc.narg(cursor_id)::uuid)))
order by fd.date desc, f.name_key, ft.name_key, fd.id
limit sqlc.arg(row_limit)::integer;

-- name: ReportsLastDeliveryBefore :many
-- The newest live delivery of the asked farms dated before its farm's range (the empty Livrări's "Ultima livrare",
-- M4 Q8): only ranges with a from_day count; at most one row, in the Livrări row order.
with r as (
  select (sqlc.arg(farm_ids)::uuid[])[g.i] as farm_id,
         (sqlc.arg(from_days)::date[])[g.i] as from_day
  from generate_subscripts(sqlc.arg(farm_ids)::uuid[], 1) as g(i)
)
select fd.date, f.name as farm_name, ft.name as type_name, fd.qty_kg
from r
join feed_delivery fd on fd.farm_id = r.farm_id
join entry n on n.id = fd.entry_id and n.deleted_at is null
join farm f on f.id = fd.farm_id
join feed_type ft on ft.id = fd.feed_type_id
where fd.deleted_at is null
  and fd.date < r.from_day
order by fd.date desc, f.name_key, ft.name_key, fd.id
limit 1;

-- name: ReportsCostInputs :many
-- The bases of the Cost rows (§5.9, M4 D6, Q6) in one statement, one row per asked cycle in the order asked (ord, from
-- 1): base_f/base_m = calc.FarmBase over [cycle_start, base_to] (placements and transfers in from another farm),
-- eggs = calc.EggsBetween over the cycle's delivery range (from egg_from only when has_egg_from, to egg_to), and
-- placed_f/placed_m = calc.SeriesPlacement of series_id (the same on every row; 0 for a null series_id). The three
-- calc readers are the reference: TestCostInputsMatchCalc checks this batch against them.
with r as (
  select g.i as ord,
         (sqlc.arg(farm_ids)::uuid[])[g.i] as farm_id,
         (sqlc.arg(cycle_starts)::date[])[g.i] as cycle_start,
         (sqlc.arg(base_tos)::date[])[g.i] as base_to,
         (sqlc.arg(has_egg_from)::boolean[])[g.i] as has_egg_from,
         (sqlc.arg(egg_froms)::date[])[g.i] as egg_from,
         (sqlc.arg(egg_tos)::date[])[g.i] as egg_to
  from generate_subscripts(sqlc.arg(farm_ids)::uuid[], 1) as g(i)
), sp as (
  select coalesce(sum(e.qty) filter (where e.sex = 'F'), 0)::integer as placed_f,
         coalesce(sum(e.qty) filter (where e.sex = 'M'), 0)::integer as placed_m
  from cycle c
  join bird_event e on e.farm_id = c.farm_id
                   and e.date between c.start_date and coalesce(c.end_date, 'infinity'::date)
  join entry n on n.id = e.entry_id and n.deleted_at is null
  where c.series_id = sqlc.narg(series_id)::uuid
    and e.type = 'placement'
    and e.deleted_at is null
    and (n.series_id is null or n.series_id = sqlc.narg(series_id)::uuid)
)
select r.ord::integer as ord, r.farm_id::uuid as farm_id, b.base_f, b.base_m, eg.eggs, sp.placed_f, sp.placed_m
from r
cross join sp
cross join lateral (
  select coalesce(sum(e.qty) filter (where e.sex = 'F'), 0)::integer as base_f,
         coalesce(sum(e.qty) filter (where e.sex = 'M'), 0)::integer as base_m
  from v_bird_event e
  join entry n on n.id = e.entry_id
  left join section ps on ps.id = e.peer_section_id
  where e.farm_id = r.farm_id
    and e.date between r.cycle_start and r.base_to
    and (e.type = 'placement'
         or (e.type = 'transfer' and e.qty > 0
             and coalesce(ps.farm_id, case when n.kind = 'transfer' then n.farm_id end, e.farm_id) <> r.farm_id))
) b
cross join lateral (
  select coalesce(sum(ep.total_eggs), 0)::bigint as eggs
  from egg_production ep
  join entry n on n.id = ep.entry_id and n.deleted_at is null
  where ep.farm_id = r.farm_id
    and ep.deleted_at is null
    and ep.total_eggs is not null
    and (not r.has_egg_from or ep.date >= r.egg_from)
    and ep.date <= r.egg_to
) eg
order by r.ord;
