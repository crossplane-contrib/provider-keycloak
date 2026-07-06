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
# Build the llms.txt index
# ---------------------------------------------------------------------------

generate_llms_txt() {
  local out=""
  out+="# provider-keycloak llms.txt"$'\n'
  out+="# See: https://llmstxt.org/"$'\n'
  out+=""$'\n'
  out+="> Provider Keycloak is a Crossplane provider for managing Keycloak resources as Kubernetes custom resources."$'\n'
  out+=""$'\n'
  out+="## Docs"$'\n'
  out+=""$'\n'

  # Collect pages: parse front matter title from each .md file,
  # derive URL from its path relative to content dir.
  while IFS= read -r -d '' file; do
    # Skip _index.md files (section indexes) — they're navigation, not content
    [[ "$(basename "$file")" == "_index.md" ]] && continue

    # Derive URL path from file path
    rel="${file#"${CONTENT_DIR}/"}"
    rel="${rel%.md}"
    url="${BASE_URL}/${rel}/"

    # Extract title from front matter
    title=$(awk '/^---/{found++; next} found==1 && /^title:/{sub(/^title:[[:space:]]*/,""); print; exit}' "$file")
    [[ -z "$title" ]] && continue

    # Extract description from front matter (first line after title, optional)
    desc=$(awk '/^---/{found++; next} found==1 && /^description:/{sub(/^description:[[:space:]]*/,""); print; exit}' "$file")

    if [[ -n "$desc" ]]; then
      out+="- [${title}](${url}): ${desc}"$'\n'
    else
      out+="- [${title}](${url})"$'\n'
    fi
  done < <(find "${CONTENT_DIR}/docs" -name "*.md" ! -name "_index.md" -print0 | sort -z)

  echo "$out"
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

    title=$(awk '/^---/{found++; next} found==1 && /^title:/{sub(/^title:[[:space:]]*/,""); print; exit}' "$file")
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
