# Feature: Global Filters

## What are global filters?

Global filters are top-level `.osheet` objects wiring single user-visible filter control
to multiple pivots, lists, and Odoo chart data sources simultaneously. Canonical
way to let user filter entire dashboard by period, company, salesperson, etc.

## Filter types observed in samples

| Type | Count | Example |
|------|------:|---------|
| `relation` | 54 | Company, Salesperson, Team, Product |
| `date` | 11 | Order Date, Invoice Date, Close Date |
| `text` | 1 | Free-text search |

## Global filter definition structure

```json
{
  "globalFilters": [
    {
      "id": "filter_company",
      "type": "relation",
      "label": "Company",
      "modelName": "res.company",
      "fields": {
        "pivot_uuid_1": { "field": "company_id", "type": "many2one" },
        "pivot_uuid_2": { "field": "company_id", "type": "many2one" },
        "list_id_1":   { "field": "company_id", "type": "many2one" }
      },
      "defaultValue": []
    },
    {
      "id": "filter_date",
      "type": "date",
      "label": "Order Date",
      "dateRange": "this_year",
      "fields": {
        "pivot_uuid_1": { "field": "date", "type": "date" }
      },
      "defaultValue": null
    }
  ]
}
```

## Key rules

1. **`fields` mapping**: Each filter must declare which field to match for every pivot/list it
   controls. Use UUID dictionary key as top-level key in `fields`.
2. **Do not duplicate**: Do not also add filter condition to pivot's `domain`. Filter
   mechanism injects it at runtime.
3. **`dateRange` values**: Common values `"this_year"`, `"this_quarter"`, `"this_month"`,
   `"last_year"`, or `null` for no default.
4. **Relation filters**: `modelName` is Odoo model of related record (not source model).
5. **Validator warning**: Global filter without any `fields` matching is WARNING — exists
   in UI but controls nothing.

## Design pattern

- Put company filter first (applies to all pivots).
- Put date filter second (most frequently used after company).
- Relation filters (Salesperson, Team) after date.
- Text filter last (rare; use only when free-text search adds value).
