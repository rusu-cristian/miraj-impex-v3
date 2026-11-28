-- name: LockFarm :exec
-- §6.1 step 0, §12.1: a transaction-level advisory lock on one farm. ledger.LockFarms takes them in ascending id
-- order. Two farm ids whose 64-bit hashes collide only serialise more than needed.
select pg_advisory_xact_lock(hashtextextended(sqlc.arg(farm_id)::uuid::text, 0));

-- name: NegativeCounts :many
-- §6.1 steps 2–3 (plan M5-D4): the asked (section, sex) pairs whose birds alive at the end of a day are below 0 on a
-- day in [from_day, greatest(last event, today)], one row per such pair in the order asked (the arrays are parallel:
-- section_ids[i] goes with sexes[i]). A running sum over v_section_sex_delta, which has rows on event days only:
-- the count of an event day holds until the day before the next event, and the last event's until
-- greatest(its day, today), as the dense per-day view shows it (TestCheckCountsMatchesView). today is the caller's
-- (ARC-6), never the database's business day. first_* is the first negative day on or after from_day and its count,
-- last_negative_date the last negative day (today while still negative), min_alive_end the lowest count,
-- negative_today whether today is negative and today_alive_end today's count then (0 otherwise). Live rows of live
-- entries only (v_bird_event).
with a as (
  select i as ord, (sqlc.arg(section_ids)::uuid[])[i] as section_id, ((sqlc.arg(sexes)::text[])[i])::sex as sex
  from generate_subscripts(sqlc.arg(section_ids)::uuid[], 1) as i
), d as (
  select x.section_id, x.sex, x.date,
         sum(x.delta) over w as alive_end,
         lead(x.date) over w as next_date
  from v_section_sex_delta x
  where x.section_id = any(sqlc.arg(section_ids)::uuid[])
    and (x.section_id, x.sex) in (select a.section_id, a.sex from a)
  window w as (partition by x.section_id, x.sex order by x.date rows between unbounded preceding and current row)
), neg as (
  select section_id, sex, alive_end,
         greatest(date, sqlc.arg(from_day)::date) as first_day,
         coalesce(next_date - 1, greatest(date, sqlc.arg(today)::date)) as last_day
  from d
  where alive_end < 0
    and coalesce(next_date - 1, greatest(date, sqlc.arg(today)::date)) >= sqlc.arg(from_day)::date
)
select s.id as section_id, s.code, n.sex,
       min(n.first_day)::date                                                                  as first_negative_date,
       ((array_agg(n.alive_end order by n.first_day))[1])::integer                             as first_alive_end,
       max(n.last_day)::date                                                                   as last_negative_date,
       min(n.alive_end)::integer                                                               as min_alive_end,
       bool_or(sqlc.arg(today)::date between n.first_day and n.last_day)::boolean               as negative_today,
       coalesce(min(n.alive_end) filter (where sqlc.arg(today)::date between n.first_day and n.last_day),
                0)::integer                                                                     as today_alive_end
from neg n
join a on a.section_id = n.section_id and a.sex = n.sex
join section s on s.id = n.section_id
group by a.ord, s.id, s.code, n.sex
order by a.ord;
