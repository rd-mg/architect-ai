package gate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	GateStartMarker = "<!-- adaptive-reasoning-gate:v2:START -->"
	GateEndMarker   = "<!-- adaptive-reasoning-gate:v2:END -->"
	GateV1Marker    = "<!-- adaptive-reasoning-gate:START -->"

	// L2 auto-scoring patterns to remove
	AutoScoringPatternOld = `You MUST state Mode: {n} as the first line`
	AutoScoringPatternAlt = `self-classify your reasoning mode`
	AutoScoringReplacement = `## Adaptive Reasoning (Pre-Injected — READ ONLY)

Your reasoning mode and posture are injected by the orchestrator at the TOP of this
prompt. Find the line: ` + "`[MODE N | D1=X, D2=X, D3=X, D4=X] {rationale}`" + `

RULES:
- Reproduce that header as the FIRST LINE of your response. Verbatim.
- If gate header has unfilled placeholders → emit [GATE_ERROR] + status: blocked.
- If attempt_count >= 2 in the gate → override to MODE 3 + +++Forensic + +++Adversarial.
- DO NOT re-compute D1-D4. DO NOT choose a different mode.`

	// L2 SKILL.md files to patch
	// L2SkillGlob = ".agent/skills/*/SKILL.md" - Handled by paths.Context now.
)

// Target represents a file where the gate should be injected.
type Target struct {
	File          string
	InsertBefore  string
	Required      bool
	ReplaceV1     bool // Has v1 gate — replace with v2
}

// CheckResult holds the result of checking a single target file.
type CheckResult struct {
	File    string
	Present bool
	Version string
}

// InjectResult holds the result of injecting into a single target file.
type InjectResult struct {
	File           string
	AlreadyPresent bool
	Replaced       bool
	Error          error
}

// PurgeResult holds the result of purging (L2 auto-scoring or gate removal).
type PurgeResult struct {
	File     string
	Modified bool
	Error    error
}

// Gate manages the Adaptive Reasoning Gate lifecycle.
type Gate struct {
	sourceFile string
	content    string
}

// New creates a new Gate with the given source file path.
func New(sourceFile string) *Gate {
	return &Gate{sourceFile: sourceFile}
}

func (g *Gate) loadContent() error {
	if g.content != "" {
		return nil
	}
	data, err := os.ReadFile(g.sourceFile)
	if err != nil {
		return fmt.Errorf("cannot read gate source %s: %w", g.sourceFile, err)
	}
	g.content = string(data)
	return nil
}

// Check verifies which target files have v2, v1, or no gate.
func (g *Gate) Check(targets []Target) []CheckResult {
	results := make([]CheckResult, len(targets))
	for i, t := range targets {
		data, err := os.ReadFile(t.File)
		if err != nil {
			results[i] = CheckResult{File: t.File, Present: false, Version: "file-not-found"}
			continue
		}
		content := string(data)
		switch {
		case strings.Contains(content, GateStartMarker):
			results[i] = CheckResult{File: t.File, Present: true, Version: "v2"}
		case strings.Contains(content, GateV1Marker):
			results[i] = CheckResult{File: t.File, Present: true, Version: "v1-outdated"}
		default:
			results[i] = CheckResult{File: t.File, Present: false, Version: "none"}
		}
	}
	return results
}

// Inject inserts the gate into target files.
func (g *Gate) Inject(targets []Target) []InjectResult {
	if err := g.loadContent(); err != nil {
		return []InjectResult{{Error: err}}
	}

	results := make([]InjectResult, len(targets))
	for i, t := range targets {
		data, err := os.ReadFile(t.File)
		if err != nil {
			results[i] = InjectResult{File: t.File, Error: err}
			continue
		}
		content := string(data)

		// Already has v2?
		if strings.Contains(content, GateStartMarker) {
			results[i] = InjectResult{File: t.File, AlreadyPresent: true}
			continue
		}

		// Has v1? Remove it first if ReplaceV1 is set.
		if t.ReplaceV1 && strings.Contains(content, GateV1Marker) {
			content = removeV1Gate(content)
			results[i].Replaced = true
		}

		// Insert gate before InsertBefore marker
		if t.InsertBefore != "" {
			idx := strings.Index(content, t.InsertBefore)
			if idx < 0 {
				results[i] = InjectResult{
					File:  t.File,
					Error: fmt.Errorf("marker '%s' not found in file", t.InsertBefore),
				}
				continue
			}
			content = content[:idx] + g.content + "\n\n" + content[idx:]
		} else {
			// Append at end
			content = content + "\n\n" + g.content
		}

		// Atomic write
		if err := atomicWrite(t.File, content); err != nil {
			results[i] = InjectResult{File: t.File, Error: err}
			continue
		}
		results[i].File = t.File
	}
	return results
}

// Purge removes the gate from target files.
func (g *Gate) Purge(targets []Target) []PurgeResult {
	results := make([]PurgeResult, 0, len(targets))
	for _, t := range targets {
		data, err := os.ReadFile(t.File)
		if err != nil {
			results = append(results, PurgeResult{File: t.File, Error: err})
			continue
		}
		content := string(data)
		if !strings.Contains(content, GateStartMarker) && !strings.Contains(content, GateV1Marker) {
			results = append(results, PurgeResult{File: t.File, Modified: false})
			continue
		}

		// Remove v2 block
		content = removeV2Gate(content)
		// Also remove any remaining v1
		content = removeV1Gate(content)
		// Clean up extra blank lines
		content = strings.TrimSpace(content) + "\n"

		if err := atomicWrite(t.File, content); err != nil {
			results = append(results, PurgeResult{File: t.File, Error: err})
			continue
		}
		results = append(results, PurgeResult{File: t.File, Modified: true})
	}
	return results
}

// PurgeL2AutoScoring removes old auto-scoring patterns from L2 SKILL.md files.
func (g *Gate) PurgeL2AutoScoring(glob string) []PurgeResult {
	matches, _ := filepath.Glob(glob)
	results := make([]PurgeResult, 0, len(matches))

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?m)Adaptive Reasoning gate: You MUST state Mode: \{n\}[^\n]*`),
		regexp.MustCompile(`(?m)self-classify your reasoning mode[^\n]*`),
		regexp.MustCompile(`(?m)chosen_mode[^\n]*`),
	}

	for _, match := range matches {
		data, err := os.ReadFile(match)
		if err != nil {
			results = append(results, PurgeResult{File: match, Error: err})
			continue
		}

		content := string(data)
		original := content
		for _, p := range patterns {
			content = p.ReplaceAllString(content, "")
		}

		// Also replace the Purpose block if it contains the old gate instruction
		if strings.Contains(original, AutoScoringPatternOld) ||
			strings.Contains(original, AutoScoringPatternAlt) {
			// Full block replacement
			content = replaceAutoScoringBlock(content)
		}

		if content == original {
			results = append(results, PurgeResult{File: match, Modified: false})
			continue
		}

		if err := atomicWrite(match, content); err != nil {
			results = append(results, PurgeResult{File: match, Error: err})
			continue
		}
		results = append(results, PurgeResult{File: match, Modified: true})
	}
	return results
}

// EnsureIncludeInTemplates adds {{ include }} directive to template files if missing.
func (g *Gate) EnsureIncludeInTemplates(templates []string) {
	includeLine := `{{ include "_shared/adaptive-reasoning-gate-v2.md" }}`
	for _, tmpl := range templates {
		data, err := os.ReadFile(tmpl)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARN: template %s not found: %v\n", tmpl, err)
			continue
		}
		content := string(data)
		if strings.Contains(content, includeLine) {
			continue
		}
		// Insert before "## Delegation Rules" or at end
		marker := "## Delegation Rules"
		if idx := strings.Index(content, marker); idx >= 0 {
			content = content[:idx] + includeLine + "\n\n" + content[idx:]
		} else {
			content = content + "\n\n" + includeLine + "\n"
		}
		if err := atomicWrite(tmpl, content); err != nil {
			fmt.Fprintf(os.Stderr, "WARN: cannot update template %s: %v\n", tmpl, err)
		}
	}
}

func removeV1Gate(content string) string {
	start := strings.Index(content, GateV1Marker)
	if start < 0 {
		return content
	}
	// Find the next heading after the v1 gate
	end := strings.Index(content[start+len(GateV1Marker):], "## ")
	if end < 0 {
		return content[:start]
	}
	return content[:start] + content[start+len(GateV1Marker)+end:]
}

func removeV2Gate(content string) string {
	start := strings.Index(content, GateStartMarker)
	if start < 0 {
		return content
	}
	end := strings.Index(content[start:], GateEndMarker)
	if end < 0 {
		return content[:start]
	}
	end += start + len(GateEndMarker)
	// Remove up to next newline after end marker
	return content[:start] + content[end:]
}

func replaceAutoScoringBlock(content string) string {
	// Replace the old instruction with the canonical replacement
	if strings.Contains(content, AutoScoringPatternOld) ||
		strings.Contains(content, AutoScoringPatternAlt) {
		// Try to find a "Purpose" section or similar block
		lines := strings.Split(content, "\n")
		var result []string
		for _, line := range lines {
			if strings.Contains(line, AutoScoringPatternOld) ||
				strings.Contains(line, AutoScoringPatternAlt) {
				continue
			}
			result = append(result, line)
		}
		return strings.Join(result, "\n")
	}
	return content
}

func atomicWrite(path, content string) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
