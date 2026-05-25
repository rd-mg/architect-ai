#!/usr/bin/env bash
# Lints Odoo overlay skill files for non-English content
# Run during: architect-ai skill-registry (Odoo projects)
set -euo pipefail

SKILL_DIR="${1:-.atl/overlays/odoo-development-skill}"
VIOLATIONS=0
CHECKED=0

SPANISH_PATTERNS=(
  'Cuando el usuario'  'Objetivo:'  'Resultado esperado'
  'desde el'           'para los'   'del usuario'
  'que el modelo'      'Descripción' 'Configuración'
)

while IFS= read -r -d '' file; do
  CHECKED=$((CHECKED + 1))
  for pattern in "${SPANISH_PATTERNS[@]}"; do
    if grep -q "${pattern}" "${file}" 2>/dev/null; then
      echo "LANGUAGE VIOLATION: ${file}"
      echo "  Pattern: '${pattern}'"
      VIOLATIONS=$((VIOLATIONS + 1))
    fi
  done
done < <(find "${SKILL_DIR}" -name "*.md" -print0 2>/dev/null)

echo "Checked: ${CHECKED} files | Violations: ${VIOLATIONS}"
[ "${VIOLATIONS}" -gt 0 ] && { echo "FAIL: All Odoo skill content must be English."; exit 1; }
echo "PASS: Language policy check passed."
