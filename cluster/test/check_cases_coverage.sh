#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root"

case_files=(
  "cluster/test/cases.txt"
  "cluster/test/cases-kc-26.4.txt"
  "cluster/test/cases-kc-26.5.txt"
  "cluster/test/cases-orgs.txt"
)

for file in "${case_files[@]}"; do
  if [[ ! -f "$file" ]]; then
    echo "missing required case file: $file"
    exit 1
  fi
done

normalize_paths() {
  sed 's/#.*//' | sed '/^[[:space:]]*$/d' | sed 's#^\./##'
}

all_demos="$(mktemp)"
all_cases="$(mktemp)"
missing_cases="$(mktemp)"
stale_cases="$(mktemp)"

trap 'rm -f "$all_demos" "$all_cases" "$missing_cases" "$stale_cases"' EXIT

find dev/demos/basic dev/demos/namespaced dev/demos/orgs \
  -type f -name '*.yaml' ! -name '000-init.yaml' \
  | sed 's#^\./##' | sort -u > "$all_demos"

cat "${case_files[@]}" | normalize_paths | sort -u > "$all_cases"

comm -23 "$all_demos" "$all_cases" > "$missing_cases"
comm -13 "$all_demos" "$all_cases" > "$stale_cases"

if [[ -s "$missing_cases" ]] || [[ -s "$stale_cases" ]]; then
  echo "e2e cases coverage check failed"

  if [[ -s "$missing_cases" ]]; then
    echo
    echo "demos missing from case files:"
    cat "$missing_cases"
  fi

  if [[ -s "$stale_cases" ]]; then
    echo
    echo "case entries without a matching demo file:"
    cat "$stale_cases"
  fi

  exit 1
fi

echo "e2e cases coverage check passed"
