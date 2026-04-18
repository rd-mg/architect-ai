---
name: patterns-ddd
Trigger: "When implementing domain logic, aggregates, invariants, value objects, or complex business rules in Odoo."
description: >
  Domain-Driven Design (DDD) tactical patterns for Odoo. Maps architectural concepts like
  Aggregate Roots, Value Objects, Domain Services, and Specifications to Odoo ORM decorators
  and structures. Use this to ensure business logic is correctly encapsulated and invariants are enforced.
globs: "**/*.{py,xml}"
---

# Odoo DDD Tactical Patterns

This skill provides guidance on mapping DDD tactical patterns to the Odoo framework (v17+).

## Core Patterns

### 1. Aggregate Root (Model with Invariants)
In Odoo, the **Aggregate Root** is the primary `models.Model`. Business invariants (rules that must always be true) are enforced using `@api.constrains`.

- **Pattern**: Use `@api.constrains` for global invariants.
- **Why**: Ensures the rule is checked on every create/write, not just UI actions.
- **Example**:
    ```python
    class SaleOrder(models.Model):
        _inherit = 'sale.order'

        @api.constrains('order_line')
        def _check_lines_on_confirm(self):
            for order in self:
                if order.state == 'sale' and not order.order_line:
                    raise ValidationError("A confirmed sale order must have at least one line.")
    ```

### 2. Value Object (Data Logic)
**Value Objects** are immutable data structures that encapsulate logic. In Odoo, these are often mapped to `fields.Selection` or related fields combined with `compute` methods.

- **Pattern**: Use `compute` fields with `@api.depends` for domain logic derived from attributes.
- **Odoo 19+ Tip**: For complex logic, extract to a pure Python class and use it within the compute method.
- **Example**: Support Ticket SLA calculation logic.

### 3. Domain Service (Cross-Model Orchestration)
When logic doesn't belong to a single entity, use a **Domain Service**.

- **Pattern**: Use `models.AbstractModel` for stateless orchestration logic.
- **Why**: Allows grouping methods without creating a database table.
- **Example**:
    ```python
    class StockOptimizer(models.AbstractModel):
        _name = 'stock.optimizer.service'

        @api.model
        def optimize_pickings(self, picking_ids):
            # Batch orchestration logic across warehouses
            pass
    ```

### 4. Specification (Reusable Queries)
**Specifications** encapsulate query logic.

- **Pattern**: Use `@api.model` methods returning domain filters.
- **Odoo 19+ Requirement**: Use the native `odoo.fields.Domain` object for advanced composition.
- **Example**:
    ```python
    @api.model
    def _get_at_risk_domain(self):
        from odoo.fields import Domain
        return Domain([('stage_id.is_close', '=', False), ('sla_deadline', '<', fields.Datetime.now())])
    ```

### 5. Domain Events
Use the Odoo **Bus** or **Tracking** system to react to state changes.

### 6. Repository
The **Odoo ORM** (`env['model'].search(...)`) acts as the Repository. Do not wrap it further unless implementing an external integration adapter.

---

## Non-Applicable Patterns (Anti-patterns)

Avoid these DDD patterns in Odoo as they conflict with the "Active Record" nature of the framework:
1. **Explicit Repositories**: Creating separate `Repository` classes adds unnecessary boilerplate; Odoo `env` is the repository.
2. **DTOs for internal logic**: Passing dictionaries instead of recordsets loses the power of lazy-loading and pre-fetching.
3. **Anemic Domain Model**: Putting all logic in "Services" and leaving Models as data containers. Keep logic in Models where possible.
4. **Manual Transaction Management**: Odoo handles transactions via its cursor; manual `commit()` is usually a security/integrity risk.
