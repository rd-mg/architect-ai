# HR & Project Patterns

Consolidated from the following source files:
- `hr-employee-patterns.md` (architect-ai)
- `project-task-patterns.md` (architect-ai)

> **Version-specific syntax** → `patterns-{version}/model-patterns.md`
> `_check_company_auto` in v18+ · `tracking=True` mandatory for chatter

---

## HR & Workforce

### Employee Onboarding Flow
```python
class HrEmployee(models.Model):
    _inherit = 'hr.employee'

    onboarding_progress = fields.Float(compute='_compute_onboarding_progress')

    def action_start_onboarding(self):
        # Create standard tasks for new hires
        tasks = self.env['hr.onboarding.task'].search([])
        for task in tasks:
            self.env['hr.employee.onboarding'].create({
                'employee_id': self.id,
                'task_id': task.id,
            })
```

### Attendance & Overtime
```python
class HrAttendance(models.Model):
    _inherit = 'hr.attendance'

    overtime_hours = fields.Float(compute='_compute_overtime', store=True)

    @api.depends('check_in', 'check_out')
    def _compute_overtime(self):
        for rec in self:
            if rec.worked_hours > 8:
                rec.overtime_hours = rec.worked_hours - 8
```

---

## Project & Task Management

### Task Progress & Costing
```python
class ProjectTask(models.Model):
    _inherit = 'project.task'

    progress = fields.Float(compute='_compute_progress', store=True)
    task_cost = fields.Monetary(compute='_compute_cost', currency_field='currency_id')

    @api.depends('effective_hours', 'planned_hours')
    def _compute_progress(self):
        for task in self:
            task.progress = (task.effective_hours / task.planned_hours) * 100 if task.planned_hours else 0
```

---

## Anti-Patterns

```python
# ❌ NEVER update hr.employee active status without checking related user status.

# ❌ NEVER use float_time widget for durations longer than 24h without custom logic.

# ❌ NEVER bypass the project.task state machine when integrating with Sales.
```

---

## Version Matrix

| Feature | v14-v16 | v17 | v18 | v19 |
|---------|---------|-----|-----|-----|
| Employee | `hr.employee` | `hr.employee` | `hr.employee` | `hr.employee` |
| Project | `project.project`| `project.project`| `project.project`| `project.project`|
| Timesheets | `hr_timesheet` | `hr_timesheet` | `hr_timesheet` | `hr_timesheet` |
| Tracking | `tracking` | `tracking` | `tracking` | `tracking` |
