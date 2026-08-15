#!/usr/bin/env python3
"""
e2e_dag.py — DAG-based E2E test selection for provider-keycloak.

Uses only Python stdlib — no third-party dependencies required.

Usage
-----
  # Compute demo subset from changed files (used by CI)
  #   stdout: "full" | "skip" | comma-separated list of demo paths
  #   stderr: E2E_TIER=... / KEYCLOAK_VERSIONS=... plus a human-readable proof
  #           explaining exactly why the tier and each demo were selected
  python3 scripts/e2e_dag.py select --changed-files <file> [--demo-dir dev/demos]
  echo "config/openidclient/config.go" | python3 scripts/e2e_dag.py select --changed-files -

  # Decide whether the FGAPv2 suite (dev/demos/fgapv2, own cluster) must run
  #   stdout: "run" | "skip"
  python3 scripts/e2e_dag.py select-fgapv2 --changed-files <file>

  # Write the proof as markdown (for $GITHUB_STEP_SUMMARY) instead of plain text
  python3 scripts/e2e_dag.py select --changed-files - --proof-file proof.md

  # Write cluster/test/e2e-index.json (run automatically by `make generate`).
  # The index answers "which e2e test uses resource X?" and holds the demo DAG.
  python3 scripts/e2e_dag.py index [--check]

  # Fail when a managed resource has no e2e demo (run by `make e2e-cases-check`)
  python3 scripts/e2e_dag.py coverage

Tier logic (highest wins)
--------------------------
  full     : go.mod | go.sum | Makefile | internal/ (except generated
             controllers) | cmd/ | build/ | .github/workflows/ci.yml |
             non-PR event
  targeted : apis/<group>/ | config/<group>/ | package/crds/ |
             dev/demos/ | cluster/test/ | generated internal/controller code
  skip     : docs/ | scripts/ | *.md | *.png | *.jpg only

A `targeted` change never selects zero demos: if no demo covers the touched
resources, the selection broadens to the demos of their API group, and if even
that is empty it falls back to `full`. Only a `skip`-tier change set (which
cannot affect provider behaviour) runs no demos at all.

Every decision is accompanied by a proof: which changed file matched which
rule, which API groups that implies, which demos seeded the subgraph and which
dependency edge pulled in each additional demo.

Suites
------
The demo directory has one subdirectory per e2e suite, each with its own
cluster and Keycloak configuration:

  basic/, namespaced/, orgs/  regular suite  (cluster/test/cases*.txt)
  fgapv2/                     FGAPv2 suite   (cluster/test/cases-fgapv2.txt)

Selection is computed per suite: ``select`` only ever returns demos of the
regular suite, ``select-fgapv2`` only decides whether the FGAPv2 suite runs.
The suites never pull demos from each other, because the Keycloak feature
``admin-fine-grained-authz`` can only be enabled as v1 or v2, not both.
"""

import argparse
import json
import re
import sys
from collections import defaultdict, deque
from pathlib import Path

REPO_ROOT = Path(__file__).parent.parent

# Generated inverted index + demo DAG, refreshed by `make generate`.
INDEX_FILE = "cluster/test/e2e-index.json"

# Managed resources that intentionally have no e2e demo, with the reason why.
UNCOVERED_FILE = "cluster/test/uncovered-resources.txt"

# ---------------------------------------------------------------------------
# Path-classification rules
#
# Each rule is (tier, name, regex). The first matching rule wins per file, and
# the highest tier across all files wins overall. Rule names are reported in
# the selection proof so every decision is traceable back to a single pattern.
# ---------------------------------------------------------------------------

# Generated per-resource controller code: belongs to an API group, so a change
# there is targeted (not a provider-wide change), same as apis/<group>/.
GENERATED_CONTROLLER_RE = re.compile(
    r"^internal/controller/(?:cluster|namespaced)/"
    r"(?:([a-z][a-z0-9]+)/[a-z0-9]+/)?zz_[a-z_]+\.go$"
)

CLASSIFICATION_RULES: list[tuple[str, str, re.Pattern]] = [
    # skip — cannot affect provider behaviour at runtime
    ("skip", "documentation", re.compile(r"^docs/")),
    ("skip", "markdown/images", re.compile(r".*\.(md|png|jpg|svg)$")),
    # targeted — generated controller code for a single API group
    ("targeted", "generated controller", GENERATED_CONTROLLER_RE),
    # full — provider-wide code, build system or CI definition
    ("full", "go module", re.compile(r"^go\.(mod|sum)$")),
    ("full", "build system", re.compile(r"^(Makefile$|build/)")),
    ("full", "provider runtime code", re.compile(r"^(internal/|cmd/)")),
    ("full", "CI workflow", re.compile(r"^\.github/workflows/ci\.yml$")),
    ("full", "e2e environment", re.compile(r"^dev/(?!demos/)")),
    # skip — other workflows do not influence the e2e run
    ("skip", "unrelated workflow", re.compile(r"^\.github/")),
    # targeted — API surface / demos
    ("targeted", "API types", re.compile(r"^apis/")),
    ("targeted", "resource config", re.compile(r"^config/")),
    ("targeted", "CRD schema", re.compile(r"^package/crds/")),
    ("targeted", "generated example", re.compile(r"^examples-generated/")),
    ("targeted", "example manifest", re.compile(r"^examples/")),
    ("targeted", "demo manifest", re.compile(r"^dev/demos/")),
    ("targeted", "e2e harness", re.compile(r"^cluster/test/")),
    # skip — helper scripts (the selection script itself is covered here; a
    # change to it only affects which tests are chosen, not their outcome)
    ("skip", "helper script", re.compile(r"^scripts/")),
]

TIER_ORDER = {"skip": 0, "targeted": 1, "full": 2}

# Extract API group from apiVersion, e.g.
#   "openidclient.keycloak.crossplane.io/v1alpha1"   (cluster-scoped)
#   "realm.keycloak.m.crossplane.io/v1alpha1"         (namespaced)
API_GROUP_RE = re.compile(r"^([a-z][a-z0-9]+)\.keycloak(?:\.m)?\.crossplane\.io/")

# Ref-style key (e.g. clientIdRef:, realmIdRef:, idRef:) — also matches list
# entries ("- idRef:") and plural list keys ("timePoliciesRefs:").
REF_KEY_RE = re.compile(r"^\s*(?:-\s+)?\w+Refs?:\s*(?:#.*)?$")

# name: value inside a Ref block, either plain ("name: x") or as a list item
# ("- name: x"). A trailing "# ..." comment is tolerated.
REF_NAME_RE = re.compile(r"^(?:-\s+)?name:\s+[\"']?([^\"'\s#]+)[\"']?\s*(?:#.*)?$")

# Metadata name definition (a trailing "# ..." comment is tolerated)
METADATA_NAME_RE = re.compile(r"^\s{2}name:\s+[\"']?([^\"'\s#]+)[\"']?\s*(?:#.*)?$")

# Top-level apiVersion line
APIVERSION_RE = re.compile(r"^apiVersion:\s+([^\s]+)")

# Top-level kind line (a trailing "# ..." comment is tolerated)
KIND_RE = re.compile(r"^kind:\s+([A-Za-z][A-Za-z0-9]+)\s*(?:#.*)?$")

# Literal realm reference, e.g. "    realmId: dev" or "    realm: \"dev-ns\"".
# Many demos address the realm by value instead of via realmIdRef; the demo
# creating that realm is still a prerequisite.
REALM_VALUE_RE = re.compile(
    r"^\s+(?:realmId|realm):\s+[\"']?([^\"'\s#]+)[\"']?\s*(?:#.*)?$"
)

# Source path → group extraction
CONFIG_GROUP_RE = re.compile(r"^config/([a-z][a-z0-9]+)/")
APIS_GROUP_RE = re.compile(r"^apis/(?:cluster/|namespaced/)?([a-z][a-z0-9]+)/")
CRD_GROUP_RE = re.compile(r"^package/crds/([a-z][a-z0-9]+)\.")
EXAMPLES_GROUP_RE = re.compile(
    r"^examples(?:-generated)?/(?:cluster/|namespaced/)?([a-z][a-z0-9]+)/"
)
RESOURCE_APIS_RE = re.compile(
    r"^apis/(?:cluster|namespaced)/([a-z][a-z0-9]+)/v[^/]+/"
    r"zz_[a-z0-9]+_(?:types|terraformed)\.go$"
)
RESOURCE_CONTROLLER_RE = re.compile(
    r"^internal/controller/(?:cluster|namespaced)/([a-z][a-z0-9]+)/[a-z0-9]+/zz_controller\.go$"
)
RESOURCE_CRD_RE = re.compile(
    r"^package/crds/([a-z][a-z0-9]+)\.keycloak(?:\.m)?\.crossplane\.io_[a-z0-9]+\.yaml$"
)
TOP_LEVEL_RESOURCE_KIND_RE = re.compile(r"^kind:\s+([A-Z][A-Za-z0-9]+)\s*$", re.MULTILINE)
NESTED_RESOURCE_KIND_RE = re.compile(r"^\s+kind:\s+([A-Z][A-Za-z0-9]+)\s*$", re.MULTILINE)
CRD_SPEC_KIND_RE = re.compile(r"^\s{4}kind:\s+([A-Z][A-Za-z0-9]+)\s*$", re.MULTILINE)
GO_SCHEMA_KIND_RE = re.compile(r"^//\s+([A-Z][A-Za-z0-9]+)\s+is the Schema\b", re.MULTILINE)
GO_CONTROLLER_KIND_RE = re.compile(r"\*svcapitypes\.([A-Z][A-Za-z0-9]+)")


# ---------------------------------------------------------------------------
# Parsing helpers (no PyYAML — simple line-by-line)
# ---------------------------------------------------------------------------

def parse_demo_file(path: Path) -> dict:
    """
    Parse a demo YAML file (multi-document) using line-by-line regex.
    Returns:
      {
        "groups":   set of API group names used
        "kinds":    dict of (kind, group) → count
        "defines":  dict of resource name → (kind, group)
        "ref_names": set of names appearing under *Ref keys, plus literal realm
                     values (cross-resource refs)
      }
    """
    groups: set[str] = set()
    kinds: dict[tuple[str, str], int] = defaultdict(int)
    defines: dict[str, tuple[str, str]] = {}
    ref_names: set[str] = set()

    try:
        lines = path.read_text(errors="replace").splitlines()
    except Exception:
        return {"groups": groups, "kinds": kinds, "defines": defines, "ref_names": ref_names}

    current_kind = ""
    current_group = ""
    current_name = ""
    in_metadata = False
    # Indentation of the enclosing *Ref/*Refs key, or None when outside one.
    ref_indent: int | None = None

    for line in lines:
        stripped = line.strip()

        # Document separator resets state
        if stripped == "---":
            if current_kind and current_group and current_name:
                defines[current_name] = (current_kind, current_group)
            current_kind = ""
            current_group = ""
            current_name = ""
            in_metadata = False
            ref_indent = None
            continue

        if not stripped or stripped.startswith("#"):
            continue

        indent = len(line) - len(line.lstrip())

        # Inside a *Ref/*Refs block: collect every "name:" (also list items,
        # e.g. "- name: x" for *Refs lists and "- idRef:" + "name: x").
        if ref_indent is not None:
            if indent > ref_indent:
                m = REF_NAME_RE.match(stripped)
                if m:
                    ref_names.add(m.group(1))
                if not REF_KEY_RE.match(line):
                    continue
            else:
                ref_indent = None

        # Track metadata block (indentation = 0 for "metadata:", depth 1 for "  name:")
        if stripped == "metadata:":
            in_metadata = True
            continue

        # Leaving metadata block on any non-indented key
        if in_metadata and indent == 0:
            in_metadata = False

        # apiVersion
        m = APIVERSION_RE.match(line)
        if m:
            av = m.group(1)
            mg = API_GROUP_RE.match(av)
            if mg:
                current_group = mg.group(1)
                groups.add(current_group)
            continue

        # kind
        m = KIND_RE.match(line)
        if m:
            current_kind = m.group(1)
            continue

        # metadata name (exactly 2-space indent + "name:")
        if in_metadata and METADATA_NAME_RE.match(line):
            m2 = METADATA_NAME_RE.match(line)
            current_name = m2.group(1)
            continue

        # *Ref:/*Refs: key (entering a ref block)
        if REF_KEY_RE.match(line):
            ref_indent = indent
            continue

        # Literal realm value (realmId: dev) — the demo defining that realm is
        # a prerequisite just like an explicit realmIdRef would be.
        m = REALM_VALUE_RE.match(line)
        if m:
            ref_names.add(m.group(1))
            continue

    # Flush last document
    if current_kind and current_group and current_name:
        defines[current_name] = (current_kind, current_group)

    # Populate kinds from defines
    for name, (kind, group) in defines.items():
        kinds[(kind, group)] += 1

    return {
        "groups": groups,
        "kinds": dict(kinds),
        "defines": defines,
        "ref_names": ref_names,
    }


# ---------------------------------------------------------------------------
# Demo discovery
# ---------------------------------------------------------------------------

# The demo directory holds one subdirectory per e2e suite. Each suite runs in
# its own cluster with its own Keycloak configuration, so selection is computed
# per suite and the suites never pull demos from each other.
#
# basic/, namespaced/ and orgs/ share one cluster: the orgs demos only need the
# additional `organization` feature (Keycloak >= 26.6) and are version-gated by
# the Makefile, while targeted runs always use the latest Keycloak.
REGULAR_VARIANTS = ("basic", "namespaced", "orgs")
"""Demo subdirectories of the regular e2e suite (cluster/test/cases*.txt)."""

FGAPV2_VARIANTS = ("fgapv2",)
"""Demo subdirectories of the FGAPv2 e2e suite (cluster/test/cases-fgapv2.txt)."""


def discover_demos(demo_dir: Path, variants: tuple[str, ...] = REGULAR_VARIANTS) -> list[Path]:
    """Sorted list of all non-init demo YAML files across the given variants."""
    result: list[Path] = []
    for variant in variants:
        sub = demo_dir / variant
        if sub.is_dir():
            for f in sorted(sub.glob("*.yaml")):
                if f.name != "000-init.yaml":
                    result.append(f)
    return result


# ---------------------------------------------------------------------------
# DAG
# ---------------------------------------------------------------------------

class DemoGraph:
    """Full demo dependency graph + inverted index, derived purely from YAML."""

    def __init__(self, demo_dir: Path, variants: tuple[str, ...] = REGULAR_VARIANTS):
        self.demo_dir = demo_dir
        self.variants = variants
        self.demos = discover_demos(demo_dir, variants)
        self._parsed: dict[Path, dict] = {}

        # group name → demos using that group
        self.group_to_demos: dict[str, list[Path]] = defaultdict(list)

        # resource name → demos that define it
        self.name_to_demos: dict[str, list[Path]] = defaultdict(list)

        # (kind, group) → demos that define it
        self.resource_defined_by: dict[tuple[str, str], list[Path]] = defaultdict(list)

        # (kind, group) → demos that USE it
        self.resource_used_by: dict[tuple[str, str], list[Path]] = defaultdict(list)

        # demo → forward deps (prerequisites)
        self.demo_deps: dict[Path, set[Path]] = defaultdict(set)

        # demo → reverse deps (dependants)
        self.demo_rdeps: dict[Path, set[Path]] = defaultdict(set)

        self._build()

    def _build(self) -> None:
        # Phase 1: parse all demos
        for demo in self.demos:
            info = parse_demo_file(demo)
            self._parsed[demo] = info

            for g in info["groups"]:
                if demo not in self.group_to_demos[g]:
                    self.group_to_demos[g].append(demo)

            for name, (kind, group) in info["defines"].items():
                if demo not in self.name_to_demos[name]:
                    self.name_to_demos[name].append(demo)
                if demo not in self.resource_defined_by[(kind, group)]:
                    self.resource_defined_by[(kind, group)].append(demo)

            for (kind, group) in info["kinds"]:
                if demo not in self.resource_used_by[(kind, group)]:
                    self.resource_used_by[(kind, group)].append(demo)

        # Phase 2: build cross-demo edges from *Ref lookups
        # Cross-demo edges are derived from *Ref names. Only genuinely
        # infrastructure-level names are ignored here; realm names such as
        # "dev"/"dev-ns" are real demo dependencies (the realm demo must run
        # first) and must not be filtered out.
        INFRA_NAMES = {"keycloak-provider-config", "crossplane-system"}
        for demo in self.demos:
            info = self._parsed[demo]
            for ref_name in info["ref_names"]:
                if ref_name in INFRA_NAMES:
                    continue
                for dep_demo in self.name_to_demos.get(ref_name, []):
                    if dep_demo != demo and dep_demo.parent == demo.parent:
                        self.demo_deps[demo].add(dep_demo)
                        self.demo_rdeps[dep_demo].add(demo)

    def demos_for_groups(self, groups: set[str]) -> set[Path]:
        result: set[Path] = set()
        for g in groups:
            result.update(self.group_to_demos.get(g, []))
        return result

    def demos_for_resources(self, resources: set[tuple[str, str]]) -> set[Path]:
        result: set[Path] = set()
        for resource in resources:
            result.update(self.resource_defined_by.get(resource, []))
            result.update(self.resource_used_by.get(resource, []))
        return result

    def expand_subgraph(self, seed: set[Path]) -> tuple[list[Path], dict[Path, str]]:
        """Expand seed via forward + reverse dep traversal, return an ordered list.

        The returned order matches the convention of ``cluster/test/cases.txt``:
        dependents first, prerequisites last, because uptest deletes the
        examples in the order they are listed. Applying in that order is safe —
        Crossplane retries reference resolution until the prerequisite exists —
        whereas deleting a prerequisite (e.g. the realm) before its dependents
        leaves them unable to resolve their references and blocks teardown.

        Also returns a mapping demo → reason describing which edge pulled it in.
        """
        included = set(seed)
        reasons: dict[Path, str] = {}
        queue: deque[Path] = deque(seed)

        # Walk forward deps
        while queue:
            d = queue.popleft()
            for dep in self.demo_deps.get(d, set()):
                if dep not in included:
                    included.add(dep)
                    reasons[dep] = f"prerequisite of {rel(d)}"
                    queue.append(dep)

        # Walk reverse deps starting from the original seed only. This keeps
        # resource-focused selections tight: prerequisites pulled in by the
        # forward walk do not fan out into every sibling demo that happens to
        # share the same dependency.
        queue = deque(seed)
        while queue:
            d = queue.popleft()
            for rdep in self.demo_rdeps.get(d, set()):
                if rdep not in included:
                    included.add(rdep)
                    reasons[rdep] = f"depends on {rel(d)}"
                    queue.append(rdep)

        # Pull in prerequisites of anything added during the reverse-dep walk.
        # A demo selected via rdeps may itself have prerequisites (forward deps)
        # that were never visited because the first forward walk only started
        # from the seed.
        queue = deque(included)
        while queue:
            d = queue.popleft()
            for dep in self.demo_deps.get(d, set()):
                if dep not in included:
                    included.add(dep)
                    reasons[dep] = f"prerequisite of {rel(d)}"
                    queue.append(dep)

        return list(reversed(self._topo_sort(included))), reasons

    def _topo_sort(self, nodes: set[Path]) -> list[Path]:
        in_deg = {n: 0 for n in nodes}
        for n in nodes:
            for dep in self.demo_deps.get(n, set()):
                if dep in nodes:
                    in_deg[n] += 1

        ready: deque[Path] = deque(sorted(n for n, d in in_deg.items() if d == 0))
        result: list[Path] = []
        while ready:
            n = ready.popleft()
            result.append(n)
            for rdep in sorted(self.demo_rdeps.get(n, set())):
                if rdep in nodes:
                    in_deg[rdep] -= 1
                    if in_deg[rdep] == 0:
                        ready.append(rdep)
        result.extend(sorted(nodes - set(result)))
        return result

    def inverted_index(self) -> dict:
        """Return JSON-serializable inverted index: resource → {defined_in, used_by}."""
        all_keys = set(self.resource_defined_by) | set(self.resource_used_by)
        idx: dict = {}
        for kind, group in sorted(all_keys):
            label = f"{kind} ({group})"
            idx[label] = {
                "defined_in": sorted(
                    str(p.relative_to(REPO_ROOT))
                    for p in self.resource_defined_by.get((kind, group), [])
                ),
                "used_by": sorted(
                    str(p.relative_to(REPO_ROOT))
                    for p in self.resource_used_by.get((kind, group), [])
                ),
            }
        return idx

    def full_dag(self) -> dict:
        result: dict = {}
        for demo in self.demos:
            key = str(demo.relative_to(REPO_ROOT))
            result[key] = {
                "groups": sorted(self._parsed[demo]["groups"]),
                "deps": sorted(str(d.relative_to(REPO_ROOT)) for d in self.demo_deps.get(demo, set())),
                "rdeps": sorted(str(d.relative_to(REPO_ROOT)) for d in self.demo_rdeps.get(demo, set())),
            }
        return result


# ---------------------------------------------------------------------------
# Tier detection
# ---------------------------------------------------------------------------

def rel(path: Path) -> str:
    """Repo-relative string for a demo path."""
    try:
        return str(path.relative_to(REPO_ROOT))
    except ValueError:
        return str(path)


def classify_file(path: str) -> tuple[str, str]:
    """Return (tier, rule name) for a single changed path."""
    for tier, name, pattern in CLASSIFICATION_RULES:
        if pattern.match(path):
            return tier, name
    # Unknown path → safe fallback
    return "targeted", "unclassified path (safe fallback)"


def detect_tier(changed_files: list[str]) -> tuple[str, list[tuple[str, str, str]]]:
    """Return (tier, [(file, tier, rule), ...]) for the changed file set."""
    tier = "skip"
    matches: list[tuple[str, str, str]] = []
    for f in changed_files:
        f = f.strip()
        if not f:
            continue
        file_tier, rule = classify_file(f)
        matches.append((f, file_tier, rule))
        if TIER_ORDER[file_tier] > TIER_ORDER[tier]:
            tier = file_tier
    return tier, matches


def extract_groups_from_paths(
    changed_files: list[str], demo_dir: Path
) -> tuple[set[str], dict[str, set[str]]]:
    """Return (groups, group → changed paths that implied it)."""
    groups: set[str] = set()
    why: dict[str, set[str]] = defaultdict(set)
    for f in changed_files:
        f = f.strip()
        matched = False
        for pat in (CONFIG_GROUP_RE, APIS_GROUP_RE, CRD_GROUP_RE,
                    EXAMPLES_GROUP_RE, GENERATED_CONTROLLER_RE):
            m = pat.match(f)
            if m and m.group(1):
                groups.add(m.group(1))
                why[m.group(1)].add(f)
                matched = True
                break
        if matched:
            continue
        # If a demo YAML changed directly, infer its group(s) by parsing it
        if f.startswith("dev/demos/"):
            p = REPO_ROOT / f
            if p.exists() and p.suffix == ".yaml":
                info = parse_demo_file(p)
                for g in info["groups"]:
                    groups.add(g)
                    why[g].add(f)
    return groups, why


def group_scoped_config_changes(
    changed_files: list[str],
) -> tuple[set[str], dict[str, set[str]]]:
    """Return (groups, group → changed paths) for hand-written group config.

    ``config/<group>/`` holds the Upjet configuration of a whole API group
    (external names, references, lookups). A change there is not attributable
    to a single resource, so every demo of that group is exercised.
    """
    groups: set[str] = set()
    why: dict[str, set[str]] = defaultdict(set)
    for f in changed_files:
        f = f.strip()
        m = CONFIG_GROUP_RE.match(f)
        if m:
            groups.add(m.group(1))
            why[m.group(1)].add(f)
    return groups, why


def parse_resource_kinds_from_file(path: Path) -> set[str]:
    try:
        text = path.read_text(errors="replace")
    except Exception:
        return set()

    if path.suffix == ".go":
        kinds = set(GO_SCHEMA_KIND_RE.findall(text))
        if kinds:
            return kinds
        return set(GO_CONTROLLER_KIND_RE.findall(text))

    if path.suffix == ".yaml":
        if RESOURCE_CRD_RE.match(str(path.relative_to(REPO_ROOT))):
            return {
                kind for kind in CRD_SPEC_KIND_RE.findall(text)
                if kind != "CustomResourceDefinition"
            }
        top_level = set(TOP_LEVEL_RESOURCE_KIND_RE.findall(text))
        if top_level:
            return top_level
        return {
            kind for kind in NESTED_RESOURCE_KIND_RE.findall(text)
            if kind != "CustomResourceDefinition"
        }

    return set()


def extract_resources_from_paths(
    changed_files: list[str], demo_dir: Path
) -> tuple[set[tuple[str, str]], dict[tuple[str, str], set[str]]]:
    resources: set[tuple[str, str]] = set()
    why: dict[tuple[str, str], set[str]] = defaultdict(set)

    for f in changed_files:
        f = f.strip()
        if not f:
            continue

        path = REPO_ROOT / f
        group = None
        for pattern in (
            RESOURCE_APIS_RE,
            RESOURCE_CRD_RE,
            RESOURCE_CONTROLLER_RE,
        ):
            match = pattern.match(f)
            if match:
                group = match.group(1)
                break

        if f.startswith("dev/demos/") and path.exists() and path.suffix == ".yaml":
            info = parse_demo_file(path)
            for kind, group_name in info["kinds"]:
                resource = (kind, group_name)
                resources.add(resource)
                why[resource].add(f)
            continue

        if not group or not path.exists():
            continue

        for kind in parse_resource_kinds_from_file(path):
            resource = (kind, group)
            resources.add(resource)
            why[resource].add(f)

    return resources, why


# ---------------------------------------------------------------------------
# CLI commands
# ---------------------------------------------------------------------------

def _proof(lines: list[str], proof_file: str | None) -> None:
    """Emit the selection proof to stderr and, optionally, a markdown file."""
    text = "\n".join(lines)
    print(text, file=sys.stderr)
    if proof_file:
        md = ["### E2E test selection proof", "", "```text", text, "```", ""]
        Path(proof_file).write_text("\n".join(md))


def cmd_select(args) -> None:
    demo_dir = REPO_ROOT / args.demo_dir
    if args.changed_files == "-":
        changed = sys.stdin.read().splitlines()
    else:
        with open(args.changed_files) as fh:
            changed = fh.read().splitlines()
    changed = [f.strip() for f in changed if f.strip()]

    tier, matches = detect_tier(changed)

    proof: list[str] = [f"Tier: {tier}", "", "Why (changed file -> rule -> tier):"]
    # Show the decisive matches first, then the rest.
    for f, file_tier, rule in sorted(
        matches, key=lambda m: (-TIER_ORDER[m[1]], m[0])
    ):
        marker = "*" if file_tier == tier else " "
        proof.append(f"  {marker} {f} -> {rule} -> {file_tier}")
    proof.append("")
    proof.append("(* = files that determined the resulting tier)")

    if tier == "full":
        proof.append("")
        proof.append("Running ALL demos against ALL Keycloak versions.")
        print("full")
        print("E2E_TIER=full", file=sys.stderr)
        print("KEYCLOAK_VERSIONS=all", file=sys.stderr)
        _proof(proof, args.proof_file)
        return

    if tier == "skip":
        proof.append("")
        proof.append("No e2e-relevant changes — running NO demos.")
        print("skip")
        print("E2E_TIER=skip", file=sys.stderr)
        print("KEYCLOAK_VERSIONS=none", file=sys.stderr)
        _proof(proof, args.proof_file)
        return

    # targeted
    graph = DemoGraph(demo_dir)
    touched_groups, group_why = extract_groups_from_paths(changed, demo_dir)
    touched_resources, resource_why = extract_resources_from_paths(changed, demo_dir)
    seed = graph.demos_for_resources(touched_resources)
    seed_why: dict[Path, str] = {}
    for kind, group in sorted(touched_resources):
        for d in graph.resource_defined_by.get((kind, group), []):
            seed_why.setdefault(d, f"defines {kind} ({group})")
        for d in graph.resource_used_by.get((kind, group), []):
            seed_why.setdefault(d, f"uses {kind} ({group})")

    # Hand-written resource configuration (config/<group>/) is group-scoped:
    # a single edit there can change external names, references or lookups of
    # every resource of that group, so the whole group is exercised — not only
    # the resources whose generated files happen to be in the change set.
    config_groups, config_why = group_scoped_config_changes(changed)
    for g in sorted(config_groups):
        for d in graph.group_to_demos.get(g, []):
            seed.add(d)
            seed_why.setdefault(
                d, f"uses API group '{g}' ({', '.join(sorted(config_why[g]))} is group-scoped)"
            )

    if not seed and not touched_resources:
        seed = graph.demos_for_groups(touched_groups)
        for g in sorted(touched_groups):
            for d in graph.group_to_demos.get(g, []):
                seed_why.setdefault(d, f"uses API group '{g}'")

    # Include directly changed demo files of this suite. Demos of the other
    # suite are ignored here: they run in a separate cluster and are selected
    # by their own command (`select-fgapv2`).
    for f in changed:
        p = REPO_ROOT / f.strip()
        if p.suffix == ".yaml" and p.exists():
            if (
                p.parent.parent == demo_dir
                and p.parent.name in REGULAR_VARIANTS
                and p.name != "000-init.yaml"
            ):
                seed.add(p)
                seed_why[p] = "changed directly"

    proof.append("")
    proof.append("Touched resources:")
    if touched_resources:
        for kind, group in sorted(touched_resources):
            proof.append(
                f"  {kind} ({group}) (from {', '.join(sorted(resource_why[(kind, group)]))})"
            )
    else:
        proof.append("  (none)")

    proof.append("")
    proof.append("Touched API groups:")
    if touched_groups:
        for g in sorted(touched_groups):
            proof.append(f"  {g} (from {', '.join(sorted(group_why[g]))})")
    else:
        proof.append("  (none)")

    if not seed and touched_resources:
        # A touched resource that no demo covers must never silently result in
        # an empty run: broaden to the demos of its API group.
        seed = graph.demos_for_groups(touched_groups)
        for g in sorted(touched_groups):
            for d in graph.group_to_demos.get(g, []):
                seed_why.setdefault(
                    d, f"uses API group '{g}' (no demo covers the touched resources)"
                )

    if not seed:
        # Nothing in the change set could be mapped to a demo. Never run an
        # empty e2e suite for a code change — fall back to the full suite.
        proof.append("")
        proof.append(
            "No demo covers the touched resources or groups — falling back to "
            "ALL demos against ALL Keycloak versions."
        )
        print("full")
        print("E2E_TIER=full (fallback: no demos match the change set)", file=sys.stderr)
        print("KEYCLOAK_VERSIONS=all", file=sys.stderr)
        _proof(proof, args.proof_file)
        return

    ordered, dep_why = graph.expand_subgraph(seed)
    paths = [f"./{p.relative_to(REPO_ROOT)}" for p in ordered]

    proof.append("")
    proof.append(f"Selected demos ({len(paths)}, in apply/delete order):")
    for p in ordered:
        reason = seed_why.get(p) or dep_why.get(p, "pulled in by the DAG")
        proof.append(f"  {rel(p)} -- {reason}")
    proof.append("")
    proof.append("Keycloak versions: latest only")

    print(",".join(paths))
    print(f"E2E_TIER=targeted ({len(paths)} demos)", file=sys.stderr)
    print("KEYCLOAK_VERSIONS=latest", file=sys.stderr)
    _proof(proof, args.proof_file)


def cmd_select_fgapv2(args) -> None:
    """Decide whether the FGAPv2 e2e suite has to run for the changed file set.

    The FGAPv2 suite needs Keycloak's ``admin-fine-grained-authz:v2`` feature,
    which is mutually exclusive with the v1 feature the regular suite relies
    on, so it runs in its own cluster over the complete
    ``cluster/test/cases-fgapv2.txt`` list. There is no subsetting: the answer
    is simply ``run`` or ``skip``.
    """
    demo_dir = REPO_ROOT / args.demo_dir
    if args.changed_files == "-":
        changed = sys.stdin.read().splitlines()
    else:
        with open(args.changed_files) as fh:
            changed = fh.read().splitlines()
    changed = [f.strip() for f in changed if f.strip()]

    tier, _ = detect_tier(changed)
    proof: list[str] = [f"FGAPv2 suite tier: {tier}", ""]

    if tier == "skip":
        proof.append("No e2e-relevant changes — NOT running the FGAPv2 suite.")
        print("skip")
        _proof(proof, args.proof_file)
        return

    if tier == "full":
        proof.append("Provider-wide change — running the FGAPv2 suite.")
        print("run")
        _proof(proof, args.proof_file)
        return

    # targeted: run the suite only if the change can affect one of its demos.
    graph = DemoGraph(demo_dir, FGAPV2_VARIANTS)
    touched_groups, group_why = extract_groups_from_paths(changed, demo_dir)
    touched_resources, resource_why = extract_resources_from_paths(changed, demo_dir)

    reasons: list[str] = []
    for resource in sorted(touched_resources):
        if graph.demos_for_resources({resource}):
            kind, group = resource
            reasons.append(
                f"  {kind} ({group}) is used by the FGAPv2 suite "
                f"(from {', '.join(sorted(resource_why[resource]))})"
            )

    if not touched_resources:
        for g in sorted(touched_groups):
            if graph.group_to_demos.get(g):
                reasons.append(
                    f"  API group '{g}' is used by the FGAPv2 suite "
                    f"(from {', '.join(sorted(group_why[g]))})"
                )

    # Hand-written group configuration affects every resource of the group.
    config_groups, config_why = group_scoped_config_changes(changed)
    for g in sorted(config_groups):
        if graph.group_to_demos.get(g):
            reasons.append(
                f"  API group '{g}' is used by the FGAPv2 suite "
                f"({', '.join(sorted(config_why[g]))} is group-scoped)"
            )

    for f in changed:
        p = REPO_ROOT / f
        if (
            p.suffix == ".yaml"
            and p.parent.parent == demo_dir
            and p.parent.name in FGAPV2_VARIANTS
        ):
            reasons.append(f"  {f} is an FGAPv2 demo and changed directly")
        elif f == "cluster/test/cases-fgapv2.txt":
            reasons.append(f"  {f} defines the FGAPv2 case list and changed")

    if reasons:
        proof.append("Running the FGAPv2 suite because:")
        proof.extend(reasons)
        print("run")
    else:
        proof.append(
            "No FGAPv2 demo covers the changed resources, API groups or files — "
            "NOT running the FGAPv2 suite."
        )
        print("skip")
    _proof(proof, args.proof_file)


def managed_resources() -> dict[tuple[str, str], str]:
    """Return {(kind, group): crd path} for every managed resource CRD.

    Only the cluster-scoped flavour is enumerated; the namespaced flavour
    (``*.keycloak.m.crossplane.io``) always mirrors it.
    """
    result: dict[tuple[str, str], str] = {}
    crd_dir = REPO_ROOT / "package" / "crds"
    for path in sorted(crd_dir.glob("*.keycloak.crossplane.io_*.yaml")):
        group = path.name.split(".", 1)[0]
        for kind in CRD_SPEC_KIND_RE.findall(path.read_text(errors="replace")):
            if kind == "CustomResourceDefinition":
                continue
            result[(kind, group)] = rel(path)
            break
    return result


def demo_covered_resources(demo_dir: Path) -> set[tuple[str, str]]:
    """Every (kind, group) used by a demo of any suite."""
    covered: set[tuple[str, str]] = set()
    for variants in (REGULAR_VARIANTS, FGAPV2_VARIANTS):
        for demo in discover_demos(demo_dir, variants):
            covered.update(parse_demo_file(demo)["kinds"])
    return covered


def read_uncovered_exceptions() -> dict[tuple[str, str], str]:
    """Parse UNCOVERED_FILE into {(kind, group): reason}.

    Format (one per line, ``#`` starts a comment):

        Kind (group): reason why Keycloak cannot accept this resource in e2e
    """
    path = REPO_ROOT / UNCOVERED_FILE
    result: dict[tuple[str, str], str] = {}
    if not path.exists():
        return result
    for lineno, raw in enumerate(path.read_text().splitlines(), start=1):
        line = raw.split("#", 1)[0].strip()
        if not line:
            continue
        m = re.match(r"^([A-Z][A-Za-z0-9]*)\s+\(([a-z][a-z0-9]*)\):\s*(\S.*)$", line)
        if not m:
            raise SystemExit(
                f"{UNCOVERED_FILE}:{lineno}: expected 'Kind (group): reason', got {raw!r}"
            )
        result[(m.group(1), m.group(2))] = m.group(3).strip()
    return result


def cmd_coverage(args) -> None:
    """Fail when a managed resource has no e2e demo and no declared exception."""
    demo_dir = REPO_ROOT / args.demo_dir
    resources = managed_resources()
    covered = demo_covered_resources(demo_dir)
    exceptions = read_uncovered_exceptions()

    missing = sorted(r for r in resources if r not in covered and r not in exceptions)
    stale = sorted(r for r in exceptions if r in covered or r not in resources)

    if missing:
        print("e2e resource coverage check failed", file=sys.stderr)
        print("", file=sys.stderr)
        print("managed resources without an e2e demo:", file=sys.stderr)
        for kind, group in missing:
            print(f"  {kind} ({group}) — {resources[(kind, group)]}", file=sys.stderr)
        print("", file=sys.stderr)
        print(
            "Add a demo under dev/demos/<suite>/ and list it in the matching "
            f"cluster/test/cases*.txt. Only if Keycloak cannot accept the resource "
            f"in a test environment, declare it in {UNCOVERED_FILE} with a reason.",
            file=sys.stderr,
        )
    if stale:
        print("", file=sys.stderr)
        print(f"stale entries in {UNCOVERED_FILE} (now covered or removed):", file=sys.stderr)
        for kind, group in stale:
            print(f"  {kind} ({group})", file=sys.stderr)

    if missing or stale:
        sys.exit(1)

    print(
        f"e2e resource coverage check passed "
        f"({len(resources) - len(exceptions)}/{len(resources)} resources covered, "
        f"{len(exceptions)} declared untestable)"
    )


def build_index(demo_dir: Path) -> dict:
    graph = DemoGraph(demo_dir)
    return {
        "_comment": (
            "Generated by scripts/e2e_dag.py (run via `make generate`). "
            "resources: which demos define/use a kind. demos: the demo dependency DAG."
        ),
        "resources": graph.inverted_index(),
        "demos": graph.full_dag(),
    }


def cmd_index(args) -> None:
    demo_dir = REPO_ROOT / args.demo_dir
    out = REPO_ROOT / INDEX_FILE
    content = json.dumps(build_index(demo_dir), indent=2) + "\n"

    if args.check:
        if not out.exists() or out.read_text() != content:
            print(
                f"{INDEX_FILE} is stale — run `make generate` and commit the result",
                file=sys.stderr,
            )
            sys.exit(1)
        return

    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(content)


# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------

def main() -> None:
    parser = argparse.ArgumentParser(
        description="DAG-based e2e test selection for provider-keycloak"
    )
    parser.add_argument(
        "--demo-dir",
        default="dev/demos",
        help="Path to the demo directory relative to repo root (default: dev/demos)",
    )
    sub = parser.add_subparsers(dest="command", required=True)

    sel = sub.add_parser("select", help="Compute minimal demo subset from changed files")
    sel.add_argument("--changed-files", required=True,
                     metavar="FILE", help="Newline-separated changed paths or '-' for stdin")
    sel.add_argument("--proof-file", metavar="FILE", default=None,
                     help="Also write the selection proof as markdown to FILE "
                          "(e.g. $GITHUB_STEP_SUMMARY)")
    sel.set_defaults(func=cmd_select)

    fgap = sub.add_parser(
        "select-fgapv2",
        help="Decide whether the FGAPv2 e2e suite must run ('run' or 'skip')",
    )
    fgap.add_argument("--changed-files", required=True,
                      metavar="FILE", help="Newline-separated changed paths or '-' for stdin")
    fgap.add_argument("--proof-file", metavar="FILE", default=None,
                      help="Also write the selection proof as markdown to FILE "
                           "(e.g. $GITHUB_STEP_SUMMARY)")
    fgap.set_defaults(func=cmd_select_fgapv2)

    idx = sub.add_parser(
        "index", help=f"Write {INDEX_FILE} (resource index + demo DAG)"
    )
    idx.add_argument("--check", action="store_true",
                     help="Verify the committed index is up to date instead of writing it")
    idx.set_defaults(func=cmd_index)

    cov = sub.add_parser(
        "coverage",
        help="Fail when a managed resource has no e2e demo and no declared exception",
    )
    cov.set_defaults(func=cmd_coverage)

    args = parser.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
