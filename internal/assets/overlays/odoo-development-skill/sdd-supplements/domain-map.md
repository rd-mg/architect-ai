# Odoo Domain Map (DDD Reference)

This map defines the bounded contexts for Odoo's core domains. Use it during
`sdd-propose` and `sdd-design` to:
- Identify which domain(s) a change touches
- Detect cross-domain impacts
- Specify the anti-corruption layer when a change crosses domains

## The Ten Core Domains

| Domain | Core Modules | Key Models | Incoming Integrations | Outgoing Integrations |
|--------|--------------|------------|----------------------|----------------------|
| **Sales** | sale, sale_management | sale.order, sale.order.line | crm.lead → sale.order | sale.order → account.move |
| **Inventory** | stock, stock_account | stock.move, stock.picking, stock.lot | purchase.order → stock.picking | stock.move → account.move.line |
| **Accounting** | account, account_payment | account.move, account.journal, account.payment | Sale, Stock, HR, Purchase all produce account.moves | Financial reports |
| **Purchase** | purchase | purchase.order | mrp.production → purchase.order | purchase.order → stock.picking |
| **Manufacturing** | mrp, mrp_account | mrp.production, mrp.bom, mrp.workorder | stock → raw materials | mrp.production → stock.move (finished goods) |
| **HR** | hr, hr_payroll, hr_expense, hr_holidays | hr.employee, hr.payslip, hr.expense | — | hr.expense → account.move, payslip → account.move |
| **CRM** | crm | crm.lead, crm.team | Website lead forms | crm.lead → sale.order |
| **Website** | website, website_sale | website, website.page | — | website_sale → sale.order |
| **POS** | point_of_sale | pos.order, pos.session, pos.config | pos.config reads product, pricelist | pos.order → account.move, stock.move |
| **Project** | project, hr_timesheet | project.project, project.task | — | hr_timesheet → account.analytic.line |

## Cross-Domain Bridges (Anti-Corruption Layers)

The following are the standard bridges. When a change needs to cross domains, route through these, NOT direct database writes:

| Bridge | Source → Target | Mechanism |
|--------|----------------|-----------|
| Sale → Accounting | sale.order.action_confirm → creates account.move | Override action_confirm, call _create_invoices |
| Stock → Accounting | stock.move._action_done → generates journal entries | Automatic via stock_account module |
| Purchase → Stock | purchase.order.button_confirm → creates stock.picking | Automatic via purchase_stock module |
| MRP → Stock | mrp.production.action_confirm → reserves components | Automatic via mrp module |
| HR → Accounting | hr.expense.action_submit_expenses → creates account.move | expense sheet flow |
| CRM → Sale | crm.lead.action_new_quotation → creates sale.order | Manual button or automation |
| Website → Sale | website_sale checkout → creates sale.order | HTTP controller posts form |
| POS → Accounting/Stock | pos.session._create_account_move → creates closing entries | At POS session close |
| Project → Accounting | hr_timesheet creates account.analytic.line | Billing integration |

## Rules for SDD Phases

### sdd-propose
Identify which domain(s) the change touches. List them. If cross-domain:
- Explain WHY cross-domain is necessary
- Identify the appropriate bridge (from the table above)
- Do NOT propose direct SQL writes across domains

### sdd-design
If change crosses domains:
- Identify the anti-corruption layer model, method, or event
- Preserve existing bridge semantics — don't reimplement standard Odoo flows
- Document all side effects on other domains

### sdd-spec
Write specs per-domain. Each domain = one spec section. Cross-domain effects go in a separate "Integration Effects" section.

### sdd-apply
- Never let one domain's code directly write to another domain's tables
- Use standard Odoo inheritance (`_inherit`) and method calls
- When adding a new cross-domain bridge, prefer events (`mail.message` post) or computed fields over direct manipulation

### sdd-verify
Validate that:
- Cross-domain interactions use documented bridges
- No raw SQL writes target another domain's tables
- Existing bridge behavior is preserved (regression test other domains)

## Anti-Pattern Examples

### ❌ BAD: Direct cross-domain write
```python
# In sale.order.py
def action_confirm(self):
    super().action_confirm()
    # WRONG: directly inserting into accounting's table
    self.env.cr.execute("INSERT INTO account_move ...")
```

### ✅ GOOD: Use the bridge
```python
# In sale.order.py
def action_confirm(self):
    result = super().action_confirm()
    # Use the standard bridge method
    self._create_invoices()
    return result
```

### ❌ BAD: Reading across domains without event
```python
# In hr.employee.py
def action_compute_timesheet_billing(self):
    # WRONG: tight coupling to account_move internals
    moves = self.env['account.move'].search([('employee_id', '=', self.id)])
    ...
```

### ✅ GOOD: Use a method exposed by the target domain
```python
# In hr.employee.py
def action_compute_timesheet_billing(self):
    # Use method exposed by account module
    return self.env['account.move'].get_employee_billing(self.ids)
```

## When to Create a New Bridge

Create a new bridge ONLY when:
1. No existing bridge covers the flow
2. The integration is durable (not a one-off)
3. The semantics are well-defined

Prefer:
- Events (mail.message, bus notifications) for loose coupling
- Exposed methods for explicit contracts
- Abstract mixins for shared behavior across models

Avoid:
- Direct SQL across domains
- Cross-domain @api.depends (fragile)
- Writing to another domain's computed fields

## High-Risk Models & Conflict Checklist

The following models are central to Odoo's core integrity. Any modification MUST be audited with extreme care during `sdd-design` and `sdd-verify`.

| Model | Primary Risks | Mandatory Verification |
|-------|---------------|------------------------|
| **res.partner** | Database locking, duplicate data, sync overhead | ALWAYS use \`_inherit\`. NO direct SQL writes. |
| **account.move** | Fiscal integrity, N+1 queries, tax logic breakage | AUDIT computed fields. VERIFY \`ondelete\` on lines. |
| **stock.move** | Inventory valuation drift, missing valuation entries | CHECK \`stock_account\` integration. NO manual state set. |
| **res.users** | Permission escalation, session bloat, auth bypass | AUDIT record rules. VERIFY group inheritance. |
| **account.payment** | Payment reconciliation breakage, orphan payments | CHECK \`payment_state\` transition logic. |

### Mandatory Conflict Protocol (ALL PHASES)
The following protocol is NOT optional and must be executed in `sdd-propose`, `sdd-design`, `sdd-apply`, and `sdd-verify`:

1. **Identify** if the change touches any high-risk model.
2. **Scan** Engram for "archived-decisions" related to these models.
3. **Audit** for "Anti-Patterns" (direct SQL, missing `ondelete`, N+1).
4. **Pass** Judgement Day Gate explicitly focusing on these models during verification.
