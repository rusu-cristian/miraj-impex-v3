-- Form states (SPEC §7.8, M3 D8, D14). The save's queries are in entries.sql; these are the reads before editing.

-- name: FormsFeedHead :one
-- The head of the Hrană zilnică form state of a farm on a day: the farm; the cycle whose [start_date, end_date]
-- contains the day (an open cycle has no end) with its series (name, closed or not, age parameters) and the guide's
-- feed_g per sex for SAPT(day) (null without a guide, a guide row or a cell); whether the farm has an open cycle
-- now (the stock rows are shown only then); and the active feed types in name order, as parallel arrays (for_sex
-- '' = ambele sexe). No row for an unknown farm.
select f.id, f.name, f.type,
       c.id as cycle_id, c.start_date as cycle_start, s.name as series_name,
       coalesce(s.status = 'closed', false)::boolean as series_closed, s.placement_date, s.age_at_placement_days,
       c.guide_id, gf.feed_g as guide_f, gm.feed_g as guide_m,
       exists (select 1 from cycle o where o.farm_id = f.id and o.end_date is null)::boolean as open_cycle,
       array(select ft.id from feed_type ft where ft.active order by ft.name_key, ft.id)::uuid[] as type_ids,
       array(select ft.name from feed_type ft where ft.active order by ft.name_key, ft.id)::text[] as type_names,
       array(select coalesce(ft.for_sex::text, '') from feed_type ft where ft.active
             order by ft.name_key, ft.id)::text[] as type_sexes
from farm f
left join cycle c on c.farm_id = f.id
                 and c.start_date <= sqlc.arg(day)::date
                 and (c.end_date is null or c.end_date >= sqlc.arg(day)::date)
left join series s on s.id = c.series_id
left join guide_week gf on gf.guide_id = c.guide_id and gf.sex = 'F' and gf.week = sapt(c.series_id, sqlc.arg(day)::date)
left join guide_week gm on gm.guide_id = c.guide_id and gm.sex = 'M' and gm.week = sapt(c.series_id, sqlc.arg(day)::date)
where f.id = sqlc.arg(farm_id)::uuid;

-- name: FormsFeedRows :many
-- Every bird-holding section (a section without children) of a farm × sex on a day, in tree order (hall, then
-- compartment; ♀ before ♂), retired ones included and flagged, with:
--   - birds_start: Σ deltas before the day (§5.3), a running sum over v_section_sex_delta;
--   - the hall's (itself when parentless) count at the end of the day and since when it is empty (the last day its
--     count went from > 0 to 0, null while it has birds or when it never had any), as CalcEmptySince;
--   - yesterday_g: the typed g of yesterday's live allocation (null when none or kg-only);
--   - last_type_id: the feed type of the latest live allocation before the day;
--   - the day's live allocation, if any: its entry, g, migrated kg and type, with the entry's author, creation
--     instant and peer farm (§3.2 actions). The caller computes its kg (the migrated kg, else g × birds_start / 1000).
with leaf as (
  select s.id, s.code, s.sort_order, s.retired_at, coalesce(p.id, s.id) as hall_id,
         coalesce(p.code, s.code) as hall_code, coalesce(p.sort_order, s.sort_order) as hall_sort
  from section s
  left join section p on p.id = s.parent_id
  where s.farm_id = sqlc.arg(farm_id)::uuid
    and not exists (select 1 from section c where c.parent_id = s.id)
), leaf_sex as (
  select l.id, l.code, l.sort_order, l.retired_at, l.hall_id, l.hall_code, l.hall_sort, x.sex::sex as sex
  from leaf l
  cross join unnest(enum_range(null::sex)) as x(sex)
), alive as (
  select x.section_id, x.sex, sum(x.delta) as alive_start
  from v_section_sex_delta x
  where x.farm_id = sqlc.arg(farm_id)::uuid
    and x.date < sqlc.arg(day)::date
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
), live_alloc as (
  select fa.section_id, fa.sex, fa.date, fa.entry_id, fa.g_per_bird, fa.kg_recorded, fa.feed_type_id,
         n.created_by, n.created_at, n.peer_farm_id
  from feed_allocation fa
  join entry n on n.id = fa.entry_id and n.deleted_at is null
  where fa.farm_id = sqlc.arg(farm_id)::uuid
    and fa.date <= sqlc.arg(day)::date
    and fa.deleted_at is null
), yesterday as (
  select section_id, sex, g_per_bird
  from live_alloc
  where date = sqlc.arg(day)::date - 1
), last_type as (
  select distinct on (section_id, sex) section_id, sex, feed_type_id
  from live_alloc
  where date < sqlc.arg(day)::date
  order by section_id, sex, date desc
), today as (
  select a.section_id, a.sex, a.entry_id, a.g_per_bird, a.kg_recorded, a.feed_type_id, a.created_by, a.created_at,
         a.peer_farm_id, u.name as author
  from live_alloc a
  left join app_user u on u.id = a.created_by
  where a.date = sqlc.arg(day)::date
)
select l.id as section_id, l.code, l.hall_id, l.hall_code, (l.retired_at is not null)::boolean as retired, l.sex,
       coalesce(a.alive_start, 0)::integer as birds_start,
       coalesce(h.alive_end, 0)::integer as hall_alive_end,
       h.empty_since as hall_empty_since,
       y.g_per_bird as yesterday_g,
       lt.feed_type_id as last_type_id,
       e.entry_id as existing_entry_id,
       e.g_per_bird as existing_g,
       e.kg_recorded as existing_kg_recorded,
       e.feed_type_id as existing_type_id,
       e.created_by as existing_created_by,
       e.created_at as existing_created_at,
       e.peer_farm_id as existing_peer_farm_id,
       e.author as existing_author
from leaf_sex l
left join alive a on a.section_id = l.id and a.sex = l.sex
left join hall h on h.hall_id = l.hall_id
left join yesterday y on y.section_id = l.id and y.sex = l.sex
left join last_type lt on lt.section_id = l.id and lt.sex = l.sex
left join today e on e.section_id = l.id and e.sex = l.sex
order by l.hall_sort, l.hall_code, l.sort_order, l.code, l.sex;

-- name: FormsDeliveryHead :one
-- The head of the Livrare hrană form state of a farm (§7.8, M4 Q14): the farm; its open cycle, if any, with its series
-- (name and age parameters, for SAPT today); and the active feed types in name order, as parallel arrays (for_sex
-- '' = ambele sexe). No row for an unknown farm.
select f.id, f.name, f.type,
       c.id as cycle_id, c.start_date as cycle_start, s.name as series_name, s.placement_date, s.age_at_placement_days,
       array(select ft.id from feed_type ft where ft.active order by ft.name_key, ft.id)::uuid[] as type_ids,
       array(select ft.name from feed_type ft where ft.active order by ft.name_key, ft.id)::text[] as type_names,
       array(select coalesce(ft.for_sex::text, '') from feed_type ft where ft.active
             order by ft.name_key, ft.id)::text[] as type_sexes
from farm f
left join cycle c on c.farm_id = f.id and c.end_date is null
left join series s on s.id = c.series_id
where f.id = sqlc.arg(farm_id)::uuid;
