---
name: odoo-quote-calculator
description: >
  Abstracts Quote Calculators using Odoo Spreadsheet in v19.
  Trigger: Cuando el usuario pide crear, leer, actualizar, vincular, borrar un quote calculator, calculadora de cotización, o spreadsheet template de Odoo.
license: Apache-2.0
metadata:
  author: rd-mg
  version: "1.0"
allowed-tools: Read, Edit, Write, Bash, mcp-server-odoo
---

## Purpose
Estandarizar cómo crear, actualizar, leer o eliminar "Quote Calculators" (Odoo Spreadsheets) vinculados a Quotation Templates en Odoo 19 de una forma agnóstica a la configuración usando `mcp-server-odoo`.

## When to Use
- Cuando se pida crear un nuevo `sale.order.spreadsheet` (Quote Calculator).
- Cuando se pida actualizar calculadoras de cotización volumétricas o de servicios.
- Cuando se pida vincular un Excel de odoo a un `sale.order.template`.
- Do not use when: Editando listas de precios estándar que no usen la vista `owl` de Odoo Spreadsheet.

## Critical Patterns
- **No inventar el JSON**: El JSON para Odoo Spreadsheet v19 es masivo (~120KB) y estricto (requiere `searchParams`, `fieldMatching`, `globalFilters`). NO LO CONSTRUYAS DESDE CERO EN EL PROMPT.
- **Flujo de Ejecución Canónico**:
  1. Usa el archivo JSON canónico ubicado en `assets/jsons/odoo19_quotation_canonical.json`.
  2. Corre el adaptador en Python `assets/adapt_spreadsheet.py` localmente en la shell para customizar el título y exportarlo en Base64 SIN usar la red.
  3. Lee e inyecta el contenido generado (Base64) directo al comando de la tool `mcp_odoo_create_record`.

## Steps
1. Revisa la marca requerida por el usuario.
2. Ejecuta `python3 skills/40-odoo/odoo-quote-calculator/assets/adapt_spreadsheet.py --input skills/40-odoo/odoo-quote-calculator/assets/jsons/odoo19_quotation_canonical.json --output skills/40-odoo/odoo-quote-calculator/assets/jsons/deploy.txt --brand "Marca"`
3. Extrae la string Base64 devuelta: `PAYLOAD=$(cat skills/40-odoo/odoo-quote-calculator/assets/jsons/deploy.txt)` *(recuerda usar `read_file` local para ti)*.
4. Usa el tool `mcp_odoo_create_record` en el modelo `sale.order.spreadsheet` definiendo: `{"name": "...", "spreadsheet_binary_data": "<base64_recolectado>"}`.
5. Para enlazar a la cotización, usa `mcp_odoo_update_record` al modelo `sale.order.template` cambiando el integer en el campo `spreadsheet_template_id`.

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

Negative example (JSON inventado a mano con texto plano):
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
- Solo utiliza `spreadsheet_binary_data` en el tool MCP. Nunca escribas directamente un string al `spreadsheet_data` field a menos que estés testeando patches de hotfix.
- Verifica si la API rechaza el enlace M2O de `sale.order.template`; si el MCP API retorna integer List `[ID]` usa solo `ID`.

## Validation Checklist
- [ ] Base JSON exportado es v19 (`version: 19`).
- [ ] Extracción alfanumérica M2O respeta referencias de ID.
