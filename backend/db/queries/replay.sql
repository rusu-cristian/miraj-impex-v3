-- name: InsertReplay :exec
-- The exact response bytes of a create that carried a clientRequestId, written in the save's transaction (§4.11,
-- §12.5). created_at keeps the database now(): only the 7-day purge reads it (M3 D3).
insert into request_replay (client_request_id, user_id, route, status, body)
values (sqlc.arg(client_request_id), sqlc.arg(user_id), sqlc.arg(route), sqlc.arg(status), sqlc.arg(body));

-- name: GetReplay :one
select user_id, route, status, body from request_replay where client_request_id = sqlc.arg(client_request_id);

-- name: EntryIDByClientRequest :one
-- The entry that holds an idempotency key (soft-deleted ones too: the key stays taken), with its author and kind: a
-- key whose reply was purged answers ALREADY_ENTERED only to its author on a route of its kind (M3 SEC-1).
select id, created_by, kind from entry where client_request_id = sqlc.arg(client_request_id)::uuid;
