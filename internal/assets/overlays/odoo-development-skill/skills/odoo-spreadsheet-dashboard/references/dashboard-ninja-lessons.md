# Reference: Dashboard Ninja UX Lessons

## Purpose

Dashboard Ninja is Odoo App Store product for advanced dashboards. This reference extracts
its **UX and functional patterns** as inspiration ONLY.

> **Do not vendor Dashboard Ninja code, JavaScript, CSS, or custom SQL runtimes.**
> Map patterns to native Odoo o-spreadsheet capabilities only.

## Mapping: Dashboard Ninja concepts → Odoo native

| Dashboard Ninja concept | Odoo native equivalent |
|------------------------|------------------------|
| KPI tile | Scorecard figure (`type: scorecard`) |
| Graph item (line/bar/pie) | Odoo chart figure (`odoo_line`, `odoo_bar`, `odoo_pie`) |
| List view item | Odoo list data source + `ODOO.LIST` formulas |
| Date filter | Global filter (`type: date`) |
| Relation filter | Global filter (`type: relation`) |
| TV/website dashboard | Odoo spreadsheet shared link (read-only URL) |
| Theming / branding | Global style dictionaries + `background` in scorecard |
| Query dashboard | Custom SQL — explicit opt-in ONLY, separate security review |

## UX patterns to carry over

1. **KPI row first** — place all scorecards in horizontal row at top before charts.
2. **Current period vs. previous** — every KPI should have `baseline` comparison (MoM or YoY).
3. **Humanize large numbers** — use `FORMAT.LARGE.NUMBER` or `humanize: true`.
4. **Colour-coded delta** — green for positive, red for negative (`baselineColorUp/Down`).
5. **Clear labels** — use `=_t("Revenue")` for translatable titles.
6. **Filter-first UX** — place company + date filters at global filter bar; avoid per-pivot domain clutter.
7. **Drill-down links** — attach `odoo://view` links to scorecards so users can click through.
8. **Consistent chart sizing** — standardise on 480×280 for charts, 240×120 for scorecards.

## Anti-patterns (from Dashboard Ninja review)

- Do not use raw SQL by default — bypasses Odoo access control.
- Do not duplicate global filter conditions in every pivot domain — defeats filter mechanism.
- Do not put calculated grids on Dashboard sheet — keep on hidden Data sheet.
- Do not hard-code company IDs or user IDs in domains.
