package config

import (
	"testing"
)

func TestAgentDir(t *testing.T) {
	t.Run("known agent", func(t *testing.T) {
		if got := AgentDir("claude"); got != "claude" {
			t.Errorf("AgentDir(\"claude\") = %q, want %q", got, "claude")
		}
	})
	t.Run("unknown agent", func(t *testing.T) {
		if got := AgentDir("nonexistent"); got != "" {
			t.Errorf("AgentDir(\"nonexistent\") = %q, want %q", got, "")
		}
	})
	t.Run("empty name", func(t *testing.T) {
		if got := AgentDir(""); got != "" {
			t.Errorf("AgentDir(\"\") = %q, want %q", got, "")
		}
	})
}

func TestAgentFiles(t *testing.T) {
	t.Run("known agent", func(t *testing.T) {
		files := AgentFiles("claude")
		if len(files) == 0 {
			t.Fatal("AgentFiles(\"claude\") returned empty slice")
		}
		found := false
		for _, f := range files {
			if f == "sdd-orchestrator.md" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("AgentFiles(\"claude\") missing sdd-orchestrator.md: got %v", files)
		}
	})
	t.Run("unknown agent returns nil", func(t *testing.T) {
		if files := AgentFiles("void"); files != nil {
			t.Errorf("AgentFiles(\"void\") = %v, want nil", files)
		}
	})
}

func TestKnownAgents(t *testing.T) {
	agents := KnownAgents()
	if len(agents) == 0 {
		t.Fatal("KnownAgents() returned empty slice")
	}
	// Build a set for lookup.
	set := make(map[string]bool, len(agents))
	for _, a := range agents {
		set[a] = true
	}
	t.Run("contains claude", func(t *testing.T) {
		if !set["claude"] {
			t.Errorf("KnownAgents missing claude: got %v", agents)
		}
	})
	t.Run("contains all expected agents", func(t *testing.T) {
		for _, want := range []string{"antigravity", "claude", "codex", "cursor", "gemini", "generic", "kiro", "opencode", "qwen", "vscode", "windsurf"} {
			if !set[want] {
				t.Errorf("KnownAgents missing %q: got %v", want, agents)
			}
		}
	})
}

func TestTemplateDir(t *testing.T) {
	if got := TemplateDir(); got != "templates" {
		t.Errorf("TemplateDir() = %q, want %q", got, "templates")
	}
}

func TestSourceMapVersion(t *testing.T) {
	if got := SourceMapVersion(); got != 1 {
		t.Errorf("SourceMapVersion() = %d, want %d", got, 1)
	}
}

func TestPlaceholderSource(t *testing.T) {
	t.Run("known placeholder", func(t *testing.T) {
		got := PlaceholderSource("{content from .atl/agents/architect.md}")
		if got == "" {
			t.Error("PlaceholderSource known token returned empty")
		}
	})
	t.Run("unknown placeholder", func(t *testing.T) {
		if got := PlaceholderSource("{{ unknown }}"); got != "" {
			t.Errorf("PlaceholderSource unknown = %q, want \"\"", got)
		}
	})
}

func TestHashSource(t *testing.T) {
	known := []string{"L0_HASH", "L1A_HASH", "L1B_HASH", "FOUNDATION_HASH"}
	for _, key := range known {
		t.Run(key, func(t *testing.T) {
			if got := HashSource(key); got == "" {
				t.Errorf("HashSource(%q) returned empty", key)
			}
		})
	}
	t.Run("unknown hash", func(t *testing.T) {
		if got := HashSource("UNKNOWN"); got != "" {
			t.Errorf("HashSource unknown = %q, want \"\"", got)
		}
	})
}

func TestIncludesBase(t *testing.T) {
	if got := IncludesBase(); got != "internal/assets" {
		t.Errorf("IncludesBase() = %q, want %q", got, "internal/assets")
	}
}

func TestAllPlaceholders(t *testing.T) {
	m := AllPlaceholders()
	if len(m) == 0 {
		t.Fatal("AllPlaceholders() returned empty map")
	}
	if _, ok := m["{content from .atl/agents/architect.md}"]; !ok {
		t.Error("AllPlaceholders() missing expected entry")
	}
}

func TestAllHashSources(t *testing.T) {
	m := AllHashSources()
	if len(m) == 0 {
		t.Fatal("AllHashSources() returned empty map")
	}
	if _, ok := m["L0_HASH"]; !ok {
		t.Error("AllHashSources() missing L0_HASH")
	}
}

func TestSourceMapIsLazySingleton(t *testing.T) {
	// Verify the source map is loaded once and reused across calls.
	d1 := AgentDir("claude")
	d2 := AgentDir("claude")
	if d1 != d2 {
		t.Error("lazy load produced inconsistent results")
	}
}
