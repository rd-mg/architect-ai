---
name: bash-expert
description: >
  Safe, portable shell scripting. Every sub-agent that runs bash MUST
  follow these patterns: strict mode, quoting discipline, error handling,
  POSIX-vs-bashism awareness. Protects against classic pitfalls
  (unquoted globs, word splitting, trap handling, signal propagation).
license: MIT
bridge: always
applies-when: "any delegation that runs bash/sh scripts, uses pipes, or chains commands"
metadata:
  author: rd-mg
  version: "1.0"
---

# Bash Expert — Mandatory Skill

## Why this is `bridge: always`

Shell mistakes are silent. A script runs, exits 0, and deletes the wrong directory. We force this skill into every sub-agent to make the common traps explicit before they happen.

---

## Compact Rules (injected into sub-agent prompts)

### Strict mode — always

```bash
#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'
```

- `-e` — exit on any error
- `-u` — error on unset variable
- `-o pipefail` — propagate errors from pipes
- `IFS=$'\n\t'` — stop word-splitting on spaces

### Quote every variable

```bash
# WRONG
rm -rf $TMPDIR

# RIGHT
rm -rf "$TMPDIR"
```

If `$TMPDIR` is empty or has spaces, the unquoted version can `rm -rf /` or fail weirdly. Always double-quote unless you specifically need word-splitting.

### Check for required tools first

```bash
for cmd in jq rg fd; do
  command -v "$cmd" >/dev/null || { echo "missing: $cmd" >&2; exit 127; }
done
```

### Capture stdout AND stderr when debugging

```bash
out=$(command 2>&1) || { echo "command failed: $out" >&2; exit 1; }
```

### Temp files — cleanup with trap

```bash
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT INT TERM
# ...use $tmp...
```

### Avoid backticks — use `$()`

```bash
# WRONG
version=`git describe`

# RIGHT
version=$(git describe)
```

### Test conditions — use `[[ ]]`, not `[ ]` (in bash)

```bash
if [[ "$var" == *substr* ]]; then ...
```

`[[ ]]` is safer: doesn't word-split, supports pattern matching, no escape-the-regex pain.

### Check exit status explicitly

```bash
if ! command; then
  echo "command failed" >&2
  exit 1
fi
```

---

## POSIX vs bashism awareness

If the script must run in `sh` (Alpine, Dash, busybox), avoid:

- `[[ ]]` → use `[ ]`
- Arrays `arr=(a b c)` → use positional params or multiple vars
- `$'\n'` → use printf
- `read -r -a` → use `while read` loop
- `<<<` here-strings → use `echo | cmd`

Mark scripts explicitly with the right shebang:
- `#!/usr/bin/env bash` for bashisms (must have bash)
- `#!/bin/sh` for POSIX (portable)

---

## Pipes and subshells

### `set -e` doesn't cross pipe boundaries without `pipefail`

```bash
set -e
false | true
echo "reached"  # YES, because `true` succeeded
```

Fix: `set -o pipefail` — exit status of pipe is rightmost non-zero.

### Subshell variable changes don't propagate

```bash
count=0
echo "a b c" | while read word; do count=$((count+1)); done
echo "$count"  # 0, NOT 3 — while runs in a subshell
```

Fix: use process substitution:
```bash
while read word; do count=$((count+1)); done < <(echo "a b c")
echo "$count"  # 3
```

---

## Signal handling

When the user hits ctrl+c, your script needs to clean up:

```bash
cleanup() {
  rm -rf "$tmp" 2>/dev/null
  [[ -n "${child_pid:-}" ]] && kill "$child_pid" 2>/dev/null
}
trap cleanup EXIT INT TERM
```

For long-running subprocesses, propagate the signal:

```bash
long_running_cmd &
child_pid=$!
wait "$child_pid"
```

---

## Common anti-patterns

**❌ Parsing `ls`**
```bash
for f in $(ls); do ...
```
Breaks on filenames with spaces. Use:
```bash
for f in *; do ...
```
or `find ... -print0 | xargs -0`.

**❌ `cat file | grep`**
```bash
cat log.txt | grep ERROR
```
Useless use of cat:
```bash
grep ERROR log.txt
```

**❌ `(( $var > 5 ))` without quotes on empty**
If `$var` is empty, `(( > 5 ))` is a syntax error. Guard:
```bash
if [[ -n "$var" && "$var" -gt 5 ]]; then ...
```

**❌ Silent failures**
```bash
some_command 2>/dev/null
```
If you redirect stderr to `/dev/null`, you lose the error message forever. Log it:
```bash
some_command 2>>/tmp/myscript.log
```

---

## When the orchestrator should NOT delegate to bash

- Data manipulation with structure (JSON, XML, YAML) → use `jq`, `yq`, or Python, not sed/awk.
- Cross-platform scripts needed on Windows → write a Go tool or Python script, not a bash script.
- Long-running services → not a shell job. Use systemd / launchd / supervisord.

---

## See also

- `ripgrep/SKILL.md` — the `rg` command is preferred over grep
- `_shared/research-routing.md` — how shell fits in the research priority
