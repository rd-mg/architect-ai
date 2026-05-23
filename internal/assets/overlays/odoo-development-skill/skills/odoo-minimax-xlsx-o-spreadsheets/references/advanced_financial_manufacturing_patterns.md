# Advanced Functional & Manufacturing Patterns

When connecting `o-spreadsheet` file (Quotation Calculator) to ERP (Odoo v19), quotation calculator must evolve beyond simple (Price x Quantity) = Total. Must consider transactional impact on other key system apps, primarily manufacturing (`mrp`) and accounting (`account`).

Below, document technical breakdown of *Advanced Quotation Calculator*.

---

## 1. Integration Scope: Sales

Section receives base data and logically interconnects it.

- **Base Revenue Mapping (`Revenue`):**
  Advanced model does not assume final `Unit Price`. Acts as bridge via formula querying injected Odoo database for hidden price list rules (`pricelist_id` rules).
  
  *Standard Applied Formula:*
  `=IFERROR(VLOOKUP([ProductID], Data!$A$2:$F$100, [PriceColumn], FALSE), 0)`

- **Dynamic Tax and Discount Reallocation:**
  By identifying original `Discount` column from quotation, calculator exposes *Net Income* scenario.
  *(List Price - Discount) + Specific region Tax.*

---

## 2. Integration Scope: Manufacturing & Costing

If product marked with MTO (Make To Order) or *Manufacture* route, salesperson needs to preview economic risk due to industrial capacity constraints.

### 2.1 Bill Of Materials (BOM) Modeling
- **Volume Ratio:** Volume demanded in initial input linked with resources required to dispatch.
- **Direct Material Cost:** Internally linked to injected materials validation sheet.

### 2.2 Capacity Costing Pattern (Labor Hours)
Matrix must simulate regular hours vs overtime:
```excel
Regular Hours Constraint -> MAX_CAPACITY_MONTH = =Assumptions!$B$12 (e.g. 160h)
Hours Used = =SUMPRODUCT([Lines_Labor_Required], [Units_Selling])

If Hours Used > MAX_CAPACITY_MONTH:
  Cost = (MAX_CAPACITY_MONTH * Base_Rate) + ((Hours Used - MAX_CAPACITY_MONTH) * Overtime_Rate)
```
Immediately tells salesperson whether accepting large order will destroy profit margin due to Overtime pay (Operational Capacity limits).

### 2.3 Production Scrap and Inventory Waste
- Every real-world BOM has scrap. In model inputs (Blue Font), there will be `Historical Scrap Rate (%)`.
- True final cost to company (`Real COGS`) = `Calculated COGS * (1 + Scrap Rate)`.

---

## 3. Integration Scope: Finance (Profitability & Executive Analysis)

As final evaluation layer of *Quotation Layout*, abbreviated executive section (Synthetic Income Statement) is projected.

### 3.1 Synthetic P&L for Sales Director
Structure must report following financial metric "Steps" so management can validate whether to proceed with "Won" Stage:

1. **Total Operating Revenue**: Gross Net Revenues.
2. **COGS (Cost Of Goods Sold)**: Calculated Direct Components + Manufacturing labor applied by template.
3. **Gross Profit**: `=Revenue - COGS`.
4. **OpEx (Assigned Spend/SG&A)**: Absorbed expense (absorption costing). Odoo metrics sometimes distribute administrative burden (e.g., 10% of Gross Profit) as standard way to uncover true `Net Operating Income`.

### 3.2 Commission Deductions
Financial macro evaluates hidden cost of sales commissions to report clean NOI:
`='Expected Revenue:sum' * VLOOKUP([SalespersonID], Assumptions_Commissions_Matrix, 2, 0)`

---

## Exportability Guidelines and Odoo 19 `formula_check.py`

Templates generated to represent these logic flows face penalties if they abuse rich visual extensions of commercial Excel that break Odoo 19:
- **NO COMPLEX OVERLAY CHARTS**: Key metrics must be delivered as text/numbers formatted as KPI cards.
- **Limited Validations**: `formula_check.py` linter in Odoo penalizes advanced array forms (`{=TRANSPOSE(...)}`) in disk export scenarios. All BOM Manufacturing cost calculations must be evaluated using simple `SUMPRODUCT`, `IF`, and standard arithmetic.
