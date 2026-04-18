<!-- adaptive-reasoning-gate:START -->
## Adaptive Reasoning (MANDATORY)

Before executing your assigned phase protocol, you MUST classify the reasoning depth required for this task. 

**Response Format**: You MUST state your chosen mode as the very first line of your response (or within the first 5 non-blank lines if a brief preamble is needed). 

**Format**: `Mode: {n}. Why: {short reason}.`

| Mode | Scenario |
|------|----------|
| **1: Fast** | Mechanical, low-risk, or repetitive tasks. You already know exactly what to do. |
| **2: Balanced** | Standard implementation, multi-file changes, or architectural alignment. Requires careful thinking but no deep experimentation. |
| **3: Deep** | High-risk, ambiguous, or complex refactors. Requires internal chain-of-thought, alternative evaluation, and edge-case analysis. |
| **deferred** | Only for sdd-orchestrator when waiting for user input. |
| **sdd-first** | Only for sdd-init or sdd-onboard during bootstrap. |

FAILURE to include this mode declaration will result in an automated re-prompt.
<!-- adaptive-reasoning-gate:END -->
