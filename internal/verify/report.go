package verify

import (
	"fmt"
	"strings"
)

// Verdict is the top-level readiness classification derived from check results.
type Verdict string

const (
	VerdictReady   Verdict = "READY"   // 0 failures, 0 warnings
	VerdictWarning Verdict = "WARNING" // 0 failures, ≥1 warnings
	VerdictBlocked Verdict = "BLOCKED" // ≥1 failures
)

// Report is the result of running a set of checks.
// The Ready bool field is preserved for backward compatibility with TUI callers.
type Report struct {
	Checks    []CheckResult
	Passed    int
	Failed    int
	Skipped   int
	Warnings  int
	Ready     bool    // true iff Failed == 0 — unchanged from before
	Verdict   Verdict // NEW
	FinalNote string
}

// BuildReport constructs a Report from raw CheckResult slice.
func BuildReport(results []CheckResult) Report {
	report := Report{Checks: append([]CheckResult(nil), results...)}
	for _, r := range results {
		switch r.Status {
		case CheckStatusPassed:
			report.Passed++
		case CheckStatusFailed:
			report.Failed++
		case CheckStatusSkipped:
			report.Skipped++
		case CheckStatusWarning:
			report.Warnings++
		}
	}
	report.Ready = report.Failed == 0
	report.Verdict = deriveVerdict(report.Failed, report.Warnings)
	report.FinalNote = verdictBanner(report.Verdict)
	return report
}

func deriveVerdict(failed, warnings int) Verdict {
	if failed > 0 {
		return VerdictBlocked
	}
	if warnings > 0 {
		return VerdictWarning
	}
	return VerdictReady
}

func verdictBanner(v Verdict) string {
	switch v {
	case VerdictReady:
		return "[ok] READY — run `claude` or `opencode` and start building."
	case VerdictWarning:
		return "[??] WARNING — usable but some components are degraded. Review items above."
	case VerdictBlocked:
		return "[!!] BLOCKED — resolve failed checks before using architect-ai."
	default:
		return ""
	}
}

// RenderReport formats the report as a human-readable string.
func RenderReport(report Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Verification: %d passed, %d failed, %d warnings, %d skipped\n",
		report.Passed, report.Failed, report.Warnings, report.Skipped)

	if len(report.Checks) > 0 {
		b.WriteString("\n")
	}
	for _, check := range report.Checks {
		fmt.Fprintf(&b, "%s %s", statusIcon(check.Status), check.ID)
		if check.Description != "" {
			fmt.Fprintf(&b, " — %s", check.Description)
		}
		b.WriteString("\n")
		if check.Error != "" {
			fmt.Fprintf(&b, "      error: %s\n", check.Error)
		}
		// Show fix only on actionable statuses
		if check.FixHint != "" &&
			(check.Status == CheckStatusFailed || check.Status == CheckStatusWarning) {
			fmt.Fprintf(&b, "      fix:   %s\n", check.FixHint)
		}
	}

	b.WriteString("\n")
	b.WriteString(report.FinalNote)
	if !strings.HasSuffix(report.FinalNote, "\n") {
		b.WriteString("\n")
	}
	return b.String()
}

func statusIcon(s CheckStatus) string {
	switch s {
	case CheckStatusPassed:
		return "[ok]"
	case CheckStatusFailed:
		return "[!!]"
	case CheckStatusWarning:
		return "[??]"
	case CheckStatusSkipped:
		return "[--]"
	default:
		return "[ ]"
	}
}
