#!/usr/bin/env python3
"""
osheet_build.py — Build a new Odoo 19 .osheet JSON from a declarative DashboardBlueprint.

Usage:
    python3 osheet_build.py blueprint.json output.osheet.json [--pretty]

Blueprint format: see assets/blueprints/monthly-sales.json for a full example.
"""

import argparse
import json
import sys
import uuid
from pathlib import Path


OSHEET_VERSION = "19.1.2"

# Default styles: index 0 = normal, 1 = bold, 2 = bold+large, 3 = header
DEFAULT_STYLES = [
    {},
    {"bold": True},
    {"bold": True, "fontSize": 14},
    {"bold": True, "textColor": "#1F4E79"},
]
DEFAULT_FORMATS = ["", "#,##0", "#,##0.00", "[$€]#,##0.00"]
DEFAULT_BORDERS = [None]


def _new_uuid() -> str:
    return str(uuid.uuid4())


def _make_pivot(spec: dict, formula_id: int) -> tuple[str, dict]:
    """Return (uuid_key, pivot_dict) from a blueprint pivot spec."""
    key = spec.get("id") or _new_uuid()
    pivot = {
        "id": key,
        "formulaId": str(formula_id),
        "model": spec["model"],
        "domain": spec.get("domain", []),
        "context": spec.get("context", {}),
        "measures": spec.get("measures", []),
        "columns": spec.get("columns", []),
        "rows": spec.get("rows", []),
        "sortedColumn": spec.get("sortedColumn"),
        "name": spec.get("name", f"Pivot {formula_id}"),
    }
    return key, pivot


def _make_list(spec: dict, list_id: int) -> tuple[str, dict]:
    """Return (str_key, list_dict) from a blueprint list spec."""
    key = str(list_id)
    lst = {
        "id": key,
        "model": spec["model"],
        "domain": spec.get("domain", []),
        "context": spec.get("context", {}),
        "orderBy": spec.get("orderBy", []),
        "columns": spec.get("columns", []),
        "name": spec.get("name", f"List {list_id}"),
    }
    return key, lst


def _make_global_filter(spec: dict, pivot_map: dict[str, str], list_map: dict[str, str]) -> dict:
    """Build a global filter dict, wiring fields to all pivots and lists."""
    fields: dict[str, dict] = {}
    for pivot_name, pivot_uuid in pivot_map.items():
        field_spec = spec.get("fields", {}).get(pivot_name)
        if field_spec:
            fields[pivot_uuid] = field_spec
    for list_name, list_id in list_map.items():
        field_spec = spec.get("fields", {}).get(list_name)
        if field_spec:
            fields[list_id] = field_spec

    return {
        "id": spec.get("id", _new_uuid()),
        "type": spec.get("type", "relation"),
        "label": spec.get("label", ""),
        "modelName": spec.get("modelName", ""),
        "fields": fields,
        "defaultValue": spec.get("defaultValue", []),
        **({"dateRange": spec["dateRange"]} if "dateRange" in spec else {}),
    }


def _make_scorecard_figure(spec: dict, fig_id: str) -> dict:
    return {
        "id": fig_id,
        "tag": "chart",
        "width": spec.get("width", 240),
        "height": spec.get("height", 120),
        "col": spec.get("col", 0),
        "row": spec.get("row", 0),
        "offset": spec.get("offset", {"x": 0, "y": 0}),
        "data": {
            "type": "scorecard",
            "title": spec.get("title", {"text": "", "bold": True}),
            "keyValue": spec.get("keyValue", ""),
            "baseline": spec.get("baseline", ""),
            "baselineMode": spec.get("baselineMode", "difference"),
            "baselineColorUp": spec.get("baselineColorUp", "#00A04A"),
            "baselineColorDown": spec.get("baselineColorDown", "#E30000"),
            "background": spec.get("background", "#FFFFFF"),
            "humanize": spec.get("humanize", True),
            "chartId": fig_id,
        },
    }


def _make_chart_figure(spec: dict, fig_id: str) -> dict:
    return {
        "id": fig_id,
        "tag": "chart",
        "width": spec.get("width", 480),
        "height": spec.get("height", 280),
        "col": spec.get("col", 0),
        "row": spec.get("row", 4),
        "offset": spec.get("offset", {"x": 0, "y": 0}),
        "data": {
            "type": spec.get("chartType", "odoo_line"),
            "title": spec.get("title", ""),
            "legendPosition": spec.get("legendPosition", "top"),
            "stacked": spec.get("stacked", False),
            "cumulative": spec.get("cumulative", False),
            "odooDataSets": spec.get("odooDataSets", []),
            "searchParams": spec.get("searchParams", {"groupBy": [], "orderBy": [], "domain": []}),
            "chartId": fig_id,
        },
    }


def build(blueprint: dict) -> dict:
    name = blueprint.get("name", "Dashboard")
    version = blueprint.get("version", OSHEET_VERSION)

    # Build pivots
    pivot_specs = blueprint.get("pivots", [])
    pivots: dict[str, dict] = {}
    pivot_name_to_uuid: dict[str, str] = {}
    for i, spec in enumerate(pivot_specs, start=1):
        key, pivot = _make_pivot(spec, i)
        pivots[key] = pivot
        pivot_name_to_uuid[spec.get("name", str(i))] = key

    # Build lists
    list_specs = blueprint.get("lists", [])
    lists: dict[str, dict] = {}
    list_name_to_id: dict[str, str] = {}
    for i, spec in enumerate(list_specs, start=1):
        key, lst = _make_list(spec, i)
        lists[key] = lst
        list_name_to_id[spec.get("name", str(i))] = key

    # Build global filters
    filter_specs = blueprint.get("globalFilters", [])
    global_filters = [_make_global_filter(s, pivot_name_to_uuid, list_name_to_id) for s in filter_specs]

    # Build figures
    figure_specs = blueprint.get("figures", [])
    figures: list[dict] = []
    figure_ids: list[str] = []
    for i, spec in enumerate(figure_specs, start=1):
        fig_id = spec.get("id", f"figure_{i}")
        figure_ids.append(fig_id)
        fig_type = spec.get("type", "scorecard")
        if fig_type == "scorecard":
            figures.append(_make_scorecard_figure(spec, fig_id))
        else:
            figures.append(_make_chart_figure(spec, fig_id))

    # Sheets
    dashboard_sheet = {
        "id": _new_uuid(),
        "name": "Dashboard",
        "figures": figures,
        "cells": {},
        "merges": [],
        "cols": {},
        "rows": {},
        "conditionalFormats": [],
        "filterTables": [],
        "isVisible": True,
    }
    data_cells: dict[str, dict] = {}
    for cell_ref, content in blueprint.get("dataCells", {}).items():
        data_cells[cell_ref] = {"content": content}

    data_sheet = {
        "id": _new_uuid(),
        "name": "Data",
        "figures": [],
        "cells": data_cells,
        "merges": [],
        "cols": {},
        "rows": {},
        "conditionalFormats": [],
        "filterTables": [],
        "isVisible": blueprint.get("dataSheetVisible", False),
    }

    osheet = {
        "version": version,
        "revisionId": blueprint.get("revisionId", _new_uuid()),
        "uniqueFigureIds": figure_ids,
        "isNotSquishable": blueprint.get("isNotSquishable", True),
        "sheets": [dashboard_sheet, data_sheet],
        "styles": blueprint.get("styles", DEFAULT_STYLES),
        "formats": blueprint.get("formats", DEFAULT_FORMATS),
        "borders": blueprint.get("borders", DEFAULT_BORDERS),
        "settings": blueprint.get("settings", {"locale": {"name": "English (US)", "code": "en_US"}}),
        "customTableStyles": {},
        "pivots": pivots,
        "pivotNextId": len(pivots) + 1,
        "lists": lists,
        "listNextId": len(lists) + 1,
        "odooLinkReferences": blueprint.get("odooLinkReferences", {}),
        "globalFilters": global_filters,
    }
    return osheet


def main():
    parser = argparse.ArgumentParser(description="Build Odoo 19 .osheet from declarative blueprint.")
    parser.add_argument("blueprint", help="Blueprint JSON file")
    parser.add_argument("output", help="Output .osheet.json file")
    parser.add_argument("--pretty", action="store_true", help="Pretty-print output JSON")
    args = parser.parse_args()

    with open(args.blueprint, encoding="utf-8") as f:
        blueprint = json.load(f)

    osheet = build(blueprint)

    output = Path(args.output)
    output.parent.mkdir(parents=True, exist_ok=True)
    indent = 2 if args.pretty else None
    with open(output, "w", encoding="utf-8") as f:
        json.dump(osheet, f, indent=indent, ensure_ascii=False)

    print(f"Built: {output}")
    print(f"  Pivots:  {len(osheet['pivots'])}")
    print(f"  Lists:   {len(osheet['lists'])}")
    print(f"  Filters: {len(osheet['globalFilters'])}")
    print(f"  Figures: {len(osheet['uniqueFigureIds'])}")


if __name__ == "__main__":
    main()
