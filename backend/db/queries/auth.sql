/* Auth: users, login attempts, sessions, farm scope (SPEC §4.3, §6.16, §7.1, §11.4).
   Every time comparison and every timestamp written takes a parameter from the injected clock (plan D5): no now(). */

-- name: GetUserByLogin :one
select id, name, login, password_hash, role, active, last_login_at
from app_user
where login = sqlc.arg(login);

-- name: GetUserByID :one
select id, name, login, password_hash, role, active, last_login_at
from app_user
where id = sqlc.arg(id);

-- name: CountActiveOwners :one
select count(*)::integer as owners
from app_user
where role = 'owner' and active;

-- name: InsertUser :one
insert into app_user (name, login, password_hash, role, password_generated_at, password_generated_by,
                      created_at, created_by, updated_at)
values (sqlc.arg(name), sqlc.arg(login), sqlc.arg(password_hash), sqlc.arg(role), sqlc.arg(now),
        sqlc.narg(created_by), sqlc.arg(now), sqlc.narg(created_by), sqlc.arg(now))
returning id, name, login, role, active, created_at;

-- name: UpdatePasswordHash :exec
-- Re-hash with new argon2id parameters after a successful login (§11.4); not a user change.
update app_user
set password_hash = sqlc.arg(password_hash)
where id = sqlc.arg(id);

-- name: SetLastLogin :exec
update app_user
set last_login_at = sqlc.arg(now)::timestamptz
where id = sqlc.arg(id);

-- name: InsertLoginAttempt :one
insert into login_attempt (login, ip, at, success)
values (sqlc.arg(login), sqlc.narg(ip), sqlc.arg(at), sqlc.arg(success))
returning id;

-- name: MarkLoginAttemptSucceeded :exec
-- A sign-in reserves its attempt as a failure before the password check (§11.4); a success flips it.
update login_attempt
set success = true
where id = sqlc.arg(id);

-- name: LockLoginThrottle :exec
-- Serialises "count the failures, then record the attempt" per login until the transaction ends, so concurrent
-- attempts can't all pass the throttle before any of them is recorded (§11.4). Take it before LockIPThrottle.
select pg_advisory_xact_lock(hashtextextended('login:' || sqlc.arg(login)::text, 0));

-- name: LockIPThrottle :exec
-- The same per client IP; always taken after LockLoginThrottle (one fixed order, no deadlock).
select pg_advisory_xact_lock(hashtextextended('ip:' || host(sqlc.arg(ip)::inet), 0));

-- name: LoginFailuresSince :one
-- Failed attempts for a login after `since`, and the latest of them (= `since` when there are none).
select count(*)::integer as failures, coalesce(max(at), sqlc.arg(since))::timestamptz as latest
from login_attempt
where login = sqlc.arg(login) and not success and at > sqlc.arg(since);

-- name: IPFailuresSince :one
-- Failed attempts from an IP after `since`, and the latest of them (= `since` when there are none).
select count(*)::integer as failures, coalesce(max(at), sqlc.arg(since))::timestamptz as latest
from login_attempt
where ip = sqlc.arg(ip)::inet and not success and at > sqlc.arg(since);

-- name: InsertSession :one
insert into user_session (user_id, token_hash, created_at, last_seen_at, expires_at, user_agent, ip)
values (sqlc.arg(user_id), sqlc.arg(token_hash), sqlc.arg(now), sqlc.arg(now), sqlc.arg(expires_at),
        sqlc.narg(user_agent), sqlc.narg(ip))
returning id;

-- name: GetActiveSession :one
-- The session of a cookie, only while it is unrevoked, unexpired at `now`, and its user is active.
select s.id as session_id, s.user_id, s.last_seen_at, s.expires_at, u.name, u.login, u.role
from user_session s
join app_user u on u.id = s.user_id
where s.token_hash = sqlc.arg(token_hash)
  and s.revoked_at is null
  and s.expires_at > sqlc.arg(now)
  and u.active;

-- name: SlideSession :execrows
-- Moves expires_at forward at most once per `seen_before` window (§11.4: hourly); 1 row = slid (Q2: resend the cookie).
update user_session
set last_seen_at = sqlc.arg(now), expires_at = sqlc.arg(expires_at)
where id = sqlc.arg(id) and revoked_at is null and last_seen_at < sqlc.arg(seen_before);

-- name: RevokeSession :execrows
update user_session
set revoked_at = sqlc.arg(now)::timestamptz
where id = sqlc.arg(id) and revoked_at is null;

-- name: RevokeUserSessions :execrows
-- Password reset, deactivation, any role or farm change: the user signs in again everywhere (§4.3).
update user_session
set revoked_at = sqlc.arg(now)::timestamptz
where user_id = sqlc.arg(user_id) and revoked_at is null;

-- name: ListUserFarms :many
-- The farms of a vet, viewer or manager (owners have no user_farm rows: use ListFarms).
select f.id, f.name, f.type
from farm f
join user_farm uf on uf.farm_id = f.id
where uf.user_id = sqlc.arg(user_id)
order by f.name_key;

-- name: ListFarms :many
select id, name, type
from farm
order by name_key;
