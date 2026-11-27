-- +goose Up
-- Grants for the two runtime roles (SPEC §4.10, §11.9). The roles themselves are created by infrastructure
-- (deploy/initdb in compose, internal/testdb in tests, the DBA in production) before this migration runs.
-- Migrations run as the owner role (MIGRATE_DATABASE_URL), which owns every object and needs no grant.
-- A later migration that adds a table or view grants on it explicitly (no default privileges: append-only
-- tables must not get update/delete by accident).

grant usage on schema public to il_app, il_maint;

-- il_app (api serve): read and write every table and view ...
grant select, insert, update, delete on all tables in schema public to il_app;
-- ... except the append-only ones: audit_log and live_event (SPEC §4.10) and request_replay (§4.11) ...
revoke update, delete on audit_log, live_event, request_replay from il_app;
-- ... and goose's version table, which /readyz only reads (SPEC §11.12).
revoke insert, update, delete on goose_db_version from il_app;
grant usage, select on all sequences in schema public to il_app;

-- il_maint (the purge loop of api serve, SPEC §9.1): read and delete expired rows only.
grant select, delete on live_event, login_attempt, request_replay to il_maint;

-- +goose Down
revoke all on all tables in schema public from il_app, il_maint;
revoke all on all sequences in schema public from il_app, il_maint;
revoke usage on schema public from il_app, il_maint;
