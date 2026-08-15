#!/usr/bin/env bash
# Post-assert hook for dev/demos/namespaced/082-connection-secret-transform.yaml.
#
# Verifies that the connection secret of a namespaced managed resource is
# republished with the keys renamed by the
# keycloak.crossplane.io/connection-secret-key-rename annotation, and that the
# transformed secret is owned by its source secret.
set -euo pipefail

KUBECTL="${KUBECTL:-kubectl}"
NAMESPACE="dev-ns"
SOURCE="conn-secret-transform-annotated"
TRANSFORMED="conn-secret-transform-annotated-transformed"
TIMEOUT_SECONDS=180

fail() {
  echo "ERROR: $*" >&2
  exit 1
}

wait_for_secret() {
  local name="$1" waited=0
  until "${KUBECTL}" get secret "${name}" --namespace "${NAMESPACE}" >/dev/null 2>&1; do
    if [ "${waited}" -ge "${TIMEOUT_SECONDS}" ]; then
      "${KUBECTL}" get secrets --namespace "${NAMESPACE}" >&2 || true
      fail "secret ${NAMESPACE}/${name} was not created within ${TIMEOUT_SECONDS}s"
    fi
    sleep 5
    waited=$((waited + 5))
  done
}

secret_value() {
  # Secret keys may contain dots, which have to be escaped in a jsonpath.
  local key="${2//./\\.}"
  "${KUBECTL}" get secret "$1" --namespace "${NAMESPACE}" -o "jsonpath={.data.${key}}"
}

assert_key_renamed() {
  local old="$1" new="$2" want got

  want="$(secret_value "${SOURCE}" "${old}")"
  [ -n "${want}" ] || fail "source secret ${NAMESPACE}/${SOURCE} has no ${old} key"

  got="$(secret_value "${TRANSFORMED}" "${new}")"
  [ -n "${got}" ] || fail "transformed secret ${NAMESPACE}/${TRANSFORMED} has no ${new} key"
  [ "${want}" = "${got}" ] || fail "transformed secret ${NAMESPACE}/${TRANSFORMED} key ${new} does not carry the value of ${SOURCE}/${old}"

  [ -z "$(secret_value "${TRANSFORMED}" "${old}")" ] || fail "transformed secret ${NAMESPACE}/${TRANSFORMED} still carries the original key ${old}"
  [ -z "$(secret_value "${SOURCE}" "${new}")" ] || fail "source secret ${NAMESPACE}/${SOURCE} was modified: it carries the renamed key ${new}"
}

echo "Verifying annotation-driven connection secret renaming for a namespaced resource..."
wait_for_secret "${SOURCE}"
wait_for_secret "${TRANSFORMED}"
assert_key_renamed "clientID" "client-id"
assert_key_renamed "clientSecret" "client-secret"

owner="$("${KUBECTL}" get secret "${TRANSFORMED}" --namespace "${NAMESPACE}" \
  -o "jsonpath={.metadata.ownerReferences[?(@.controller==true)].name}")"
[ "${owner}" = "${SOURCE}" ] || fail "transformed secret ${NAMESPACE}/${TRANSFORMED} is controlled by '${owner}', want '${SOURCE}'"

# The transformed secret must carry the marker labels the controller relies on
# to recognise its own output and to collect it when it becomes stale. A label
# selector avoids escaping the dots of the label key in a jsonpath.
matched="$("${KUBECTL}" get secret "${TRANSFORMED}" --namespace "${NAMESPACE}" -o name \
  --selector "keycloak.crossplane.io/connection-secret-transform=true,keycloak.crossplane.io/connection-secret-source=${SOURCE}")"
[ -n "${matched}" ] || fail "transformed secret ${NAMESPACE}/${TRANSFORMED} does not carry the transform/source labels of ${SOURCE}"

echo "Connection secret transform assertions passed."
