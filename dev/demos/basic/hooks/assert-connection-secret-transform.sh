#!/usr/bin/env bash
# Post-assert hook for dev/demos/basic/082-connection-secret-transform.yaml.
#
# Verifies that the connection secrets of the two clients are republished with
# renamed keys - once configured on the ProviderConfig, once with annotations
# on the resource itself - and that the transformed secrets are owned by their
# source secret so they are garbage-collected with the managed resource.
set -euo pipefail

KUBECTL="${KUBECTL:-kubectl}"
NAMESPACE="dev"
TIMEOUT_SECONDS=180

fail() {
  echo "ERROR: $*" >&2
  exit 1
}

# wait_for_secret <name>
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

# secret_value <name> <key> - empty output when the key does not exist
secret_value() {
  # Secret keys may contain dots, which have to be escaped in a jsonpath.
  local key="${2//./\\.}"
  "${KUBECTL}" get secret "$1" --namespace "${NAMESPACE}" -o "jsonpath={.data.${key}}"
}

# assert_key_renamed <source secret> <transformed secret> <old key> <new key>
assert_key_renamed() {
  local src="$1" dst="$2" old="$3" new="$4" want got

  want="$(secret_value "${src}" "${old}")"
  [ -n "${want}" ] || fail "source secret ${NAMESPACE}/${src} has no ${old} key"

  got="$(secret_value "${dst}" "${new}")"
  [ -n "${got}" ] || fail "transformed secret ${NAMESPACE}/${dst} has no ${new} key"
  [ "${want}" = "${got}" ] || fail "transformed secret ${NAMESPACE}/${dst} key ${new} does not carry the value of ${src}/${old}"

  [ -z "$(secret_value "${dst}" "${old}")" ] || fail "transformed secret ${NAMESPACE}/${dst} still carries the original key ${old}"

  # The source secret must be left untouched, otherwise upjet can no longer
  # rebuild the Terraform state from it.
  [ -z "$(secret_value "${src}" "${new}")" ] || fail "source secret ${NAMESPACE}/${src} was modified: it carries the renamed key ${new}"
}

# assert_owned_by <transformed secret> <source secret>
assert_owned_by() {
  local dst="$1" src="$2" owner matched
  owner="$("${KUBECTL}" get secret "${dst}" --namespace "${NAMESPACE}" \
    -o "jsonpath={.metadata.ownerReferences[?(@.controller==true)].name}")"
  [ "${owner}" = "${src}" ] || fail "transformed secret ${NAMESPACE}/${dst} is controlled by '${owner}', want '${src}'"

  # The marker labels are what let the controller recognise its own output and
  # collect it once it becomes stale. Matching them with a label selector
  # avoids escaping the dots of the label key in a jsonpath.
  matched="$("${KUBECTL}" get secret "${dst}" --namespace "${NAMESPACE}" -o name \
    --selector "keycloak.crossplane.io/connection-secret-transform=true,keycloak.crossplane.io/connection-secret-source=${src}")"
  [ -n "${matched}" ] || fail "transformed secret ${NAMESPACE}/${dst} does not carry the transform/source labels of ${src}"
}

echo "Verifying ProviderConfig-driven connection secret renaming..."
wait_for_secret "conn-secret-transform-pc"
wait_for_secret "conn-secret-transform-pc-transformed"
assert_key_renamed "conn-secret-transform-pc" "conn-secret-transform-pc-transformed" "clientID" "client-id"
assert_key_renamed "conn-secret-transform-pc" "conn-secret-transform-pc-transformed" "clientSecret" "client-secret"
assert_owned_by "conn-secret-transform-pc-transformed" "conn-secret-transform-pc"

echo "Verifying annotation-driven connection secret renaming..."
wait_for_secret "conn-secret-transform-annotated"
wait_for_secret "conn-secret-transform-envoy"
assert_key_renamed "conn-secret-transform-annotated" "conn-secret-transform-envoy" "clientID" "client-id"
assert_key_renamed "conn-secret-transform-annotated" "conn-secret-transform-envoy" "clientSecret" "client-secret"
assert_key_renamed "conn-secret-transform-annotated" "conn-secret-transform-envoy" "serviceAccountUserId" "service-account-user-id"
assert_owned_by "conn-secret-transform-envoy" "conn-secret-transform-annotated"

# The default ProviderConfig configures no renaming, so the annotated client
# must not have produced the default "-transformed" secret in addition to the
# name it asked for.
if "${KUBECTL}" get secret "conn-secret-transform-annotated-transformed" --namespace "${NAMESPACE}" >/dev/null 2>&1; then
  fail "unexpected secret ${NAMESPACE}/conn-secret-transform-annotated-transformed: the transform name annotation was ignored"
fi

echo "Connection secret transform assertions passed."
