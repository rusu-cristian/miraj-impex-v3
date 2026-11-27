#!/usr/bin/env bash
# Dev/e2e only. The postgres image runs this once, when the data volume is empty.
# Creates the three roles and the database `impex` owned by il_owner (SPEC §11.9, plan T8).
# The roles must exist before `api migrate up`: migration 0002_roles.sql only grants.
# Passwords come from the compose environment (.env); psql variables keep them out of the SQL text.
set -eo pipefail

: "${IL_OWNER_PASSWORD:?IL_OWNER_PASSWORD is not set}"
: "${IL_APP_PASSWORD:?IL_APP_PASSWORD is not set}"
: "${IL_MAINT_PASSWORD:?IL_MAINT_PASSWORD is not set}"

psql --no-password -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname postgres \
  -v owner_pw="$IL_OWNER_PASSWORD" -v app_pw="$IL_APP_PASSWORD" -v maint_pw="$IL_MAINT_PASSWORD" <<'SQL'
create role il_owner login password :'owner_pw';
create role il_app   login password :'app_pw';
create role il_maint login password :'maint_pw';
create database impex owner il_owner;
-- Only the three roles may connect (the owner keeps its rights as owner).
revoke all on database impex from public;
grant connect on database impex to il_app, il_maint;
SQL
