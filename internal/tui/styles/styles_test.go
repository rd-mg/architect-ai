package styles_test

import (
	"strings"
	"testing"

	styles "github.com/rd-mg/architect-ai/internal/tui/styles"
)

func TestTagline(t *testing.T) {
	t.Run("with version", func(t *testing.T) {
		got := styles.Tagline("v3.3.0")
		want := "Architect AI Stack v3.3.0 — One command. Any agent. Any OS."
		if got != want {
			t.Errorf("Tagline() = %q, want %q", got, want)
		}
	})
	t.Run("empty version", func(t *testing.T) {
		got := styles.Tagline("")
		want := "Architect AI Stack  — One command. Any agent. Any OS."
		if got != want {
			t.Errorf("Tagline(\"\") = %q, want %q", got, want)
		}
	})
}

func TestCursor(t *testing.T) {
	if got, want := styles.Cursor, "▸ "; got != want {
		t.Errorf("Cursor = %q, want %q", got, want)
	}
}

func TestRenderLogo(t *testing.T) {
	t.Run("returns non-empty", func(t *testing.T) {
		got := styles.RenderLogo()
		if got == "" {
			t.Error("RenderLogo() returned empty string")
		}
	})
	t.Run("contains braille characters", func(t *testing.T) {
		got := styles.RenderLogo()
		if !strings.Contains(got, "⠀") && !strings.Contains(got, "⣀") && !strings.Contains(got, "⣿") {
			t.Error("RenderLogo() does not contain braille art characters")
		}
	})
	t.Run("multiple lines", func(t *testing.T) {
		got := styles.RenderLogo()
		lines := strings.Split(got, "\n")
		if len(lines) < 5 {
			t.Errorf("RenderLogo() has %d lines, expected at least 5", len(lines))
		}
	})
}

func TestStylesRender(t *testing.T) {
	t.Run("TitleStyle renders", func(t *testing.T) {
		got := styles.TitleStyle.Render("title")
		if got == "" || !strings.Contains(got, "title") {
			t.Errorf("TitleStyle.Render(\"title\") = %q, want non-empty containing \"title\"", got)
		}
	})
	t.Run("ErrorStyle renders", func(t *testing.T) {
		got := styles.ErrorStyle.Render("error")
		if got == "" || !strings.Contains(got, "error") {
			t.Errorf("ErrorStyle.Render(\"error\") = %q, want non-empty containing \"error\"", got)
		}
	})
	t.Run("SuccessStyle renders", func(t *testing.T) {
		got := styles.SuccessStyle.Render("ok")
		if got == "" || !strings.Contains(got, "ok") {
			t.Errorf("SuccessStyle.Render(\"ok\") = %q, want non-empty containing \"ok\"", got)
		}
	})
}

func TestRenderLogoNotPanic(t *testing.T) {
	// Multiple calls should be stable and not panic.
	for i := 0; i < 10; i++ {
		_ = styles.RenderLogo()
	}
}
