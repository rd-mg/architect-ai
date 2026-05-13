# Audit: Skill Registry Serial I/O

## Call Chain
The `WriteLocalSkillRegistry` function in `internal/cli/skill_registry.go` executes the following collection steps sequentially:

1. `collectUserSkills(homeDir)` (filesystem walk)
2. `collectProjectSkills(projectRoot)` (filesystem walk)
3. `collectOverlayContent(projectRoot)` (filesystem walk)
4. Registry building (deduplication, grouping, markdown section injection)
5. `filemerge.WriteFileAtomic(registryPath, ...)` (IO operation)

## Evidence
Code snippet from `internal/cli/skill_registry.go`:
```go
// 1. Collect all entries
// User skills
userSkills, err := collectUserSkills(homeDir)
if err == nil {
    allSkills = append(allSkills, userSkills...)
}

// Project skills
projectSkills, err := collectProjectSkills(projectRoot)
if err == nil {
    allSkills = append(allSkills, projectSkills...)
}

// Overlay content
overlaySkills, overlayAssets, err := collectOverlayContent(projectRoot)
if err == nil {
    allSkills = append(allSkills, overlaySkills...)
    assets = append(assets, overlayAssets...)
}
```

The implementation is strictly serial; there is no use of `errgroup`, `sync.WaitGroup`, or goroutines to perform these collection steps in parallel.

## Impact
- Current: Serial execution requiring 3 consecutive filesystem walks.
- Proposed: Parallel execution using `errgroup` or `sync.WaitGroup` to perform all 3 walks concurrently.
- Estimated improvement: ~60% reduction in registry regeneration time (based on max(T_walk_user, T_walk_project, T_walk_overlay)).

## Verdict
MASTER-PLAN Phase 2 claim VALIDATED
