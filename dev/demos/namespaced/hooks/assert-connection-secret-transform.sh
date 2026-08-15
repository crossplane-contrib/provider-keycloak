#!/usr/bin/env bash
# Post-assert hook for dev/demos/namespaced/082-connection-secret-transform.yaml.
#
# Verifies that the connection secret of a namespaced managed resource gains
# the keys renamed by the
# keycloak.crossplane.io/connection-secret-key-rename annotation, in the
# default InPlace mode: the keys are added to the connection secret itself and
# no second secret is created.
set -euo pipefail

KUBECTL="${KUBECTL:-kubectl}"
NAMESPACE="dev-ns"
SOURCE="conn-secret-transform-annotated"
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

# The renamed key carries the value of the original one, in the very same
# secret. The original key stays: the renaming is additive, because the
# managed resource's own controller republishes the keys it owns on every
# reconcile.
assert_key_aliased() {
  local old="$1" new="$2" want got

  want="$(secret_value "${SOURCE}" "${old}")"
  [ -n "${want}" ] || fail "secret ${NAMESPACE}/${SOURCE} has no ${old} key"

  got="$(secret_value "${SOURCE}" "${new}")"
  [ -n "${got}" ] || fail "secret ${NAMESPACE}/${SOURCE} has no ${new} key"
  [ "${want}" = "${got}" ] || fail "secret ${NAMESPACE}/${SOURCE} key ${new} does not carry the value of ${old}"
}

echo "Verifying annotation-driven in-place connection secret renaming for a namespaced resource..."
wait_for_secret "${SOURCE}"

# The keys are added asynchronously, after the connection secret itself has
# been published, so give the transform controller a moment to catch up.
waited=0
until [ -n "$(secret_value "${SOURCE}" "client-secret")" ]; do
  if [ "${waited}" -ge "${TIMEOUT_SECONDS}" ]; then
    fail "secret ${NAMESPACE}/${SOURCE} did not gain the renamed keys within ${TIMEOUT_SECONDS}s"
  fi
  sleep 5
  waited=$((waited + 5))
done

assert_key_aliased "clientID" "client-id"
assert_key_aliased "clientSecret" "client-secret"

# The keys the controller added are recorded on the secret, which is what
# limits it to writing exactly those keys again.
managed="$("${KUBECTL}" get secret "${SOURCE}" --namespace "${NAMESPACE}" \
  -o "jsonpath={.metadata.annotations['keycloak\.crossplane\.io/connection-secret-transform-keys']}")"
for key in "client-id" "client-secret"; do
  case ",${managed}," in
    *",${key},"*) ;;
    *) fail "secret ${NAMESPACE}/${SOURCE} does not record ${key} as a managed key (got '${managed}')" ;;
  esac
done

# InPlace is the default, so no second secret may be created.
if "${KUBECTL}" get secret "${SOURCE}-transformed" --namespace "${NAMESPACE}" >/dev/null 2>&1; then
  fail "unexpected secret ${NAMESPACE}/${SOURCE}-transformed: the InPlace mode must not publish a second secret"
fi

echo "Connection secret transform assertions passed."
