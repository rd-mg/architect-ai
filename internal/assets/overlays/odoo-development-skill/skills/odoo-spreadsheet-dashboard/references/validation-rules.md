# Reference: Structural Validation Rules

## Severity levels

- `CRITICAL` — must fix before Odoo import; likely to cause import error or silent data corruption.
- `WARNING` — should fix; may cause runtime issues or confusing behaviour.
- `INFO` — informational; no immediate risk.

## Critical rules

| ID | Rule | Check |
|----|------|-------|
| V01 | `version` key must be present | `"version"` in top-level keys |
| V02 | `sheets` array must be present and non-empty | `len(doc["sheets"]) > 0` |
| V03 | Scorecard figure must have `keyValue` | `data.keyValue` present and non-empty |
| V04 | Chart figure must have `data.type` | `data.type` present and non-empty |
| V05 | `odoo://view` URL must have valid JSON payload | parse JSON after `odoo://view/` |
| V06 | `odoo://view` action must have `modelName` | `action.modelName` present in payload |
| V07 | `odoo://view` action must have `viewType` | `viewType` present in payload |
| V08 | Pivot formula reference must resolve | match `formulaId` OR UUID dict key |
| V09 | List formula reference must resolve | match numeric `id` OR dict key |

## Warning rules

| ID | Rule | Check |
|----|------|-------|
| W01 | Style index reference must be in-bounds | `0 <= styleId < len(styles)` |
| W02 | Format index reference must be in-bounds | `0 <= formatId < len(formats)` |
| W03 | Border index reference must be in-bounds | `0 <= borderId < len(borders)` |
| W04 | Unknown top-level key detected | key not in known key set |
| W05 | Global filter has no `fields` mapping | `filter.fields` empty or absent |
| W06 | `uniqueFigureIds` count != actual figure count | counts mismatch |
| W07 | `pivotNextId` < number of pivots | nextId not incremented after pivot add |
| W08 | `listNextId` < number of lists | nextId not incremented after list add |

## Critical implementation note: pivot formulaId resolution

**V08 must resolve pivots by BOTH forms**:
- UUID dictionary key (e.g. `"a3f2c1d0-..."`)
- Numeric `formulaId` value (e.g. `"6"`, `"10"`)

Odoo generates UUID keys but formulas use short numeric IDs. Validator that only checks
dictionary keys will INCORRECTLY reject valid dashboards (observed in POS Restaurant sample
with formulaIds 6, 7, 9, 10 and UUID keys).

```python
def _resolve_pivot(doc, ref):
    """Return True if ref matches any pivot by UUID key or formulaId."""
    for key, pivot in doc.get("pivots", {}).items():
        if ref == key or ref == str(pivot.get("formulaId", "")):
            return True
    return False
```

## Validator output format

```json
{
  "file": "Sales.osheet.json",
  "status": "PASS",
  "issues": [
    { "severity": "WARNING", "rule": "W04", "message": "Unknown top-level key: isNotSquishable" }
  ]
}
```

`status` is `PASS` if no CRITICAL issues, `FAIL` if any CRITICAL issue exists.

## Strict mode

`--strict` treats WARNING as CRITICAL. Use for pre-production validation.
