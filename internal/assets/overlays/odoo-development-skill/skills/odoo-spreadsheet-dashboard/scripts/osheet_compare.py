#!/usr/bin/env python3
"""
osheet_compare.py — Compare two Odoo 19 .osheet JSON files.

Usage:
    python3 osheet_compare.py original.json recreated.json --mode exact
    python3 osheet_compare.py original.json refactored.json --mode semantic
    python3 osheet_compare.py original.json refactored.json --mode profile

Modes:
    exact    — canonical SHA-256 equivalence (for lossless recreation validation)
    semantic — same top-level structure, sheet names, pivot/list counts, figure counts
    profile  — full profile equivalence (same models, chart types, formula counts)

Exit code: 0 = PASS, 1 = FAIL.
"""

import argparse
import hashlib
import json
import sys
from pathlib import Path


def _normalise(obj):
    if isinstance(obj, dict):
        return {k: _normalise(v) for k, v in sorted(obj.items())}
    if isinstance(obj, list):
        return [_normalise(i) for i in obj]
    return obj


def _sha256(doc: dict) -> str:
    return hashlib.sha256(
        json.dumps(_normalise(doc), ensure_ascii=False, separators=(",", ":")).encode("utf-8")
    ).hexdigest()


def _load(path: Path) -> dict:
    with open(path, encoding="utf-8") as f:
        return json.load(f)


def _sheet_summary(doc: dict) -> dict:
    sheets = doc.get("sheets", [])
    figure_count = sum(len(s.get("figures", [])) for s in sheets)
    chart_types: dict[str, int] = {}
    for s in sheets:
        for fig in s.get("figures", []):
            t = fig.get("data", {}).get("type", "unknown")
            chart_types[t] = chart_types.get(t, 0) + 1
    return {
        "sheet_names": [s.get("name", "") for s in sheets],
        "sheet_count": len(sheets),
        "pivot_count": len(doc.get("pivots", {})),
        "list_count": len(doc.get("lists", {})),
        "filter_count": len(doc.get("globalFilters", [])),
        "figure_count": figure_count,
        "chart_types": chart_types,
    }


def _cell_formula_counts(doc: dict) -> dict[str, int]:
    import re
    FORMULA_PATTERN = re.compile(
        r"\b(PIVOT\.VALUE|PIVOT\.HEADER|PIVOT|ODOO\.LIST\.HEADER|ODOO\.LIST"
        r"|ODOO\.BALANCE|ODOO\.DEBIT|ODOO\.CREDIT|ODOO\.ACCOUNT\.GROUP"
        r"|FORMAT\.LARGE\.NUMBER|IFERROR|_[Tt])\s*\(",
        re.IGNORECASE,
    )
    counts: dict[str, int] = {}
    for sheet in doc.get("sheets", []):
        for cell in sheet.get("cells", {}).values():
            content = cell.get("content", "")
            if content.startswith("="):
                for m in FORMULA_PATTERN.finditer(content):
                    name = m.group(1).upper()
                    counts[name] = counts.get(name, 0) + 1
    return counts


def _models(doc: dict) -> set[str]:
    models: set[str] = set()
    for piv in doc.get("pivots", {}).values():
        if m := piv.get("model"):
            models.add(m)
    for lst in doc.get("lists", {}).values():
        if m := lst.get("model"):
            models.add(m)
    return models


def compare_exact(a: dict, b: dict) -> tuple[bool, str]:
    sha_a = _sha256(a)
    sha_b = _sha256(b)
    if sha_a == sha_b:
        return True, f"SHA-256 match: {sha_a}"
    return False, f"SHA-256 mismatch:\n  A: {sha_a}\n  B: {sha_b}"


def compare_semantic(a: dict, b: dict) -> tuple[bool, str]:
    sa = _sheet_summary(a)
    sb = _sheet_summary(b)
    diffs = []
    for key in ("sheet_names", "sheet_count", "pivot_count", "list_count", "filter_count", "figure_count"):
        if sa[key] != sb[key]:
            diffs.append(f"  {key}: A={sa[key]} B={sb[key]}")
    if diffs:
        return False, "Semantic differences:\n" + "\n".join(diffs)
    return True, f"Semantic match: {sa['sheet_count']} sheets, {sa['pivot_count']} pivots, {sa['figure_count']} figures"


def compare_profile(a: dict, b: dict) -> tuple[bool, str]:
    ok_sem, msg_sem = compare_semantic(a, b)
    if not ok_sem:
        return False, msg_sem

    diffs = []
    models_a = _models(a)
    models_b = _models(b)
    if models_a != models_b:
        diffs.append(f"  models: A={sorted(models_a)} B={sorted(models_b)}")

    funcs_a = _cell_formula_counts(a)
    funcs_b = _cell_formula_counts(b)
    all_funcs = set(funcs_a) | set(funcs_b)
    for fn in sorted(all_funcs):
        ca = funcs_a.get(fn, 0)
        cb = funcs_b.get(fn, 0)
        if ca != cb:
            diffs.append(f"  formula {fn}: A={ca} B={cb}")

    sa = _sheet_summary(a)
    sb = _sheet_summary(b)
    if sa["chart_types"] != sb["chart_types"]:
        diffs.append(f"  chart_types: A={sa['chart_types']} B={sb['chart_types']}")

    if diffs:
        return False, "Profile differences:\n" + "\n".join(diffs)
    return True, "Profile match: models, formulas, and chart types are equivalent"


def main():
    parser = argparse.ArgumentParser(description="Compare two .osheet JSON files.")
    parser.add_argument("a", help="Original file")
    parser.add_argument("b", help="Recreated/refactored file")
    parser.add_argument("--mode", choices=["exact", "semantic", "profile"], default="semantic",
                        help="Comparison mode (default: semantic)")
    args = parser.parse_args()

    a = _load(Path(args.a))
    b = _load(Path(args.b))

    if args.mode == "exact":
        ok, msg = compare_exact(a, b)
    elif args.mode == "semantic":
        ok, msg = compare_semantic(a, b)
    else:
        ok, msg = compare_profile(a, b)

    status = "PASS" if ok else "FAIL"
    print(f"[{status}] mode={args.mode}")
    print(msg)

    if not ok:
        sys.exit(1)


if __name__ == "__main__":
    main()
