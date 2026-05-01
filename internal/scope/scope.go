package scope

import (
	"path/filepath"
	"strings"
)

type PathClass string

const (
	PathSourceRefactor  PathClass = "source_refactor"
	PathGeneratedDotdir PathClass = "generated_dotdir"
	PathBuildArtifact   PathClass = "build_artifact"
	PathUnknown         PathClass = "unknown"
)

func ClassifyRefactorPath(path string) PathClass {
	clean := filepath.Clean(path)
	
	// Fast path for exact root files
	if clean == "README.md" || clean == "PRD.md" || clean == "CONTRIBUTING.md" || 
		clean == "go.mod" || clean == "go.sum" || clean == "package.json" || 
		clean == "skills-lock.json" {
		return PathSourceRefactor
	}

	parts := strings.Split(clean, string(filepath.Separator))
	
	for _, part := range parts {
		if part == "." || part == "" {
			continue
		}
		if strings.HasPrefix(part, ".") && part != ".agent" {
			return PathGeneratedDotdir
		}
	}

	// Reject dist, build, coverage, vendor, tmp, .tmp, node_modules
	for _, part := range parts {
		switch part {
		case "node_modules", "dist", "build", "coverage", "vendor", ".tmp":
			return PathBuildArtifact
		}
	}

	// Check openspec archives
	if strings.HasPrefix(clean, "openspec"+string(filepath.Separator)+"archive") || 
	   strings.HasPrefix(clean, "openspec"+string(filepath.Separator)+"changes"+string(filepath.Separator)+"archive") {
		return PathUnknown // or PathBuildArtifact
	}

	switch {
	case strings.HasPrefix(clean, "cmd"+string(filepath.Separator)):
		return PathSourceRefactor
	case strings.HasPrefix(clean, "internal"+string(filepath.Separator)):
		return PathSourceRefactor
	case strings.HasPrefix(clean, "docs"+string(filepath.Separator)):
		return PathSourceRefactor
	case strings.HasPrefix(clean, "openspec"+string(filepath.Separator)):
		return PathSourceRefactor
	case strings.HasPrefix(clean, "scripts"+string(filepath.Separator)):
		return PathSourceRefactor
	case strings.HasPrefix(clean, "testdata"+string(filepath.Separator)):
		return PathSourceRefactor
	default:
		return PathUnknown
	}
}

func ShouldRefactorPath(path string) bool {
	return ClassifyRefactorPath(path) == PathSourceRefactor
}

func ShouldSkipRefactorPath(path string) bool {
	class := ClassifyRefactorPath(path)
	return class == PathGeneratedDotdir || class == PathBuildArtifact
}
