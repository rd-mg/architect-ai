# Design: English Language Sweep

## Overview
Implement a regex-based lint script to detect Spanish "voice" leaks while allowlisting "trigger" data. This ensures the framework speaks English exclusively as per directive #3.

## Components

### 1. Language Policy (`docs/language-policy.md`)
- Defines the distinction between **Voice** (English only) and **Data** (Triggers, multi-lingual).
- Documents the `<!-- trigger-phrase-allowlist -->` annotation.

### 2. Lint Script (`scripts/lint-language.sh`)
- Uses `rg` for high-performance scanning.
- **Safety Guards (from Judge Review)**:
  - `set -euo pipefail` for robust execution.
  - Automatic `cd` to git toplevel.
  - CRLF sanitation for the allowlist file.
  - Warning on missing allowlist paths.
- **Word List**: Curated regex for Spanish voice (e.g., `gracias`, `por favor`, `ahora`).

### 3. Allowlist (`scripts/language-allowlist.txt`)
- Explicit list of files permitted to contain Spanish trigger phrases.

### 4. CI Integration
- `Makefile` target `lint-language`.
- GitHub Actions step in `ci.yml`.

## Implementation Strategy
1. **Bootstrap**: Create policy and script.
2. **Sweep**: Refactor the 9 orchestrator files and `engram-protocol.md`.
3. **Verify**: Run lint and unit tests.
