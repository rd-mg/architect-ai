# Design: Phase 3 - Skill Registry v3 Tiers + Researcher Universal + bash-expert/fish

## Architecture
The Skill Registry (v3) uses a declarative "Context Kubernetes" model via a YAML manifest (`.atl/skill-manifest.yaml`). Skills are divided into 3 Tiers (Foundation, Context Activated, On-Demand) to eliminate context bloat caused by indiscriminately loading `bridge:always` skills. Tier 1 skills are merged into a single `_shared/foundation.md` file at install time by the Go installer.

Research is centralized through a single `researcher` agent (Tier 3) that implements the canonical priority chain. All L1 orchestrators and SDD phase agents delegate to researcher instead of implementing their own routing. This eliminates duplicated research logic across the ecosystem.

The researcher operates as a stateless query-response agent: receives (query, scope_hint), returns (summary, sources, confidence). The priority chain is sequential with early exit — each step is skipped if the previous yielded sufficient results.

## FMEA Matrix
| Component | Failure Mode | Effect | Likelihood | Severity | RPN | Mitigation |
|---|---|---|---|---|---|---|
| Researcher | Single point of failure for all research | All agents lose investigation capability | 1 | 4 | 4 | Inline fallback: Engram + rg available without delegation. |
| Shell Detection | Fish detection fails on non-standard $SHELL paths | Wrong syntax generated | 1 | 2 | 2 | Fallback to bash syntax. Check both $SHELL and `fish --version`. |
| Context7 MCP | Context7 server unavailable | Research chain incomplete | 2 | 2 | 4 | Skip and proceed to next source. Log warning. |
| Ripgrep Patterns | Domain pattern too aggressive | Relevant results filtered out | 2 | 2 | 4 | Always include a `--type-not-add` escape hatch. Manual override via raw rg. |
| Tier 2 Injection | Context detection misses | Useful skill not loaded | 2 | 2 | 4 | Explicit fallback allowed: orchestrator can retry if sub-agent complains. |

## Approach
1. Implement Go installer `internal/skill/registry/generator.go` to create manifest and `foundation.md`.
2. Implement `skill-resolver.md` v3.0 to use the new tiered approach.
3. Rewrite `researcher/SKILL.md` with structured priority chain and scope_hint routing.
4. Add `bash-expert-fish` patterns: detect shell, generate correct syntax.
5. Create `ripgrep/domains/` with per-language pattern files.
6. Demote `mcp-notebooklm-orchestrator` to Tier 3 (On-Demand).

## Contracts
- Researcher input: `{query: string, scope_hint: "code"|"docs"|"api"|"concept"}`.
- Researcher output: `{summary: string, sources: [{type, ref, excerpt}], confidence: float}`.
- Shell detector: `detectShell()` → `"bash"` | `"fish"` | `"zsh"`.
- Ripgrep domain resolver: `domainPatterns(language)` → `{include, exclude, typeFlags}`.

## Key Decisions
- **Researcher as L2 (not L1)**: Researcher is a specialized executor, not an orchestrator. It receives tasks and returns results — no routing authority.
- **scope_hint over auto-detection**: Explicit hints are more reliable than inferring research scope from query text. Reduces ambiguity.
- **Fish as first-class**: Not a "compatibility mode" — full Fish syntax generation when detected. Rationale: significant developer population uses Fish as default shell.
