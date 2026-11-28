-- name: InsertAudit :exec
-- Audit log (SPEC §4.10), append-only: a setup or user row with no entry, farm scope or reason. `at` comes from
-- the injected clock (plan D5); user_id null = system (create-owner, Q6).
insert into audit_log (at, user_id, action, entity, entity_id, new_values)
values (sqlc.arg(at), sqlc.narg(user_id), sqlc.arg(action), sqlc.arg(entity), sqlc.arg(entity_id),
        sqlc.narg(new_values));

-- name: InsertAuditLog :exec
-- One audit row per save action, every §4.10 column (M3 D1; the name InsertEntryAudit is migratehistory.sql's).
-- `at` comes from the injected clock (M3 D3); changes is never null ('[]' for a create), old_values/new_values may be.
-- InsertAuditLogID below must keep exactly these parameters, in this order: audit.WriteID converts
-- InsertAuditLogParams to InsertAuditLogIDParams (a Go struct conversion that compiles only while the two match).
insert into audit_log (at, user_id, action, entity, entity_id, entry_id, entry_kind, record_types, farm_ids,
                       record_date, reason, same_day_author, summary, changes, old_values, new_values, source,
                       voice_transcript)
values (sqlc.arg(at), sqlc.narg(user_id), sqlc.arg(action), sqlc.arg(entity), sqlc.arg(entity_id),
        sqlc.narg(entry_id), sqlc.narg(entry_kind), sqlc.arg(record_types)::text[], sqlc.arg(farm_ids)::uuid[],
        sqlc.narg(record_date), sqlc.narg(reason), sqlc.arg(same_day_author), sqlc.narg(summary),
        sqlc.arg(changes)::jsonb, sqlc.narg(old_values)::jsonb, sqlc.narg(new_values)::jsonb, sqlc.arg(source),
        sqlc.narg(voice_transcript));

-- name: InsertAuditLogID :one
-- InsertAuditLog returning the row's bigserial id (M4 D3): the price batch's undo token names its audit row. Keep its
-- parameters identical to InsertAuditLog's (audit.WriteID converts one params struct into the other).
insert into audit_log (at, user_id, action, entity, entity_id, entry_id, entry_kind, record_types, farm_ids,
                       record_date, reason, same_day_author, summary, changes, old_values, new_values, source,
                       voice_transcript)
values (sqlc.arg(at), sqlc.narg(user_id), sqlc.arg(action), sqlc.arg(entity), sqlc.arg(entity_id),
        sqlc.narg(entry_id), sqlc.narg(entry_kind), sqlc.arg(record_types)::text[], sqlc.arg(farm_ids)::uuid[],
        sqlc.narg(record_date), sqlc.narg(reason), sqlc.arg(same_day_author), sqlc.narg(summary),
        sqlc.arg(changes)::jsonb, sqlc.narg(old_values)::jsonb, sqlc.narg(new_values)::jsonb, sqlc.arg(source),
        sqlc.narg(voice_transcript))
returning id;
