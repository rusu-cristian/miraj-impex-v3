-- The Zilnic form state (SPEC §6.2, §7.8 GET /api/forms/daily, plan M5 T4, Q2). Owned by entries/dailyform.go; the
-- deaths window and the hens at the start of the day are calc's readers (CalcDeathsWindow, CalcHensStart).

-- name: DailyFormHead :one
-- The head of the Zilnic form state of a farm on a day: the farm; the cycle whose [start_date, end_date] contains the
-- day (an open cycle has no end) with its series (name, age parameters) and the guide's lay % ♀ for SAPT(day) (null
-- without a guide, a guide row or a cell, §5.6 guide_lay); the closed series whose cycle holds the day
-- (closed_series_on, 0003; empty when none, §3.1); and the farm's latest cycle that ended before the day, with its series
-- (the empty-farm state, M2 Q4 b). No row for an unknown farm.
select f.id, f.name, f.type,
       c.id as cycle_id, c.start_date as cycle_start, s.name as series_name, s.placement_date, s.age_at_placement_days,
       gl.lay_pct as guide_lay_pct,
       coalesce(closed_series_on(f.id, sqlc.arg(day)::date), '')::text as closed_series,
       ls.name as last_series, lc.start_date as last_start, lc.end_date as last_end,
       ls.placement_date as last_placement_date, ls.age_at_placement_days as last_age_at_placement_days
from farm f
left join cycle c on c.farm_id = f.id
                 and c.start_date <= sqlc.arg(day)::date
                 and (c.end_date is null or c.end_date >= sqlc.arg(day)::date)
left join series s on s.id = c.series_id
left join guide_week gl on gl.guide_id = c.guide_id and gl.sex = 'F' and gl.week = sapt(c.series_id, sqlc.arg(day)::date)
left join cycle lc on lc.id = (
  select pc.id
  from cycle pc
  where pc.farm_id = f.id
    and pc.end_date < sqlc.arg(day)::date
  order by pc.end_date desc, pc.start_date desc
  limit 1
)
left join series ls on ls.id = lc.series_id
where f.id = sqlc.arg(farm_id)::uuid;

-- name: DailyFormSections :many
-- Every bird-holding section (a section without children) of a farm × sex on a day, in tree order (hall, then
-- compartment; ♀ before ♂), retired ones included and flagged, with:
--   - parent_id (null for a parentless section, its own egg unit) and the hall (itself when parentless), retired or not;
--   - alive_start: Σ deltas before the day (§5.3), a running sum over v_section_sex_delta (ARC-6);
--   - inflow: a placement or a transfer in on the day (the sexes present, §5.3; Z3, plan M5 Q3);
--   - the hall's count at the end of the day and since when it is empty (as FormsFeedRows, the summary's tiles);
--   - the day's live Zilnic values (bird_event mortality, cull, sexing_error of live entries): has_* = a row exists,
--     mortality and cull as positive counts (−qty), the sexing error signed as recorded, the cull reason (empty when none).
with leaf as (
  select s.id, s.code, s.sort_order, s.retired_at, s.parent_id, coalesce(p.id, s.id) as hall_id,
         coalesce(p.code, s.code) as hall_code, coalesce(p.sort_order, s.sort_order) as hall_sort,
         coalesce(p.retired_at, s.retired_at) as hall_retired_at
  from section s
  left join section p on p.id = s.parent_id
  where s.farm_id = sqlc.arg(farm_id)::uuid
    and not exists (select 1 from section c where c.parent_id = s.id)
), leaf_sex as (
  select l.id, l.code, l.sort_order, l.retired_at, l.parent_id, l.hall_id, l.hall_code, l.hall_sort, l.hall_retired_at,
         x.sex::sex as sex
  from leaf l
  cross join unnest(enum_range(null::sex)) as x(sex)
), alive as (
  select x.section_id, x.sex,
         coalesce(sum(x.delta) filter (where x.date < sqlc.arg(day)::date), 0) as alive_start,
         bool_or(x.date = sqlc.arg(day)::date and (x.placed > 0 or x.transfer_in > 0)) as inflow
  from v_section_sex_delta x
  where x.farm_id = sqlc.arg(farm_id)::uuid
    and x.date <= sqlc.arg(day)::date
  group by x.section_id, x.sex
), hall_day as (
  select l.hall_id, x.date, sum(x.delta) as delta
  from leaf l
  join v_section_sex_delta x on x.section_id = l.id
  where x.farm_id = sqlc.arg(farm_id)::uuid
    and x.date <= sqlc.arg(day)::date
  group by l.hall_id, x.date
), hall_running as (
  select hall_id, date, sum(delta) over w as alive_end,
         coalesce(sum(delta) over w - delta, 0) as alive_before
  from hall_day
  window w as (partition by hall_id order by date rows between unbounded preceding and current row)
), hall_scope as (
  select hall_id,
         (array_agg(alive_end order by date desc))[1] as alive_end,
         max(date) filter (where alive_end = 0 and alive_before > 0) as emptied
  from hall_running
  group by hall_id
), hall as (
  select hall_id, alive_end, (case when alive_end = 0 then emptied end)::date as empty_since
  from hall_scope
), saved as (
  select b.section_id, b.sex,
         bool_or(b.type = 'mortality') as has_mortality,
         coalesce(-sum(b.qty) filter (where b.type = 'mortality'), 0) as mortality,
         bool_or(b.type = 'cull') as has_cull,
         coalesce(-sum(b.qty) filter (where b.type = 'cull'), 0) as cull,
         coalesce(max(b.cull_reason::text) filter (where b.type = 'cull'), '') as cull_reason,
         bool_or(b.type = 'sexing_error') as has_sexing_error,
         coalesce(sum(b.qty) filter (where b.type = 'sexing_error'), 0) as sexing_error
  from bird_event b
  join entry n on n.id = b.entry_id and n.deleted_at is null
  where b.farm_id = sqlc.arg(farm_id)::uuid
    and b.date = sqlc.arg(day)::date
    and b.deleted_at is null
    and b.type in ('mortality', 'cull', 'sexing_error')
  group by b.section_id, b.sex
)
select l.id as section_id, l.code, l.parent_id, l.hall_id, l.hall_code, (l.retired_at is not null)::boolean as retired,
       (l.hall_retired_at is not null)::boolean as hall_retired, l.sex,
       coalesce(a.alive_start, 0)::integer as alive_start,
       coalesce(a.inflow, false)::boolean as inflow,
       coalesce(h.alive_end, 0)::integer as hall_alive_end,
       h.empty_since as hall_empty_since,
       coalesce(v.has_mortality, false)::boolean as has_mortality,
       coalesce(v.mortality, 0)::integer as mortality,
       coalesce(v.has_cull, false)::boolean as has_cull,
       coalesce(v.cull, 0)::integer as cull,
       coalesce(v.cull_reason, '')::text as cull_reason,
       coalesce(v.has_sexing_error, false)::boolean as has_sexing_error,
       coalesce(v.sexing_error, 0)::integer as sexing_error
from leaf_sex l
left join alive a on a.section_id = l.id and a.sex = l.sex
left join hall h on h.hall_id = l.hall_id
left join saved v on v.section_id = l.id and v.sex = l.sex
order by l.hall_sort, l.hall_code, l.sort_order, l.code, l.sex;

-- name: DailyFormEggs :many
-- The day's live egg_production rows of a farm (one per egg unit, egg_production_uq), of live entries. total_eggs is
-- generated and not read: the client sums the categories as the save does (§5.6).
select g.section_id, g.floor_eggs, g.cat1, g.cat2, g.cat3, g.hatching_eggs, g.avg_egg_weight_g
from egg_production g
join entry n on n.id = g.entry_id and n.deleted_at is null
where g.farm_id = sqlc.arg(farm_id)::uuid
  and g.date = sqlc.arg(day)::date
  and g.deleted_at is null;

-- name: DailyFormEntries :many
-- The live entries holding the day's Zilnic values at a farm (bird_event mortality, cull, sexing_error; egg_production),
-- one row per entry × section holding a value (egg = the section is an egg unit's eggs), oldest entry first (by
-- creation instant, then id), then in tree order: the form's `existing` list with its author, creation instant and
-- kind (§3.2 actions; only served kinds may be changed, Q1).
with r as (
  select b.entry_id, b.section_id, false as egg
  from bird_event b
  where b.farm_id = sqlc.arg(farm_id)::uuid
    and b.date = sqlc.arg(day)::date
    and b.deleted_at is null
    and b.type in ('mortality', 'cull', 'sexing_error')
  union
  select g.entry_id, g.section_id, true as egg
  from egg_production g
  where g.farm_id = sqlc.arg(farm_id)::uuid
    and g.date = sqlc.arg(day)::date
    and g.deleted_at is null
)
select n.id as entry_id, n.kind, n.created_by, n.created_at, n.peer_farm_id, u.name as author, s.code,
       r.egg::boolean as egg
from r
join entry n on n.id = r.entry_id and n.deleted_at is null
join section s on s.id = r.section_id
left join section p on p.id = s.parent_id
left join app_user u on u.id = n.created_by
order by n.created_at, n.id, r.egg, coalesce(p.sort_order, s.sort_order), coalesce(p.code, s.code), s.sort_order, s.code;
