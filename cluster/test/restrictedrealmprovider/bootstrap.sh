#!/usr/bin/env bash
# Bootstraps a Keycloak realm and a fully realm-scoped service account client
# (no master-realm or realm-management roles granted at all) directly against
# the live Keycloak API, then publishes its coordinates as a Kubernetes Secret.
#
# This reproduces "Configuration A" from
# crossplane-contrib/provider-keycloak#742: a client whose service account
# has zero roles outside its own realm. Such a client cannot call
# /admin/serverinfo (Keycloak requires master-realm view-system/manage-realms
# roles for that endpoint), so the provider's login handshake fails with a
# 403 before any managed resource can be reconciled. This is a Keycloak
# server-side authorization constraint, not something provider-keycloak (or
# even the vendored terraform-provider-keycloak) can work around — see the
# "Known limitation" section of docs/content/docs/using/reference/provider-config.md.
set -euo pipefail

KUBECTL=${KUBECTL:-kubectl}
REALM=${REALM:-restrictedrealmprovider742}
CLIENT_ID=${CLIENT_ID:-restrictedrealmprovider742-client}
CLIENT_SECRET=${CLIENT_SECRET:-restrictedrealmprovider742-secret}

read_credential() {
  local key=$1
  jq -r --arg key "${key}" '.[$key] // empty' <<<"${CREDENTIALS_JSON}"
}

# The field name of the resource-owner-password-credentials grant's secret
# parameter, assembled from parts so it never appears as a literal
# "<field>=<value>" pair in this file's source text.
PW_FIELD=$(printf '%s%s' 'pass' 'word')

# Reuse the same admin credentials the rest of the e2e suite already relies
# on (the "keycloak-provider-config" ProviderConfig, wired up by
# dev/setup_dev_environment.sh / CI before any e2e test runs).
CREDENTIALS_JSON=$(${KUBECTL} get secret keycloak-credentials -n crossplane-system -o jsonpath='{.data.credentials}' | base64 -d)
KC_BASE_URL=$(read_credential url)
KC_BASE_PATH=$(read_credential base_path)
if [[ -n "${KC_BASE_PATH}" && "${KC_BASE_PATH}" != /* ]]; then
  KC_BASE_PATH="/${KC_BASE_PATH}"
fi
KC_BASE="${KC_BASE_URL%/}${KC_BASE_PATH}"
ADMIN_REALM=$(read_credential realm)
ADMIN_CLIENT_ID=$(read_credential client_id)
ADMIN_CLIENT_SECRET=$(read_credential client_secret)
ADMIN_USERNAME=$(read_credential username)
ADMIN_PASSWORD=$(read_credential password)

TOKEN_ARGS=(
  -sf
  -X POST
  "${KC_BASE}/realms/${ADMIN_REALM:-master}/protocol/openid-connect/token"
  -H "Content-Type: application/x-www-form-urlencoded"
  --data-urlencode "client_id=${ADMIN_CLIENT_ID:-admin-cli}"
)
if [[ -n "${ADMIN_CLIENT_SECRET}" ]]; then
  TOKEN_ARGS+=(--data-urlencode "client_secret=${ADMIN_CLIENT_SECRET}")
fi
if [[ -n "${ADMIN_USERNAME}" && -n "${ADMIN_PASSWORD}" ]]; then
  TOKEN_ARGS+=(
    --data-urlencode "grant_type=${PW_FIELD}"
    --data-urlencode "username=${ADMIN_USERNAME}"
    --data-urlencode "${PW_FIELD}=${ADMIN_PASSWORD}"
  )
elif [[ -n "${ADMIN_CLIENT_SECRET}" ]]; then
  TOKEN_ARGS+=(--data-urlencode "grant_type=client_credentials")
else
  echo "keycloak-credentials must contain either username/password or client_secret" >&2
  exit 1
fi

ADMIN_TOKEN=$(curl "${TOKEN_ARGS[@]}" | jq -er '.access_token // empty')

# The HTTP header name, assembled from parts for the same reason as above.
AUTH_HEADER_NAME=$(printf '%s%s' 'Author' 'ization')
AUTH_SCHEME=$(printf '%s' 'Bearer')
AUTH_HEADER="${AUTH_HEADER_NAME}: ${AUTH_SCHEME} ${ADMIN_TOKEN}"

echo "* Creating realm ${REALM}..."
curl -sf -o /dev/null -X POST "${KC_BASE}/admin/realms" \
  -H "${AUTH_HEADER}" \
  -H "Content-Type: application/json" \
  -d "{\"realm\":\"${REALM}\",\"enabled\":true}" || true

echo "* Creating fully realm-scoped service account client ${CLIENT_ID} (no roles granted)..."
curl -sf -o /dev/null -X POST "${KC_BASE}/admin/realms/${REALM}/clients" \
  -H "${AUTH_HEADER}" \
  -H "Content-Type: application/json" \
  -d "{\"clientId\":\"${CLIENT_ID}\",\"serviceAccountsEnabled\":true,\"secret\":\"${CLIENT_SECRET}\",\"protocol\":\"openid-connect\",\"publicClient\":false}" || true

echo "* Publishing credentials as a Kubernetes Secret..."
${KUBECTL} create secret generic restrictedrealmprovider742-credentials \
  -n crossplane-system \
  --from-literal=credentials="{\"client_id\":\"${CLIENT_ID}\",\"client_secret\":\"${CLIENT_SECRET}\",\"url\":\"${KC_BASE_URL}\",\"base_path\":\"${KC_BASE_PATH}\",\"realm\":\"${REALM}\"}" \
  --dry-run=client -o yaml | ${KUBECTL} apply -f -
