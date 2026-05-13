<!-- architect-ai:prompt-caching-anchor:start -->
# SDD Phase Protocol: sdd-design
Project: architect-ai
Adapter: Antigravity
Version: 1.1
<!-- architect-ai:prompt-caching-anchor:end -->

## Dependencies
- **Reads**: proposal artifact, spec artifact (if exists)
- **Writes**: `design` artifact

## Cognitive Posture
+++Critical + +++Systemic — Architecture needs both rigor and system view.

## Model
opus — architectural decisions

## Sub-Agent Launch Template

```
+++Critical
Evaluate objectively based on evidence. For each claim made or implied:
(1) What evidence supports it? (2) What evidence contradicts it?
(3) What alternative explanation exists?

+++Systemic
Analyze 2nd and 3rd order effects. What breaks elsewhere? What new
dependencies are created? What becomes harder to change later?

## Project Standards (auto-resolved)
{matching compact rules}

## Available Tools
{verified tool list}

## Phase: sdd-design

Task: Produce the architecture design for "{change-name}". Based on proposal
and spec (if present), produce a design document covering:

## Mandatory Sections
## Mandatory Sections
- Architecture diagram (ASCII or Mermaid)
- Module/component boundaries
- Data flow & Interface contracts
- State management & Error propagation
- Integration points with existing code
- **Poka-Yoke Checklist**: (1) Type Safety (2) State Integrity (3) Resource Cleanup (4) Interface Simplicity. Mark each as //N/A.
- **Architecture Decision Records (ADR)**: Table of key decisions made, rationale, and consequences.
- **YAGNI Gate**: Table of proposed abstractions. For each, state: (1) Current need (2) Anticipated implementations (3) Cost of direct implementation. If only 1 implementation exists, abstraction is REJECTED.
- Alternative designs considered and why rejected
- Open questions (if any remain)


## Empirical Verification Loop (+++Empirical)
- **MANDATORY**: Before concluding, you MUST perform an empirical verification of your findings/artifacts.
- Examples: run a script, check a file, verify a tool output, or perform a manual check of the logic.
- Record the evidence in the `empirical_proof` field of the return handshake.

## Artifact Store: {mode}

## Persistence (MANDATORY)
mem_save(
  title: "sdd/{change-name}/design",
  topic_key: "sdd/{change-name}/design",
  type: "architecture",
  project: "{project}",
  content: "{your design markdown}"
)

## Size Budget: 800 words max

## Return Envelope & Compliance per sdd-phase-common.md (Sections A-F)
```

## Result Processing

- Validate all mandatory sections present
- **Decision Check**: Reject if `Alternative designs considered` is empty.
- **Simplicity Check**: Reject if `YAGNI Gate` shows premature abstractions.
- **Poka-Yoke Check**: Ensure checklist is filled and items with  have justification.
- Update state: `specifying` → `designing`
- Next recommended: `sdd-tasks`

## Failure Handling

- If sub-agent cannot identify integration points → return `partial`, suggest sdd-explore round
- If design conflicts with active constraints in Context Pack → return `blocked`, escalate to user
- If Open Questions is non-empty → present to user, wait for resolution before sdd-tasks
