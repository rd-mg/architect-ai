package agent

import (
	"testing"
)

func TestCrossCall_AllowedPairs(t *testing.T) {
	matrix := NewCrossCallMatrix()

	tests := []struct {
		caller   string
		callee   string
		expected bool
	}{
		{"solver", "researcher", true},
		{"solver", "ideator", false},
		{"architect", "sdd-orchestrator", true},
		{"sdd-explore", "solver", true},
		{"researcher", "solver", false},
		{"general-orchestrator", "solver", true},
		{"general-orchestrator", "sdd-orchestrator", false},
		{"solver", "odoo-expert", true},
		{"ideator", "generalist", true},
		{"generalist", "solver", false},
	}

	for _, tt := range tests {
		t.Run(tt.caller+"_calls_"+tt.callee, func(t *testing.T) {
			got := matrix.CanCall(tt.caller, tt.callee)
			if got != tt.expected {
				t.Errorf("CanCall(%q, %q) = %v; want %v", tt.caller, tt.callee, got, tt.expected)
			}
		})
	}
}

func TestCrossCall_NoLoops(t *testing.T) {
	matrix := NewCrossCallMatrix()

	// A list of some known allowed pairs
	allowedPairs := []struct {
		caller string
		callee string
	}{
		{"architect", "sdd-orchestrator"},
		{"architect", "general-orchestrator"},
		{"sdd-orchestrator", "sdd-explore"},
		{"sdd-orchestrator", "sdd-apply"},
		{"general-orchestrator", "researcher"},
		{"general-orchestrator", "solver"},
		{"general-orchestrator", "ideator"},
		{"general-orchestrator", "generalist"},
		{"sdd-explore", "researcher"},
		{"sdd-explore", "solver"},
		{"solver", "researcher"},
		{"solver", "generalist"},
		{"solver", "odoo-expert"},
		{"ideator", "researcher"},
		{"ideator", "generalist"},
		{"generalist", "researcher"},
		{"generalist", "odoo-expert"},
	}

	for _, pair := range allowedPairs {
		t.Run(pair.caller+"_to_"+pair.callee+"_is_acyclic", func(t *testing.T) {
			// If A can call B, then B must NOT be able to call A
			if matrix.CanCall(pair.caller, pair.callee) {
				if matrix.CanCall(pair.callee, pair.caller) {
					t.Errorf("Loop detected: %s can call %s AND %s can call %s", pair.caller, pair.callee, pair.callee, pair.caller)
				}
			} else {
				t.Skipf("Skipping loop check since %s calling %s is not currently allowed in matrix", pair.caller, pair.callee)
			}
		})
	}
}

func TestCrossCall_ReturnContract(t *testing.T) {
	matrix := NewCrossCallMatrix()

	tests := []struct {
		name    string
		input   map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid complete contract",
			input: map[string]interface{}{
				"status":           "COMPLETE",
				"result":           "success data",
				"source":           "local",
				"confidence":       "high",
				"reason_if_failed": nil,
			},
			wantErr: false,
		},
		{
			name: "valid failed contract with reason",
			input: map[string]interface{}{
				"status":           "FAILED",
				"result":           "",
				"source":           "engram",
				"confidence":       "low",
				"reason_if_failed": "connection timeout",
			},
			wantErr: false,
		},
		{
			name: "missing status",
			input: map[string]interface{}{
				"result":           "data",
				"source":           "local",
				"confidence":       "high",
				"reason_if_failed": nil,
			},
			wantErr: true,
		},
		{
			name: "invalid status value",
			input: map[string]interface{}{
				"status":           "INVALID_STATUS",
				"result":           "data",
				"source":           "local",
				"confidence":       "high",
				"reason_if_failed": nil,
			},
			wantErr: true,
		},
		{
			name: "missing result",
			input: map[string]interface{}{
				"status":           "COMPLETE",
				"source":           "local",
				"confidence":       "high",
				"reason_if_failed": nil,
			},
			wantErr: true,
		},
		{
			name: "invalid source value",
			input: map[string]interface{}{
				"status":           "COMPLETE",
				"result":           "data",
				"source":           "invalid_source",
				"confidence":       "high",
				"reason_if_failed": nil,
			},
			wantErr: true,
		},
		{
			name: "invalid confidence value",
			input: map[string]interface{}{
				"status":           "COMPLETE",
				"result":           "data",
				"source":           "local",
				"confidence":       "very high",
				"reason_if_failed": nil,
			},
			wantErr: true,
		},
		{
			name: "failed status without reason",
			input: map[string]interface{}{
				"status":           "FAILED",
				"result":           "",
				"source":           "local",
				"confidence":       "medium",
				"reason_if_failed": nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := matrix.ValidateReturn(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateReturn() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
