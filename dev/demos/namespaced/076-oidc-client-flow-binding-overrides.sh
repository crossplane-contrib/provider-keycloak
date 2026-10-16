#!/usr/bin/env bash
# uptest post-assert hook for 076-oidc-client-flow-binding-overrides.yaml (#750).
set -euo pipefail
exec "$(dirname "$0")/../../../cluster/test/hack/checkClientFlowBindingOverrides.sh" \
  keycloak.m.crossplane.io flow-overrides-750-ns flow-overrides-750-ns-browser dev-ns
