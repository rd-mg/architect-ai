# Phase 2 — Skill Registry Kernel Refactor

> **Cognitive Mode**: +++Systemic +++Empirical  
> **CCLD Tag**: `[PHASE-2][SKILL-REGISTRY][PARALLEL-IO][INDEXED]`  
> **Status**: BLOCKED until Phase 0 Task B completes  
> **Estimated Duration**: 1–2 sessions  
> **Depends On**: `audit/skill-registry-io.md`

---

## 2.1 Objective

Refactor the skill registry system on two fronts:

1. **Go layer** (`cli/skill_registry.go`): Parallelize the three independent I/O collection calls using `errgroup`. Add concurrent-safe deduplication.
2. **Markdown layer** (`.atl/skill-registry.md`): Introduce a section-indexed format that allows agents to load only the section they need instead of the full file.

**Target Outcome**: Skill registry generation latency reduced by ~60% on large Odoo projects. Agent context injection reduced from full-file read to section-targeted read.

---

## 2.2 Root Cause (from Phase 0 Audit)

### 2.2.1 Sequential I/O in `WriteLocalSkillRegistry`

```go
// internal/cli/skill_registry.go — current
func WriteLocalSkillRegistry(projectRoot string) error {
    // ...
    userSkills, err := collectUserSkills(homeDir)     // I/O: walk ~/.gemini/skills/
    if err == nil {
        allSkills = append(allSkills, userSkills...)   // sequential append
    }
    projectSkills, err := collectProjectSkills(projectRoot) // I/O: walk .agent/skills/
    if err == nil {
        allSkills = append(allSkills, projectSkills...)      // sequential append
    }
    overlaySkills, overlayAssets, err := collectOverlayContent(projectRoot) // I/O: walk .atl/overlays/
    if err == nil {
        allSkills = append(allSkills, overlaySkills...)      // sequential append
        assets = append(assets, overlayAssets...)
    }
    // then deduplicateSkills — also sequential
}
```

**Problem**: All three collect operations are independent filesystem walks. Combined they read O(30–50) files on a typical Odoo project. No goroutine fan-out.

### 2.2.2 Full-File Registry Read by Agents

`.atl/skill-registry.md` format:
```markdown
## System Skills
| Trigger | Skill | Path |

## SharedRule Skills
...
## Project Skills
...
## Overlay Skills
...
## User Skills
...
```

**Problem**: Every agent delegated by the orchestrator reads the entire registry file to find its relevant compact rules. No section anchors. No per-kind index.

### 2.2.3 `deduplicateSkills` — Append-Unsafe for Concurrent Input

```go
func deduplicateSkills(skills []skillEntry) []skillEntry {
    seen := make(map[string]int)
    result := make([]skillEntry, 0, len(skills))
    for _, s := range skills {
        ...
    }
    return result
}
```

**Problem**: Operates on a single slice built by sequential appends. Cannot accept concurrent input without a sync wrapper.

---

## 2.3 Go Layer Refactoring

### 2.3.1 Concurrent Collection with `errgroup`

```go
// internal/cli/skill_registry.go — NEW

type collectionResult struct {
    skills []skillEntry
    assets []assetEntry // only for overlay
    err    error
    origin string
}

func WriteLocalSkillRegistry(projectRoot string) error {
    homeDir, err := osUserHomeDir()
    if err != nil {
        return fmt.Errorf("resolve user home directory: %w", err)
    }

    // Launch all three collections concurrently
    results := make([]collectionResult, 3)
    g, _ := errgroup.WithContext(context.Background())

    g.Go(func() error {
        skills, err := collectUserSkills(homeDir)
        results[0] = collectionResult{skills: skills, err: err, origin: "user"}
        return nil // non-fatal: log and continue
    })

    g.Go(func() error {
        skills, err := collectProjectSkills(projectRoot)
        results[1] = collectionResult{skills: skills, err: err, origin: "project"}
        return nil
    })

    g.Go(func() error {
        skills, assets, err := collectOverlayContent(projectRoot)
        results[2] = collectionResult{skills: skills, assets: assets, err: err, origin: "overlay"}
        return nil
    })

    _ = g.Wait()

    // Fan-in: merge results (now safe — all goroutines done)
    var allSkills []skillEntry
    var allAssets []assetEntry
    for _, r := range results {
        if r.err == nil {
            allSkills = append(allSkills, r.skills...)
            allAssets = append(allAssets, r.assets...)
        }
        // log r.err if non-nil but continue (non-fatal)
    }

    allSkills = deduplicateSkills(allSkills)  // safe: single-threaded fan-in
    // ... rest of registry building unchanged
}
```

**Notes**:
- `errgroup` errors are non-fatal for individual collectors (a missing user skills dir is not an error).
- `results` slice is pre-allocated by index — no mutex needed for the goroutine writes.
- `deduplicateSkills` is called after fan-in, still single-threaded — no change needed.

---

### 2.3.2 `EnsureProjectRegistryReady` — Parallel Overlay Bootstrap

```go
// internal/cli/skill_registry.go — MODIFY

// BootstrapResult already exists. Add parallel overlay install.

func EnsureProjectRegistryReady(projectRoot string) (BootstrapResult, error) {
    // ... existing detection logic ...
    
    if result.IsOdooProject && len(overlaysToInstall) > 1 {
        // Install multiple overlays in parallel
        g, ctx := errgroup.WithContext(context.Background())
        var mu sync.Mutex
        
        for _, overlay := range overlaysToInstall {
            overlay := overlay
            g.Go(func() error {
                action, err := installOverlay(ctx, projectRoot, overlay)
                if err != nil {
                    return fmt.Errorf("overlay %q: %w", overlay.Name, err)
                }
                mu.Lock()
                result.Actions[overlay.Name] = action
                mu.Unlock()
                return nil
            })
        }
        
        if err := g.Wait(); err != nil {
            return result, err
        }
    }
    
    return result, WriteLocalSkillRegistry(projectRoot)
}
```

---

### 2.3.3 `collectOverlayContent` — Internal Walk Parallelism

For Odoo projects with multiple overlays (e.g., `odoo-development-skill` contains 6+ sub-skills), the walk can be parallelized per sub-skill directory.

```go
// internal/cli/skill_registry.go — MODIFY collectOverlayContent

func collectOverlayContent(projectRoot string) ([]skillEntry, []assetEntry, error) {
    overlayRoot := filepath.Join(projectRoot, ".atl", "overlays")
    entries, err := os.ReadDir(overlayRoot)
    if err != nil {
        return nil, nil, nil // not found is OK
    }
    
    type overlayResult struct {
        skills []skillEntry
        assets []assetEntry
    }
    
    results := make([]overlayResult, len(entries))
    g, _ := errgroup.WithContext(context.Background())
    
    for i, entry := range entries {
        if !entry.IsDir() {
            continue
        }
        i, entry := i, entry
        g.Go(func() error {
            skills, assets := walkOverlayDir(filepath.Join(overlayRoot, entry.Name()), entry.Name())
            results[i] = overlayResult{skills: skills, assets: assets}
            return nil
        })
    }
    
    _ = g.Wait()
    
    var allSkills []skillEntry
    var allAssets []assetEntry
    for _, r := range results {
        allSkills = append(allSkills, r.skills...)
        allAssets = append(allAssets, r.assets...)
    }
    
    return allSkills, allAssets, nil
}
```

---

## 2.4 Markdown Registry Format — Indexed Sections

### 2.4.1 Problem with Current Format

Current `.atl/skill-registry.md`:
- Flat sequential sections
- No machine-readable section anchors
- No per-skill compact rules index separate from skill path
- Agents read entire file for every sub-agent spawn

### 2.4.2 New Format: Section-Anchored with Compact Rules Index

```markdown
<!-- architect-ai:registry:version:2 -->

# Skill Registry

**Delegator use only.** Sub-agents do NOT read this file.

<!-- architect-ai:registry:index:start -->
## Quick Index
| Skill | Kind | Trigger | Section |
|---|---|---|---|
| ripgrep | System | pattern-search | #skill-ripgrep |
| bash-expert | System | bash,shell | #skill-bash-expert |
| sdd-apply | SharedRule | sdd-apply,implement | #skill-sdd-apply |
| go-testing | Project | go,test | #skill-go-testing |
<!-- architect-ai:registry:index:end -->

<!-- architect-ai:registry:system:start -->
## System Skills

### ripgrep {#skill-ripgrep}
**Trigger**: pattern-search, rg, ripgrep  
**Compact Rules**: Use `rg` instead of `grep`. Never use raw `find | xargs grep`. Pattern: `rg "pattern" --type go -l`. JSON output: `rg --json`.

### bash-expert {#skill-bash-expert}
**Trigger**: bash, shell, script  
**Compact Rules**: ...
<!-- architect-ai:registry:system:end -->

<!-- architect-ai:registry:sharedrule:start -->
## SharedRule Skills
...
<!-- architect-ai:registry:sharedrule:end -->

<!-- architect-ai:registry:project:start -->
## Project Skills
...
<!-- architect-ai:registry:project:end -->

<!-- architect-ai:registry:overlay:start -->
## Overlay Skills
...
<!-- architect-ai:registry:overlay:end -->

<!-- architect-ai:registry:user:start -->
## User Skills
...
<!-- architect-ai:registry:user:end -->
```

**Benefits**:
- Orchestrator can extract only the `Quick Index` section to find which skills a sub-agent needs.
- Sub-agent prompts inject only the compact rules from the targeted `#skill-X` section.
- Section anchors are grep-able: `rg "architect-ai:registry:project:start" .atl/skill-registry.md`.

### 2.4.3 Registry Writer Update

```go
// internal/cli/skill_registry.go — MODIFY WriteLocalSkillRegistry section builder

// Section markers as constants
const (
    registryVersion    = "<!-- architect-ai:registry:version:2 -->"
    sectionStart       = "<!-- architect-ai:registry:%s:start -->"
    sectionEnd         = "<!-- architect-ai:registry:%s:end -->"
    indexStart         = "<!-- architect-ai:registry:index:start -->"
    indexEnd           = "<!-- architect-ai:registry:index:end -->"
)

// Build Quick Index section first
func buildQuickIndex(skillsByKind map[string][]skillEntry) string {
    var b strings.Builder
    b.WriteString(indexStart + "\n## Quick Index\n\n")
    b.WriteString("| Skill | Kind | Trigger | Anchor |\n|---|---|---|---|\n")
    for _, kind := range []string{"System", "SharedRule", "Project", "Overlay", "User"} {
        for _, s := range skillsByKind[kind] {
            anchor := "#skill-" + strings.ToLower(strings.ReplaceAll(s.Name, " ", "-"))
            b.WriteString(fmt.Sprintf("| %s | %s | %s | [link](%s) |\n", s.Name, kind, s.Trigger, anchor))
        }
    }
    b.WriteString(indexEnd + "\n")
    return b.String()
}
```

---

## 2.5 Orchestrator Skill Loading — Targeted Injection Protocol

Update the `general-orchestrator.md` skill resolution section to use the new indexed format:

```markdown
## Skill Resolution (Updated Protocol)

When building a sub-agent prompt:
1. Read ONLY `<!-- architect-ai:registry:index:start -->` → `<!-- architect-ai:registry:index:end -->` from `.atl/skill-registry.md` to get the Quick Index.
2. Identify which skills are mandatory (always inject: ripgrep, bash-expert, context-guardian) and which are task-matched.
3. For each required skill, extract ONLY that skill's section using the anchor: read from `### {skill-name} {#skill-{slug}}` to the next `###` heading.
4. Inject only extracted compact rules into the sub-agent prompt — NOT the full SKILL.md content.

This replaces the current `{content of _shared/context-mode-routing-policy.md}` pattern.
```

---

## 2.6 Files to Create / Modify

| File | Action | Notes |
|---|---|---|
| `internal/cli/skill_registry.go` | MODIFY | Concurrent collection, section-anchored writer |
| `internal/cli/skill_registry_test.go` | MODIFY | Add concurrency tests, race detector tests |
| `.atl/skill-registry.md` | REGENERATE | New format (via `architect-ai skill-registry`) |
| `.agent/skills/_shared/skill-resolver.md` | MODIFY | Targeted section injection protocol |
| `internal/assets/generic/skill-resolver.md` (and per-agent variants) | MODIFY | Same targeted injection update |

---

## 2.7 Testing Requirements

| Test | Verifies |
|---|---|
| `TestWriteLocalSkillRegistryConcurrent` | Three collectors complete; order-independent; results merged correctly |
| `TestWriteLocalSkillRegistryRace` | `-race` detector clean |
| `TestRegistryIndexSectionParseable` | Quick Index section can be extracted by anchor grep |
| `TestCollectOverlayContentParallel` | Multiple overlays collected concurrently; no duplicate entries |
| `TestDeduplicateSkillsPreservesKindPriority` | Project > Overlay > User precedence maintained after concurrent fan-in |

---

## 2.8 Acceptance Criteria

- [ ] `WriteLocalSkillRegistry` runs three collectors in parallel — proven by `-race` test
- [ ] New registry format includes `Quick Index` section parseable without reading full file
- [ ] Section markers are machine-readable (`rg "architect-ai:registry:.*:start"`)
- [ ] Existing registry content (all skill entries) is unchanged — only format updated
- [ ] `RunSkillRegistry` CLI command regenerates registry in new format on `--refresh-overlays`
- [ ] `skill-resolver.md` updated to targeted injection protocol
- [ ] All existing `skill_registry_test.go` tests pass

---

## 2.9 Sub-Agent Delegation

```
[PHASE-2 ORCHESTRATOR]
    │
    ├── [2A] go-writer-agent     → skill_registry.go concurrent WriteLocalSkillRegistry
    ├── [2B] go-writer-agent     → skill_registry.go concurrent collectOverlayContent
    ├── [2C] go-writer-agent     → skill_registry.go section-anchored writer + Quick Index
    ├── [2D] go-tester-agent     → skill_registry_test.go concurrency + race tests (depends 2A-2C)
    ├── [2E] md-writer-agent     → .atl/skill-registry.md regenerate in new format
    └── [2F] md-writer-agent     → skill-resolver.md targeted injection protocol
```

2A, 2B, 2C can launch in parallel (different functions).  
2D launches after 2A–2C.  
2E, 2F can launch in parallel after 2C.
