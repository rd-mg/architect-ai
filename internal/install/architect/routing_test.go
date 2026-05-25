package architect

import "testing"

func TestRouteSDD_SlashCommand(t *testing.T) {
	decision, _, _ := ClassifyIntent("/sdd-new auth-feature", TaskContext{})
	if decision != RouteSDD {
		t.Errorf("expected RouteSDD, got %s", decision)
	}
}

func TestRouteSDD_NaturalLanguage(t *testing.T) {
	for _, msg := range []string{
		"use sdd to add authentication",
		"start sdd for the payment flow",
		"haceme un sdd para el login",
	} {
		decision, _, _ := ClassifyIntent(msg, TaskContext{})
		if decision != RouteSDD {
			t.Errorf("expected RouteSDD for %q, got %s", msg, decision)
		}
	}
}

func TestRouteGeneral_AllNonSDDTasks(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		tc   TaskContext
	}{
		{"simple git status", "git status", TaskContext{}},
		{"debug request", "debug the null pointer in auth service", TaskContext{}},
		{"research query", "how does X work", TaskContext{}},
		{"brainstorm", "brainstorm options for Y", TaskContext{}},
		{"refactor", "refactor the auth module", TaskContext{}},
		{"read file", "what's in README.md?", TaskContext{}},
		{"run tests", "run the tests", TaskContext{InvolvesTests: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, reason, mandatory := ClassifyIntent(tt.msg, tt.tc)
			if decision != RouteGeneral {
				t.Errorf("expected RouteGeneral, got %s (%s)", decision, reason)
			}
			// Tests/builds trigger mandatory delegation
			if tt.tc.InvolvesTests || tt.tc.InvolvesBuild {
				if !mandatory {
					t.Error("expected mandatory delegation for tests/builds")
				}
			}
		})
	}
}

func TestMandatoryTriggers_AllBranches(t *testing.T) {
	tests := []struct {
		name string
		tc   TaskContext
		rule string
	}{
		{"4-file rule", TaskContext{FilesReferenced: 4}, "4-file rule"},
		{"multi-file write rule", TaskContext{FilesToWrite: 2}, "multi-file write rule"},
		{"PR rule", TaskContext{IsPRCreation: true}, "PR rule"},
		{"incident rule", TaskContext{IsIncident: true}, "incident rule"},
		{"long-session rule 1", TaskContext{ToolCallCount: 20}, "long-session rule"},
		{"long-session rule 2", TaskContext{ExploratoryReads: 5}, "long-session rule"},
		{"long-session rule 3", TaskContext{NonMechanicalEdits: 2}, "long-session rule"},
		{"execution rule 1", TaskContext{InvolvesTests: true}, "execution rule"},
		{"execution rule 2", TaskContext{InvolvesBuild: true}, "execution rule"},
		{"no rule", TaskContext{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trigger := CheckMandatoryTriggers(tt.tc)
			if tt.rule == "" {
				if trigger.Fired {
					t.Error("expected no trigger to fire")
				}
			} else {
				if !trigger.Fired {
					t.Error("expected trigger to fire")
				}
				if trigger.Rule != tt.rule {
					t.Errorf("expected rule %s, got %s", tt.rule, trigger.Rule)
				}
			}
		})
	}
}

func TestMandatoryTriggerOverridesSDD(t *testing.T) {
	tc := TaskContext{FilesReferenced: 5}
	decision, _, mandatory := ClassifyIntent("/sdd-apply big-change", tc)
	if decision != RouteSDD {
		t.Error("SDD intent should still route to SDD even with mandatory trigger")
	}
	if !mandatory {
		t.Error("should be mandatory delegation")
	}
}

func TestMandatoryTriggerOverridesGeneral(t *testing.T) {
	tc := TaskContext{FilesReferenced: 5}
	decision, _, mandatory := ClassifyIntent("do some general complex stuff", tc)
	if decision != RouteGeneral {
		t.Error("general intent should still route to General even with mandatory trigger")
	}
	if !mandatory {
		t.Error("should be mandatory delegation")
	}
}

func TestModelRouting_Opus(t *testing.T) {
	for _, phase := range []string{"sdd-propose", "sdd-design", "sdd-orchestrator", "architect"} {
		if got := ModelForPhase(phase); got != "opus" {
			t.Errorf("phase %s: expected opus, got %s", phase, got)
		}
	}
}

func TestModelRouting_Haiku(t *testing.T) {
	for _, phase := range []string{"sdd-archive", "sdd-init", "researcher", "generalist", "analyst"} {
		if got := ModelForPhase(phase); got != "haiku" {
			t.Errorf("phase %s: expected haiku, got %s", phase, got)
		}
	}
}

func TestModelRouting_Default(t *testing.T) {
	if got := ModelForPhase("unknown-phase-xyz"); got != "sonnet" {
		t.Errorf("unknown phase should default to sonnet, got %s", got)
	}
}
