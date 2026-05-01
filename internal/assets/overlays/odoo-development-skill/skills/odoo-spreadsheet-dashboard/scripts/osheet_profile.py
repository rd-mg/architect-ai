#!/usr/bin/env python3
"""
osheet_profile.py — Profile an Odoo 19 .osheet JSON file.

Usage:
    python3 osheet_profile.py <file_or_dir> [--markdown] [--out PATH] [--json]
"""

import argparse
import json
import os
import re
import sys
from pathlib import Path


KNOWN_TOP_LEVEL_KEYS = {
    "version", "sheets", "styles", "formats", "borders", "revisionId",
    "uniqueFigureIds", "settings", "pivots", "pivotNextId", "customTableStyles",
    "globalFilters", "lists", "listNextId", "odooLinkReferences", "isNotSquishable",
}

# Odoo and helper formula names to detect in cell content
FORMULA_PATTERN = re.compile(
    r"\b(PIVOT\.VALUE|PIVOT\.HEADER|PIVOT|ODOO\.LIST\.HEADER|ODOO\.LIST"
    r"|ODOO\.BALANCE|ODOO\.DEBIT|ODOO\.CREDIT|ODOO\.ACCOUNT\.GROUP"
    r"|FORMAT\.LARGE\.NUMBER|IFERROR|CONTRACTION|EXPANSION|CHOOSECOLS"
    r"|CONCATENATE|YEAR|LEFT|RIGHT|ROUND|_[Tt])\s*\(",
    re.IGNORECASE,
)


def _iter_osheets(path: Path):
    if path.is_file():
        yield path
    elif path.is_dir():
        for p in sorted(path.iterdir()):
            if p.suffix in (".json",) and "osheet" in p.name.lower():
                yield p


def _extract_formulas(doc: dict) -> dict:
    counts: dict[str, int] = {}
    for sheet in doc.get("sheets", []):
        for cell in sheet.get("cells", {}).values():
            content = cell.get("content", "")
            if content.startswith("="):
                for m in FORMULA_PATTERN.finditer(content):
                    name = m.group(1).upper().rstrip("(")
                    counts[name] = counts.get(name, 0) + 1
    return counts


def _get_pivot_ids(doc: dict) -> list[dict]:
    result = []
    for key, piv in doc.get("pivots", {}).items():
        result.append({
            "key": key,
            "formulaId": piv.get("formulaId", ""),
            "model": piv.get("model", ""),
            "name": piv.get("name", ""),
        })
    return result


def _get_list_ids(doc: dict) -> list[dict]:
    result = []
    for key, lst in doc.get("lists", {}).items():
        result.append({
            "key": key,
            "id": lst.get("id", key),
            "model": lst.get("model", ""),
            "name": lst.get("name", ""),
        })
    return result


def _get_filter_summary(doc: dict) -> list[dict]:
    result = []
    for f in doc.get("globalFilters", []):
        result.append({
            "id": f.get("id", ""),
            "type": f.get("type", ""),
            "label": f.get("label", ""),
            "fields_count": len(f.get("fields", {})),
        })
    return result


def _get_figures(doc: dict) -> list[dict]:
    result = []
    for sheet in doc.get("sheets", []):
        for fig in sheet.get("figures", []):
            data = fig.get("data", {})
            result.append({
                "id": fig.get("id", ""),
                "type": data.get("type", "unknown"),
                "sheet": sheet.get("name", ""),
                "width": fig.get("width", 0),
                "height": fig.get("height", 0),
            })
    return result


def _get_models(doc: dict) -> set[str]:
    models: set[str] = set()
    for piv in doc.get("pivots", {}).values():
        if m := piv.get("model"):
            models.add(m)
    for lst in doc.get("lists", {}).values():
        if m := lst.get("model"):
            models.add(m)
    for sheet in doc.get("sheets", []):
        for fig in sheet.get("figures", []):
            data = fig.get("data", {})
            for ds in data.get("odooDataSets", []):
                if m := ds.get("metaData", {}).get("resModel"):
                    models.add(m)
    return models


def _get_odoo_links(doc: dict) -> list[str]:
    return list(doc.get("odooLinkReferences", {}).keys())


def _get_unknown_keys(doc: dict) -> list[str]:
    return [k for k in doc if k not in KNOWN_TOP_LEVEL_KEYS]


def profile_file(path: Path) -> dict:
    with open(path, encoding="utf-8") as f:
        doc = json.load(f)

    sheets = doc.get("sheets", [])
    figures = _get_figures(doc)
    formula_counts = _extract_formulas(doc)

    chart_types: dict[str, int] = {}
    for fig in figures:
        t = fig["type"]
        chart_types[t] = chart_types.get(t, 0) + 1

    return {
        "file": path.name,
        "version": doc.get("version", ""),
        "sheets": [{"name": s.get("name", ""), "visible": s.get("isVisible", True)} for s in sheets],
        "sheet_count": len(sheets),
        "pivot_count": len(doc.get("pivots", {})),
        "list_count": len(doc.get("lists", {})),
        "filter_count": len(doc.get("globalFilters", [])),
        "figure_count": len(figures),
        "formula_count": sum(formula_counts.values()),
        "formula_functions": formula_counts,
        "chart_types": chart_types,
        "pivots": _get_pivot_ids(doc),
        "lists": _get_list_ids(doc),
        "filters": _get_filter_summary(doc),
        "figures": figures,
        "models": sorted(_get_models(doc)),
        "odoo_links": _get_odoo_links(doc),
        "unknown_keys": _get_unknown_keys(doc),
        "top_level_keys": sorted(doc.keys()),
    }


def _as_markdown(profiles: list[dict]) -> str:
    lines = ["# Odoo .osheet Profile Report", ""]

    # Summary table
    lines.append("| file | version | sheets | pivots | lists | filters | figures | formulas | models | chart types |")
    lines.append("|------|--------:|-------:|-------:|------:|--------:|--------:|---------:|--------|-------------|")
    for p in profiles:
        models = ", ".join(p["models"])
        ctypes = ", ".join(p["chart_types"].keys())
        lines.append(
            f"| {p['file']} | {p['version']} | {p['sheet_count']} | {p['pivot_count']} "
            f"| {p['list_count']} | {p['filter_count']} | {p['figure_count']} "
            f"| {p['formula_count']} | {models} | {ctypes} |"
        )

    lines.append("")
    lines.append("## Aggregate formula functions")
    lines.append("")
    all_funcs: dict[str, int] = {}
    for p in profiles:
        for k, v in p["formula_functions"].items():
            all_funcs[k] = all_funcs.get(k, 0) + v
    for k, v in sorted(all_funcs.items(), key=lambda x: -x[1]):
        lines.append(f"- `{k}`: {v}")

    lines.append("")
    lines.append("## Aggregate chart types")
    lines.append("")
    all_charts: dict[str, int] = {}
    for p in profiles:
        for k, v in p["chart_types"].items():
            all_charts[k] = all_charts.get(k, 0) + v
    for k, v in sorted(all_charts.items(), key=lambda x: -x[1]):
        lines.append(f"- `{k}`: {v}")

    return "\n".join(lines)


def main():
    parser = argparse.ArgumentParser(description="Profile Odoo 19 .osheet JSON files.")
    parser.add_argument("path", help="File or directory to profile")
    parser.add_argument("--markdown", action="store_true", help="Output Markdown instead of JSON")
    parser.add_argument("--json", action="store_true", help="Output JSON (default)")
    parser.add_argument("--out", help="Write output to file instead of stdout")
    args = parser.parse_args()

    target = Path(args.path)
    if not target.exists():
        print(f"ERROR: {target} does not exist", file=sys.stderr)
        sys.exit(1)

    profiles = [profile_file(p) for p in _iter_osheets(target)]
    if not profiles:
        print("ERROR: no .osheet JSON files found", file=sys.stderr)
        sys.exit(1)

    if args.markdown:
        output = _as_markdown(profiles)
    else:
        output = json.dumps(profiles, indent=2, ensure_ascii=False)

    if args.out:
        Path(args.out).write_text(output, encoding="utf-8")
        print(f"Written to {args.out}", file=sys.stderr)
    else:
        print(output)


if __name__ == "__main__":
    main()
