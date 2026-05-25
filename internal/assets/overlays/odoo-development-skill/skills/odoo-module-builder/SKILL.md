---
name: odoo-module-builder
description: "Scaffold complete Odoo module. Version-specific blueprints. Zero legacy references."
bridge: false
on-demand: true
risk_level: medium
---

## Version Blueprint Selection [MANDATORY first step]
```bash
ODOO_VERSION=$(grep "odoo_version:" .atl/config.yaml 2>/dev/null \
  | awk '{print $2}' | tr -d '"' || echo "18")

case "${ODOO_VERSION}" in
  14|15) BLUEPRINT="blueprints/v14-v15" ;;
  16|17) BLUEPRINT="blueprints/v16-v17" ;;
  18)    BLUEPRINT="blueprints/v18" ;;
  19)    BLUEPRINT="blueprints/v19" ;;
  *)
    echo "WARN: unknown version ${ODOO_VERSION}, defaulting to v18"
    BLUEPRINT="blueprints/v18"
    ;;
esac
echo "Using blueprint: ${BLUEPRINT}"
```

## Research Priority [LOCAL FIRST — never external first]
1. Engram: `mem_search("odoo ${ODOO_VERSION} module scaffold")`
2. rg in existing project modules (look at similar modules)
3. Blueprint from `${BLUEPRINT}/` (version-specific, no cross-refs)
4. Context7 ONLY for edge cases not in blueprint

## Recovery Strategies [MANDATORY in every Odoo SKILL.md]
- Blueprint missing → use v18 as fallback + ADD RISK note
- scaffold_module.py fails → create files manually from blueprint templates
- Version unknown → ASK USER explicitly before proceeding
- Import error in generated __init__.py → check model names match class names

## Anti-Pattern Prohibition
NEVER add "IMPORTANT: for v19 remember to change X to Y" blocks.
blueprint for v19 already has the correct syntax.
Cross-version instructions inside a versioned blueprint = REJECTED.
