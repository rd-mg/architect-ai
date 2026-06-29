package antigravitycli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/rd-mg/architect-ai/internal/model"
	"github.com/rd-mg/architect-ai/internal/system"
)

func TestAntigravityCLIAdapterIdentity(t *testing.T) {
	a := NewAdapter()
	if a.Agent() != model.AgentAntigravityCLI {
		t.Errorf("wrong agent ID: got %q want %q", a.Agent(), model.AgentAntigravityCLI)
	}
	if a.SupportsAutoInstall() {
		t.Error("antigravity-cli should not support auto-install")
	}
	if !a.SupportsMCP() {
		t.Error("antigravity-cli should support MCP")
	}
}

func TestAntigravityCLIPathsResolveCorrectly(t *testing.T) {
	a := NewAdapter()
	home := "/home/testuser"
	want := "/home/testuser/.gemini/antigravity-cli/plugins/architect-ai/mcp_config.json"
	got := a.MCPConfigPath(home, "any")
	if got != want {
		t.Errorf("MCPConfigPath: got %q want %q", got, want)
	}
}

func TestDetect(t *testing.T) {
	tests := []struct {
		name            string
		stat            statResult
		lookPathErr     error
		wantInstalled   bool
		wantBinaryPath  string
		wantConfigPath  string
		wantConfigFound bool
		wantErr         bool
	}{
		{
			name:            "binary on PATH and plugin dir found",
			stat:            statResult{isDir: true},
			lookPathErr:     nil,
			wantInstalled:   true,
			wantBinaryPath:  "/usr/local/bin/agy",
			wantConfigPath:  filepath.Join("/tmp/home", ".gemini", "antigravity-cli", "plugins"),
			wantConfigFound: true,
		},
		{
			name:            "binary on PATH but plugin dir missing",
			stat:            statResult{err: os.ErrNotExist},
			lookPathErr:     nil,
			wantInstalled:   true,
			wantBinaryPath:  "/usr/local/bin/agy",
			wantConfigPath:  filepath.Join("/tmp/home", ".gemini", "antigravity-cli", "plugins"),
			wantConfigFound: false,
		},
		{
			name:            "no binary but plugin dir found",
			stat:            statResult{isDir: true},
			lookPathErr:     errors.New("not found"),
			wantInstalled:   false,
			wantBinaryPath:  "",
			wantConfigPath:  filepath.Join("/tmp/home", ".gemini", "antigravity-cli", "plugins"),
			wantConfigFound: true,
		},
		{
			name:    "stat error bubbles up",
			stat:    statResult{err: errors.New("permission denied")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Adapter{
				lookPath: func(string) (string, error) {
					return "/usr/local/bin/agy", tt.lookPathErr
				},
				statPath: func(string) statResult {
					return tt.stat
				},
			}

			installed, binaryPath, configPath, configFound, err := a.Detect(context.Background(), "/tmp/home")
			if (err != nil) != tt.wantErr {
				t.Fatalf("Detect() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if installed != tt.wantInstalled {
				t.Fatalf("Detect() installed = %v, want %v", installed, tt.wantInstalled)
			}

			if installed && binaryPath != tt.wantBinaryPath {
				t.Fatalf("Detect() binaryPath = %q, want %q", binaryPath, tt.wantBinaryPath)
			}

			if configPath != tt.wantConfigPath {
				t.Fatalf("Detect() configPath = %q, want %q", configPath, tt.wantConfigPath)
			}

			if configFound != tt.wantConfigFound {
				t.Fatalf("Detect() configFound = %v, want %v", configFound, tt.wantConfigFound)
			}
		})
	}
}

func TestInstallCommand(t *testing.T) {
	a := NewAdapter()

	_, err := a.InstallCommand(system.PlatformProfile{OS: "darwin"})
	if err == nil {
		t.Fatal("InstallCommand() expected error for CLI agent, got nil")
	}

	var notInstallable AgentNotInstallableError
	if !errors.As(err, &notInstallable) {
		t.Fatalf("InstallCommand() error type = %T, want AgentNotInstallableError", err)
	}

	if notInstallable.Agent != model.AgentAntigravityCLI {
		t.Fatalf("AgentNotInstallableError.Agent = %q, want %q", notInstallable.Agent, model.AgentAntigravityCLI)
	}
}

func TestCapabilities(t *testing.T) {
	a := NewAdapter()

	if !a.SupportsSkills() {
		t.Fatal("SupportsSkills() = false, want true")
	}

	if !a.SupportsSystemPrompt() {
		t.Fatal("SupportsSystemPrompt() = false, want true")
	}

	if !a.SupportsMCP() {
		t.Fatal("SupportsMCP() = false, want true")
	}

	if a.SupportsOutputStyles() {
		t.Fatal("SupportsOutputStyles() = true, want false")
	}

	if !a.SupportsSlashCommands() {
		t.Fatal("SupportsSlashCommands() = false, want true")
	}

	if got := a.OutputStyleDir("/tmp/home"); got != "" {
		t.Fatalf("OutputStyleDir() = %q, want empty string", got)
	}

	if got := a.CommandsDir("/tmp/home"); got != "" {
		t.Fatalf("CommandsDir() = %q, want empty string", got)
	}
}

func TestStrategies(t *testing.T) {
	a := NewAdapter()

	if got := a.SystemPromptStrategy(); got != model.StrategyMarkdownSections {
		t.Fatalf("SystemPromptStrategy() = %v, want StrategyMarkdownSections", got)
	}

	if got := a.MCPStrategy(); got != model.StrategyMCPConfigFile {
		t.Fatalf("MCPStrategy() = %v, want StrategyMCPConfigFile", got)
	}
}

func TestConfigPathsCrossPlatform(t *testing.T) {
	a := NewAdapter()
	home := "/tmp/home"

	if got := a.GlobalConfigDir(home); got != filepath.Join(home, ".gemini", "antigravity-cli", "plugins", "architect-ai") {
		t.Fatalf("GlobalConfigDir() = %q, want %q", got, filepath.Join(home, ".gemini", "antigravity-cli", "plugins", "architect-ai"))
	}

	if got := a.SkillsDir(home); got != filepath.Join(home, ".gemini", "antigravity-cli", "skills") {
		t.Fatalf("SkillsDir() = %q, want %q", got, filepath.Join(home, ".gemini", "antigravity-cli", "skills"))
	}

	if got := a.MCPConfigPath(home, "ctx7"); got != filepath.Join(home, ".gemini", "antigravity-cli", "plugins", "architect-ai", "mcp_config.json") {
		t.Fatalf("MCPConfigPath() = %q, want %q", got, filepath.Join(home, ".gemini", "antigravity-cli", "plugins", "architect-ai", "mcp_config.json"))
	}

	if got := a.SystemPromptFile(home); got != filepath.Join(home, ".gemini", "GEMINI.md") {
		t.Fatalf("SystemPromptFile() = %q, want %q", got, filepath.Join(home, ".gemini", "GEMINI.md"))
	}

	if got := a.SettingsPath(home); got != filepath.Join(home, ".gemini", "antigravity-cli", "settings.json") {
		t.Fatalf("SettingsPath() = %q, want %q", got, filepath.Join(home, ".gemini", "antigravity-cli", "settings.json"))
	}

	if got := a.SystemPromptDir(home); got != filepath.Join(home, ".gemini", "antigravity-cli", "skills") {
		t.Fatalf("SystemPromptDir() = %q, want %q", got, filepath.Join(home, ".gemini", "antigravity-cli", "skills"))
	}
}
