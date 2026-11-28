-- name: SeriesGet :one
-- One series (§7.3 series page, M4 D7) with its age parameters (§5.1). The birds placed in it are
-- calc.SeriesPlacement's (CalcSeriesPlacement), over the caller's farms only (M4 COR-1).
select s.id, s.name, s.status, s.placement_date, s.placement_inferred,
       s.age_at_placement_days::integer as age_at_placement_days
from series s
where s.id = sqlc.arg(series_id)::uuid;

-- name: SeriesFarms :many
-- The farms of a series page (M4 Q5): every farm the caller sees (all of them when all_farms, else the listed ids),
-- and every farm with a cycle of the series, with whether the caller sees it. Growth farms first, then laying, each
-- by name_key.
select f.id, f.name, f.type,
       (sqlc.arg(all_farms)::boolean or f.id = any(sqlc.arg(farm_ids)::uuid[]))::boolean as visible
from farm f
where sqlc.arg(all_farms)::boolean
   or f.id = any(sqlc.arg(farm_ids)::uuid[])
   or exists (select 1 from cycle c where c.farm_id = f.id and c.series_id = sqlc.arg(series_id)::uuid)
order by f.type, f.name_key;

-- name: SeriesCycles :many
-- Every cycle of a series on any farm, oldest first, with its guide snapshot (B4).
select c.id, c.farm_id, c.start_date, c.end_date, c.start_inferred, c.ended_by,
       c.guide_id, g.name as guide_name, g.type as guide_type
from cycle c
left join guide g on g.id = c.guide_id
where c.series_id = sqlc.arg(series_id)::uuid
order by c.start_date, c.id;

-- name: SeriesTransfers :many
-- The live transfers between two farms the caller sees (all of them when all_farms) dated inside a cycle of the
-- series at either farm (§7.3 cycle track links, M4 Q5), one row per day and direction, with the birds that arrived
-- per sex (the destination side of each pair, as FarmsTransfers). Moves inside one farm are not links.
select n.date, n.farm_id as from_farm_id, ff.name as from_farm, tf.name as to_farm,
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
  and (sqlc.arg(all_farms)::boolean or n.farm_id = any(sqlc.arg(farm_ids)::uuid[]))
  and (sqlc.arg(all_farms)::boolean or n.peer_farm_id = any(sqlc.arg(farm_ids)::uuid[]))
  and exists (select 1
              from cycle c
              where c.series_id = sqlc.arg(series_id)::uuid
                and c.farm_id in (n.farm_id, n.peer_farm_id)
                and n.date between c.start_date and coalesce(c.end_date, 'infinity'::date))
group by n.date, n.farm_id, ff.name, n.peer_farm_id, tf.name
order by n.date, ff.name, tf.name;
