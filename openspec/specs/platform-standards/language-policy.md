# Spec: English Language Sweep

## Requirements

1. **New File**: `docs/language-policy.md` defining "voice" vs "data".
2. **New File**: `scripts/lint-language.sh` to enforce the policy.
3. **CI Integration**: `Makefile` and GitHub Actions updated to run the lint.
4. **Content Cleanup**: Remove Spanish voice from all `internal/assets` files (orchestrators, personas, protocols).
5. **Trigger Preservation**: Keep Spanish trigger phrases (e.g., "iniciar sdd") but annotate them for the lint script.

## Acceptance Criteria

- [ ] `make lint-language` exists and exits 0 on a clean repo.
- [ ] `make lint-language` fails if Spanish voice (e.g., "estás en lo cierto") is added to a non-allowlisted file.
- [ ] No Spanish voice remains in `internal/assets/`.
- [ ] All 9 orchestrator files have annotated trigger phrases.
- [ ] CI pipeline includes the language lint step.
