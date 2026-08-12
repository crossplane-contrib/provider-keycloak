#!/usr/bin/env python3
"""
Check for new releases of the Keycloak Terraform provider
(github.com/keycloak/terraform-provider-keycloak) and act on it.

This is the automation half of a manual "is there a new upstream provider
release?" check:

  --mode issue (default)
      Compares the latest upstream release against the version pinned in
      `Makefile` (`TERRAFORM_PROVIDER_VERSION`). If a newer version is found
      and no open issue already tracks it, files a tracking issue with a
      checklist of everything a version bump requires (Makefile, go.mod,
      schema/CRD regeneration, breaking-change review, etc).

  --mode bump
      Performs the actual version bump in the *current* working tree:
        - updates `TERRAFORM_PROVIDER_VERSION` in `Makefile`
        - updates the `terraform-provider-keycloak` pseudo-version in `go.mod`
          via `go get` + `go mod tidy`
        - regenerates `config/schema.json` and the CRDs via `make generate`
        - diffs the old and new `config/schema.json` (via version_diff.py) to
          surface potential breaking changes (schema version bumps, removed
          resources/attributes)
      It does not commit, push, or open a PR itself -- that is left to the
      calling workflow so it can control git identity/branching, but this
      script prints a ready-to-use PR body summarizing the changes.

Usage:
  check_provider_release.py --repo <owner/repo> --mode issue [--dry-run]
  check_provider_release.py --repo <owner/repo> --mode bump [--dry-run] \\
      [--pr-body-file PATH]

Requires the `gh` CLI to be installed and authenticated (GH_TOKEN/GITHUB_TOKEN)
for --mode issue (and for looking up the release in --mode bump), unless
--dry-run is used together with --skip-issue-lookup.

Exit codes:
  0  Ran successfully (regardless of whether a new version was found)
  1  (--mode bump only) a new version was found and processed
  2  Error (missing files, invalid JSON/semver, gh/go/make failure, etc.)
"""

import argparse
import json
import re
import subprocess
import sys
from pathlib import Path

UPSTREAM_REPO = "keycloak/terraform-provider-keycloak"
GO_MODULE = "github.com/keycloak/terraform-provider-keycloak"
MAKEFILE = "Makefile"
SCHEMA_JSON = "config/schema.json"
GENERATED_LST = "config/generated.lst"

MAKEFILE_VERSION_RE = re.compile(
    r"^(export TERRAFORM_PROVIDER_VERSION\s*\?=\s*)(\S+)\s*$", re.MULTILINE
)


def run(args, check=True, cwd=None):
    """Run a subprocess command and return its stdout."""
    result = subprocess.run(
        args, capture_output=True, text=True, check=False, cwd=cwd
    )
    if check and result.returncode != 0:
        raise RuntimeError(
            f"{' '.join(args)} failed (exit {result.returncode}): {result.stderr.strip()}"
        )
    return result.stdout


def run_gh(args, check=True):
    return run(["gh", *args], check=check)


def parse_version(version):
    """Parse a version string like 'v5.9.0' or '5.9.0' into a comparable tuple."""
    v = version.strip()
    if v.startswith("v"):
        v = v[1:]
    # Drop any pre-release/build metadata suffix (e.g. "5.9.0-rc1").
    v = re.split(r"[-+]", v, maxsplit=1)[0]
    parts = v.split(".")
    try:
        return tuple(int(p) for p in parts)
    except ValueError as e:
        raise ValueError(f"cannot parse version {version!r}: {e}") from e


def get_current_version(makefile_path=MAKEFILE):
    """Read the currently pinned TERRAFORM_PROVIDER_VERSION from the Makefile."""
    text = Path(makefile_path).read_text()
    m = MAKEFILE_VERSION_RE.search(text)
    if not m:
        raise RuntimeError(
            f"could not find 'export TERRAFORM_PROVIDER_VERSION ?= ...' in {makefile_path}"
        )
    return m.group(2)


def set_makefile_version(new_version, makefile_path=MAKEFILE):
    """Rewrite TERRAFORM_PROVIDER_VERSION in the Makefile to new_version."""
    text = Path(makefile_path).read_text()
    new_text, count = MAKEFILE_VERSION_RE.subn(rf"\g<1>{new_version}", text)
    if count == 0:
        raise RuntimeError(f"failed to update TERRAFORM_PROVIDER_VERSION in {makefile_path}")
    Path(makefile_path).write_text(new_text)


def get_latest_release(repo=UPSTREAM_REPO):
    """Return (tag_name, html_url, body) for the latest release of repo."""
    out = run_gh(
        [
            "api",
            f"repos/{repo}/releases/latest",
            "--jq",
            "{tag_name: .tag_name, html_url: .html_url, body: .body}",
        ]
    )
    data = json.loads(out)
    return data["tag_name"], data["html_url"], data.get("body") or ""


def fetch_open_issues(repo):
    """Fetch all open issues (number, title, body, url) for the given repo."""
    out = run_gh(
        [
            "issue",
            "list",
            "--repo",
            repo,
            "--state",
            "open",
            "--limit",
            "500",
            "--json",
            "number,title,body,url",
        ]
    )
    return json.loads(out)


def version_mentioned(version, text):
    """Return True if version (optionally prefixed with 'v') is mentioned as a
    whole token in text, e.g. matches 'v5.9.0' or '5.9.0' but not '5.9.01'."""
    if not text:
        return False
    pattern = r"(?<![\w.-])v?" + re.escape(version) + r"(?![\w.-])"
    return re.search(pattern, text) is not None


def find_existing_issue(version, issues):
    for issue in issues:
        if version_mentioned(version, issue.get("title", "")) or version_mentioned(
            version, issue.get("body", "")
        ):
            return issue
    return None


def build_issue_body(current_version, new_version, release_url, release_notes):
    notes_section = ""
    if release_notes.strip():
        trimmed = release_notes.strip()
        if len(trimmed) > 3000:
            trimmed = trimmed[:3000] + "\n\n_(truncated)_"
        notes_section = f"\n<details>\n<summary>Upstream release notes</summary>\n\n{trimmed}\n\n</details>\n"

    return f"""A new release of the [Keycloak Terraform provider]({release_url}) is
available: `v{new_version}` (currently pinned: `v{current_version}`).
{notes_section}
## Steps to bump the provider version

- [ ] Update `export TERRAFORM_PROVIDER_VERSION` in `Makefile` to `{new_version}`.
- [ ] Update the `github.com/keycloak/terraform-provider-keycloak` pseudo-version
      in `go.mod` (e.g. `go get {GO_MODULE}@v{new_version} && go mod tidy`).
- [ ] Regenerate the provider schema and CRDs with `make generate`.
- [ ] Review `config/schema.json` for breaking changes: run
      `make schema-diff OLD_PROVIDER_VERSION={current_version}` (or
      `scripts/version_diff.py {GENERATED_LST} <old-schema.json> {SCHEMA_JSON}`)
      and check for removed resources/attributes or schema version bumps on
      already-generated resources, which may require a migration.
- [ ] Check `make schema-diff` output / the `schema-diff-issues` workflow for
      any newly available Terraform resources that could be exposed as
      managed resources.
- [ ] Update examples/docs if any generated field names or types changed.
- [ ] Run `make test` and, if applicable, `make e2e`.

_This issue was filed automatically by the `provider-release-check` workflow._
"""


def cmd_issue(args):
    try:
        current_version = get_current_version()
    except RuntimeError as e:
        print(f"Error: {e}", file=sys.stderr)
        return 2

    try:
        latest_tag, release_url, release_notes = get_latest_release()
    except (RuntimeError, json.JSONDecodeError) as e:
        print(f"Error fetching latest release: {e}", file=sys.stderr)
        return 2

    try:
        latest_version = latest_tag[1:] if latest_tag.startswith("v") else latest_tag
        is_newer = parse_version(latest_version) > parse_version(current_version)
    except ValueError as e:
        print(f"Error: {e}", file=sys.stderr)
        return 2

    print(f"Currently pinned version: {current_version}")
    print(f"Latest upstream release:  {latest_version} ({release_url})")

    if not is_newer:
        print("No newer version available.")
        return 0

    try:
        issues = fetch_open_issues(args.repo)
    except RuntimeError as e:
        print(f"Error fetching open issues: {e}", file=sys.stderr)
        return 2

    existing = find_existing_issue(latest_version, issues)
    if existing:
        print(f"[skip] v{latest_version} already tracked by #{existing['number']} ({existing['url']})")
        return 0

    title = f"chore: bump terraform-provider-keycloak to v{latest_version}"
    body = build_issue_body(current_version, latest_version, release_url, release_notes)

    if args.dry_run:
        print(f'[dry-run] would create issue: "{title}"')
        return 0

    out = run_gh(
        [
            "issue",
            "create",
            "--repo",
            args.repo,
            "--title",
            title,
            "--body",
            body,
            "--label",
            args.label,
        ]
    )
    print(f"[created] v{latest_version}: {out.strip()}")
    return 0


def cmd_bump(args):
    try:
        current_version = get_current_version()
    except RuntimeError as e:
        print(f"Error: {e}", file=sys.stderr)
        return 2

    try:
        latest_tag, release_url, _ = get_latest_release()
    except (RuntimeError, json.JSONDecodeError) as e:
        print(f"Error fetching latest release: {e}", file=sys.stderr)
        return 2

    latest_version = latest_tag[1:] if latest_tag.startswith("v") else latest_tag
    try:
        is_newer = parse_version(latest_version) > parse_version(current_version)
    except ValueError as e:
        print(f"Error: {e}", file=sys.stderr)
        return 2

    print(f"Currently pinned version: {current_version}")
    print(f"Latest upstream release:  {latest_version} ({release_url})")

    if not is_newer:
        print("No newer version available; nothing to bump.")
        return 0

    if args.dry_run:
        print(f"[dry-run] would bump {current_version} -> {latest_version}")
        return 1

    old_schema = None
    schema_path = Path(SCHEMA_JSON)
    if schema_path.exists():
        old_schema = schema_path.read_text()

    print(f"Bumping TERRAFORM_PROVIDER_VERSION: {current_version} -> {latest_version}")
    set_makefile_version(latest_version)

    print(f"Updating go.mod dependency on {GO_MODULE}@v{latest_version}...")
    run(["go", "get", f"{GO_MODULE}@v{latest_version}"])
    run(["go", "mod", "tidy"])

    print("Regenerating provider schema and CRDs (make generate)...")
    run(["make", "generate"])

    breaking_summary = "No previous schema.json snapshot available to diff against."
    if old_schema is not None:
        import tempfile

        with tempfile.NamedTemporaryFile("w", suffix=".json", delete=False) as f:
            f.write(old_schema)
            old_schema_path = f.name
        diff_result = subprocess.run(
            [sys.executable, "scripts/version_diff.py", GENERATED_LST, old_schema_path, SCHEMA_JSON],
            capture_output=True,
            text=True,
        )
        breaking_summary = diff_result.stdout.strip() or "No changes detected."

    pr_body = f"""## Summary

Bumps the Keycloak Terraform provider from `v{current_version}` to
`v{latest_version}` ([release notes]({release_url})).

- Updated `TERRAFORM_PROVIDER_VERSION` in `Makefile`.
- Updated the `{GO_MODULE}` pseudo-version in `go.mod`/`go.sum`.
- Regenerated `config/schema.json` and CRDs via `make generate`.

## Schema diff (potential breaking changes)

```
{breaking_summary}
```

Please review the diff above for removed resources/attributes or schema
version bumps on already-generated resources -- these may require a manual
migration before merging.

_This PR was generated automatically by the `provider-release-check` workflow._
"""

    if args.pr_body_file:
        Path(args.pr_body_file).write_text(pr_body)
        print(f"Wrote PR body to {args.pr_body_file}")
    else:
        print(pr_body)

    print(f"Bump complete: v{current_version} -> v{latest_version}")
    return 1


def main():
    parser = argparse.ArgumentParser(
        description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter
    )
    parser.add_argument("--repo", required=True, help="owner/repo to search/create issues in")
    parser.add_argument("--mode", choices=["issue", "bump"], default="issue")
    parser.add_argument("--dry-run", action="store_true", help="Only print what would happen")
    parser.add_argument("--label", default="dependencies", help="Label to apply to created issues")
    parser.add_argument(
        "--pr-body-file",
        help="(--mode bump) write the generated PR description to this file instead of stdout",
    )
    args = parser.parse_args()

    if args.mode == "issue":
        sys.exit(cmd_issue(args))
    else:
        sys.exit(cmd_bump(args))


if __name__ == "__main__":
    main()
