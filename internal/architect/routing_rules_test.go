package architect

import (
	"strings"
	"testing"
)

func TestRoutingRulesMarkdown_ContainsImmutabilityRule(t *testing.T) {
	rules := RoutingRulesMarkdown()
	if !strings.Contains(rules, "IMMUTABILITY RULE") {
		t.Error("RoutingRulesMarkdown() should contain IMMUTABILITY RULE")
	}
	if !strings.Contains(rules, "L0 NEVER executes any tool directly") {
		t.Error("RoutingRulesMarkdown() should state L0 never executes tools directly")
	}
}

func TestRoutingRulesMarkdown_NoModeA(t *testing.T) {
	rules := RoutingRulesMarkdown()
	if strings.Contains(rules, "Mode A (Gemini inline") {
		t.Error("RoutingRulesMarkdown() should NOT contain Mode A inline execution")
	}
	if strings.Contains(rules, "Use bash/read/write tools directly") {
		t.Error("RoutingRulesMarkdown() should NOT contain inline execution instructions")
	}
}

func TestRoutingRulesMarkdown_ContainsModeBAndC(t *testing.T) {
	rules := RoutingRulesMarkdown()
	if !strings.Contains(rules, "Mode B") {
		t.Error("RoutingRulesMarkdown() should contain Mode B")
	}
	if !strings.Contains(rules, "Mode C") {
		t.Error("RoutingRulesMarkdown() should contain Mode C")
	}
	if !strings.Contains(rules, "sdd-orchestrator") {
		t.Error("RoutingRulesMarkdown() should mention sdd-orchestrator")
	}
	if !strings.Contains(rules, "general-orchestrator") {
		t.Error("RoutingRulesMarkdown() should mention general-orchestrator")
	}
}

func TestRoutingRulesMarkdown_ContainsDate(t *testing.T) {
	rules := RoutingRulesMarkdown()
	if !strings.Contains(rules, "Last generated:") {
		t.Error("RoutingRulesMarkdown() should contain generation date")
	}
}
