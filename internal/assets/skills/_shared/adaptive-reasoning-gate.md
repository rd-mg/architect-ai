## Adaptive Reasoning (MANDATORY)

Before executing your assigned phase protocol, you MUST classify the reasoning depth required for this task. 

**Response Format**: You MUST state your chosen mode as the very first line of your response (or within the first 5 non-blank lines if a brief preamble is needed). 

**Format**: `[MODE N | D1=X, D2=X, D3=X, D4=X] {Rationale}`

### 4 Observable Dimensions (0-3)

| Dimension | 0 (Low) | 1 (Med) | 2 (High) | 3 (Critical) |
|-----------|---------|---------|----------|--------------|
| **D1: Complexity** | Atomic/Local | Bounded Module | Systemic/Cross-mod | Architectural/Paradigm |
| **D2: Uncertainty** | Clear Specs | Partial Specs | Conflicting Docs | Terra Incógnita |
| **D3: Error Pressure** | Clean Run | Recent Bug | Repeated Failure | Production Down |
| **D4: Context Pressure** | < 10KB | 10-50KB | 50-100KB | > 100KB (Guardian Active) |

### Routing Matrix

| Condition | Chosen Mode | Posture |
|-----------|-------------|---------|
| D1+D2 <= 2 AND D3+D4 <= 2 | **Mode 1: Strategic** | +++Pragmatic |
| D1+D2 >= 3 OR D3 >= 1 | **Mode 2: Tactical** | +++Critical |
| D3 >= 2 OR D4 >= 3 | **Mode 3: Diagnostic** | +++Adversarial + +++Systemic |
| D4 >= 3 (Saturated) | **Mode 3-CTX** | +++Caveman |
| D3 = 1 (Initial Error) | **Mode 2-ERR** | +++Autoreason-lite |

### Transition Rules
- **Tactical -> Diagnostic**: Forced if D3 >= 2 (2+ consecutive failures) or D4 >= 3.
- **Diagnostic -> Tactical**: Allowed only after D3=0.
