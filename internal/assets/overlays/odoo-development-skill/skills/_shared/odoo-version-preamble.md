# Odoo Version Context (Shared Preamble)

Odoo version for this session was detected during sdd-init and cached in Engram.

To retrieve it:
mem_search(query: "sdd-init/{project}/odoo-versions")
→ mem_get_observation(id) → {versions: [18], primary: 18}

If NOT found (session started without sdd-init):
rg '"version"' __manifest__.py
# Format: "version": "18.0.1.0.0" → major version = 18
Persist result before continuing.

ALL version-specific decisions below use this cached value.
