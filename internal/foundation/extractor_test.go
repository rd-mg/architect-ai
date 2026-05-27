package foundation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- extractCompactRules (unexported) ---

func TestExtractCompactRules_CompactRulesSection(t *testing.T) {
	data := []byte(`---
name: test-skill
---

## Compact Rules

- rule one
- rule two
- rule three

## Another Section

not included
`)
	got := extractCompactRules(data, 20)
	if !strings.Contains(got, "## Compact Rules") {
		t.Errorf("result should contain the heading")
	}
	if !strings.Contains(got, "rule two") {
		t.Errorf("result should contain bullet lines")
	}
	if strings.Contains(got, "Another Section") {
		t.Errorf("result should stop before the next top-level heading")
	}
	if strings.Contains(got, "not included") {
		t.Errorf("result should not include content from later section")
	}
}

func TestExtractCompactRules_QuickReferenceSection(t *testing.T) {
	data := []byte(`---
name: test-skill
---

## Quick Reference

first rule
second rule

## Something Else

tail
`)
	got := extractCompactRules(data, 20)
	if !strings.Contains(got, "## Quick Reference") {
		t.Errorf("result should contain the Quick Reference heading")
	}
	if !strings.Contains(got, "second rule") {
		t.Errorf("result should include content from Quick Reference")
	}
	if strings.Contains(got, "Something Else") {
		t.Errorf("result should stop before the next top-level heading")
	}
}

func TestExtractCompactRules_RespectsMaxLines(t *testing.T) {
	data := []byte(`## Compact Rules
line 1
line 2
line 3
line 4
line 5
`)
	got := extractCompactRules(data, 3)
	lines := strings.Split(got, "\n")
	// Heading + 2 lines = 3 total (maxLines restricts total slice length)
	if len(lines) > 3 {
		t.Errorf("expected at most 3 lines, got %d: %q", len(lines), got)
	}
}

func TestExtractCompactRules_FallbackAfterFrontmatter(t *testing.T) {
	data := []byte(`---
name: test
version: "1"
---
first line after frontmatter
second line
third line
`)
	got := extractCompactRules(data, 2)
	if !strings.Contains(got, "first line after frontmatter") {
		t.Errorf("fallback should include first content line after frontmatter")
	}
	if strings.Contains(got, "third line") {
		t.Errorf("fallback should respect maxLines")
	}
}

func TestExtractCompactRules_EmptyFile(t *testing.T) {
	got := extractCompactRules([]byte{}, 10)
	if got != "" {
		t.Errorf("empty file should produce empty result, got %q", got)
	}
}

func TestExtractCompactRules_NoFrontmatter(t *testing.T) {
	data := []byte(`just a line
another line
`)
	got := extractCompactRules(data, 5)
	if !strings.Contains(got, "just a line") {
		t.Errorf("fallback should include lines even without frontmatter")
	}
}

func TestExtractCompactRules_OnlyFrontmatter(t *testing.T) {
	data := []byte("---\nname: test\n---\n")
	got := extractCompactRules(data, 10)
	// After frontmatter there's only an empty line.
	if got != "" {
		t.Errorf("only frontmatter should give empty result, got %q", got)
	}
}

func TestExtractCompactRules_StopsAtNextHeading(t *testing.T) {
	data := []byte(`## Compact Rules
keep this

## Next Major Section
skip this
`)
	got := extractCompactRules(data, 20)
	if strings.Contains(got, "Next Major Section") {
		t.Errorf("should stop at the next major heading")
	}
}

// --- Extractor ---

func TestExtractor_Extract_Success(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`---
name: my-skill
---

## Compact Rules
- always use rg
- never use grep
`), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	ext := NewExtractor(dir)
	result := ext.Extract(SkillRef{Name: "my-skill", CompactLines: 10})

	if result.Err != nil {
		t.Fatalf("Extract() error = %v", result.Err)
	}
	if result.Name != "my-skill" {
		t.Errorf("Name = %q, want %q", result.Name, "my-skill")
	}
	if !strings.Contains(result.Content, "always use rg") {
		t.Errorf("Content should include extracted rules")
	}
}

func TestExtractor_Extract_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	ext := NewExtractor(dir)

	result := ext.Extract(SkillRef{Name: "nonexistent", CompactLines: 10})

	if result.Err == nil {
		t.Fatal("expected error for missing SKILL.md, got nil")
	}
	if result.Content != "" {
		t.Errorf("expected empty content on error, got %q", result.Content)
	}
}

func TestExtractor_ExtractAll_MultipleSkills(t *testing.T) {
	dir := t.TempDir()

	// skill-a has compact rules
	skillADir := filepath.Join(dir, "skill-a")
	os.MkdirAll(skillADir, 0o755)
	os.WriteFile(filepath.Join(skillADir, "SKILL.md"), []byte("## Compact Rules\n- rule from a\n"), 0o644)

	// skill-b is missing
	// (no dir created)

	// skill-c has compact rules
	skillCDir := filepath.Join(dir, "skill-c")
	os.MkdirAll(skillCDir, 0o755)
	os.WriteFile(filepath.Join(skillCDir, "SKILL.md"), []byte("## Compact Rules\n- rule from c\n"), 0o644)

	ext := NewExtractor(dir)
	skills := []SkillRef{
		{Name: "skill-a", CompactLines: 5},
		{Name: "skill-b", CompactLines: 5},
		{Name: "skill-c", CompactLines: 5},
	}

	results := ext.ExtractAll(skills)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// skill-a: success
	if results[0].Err != nil {
		t.Errorf("skill-a: unexpected error: %v", results[0].Err)
	}
	if !strings.Contains(results[0].Content, "rule from a") {
		t.Errorf("skill-a: missing expected content")
	}

	// skill-b: error (missing)
	if results[1].Err == nil {
		t.Error("skill-b: expected error for missing skill")
	}

	// skill-c: success
	if results[2].Err != nil {
		t.Errorf("skill-c: unexpected error: %v", results[2].Err)
	}
	if !strings.Contains(results[2].Content, "rule from c") {
		t.Errorf("skill-c: missing expected content")
	}
}

// --- Generator ---

func TestGenerator_Generate_IncludesAllResults(t *testing.T) {
	gen := NewGenerator()
	results := []ExtractResult{
		{Name: "ripgrep", Content: "- use rg"},
		{Name: "bash-expert", Content: "- use set -euo"},
	}

	block := gen.Generate(results)

	if !strings.Contains(block.Content, "Project Foundation Standards") {
		t.Error("missing header")
	}
	if !strings.Contains(block.Content, "architect-ai:foundation:start") {
		t.Error("missing start marker")
	}
	if !strings.Contains(block.Content, "architect-ai:foundation:end") {
		t.Error("missing end marker")
	}
	if !strings.Contains(block.Content, "- use rg") {
		t.Error("missing first skill content")
	}
	if !strings.Contains(block.Content, "- use set -euo") {
		t.Error("missing second skill content")
	}
	if len(block.SkillNames) != 2 {
		t.Errorf("SkillNames length = %d, want 2", len(block.SkillNames))
	}
}

func TestGenerator_Generate_EmptyContentGetsPlaceholder(t *testing.T) {
	gen := NewGenerator()
	results := []ExtractResult{
		{Name: "missing-skill", Content: ""},
	}

	block := gen.Generate(results)

	if !strings.Contains(block.Content, "compact rules not available") {
		t.Errorf("empty result should produce placeholder, got: %s", block.Content)
	}
}

func TestFoundationBlock_WriteToFile_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "_generated", "foundation.md")

	block := FoundationBlock{
		Content:    "## Test\n\ncontent\n",
		SkillNames: []string{"test"},
	}

	if err := block.WriteToFile(outPath); err != nil {
		t.Fatalf("WriteToFile() error = %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "## Test\n\ncontent\n" {
		t.Errorf("unexpected content: %q", string(data))
	}
}

func TestFoundationBlock_WriteToFile_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	// Deeply nested path that does not exist yet.
	outPath := filepath.Join(dir, "a", "b", "c", "out.md")

	block := FoundationBlock{
		Content: "hello",
	}

	if err := block.WriteToFile(outPath); err != nil {
		t.Fatalf("WriteToFile() error = %v", err)
	}

	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Error("file should exist after WriteToFile")
	}
}

func TestFoundationBlock_WriteToFile_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "foundation.md")

	block := FoundationBlock{
		Content: "final content",
	}

	if err := block.WriteToFile(outPath); err != nil {
		t.Fatalf("WriteToFile() error = %v", err)
	}

	// The .tmp file should not exist after the atomic rename.
	if _, err := os.Stat(outPath + ".tmp"); !os.IsNotExist(err) {
		t.Error(".tmp file should not exist after atomic rename")
	}
}

// --- Integration: Extractor + Generator round trip ---

func TestExtractAndGenerate_Integration(t *testing.T) {
	dir := t.TempDir()

	// Create skill dirs with SKILL.md files.
	for _, skill := range []struct {
		name    string
		content string
	}{
		{"ripgrep", "## Compact Rules\n- use rg\n- never grep\n"},
		{"bash-expert", "## Compact Rules\n- quote vars\n- set -euo\n"},
	} {
		skillDir := filepath.Join(dir, skill.name)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			t.Fatalf("MkdirAll %s: %v", skill.name, err)
		}
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skill.content), 0o644); err != nil {
			t.Fatalf("WriteFile %s: %v", skill.name, err)
		}
	}

	ext := NewExtractor(dir)
	skills := []SkillRef{
		{Name: "ripgrep", CompactLines: 10},
		{Name: "bash-expert", CompactLines: 10},
	}

	results := ext.ExtractAll(skills)
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("unexpected extraction error for %s: %v", r.Name, r.Err)
		}
	}

	gen := NewGenerator()
	block := gen.Generate(results)

	if !strings.Contains(block.Content, "use rg") {
		t.Error("integration: missing ripgrep rule")
	}
	if !strings.Contains(block.Content, "set -euo") {
		t.Error("integration: missing bash rule")
	}
	if !strings.Contains(block.Content, "architect-ai:foundation:end") {
		t.Error("integration: missing end marker")
	}

	// Write to file and verify.
	outPath := filepath.Join(dir, "_generated", "foundation.md")
	if err := block.WriteToFile(outPath); err != nil {
		t.Fatalf("WriteToFile() error = %v", err)
	}
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Error("integration: foundation.md was not created")
	}
}
