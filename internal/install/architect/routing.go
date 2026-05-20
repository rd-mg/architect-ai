package architect

import (
	"regexp"
	"strings"
)

// RoutingDecision is L0's routing choice per turn
type RoutingDecision string

const (
	RouteInline  RoutingDecision = "inline"  // Mode A
	RouteSDD     RoutingDecision = "sdd"     // Mode B
	RouteGeneral RoutingDecision = "general" // Mode C
)

// ExecutionMode for the session
type ExecutionMode string

const (
	ModeInteractive ExecutionMode = "interactive"
	ModeAutomatic   ExecutionMode = "automatic"
)

// TaskContext describes the complexity of a task for delegation decisions
type TaskContext struct {
	FilesReferenced    int
	FilesToWrite       int
	InvolvesTests      bool
	InvolvesBuild      bool
	IsPRCreation       bool
	IsIncident         bool
	ToolCallCount      int // running count in this session
	ExploratoryReads   int // running count of exploratory reads
	NonMechanicalEdits int // running count of multi-file edits without delegation
}

// MandatoryTrigger represents a delegation trigger result
type MandatoryTrigger struct {
	Fired  bool
	Rule   string
	Reason string
}

// CheckMandatoryTriggers returns the first trigger that fires (if any)
func CheckMandatoryTriggers(tc TaskContext) MandatoryTrigger {
	switch {
	case tc.FilesReferenced >= 4:
		return MandatoryTrigger{true, "4-file rule",
			"understanding requires 4+ files → delegate exploration"}
	case tc.FilesToWrite >= 2:
		return MandatoryTrigger{true, "multi-file write rule",
			"implementation touches 2+ non-trivial files → delegate"}
	case tc.IsPRCreation:
		return MandatoryTrigger{true, "PR rule",
			"before PR → fresh context review required"}
	case tc.IsIncident:
		return MandatoryTrigger{true, "incident rule",
			"after incident → stop + delegate audit"}
	case tc.ToolCallCount >= 20 || tc.ExploratoryReads >= 5 || tc.NonMechanicalEdits >= 2:
		return MandatoryTrigger{true, "long-session rule",
			"session complexity → pause + delegate"}
	case tc.InvolvesTests || tc.InvolvesBuild:
		return MandatoryTrigger{true, "execution rule",
			"tests/builds always delegated"}
	}
	return MandatoryTrigger{Fired: false}
}

// sddPatterns match SDD_INTENT messages
var sddPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)/sdd-(new|continue|ff|init|explore|apply|verify|archive|onboard|hotfix)`),
	regexp.MustCompile(`(?i)\b(use sdd|start sdd|begin sdd|sdd mode|spec-driven)\b`),
	regexp.MustCompile(`(?i)(haceme un sdd|hacer sdd|iniciar sdd|aplicar sdd)`),
}

// simpleTaskKeywords suggest Mode A (inline) execution
var simpleTaskKeywords = []string{
	"git status", "git log", "pwd", "ls ", "cat ",
	"what is", "what does", "explain ", "show me ",
	"rename ", "format this", "what's in",
}

// ClassifyIntent determines the routing decision for a user message.
// Returns (decision, reason, isMandatory).
func ClassifyIntent(message string, tc TaskContext) (RoutingDecision, string, bool) {
	// Step 1: Check mandatory triggers (override everything)
	if trigger := CheckMandatoryTriggers(tc); trigger.Fired {
		// Still need to determine SDD vs General
		if isSDDMessage(message) {
			return RouteSDD, trigger.Reason, true
		}
		return RouteGeneral, trigger.Reason, true
	}

	// Step 2: SDD intent detection
	if isSDDMessage(message) {
		return RouteSDD, "SDD intent detected", false
	}

	// Step 3: Simple task → inline execution (Mode A)
	if isSimpleTask(message, tc) {
		return RouteInline, "simple task — inline execution", false
	}

	// Step 4: Default → delegate to general (Mode C)
	return RouteGeneral, "complex non-SDD task", false
}

func isSDDMessage(message string) bool {
	for _, p := range sddPatterns {
		if p.MatchString(message) {
			return true
		}
	}
	return false
}

func isSimpleTask(message string, tc TaskContext) bool {
	// Must not involve multiple files, tests, or builds
	if tc.FilesReferenced > 3 || tc.FilesToWrite > 1 ||
		tc.InvolvesTests || tc.InvolvesBuild || tc.IsPRCreation {
		return false
	}
	msg := strings.ToLower(message)
	for _, kw := range simpleTaskKeywords {
		if strings.Contains(msg, kw) {
			return true
		}
	}
	return false
}

// ModelForPhase returns the appropriate model alias for a phase/agent
func ModelForPhase(phase string) string {
	routing := map[string]string{
		// Opus: architectural decisions
		"architect":        "opus",
		"orchestrator":     "opus",
		"sdd-orchestrator": "opus",
		"sdd-propose":      "opus",
		"sdd-design":       "opus",
		// Sonnet: implementation + validation
		"sdd-explore":          "sonnet",
		"sdd-spec":             "sonnet",
		"sdd-tasks":            "sonnet",
		"sdd-apply":            "sonnet",
		"sdd-verify":           "sonnet",
		"sdd-onboard":          "sonnet",
		"general-orchestrator": "sonnet",
		"solver":               "sonnet",
		"ideator":              "sonnet",
		// Haiku: mechanical/archival/fast
		"sdd-archive": "haiku",
		"sdd-init":    "haiku",
		"researcher":  "haiku",
		"generalist":  "haiku",
		"analyst":     "haiku",
	}
	if m, ok := routing[phase]; ok {
		return m
	}
	return "sonnet" // safe default
}
