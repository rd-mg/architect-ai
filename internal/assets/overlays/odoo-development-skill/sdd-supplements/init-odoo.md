# SDD Init — Odoo Context

When running sdd-init in an Odoo project:

## Version Detection (ONCE — cache result)

1. Find all `__manifest__.py` in the project:
   rg '"version"' --glob '__manifest__.py' -l
2. Extract major versions (first digit of X.Y.Z.W):
   rg '"version":\s+"(\d+)\.' --glob '__manifest__.py' -o -r '$1' | sort -u
3. Persist result to Engram:
   mem_save(
     topic_key: "sdd-init/{project}/odoo-versions",
     content: {versions: [18, 19], primary: 19}
   )

## Overlay Validation

Confirm active supplements for detected version:
- patterns-{version}/ bundle bridged ✓
- migration-{from}-{to}/ bundled if multi-version ✓
- patterns-agnostic/ always bridged ✓

## Session Context Block

After detection, emit to orchestrator (included in every subsequent sub-agent prompt):

> Odoo version(s): {versions}. Primary: {primary}.
> Version bundle: patterns-{primary}/
> Reference guides: skills/odoo-{primary}.0/references/
