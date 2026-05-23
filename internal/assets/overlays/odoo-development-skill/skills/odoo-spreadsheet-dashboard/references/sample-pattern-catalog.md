# Reference: Validated Sample Pattern Catalog

## Source

Profiled from 13 Odoo 19.1.2 `.osheet.json` samples on 2026-05-01.
All 13 samples passed `osheet_validate.py` and lossless `osheet_recipe.py` roundtrip.

## Profile table

| File | Sheets | Pivots | Lists | Filters | Figures | Formulas | Models | Chart types |
|------|-------:|-------:|------:|--------:|--------:|---------:|--------|-------------|
| Accounting | 2 | 0 | 0 | 1 | 5 | 403 | account.invoice.report | scorecard, odoo_line |
| Invoicing | 2 | 7 | 1 | 5 | 6 | 96 | account.invoice.report, account.move | scorecard, odoo_line |
| Leads | 2 | 14 | 0 | 9 | 7 | 58 | crm.lead | scorecard, odoo_line |
| MRR Evolution | 3 | 7 | 2 | 6 | 8 | 119 | account.move.line, sale.order.log.report | scorecard, odoo_line, odoo_bar |
| POS Restaurant | 2 | 8 | 0 | 4 | 11 | 129 | pos.order, report.pos.order | scorecard, combo |
| Pipeline | 2 | 9 | 1 | 8 | 8 | 108 | crm.lead | scorecard, odoo_bar |
| Point of Sale | 2 | 6 | 1 | 5 | 5 | 90 | pos.order, report.pos.order | scorecard, odoo_line |
| Product | 2 | 2 | 0 | 3 | 5 | 14 | sale.report | odoo_bar, scorecard |
| Rental | 2 | 6 | 0 | 5 | 5 | 39 | sale.rental.report | scorecard, odoo_line |
| Retention | 2 | 2 | 0 | 0 | 10 | 178 | sale.order.log.report | line, scorecard |
| Sales | 2 | 10 | 2 | 9 | 7 | 114 | sale.order, sale.report | scorecard, odoo_line |
| Salesperson | 2 | 12 | 4 | 4 | 5 | 174 | sale.order, sale.order.log.report, sale.report, sale.subscription.report | scorecard, odoo_bar |
| Subscriptions | 2 | 12 | 2 | 7 | 9 | 112 | sale.order, sale.order.log.report, sale.subscription.report | scorecard, odoo_pie |

## Aggregate findings

- **Files**: 13
- **All have**: version, sheets, styles, formats, borders, revisionId, uniqueFigureIds, settings, pivots, pivotNextId, customTableStyles, globalFilters, lists, listNextId, odooLinkReferences
- **12 of 13 have**: isNotSquishable (set to true)
- **Dominant figure**: scorecard (52 of 76 total figures)
- **Dominant formula**: ODOO.LIST (480), _T (308), PIVOT.VALUE (209)
- **Filter types**: relation (54), date (11), text (1)
- **Chart types**: scorecard (52), odoo_line (7), line (6), odoo_bar (5), odoo_pie (4), combo (2)

## Model coverage

| Domain | Odoo models |
|--------|-------------|
| Sales | sale.report, sale.order, sale.order.log.report |
| CRM | crm.lead |
| POS | pos.order, report.pos.order |
| Accounting | account.invoice.report, account.move, account.move.line |
| Subscriptions | sale.subscription.report |
| Rental | sale.rental.report |

## Key patterns observed

1. Every dashboard uses `Dashboard` + `Data` two-sheet layout.
2. MRR Evolution uses 3 sheets (adds `MRR` calculation sheet).
3. Accounting dashboard uses 403 formulas with 0 pivots — pure `ODOO.BALANCE` + `ODOO.ACCOUNT.GROUP`.
4. POS Restaurant uses `combo` chart type (only sample).
5. Retention dashboard uses spreadsheet-range `line` charts (local calc, no Odoo chart).
6. Sales dashboard (10 pivots, 2 lists, 9 filters) is most complex sample.
