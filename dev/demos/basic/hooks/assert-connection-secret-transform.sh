#!/usr/bin/env bash
# Post-assert hook for dev/demos/basic/082-connection-secret-transform.yaml.
#
# Verifies both transform modes:
#
#   1. InPlace (the default), configured on the ProviderConfig: the renamed
#      keys are added to the client's own connection secret and no second
#      secret is created.
#   2. SeparateSecret, configured with annotations on the resource itself: the
#      connection secret is left untouched and a transformed secret is
#      published next to it, owned by its source secret so that it is
#      garbage-collected with the managed resource.
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

# assert_key_aliased <secret> <old key> <new key> - InPlace mode: <new key>
# carries the value of <old key> in the very same secret, and <old key> is
# still there (the renaming is additive, since the managed resource's own
# controller republishes the keys it owns on every reconcile).
assert_key_aliased() {
  local secret="$1" old="$2" new="$3" want got

  want="$(secret_value "${secret}" "${old}")"
  [ -n "${want}" ] || fail "secret ${NAMESPACE}/${secret} has no ${old} key"

  got="$(secret_value "${secret}" "${new}")"
  [ -n "${got}" ] || fail "secret ${NAMESPACE}/${secret} has no ${new} key"
  [ "${want}" = "${got}" ] || fail "secret ${NAMESPACE}/${secret} key ${new} does not carry the value of ${old}"
}

# assert_managed_keys <secret> <key>... - the keys the controller added are
# recorded on the secret, which is what limits it to writing those keys only.
assert_managed_keys() {
  local secret="$1" managed key
  shift
  managed="$("${KUBECTL}" get secret "${secret}" --namespace "${NAMESPACE}" \
    -o "jsonpath={.metadata.annotations['keycloak\.crossplane\.io/connection-secret-transform-keys']}")"
  for key in "$@"; do
    case ",${managed}," in
      *",${key},"*) ;;
      *) fail "secret ${NAMESPACE}/${secret} does not record ${key} as a managed key (got '${managed}')" ;;
    esac
  done
}

# assert_added_value <secret> <key> <expected value>
assert_added_value() {
  local secret="$1" key="$2" want="$3" got
  got="$(secret_value "${secret}" "${key}" | base64 -d)"
  [ "${got}" = "${want}" ] || fail "secret ${NAMESPACE}/${secret} key ${key} is '${got}', want '${want}'"
}

# assert_added_suffix <secret> <key> <expected suffix>
assert_added_suffix() {
  local secret="$1" key="$2" want="$3" got
  got="$(secret_value "${secret}" "${key}" | base64 -d)"
  case "${got}" in
    *"${want}") ;;
    *) fail "secret ${NAMESPACE}/${secret} key ${key} is '${got}', want a value ending in '${want}'" ;;
  esac
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

echo "Verifying ProviderConfig-driven in-place connection secret renaming..."
wait_for_secret "conn-secret-transform-pc"
# The keys are added asynchronously, after the connection secret itself has
# been published, so give the transform controller a moment to catch up.
waited=0
until [ -n "$(secret_value "conn-secret-transform-pc" "client-secret")" ]; do
  if [ "${waited}" -ge "${TIMEOUT_SECONDS}" ]; then
    fail "secret ${NAMESPACE}/conn-secret-transform-pc did not gain the renamed keys within ${TIMEOUT_SECONDS}s"
  fi
  sleep 5
  waited=$((waited + 5))
done
assert_key_aliased "conn-secret-transform-pc" "clientID" "client-id"
assert_key_aliased "conn-secret-transform-pc" "clientSecret" "client-secret"
assert_managed_keys "conn-secret-transform-pc" "client-id" "client-secret" "issuerUrl" "discoveryUrl" "realmName"

# Keycloak-derived fields ("keycloak:" sources): the realm's OIDC issuer is
# built from the ProviderConfig's Keycloak URL and the client's realm, so it
# exists on neither object on its own.
assert_added_value "conn-secret-transform-pc" "realmName" "dev"
assert_added_suffix "conn-secret-transform-pc" "issuerUrl" "/realms/dev"
assert_added_suffix "conn-secret-transform-pc" "discoveryUrl" "/realms/dev/.well-known/openid-configuration"

# InPlace is the default, so no second secret may be created for it.
if "${KUBECTL}" get secret "conn-secret-transform-pc-transformed" --namespace "${NAMESPACE}" >/dev/null 2>&1; then
  fail "unexpected secret ${NAMESPACE}/conn-secret-transform-pc-transformed: the InPlace mode must not publish a second secret"
fi

echo "Verifying annotation-driven connection secret renaming into a separate secret..."
wait_for_secret "conn-secret-transform-annotated"
wait_for_secret "conn-secret-transform-envoy"
assert_key_renamed "conn-secret-transform-annotated" "conn-secret-transform-envoy" "clientID" "client-id"
assert_key_renamed "conn-secret-transform-annotated" "conn-secret-transform-envoy" "clientSecret" "client-secret"
assert_key_renamed "conn-secret-transform-annotated" "conn-secret-transform-envoy" "serviceAccountUserId" "service-account-user-id"
assert_owned_by "conn-secret-transform-envoy" "conn-secret-transform-annotated"

# In SeparateSecret mode the connection secret keeps only the keys the
# provider itself publishes.
[ -z "$(secret_value "conn-secret-transform-annotated" "client-id")" ] || fail "source secret ${NAMESPACE}/conn-secret-transform-annotated was modified in SeparateSecret mode"

# The default ProviderConfig configures no renaming, so the annotated client
# must not have produced the default "-transformed" secret in addition to the
# name it asked for.
if "${KUBECTL}" get secret "conn-secret-transform-annotated-transformed" --namespace "${NAMESPACE}" >/dev/null 2>&1; then
  fail "unexpected secret ${NAMESPACE}/conn-secret-transform-annotated-transformed: the transform name annotation was ignored"
fi

echo "Connection secret transform assertions passed."
