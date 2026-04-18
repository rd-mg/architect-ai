# Odoo DDD Tactical Specification

## Purpose
Adaptive Reasoning gate: Mode: 1. Why: Defining static documentation requirements for Odoo DDD skill.

Provide structured guidance for mapping Domain-Driven Design (DDD) tactical patterns to Odoo-native constructs. This spec ensures that teams can apply DDD principles without fighting the Odoo ORM's Active Record nature.

## Requirements

### Requirement: Aggregate Root Representation
The Odoo model SHALL be treated as the Aggregate Root.
Business invariants MUST be enforced using `@api.constrains` decorators.

#### Scenario: Invariant enforcement
- GIVEN a Sale Order (Aggregate Root)
- WHEN a user attempts to confirm the order with zero lines
- THEN the system MUST raise a ValidationError
- AND the transaction MUST roll back.

### Requirement: Value Object Implementation
Immutable domain concepts SHOULD be implemented as a combination of `fields.Selection` (for state/type) and `@api.depends` computed fields (for derived logic).
Separate `models.Model` entities SHOULD NOT be created for small immutable concepts unless identity is required.

#### Scenario: Derived SLA logic
- GIVEN a Support Ticket with a 'High' priority Value Object
- WHEN the ticket is created
- THEN the system MUST compute the SLA hours to exactly 24 based on the priority.

### Requirement: Domain Service Isolation
Logic spanning multiple aggregates or complex calculations SHALL be implemented in `models.AbstractModel` using `@api.model` methods.
This ensures logic is not coupled to a specific database record until necessary.

#### Scenario: Cross-aggregate calculation
- GIVEN multiple Stock Picking aggregates
- WHEN a global optimization service is invoked
- THEN the service MUST process all pickings in batch and return an optimized route.

### Pattern: Specification (Reusable Query)
- **Odoo Mapping**: `@api.model` methods returning domain lists or `odoo.fields.Domain` objects.
- **Requirement**: SHOULD encapsulate complex filtering logic to avoid duplication in Python/XML.
- **Requirement**: MUST return a valid Odoo domain filter.
- **Odoo 19+ Requirement**: SHOULD use the native `odoo.fields.Domain` object for advanced domain composition and optimization.
- **Scenario**: Finding "at risk" Support Tickets (not closed, SLA expired).
    - **Implementation**: `ticket._get_at_risk_domain()` returning `[('stage_id.is_close', '=', False), ('sla_deadline', '<', fields.Datetime.now())]`.
