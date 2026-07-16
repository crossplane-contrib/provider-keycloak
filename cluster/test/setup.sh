#!/usr/bin/env bash
set -aeuo pipefail

SCRIPT_DIR=$(cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd)
echo "Run Setup..."

${KUBECTL} apply -f ${SCRIPT_DIR}/../../dev/demos/basic/000-init.yaml
${KUBECTL} apply -f ${SCRIPT_DIR}/../../dev/demos/namespaced/000-init.yaml

# Wait for all provider ManagedResourceDefinitions to be established.
# This ensures all CRDs are available in the API discovery cache before
# tests attempt to apply resources. Without this, a race condition can
# cause "no matches for kind" errors for CRDs established just before
# uptest runs (observed with OidcOpenShiftV4IdentityProvider and
# ClientRegexPolicy in Keycloak 26.4.x).
echo "Waiting for all ManagedResourceDefinitions to be established..."
if ${KUBECTL} api-resources --api-group=apiextensions.crossplane.io --no-headers | grep -q '^managedresourcedefinitions'; then
  if ${KUBECTL} get managedresourcedefinitions.apiextensions.crossplane.io --no-headers 2>/dev/null | grep -q .; then
    ${KUBECTL} wait managedresourcedefinitions.apiextensions.crossplane.io \
      --all --for=condition=Established --timeout=10m
  else
    echo "No ManagedResourceDefinitions found; waiting for Keycloak CRDs instead..."
    mapfile -t keycloak_crds < <(${KUBECTL} get crd -o name | grep -E 'keycloak\.(crossplane\.io|m\.crossplane\.io)$' || true)
    if [ "${#keycloak_crds[@]}" -gt 0 ]; then
      ${KUBECTL} wait --for=condition=Established --timeout=10m "${keycloak_crds[@]}"
    fi
  fi
else
  echo "ManagedResourceDefinition API is not available; waiting for Keycloak CRDs instead..."
  mapfile -t keycloak_crds < <(${KUBECTL} get crd -o name | grep -E 'keycloak\.(crossplane\.io|m\.crossplane\.io)$' || true)
  if [ "${#keycloak_crds[@]}" -gt 0 ]; then
    ${KUBECTL} wait --for=condition=Established --timeout=10m "${keycloak_crds[@]}"
  fi
fi

# Apply org init manifest if KEYCLOAK_VERSION >= 26.6
if [ -n "${KEYCLOAK_VERSION:-}" ]; then
  MIN_KC_VERSION_ORGS="26.6"
  if [ "$(printf '%s\n%s' "$MIN_KC_VERSION_ORGS" "$KEYCLOAK_VERSION" | sort -V | head -n1)" = "$MIN_KC_VERSION_ORGS" ]; then
    echo "Keycloak version ${KEYCLOAK_VERSION} >= ${MIN_KC_VERSION_ORGS}, applying org init manifests..."
    ${KUBECTL} apply -f ${SCRIPT_DIR}/../../dev/demos/orgs/000-init.yaml
  fi
fi

# Uptest creates by default these chainsaw test files in following folder /tmp/uptest-e2e/case/
# * 00-apply.yaml
# * 02-import.yaml
# * 03-delete.yaml

# Chainsaw sometimes runs setup.sh with a working directory different from the
# generated case directory, so resolve files from either location.
CASE_DIR="${PWD}"
if [[ ! -f "${CASE_DIR}/00-apply.yaml" && -f "/tmp/uptest-e2e/case/00-apply.yaml" ]]; then
  CASE_DIR="/tmp/uptest-e2e/case"
fi

rewrite_file() {
  local src="$1"
  local tmp
  tmp="${src}.new"
  sed "s/exec: 20m0s/exec: 60m0s/g" "${src}" | \
    sed "s/apply: 20m0s/apply: 60m0s/g" | \
    sed "s/assert: 20m0s/assert: 60m0s/g" > "${tmp}"
  mv "${tmp}" "${src}"
}

# Increase timeouts
rewrite_file "${CASE_DIR}/00-apply.yaml"
rewrite_file "${CASE_DIR}/02-import.yaml"
sed "s/exec: 20m0s/exec: 60m0s/g" "${CASE_DIR}/03-delete.yaml" > "${CASE_DIR}/03-delete.yaml.new"
mv "${CASE_DIR}/03-delete.yaml.new" "${CASE_DIR}/03-delete.yaml"


# We want to add more import tests that:
# 1. test if it finds the resource if the external name is set to a different value
# 2. test if it finds the resource if the external name is not set

#cp ${SCRIPT_DIR}/hack/patchIncorrectExternalName.sh /tmp/patchIncorrectExternalName.sh
#cp ${SCRIPT_DIR}/hack/patchIncorrectExternalName-ns.sh /tmp/patchIncorrectExternalName-ns.sh
#sed "s/patch.sh/patchIncorrectExternalName.sh/g" 02-import.yaml | sed "s/patch-ns.sh/patchIncorrectExternalName-ns.sh/g" | sed "s/curl/#curl/g" > 02-import-IncorrectExtName.yaml

#cp ${SCRIPT_DIR}/hack/patchRemoveExternalName.sh /tmp/patchRemoveExternalName.sh
#cp ${SCRIPT_DIR}/hack/patchRemoveExternalName-ns.sh /tmp/patchRemoveExternalName-ns.sh
#sed "s/patch.sh/patchRemoveExternalName.sh/g" 02-import.yaml | sed "s/patch-ns.sh/patchRemoveExternalName-ns.sh/g" | sed "s/curl/#curl/g" > 02-import-NoExtName.yaml

#echo "---" >> 02-import.yaml
#cat 02-import-IncorrectExtName.yaml >> 02-import.yaml
#echo "---" >> 02-import.yaml
#cat 02-import-NoExtName.yaml >> 02-import.yaml

#rm 02-import-IncorrectExtName.yaml
#rm 02-import-NoExtName.yaml

cp ${SCRIPT_DIR}/hack/deleteOrdered.sh /tmp/deleteOrdered.sh
sed 's/retry_kubectl "/eval "\/tmp\/deleteOrdered.sh /g' "${CASE_DIR}/03-delete.yaml" > "${CASE_DIR}/03-delete.yaml.new"
mv "${CASE_DIR}/03-delete.yaml.new" "${CASE_DIR}/03-delete.yaml"

