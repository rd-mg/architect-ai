---
name: bash-expert
description: >
  Safe, portable shell scripting for bash AND fish.
  Every sub-agent running shell MUST follow these patterns.
  Prefers rg over grep. Handles fish-specific syntax.
  Part of foundation (Tier 1) — always injected.
bridge: true
tier: foundation
version: "2.0"
---

# Shell Expert (bash + fish) v2.0

## Shell Detection (MANDATORY — first step, every script)
```bash
ACTIVE_SHELL=$(basename "${SHELL:-bash}")
# Use appropriate section below
```

## BASH Rules

### Strict mode header (every script)
```bash
#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'
```

### Variable quoting (always)
```bash
# WRONG
rm -rf $TMPDIR
echo $MY_VAR

# RIGHT
rm -rf "${TMPDIR}"
echo "${MY_VAR}"
```

### rg instead of grep (MANDATORY — grep -r forbidden)
```bash
# WRONG — slow, ignores .gitignore
grep -r "pattern" .
find . -name "*.go" -exec grep "pattern" {} \;

# RIGHT — fast, .gitignore-aware
rg "pattern" --type go
rg -l "pattern" .                          # file list only
rg -c "pattern" --type py                  # count per file
rg -w "exactFunction" --type go            # word boundary
rg -C 3 "pattern" --type go               # 3 lines context
rg --json "pattern" . | jq '.data.lines.text'  # for parsing
rg -U "multi.*\nline" --type go            # multi-line
rg -l "pattern" -g "!vendor/" -g "!node_modules/"  # exclude dirs
```

### Error handling
```bash
# Trap for cleanup
cleanup() { local e=$?; rm -f "${TMPFILE:-}"; exit "${e}"; }
trap cleanup EXIT INT TERM

# Check command availability
for cmd in rg jq; do
  command -v "${cmd}" > /dev/null || { echo "ERROR: ${cmd} not found" >&2; exit 127; }
done

# Capture stdout + stderr
output=$(some_command 2>&1) || { echo "Failed: ${output}" >&2; exit 1; }
```

### Safe file operations
```bash
# Atomic write
tmp=$(mktemp)
cat > "${tmp}" << 'EOF'
content
EOF
mv "${tmp}" "${target}"

# Check existence before read
[ -f "${file}" ] || { echo "Missing: ${file}" >&2; exit 1; }

# Never: rm -rf without quoting and validation
[ -n "${DIR}" ] && [ -d "${DIR}" ] && rm -rf "${DIR}"
```

## FISH Rules

### NO set -euo pipefail in fish
```fish
#!/usr/bin/env fish
# Fish handles errors differently — use explicit checks
```

### Error handling (fish style)
```fish
function check_cmd
    if not command -q $argv[1]
        echo "ERROR: $argv[1] not found" >&2
        exit 127
    end
end
check_cmd rg
check_cmd jq
```

### Variable syntax (fish)
```fish
# WRONG (bash style in fish)
export MY_VAR="value"
echo "$MY_VAR"

# RIGHT (fish style)
set -x MY_VAR "value"    # export
set MY_VAR "value"       # local
echo $MY_VAR             # no quotes needed in fish
set -e MY_VAR            # unset
```

### Conditionals and loops (fish)
```fish
if test -f file.txt
    echo "exists"
else if test -d dir/
    echo "directory"
else
    echo "not found"
end

for file in *.go
    echo $file
end

# Command substitution
set files (rg -l "pattern" .)
```

### Error propagation (fish)
```fish
# Fish doesn't propagate pipe errors like bash — use explicit checks
rg "pattern" . ; or begin
    echo "rg failed or no results" >&2
    exit 1
end
```

### rg in fish (same patterns, fish quoting)
```fish
rg "pattern" --type go
rg -l "pattern" .
rg --json "pattern" . | python3 -c "
import sys, json
for line in sys.stdin:
    try:
        o = json.loads(line)
        if o.get('type') == 'match':
            print(o['data']['path']['text'])
    except: pass
"
```

## Cross-Shell rg Patterns

### Pattern 1: Domain-specific search
```bash
# Backend code only (no tests, no vendor)
rg "pattern" --type go -g "!*_test.go" -g "!vendor/"

# Odoo Python models only
rg "pattern" --type py -g "models/*.py" -g "!tests/"

# XML views only (Odoo)
rg "pattern" --type xml -g "views/*.xml"
```

### Pattern 2: Negative assertion (security)
```bash
# MUST NOT exist — violation detection
rg -l "forbidden_pattern" . \
  && echo "SECURITY VIOLATION" \
  || echo "CLEAN"
```

### Pattern 3: Count to estimate scope
```bash
# Before starting work, estimate file count
AFFECTED=$(rg -l "old_function_name" --type go | wc -l)
echo "Estimated ${AFFECTED} files to change"
[ "${AFFECTED}" -gt 10 ] && echo "WARN: large scope — consider splitting"
```

### Pattern 4: JSON structured output for agent parsing
```bash
rg --json "function_name" --type go \
  | jq -r 'select(.type=="match") | "\(.data.path.text):\(.data.line_number)"'
```
