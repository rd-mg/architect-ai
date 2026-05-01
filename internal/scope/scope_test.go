package scope

import (
	"path/filepath"
	"testing"
)

func TestShouldRefactorSourcePath(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		activeChange string
		expected     bool
	}{
		{"root readme", "README.md", "", true},
		{"root go.mod", "go.mod", "", true},
		{"cmd dir", filepath.Join("cmd", "main.go"), "", true},
		{"internal dir", filepath.Join("internal", "scope", "scope.go"), "", true},
		{"docs dir", filepath.Join("docs", "design.md"), "", true},
		{"openspec root", "openspec/config.yaml", "", true},
		{"openspec specs", filepath.Join("openspec", "specs", "auth.md"), "", true},
		{"openspec active change", filepath.Join("openspec", "changes", "my-change", "delta.md"), "my-change", true},
		{"openspec inactive change", filepath.Join("openspec", "changes", "other-change", "delta.md"), "my-change", false},
		{"openspec archive", filepath.Join("openspec", "archive", "old.md"), "", false},
		{"openspec changes archive", filepath.Join("openspec", "changes", "archive", "old.md"), "", false},
		{"dot dir root", filepath.Join(".agent", "workflows", "sdd.md"), "", false},
		{"dot dir nested", filepath.Join("cmd", ".hidden", "file.go"), "", false},
		{"node modules", filepath.Join("node_modules", "pkg", "index.js"), "", false},
		{"vendor", filepath.Join("vendor", "pkg", "lib.go"), "", false},
		{"dot tmp", filepath.Join(".tmp", "cache.bin"), "", false},
		{"regular tmp (allowed)", filepath.Join("tmp", "cache.bin"), "", false}, // Denylisted in v4.1
		{"unknown root file", "random.txt", "", false},
		{"unknown dir", filepath.Join("random", "file.go"), "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldRefactorSourcePath(tt.path, tt.activeChange); got != tt.expected {
				t.Errorf("ShouldRefactorSourcePath(%q, %q) = %v, want %v", tt.path, tt.activeChange, got, tt.expected)
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
