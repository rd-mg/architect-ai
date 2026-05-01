#!/usr/bin/env python3
"""
osheet_validate.py — Validate Odoo 19 .osheet JSON files for structural correctness.

Usage:
    python3 osheet_validate.py <file_or_glob> [--strict] [--json]

Exit code: 0 = all PASS, 1 = any FAIL (CRITICAL issues found).
"""

import argparse
import glob
import json
import re
import sys
from pathlib import Path


REQUIRED_TOP_LEVEL = {"version", "sheets"}
KNOWN_TOP_LEVEL_KEYS = {
    "version", "sheets", "styles", "formats", "borders", "revisionId",
    "uniqueFigureIds", "settings", "pivots", "pivotNextId", "customTableStyles",
    "globalFilters", "lists", "listNextId", "odooLinkReferences", "isNotSquishable",
}

FORMULA_REF_PATTERN = re.compile(
    r"\b(PIVOT\.VALUE|PIVOT\.HEADER|PIVOT)\s*\(\s*([0-9]+)",
    re.IGNORECASE,
)
LIST_REF_PATTERN = re.compile(
    r"\b(ODOO\.LIST(?:\.HEADER)?)\s*\(\s*([0-9]+)",
    re.IGNORECASE,
)
ODOO_VIEW_PATTERN = re.compile(r"odoo://view/(.+)$")


def _resolve_pivot(doc: dict, ref: str) -> bool:
    """V08: Accept pivot ref by UUID dict key OR by formulaId (numeric string)."""
    for key, pivot in doc.get("pivots", {}).items():
        if ref == key or ref == str(pivot.get("formulaId", "")):
            return True
    return False


def _resolve_list(doc: dict, ref: str) -> bool:
    """V09: Accept list ref by dict key OR by id field."""
    for key, lst in doc.get("lists", {}).items():
        if ref == key or ref == str(lst.get("id", "")):
            return True
    return False


def _validate_odoo_view_url(url: str) -> list[dict]:
    issues = []
    m = ODOO_VIEW_PATTERN.search(url)
    if not m:
        return issues
    payload_str = m.group(1)
    try:
        payload = json.loads(payload_str)
    except json.JSONDecodeError:
        issues.append({"severity": "CRITICAL", "rule": "V05", "message": f"Malformed JSON in odoo://view URL: {payload_str[:80]}"})
        return issues
    # V06
    action = payload.get("action", {})
    if not action.get("modelName"):
        issues.append({"severity": "CRITICAL", "rule": "V06", "message": "odoo://view action missing modelName"})
    # V07
    if not payload.get("viewType") and not action.get("viewType"):
        issues.append({"severity": "CRITICAL", "rule": "V07", "message": "odoo://view missing viewType"})
    return issues


def validate_file(path: Path, strict: bool = False) -> dict:
    try:
        with open(path, encoding="utf-8") as f:
            doc = json.load(f)
    except Exception as e:
        return {
            "file": str(path),
            "status": "FAIL",
            "issues": [{"severity": "CRITICAL", "rule": "LOAD", "message": f"Failed to load: {e}"}],
        }

    issues: list[dict] = []

    # V01 — version
    if "version" not in doc:
        issues.append({"severity": "CRITICAL", "rule": "V01", "message": "Missing top-level 'version' key"})

    # V02 — sheets
    if not doc.get("sheets"):
        issues.append({"severity": "CRITICAL", "rule": "V02", "message": "Missing or empty 'sheets' array"})

    styles_len = len(doc.get("styles", []))
    formats_len = len(doc.get("formats", []))
    borders_len = len(doc.get("borders", []))

    seen_pivot_refs: set[str] = set()
    seen_list_refs: set[str] = set()

    for sheet in doc.get("sheets", []):
        sheet_name = sheet.get("name", "?")

        # Figures
        for fig in sheet.get("figures", []):
            data = fig.get("data", {})
            fig_id = fig.get("id", "?")
            fig_type = data.get("type", "")

            # V03 — scorecard missing keyValue
            if fig_type == "scorecard" and not data.get("keyValue"):
                issues.append({"severity": "CRITICAL", "rule": "V03", "message": f"Scorecard {fig_id} missing keyValue"})

            # V04 — chart missing data.type
            if fig.get("tag") == "chart" and not fig_type:
                issues.append({"severity": "CRITICAL", "rule": "V04", "message": f"Figure {fig_id} missing data.type"})

        # Cells — check formula refs and style refs
        for cell_ref, cell in sheet.get("cells", {}).items():
            content = cell.get("content", "")

            # Pivot formula refs
            for m in FORMULA_REF_PATTERN.finditer(content):
                ref = m.group(2)
                seen_pivot_refs.add(ref)

            # List formula refs
            for m in LIST_REF_PATTERN.finditer(content):
                ref = m.group(2)
                seen_list_refs.add(ref)

            # W01 — style index
            if (style_id := cell.get("style")) is not None:
                if not isinstance(style_id, int) or style_id >= styles_len:
                    issues.append({"severity": "WARNING", "rule": "W01", "message": f"Sheet '{sheet_name}' cell {cell_ref}: undefined styleId {style_id}"})

            # W02 — format index
            if (fmt_id := cell.get("format")) is not None:
                if not isinstance(fmt_id, int) or fmt_id >= formats_len:
                    issues.append({"severity": "WARNING", "rule": "W02", "message": f"Sheet '{sheet_name}' cell {cell_ref}: undefined formatId {fmt_id}"})

    # V08 — pivot formula refs resolve
    for ref in seen_pivot_refs:
        if not _resolve_pivot(doc, ref):
            issues.append({"severity": "CRITICAL", "rule": "V08", "message": f"Pivot reference '{ref}' not found in pivots (checked key and formulaId)"})

    # V09 — list formula refs resolve
    for ref in seen_list_refs:
        if not _resolve_list(doc, ref):
            issues.append({"severity": "CRITICAL", "rule": "V09", "message": f"List reference '{ref}' not found in lists (checked key and id)"})

    # odoo://view links
    for link_id, link_obj in doc.get("odooLinkReferences", {}).items():
        url = link_obj.get("url", "")
        if "odoo://view" in url:
            issues.extend(_validate_odoo_view_url(url))

    # W04 — unknown top-level keys
    for k in doc:
        if k not in KNOWN_TOP_LEVEL_KEYS:
            issues.append({"severity": "WARNING", "rule": "W04", "message": f"Unknown top-level key: {k}"})

    # W05 — global filter without fields
    for gf in doc.get("globalFilters", []):
        if not gf.get("fields"):
            issues.append({"severity": "WARNING", "rule": "W05", "message": f"Global filter '{gf.get('id', '?')}' has no fields mapping"})

    # W06 — uniqueFigureIds count
    declared_ids = set(doc.get("uniqueFigureIds", []))
    actual_ids: set[str] = set()
    for sheet in doc.get("sheets", []):
        for fig in sheet.get("figures", []):
            if fid := fig.get("id"):
                actual_ids.add(fid)
    if declared_ids != actual_ids:
        issues.append({"severity": "WARNING", "rule": "W06", "message": f"uniqueFigureIds mismatch: declared={len(declared_ids)}, actual={len(actual_ids)}"})

    # W07 — pivotNextId
    pivot_count = len(doc.get("pivots", {}))
    pivot_next = doc.get("pivotNextId", pivot_count)
    if pivot_next < pivot_count:
        issues.append({"severity": "WARNING", "rule": "W07", "message": f"pivotNextId ({pivot_next}) < pivot count ({pivot_count})"})

    # W08 — listNextId
    list_count = len(doc.get("lists", {}))
    list_next = doc.get("listNextId", list_count)
    if list_next < list_count:
        issues.append({"severity": "WARNING", "rule": "W08", "message": f"listNextId ({list_next}) < list count ({list_count})"})

    # In strict mode: treat WARNING as CRITICAL
    if strict:
        for issue in issues:
            if issue["severity"] == "WARNING":
                issue["severity"] = "CRITICAL"

    has_critical = any(i["severity"] == "CRITICAL" for i in issues)
    return {
        "file": str(path),
        "status": "FAIL" if has_critical else "PASS",
        "issues": issues,
    }


def main():
    parser = argparse.ArgumentParser(description="Validate Odoo 19 .osheet JSON files.")
    parser.add_argument("paths", nargs="+", help="Files or glob patterns to validate")
    parser.add_argument("--strict", action="store_true", help="Treat WARNINGs as CRITICAL")
    parser.add_argument("--json", action="store_true", help="Output JSON report")
    args = parser.parse_args()

    # Expand globs
    files: list[Path] = []
    for pattern in args.paths:
        expanded = glob.glob(pattern)
        if expanded:
            files.extend(Path(p) for p in expanded)
        else:
            files.append(Path(pattern))

    results = []
    for f in files:
        result = validate_file(f, strict=args.strict)
        results.append(result)

    if args.json:
        print(json.dumps(results, indent=2, ensure_ascii=False))
    else:
        any_fail = False
        for r in results:
            status = r["status"]
            print(f"[{status}] {r['file']}")
            for issue in r["issues"]:
                print(f"  {issue['severity']}: [{issue['rule']}] {issue['message']}")
            if status == "FAIL":
                any_fail = True
        print()
        passed = sum(1 for r in results if r["status"] == "PASS")
        print(f"Results: {passed}/{len(results)} PASS")
        if any_fail:
            sys.exit(1)

    any_fail = any(r["status"] == "FAIL" for r in results)
    if any_fail:
        sys.exit(1)


if __name__ == "__main__":
    main()
