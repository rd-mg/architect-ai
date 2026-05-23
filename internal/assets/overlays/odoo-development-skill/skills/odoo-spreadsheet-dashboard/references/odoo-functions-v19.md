# Reference: Odoo 19 Formula Functions

## Inventory from 13 validated samples

| Function | Count | Category |
|----------|------:|----------|
| `ODOO.LIST` | 480 | List data |
| `_T` / `=_t(...)` | 308 | Translation |
| `PIVOT.VALUE` | 209 | Pivot data |
| `IFERROR` | 115 | Error handling |
| `FORMAT.LARGE.NUMBER` | 107 | Number display |
| `ODOO.BALANCE` | 81 | Accounting |
| `PIVOT` | 57 | Pivot table |
| `ODOO.LIST.HEADER` | 48 | List header |
| `ODOO.ACCOUNT.GROUP` | 40 | Accounting group |
| `YEAR` | 12 | Date |
| `PIVOT.HEADER` | 12 | Pivot header |
| `LEFT`, `RIGHT` | 10 each | String |
| `CHOOSECOLS` | 10 | Array |
| `CONCATENATE` | 9 | String |
| `ODOO.DEBIT` | 6 | Accounting |
| `ODOO.CREDIT` | 6 | Accounting |
| `ROUND` | 5 | Math |
| `CONTRACTION`, `EXPANSION` | 5 each | Custom metric |

## Odoo-specific functions

### Pivot functions

```
=PIVOT(formulaId)
  → Inserts full pivot table starting at this cell

=PIVOT.VALUE(formulaId, "measure" [, "groupBy", value, ...])
  → Returns single aggregated value

=PIVOT.HEADER(formulaId, "groupBy", value [, ...])
  → Returns pivot row/column header label
```

`formulaId` is numeric string from pivot's `formulaId` field, NOT UUID key.

### List functions

```
=ODOO.LIST(listId, rowIndex, "fieldName")
  → Returns single field value from list row rowIndex (1-based)

=ODOO.LIST.HEADER(listId, "fieldName")
  → Returns display label for list column
```

### Accounting functions

```
=ODOO.BALANCE("accountCode", fromDate, toDate [, options])
=ODOO.DEBIT("accountCode", fromDate, toDate [, options])
=ODOO.CREDIT("accountCode", fromDate, toDate [, options])
=ODOO.ACCOUNT.GROUP("groupCode", fromDate, toDate [, options])
```

### Helper / utility functions

```
=_t("Label text")                   → Translatable string literal
=FORMAT.LARGE.NUMBER(value)         → Format as K / M / B
=IFERROR(expr, fallback)            → Return fallback on any error
=CONTRACTION(value1, value2)        → Month-over-month decrease (custom)
=EXPANSION(value1, value2)          → Month-over-month increase (custom)
```

Note: `CONTRACTION` and `EXPANSION` are Odoo-specific functions for subscription MRR metrics.

## Standard functions used in dashboard context

- `YEAR(date)` — extract year from date cell
- `LEFT(text, n)`, `RIGHT(text, n)` — string truncation for labels
- `CONCATENATE(a, b, ...)` — label construction
- `CHOOSECOLS(array, col1, col2, ...)` — select columns from PIVOT-expanded range
- `ROUND(value, n)` — rounding before display

## Best practices

1. Always wrap `PIVOT.VALUE` and `ODOO.LIST` in `IFERROR(..., 0)` or `IFERROR(..., "")`.
2. Use `=_t("...")` for any user-visible string that may need translation.
3. Use `FORMAT.LARGE.NUMBER` in `Data` sheet; reference cell from scorecard `keyValue`.
4. Never hard-code number in scorecard `keyValue` — always use formula.
