---
name: odoo-spreadsheet-ecosystem
description: "Unified router and ecosystem for ALL Odoo spreadsheet tasks: Dashboards (.osps), quote calculators, XLSX creation, and direct XML edits."
bridge: false
on-demand: true
risk_level: high
---

# Odoo Spreadsheet Ecosystem Skill [UNIFIED ROUTER]

This is the primary router and source of truth for all Odoo-related spreadsheet tasks. Do NOT spawn sub-agents. Handle all tasks through this ecosystem.

## Task Routing Table

| Intent | Target Sub-Guide | Execution Script |
|--------|------------------|------------------|
| **Native Dashboard** (.osheet/.osps JSON) | `features/pivots-lists.md` | `python3 scripts/osheet_build.py` |
| **Strict XLSX Import** (Formulas & Colors) | `references/create.md` | `python3 scripts/xlsx_pack.py` |
| **XLSX Direct Editing** (Insert/Shift/Append) | `references/edit.md` | `python3 scripts/xlsx_insert_row.py` |
| **Quote Calculator** (.osps Base64) | `references/quotes.md` | `python3 scripts/adapt_spreadsheet.py` |

## Universal Rules

1. **Formula-First**: Every calculated cell MUST use an Excel formula (`<f>SUM(...)</f>`) or Odoo JSON `=formula`. Never write hardcoded values for computed data.
2. **No openpyxl Round-Trip**: Editing spreadsheets must be done using unpack → XML direct-edit → repack workflow to prevent losing VBA, macros, or pivot linkages.
3. **Strict Linter Audit**: Always validate outputs via `formula_check.py` or `osheet_validate.py` before presenting results to the user.

## Standard Color Palette

| Cell Role | Font Color | Hex Code |
|-----------|-----------|----------|
| **Hard-coded Input / Assumption** | Blue | `0000FF` |
| **Formula / Computed Result** | Black | `000000` |
| **Cross-sheet Reference Formula** | Green | `00B050` |

## Odoo Research Priority [MANDATORY]

All research query flows MUST respect the Local-First Fallback Chain:
1. Engram: `mem_search("odoo ${ODOO_VERSION} <topic>")`
2. rg in Local Workspace (`${ODOO_COMMUNITY}/addons/`, etc.)
3. Context7 MCP: `context7.resolve_library_id("odoo")`
4. researcher agent: `scope_hint="docs"`, `max_depth="standard"`
5. Web Search (Google/GitHub): ONLY if all local sources are exhausted or fail.

## Recovery Strategies [MANDATORY in every Odoo SKILL.md]
- **XML Ordering Breakage**: Use standard Python `xml.etree.ElementTree` serialization or manual regex patching to preserve original tag order and attributes.
- **Chart/Global Filter Mismatch**: Inspect pivot/list `formulaId` linkages and ensure chart fields match global filter domain keys.
- **Workspace Source Missing**: fallback to Engram knowledge nodes and Context7 docs.
