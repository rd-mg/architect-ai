# Phase Protocol: sdd-propose

## Dependencies
- **Reads**: exploration artifact (optional)
- **Writes**: `proposal` artifact

## Cognitive Posture
+++Critical — Evaluate feasibility with rigor. Identify biases and unproven assumptions before committing.

## Model
opus — architectural decisions

## Sub-Agent Launch Template

```
+++Critical
Evaluate objectively based on evidence. For each claim made or implied:
(1) What evidence supports it? (2) What evidence contradicts it?
(3) What alternative explanation exists? Do not accept aesthetic preferences
as evidence.

## Project Standards (auto-resolved)
{matching compact rules}

## Available Tools
{verified tool list}

## Phase: sdd-propose

Task: Create a change proposal for "{change-name}". Read exploration (if any).
Produce: proposal.md with scope, approach, affected areas, rollback plan,
success criteria, capabilities section.

## Step 0: Hypothesis Branching (Sequential Thinking)
- **MANDATORY**: Call `sequential_thinking` with at least 2 branches (using `branchId`) to explore alternative architectural approaches before committing to one in the proposal.

## Mandatory Sections
- Scope (what's in, what's out)
- Approach (high-level strategy)
- Affected Areas (concrete file paths where possible)
- Rollback Plan (how to undo if this fails)
- Success Criteria (observable conditions for "done")
- Capabilities (contract with sdd-spec — new/modified/none)
- **Pre-mortem**: Address: (1) What is most likely to break? (2) What dependency is the weakest link? (3) How will we detect failure in production? (4) Who is affected if this fails?
- **Open Assumptions**: Table with ≥ 2 rows (Assumption | Impact if False). If 0 assumptions, justify why.
- **Viability Score**: Score 1-15 (Sum of 3 dimensions: 1-5 Complexity, 1-5 Clarity, 1-5 Tooling). If score < 8, initialization is BLOCKED.

## Artifact Store: {mode}

## Persistence (MANDATORY)
mem_save(
  title: "sdd/{change-name}/proposal",
  topic_key: "sdd/{change-name}/proposal",
  type: "architecture",
  project: "{project}",
  content: "{your proposal markdown with Pre-mortem, Assumptions, and Viability Score}"
)

## Size Budget: 600 words max. Use bullets and tables over prose.

## Return Envelope & Compliance per sdd-phase-common.md (Sections A-F)
```

## Result Processing

- Check `executive_summary` length — must be < 100 words
- Validate `Capabilities` section is filled (not "TODO")
- **Viability Gate**: If Viability Score < 8, set status to `blocked` and recommend `sdd-explore`.
- **Pre-mortem Check**: Reject if Pre-mortem section is missing or incomplete.
- Update state: `exploring` → `proposing`
- Next recommended: `sdd-spec` or `sdd-design`

## Failure Handling

- If proposal lacks rollback plan → return `partial`, ask sub-agent to complete
- If Capabilities section says "unclear" → route to sdd-explore for more investigation
