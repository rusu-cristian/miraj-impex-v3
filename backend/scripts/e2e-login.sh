#!/usr/bin/env bash
# End-to-end login through the dev reverse proxy (plan T8, criterion 4). Run with `make e2e` from backend/,
# which starts the compose stack first. Needs docker (compose v2), curl and jq on the host.
#
# Checks: create-owner from the image; login via https://localhost:8443 with the right Origin (200, cookie flags);
# GET /api/auth/me (200, role owner, today); wrong Origin (403); logout (204) and the old cookie then gets 401;
# six bad logins (429 + Retry-After); login_attempt.ip is the host-side client IP; non-/api paths get the placeholder.
# Secrets (password, session token) stay in files under a private temp dir, never on a command line.
set -euo pipefail

cd "$(dirname "$0")/.."

BASE_URL="${E2E_BASE_URL:-https://localhost:8443}"
# Must equal the API's APP_ORIGIN (compose.yaml, .env).
ORIGIN="${E2E_ORIGIN:-$BASE_URL}"
PLACEHOLDER='Impex Line dev proxy: the front end is not served here.'
COOKIE_RE='^il_session=[A-Za-z0-9_-]{43}; Path=/; Max-Age=2592000; HttpOnly; Secure; SameSite=Lax$'

run_id="$(date +%s)"
owner_login="e2e.owner.${run_id}"
nobody_login="e2e.nobody.${run_id}"

tmp="$(mktemp -d)"
chmod 700 "$tmp"
trap 'rm -rf "$tmp"' EXIT

fail() { echo "e2e: FAIL: $*" >&2; exit 1; }
ok() { echo "e2e: ok: $*"; }

compose() { docker compose "$@"; }

# psql as the container superuser over the local socket (dev only). SQL comes on stdin so -v variables work.
db_sql() { compose exec -T db psql -X -q -v ON_ERROR_STOP=1 -U postgres -d impex -tA "$@"; }

# req METHOD PATH [curl args...] -> prints the status; body in $tmp/body, headers in $tmp/headers.
req() {
  local method="$1" path="$2"
  shift 2
  curl --silent --show-error --max-time 15 --cacert "$tmp/ca.crt" \
    -o "$tmp/body" -D "$tmp/headers" -w '%{http_code}' -X "$method" "$@" "$BASE_URL$path"
}

expect_status() { # expect_status WANT GOT WHAT
  [ "$2" = "$1" ] || fail "$3: status $2, want $1; body: $(head -c 300 "$tmp/body")"
}

header_value() { # header_value NAME -> value of the first matching response header (HTTP/2 lower-cases names)
  grep -i "^$1:" "$tmp/headers" | head -n 1 | sed -E 's/^[^:]+:[[:space:]]*//' | tr -d '\r'
}

login_body() { # login_body LOGIN PASSWORD_FILE OUT_FILE
  jq -n --arg login "$1" --rawfile password "$2" '{login: $login, password: ($password | rtrimstr("\n"))}' >"$3"
}

# --- 0. Stack is up and ready -----------------------------------------------------------------------------------------
compose ps --status running --services | grep -qx proxy || fail "the compose stack is not running (make up)"
compose exec -T proxy wget -q -O /dev/null http://api:8080/readyz || fail "api /readyz is not 200"
ok "api /readyz 200"
compose cp proxy:/data/caddy/pki/authorities/local/root.crt "$tmp/ca.crt" >/dev/null 2>&1 \
  || fail "cannot copy Caddy's local root CA from the proxy container"

# Earlier runs' failed logins count towards the per-IP limit (20 in 15 min): start clean.
db_sql <<'SQL' >/dev/null
delete from login_attempt where login like 'e2e.%';
SQL

# --- 1. create-owner from the image ------------------------------------------------------------------------------------
# --force: earlier runs leave active owners behind. The password is printed once on stdout (SPEC §11.9).
compose run --rm --no-deps -T --entrypoint /usr/local/bin/create-owner api \
  --name "Proprietar E2E" --login "$owner_login" --force >"$tmp/create-owner.out" \
  || fail "create-owner exited non-zero"
# Generated passwords use 12 characters from this alphabet (SPEC §11.4); take the last such token.
tr -c 'abcdefghjkmnpqrstuvwxyz23456789' '\n' <"$tmp/create-owner.out" | grep -xE '.{12}' | tail -n 1 >"$tmp/password" || true
[ -s "$tmp/password" ] || fail "create-owner printed no 12-character password"
ok "create-owner $owner_login"

# --- 2. Login with the right Origin ------------------------------------------------------------------------------------
login_body "$owner_login" "$tmp/password" "$tmp/login.json"
status="$(req POST /api/auth/login -H "Origin: $ORIGIN" -H 'Content-Type: application/json' --data-binary "@$tmp/login.json")"
expect_status 200 "$status" "login"
set_cookie="$(header_value set-cookie)"
[[ "$set_cookie" =~ $COOKIE_RE ]] || fail "login Set-Cookie flags: got '${set_cookie%%=*}=…${set_cookie#*;}'"
printf 'Cookie: %s\n' "${set_cookie%%;*}" >"$tmp/cookie"
jq -e '.user.role == "owner" and (.today | type == "string")' "$tmp/body" >/dev/null || fail "login body shape"
ok "login 200, cookie il_session; Path=/; Max-Age=2592000; HttpOnly; Secure; SameSite=Lax"

# --- 3. me ---------------------------------------------------------------------------------------------------------------
status="$(req GET /api/auth/me -H "@$tmp/cookie")"
expect_status 200 "$status" "me"
jq -e '.user.role == "owner" and (.today | type == "string" and length > 0)' "$tmp/body" >/dev/null \
  || fail "me: want user.role owner and today; body: $(head -c 300 "$tmp/body")"
ok "me 200, role owner, today $(jq -r .today "$tmp/body")"

# --- 4. Wrong Origin -----------------------------------------------------------------------------------------------------
status="$(req POST /api/auth/login -H 'Origin: https://evil.example' -H 'Content-Type: application/json' \
  --data-binary "@$tmp/login.json")"
expect_status 403 "$status" "login with a wrong Origin"
jq -e '.error.code == "FORBIDDEN"' "$tmp/body" >/dev/null || fail "wrong Origin: want error.code FORBIDDEN"
ok "wrong Origin 403 FORBIDDEN"

# --- 5. Logout, then the old cookie is refused ---------------------------------------------------------------------------
status="$(req POST /api/auth/logout -H "@$tmp/cookie" -H "Origin: $ORIGIN" -H 'Content-Type: application/json' \
  --data-binary '')"
expect_status 204 "$status" "logout"
status="$(req GET /api/auth/me -H "@$tmp/cookie")"
expect_status 401 "$status" "me after logout (old cookie)"
ok "logout 204; me with the old cookie 401"

# --- 6. Throttling: six bad logins for one login -------------------------------------------------------------------------
# A spoofed X-Forwarded-For from the client must not be believed (checked in step 7).
printf 'wrong-password\n' >"$tmp/bad-password"
login_body "$nobody_login" "$tmp/bad-password" "$tmp/bad.json"
for i in 1 2 3 4 5; do
  status="$(req POST /api/auth/login -H "Origin: $ORIGIN" -H 'Content-Type: application/json' \
    -H 'X-Forwarded-For: 203.0.113.9' --data-binary "@$tmp/bad.json")"
  expect_status 401 "$status" "bad login $i"
done
status="$(req POST /api/auth/login -H "Origin: $ORIGIN" -H 'Content-Type: application/json' \
  -H 'X-Forwarded-For: 203.0.113.9' --data-binary "@$tmp/bad.json")"
expect_status 429 "$status" "bad login 6"
retry_after="$(header_value retry-after)"
[[ "$retry_after" =~ ^[1-9][0-9]*$ ]] || fail "429 without a positive Retry-After header (got '$retry_after')"
jq -e '.error.code == "RATE_LIMITED" and (.error.retryAfterSeconds | type == "number")' "$tmp/body" >/dev/null \
  || fail "429 body: want error.code RATE_LIMITED and retryAfterSeconds"
ok "6th bad login 429, Retry-After $retry_after"

# --- 7. login_attempt.ip is the host-side client IP ----------------------------------------------------------------------
# What the proxy saw as the TCP peer for our requests (Caddy access log), versus the proxy's own address.
client_ip="$(compose logs --no-log-prefix proxy \
  | jq -rR 'fromjson? | select(.msg == "handled request" and .request.uri == "/api/auth/login") | .request.remote_ip' \
  | tail -n 1)"
[ -n "$client_ip" ] || fail "no login request in the proxy access log"
proxy_ip="$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}} {{end}}' "$(compose ps -q proxy)" | awk '{print $1}')"
db_sql -v login="$nobody_login" <<'SQL' >"$tmp/ips"
select distinct host(ip) from login_attempt where login = :'login';
SQL
[ "$(wc -l <"$tmp/ips" | tr -d ' ')" = 1 ] || fail "login_attempt ips for $nobody_login: $(tr '\n' ' ' <"$tmp/ips")"
attempt_ip="$(cat "$tmp/ips")"
[ "$attempt_ip" != "$proxy_ip" ] || fail "login_attempt.ip is the proxy's address $proxy_ip"
[ "$attempt_ip" != "203.0.113.9" ] || fail "login_attempt.ip took the client-sent X-Forwarded-For"
[ "$attempt_ip" = "$client_ip" ] || fail "login_attempt.ip $attempt_ip, want the host-side client IP $client_ip"
ok "login_attempt.ip $attempt_ip (proxy is $proxy_ip)"

# --- 8. Placeholder outside /api -----------------------------------------------------------------------------------------
status="$(req GET /)"
expect_status 200 "$status" "GET /"
grep -qF "$PLACEHOLDER" "$tmp/body" || fail "GET / did not return the placeholder"
ok "GET / placeholder"

echo "e2e: PASS"
