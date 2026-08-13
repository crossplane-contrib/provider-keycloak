#!/usr/bin/env python3
"""
e2e_dag.py — DAG-based E2E test selection for provider-keycloak.

Uses only Python stdlib — no third-party dependencies required.

Usage
-----
  # Compute demo subset from changed files
  #   stdout: "full" | "skip" | comma-separated list of demo paths
  #   stderr: E2E_TIER=... and KEYCLOAK_VERSIONS=... lines for CI
  python3 scripts/e2e_dag.py select --changed-files <file> [--demo-dir dev/demos]
  echo "config/openidclient/config.go" | python3 scripts/e2e_dag.py select --changed-files -

  # Print inverted index: which e2e test uses resource X?
  python3 scripts/e2e_dag.py index [--demo-dir dev/demos]

  # Query which tests use a specific resource kind
  python3 scripts/e2e_dag.py query ClientTimePolicy [--group openidclient]

  # Print the full DAG as JSON
  python3 scripts/e2e_dag.py dag [--demo-dir dev/demos]

Tier logic (highest wins)
--------------------------
  full     : go.mod | go.sum | Makefile | internal/ | cmd/ | build/ |
             .github/workflows/ci.yml | non-PR event
  targeted : apis/<group>/ | config/<group>/ | package/crds/ |
             dev/demos/ | cluster/test/
  skip     : docs/ | scripts/ | *.md | *.png | *.jpg only
"""

import argparse
import json
import re
import sys
from collections import defaultdict, deque
from pathlib import Path

REPO_ROOT = Path(__file__).parent.parent

# ---------------------------------------------------------------------------
# Path-classification patterns
# ---------------------------------------------------------------------------

FULL_PATTERNS = re.compile(
    r"^(go\.mod$|go\.sum$|Makefile$|internal/|cmd/|build/|"
    r"\.github/workflows/ci\.yml$)"
)

TARGETED_PATTERNS = re.compile(
    r"^(apis/|config/|package/crds/|dev/demos/|cluster/test/)"
)

SKIP_PATTERNS = re.compile(
    r".*\.(md|png|jpg|svg)$|^(docs/|scripts/|AGENTS\.md|SKILL\.md|README\.md)"
)

# Extract API group from apiVersion, e.g.
#   "openidclient.keycloak.crossplane.io/v1alpha1"   (cluster-scoped)
#   "realm.keycloak.m.crossplane.io/v1alpha1"         (namespaced)
API_GROUP_RE = re.compile(r"^([a-z][a-z0-9]+)\.keycloak(?:\.m)?\.crossplane\.io/")

# Ref-style key (e.g. clientIdRef:, realmIdRef:, idRef:)
REF_LINE_RE = re.compile(r"^\s+\w+Ref:\s*$")

# name: value under a Ref block
NAME_VALUE_RE = re.compile(r"^\s+name:\s+[\"']?([^\"'\s]+)[\"']?\s*$")

# Metadata name definition
METADATA_NAME_RE = re.compile(r"^\s{2}name:\s+[\"']?([^\"'\s]+)[\"']?\s*$")

# apiVersion line
APIVERSION_RE = re.compile(r"^\s*apiVersion:\s+([^\s]+)")

# kind line
KIND_RE = re.compile(r"^\s*kind:\s+([A-Za-z][A-Za-z0-9]+)\s*$")

# Source path → group extraction
CONFIG_GROUP_RE = re.compile(r"^config/([a-z][a-z0-9]+)/")
APIS_GROUP_RE = re.compile(r"^apis/(?:cluster/|namespaced/)?([a-z][a-z0-9]+)/")
CRD_GROUP_RE = re.compile(r"^package/crds/([a-z][a-z0-9]+)\.")


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
        "ref_names": set of names appearing under *Ref keys (cross-resource refs)
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
    in_ref = False

    for line in lines:
        # Document separator resets state
        if line.strip() == "---":
            if current_kind and current_group and current_name:
                defines[current_name] = (current_kind, current_group)
            current_kind = ""
            current_group = ""
            current_name = ""
            in_metadata = False
            in_ref = False
            continue

        # Track metadata block (indentation = 0 for "metadata:", depth 1 for "  name:")
        if line.strip() == "metadata:":
            in_metadata = True
            in_ref = False
            continue

        # Leaving metadata block on any non-indented key
        if in_metadata and line and not line[0].isspace():
            in_metadata = False

        # apiVersion
        m = APIVERSION_RE.match(line)
        if m:
            av = m.group(1)
            mg = API_GROUP_RE.match(av)
            if mg:
                current_group = mg.group(1)
                groups.add(current_group)
            in_ref = False
            continue

        # kind
        m = KIND_RE.match(line)
        if m:
            current_kind = m.group(1)
            in_ref = False
            continue

        # metadata name (exactly 2-space indent + "name:")
        if in_metadata and METADATA_NAME_RE.match(line):
            m2 = METADATA_NAME_RE.match(line)
            current_name = m2.group(1)
            continue

        # *Ref: key (entering a ref block)
        if REF_LINE_RE.match(line):
            in_ref = True
            continue

        # name: under a Ref block (deeper indent = 6+ spaces typically)
        if in_ref and re.match(r"^\s{6,}name:\s+", line):
            m2 = re.match(r"^\s+name:\s+[\"']?([^\"'\s#]+)[\"']?", line)
            if m2:
                ref_names.add(m2.group(1))
            in_ref = False
            continue

        # Any other non-empty line clears ref state if indentation drops
        if in_ref and line.strip() and not re.match(r"^\s{6,}", line):
            in_ref = False

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

def discover_demos(demo_dir: Path) -> list[Path]:
    """Sorted list of all non-init demo YAML files across basic/ and namespaced/."""
    result: list[Path] = []
    for variant in ("basic", "namespaced"):
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

    def __init__(self, demo_dir: Path):
        self.demo_dir = demo_dir
        self.demos = discover_demos(demo_dir)
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
        INFRA_NAMES = {"keycloak-provider-config", "dev", "crossplane-system"}
        for demo in self.demos:
            info = self._parsed[demo]
            for ref_name in info["ref_names"]:
                if ref_name in INFRA_NAMES:
                    continue
                for dep_demo in self.name_to_demos.get(ref_name, []):
                    if dep_demo != demo:
                        self.demo_deps[demo].add(dep_demo)
                        self.demo_rdeps[dep_demo].add(demo)

    def demos_for_groups(self, groups: set[str]) -> set[Path]:
        result: set[Path] = set()
        for g in groups:
            result.update(self.group_to_demos.get(g, []))
        return result

    def expand_subgraph(self, seed: set[Path]) -> list[Path]:
        """Expand seed via forward + reverse dep traversal, return topo-sorted list."""
        included = set(seed)
        queue: deque[Path] = deque(seed)

        # Walk forward deps
        while queue:
            d = queue.popleft()
            for dep in self.demo_deps.get(d, set()):
                if dep not in included:
                    included.add(dep)
                    queue.append(dep)

        # Walk reverse deps from the full included set
        for d in list(included):
            for rdep in self.demo_rdeps.get(d, set()):
                if rdep not in included:
                    included.add(rdep)
                    queue.append(rdep)
        while queue:
            d = queue.popleft()
            for rdep in self.demo_rdeps.get(d, set()):
                if rdep not in included:
                    included.add(rdep)
                    queue.append(rdep)

        return self._topo_sort(included)

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

def detect_tier(changed_files: list[str]) -> str:
    tier = "skip"
    for f in changed_files:
        f = f.strip()
        if not f:
            continue
        if FULL_PATTERNS.search(f):
            return "full"
        if TARGETED_PATTERNS.search(f):
            tier = "targeted"
        elif not SKIP_PATTERNS.match(f) and tier == "skip":
            tier = "targeted"  # unknown path → safe fallback
    return tier


def extract_groups_from_paths(changed_files: list[str], demo_dir: Path) -> set[str]:
    groups: set[str] = set()
    for f in changed_files:
        f = f.strip()
        for pat in (CONFIG_GROUP_RE, APIS_GROUP_RE, CRD_GROUP_RE):
            m = pat.match(f)
            if m:
                groups.add(m.group(1))
                break
        # If a demo YAML changed directly, infer its group(s) by parsing it
        if f.startswith("dev/demos/"):
            p = REPO_ROOT / f
            if p.exists() and p.suffix == ".yaml":
                info = parse_demo_file(p)
                groups.update(info["groups"])
    return groups


# ---------------------------------------------------------------------------
# CLI commands
# ---------------------------------------------------------------------------

def cmd_select(args) -> None:
    demo_dir = REPO_ROOT / args.demo_dir
    if args.changed_files == "-":
        changed = sys.stdin.read().splitlines()
    else:
        with open(args.changed_files) as fh:
            changed = fh.read().splitlines()
    changed = [f.strip() for f in changed if f.strip()]

    tier = detect_tier(changed)

    if tier == "full":
        print("full")
        print("E2E_TIER=full", file=sys.stderr)
        print("KEYCLOAK_VERSIONS=all", file=sys.stderr)
        return

    if tier == "skip":
        print("skip")
        print("E2E_TIER=skip", file=sys.stderr)
        print("KEYCLOAK_VERSIONS=none", file=sys.stderr)
        return

    # targeted
    graph = DemoGraph(demo_dir)
    touched_groups = extract_groups_from_paths(changed, demo_dir)
    seed = graph.demos_for_groups(touched_groups)

    # Include directly changed demo files
    for f in changed:
        p = REPO_ROOT / f.strip()
        if p.suffix == ".yaml" and p.exists():
            if p.parent.parent == demo_dir and p.name != "000-init.yaml":
                seed.add(p)

    if not seed:
        # Changed config/ paths don't map to any known demo group → run full
        print("full")
        print("E2E_TIER=full (fallback: no demos match changed groups)", file=sys.stderr)
        print("KEYCLOAK_VERSIONS=all", file=sys.stderr)
        return

    ordered = graph.expand_subgraph(seed)
    paths = [f"./{p.relative_to(REPO_ROOT)}" for p in ordered]
    print(",".join(paths))
    print(f"E2E_TIER=targeted ({len(paths)} demos)", file=sys.stderr)
    print("KEYCLOAK_VERSIONS=latest", file=sys.stderr)


def cmd_index(args) -> None:
    demo_dir = REPO_ROOT / args.demo_dir
    graph = DemoGraph(demo_dir)
    json.dump(graph.inverted_index(), sys.stdout, indent=2)
    print()


def cmd_dag(args) -> None:
    demo_dir = REPO_ROOT / args.demo_dir
    graph = DemoGraph(demo_dir)
    json.dump(graph.full_dag(), sys.stdout, indent=2)
    print()


def cmd_query(args) -> None:
    """Which e2e tests use resource kind X (optionally filtered by group)?"""
    demo_dir = REPO_ROOT / args.demo_dir
    graph = DemoGraph(demo_dir)
    kind_lower = args.kind.lower()

    matches: dict[str, dict] = {}
    for (kind, group), demos in graph.resource_used_by.items():
        if kind.lower() == kind_lower:
            if args.group and group != args.group:
                continue
            label = f"{kind} ({group})"
            matches[label] = {
                "used_by": sorted(str(p.relative_to(REPO_ROOT)) for p in demos),
                "defined_in": sorted(
                    str(p.relative_to(REPO_ROOT))
                    for p in graph.resource_defined_by.get((kind, group), [])
                ),
            }

    if not matches:
        print(f"No demos found for kind '{args.kind}'", file=sys.stderr)
        sys.exit(1)

    json.dump(matches, sys.stdout, indent=2)
    print()


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
    sel.set_defaults(func=cmd_select)

    idx = sub.add_parser("index", help="Print inverted index (resource → demos) as JSON")
    idx.set_defaults(func=cmd_index)

    dag_cmd = sub.add_parser("dag", help="Print full dependency graph as JSON")
    dag_cmd.set_defaults(func=cmd_dag)

    q = sub.add_parser("query", help="Which e2e tests use resource kind X?")
    q.add_argument("kind", help="CRD kind name (e.g. ClientTimePolicy)")
    q.add_argument("--group", help="Optional API group filter (e.g. openidclient)")
    q.set_defaults(func=cmd_query)

    args = parser.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
