# SDD Verify — Odoo Context

Apply domain-specific checks IN ADDITION to standard sdd-verify.

## Odoo Version
{See shared preamble — version cached from sdd-init.}

## Deterministic Checklist

Apply ALL checks. Each automatable via `ripgrep` or file existence.

### Manifest Checks
- [ ] `__manifest__.py` exists in module root
- [ ] `version` field is X.Y.Z.W format
- [ ] `version` incremented since last commit on this module
  ```bash
  git log --oneline --diff-filter=M -- __manifest__.py | head -5
  ```
- [ ] `depends` list contains all imported modules (cross-check with `_inherit` and Python imports)
- [ ] `external_dependencies` declared if new Python libraries used
- [ ] `data` list in correct load order (security → data → views → menus)

### Security Checks
- [ ] `security/ir.model.access.csv` exists for EVERY new model
  ```bash
  # Find new models
  rg "_name\s*=\s*['\"]" models/ --files-with-matches
  # Check each appears in ir.model.access.csv
  cat security/ir.model.access.csv
  ```
- [ ] Record rules exist for multi-company fields (`company_id`)
- [ ] No `sudo()` without documented justification
  ```bash
  rg "\.sudo\(\)" --files-with-matches
  # Each match should have a comment like "# sudo: accessing res.users as portal user"
  ```
- [ ] No raw SQL without parameterization
  ```bash
  rg "cr\.execute\(" -A 2 --files-with-matches
  # Each should use %s placeholders, not f-strings or string formatting
  ```
- [ ] No user-controlled input in XML IDs or domains

### Code Quality Checks
- [ ] `hasclass()` used in XPath (not `contains(@class, ...)`)
  ```bash
  rg "contains\(@class" views/
  # Expected: no matches
  ```
- [ ] No `attrs=` in v17+ code
  ```bash
  rg "attrs\s*=\s*\"" views/
  # Expected (v17+): no matches
  ```
- [ ] No `<tree>` in v18+ views
  ```bash
  rg "<tree" views/
  # Expected (v18+): no matches
  ```
- [ ] OWL version matches Odoo version:
  - v15: OWL 1.x (`require('@odoo/owl')`)
  - v16-18: OWL 2.x (`import { Component } from "@odoo/owl"`)
  - v19: OWL 3.x (new patterns)
- [ ] No `@api.model_create_multi` missing on `create` overrides
- [ ] All computed fields have complete `@api.depends(...)`
- [ ] No unbounded `search([])` (must have limit, offset, or narrow domain)
- [ ] No N+1 queries (browsing inside loops over large recordsets)

### Documentation Checks
- [ ] `README.md` exists in module root
- [ ] README describes: purpose, features, configuration, dependencies
- [ ] Changelog entry exists for this version (in README or CHANGELOG.md)

### Migration Checks (Conditional)
If schema changes made in this batch:
- [ ] Pre-migrate script in `migrations/{version}/pre-migrate.py`
- [ ] Post-migrate script if data transformation needed
- [ ] New required fields have default values or computed defaults
- [ ] Field renames/removals have proper migration logic

### Test Checks
- [ ] Test file exists for each new capability in `tests/`
- [ ] Tests pass: `python odoo-bin -i {module} --test-enable --stop-after-init`
- [ ] Tests cover happy path AND at least one error case
- [ ] No `print()` or `import pdb` left in code
- [ ] Judgement Day Gate executed and PASSED (Required for COMPLETE status)

## Adversarial Review Focus Areas

Apply adaptive-reasoning Mode 2 (adversarial-review) with Odoo-specific lenses:

### Pass A: Local correctness lens
- Does code do what spec says?
- Are all declared fields actually used?
- Do views reference fields that exist?
- Do method signatures match their invocations?

### Pass B: System impact lens
- What OTHER modules inherit from this? Did we break them?
- What happens during module upgrade from previous version?
- What happens if user installs this module on DB with existing data?
- What happens if user uninstalls this module? Orphan records?
- Does change affect multi-company behavior?
- Does change interact with Enterprise features?
- Does this interact with studio module (user customizations)?

## Human-Required Checklist (Agent Reminds, Does Not Execute)

Include these reminders in return envelope:

- [ ] Module installs on clean database without errors
- [ ] Module upgrades on existing database without errors
- [ ] Feature tested with correct user roles (not only admin)
- [ ] Multi-company behavior verified (if applicable)
- [ ] Multi-currency behavior verified (if applicable)
- [ ] No JavaScript console errors in browser
- [ ] Main flows tested end-to-end on staging
- [ ] Edge cases identified by developer tested

## Output Format

In verify-report, include:

```markdown
## Odoo-Specific Verification

### Deterministic Checks
 Manifest version bumped (18.0.1.0.0 → 18.0.1.1.0)
 ir.model.access.csv exists for acme.approval.request
 README.md missing in module root — BLOCKER
 sudo() used in models/approval.py:45 without justification comment

### Adversarial Findings
- CRITICAL: Uninstall leaves orphan records in acme_approval_log
- WARNING (real): Multi-company not respected in acme.approval.request.search
- SUGGESTION: Consider adding tracking=True to state field

### Human Verification Required
- [ ] Test on clean DB
- [ ] Test upgrade from 18.0.1.0.0
- [ ] Test as non-admin user

### Verdict
NEEDS CHANGES (README missing, CRITICAL uninstall issue, sudo() needs justification)
```

> [!IMPORTANT]
> **JUDGEMENT DAY GATE**: Module cannot be marked COMPLETE until Judgement Day Gate (Pass/Fail) is executed and reported below.

## Boundaries

- Do NOT mark APPROVED with any BLOCKER or CRITICAL finding unresolved
- Do NOT skip deterministic checks because "code looks fine"
- Do NOT run tests in production-adjacent environments without user confirmation

## JUDGEMENT DAY GATE (execute before marking verify as COMPLETE)

### When to activate
- sdd-verify sub-agent about to report "PASS" to Orchestrator
- Mode 1 active (expansive reasoning available — context not saturated)
- First complete verification of module (not partial re-verifications)

### When NOT to activate
- Mode 3 active (saturated context — Judgement Day too expensive)
- Re-verification of already approved sub-component
- Trivial task (Complexity = 0-1 in Classifier)

### Execution protocol

**INPUT:** Approved Brief (mem_context of sdd/{module}/brief/v{N})
**POSTURE:** +++Critical (mandatory — Judgement Day without Critical makes no sense)
**TOKEN BUDGET:** max 600 tokens (if exceeded → force truncate)

**PROMPT BLOCK (inject in Layer 8 when activated):**

> [!IMPORTANT]
> [JUDGEMENT DAY GATE — final verification of the Brief]
> Before closing sdd-verify as COMPLETE, audit active Brief against these 3 critical criteria for Odoo:
>
> 1. MODULE INTEGRITY: Can module be uninstalled without leaving orphan records in core tables? Do all relational fields have ondelete defined?
> 2. CORE COLLISION: Does design directly modify core models (res.partner, account.move, res.users) without using inheritance? Or create duplicate functionality OCA already provides?
> 3. SCALABILITY: Are there N+1 operations in computed fields? Do search filters use indexes? Does design work with 1M records?
>
> For each criterion: PASS or FAIL + 1-sentence description.
> Final verdict: PASS (all 3 OK) or FAIL (any serious failure).
>
> If FAIL: specify only minimum required change. Do not rewrite Brief.
> If PASS: return "JUDGEMENT DAY: PASS" and proceed with closure.

### Result and post-Judgement Day action

- **PASS** → after_model hook saves:
  `mem_save("sdd/{module}/brief/v{N}", {brief + "judgement_day": "PASS"})`
  Orchestrator advances to next phase.
- **FAIL** → Orchestrator re-opens sdd-design with corrections as input.
  brief/v{N+1} created (version incremented).
  Failure counter does NOT increment (it's design correction, not error).

### Timeout and protection
If sub-agent takes more than 2 cycles in Judgement Day → assume PASS with warning:
"JUDGEMENT DAY: TIMEOUT — Brief approved with warning. Manually review criteria 1-3 before deploy."
