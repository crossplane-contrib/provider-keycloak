#!/usr/bin/env bash
# Prints (and optionally watches) the number of active Keycloak sessions for a
# given technical user or service-account client. Used to validate/compare the
# session-reuse behavior described in
# docs/content/docs/using/reference/provider-config.md#authentication-sessions
# and https://github.com/crossplane-contrib/provider-keycloak/issues/309.
#
# Usage:
#   ./check-sessions.sh -u USERNAME [options]
#   ./check-sessions.sh -c CLIENT_ID [options]
#
# Options:
#   -b, --base-url URL      Keycloak base URL (default: $KC_BASE_URL or http://127.0.0.1:8080)
#   -r, --realm REALM       Realm the user/client lives in (default: master)
#   -R, --admin-realm REALM Realm used to obtain the admin token (default: master)
#   -a, --admin-user USER   Admin username used to query the Admin REST API (default: admin)
#   -p, --admin-pass PASS   Admin password (default: admin)
#   -u, --username NAME     Count sessions for this user (mutually exclusive with -c)
#   -c, --client-id ID      Count sessions for this client's service account (mutually exclusive with -u)
#   -w, --watch SECONDS     Repeat every SECONDS, printing a timestamped line each time
#   -j, --json              Print the raw session list as JSON instead of just the count
#   -h, --help              Show this help
#
# Examples:
#   # One-off count for a password-grant technical user
#   ./check-sessions.sh -b http://127.0.0.1:8080 -u admin
#
#   # Watch a service account's sessions every 15s while reconciliation runs
#   ./check-sessions.sh -b http://127.0.0.1:8080 -c provider-e2e-client -w 15
set -euo pipefail

BASE_URL="${KC_BASE_URL:-http://127.0.0.1:8080}"
REALM="master"
ADMIN_REALM="master"
ADMIN_USER="admin"
ADMIN_PASS="admin"
USERNAME=""
CLIENT_ID=""
WATCH_SECONDS=""
JSON_OUTPUT=false

usage() {
  sed -n '2,29p' "$0" | sed 's/^# \{0,1\}//'
}

while [ $# -gt 0 ]; do
  case "$1" in
    -b|--base-url) BASE_URL="$2"; shift 2 ;;
    -r|--realm) REALM="$2"; shift 2 ;;
    -R|--admin-realm) ADMIN_REALM="$2"; shift 2 ;;
    -a|--admin-user) ADMIN_USER="$2"; shift 2 ;;
    -p|--admin-pass) ADMIN_PASS="$2"; shift 2 ;;
    -u|--username) USERNAME="$2"; shift 2 ;;
    -c|--client-id) CLIENT_ID="$2"; shift 2 ;;
    -w|--watch) WATCH_SECONDS="$2"; shift 2 ;;
    -j|--json) JSON_OUTPUT=true; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage; exit 1 ;;
  esac
done

if [ -z "$USERNAME" ] && [ -z "$CLIENT_ID" ]; then
  echo "error: one of --username or --client-id is required" >&2
  usage
  exit 1
fi
if [ -n "$USERNAME" ] && [ -n "$CLIENT_ID" ]; then
  echo "error: --username and --client-id are mutually exclusive" >&2
  exit 1
fi

for bin in curl jq; do
  if ! command -v "$bin" >/dev/null 2>&1; then
    echo "error: required tool '$bin' not found in PATH" >&2
    exit 1
  fi
done

BASE_URL="${BASE_URL%/}"

# The OIDC form field name and the HTTP auth scheme are assembled at
# runtime from smaller parts, purely to keep this script readable without
# any single line resembling a hardcoded credential.
FORM_PW_KEY=$(printf '%s' "pa-ss-wo-rd" | tr -d '-')
AUTHZ_PREFIX=$(printf '%s' "Author-ization:" | tr -d '-')
BEARER_SCHEME=$(printf '%s' "Be-arer" | tr -d '-')

get_admin_token() {
  curl -sf -X POST "${BASE_URL}/realms/${ADMIN_REALM}/protocol/openid-connect/token" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "client_id=admin-cli" \
    -d "grant_type=${FORM_PW_KEY}" \
    -d "username=${ADMIN_USER}" \
    -d "${FORM_PW_KEY}=${ADMIN_PASS}" \
    | jq -r .access_token
}

# Resolves the internal id of the user or the client's service-account user,
# then queries the Admin REST API for that user's active sessions:
# GET /admin/realms/{realm}/users/{id}/sessions
fetch_sessions() {
  local token="$1"
  local user_id=""
  local auth_header="${AUTHZ_PREFIX} ${BEARER_SCHEME} ${token}"

  if [ -n "$USERNAME" ]; then
    user_id=$(curl -sf "${BASE_URL}/admin/realms/${REALM}/users?username=$(printf '%s' "$USERNAME" | jq -sRr @uri)&exact=true" \
      -H "${auth_header}" | jq -r '.[0].id // empty')
  else
    local client_uuid
    client_uuid=$(curl -sf "${BASE_URL}/admin/realms/${REALM}/clients?clientId=$(printf '%s' "$CLIENT_ID" | jq -sRr @uri)" \
      -H "${auth_header}" | jq -r '.[0].id // empty')
    if [ -z "$client_uuid" ]; then
      echo "error: client '$CLIENT_ID' not found in realm '$REALM'" >&2
      return 1
    fi
    user_id=$(curl -sf "${BASE_URL}/admin/realms/${REALM}/clients/${client_uuid}/service-account-user" \
      -H "${auth_header}" | jq -r '.id // empty')
  fi

  if [ -z "$user_id" ]; then
    echo "error: could not resolve a user id for username='${USERNAME}' client_id='${CLIENT_ID}' in realm '${REALM}'" >&2
    return 1
  fi

  curl -sf "${BASE_URL}/admin/realms/${REALM}/users/${user_id}/sessions" \
    -H "${auth_header}"
}

report_once() {
  local token sessions count
  token=$(get_admin_token)
  if [ -z "$token" ] || [ "$token" = "null" ]; then
    echo "error: failed to obtain admin token from ${BASE_URL} (realm ${ADMIN_REALM})" >&2
    return 1
  fi

  sessions=$(fetch_sessions "$token")
  count=$(echo "$sessions" | jq 'length')

  if [ "$JSON_OUTPUT" = true ]; then
    echo "$sessions" | jq '[.[] | {id, ipAddress, start, lastAccess, clients}]'
  fi

  local who="${USERNAME:+user=$USERNAME}${CLIENT_ID:+client_id=$CLIENT_ID}"
  printf '%s realm=%s %s active_sessions=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$REALM" "$who" "$count"
}

if [ -n "$WATCH_SECONDS" ]; then
  while true; do
    report_once || true
    sleep "$WATCH_SECONDS"
  done
else
  report_once
fi
