# Spec Delta: Release Metadata Contract

## Requirement: Canonical Metadata Source
- All user-facing repository, homepage, and documentation URLs must originate from a single canonical source of truth.
- Canonical Repository: `github.com/rd-mg/architect-ai`
- Canonical Documentation: `github.com/rd-mg/architect-ai/docs`
- Canonical Owner: `rd-mg` (or verified repository owner)

## Requirement: Consistency Alignment
- `.goreleaser.yaml` must use the canonical homepage and repository URLs.
- `internal/app/help.go` must use the canonical documentation URL.
- README.md and all sub-feature documentation must point to the canonical root.

## Verification
- `TestHelpContainsCanonicalDocsURL` must pass.
- `TestReleaseMetadataUsesCanonicalHomepage` must pass.
