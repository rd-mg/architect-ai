# Feature: Pivots and Lists

## Pivots

Pivots are Odoo data sources powering `PIVOT`, `PIVOT.VALUE`, and `PIVOT.HEADER` formulas.

### Pivot definition structure

```json
{
  "pivots": {
    "a3f2c1d0-...": {
      "id": "a3f2c1d0-...",
      "formulaId": "1",
      "model": "sale.report",
      "domain": [],
      "context": {},
      "measures": [
        { "id": "price_subtotal", "fieldName": "price_subtotal", "aggregator": "sum" }
      ],
      "columns": [],
      "rows": [
        { "fieldName": "date", "granularity": "month", "order": "asc" }
      ],
      "sortedColumn": null,
      "name": "Sales by Month"
    }
  }
}
```

### Critical: formulaId vs dictionary key

- **Dictionary key** is UUID (e.g. `"a3f2c1d0-..."`).
- `formulaId` is short numeric string (e.g. `"1"`, `"2"`).
- Formulas use `formulaId`: `=PIVOT.VALUE(1,"price_subtotal")`.
- Validator MUST accept both forms as valid pivot references.

### Pivot formulas

| Formula | Usage |
|---------|-------|
| `=PIVOT(formulaId)` | Full pivot table in range |
| `=PIVOT.VALUE(formulaId, "measure", "groupBy", value, ...)` | Single cell value |
| `=PIVOT.HEADER(formulaId, "groupBy", value, ...)` | Row/column header |

Wrap in `IFERROR` to handle empty data: `=IFERROR(PIVOT.VALUE(1,"revenue"),0)`.

## Lists

Lists are Odoo data sources powering `ODOO.LIST` and `ODOO.LIST.HEADER` formulas.

### List definition structure

```json
{
  "lists": {
    "1": {
      "id": "1",
      "model": "sale.order",
      "domain": [],
      "context": {},
      "orderBy": [{ "name": "date_order", "asc": false }],
      "columns": [
        { "name": "name", "type": "char" },
        { "name": "partner_id", "type": "many2one" },
        { "name": "amount_total", "type": "monetary" }
      ],
      "name": "Recent Orders"
    }
  }
}
```

### List formulas

| Formula | Usage |
|---------|-------|
| `=ODOO.LIST(listId, rowIndex, "fieldName")` | Single cell value |
| `=ODOO.LIST.HEADER(listId, "fieldName")` | Column header |

## Placement

- Define all pivot and list data sources in top-level `.osheet` `pivots` / `lists` keys.
- Place `PIVOT.VALUE`, `ODOO.LIST` formulas in `Data` sheet.
- Place `PIVOT`, `ODOO.LIST.HEADER` rendering ranges in `Data` sheet.
- Reference computed cells from `Dashboard` sheet only if needed for scorecard `keyValue`.
