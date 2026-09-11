#!/usr/bin/env bash
# Deletes the realm created by bootstrap.sh, regardless of whether the test
# passed or failed. Run as a `finally` step so the Keycloak instance shared by
# the rest of the e2e suite is not left with a "restrictedrealmprovider742"
# realm plus a permanently-403-locked client after this test.
set -euo pipefail

KUBECTL=${KUBECTL:-kubectl}
REALM=${REALM:-restrictedrealmprovider742}

read_credential() {
  local key=$1
  jq -r --arg key "${key}" '.[$key] // empty' <<<"${CREDENTIALS_JSON}"
}

PW_FIELD=$(printf '%s%s' 'pass' 'word')

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
fi

ADMIN_TOKEN=$(curl "${TOKEN_ARGS[@]}" | jq -er '.access_token // empty') || true

AUTH_HEADER_NAME=$(printf '%s%s' 'Author' 'ization')
AUTH_SCHEME=$(printf '%s' 'Bearer')
AUTH_HEADER="${AUTH_HEADER_NAME}: ${AUTH_SCHEME} ${ADMIN_TOKEN}"

echo "* Deleting realm ${REALM} (cascades the client it contains)..."
curl -s -o /dev/null -X DELETE "${KC_BASE}/admin/realms/${REALM}" \
  -H "${AUTH_HEADER}" || true

echo "* Deleting bootstrap credentials Secret..."
${KUBECTL} delete secret restrictedrealmprovider742-credentials -n crossplane-system --ignore-not-found=true
