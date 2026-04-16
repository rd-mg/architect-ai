#!/usr/bin/env python3
"""
Odoo Quote Calculator Adapter
Reads the canonical Odoo 19 Spreadsheet JSON payload, allows rebranding,
injects dynamic queries for local database products to replace hardcoded data,
and outputs the Base64 file that agents should load into MCP Tool Payload.

Usage:
  python3 adapt_spreadsheet.py --input jsons/odoo19_quotation_canonical.json --output jsons/payload.b64.txt --brand "My Brand"
"""
import json
import base64
import argparse
import sys
import re

def update_formula_references(data, old_name, new_name):
    """
    Search and replace all references to old_name with new_name in all formulas.
    Handles 'Sheet Name'! references.
    """
    # Spreadsheet formulas use 'SheetName'!A1 or SheetName!A1
    # We pattern match for both.
    pattern = re.compile(f"(['\"]?){re.escape(old_name)}(['\"]?)!")
    replacement = f"\\1{new_name}\\2!"
    
    for sheet in data.get("sheets", []):
        cells = sheet.get("cells", {})
        for cell_id, cell in cells.items():
            content = cell.get("content", "")
            if isinstance(content, str) and content.startswith("="):
                cell["content"] = pattern.sub(replacement, content)
                
        # Also update charts/figures
        for fig in sheet.get("figures", []):
            fig_data = fig.get("data", {})
            # Charts usually have dataSets with dataRange
            for ds in fig_data.get("dataSets", []):
                dr = ds.get("dataRange", "")
                if dr:
                    ds["dataRange"] = pattern.sub(replacement, dr)
            # Gauges/Scorecards have dataRange or keyValue
            if "dataRange" in fig_data:
                 fig_data["dataRange"] = pattern.sub(replacement, fig_data["dataRange"])
            if "keyValue" in fig_data:
                 fig_data["keyValue"] = pattern.sub(replacement, fig_data["keyValue"])
            if "labelRange" in fig_data:
                 fig_data["labelRange"] = pattern.sub(replacement, fig_data["labelRange"])

def adapt_spreadsheet(input_path, output_path, brand_name):
    """
    Main adaptation logic.
    """
    # Load canonical base
    try:
        with open(input_path, 'r') as f:
            data = json.load(f)
    except FileNotFoundError:
        print(f"Error: {input_path} no encontrado.")
        sys.exit(1)

    # 1. Update Sheet Names & Header Branding

    for sheet in data.get("sheets", []):
        old_name = sheet.get("name")
        if old_name == "Transport & Assembly":
            new_name = f"{brand_name} Logistics"
            sheet["name"] = new_name
            # Update all references to this sheet in the whole workbook
            update_formula_references(data, old_name, new_name)
            
            if "A1" in sheet.get("cells", {}):
                sheet["cells"]["A1"]["content"] = f"{brand_name} Logistics Calculation"
            
            # Rebrand Carpenter to Welder in this sheet
            for cid, cell in sheet.get("cells", {}).items():
                content = cell.get("content", "")
                if isinstance(content, str):
                    if "Carpenter" in content:
                        cell["content"] = content.replace("Carpenter", "Welder")
                        
        elif old_name == "Instructions":
            # Instructions doesn't usually have external references, but we play safe
            new_name = "Instructions" # Keep consistent or change if needed
            if "A1" in sheet.get("cells", {}):
                sheet["cells"]["A1"]["content"] = f"{brand_name} Template Instructions"
            # Rebrand Carpenter to Welder in instructions
            for cid, cell in sheet.get("cells", {}).items():
                content = cell.get("content", "")
                if isinstance(content, str) and "Carpenter" in content:
                    cell["content"] = content.replace("Carpenter", "Welder")
    
    # 2. Update Charts
    for sheet in data.get("sheets", []):
        for fig in sheet.get("figures", []):
            fig_data = fig.get("data", {})
            title = fig_data.get("title", {})
            if isinstance(title, dict):
                text = title.get("text", "")
                if "Cost repartition" in text:
                    title["text"] = f"{brand_name} Cost Distribution"

    # 3. Inject Dynamic Database Products (fixes hardcoded demo data)
    if "lists" not in data:
        data["lists"] = {}

    # Update List 1 (Sale Order Lines) to include synchronization columns (write-back)
    if "1" in data["lists"]:
        data["lists"]["1"]["columns"] = ["product_id", "product_uom_qty", "price_unit", "discount", "price_subtotal"]
        # Inject fieldMatching for price_unit to allow write-back
        if "fieldMatching" not in data["lists"]["1"]:
            data["lists"]["1"]["fieldMatching"] = {}
        data["lists"]["1"]["fieldMatching"]["price_unit"] = {
            "chain": "price_unit",
            "type": "number"
        }

    data["lists"]["2"] = {
        "model": "product.product",
        "domain": [["sale_ok", "=", True]],
        "orderBy": [],
        "context": {},
        "columns": ["display_name", "list_price", "standard_price", "uom_id", "volume", "weight", "default_code"],
        "name": "Database Products",
        "fieldMatching": {}
    }

    # Ensure Odoo version metadata is correctly set for v19
    data["odooVersion"] = 19

    logistics_sheet = f"{brand_name} Logistics"
    
    for sheet in data.get("sheets", []):
        if sheet.get("name") == "Products":
            # Hardcoded products cleanup, replace with dynamic ODOO.LIST fetches
            sheet["cells"] = {
                "A2": {"style": 17, "content": "Name", "border": 16},
                "B2": {"style": 17, "content": "Sales Price", "border": 16},
                "C2": {"style": 17, "content": "Cost", "border": 16},
                "D2": {"style": 17, "content": "Unit of Measure", "border": 16},
                "E2": {"style": 17, "content": "Volume (m3)", "border": 16},
                "F2": {"style": 17, "content": "Weight", "border": 16},
                "G2": {"style": 17, "content": "Internal Ref", "border": 16},
            }
            # Generate formulas for up to 150 products dynamically
            for i in range(1, 151):
                r = i + 2
                sheet["cells"][f"A{r}"] = {"content": f'=ODOO.LIST(2,{i},"display_name")'}
                sheet["cells"][f"B{r}"] = {"content": f'=ODOO.LIST(2,{i},"list_price")'}
                sheet["cells"][f"C{r}"] = {"content": f'=ODOO.LIST(2,{i},"standard_price")'}
                sheet["cells"][f"D{r}"] = {"content": f'=ODOO.LIST(2,{i},"uom_id")'}
                sheet["cells"][f"E{r}"] = {"content": f'=ODOO.LIST(2,{i},"volume")'}
                sheet["cells"][f"F{r}"] = {"content": f'=ODOO.LIST(2,{i},"weight")'}
                sheet["cells"][f"G{r}"] = {"content": f'=ODOO.LIST(2,{i},"default_code")'}
                
        elif sheet.get("name") == "SO Lines":
            # Update VLOOKUP bounds to cover the new dynamic product list range
            for cid, cell in list(sheet.get("cells", {}).items()):
                content = cell.get("content", "")
                if not isinstance(content, str):
                    continue
                    
                if "Products!$A$2:$G$57" in content:
                    cell["content"] = content.replace("Products!$A$2:$G$57", "Products!$A$2:$G$152")
                
                # Dynamic Price injection via conditional logic
                # Target column D cells that contain price_unit formulas
                if cid.startswith("D") and cid != "D1":
                    if 'ODOO.LIST(1,' in content and '"price_unit"' in content:
                        row = cid[1:]
                        logistics_ref = f"'{logistics_sheet}'"
                        
                        # Addition 1: Fabrication Cost (Apprentice B17 + Welder B23)
                        # Addition 2: Delivery Cost (Transport B10)
                        fab_cond = f'IF(A{row}="[SRV-FAB-002] Plastic Fabrication", {logistics_ref}!B17+{logistics_ref}!B23, 0)'
                        del_cond = f'IF(A{row}="[LOG-DEL] Delivery Fee (Smart Calculation)", {logistics_ref}!B10, 0)'
                        
                        cell["content"] = f"{content}+{fab_cond}+{del_cond}"

    # 4. Base64 Encode
    json_str = json.dumps(data)
    b64_data = base64.b64encode(json_str.encode()).decode('utf-8')

    # Save to payload text file
    with open(output_path, 'w') as f:
        f.write(b64_data)
        
    print(f"Success! Base64 payload generated at: {output_path}")
    print("Agent Instruction: Read the content and use it inside mcp_odoo_create_record inside the 'spreadsheet_binary_data' key.")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Adapt Odoo 19 Quote Calculator JSON")
    parser.add_argument("--input", required=True, help="Path to input JSON")
    parser.add_argument("--output", required=True, help="Path to output Base64 string")
    parser.add_argument("--brand", default="Custom", help="Brand name to inject into the calculator")
    
    args = parser.parse_args()
    adapt_spreadsheet(args.input, args.output, args.brand)
