#!/usr/bin/env bash
# gen-llms.sh — regenerate docs/static/llms.txt and docs/static/llms-full.txt
# from the Hugo content tree.
#
# Usage:
#   ./docs/scripts/gen-llms.sh              # write files
#   ./docs/scripts/gen-llms.sh --check      # exit non-zero if files are stale

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
CONTENT_DIR="${REPO_ROOT}/docs/content"
STATIC_DIR="${REPO_ROOT}/docs/static"
BASE_URL="https://crossplane-contrib.github.io/provider-keycloak"

CHECK_MODE=false
if [[ "${1:-}" == "--check" ]]; then
  CHECK_MODE=true
fi

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

# Extract front-matter field value from a markdown file.
# Strips surrounding YAML quotes if present.
# Usage: extract_field <file> <field>
extract_field() {
  local file="$1" field="$2"
  awk '/^---/{found++; next} found==1 && /^'"${field}"':/{sub(/^'"${field}"':[[:space:]]*/,""); gsub(/^["'"'"']|["'"'"']$/,""); print; exit}' "$file"
}

# Emit one llms.txt line for a file.
# Prefers llmsDescription over description; skips if no title.
emit_line() {
  local file="$1"
  local rel="${file#"${CONTENT_DIR}/"}"
  rel="${rel%.md}"
  local url="${BASE_URL}/${rel}/"
  local title; title=$(extract_field "$file" "title")
  [[ -z "$title" ]] && return
  local desc; desc=$(extract_field "$file" "llmsDescription")
  [[ -z "$desc" ]] && desc=$(extract_field "$file" "description")
  if [[ -n "$desc" ]]; then
    echo "- [${title}](${url}): ${desc}"
  else
    echo "- [${title}](${url})"
  fi
}

# ---------------------------------------------------------------------------
# Build the llms.txt index
# ---------------------------------------------------------------------------

generate_llms_txt() {
  cat <<'HEADER'
# provider-keycloak llms.txt
# See: https://llmstxt.org/

> provider-keycloak is a Crossplane provider that manages Keycloak (IAM/SSO) resources as Kubernetes custom resources. It is generated with Upjet from the Keycloak Terraform provider. Declare Keycloak realms, clients, users, groups, roles, identity providers, and more as YAML; the provider reconciles them continuously.

## Getting Started

HEADER

  while IFS= read -r -d '' file; do
    emit_line "$file"
  done < <(find "${CONTENT_DIR}/docs/using/getting-started" -name "*.md" ! -name "_index.md" -print0 | sort -z)

  echo ""
  echo "## Resources"
  echo ""

  while IFS= read -r -d '' file; do
    emit_line "$file"
  done < <(find "${CONTENT_DIR}/docs/using/resources" -name "*.md" ! -name "_index.md" -print0 | sort -z)

  echo ""
  echo "## Reference"
  echo ""

  while IFS= read -r -d '' file; do
    emit_line "$file"
  done < <(find "${CONTENT_DIR}/docs/using/reference" -name "*.md" ! -name "_index.md" -print0 | sort -z)

  echo ""
  echo "## AI"
  echo ""

  while IFS= read -r -d '' file; do
    emit_line "$file"
  done < <(find "${CONTENT_DIR}/docs/ai-usage" -name "*.md" ! -name "_index.md" -print0 | sort -z)

  echo ""
  echo "## Optional"
  echo ""

  while IFS= read -r -d '' file; do
    emit_line "$file"
  done < <(find "${CONTENT_DIR}/docs/developing" -name "*.md" ! -name "_index.md" -print0 | sort -z)
}

# ---------------------------------------------------------------------------
# Build the llms-full.txt content dump
# ---------------------------------------------------------------------------

generate_llms_full_txt() {
  local out=""
  out+="# provider-keycloak — full content for AI assistants"$'\n'
  out+="# Generated from docs/content. See https://llmstxt.org/"$'\n'
  out+=""$'\n'

  while IFS= read -r -d '' file; do
    rel="${file#"${CONTENT_DIR}/"}"
    rel="${rel%.md}"
    url="${BASE_URL}/${rel}/"

    title=$(extract_field "$file" "title")
    [[ -z "$title" ]] && title="$(basename "$rel")"

    out+="## ${title}"$'\n'
    out+="URL: ${url}"$'\n'
    out+=""$'\n'
    # Strip front matter, append body
    body=$(awk '/^---/{found++; next} found>=2' "$file")
    out+="${body}"$'\n'
    out+=""$'\n'
  done < <(find "${CONTENT_DIR}/docs" -name "*.md" ! -name "_index.md" -print0 | sort -z)

  echo "$out"
}

# ---------------------------------------------------------------------------
# Write or check
# ---------------------------------------------------------------------------

LLMS_TXT="${STATIC_DIR}/llms.txt"
LLMS_FULL_TXT="${STATIC_DIR}/llms-full.txt"

NEW_LLMS=$(generate_llms_txt)
NEW_LLMS_FULL=$(generate_llms_full_txt)

if $CHECK_MODE; then
  FAIL=false

  CURRENT_LLMS=""
  [[ -f "$LLMS_TXT" ]] && CURRENT_LLMS=$(cat "$LLMS_TXT")

  CURRENT_LLMS_FULL=""
  [[ -f "$LLMS_FULL_TXT" ]] && CURRENT_LLMS_FULL=$(cat "$LLMS_FULL_TXT")

  if [[ "$NEW_LLMS" != "$CURRENT_LLMS" ]]; then
    echo "ERROR: docs/static/llms.txt is stale. Run 'make docs-gen' to update." >&2
    FAIL=true
  fi

  if [[ "$NEW_LLMS_FULL" != "$CURRENT_LLMS_FULL" ]]; then
    echo "ERROR: docs/static/llms-full.txt is stale. Run 'make docs-gen' to update." >&2
    FAIL=true
  fi

  if $FAIL; then
    exit 1
  fi

  echo "docs/static/llms.txt and llms-full.txt are up to date."
else
  printf '%s' "$NEW_LLMS" > "$LLMS_TXT"
  printf '%s' "$NEW_LLMS_FULL" > "$LLMS_FULL_TXT"
  echo "Generated ${LLMS_TXT}"
  echo "Generated ${LLMS_FULL_TXT}"
fi
