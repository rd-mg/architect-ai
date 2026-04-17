# Phase Protocol: sdd-onboard

## Dependencies
- **Reads**: project context from sdd-init (if exists)
- **Writes**: potentially runs full SDD cycle end-to-end

## Cognitive Posture
+++Socratic — Guided question-driven walkthrough for new users.

## Model
sonnet — conversational guidance

## When Triggered
- User invokes `/sdd-onboard` explicitly
- First-time usage in a new project

## Procedure (Orchestrator Handles This — Not a Single Sub-Agent)

1. **Check for prior init**:
   - `mem_search(query: "sdd-init/{project}")`
   - If not found, run `sdd-init` silently first

2. **Ask key questions** (Socratic style):
   - What are you trying to build?
   - What's the scope — a new feature, refactor, bug fix, exploration?
   - Any hard constraints the agent should know (can't break X, must support Y)?
   - Preferred execution mode: automatic or interactive?
   - Preferred artifact store: engram, openspec, hybrid, or none?

3. **Walk through each phase** in interactive mode:
   - Explain what the phase does (1-2 sentences)
   - Run the phase
   - Show result summary
   - Ask "continue to next phase, refine this one, or stop?"

4. **Provide resources**:
   - At the end, point the user to:
     - `docs/cognitive-modes.md` — understanding postures
     - `docs/caveman-integration.md` — understanding output compression
     - `.atl/skill-registry.md` — what skills are active
     - The generated artifacts in Engram or `openspec/changes/`

## No Explicit Sub-Agent Launch

This phase is NOT a single sub-agent invocation. The orchestrator runs the
full SDD cycle interactively, using the normal phase protocols, but with
extra conversational framing between phases.

## Artifact Store: (user-selected in step 2)

## Persistence

The orchestrator persists a note that onboarding was completed:

```
mem_save(
  title: "sdd-onboard/{project}/completed",
  topic_key: "sdd-onboard/{project}/completed",
  type: "workflow-state",
  project: "{project}",
  content: "Onboarding completed. User preferences: execution_mode={mode}, artifact_store={mode}. First change: {change-name}."
)
```

## Return Envelope per sdd-phase-common.md Section D

## Result Processing

- Cache user preferences for the session
- No further phases needed — user is now onboarded

## Failure Handling

- If user stops mid-onboarding → save state, can resume later
- If a phase fails during onboarding → explain in user-friendly terms, suggest recovery
- If user seems overwhelmed → offer to skip to a specific phase or exit onboarding
