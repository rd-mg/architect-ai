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
  version: "2.0"
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

## Advanced Search Patterns

### AST-level Function Signature Search (Go, Python, TypeScript)
- **Go Function Search**: Find Go functions with receiver or standard signatures:
  `rg -U "func\s+\([^)]+\)\s+\w+\s*\([^)]*\)" --type go`
  `rg -w "func \w+" --type go`
- **Python Method/Class Search**: Find class definitions or methods:
  `rg "class \w+(\(\w+\))?:" --type py`
  `rg "def \w+\([^)]*\):" --type py`
- **TypeScript/JavaScript Method Search**: Find functions/classes/interfaces:
  `rg "export (class|interface|type) \w+" --type ts`
  `rg "function \w+\s*\(" --type ts --type js`

### Import and Dependency Graph Search Patterns
- **Go Imports**: Find where a specific package is imported:
  `rg -U "import\s+\(\s*[^)]*\"github.com/rd-mg/architect-ai/[^\"]+\"\s*\)" --type go`
  `rg "import\s+\"github.com/rd-mg/architect-ai/[^\"]+\"" --type go`
- **Python Imports**: Find imports of standard or local modules:
  `rg "^(import \w+|from \w+ import)" --type py`
- **TypeScript/ES6 Imports**: Find imports of packages:
  `rg "import\s+.*\s+from\s+'[^']+'" --type ts --type js`

### Multi-File Context Patterns
- **Context lines around match (A=after, B=before)**:
  `rg "pattern" -A 5 -B 3 --no-heading --type go`
- **Include line numbers and path details**:
  `rg -n --with-filename "pattern" --type go`

### Performance & Filtering Patterns
- **Target specific globs**:
  `rg "pattern" --glob "internal/skills/**"`
- **Limit results to avoid flooding context**:
  `rg "pattern" --max-count 10`
- **Select specific language type**:
  `rg "pattern" --type go --type-not-add python`
