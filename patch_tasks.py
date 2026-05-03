import glob
import re

def patch_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    if 'Each task must include a risk rating.' in content:
        print(f"Skipping {filepath}, already patched")
        return

    # Replace from #### Task File Format up to ### Task Writing Rules
    pattern = re.compile(r'#### Task File Format.*?(?=### Task Writing Rules)', re.DOTALL)
    
    new_block = """#### Task File Format

Each task must include a risk rating.

```markdown
# Tasks: {Change Title}

## Phase 1: {Phase Name} (e.g., Infrastructure / Foundation)

- [ ] 1.1 {Concrete action — what file, what change}
      Acceptance: {condition}
      Risk: LOW | MEDIUM | HIGH
      Risk-reason: {required only when HIGH — one sentence explaining why}
- [ ] 1.2 {Concrete action}
      Acceptance: {condition}
      Risk: LOW | MEDIUM | HIGH
      Risk-reason: {required only when HIGH — one sentence explaining why}

## Phase 2: {Phase Name} (e.g., Core Implementation)

- [ ] 2.1 {Concrete action}
      Acceptance: {condition}
      Risk: LOW | MEDIUM | HIGH
      Risk-reason: {required only when HIGH — one sentence explaining why}
```

### Risk classification

| Level | Criteria |
|-------|----------|
| `LOW` | Single file, additive only, no behavior change. Examples: add log line, rename var, add comment, write a new unit test. |
| `MEDIUM` | Multi-file, new logic, no external contract change. Examples: add internal function, refactor method, add CLI flag. |
| `HIGH` | API contract, public interface, database schema, security boundary, irreversible data operation, or external integration. Examples: new MCP tool, schema migration, auth change, public function rename. |

When risk is HIGH, `Risk-reason` is **mandatory**. It must name the specific risk: "Changes the public MCP tool API consumed by all agents" or "Deletes rows from the state table — no rollback."

"""
    
    new_content, count = pattern.subn(new_block, content)
    
    if count == 0:
        print(f"Failed to find match in {filepath}")
        return

    with open(filepath, 'w') as f:
        f.write(new_content)
    print(f"Patched {filepath}")

files = glob.glob('internal/assets/*/sdd-phase-protocols/sdd-tasks.md')
files.extend(glob.glob('internal/assets/*/agents/sdd-tasks.md'))
files.extend(glob.glob('internal/assets/*/commands/sdd-tasks.md'))
for f in files:
    patch_file(f)

