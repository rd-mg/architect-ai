## Odoo Project Detection [sdd-init + sdd-explore MANDATORY — all platforms]

### Detection (bash + fish compatible)
```bash
# Signal 1: __manifest__.py anywhere
MANIFEST=$(rg -l "__manifest__" --type py -l 2>/dev/null | head -1)

# Signal 2: odoo in requirements / pyproject
ODOO_DEP=$(rg -l "^odoo" requirements.txt pyproject.toml 2>/dev/null | head -1)

# Signal 3: .atl/config.yaml has odoo_version
ODOO_CFG=$(rg "odoo_version" .atl/ -l 2>/dev/null | head -1)

IS_ODOO="false"
[ -n "${MANIFEST}" ] || [ -n "${ODOO_DEP}" ] || [ -n "${ODOO_CFG}" ] && IS_ODOO="true"
```

### Version Extraction
```bash
if [ "${IS_ODOO}" = "true" ] && [ -n "${MANIFEST}" ]; then
  ODOO_VERSION=$(python3 -c "
import ast, re, sys
txt = open('${MANIFEST}').read()
# Try ast parse first
try:
    d = ast.literal_eval(txt)
    v = d.get('version','')
    print(v.split('.')[0] if v else 'unknown')
except:
    # fallback: regex
    m = re.search(r'[\"\\']version[\"\\']\\s*:\\s*[\"\\'](\d+)\\.', txt)
    print(m.group(1) if m else 'unknown')
" 2>/dev/null || echo "unknown")
fi
```

### Skill Registry Load
```bash
if [ "${IS_ODOO}" = "true" ]; then
  REGISTRY=".atl/skill-registry.md"
  OVERLAY_DIR=".atl/overlays/odoo-development-skill"

  [ -f "${REGISTRY}" ] && {
    ODOO_SKILLS=$(rg "^- odoo-" "${REGISTRY}" | sed 's/^- //' | tr '\n' ',')
    echo "odoo_skills_available=${ODOO_SKILLS}"
  }

  [ -d "${OVERLAY_DIR}" ] || {
    echo "WARN: Odoo overlay not installed."
    echo "Fix: architect-ai overlay install odoo-development-skill"
  }
fi
```

### Emit to Engram
```
mem_save(
  topic_key: "project/meta/odoo-detection",
  content: {
    is_odoo: true,
    odoo_version: "${ODOO_VERSION}",
    manifest_path: "${MANIFEST}",
    overlay_installed: bool,
    available_skills: [${ODOO_SKILLS}]
  }
)
```
