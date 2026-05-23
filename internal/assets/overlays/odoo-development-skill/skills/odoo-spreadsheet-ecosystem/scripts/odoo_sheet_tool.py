#!/usr/bin/env python3
"""
odoo_sheet_tool.py — Unified CLI for Odoo Spreadsheet operations
Replaces: osheet_build.py, osheet_validate.py, osheet_compare.py,
          json_builder.py, formula_check.py, xlsx_reader.py, xlsx_pack.py
"""
import argparse, json, os, re, sys, zipfile
from pathlib import Path


# ── Validation ────────────────────────────────────────────────────────────────

def validate_osheet(data: dict) -> list[str]:
    issues = []
    for key in ["version", "sheets"]:
        if key not in data:
            issues.append(f"MISSING key: {key}")
    for sheet in data.get("sheets", []):
        for ref, cell in sheet.get("cells", {}).items():
            if isinstance(cell, dict):
                content = cell.get("content", "")
                if not isinstance(content, str):
                    issues.append(f"NON-STRING: {sheet.get('name','?')}!{ref} type={type(content).__name__}")
    for sheet in data.get("sheets", []):
        for ref, cell in sheet.get("cells", {}).items():
            if isinstance(cell, dict):
                c = str(cell.get("content", ""))
                for err in ["#ERROR", "#BAD_EXPR", "#SPILL!", "#DIV/0!"]:
                    if err in c:
                        issues.append(f"ERROR_LITERAL: {sheet.get('name','?')}!{ref}: {err}")
    seen_ids = set()
    for fid in [p.get("formulaId") for p in data.get("pivots", {}).values()]:
        if fid in seen_ids:
            issues.append(f"DUPLICATE formulaId: {fid}")
        seen_ids.add(fid)
    for sheet in data.get("sheets", []):
        for ref, cell in sheet.get("cells", {}).items():
            if isinstance(cell, dict):
                c = str(cell.get("content", ""))
                if re.search(r'=[^=].*/', c) and "IFERROR" not in c:
                    issues.append(f"UNPROTECTED_DIVISION: {sheet.get('name','?')}!{ref}")
    return issues


# ── Builders ──────────────────────────────────────────────────────────────────

def _base_osheet(name: str) -> dict:
    return {
        "version": "18.5.10",
        "sheets": [{"id": "s1", "name": name, "cells": {}, "figures": [], "tables": [], "cols": {}, "rows": {}}],
        "styles": {
            "1": {"fillColor": "#6C4E65", "textColor": "#FFFFFF", "bold": True, "fontSize": 11},
            "2": {"fillColor": "#E4D9E1", "bold": True},
        },
        "borders": {}, "formats": {"currency": "$#,##0.00", "percent": "0.00%"},
        "pivots": {}, "globalFilters": [],
        "settings": {"locale": "en_US"},
        "revisionId": "1", "pivotNextId": 1, "uniqueFigureIds": [],
    }


def build_dashboard(spec: dict) -> dict:
    wb = _base_osheet(spec.get("dashboard_name", "Dashboard"))
    wb["sheets"].append({"id": "summary", "name": "Sales Period Summary", "cells": {}, "figures": [], "tables": [], "cols": {}, "rows": {}})
    return wb


def build_quote(spec: dict) -> dict:
    wb = _base_osheet(spec.get("quote_name", "Quote Dashboard"))
    for sid, name in [("calc", "Quote Calculator"), ("pricing", "Product Pricing"),
                      ("options", "Optional Products"), ("approval", "Approval Rules"),
                      ("summary", "Quote Summary")]:
        wb["sheets"].append({"id": sid, "name": name, "cells": {}, "figures": [], "tables": [], "cols": {}, "rows": {}})
    return wb


def build_generic(spec: dict) -> dict:
    return _base_osheet(spec.get("sheet_name", "Sheet1"))


# ── XLSX Operations ───────────────────────────────────────────────────────────

def xlsx_validate(args):
    p = Path(args.file)
    if not zipfile.is_zipfile(p):
        print(f"NOT VALID XLSX: {p}", file=sys.stderr); sys.exit(1)
    with zipfile.ZipFile(p) as z:
        for req in ["xl/workbook.xml", "[Content_Types].xml"]:
            if req not in z.namelist():
                print(f"MISSING: {req}", file=sys.stderr); sys.exit(1)
    print(f"VALID XLSX: {p}")


def xlsx_read(args):
    p = Path(args.file)
    with zipfile.ZipFile(p) as z:
        sheets = [n for n in z.namelist() if n.startswith("xl/worksheets/")]
        print(f"Sheets ({len(sheets)}):")
        for s in sheets[:10]:
            print(f"  {s}")


def xlsx_unpack(args):
    p, out = Path(args.file), Path(args.out or Path(args.file).stem + "_unpacked")
    out.mkdir(parents=True, exist_ok=True)
    with zipfile.ZipFile(p) as z:
        z.extractall(out)
    print(f"Unpacked: {out}")


def xlsx_pack(args):
    src, out = Path(args.src), Path(args.out or "output.xlsx")
    with zipfile.ZipFile(out, "w", zipfile.ZIP_DEFLATED) as z:
        for f in src.rglob("*"):
            if f.is_file():
                z.write(f, f.relative_to(src))
    print(f"Packed: {out}")


# ── Compare ───────────────────────────────────────────────────────────────────

def compare(args):
    a = json.loads(Path(args.file_a).read_text())
    b = json.loads(Path(args.file_b).read_text())
    diffs = []
    a_sheets = {s["name"] for s in a.get("sheets", [])}
    b_sheets = {s["name"] for s in b.get("sheets", [])}
    for n in a_sheets - b_sheets: diffs.append(f"REMOVED sheet: {n}")
    for n in b_sheets - a_sheets: diffs.append(f"ADDED sheet: {n}")
    a_pivots = set(a.get("pivots", {}).keys())
    b_pivots = set(b.get("pivots", {}).keys())
    for p in a_pivots - b_pivots: diffs.append(f"REMOVED pivot: {p}")
    for p in b_pivots - a_pivots: diffs.append(f"ADDED pivot: {p}")
    if diffs:
        print("DIFFERENCES:"); [print(f"  {d}") for d in diffs]
    else:
        print("IDENTICAL")


# ── CLI ───────────────────────────────────────────────────────────────────────

def main():
    ap = argparse.ArgumentParser(description="Odoo Sheet Tool")
    sub = ap.add_subparsers(dest="command", required=True)

    p_build = sub.add_parser("build")
    p_build.add_argument("--type", choices=["dashboard", "quote", "generic"], default="generic")
    p_build.add_argument("--blueprint", help="JSON spec file")
    p_build.add_argument("--out", help="Output path")

    p_val = sub.add_parser("validate")
    p_val.add_argument("--file", required=True)

    p_xlsx = sub.add_parser("xlsx")
    p_xlsx.add_argument("--action", choices=["read", "validate", "unpack", "pack"], required=True)
    p_xlsx.add_argument("--file"); p_xlsx.add_argument("--src"); p_xlsx.add_argument("--out")

    p_cmp = sub.add_parser("compare")
    p_cmp.add_argument("--file-a", required=True); p_cmp.add_argument("--file-b", required=True)

    args = ap.parse_args()

    if args.command == "build":
        spec = json.loads(Path(args.blueprint).read_text()) if args.blueprint else {}
        builders = {"dashboard": build_dashboard, "quote": build_quote, "generic": build_generic}
        result = builders[args.type](spec)
        out = Path(args.out or f"output_{args.type}.osheet.json")
        out.write_text(json.dumps(result, indent=2, ensure_ascii=False))
        issues = validate_osheet(result)
        print(f"Built: {out}")
        if issues:
            print("VALIDATION WARNINGS:")
            for i in issues: print(f"  {i}")
        else:
            print("Validation: PASS")

    elif args.command == "validate":
        p = Path(args.file)
        try:
            data = json.loads(p.read_text())
        except json.JSONDecodeError as e:
            print(f"INVALID JSON: {e}", file=sys.stderr); sys.exit(1)
        issues = validate_osheet(data)
        if issues:
            print("VALIDATION FAILED:")
            for i in issues: print(f"  {i}")
            sys.exit(1)
        print("VALID: all checks passed")

    elif args.command == "xlsx":
        actions = {"read": xlsx_read, "validate": xlsx_validate,
                   "unpack": xlsx_unpack, "pack": xlsx_pack}
        actions[args.action](args)

    elif args.command == "compare":
        compare(args)


if __name__ == "__main__":
    main()
