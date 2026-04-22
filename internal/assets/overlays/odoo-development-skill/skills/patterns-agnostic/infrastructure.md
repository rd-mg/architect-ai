# Infrastructure Patterns

Consolidated from the following source files:
- `cron-server-action-patterns.md` (architect-ai)
- `security-acl-rules-patterns.md` (architect-ai)
- `translation-i18n-patterns.md` (architect-ai)
- `performance-profiling-patterns.md` (architect-ai)

> **Version-specific syntax** → `patterns-{version}/infrastructure.md`
> `ir.cron` API stability · `groups` XML syntax · `_sql_constraints` vs `SQL()`

---

## Security & ACL

### Record Rules (domain_force)
```xml
<record id="rule_my_model_user" model="ir.rule">
    <field name="name">My Model: User Access</field>
    <field name="model_id" ref="model_my_model"/>
    <field name="domain_force">[('user_id', '=', user.id)]</field>
    <field name="groups" eval="[(4, ref('base.group_user'))]"/>
</record>
```

### Access Rights (CSV)
```csv
id,name,model_id:id,group_id:id,perm_read,perm_write,perm_create,perm_unlink
access_my_model_user,my.model.user,model_my_model,base.group_user,1,1,1,0
```

---

## Scheduled Actions (Cron)

```xml
<record id="ir_cron_my_task" model="ir.cron">
    <field name="name">My Periodic Task</field>
    <field name="model_id" ref="model_my_model"/>
    <field name="state">code</field>
    <field name="code">model._run_my_task()</field>
    <field name="interval_number">1</field>
    <field name="interval_type">days</field>
    <field name="numbercall">-1</field>
</record>
```

---

## Anti-Patterns

```python
# ❌ NEVER use [('id', 'in', ids)] for large lists: Odoo 17+ optimizes this, but old versions slow down.
# ✅ CORRECT: Use domain filters or _read_group.

# ❌ NEVER define record rules with broad permissions for everyone.
# ✅ CORRECT: Scope rules to specific groups to avoid overlaps.

# ❌ NEVER hardcode translations in Python: use _("string").
```

---

## Version Matrix

| Feature | v14-v16 | v17 | v18 | v19 |
|---------|---------|-----|-----|-----|
| I18n | `_("str")` | `_("str")` | `_("str")` | `_("str")` |
| Performance | `profiler` | `profiler` | `profiler` | `profiler` |
| Security | `ir.rule` | `ir.rule` | `ir.rule` | `ir.rule` |
| SQL Index | `index=True` | `index=True` | `index=True` | `models.Index` |
