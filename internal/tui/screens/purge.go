package screens

import (
	"strings"

	"github.com/rd-mg/architect-ai/internal/tui/styles"
)

// PurgeScopeChoice represents a single toggleable purge category in the TUI.
// Ordered for display top-to-bottom in the checkbox list.
type PurgeScopeChoice int

const (
	PurgeChoiceManagedConfig PurgeScopeChoice = iota
	PurgeChoiceEngramProject
	PurgeChoiceWorkspaceATL
	PurgeChoiceGlobalDir
	PurgeChoiceBinary
)

// PurgeChoice represents one row in the purge-scope screen.
type PurgeChoice struct {
	Kind        PurgeScopeChoice
	Label       string
	Description string
	Selected    bool
	Dangerous   bool // shows ⚠ glyph and red color
}

// DefaultPurgeChoices returns the default checkbox selections.
// Conservative defaults: pre-select managed config + Engram + .atl/ (safe
// project-level cleanup). Do NOT pre-select global dir or binary.
func DefaultPurgeChoices() []PurgeChoice {
	return []PurgeChoice{
		{
			Kind:        PurgeChoiceManagedConfig,
			Label:       "Managed config (all installed agents)",
			Description: "Removes whatever Managed uninstall already removes.",
			Selected:    true,
		},
		{
			Kind:        PurgeChoiceEngramProject,
			Label:       "Engram project memories",
			Description: "Deletes all memories for this project. Other projects untouched.",
			Selected:    true,
		},
		{
			Kind:        PurgeChoiceWorkspaceATL,
			Label:       "Workspace .atl/ directory",
			Description: "Removes the skill registry and overlay manifests in this repo.",
			Selected:    true,
		},
		{
			Kind:        PurgeChoiceGlobalDir,
			Label:       "Global ~/.architect-ai/ (backups, state)",
			Description: "Removes ALL backups and global state. Affects every project.",
			Selected:    false,
			Dangerous:   true,
		},
		{
			Kind:        PurgeChoiceBinary,
			Label:       "Uninstall the architect-ai binary",
			Description: "Runs brew uninstall / apt remove / pacman -R if applicable.",
			Selected:    false,
			Dangerous:   true,
		},
	}
}

// RenderPurge renders the checkbox list for selecting purge scope.
// cursor is the row under the highlighted selector.
func RenderPurge(choices []PurgeChoice, cursor int) string {
	var b strings.Builder

	b.WriteString(styles.RenderLogo())
	b.WriteString("\n\n")
	b.WriteString(styles.HeadingStyle.Render("Uninstall & Purge All"))
	b.WriteString("\n\n")

	b.WriteString(styles.WarningStyle.Render("⚠  This is IRREVERSIBLE beyond the pre-purge snapshot."))
	b.WriteString("\n")
	b.WriteString(styles.SubtextStyle.Render("   A snapshot is captured automatically before any removal."))
	b.WriteString("\n\n")

	b.WriteString(styles.HeadingStyle.Render("Select what to purge"))
	b.WriteString("\n\n")

	for i, c := range choices {
		cursorMark := "  "
		if i == cursor {
			cursorMark = "▸ "
		}
		check := "[ ]"
		if c.Selected {
			check = "[x]"
		}

		line := cursorMark + check + " " + c.Label
		if c.Dangerous {
			line += "  ⚠"
			b.WriteString(styles.ErrorStyle.Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
		b.WriteString(styles.SubtextStyle.Render("     " + c.Description))
		b.WriteString("\n\n")
	}

	b.WriteString(styles.HelpStyle.Render("j/k: navigate • space: toggle • enter: continue • esc: cancel"))

	return styles.FrameStyle.Render(b.String())
}
