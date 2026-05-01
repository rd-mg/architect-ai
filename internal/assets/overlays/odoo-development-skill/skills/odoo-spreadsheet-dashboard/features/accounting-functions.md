# Feature: Accounting Formulas

## Odoo 19 Accounting Functions

These are Odoo-specific spreadsheet functions for querying journal entry aggregates.
They require `account.move.line` data access and respect Odoo domain and record-rule filtering.

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
- These functions are best placed in the `Data` sheet and referenced by scorecard `keyValue`.
- The `Accounting.osheet(1).json` sample uses 403 formulas, mostly `ODOO.BALANCE` and
  `ODOO.ACCOUNT.GROUP`, to power 5 scorecard figures with 0 pivots and 0 lists.
- Accounting dashboards commonly use a `line` (spreadsheet-range) chart backed by a grid of
  `ODOO.BALANCE` cells across months — not an Odoo chart — because account groupings do not
  map cleanly to pivot `groupBy` semantics.

## Key rules

1. Do NOT use accounting functions in a live cell on the `Dashboard` sheet — back them in `Data`.
2. Always wrap in `IFERROR` — accounts without entries return an error, not zero.
3. Respect the Odoo date range API; do not pass raw timestamps without testing.
4. These functions execute under Odoo's current user record rules — no bypass is possible or desired.

## Security note

`ODOO.BALANCE` and friends query `account.move.line`. A user without accounting access will get
a permission error at dashboard open time. This is intentional Odoo behaviour.
Generate dashboards with accounting functions only for audiences that have accounting read access,
or gate the dashboard with the appropriate access group in Odoo.
