---
name: bash-expert
trigger: "When writing Bash scripts, hardening shell automation, or building interactive command-line pipelines in Bash or Fish terminals."
description: Handles Bash shell scripting and shell-safe automation. While specialized in Bash scripts, it is shell-aware and provides compatible patterns for interactive environments like Fish.
license: Apache-2.0
metadata:
  author: rd-mg
  version: "1.1"
---

## Purpose
Use this skill when shell execution is the correct medium. While it enforces 
strict Bash discipline for scripts, it also ensures that interactive commands 
generated for the current session (e.g., in a Fish terminal) are compatible 
with the host environment to prevent syntax errors.

## When to Use
- User asks to write or refactor a Bash script.
- User asks to improve shell safety, quoting, traps, or error handling.
- User asks to build automation with native Unix pipelines.
- User asks to review or harden Bash code against common shell pitfalls.
- User asks for Linux CLI workflow automation where Bash is a natural fit.
- Do not use when: the task requires portability to a non-shell runtime, or a
  more suitable language is clearly required for complex application logic.

## Critical Patterns
- Start Bash scripts with a robust header unless the file has a documented
  reason not to use one.
- ShellCheck compliance is required. Any suppression must be explicit and
  justified near the suppression point.
- Quote variables by default.
- Use `[[ ]]` for Bash-specific conditionals and `[ ]` only when true POSIX
  shell compatibility is required.
- Never use `eval`.
- Prefer native command pipelines for intermediate data shaping before adding a
  second language.
- Clean up temporary resources with traps.
- Keep functions scoped and readable; use `local` in Bash functions.

### Robust script header
```bash
#!/bin/bash
set -euo pipefail
IFS=$'\n\t'
```

### ShellCheck standards
Every script must pass a strict linter. Exclusions (`SCXXXX`) MUST be justified.
- **Quoting**: ALWAYS use quotes for variables (`"$var"`).
- **Conditionals**: Use `[[ ]]` for Bash-specific scripts. For POSIX `/bin/sh`
  portability, use `[ ]` and ensure compatibility.
- **Arithmetic**: For floating-point operations in portable scripts, prefer
  `awk` over `bc`.
- **Functions**: Declare variables as `local` within functions or use clear
  naming conventions where portability prevents `local`.
- **Eval**: NEVER use `eval`. Zero exceptions.

### Cleanup traps
Ensure cleanup of temporary files upon failure or termination.
```bash
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT ERR
```

### Interactive Shell Compatibility (Fish)
When generating one-liners for the current terminal, check if the shell is 
Fish. Fish is NOT POSIX-compliant.
- **Variables**: Use `set -x VAR val` instead of `export VAR=val`.
- **Command Sub**: Use `(command)` instead of `$(command)`.
- **Logic**: Use `; and` or `; or` instead of `&&` or `||` in older versions 
  (though modern Fish supports `&&`/`||`, `test` is safer).
- **Redirection**: Use `> file 2>&1` for portability; avoid Bash-specific `&>`.

### Pipeline-first philosophy
Prefer native pipelines (`grep | awk | sort`) over processing with Python or AI for intermediate data transformations.

## Inputs
- Script goal or operational task.
- Existing shell file or command sequence.
- Target shell constraints: Bash-only or POSIX-sensitive.
- Expected inputs, outputs, and failure modes.
- Any environment or filesystem cleanup requirements.

## Outputs
- One Bash-oriented implementation or refactor.
- ShellCheck-friendly script structure.
- Safe error handling and cleanup behavior.
- Native pipeline usage where appropriate.

## Steps
1. Confirm the shell environment (e.g., `bash` script vs. `fish` interactive).
2. Add or preserve a strict script header when the file is Bash-specific.
3. Structure conditionals, loops, and functions using ShellCheck-friendly
   idioms.
4. Quote variables and command substitutions defensively.
5. Add cleanup traps when temporary files or directories are used.
6. For interactive commands, apply Fish-compatible syntax if detected.
7. Prefer native pipelines for intermediate transformations.
8. Avoid `eval` and replace it with explicit command construction.
9. Run or reason against ShellCheck expectations before considering the script
   complete.

## Code Examples
Positive example (Bash Script):
```bash
#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT ERR

count_lines() {
  local file_path="$1"
  wc -l < "$file_path"
}
```

Positive example (Fish Interactive):
```fish
set -x PROJECT_ROOT (pwd); and ls $PROJECT_ROOT
```

Negative example:
```bash
#!/bin/bash
tmp=$(mktemp -d)
eval "$user_input"
if [ $flag = yes ]; then
  cat $file | grep foo
fi
```

## Commands
```bash
# Lint a Bash script
shellcheck path/to/script.sh

# Run a script with Bash explicitly
bash path/to/script.sh

# Find shell scripts in the repository
rg --files -g '*.sh'
```

## Resources
- `AGENTS.md` — canonical repository rules and skill formatting requirements.
- `skills/skill-creator/SKILL.md` — baseline for skill structure and packaging.

## Guardrails
- Do not use `eval`.
- Do not leave temporary files without cleanup traps.
- Do not suppress ShellCheck warnings without justification.
- Do not introduce unquoted variables in commands or conditionals.
- Do not add Python for simple text transformations that native pipelines can solve.

## Validation Checklist
- [ ] Frontmatter includes required repository fields.
- [ ] The skill preserves the robust Bash header pattern.
- [ ] ShellCheck standards are explicit and enforceable.
- [ ] Cleanup traps are documented for temporary resources.
- [ ] Interactive shell compatibility (Fish) is addressed.
- [ ] Pipeline-first guidance is explicit.
- [ ] Positive and negative examples exist.
- [ ] Commands section exists.
- [ ] The skill is registered in `AGENTS.md` and `.atl/skill-registry.md`.
