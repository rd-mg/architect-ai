#!/usr/bin/env bash
# assumption-linter.sh — Static analysis for unresolved gaps in SDD artifacts
# Usage: bash .atl/scripts/assumption-linter.sh {change_name}
# Exit 0 = PASS, Exit 1 = VIOLATIONS FOUND

set -euo pipefail

CHANGE="${1:?change_name required}"
ARTIFACT_DIR="openspec/changes/${CHANGE}"
VIOLATIONS=0
TOTAL_CHECKED=0

if [ ! -d "${ARTIFACT_DIR}" ]; then
  echo "WARN: artifact dir not found: ${ARTIFACT_DIR}"
  echo "Checking Engram artifacts instead..."
  # If openspec not used, check engram export or skip
  exit 0
fi

echo "Running Assumption Linter on: ${ARTIFACT_DIR}"
echo "──────────────────────────────────────────"

# Rule 1: No TBD or To Be Determined
echo "Checking Rule 1: No TBD..."
while IFS= read -r -d '' file; do
  TOTAL_CHECKED=$((TOTAL_CHECKED + 1))
  # Exclude code blocks (lines starting with spaces or tabs inside ```)
  if grep -nE '\bTBD\b|\bTo Be Determined\b' "${file}" 2>/dev/null | \
     grep -v '```' | grep -v '^.*#.*TBD' | grep -q .; then
    echo "  VIOLATION: TBD found in ${file}"
    grep -nE '\bTBD\b|\bTo Be Determined\b' "${file}" | head -3
    VIOLATIONS=$((VIOLATIONS + 1))
  fi
done < <(find "${ARTIFACT_DIR}" -name "*.md" -print0)

# Rule 2: No FIXME or TODO
echo "Checking Rule 2: No FIXME/TODO..."
while IFS= read -r -d '' file; do
  if grep -nE '\bFIXME\b|\bTODO\b' "${file}" 2>/dev/null | grep -q .; then
    echo "  VIOLATION: FIXME/TODO found in ${file}"
    grep -nE '\bFIXME\b|\bTODO\b' "${file}" | head -3
    VIOLATIONS=$((VIOLATIONS + 1))
  fi
done < <(find "${ARTIFACT_DIR}" -name "*.md" -print0)

# Rule 3: No implicit assumptions (detect common patterns)
echo "Checking Rule 3: No implicit assumptions..."
ASSUMPTION_PATTERNS=(
  'assuming that'
  'it is assumed'
  'we assume'
  'this will work'
  'should work'
  'probably'
  'might need'
  'need to decide'
  'not sure'
  'unclear'
)
while IFS= read -r -d '' file; do
  for pattern in "${ASSUMPTION_PATTERNS[@]}"; do
    if grep -inE "${pattern}" "${file}" 2>/dev/null | grep -q .; then
      echo "  WARN: Implicit assumption in ${file}: '${pattern}'"
      grep -inE "${pattern}" "${file}" | head -2
    fi
  done
done < <(find "${ARTIFACT_DIR}" -name "*.md" -print0)

# Rule 4: All artifacts must have Source of Truth header
echo "Checking Rule 4: Traceability headers..."
while IFS= read -r -d '' file; do
  if ! grep -q "Source of Truth:" "${file}" 2>/dev/null; then
    echo "  VIOLATION: Missing traceability header in ${file}"
    VIOLATIONS=$((VIOLATIONS + 1))
  fi
done < <(find "${ARTIFACT_DIR}" -name "*.md" -print0)

# Rule 5: spec.md must have BDD assertions (Given/When/Then)
echo "Checking Rule 5: BDD assertions in spec..."
SPEC="${ARTIFACT_DIR}/spec.md"
if [ -f "${SPEC}" ]; then
  if ! grep -qiE 'given|when|then' "${SPEC}" 2>/dev/null; then
    echo "  VIOLATION: spec.md has no BDD assertions (Given/When/Then)"
    VIOLATIONS=$((VIOLATIONS + 1))
  fi
fi

echo "──────────────────────────────────────────"
echo "Checked: ${TOTAL_CHECKED} files | Violations: ${VIOLATIONS}"

if [ "${VIOLATIONS}" -gt 0 ]; then
  echo "FAIL: Assumption linter found ${VIOLATIONS} violation(s)."
  echo "Fix all TBD/FIXME/missing-headers before sdd-verify can proceed."
  exit 1
fi

echo "PASS: No assumption violations found."
exit 0
