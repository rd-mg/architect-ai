# Adaptive-Reasoning v1.0 — Architecture Guide

**Status**: Stable | **Introduced in**: V3 | **Related skill**: `internal/assets/skills/adaptive-reasoning/SKILL.md`

---

## What Changed

Prior to V3, architect-ai shipped three separate skills that handled reasoning decisions:

- `adaptive-reasoning` — a classifier that routed to other skills
- `judgment-day` — a two-pass adversarial review skill
- `autoreason-lite` — an A/B/AB comparison skill for pre-implementation refinement

Using them together required the orchestrator to delegate to external sub-agents. Each delegation consumed ~1500 tokens for the fresh sub-agent context setup, lost the calling context (fresh sub-agent = no memory of what triggered it), and added latency from serialized sub-agent spawning.

**V3 absorbs all three into a single skill.** The new `adaptive-reasoning` classifies AND executes inline in the same context. The three old skills become three **modes** of the new skill.

## The Three Modes

| Mode | Formerly | Purpose |
|------|----------|---------|
| 1. `direct-exec` | (implicit — just proceed) | Atomic, low-risk, testable work — no reasoning overlay needed |
| 2. `adversarial-review` | `judgment-day` | Defect discovery via two independent passes with different lenses |
| 3. `bounded-synthesis` | `autoreason-lite` | A/B/AB comparison for pre-implementation refinement |

A fourth route, `native-sdd-first`, triggers when scope is multi-step or system-level and the right answer is "decompose via SDD phases first, then classify narrower".

## Classification: Four Observable Dimensions

Before routing, the skill scores the task on four dimensions:

| Dimension | Values (weakest → strongest) |
|-----------|------------------------------|
| `scope` | atomic · bounded · multi-step · system-level |
| `ambiguity` | clear · partial · conflicting · unknown |
| `risk` | low · medium · high · critical |
| `verification` | syntax-only · testable · review-heavy · multi-gate |

All four dimensions MUST be scored. The result is the input to the routing matrix.

## Routing Matrix

Apply in order. Stop at the first match.

| Signal | Mode | Why |
|--------|------|-----|
| `verification=testable` OR `multi-gate` AND goal is code acceptance | **direct-exec → defer to validators** | Tests/builds/linters decide; no reasoning overlay substitutes |
| `verification=review-heavy` OR `multi-gate` AND goal is DEFECT DISCOVERY | **Mode 2: adversarial-review** | Two passes expose what one misses |
| `ambiguity` is `partial`/`conflicting` AND incumbent draft `A` exists AND competing draft `B` exists AND goal is pre-implementation refinement | **Mode 3: bounded-synthesis** | A/B/AB comparison produces defensible final artifact |
| `scope=multi-step` OR `system-level` AND ambiguity is non-trivial | **native-sdd-first** | Decomposition before narrower modes |
| Everything else | **Mode 1: direct-exec** | No extra reasoning layer needed |

## Mode 2: adversarial-review (Inline Two-Pass)

When the goal is to find what's wrong.

### Procedure (executes inline, no delegation)

1. **Confirm target and scope** — if ambiguous, ask once; otherwise state assumption
2. **Pass A** — one serious analysis with a primary lens (local correctness for code, feasibility for specs, best-supported for research)
3. **Pass B** — materially different lens (system impact for code, failure-mode for specs, adversarial falsification for research) — kept independent of Pass A
4. **Agreement trap check** — if passes converge quickly, ask what shared assumption could make both wrong
5. **Synthesis** — weigh evidence, reasoning, requirement alignment, and remaining uncertainty
6. **Classify findings** — Confirmed, Suspect, Contradiction, Info
7. **Apply severity** — CRITICAL, WARNING (real), WARNING (theoretical), SUGGESTION
8. **Verdict** — APPROVED, CONDITIONALLY APPROVED, NEEDS CHANGES, UNRESOLVED

The verdict is ANALYTICAL. It is never presented as merge permission.

## Mode 3: bounded-synthesis (A/B/AB Comparison)

When pre-implementation refinement is needed and both an incumbent draft and a competing draft exist.

### Procedure (executes inline, no delegation)

1. **Applicability check** — all five conditions must hold (proposal/spec/design class, A exists, B exists, goal is refinement, "no change" is acceptable)
2. **Restate target** — decision target, constraints, success condition
3. **Normalize candidates** — A (incumbent), B (competing), same comparison frame
4. **Produce AB synthesis** — one synthesis combining the strongest traits of A and B
5. **Evaluate rubric** — correctness, safety, contract compatibility, blast radius, operability, simplicity, testability
6. **Conservative selection** — keep A on ties, prefer least churn, adopt B or AB only when gain is substantive
7. **Return result** — chosen option + minimal delta list + unresolved risks

## Output Record Contract

Every routing decision MUST emit this machine-readable record:

```text
mode: direct-exec | adversarial-review | bounded-synthesis | native-sdd-first
scope: atomic | bounded | multi-step | system-level
ambiguity: clear | partial | conflicting | unknown
risk: low | medium | high | critical
verification: syntax-only | testable | review-heavy | multi-gate
confidence: high | medium | low
escalation_flag: none | risk-override | cost-constrained | multi-gate-override
reason: <1-2 lines>
```

This record is a stable contract surface. Field names and mode values must not be renamed.

## Why Inline Execution Matters

Before V3, a typical verify flow looked like:

```
Orchestrator → sdd-verify → (delegate to judgment-day) → fresh sub-agent
                              → cold start, read registry, read skill, do two passes
                              → return verdict
                           → sdd-verify returns
Orchestrator → user
```

Each arrow in that chain costs tokens. The delegation to judgment-day consumed ~1500 extra tokens for context setup alone.

After V3:

```
Orchestrator → sdd-verify (uses adaptive-reasoning Mode 2 inline)
                → same context, do two passes
                → return verdict
Orchestrator → user
```

No delegation overhead. The same two-pass logic executes in the sub-agent's original context.

## Boundaries (What This Skill Does NOT Do)

- **Does not create SDD phases** — `sdd-propose`, `sdd-spec`, etc. still own artifact creation
- **Does not replace `sdd-verify`** — `sdd-verify` USES adversarial-review as one of its techniques
- **Does not decide code acceptance** — deterministic validators (tests, builds, linters) are authoritative for correctness
- **Does not own persistence** — phase skills own persistence of the final artifact
- **Does not delegate** — modes execute inline in the same response

## Escalation Rules

Priority order (stop at first match):

1. `verification=multi-gate` → always defer to deterministic validators (no reasoning overlay will make a test pass that's failing)
2. `risk=critical` → prefer Mode 2 or Mode 3 before Mode 1 (don't skip review for critical changes)
3. `cost-constrained` context → prefer the simplest mode that preserves correctness

## Compatibility

- Legacy prompts that reference `judgment-day` or `autoreason-lite` should be updated, but if not, the intent is clear — treat as Mode 2 or Mode 3 invocation
- The old skills are archived under `internal/assets/skills/_archived/` with a README explaining the migration
- Rolling back is possible via the procedure in `_archived/README.md`

## Anti-Patterns

- Using Mode 2 (adversarial-review) when the goal is synthesis (use Mode 3 instead)
- Using Mode 3 when only one credible draft exists (use Mode 1 instead)
- Using Mode 3 for implementation or acceptance work
- Letting both passes in Mode 2 share a hidden premise without checking
- Generating many alternatives in Mode 3 just because ambiguity exists
- Claiming Mode 2 verdict `APPROVED` = merge permission

## Migration From Pre-V3

If you have existing orchestrator prompts referencing the old skills:

| Old Reference | New Equivalent |
|---------------|----------------|
| `Launch judgment-day sub-agent` | Apply adaptive-reasoning Mode 2 inline |
| `Delegate to autoreason-lite` | Apply adaptive-reasoning Mode 3 inline |
| `Routes: judgment-day, autoreason-lite, native-owner, ...` | Modes: adversarial-review, bounded-synthesis, direct-exec, native-sdd-first |
| `Judge A / Judge B procedure` | Pass A / Pass B in Mode 2 |
| `A/B/AB comparison` | Same semantics, now in Mode 3 |

The `_archived/README.md` has the full rollback procedure if the new integration fails.

## Related

- `internal/assets/skills/adaptive-reasoning/SKILL.md` — authoritative reference
- `internal/assets/skills/_archived/README.md` — archived skills + migration
- `docs/cognitive-modes.md` — how postures interact with reasoning modes
- `internal/assets/claude/sdd-phase-protocols/sdd-verify.md` — how adversarial-review integrates with verify phase
