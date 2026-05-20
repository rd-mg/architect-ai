package architect

import "testing"

func TestRouteInline_SimpleGitStatus(t *testing.T) {
	decision, _, mandatory := ClassifyIntent("git status", TaskContext{})
	if decision != RouteInline {
		t.Errorf("expected RouteInline, got %s", decision)
	}
	if mandatory {
		t.Error("git status should not trigger mandatory delegation")
	}
}

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

func TestRouteGeneral_ComplexTask(t *testing.T) {
	decision, _, _ := ClassifyIntent("debug the null pointer in auth service", TaskContext{
		FilesReferenced: 2,
		FilesToWrite:    1,
	})
	if decision != RouteGeneral {
		t.Errorf("expected RouteGeneral, got %s", decision)
	}
}

func TestMandatory4FileRule(t *testing.T) {
	tc := TaskContext{FilesReferenced: 5}
	trigger := CheckMandatoryTriggers(tc)
	if !trigger.Fired {
		t.Error("4-file rule should fire for 5 files referenced")
	}
	if trigger.Rule != "4-file rule" {
		t.Errorf("wrong rule: %s", trigger.Rule)
	}
}

func TestMandatoryLongSessionRule(t *testing.T) {
	tc := TaskContext{ToolCallCount: 21}
	trigger := CheckMandatoryTriggers(tc)
	if !trigger.Fired {
		t.Error("long-session rule should fire at 21 tool calls")
	}
}

func TestMandatoryTriggerOverridesSDD(t *testing.T) {
	// Even an SDD message with 5 files referenced → mandatory delegation
	tc := TaskContext{FilesReferenced: 5}
	decision, _, mandatory := ClassifyIntent("/sdd-apply big-change", tc)
	if decision != RouteSDD {
		t.Error("SDD intent should still route to SDD even with mandatory trigger")
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
	// Unknown phase → safe default sonnet
	if got := ModelForPhase("unknown-agent"); got != "sonnet" {
		t.Errorf("expected sonnet default, got %s", got)
	}
}

func TestClassifyIntent_MandatoryTests(t *testing.T) {
	tc := TaskContext{InvolvesTests: true}
	_, _, mandatory := ClassifyIntent("run the auth tests", tc)
	if !mandatory {
		t.Error("InvolvesTests=true should trigger mandatory delegation")
	}
}

func TestClassifyIntent_MandatoryPR(t *testing.T) {
	tc := TaskContext{IsPRCreation: true}
	_, _, mandatory := ClassifyIntent("create a PR for this change", tc)
	if !mandatory {
		t.Error("IsPRCreation=true should trigger mandatory delegation")
	}
}
