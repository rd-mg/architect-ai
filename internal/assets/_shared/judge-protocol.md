<!-- architect-ai:judge-protocol:v1 -->
## Judge Protocol (Adversarial Verification)

MANDATORY in sdd-verify. TWO SEPARATE judge agents. Never combine into one.

### Judge Primary — +++Adversarial
Role: Correctness. Happy-path failures. Implementation vs spec.
Model: sonnet (or agent default)
Focus areas:
- Does implementation match every spec capability?
- Are test assertions actually meaningful (not trivially true)?
- Hostile inputs: null, empty, max, concurrent, unicode edge cases
- No TODO/FIXME/XXX in changed code

### Judge Secondary — +++Forensic
Role: Robustness. Failure modes. Production survivability.
Model: sonnet (or agent default)
Focus areas:
- FMEA failure modes handled in implementation?
- Error paths explicitly tested?
- Race conditions under concurrent load?
- Resource leak paths (file handles, goroutines, connections)?
- Upgrade/downgrade hazards?

### Verdict Scale (both judges)
APPROVED                — meets all criteria
CONDITIONALLY_APPROVED  — minor issues, can ship with follow-up task
NEEDS_CHANGES           — blocking issues, must fix before archive
UNRESOLVED              — cannot determine without more information

### Synthesis Rules
| Primary      | Secondary    | Final verdict |
|-------------|-------------|---------------|
| APPROVED    | APPROVED    | APPROVED      |
| APPROVED    | COND_APPR   | COND_APPR     |
| COND_APPR   | APPROVED    | COND_APPR     |
| COND_APPR   | COND_APPR   | COND_APPR     |
| Any NEEDS_CHANGES | Any | NEEDS_CHANGES |
| Any UNRESOLVED | Any   | UNRESOLVED (escalate) |

Combined report: mem_save("sdd/{change-name}/verify-report", combined)
