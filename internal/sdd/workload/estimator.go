package workload

import "fmt"

// Task represents a single SDD task with size estimate
type Task struct {
    ID           string
    Description  string
    LinesAdded   int
    LinesDeleted int
    FilesAffected []string
    DependsOn    []string
    ParallelWith []string
}

// BudgetRisk represents the PR size risk level
type BudgetRisk string

const (
    RiskLow    BudgetRisk = "low"
    RiskMedium BudgetRisk = "medium"
    RiskHigh   BudgetRisk = "high"
)

// Forecast is the output of the workload estimator
type Forecast struct {
    EstimatedLinesChanged  int
    BudgetRisk             BudgetRisk
    ChainedPRsRecommended  bool
    DecisionNeededBeforeApply bool
    TasksCount             int
    ParallelTasks          int
    SequentialTasks        int
    TaskDependencyOrder    []TaskDependency
}

// TaskDependency describes ordering constraints between tasks
type TaskDependency struct {
    TaskID       string   `json:"task_id"`
    DependsOn    []string `json:"depends_on"`
    ParallelWith []string `json:"parallel_with"`
}

// Estimate computes the Review Workload Forecast for a set of tasks
func Estimate(tasks []Task) Forecast {
    totalLines := 0
    parallelCount := 0
    sequentialCount := 0
    deps := make([]TaskDependency, 0, len(tasks))

    for _, t := range tasks {
        totalLines += t.LinesAdded + t.LinesDeleted

        if len(t.DependsOn) == 0 {
            parallelCount++
        } else {
            sequentialCount++
        }

        deps = append(deps, TaskDependency{
            TaskID:       t.ID,
            DependsOn:    t.DependsOn,
            ParallelWith: t.ParallelWith,
        })
    }

    risk := classifyRisk(totalLines)

    return Forecast{
        EstimatedLinesChanged:     totalLines,
        BudgetRisk:                risk,
        ChainedPRsRecommended:     totalLines > 400,
        DecisionNeededBeforeApply: risk == RiskMedium || risk == RiskHigh,
        TasksCount:                len(tasks),
        ParallelTasks:             parallelCount,
        SequentialTasks:           sequentialCount,
        TaskDependencyOrder:       deps,
    }
}

func classifyRisk(totalLines int) BudgetRisk {
    switch {
    case totalLines <= 400:
        return RiskLow
    case totalLines <= 800:
        return RiskMedium
    default:
        return RiskHigh
    }
}

// SliceForDelivery returns the first N tasks that fit within the line budget
// Used for auto-chain strategy
func SliceForDelivery(tasks []Task, maxLines int) (slice []Task, remaining []Task) {
    linesSoFar := 0
    for i, t := range tasks {
        taskLines := t.LinesAdded + t.LinesDeleted
        if linesSoFar+taskLines > maxLines && i > 0 {
            // Don't add this task — it would exceed budget
            return tasks[:i], tasks[i:]
        }
        linesSoFar += taskLines
    }
    // All tasks fit
    return tasks, nil
}

// FormatForecastLITE returns a human-readable summary for LITE caveman output
func FormatForecastLITE(f Forecast) string {
    return fmt.Sprintf(
        "Review Workload: ~%d lines changed | Risk: %s | Chained PRs: %v | Tasks: %d (%d parallel, %d sequential)",
        f.EstimatedLinesChanged,
        f.BudgetRisk,
        f.ChainedPRsRecommended,
        f.TasksCount,
        f.ParallelTasks,
        f.SequentialTasks,
    )
}
