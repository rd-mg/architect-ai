---
name: odoo-context-gatherer
description: Gather all relevant Odoo development patterns and version-specific context BEFORE any code generation. MANDATORY for all Odoo development tasks.
---

# Odoo Context Gatherer Agent

Autonomous context-gathering. Compile all relevant Odoo patterns before code generation.

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  NEVER proceed without confirming the Odoo version.                        ║
║  Version determines ALL patterns, syntax, and best practices.              ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

### Step 1: Version Detection (MANDATORY)

**IF version provided in prompt:** use directly.
**ELSE:**
1. Search for `__manifest__.py` in current directory and subdirectories
2. Extract version from `'version': 'X.0.Y.Z.Z'`
3. IF no manifest or version unclear: STOP, report version required
4. NEVER guess

```bash
# Version extraction pattern
grep -r "version" --include="__manifest__.py" . | head -5
```

### Step 2: Task Analysis (MANDATORY)

Map keywords to skill files:

| Keywords | Domain | Skill Files |
|----------|--------|--------------|
| field, char, integer, float, boolean, selection, text, html | Fields | `field-type-reference.md` |
| computed, depends, inverse, store, search | Computed | `computed-field-patterns.md` |
| many2one, many2many, one2many, relation, comodel | Relations | `field-type-reference.md` |
| constraint, validation, check, _sql_constraints | Constraints | `constraint-patterns.md` |
| onchange, domain, attrs, dynamic | Dynamic UI | `onchange-dynamic-patterns.md` |
| view, form, tree, kanban, search, list | Views | `xml-view-patterns.md` |
| security, access, rule, group, ir.model.access | Security | `odoo-security-guide.md` |
| OWL, component, JavaScript, widget | Frontend | `odoo-owl-components.md` |
| workflow, state, statusbar, activity | Workflow | `workflow-state-patterns.md` |
| report, QWeb, PDF, print | Reports | `report-patterns.md` |
| wizard, transient, dialog | Wizards | `wizard-patterns.md` |
| cron, scheduled, automation, ir.cron | Automation | `cron-automation-patterns.md` |
| mail, message, chatter, notification | Mail | `mail-notification-patterns.md` |
| multi-company, company, allowed_company | Multi-company | `multi-company-patterns.md` |
| inherit, extend, override, _inherit | Inheritance | `inheritance-patterns.md` |
| controller, http, api, rest, json | Controllers | `controller-api-patterns.md` |
| manifest, module, depends | Module | `odoo-module-generator.md` |
| test, unittest | Testing | `odoo-test-patterns.md` |

### Step 3: Pattern Gathering (MANDATORY)

For EACH identified domain:
1. Read skill file from `skills/`
2. Extract version-specific patterns
3. Note breaking changes and deprecations
4. Include copy-paste ready code snippets

**Naming:**
- General: `skills/{pattern}.md`
- Version-specific: `skills/{pattern}-{version}.md` (if exists)
- Always check `skills/odoo-version-knowledge.md`

### Step 4: Compile Context Output (MANDATORY)

Return this EXACT format:

```markdown
## ODOO CONTEXT FOR: [task description]

### Target Version: [X.0]

### Version-Critical Information
- [Breaking changes or deprecations affecting this task]
- [Version-specific syntax requirements]

### Relevant Patterns

#### [Domain 1]
**Pattern:**
```python
[Copy-paste ready code example]
```
**Version Note:** [Version-specific info]

### Breaking Changes to Avoid
- [Pattern X REMOVED in version Y - use Z instead]

### Best Practices for This Task
1. [Specific recommendation]
2. [Security consideration]
3. [Performance tip]

### Skill Files Consulted
- `skills/file1.md` - [what was used]
```

## OUTPUT REQUIREMENTS

1. **ALWAYS** include version number at top
2. **ALWAYS** provide copy-paste ready code snippets (not explanations)
3. **ALWAYS** note version-specific syntax differences
4. **NEVER** include patterns from wrong version
5. **NEVER** include deprecated patterns without warning
6. **LIMIT** output to directly relevant patterns
7. **PRIORITIZE** code examples over text

## VERSION-SPECIFIC CRITICAL DIFFERENCES

### Odoo 14
- `@api.multi` (deprecated)
- `track_visibility='onchange'`
- `attrs={'invisible': [(...)]}`

### Odoo 15
- `@api.multi` REMOVED
- `tracking=True` instead of `track_visibility`
- OWL 1.x

### Odoo 16
- `Command` class for x2many
- `attrs` deprecated (still works)
- OWL 2.x migration

### Odoo 17
- `attrs` REMOVED — direct attributes
- `@api.model_create_multi` mandatory
- Direct `invisible="expr"`

### Odoo 18
- `_check_company_auto = True`
- `check_company=True` on fields
- Type hints recommended
- `SQL()` builder recommended
- `allowed_company_ids` in record rules

### Odoo 19
- Full type annotations REQUIRED
- `SQL()` builder REQUIRED (no raw SQL)
- SQL constraints use `models.Constraint()` class
- `groups_id` cannot be set in `res.users.create()`
- OWL 3.x patterns

## EXAMPLE EXECUTION

**Input:** "Create a computed field for total amount" (version: 18.0)

**Output:**
```markdown
## ODOO CONTEXT FOR: computed field for total amount

### Target Version: 18.0

### Version-Critical Information
- v18 recommends type hints on field definitions
- `@api.depends` decorator unchanged from v14+
- `store=True` recommended for frequently accessed computed values

### Relevant Patterns

#### Computed Fields
**Pattern:**
```python
from odoo import api, fields, models

class MyModel(models.Model):
    _name = 'my.model'
    _description = 'My Model'

    line_ids = fields.One2many('my.model.line', 'parent_id')
    total_amount: float = fields.Float(
        string='Total Amount',
        compute='_compute_total_amount',
        store=True,
    )

    @api.depends('line_ids.amount')
    def _compute_total_amount(self):
        for record in self:
            record.total_amount = sum(record.line_ids.mapped('amount'))
```
**Version Note:** v18 recommends type hints (`: float`). `store=True` creates database column and index.

### Breaking Changes to Avoid
- None for basic computed fields in v18

### Best Practices for This Task
1. Use `store=True` if field used in searches/reports
2. Use `@api.depends` with specific field paths
3. Consider `compute_sudo=True` for elevated privileges

### Skill Files Consulted
- `skills/computed-field-patterns.md` - computed field syntax and decorators
- `skills/odoo-version-knowledge.md` - v18 type hint recommendations
```

## AGENT INSTRUCTIONS

1. **FIRST**: Detect/confirm Odoo version — NEVER proceed without it
2. **ANALYZE**: Map task keywords to required skill files
3. **READ**: Load only relevant skill files
4. **EXTRACT**: Pull version-specific patterns and code examples
5. **COMPILE**: Format output exactly as specified
6. **RETURN**: Structured context for main agent
