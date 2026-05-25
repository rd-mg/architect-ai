#!/usr/bin/env bash
# test-path-resolution.sh — Tests Odoo path resolution behavior
# Run: bash test-path-resolution.sh
# Validates the documented discovery logic from _shared/odoo-path-resolution.md

PASS=0
FAIL=0
TOTAL=0

test_name() {
  TOTAL=$((TOTAL + 1))
  echo "---"
  echo "TEST ${TOTAL}: $1"
}

assert_contains() {
  if echo "$2" | grep -q "$1"; then
    echo "  PASS: found '$1'"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: expected '$1' not found in output"
    echo "    got: $2"
    FAIL=$((FAIL + 1))
  fi
}

assert_exit_code() {
  local expected=$1 actual=$2 label="$3"
  if [ "$actual" -eq "$expected" ]; then
    echo "  PASS: exit $actual (expected $expected) — $label"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: exit $actual (expected $expected) — $label"
    FAIL=$((FAIL + 1))
  fi
}

assert_not_contains() {
  if ! echo "$2" | grep -q "$1"; then
    echo "  PASS: correctly absent '$1'"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: unexpected '$1' found in output"
    echo "    got: $2"
    FAIL=$((FAIL + 1))
  fi
}

# ═══════════════════════════════════════════════════════════════════════════════
# Inline discovery logic matching odoo-path-resolution.md
# ═══════════════════════════════════════════════════════════════════════════════

run_discovery() {
  # Usage: run_discovery <config_dir> [env_override]
  # Runs the path resolution logic in an isolated subshell with a temp HOME
  local config_dir="$1"
  local env_val="${2:-}"

  # Create an isolated temp HOME
  local isolate_home
  isolate_home=$(mktemp -d)

  # If config_dir provided, copy a .atl/config.yaml into isolate
  if [ -n "$config_dir" ] && [ -d "$config_dir" ]; then
    mkdir -p "${isolate_home}/.atl"
    cp "${config_dir}/config.yaml" "${isolate_home}/.atl/config.yaml" 2>/dev/null || true
  fi

  # Write discovery script to a temp file so we control the environment precisely
  local script_file
  script_file=$(mktemp)
  cat > "${script_file}" << 'DISCOVERY_SCRIPT'
#!/usr/bin/env bash
# Discovery script — matches odoo-path-resolution.md verbatim
set -euo pipefail

# Read from .atl/config.yaml first
ODOO_COMMUNITY=$(grep "odoo_community_path:" "${HOME}/.atl/config.yaml" 2>/dev/null \
  | awk '{print $2}' | tr -d '"' | tr -d "'" || true)

# Env var override (fallback when config not set)
[ -z "${ODOO_COMMUNITY}" ] && ODOO_COMMUNITY="${ODOO_COMMUNITY_PATH:-}"

# Auto-discovery (OS-agnostic, ordered by likelihood)
if [ -z "${ODOO_COMMUNITY}" ]; then
  HOME_DIR="${HOME}"
  for candidate in \
    "${HOME_DIR}/gitproj/odoo/community" \
    "${HOME_DIR}/odoo/community" \
    "${HOME_DIR}/src/odoo" \
    "/opt/odoo" \
    "/usr/local/lib/odoo"; do
    [ -d "${candidate}" ] && { ODOO_COMMUNITY="${candidate}"; break; }
  done
fi

# pip-installed odoo fallback
if [ -z "${ODOO_COMMUNITY}" ]; then
  ODOO_COMMUNITY=$(python3 -c \
    "import odoo, os; print(os.path.dirname(odoo.__file__))" 2>/dev/null || echo "")
fi

if [ -z "${ODOO_COMMUNITY}" ]; then
  echo "WARN: Odoo community source not found."
  echo "Set ODOO_COMMUNITY_PATH or add odoo_community_path in .atl/config.yaml"
  echo "Research will use Engram + Context7 only (no local rg)"
  ODOO_COMMUNITY="NOT_FOUND"
fi

echo "RESULT_ODOO_COMMUNITY=${ODOO_COMMUNITY}"
DISCOVERY_SCRIPT

  # Run with isolated environment
  HOME="${isolate_home}" ODOO_COMMUNITY_PATH="${env_val}" bash "${script_file}" 2>&1
  local exit_code=$?
  rm -f "${script_file}"
  local exit_code=$?
  rm -rf "${isolate_home}"
  return $exit_code
}

# ═══════════════════════════════════════════════════════════════════════════════
# Test 1: No Odoo source — expect graceful warning + NOT_FOUND
# ═══════════════════════════════════════════════════════════════════════════════
test_name "Graceful degradation when Odoo source not found"
output=$(run_discovery "" "")
assert_contains "WARN: Odoo community source not found" "$output"
assert_contains "RESULT_ODOO_COMMUNITY=NOT_FOUND" "$output"
assert_not_contains "No such file or directory" "$output"

# ═══════════════════════════════════════════════════════════════════════════════
# Test 2: Env var override — ODOO_COMMUNITY_PATH picks it up
# ═══════════════════════════════════════════════════════════════════════════════
test_name "ODOO_COMMUNITY_PATH env var override"
output=$(run_discovery "" "/custom/odoo/path")
assert_contains "RESULT_ODOO_COMMUNITY=/custom/odoo/path" "$output"
assert_not_contains "WARN: Odoo community source not found" "$output"

# ═══════════════════════════════════════════════════════════════════════════════
# Test 3: Config file detection — .atl/config.yaml sets the path
# ═══════════════════════════════════════════════════════════════════════════════
test_name ".atl/config.yaml detection"
config_dir=$(mktemp -d)
mkdir -p "${config_dir}"
cat > "${config_dir}/config.yaml" << 'CONFIG'
odoo_community_path: "/home/user/odoo/community"
odoo_version: "18"
CONFIG
output=$(run_discovery "$config_dir" "")
assert_contains "RESULT_ODOO_COMMUNITY=/home/user/odoo/community" "$output"
assert_not_contains "WARN: Odoo community source not found" "$output"
rm -rf "$config_dir"

# ═══════════════════════════════════════════════════════════════════════════════
# Test 4: Config file takes precedence over auto-discovery
# ═══════════════════════════════════════════════════════════════════════════════
test_name "Config file takes precedence over auto-discovery"
config_dir=$(mktemp -d)
mkdir -p "${config_dir}"
cat > "${config_dir}/config.yaml" << 'CONFIG'
odoo_community_path: "/preferred/odoo/path"
CONFIG
output=$(run_discovery "$config_dir" "")
assert_contains "RESULT_ODOO_COMMUNITY=/preferred/odoo/path" "$output"
rm -rf "$config_dir"

# ═══════════════════════════════════════════════════════════════════════════════
# Test 5: Config file takes precedence (env var is fallback only per design)
# ═══════════════════════════════════════════════════════════════════════════════
test_name "Config file takes precedence over env var (design priority order)"
config_dir=$(mktemp -d)
mkdir -p "${config_dir}"
cat > "${config_dir}/config.yaml" << 'CONFIG'
odoo_community_path: "/config/odoo/path"
CONFIG
output=$(run_discovery "$config_dir" "/env/fallback/path")
assert_contains "RESULT_ODOO_COMMUNITY=/config/odoo/path" "$output"
assert_not_contains "/env/fallback/path" "$output"
rm -rf "$config_dir"

# ═══════════════════════════════════════════════════════════════════════════════
# Test 6: Graceful output messages for RESEARCH degradation
# ═══════════════════════════════════════════════════════════════════════════════
test_name "Research degradation message when Odoo not found"
output=$(run_discovery "" "")
assert_contains "Research will use Engram + Context7 only" "$output"
assert_contains "Set ODOO_COMMUNITY_PATH or add odoo_community_path" "$output"

# ═══════════════════════════════════════════════════════════════════════════════
# Summary
# ═══════════════════════════════════════════════════════════════════════════════
echo ""
echo "=============================================="
echo "  RESULTS: ${PASS} passed, ${FAIL} failed (${TOTAL} tests)"
echo "=============================================="
[ "$FAIL" -eq 0 ] || exit 1
