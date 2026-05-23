# Phase Protocol: sdd-spec

## Dependencies
- **Reads**: proposal artifact
- **Writes**: `spec` artifact (detailed specifications per capability)

## Cognitive Posture
+++Systemic — Detect cross-domain dependencies, 2nd/3rd order effects.

## Model
opus → sonnet (for writing structured output)

## Sub-Agent Launch Template

```
+++Systemic
Analyze 2nd and 3rd order effects before specifying. What OTHER subsystems
could break? What new dependencies created? What becomes harder to change
later? Prefer reversible decisions over optimal-but-irreversible ones.

## Project Standards (auto-resolved)
{matching compact rules}

## Available Tools
{verified tool list}

## Phase: sdd-spec

Task: Translate proposal for "{change-name}" into detailed specifications.
One spec entry per capability listed in proposal.md's Capabilities section.

## Mandatory Sections Per Capability
- Purpose (one sentence)
- Preconditions
- Behavior (what system does)
- Postconditions
- Error handling
- Invariants (what must stay true across change)
- Test hooks (how this can be verified)
- **FMEA Table**: Required for external I/O or user input. Columns: Failure Mode | Impact | Severity (1-5) | Detection.
- **Sad-path BDD**: Required for any FMEA severity ≥ 3. Format: Given-When-Then for failure case.
- **UI & State Modeling**: Required for UI with async states. Use Mermaid `stateDiagram-v2` for FSM.
- **Accessibility Contract**: Required for UI. Specify keyboard navigation, ARIA roles, and focus management.

## Artifact Store: {mode}

## Persistence (MANDATORY)
mem_save(
  title: "sdd/{change-name}/spec",
  topic_key: "sdd/{change-name}/spec",
  type: "architecture",
  project: "{project}",
  content: "{your spec markdown with FMEA, BDD, FSM, and A11y where applicable}"
)

## Size Budget: 1000 words max

## Return Envelope & Compliance per sdd-phase-common.md (Sections A-F)
```

## Result Processing

- Validate one capability per spec section
- Check each capability has mandatory fields
- **Validation Gate**: Verify presence of FMEA and Sad-path BDD if I/O or User Input detected.
- **UI Gate**: Verify FSM and A11y contract if UI capabilities present.
- Update state: `proposing` → `specifying`
- Next recommended: `sdd-design` or `sdd-tasks`

## Failure Handling

- If proposal's Capabilities section was "unclear" → return `blocked` with reason
- If capability can't be specified without more investigation → flag as risk, do NOT fabricate behavior
