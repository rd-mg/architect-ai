package agent

import (
	"errors"
	"fmt"
	"strings"
)

// CrossCallMatrix represents the cross-agent calling permissions.
type CrossCallMatrix struct {
	allowed map[string]map[string]bool
}

// NewCrossCallMatrix creates a new CrossCallMatrix initialized with the calling allowed matrix.
func NewCrossCallMatrix() *CrossCallMatrix {
	allowed := map[string]map[string]bool{
		"architect": {
			"sdd-orchestrator":    true,
			"general-orchestrator": true,
		},
		"sdd-orchestrator": {
			"sdd-explore": true,
			"sdd-apply":   true,
			"sdd-verify":  true,
			"sdd-propose": true,
			"sdd-spec":    true,
			"sdd-tasks":   true,
			"sdd-archive": true,
			"sdd-onboard": true,
		},
		"general-orchestrator": {
			"researcher": true,
			"solver":     true,
			"ideator":    true,
			"generalist": true,
			"analyst":    true,
		},
		"researcher": {
			"context7":   true,
			"notebooklm": true,
			"engram":     true,
			"web":        true,
		},
		"solver": {
			"researcher": true,
			"generalist": true,
			"odoo-expert": true,
			"odoo-skill-finder": true,
			"odoo-database-query": true,
			"odoo-code-reviewer": true,
			"odoo-upgrade-analyzer": true,
			"odoo-spreadsheet-dashboard-architect": true,
		},
		"ideator": {
			"researcher": true,
			"generalist": true,
		},
		"generalist": {
			"researcher": true,
			"odoo-expert": true,
			"odoo-skill-finder": true,
			"odoo-database-query": true,
			"odoo-code-reviewer": true,
			"odoo-upgrade-analyzer": true,
			"odoo-spreadsheet-dashboard-architect": true,
		},
	}

	return &CrossCallMatrix{
		allowed: allowed,
	}
}

// normalizeName converts names to lowercase and strips tier prefixes (e.g. "l0 ", "l2 ").
func normalizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.TrimPrefix(name, "l0 ")
	name = strings.TrimPrefix(name, "l1a ")
	name = strings.TrimPrefix(name, "l1b ")
	name = strings.TrimPrefix(name, "l2 ")
	name = strings.TrimPrefix(name, "l3 ")
	return strings.TrimSpace(name)
}

// CanCall checks if the caller is allowed to call the callee.
func (m *CrossCallMatrix) CanCall(caller, callee string) bool {
	normCaller := normalizeName(caller)
	normCallee := normalizeName(callee)

	// L2 SDD phases can call researcher and solver
	if strings.HasPrefix(normCaller, "sdd-") && normCaller != "sdd-orchestrator" {
		if normCallee == "researcher" || normCallee == "solver" {
			return true
		}
	}

	// L3 Odoo agents can call Engram, rg, bash
	if strings.HasPrefix(normCaller, "odoo-") {
		if normCallee == "engram" || normCallee == "rg" || normCallee == "bash" {
			return true
		}
	}

	// Direct lookup
	if callees, exists := m.allowed[normCaller]; exists {
		return callees[normCallee]
	}

	return false
}

// ValidateReturn validates if the given return map complies with the return contract.
func (m *CrossCallMatrix) ValidateReturn(result map[string]interface{}) error {
	// 1. Validate 'status'
	statusVal, exists := result["status"]
	if !exists {
		return errors.New("missing status field")
	}
	status, ok := statusVal.(string)
	if !ok {
		return errors.New("status field must be a string")
	}
	validStatuses := map[string]bool{
		"COMPLETE": true,
		"PARTIAL":  true,
		"FAILED":   true,
		"BLOCKED":  true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status value: %q", status)
	}

	// 2. Validate 'result'
	resultVal, exists := result["result"]
	if !exists {
		return errors.New("missing result field")
	}
	_, ok = resultVal.(string)
	if !ok {
		return errors.New("result field must be a string")
	}

	// 3. Validate 'source'
	sourceVal, exists := result["source"]
	if !exists {
		return errors.New("missing source field")
	}
	source, ok := sourceVal.(string)
	if !ok {
		return errors.New("source field must be a string")
	}
	validSources := map[string]bool{
		"engram":   true,
		"local":    true,
		"context7": true,
		"web":      true,
	}
	if !validSources[source] {
		return fmt.Errorf("invalid source value: %q", source)
	}

	// 4. Validate 'confidence'
	confidenceVal, exists := result["confidence"]
	if !exists {
		return errors.New("missing confidence field")
	}
	confidence, ok := confidenceVal.(string)
	if !ok {
		return errors.New("confidence field must be a string")
	}
	validConfidences := map[string]bool{
		"high":   true,
		"medium": true,
		"low":    true,
	}
	if !validConfidences[confidence] {
		return fmt.Errorf("invalid confidence value: %q", confidence)
	}

	// 5. Validate 'reason_if_failed'
	reasonVal, exists := result["reason_if_failed"]
	if !exists {
		return errors.New("missing reason_if_failed field")
	}

	// 6. If FAILED or BLOCKED, reason_if_failed must be non-empty string
	if status == "FAILED" || status == "BLOCKED" {
		if reasonVal == nil {
			return fmt.Errorf("reason_if_failed cannot be nil when status is %q", status)
		}
		reasonStr, ok := reasonVal.(string)
		if !ok {
			return fmt.Errorf("reason_if_failed must be a string when status is %q", status)
		}
		if strings.TrimSpace(reasonStr) == "" {
			return fmt.Errorf("reason_if_failed cannot be empty when status is %q", status)
		}
	} else {
		// COMPLETE or PARTIAL: reason_if_failed can be nil or a string (possibly empty)
		if reasonVal != nil {
			if _, ok := reasonVal.(string); !ok {
				return errors.New("reason_if_failed must be a string or nil")
			}
		}
	}

	return nil
}
