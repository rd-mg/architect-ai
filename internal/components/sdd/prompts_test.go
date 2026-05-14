package sdd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rd-mg/architect-ai/internal/agents"
	"github.com/rd-mg/architect-ai/internal/model"
	"github.com/rd-mg/architect-ai/internal/state"
)

// TestSharedPromptDir verifies the expected directory path is returned.
func TestSharedPromptDir(t *testing.T) {
	home := "/home/testuser"
	want := "/home/testuser/.config/opencode/prompts/sdd"
	got := SharedPromptDir(home)
	if got != want {
		t.Fatalf("SharedPromptDir(%q) = %q, want %q", home, got, want)
	}
}

// TestWriteSharedPromptFilesCreates10Files verifies that WriteSharedPromptFiles
// creates exactly the 10 expected prompt files under {homeDir}/.config/opencode/prompts/sdd/.
func TestWriteSharedPromptFilesCreates10Files(t *testing.T) {
	home := t.TempDir()
	manifest := state.NewManagedManifest(model.AgentOpenCode, "/test")

	changed, err := WriteSharedPromptFiles(home, manifest)
	if err != nil {
		t.Fatalf("WriteSharedPromptFiles() error = %v", err)
	}
	if !changed {
		t.Fatal("WriteSharedPromptFiles() first call changed = false, want true")
	}

	expectedFiles := []string{
		"sdd-init.md",
		"sdd-explore.md",
		"sdd-propose.md",
		"sdd-spec.md",
		"sdd-design.md",
		"sdd-tasks.md",
		"sdd-apply.md",
		"sdd-verify.md",
		"sdd-archive.md",
		"sdd-onboard.md",
	}

	promptDir := SharedPromptDir(home)
	for _, fileName := range expectedFiles {
		path := filepath.Join(promptDir, fileName)
		info, statErr := os.Stat(path)
		if statErr != nil {
			t.Errorf("prompt file %q not found: %v", path, statErr)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("prompt file %q is empty", path)
		}
	}
}

// TestWriteSharedPromptFilesIdempotent verifies that calling WriteSharedPromptFiles
// twice returns changed=false on the second call.
func TestWriteSharedPromptFilesIdempotent(t *testing.T) {
	home := t.TempDir()
	manifest := state.NewManagedManifest(model.AgentOpenCode, "/test")

	first, err := WriteSharedPromptFiles(home, manifest)
	if err != nil {
		t.Fatalf("WriteSharedPromptFiles() first error = %v", err)
	}
	if !first {
		t.Fatal("WriteSharedPromptFiles() first call changed = false, want true")
	}

	second, err := WriteSharedPromptFiles(home, manifest)
	if err != nil {
		t.Fatalf("WriteSharedPromptFiles() second error = %v", err)
	}
	if second {
		t.Fatal("WriteSharedPromptFiles() second call changed = true, want false (idempotent)")
	}
}

// TestWriteSharedPromptFilesContent verifies each prompt file contains the
// executor-scoped sub-agent prompt content for the correct phase.
func TestWriteSharedPromptFilesContent(t *testing.T) {
	home := t.TempDir()
	manifest := state.NewManagedManifest(model.AgentOpenCode, "/test")

	if _, err := WriteSharedPromptFiles(home, manifest); err != nil {
		t.Fatalf("WriteSharedPromptFiles() error = %v", err)
	}

	promptDir := SharedPromptDir(home)

	phases := []struct {
		file  string
		phase string
	}{
		{"sdd-init.md", "init"},
		{"sdd-explore.md", "explore"},
		{"sdd-propose.md", "propose"},
		{"sdd-spec.md", "spec"},
		{"sdd-design.md", "design"},
		{"sdd-tasks.md", "tasks"},
		{"sdd-apply.md", "apply"},
		{"sdd-verify.md", "verify"},
		{"sdd-archive.md", "archive"},
		{"sdd-onboard.md", "onboard"},
	}

	for _, tc := range phases {
		path := filepath.Join(promptDir, tc.file)
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Errorf("ReadFile(%q) error = %v", path, readErr)
			continue
		}

		content := string(data)

		// Each file must contain the phase name (executor-scoped prompt).
		if !strings.Contains(content, tc.phase) {
			t.Errorf("prompt file %q missing phase %q in content", tc.file, tc.phase)
		}

		// Each file must contain the key executor-scoped markers.
		for _, marker := range []string{"not the orchestrator", "Do NOT delegate", "Do NOT launch sub-agents"} {
			if !strings.Contains(content, marker) {
				t.Errorf("prompt file %q missing required marker %q", tc.file, marker)
			}
		}
	}
}

// TestInjectOpenCodeMultiModeSubagentPromptsUseFilePaths verifies that after
// injection in multi-mode, each sub-agent's prompt field in opencode.json
// contains a {file:...} reference (not an inline string).
func TestInjectOpenCodeMultiModeSubagentPromptsUseFilePaths(t *testing.T) {
	home := t.TempDir()
	mockNoPackageManager(t)

	if _, err := Inject(home, opencodeAdapter(), "multi"); err != nil {
		t.Fatalf("Inject(multi) error = %v", err)
	}

	settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(opencode.json) error = %v", err)
	}

	promptDir := SharedPromptDir(home)

	text := string(content)
	for _, phase := range []string{"sdd-init", "sdd-explore", "sdd-propose", "sdd-spec", "sdd-design", "sdd-tasks", "sdd-apply", "sdd-verify", "sdd-archive", "sdd-onboard"} {
		expectedRef := "{file:" + filepath.Join(promptDir, phase+".md") + "}"
		if !strings.Contains(text, expectedRef) {
			t.Errorf("opencode.json sub-agent %q missing {file:...} reference %q", phase, expectedRef)
		}
	}
}

// TestInjectOpenCodeMultiModeOrchestratorPromptIsStillInlined verifies that
// the orchestrator prompt is still inlined (not a file reference) after injection.
func TestInjectOpenCodeMultiModeOrchestratorPromptIsStillInlined(t *testing.T) {
	home := t.TempDir()
	mockNoPackageManager(t)

	if _, err := Inject(home, opencodeAdapter(), "multi"); err != nil {
		t.Fatalf("Inject(multi) error = %v", err)
	}

	settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(opencode.json) error = %v", err)
	}

	text := string(content)

	// The orchestrator still uses {file:./AGENTS.md} from the overlay (not from prompts/).
	// We check that there's NO file reference inside the prompts/sdd/ dir for orchestrator.
	promptDir := SharedPromptDir(home)
	if strings.Contains(text, filepath.Join(promptDir, "sdd-orchestrator.md")) {
		t.Fatal("orchestrator should NOT use a file reference from prompts/sdd/")
	}
}

// TestInjectOpenCodeMultiModeIdempotentWithPromptFiles verifies that the second
// Inject call returns changed=false when prompt files are already on disk.
func TestInjectOpenCodeMultiModeIdempotentWithPromptFiles(t *testing.T) {
	home := t.TempDir()
	mockNoPackageManager(t)

	first, err := Inject(home, opencodeAdapter(), "multi")
	if err != nil {
		t.Fatalf("Inject(multi) first error = %v", err)
	}
	if !first.Changed {
		t.Fatal("Inject(multi) first changed = false")
	}

	second, err := Inject(home, opencodeAdapter(), "multi")
	if err != nil {
		t.Fatalf("Inject(multi) second error = %v", err)
	}
	if second.Changed {
		t.Fatal("Inject(multi) second changed = true — should be idempotent with prompt files")
	}
}

// TestL0ThinkingAgentMindsetInjected verifies that the L0 Thinking Agent mindset
// is injected into the general-orchestrator (Thinking Agent) prompt.
func TestL0ThinkingAgentMindsetInjected(t *testing.T) {
	home := t.TempDir()
	mockNoPackageManager(t)

	// We use gemini as a representative adapter for L0/L1 topology
	geminiAdapter, _ := agents.NewAdapter("gemini-cli")
	
	_, err := Inject(home, geminiAdapter, "")
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	promptPath := geminiAdapter.SystemPromptFile(home)
	data, _ := os.ReadFile(promptPath)
	content := string(data)

	// Check for L0 Thinking Agent header and Sentinel mindset
	if !strings.Contains(content, "# Thinking Agent (L0 Strategic Sentinel)") {
		t.Error("System prompt missing L0 Thinking Agent header")
	}
	if !strings.Contains(content, "## Mindset & Strategic Supervision") {
		t.Error("System prompt missing Sentinel mindset section")
	}
	if !strings.Contains(content, "## Intention Gate (MANDATORY)") {
		t.Error("System prompt missing Intention Gate section")
	}
}

// TestArchitectureGuardrailsInjected verifies that the Architecture Guardrails
// skill content is injected into the Thinking Agent prompt.
func TestArchitectureGuardrailsInjected(t *testing.T) {
	home := t.TempDir()
	mockNoPackageManager(t)

	geminiAdapter, _ := agents.NewAdapter("gemini-cli")
	_, err := Inject(home, geminiAdapter, "")
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	promptPath := geminiAdapter.SystemPromptFile(home)
	data, _ := os.ReadFile(promptPath)
	content := string(data)

	// Check for Guardrails markers and key content
	if !strings.Contains(content, "<!-- architect-ai:architecture-guardrails:START -->") {
		t.Error("System prompt missing Architecture Guardrails start marker")
	}
	if !strings.Contains(content, "## Core Guardrails (REQUIRED)") {
		t.Error("System prompt missing Core Guardrails heading")
	}
	if !strings.Contains(content, "Thin Adapters") {
		t.Error("System prompt missing 'Thin Adapters' guardrail")
	}
}

// TestSequentialThinkingHarmonizationInjected verifies that the harmonization
// instructions for sequential_thinking are present in the system prompts.
func TestSequentialThinkingHarmonizationInjected(t *testing.T) {
	home := t.TempDir()
	mockNoPackageManager(t)

	// We use gemini as a representative adapter for L0/L1 topology
	adapter, _ := agents.NewAdapter("gemini-cli")
	_, err := Inject(home, adapter, "")
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	promptPath := adapter.SystemPromptFile(home)
	data, _ := os.ReadFile(promptPath)
	content := string(data)

	// Check for Intention Gate using sequential_thinking in L0
	if !strings.Contains(content, "## Intention Gate (MANDATORY)") {
		t.Error("Thinking Agent prompt missing Intention Gate section")
	}
	if !strings.Contains(content, "use the `sequential_thinking` tool (if available) to analyze the user request") {
		t.Error("Thinking Agent prompt missing Sequential Thinking mandate")
	}
}


// TestL1TacticalOrchestratorHeaderInjected verifies that the SDD orchestrator
// prompt contains the L1 Tactical Orchestrator header.
func TestL1TacticalOrchestratorHeaderInjected(t *testing.T) {
	home := t.TempDir()
	mockNoPackageManager(t)

	geminiAdapter, _ := agents.NewAdapter("gemini-cli")
	_, err := Inject(home, geminiAdapter, "")
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	promptPath := geminiAdapter.SystemPromptFile(home)
	data, _ := os.ReadFile(promptPath)
	content := string(data)

	// Check for L1 Tactical Orchestrator header and supervision note
	if !strings.Contains(content, "# Agent Teams Lite — L1 Tactical Orchestrator") {
		t.Error("System prompt missing L1 Tactical Orchestrator header")
	}
	if !strings.Contains(content, "You operate under the strategic guidance of the **L0 Thinking Agent (Strategic Sentinel)**") {
		t.Error("System prompt missing supervision link to L0 Sentinel")
	}
}
