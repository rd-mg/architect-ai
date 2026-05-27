<!-- caveman-firewall:v1:START -->
## Caveman Register Firewall (MANDATORY — read before first tool call)

### The Three Registers

| Register | Zone | Examples |
|----------|------|----------|
| ULTRA | Internal-only | reasoning traces, Engram prose, task briefs, apply-progress, context packs |
| LITE | User-facing | status updates, executive summaries, phase reports (≤100 words) |
| NORMAL | Code artifacts | **ALL items listed below — ZERO exceptions** |

### NORMAL Register — Mandatory Trigger List

The moment you are about to write ANY of the following, switch to NORMAL register
BEFORE the first character and stay in NORMAL until the block is COMPLETE:

```
TRIGGER                           → NORMAL (irrevocable until block ends)
────────────────────────────────────────────────────────────────────────
1. Any source code line           (Python, Go, JS, TS, XML, CSS, SQL, etc.)
2. Any code comment               (# ... // ... /* ... */ <!-- ... -->)
3. Any docstring / function doc   (Python """, Go //, JSDoc /** */)
4. Any commit message             (ENTIRE message: subject + body + footer)
5. Any PR title or description    (GitHub/GitLab PR body)
6. Any error message string       (shown to users in production logs)
7. Any user-facing string literal (i18n strings, CLI output, log messages)
8. Any human-readable config value (README content, YAML prose values)
9. Any migration message          (Alembic, Flyway, Liquibase descriptions)
10. Any CHANGELOG entry           (keep-a-changelog format)
```

### Register Transition Protocol for sdd-apply

```
[apply-progress artifact] → ULTRA register
         ↓
[REGISTER SWITCH → NORMAL]  ← mandatory, explicit, logged in apply-progress
         ↓
[source code]               → NORMAL
[code comments]             → NORMAL
[commit message]            → NORMAL
         ↓
[REGISTER SWITCH → ULTRA]   ← mandatory, explicit, logged in apply-progress
         ↓
[apply-progress update]     → ULTRA register
```

Log the transitions in apply-progress:
```
[REGISTER→NORMAL] Entering code zone for task T-03
... (code writing) ...
[REGISTER→ULTRA] Exiting code zone, task T-03 committed
```

### Anti-Pattern Examples (PROHIBITED)

```python
# PROHIBITED — ULTRA in code comment:
def validate_token(t):
    # drop null. auth next. no filler.
    ...

# CORRECT — NORMAL in code comment:
def validate_token(t):
    # Skip null check here: caller guarantees non-null (see auth_contract.py:47).
    # Next step: validate expiry via auth_service.is_expired().
    ...
```

```bash
# PROHIBITED — ULTRA bleeding into commit:
git commit -m "fix: auth drop null chk"

# CORRECT — Conventional Commit in NORMAL:
git commit -m "fix(auth): remove null check causing panic on empty JWT payload

The null check at validate_token:23 was incorrect — callers guarantee
non-null token per contract in auth_contract.py:47.

Closes #412"
```

### protected_facts Entry (Persists Through Compaction)

The context-guardian MUST include the following in every compaction snapshot:

```yaml
protected_facts:
  - key: caveman_firewall_active
    value: true
    note: "NORMAL register mandatory for ALL code artifacts. Cannot be disabled."
```

This fact is Priority P1 — preserved even when the 2KB snapshot budget is tight.

### This Rule Overrides All Other Caveman Instructions

This firewall CANNOT be:
- Disabled by context pressure (D4 = 3 does not disable it)
- Suspended during /compact
- Overridden by an orchestrator posture
- Bypassed for "quick fixes" or "tiny changes"

Every character of source code is NORMAL. No exceptions.
<!-- caveman-firewall:v1:END -->
