## Delivery Strategy Selection [Ask ONCE — on first SDD command of the session]

When the user sends the FIRST SDD intent (combined with Execution Mode question):

Ask (LITE register):
```
Before we start, two quick questions:

1. Execution mode?
   [i] interactive (default) — I pause after each phase
   [a] automatic — run all phases, show final result

2. Delivery strategy for large changes?
   [r] ask-on-risk (default) — I ask when a change exceeds 400 lines
   [c] auto-chain — I automatically split into chained PRs ≤ 400 lines each
   [s] single-pr — I always do one PR (you confirm any exceptions)
   [e] exception-ok — I proceed even for large PRs, just document it
```

Cache answers:
- `session.execution_mode` = interactive | automatic
- `session.delivery_strategy` = ask-on-risk | auto-chain | single-pr | exception-ok

Pass BOTH to every sdd-orchestrator delegation.

Default if no answer:
- execution_mode = interactive
- delivery_strategy = ask-on-risk
