# SDD Phase — Common Protocol

Boilerplate identical across all SDD phase skills. Sub-agents MUST load this alongside phase-specific SKILL.md.

Executor boundary: SDD phase agents are EXECUTOR, not an orchestrator. Do phase work yourself. Do NOT launch sub-agents, do NOT call `delegate`/`task`, do NOT bounce work back unless phase skill explicitly says stop and report blocker.

## A. Skill Loading

1. Check if orchestrator injected `## Project Standards (auto-resolved)` block in launch prompt. If yes, follow those rules — pre-digested compact rules from skill registry. **Do NOT read any SKILL.md files.**
2. If no Project Standards block, check for `SKILL: Load` instructions. If present, load those exact skill files.
3. If neither: fallback search for skill registry:
   a. `mem_search(query: "skill-registry", project: "{project}")` — if found, `mem_get_observation(id)` for full content
   b. Fallback: read `.atl/skill-registry.md` from project root if exists
   c. From registry **Compact Rules** section, apply rules whose triggers match current task.
4. No registry: proceed with phase skill only.

Preferred path is (1) — compact rules pre-injected by orchestrator. Paths (2) and (3) are fallbacks. Searching registry is SKILL LOADING, not delegation. If `## Project Standards` present, IGNORE any `SKILL: Load` instructions — redundant.

## A2. Cognitive Posture Reception

Check if orchestrator injected `+++{Posture}` block at top of prompt. Valid postures:

- `+++Socratic` — formulate 3 clarifying questions before acting
- `+++Critical` — evaluate claims against evidence
- `+++Systemic` — analyze 2nd/3rd order effects
- `+++Adversarial` — actively try to break the artifact
- `+++Pragmatic` — minimum viable solution, no gold-plating
- `+++Forensic` — trace evidence chains, mark validation state per fact

If posture present:
1. Internalize posture before reading task instructions
2. Apply posture's behavior throughout work
3. Reflect posture in return envelope (Socratic returns questions; Adversarial returns findings; etc.)

Multiple postures present (e.g., Critical + Systemic for sdd-design): apply BOTH simultaneously — do not choose one.

No posture: proceed with default analytical behavior.

## A4. Sequential Thinking Harmonization

Use `sequential_thinking` tool (if available) to execute Cognition Mode and formulate Adaptive Reasoning evaluation.

1. **Hierarchy**: `sequential_thinking` is mechanical engine.
   - **Cognition Modes (The Lens)**: Postures like `+++Divergent` dictate *how* you think. Use `branchId` in sequential thoughts to explore alternative hypotheses in divergent postures.
   - **Adaptive Reasoning (The Structure)**: Dictates final output depth and formatting. Use thoughts to silently evaluate D1-D4 and formulate response strategy.
2. **Supplement, Not Replace**: Sequential thinking steps are internal logic. MUST still produce final synthesized response (e.g., `[MODE N | ...]` and return envelope) as required.

## A3. Tool Availability Check

Check for `## Available Tools` block in launch prompt. Lists tools orchestrator verified operational:

```
## Available Tools
- mem_search, mem_save, mem_get_observation: Engram memory (verified available)
- context7_resolve, context7_get_docs: Context7 documentation
```

Use ONLY listed tools. If normally would call `mem_save` but Engram NOT listed, fall back to `none` persistence mode (return results inline, do not attempt `mem_save`).

No `## Available Tools` block: assume standard availability.

## B. Artifact Retrieval (Engram Mode)

**CRITICAL**: `mem_search` returns 300-char PREVIEWS, not full content. MUST call `mem_get_observation(id)` for EVERY artifact. **Skipping produces wrong output.**

**Run all searches in parallel** — do NOT search sequentially.

```
mem_search(query: "sdd/{change-name}/{artifact-type}", project: "{project}") → save ID
```

Then **run all retrievals in parallel**:

```
mem_get_observation(id: {saved_id}) → full content (REQUIRED)
```

Do NOT use search previews as source material.

## C. Artifact Persistence

Every phase producing artifact MUST persist it. Skipping BREAKS pipeline — downstream phases won't find output.

### Engram mode

```
mem_save(
  title: "sdd/{change-name}/{artifact-type}",
  topic_key: "sdd/{change-name}/{artifact-type}",
  type: "architecture",
  project: "{project}",
  content: "{your full artifact markdown}"
)
```

`topic_key` enables upserts — saving again updates, not duplicates.

### OpenSpec mode

File already written during phase's main step. No additional action needed.

**REQUIRED: state.yaml maintenance**:
If mode `openspec` or `hybrid`, MUST update `openspec/changes/{change-name}/state.yaml` to reflect new phase status and artifact.

#### Atomic Write Pattern (state.yaml)
1. Write to `state.yaml.tmp`
2. Rename to `state.yaml`

#### Validation
After every write to `state.yaml`, call `architect-ai sdd-status {change-name}`. If fails, fix file immediately.

**Reminder for human**: If using `openspec` mode, add `git add openspec/` reminder in return envelope so user knows to commit artifacts.

### Hybrid mode

Do BOTH: write file to filesystem AND call `mem_save` as above. Follow **state.yaml maintenance** rules from OpenSpec section above.

### None mode

Return result inline only. Do not write files or call `mem_save`.

## D. Return Envelope

Every phase MUST return structured envelope to orchestrator.

### D1. Human-Readable Summary

- `status`: `success`, `partial`, or `blocked`
- `executive_summary`: 1-3 sentence summary (LITE caveman style, user-facing)
- `artifacts`: list of artifact keys/paths written
- `risks`: risks discovered, or "None"
- `next_recommended`: next SDD phase to run, or "none"
- `cognitive_posture`: posture applied, e.g., `+++Socratic` or `none`

### D2. Unified Handshake Schema (JSON-in-Markdown)

**MANDATORY**: MUST include this JSON block at very end of response. Used by **Observer Agent ("Gentleman Angel")** for automated monitoring and state-sync.

```json
{
  "status": "success|partial|blocked",
  "change": "{change-name}",
  "phase": "{current-phase}",
  "artifacts": ["path/to/artifact1", "key:topic/key/2"],
  "defects_found": 0,
  "empirical_proof": "Brief description of test result, log entry, or evidence",
  "estimated_tokens": 1200,
  "nonce": "{session-uuid-if-provided}"
}
```

### D3. Detailed Metrics & Triage (Internal)

- `detailed_report`: (optional) full phase output
- `findings_triage`: (mandatory for sdd-verify) summary object `{ blocking: N, warning: M, suggestion: K }`
- `pre_mortem`: (mandatory for sdd-propose) summary of top risks and viability score
- `skill_resolution`: how skills loaded — `injected`, `fallback-registry`, `fallback-path`, or `none`

### Size Budget

`executive_summary` MUST be under 100 words. Full artifact MUST respect phase-specific word limit:

| Phase | Word Budget |
|-------|-------------|
| sdd-propose | 450 |
| sdd-design | 800 |
| sdd-tasks | 530 |
| sdd-explore | 600 |
| sdd-spec | 1000 |
| sdd-apply (apply-progress) | 400 |
| sdd-verify | 700 |
| sdd-archive | 200 |

If artifact exceeds budget, compress via:
1. Remove redundant framing
2. Collapse lists into tables
3. Use fragments instead of full sentences in checklists
4. Move supporting detail to Engram and reference via topic_key

NEVER exceed budget by more than 20%. If cannot fit content, split into multiple smaller artifacts or escalate as `partial` status.

### Example envelope

```markdown
**Status**: success
**Summary**: Proposal created for `add-dark-mode`. Defined scope, approach, and rollback plan.
**Pre-mortem**: Viability 9/10. Risk: theme flickering (Med) - mitigated via hydration guard.
**Artifacts**: Engram `sdd/add-dark-mode/proposal` | `openspec/changes/add-dark-mode/proposal.md`
**Next**: sdd-spec or sdd-design
**Risks**: None
**Skill Resolution**: injected — 3 skills (react-19, typescript, tailwind-4)
**Cognitive Posture**: +++Critical
**Estimated Tokens**: 850
```

```markdown
**Status**: success
**Summary**: Verified `add-dark-mode` implementation. Matches specs and design.
**Triage**: { blocking: 0, warning: 1, suggestion: 2 }
**Artifacts**: Engram `sdd/add-dark-mode/verify` | `openspec/changes/add-dark-mode/verify.md`
**Next**: sdd-archive
**Risks**: None
**Skill Resolution**: injected
**Estimated Tokens**: 1200
```

(Other values for Skill Resolution: `fallback-registry`, `fallback-path`, or `none — no registry found`)

## E. Permission Tiers

Every tool call MUST align with permission tiers set by orchestrator or user.

- **ALWAYS**: Safe, read-only, or idempotent actions (e.g., `mem_search`, `rg`, `ls`).
- **ASK FIRST**: Mutative, external, or resource-heavy actions (e.g., `mem_save` first time, `run_command` with side effects, `git push`).
- **NEVER**: Destructive actions without multi-factor/manual override (e.g., `rm -rf /`, `git push --force` to main).

Unsure of tool tier → DEFAULT to **ASK FIRST**.

## F. Anti-Overengineering Constraints

- **F1: Abstraction Gate**: Do NOT create interface/wrapper unless at least 2 distinct implementations planned.
- **F2: Scale Check**: Design for project's CURRENT scale, not theoretical future scale.
- **F3: Dependency Minimization**: Prefer built-in language features over new external libraries.
- **F4: Cognitive Load**: Single function/component should fit on one screen, do one thing.
- **F5: YAGNI**: Do NOT implement features or edge-case handling until explicitly required by spec.

## G. Caveman Output Mode

Apply caveman compression per persona file:

- **Artifacts stored to Engram / OpenSpec**: ULTRA mode. Telegraphic. Fragments OK. Drop articles and filler.
- **`executive_summary` field in return envelope**: LITE mode. No filler, grammar intact, professional.
- **Code, commits, PRs**: Normal English (no compression).
- **Security warnings, irreversible action confirmations**: Normal English (clarity over brevity).

ULTRA example (artifact content):
```
Change: add dark mode toggle. Affected: settings.py, theme.js. New dep: colorContext hook. Risk: cache invalidation on switch. Rollback: feature flag.
```

LITE example (executive_summary):
```
Proposal created for add-dark-mode change. Affects settings.py and theme.js.
Main risk is cache invalidation on theme switch; rollback via feature flag.
```

## H. Fallback and Recovery Behavior (MANDATORY)

Autonomous agent expected to reach goal. On roadblocks:

1. **Unresolved Placeholders**: If orchestrator passes raw `{change-name}` or `{project}`, determine dynamically:
   - `{change-name}`: Use `glob` on `openspec/changes/*` or `mem_search(query: "sdd")` to find active change.
   - `{project}`: Use current directory name or git root.
2. **Tool/Step Failures**: If tool fails (e.g., `mem_search` returns nothing), use alternative tool (`glob`, `read`, or `grep` filesystem) or gracefully omit non-blocking step.
3. **Missing Artifacts**: If required artifact missing, try to reconstruct from recent context or git history. If impossible, record as deviation/risk and proceed with remaining tasks.
4. **Resilience**: Never give up on first error. Attempt fix, try workaround, or skip failing non-critical step to complete core objective.
