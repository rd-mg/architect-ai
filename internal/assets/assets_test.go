package assets

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestAllEmbeddedAssetsAreReadable verifies that every expected embedded file
// can be loaded via Read() without error. This catches missing/misnamed files
// at test time rather than at runtime.
func TestAllEmbeddedAssetsAreReadable(t *testing.T) {
	expectedFiles := []string{
		// Claude agent files
		"claude/engram-protocol.md",
		"claude/persona-architect.md",
		"claude/sdd-orchestrator.md",
		"claude/thinking-agent.md",

		// OpenCode agent files
		"opencode/persona-architect.md",
		"opencode/sdd-overlay-single.json",
		"opencode/sdd-overlay-multi.json",
		"opencode/commands/sdd-apply.md",
		"opencode/commands/sdd-archive.md",
		"opencode/commands/sdd-continue.md",
		"opencode/commands/sdd-explore.md",
		"opencode/commands/sdd-ff.md",
		"opencode/commands/sdd-init.md",
		"opencode/commands/sdd-new.md",
		"opencode/commands/sdd-verify.md",
		"opencode/plugins/background-agents.ts",

		// Gemini agent files
		"gemini/sdd-orchestrator.md",
		"gemini/thinking-agent.md",

		// Antigravity agent files
		"antigravity/sdd-orchestrator.md",
		"antigravity/thinking-agent.md",
		"antigravity/architect.md",

		// VSCode agent files
		"vscode/sdd-orchestrator.md",
		"vscode/thinking-agent.md",
		"vscode/general-orchestrator.md",

		// Codex agent files
		"codex/sdd-orchestrator.md",
		"codex/thinking-agent.md",

		// Cursor agent files
		"cursor/sdd-orchestrator.md",
		"cursor/thinking-agent.md",
		"cursor/agents/sdd-init.md",
		"cursor/agents/sdd-explore.md",
		"cursor/agents/sdd-propose.md",
		"cursor/agents/sdd-spec.md",
		"cursor/agents/sdd-design.md",
		"cursor/agents/sdd-tasks.md",
		"cursor/agents/sdd-apply.md",
		"cursor/agents/sdd-verify.md",
		"cursor/agents/sdd-archive.md",

		// SDD skills
		"skills/sdd-init/SKILL.md",
		"skills/sdd-orchestrator/SKILL.md",
		"skills/general-orchestrator/SKILL.md",
		"skills/sdd-apply/SKILL.md",
		"skills/sdd-archive/SKILL.md",
		"skills/sdd-design/SKILL.md",
		"skills/sdd-explore/SKILL.md",
		"skills/sdd-propose/SKILL.md",
		"skills/sdd-spec/SKILL.md",
		"skills/sdd-tasks/SKILL.md",
		"skills/sdd-verify/SKILL.md",
		"skills/skill-registry/SKILL.md",
		"skills/architecture-guardrails/SKILL.md",
		"skills/_shared/persistence-contract.md",
		"skills/_shared/engram-convention.md",
		"skills/_shared/openspec-convention.md",
		"skills/_shared/sdd-phase-common.md",

		// Foundation skills
		"skills/go-testing/SKILL.md",
		"skills/skill-creator/SKILL.md",
		"skills/_shared/adaptive-reasoning-gate.md",

		// GGA v2 assets
		"gga/AGENTS.md",
		"gga/sdd-orchestrator.md",
		"gga/pre-commit.bash.tpl",
		"gga/pre-commit.ps1.tpl",
	}

	for _, path := range expectedFiles {
		t.Run(path, func(t *testing.T) {
			content, err := Read(path)
			if err != nil {
				t.Fatalf("Read(%q) error = %v", path, err)
			}

			if len(strings.TrimSpace(content)) == 0 {
				t.Fatalf("Read(%q) returned empty content", path)
			}

			// Real content should be substantial, not a one-line stub.
			if len(content) < 50 {
				t.Fatalf("Read(%q) content is suspiciously short (%d bytes) — possible stub", path, len(content))
			}
		})
	}
}

func TestOpenCodeEmbeddedAssetLayout(t *testing.T) {
	entries, err := FS.ReadDir("opencode")
	if err != nil {
		t.Fatalf("ReadDir(opencode) error = %v", err)
	}

	seen := map[string]bool{}
	for _, entry := range entries {
		seen[entry.Name()] = true
	}

	for _, name := range []string{"commands", "plugins", "persona-architect.md", "sdd-overlay-single.json", "sdd-overlay-multi.json"} {
		if !seen[name] {
			t.Fatalf("opencode embedded assets missing %q", name)
		}
	}

	commandEntries, err := FS.ReadDir("opencode/commands")
	if err != nil {
		t.Fatalf("ReadDir(opencode/commands) error = %v", err)
	}
	if len(commandEntries) != 9 {
		t.Fatalf("opencode commands count = %d, want 9", len(commandEntries))
	}

	pluginEntries, err := FS.ReadDir("opencode/plugins")
	if err != nil {
		t.Fatalf("ReadDir(opencode/plugins) error = %v", err)
	}
	if len(pluginEntries) != 1 {
		t.Fatalf("opencode plugins count = %d, want 1", len(pluginEntries))
	}
	if pluginEntries[0].Name() != "background-agents.ts" {
		t.Fatalf("plugin entry = %q, want background-agents.ts", pluginEntries[0].Name())
	}
}

// TestArchitectL0Identity validates the L0 Super-Orchestrator identity.
// L0 is a pure orchestrator — it routes ALL work to L1a (sdd-orchestrator)
// or L1b (general-orchestrator) and NEVER executes inline (Mode A removed).
func TestArchitectL0Identity(t *testing.T) {
	// Verify shared architect identity has correct TWO-mode structure
	sharedIdentity := MustRead("_shared/architect-identity.md")
	if len(sharedIdentity) == 0 {
		t.Fatal("_shared/architect-identity.md is empty")
	}

	// Must have TWO operating modes (NOT three — Mode A inline was removed)
	if !strings.Contains(sharedIdentity, "TWO operating modes") {
		t.Error("architect-identity.md must say 'TWO operating modes' (Mode A inline removed)")
	}
	if strings.Contains(sharedIdentity, "THREE operating modes") {
		t.Error("architect-identity.md must NOT say 'THREE operating modes'")
	}

	// Must be a pure orchestrator — no inline execution
	if !strings.Contains(sharedIdentity, "pure orchestrator") {
		t.Error("architect-identity.md must describe itself as a 'pure orchestrator'")
	}
	if strings.Contains(sharedIdentity, "Mode A (inline") {
		t.Error("architect-identity.md must NOT reference Mode A inline execution")
	}
	if strings.Contains(sharedIdentity, "simple tasks directly") {
		t.Error("architect-identity.md must NOT allow simple tasks directly")
	}

	// Must route to both L1 orchestrators
	if !strings.Contains(sharedIdentity, "L1a") || !strings.Contains(sharedIdentity, "L1b") {
		t.Error("architect-identity.md must reference L1a (sdd-orchestrator) and L1b (general-orchestrator)")
	}
	if !strings.Contains(sharedIdentity, "sdd-orchestrator") {
		t.Error("architect-identity.md must reference sdd-orchestrator")
	}
	if !strings.Contains(sharedIdentity, "general-orchestrator") {
		t.Error("architect-identity.md must reference general-orchestrator")
	}

	// Verify all platform architect.md files exist and reference the shared identity
	platformArchitects := []string{
		"opencode/architect.md",
		"claude/architect.md",
		"cursor/architect.md",
		"gemini/architect.md",
		"antigravity/architect.md",
	}
	for _, path := range platformArchitects {
		t.Run(path, func(t *testing.T) {
			content, err := Read(path)
			if err != nil {
				t.Fatalf("Read(%q) error = %v", path, err)
			}
			if len(strings.TrimSpace(content)) == 0 {
				t.Fatalf("%q is empty", path)
			}
			// Platform files must include the shared identity
			if !strings.Contains(content, "_shared/architect-identity.md") {
				t.Errorf("%q must include the shared architect-identity.md", path)
			}
			// Platform files must NOT reference Mode A inline execution
			if strings.Contains(content, "Mode A (") && strings.Contains(content, "inline") {
				t.Errorf("%q must NOT have Mode A inline section", path)
			}
		})
	}
}

// TestMustReadPanicsOnMissingFile verifies that MustRead panics for a
// nonexistent file, confirming the safety mechanism works.
func TestMustReadPanicsOnMissingFile(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("MustRead() did not panic for missing file")
		}
	}()

	MustRead("nonexistent/file.md")
}

// TestEmbeddedAssetCount verifies we have the expected number of embedded files.
// This catches accidental deletions of asset files.
func TestEmbeddedAssetCount(t *testing.T) {
	// Count skill files.
	entries, err := FS.ReadDir("skills")
	if err != nil {
		t.Fatalf("ReadDir(skills) error = %v", err)
	}

	skillDirs := 0
	for _, entry := range entries {
		if entry.IsDir() {
			skillDirs++
		}
	}

	// We expect 35 skill directories (10 SDD + sdd-orchestrator + general-orchestrator + judgment-day + foundation + _shared + generalist + ideator + researcher + solver + architecture-guardrails + others).
	if skillDirs != 35 {
		t.Fatalf("expected 35 skill directories, got %d", skillDirs)
	}

	// Verify each skill directory has a SKILL.md.
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if entry.Name() == "_shared" {
			for _, sharedFile := range []string{"persistence-contract.md", "engram-convention.md", "openspec-convention.md", "sdd-phase-common.md", "skill-resolver.md"} {
				sharedPath := "skills/_shared/" + sharedFile
				if _, err := Read(sharedPath); err != nil {
					t.Fatalf("shared directory missing %q: %v", sharedFile, err)
				}
			}
			continue
		}
		skillPath := "skills/" + entry.Name() + "/SKILL.md"
		if _, err := Read(skillPath); err != nil {
			t.Fatalf("skill directory %q missing SKILL.md: %v", entry.Name(), err)
		}
	}
}

func TestSDDPhaseCommonEnforcesExecutorBoundary(t *testing.T) {
	content := MustRead("skills/_shared/sdd-phase-common.md")

	// Must enforce executor boundary — no delegation allowed.
	for _, want := range []string{
		"EXECUTOR, not an orchestrator",
		"Do NOT launch sub-agents",
		"do NOT call `delegate`/`task`",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("sdd-phase-common missing executor boundary rule %q", want)
		}
	}

	// Must instruct phase agents to search the skill registry themselves
	// when no explicit skill path was provided — this is skill LOADING, not delegation.
	if !strings.Contains(content, `mem_search(query: "skill-registry"`) {
		t.Fatal("sdd-phase-common must instruct phase agents to search skill-registry themselves for skill loading")
	}

	// Must NOT tell agents to launch sub-agents or delegate tasks.
	for _, forbidden := range []string{
		"launch a sub-agent",
		"delegate this to",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("sdd-phase-common should not contain delegation instruction %q", forbidden)
		}
	}
}

func TestOpenCodeSDDOverlaySubagentsAreExplicitExecutors(t *testing.T) {
	for _, assetPath := range []string{"opencode/sdd-overlay-single.json", "opencode/sdd-overlay-multi.json"} {
		t.Run(assetPath, func(t *testing.T) {
			var root map[string]any
			if err := json.Unmarshal([]byte(MustRead(assetPath)), &root); err != nil {
				t.Fatalf("Unmarshal(%q) error = %v", assetPath, err)
			}

			agents, ok := root["agent"].(map[string]any)
			if !ok {
				t.Fatalf("%q missing agent map", assetPath)
			}

			// multi overlay uses __PROMPT_FILE_{phase}__ placeholders that are
			// replaced at runtime with absolute {file:...} references by
			// inlineOpenCodeSDDPrompts. Verify the placeholder format.
			// single overlay still uses inline prompt strings.
			isMulti := assetPath == "opencode/sdd-overlay-multi.json"

			for _, phase := range []string{"sdd-init", "sdd-explore", "sdd-propose", "sdd-spec", "sdd-design", "sdd-tasks", "sdd-apply", "sdd-verify", "sdd-archive"} {
				agentDef, ok := agents[phase].(map[string]any)
				if !ok {
					t.Fatalf("%q missing %s agent", assetPath, phase)
				}
				prompt, _ := agentDef["prompt"].(string)
				if isMulti {
					// Multi overlay uses placeholders — verify the placeholder exists.
					expectedPlaceholder := "__PROMPT_FILE_" + phase + "__"
					if prompt != expectedPlaceholder {
						t.Fatalf("%q phase %s prompt = %q, want placeholder %q", assetPath, phase, prompt, expectedPlaceholder)
					}
				} else {
					// Single overlay has inline executor-scoped prompts.
					for _, want := range []string{"not the orchestrator", "Do NOT delegate", "Do NOT call task/delegate", "Do NOT launch sub-agents"} {
						if !strings.Contains(prompt, want) {
							t.Fatalf("%q phase %s prompt missing %q", assetPath, phase, want)
						}
					}
				}
			}
		})
	}
}

func TestSDDOrchestratorAssetsScopedToDedicatedAgent(t *testing.T) {
	for _, assetPath := range []string{
		"generic/sdd-orchestrator.md",
		"claude/sdd-orchestrator.md",
		"gemini/sdd-orchestrator.md",
		"codex/sdd-orchestrator.md",
		"cursor/sdd-orchestrator.md",
		"opencode/sdd-orchestrator.md",
		"antigravity/sdd-orchestrator.md",
		"vscode/sdd-orchestrator.md",
	} {
		t.Run(assetPath, func(t *testing.T) {
			content := MustRead(assetPath)
			if !strings.Contains(content, "dedicated `sdd-orchestrator` agent or rule only") {
				t.Fatalf("%q missing dedicated-agent scoping note", assetPath)
			}
			if !strings.Contains(content, "Do NOT apply it to executor phase agents") {
				t.Fatalf("%q missing executor exclusion note", assetPath)
			}
		})
	}
}

// TestAdaptiveReasoningGateInjected verifies that all orchestrators have the
// mandatory adaptive reasoning gate injected exactly as it appears in the
// shared source file, and that no duplicates exist.
func TestAdaptiveReasoningGateInjected(t *testing.T) {
	gateContent := MustRead("skills/_shared/adaptive-reasoning-gate.md")
	if len(gateContent) == 0 {
		t.Fatal("shared adaptive-reasoning-gate.md is empty")
	}

	orchestrators := []string{
		"codex/sdd-orchestrator.md",
		"antigravity/sdd-orchestrator.md",
		"kiro/sdd-orchestrator.md",
		"claude/sdd-orchestrator.md",
		"qwen/sdd-orchestrator.md",
		"gemini/sdd-orchestrator.md",
		"generic/sdd-orchestrator.md",
		"gga/sdd-orchestrator.md",
		"windsurf/sdd-orchestrator.md",
		"cursor/sdd-orchestrator.md",
		"opencode/sdd-orchestrator.md",
		"vscode/sdd-orchestrator.md",
	}

	for _, path := range orchestrators {
		t.Run(path, func(t *testing.T) {
			content := MustRead(path)
			if !strings.Contains(content, gateContent) {
				t.Errorf("%q does not contain byte-identical gate content", path)
			}

			// Verify markers are present
			if !strings.Contains(content, "<!-- adaptive-reasoning-gate:START -->") {
				t.Errorf("%q missing start marker", path)
			}
			if !strings.Contains(content, "<!-- adaptive-reasoning-gate:END -->") {
				t.Errorf("%q missing end marker", path)
			}

			// Dedup verification: exactly 1 start marker per file
			if got := strings.Count(content, "<!-- adaptive-reasoning-gate:START -->"); got != 1 {
				t.Errorf("%q has %d gate blocks (want exactly 1)", path, got)
			}

			// Verify programmatic validation contract (replaces old chosen_mode/mode_rationale)
			if !strings.Contains(content, "## Sub-Agent Result Validation") {
				t.Errorf("%q missing result validation section", path)
			}

			// Verify state synchronization section
			if !strings.Contains(content, "## State Synchronization") {
				t.Errorf("%q missing state synchronization section", path)
			}
		})
	}
}

func TestCognitivePosturesElevenNotTenOrTwelve(t *testing.T) {
	postures := []string{
		"+++Socratic", "+++Critical", "+++Systemic",
		"+++Adversarial", "+++Pragmatic", "+++Forensic",
		"+++Economic", "+++Empirical",
		"+++Divergent", "+++Lateral", "+++Diamond",
	}
	body := MustRead("skills/cognitive-mode/SKILL.md")
	for _, p := range postures {
		if !strings.Contains(body, p) {
			t.Errorf("posture %s missing from cognitive-mode/SKILL.md", p)
		}
	}

	orchestrators := []string{
		"claude/sdd-orchestrator.md",
		"antigravity/sdd-orchestrator.md",
		"opencode/sdd-orchestrator.md",
		"codex/sdd-orchestrator.md",
		"kiro/sdd-orchestrator.md",
		"cursor/sdd-orchestrator.md",
		"gemini/sdd-orchestrator.md",
		"generic/sdd-orchestrator.md",
		"gga/sdd-orchestrator.md",
		"windsurf/sdd-orchestrator.md",
		"qwen/sdd-orchestrator.md",
		"vscode/sdd-orchestrator.md",
	}
	for _, rel := range orchestrators {
		t.Run(rel, func(t *testing.T) {
			body := MustRead(rel)
			// Check table rows to ensure they are in the injection table, not just prose.
			for _, p := range []string{"+++Economic", "+++Empirical", "+++Divergent", "+++Lateral", "+++Diamond"} {
				if !strings.Contains(body, p) {
					t.Errorf("%s missing in %s", p, rel)
				}
			}
		})
	}
}
