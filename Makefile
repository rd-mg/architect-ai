.PHONY: lint-language
lint-language:
	@./scripts/lint-language.sh

.PHONY: test
test:
	go test ./...
