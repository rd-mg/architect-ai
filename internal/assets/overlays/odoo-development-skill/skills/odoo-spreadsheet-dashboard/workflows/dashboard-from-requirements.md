# Workflow: Create a New Dashboard from Requirements

## Purpose

Transform user requirements into a valid Odoo 19 `.osheet` JSON dashboard.

## Steps

1. **Gather requirements** (ask if missing):
   - Target Odoo models (e.g. `sale.report`, `crm.lead`, `account.invoice.report`)
   - KPIs / scorecards needed
   - Charts needed (line, bar, pie; Odoo or spreadsheet-range)
   - Lists needed
   - Global filter fields (date, relation, text)
   - Colour scheme / theme
   - Dashboard name

2. **Author the blueprint** JSON (see `assets/blueprints/monthly-sales.json` for reference):
   ```json
   {
     "name": "Sales KPI",
     "version": "19.1.2",
     "sheets": [
       {
         "name": "Dashboard",
         "type": "dashboard"
       },
       {
         "name": "Data",
         "type": "data"
       }
     ],
     "pivots": [...],
     "lists": [...],
     "globalFilters": [...],
     "figures": [...]
   }
   ```

3. **Build the `.osheet` JSON** from the blueprint:
   ```bash
   python3 scripts/osheet_build.py blueprint.json output.osheet.json --pretty
   ```

4. **Validate the output**:
   ```bash
   python3 scripts/osheet_validate.py output.osheet.json --strict
   ```
   Fix any errors, then re-validate.

5. **Profile the output** to confirm the generated structure matches expectations:
   ```bash
   python3 scripts/osheet_profile.py output.osheet.json --markdown
   ```

6. **Deliver** the `.osheet.json` file to the user, ready for Odoo import via Documents or the
   spreadsheet REST endpoint.

## Key design rules

- Two-sheet layout: `Dashboard` (figures only) + `Data` (pivot formulas, list formulas).
- Scorecards must include `keyValue`, `baseline`, and `baselineMode`.
- Every pivot must have a `formulaId` that matches the formula references in `Data` cells.
- Global filters must declare `fields` matching for every pivot and list they control.
- Respect Odoo access groups: do not hard-code domain workarounds that bypass record rules.
