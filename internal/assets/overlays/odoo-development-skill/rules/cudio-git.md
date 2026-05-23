---
name: cudio-git
project_scope: cudio
---

# Cudio Git & Branch Convention

Organization-specific Git workflow for Cudio projects. Extends `branch-pr` skill.

## Branch Structure

Cudio repos follow:
- **1 Production branch** — stable, production-ready
- **Multiple Staging branches** — testing/validation
- **Multiple Development branches** — active work

## Branch Naming

Format: `{prefix}/{task-id}-{brief-description}`

### Prefix Table

| Prefix | Use Case |
|--------|----------|
| `feat/` | New features or enhancements |
| `fix/` | Bug fixes |
| `refactor/` | Code refactoring without functionality changes |
| `docs/` | Documentation updates |
| `test/` | Test-related changes |
| `chore/` | Maintenance tasks and build changes |

### Examples

- `feat/13123-new-invoice-report`
- `fix/14567-payment-validation-error`
- `refactor/15890-optimize-inventory-queries`
- `docs/16001-update-readme`
- `test/17234-add-approval-tests`
- `chore/18045-bump-dependencies`

### Requirements

- Always include task/ticket ID when available
- Use kebab-case for description
- Keep descriptions concise but descriptive
- Create branches from appropriate source (typically `staging-dev`)

### Validation Regex
```
^(feat|fix|refactor|docs|test|chore)/\d+-[a-z0-9]+(-[a-z0-9]+)*$
```

## Commit Message Format

### Title Format
```
[TAG][TASK_ID] module_name: Brief description (< 50 chars)
```

### Tag Table (per Odoo Git Guidelines)

| Tag | Use Case |
|-----|----------|
| `[ADD]` | New feature or module |
| `[FIX]` | Bug fix |
| `[IMP]` | Improvement or enhancement |
| `[REF]` | Refactor (no behavior change) |
| `[REM]` | Removal (module, field, feature) |
| `[MOV]` | Move or rename |
| `[REV]` | Revert a previous commit |

### Examples

- `[ADD][1234] acme_google_drive_import: Add Google Drive import module`
- `[FIX][5678] acme_account_invoice_report: Fix discount calculation error`
- `[IMP][9012] cudio_stock_customization: Optimize stock move queries`
- `[REF][3456] acme_sale_custom_approval: Extract approval state machine`
- `[REM][7890] acme_hr_custom: Remove deprecated payroll fields`

### Commit Body

Include detailed description:

```
Long version of the change description, including the rationale
for the change or a summary of the feature being introduced.

- Explain what was changed and why
- Include any breaking changes
- Reference related issues or documentation

task-1234 (related to task)
```

### Validation Regex
```
^\[(ADD|FIX|IMP|REF|REM|MOV|REV)\]\[\d+\] [a-z][a-z0-9_]*: .{1,50}$
```

## Development Workflow

### Step 1: Branch Synchronization
Before starting:
1. Switch to development branch
2. Pull latest from origin branch
3. Resolve conflicts

### Step 2: Development and Testing
1. Implement changes following Odoo coding guidelines and Cudio conventions
2. Test thoroughly in development environment
3. Ensure existing functionality remains intact

### Step 3: Pre-PR Branch Update
Before creating PR:
1. Update branch from origin to avoid conflicts
2. Resolve merge conflicts
3. Test again

## Pull Request Process

### Creating the Pull Request

1. Create PR targeting appropriate origin branch
2. Use descriptive title following commit message format
3. Include detailed description of changes
4. Add relevant labels and assignees

### Code Review with Copilot

1. Assign GitHub Copilot as reviewer
2. Review Copilot's feedback carefully:
   - Validate suggested changes are correct
   - Ensure recommendations don't introduce errors
   - Verify changes don't break existing functionality
3. Resolve each Copilot comment:
   - Either accept suggestion by merging
   - Or manually address concern and mark resolved
   - Add explanatory comments when rejecting
4. Test merged recommendations:
   - Always test changes from Copilot recommendations
   - Ensure functionality works as expected
   - Validate no new issues introduced

### Important Notes on Copilot Review

- **Critical Validation**: Always verify Copilot's suggestions are appropriate
- **Manual Testing**: Test any changes based on Copilot recommendations
- **Documentation**: Comment on why suggestions were accepted or rejected
- **Resolution Requirement**: All Copilot comments MUST be resolved before merging

### IDE-Based Review (Optional)

Optionally perform IDE-based review (e.g., VSCode with Copilot agent) before opening PR.

### Reviewers

Lead developer (or at least one team member) MUST be added as reviewer on PR.

## Merging Process

### Squash and Merge

- Use "Squash and Merge" when merging PR
- Update final commit message to Cudio standards:
  - Title: `[TAG][TASK_ID] module_name: Brief description (< 50 chars)`
  - Description: comprehensive and accurate

### Final Commit Message Example

```
[ADD][5678] acme_account_invoice_report: Add custom invoice report module

This commit introduces a new module for generating custom invoice
reports with enhanced filtering capabilities. The module includes:

- Custom report templates with company branding
- Advanced filtering by date range, customer, and product category
- Export functionality to PDF and Excel formats
- Integration with existing invoice workflows

Resolves task-1234
```

## Post-Merge Actions

1. Delete development branch after merge (if no further development)
2. Verify changes in target environment
3. Update task/ticket status to reflect completion

## Quality Checklist (Before PR)

- [ ] Code follows Cudio coding guidelines
- [ ] Branch name follows naming convention
- [ ] Commit messages follow specified format
- [ ] Branch updated from origin to avoid conflicts
- [ ] All Copilot review comments resolved
- [ ] Changes tested thoroughly
- [ ] PR description complete and accurate

## Forbidden Patterns

-  Co-Authored-By trailers (AI attribution)
-  Merge commits (use squash and merge)
-  Force-pushes to shared branches (staging, production)
-  Committing directly to staging or production
-  Branch names without task IDs (when task ID exists)
-  Commit messages without tags

## Compact Rule Summary (for skill registry)

```
### cudio-git
- Branches: `{feat|fix|refactor|docs|test|chore}/{task-id}-{kebab-description}`
- Commits: `[TAG][TASK_ID] module_name: Brief description (< 50 chars)`
- Tags: [ADD] [FIX] [IMP] [REF] [REM] [MOV] [REV]
- Merge: ALWAYS Squash and Merge. Delete branch after merge.
- No Co-Authored-By trailers.
- No force-push to staging/production.
- Copilot review required before merge.
- Lead developer must be assigned reviewer.
```
