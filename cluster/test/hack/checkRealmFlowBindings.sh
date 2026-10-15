#!/usr/bin/env bash
# Regression check for crossplane-contrib/provider-keycloak#751, run as an
# uptest post-assert hook of the realm-flow-bindings demos.
#
# Usage: checkRealmFlowBindings.sh <realm-kind-group> <realm-name> <expected-browser-flow> [namespace]
#
# A Realm that does not declare its flow bindings must not take ownership of
# them, otherwise every Realm update resets the binding applied by a Bindings
# resource (e.g. browserFlow back to "browser") and both oscillate. Asserts
#   1. no flow binding is late-initialized into the Realm's spec.forProvider,
#   2. a Realm update triggered by an unrelated field keeps the browser flow
#      that the Bindings resource applied.
set -euo pipefail

KUBECTL="${KUBECTL:-kubectl}"
KIND_GROUP="$1"
NAME="$2"
EXPECTED_FLOW="$3"
NS_ARGS=()
if [[ -n "${4:-}" ]]; then
  NS_ARGS=(--namespace "$4")
fi

get() {
  "${KUBECTL}" get "${NS_ARGS[@]}" "${KIND_GROUP}/${NAME}" -o "jsonpath=$1"
}

wait_for() {
  local jsonpath="$1" expected="$2" actual=""
  for _ in $(seq 1 60); do
    actual="$(get "${jsonpath}")"
    if [[ "${actual}" == "${expected}" ]]; then
      return 0
    fi
    sleep 5
  done
  echo "timed out waiting for ${jsonpath} == '${expected}' on ${KIND_GROUP}/${NAME} (last: '${actual}')"
  return 1
}

check_not_late_initialized() {
  local field value
  for field in browserFlow registrationFlow directGrantFlow resetCredentialsFlow clientAuthenticationFlow dockerAuthenticationFlow; do
    value="$(get "{.spec.forProvider.${field}}")"
    if [[ -n "${value}" ]]; then
      echo "spec.forProvider.${field} was late-initialized to '${value}' on ${KIND_GROUP}/${NAME}"
      return 1
    fi
  done
}

echo "Checking ${KIND_GROUP}/${NAME} does not own its flow bindings"
wait_for '{.status.atProvider.browserFlow}' "${EXPECTED_FLOW}"
check_not_late_initialized

# Force a full realm update through an unrelated field.
display_name="flow-bindings-update-$(date +%s)"
"${KUBECTL}" patch "${NS_ARGS[@]}" "${KIND_GROUP}/${NAME}" --type=merge \
  -p "{\"spec\":{\"forProvider\":{\"displayName\":\"${display_name}\"}}}"
wait_for '{.status.atProvider.displayName}' "${display_name}"
wait_for '{.status.conditions[?(@.type=="Synced")].status}' "True"

actual="$(get '{.status.atProvider.browserFlow}')"
if [[ "${actual}" != "${EXPECTED_FLOW}" ]]; then
  echo "Realm update reset browserFlow to '${actual}', expected '${EXPECTED_FLOW}' from the Bindings resource"
  exit 1
fi
check_not_late_initialized
echo "OK: ${KIND_GROUP}/${NAME} kept browserFlow '${EXPECTED_FLOW}' after update"
