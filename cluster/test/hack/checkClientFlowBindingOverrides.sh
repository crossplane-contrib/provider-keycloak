#!/usr/bin/env bash
# Regression check for crossplane-contrib/provider-keycloak#750, run as an
# uptest post-assert hook of the client-flow-binding-overrides demos.
#
# Usage: checkClientFlowBindingOverrides.sh <api-group-suffix> <client-name> <flow-name> [namespace]
#   <api-group-suffix> is "keycloak.crossplane.io" (cluster-scoped) or
#   "keycloak.m.crossplane.io" (namespaced).
#
# The bug only shows on an already existing Client whose references resolve
# in different reconciles: the reference resolver's server-side apply pruned
# the fields it had applied before, so a browserIdRef override added later
# was removed from spec.forProvider and never reached Keycloak. Asserts that,
# after an applier adds the override to the existing Client via server-side
# apply (as Argo CD does),
#   1. the browser flow override reaches Keycloak (status.atProvider),
#   2. the override and realmId stay in spec.forProvider across reconciles.
set -euo pipefail

KUBECTL="${KUBECTL:-kubectl}"
GROUP_SUFFIX="$1"
CLIENT="$2"
FLOW="$3"
NAMESPACE="${4:-}"
NS_ARGS=()
NS_META=""
if [[ -n "${NAMESPACE}" ]]; then
  NS_ARGS=(--namespace "${NAMESPACE}")
  NS_META="  namespace: ${NAMESPACE}"
fi
CLIENT_KG="client.openidclient.${GROUP_SUFFIX}"
FLOW_KG="flow.authenticationflow.${GROUP_SUFFIX}"

get() {
  "${KUBECTL}" get "${NS_ARGS[@]}" "$1" -o "jsonpath=$2"
}

wait_for() {
  local kg="$1" jsonpath="$2" expected="$3" actual=""
  for _ in $(seq 1 60); do
    actual="$(get "${kg}" "${jsonpath}")"
    if [[ "${actual}" == "${expected}" ]]; then
      return 0
    fi
    sleep 5
  done
  echo "timed out waiting for ${jsonpath} == '${expected}' on ${kg} (last: '${actual}')"
  return 1
}

check_spec() {
  local ref realm
  ref="$(get "${CLIENT_KG}/${CLIENT}" '{.spec.forProvider.authenticationFlowBindingOverrides[0].browserIdRef.name}')"
  if [[ "${ref}" != "${FLOW}" ]]; then
    echo "spec.forProvider.authenticationFlowBindingOverrides was pruned from ${CLIENT_KG}/${CLIENT} (browserIdRef.name: '${ref}')"
    return 1
  fi
  realm="$(get "${CLIENT_KG}/${CLIENT}" '{.spec.forProvider.realmId}')"
  if [[ -z "${realm}" ]]; then
    echo "spec.forProvider.realmId was pruned from ${CLIENT_KG}/${CLIENT}"
    return 1
  fi
}

flow_id="$(get "${FLOW_KG}/${FLOW}" '{.status.atProvider.id}')"
flow_id="${flow_id##*/}"
if [[ -z "${flow_id}" ]]; then
  echo "${FLOW_KG}/${FLOW} has no status.atProvider.id"
  exit 1
fi

echo "Adding a browserIdRef override to existing ${CLIENT_KG}/${CLIENT} via server-side apply"
"${KUBECTL}" apply --server-side --field-manager=e2e-flow-binding-overrides -f - <<EOF
apiVersion: openidclient.${GROUP_SUFFIX}/v1alpha2
kind: Client
metadata:
  name: ${CLIENT}
${NS_META}
spec:
  forProvider:
    authenticationFlowBindingOverrides:
      - browserIdRef:
          name: ${FLOW}
EOF

wait_for "${CLIENT_KG}/${CLIENT}" '{.status.atProvider.authenticationFlowBindingOverrides[0].browserId}' "${flow_id}"
wait_for "${CLIENT_KG}/${CLIENT}" '{.status.conditions[?(@.type=="Synced")].status}' "True"
check_spec

# Let a few more reconciles (and reference resolutions) pass, then re-check
# that nothing was pruned in the meantime.
sleep 30
check_spec
actual="$(get "${CLIENT_KG}/${CLIENT}" '{.status.atProvider.authenticationFlowBindingOverrides[0].browserId}')"
if [[ "${actual}" != "${flow_id}" ]]; then
  echo "browser flow override drifted to '${actual}', expected '${flow_id}'"
  exit 1
fi
echo "OK: ${CLIENT_KG}/${CLIENT} kept the browser flow override '${flow_id}'"
