package workload

import "testing"

func TestEstimate_LowRisk(t *testing.T) {
    tasks := []Task{
        {ID: "T1", LinesAdded: 50, LinesDeleted: 10},
        {ID: "T2", LinesAdded: 80, LinesDeleted: 20},
    }
    f := Estimate(tasks)
    if f.EstimatedLinesChanged != 160 {
        t.Errorf("expected 160, got %d", f.EstimatedLinesChanged)
    }
    if f.BudgetRisk != RiskLow {
        t.Errorf("expected low risk, got %s", f.BudgetRisk)
    }
    if f.ChainedPRsRecommended {
        t.Error("160 lines should not recommend chained PRs")
    }
}

func TestEstimate_MediumRisk(t *testing.T) {
    tasks := []Task{
        {ID: "T1", LinesAdded: 300, LinesDeleted: 150},
    }
    f := Estimate(tasks)
    if f.BudgetRisk != RiskMedium {
        t.Errorf("expected medium risk for 450 lines, got %s", f.BudgetRisk)
    }
    if !f.ChainedPRsRecommended {
        t.Error("450 lines should recommend chained PRs")
    }
    if !f.DecisionNeededBeforeApply {
        t.Error("medium risk should require decision before apply")
    }
}

func TestEstimate_HighRisk(t *testing.T) {
    tasks := []Task{
        {ID: "T1", LinesAdded: 500, LinesDeleted: 400},
    }
    f := Estimate(tasks)
    if f.BudgetRisk != RiskHigh {
        t.Errorf("expected high risk for 900 lines, got %s", f.BudgetRisk)
    }
}

func TestEstimate_DependencyOrdering(t *testing.T) {
    tasks := []Task{
        {ID: "T1", DependsOn: []string{}, ParallelWith: []string{"T2"}},
        {ID: "T2", DependsOn: []string{}, ParallelWith: []string{"T1"}},
        {ID: "T3", DependsOn: []string{"T1", "T2"}},
    }
    f := Estimate(tasks)
    if f.ParallelTasks != 2 {
        t.Errorf("expected 2 parallel tasks, got %d", f.ParallelTasks)
    }
    if f.SequentialTasks != 1 {
        t.Errorf("expected 1 sequential task, got %d", f.SequentialTasks)
    }
}

func TestSliceForDelivery_SplitsCorrectly(t *testing.T) {
    tasks := []Task{
        {ID: "T1", LinesAdded: 200, LinesDeleted: 0},  // 200 lines
        {ID: "T2", LinesAdded: 150, LinesDeleted: 50}, // 200 lines → total 400 at limit
        {ID: "T3", LinesAdded: 100, LinesDeleted: 0},  // 100 lines → would exceed 400
    }
    slice, remaining := SliceForDelivery(tasks, 400)
    if len(slice) != 2 {
        t.Errorf("expected 2 tasks in first slice, got %d", len(slice))
    }
    if len(remaining) != 1 {
        t.Errorf("expected 1 remaining task, got %d", len(remaining))
    }
    if remaining[0].ID != "T3" {
        t.Errorf("expected T3 as remaining, got %s", remaining[0].ID)
    }
}

func TestSliceForDelivery_AllFit(t *testing.T) {
    tasks := []Task{
        {ID: "T1", LinesAdded: 50},
        {ID: "T2", LinesAdded: 50},
    }
    slice, remaining := SliceForDelivery(tasks, 400)
    if len(slice) != 2 {
        t.Errorf("expected all 2 tasks, got %d", len(slice))
    }
    if len(remaining) != 0 {
        t.Error("expected no remaining tasks")
    }
}

func TestFormatForecastLITE(t *testing.T) {
    f := Forecast{
        EstimatedLinesChanged: 450,
        BudgetRisk:            RiskMedium,
        ChainedPRsRecommended: true,
        TasksCount:            5,
        ParallelTasks:         3,
        SequentialTasks:       2,
    }
    result := FormatForecastLITE(f)
    if result == "" {
        t.Error("format should not return empty string")
    }
}
