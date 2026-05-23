---
name: odoo-spreadsheet-dashboard
trigger: "When user asks to create, edit, validate, recreate, or refactor Odoo 19 native dashboard spreadsheet (.osheet / .osps JSON). Triggers on 'odoo dashboard', 'osheet', 'scorecard', 'odoo pivot', 'PIVOT.VALUE', 'dashboard json'."
description: >
  Design, generate, validate, recreate, and refactor Odoo 19 native dashboard spreadsheet
  files (.osheet / .osps JSON). Use when user asks to create new Odoo 19 dashboard,
  work with o-spreadsheet JSON, validate .osheet files, recreate sample dashboards for
  regression tests, build scorecards / pivots / charts / global filters, or inspect existing
  dashboard JSON files.
license: MIT
metadata:
  version: "1.0"
  category: odoo-productivity
  odoo_version: "19.x"
  sources:
    - Odoo 19 o-spreadsheet documentation
    - Odoo 19 spreadsheet module source (addons/spreadsheet)
    - 13 validated Odoo 19 .osheet JSON samples (profiled 2026-05-01)
---

# Odoo 19 Spreadsheet Dashboard Skill

Handle request directly. Do NOT spawn sub-agents. Always write output file user requests.

> **Scope guard**: This skill covers NATIVE Odoo dashboard JSON semantics.
> XLSX import/export and generic spreadsheet manipulation remain in `odoo-minimax-xlsx-o-spreadsheets`.
> Do NOT merge two concerns unless user explicitly asks for XLSX round-trip.

## Task Routing

| Task | Workflow / Script |
|------|-------------------|
| Inspect / profile existing .osheet file | `workflows/profile-and-analyze.md` + `scripts/osheet_profile.py` |
| Design and create new dashboard | `workflows/dashboard-from-requirements.md` + `scripts/osheet_build.py` |
| Recreate or regression-test sample | `workflows/recreate-sample-json.md` + `scripts/osheet_recipe.py` |
| Refactor / upgrade existing dashboard JSON | `workflows/refactor-dashboard-json.md` |
| Validate before import | `workflows/validate-dashboard-json.md` + `scripts/osheet_validate.py` |

## Core concepts (load on demand)

| Area | Feature doc |
|------|-------------|
| Scorecards (KPI tiles) | `features/scorecards.md` |
| Odoo charts (line / bar / pie) | `features/odoo-charts.md` |
| Pivots and list data sources | `features/pivots-lists.md` |
| Global filters | `features/global-filters.md` |
| Accounting formulas | `features/accounting-functions.md` |
| Layout and theme | `features/layout-and-theme.md` |

## Reference materials (load on demand)

| Reference | File |
|-----------|------|
| .osheet top-level data model | `references/osheet-data-model-v19.md` |
| Dashboard semantics and patterns | `references/odoo-dashboard-semantics.md` |
| Odoo formula functions (v19) | `references/odoo-functions-v19.md` |
| Dashboard Ninja UX lessons | `references/dashboard-ninja-lessons.md` |
| Validated sample pattern catalog | `references/sample-pattern-catalog.md` |
| Structural validation rules | `references/validation-rules.md` |

## Scripts

```bash
# Profile existing dashboard
python3 scripts/osheet_profile.py Sales.osheet.json --markdown

# Validate before import
python3 scripts/osheet_validate.py Sales.osheet.json --strict

# Lossless recipe export → rebuild → compare
python3 scripts/osheet_recipe.py export Sales.osheet.json Sales.recipe.json
python3 scripts/osheet_recipe.py build  Sales.recipe.json Sales.recreated.osheet.json
python3 scripts/osheet_compare.py Sales.osheet.json Sales.recreated.osheet.json --mode exact

# Build new dashboard from declarative blueprint
python3 scripts/osheet_build.py assets/blueprints/monthly-sales.json monthly-sales.osheet.json --pretty
python3 scripts/osheet_validate.py monthly-sales.osheet.json --strict
```

## Key rules

1. **Pivot formulaId**: Pivot formulas reference `formulaId` (e.g. `1`, `2`), not UUID dictionary key. Validator MUST accept both.
2. **List id**: List formulas reference numeric `id`, not dictionary key.
3. **Global filters**: Preserve `fields` matching across pivots, lists, chart data sources. Do not duplicate per-pivot domain conditions.
4. **Scorecard first**: Scorecards are dominant figure type. Treat as first-class, not generic chart fallbacks.
5. **Dashboard + Data layout**: Use two-sheet pattern by default — `Dashboard` (presentation) + `Data` (backing formulas, hidden areas).
6. **Safety**: Generate native Odoo data sources. Let Odoo enforce runtime access rights and record rules. Do not generate raw SQL dashboards by default.
7. **Unknown keys**: Preserve unknown top-level keys; warn, do not delete.

## Odoo Research Priority [MANDATORY]

All research query flows MUST respect Local-First Fallback Chain:
1. Engram: `mem_search("odoo ${ODOO_VERSION} <topic>")`
2. rg in Local Workspace (`${ODOO_COMMUNITY}/addons/`, etc.)
3. Context7 MCP: `context7.resolve_library_id("odoo")`
4. researcher agent: `scope_hint="docs"`, `max_depth="standard"`
5. Web Search (Google/GitHub): ONLY if all local sources exhausted or fail.

## Recovery Strategies [MANDATORY in every Odoo SKILL.md]
- Pivot formulaId mismatch → fallback to checking UUID key mapping in pivot dictionary
- Chart global filter breakage → verify `fieldMatching` specifies matching field for each data source
- Workspace source missing → use Engram knowledge nodes and Context7 docs
