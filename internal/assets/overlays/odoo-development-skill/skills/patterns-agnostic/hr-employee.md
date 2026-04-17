# Hr Employee Patterns

Consolidated from the following source files:
- `hr-employee-patterns.md`
- `project-task-patterns.md`

---


## Source: hr-employee-patterns.md

# HR and Employee Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  HR & EMPLOYEE PATTERNS                                                      ║
║  Employee management, contracts, attendance, and HR workflows                ║
║  Use for HR modules, time tracking, and workforce management                 ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Module Setup

### Manifest Dependencies
```python
{
    'name': 'My HR Module',
    'version': '18.0.1.0.0',
    'depends': ['hr'],  # Core HR
    # Optional: 'hr_contract', 'hr_attendance', 'hr_holidays', 'hr_expense'
    'data': [
        'security/ir.model.access.csv',
        'views/hr_views.xml',
    ],
}
```

---

## Extending Employee Model

### Add Custom Fields
```python
from odoo import api, fields, models
from datetime import date
from dateutil.relativedelta import relativedelta


class HrEmployee(models.Model):
    _inherit = 'hr.employee'

    # Personal Info
    x_emergency_contact = fields.Char(string='Emergency Contact')
    x_emergency_phone = fields.Char(string='Emergency Phone')
    x_blood_type = fields.Selection([
        ('a+', 'A+'), ('a-', 'A-'),
        ('b+', 'B+'), ('b-', 'B-'),
        ('ab+', 'AB+'), ('ab-', 'AB-'),
        ('o+', 'O+'), ('o-', 'O-'),
    ], string='Blood Type')

    # Employment Info
    x_employee_number = fields.Char(
        string='Employee Number',
        copy=False,
        readonly=True,
        default='New',
    )
    x_hire_date = fields.Date(string='Hire Date')
    x_probation_end = fields.Date(
        string='Probation End Date',
        compute='_compute_probation_end',
        store=True,
    )
    x_years_of_service = fields.Float(
        string='Years of Service',
        compute='_compute_years_of_service',
    )
    x_employment_type = fields.Selection([
        ('full_time', 'Full Time'),
        ('part_time', 'Part Time'),
        ('contractor', 'Contractor'),
        ('intern', 'Intern'),
    ], string='Employment Type', default='full_time')

    # Skills & Certifications
    x_skill_ids = fields.Many2many(
        'hr.skill',
        string='Skills',
    )
    x_certification_ids = fields.One2many(
        'hr.employee.certification',
        'employee_id',
        string='Certifications',
    )

    @api.model_create_multi
    def create(self, vals_list):
        for vals in vals_list:
            if vals.get('x_employee_number', 'New') == 'New':
                vals['x_employee_number'] = self.env['ir.sequence'].next_by_code(
                    'hr.employee.number'
                ) or 'New'
        return super().create(vals_list)

    @api.depends('x_hire_date')
    def _compute_probation_end(self):
        for employee in self:
            if employee.x_hire_date:
                employee.x_probation_end = employee.x_hire_date + relativedelta(months=3)
            else:
                employee.x_probation_end = False

    def _compute_years_of_service(self):
        today = date.today()
        for employee in self:
            if employee.x_hire_date:
                delta = relativedelta(today, employee.x_hire_date)
                employee.x_years_of_service = delta.years + (delta.months / 12)
            else:
                employee.x_years_of_service = 0.0
```

### Employee Certification Model
```python
class HrEmployeeCertification(models.Model):
    _name = 'hr.employee.certification'
    _description = 'Employee Certification'

    employee_id = fields.Many2one(
        'hr.employee',
        string='Employee',
        required=True,
        ondelete='cascade',
    )
    name = fields.Char(string='Certification Name', required=True)
    issuing_org = fields.Char(string='Issuing Organization')
    issue_date = fields.Date(string='Issue Date')
    expiry_date = fields.Date(string='Expiry Date')
    certificate_file = fields.Binary(string='Certificate File')
    certificate_filename = fields.Char(string='Filename')
    is_expired = fields.Boolean(
        string='Expired',
        compute='_compute_is_expired',
    )

    def _compute_is_expired(self):
        today = date.today()
        for cert in self:
            cert.is_expired = cert.expiry_date and cert.expiry_date < today
```

---

## Department Extensions

### Custom Department Fields
```python
class HrDepartment(models.Model):
    _inherit = 'hr.department'

    x_budget = fields.Monetary(
        string='Department Budget',
        currency_field='x_currency_id',
    )
    x_currency_id = fields.Many2one(
        'res.currency',
        default=lambda self: self.env.company.currency_id,
    )
    x_cost_center = fields.Char(string='Cost Center')
    x_location_id = fields.Many2one('res.partner', string='Location')

    x_employee_count = fields.Integer(
        string='Employee Count',
        compute='_compute_employee_count',
    )

    def _compute_employee_count(self):
        for dept in self:
            dept.x_employee_count = self.env['hr.employee'].search_count([
                ('department_id', '=', dept.id),
                ('active', '=', True),
            ])
```

---

## Job Positions

### Extend Job Model
```python
class HrJob(models.Model):
    _inherit = 'hr.job'

    x_min_salary = fields.Monetary(
        string='Minimum Salary',
        currency_field='x_currency_id',
    )
    x_max_salary = fields.Monetary(
        string='Maximum Salary',
        currency_field='x_currency_id',
    )
    x_currency_id = fields.Many2one(
        'res.currency',
        default=lambda self: self.env.company.currency_id,
    )
    x_required_skills = fields.Many2many(
        'hr.skill',
        string='Required Skills',
    )
    x_education_level = fields.Selection([
        ('high_school', 'High School'),
        ('bachelor', 'Bachelor\'s Degree'),
        ('master', 'Master\'s Degree'),
        ('phd', 'PhD'),
    ], string='Education Required')
    x_experience_years = fields.Integer(string='Experience Required (Years)')
```

---

## Attendance Integration

### Custom Attendance Logic
```python
class HrAttendance(models.Model):
    _inherit = 'hr.attendance'

    x_location = fields.Char(string='Check-in Location')
    x_device_id = fields.Char(string='Device ID')
    x_is_remote = fields.Boolean(string='Remote Work')
    x_overtime_hours = fields.Float(
        string='Overtime Hours',
        compute='_compute_overtime',
        store=True,
    )

    @api.depends('check_in', 'check_out')
    def _compute_overtime(self):
        for attendance in self:
            if attendance.worked_hours > 8:
                attendance.x_overtime_hours = attendance.worked_hours - 8
            else:
                attendance.x_overtime_hours = 0.0


class HrEmployee(models.Model):
    _inherit = 'hr.employee'

    def action_check_in(self, location=None):
        """Custom check-in with location."""
        self.ensure_one()
        return self.env['hr.attendance'].create({
            'employee_id': self.id,
            'check_in': fields.Datetime.now(),
            'x_location': location,
        })

    def action_check_out(self):
        """Custom check-out."""
        self.ensure_one()
        attendance = self.env['hr.attendance'].search([
            ('employee_id', '=', self.id),
            ('check_out', '=', False),
        ], limit=1)

        if attendance:
            attendance.write({'check_out': fields.Datetime.now()})
            return attendance
        return False
```

---

## Leave Management

### Custom Leave Types
```python
class HrLeaveType(models.Model):
    _inherit = 'hr.leave.type'

    x_requires_approval = fields.Boolean(
        string='Requires Manager Approval',
        default=True,
    )
    x_max_days_per_request = fields.Integer(
        string='Max Days Per Request',
        default=0,
        help='0 = no limit',
    )
    x_requires_attachment = fields.Boolean(
        string='Requires Attachment',
        help='E.g., medical certificate for sick leave',
    )


class HrLeave(models.Model):
    _inherit = 'hr.leave'

    x_attachment_ids = fields.Many2many(
        'ir.attachment',
        string='Attachments',
    )

    @api.constrains('holiday_status_id', 'number_of_days', 'x_attachment_ids')
    def _check_leave_requirements(self):
        for leave in self:
            leave_type = leave.holiday_status_id

            # Check max days
            if leave_type.x_max_days_per_request > 0:
                if leave.number_of_days > leave_type.x_max_days_per_request:
                    raise ValidationError(
                        f"Maximum {leave_type.x_max_days_per_request} days allowed per request."
                    )

            # Check attachment requirement
            if leave_type.x_requires_attachment and not leave.x_attachment_ids:
                raise ValidationError(
                    f"Attachment required for {leave_type.name}."
                )
```

---

## Employee Onboarding

### Onboarding Checklist
```python
class HrOnboardingTask(models.Model):
    _name = 'hr.onboarding.task'
    _description = 'Onboarding Task'
    _order = 'sequence, id'

    name = fields.Char(string='Task', required=True)
    description = fields.Text(string='Description')
    sequence = fields.Integer(default=10)
    department_id = fields.Many2one('hr.department', string='Department')
    responsible_id = fields.Many2one('res.users', string='Responsible')
    days_after_hire = fields.Integer(
        string='Days After Hire',
        help='When task should be completed',
    )
    is_mandatory = fields.Boolean(string='Mandatory', default=True)


class HrEmployeeOnboarding(models.Model):
    _name = 'hr.employee.onboarding'
    _description = 'Employee Onboarding Progress'

    employee_id = fields.Many2one(
        'hr.employee',
        string='Employee',
        required=True,
        ondelete='cascade',
    )
    task_id = fields.Many2one(
        'hr.onboarding.task',
        string='Task',
        required=True,
    )
    state = fields.Selection([
        ('pending', 'Pending'),
        ('in_progress', 'In Progress'),
        ('done', 'Done'),
        ('skipped', 'Skipped'),
    ], string='Status', default='pending')
    completed_date = fields.Date(string='Completed Date')
    completed_by = fields.Many2one('res.users', string='Completed By')
    notes = fields.Text(string='Notes')

    def action_complete(self):
        self.write({
            'state': 'done',
            'completed_date': fields.Date.today(),
            'completed_by': self.env.uid,
        })


class HrEmployee(models.Model):
    _inherit = 'hr.employee'

    x_onboarding_ids = fields.One2many(
        'hr.employee.onboarding',
        'employee_id',
        string='Onboarding Tasks',
    )
    x_onboarding_progress = fields.Float(
        string='Onboarding Progress',
        compute='_compute_onboarding_progress',
    )

    def _compute_onboarding_progress(self):
        for employee in self:
            total = len(employee.x_onboarding_ids)
            done = len(employee.x_onboarding_ids.filtered(
                lambda t: t.state == 'done'
            ))
            employee.x_onboarding_progress = (done / total * 100) if total else 0

    def action_create_onboarding(self):
        """Create onboarding tasks for new employee."""
        self.ensure_one()

        # Get tasks for employee's department or general
        tasks = self.env['hr.onboarding.task'].search([
            '|',
            ('department_id', '=', self.department_id.id),
            ('department_id', '=', False),
        ])

        for task in tasks:
            self.env['hr.employee.onboarding'].create({
                'employee_id': self.id,
                'task_id': task.id,
            })

        return True
```

---

## Views

### Employee Form Extension
```xml
<?xml version="1.0" encoding="utf-8"?>
<odoo>
    <record id="view_employee_form_inherit" model="ir.ui.view">
        <field name="name">hr.employee.form.inherit</field>
        <field name="model">hr.employee</field>
        <field name="inherit_id" ref="hr.view_employee_form"/>
        <field name="arch" type="xml">
            <field name="job_id" position="before">
                <field name="x_employee_number"/>
            </field>

            <xpath expr="//page[@name='public']" position="after">
                <page string="Employment" name="employment">
                    <group>
                        <group>
                            <field name="x_hire_date"/>
                            <field name="x_probation_end"/>
                            <field name="x_years_of_service"/>
                        </group>
                        <group>
                            <field name="x_employment_type"/>
                        </group>
                    </group>
                </page>
                <page string="Emergency" name="emergency">
                    <group>
                        <field name="x_emergency_contact"/>
                        <field name="x_emergency_phone"/>
                        <field name="x_blood_type"/>
                    </group>
                </page>
                <page string="Skills &amp; Certifications" name="skills">
                    <field name="x_skill_ids" widget="many2many_tags"/>
                    <field name="x_certification_ids">
                        <tree editable="bottom">
                            <field name="name"/>
                            <field name="issuing_org"/>
                            <field name="issue_date"/>
                            <field name="expiry_date"/>
                            <field name="is_expired"/>
                        </tree>
                    </field>
                </page>
            </xpath>

            <div name="button_box" position="inside">
                <button class="oe_stat_button" type="object"
                        name="action_view_onboarding"
                        icon="fa-tasks">
                    <field string="Onboarding" name="x_onboarding_progress"
                           widget="percentpie"/>
                </button>
            </div>
        </field>
    </record>
</odoo>
```

---

## Scheduled Actions

### Probation Reminder Cron
```python
@api.model
def _cron_probation_reminder(self):
    """Send reminder for employees ending probation."""
    in_7_days = date.today() + timedelta(days=7)

    employees = self.env['hr.employee'].search([
        ('x_probation_end', '=', in_7_days),
    ])

    template = self.env.ref('my_module.email_template_probation_reminder')
    for employee in employees:
        if employee.parent_id.work_email:
            template.send_mail(employee.id)
```

### Certification Expiry Alert
```python
@api.model
def _cron_certification_expiry(self):
    """Alert for expiring certifications."""
    in_30_days = date.today() + timedelta(days=30)

    expiring = self.env['hr.employee.certification'].search([
        ('expiry_date', '<=', in_30_days),
        ('expiry_date', '>=', date.today()),
    ])

    for cert in expiring:
        cert.employee_id.message_post(
            body=f"Certification '{cert.name}' expires on {cert.expiry_date}",
            message_type='notification',
        )
```

---

## Best Practices

1. **Privacy** - Use `groups="hr.group_hr_user"` for sensitive fields
2. **Employee self-service** - Separate views for employees vs HR
3. **Multi-company** - Filter employees by company
4. **Manager hierarchy** - Use `parent_id` for reporting structure
5. **Document management** - Attach contracts, certificates
6. **Activity scheduling** - Use activities for HR tasks
7. **Audit trail** - Track changes to sensitive data
8. **Integration** - Connect with payroll, expense, timesheet

---


## Source: project-task-patterns.md

# Project and Task Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  PROJECT & TASK PATTERNS                                                     ║
║  Project management, task workflows, and time tracking                       ║
║  Use for project modules, task automation, and resource planning             ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Module Setup

### Manifest Dependencies
```python
{
    'name': 'My Project Module',
    'version': '18.0.1.0.0',
    'depends': ['project'],  # Core project
    # Optional: 'hr_timesheet', 'project_forecast', 'sale_project'
    'data': [
        'security/ir.model.access.csv',
        'views/project_views.xml',
    ],
}
```

---

## Extending Projects

### Add Custom Fields
```python
from odoo import api, fields, models


class ProjectProject(models.Model):
    _inherit = 'project.project'

    x_project_code = fields.Char(string='Project Code')
    x_project_type = fields.Selection([
        ('internal', 'Internal'),
        ('client', 'Client'),
        ('rd', 'R&D'),
    ], string='Project Type', default='client')
    x_budget = fields.Monetary(string='Budget', currency_field='x_currency_id')
    x_currency_id = fields.Many2one(
        'res.currency',
        default=lambda self: self.env.company.currency_id,
    )
    x_start_date = fields.Date(string='Start Date')
    x_end_date = fields.Date(string='End Date')
    x_department_id = fields.Many2one('hr.department', string='Department')
    x_project_manager_id = fields.Many2one(
        'res.users',
        string='Project Manager',
        default=lambda self: self.env.user,
    )

    # Computed fields
    x_progress = fields.Float(
        string='Progress %',
        compute='_compute_progress',
        store=True,
    )
    x_total_hours = fields.Float(
        string='Total Hours',
        compute='_compute_hours',
    )
    x_remaining_budget = fields.Monetary(
        string='Remaining Budget',
        compute='_compute_remaining_budget',
        currency_field='x_currency_id',
    )

    @api.depends('task_ids.stage_id', 'task_ids.x_progress')
    def _compute_progress(self):
        for project in self:
            tasks = project.task_ids.filtered(lambda t: t.active)
            if tasks:
                project.x_progress = sum(tasks.mapped('x_progress')) / len(tasks)
            else:
                project.x_progress = 0.0

    def _compute_hours(self):
        for project in self:
            project.x_total_hours = sum(
                project.task_ids.mapped('effective_hours')
            )

    def _compute_remaining_budget(self):
        for project in self:
            spent = sum(project.task_ids.mapped('x_cost'))
            project.x_remaining_budget = project.x_budget - spent
```

### Project Stages
```python
class ProjectProjectStage(models.Model):
    _name = 'project.project.stage'
    _description = 'Project Stage'
    _order = 'sequence, id'

    name = fields.Char(string='Stage Name', required=True)
    sequence = fields.Integer(default=10)
    fold = fields.Boolean(string='Folded in Kanban')
    description = fields.Text(string='Description')


class ProjectProject(models.Model):
    _inherit = 'project.project'

    x_stage_id = fields.Many2one(
        'project.project.stage',
        string='Stage',
        tracking=True,
        group_expand='_read_group_stage_ids',
    )

    @api.model
    def _read_group_stage_ids(self, stages, domain, order):
        """Show all stages in kanban."""
        return stages.search([], order=order)
```

---

## Extending Tasks

### Add Custom Fields
```python
class ProjectTask(models.Model):
    _inherit = 'project.task'

    x_task_type = fields.Selection([
        ('feature', 'Feature'),
        ('bug', 'Bug Fix'),
        ('improvement', 'Improvement'),
        ('support', 'Support'),
    ], string='Task Type', default='feature')
    x_priority_level = fields.Selection([
        ('0', 'Low'),
        ('1', 'Normal'),
        ('2', 'High'),
        ('3', 'Critical'),
    ], string='Priority Level', default='1')
    x_estimated_hours = fields.Float(string='Estimated Hours')
    x_progress = fields.Float(
        string='Progress %',
        compute='_compute_progress',
        store=True,
    )
    x_cost = fields.Monetary(
        string='Cost',
        compute='_compute_cost',
        currency_field='x_currency_id',
    )
    x_currency_id = fields.Many2one(
        related='project_id.x_currency_id',
    )
    x_reviewer_id = fields.Many2one('res.users', string='Reviewer')
    x_due_warning = fields.Boolean(
        string='Due Warning',
        compute='_compute_due_warning',
    )

    @api.depends('effective_hours', 'x_estimated_hours')
    def _compute_progress(self):
        for task in self:
            if task.x_estimated_hours:
                progress = (task.effective_hours / task.x_estimated_hours) * 100
                task.x_progress = min(progress, 100)
            else:
                task.x_progress = 0.0

    def _compute_cost(self):
        for task in self:
            # Calculate cost from timesheets
            cost = sum(
                ts.unit_amount * ts.employee_id.hourly_cost
                for ts in task.timesheet_ids
            )
            task.x_cost = cost

    @api.depends('date_deadline')
    def _compute_due_warning(self):
        today = fields.Date.today()
        warning_days = 3
        for task in self:
            if task.date_deadline:
                days_until = (task.date_deadline - today).days
                task.x_due_warning = 0 <= days_until <= warning_days
            else:
                task.x_due_warning = False
```

### Task Dependencies
```python
class ProjectTask(models.Model):
    _inherit = 'project.task'

    x_depends_on_ids = fields.Many2many(
        'project.task',
        'project_task_dependency_rel',
        'task_id',
        'depends_on_id',
        string='Depends On',
    )
    x_blocking_ids = fields.Many2many(
        'project.task',
        'project_task_dependency_rel',
        'depends_on_id',
        'task_id',
        string='Blocking',
    )
    x_is_blocked = fields.Boolean(
        string='Is Blocked',
        compute='_compute_is_blocked',
    )

    @api.depends('x_depends_on_ids.stage_id')
    def _compute_is_blocked(self):
        done_stages = self.env['project.task.type'].search([
            ('fold', '=', True),
        ])
        for task in self:
            blocking = task.x_depends_on_ids.filtered(
                lambda t: t.stage_id not in done_stages
            )
            task.x_is_blocked = bool(blocking)

    @api.constrains('x_depends_on_ids')
    def _check_circular_dependency(self):
        for task in self:
            if task in task._get_all_dependencies():
                raise ValidationError("Circular dependency detected!")

    def _get_all_dependencies(self, visited=None):
        """Recursively get all dependencies."""
        if visited is None:
            visited = set()
        dependencies = self.env['project.task']
        for dep in self.x_depends_on_ids:
            if dep.id not in visited:
                visited.add(dep.id)
                dependencies |= dep
                dependencies |= dep._get_all_dependencies(visited)
        return dependencies
```

### Task Checklists
```python
class ProjectTaskChecklist(models.Model):
    _name = 'project.task.checklist'
    _description = 'Task Checklist Item'
    _order = 'sequence, id'

    task_id = fields.Many2one(
        'project.task',
        string='Task',
        required=True,
        ondelete='cascade',
    )
    name = fields.Char(string='Item', required=True)
    sequence = fields.Integer(default=10)
    is_done = fields.Boolean(string='Done')
    done_date = fields.Datetime(string='Done Date')
    done_by = fields.Many2one('res.users', string='Done By')

    def action_toggle_done(self):
        for item in self:
            if item.is_done:
                item.write({
                    'is_done': False,
                    'done_date': False,
                    'done_by': False,
                })
            else:
                item.write({
                    'is_done': True,
                    'done_date': fields.Datetime.now(),
                    'done_by': self.env.uid,
                })


class ProjectTask(models.Model):
    _inherit = 'project.task'

    x_checklist_ids = fields.One2many(
        'project.task.checklist',
        'task_id',
        string='Checklist',
    )
    x_checklist_progress = fields.Float(
        string='Checklist Progress',
        compute='_compute_checklist_progress',
    )

    @api.depends('x_checklist_ids.is_done')
    def _compute_checklist_progress(self):
        for task in self:
            total = len(task.x_checklist_ids)
            done = len(task.x_checklist_ids.filtered('is_done'))
            task.x_checklist_progress = (done / total * 100) if total else 0
```

---

## Task Automation

### Auto-Assignment
```python
class ProjectTask(models.Model):
    _inherit = 'project.task'

    @api.model_create_multi
    def create(self, vals_list):
        tasks = super().create(vals_list)
        for task in tasks:
            if not task.user_ids and task.project_id.x_project_manager_id:
                task.user_ids = task.project_id.x_project_manager_id
        return tasks

    def write(self, vals):
        result = super().write(vals)

        # Auto-assign reviewer when task moves to review stage
        if 'stage_id' in vals:
            review_stage = self.env.ref('my_module.task_stage_review', raise_if_not_found=False)
            if review_stage and self.stage_id == review_stage:
                if not self.x_reviewer_id:
                    self.x_reviewer_id = self.project_id.x_project_manager_id

        return result
```

### Stage Change Notifications
```python
class ProjectTask(models.Model):
    _inherit = 'project.task'

    def write(self, vals):
        old_stages = {task.id: task.stage_id for task in self}
        result = super().write(vals)

        if 'stage_id' in vals:
            for task in self:
                old_stage = old_stages.get(task.id)
                if old_stage != task.stage_id:
                    task._notify_stage_change(old_stage)

        return result

    def _notify_stage_change(self, old_stage):
        """Send notification on stage change."""
        self.message_post(
            body=f"Stage changed from '{old_stage.name}' to '{self.stage_id.name}'",
            message_type='notification',
        )

        # Notify assignees
        for user in self.user_ids:
            self.activity_schedule(
                'mail.mail_activity_data_todo',
                summary=f'Task moved to {self.stage_id.name}',
                user_id=user.id,
            )
```

---

## Views

### Project Form Extension
```xml
<?xml version="1.0" encoding="utf-8"?>
<odoo>
    <record id="view_project_form_inherit" model="ir.ui.view">
        <field name="name">project.project.form.inherit</field>
        <field name="model">project.project</field>
        <field name="inherit_id" ref="project.edit_project"/>
        <field name="arch" type="xml">
            <field name="partner_id" position="after">
                <field name="x_project_code"/>
                <field name="x_project_type"/>
            </field>

            <xpath expr="//page[@name='settings']" position="before">
                <page string="Planning" name="planning">
                    <group>
                        <group>
                            <field name="x_start_date"/>
                            <field name="x_end_date"/>
                            <field name="x_department_id"/>
                        </group>
                        <group>
                            <field name="x_budget"/>
                            <field name="x_remaining_budget"/>
                            <field name="x_progress" widget="progressbar"/>
                        </group>
                    </group>
                </page>
            </xpath>
        </field>
    </record>
</odoo>
```

### Task Form Extension
```xml
<record id="view_task_form_inherit" model="ir.ui.view">
    <field name="name">project.task.form.inherit</field>
    <field name="model">project.task</field>
    <field name="inherit_id" ref="project.view_task_form2"/>
    <field name="arch" type="xml">
        <field name="priority" position="after">
            <field name="x_task_type"/>
            <field name="x_priority_level"/>
        </field>

        <field name="user_ids" position="after">
            <field name="x_reviewer_id"/>
        </field>

        <xpath expr="//page[@name='description_page']" position="after">
            <page string="Planning" name="planning">
                <group>
                    <group>
                        <field name="x_estimated_hours"/>
                        <field name="effective_hours"/>
                        <field name="x_progress" widget="progressbar"/>
                    </group>
                    <group>
                        <field name="x_depends_on_ids" widget="many2many_tags"/>
                        <field name="x_is_blocked"/>
                    </group>
                </group>
            </page>
            <page string="Checklist" name="checklist">
                <field name="x_checklist_ids">
                    <tree editable="bottom">
                        <field name="sequence" widget="handle"/>
                        <field name="name"/>
                        <field name="is_done"/>
                        <field name="done_by" readonly="1"/>
                        <field name="done_date" readonly="1"/>
                    </tree>
                </field>
                <field name="x_checklist_progress" widget="progressbar"/>
            </page>
        </xpath>
    </field>
</record>
```

---

## Scheduled Actions

### Overdue Task Alerts
```python
@api.model
def _cron_check_overdue_tasks(self):
    """Alert on overdue tasks."""
    today = fields.Date.today()
    overdue_tasks = self.env['project.task'].search([
        ('date_deadline', '<', today),
        ('stage_id.fold', '=', False),  # Not done
    ])

    for task in overdue_tasks:
        task.message_post(
            body="This task is overdue!",
            message_type='notification',
            subtype_xmlid='mail.mt_comment',
        )

        # Notify manager
        if task.project_id.x_project_manager_id:
            task.activity_schedule(
                'mail.mail_activity_data_todo',
                summary=f'Overdue task: {task.name}',
                user_id=task.project_id.x_project_manager_id.id,
            )
```

---

## Best Practices

1. **Use stages** for workflow, not custom fields
2. **Track dependencies** for complex projects
3. **Time tracking** - integrate with hr_timesheet
4. **Progress calculation** - automate based on subtasks/checklists
5. **Notifications** - notify on stage changes and deadlines
6. **Budget tracking** - link costs to timesheets
7. **Hierarchy** - use parent tasks for organization
8. **Templates** - create project templates for recurring types
9. **Access rights** - control visibility by project
10. **Reporting** - track velocity, burndown, etc.

---

