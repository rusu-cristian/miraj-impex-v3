-- name: InsertAudit :exec
-- Audit log (SPEC §4.10), append-only: a setup or user row with no entry, farm scope or reason. `at` comes from
-- the injected clock (plan D5); user_id null = system (create-owner, Q6).
insert into audit_log (at, user_id, action, entity, entity_id, new_values)
values (sqlc.arg(at), sqlc.narg(user_id), sqlc.arg(action), sqlc.arg(entity), sqlc.arg(entity_id),
        sqlc.narg(new_values));
