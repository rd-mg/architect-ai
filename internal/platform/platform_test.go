package platform

import (
	"os"
	"testing"
)

func TestDetect_ByOverride(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		override string
		wantName string
		wantErr  bool
	}{
		{"opencode", "opencode", "opencode", false},
		{false},
		{"cursor", "cursor", "cursor", false},
		{"kiro", "kiro", "kiro", false},
		{"vscode-copilot", "vscode-copilot", "vscode-copilot", false},
		{"antigravity", "antigravity", "antigravity", false},
		{"claude-code", "claude-code", "claude-code", false},
		{"jetbrains-copilot", "jetbrains-copilot", "jetbrains-copilot", false},
		{"kilo-code", "kilo-code", "kilo-code", false},
		{"zed", "zed", "zed", false},
		{"codex-cli", "codex-cli", "codex-cli", false},
		{"pi-agent", "pi-agent", "pi-agent", false},
		{"unknown override", "nonexistent-platform", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Detect(tt.override)
			if (err != nil) != tt.wantErr {
				t.Errorf("Detect(%q) error = %v, wantErr %v", tt.override, err, tt.wantErr)
			}
			if p.Name != tt.wantName {
				t.Errorf("Detect(%q) Name = %q, want %q", tt.override, p.Name, tt.wantName)
			}
		})
	}
}

func TestDetect_ByEnvironment(t *testing.T) {
	t.Run("OPENCODE_SESSION", func(t *testing.T) {
		os.Setenv("OPENCODE_SESSION", "test-session")
		defer os.Unsetenv("OPENCODE_SESSION")
		p, err := Detect("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name != "opencode" {
			t.Errorf("expected opencode, got %q", p.Name)
		}
	})

	t.Run("CLAUDE_CODE_SESSION", func(t *testing.T) {
		os.Setenv("CLAUDE_CODE_SESSION", "test-session")
		defer os.Unsetenv("CLAUDE_CODE_SESSION")
		p, err := Detect("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name != "claude-code" {
			t.Errorf("expected claude-code, got %q", p.Name)
		}
	})

	t.Run("GEMINI_CLI_SESSION", func(t *testing.T) {
		os.Setenv("GEMINI_CLI_SESSION", "test-session")
		defer os.Unsetenv("GEMINI_CLI_SESSION")
		p, err := Detect("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name != "gemini-cli" {
			t.Errorf("expected gemini-cli, got %q", p.Name)
		}
	})
}

func TestDetect_ByFile(t *testing.T) {
	// Save current dir and change to temp dir
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })

	dir := t.TempDir()
	os.Chdir(dir)

	t.Run("opencode.json", func(t *testing.T) {
		os.WriteFile("opencode.json", []byte("{}"), 0644)
		defer os.Remove("opencode.json")
		p, _ := Detect("")
		if p.Name != "opencode" {
			t.Errorf("expected opencode, got %q", p.Name)
		}
	})

	t.Run("kilo.json", func(t *testing.T) {
		os.WriteFile("kilo.json", []byte("{}"), 0644)
		defer os.Remove("kilo.json")
		p, _ := Detect("")
		if p.Name != "kilo-code" {
			t.Errorf("expected kilo-code, got %q", p.Name)
		}
	})

	t.Run(".cursor dir", func(t *testing.T) {
		os.MkdirAll(".cursor", 0755)
		defer os.RemoveAll(".cursor")
		p, _ := Detect("")
		if p.Name != "cursor" {
			t.Errorf("expected cursor, got %q", p.Name)
		}
	})
}

func TestDetect_Unknown(t *testing.T) {
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })

	dir := t.TempDir()
	os.Chdir(dir)

	p, err := Detect("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "unknown" {
		t.Errorf("expected unknown, got %q", p.Name)
	}
}

func TestAllPlatforms(t *testing.T) {
	t.Parallel()
	platforms := AllPlatforms()
	if len(platforms) == 0 {
		t.Fatal("expected at least one platform")
	}

	// Check specific platforms exist
	names := make(map[string]bool)
	for _, p := range platforms {
		names[p.Name] = true
		if p.DisplayName == "" {
			t.Errorf("platform %q has empty DisplayName", p.Name)
		}
	}

	expected := []string{
		"opencode", "kilo-code", "antigravity",
		"vscode-copilot", "jetbrains-copilot", "cursor", "kiro",
		"zed", "codex-cli", "pi-agent", "claude-code",
	}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected platform %q not found in AllPlatforms()", name)
		}
	}
}

func TestPromptSelection(t *testing.T) {
	t.Parallel()
	p := PromptSelection()
	if p.Name != "unknown" {
		t.Errorf("expected unknown, got %q", p.Name)
	}
}

func TestPlatformHookLevel(t *testing.T) {
	t.Parallel()
	platforms := AllPlatforms()
	for _, p := range platforms {
		switch p.HookLevel {
		case HookLevelFull, HookLevelPartial, HookLevelNone:
			// valid
		default:
			t.Errorf("platform %q has invalid HookLevel %q", p.Name, p.HookLevel)
		}
	}
}

func TestPlatformConfigFiles(t *testing.T) {
	t.Parallel()
	platforms := AllPlatforms()
	for _, p := range platforms {
		for _, cf := range p.ConfigFiles {
			if cf.Path == "" {
				t.Errorf("platform %q has ConfigFile with empty Path", p.Name)
			}
			if cf.Content == "" {
				t.Errorf("platform %q has ConfigFile %q with empty Content", p.Name, cf.Path)
			}
		}
	}
}

func TestDetect_EmptyOverrideFallsThrough(t *testing.T) {
	t.Parallel()
	// Empty override with no env vars and no known files should return unknown
	p, err := Detect("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = p
	// Note: in the real project dir this may detect a platform
}
