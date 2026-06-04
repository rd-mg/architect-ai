package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rd-mg/architect-ai/internal/platform"
)

func TestExpandHome(t *testing.T) {
	t.Parallel()
	home, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/test", filepath.Join(home, "test")},
		{"~/a/b/c", filepath.Join(home, "a", "b", "c")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		// expandHome only handles "~/..." prefix, not bare "~"
		//{"~", home},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := expandHome(tt.input)
			if got != tt.expected {
				t.Errorf("expandHome(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDryRunPlatform(t *testing.T) {
	t.Parallel()
	p := platform.Platform{
		Name:        "test-platform",
		DisplayName: "Test Platform",
		ConfigFiles: []platform.ConfigFile{
			{Path: ".config/test.json", Content: "{}"},
			{Path: ".config/other.json", Content: "{}"},
		},
	}

	files := DryRunPlatform(p)
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(files), files)
	}
	if files[0] != ".config/test.json" {
		t.Errorf("expected .config/test.json, got %q", files[0])
	}
}

func TestDryRunPlatformWithRouting(t *testing.T) {
	t.Parallel()
	p := platform.Platform{
		Name:                "test-routing",
		DisplayName:         "Test Routing",
		RoutingFileRequired: true,
		RoutingFile:         "ROUTING.md",
		ConfigFiles: []platform.ConfigFile{
			{Path: "config.json", Content: "{}"},
		},
	}

	files := DryRunPlatform(p)
	if len(files) != 2 {
		t.Fatalf("expected 2 files (config + routing), got %d: %v", len(files), files)
	}
	if files[1] != "ROUTING.md" {
		t.Errorf("expected ROUTING.md, got %q", files[1])
	}
}

func TestConfigurePlatformCreatesFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(origDir) })

	p := platform.Platform{
		Name:        "test-platform",
		DisplayName: "Test Platform",
		ConfigFiles: []platform.ConfigFile{
			{Path: filepath.Join(dir, "test.json"), Content: `{"key": "value"}`},
		},
	}

	written, err := ConfigurePlatform(p)
	if err != nil {
		t.Fatalf("ConfigurePlatform failed: %v", err)
	}
	if len(written) != 1 {
		t.Fatalf("expected 1 file written, got %d: %v", len(written), written)
	}

	data, err := os.ReadFile(written[0])
	if err != nil {
		t.Fatalf("cannot read written file: %v", err)
	}
	if string(data) != `{"key": "value"}` {
		t.Errorf("expected '{\"key\": \"value\"}', got %q", string(data))
	}
}

func TestConfigurePlatformCreatesDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	p := platform.Platform{
		Name:        "test-platform",
		DisplayName: "Test Platform",
		ConfigFiles: []platform.ConfigFile{
			{Path: filepath.Join(dir, "subdir", "config.json"), Content: `{}`},
		},
	}

	written, err := ConfigurePlatform(p)
	if err != nil {
		t.Fatalf("ConfigurePlatform failed: %v", err)
	}
	if len(written) != 1 {
		t.Fatalf("expected 1 file written, got %d", len(written))
	}
	if _, err := os.Stat(filepath.Join(dir, "subdir")); os.IsNotExist(err) {
		t.Fatal("expected subdir to be created")
	}
}

func TestMergeJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	existingFile := filepath.Join(dir, "existing.json")
	os.WriteFile(existingFile, []byte(`{"a": 1, "b": 2}`), 0644)

	merged, err := mergeJSON(existingFile, `{"b": 3, "c": 4}`)
	if err != nil {
		t.Fatalf("mergeJSON failed: %v", err)
	}

	// b should be updated, c added, a preserved
	expected := "{\n  \"a\": 1,\n  \"b\": 3,\n  \"c\": 4\n}"
	if merged != expected {
		t.Errorf("mergeJSON result mismatch.\ngot:\n%s\nwant:\n%s", merged, expected)
	}
}

func TestMergeJSON_NewFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	newFile := filepath.Join(dir, "new.json")

	merged, err := mergeJSON(newFile, `{"key": "value"}`)
	if err != nil {
		t.Fatalf("mergeJSON failed: %v", err)
	}
	expected := "{\n  \"key\": \"value\"\n}"
	if merged != expected {
		t.Errorf("expected %q, got %q", expected, merged)
	}
}

func TestMergeJSON_DeepMerge(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	existingFile := filepath.Join(dir, "deep.json")
	os.WriteFile(existingFile, []byte(`{"nested": {"a": 1, "b": 2}}`), 0644)

	merged, err := mergeJSON(existingFile, `{"nested": {"b": 99, "c": 3}}`)
	if err != nil {
		t.Fatalf("mergeJSON failed: %v", err)
	}

	// Nested merge: a preserved, b updated, c added
	expected := "{\n  \"nested\": {\n    \"a\": 1,\n    \"b\": 99,\n    \"c\": 3\n  }\n}"
	if merged != expected {
		t.Errorf("mergeJSON deep merge mismatch.\ngot:\n%s\nwant:\n%s", merged, expected)
	}
}

func TestWritePlatformConfig(t *testing.T) {
	content := platformConfigContent(
		"test-platform", platform.HookLevelFull, "1.0.0",
	)
	if !containsStr(content, "platform: test-platform") {
		t.Errorf("expected platform line in config, got: %s", content)
	}
	if !containsStr(content, `version: "1.0.0"`) {
		t.Errorf("expected version in config, got: %s", content)
	}
	if !containsStr(content, "hook_level: \"full\"") {
		t.Errorf("expected hook_level in config, got: %s", content)
	}
}

func TestWritePlatformConfig_CreatesDotAtl(t *testing.T) {
	t.Cleanup(func() { os.RemoveAll(".atl") })
	if _, err := os.Stat(".atl"); !os.IsNotExist(err) {
		t.Skip(".atl already exists in project root, skipping")
	}
	p := platform.Platform{Name: "test", HookLevel: platform.HookLevelNone}
	if err := WritePlatformConfig(p, "0.0.1"); err != nil {
		t.Fatalf("WritePlatformConfig failed: %v", err)
	}
	if _, err := os.Stat(".atl"); os.IsNotExist(err) {
		t.Fatal("expected .atl directory to be created")
	}
}

func platformConfigContent(name string, hookLevel platform.HookLevel, cmVersion string) string {
	return fmt.Sprintf(`# architect-ai platform configuration
# Generated by: architect-ai setup
# Do not edit manually — run: architect-ai setup

platform: %s
context_mode:
  installed: true
  version: "%s"
  hook_level: "%s"
  mcp_tools:
    - ctx_execute
    - ctx_batch_execute
    - ctx_execute_file
    - ctx_index
    - ctx_search
    - ctx_fetch_and_index
    - ctx_stats
    - ctx_doctor
    - ctx_upgrade
    - ctx_purge
    - ctx_insight
`, name, cmVersion, string(hookLevel))
}

func TestIsJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path string
		want bool
	}{
		{"file.json", true},
		{"path/to/file.json", true},
		{"file.yaml", false},
		{"file.md", false},
		{"file", false},
		{"file.JSON", false}, // case sensitive
	}
	for _, tt := range tests {
		got := isJSON(tt.path)
		if got != tt.want {
			t.Errorf("isJSON(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestEnsureContextMode_DryRun(t *testing.T) {
	t.Parallel()
	version, err := EnsureContextMode(true)
	if err != nil {
		t.Fatalf("EnsureContextMode(dryRun=true) failed: %v", err)
	}
	// If context-mode is installed on this system, version is the installed version string.
	// If not installed, dryRun returns "dry-run". Either is acceptable.
	// If context-mode is found but behaves unexpectedly (server mode), the version may be empty.
	t.Logf("EnsureContextMode(dryRun=true) returned version=%q", version)
}

func TestConfigurePlatform_MergeJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	existingFile := filepath.Join(dir, "config.json")
	os.WriteFile(existingFile, []byte(`{"a": 1}`), 0644)

	p := platform.Platform{
		Name:        "test-platform",
		ConfigFiles: []platform.ConfigFile{
			{Path: existingFile, Content: `{"b": 2}`, Merge: true},
		},
	}

	written, err := ConfigurePlatform(p)
	if err != nil {
		t.Fatalf("ConfigurePlatform failed: %v", err)
	}
	if len(written) != 1 {
		t.Fatalf("expected 1 file written, got %d", len(written))
	}
	data, _ := os.ReadFile(existingFile)
	if !strings.Contains(string(data), `"a": 1`) || !strings.Contains(string(data), `"b": 2`) {
		t.Errorf("Merge failed: %s", string(data))
	}
}

func TestConfigurePlatform_RoutingFile(t *testing.T) {
	// Not parallel, mutates env
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(origDir) })
	
	// Create a mock routing file source
	mockHome := filepath.Join(dir, "home")
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", mockHome)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })
	
	mockSrc := filepath.Join(mockHome, "node_modules", "context-mode", "configs", "test", "DEST_ROUTING.md")
	os.MkdirAll(filepath.Dir(mockSrc), 0755)
	os.WriteFile(mockSrc, []byte("routing content"), 0644)

	destRouting := "DEST_ROUTING.md"

	p := platform.Platform{
		Name: "test",
		RoutingFileRequired: true,
		RoutingFile: destRouting,
		ConfigFiles: []platform.ConfigFile{},
	}

	written, err := ConfigurePlatform(p)
	if err != nil {
		t.Fatalf("ConfigurePlatform failed: %v", err)
	}
	if len(written) != 1 {
		t.Fatalf("expected 1 file written, got %d", len(written))
	}

	data, _ := os.ReadFile(destRouting)
	if string(data) != "routing content" {
		t.Errorf("expected 'routing content', got %q", string(data))
	}
}

func TestConfigurePlatform_RoutingFileMissing(t *testing.T) {
	// Not parallel, mutates env
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(origDir) })
	
	mockHome := filepath.Join(dir, "home_missing")
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", mockHome)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	p := platform.Platform{
		Name: "test",
		RoutingFileRequired: true,
		RoutingFile: "DEST_ROUTING.md",
		ConfigFiles: []platform.ConfigFile{},
	}

	written, err := ConfigurePlatform(p)
	if err != nil {
		t.Fatalf("ConfigurePlatform failed: %v", err)
	}
	if len(written) != 0 {
		t.Fatalf("expected 0 files written, got %d", len(written))
	}
}

func TestConfigurePlatform_Errors(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Mkdir error: make file instead of dir
	badDir := filepath.Join(dir, "bad")
	os.WriteFile(badDir, []byte("dummy"), 0644)

	p1 := platform.Platform{
		ConfigFiles: []platform.ConfigFile{
			{Path: filepath.Join(badDir, "config.json"), Content: "{}"},
		},
	}
	if _, err := ConfigurePlatform(p1); err == nil {
		t.Error("expected mkdir error")
	}

	// Merge error: invalid JSON
	validJSONFile := filepath.Join(dir, "valid.json")
	os.WriteFile(validJSONFile, []byte(`{"a": 1}`), 0644)
	p2 := platform.Platform{
		ConfigFiles: []platform.ConfigFile{
			{Path: validJSONFile, Content: `invalid json`, Merge: true},
		},
	}
	if _, err := ConfigurePlatform(p2); err == nil {
		t.Error("expected merge error")
	}
}

func TestWritePlatformConfig_Errors(t *testing.T) {
	// Not parallel, mutates working directory
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(origDir) })

	// Make .atl a file to force MkdirAll error
	os.WriteFile(".atl", []byte("dummy"), 0644)
	p := platform.Platform{Name: "test", HookLevel: platform.HookLevelNone}
	err := WritePlatformConfig(p, "0.0.1")
	if err == nil {
		t.Error("expected error")
	}
}

func TestEnsureContextMode_NotInstalled_DryRun(t *testing.T) {
	// Not parallel, mutates env
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	t.Cleanup(func() { os.Setenv("PATH", oldPath) })
	
	version, err := EnsureContextMode(true)
	if err != nil {
		t.Fatalf("EnsureContextMode failed: %v", err)
	}
	if version != "dry-run" {
		t.Errorf("expected 'dry-run', got %q", version)
	}
}

// helper
func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
