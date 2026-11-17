#!/usr/bin/env bash
# uptest post-assert hook for 075-realm-flow-bindings-ownership.yaml (#751).
set -euo pipefail
exec "$(dirname "$0")/../../../cluster/test/hack/checkRealmFlowBindings.sh" \
  realm.realm.keycloak.crossplane.io flow-bindings-751 flow-bindings-751-browser
