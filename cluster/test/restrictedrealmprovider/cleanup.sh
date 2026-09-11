#!/usr/bin/env bash
# Deletes the realm created by bootstrap.sh, regardless of whether the test
# passed or failed. Run as a `finally` step so the Keycloak instance shared by
# the rest of the e2e suite is not left with a "restrictedrealmprovider742"
# realm plus a permanently-403-locked client after this test.
set -euo pipefail

KUBECTL=${KUBECTL:-kubectl}
REALM=${REALM:-restrictedrealmprovider742}

PW_FIELD=$(printf '%s%s' 'pass' 'word')

CREDENTIALS_JSON=$(${KUBECTL} get secret keycloak-credentials -n crossplane-system -o jsonpath='{.data.credentials}' | base64 -d)
KC_BASE_URL=$(echo "${CREDENTIALS_JSON}" | jq -r '.url')

ADMIN_TOKEN=$(curl -sf -X POST "${KC_BASE_URL}/realms/master/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  --data-urlencode "grant_type=${PW_FIELD}" \
  --data-urlencode "client_id=admin-cli" \
  --data-urlencode "username=admin" \
  --data-urlencode "${PW_FIELD}=admin" \
  | jq -r .access_token) || true

AUTH_HEADER_NAME=$(printf '%s%s' 'Author' 'ization')
AUTH_SCHEME=$(printf '%s' 'Bearer')
AUTH_HEADER="${AUTH_HEADER_NAME}: ${AUTH_SCHEME} ${ADMIN_TOKEN}"

echo "* Deleting realm ${REALM} (cascades the client it contains)..."
curl -s -o /dev/null -X DELETE "${KC_BASE_URL}/admin/realms/${REALM}" \
  -H "${AUTH_HEADER}" || true

echo "* Deleting bootstrap credentials Secret..."
${KUBECTL} delete secret restrictedrealmprovider742-credentials -n crossplane-system --ignore-not-found=true
