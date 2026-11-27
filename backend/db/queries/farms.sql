-- name: FarmsGetFarm :one
-- One farm with its own guide (the guide its next cycle copies, B1).
select f.id, f.name, f.type, f.guide_id, g.name as guide_name
from farm f
left join guide g on g.id = f.guide_id
where f.id = sqlc.arg(farm_id)::uuid;

-- name: FarmsListFarms :many
-- The farms a caller sees (§7.2): every farm when all_farms, else the listed ids; ordered by name_key like
-- /api/auth/me. Each with its own guide and its open cycle (null columns when it has none) with the series' age
-- parameters (§5.1).
select f.id, f.name, f.type, f.guide_id, g.name as guide_name,
       c.id as cycle_id, c.series_id, s.name as series_name, c.start_date,
       s.placement_date, s.age_at_placement_days
from farm f
left join guide g on g.id = f.guide_id
left join cycle c on c.farm_id = f.id and c.end_date is null
left join series s on s.id = c.series_id
where sqlc.arg(all_farms)::boolean or f.id = any(sqlc.arg(farm_ids)::uuid[])
order by f.name_key;

-- name: FarmsSections :many
-- The active (not retired) sections of the farms a caller sees (§7.2), siblings by sort_order then code. is_hall:
-- the section has children (retired ones included: a hall never holds birds, §4.4).
select s.id, s.farm_id, s.code, s.parent_id,
       exists (select 1 from section c where c.parent_id = s.id)::boolean as is_hall
from section s
where s.retired_at is null
  and (sqlc.arg(all_farms)::boolean or s.farm_id = any(sqlc.arg(farm_ids)::uuid[]))
order by s.sort_order, s.code;

-- name: FarmsSeries :many
-- The series with a cycle on the farms a caller sees (§7.2, M2 D7), newest first, with the first day and the last
-- day of those cycles (an open cycle counts up to today) and the age parameters (§5.1) for saptSpan.
select s.id, s.name, s.status, s.placement_date, s.age_at_placement_days::integer as age_at_placement_days,
       min(c.start_date)::date as first_day,
       max(coalesce(c.end_date, sqlc.arg(today)::date))::date as last_day
from series s
join cycle c on c.series_id = s.id
where sqlc.arg(all_farms)::boolean or c.farm_id = any(sqlc.arg(farm_ids)::uuid[])
group by s.id, s.name, s.status, s.placement_date, s.age_at_placement_days, s.name_key
order by s.placement_date desc, s.name_key;

-- name: FarmsActiveFeedTypes :many
-- The active feed types (§7.2), ordered by name_key. for_sex null = both sexes.
select id, name, for_sex, active
from feed_type
where active
order by name_key;

-- name: FarmsTransfers :many
-- The live transfers between a farm and another farm (§7.3 farm track links), one row per day and direction, with
-- the birds that arrived per sex (the destination side of each pair). Moves inside one farm are not links.
select n.date, ff.name as from_farm, tf.name as to_farm,
       coalesce(sum(e.qty) filter (where e.sex = 'F'), 0)::integer as f,
       coalesce(sum(e.qty) filter (where e.sex = 'M'), 0)::integer as m
from entry n
join farm ff on ff.id = n.farm_id
join farm tf on tf.id = n.peer_farm_id
join bird_event e on e.entry_id = n.id
                 and e.farm_id = n.peer_farm_id
                 and e.type = 'transfer'
                 and e.deleted_at is null
where n.kind = 'transfer'
  and n.deleted_at is null
  and n.peer_farm_id <> n.farm_id
  and (n.farm_id = sqlc.arg(farm_id)::uuid or n.peer_farm_id = sqlc.arg(farm_id)::uuid)
group by n.date, n.farm_id, ff.name, n.peer_farm_id, tf.name
order by n.date, ff.name, tf.name;
