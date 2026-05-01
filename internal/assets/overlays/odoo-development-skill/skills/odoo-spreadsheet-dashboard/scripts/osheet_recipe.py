#!/usr/bin/env python3
"""
osheet_recipe.py — Lossless recipe export / rebuild / catalog for Odoo 19 .osheet JSON.

Commands:
    export <source.osheet.json> <recipe.json>          — export normalised recipe
    build  <recipe.json>        <output.osheet.json>   — rebuild from recipe
    catalog <dir_of_osheets>    <output_dir>           — export all samples to recipes

A recipe is the canonical, deterministic normalisation of an .osheet payload.
export → build → compare --mode exact must yield PASS (SHA-256 match).
"""

import argparse
import hashlib
import json
import sys
from pathlib import Path


def _normalise(obj):
    """Recursively sort dict keys for deterministic JSON serialization."""
    if isinstance(obj, dict):
        return {k: _normalise(v) for k, v in sorted(obj.items())}
    if isinstance(obj, list):
        return [_normalise(i) for i in obj]
    return obj


def _canonical_bytes(doc: dict) -> bytes:
    """Return canonical JSON bytes (sorted keys, no trailing whitespace)."""
    return json.dumps(_normalise(doc), ensure_ascii=False, separators=(",", ":")).encode("utf-8")


def _sha256(doc: dict) -> str:
    return hashlib.sha256(_canonical_bytes(doc)).hexdigest()


def cmd_export(source: Path, recipe_out: Path):
    with open(source, encoding="utf-8") as f:
        doc = json.load(f)

    recipe = {
        "schema": "osheet-recipe-v1",
        "source_file": source.name,
        "source_sha256": _sha256(doc),
        "payload": _normalise(doc),
    }

    recipe_out.parent.mkdir(parents=True, exist_ok=True)
    with open(recipe_out, "w", encoding="utf-8") as f:
        json.dump(recipe, f, indent=2, ensure_ascii=False)

    print(f"Exported recipe: {recipe_out}")
    print(f"Source SHA-256: {recipe['source_sha256']}")


def cmd_build(recipe_in: Path, output: Path):
    with open(recipe_in, encoding="utf-8") as f:
        recipe = json.load(f)

    if recipe.get("schema") != "osheet-recipe-v1":
        print("ERROR: not an osheet-recipe-v1 file", file=sys.stderr)
        sys.exit(1)

    payload = recipe["payload"]

    output.parent.mkdir(parents=True, exist_ok=True)
    with open(output, "w", encoding="utf-8") as f:
        json.dump(payload, f, indent=2, ensure_ascii=False)

    rebuilt_sha = _sha256(payload)
    source_sha = recipe.get("source_sha256", "")
    match = rebuilt_sha == source_sha
    print(f"Built: {output}")
    print(f"Rebuilt SHA-256:  {rebuilt_sha}")
    print(f"Source SHA-256:   {source_sha}")
    print(f"SHA-256 match:    {'YES — lossless' if match else 'NO — diverged'}")
    if not match:
        sys.exit(1)


def cmd_catalog(source_dir: Path, output_dir: Path):
    output_dir.mkdir(parents=True, exist_ok=True)
    files = sorted(p for p in source_dir.iterdir() if p.suffix == ".json")
    if not files:
        print(f"No .json files found in {source_dir}", file=sys.stderr)
        sys.exit(1)

    results = []
    for source in files:
        recipe_name = source.stem.replace(" ", "_").replace("(", "").replace(")", "") + ".recipe.json"
        recipe_out = output_dir / recipe_name
        try:
            with open(source, encoding="utf-8") as f:
                doc = json.load(f)
            recipe = {
                "schema": "osheet-recipe-v1",
                "source_file": source.name,
                "source_sha256": _sha256(doc),
                "payload": _normalise(doc),
            }
            with open(recipe_out, "w", encoding="utf-8") as f:
                json.dump(recipe, f, indent=2, ensure_ascii=False)
            results.append((source.name, recipe_out.name, "OK"))
        except Exception as e:
            results.append((source.name, "", f"ERROR: {e}"))

    print(f"Cataloged {len(results)} files → {output_dir}")
    for name, out, status in results:
        print(f"  {status:6}  {name} → {out}")


def main():
    parser = argparse.ArgumentParser(description="Osheet recipe: lossless export/build/catalog.")
    subparsers = parser.add_subparsers(dest="cmd", required=True)

    exp = subparsers.add_parser("export", help="Export .osheet to recipe")
    exp.add_argument("source", help="Source .osheet.json file")
    exp.add_argument("recipe", help="Output recipe.json file")

    bld = subparsers.add_parser("build", help="Build .osheet from recipe")
    bld.add_argument("recipe", help="Input recipe.json file")
    bld.add_argument("output", help="Output .osheet.json file")

    cat = subparsers.add_parser("catalog", help="Export all samples in a directory")
    cat.add_argument("source_dir", help="Directory of .osheet.json files")
    cat.add_argument("output_dir", help="Output directory for recipe files")

    args = parser.parse_args()

    if args.cmd == "export":
        cmd_export(Path(args.source), Path(args.recipe))
    elif args.cmd == "build":
        cmd_build(Path(args.recipe), Path(args.output))
    elif args.cmd == "catalog":
        cmd_catalog(Path(args.source_dir), Path(args.output_dir))


if __name__ == "__main__":
    main()
