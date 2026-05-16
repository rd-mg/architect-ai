## Semantic Audit Protocol [sdd-verify MANDATORY — execute BEFORE test runner]

### Step 1: Retrieve Contracts from Engram
```
spec   = mem_get_observation("sdd/{change_name}/spec")
design = mem_get_observation("sdd/{change_name}/design")

IF both null → ADD RISK "no contracts in Engram", SKIP audit, proceed to test runner.
IF spec null only → partial audit using design only.
```

### Step 2: Parse Assertions

From spec:
- MUST / SHALL / MUST NOT clauses → existence/absence assertions
- BDD Given/When/Then blocks → behavioral assertions  
- FMEA listed failure modes → negative (MUST NOT occur) assertions

From design:
- Function/method signatures → signature assertions
- API endpoint path + method → endpoint assertions
- Schema field names + types → schema assertions

### Step 3: Verify via rg

```bash
# Existence (function/method)
rg -l "def {name}|func {name}" . \
  || echo "MISSING:func:{name}"

# Signature match
rg "def {name}\({params}" --type py \
  || echo "MISMATCH:sig:{name}"

# Go signature
rg "func {Name}\(" --type go \
  || echo "MISMATCH:sig:{Name}"

# API endpoint
rg "route.*\"{path}\"|\"/{path}\"" . \
  || echo "MISSING:endpoint:{path}"

# Odoo field existence
rg "'{field}' = fields\." --type py \
  || echo "MISSING:odoo_field:{field}"

# Negative assertion (MUST NOT exist)
rg -l "{forbidden}" . \
  && echo "VIOLATION:{forbidden}" \
  || echo "CLEAN:{forbidden}"
```

### Step 4: Score and Gate
```
violations       = count(MISSING | MISMATCH | VIOLATION results)
total_assertions = count of all assertions checked
score            = violations / max(total_assertions, 1)

score = 0.0        → PASS  — proceed to test runner
0 < score ≤ 0.10   → WARN  — proceed, add violations to RISKS
score > 0.10       → REJECT — halt, report delta, preserve apply branch
```

### Step 5: Persist Result
```
mem_save(
  topic_key: "sdd/{change_name}/semantic-audit",
  content: {
    assertions_checked, violations, score, status,
    apply_branch, timestamp
  }
)
```
