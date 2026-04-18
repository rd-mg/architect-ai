#!/usr/bin/env bash
# scripts/lint-language.sh
# Enforces English-only voice per docs/language-policy.md.
# Exits 0 on success, non-zero on violation.

set -euo pipefail

# Guard: must run from repo root (Judge remediation D3)
cd "$(git rev-parse --show-toplevel)" 2>/dev/null || {
  echo "ERROR: must run from inside the repo." >&2
  exit 2
}

# Spanish words that indicate voice (not just trigger data).
SPANISH_VOICE_PATTERN='\b(español|en español|gracias|por favor|ahora|también|porque|entonces|sin embargo|además|estás en lo cierto|nota|aquí|este|esta|estos|estas)\b'

# Scopes to scan (Judge remediation D5: complete scan roots)
SCAN_ROOTS=(
  "internal/assets"
  "docs"
  "README.md"
  "CONTRIBUTING.md"
  "PRD.md"
  "PRD-AGENT-BUILDER.md"
  "GEMINI.md"
  "AGENTS.md"
  "CONTRIBUTORS.md"
)

# Paths where trigger phrases are permitted.
ALLOWLIST_FILE="scripts/language-allowlist.txt"

if ! command -v rg >/dev/null 2>&1; then
  echo "ERROR: ripgrep (rg) is required. Install: https://github.com/BurntSushi/ripgrep#installation" >&2
  exit 127
fi

# Build -g exclude globs from allowlist
exclude_args=()
if [[ -f "$ALLOWLIST_FILE" ]]; then
  while IFS= read -r path || [[ -n "$path" ]]; do
    [[ -z "$path" || "$path" =~ ^# ]] && continue
    # CRLF sanitation (Judge remediation D2)
    path=${path%$'\r'}
    if [[ ! -f "$path" ]]; then
      # Warning on missing path (Judge remediation D7)
      echo "WARN: allowlist path missing: $path" >&2
    fi
    exclude_args+=(-g "!$path")
  done < "$ALLOWLIST_FILE"
fi

echo "Scanning for English-only voice violations..."
echo "  Pattern:   $SPANISH_VOICE_PATTERN"
echo "  Roots:     ${SCAN_ROOTS[*]}"
echo "  Allowlist: $ALLOWLIST_FILE (${#exclude_args[@]} path rules)"
echo

# Use rg with case-insensitive, word-boundary match.
# Returns 1 if no match (good); 0 if match (bad).
set +e
rg -i --no-heading "${exclude_args[@]}" "$SPANISH_VOICE_PATTERN" "${SCAN_ROOTS[@]}"
rc=$?
set -e

if [[ $rc -eq 0 ]]; then
  echo
  echo "VIOLATION: Spanish voice detected above."
  echo "Fix per docs/language-policy.md, or (if this is a trigger phrase)"
  echo "move it under an allowed block and update $ALLOWLIST_FILE."
  exit 1
fi

echo "OK: no English-only voice violations."
exit 0
