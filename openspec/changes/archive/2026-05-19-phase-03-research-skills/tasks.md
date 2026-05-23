# Tasks: Phase 3 - Skill Registry v3 Tiers + Researcher Universal + bash-expert/fish

## Status: COMPLETED

## Source Doc
`/home/rdmachadog/Documents/fix_achitect-ai/03-phase-research-skills-extension.md`

## Implementation Tasks

### 3.1 Skill Registry v3 & Resolver
- [x] `internal/skill/registry/generator.go` — generate `.atl/skill-manifest.yaml` with Tiers 1-3
- [x] `internal/skill/registry/generator.go` — merge Tier 1 skills into `_shared/foundation.md`
- [x] `internal/skill/registry/generator_test.go` — tests for manifest generation and NotebookLM location
- [x] `internal/assets/_shared/skill-resolver.md` — v3.0 conditional injection protocol

### 3.2 Researcher Universal Skill
- [x] `internal/assets/skills/researcher/SKILL.md` — v2: single routing source for all investigation
  - [x] Routing protocol: Engram → rg local → Context7 → NLM → web
  - [x] Output contract JSON: `{status, result, source, confidence, reason_if_failed}`
  - [x] Termination rule: terminate after delivering result
  - [x] Inline fallback when researcher unavailable (≤3 rg searches)
- [x] `internal/assets/_shared/research-routing-protocol.md` — shared fragment injected into all orchestrators

### 3.3 bash-expert + fish-expert
- [x] `internal/assets/skills/bash-expert/SKILL.md` — update with fish shell rules
  - [x] Add fish section: `fish_add_path`, `set -gx`, `not string match` patterns
  - [x] Strict mode: `set -euxo pipefail` for bash, `set -e` for fish
  - [x] rg-first rule: `grep` blocked, always `rg` with `-l`, `-A`, `-B` flags
  - [x] `fd` over `find` preference documented

### 3.4 ripgrep Advanced Patterns
- [x] `internal/assets/skills/ripgrep/SKILL.md` — v2 with AST-level patterns
  - [x] Function signature patterns per language (Go, Python, TypeScript)
  - [x] Import/dependency graph search patterns
  - [x] Multi-file context patterns (`-A 5 -B 3 --no-heading`)
  - [x] Performance patterns: `--type`, `--glob`, `--max-count`

### Go Implementation (Other)
- [x] `internal/skill/researcher_router.go` — ResearchRouter struct with priority-ordered source list
- [x] `internal/skill/researcher_router_test.go` — unit tests with mock sources

### Tests
- [x] TestResearcherRouter_EngramHit — engram returns result, no rg call made
- [x] TestResearcherRouter_RgFallback — engram miss, rg finds result
- [x] TestResearcherRouter_FullChainMiss — all sources miss, graceful degradation
- [x] TestBashExpert_FishPatterns — fish syntax snippets validated
- [x] TestRipgrepPatterns_GoFunctionSearch — rg pattern finds Go function signatures
