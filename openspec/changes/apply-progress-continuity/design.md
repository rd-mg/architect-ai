# Design: Apply-Progress Continuity

## Architecture Decisions

### Decision: Symmetric Resumption
- **Approach**: Orchestrators will evaluate `artifact_store` mode before every `sdd-apply` launch.
- **Why**: Prevents "blind" implementation resumption where an agent starts from task 1.1 even if 1.5 was finished in a prior session.

### Decision: Filesystem Authority (Hybrid Mode)
- **Approach**: In `hybrid` mode, if store drift is detected, the `apply-progress.md` file is used as the baseline.
- **Why**: Users can easily edit local files. Engram is more difficult to manual-correct; therefore, the local file is a more reliable "human-controlled" source of truth.

### Decision: Pre-Launch Guard
- **Approach**: Skill-level gate in `sdd-verify`.
- **Why**: Prevents misleading "Pass" results when implementation is only 10% complete.

## Implementation Details

### Merge Algorithm (Pseudo-code)
```
IF hybrid:
  FILE = read(apply-progress.md)
  MEM = read(engram/apply-progress)
  COMBINED = merge(FILE, MEM) // FILE is base
  write(apply-progress.md, COMBINED)
  save(engram, COMBINED)
```

### Verification Guard Logic
```
STATUS = get_phase_status("sdd-apply")
IF STATUS != "completed":
  ERROR("Apply incomplete")
```
