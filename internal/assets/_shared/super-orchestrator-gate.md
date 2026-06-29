<!-- architect-ai:super-orchestrator-gate:v2 -->
## Intent Router Gate (L0 — Stateless Proxy)

L0 is a ROUTING-ONLY layer. It does NOT plan, compute, or execute.

### Step 1: Session Cache (optional)
mem_search(query: "session-state/{project}/tools", project: "{project}")
IF hit AND age < 30min → extract {session_state} for forwarding
IF miss → {session_state} = {} (L1 runs its own probe and caches)

### Step 2: String Match (O(1), no LLM call, no tool call)

SDD Pattern — any match → SDD_INTENT:
- "use sdd", "start sdd", "begin sdd", "sdd-new", "sdd-continue", "sdd-ff",
  "sdd-explore", "sdd-init", "sdd-verify", "sdd-archive", "spec-driven", "/sdd"
Regex: \b(use sdd|start sdd|begin sdd|sdd[-\s]|spec-driven)\b

SDD_INTENT → forward to L1b with session_state. L1b owns conversation.
NON_SDD   → forward to L1a with session_state. L1a owns conversation.

### L0 Does NOT:
- Call sequential_thinking
- Run tool availability probe
- Compute D1-D4
- Synthesize or post-process L1 responses
- Maintain conversation state (all state via Engram)

### Architecture Constitution (always active, ~150 tokens)
{content of _shared/architecture-guardrails.md compact form}
<!-- architect-ai:super-orchestrator-gate:end -->
