# Skill: Odoo Security Hardening (v19)

## 1. Description
The `odoo-security-hardening-19` skill defines the architectural mandates for secure module development in Odoo 19. It enforces the transition from legacy security conventions to explicit API-level security.

## 2. Core Mandatory Rules
- **Method Exposure**: The underscore prefix (`_`) is officially insufficient for RPC blocking. **MUST** use `@api.private`.
- **Privilege Escalation**: Assigning groups during `res.users.create()` is blocked by the ORM. **MUST** use a two-step `create()` then `write()` flow.

## 3. Implementation Patterns (Bad vs Good)

### A. Blocking RPC Access
```python
# ❌ LEGACY: Underscore prefix is no longer sufficient for security.
def _sensitive_action(self):
    pass

# ✅ v19 STANDARD: Explicit security gating
from odoo import api

@api.private
def _sensitive_action(self):
    # This method is now blocked from external RPC calls.
    pass
```

### B. Safe User Creation
```python
# ❌ LEGACY: Privilege escalation risk (Blocked in Odoo 19)
user = self.env['res.users'].create({
    'name': 'User',
    'groups_id': [Command.link(self.env.ref('base.group_system').id)]
})

# ✅ v19 STANDARD: Secure two-step approach
user = self.env['res.users'].create({'name': 'User'})
user.write({'groups_id': [Command.link(self.env.ref('base.group_system').id)]})
```

## 4. Verification Workflow
- Ensure all business-logic methods that shouldn't be exposed are decorated.
- Use static analysis (Compliance Checker) to detect usage of restricted ORM parameters in user creation.

## 5. Maintenance
- Monitor Odoo 19 `security/ir.model.access.csv` changes.
- Ensure all custom modules audit their `groups_id` assignments during creation.
