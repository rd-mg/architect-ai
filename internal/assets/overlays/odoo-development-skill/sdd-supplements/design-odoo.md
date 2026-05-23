# SDD Design — Odoo Context

Follow IN ADDITION to standard sdd-design.

## Odoo Version
{See shared preamble — version cached from sdd-init.}

## Architecture Layers

Every Odoo design MUST explicitly address these layers:

1. **Data layer** (models, fields, inheritance)
2. **Business logic layer** (computed fields, constraints, onchange, Python methods)
3. **View layer** (XML views, widgets, QWeb templates)
4. **Security layer** (access rules, record rules, groups)
5. **Integration layer** (mail, cron, external APIs, webhooks)
6. **Migration layer** (pre/post-migrate scripts if schema changes)

## Global Collision Check (MANDATORY — BEFORE_MODEL HOOK)

Before starting design, verify approach does not collide with established global decisions or existing patterns for target models.

1. **Search Engram**: `mem_search` with target model name + "decision" or "pattern" keywords.
2. **Detect Collision**: If Engram contradicts proposed design (e.g., "NEVER modify res.partner directly"):
   - Transition to **+++Autoreason-lite** posture.
   - Address collision in design rationale.
   - Adjust implementation strategy to comply with global mandate.

## Domain Bounded Contexts

| Odoo App | Bounded Context | Primary Aggregate | Common Anti-Corruption Layer |
|----------|----------------|-------------------|------------------------------|
| Sales | SaleContext | sale.order | sale.order.line (bridge to stock) |
| Inventory | StockContext | stock.picking | stock.move (bridge to accounting) |
| Accounting | AccountContext | account.move | account.move.line (bridge to sale) |
| Purchase | PurchaseContext | purchase.order | purchase.order.line |
| Manufacturing | MrpContext | mrp.production | mrp.bom |
| HR | HRContext | hr.employee | hr.contract (bridge to payroll) |
| CRM | CRMContext | crm.lead | crm.stage |
| Website/Portal | PortalContext | website.visitor | portal.mixin |

**Rule**: Change spanning 2+ bounded contexts REQUIRES explicit anti-corruption layer
(model or method that bridges, not direct cross-domain writes).

Full domain map: `sdd-supplements/domain-map.md`

## DDD Tactical Patterns

When implementing complex domain logic, use tactical patterns from `skills/patterns-ddd/SKILL.md`:
1. **Aggregate Roots**: Enforce invariants via `@api.constrains`.
2. **Value Objects**: Use compute fields + `@api.depends` for domain logic.
3. **Domain Services**: Use `models.AbstractModel` for orchestration.
4. **Specifications**: Use `@api.model` returning `Domain` objects (Odoo 19+) or lists.

In your design, state:

```markdown
## Domain Placement
Primary domain: {Sales | Inventory | Accounting | Purchase | MRP | HR | CRM | Website | POS | Project}
Secondary domains touched: {list}
Anti-corruption layer: {model or method bridging domains, if applicable}
```

Rule: **Never let one domain's code directly write to another domain's tables.** Use Odoo's standard inheritance mechanisms (`_inherit`, `_inherits`) or method calls, not direct SQL.

## Inheritance Strategy

Pick ONE inheritance strategy per model and justify:

| Strategy | When to Use | Syntax |
|----------|-------------|--------|
| Classical inheritance | Extending model in place (adding fields/methods) | `_inherit = 'existing.model'` |
| Prototypal inheritance | Creating new model copying another's structure | `_inherit = [...], _name = 'new.model'` |
| Delegation inheritance | Composing models where new model "is a" existing model | `_inherits = {'base.model': 'base_id'}` |
| New abstract mixin | Reusable fields/methods across models | `_name = 'mixin.name', _inherit = ['mail.thread']` |

Design section format:
```markdown
## Model Design

### Model: acme.approval.request
- Strategy: New model
- Inherits from: mail.thread (for tracking)
- Relations: Many2one to sale.order
- Key fields: state (selection), approver_ids (Many2many to res.users)
```

## View Inheritance

For every view modification, specify exact XPath/position:

```markdown
## Views

### sale.order.form.view (inherited)
- Target view: sale.order.view_order_form
- XPath: //field[@name='user_id']
- Position: after
- Content: Add field approval_state
- Rationale: Approval state must be visible next to salesperson
```

Rules:
- Use `hasclass()` NOT `contains(@class, ...)`
- Never use `replace` unless absolutely necessary; prefer `after`/`before`/`inside`
- Always verify XML IDs exist in base view before designing inheritance

## Security Design

Every new model requires explicit security design:

```markdown
## Security

### Access (ir.model.access.csv)
| User Group | Read | Write | Create | Delete |
|-----------|:----:|:-----:|:------:|:------:|
| base.group_user |  | | | |
| sales_team.group_sale_salesman |  |  |  | |
| sales_team.group_sale_manager |  |  |  |  |

### Record Rules (ir.rule)
- {Rule name}: {domain filter} — {reason}
- Example: "Own approval requests only": `[('user_id', '=', user.id)]` for base.group_user

### Company Filtering
Multi-company relevant: {yes/no}
If yes: all queries MUST respect company_ids context.
```

## Integration Contracts

For integrations with other modules, define the contract:

```markdown
## Integration Contracts

### With sale.order
- Hook: Override `action_confirm` to trigger approval workflow
- Signal: Emit `approval.request.created` event
- Expected behavior of sale.order: Remains in `sent` state until approval completes

### With mail module
- Templates: 1 new template `email_template_approval_request`
- Activity types: 1 new `mail.activity.type` with `res_model=acme.approval.request`
```

## Migration Strategy

If schema changes, include migration design:

```markdown
## Migration

### From unversioned → 18.0.1.0.0
- pre-migrate: None (new module)
- post-migrate: None (new module)

### From 18.0.1.0.0 → 18.0.1.1.0
- pre-migrate: Rename field `approver` → `approver_id`
- post-migrate: Populate `state` field from legacy `status` field
```

## Size Budget

Respect 800-word limit from standard sdd-design protocol. If exceeded:
- Split by layer (data design, business logic design, view design) across multiple artifacts
- Move large migration SQL blocks to separate artifact referenced by topic_key

## Odoo YAGNI Patterns
When applying **YAGNI Gate**, reject these common Odoo over-engineering patterns:
1. **Premature Abstract Models**: Don't create `models.AbstractModel` unless used by ≥ 3 active models in this project.
2. **Over-Inheritance**: Avoid inheriting `mail.thread` or `mail.activity.mixin` for internal utility models with low churn.
3. **Irrelevant Multi-company**: Don't add `company_id` fields or domain rules unless project has explicit multi-company requirements.
4. **Excessive Config Parameters**: Avoid creating `ir.config_parameter` entries for simple constants that can be module attributes.
5. **Generic Views**: Don't create complex inherited views with extensive XPaths for views with only 1-2 small changes.

## Boundaries
- Do NOT include implementation code
- Do NOT decide file paths beyond module root (that's sdd-tasks)
- Do NOT skip security section, EVEN for internal-only modules
- Do NOT design for "all versions" — pick target and design for it
