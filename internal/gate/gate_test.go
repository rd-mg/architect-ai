package gate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rd-mg/architect-ai/internal/paths"
)

func TestNew(t *testing.T) {
	t.Parallel()
	g := New("testdata/source.md")
	if g.sourceFile != "testdata/source.md" {
		t.Errorf("expected sourceFile 'testdata/source.md', got %q", g.sourceFile)
	}
}

func TestCheck(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Create source file (needed for gate loadContent in Inject)
	sourceFile := filepath.Join(dir, "source.md")
	os.WriteFile(sourceFile, []byte("gate content"), 0644)

	v2File := filepath.Join(dir, "v2.md")
	os.WriteFile(v2File, []byte("some content\n"+GateStartMarker+"\nmore content"), 0644)

	v1File := filepath.Join(dir, "v1.md")
	os.WriteFile(v1File, []byte("prefix\n"+GateV1Marker+"\nsuffix"), 0644)

	noGateFile := filepath.Join(dir, "none.md")
	os.WriteFile(noGateFile, []byte("just regular content"), 0644)

	missingFile := filepath.Join(dir, "missing.md")

	g := New(sourceFile)

	tests := []struct {
		name   string
		target Target
		want   CheckResult
	}{
		{
			name:   "v2 present",
			target: Target{File: v2File},
			want:   CheckResult{File: v2File, Present: true, Version: "v2"},
		},
		{
			name:   "v1 present",
			target: Target{File: v1File},
			want:   CheckResult{File: v1File, Present: true, Version: "v1-outdated"},
		},
		{
			name:   "no gate",
			target: Target{File: noGateFile},
			want:   CheckResult{File: noGateFile, Present: false, Version: "none"},
		},
		{
			name:   "file not found",
			target: Target{File: missingFile},
			want:   CheckResult{File: missingFile, Present: false, Version: "file-not-found"},
		},
	}

	results := g.Check(testsTargets(tests))
	for i, tt := range tests {
		r := results[i]
		if r.File != tt.want.File {
			t.Errorf("%s: File = %q, want %q", tt.name, r.File, tt.want.File)
		}
		if r.Present != tt.want.Present {
			t.Errorf("%s: Present = %v, want %v", tt.name, r.Present, tt.want.Present)
		}
		if r.Version != tt.want.Version {
			t.Errorf("%s: Version = %q, want %q", tt.name, r.Version, tt.want.Version)
		}
	}
}

func TestInject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	sourceContent := GateStartMarker + "\ngate block content\n" + GateEndMarker
	sourceFile := filepath.Join(dir, "source.md")
	os.WriteFile(sourceFile, []byte(sourceContent), 0644)

	targetFile := filepath.Join(dir, "target.md")
	os.WriteFile(targetFile, []byte("## Delegation Rules\nsome rules"), 0644)

	noMarkerFile := filepath.Join(dir, "nomarker.md")
	os.WriteFile(noMarkerFile, []byte("no marker here"), 0644)

	g := New(sourceFile)

	t.Run("inject before marker", func(t *testing.T) {
		results := g.Inject([]Target{{
			File:         targetFile,
			InsertBefore: "## Delegation Rules",
		}})
		if results[0].Error != nil {
			t.Fatalf("unexpected error: %v", results[0].Error)
		}
		if results[0].AlreadyPresent {
			t.Fatal("expected injection, not already-present")
		}
		data, _ := os.ReadFile(targetFile)
		if !strings.Contains(string(data), GateStartMarker) {
			t.Fatal("expected gate content in target file")
		}
	})

	t.Run("already present", func(t *testing.T) {
		results := g.Inject([]Target{{
			File:         targetFile,
			InsertBefore: "## Delegation Rules",
		}})
		if !results[0].AlreadyPresent {
			t.Fatal("expected already-present")
		}
	})

	t.Run("marker not found", func(t *testing.T) {
		results := g.Inject([]Target{{
			File:         noMarkerFile,
			InsertBefore: "## Nonexistent",
		}})
		if results[0].Error == nil {
			t.Fatal("expected error for missing marker")
		}
	})

	t.Run("file not found", func(t *testing.T) {
		results := g.Inject([]Target{{
			File: filepath.Join(dir, "nonexistent.md"),
		}})
		if results[0].Error == nil {
			t.Fatal("expected error for missing file")
		}
	})
}

func TestInjectReplaceV1(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	sourceContent := GateStartMarker + "\ngate block content\n" + GateEndMarker
	sourceFile := filepath.Join(dir, "source.md")
	os.WriteFile(sourceFile, []byte(sourceContent), 0644)

	v1File := filepath.Join(dir, "v1target.md")
	os.WriteFile(v1File, []byte("prefix\n"+GateV1Marker+"\n## Heading\ncontent"), 0644)

	g := New(sourceFile)
	results := g.Inject([]Target{{
		File:         v1File,
		InsertBefore: "## Heading",
		ReplaceV1:    true,
	}})
	if results[0].Error != nil {
		t.Fatalf("unexpected error: %v", results[0].Error)
	}
	if !results[0].Replaced {
		t.Fatal("expected ReplaceV1 to be true")
	}
	data, _ := os.ReadFile(v1File)
	if strings.Contains(string(data), GateV1Marker) {
		t.Fatal("v1 marker should have been removed")
	}
	if !strings.Contains(string(data), GateStartMarker) {
		t.Fatal("expected new gate content")
	}
}

func TestPurge(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	sourceFile := filepath.Join(dir, "source.md")
	os.WriteFile(sourceFile, []byte("dummy"), 0644)

	v2File := filepath.Join(dir, "v2.md")
	os.WriteFile(v2File, []byte("before\n"+GateStartMarker+"\n"+GateEndMarker+"\nafter"), 0644)

	v1File := filepath.Join(dir, "v1.md")
	os.WriteFile(v1File, []byte("before\n"+GateV1Marker+"\n## Next\ncontent"), 0644)

	g := New(sourceFile)

	t.Run("purge v2", func(t *testing.T) {
		results := g.Purge([]Target{{File: v2File}})
		if results[0].Error != nil {
			t.Fatalf("unexpected error: %v", results[0].Error)
		}
		if !results[0].Modified {
			t.Fatal("expected modified=true")
		}
		data, _ := os.ReadFile(v2File)
		if strings.Contains(string(data), GateStartMarker) {
			t.Fatal("v2 marker should have been removed")
		}
	})

	t.Run("purge v1", func(t *testing.T) {
		results := g.Purge([]Target{{File: v1File}})
		if results[0].Error != nil {
			t.Fatalf("unexpected error: %v", results[0].Error)
		}
		if !results[0].Modified {
			t.Fatal("expected modified=true")
		}
		data, _ := os.ReadFile(v1File)
		if strings.Contains(string(data), GateV1Marker) {
			t.Fatal("v1 marker should have been removed")
		}
	})

	t.Run("no gate found", func(t *testing.T) {
		cleanFile := filepath.Join(dir, "clean.md")
		os.WriteFile(cleanFile, []byte("no gate"), 0644)
		results := g.Purge([]Target{{File: cleanFile}})
		if results[0].Error != nil {
			t.Fatalf("unexpected error: %v", results[0].Error)
		}
		if results[0].Modified {
			t.Fatal("expected modified=false for file with no gate")
		}
	})

	t.Run("file not found", func(t *testing.T) {
		results := g.Purge([]Target{{File: filepath.Join(dir, "missing.md")}})
		if results[0].Error == nil {
			t.Fatal("expected error for missing file")
		}
	})
}

func TestPurgeL2AutoScoring_SKILLNotFound(t *testing.T) {
	t.Parallel()
	// This uses a glob that won't match anything in the test environment
	dir := t.TempDir()
	sourceFile := filepath.Join(dir, "source.md")
	os.WriteFile(sourceFile, []byte("dummy"), 0644)

	g := New(sourceFile)
	ctx := paths.New(".", true)
	results := g.PurgeL2AutoScoring(ctx.L2SkillGlob())
	// Should return empty results (no matching files) without error
	if len(results) != 0 {
		t.Logf("PurgeL2AutoScoring returned %d results (no glob matches expected)", len(results))
	}
}

func TestEnsureIncludeInTemplates(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	sourceFile := filepath.Join(dir, "source.md")
	os.WriteFile(sourceFile, []byte("dummy"), 0644)

	tmplFile := filepath.Join(dir, "template.md")
	os.WriteFile(tmplFile, []byte("## Delegation Rules\nsome rules"), 0644)

	alreadyHasInclude := filepath.Join(dir, "already.md")
	os.WriteFile(alreadyHasInclude, []byte(`{{ include "_shared/adaptive-reasoning-gate-v2.md" }}`+"\n## Delegation Rules"), 0644)

	noMarkerFile := filepath.Join(dir, "nomarker.md")
	os.WriteFile(noMarkerFile, []byte("no delegation rules section"), 0644)

	g := New(sourceFile)

	t.Run("insert before Delegation Rules", func(t *testing.T) {
		g.EnsureIncludeInTemplates([]string{tmplFile})
		data, _ := os.ReadFile(tmplFile)
		if !strings.Contains(string(data), "{{ include") {
			t.Fatal("expected include directive in template")
		}
	})

	t.Run("already has include", func(t *testing.T) {
		g.EnsureIncludeInTemplates([]string{alreadyHasInclude})
		data, _ := os.ReadFile(alreadyHasInclude)
		count := strings.Count(string(data), "{{ include")
		if count != 1 {
			t.Fatalf("expected exactly 1 include directive, got %d", count)
		}
	})

	t.Run("appends at end when no marker", func(t *testing.T) {
		g.EnsureIncludeInTemplates([]string{noMarkerFile})
		data, _ := os.ReadFile(noMarkerFile)
		if !strings.Contains(string(data), "{{ include") {
			t.Fatal("expected include directive appended at end")
		}
	})

	t.Run("nonexistent file does not error", func(t *testing.T) {
		g.EnsureIncludeInTemplates([]string{filepath.Join(dir, "notexist.md")})
		// Should not panic
	})
}

func TestAtomicWrite(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	if err := atomicWrite(path, "hello world"); err != nil {
		t.Fatalf("atomicWrite failed: %v", err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "hello world" {
		t.Fatalf("expected 'hello world', got %q", string(data))
	}
}

// helper: extracts targets from test cases
func testsTargets(tt []struct {
	name   string
	target Target
	want   CheckResult
}) []Target {
	targets := make([]Target, len(tt))
	for i, tc := range tt {
		targets[i] = tc.target
	}
	return targets
}
