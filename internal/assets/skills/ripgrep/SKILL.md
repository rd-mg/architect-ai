---
name: ripgrep
description: >
  Fast, correct pattern search in code. Always preferred over grep for
  performance and for respecting .gitignore by default. Every sub-agent
  that searches files MUST use ripgrep (rg), not grep, find+grep, or
  language-level regex loops over file trees.
license: MIT
bridge: always
applies-when: "any delegation that involves file search, pattern matching, or code-wide refactor discovery"
metadata:
  author: rd-mg
  version: "1.0"
---

# Ripgrep (rg) — Mandatory Skill

## Why this is `bridge: always`

Every SDD phase touches file search at some point:
- **Explore**: find all callers of a function
- **Propose/Spec**: survey existing patterns
- **Apply**: find edit targets
- **Verify**: check that a pattern is NOT present (negative assertions)

Using `grep -r` or `find ... -exec grep` on a medium-sized repo is 10-50× slower than `rg` and ignores `.gitignore`, causing hits in `node_modules/`, `.venv/`, `dist/`, etc. that pollute the sub-agent's context.

Marking this skill `bridge: always` means the orchestrator injects it into **every** sub-agent prompt, unconditionally.

---

## Compact Rules (injected into sub-agent prompts)

```
Use `rg` for all code search. NEVER use `grep -r` or `find ... -exec grep`.
  - Case-insensitive: `rg -i`
  - Fixed string (no regex): `rg -F`
  - Show context: `rg -C 3`
  - Count only: `rg -c`
  - File list only: `rg -l`
  - Multi-line pattern: `rg -U`
  - Limit by path: `rg "foo" path/to/dir` (ripgrep does NOT walk up)
  - Include/exclude globs: `rg -g '!**/*.test.go'`
  - Respects .gitignore by default. To search ignored files: `rg -uuu`
  - JSON output for piping: `rg --json`

Do NOT escape `/`, `:`, or alphanumerics in patterns. ripgrep uses Rust regex syntax (PCRE-lite).

For binary safety: `rg -a` to force text mode. Default skips binary files.

When reporting findings, quote file:line:content, never just a file path.
```

---

## Typical patterns

### Find all usages of a symbol
```bash
rg -w "computeTotal" --type go
# -w = word boundary (doesn't match computeTotalPrice)
# --type go restricts to .go files
```

### Find negative — "nothing imports this"
```bash
rg -l "import.*mypkg" || echo "no callers"
```

### Multi-line match — function signatures
```bash
rg -U "func\s+\w+\s*\([^)]*\)\s+error"
```

### Count occurrences across the repo
```bash
rg -c "TODO" | sort -t: -k2 -n -r | head
```

### Find + replace (stream to xargs sed — read-only rg)
```bash
rg -l "old_name" | xargs sed -i 's/old_name/new_name/g'
```

---

## When NOT to use ripgrep

- **Structured code queries**: use the language server / tree-sitter / `gopls`, not regex. Example: finding all implementers of an interface.
- **Symbol rename**: use IDE refactor, not rg | sed (regex can't distinguish `foo.bar` from `foo_bar` in the wrong contexts).
- **Cross-file semantic checks**: regex is text. Use `go vet`, `eslint`, etc.

If the task needs semantic understanding, the orchestrator should note that in the sub-agent prompt and the sub-agent should use the language toolchain instead.

---

## Installation check

Sub-agents should never fail silently on missing `rg`. First command in any search-heavy sub-agent:

```bash
command -v rg >/dev/null || {
  echo "ripgrep not installed. Install: brew install ripgrep / apt-get install ripgrep / pacman -S ripgrep" >&2
  exit 127
}
```

---

## See also

- `bash-expert/SKILL.md` — shell patterns that complement ripgrep usage
- `_shared/research-routing.md` — when ripgrep is the right tool vs. NotebookLM/Context7
