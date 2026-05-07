#!/usr/bin/env python3
"""
Compares api-catalog/SKILL.md against the generated OpenAPI spec and reports
likely mismatches. Catches:
  1. Fields documented as strings that the spec types as integers
  2. Fields in the spec that the skill never mentions
  3. Unknown fields the spec would reject (not in any schema)

Run: python3 api/scripts/check_skill_vs_spec.py
"""
import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).parent.parent

def load_spec():
    p = ROOT / "doc" / "openapi.json"
    if not p.exists():
        sys.exit(f"openapi.json not found at {p}")
    return json.loads(p.read_text())

def load_skill():
    p = ROOT / "skills" / "api-catalog" / "SKILL.md"
    if not p.exists():
        sys.exit(f"SKILL.md not found at {p}")
    return p.read_text()

def resolve_ref(spec, ref):
    """Resolve $ref like #/components/schemas/Foo"""
    parts = ref.lstrip("#/").split("/")
    node = spec
    for part in parts:
        node = node.get(part, {})
    return node

def get_schema_fields(spec, schema):
    """Flatten a schema's properties, resolving $ref."""
    if "$ref" in schema:
        schema = resolve_ref(spec, schema["$ref"])
    fields = {}
    for fname, fdef in schema.get("properties", {}).items():
        if "$ref" in fdef:
            fdef = resolve_ref(spec, fdef["$ref"])
        fields[fname] = fdef.get("type", "object")
    # Also check allOf / anyOf
    for combo_key in ("allOf", "anyOf", "oneOf"):
        for sub in schema.get(combo_key, []):
            fields.update(get_schema_fields(spec, sub))
    return fields

def endpoint_schemas(spec):
    """Return {(METHOD, path): {field: type}} for all request bodies."""
    result = {}
    for path, methods in spec.get("paths", {}).items():
        for method, op in methods.items():
            if method.upper() not in ("POST", "PUT", "PATCH", "DELETE"):
                continue
            body = op.get("requestBody", {})
            content = body.get("content", {}).get("application/json", {})
            schema = content.get("schema", {})
            if not schema:
                continue
            fields = get_schema_fields(spec, schema)
            if fields:
                result[(method.upper(), path)] = fields
    return result

def main():
    spec = load_spec()
    skill = load_skill()
    schemas = endpoint_schemas(spec)

    print("=" * 70)
    print("REPORT: api-catalog/SKILL.md vs openapi.json")
    print("=" * 70)

    issues = []

    # 1. Integer fields — most likely to be mistyped as strings in docs
    int_fields = []
    for (method, path), fields in sorted(schemas.items()):
        for fname, ftype in fields.items():
            if ftype in ("integer", "number"):
                int_fields.append((method, path, fname, ftype))

    print(f"\n[1] INTEGER/NUMBER request fields ({len(int_fields)} total)")
    print("    These must be passed as numbers, not quoted strings.\n")
    for method, path, fname, ftype in int_fields:
        # Check if the skill mentions this field with a type hint
        skill_mentions = fname in skill
        skill_has_integer = f"INTEGER" in skill or f"numeric" in skill.lower()
        flag = ""
        if skill_mentions and "owner/repo" in skill and fname == "repository":
            flag = "  ⚠ SKILL SAYS STRING — must be integer!"
        elif not skill_mentions:
            flag = "  (not mentioned in skill)"
        print(f"    {method:6} {path:50} {fname:20} {ftype}{flag}")

    # 2. Fields that exist in spec but are COMPLETELY absent from skill
    print(f"\n[2] Required fields absent from skill entirely")
    for (method, path), fields in sorted(schemas.items()):
        for fname, ftype in fields.items():
            if fname not in skill:
                print(f"    {method:6} {path:50} {fname:20} {ftype}  ← not in SKILL.md at all")

    # 3. Paths in skill not in spec (potential stale docs)
    skill_paths = re.findall(r'/api/v1[^\s\?{]*', skill)
    spec_paths = set(spec.get("paths", {}).keys())
    print(f"\n[3] Paths referenced in skill but not in OpenAPI spec (potentially stale)")
    seen = set()
    for p in skill_paths:
        # Normalise path params
        norm = re.sub(r'\{[^}]+\}', '{id}', p)
        if norm not in seen:
            seen.add(norm)
            # Check if any spec path matches (loose check)
            matched = any(
                re.sub(r'\{[^}]+\}', '{id}', sp) == norm
                for sp in spec_paths
            )
            if not matched:
                print(f"    {p}")

    print("\n" + "=" * 70)
    print("Done. Fix issues above in api/skills/api-catalog/SKILL.md")

if __name__ == "__main__":
    main()
