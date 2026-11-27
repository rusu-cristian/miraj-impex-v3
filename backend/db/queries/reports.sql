-- name: ReportsFeedFarms :many
-- The farms of the Stoc report (§7.12, M3 D9): every farm when all_farms, else the listed ids; ordered by name_key.
-- Each with its latest cycle (by start date, as CalcFarmCycles orders them): has_open_cycle when it has no end
-- date; otherwise the farm shows one "fără ciclu deschis" row, with the latest cycle's series and end date (null
-- without any cycle).
select f.id, f.name,
       (c.id is not null and c.end_date is null)::boolean as has_open_cycle,
       c.end_date, s.name as series_name
from farm f
left join lateral (
  select c.id, c.series_id, c.end_date
  from cycle c
  where c.farm_id = f.id
  order by c.start_date desc
  limit 1
) c on true
left join series s on s.id = c.series_id
where sqlc.arg(all_farms)::boolean or f.id = any(sqlc.arg(farm_ids)::uuid[])
order by f.name_key;
