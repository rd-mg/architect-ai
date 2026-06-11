.PHONY: lint-language
lint-language:
	@./scripts/lint-language.sh

.PHONY: test
test:
	go test ./...

# ─── architect-ai Hardening v0.3 ─────────────────────────────────────────────

.PHONY: hardening hardening-verify hardening-go test-race

## hardening-go: Build and test all Go changes for hardening v0.3
hardening-go:
	@echo "Building all Go packages..."
	@go build ./... 2>&1 || (echo "BUILD FAILED" && exit 1)
	@echo "  ✓ Build passed"
	@echo "Running tests..."
	@go test ./internal/components/openspec/... -count=1 -timeout 120s 2>&1 || (echo "openspec TESTS FAILED" && exit 1)
	@go test ./internal/metering/... -count=1 -timeout 60s 2>&1 || (echo "metering TESTS FAILED" && exit 1)
	@go test ./internal/reasoning/gate/... -count=1 -timeout 60s 2>&1 || (echo "gate TESTS FAILED" && exit 1)
	@go test ./internal/verify/... -count=1 -timeout 60s 2>&1 || (echo "verify TESTS FAILED" && exit 1)
	@go test ./internal/components/engram/engramkeys/... -count=1 -timeout 60s 2>&1 || (echo "engramkeys TESTS FAILED" && exit 1)
	@go test ./internal/components/filemerge/... -count=1 -timeout 60s 2>&1 || (echo "filemerge TESTS FAILED" && exit 1)
	@go test ./internal/components/mcp/... -count=1 -timeout 60s 2>&1 || (echo "mcp TESTS FAILED" && exit 1)
	@go test ./internal/tui/... -count=1 -timeout 60s 2>&1 || (echo "tui TESTS FAILED" && exit 1)
	@go test ./internal/cli/... -count=1 -timeout 60s 2>&1 || (echo "cli TESTS FAILED" && exit 1)
	@echo "  ✓ All tests passed"

## test-race: Run all tests with the race detector
test-race:
	@go test -race \
	  ./internal/components/openspec/... \
	  ./internal/metering/... \
	  ./internal/reasoning/gate/... \
	  ./internal/tui/... \
	  -count=1 -timeout 180s

## hardening: Build, test, and verify (full pipeline)
hardening: hardening-go
	@echo ""
	@echo "════════════════════════════════════════"
	@echo "architect-ai hardening v0.3 complete ✓"
	@echo "════════════════════════════════════════"
