#!/usr/bin/env python3
"""
File GitHub issues for gaps in the provider configuration, in two passes:

1. Coverage: Terraform provider resources present in config/schema.json but
   not yet generated (per config/generated.lst), i.e. not exposed as managed
   resources at all.
2. Quality: actionable findings reported by `make config-audit`, i.e.
   resources that *are* exposed but whose reference wiring is inconsistent
   with their siblings (drift) or points at only one member of a type family
   that is elsewhere modelled with config/multitypes (missing-multitype).

Both passes file one issue per finding, skipping anything an open issue
already tracks.

This is the automation half of `make schema-diff` and `make config-audit`:
instead of a human reading the output and manually opening issues, this script
does it (or, in --dry-run mode, just reports what it *would* do).

Usage:
  schema_diff_issues.py --repo <owner/repo> [--generated-lst PATH] [--schema PATH]
                         [--dry-run] [--label LABEL]
                         [--skip-coverage] [--skip-config-audit]
                         [--config-audit-json PATH] [--audit-detectors LIST]

Requires the `gh` CLI to be installed and authenticated (GH_TOKEN/GITHUB_TOKEN),
unless --dry-run is combined with --skip-issue-lookup, in which case no GitHub
API calls are made at all.
The config-audit pass shells out to `go run ./cmd/configaudit --format=json`
unless --config-audit-json points at a pre-rendered report ("-" reads stdin).

Exit codes:
  0  Ran successfully (regardless of whether any resources/issues were found)
  2  Error (missing files, invalid JSON, gh CLI failure, etc.)
"""

import argparse
import json
import re
import shlex
import subprocess
import sys

# How the config-audit report is produced when it is not passed in.
DEFAULT_CONFIG_AUDIT_COMMAND = ["go", "run", "./cmd/configaudit", "--format=json"]

# Detectors filed as issues by default. "unclassified" is left out: roughly a
# third of reference-shaped attributes are correct omissions (provider_id,
# tenant_id, entity_id, ...), so it would file mostly noise until there is a
# place to record the reason - see design/0001.
DEFAULT_AUDIT_DETECTORS = ["drift", "missing-multitype"]

# Marker embedded in filed issues so re-runs recognise their own issues even
# after the title or the wording of the finding changes.
AUDIT_MARKER = "config-audit-key"


def load_generated_list(path):
    """Load the list of generated resource names from a file."""
    resources = set()
    with open(path) as f:
        for line in f:
            line = line.strip()
            if line and not line.startswith("#"):
                resources.add(line)
    return resources


def load_schema_resource_names(path):
    """Load and return the resource type names from a provider schema JSON file."""
    with open(path) as f:
        data = json.load(f)
    provider_schemas = data.get("provider_schemas", {})
    for _, provider_data in provider_schemas.items():
        return set(provider_data.get("resource_schemas", {}).keys())
    return set()


def run_gh(args, check=True):
    """Run a `gh` CLI command and return its stdout."""
    result = subprocess.run(
        ["gh", *args], capture_output=True, text=True, check=False
    )
    if check and result.returncode != 0:
        raise RuntimeError(
            f"gh {' '.join(args)} failed (exit {result.returncode}): {result.stderr.strip()}"
        )
    return result.stdout


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


def resource_mentioned(resource, text):
    """Return True if resource is mentioned as a whole token in text (not as a substring of a longer resource name)."""
    if not text:
        return False
    pattern = r"(?<![\w-])" + re.escape(resource) + r"(?![\w-])"
    return re.search(pattern, text) is not None


def find_existing_issue(resource, issues):
    """Return the first open issue that already mentions this resource, if any."""
    for issue in issues:
        if resource_mentioned(resource, issue.get("title", "")) or resource_mentioned(
            resource, issue.get("body", "")
        ):
            return issue
    return None


def find_issue_by_marker(marker, issues):
    """Return the first open issue whose body carries this exact marker, if any."""
    for issue in issues:
        if marker in (issue.get("body") or ""):
            return issue
    return None


def build_issue_body(resource):
    return f"""The Terraform provider resource `{resource}` is present in
`config/schema.json` but missing from `config/generated.lst`, so it is not
yet exposed as a Crossplane managed resource.

## Steps to add it

1. Add an entry for `{resource}` to `config/external_name.go`.
2. Create/update `config/<group>/config.go` with any cross-resource
   references (`r.References[...]`) and, if the resource can already exist in
   Keycloak before being imported, wire it to
   `lookup.BuildIdentifyingPropertiesLookup` to avoid `409 Conflict` errors on
   create.
3. Run `make generate` to regenerate the Go types and CRDs.
4. Add a hand-authored example manifest under `examples/<group>/`.
5. Optionally add a docs page under `docs/content/docs/using/resources/`.
6. Optionally add the resource to `cluster/test/cases.txt` and an e2e
   manifest so it is covered by `make e2e`.

See the "Adding a New Resource" section in `SKILL.md`/`AGENTS.md` for the full
playbook.

_This issue was filed automatically by the `schema-diff-issues` workflow._
"""


def create_issue(repo, title, body, label, dry_run):
    """Create an issue, or report what would be created. Returns True on success."""
    if dry_run:
        print(f'[dry-run] would create issue: "{title}"')
        return True
    try:
        out = run_gh(
            [
                "issue",
                "create",
                "--repo",
                repo,
                "--title",
                title,
                "--body",
                body,
                "--label",
                label,
            ]
        )
    except RuntimeError as e:
        print(f"[error] failed to create issue \"{title}\": {e}", file=sys.stderr)
        return False
    print(f"[created] {title}: {out.strip()}")
    return True


def coverage_pass(args, issues):
    """File one issue per Terraform resource that is not exposed as a managed resource."""
    try:
        generated = load_generated_list(args.generated_lst)
    except FileNotFoundError:
        print(f"Error: generated list file not found: {args.generated_lst}", file=sys.stderr)
        sys.exit(2)

    try:
        schema_resources = load_schema_resource_names(args.schema)
    except (FileNotFoundError, json.JSONDecodeError) as e:
        print(f"Error loading schema: {e}", file=sys.stderr)
        sys.exit(2)

    missing = sorted(schema_resources - generated)
    if not missing:
        print("No missing resources: config/generated.lst already covers all resources in config/schema.json.")
        return 0

    print(f"Found {len(missing)} resource(s) missing from {args.generated_lst}:")
    for r in missing:
        print(f"  - {r}")
    print()

    # Issues filed by the config-audit pass quote resource names in their body,
    # so they must not make a genuinely missing resource look tracked.
    issues = [i for i in issues if AUDIT_MARKER not in (i.get("body") or "")]

    created = 0
    for resource in missing:
        existing = find_existing_issue(resource, issues)
        if existing:
            print(f"[skip] {resource}: already tracked by #{existing['number']} ({existing['url']})")
            continue
        if create_issue(
            args.repo,
            f"feat: expose {resource} managed resource",
            build_issue_body(resource),
            args.label,
            args.dry_run,
        ):
            created += 1

    print()
    print(f"Coverage: {created} issue(s) {'would be ' if args.dry_run else ''}created, "
          f"{len(missing) - created} resource(s) already tracked or skipped.")
    return created


def load_config_audit_report(path, command):
    """Load a config-audit JSON report, either from a file/stdin or by running the audit."""
    if path == "-":
        return json.load(sys.stdin)
    if path:
        with open(path) as f:
            return json.load(f)
    result = subprocess.run(command, capture_output=True, text=True, check=False)
    if result.returncode != 0:
        raise RuntimeError(
            f"{' '.join(command)} failed (exit {result.returncode}): {result.stderr.strip()}"
        )
    return json.loads(result.stdout)


def finding_key(finding):
    """Mirror configaudit.Finding.Key() so issue dedup is stable across runs."""
    detector = finding.get("detector", "")
    attribute = finding.get("attribute", "")
    if finding.get("resource"):
        return f"{detector}/{finding['resource']}/{attribute}"
    if finding.get("shape"):
        return f"{detector}/{attribute}/{finding['shape']}"
    return f"{detector}/{attribute}"


def actionable_findings(report, detectors):
    """Return the findings that still need a decision, for the selected detectors.

    Mirrors configaudit.Report.Actionable: satisfied findings are already
    handled, protocol-specific missing-multitype candidates are the ones
    config-audit itself does not count as actionable (an OpenID-only resource
    legitimately wires only the OpenID member of a family), and a multi-type
    classification names no resource to change - the resource missing a family
    member is filed from the per-resource missing-multitype finding instead.
    """
    out = []
    for finding in report.get("findings", []):
        if finding.get("detector") not in detectors:
            continue
        if finding.get("status") != "open" or finding.get("protocolSpecific"):
            continue
        if finding.get("class") == "multitype":
            continue
        out.append(finding)
    out.sort(key=finding_key)
    return out


def describe_shape(shape):
    """Render a config-audit shape ("TypeString/optional") for a human ("optional string").

    Attributes sharing a name but not a schema shape are separate findings, so
    the shape has to appear in the title to keep issues distinguishable.
    """
    if not shape or "/" not in shape:
        return shape
    type_key, optionality = shape.rsplit("/", 1)
    return f"{optionality} {type_key.removeprefix('Type').lower()}"


def audit_issue_title(finding):
    attribute = finding.get("attribute", "")
    cls = finding.get("class", "")
    shape = describe_shape(finding.get("shape", ""))
    qualifier = f" ({shape})" if shape else ""
    if cls == "gap":
        unwired = finding.get("unwiredOn") or []
        return (
            f"fix(config): decide whether `{attribute}`{qualifier} should be wired on "
            f"{len(unwired)} resource(s) that leave it unwired"
        )
    if cls == "multitype":
        return (
            f"fix(config): `{attribute}`{qualifier} resolves to several target types "
            "without a multi-type field"
        )
    if cls == "missing-multitype":
        return (
            f"fix(config): decide whether `{finding.get('resource', '')}.{attribute}` "
            "needs a multi-type reference"
        )
    return f"fix(config): review `{attribute}` ({cls})"


def audit_issue_body(finding):
    key = finding_key(finding)
    lines = [
        f"`make config-audit` reports this finding on the current tree:",
        "",
        f"> {finding.get('detail', '')}",
        "",
    ]

    wired_to = finding.get("wiredTo") or {}
    if wired_to:
        lines.append("**Wired to**")
        lines.append("")
        for target in sorted(wired_to):
            resources = ", ".join(f"`{r}`" for r in wired_to[target])
            lines.append(f"- `{target}`: {resources}")
        lines.append("")
    if finding.get("unwiredOn"):
        lines.append("**Unwired on**")
        lines.append("")
        for resource in finding["unwiredOn"]:
            lines.append(f"- `{resource}`")
        lines.append("")
    if finding.get("family"):
        family = ", ".join(f"`{m}`" for m in finding["family"])
        missing = ", ".join(f"`{m}`" for m in finding.get("missing", []))
        lines.append(f"**Type family**: {family}")
        lines.append("")
        lines.append(f"**Not wired on this resource**: {missing}")
        lines.append("")

    lines.extend(
        [
            "## What to decide",
            "",
            "This is a *question*, not a confirmed bug: the audit cross-checks the",
            "Terraform schema against the configuration the generator builds, and it",
            "cannot know Keycloak's semantics. Either fix the wiring in",
            "`config/<group>/config.go` and run `make generate`, or record why the",
            "current state is correct.",
            "",
            "Before changing anything, check the pinned upstream provider — its docs",
            "are checked out by `make pull-docs` into `.work/keycloak/keycloak/` and its",
            "source is at `go list -m -f '{{.Dir}}' github.com/keycloak/terraform-provider-keycloak`,",
            "both at the version this repository builds against.",
            "",
            "## Reproduce",
            "",
            "```bash",
            "make config-audit",
            "make config-audit CONFIG_AUDIT_ARGS='--format=json'",
            "```",
            "",
            "_This issue was filed automatically by the `schema-diff-issues` workflow._",
            f"<!-- {AUDIT_MARKER}: {key} -->",
        ]
    )
    return "\n".join(lines) + "\n"


def config_audit_pass(args, issues):
    """File one issue per actionable config-audit finding."""
    try:
        report = load_config_audit_report(args.config_audit_json, args.config_audit_command)
    except (FileNotFoundError, json.JSONDecodeError, RuntimeError) as e:
        print(f"Error loading config-audit report: {e}", file=sys.stderr)
        sys.exit(2)

    detectors = [d.strip() for d in args.audit_detectors.split(",") if d.strip()]
    findings = actionable_findings(report, detectors)
    if not findings:
        print(f"No actionable config-audit findings for detector(s): {', '.join(detectors)}.")
        return 0

    print(f"Found {len(findings)} actionable config-audit finding(s):")
    for finding in findings:
        print(f"  - {finding_key(finding)}: {finding.get('detail', '')}")
    print()

    created = 0
    for finding in findings:
        key = finding_key(finding)
        marker = f"<!-- {AUDIT_MARKER}: {key} -->"
        existing = find_issue_by_marker(marker, issues)
        if existing:
            print(f"[skip] {key}: already tracked by #{existing['number']} ({existing['url']})")
            continue
        if create_issue(
            args.repo,
            audit_issue_title(finding),
            audit_issue_body(finding),
            args.label,
            args.dry_run,
        ):
            created += 1

    print()
    print(f"Config audit: {created} issue(s) {'would be ' if args.dry_run else ''}created, "
          f"{len(findings) - created} finding(s) already tracked or skipped.")
    return created


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo", required=True, help="owner/repo to search/create issues in")
    parser.add_argument("--generated-lst", default="config/generated.lst")
    parser.add_argument("--schema", default="config/schema.json")
    parser.add_argument("--dry-run", action="store_true", help="Only print what would happen; do not create issues")
    parser.add_argument("--label", default="enhancement", help="Label to apply to created issues")
    parser.add_argument(
        "--skip-issue-lookup",
        action="store_true",
        help="Do not query GitHub for existing issues; only valid together with --dry-run",
    )
    parser.add_argument("--skip-coverage", action="store_true", help="Skip the missing-resource pass")
    parser.add_argument("--skip-config-audit", action="store_true", help="Skip the config-audit pass")
    parser.add_argument(
        "--config-audit-json",
        default="",
        help="Path to a config-audit JSON report ('-' for stdin); by default the audit is run",
    )
    parser.add_argument(
        "--config-audit-command",
        default=" ".join(DEFAULT_CONFIG_AUDIT_COMMAND),
        help="Command used to produce the config-audit JSON report",
    )
    parser.add_argument(
        "--audit-detectors",
        default=",".join(DEFAULT_AUDIT_DETECTORS),
        help="Comma-separated config-audit detectors to file issues for. 'unclassified' is "
             "excluded by default: roughly a third of those attributes are correct omissions",
    )
    args = parser.parse_args()
    args.config_audit_command = shlex.split(args.config_audit_command)

    if args.skip_coverage and args.skip_config_audit:
        print("Nothing to do: both passes are skipped.")
        return

    if args.skip_issue_lookup:
        if not args.dry_run:
            print("Error: --skip-issue-lookup requires --dry-run", file=sys.stderr)
            sys.exit(2)
        issues = []
    else:
        try:
            issues = fetch_open_issues(args.repo)
        except RuntimeError as e:
            print(f"Error fetching open issues: {e}", file=sys.stderr)
            sys.exit(2)

    if not args.skip_coverage:
        print("== coverage: resources not exposed as managed resources ==")
        coverage_pass(args, issues)
        print()

    if not args.skip_config_audit:
        print("== config audit: reference wiring that needs a decision ==")
        config_audit_pass(args, issues)


if __name__ == "__main__":
    main()
