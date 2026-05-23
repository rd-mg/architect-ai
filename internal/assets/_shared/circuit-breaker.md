# Circuit Breaker Protocol

Prevent recursive loops (Ralph Loops) where agent repeatedly fails phase and retries indefinitely.

## Rules

1. **Attempt Limit**: Max 3 per SDD phase.
2. **Counting**: Retrieved from `circuit_breaker.attempt_counts.{phase}` in `.atl/sdd-state.yaml`. First run = 0, incremented to 1 on start.
3. **Escalation**:
   - **Attempt 1 Failed**: Update approach, increment to 2, retry.
   - **Attempt 2 Failed**: Request user context, increment to 3, retry.
   - **Attempt 3 Failed**: Circuit Breaker trips:
     - Write `status: abandoned` for phase in `.atl/sdd-state.yaml`
     - Add phase to `circuit_breaker.abandoned_phases`
     - Emit Result Contract with `status: abandoned`
     - Exit Code 2
4. **Orchestrator on Trip**: Intercept Exit Code 2, save logs/state, emit diagnostic, **immediately halt** flow. No advance.
5. **Recovery**:
   - User resolves blocker manually → `/sdd-continue`
   - User force-forwards → `/sdd-ff`
   - User archives as abandoned → `/sdd-archive --status=abandoned`
