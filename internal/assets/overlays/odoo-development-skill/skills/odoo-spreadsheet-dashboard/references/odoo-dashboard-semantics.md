# Reference: Odoo 19 Dashboard Semantics

## What is an Odoo 19 spreadsheet dashboard?

In Odoo 19, a dashboard is a native `.osheet` JSON document stored in `spreadsheet.spreadsheet`
(or via `documents.document`) and rendered in the browser using the `o-spreadsheet` library.
It is NOT an embedded HTML component or a Dashboard Ninja dashboard — it is a first-class
spreadsheet with Odoo-specific formula extensions.

## Dashboard vs. Data sheet pattern

| Sheet | Purpose |
|-------|---------|
| `Dashboard` | User-visible figures: scorecards, charts |
| `Data` | Backing calculations: PIVOT.VALUE, ODOO.LIST, ODOO.BALANCE grids |

The Data sheet is typically hidden (`isVisible: false`) in production.

## Odoo data insertion model

Odoo 19 supports three ways to insert live data into a spreadsheet:
1. **Pivots** — aggregate data from any Odoo model using `PIVOT.VALUE`, `PIVOT.HEADER`, `PIVOT`.
2. **Lists** — row-by-row records from any Odoo model using `ODOO.LIST`, `ODOO.LIST.HEADER`.
3. **Accounting functions** — `ODOO.BALANCE`, `ODOO.DEBIT`, `ODOO.CREDIT`, `ODOO.ACCOUNT.GROUP`.

Charts on top of pivot/list data are the standard composition. Do not query custom SQL by default.

## Dashboard conversion

When a user "converts a spreadsheet to a dashboard" in Odoo UI:
- The sheet view changes to `dashboard` mode.
- Figures become the primary content; cell formulas are hidden from direct edit.
- Global filters appear in the top bar.
- The `isNotSquishable` flag is typically set to prevent column collapse.

## Drill-down links (`odoo://view`)

Scorecard and chart figures can include `odoo://view/{json}` URIs that open Odoo views:

```
odoo://view/{"model":"sale.order","viewType":"list","action":{"modelName":"sale.order","domain":[],"viewType":"list"}}
```

Validator checks: valid JSON payload, `action.modelName` present, `viewType` present.

## Access and sharing

- Dashboards stored in `documents.document` inherit folder-level access rules.
- Dashboards stored in `spreadsheet.spreadsheet` use `spreadsheet` access groups.
- Pivots and lists execute under the CURRENT USER's Odoo rights — no privilege escalation.
- If a user lacks access to `account.move.line`, `ODOO.BALANCE` returns an error.

## Odoo 19 formula execution model

Formulas are evaluated client-side by `o-spreadsheet`. Odoo functions (`PIVOT.VALUE`,
`ODOO.LIST`, `ODOO.BALANCE`) make RPC calls to the server. Formula results are cached per
spreadsheet session and refreshed on user action or global filter change.

## What this skill does NOT cover

- Generating custom SQL dashboards (requires explicit opt-in + security review).
- Dashboard Ninja proprietary features (tile subscriptions, external data, advanced queries).
- XLSX round-trips (use `odoo-minimax-xlsx-o-spreadsheets` skill).
- Odoo Studio report views.
