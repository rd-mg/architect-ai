# Feature: Accounting Formulas

## Odoo 19 Accounting Functions

Odoo-specific spreadsheet functions for querying journal entry aggregates.
Require `account.move.line` data access and respect Odoo domain and record-rule filtering.

| Formula | Purpose | Observed count (13 samples) |
|---------|---------|------|
| `ODOO.BALANCE(accountCode, dateRange, ...)` | Net balance for account | 81 |
| `ODOO.DEBIT(accountCode, dateRange, ...)` | Total debit for account | 6 |
| `ODOO.CREDIT(accountCode, dateRange, ...)` | Total credit for account | 6 |
| `ODOO.ACCOUNT.GROUP(groupCode, dateRange, ...)` | Account group aggregate | 40 |

## Common formula signatures

```
=ODOO.BALANCE("1100", TODAY()-365, TODAY())
=ODOO.BALANCE("4000", YEAR(TODAY())&"-01-01", YEAR(TODAY())&"-12-31")
=ODOO.ACCOUNT.GROUP("income", "this_year")
=ODOO.DEBIT("2000", "this_month")
=ODOO.CREDIT("2000", "this_month")
```

## Usage patterns

- Use `IFERROR(..., 0)` to handle accounts with no entries.
- Best placed in `Data` sheet and referenced by scorecard `keyValue`.
- `Accounting.osheet(1).json` sample uses 403 formulas, mostly `ODOO.BALANCE` and
  `ODOO.ACCOUNT.GROUP`, to power 5 scorecard figures with 0 pivots and 0 lists.
- Accounting dashboards commonly use `line` (spreadsheet-range) chart backed by grid of
  `ODOO.BALANCE` cells across months — not Odoo chart — because account groupings do not
  map cleanly to pivot `groupBy` semantics.

## Key rules

1. Do NOT use accounting functions in live cell on `Dashboard` sheet — back them in `Data`.
2. Always wrap in `IFERROR` — accounts without entries return error, not zero.
3. Respect Odoo date range API; do not pass raw timestamps without testing.
4. These functions execute under Odoo's current user record rules — no bypass possible or desired.

## Security note

`ODOO.BALANCE` and friends query `account.move.line`. User without accounting access gets
permission error at dashboard open time. Intentional Odoo behaviour.
Generate dashboards with accounting functions only for audiences with accounting read access,
or gate dashboard with appropriate access group in Odoo.
