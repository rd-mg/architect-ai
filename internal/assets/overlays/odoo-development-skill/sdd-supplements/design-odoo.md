# SDD Design — Odoo Context

When designing in an Odoo project, follow this protocol IN ADDITION to the standard sdd-design behavior.

## Architecture Layers

Every Odoo design MUST explicitly address these layers:

1. **Data layer** (models, fields, inheritance)
2. **Business logic layer** (computed fields, constraints, onchange, Python methods)
3. **View layer** (XML views, widgets, QWeb templates)
4. **Security layer** (access rules, record rules, groups)
5. **Integration layer** (mail, cron, external APIs, webhooks)
6. **Migration layer** (pre/post-migrate scripts if schema changes)

## Global Collision Check (MANDATORY — BEFORE_MODEL HOOK)

Before starting the design, you MUST verify that your approach does not collide with established global decisions or existing patterns for the target models. This is your `before_model` requirement.

1. **Search Engram**: Call `mem_search` with the target model name and "decision" or "pattern" as keywords.
2. **Detect Collision**: If an Engram exists that contradicts your proposed design (e.g., "NEVER modify res.partner directly"), you MUST:
   - Transition to **+++Autoreason-lite** posture.
   - Address the collision in the design rationale.
   - Adjust the implementation strategy to comply with the global mandate.

## Domain Bounded Contexts

Reference `sdd-supplements/domain-map.md` for the DDD bounded contexts.

## DDD Tactical Patterns

When implementing complex domain logic, use the tactical patterns defined in `skills/patterns-ddd/SKILL.md`:
1. **Aggregate Roots**: Enforce invariants using `@api.constrains`.
2. **Value Objects**: Use compute fields + `@api.depends` for domain logic.
3. **Domain Services**: Use `models.AbstractModel` for orchestration.
4. **Specifications**: Use `@api.model` returning `Domain` objects (Odoo 19+) or lists.

In your design, state:

```markdown
## Domain Placement
Primary domain: {Sales | Inventory | Accounting | Purchase | MRP | HR | CRM | Website | POS | Project}
Secondary domains touched: {list}
Anti-corruption layer: {model or method that bridges domains, if applicable}
```

Rule: **Never let one domain's code directly write to another domain's tables.** Use Odoo's standard inheritance mechanisms (`_inherit`, `_inherits`) or method calls, not direct SQL.

## Inheritance Strategy

Pick ONE inheritance strategy per model and justify:

| Strategy | When to Use | Syntax |
|----------|-------------|--------|
| Classical inheritance | Extending a model in place (adding fields/methods) | `_inherit = 'existing.model'` |
| Prototypal inheritance | Creating a new model that copies another's structure | `_inherit = [...], _name = 'new.model'` |
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

For every view modification, specify the exact XPath/position:

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
- Always verify XML IDs exist in the base view before designing the inheritance

## Security Design

Every new model requires explicit security design:

```markdown
## Security

### Access (ir.model.access.csv)
| User Group | Read | Write | Create | Delete |
|-----------|:----:|:-----:|:------:|:------:|
| base.group_user | ✓ | | | |
| sales_team.group_sale_salesman | ✓ | ✓ | ✓ | |
| sales_team.group_sale_manager | ✓ | ✓ | ✓ | ✓ |

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

Respect the 800-word limit from the standard sdd-design protocol. If the design exceeds it:
- Split by layer (data design, business logic design, view design) across multiple artifacts
- Move large migration SQL blocks to a separate artifact referenced by topic_key

## Boundaries

- Do NOT include implementation code
- Do NOT decide file paths beyond the module root (that's sdd-tasks)
- Do NOT skip the security section, EVEN for internal-only modules
- Do NOT design for "all versions" — pick the target and design for it
