package pipeline

import "context"

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

type Orchestrator struct {
	runner   Runner
	policy   RollbackPolicy
	stepByID map[string]Step
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

func (o *Orchestrator) Execute(plan StagePlan) ExecutionResult {
	ctx := context.Background()

	o.indexGroupSteps(plan.Prepare)
	o.indexGroupSteps(plan.Apply)

	prepareResult := o.runner.RunGroups(ctx, StagePrepare, plan.Prepare)
	if !prepareResult.Success {
		return ExecutionResult{Prepare: prepareResult, Err: prepareResult.Err}
	}

	applyResult := o.runner.RunGroups(ctx, StageApply, plan.Apply)
	result := ExecutionResult{Prepare: prepareResult, Apply: applyResult}
	if applyResult.Success {
		return result
	}

	result.Err = applyResult.Err
	if o.policy.ShouldRollback(StageApply, applyResult.Err) {
		result.Rollback = ExecuteRollback(applyResult.Steps, o.stepByID)
		if !result.Rollback.Success {
			result.Err = result.Rollback.Err
		}
	}

	return result
}

func (o *Orchestrator) indexGroupSteps(groups []StepGroup) {
	for _, group := range groups {
		for _, step := range group.Steps {
			o.stepByID[step.ID()] = step
		}
	}
}