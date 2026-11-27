-- name: LockFarm :exec
-- §6.1 step 0, §12.1: a transaction-level advisory lock on one farm. ledger.LockFarms takes them in ascending id
-- order. Two farm ids whose 64-bit hashes collide only serialise more than needed.
select pg_advisory_xact_lock(hashtextextended(sqlc.arg(farm_id)::uuid::text, 0));
