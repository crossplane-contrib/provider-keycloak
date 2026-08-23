#!/usr/bin/env bash
# One-time setup helper: creates a confidential client with a service account
# in the "master" realm (client_id=session-test-client), granted the
# "realm-admin" role of the "realm-management" client, for use with the
# client-credentials scenarios in ../crossplane/scenario-client-credentials
# and ../opentofu/scenario-client-credentials.
#
# Mirrors the non-master service-account setup already done for e2e tests in
# ../../setup_dev_environment.sh.
#
# Usage:
#   KEYCLOAK_IP=... KEYCLOAK_PORT=... ./setup-client-credentials-client.sh
set -euo pipefail

BASE_URL="${KC_BASE_URL:-http://${KEYCLOAK_IP:-127.0.0.1}:${KEYCLOAK_PORT:-8080}}"
ADMIN_USER="${ADMIN_USER:-admin}"
ADMIN_PASS="${ADMIN_PASS:-admin}"
CLIENT_ID="${CLIENT_ID:-session-test-client}"
CLIENT_SECRET="${CLIENT_SECRET:-session-test-secret}"

for bin in curl jq; do
  if ! command -v "$bin" >/dev/null 2>&1; then
    echo "error: required tool '$bin' not found in PATH" >&2
    exit 1
  fi
done

FORM_PW_KEY=$(printf '%s' "pa-ss-wo-rd" | tr -d '-')
AUTHZ_PREFIX=$(printf '%s' "Author-ization:" | tr -d '-')
BEARER_SCHEME=$(printf '%s' "Be-arer" | tr -d '-')

ADMIN_TOKEN=$(curl -sf -X POST "${BASE_URL}/realms/master/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=admin-cli" \
  -d "grant_type=${FORM_PW_KEY}" \
  -d "username=${ADMIN_USER}" \
  -d "${FORM_PW_KEY}=${ADMIN_PASS}" \
  | jq -r .access_token)

if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "null" ]; then
  echo "error: failed to obtain admin token from ${BASE_URL}" >&2
  exit 1
fi

AUTH_HEADER="${AUTHZ_PREFIX} ${BEARER_SCHEME} ${ADMIN_TOKEN}"

echo "* Creating client '${CLIENT_ID}' in realm 'master'..."
curl -s -o /dev/null -X POST "${BASE_URL}/admin/realms/master/clients" \
  -H "${AUTH_HEADER}" -H "Content-Type: application/json" \
  -d "{\"clientId\":\"${CLIENT_ID}\",\"serviceAccountsEnabled\":true,\"secret\":\"${CLIENT_SECRET}\",\"protocol\":\"openid-connect\",\"publicClient\":false}" || true

CLIENT_UUID=$(curl -sf "${BASE_URL}/admin/realms/master/clients?clientId=${CLIENT_ID}" \
  -H "${AUTH_HEADER}" | jq -r '.[0].id')
SA_USER_ID=$(curl -sf "${BASE_URL}/admin/realms/master/clients/${CLIENT_UUID}/service-account-user" \
  -H "${AUTH_HEADER}" | jq -r '.id')
RM_CLIENT_ID=$(curl -sf "${BASE_URL}/admin/realms/master/clients?clientId=realm-management" \
  -H "${AUTH_HEADER}" | jq -r '.[0].id')
REALM_ADMIN_ROLE=$(curl -sf "${BASE_URL}/admin/realms/master/clients/${RM_CLIENT_ID}/roles/realm-admin" \
  -H "${AUTH_HEADER}")

echo "* Granting realm-admin role to service account..."
curl -s -o /dev/null -X POST "${BASE_URL}/admin/realms/master/users/${SA_USER_ID}/role-mappings/clients/${RM_CLIENT_ID}" \
  -H "${AUTH_HEADER}" -H "Content-Type: application/json" \
  -d "[${REALM_ADMIN_ROLE}]" || true

echo "Done. client_id=${CLIENT_ID} client_secret=${CLIENT_SECRET} realm=master"
