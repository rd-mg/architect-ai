# SDD Archive — Odoo Context

## Pre-Archive Checklist (Odoo-specific)

These checks MUST pass before closing the change:

### Module Health
- [ ] `__manifest__.py` version reflects all changes in this change
- [ ] All modules touched have CHANGELOG.md or README version entry
- [ ] `ir.model.access.csv` has entries for ALL models (new and inherited)

### OCA Compliance (if applicable)
- [ ] License declared in manifest (`OPL-1`, `LGPL-3`, or `AGPL-3`)
- [ ] No `print()` statements in Python files
- [ ] No hardcoded IDs (use `xml_id` references, not raw integer IDs)
- [ ] All fields have `string=` attribute

### Uninstall Safety
- [ ] Module uninstall does not orphan records in related models
- [ ] Any `post_init_hook` has corresponding `uninstall_hook` if data was created

### Data Integrity
- [ ] All `ir.rule` domains use valid field paths
- [ ] Demo data (if any) does not conflict with base data

## Archive Artifact Addition

Add to standard archive report:
```markdown
## Odoo Compliance
- Versions tested: {list}
- OCA compliance: {PASS | PARTIAL | N/A}
- Module uninstall safety: {SAFE | RISKY — describe}
- Open regressions: {none | list}
```
