#!/usr/bin/env python3
"""
Compare two Terraform provider schema.json files and report:
  - New resources (not in old schema)
  - Removed resources (not in new schema)
  - Schema version changes for generated resources
  - Attribute changes (added/removed) for generated resources

Usage:
  version_diff.py <generated.lst> <old_schema.json> <new_schema.json>

Arguments:
  generated.lst    File listing currently generated resource names (one per line).
                   Lines starting with '#' are ignored.
  old_schema.json  The previous provider schema JSON file.
  new_schema.json  The current provider schema JSON file.

Exit codes:
  0  No changes detected
  1  Changes detected (new/removed resources or schema changes)
  2  Error (missing files, invalid JSON, etc.)
"""

import json
import sys


def load_generated_list(path):
    """Load the list of generated resource names from a file."""
    resources = set()
    with open(path) as f:
        for line in f:
            line = line.strip()
            if line and not line.startswith("#"):
                resources.add(line)
    return resources


def load_schema(path):
    """Load and return the resource_schemas from a provider schema JSON file."""
    with open(path) as f:
        data = json.load(f)
    # Navigate to the resource schemas; handle both wrapped and flat formats
    provider_schemas = data.get("provider_schemas", {})
    for provider_key, provider_data in provider_schemas.items():
        return provider_data.get("resource_schemas", {}), provider_data.get("data_source_schemas", {})
    return {}, {}


def get_schema_version(resource_schema):
    """Extract the schema version from a resource schema block."""
    return resource_schema.get("version", 0)


def collect_attributes(block, prefix=""):
    """Recursively collect all attribute paths from a schema block."""
    attrs = {}
    if not block:
        return attrs

    block_body = block.get("block", block)

    for attr_name, attr_def in block_body.get("attributes", {}).items():
        full_path = f"{prefix}{attr_name}"
        attrs[full_path] = {
            "type": attr_def.get("type"),
            "required": attr_def.get("required", False),
            "optional": attr_def.get("optional", False),
            "computed": attr_def.get("computed", False),
            "sensitive": attr_def.get("sensitive", False),
        }

    for bt_name, bt_def in block_body.get("block_types", {}).items():
        nested_prefix = f"{prefix}{bt_name}."
        nested_attrs = collect_attributes(bt_def, nested_prefix)
        attrs.update(nested_attrs)

    return attrs


def diff_attributes(old_attrs, new_attrs):
    """Compare two attribute dicts and return added, removed, and changed."""
    old_keys = set(old_attrs.keys())
    new_keys = set(new_attrs.keys())

    added = sorted(new_keys - old_keys)
    removed = sorted(old_keys - new_keys)

    changed = []
    for key in sorted(old_keys & new_keys):
        old_val = old_attrs[key]
        new_val = new_attrs[key]
        if old_val != new_val:
            changed.append((key, old_val, new_val))

    return added, removed, changed


def format_attr_flags(attr):
    """Format attribute flags into a human-readable string."""
    flags = []
    if attr.get("required"):
        flags.append("required")
    if attr.get("optional"):
        flags.append("optional")
    if attr.get("computed"):
        flags.append("computed")
    if attr.get("sensitive"):
        flags.append("sensitive")
    return ", ".join(flags) if flags else "none"


def main():
    if len(sys.argv) != 4:
        print(__doc__)
        sys.exit(2)

    generated_lst_path = sys.argv[1]
    old_schema_path = sys.argv[2]
    new_schema_path = sys.argv[3]

    try:
        generated = load_generated_list(generated_lst_path)
    except FileNotFoundError:
        print(f"Error: generated list file not found: {generated_lst_path}", file=sys.stderr)
        sys.exit(2)

    try:
        old_resources, old_datasources = load_schema(old_schema_path)
    except (FileNotFoundError, json.JSONDecodeError) as e:
        print(f"Error loading old schema: {e}", file=sys.stderr)
        sys.exit(2)

    try:
        new_resources, new_datasources = load_schema(new_schema_path)
    except (FileNotFoundError, json.JSONDecodeError) as e:
        print(f"Error loading new schema: {e}", file=sys.stderr)
        sys.exit(2)

    old_resource_names = set(old_resources.keys())
    new_resource_names = set(new_resources.keys())

    has_changes = False

    # --- New Resources ---
    new_res = sorted(new_resource_names - old_resource_names)
    if new_res:
        has_changes = True
        print("=" * 70)
        print("NEW RESOURCES (not in previous schema)")
        print("=" * 70)
        for r in new_res:
            status = "[generated]" if r in generated else "[NOT generated]"
            print(f"  + {r}  {status}")
        print()

    # --- Removed Resources ---
    removed_res = sorted(old_resource_names - new_resource_names)
    if removed_res:
        has_changes = True
        print("=" * 70)
        print("REMOVED RESOURCES (no longer in schema)")
        print("=" * 70)
        for r in removed_res:
            status = "[was generated]" if r in generated else "[was NOT generated]"
            print(f"  - {r}  {status}")
        print()

    # --- Schema Version Changes (generated resources only) ---
    version_changes = []
    for r in sorted(generated & old_resource_names & new_resource_names):
        old_ver = get_schema_version(old_resources[r])
        new_ver = get_schema_version(new_resources[r])
        if old_ver != new_ver:
            version_changes.append((r, old_ver, new_ver))

    if version_changes:
        has_changes = True
        print("=" * 70)
        print("SCHEMA VERSION CHANGES (generated resources)")
        print("=" * 70)
        for r, old_ver, new_ver in version_changes:
            print(f"  {r}: version {old_ver} -> {new_ver}")
        print()

    # --- Attribute Changes (generated resources only) ---
    attr_changes_found = False
    for r in sorted(generated & old_resource_names & new_resource_names):
        old_attrs = collect_attributes(old_resources[r])
        new_attrs = collect_attributes(new_resources[r])
        added, removed, changed = diff_attributes(old_attrs, new_attrs)

        if added or removed or changed:
            if not attr_changes_found:
                print("=" * 70)
                print("ATTRIBUTE CHANGES (generated resources)")
                print("=" * 70)
                attr_changes_found = True
                has_changes = True

            print(f"\n  {r}:")
            for a in added:
                flags = format_attr_flags(new_attrs[a])
                print(f"    + {a}  ({flags})")
            for a in removed:
                print(f"    - {a}")
            for key, old_val, new_val in changed:
                old_flags = format_attr_flags(old_val)
                new_flags = format_attr_flags(new_val)
                print(f"    ~ {key}: ({old_flags}) -> ({new_flags})")

    if attr_changes_found:
        print()

    # --- Summary of not-yet-generated resources ---
    not_generated = sorted(new_resource_names - generated)
    if not_generated:
        print("=" * 70)
        print(f"RESOURCES NOT YET GENERATED ({len(not_generated)} total)")
        print("=" * 70)
        for r in not_generated:
            new_indicator = " [NEW]" if r in new_res else ""
            print(f"  {r}{new_indicator}")
        print()

    # --- Data Source Changes ---
    old_ds_names = set(old_datasources.keys())
    new_ds_names = set(new_datasources.keys())
    new_ds = sorted(new_ds_names - old_ds_names)
    removed_ds = sorted(old_ds_names - new_ds_names)

    if new_ds:
        has_changes = True
        print("=" * 70)
        print("NEW DATA SOURCES")
        print("=" * 70)
        for d in new_ds:
            print(f"  + {d}")
        print()

    if removed_ds:
        has_changes = True
        print("=" * 70)
        print("REMOVED DATA SOURCES")
        print("=" * 70)
        for d in removed_ds:
            print(f"  - {d}")
        print()

    if not has_changes:
        print("No changes detected between the two schema versions.")

    sys.exit(1 if has_changes else 0)


if __name__ == "__main__":
    main()
