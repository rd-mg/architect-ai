# Proposal: Phase 11 — Source of Truth Enforcement: Zero-Deviation + Traceability + Gap Detector

> **Source of Truth:** `/home/rdmachadog/Documents/fix_achitect-ai/11-phase-sot-enforcement.md`
> **Change:** phase-11-sot-enforcement
> **Phase:** sdd-propose
> **Generated:** 2026-05-22T19:17:00Z
> **Author:** sdd-propose (architect-ai)

## Intent

Eliminate hallucination, creative deviations, and "Pending Decision" stubs during SDD generation and application. Ensure 100% traceability to the original Source of Truth and guarantee zero gaps when translating specs into code.

## Problem Context

1. **Agent Deviation**: `sdd-apply` re-designs architecture instead of implementing the spec. The LLM "helps" by inventing better code than specified.
2. **Traceability Loss**: Spec → Design → Tasks → Code loses connection with the origin document. By verify phase, nobody knows if code fulfills the original contract.
3. **Implicit Assumptions**: The agent leaves `Pending Decision` or assumes behavior when spec is ambiguous, instead of halting to ask.

## Proposed Approach

1. **Zero-Deviation Coder Protocol**: Mandatory identity block for `sdd-apply`. First attempt = verbatim transcription. Modification only authorized on test failure (attempt 2+).
2. **Traceability Auto-Injection**: Every SDD artifact MUST include `> **Source of Truth:** {path}` header tracing to its predecessor.
3. **Assumption Linter** (`assumption-linter.sh`): Pre-flight validation scanning for Pending Decision, FixMe, Task, implicit assumptions, missing traceability headers.
4. **Hard-Stop Protocol**: Agent halts with BLOCKED JSON when spec is ambiguous. `--force-assume` flag for minor ambiguities only.
5. **Gap Detector Integration**: `sdd-verify` Step 0a runs assumption linter, Step 0b checks traceability, Step 0d validates all prior phase statuses.

## MANDATORY IMPLEMENTATION DIRECTIVE
**For `sdd-apply` (Coder):**
All code definitions, declarations, directives, and logic provided in the Source of Truth MUST be implemented EXACTLY as written on the first attempt. The coder MUST NOT analyze, re-design, or deviate from the source code during the initial implementation. Modification of the initial implementation is ONLY permitted if the tests fail.
