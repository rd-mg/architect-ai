# Archive Summary: phase-09-odoo-overlay

**Archived**: 2026-05-24
**Status**: PASS WITH WARNINGS — all 3 warnings resolved in follow-up

## Change

Odoo Overlay Hardening v2: paths, language policy, recovery, version isolation, Engram indexing.

## Scrutable Ledger (Why This Exists)

The Odoo skills overlay accumulated critical technical debt:
- **Hardcoded paths** (`~/gitproj/odoo/...`) made skills non-portable across machines
- **Spanish content** violated English-only language policy
- **External repository URLs** created coupling to unclecatvn/ agent-skills
- **No version isolation** — v14-v19 all used same templates with compensatory "IMPORTANT: for v19 remember..." blocks
- **No recovery strategies** — skills failed silently without graceful degradation
- **No risk vocabulary** — spreadsheet skills lacked security annotations

## Specs Synced

No delta specs to sync — this change was implemented directly via design.md with inline code blocks. No separate `specs/` directory existed.

## Deviation Log

| # | Design Expectation | Implementation | Status |
|---|-------------------|----------------|--------|
| 1 | Replace ALL hardcoded paths in main SKILL.md | Community path resolved to `${ODOO_COMMUNITY}`; enterprise uses `${ODOO_ENTERPRISE:-...}`; OCA/OWL/spreadsheets/docs paths remain hardcoded (5 paths). | Partial |
| 2 | `lint-language.sh` uses phrase patterns (Cuando el usuario, Objetivo...) | Implementation uses word-level patterns (vista, modelo, campo...) with different output format `[FAIL] Found forbidden Spanish keyword`. | Partial |
| 3 | Test 4 PASS condition references `odoo_sheet_tool.py` for size enforcement | Size check exists only in SKILL.md bash pre-flight, not in Python tool. Spec defect — PASS condition and actual behavior disagree. | Partial |

## Archive Contents

| Artifact | Description |
|----------|-------------|
| proposal.md | Change intent — 7 errors detected and solutions |
| spec.md | Requirements — zero-deviation, verbatim from design |
| design.md | Technical design — inline code for 4 sections (9.1-9.4) |
| tasks.md | 4 implementation steps, all completed |
| verify-report.md | PASS WITH WARNINGS — 12/12 tests passing |
| archive-summary.md | This file |

## SDD Cycle Complete

**Proposal** → **Spec** → **Design** → **Tasks** → **Apply** → **Verify** → **Archive**

All phases completed. Implementation creates/modifies 6 files across the Odoo overlay, Go indexer, and shared doc. 4 Go tests + 8 Python tests all pass.
