package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitignoreCovers_ExactMatch(t *testing.T) {
	content := ".env.mcp\n"
	if !gitignoreCovers(content, ".env.mcp") {
		t.Error("exact match should be covered")
	}
}

func TestGitignoreCovers_CommentNotCovered(t *testing.T) {
	content := "# .env.mcp\n"
	if gitignoreCovers(content, ".env.mcp") {
		t.Error("commented pattern should NOT be covered")
	}
}

func TestGitignoreCovers_NegationNotCovered(t *testing.T) {
	content := "!.env.mcp\n"
	if gitignoreCovers(content, ".env.mcp") {
		t.Error("negated pattern should NOT be covered")
	}
}

func TestGitignoreCovers_GlobMatch(t *testing.T) {
	content := "*.mcp\n"
	if !gitignoreCovers(content, ".env.mcp") {
		t.Error("glob *.mcp should cover .env.mcp")
	}
}

func TestGitignoreCovers_NegationOnly(t *testing.T) {
	content := "!.env.mcp\n"
	if gitignoreCovers(content, ".env.mcp") {
		t.Error("negation-only content should NOT cover the pattern")
	}
}

func TestEnsureGitignored_AddsWhenMissing(t *testing.T) {
	dir := t.TempDir()
	gitignorePath := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("*.log\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	ensureGitignored(gitignorePath, ".env.mcp")

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(data), ".env.mcp") {
		t.Errorf(".env.mcp should have been added, got:\n%s", data)
	}
}

func TestEnsureGitignored_NoDoubleAdd(t *testing.T) {
	dir := t.TempDir()
	gitignorePath := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("*.log\n.env.mcp\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	ensureGitignored(gitignorePath, ".env.mcp")

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	count := strings.Count(string(data), ".env.mcp")
	if count != 1 {
		t.Errorf("expected .env.mcp to appear exactly once, got %d times", count)
	}
}
