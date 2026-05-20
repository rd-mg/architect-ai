# Circuit Breaker Protocol

To prevent recursive loops (Ralph Loops) where an agent repeatedly fails a phase and retries indefinitely, the Circuit Breaker monitors the number of execution attempts.

## Protocol Rules

1. **Attempt Limit**: Max 3 attempts are allowed per SDD phase.
2. **Attempt Counting**:
   - The attempt count for the active phase is retrieved from `circuit_breaker.attempt_counts.{phase}` in `.atl/sdd-state.yaml`.
   - On the first run, the attempt count is 0. Upon execution start, the count is incremented to 1.
3. **Escalation Path**:
   - **Attempt 1 Failed**: The agent updates its approach, increments the count to 2, and retries.
   - **Attempt 2 Failed**: The agent requests additional context from the user, updates the count to 3, and retries.
   - **Attempt 3 Failed**: The Circuit Breaker trips. The agent:
     - Writes `status: abandoned` for the phase in `.atl/sdd-state.yaml`.
     - Adds the phase name to `circuit_breaker.abandoned_phases`.
     - Emits a Result Contract with `status: abandoned`.
     - Exits with **Exit Code 2**.
4. **Orchestrator Behavior on Trip**:
   - The `sdd-orchestrator` intercepts Exit Code 2, saves all logs/state, emits a clear diagnostic message, and **immediately halts** development flow. It does not advance.
5. **Recovery Options**:
   - The user can intervene to resolve the blocker manually and run `/sdd-continue`.
   - The user can force-forward the phase (accepting the risks) with `/sdd-ff`.
   - The user can archive the change as abandoned via `/sdd-archive --status=abandoned`.
