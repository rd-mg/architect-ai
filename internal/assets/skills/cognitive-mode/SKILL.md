---
name: cognitive-mode
description: >
  11 cognitive postures for structured reasoning. Applied by Adaptive Reasoning Gate v3.
  Hard ceiling: 2 postures active simultaneously. Part of foundation (Tier 1).
tier: foundation
version: "2.0"
---

# Cognitive Mode Postures v2.0

## Hard Rule: MAX 2 Active Postures
Three or more postures simultaneously = attention head fragmentation = incoherent output.
The Gate selects postures deterministically. Agents do NOT choose their own postures.

## Posture Descriptions

### +++Pragmatic
**When**: Mode 1. sdd-apply. Archive. Simple tasks.
**Behavior**: Implement the minimum viable solution. Avoid abstractions not in spec.
No "while I'm here" improvements. No anticipatory engineering.
**Anti-pattern**: Do not suggest improvements. Do not refactor adjacent code.

### +++Critical
**When**: Mode 2 primary. Evidence evaluation required.
**Behavior**: Every claim requires evidence from codebase, tests, or specs.
Evaluate risks before proposing. State uncertainty explicitly.
**Anti-pattern**: Do not cite memory as evidence. Always cite file:line.

### +++Systemic
**When**: Mode 2 secondary. D1 ≥ 2 (cross-module changes).
**Behavior**: For every proposed change, ask: what second/third-order effects does this have?
Check callers. Check downstream. Check shared state.
**Anti-pattern**: Do not implement in isolation without checking callers.

### +++Adversarial
**When**: D5 ≥ 2. Mode 3 with security context.
**Behavior**: Actively try to break the design before proposing it.
List 3 ways the proposed approach could fail. Address each.
**Anti-pattern**: Do not just point out problems — propose mitigations.

### +++Forensic
**When**: Mode 3. D3 ≥ 1. Circuit breaker active.
**Behavior**: Trace evidence chain from symptom to root cause.
Every claim needs file:line provenance. Establish what IS working before diagnosing what is NOT.
**Anti-pattern**: Do not hypothesize without evidence. No "probably".

### +++Socratic
**When**: D2 ≥ 2. Requirements unclear.
**Behavior**: Ask 3 clarifying questions BEFORE starting any work.
Do not start coding with insufficient specs.
**Anti-pattern**: Do not assume intent. Ambiguity = ask, not guess.

### +++Economic
**When**: Task involves API quota, cost/ROI decisions, latency/cost tradeoffs.
**Auto-trigger**: sdd-propose or sdd-tasks when budget constraints mentioned.
**Behavior**: For each option, evaluate: token cost, latency impact, maintenance overhead, reversibility.
Reject technically superior options if economic cost is disproportionate.
**Anti-pattern**: Do not optimize for performance alone if cost is prohibitive.

### +++Empirical
**When**: Task requires benchmarks, measurements, or evidence-based decisions.
**Auto-trigger**: researcher always. sdd-archive always. sdd-explore when D2 ≥ 1.
**Behavior**: Base all conclusions STRICTLY on gathered evidence. No speculative claims.
Distinguish "measured" from "estimated" from "assumed".
**Anti-pattern**: Do not claim performance improvements without benchmark data.

### +++Divergent
**When**: ideator Phase 1. Creative generation needed.
**Behavior**: Generate WITHOUT filtering. Push beyond first 3 obvious answers.
Include ideas that seem impractical. 6-8 distinct ideas minimum.
**Anti-pattern**: Do not self-censor during generation. Filtering happens in evaluation phase.

### +++Lateral
**When**: ideator Phase 1 alongside +++Divergent. solver on deadlock (3 failed hypotheses).
**Behavior**: Apply lateral thinking: Reversal, Random Entry, Assumption Challenge, Zooming Out.
Challenge assumptions about the problem formulation itself.
**Anti-pattern**: Do not apply lateral thinking to code implementation — only to problem framing.

### +++Diamond
**When**: ideator Phase 3 (synthesis). After evaluation.
**Behavior**: Converge from 6-8 generated ideas to Top 3 with concrete next steps.
Each synthesized option has: concept + pros/cons + immediate next step.
**Anti-pattern**: Do not deliver Diamond synthesis without prior Divergent generation phase.