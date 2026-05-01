package scope

import (
	"path/filepath"
	"strings"
)

func hasDotSegment(path string) bool {
	slash := filepath.ToSlash(filepath.Clean(path))
	for _, part := range strings.Split(slash, "/") {
		if part == "" || part == "." {
			continue
		}
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}

func ShouldRefactorSourcePath(path string, activeChange string) bool {
	slash := filepath.ToSlash(filepath.Clean(path))
	if hasDotSegment(slash) {
		return false
	}
	if strings.HasPrefix(slash, "openspec/archive/") || strings.HasPrefix(slash, "openspec/changes/archive/") {
		return false
	}
	if activeChange != "" && strings.HasPrefix(slash, "openspec/changes/"+activeChange+"/") {
		return true
	}
	if slash == "." || slash == "" {
		return true
	}
	switch {
	case strings.HasPrefix(slash, "cmd/"):
		return true
	case strings.HasPrefix(slash, "internal/"):
		return true
	case strings.HasPrefix(slash, "docs/"):
		return true
	case strings.HasPrefix(slash, "scripts/"):
		return true
	case strings.HasPrefix(slash, "testdata/"):
		return true
	case strings.HasPrefix(slash, "openspec/specs/"):
		return true
	case slash == "openspec/config.yaml":
		return true
	case slash == "README.md" || slash == "PRD.md" || slash == "CONTRIBUTING.md" || slash == "go.mod" || slash == "go.sum" || slash == "package.json" || slash == "skills-lock.json" || slash == "AGENTS.md":
		return true
	default:
		return false
	}
}

func ShouldRefactorPath(path string) bool {
	return ShouldRefactorSourcePath(path, "")
}

func ShouldSkipRefactorPath(path string) bool {
	return !ShouldRefactorSourcePath(path, "")
}
