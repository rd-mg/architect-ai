---
name: odoo-quote-calculator
description: >
  Abstracts Quote Calculators using Odoo Spreadsheet in v19.
  Trigger: When the user asks to create, read, update, link, or delete a quote calculator, quote calculator, or spreadsheet template of Odoo.
license: Apache-2.0
metadata:
  author: rd-mg
  version: "1.0"
allowed-tools: Read, Edit, Write, Bash, mcp-server-odoo
---

## Purpose
Standardize how to create, update, read, or delete "Quote Calculators" (Odoo Spreadsheets) linked to Quotation Templates in Odoo 19 in a configuration-agnostic way using `mcp-server-odoo`.

## When to Use
- When asked to create a new `sale.order.spreadsheet` (Quote Calculator).
- When asked to update volumetric or service quote calculators.
- When asked to link an Odoo Excel to a `sale.order.template`.
- Do not use when: Editing standard price lists that don't use Odoo Spreadsheet's `owl` view.

## Critical Patterns
- **Do not invent JSON**: JSON for Odoo Spreadsheet v19 is massive (~120KB) and strict (requires `searchParams`, `fieldMatching`, `globalFilters`). DO NOT BUILD IT FROM SCRATCH IN THE PROMPT.
- **Canonical Execution Flow**:
  1. Use the canonical JSON file located at `assets/jsons/odoo19_quotation_canonical.json`.
  2. Run the Python adapter `assets/adapt_spreadsheet.py` locally in shell to customize the title and export it as Base64 WITHOUT using the network.
  3. Read and inject the generated content (Base64) directly into the `mcp_odoo_create_record` tool command.

## Steps
1. Review the brand required by the user.
2. Execute `python3 skills/40-odoo/odoo-quote-calculator/assets/adapt_spreadsheet.py --input skills/40-odoo/odoo-quote-calculator/assets/jsons/odoo19_quotation_canonical.json --output skills/40-odoo/odoo-quote-calculator/assets/jsons/deploy.txt --brand "Brand"`
3. Extract the returned Base64 string: `PAYLOAD=$(cat skills/40-odoo/odoo-quote-calculator/assets/jsons/deploy.txt)` *(remember to use local `read_file` for yourself)*.
4. Use the `mcp_odoo_create_record` tool on the `sale.order.spreadsheet` model defining: `{"name": "...", "spreadsheet_binary_data": "<collected_base64>"}`.
5. To link to the quote, use `mcp_odoo_update_record` on the `sale.order.template` model changing the integer in the `spreadsheet_template_id` field.

## Code Examples
Positive example (Deploy agnóstico vía MCP):
```json
{
  "model": "sale.order.spreadsheet",
  "values": {
    "name": "Marca Logistics Calculator",
    "spreadsheet_binary_data": "eyJ2ZXJzaW..." // Extraido desde JSons/deploy.txt
  }
}
```

Negative example (hand-crafted JSON with plain text):
```json
{
  "model": "sale.order.spreadsheet",
  "values": {
    "spreadsheet_data": "{\"sheets\":[...], \"lists\":{...}}"
  }
}
```

## Commands
```bash
# Adapt and Generate Payload
python3 skills/40-odoo/odoo-quote-calculator/assets/adapt_spreadsheet.py \
  --input skills/40-odoo/odoo-quote-calculator/assets/jsons/odoo19_quotation_canonical.json \
  --output skills/40-odoo/odoo-quote-calculator/assets/jsons/payload.b64.txt \
  --brand "MyCustomBrand"
```

## Resources
- **Templates**: Canónico base Odoo19 ubicado en `assets/jsons/odoo19_quotation_canonical.json`
- **Adapter**: Script puente agnóstico a credenciales en `assets/adapt_spreadsheet.py`

## Guardrails
- Only use `spreadsheet_binary_data` in the MCP tool. Never directly write a string to the `spreadsheet_data` field unless you are testing hotfix patches.
- Verify if the API rejects the M2O link of `sale.order.template`; if the MCP API returns integer List `[ID]` use only `ID`.

## Checklist
- [ ] M2O alphanumeric extraction respects ID references.
