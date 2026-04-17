package screens

import (
	"fmt"
	"strings"

	"github.com/rd-mg/architect-ai/internal/components/uninstall"
	"github.com/rd-mg/architect-ai/internal/tui/styles"
)

// RenderPurgeResult shows the post-purge report with what was removed and
// where the snapshot lives. exitOnEnter indicates whether pressing enter
// returns to the welcome screen or exits the TUI.
func RenderPurgeResult(res uninstall.PurgeResult) string {
	var b strings.Builder

	b.WriteString(styles.RenderLogo())
	b.WriteString("\n\n")
	b.WriteString(styles.HeadingStyle.Render("Purge complete"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("Duration: %d ms\n", res.PurgeDurationMs))
	b.WriteString(fmt.Sprintf("Snapshot: %s\n\n", res.SnapshotPath))

	// Removed breakdown
	b.WriteString(styles.HeadingStyle.Render("Removed"))
	b.WriteString("\n\n")

	marks := func(ok bool, label string) string {
		if ok {
			return "  ✓ " + label + "\n"
		}
		return "  ─ " + label + " (skipped)\n"
	}

	if res.ScopeRequested.ManagedConfig {
		b.WriteString(marks(len(res.Result.ChangedFiles)+len(res.Result.RemovedFiles) > 0, "Managed config"))
	}
	if res.ScopeRequested.EngramProject {
		if res.EngramRemoved {
			b.WriteString("  ✓ Engram project memories\n")
		} else {
			b.WriteString(styles.ErrorStyle.Render("  ✗ Engram: " + res.EngramError))
			b.WriteString("\n")
		}
	}
	if res.ScopeRequested.WorkspaceATL {
		b.WriteString(marks(res.ATLRemoved, "Workspace .atl/"))
	}
	if res.ScopeRequested.GlobalArchitectAI {
		b.WriteString(marks(res.GlobalRemoved, "Global ~/.architect-ai/"))
	}
	if res.ScopeRequested.Binary {
		if res.BinaryRemoved {
			b.WriteString(fmt.Sprintf("  ✓ Binary removed via: %s\n", res.BinaryCommandUsed))
		} else {
			b.WriteString(styles.ErrorStyle.Render("  ✗ Binary: " + res.BinaryError))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Manual actions — if any
	if len(res.Result.ManualActions) > 0 {
		b.WriteString(styles.HeadingStyle.Render("Manual follow-up needed"))
		b.WriteString("\n\n")
		for _, a := range res.Result.ManualActions {
			b.WriteString(styles.WarningStyle.Render("  • " + a))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Restore instructions
	b.WriteString(styles.HeadingStyle.Render("To restore"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  architect-ai restore %s\n\n", res.SnapshotPath))

	b.WriteString(styles.HelpStyle.Render("enter: exit • q: quit"))

	return styles.FrameStyle.Render(b.String())
}
