package pipeline

import (
	"context"
	"fmt"
)

// OrchestratorOption configures the orchestrator.
type OrchestratorOption func(*Orchestrator)

// WithFailurePolicy sets the failure policy for the apply stage runner.
func WithFailurePolicy(policy FailurePolicy) OrchestratorOption {
	return func(o *Orchestrator) {
		o.runner.FailurePolicy = policy
	}
}

// WithProgressFunc sets a callback that receives progress events during execution.
func WithProgressFunc(fn ProgressFunc) OrchestratorOption {
	return func(o *Orchestrator) {
		o.runner.OnProgress = fn
	}
}

type QualityCheckFunc func(StageResult) bool

type Orchestrator struct {
	runner       Runner
	policy       RollbackPolicy
	stepByID     map[string]Step
	QualityCheck QualityCheckFunc // Returns true if quality threshold is met
}

func NewOrchestrator(policy RollbackPolicy, opts ...OrchestratorOption) *Orchestrator {
	o := &Orchestrator{
		runner:   Runner{},
		policy:   policy,
		stepByID: map[string]Step{},
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

func (o *Orchestrator) Execute(ctx context.Context, plan StagePlan) ExecutionResult {
	o.indexSteps(plan.Prepare)
	o.indexSteps(plan.Apply)

	o.runner.Nonce = plan.Nonce

	prepareResult := o.runner.Run(ctx, StagePrepare, plan.Prepare)
	if !prepareResult.Success {
		return ExecutionResult{Prepare: prepareResult, Err: prepareResult.Err}
	}

	applyResult := o.runner.Run(ctx, StageApply, plan.Apply)
	result := ExecutionResult{Prepare: prepareResult, Apply: applyResult}

	// Recursive Reasoning Gate
	if o.QualityCheck != nil && !o.QualityCheck(applyResult) {
		result.NextRecommendedStage = "design" // Recursive Gate trigger
		result.Err = fmt.Errorf("quality threshold not met in %s stage", StageApply)
		return result
	}

	if applyResult.Success {
		return result
	}

	result.Err = applyResult.Err
	if o.policy.ShouldRollback(StageApply, applyResult.Err) {
		result.Rollback = ExecuteRollback(ctx, applyResult.Steps, o.stepByID)
		if !result.Rollback.Success {
			result.Err = result.Rollback.Err
		}
	}

	return result
}

func (o *Orchestrator) indexSteps(steps []Step) {
	for _, step := range steps {
		o.stepByID[step.ID()] = step
	}
}
