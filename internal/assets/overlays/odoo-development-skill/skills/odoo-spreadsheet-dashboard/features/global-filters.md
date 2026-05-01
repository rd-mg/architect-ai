# Feature: Global Filters

## What are global filters?

Global filters are top-level `.osheet` objects that wire a single user-visible filter control
to multiple pivots, lists, and Odoo chart data sources simultaneously. They are the canonical
way to let a user filter an entire dashboard by period, company, salesperson, etc.

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
   controls. Use the UUID dictionary key as the top-level key in `fields`.
2. **Do not duplicate**: Do not also add the filter condition to a pivot's `domain`. The filter
   mechanism injects it at runtime.
3. **`dateRange` values**: Common values are `"this_year"`, `"this_quarter"`, `"this_month"`,
   `"last_year"`, or `null` for no default.
4. **Relation filters**: `modelName` is the Odoo model of the related record (not the source model).
5. **Validator warning**: A global filter without any `fields` matching is a WARNING — it exists
   in the UI but controls nothing.

## Design pattern

- Put company filter first (applies to all pivots).
- Put date filter second (most frequently used after company).
- Relation filters (Salesperson, Team) after date.
- Text filter last (rare; use only when free-text search adds value).
