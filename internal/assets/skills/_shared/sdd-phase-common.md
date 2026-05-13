# SDD Phase — Common Protocol

Boilerplate identical across all SDD phase skills. Sub-agents MUST load this alongside their phase-specific SKILL.md.

Executor boundary: every SDD phase agent is an EXECUTOR, not an orchestrator. Do the phase work yourself. Do NOT launch sub-agents, do NOT call `delegate`/`task`, and do NOT bounce work back unless the phase skill explicitly says to stop and report a blocker.

## A. Skill Loading

1. Check if the orchestrator injected a `## Project Standards (auto-resolved)` block in your launch prompt. If yes, follow those rules — they are pre-digested compact rules from the skill registry. **Do NOT read any SKILL.md files.**
2. If no Project Standards block was provided, check for `SKILL: Load` instructions. If present, load those exact skill files.
3. If neither was provided, search for the skill registry as a fallback:
   a. `mem_search(query: "skill-registry", project: "{project}")` — if found, `mem_get_observation(id)` for full content
   b. Fallback: read `.atl/skill-registry.md` from the project root if it exists
   c. From the registry's **Compact Rules** section, apply rules whose triggers match your current task.
4. If no registry exists, proceed with your phase skill only.

NOTE: the preferred path is (1) — compact rules pre-injected by the orchestrator. Paths (2) and (3) are fallbacks for backwards compatibility. Searching the registry is SKILL LOADING, not delegation. If `## Project Standards` is present, IGNORE any `SKILL: Load` instructions — they are redundant.

## A2. Cognitive Posture Reception (NEW in v2)

Check if the orchestrator injected a `+++{Posture}` block at the top of your prompt. Valid postures:

- `+++Socratic` — formulate 3 clarifying questions before acting
- `+++Critical` — evaluate claims against evidence
- `+++Systemic` — analyze 2nd/3rd order effects
- `+++Adversarial` — actively try to break the artifact
- `+++Pragmatic` — minimum viable solution, no gold-plating
- `+++Forensic` — trace evidence chains, mark validation state per fact

If a posture is present:
1. Read and internalize the posture before reading task instructions
2. Apply the posture's behavior throughout your work
3. Reflect the posture in your return envelope (Socratic returns questions; Adversarial returns findings; etc.)

If multiple postures are present (e.g., Critical + Systemic for sdd-design), apply BOTH simultaneously — do not choose one.

If no posture is present, proceed with default analytical behavior.

## A3. Tool Availability Check (NEW in v2)

Check for an `## Available Tools` block in your launch prompt. If present, it lists the tools the orchestrator has verified are operational:

```
## Available Tools
- mem_search, mem_save, mem_get_observation: Engram memory (verified available)
- context7_resolve, context7_get_docs: Context7 documentation
```

Use ONLY the listed tools. If you would normally call `mem_save` but Engram is NOT listed, fall back to the behavior for the `none` persistence mode (return results inline, do not attempt `mem_save`).

If no `## Available Tools` block exists, assume standard availability and proceed normally.

## B. Artifact Retrieval (Engram Mode)

**CRITICAL**: `mem_search` returns 300-char PREVIEWS, not full content. You MUST call `mem_get_observation(id)` for EVERY artifact. **Skipping this produces wrong output.**

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

Every phase that produces an artifact MUST persist it. Skipping this BREAKS the pipeline — downstream phases will not find your output.

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

File was already written during the phase's main step. No additional action needed.

**REQUIRED: state.yaml maintenance**:
If mode is `openspec` or `hybrid`, you MUST update `openspec/changes/{change-name}/state.yaml` to reflect the new phase status and artifact.

#### Atomic Write Pattern (state.yaml)
1. Write to `state.yaml.tmp`
2. Rename to `state.yaml`

#### Validation
After every write to `state.yaml`, call `architect-ai sdd-status {change-name}`. If it fails, fix the file immediately.

**Reminder for human**: If using `openspec` mode, add a `git add openspec/` reminder in the return envelope so the user knows to commit artifacts.

### Hybrid mode

Do BOTH: write the file to the filesystem AND call `mem_save` as above.
Follow the **state.yaml maintenance** rules from the OpenSpec section above.

### None mode

Return result inline only. Do not write any files or call `mem_save`.

## D. Return Envelope

Every phase MUST return a structured envelope to the orchestrator.

### D1. Human-Readable Summary

Provide a concise summary of the work done:

- `status`: `success`, `partial`, or `blocked`
- `executive_summary`: 1-3 sentence summary (LITE caveman style, user-facing)
- `artifacts`: list of artifact keys/paths written
- `risks`: risks discovered, or "None"
- `next_recommended`: the next SDD phase to run, or "none"
- `cognitive_posture`: the posture applied, e.g., `+++Socratic` or `none` (NEW in v2)

### D2. Unified Handshake Schema (JSON-in-Markdown)

**MANDATORY**: You MUST include the following JSON block at the very end of your response. This is used by the **Observer Agent ("Gentleman Angel")** for automated monitoring and state-sync.

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
- `skill_resolution`: how skills were loaded — `injected`, `fallback-registry`, `fallback-path`, or `none`

### Size Budget (NEW in v2)
...
The `executive_summary` MUST be under 100 words. The full artifact MUST respect the phase-specific word limit:

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

If your artifact exceeds the budget, compress via:
1. Remove redundant framing
2. Collapse lists into tables
3. Use fragments instead of full sentences in checklists
4. Move supporting detail to Engram and reference via topic_key

NEVER exceed budget by more than 20%. If you cannot fit the content, split the work into multiple smaller artifacts or escalate as `partial` status.

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

Every tool call MUST align with the permission tiers set by the orchestrator or user.

- **ALWAYS**: Safe, read-only, or idempotent actions (e.g., `mem_search`, `rg`, `ls`).
- **ASK FIRST**: Mutative, external, or resource-heavy actions (e.g., `mem_save` first time, `run_command` with side effects, `git push`).
- **NEVER**: Destructive actions without multi-factor/manual override (e.g., `rm -rf /`, `git push --force` to main).

If you are unsure of a tool's tier, DEFAULT to **ASK FIRST**.

## F. Anti-Overengineering Constraints 

Follow these constraints to ensure simple, maintainable solutions:

- **F1: Abstraction Gate**: Do NOT create an interface/wrapper unless there are at least 2 distinct implementations planned.
- **F2: Scale Check**: Design for the project's CURRENT scale, not its theoretical future scale.
- **F3: Dependency Minimization**: Prefer built-in language features over adding new external libraries.
- **F4: Cognitive Load**: A single function/component should ideally fit on one screen and do one thing.
- **F5: YAGNI (You Ain't Gonna Need It)**: Do NOT implement features or edge-case handling until they are explicitly required by a spec.

## G. Caveman Output Mode 

When producing artifacts, apply caveman compression per the persona file:

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

You are an autonomous agent expected to reach the goal. If you encounter roadblocks:

1. **Unresolved Placeholders**: If the orchestrator passes raw variables like `{change-name}` or `{project}`, DO NOT fail. Determine them dynamically:
   - For `{change-name}`: Use `glob` on `openspec/changes/*` or `mem_search(query: "sdd")` to find the active change.
   - For `{project}`: Use the current directory name or git root.
2. **Tool/Step Failures**: If a specific tool fails (e.g., `mem_search` returns nothing), do not get stuck. Use an alternative tool (e.g., `glob`, `read`, or `grep` the filesystem) or gracefully omit the step if it is non-blocking.
3. **Missing Artifacts**: If a required artifact is missing, try to reconstruct it from recent context or git history. If it is impossible to reconstruct, record it as a deviation/risk and proceed with the remaining tasks.
4. **Resilience**: Never give up on the first error. Attempt a fix, try a workaround, or skip the failing non-critical step to complete the core objective.
