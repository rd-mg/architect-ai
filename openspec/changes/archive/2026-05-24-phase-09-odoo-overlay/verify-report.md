## Verification Report

**Change**: phase-09-odoo-overlay
**Version**: N/A
**Mode**: Standard

---

### Completeness
| Metric | Value |
|--------|-------|
| Tasks total | 4 |
| Tasks complete | 4 |
| Tasks incomplete | 0 |

All tasks marked [x] — complete.

---

### Build & Tests Execution

**Build**: ✅ Passed
```
go build ./... → exit 0 (no output)
```

**Tests**: ✅ 12 passed / ❌ 0 failed / ⏭ 0 skipped

**Go Tests** (4/4 PASS):
```
=== RUN   TestIndexAll_BasicIndexing --- PASS
=== RUN   TestIndexAll_Idempotent --- PASS
=== RUN   TestIndexAll_PartialFailure --- PASS
=== RUN   TestIndexAll_TopicKeyFormat --- PASS
```

**Python Tests** (8/8 PASS):
```
test_build_dashboard PASSED
test_build_quote PASSED
test_validate_osheet_duplicate_formula_id PASSED
test_validate_osheet_error_literal PASSED
test_validate_osheet_missing_keys PASSED
test_validate_osheet_non_string_cell PASSED
test_validate_osheet_unprotected_division PASSED
test_xlsx_operations PASSED
```

**Coverage**: Not available (no coverage threshold configured)

---

### Spec Compliance Matrix

| # | Requirement | Scenario | Test | Result |
|---|-------------|----------|------|--------|
| T1 | Path Resolution | WARN when ODOO_COMMUNITY unset | `_shared/odoo-path-resolution.md` (static doc — no automated test) | ⚠️ PARTIAL |
| T2 | Language Linter | Spanish patterns → exit 1 | `lint-language.sh` (no automated test; script modifications differ from design) | ⚠️ PARTIAL |
| T3 | Module Builder Version Isolation | v19 blueprints ≠ legacy attrs | Blueprint v19 has `SQL()`, `<list>`, `<chatter/>` — verified via source read | ✅ COMPLIANT |
| T4 | XLSX Security Block | 50MB file → BLOCKED | Pre-flight bash in `odoo-minimax SKILL.md`; Python `xlsx_validate` does NOT check size | ⚠️ PARTIAL |
| T5 | osheet Validate | `"B4": 2026` (int) → NON-STRING | `test_validate_osheet_non_string_cell` (PASSED) + `validate_osheet()` impl | ✅ COMPLIANT |

**Compliance summary**: 2/5 scenarios fully compliant, 3/5 partial

---

### Correctness (Static — Structural Evidence)
| Requirement | Status | Notes |
|------------|--------|-------|
| Error 1 — Hardcoded Paths | ⚠️ Partial | Community → `${ODOO_COMMUNITY}` ✓. Enterprise → `${ODOO_ENTERPRISE:-...}` partial. OCA/OWL/spreadsheets/docs → still hardcoded `~/gitproj/odoo/...` in main SKILL.md. ripgrep-odoo SKILL.md, explore-odoo.md, copilot-instructions.md unaffected (out of scope). |
| Error 2 — Language Policy | ⚠️ Partial | Script exists, checks Spanish, exits 1. But patterns and output format differ from design. |
| Error 3 — External Repository Coupling | ✅ Implemented | External References Policy in main SKILL.md documents correct access chain. No external URLs found. |
| Error 4 — Legacy Context in Module Builder | ✅ Implemented | Version selection via case statement. 4 blueprint dirs (v14-v15, v16-v17, v18, v19). v19 uses `SQL()`, `<list>`, `<chatter/>`. No "IMPORTANT: for v19 remember..." blocks. |
| Error 5 — Research Priority Inversion | ✅ Implemented | Research priority in module builder SKILL.md: Engram → rg → Blueprint → Context7. Correct order. |
| Error 6 — Missing Risk Vocabulary | ✅ Implemented | odoo-minimax SKILL.md has `risk_level: high`, `security_warning:`, `max_file_size_mb: 10`, pre-flight bash block with size check + decompressed check. |
| Error 7 — Missing Recovery Strategies | ✅ Implemented | All 5 Odoo sub-skills have `## Recovery Strategies [MANDATORY in every Odoo SKILL.md]`. |
| Spreadsheet Macro-Skill Fusion (9.2) | ✅ Implemented | Unified routing table, universal rules, standard color palette, Recovery Strategies. Matches design. |
| odoo_sheet_tool.py (9.3) | ✅ Implemented | Unified CLI with build/validate/xlsx/compare. Matches design verbatim. |
| Engram Indexer (9.4) | ✅ Implemented | indexer.go + indexer_test.go. 4 Go tests PASS. Idempotent SaveIdempotent. Parallel goroutines. |

---

### Coherence (Design)
| Decision | Followed? | Notes |
|----------|-----------|-------|
| Replace hardcoded paths in main SKILL.md | ⚠️ Partial | Community path uses `${ODOO_COMMUNITY}` ✓. Enterprise uses `${ODOO_ENTERPRISE:-...}`. OCA/OWL/spreadsheets/docs still hardcoded. |
| New odoo-path-resolution.md shared doc | ✅ Yes | Created at `_shared/odoo-path-resolution.md`. Matches design verbatim. |
| lint-language.sh with phrase patterns | ⚠️ Deviated | Existing script uses word-level patterns (vista, modelo...) NOT the design's phrase patterns (Cuando el usuario, Objetivo...). Output uses `[FAIL] Found forbidden Spanish keyword` not design's `LANGUAGE VIOLATION`. |
| Unified spreadsheet SKILL.md router | ✅ Yes | Routing table, universal rules, color palette all match design. |
| Recovery Strategies heading format | ✅ Yes | `## Recovery Strategies [MANDATORY in every Odoo SKILL.md]` heading applied to module builder (apply change) and present in all 5 sub-skills. |
| Engram indexer with errgroup parallelism | ✅ Yes | Implementation matches design. |
| File Changes | ⚠️ Partial | Changed files match. Additional files with hardcoded paths remain (ripgrep-odoo, explore-odoo.md, copilot-instructions.md) — noted as out-of-scope. |

---

### Issues Found

| Level | Description | Status |
|-------|-------------|--------|
| **[WARNING]** | `lint-language.sh` implementation differs from design: uses word-level Spanish keywords (vista, modelo, campo...) instead of design's phrase patterns (Cuando el usuario, Objetivo...). Output format is `[FAIL] Found forbidden Spanish keyword` not `LANGUAGE VIOLATION` as spec expects. | Open |
| **[WARNING]** | Main `SKILL.md` still has hardcoded `~/gitproj/odoo/...` paths for OCA (line 104), OWL (105), spreadsheets (106), developer docs (107), functional docs (108). Only community path is fully resolved to variable. | Open |
| **[WARNING]** | Outside scope but notable: `ripgrep-odoo/SKILL.md` (25+ matches), `explore-odoo.md` (3 matches), `copilot-instructions.md` (7+ matches) still use `~/gitproj/odoo/...` hardcoded paths. | Open |
| **[WARNING]** | Test 4 (XLSX Security Block) PASS condition is inconsistent: spec says `odoo_sheet_tool.py xlsx --action validate exits 1 for oversized file` but the Python script's validate only checks ZIP structure — a 50MB valid XLSX would PASS validation. The size check exists only in the SKILL.md bash pre-flight, not in the Python tool. | Open |
| **[SUGGESTION]** | Spec Test 4 condition should reference bash pre-flight check instead of `odoo_sheet_tool.py xlsx --action validate` for file size enforcement. | Open |

---

### Adversarial Findings

**[PASS 2: ADVERSARIAL REVIEW]**

**Finding 1 — False Positive in lint-language.sh assessment**: The apply summary stated this file was "already correct (no changes needed)" but the existing implementation substantially differs from the design. The design specifies phrase-level Spanish patterns for the `SPANISH_PATTERNS` array (`'Cuando el usuario'`, `'Objetivo:'`, etc.) while the implementation uses word-level patterns (`'vista'`, `'modelo'`, `'campo'`, `'nombre'`, etc.). Both functionally catch Spanish content, but the word-level approach has higher false-positive risk (e.g., `'nombre'` could appear in a French comment, `'vista'` in a code context). The output format also differs. **Verdict**: Functional but deviates from design; WARNING warranted.

**Finding 2 — Path hardening gap**: The design's Error 1 says "Replace ALL hardcoded paths" but only the community path was resolved to `${ODOO_COMMUNITY}` in the main SKILL.md. Enterprise uses env var with hardcoded default, and 5 other paths remain fully hardcoded. Three additional files (ripgrep-odoo, explore-odoo, copilot-instructions) were untouched. **Verdict**: Core goal partially met; remaining hardcoded paths are a drift from design intent.

**Finding 3 — Test 4 spec ambiguity**: The PASS condition references the Python tool for size enforcement that doesn't implement it. The bash pre-flight (in SKILL.md) would correctly block a 50MB file, but the verification criteria point to the wrong component. **Verdict**: Spec defect — PASS condition doesn't match expected behavior.

**Finding 4 — No critical bypasses found**: The Go tests exercise concurrency, idempotency, and partial failure correctly. The Python tests cover validation, build, and xlsx operations. Build compiles cleanly. No test-fixing or contract violations detected. All existing tests pass.

---

### Verdict
**PASS WITH WARNINGS**

Implementation is functionally complete across all 4 tasks. All tests pass (12/12). Build compiles. Core behavioral requirements (version isolation, osheet validation, recovery strategies, engram indexing) are verified. However, 3 warnings remain regarding partial path hardening, lint-language.sh implementation drift, and a spec inconsistency in Test 4's PASS condition.

### Return Envelope (Internal)
```json
{
  "status": "success",
  "findings_triage": {
    "blocking": 0,
    "warning": 4,
    "suggestion": 1
  },
  "ready_for_archive": true
}
```
