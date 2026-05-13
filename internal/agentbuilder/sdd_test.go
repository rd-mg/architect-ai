package agentbuilder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeSDDAgent creates a GeneratedAgent with a given SDDConfig for testing.
func makeSDDAgent(name, trigger string, cfg *SDDIntegration) *GeneratedAgent {
	return &GeneratedAgent{
		Name:      name,
		Title:     "Test SDD Agent",
		Trigger:   trigger,
		Content:   "# Test SDD Agent\n",
		SDDConfig: cfg,
	}
}

func TestInjectSDDReference_InjectIntoEmptyFile(t *testing.T) {
	dir := t.TempDir()
	promptFile := filepath.Join(dir, "system-prompt.md")

	// Start with an empty file.
	if err := os.WriteFile(promptFile, []byte(""), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	agent := makeSDDAgent("my-skill", "When the user asks to do X", &SDDIntegration{
		Mode:        SDDPhaseSupport,
		TargetPhase: "apply",
	})

	if err := InjectSDDReference(agent, promptFile); err != nil {
		t.Fatalf("InjectSDDReference: %v", err)
	}

	data, err := os.ReadFile(promptFile)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)
	marker := "<!-- architect-ai:custom-agent:my-skill -->"
	if !strings.Contains(content, marker) {
		t.Errorf("marker not found in file;\ngot:\n%s", content)
	}
}

func TestInjectSDDReference_ExistingContentPreserved(t *testing.T) {
	dir := t.TempDir()
	promptFile := filepath.Join(dir, "system-prompt.md")

	existing := "# My System Prompt\n\nSome existing instructions here.\n"
	if err := os.WriteFile(promptFile, []byte(existing), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	agent := makeSDDAgent("my-skill", "When X happens", &SDDIntegration{
		Mode:        SDDPhaseSupport,
		TargetPhase: "spec",
	})

	if err := InjectSDDReference(agent, promptFile); err != nil {
		t.Fatalf("InjectSDDReference: %v", err)
	}

	data, _ := os.ReadFile(promptFile)
	content := string(data)

	if !strings.Contains(content, "My System Prompt") {
		t.Errorf("existing content was not preserved;\ngot:\n%s", content)
	}
	if !strings.Contains(content, "<!-- architect-ai:custom-agent:my-skill -->") {
		t.Errorf("marker not found in file;\ngot:\n%s", content)
	}
}

func TestInjectSDDReference_DuplicateInjection_MarkerReplacedNotDuplicated(t *testing.T) {
	dir := t.TempDir()
	promptFile := filepath.Join(dir, "system-prompt.md")

	if err := os.WriteFile(promptFile, []byte("# Prompt\n"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	agent := makeSDDAgent("dedup-skill", "When dedup needed", &SDDIntegration{
		Mode:        SDDPhaseSupport,
		TargetPhase: "verify",
	})

	// Inject twice.
	if err := InjectSDDReference(agent, promptFile); err != nil {
		t.Fatalf("first inject: %v", err)
	}
	if err := InjectSDDReference(agent, promptFile); err != nil {
		t.Fatalf("second inject: %v", err)
	}

	data, _ := os.ReadFile(promptFile)
	content := string(data)

	// Count marker occurrences — should appear exactly once.
	marker := "<!-- architect-ai:custom-agent:dedup-skill -->"
	count := strings.Count(content, marker)
	if count != 1 {
		t.Errorf("marker appears %d times, want exactly 1;\ngot:\n%s", count, content)
	}
}

func TestInjectSDDReference_NewPhaseMode_DependencyGraphReferenced(t *testing.T) {
	dir := t.TempDir()
	promptFile := filepath.Join(dir, "system-prompt.md")

	if err := os.WriteFile(promptFile, []byte("# Prompt\n"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	agent := makeSDDAgent("new-phase-skill", "When new phase starts", &SDDIntegration{
		Mode:      SDDNewPhase,
		PhaseName: "my-phase",
	})

	if err := InjectSDDReference(agent, promptFile); err != nil {
		t.Fatalf("InjectSDDReference: %v", err)
	}

	data, _ := os.ReadFile(promptFile)
	content := string(data)

	// New phase references the dependency graph.
	if !strings.Contains(content, "dependency graph") {
		t.Errorf("new-phase block should reference dependency graph;\ngot:\n%s", content)
	}
	if !strings.Contains(content, "my-phase") {
		t.Errorf("new-phase block should reference phase name 'my-phase';\ngot:\n%s", content)
	}
}

func TestInjectSDDReference_PhaseSupportMode_TargetPhaseReferenced(t *testing.T) {
	dir := t.TempDir()
	promptFile := filepath.Join(dir, "system-prompt.md")

	if err := os.WriteFile(promptFile, []byte("# Prompt\n"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	agent := makeSDDAgent("support-skill", "When supporting design phase", &SDDIntegration{
		Mode:        SDDPhaseSupport,
		TargetPhase: "design",
	})

	if err := InjectSDDReference(agent, promptFile); err != nil {
		t.Fatalf("InjectSDDReference: %v", err)
	}

	data, _ := os.ReadFile(promptFile)
	content := string(data)

	if !strings.Contains(content, "design") {
		t.Errorf("phase-support block should reference target phase 'design';\ngot:\n%s", content)
	}
	if !strings.Contains(content, "sdd-design") {
		t.Errorf("phase-support block should reference sdd-design trigger;\ngot:\n%s", content)
	}
}

func TestInjectSDDReference_StandaloneMode_IsNoop(t *testing.T) {
	dir := t.TempDir()
	promptFile := filepath.Join(dir, "system-prompt.md")

	original := "# My Prompt\n\nNo changes expected.\n"
	if err := os.WriteFile(promptFile, []byte(original), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	agent := makeSDDAgent("standalone-skill", "Never triggered", &SDDIntegration{
		Mode: SDDStandalone,
	})

	if err := InjectSDDReference(agent, promptFile); err != nil {
		t.Fatalf("InjectSDDReference: %v", err)
	}

	data, _ := os.ReadFile(promptFile)
	if string(data) != original {
		t.Errorf("standalone mode should be a no-op;\ngot:\n%s\nwant:\n%s", string(data), original)
	}
}

func TestInjectSDDReference_NilAgent_IsNoop(t *testing.T) {
	dir := t.TempDir()
	promptFile := filepath.Join(dir, "system-prompt.md")

	if err := os.WriteFile(promptFile, []byte("# Prompt\n"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := InjectSDDReference(nil, promptFile); err != nil {
		t.Fatalf("expected no error for nil agent, got: %v", err)
	}
}

func TestResolveMatchingSkills(t *testing.T) {
	dir := t.TempDir()
	registryPath := filepath.Join(dir, "skill-registry.md")

	registryContent := `# Skill Registry

<!-- architect-ai:registry:project -->
## Project Skills

| Trigger | Skill | Path |
|---------|-------|------|
| When writing Go tests, using teatest, or adding test coverage. | go-testing | internal/assets/skills/go-testing/SKILL.md |
| Safe, portable shell scripting | bash-expert | internal/assets/skills/bash-expert/SKILL.md |
| "Delegated by General Orchestrator for /solve intents." | analyst | internal/assets/skills/analyst/SKILL.md |
<!-- /architect-ai:registry:project -->

<!-- architect-ai:registry:compact-rules -->
## compact-rules
### go-testing
- Use table-driven tests
- Use teatest for TUI

### bash-expert
- Use set -euo pipefail
- Quote variables

### analyst
- Unified Handshake mandatory
- Empirical proof required
<!-- /architect-ai:registry:compact-rules -->
`
	if err := os.WriteFile(registryPath, []byte(registryContent), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	tests := []struct {
		name      string
		taskPaths []string
		want      []string // skill names that should be in the output
		dontWant  []string // skill names that should NOT be in the output
	}{
		{
			name:      "matches go-testing",
			taskPaths: []string{"internal/pipeline/runner_test.go"},
			want:      []string{"go-testing"},
			dontWant:  []string{"bash-expert", "analyst"},
		},
		{
			name:      "matches bash-expert",
			taskPaths: []string{"scripts/install.sh"},
			want:      []string{"bash-expert"},
			dontWant:  []string{"go-testing", "analyst"},
		},
		{
			name:      "matches analyst via keyword",
			taskPaths: []string{"solve-this-bug.txt"},
			want:      []string{"analyst"},
			dontWant:  []string{"go-testing", "bash-expert"},
		},
		{
			name:      "matches multiple",
			taskPaths: []string{"main.go", "deploy.sh"},
			want:      []string{"go-testing", "bash-expert"},
			dontWant:  []string{"analyst"},
		},
		{
			name:      "no match",
			taskPaths: []string{"README.md"},
			want:      []string{},
			dontWant:  []string{"go-testing", "bash-expert", "analyst"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveMatchingSkills(registryPath, tt.taskPaths)
			if err != nil {
				t.Fatalf("ResolveMatchingSkills: %v", err)
			}

			for _, want := range tt.want {
				if !strings.Contains(got, "### "+want) {
					t.Errorf("missing expected skill %q in output;\ngot:\n%s", want, got)
				}
			}

			for _, dontWant := range tt.dontWant {
				if strings.Contains(got, "### "+dontWant) {
					t.Errorf("found unexpected skill %q in output;\ngot:\n%s", dontWant, got)
				}
			}
		})
	}
}
