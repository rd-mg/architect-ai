# Cudio Git & Branch Convention

Organization-specific Git workflow rules for Cudio projects. Extends the
generic `branch-pr` skill.

## Branch Structure

Cudio repositories follow this hierarchy:
- **1 Production branch** — stable, production-ready code
- **Multiple Staging branches** — testing and validation
- **Multiple Development branches** — active development work

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

- Always include the task/ticket ID when available
- Use kebab-case for the description
- Keep descriptions concise but descriptive
- Create branches from the appropriate source branch (typically `staging-dev`)

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

Include a detailed description:

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
Before starting development:
1. Switch to your development branch
2. Pull the latest changes from the origin branch
3. Resolve any conflicts if they exist

### Step 2: Development and Testing
1. Implement changes following Odoo coding guidelines and Cudio conventions
2. Test thoroughly in your development environment
3. Ensure all existing functionality remains intact

### Step 3: Pre-PR Branch Update
Before creating a Pull Request:
1. Update your branch from the origin branch to avoid conflicts
2. Resolve any merge conflicts
3. Test again to ensure everything works correctly

## Pull Request Process

### Creating the Pull Request

1. Create a PR targeting the appropriate origin branch
2. Use a descriptive title following the commit message format
3. Include a detailed description of changes
4. Add relevant labels and assignees

### Code Review with Copilot

1. Assign GitHub Copilot as a reviewer
2. Review Copilot's feedback carefully:
   - Validate that suggested changes are correct
   - Ensure recommendations don't introduce errors
   - Verify changes don't break existing functionality
3. Resolve each Copilot comment:
   - Either accept the suggestion by merging it
   - Or manually address the concern and mark as resolved
   - Add explanatory comments when rejecting suggestions
4. Test merged recommendations:
   - Always test changes made based on Copilot recommendations
   - Ensure functionality works as expected
   - Validate that no new issues are introduced

### Important Notes on Copilot Review

- **Critical Validation**: Always verify Copilot's suggestions are appropriate for the specific use case
- **Manual Testing**: Test any changes made based on Copilot recommendations
- **Documentation**: Comment on why certain suggestions were accepted or rejected
- **Resolution Requirement**: All Copilot comments MUST be resolved before merging

### IDE-Based Review (Optional)

Optionally perform an IDE-based review (e.g., PyCharm Copilot agent) for additional coverage before opening the PR.

### Reviewers

The lead developer (or at least one other team member) MUST be added as a reviewer on the PR.

## Merging Process

### Squash and Merge

- Use the "Squash and Merge" option when merging the PR
- Update the final commit message to follow Cudio standards:
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

1. Delete the development branch after successful merge (if no further development needed)
2. Verify the changes in the target environment
3. Update task/ticket status to reflect completion

## Quality Checklist (Before PR)

- [ ] Code follows Cudio coding guidelines
- [ ] Branch name follows naming convention
- [ ] Commit messages follow the specified format
- [ ] Branch is updated from origin to avoid conflicts
- [ ] All Copilot review comments resolved
- [ ] Changes have been tested thoroughly
- [ ] PR description is complete and accurate

## Forbidden Patterns

- ❌ Co-Authored-By trailers (AI attribution)
- ❌ Merge commits (use squash and merge)
- ❌ Force-pushes to shared branches (staging, production)
- ❌ Committing directly to staging or production
- ❌ Branch names without task IDs (when a task ID exists)
- ❌ Commit messages without tags

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
