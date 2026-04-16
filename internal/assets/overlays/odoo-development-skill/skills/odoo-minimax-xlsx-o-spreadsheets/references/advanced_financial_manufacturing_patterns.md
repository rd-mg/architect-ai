# Advanced Functional & Manufacturing Patterns

When connecting an `o-spreadsheet` file (Quotation Calculator) to the ERP (Odoo v19), the quotation calculator must evolve beyond the simple (Price x Quantity) = Total. It must consider the transactional impact on other key system apps, primarily manufacturing (`mrp`) and accounting (`account`).

Below, we document the technical breakdown of an *Advanced Quotation Calculator*.

---

## 1. Integration Scope: Sales

This section receives base data and logically interconnects it.

- **Base Revenue Mapping (`Revenue`):**
  An advanced model does not assume the final `Unit Price`. It acts as a bridge via a formula that queries the injected Odoo database to search for hidden price list rules (`pricelist_id` rules).
  
  *Standard Applied Formula:*
  `=IFERROR(VLOOKUP([ProductID], Data!$A$2:$F$100, [PriceColumn], FALSE), 0)`

- **Dynamic Tax and Discount Reallocation:**
  By identifying the original `Discount` column from the quotation, the calculator exposes a *Net Income* scenario.
  *(List Price - Discount) + Specific region Tax.*

---

## 2. Integration Scope: Manufacturing & Costing

If the product is marked with the MTO (Make To Order) or *Manufacture* route, the salesperson needs to preview the economic risk due to industrial capacity constraints.

### 2.1 Bill Of Materials (BOM) Modeling
- **Volume Ratio:** The volume demanded in the initial input is linked with the resources required to dispatch.
- **Direct Material Cost:** Internally linked to the injected materials validation sheet.

### 2.2 Capacity Costing Pattern (Labor Hours)
The matrix must simulate regular hours vs overtime:
```excel
Regular Hours Constraint -> MAX_CAPACITY_MONTH = =Assumptions!$B$12 (e.g. 160h)
Hours Used = =SUMPRODUCT([Lines_Labor_Required], [Units_Selling])

If Hours Used > MAX_CAPACITY_MONTH:
  Cost = (MAX_CAPACITY_MONTH * Base_Rate) + ((Hours Used - MAX_CAPACITY_MONTH) * Overtime_Rate)
```
This immediately tells the salesperson whether accepting a large order will destroy their profit margin due to Overtime pay (Operational Capacity limits).

### 2.3 Production Scrap and Inventory Waste
- Every real-world BOM has scrap. In the model inputs (Blue Font), there will be a `Historical Scrap Rate (%)`.
- The true final cost to the company (`Real COGS`) = `Calculated COGS * (1 + Scrap Rate)`.

---

## 3. Integration Scope: Finance (Profitability & Executive Analysis)

As a final evaluation layer of the *Quotation Layout*, an abbreviated executive section (Synthetic Income Statement) is projected.

### 3.1 Synthetic P&L for the Sales Director
The structure must report the following financial metric "Steps" so management can validate whether to proceed with a "Won" Stage:

1. **Total Operating Revenue**: Gross Net Revenues.
2. **COGS (Cost Of Goods Sold)**: Calculated Direct Components + Manufacturing labor applied by the template.
3. **Gross Profit**: `=Revenue - COGS`.
4. **OpEx (Assigned Spend/SG&A)**: This is an absorbed expense (absorption costing). Odoo metrics sometimes distribute an administrative burden (e.g., 10% of Gross Profit) as a standard way to uncover the true `Net Operating Income`.

### 3.2 Commission Deductions
A financial macro will evaluate the hidden cost of sales commissions to report a clean NOI:
`='Expected Revenue:sum' * VLOOKUP([SalespersonID], Assumptions_Commissions_Matrix, 2, 0)`

---

## Exportability Guidelines and Odoo 19 `formula_check.py`

Templates generated to represent these logic flows face penalties if they abuse rich visual extensions of commercial Excel that break Odoo 19:
- **NO COMPLEX OVERLAY CHARTS**: Key metrics must be delivered as text/numbers formatted as KPI cards.
- **Limited Validations**: The `formula_check.py` linter in Odoo will penalize advanced array forms (`{=TRANSPOSE(...)}`) in disk export scenarios. All BOM Manufacturing cost calculations must be evaluated using simple `SUMPRODUCT`, `IF`, and standard arithmetic.
