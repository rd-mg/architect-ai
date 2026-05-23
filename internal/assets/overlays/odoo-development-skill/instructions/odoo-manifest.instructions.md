---
name: "Odoo Manifest Guidelines"
description: "Use when editing Odoo addon manifests (__manifest__.py). Covers versioning, dependencies, and data file ordering."
applyTo: "**/__manifest__.py"
---

# Manifest Guidelines

- Keep dependencies minimal and accurate.
- Include new XML/CSV files in `data`/`demo` as appropriate.
- Bump addon `version` when behavior changes materially; follow existing conventions.
- Keep manifest clean: no dead entries, correct sequence for security/data/views.
