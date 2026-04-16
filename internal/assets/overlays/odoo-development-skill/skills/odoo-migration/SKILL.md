---
name: odoo-migration
description: "Plan and execute safe Odoo addon migrations (version upgrades, data migrations, and behavior changes). Use when upgrading between Odoo versions or adding migration scripts/checklists."
argument-hint: "From Odoo X -> Y, which addon(s), and what breaks?"
---

# Skill Instructions

## What this skill helps accomplish
This skill helps you safely plan and execute Odoo addon migrations, including version upgrades, data migrations, and behavior changes. It provides a structured approach to identifying risks, making incremental changes, and validating the final result across different Odoo environments (versions 13.0 to 19.0).

## When to use this skill
- Upgrading an addon to a newer Odoo major/minor version.
- Writing migration scripts or validating upgrade readiness.
- Reviewing breaking changes across addons (manifests, views, models, security).

## Step-by-step procedure
1. Identify the target addon(s) and current `version` in `__manifest__.py`.
2. Inventory risky surfaces:
   - External IDs and view inheritance changes.
   - Security rules (`ir.model.access.csv`, record rules).
   - Stored computed fields and schema-impacting changes.
3. Make changes incrementally:
   - One behavior change per PR when possible.
   - Keep compatibility where feasible.
4. Add/adjust tests in the addon's `tests/` folder when behavior changes.
5. Validation:
   - If an Odoo environment is available: run the module tests and upgrade the module on a copy of data.
   - Otherwise: run `python -m compileall .` and perform a targeted review of XML inheritance + manifest entries.

## Examples of expected input and output

**Input:**
"Plan a migration for `sws_account` from Odoo 16.0 to 17.0."

**Output:**
- A detailed migration checklist.
- A list of files to touch.
- A risk list + mitigations.
