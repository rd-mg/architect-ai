package screens

import (
	"strings"

	"github.com/rd-mg/architect-ai/internal/tui/styles"
)

// ConfirmKeyword is the word the user must type to proceed with the purge.
const ConfirmKeyword = "PURGE"

// RenderPurgeConfirm shows the last-chance confirmation screen.
// typedInput is what the user has typed so far; matches (case-sensitive)
// against ConfirmKeyword. errorMsg is shown inline if the user pressed
// enter without a match.
func RenderPurgeConfirm(choices []PurgeChoice, typedInput string, errorMsg string) string {
	var b strings.Builder

	b.WriteString(styles.RenderLogo())
	b.WriteString("\n\n")
	b.WriteString(styles.ErrorStyle.Render("Final Confirmation"))
	b.WriteString("\n\n")

	b.WriteString("You are about to purge the following:\n\n")
	for _, c := range choices {
		if !c.Selected {
			continue
		}
		if c.Dangerous {
			b.WriteString(styles.ErrorStyle.Render("  ✗ " + c.Label))
		} else {
			b.WriteString("  ✗ " + c.Label)
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")

	b.WriteString(styles.WarningStyle.Render("This is IRREVERSIBLE beyond the pre-purge snapshot."))
	b.WriteString("\n\n")

	b.WriteString("Type ")
	b.WriteString(styles.ErrorStyle.Render(ConfirmKeyword))
	b.WriteString(" to confirm (case-sensitive):\n\n")

	// Show typed input with a cursor caret
	inputDisplay := typedInput + "▌"
	b.WriteString("  > " + inputDisplay)
	b.WriteString("\n")

	if errorMsg != "" {
		b.WriteString("\n")
		b.WriteString(styles.ErrorStyle.Render("  ✗ " + errorMsg))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("enter: confirm • esc: cancel • backspace: edit"))

	return styles.FrameStyle.Render(b.String())
}

// IsConfirmed returns true when the typed input exactly matches ConfirmKeyword.
// Case-sensitive, trimmed — "purge" or " PURGE " do NOT pass.
func IsConfirmed(typedInput string) bool {
	return strings.TrimSpace(typedInput) == ConfirmKeyword && !strings.Contains(typedInput, " ")
}
