#!/usr/bin/env bash
# This script adds +nullable markers to generated type fields that need to
# accept null values in server-side apply patches. This is needed because
# upjet does not natively support the +nullable marker.
#
# Background (issue #425):
# When a compositeRolesSelector no longer matches any roles, the reference
# resolver produces nil for CompositeRoles and CompositeRolesRefs. The managed
# reconciler's JSON merge patch then emits "compositeRoles": null to clear the
# field. Without nullable: true in the CRD schema, this valid patch is rejected.
set -euo pipefail

REPO_ROOT="${1:-.}"

# Add +nullable marker before CompositeRoles and CompositeRolesRefs field
# declarations in the generated role types for both cluster and namespaced APIs.
for f in \
  "${REPO_ROOT}/apis/cluster/role/v1alpha1/zz_role_types.go" \
  "${REPO_ROOT}/apis/namespaced/role/v1alpha1/zz_role_types.go"; do
  if [ ! -f "$f" ]; then
    echo "WARNING: $f not found, skipping" >&2
    continue
  fi
  # Skip if already patched (idempotency check)
  if grep -q '+nullable' "$f"; then
    continue
  fi
  # Insert "// +nullable" before each CompositeRoles []*string declaration
  sed -i '/^\tCompositeRoles \[\]\*string.*tf:"composite_roles/i \\t// +nullable' "$f"
  # Insert "// +nullable" before each CompositeRolesRefs [] declaration
  sed -i '/^\tCompositeRolesRefs \[/i \\t// +nullable' "$f"
done
