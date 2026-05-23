# Business Logic & Design Patterns for Odoo 19 Quotation Calculators

To develop professional `Quotation Calculator` templates integrated directly with sales orders (`sale.order`) in Odoo 19's `o-spreadsheet` ecosystem, mandatory to comply with strict structural rules derived from financial modeling and enterprise software design.

This standard ensures templates exported and injected by Odoo do not suffer from data degradation and offer transparent auditing.

## 1. Structural and Visual Design Patterns (Semantic Rules)

In serious models and calculators (based on Big 4 and investment banking guidelines), spreadsheet is not decorative; appearance of each cell indicates its role within internal calculation logic.

### 1.1 Principle of Assumption Separation
Never combine hardcoded values with mathematical formulas.
- Always create isolated sheet/zone called **`Assumptions`**.
- This sheet manages variables (e.g., exchange rates, target margin rates, VAT percentage, max working hours capacity, scrap rates).
- Rest of workbook imports these values solely via cell references (e.g., `=Assumptions!$B$3`).

### 1.2 Financial Color-Coding Standards
Every cell rendered in `o-spreadsheet` JSON must map to following standardization in its `styleId`:

| Data Role | Font Color | RGB Representation | Primary Use |
| :--- | :--- | :--- | :--- |
| **Input / Hardcode** | <span style="color:blue">Blue</span> | `000000FF` / `#0000FF` | Manually adjustable values, financial assumptions (Margins, Rates). |
| **Formulas / Calculations** | **Black** | `00000000` / `#000000` | All cells whose content processes mathematics. |
| **Internal Cross-Reference** | <span style="color:green">Green</span> | `00008000` / `#008000` | Formulas fetching value from another sheet (`=Data!A1`). |
| **Immediate Attention** | <span style="color:blue">Blue (Yellow Fill)</span> | Fill: `#FFFF00` | Flags Odoo user MUST fill before approving budget. |

### 1.3 Format Codes (Type Standardization)

JSON defining calculator in Odoo (in `formats` section) must treat numbers according to strict accounting semantics:

- Zeros Format: `#,##0;(#,##0);"-"` (Any blank or zero value in sums displayed as dash to avoid visual noise).
- Multipliers and P/E Ratios: `0.0x`.
- Years and Categories: Must be formatted as strings (`"2024"`, not `2,024`) within `formatCode` assigned by Odoo.

## 2. Data Injection from Odoo (Pivot Fields)

Odoo `o-spreadsheet` engine performs "binding" via hidden injected sheet inside file (fully connected to context of current `sale.order`).

### Typically Injected `sale.order.line` Model:
1. `Product ID` and `Product Name`.
2. `Quantity`.
3. `Unit Price` and `Discount`.
4. `Salesperson` (for base commission distribution).
5. `Amount` (Original pre-approved amount before calculator manipulation).

### 2.1 Native Pivots (PIVOT Pattern)
Unlike flat Excel models, Odoo's `o-spreadsheet` supports formulas such as:
`=PIVOT("1")`

This formula (as detailed in engine core documentation) inserts array injected by Odoo and updates reactively via `model_id`. Any update in sales pipeline resonates without breaking math connected to these Pivots.

## 3. Formula Execution Safety (Critical Formulas)

To guarantee data consistency when Odoo asynchronously evaluates Quotation Document, always use these preventive abstractions:

1. **Mandatory `IFERROR` Protection:**
   Every cross-import must have fail-safes: `IFERROR(VLOOKUP(...), "-")`.
   Single uncalculated row returning `#N/A` breaks executive financial summaries.

2. **`SUMPRODUCT` for Complex Costing:**
   To calculate total prices based on system configurations, use:
   `=SUMPRODUCT(Cost[Materials], Quantities[Required])`
   This formula preserves JSON processing cycles much better than hidden additive columns.

3. **Strict Export Restriction:**
   If this calculator exported as traditional `xlsx` to client user, **all structural formulas must be validated** against compatibility standard. Odoo does not export 3D charts or non-semantic Excel decorations.
