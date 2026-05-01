# Workflow: Validate a Dashboard JSON Before Import

## Purpose

Catch structural errors in `.osheet.json` files offline, before importing into Odoo, to avoid
silent corruption or runtime import failures.

## Steps

1. **Run standard validation**:
   ```bash
   python3 scripts/osheet_validate.py file.osheet.json
   ```

2. **Run strict validation** (recommended before production import):
   ```bash
   python3 scripts/osheet_validate.py file.osheet.json --strict
   ```

3. **Validate a directory** and emit JSON report for CI:
   ```bash
   python3 scripts/osheet_validate.py assets/sample-osheets/*.json --json
   ```

4. **Review findings** by severity:
   - `CRITICAL` — must fix before import (missing required keys, malformed `odoo://view` JSON,
     scorecard without `keyValue`, chart without `data.type`).
   - `WARNING` — should fix (undefined style/format/border references, unknown top-level keys,
     mismatched `formulaId` counts).
   - `INFO` — informational only.

5. **Fix and re-validate** until zero CRITICAL errors remain.

## Validator coverage

| Check | Severity |
|-------|----------|
| Missing `version` key | CRITICAL |
| Missing `sheets` array | CRITICAL |
| Scorecard figure missing `keyValue` | CRITICAL |
| Chart figure missing `data.type` | CRITICAL |
| Malformed `odoo://view` JSON payload | CRITICAL |
| `odoo://view` missing `action.modelName` | CRITICAL |
| `odoo://view` missing `viewType` | CRITICAL |
| Pivot formula references undefined pivot (by key OR formulaId) | CRITICAL |
| List formula references undefined list (by key OR id) | CRITICAL |
| Undefined sheet-level `styleId` reference | WARNING |
| Undefined sheet-level `formatId` reference | WARNING |
| Undefined sheet-level `borderId` reference | WARNING |
| Unknown top-level key | WARNING |
| Global filter without `fields` matching | WARNING |

## Notes

- The validator resolves pivots by BOTH UUID dictionary key AND numeric `formulaId`.
  This is mandatory — all 13 reference samples use UUID keys but numeric `formulaId` in formulas.
- The validator does NOT connect to a live Odoo instance by default.
  Online model/field verification is available via `--odoo-url` flag (opt-in).
