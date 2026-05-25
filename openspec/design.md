# Design: FASE 4 — Adaptive Reasoning Gate v3 + Sequential Thinking + Circuit Breaker Integration

> **Source of Truth:** `/home/rdmachadog/Documents/fix_achitect-ai/04-phase-adaptive-reasoning-v3.md`

## Architecture & Code Implementations

## 4.1 `internal/assets/skills/adaptive-reasoning/SKILL.md` (v3.0 final)

```markdown
---
name: adaptive-reasoning
description: >
  Single-entry classifier and cross-agent reasoning engine v3.0.
  Scores D1-D5 dimensions, routes to Mode 1/2/3, selects postures explicitly,
  triggers sequential thinking for complex tasks, integrates with circuit breaker.
  Part of foundation (Tier 1) — always injected in every agent.
tier: foundation
version: "3.0"
---

# Adaptive Reasoning Gate v3.0

<!-- architect-ai:caveman:identity-start -->
## Output Register [MANDATORY]
Language: English. Caveman: LITE for user output. ULTRA for Gate header and internal reasoning.
<!-- architect-ai:caveman:identity-end -->

## Operating Contract (non-negotiable)

1. **Self-Classification FIRST**: Score D1-D5 before EVERY response. No exceptions.
2. **Response Header MANDATORY**: First line of every response MUST match:
   `[MODE N | D1=X D2=X D3=X D4=X D5=X | POSTURE: +++P1 [+++P2]]`
   If this line is absent from your response, the orchestrator will retry the phase.
3. **Deterministic Routing**: Mode AND postures are decided by the table, not by LLM intuition.
4. **Hard Ceiling**: MAX 2 active postures simultaneously. Three postures = cognitive incoherence.
5. **D5 Ambiguity Rule**: If you cannot determine D5 with certainty AND the task touches
   authentication / credentials / secrets / user data → assume D5=2.
6. **Circuit Breaker Integration**: After classifying, check .atl/sdd-state.yaml for the
   current phase's attempt_count. If attempt_count >= 2, escalate to Mode 3 automatically.

## Dimensions (D1-D5)

| Dim | Label | 0 | 1 | 2 | 3 |
|---|---|---|---|---|---|
| **D1** | Complexity | Atomic, single file | Bounded module | Cross-module systemic | Architectural paradigm change |
| **D2** | Uncertainty | Specs clear and complete | Partial specs | Conflicting docs or unknown domain | Terra incognita |
| **D3** | Error Pressure | First run | Recent failure | Repeated failure | Production down / data loss risk |
| **D4** | Context Pressure | < 10 KB context used | 10-50 KB | 50-100 KB | > 100 KB (Guardian must fire) |
| **D5** | Security/Risk | No credentials, no PII | User data, normal | Auth / tokens / secrets / env vars | Crypto / PII / live production |

## D5 Ambiguity Resolution (MANDATORY)

Before assigning D5=0, verify:
```
IF task description contains ANY of these keywords:
  login, auth, token, password, secret, key, credential, session, cookie,
  oauth, jwt, user_id, role, permission, admin, sudo, encrypt, hash, salt
→ D5 >= 1 (at minimum)

IF context shows the agent will READ or WRITE files containing above keywords:
→ D5 >= 2

IF still ambiguous:
→ D5 = 2 (conservative default for security)
```

## Routing Matrix v3

| Condition | Mode | Label | Default Postures |
|---|---|---|---|
| D1+D2 ≤ 2 AND D3=0 AND D4 ≤ 1 AND D5=0 | **1** | Strategic | +++Pragmatic |
| D1+D2 ≥ 3 OR D3=1 | **2** | Tactical | +++Critical [++++Systemic if D1=3] |
| D3 ≥ 2 OR D4 ≥ 3 | **3** | Diagnostic | +++Forensic ++++Pragmatic |
| D5 = 2 | **Force ≥ 2** | + Security Review | +++Adversarial +++Critical |
| D5 = 3 | **Force 3** | + Parallel Review | +++Adversarial +++Forensic |
| attempt_count ≥ 2 (circuit breaker) | **Force 3** | Diagnostic fallback | +++Forensic +++Pragmatic |

## Explicit Posture Selection Table

After determining Mode, select postures from this table (MAX 2):

| Mode | D-Score pattern | Posture 1 | Posture 2 | Rationale |
|---|---|---|---|---|
| 1 | D1=0-1, D2=0-1 | +++Pragmatic | — | Direct execution, minimal overhead |
| 2 | D1=2-3 | +++Critical | +++Systemic | Cross-domain evaluation needed |
| 2 | D2=2-3 | +++Socratic | +++Critical | Clarify before acting |
| 2 | D3=1 | +++Forensic | +++Critical | Investigate recent failure |
| 2 | D5=2 | +++Adversarial | +++Critical | Security review mandatory |
| 2 | task involves cost/ROI/quota | +++Critical | +++Economic | Cost-aware evaluation |
| 2 | task needs measurement/benchmark | +++Empirical | +++Critical | Evidence-based decisions |
| 3 | D3 ≥ 2 | +++Forensic | +++Pragmatic | Stabilize, minimal blast radius |
| 3 | D1=3 (paradigm change) | +++Systemic | +++Adversarial | Deep impact analysis |
| 3 | D5=3 | +++Adversarial | +++Forensic | Security emergency + parallel review |
| 3 | attempt_count ≥ 2 | +++Forensic | +++Pragmatic | Break the pattern, minimal fix |

## SDD Phase-to-Mode Map (inject at delegation time)

| Phase | Default Mode | Posture 1 | Posture 2 | Sequential Thinking? |
|---|---|---|---|---|
| sdd-init | 1 | +++Pragmatic | — | No |
| sdd-onboard | 2 | +++Socratic | — | No |
| sdd-explore | 2 | +++Socratic | +++Empirical | If D1+D2 ≥ 5 |
| sdd-propose | 2 | +++Critical | +++Economic | If D1+D2 ≥ 5 |
| sdd-spec | 2 | +++Systemic | +++Critical | If D1+D2 ≥ 5 |
| sdd-design | 2 | +++Systemic | +++Adversarial | ALWAYS if D1 ≥ 2 |
| sdd-tasks | 1 | +++Pragmatic | +++Economic | No |
| sdd-apply | 1 | +++Pragmatic | — | No |
| sdd-verify | 2 | +++Critical | +++Adversarial | If D3 ≥ 1 |
| sdd-archive | 1 | +++Empirical | — | No |

## Non-SDD Agent-to-Mode Map

| Agent | Mode | Posture 1 | Posture 2 | Sequential Thinking? |
|---|---|---|---|---|
| general-orchestrator | 1 | +++Pragmatic | — | If D1+D2 ≥ 5 |
| researcher | 2 | +++Empirical | +++Socratic | If D1+D2 ≥ 5 |
| solver | 3 | +++Forensic | +++Systemic | If D1+D2 ≥ 4 |
| ideator | 2 | +++Divergent | +++Lateral | No (creativity ≠ deep reasoning) |
| generalist | 1 | +++Pragmatic | — | No |
| analyst | 2 | +++Empirical | +++Critical | If D1+D2 ≥ 5 |

## Sequential Thinking Activation Rule

```
IF (D1 + D2) >= 5
OR (D1 + D2) >= 3 AND D5 >= 2
OR task_type IN ["architectural_decision", "security_review", "multi-file_refactor"]:
  → MANDATORY: invoke sequential_thinking MCP BEFORE code generation
  → MIN_BRANCHES = 2 (evaluate at least 2 competing approaches)
  → MIN_THOUGHTS = 5
  → REQUIRE: at least 1 "revisit" thought challenging prior assumption

ELSE:
  → SKIP sequential thinking (overhead not justified for simple tasks)
```

## Sequential Thinking Fallback (MCP unavailable)

When sequential_thinking MCP server is down OR unavailable:

```
MANDATORY BRANCH ANALYSIS (inline — replaces MCP):

[SEQUENTIAL THINKING — inline fallback]

Branch A: {approach_name}
  Implementation: {how}
  Tradeoffs: {pros / cons}
  Risk: {what could go wrong}
  Token cost estimate: {rough size}

Branch B: {alternative_approach}
  Implementation: {how}
  Tradeoffs: {pros / cons}
  Risk: {what could go wrong}
  Token cost estimate: {rough size}

[If D1=3 or D5 >= 2: add Branch C — adversarial / do-nothing option]
Branch C: {adversarial_or_do_nothing}
  Why this matters: {what if we don't do either A or B}

Decision: Branch {X}
Rationale: {specific evidence from codebase or specs}
Rejected: {brief why not for others}
[END SEQUENTIAL THINKING]
```

## D5=3 Parallel Review Protocol

When D5=3 (crypto/PII/live production):
```
MANDATORY: Launch 2 independent review sub-agents BEFORE merging any code.

Sub-agent 1 (executor): implements the change
Sub-agent 2 (adversarial reviewer): receives ONLY the diff + spec, reviews independently

Consensus protocol:
- Both approve → proceed to merge
- Reviewer rejects → executor MUST address all rejections before re-review
- Second rejection → STATUS: BLOCKED, escalate to human

This cannot be bypassed by execution_mode="automatic".
```

## Circuit Breaker Integration

At the START of every response (after reading sdd-state.yaml):
```bash
PHASE="${current_phase}"
ATTEMPTS=$(grep -A5 "  ${PHASE}:" .atl/sdd-state.yaml | grep -o "attempt_counts.*[0-9]" | grep -o "[0-9]" | tail -1)
ATTEMPTS=${ATTEMPTS:-0}

if [ "${ATTEMPTS}" -ge 2 ]; then
  # Force Mode 3 regardless of D-scores
  echo "CIRCUIT BREAKER ACTIVE: ${PHASE} has ${ATTEMPTS} prior attempts."
  echo "Forcing Mode 3 + +++Forensic to break the pattern."
  # Override mode selection: Mode 3, +++Forensic +++Pragmatic
fi
```

## Ralph Loop Prevention (exit code 2)

If Mode 3 is triggered by circuit breaker AND this is attempt 3:
```
DO NOT choose another approach. Instead:
1. Emit: "RALPH LOOP PREVENTION: 3 attempts in Mode 3. Aborting."
2. Return Result Contract with status: "abandoned"
3. Exit code 2
4. Record in sdd-state.yaml: abandoned_phases += [current_phase]

The orchestrator will pause the SDD cycle and present options to the user.
A Ralph Loop is worse than an abandoned phase — it burns all API quota.
```
```

---

## 4.2 Cognitive Mode Postures Catalog

### `internal/assets/skills/cognitive-mode/SKILL.md` (v2.0 — all 11 postures)

```markdown
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
```

---

## 4.3 Go Installer — Gate Header Validator

```go
// internal/reasoning/gate/validator.go
// Validates that agent responses contain the mandatory Gate header.
// Used during golden testing and E2E test runs.
package gate

import (
    "fmt"
    "regexp"
    "strings"
)

// GateHeader represents a parsed Adaptive Reasoning Gate header
type GateHeader struct {
    Mode    int
    D1, D2, D3, D4, D5 int
    Posture1 string
    Posture2 string // empty if only one posture
    Raw      string
}

// headerPattern matches: [MODE N | D1=X D2=X D3=X D4=X D5=X | POSTURE: +++P1 [+++P2]]
var headerPattern = regexp.MustCompile(
    `^\[MODE ([123]) \| D1=([0-3]) D2=([0-3]) D3=([0-3]) D4=([0-3]) D5=([0-3]) \| POSTURE: (\+\+\+\w+)(?: (\+\+\+\w+))?\]`)

// ParseHeader parses the gate header from the first line of a response
func ParseHeader(firstLine string) (*GateHeader, error) {
    m := headerPattern.FindStringSubmatch(strings.TrimSpace(firstLine))
    if m == nil {
        return nil, fmt.Errorf("invalid or missing gate header: %q\nExpected format: [MODE N | D1=X D2=X D3=X D4=X D5=X | POSTURE: +++P]", firstLine)
    }

    parseInt := func(s string) int { var n int; fmt.Sscan(s, &n); return n }

    return &GateHeader{
        Mode:     parseInt(m[1]),
        D1:       parseInt(m[2]),
        D2:       parseInt(m[3]),
        D3:       parseInt(m[4]),
        D4:       parseInt(m[5]),
        D5:       parseInt(m[6]),
        Posture1: m[7],
        Posture2: m[8],
        Raw:      firstLine,
    }, nil
}

// ValidateDecision checks that the Mode and Postures are consistent with D-scores
func ValidateDecision(h *GateHeader) []string {
    var issues []string

    // D5 >= 2 must have +++Adversarial
    if h.D5 >= 2 {
        hasAdversarial := h.Posture1 == "+++Adversarial" || h.Posture2 == "+++Adversarial"
        if !hasAdversarial {
            issues = append(issues, fmt.Sprintf("D5=%d requires +++Adversarial posture", h.D5))
        }
        if h.Mode < 2 {
            issues = append(issues, fmt.Sprintf("D5=%d requires Mode >= 2, got Mode %d", h.D5, h.Mode))
        }
    }

    // D3 >= 2 must be Mode 3
    if h.D3 >= 2 && h.Mode < 3 {
        issues = append(issues, fmt.Sprintf("D3=%d requires Mode 3, got Mode %d", h.D3, h.Mode))
    }

    // Two postures must not be the same
    if h.Posture2 != "" && h.Posture1 == h.Posture2 {
        issues = append(issues, fmt.Sprintf("duplicate postures: %s and %s", h.Posture1, h.Posture2))
    }

    // Divergent/Lateral/Diamond only valid in ideator context (warning, not error)
    creativePostures := map[string]bool{"+++Divergent": true, "+++Lateral": true, "+++Diamond": true}
    if creativePostures[h.Posture1] && h.Mode == 3 {
        issues = append(issues, "creative postures (Divergent/Lateral/Diamond) are unusual for Mode 3")
    }

    return issues
}

// ExtractFirstLine returns the first non-empty line of a response
func ExtractFirstLine(response string) string {
    for _, line := range strings.Split(response, "\n") {
        if line = strings.TrimSpace(line); line != "" {
            return line
        }
    }
    return ""
}
```

```go
// internal/reasoning/gate/validator_test.go
package gate

import "testing"

func TestParseHeader_Valid(t *testing.T) {
    line := "[MODE 2 | D1=2 D2=1 D3=0 D4=1 D5=0 | POSTURE: +++Critical +++Systemic]"
    h, err := ParseHeader(line)
    if err != nil {
        t.Fatalf("ParseHeader: %v", err)
    }
    if h.Mode != 2 { t.Errorf("Mode: want 2, got %d", h.Mode) }
    if h.D1 != 2   { t.Errorf("D1: want 2, got %d", h.D1) }
    if h.D5 != 0   { t.Errorf("D5: want 0, got %d", h.D5) }
    if h.Posture1 != "+++Critical" { t.Errorf("Posture1: %s", h.Posture1) }
    if h.Posture2 != "+++Systemic" { t.Errorf("Posture2: %s", h.Posture2) }
}

func TestParseHeader_SinglePosture(t *testing.T) {
    line := "[MODE 1 | D1=0 D2=0 D3=0 D4=0 D5=0 | POSTURE: +++Pragmatic]"
    h, err := ParseHeader(line)
    if err != nil { t.Fatalf("ParseHeader: %v", err) }
    if h.Posture1 != "+++Pragmatic" { t.Errorf("Posture1: %s", h.Posture1) }
    if h.Posture2 != "" { t.Errorf("Posture2 should be empty: %s", h.Posture2) }
}

func TestParseHeader_Invalid(t *testing.T) {
    cases := []string{
        "This is a normal response without gate header",
        "[MODE 4 | D1=0 D2=0 D3=0 D4=0 D5=0 | POSTURE: +++Pragmatic]", // MODE 4 invalid
        "",
    }
    for _, c := range cases {
        _, err := ParseHeader(c)
        if err == nil {
            t.Errorf("expected error for invalid header: %q", c)
        }
    }
}

func TestValidateDecision_D5RequiresAdversarial(t *testing.T) {
    h := &GateHeader{Mode: 2, D5: 2, Posture1: "+++Critical", Posture2: "+++Systemic"}
    issues := ValidateDecision(h)
    if len(issues) == 0 {
        t.Error("D5=2 without +++Adversarial should produce issues")
    }
}

func TestValidateDecision_D5AdversarialCorrect(t *testing.T) {
    h := &GateHeader{Mode: 2, D5: 2, Posture1: "+++Adversarial", Posture2: "+++Critical"}
    issues := ValidateDecision(h)
    if len(issues) > 0 {
        t.Errorf("valid D5=2 setup should have no issues, got: %v", issues)
    }
}

func TestValidateDecision_D3ForcesMode3(t *testing.T) {
    h := &GateHeader{Mode: 2, D3: 2, Posture1: "+++Forensic", Posture2: "+++Pragmatic"}
    issues := ValidateDecision(h)
    if len(issues) == 0 {
        t.Error("D3=2 with Mode 2 should produce issue (requires Mode 3)")
    }
}

func TestValidateDecision_DuplicatePostures(t *testing.T) {
    h := &GateHeader{Mode: 2, Posture1: "+++Critical", Posture2: "+++Critical"}
    issues := ValidateDecision(h)
    if len(issues) == 0 {
        t.Error("duplicate postures should produce issue")
    }
}

func TestExtractFirstLine(t *testing.T) {
    response := "\n\n[MODE 1 | D1=0 D2=0 D3=0 D4=0 D5=0 | POSTURE: +++Pragmatic]\n\nSome content"
    got := ExtractFirstLine(response)
    if got != "[MODE 1 | D1=0 D2=0 D3=0 D4=0 D5=0 | POSTURE: +++Pragmatic]" {
        t.Errorf("unexpected first line: %q", got)
    }
}
```

---

## Criterios de Verificación

### Test 1: Gate Header Mandatory
```
Input: ANY agent response
Expected: First non-empty line matches `^\[MODE [123] \| D1=[0-3] ...`
PASS if: Zero responses in golden tests without the gate header
```

### Test 2: D5 Ambiguity Default
```
Input: Task description contains "login endpoint"
Agent cannot determine D5 with certainty
Expected: D5 >= 1 (not D5=0)
PASS if: D5=0 never appears when task mentions auth-related keywords
```

### Test 3: Sequential Thinking Triggers
```
Input: sdd-design task with D1=3, D2=2 (D1+D2=5)
Expected: Sequential thinking MCP invoked OR inline fallback template used
Expected: At least 2 branches evaluated before code generation
PASS if: Branch analysis present in response before first code block
```

### Test 4: Sequential Thinking Fallback
```
Setup: sequential_thinking MCP server unavailable (connection refused)
Input: Task with D1+D2 >= 5
Expected: Inline fallback activates automatically (no blocking wait for MCP)
Expected: Branch A / Branch B / Decision format appears in response
PASS if: No infinite wait; fallback within 2 seconds of MCP failure detection
```

### Test 5: Circuit Breaker Forces Mode 3
```
Setup: sdd-apply in sdd-state.yaml has attempt_count: 2
Input: sdd-apply starts
Expected: Gate reads attempt_count = 2
Expected: Mode 3 forced regardless of D1-D4 scores
Expected: +++Forensic posture selected
PASS if: Gate header shows MODE 3 even for D1=0,D2=0 task
```

### Test 6: D5=3 Parallel Review
```
Input: sdd-design task touching auth token generation (D5=3)
Expected: Orchestrator launches 2 sub-agents: executor + adversarial reviewer
Expected: Both sub-agents produce independent output
Expected: Merge only after consensus
PASS if: Two separate delegation calls appear in orchestrator trace
```

---

## Resultados Esperados

| Métrica | Antes | Después |
|---|---|---|
| D5 ambiguity handling | ❌ Defaults to 0 (unsafe) | ✅ Defaults to 2 when keyword detected |
| +++Economic auto-trigger | ❌ Never automatic | ✅ sdd-propose + sdd-tasks with cost mentions |
| +++Empirical auto-trigger | ⚠️ Only researcher | ✅ researcher + sdd-archive + sdd-explore D2≥1 |
| Sequential thinking when MCP down | ❌ Blocks | ✅ Inline fallback activates automatically |
| Circuit breaker integration | ❌ None | ✅ attempt_count check forces Mode 3 |
| Ralph Loop prevention | ❌ None | ✅ exit code 2 after 3 Mode 3 failures |
| Gate header validation | ❌ No validator | ✅ ParseHeader + ValidateDecision in tests |
