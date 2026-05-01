package scope

import (
	"path/filepath"
	"testing"
)

func TestClassifyRefactorPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected PathClass
	}{
		{"root readme", "README.md", PathSourceRefactor},
		{"root go.mod", "go.mod", PathSourceRefactor},
		{"cmd dir", filepath.Join("cmd", "main.go"), PathSourceRefactor},
		{"internal dir", filepath.Join("internal", "scope", "scope.go"), PathSourceRefactor},
		{"docs dir", filepath.Join("docs", "design.md"), PathSourceRefactor},
		{"openspec root", filepath.Join("openspec", "config.yaml"), PathSourceRefactor},
		{"openspec archive", filepath.Join("openspec", "archive", "old.md"), PathUnknown},
		{"openspec changes archive", filepath.Join("openspec", "changes", "archive", "old.md"), PathUnknown},
		{"dot dir root", filepath.Join(".agent", "workflows", "sdd.md"), PathGeneratedDotdir},
		{"dot dir nested", filepath.Join("cmd", ".hidden", "file.go"), PathGeneratedDotdir},
		{"node modules", filepath.Join("node_modules", "pkg", "index.js"), PathBuildArtifact},
		{"vendor", filepath.Join("vendor", "pkg", "lib.go"), PathBuildArtifact},
		{"tmp", filepath.Join("tmp", "cache.bin"), PathBuildArtifact},
		{"unknown root file", "random.txt", PathUnknown},
		{"unknown dir", filepath.Join("random", "file.go"), PathUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClassifyRefactorPath(tt.path); got != tt.expected {
				t.Errorf("ClassifyRefactorPath(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestShouldRefactorPath(t *testing.T) {
	if ShouldRefactorPath(filepath.Join(".atl", "config.yaml")) {
		t.Error("ShouldRefactorPath(.atl/config.yaml) should be false")
	}
	if !ShouldRefactorPath(filepath.Join("internal", "scope", "scope.go")) {
		t.Error("ShouldRefactorPath(internal/scope/scope.go) should be true")
	}
}
