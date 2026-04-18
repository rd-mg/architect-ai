# Phase Protocol: sdd-spec

## Dependencies
- **Reads**: proposal artifact
- **Writes**: `spec` artifact (detailed specifications per capability)

## Cognitive Posture
+++Systemic — Detect cross-domain dependencies, 2nd/3rd order effects.

## Model
opus → sonnet (for writing the structured output)

## Sub-Agent Launch Template

```
+++Systemic
Analyze 2nd and 3rd order effects before specifying. What OTHER subsystems
could break? What new dependencies are created? What becomes harder to change
later? Prefer reversible decisions over optimal-but-irreversible ones.

## Project Standards (auto-resolved)
{matching compact rules}

## Available Tools
{verified tool list}

## Phase: sdd-spec

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.

Task: Translate the proposal for "{change-name}" into detailed specifications.
One spec entry per capability listed in proposal.md's Capabilities section.

## Mandatory Sections Per Capability
- Purpose (one sentence)
- Preconditions
- Behavior (what the system does)
- Postconditions
- Error handling
- Invariants (what must stay true across the change)
- Test hooks (how this can be verified)

## Artifact Store: {mode}

## Persistence (MANDATORY)
mem_save(
  title: "sdd/{change-name}/spec",
  topic_key: "sdd/{change-name}/spec",
  type: "architecture",
  project: "{project}",
  content: "{your spec markdown}"
)

## Size Budget: 1000 words max

## Return Envelope per sdd-phase-common.md Section D
```

## Result Processing

- Validate one capability per spec section
- Check each capability has all 7 mandatory fields
- Update state: `proposing` → `specifying`
- Next recommended: `sdd-design` or `sdd-tasks`

## Failure Handling

- If proposal's Capabilities section was "unclear" → return `blocked` with reason
- If a capability can't be specified without more investigation → flag as risk, do NOT fabricate behavior
