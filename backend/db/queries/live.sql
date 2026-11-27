-- name: EmitLiveEvent :one
-- The live outbox (§4.11, §9): the live_event row and its pg_notify in one statement, inside the save's transaction.
-- PostgreSQL delivers the notification at commit and drops it on rollback. `at` keeps the database now(): only the
-- purge, on the real clock, reads it (M3 D3).
with ins as (
  insert into live_event (farm_ids, type, payload)
  values (sqlc.arg(farm_ids)::uuid[], sqlc.arg(type), sqlc.arg(payload)::jsonb)
  returning id
)
select ins.id from ins cross join lateral pg_notify('live_event', ins.id::text);
