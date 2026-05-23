# Feature: Scorecards (KPI Tiles)

## What is a scorecard?

Scorecard is Odoo 19 dashboard figure with `"tag": "chart"` and `"data": { "type": "scorecard" }`.
Displays single metric with optional baseline comparison — primary KPI tile in every
dashboard sample analysed (52 scorecards across 13 samples).

## Scorecard figure structure

```json
{
  "id": "figure_1",
  "width": 240,
  "height": 120,
  "tag": "chart",
  "col": 0,
  "row": 0,
  "offset": { "x": 0, "y": 0 },
  "data": {
    "type": "scorecard",
    "title": { "text": "Revenue", "bold": true },
    "keyValue": "=PIVOT.VALUE(1,\"revenue\")",
    "baseline": "=PIVOT.VALUE(2,\"revenue\")",
    "baselineMode": "difference",
    "baselineColorUp": "#00A04A",
    "baselineColorDown": "#E30000",
    "background": "#FFFFFF",
    "humanize": true,
    "chartId": "figure_1"
  }
}
```

## Required fields

| Field | Type | Description |
|-------|------|-------------|
| `type` | `"scorecard"` | Must be literal `scorecard` |
| `keyValue` | formula or value | KPI metric (validator: CRITICAL if missing) |
| `baseline` | formula or value | Comparison value (optional but recommended) |
| `baselineMode` | `"difference"` \| `"percentage"` | How delta shown |
| `baselineColorUp` | hex string | Colour when delta positive |
| `baselineColorDown` | hex string | Colour when delta negative |
| `background` | hex string | Tile background colour |
| `humanize` | boolean | Format large numbers (1M, 500K) |

## Common patterns

- `keyValue` = `PIVOT.VALUE(formulaId, "measure_field")` or `IFERROR(PIVOT.VALUE(...), 0)`
- `baseline` = previous-period pivot value or different measure
- `title.text` = `=_t("Revenue")` for translatable labels
- Use `FORMAT.LARGE.NUMBER` in backing Data cell and reference it from `keyValue`

## Design rules

1. Place scorecards only on `Dashboard` sheet.
2. Back complex calculations in `Data` sheet; reference with `=Data.A1`.
3. Always set `baselineColorUp` and `baselineColorDown` — do not leave defaults.
4. Group related scorecards in row before charts.
