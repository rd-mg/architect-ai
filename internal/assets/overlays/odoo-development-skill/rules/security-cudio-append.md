
---

## Cudio Security Standards Appendix

Additional security standards enforced on Cudio Inc. projects. These EXTEND the base security rules above.

### ORM-Only for Writes

NEVER write to the database using raw SQL. Always use ORM methods:
- `env['model'].create(vals)` or `create([vals1, vals2])`
- `record.write(vals)`
- `record.unlink()`

Raw SQL writes bypass ACL, record rules, tracking, and audit trails. They are NEVER acceptable in application code.

The only exception is **performance-critical migrations** where the migration comment MUST explain why ORM is inadequate.

### SQL Injection Prevention

When raw SELECT queries are necessary (rare), ALWAYS use parameterized queries:

Anti-pattern:
```python
# DANGEROUS — SQL injection vulnerability
user_filter = self.env.context.get('user_name')
self.env.cr.execute(f"SELECT id FROM res_users WHERE login = '{user_filter}'")
```

Correct:
```python
# SAFE — parameterized query
self.env.cr.execute(
    "SELECT id FROM res_users WHERE login = %s",
    (user_filter,)
)
```

For Odoo 19+: Use the `SQL()` builder:
```python
from odoo.tools import SQL
query = SQL("SELECT id FROM res_users WHERE login = %s", user_filter)
self.env.cr.execute(query)
```

### sudo() Justification

Every `sudo()` call MUST have a comment explaining why bypassing ACL is necessary.

Anti-pattern:
```python
# BAD — no justification
partner = self.env['res.partner'].sudo().search([('name', '=', name)])
```

Correct:
```python
# GOOD — justified
# sudo: accessing res.partner as portal user to display company info on checkout page
partner = self.env['res.partner'].sudo().search([('name', '=', name)])
```

Common acceptable justifications:
- Portal users accessing related records for display
- Cron jobs running as SUPERUSER
- Integration endpoints verified via separate auth layer
- Computed fields that need cross-user visibility

Unacceptable (refactor instead):
- "It was easier" — write the proper ACL
- "It's temporary" — make it permanent or don't do it
- "The user won't see this" — they might, in an edge case

### Record Rules for Multi-Company Fields

Every model with a `company_id` field MUST have record rules that respect multi-company boundaries:

```xml
<record id="rule_my_model_multi_company" model="ir.rule">
    <field name="name">My Model: multi-company</field>
    <field name="model_id" ref="model_acme_my_model"/>
    <field name="global" eval="True"/>
    <field name="domain_force">[
        '|', ('company_id', '=', False),
        ('company_id', 'in', company_ids)
    ]</field>
</record>
```

### Default Groups Principle

Default groups for new models:
- **Read-only**: `base.group_user` (all internal users can see the record)
- **Create/Write**: Specific business group (e.g., `sales_team.group_sale_salesman`)
- **Unlink**: Manager group only (e.g., `sales_team.group_sale_manager`)

Never grant unrestricted write/unlink to `base.group_user` for business records.

### Portal Security

For portal-accessible records:
- Use `website.website` record rule pattern
- Never expose sensitive fields in public templates
- Always validate access tokens via `access_token` field
- Use `request.env['model'].sudo()` only AFTER token validation

### Password/Secret Management

NEVER:
- Hardcode API keys or passwords in code
- Commit `.env` files or config with secrets
- Log passwords, tokens, or API keys
- Store passwords in plain text fields

ALWAYS:
- Use `ir.config_parameter` with `password=True` flag for secrets
- Read secrets from environment variables for production
- Use Odoo's `res.users.identitycheck` flow for sensitive operations
- Use `sudo()` only after access verification, never before

### Input Validation

All user-provided input MUST be validated before use:

```python
@api.constrains('email')
def _check_email_format(self):
    for record in self:
        if record.email and '@' not in record.email:
            raise ValidationError(_("Invalid email format"))
```

For HTTP endpoints:
```python
@http.route('/my-endpoint', type='json', auth='user')
def my_endpoint(self, **kwargs):
    # Validate each expected field
    required = ['name', 'email']
    missing = [f for f in required if not kwargs.get(f)]
    if missing:
        return {'error': f'Missing fields: {missing}'}
    ...
```

### Cron Job Security

Cron jobs run as SUPERUSER by default. Mitigate by:
- Setting `user_id` explicitly in the cron definition (non-SUPERUSER when possible)
- Filtering records by company/user within the job logic
- Logging all actions performed by the cron

### Audit Trail

For records with compliance/audit needs:
- Add `tracking=True` on status fields
- Inherit from `mail.thread` for chatter-based audit
- Use `mail.activity.mixin` for pending actions

Do NOT use `delete` on audited records. Archive (`active=False`) or create reversal records.

### Cross-Module Security

When a module depends on another:
- Never assume the parent module's ACL is sufficient
- Define explicit ACL for any new field added to an inherited model
- Test with a user who has access to the child but not the parent

### Security Review Checklist

Before any PR is merged, verify:
- [ ] `ir.model.access.csv` covers every new model
- [ ] Record rules exist for multi-company fields
- [ ] All `sudo()` calls have justification comments
- [ ] No raw SQL writes (only ORM)
- [ ] User input validated on server side
- [ ] No secrets in code or committed files
- [ ] Portal endpoints use access token validation
- [ ] Cron jobs log their actions
- [ ] Sensitive fields have appropriate groups
