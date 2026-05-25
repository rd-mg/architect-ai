---
name: odoo-spreadsheet-ecosystem
description: >
  Unified skill for ALL Odoo Spreadsheet tasks: dashboards (.osheet.json),
  quote calculators, XLSX manipulation, pivot-based reports.
  Routes internally to correct sub-guide based on intent.
  Zero new MCP servers — local scripts only.
bridge: false
on-demand: true
output: "A complete .osheet.json file OR modified .xlsx file"
---

## Routing Table [Execute FIRST — classify intent before anything else]

| Intent keywords | Sub-guide | Script |
|---|---|---|
| dashboard, KPI, scorecard, chart | `dashboard/GUIDE.md` | `odoo_sheet_tool.py build --type dashboard` |
| quote, pricing, margin, discount | `quote-calculator/GUIDE.md` | `odoo_sheet_tool.py build --type quote` |
| xlsx, Excel, OOXML, spreadsheet analysis | `xlsx-tools/GUIDE.md` | `odoo_sheet_tool.py xlsx --action {read\|validate}` |
| .osheet.json, pivot report, generic | `_shared/osheet-schema-notes.md` | `odoo_sheet_tool.py build --type generic` |

## Universal Rules (apply to ALL output types)
- All cell content MUST be string: `"B4": "2026"` not `"B4": 2026`
- All divisions protected: `=IFERROR(IF(B5=0,0,C5/B5),0)`
- No invented model fields — verify with rg before using
- No hardcoded team_id unless user requests
- Months Jan-Dec always visible (explicit PIVOT.VALUE per month)
- Scorecards point to summary cells, not isolated pivots
- Run validate after build: `python3 scripts/odoo_sheet_tool.py validate --file {output}`

## Standard Color Palette
```
#6C4E65  Primary header
#E4D9E1  Alternate rows
#FFFFFF  Text on dark
#D9EAD3  Positive
#FFF2CC  Warning
#F4CCCC  Alert
```

## Recovery Strategies [MANDATORY in every Odoo SKILL.md]
- Script fails → construct minimal valid JSON from `_shared/osheet-schema-notes.md`
- Validation fails → fix each issue one-by-one; re-validate after each
- JSON too large (>500KB) → simplify; note limitations to user
