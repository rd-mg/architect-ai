import glob
import re

def patch_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    if 'Risk: LOW | MEDIUM | HIGH' in content:
        print(f"Skipping {filepath}, already patched")
        return

    pattern = re.compile(r'## Task Format.*?## Research Procedure', re.DOTALL)
    
    new_block = """## Task Format

Each task must include a risk rating.

```markdown
- [ ] {number} {action verb} {target}
      Acceptance: {condition}
      Risk: LOW | MEDIUM | HIGH
      Risk-reason: {required only when HIGH — one sentence explaining why}
```

### Risk classification

| Level | Criteria |
|-------|----------|
| `LOW` | Single file, additive only, no behavior change. |
| `MEDIUM` | Multi-file, new logic, no external contract change. |
| `HIGH` | API contract, public interface, schema, security. |

When risk is HIGH, `Risk-reason` is **mandatory**.

## Research Procedure"""
    
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

