# Tasks: OpenSpec Mode-Branching Consolidation

## Phase 1: Shared Protocol
- [x] Create `internal/assets/skills/_shared/mode-branching.md` with the canonical logic.

## Phase 2: Refactor Skills (SDD Core)
- [x] Refactor `internal/assets/skills/sdd-propose/SKILL.md`
- [x] Refactor `internal/assets/skills/sdd-spec/SKILL.md`
- [x] Refactor `internal/assets/skills/sdd-design/SKILL.md`
- [x] Refactor `internal/assets/skills/sdd-tasks/SKILL.md`
- [x] Refactor `internal/assets/skills/sdd-apply/SKILL.md`
- [x] Refactor `internal/assets/skills/sdd-verify/SKILL.md`
- [x] Refactor `internal/assets/skills/sdd-archive/SKILL.md`

## Phase 3: Refactor Skills (Supporting)
- [x] Refactor `internal/assets/skills/sdd-explore/SKILL.md`
- [x] Refactor `internal/assets/skills/sdd-init/SKILL.md`
- [x] Refactor `internal/assets/skills/sdd-onboard/SKILL.md`
- [x] Refactor `internal/assets/skills/skill-registry/SKILL.md`

## Phase 4: Verification
- [x] Run `rg "mode-branching.md" internal/assets/skills/sdd-*/SKILL.md | wc -l` (expect 10)
- [x] Run `rg "mode-branching.md" internal/assets/skills/skill-registry/SKILL.md | wc -l` (expect 1)
- [x] Verify no "IF mode is" blocks remain (except in the shared file).
