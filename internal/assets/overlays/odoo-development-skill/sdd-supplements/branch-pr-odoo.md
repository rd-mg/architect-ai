# Overlay Branch & PR Supplement — Odoo (Cudio)

Overlay-specific branch and PR rules that EXTEND the base `branch-pr` skill.
Base branch-pr rules still apply. These are ADDITIONAL constraints for
Cudio Odoo projects.

## Branch Naming Override

Format: `{prefix}/{task-id}-{brief-description}`

**Source**: `rules/cudio-git.md` section "Branch Naming"

Validation regex (applied by agent before branch creation):
```
^(feat|fix|refactor|docs|test|chore)/\d+-[a-z0-9]+(-[a-z0-9]+)*$
```

If the proposed branch name fails validation, the agent MUST NOT proceed with
branch creation. Instead, return with a suggestion:

> Branch name `{proposed}` does not match Cudio convention.
> Expected format: `{feat|fix|refactor|docs|test|chore}/{task-id}-{description}`
> Example: `feat/12345-add-approval-workflow`

## Commit Message Override

Format: `[TAG][TASK_ID] module_name: Brief description (< 50 chars)`

**Source**: `rules/cudio-git.md` section "Commit Message Format"

Validation regex:
```
^\[(ADD|FIX|IMP|REF|REM|MOV|REV)\]\[\d+\] [a-z][a-z0-9_]*: .{1,50}$
```

Body requirements:
- Rationale for the change
- List of what was changed and why
- Any breaking changes
- Task reference (e.g., `task-1234`)

## PR Title & Body Requirements

In addition to base branch-pr requirements:

### Title
Must match the commit message title format:
```
[TAG][TASK_ID] module_name: Brief description
```

### Body Required Sections

```markdown
## Module(s) Affected
- {module_name} (v{version})
- ...

## Odoo Version
{18.0 | 19.0 | ...}

## Change Summary
{1-2 paragraph description}

## Migration Included
{Yes/No — if Yes, list pre-migrate and post-migrate scripts}

## Manifest Version Change
- Old: {old version}
- New: {new version}
- Rationale: {why this bump level (Z vs W)}

## Testing Done
- [ ] Tested on staging
- [ ] Tested with non-admin user
- [ ] Multi-company behavior verified (if applicable)
- [ ] Clean database install verified
- [ ] Upgrade path verified

## Related Task
task-{task-id}
```

## Merge Strategy

- **ALWAYS** use "Squash and Merge"
- Final commit message MUST follow the Cudio `[TAG][TASK_ID]` format
- Delete branch after merge

The agent MUST NOT propose merge commits or rebase-and-merge strategies.

## Copilot Review Guidance

When the agent performs code review (via adaptive-reasoning adversarial-review
or sdd-verify), apply the Cudio code review standards:

- Validate that Copilot/AI suggestions don't introduce errors
- Verify changes don't break existing functionality
- Document why suggestions were accepted or rejected
- Ensure all review comments are resolved before marking as APPROVED

The Cudio process REQUIRES:
1. GitHub Copilot review assigned
2. All Copilot comments resolved (accepted with justification, or rejected with explanation)
3. Lead developer (or at least one team member) assigned as reviewer
4. Reviewer approval before squash-and-merge

## Pre-Merge Checklist (Automated)

Before marking the PR as ready-to-merge, the agent validates:

- [ ] Branch name matches Cudio convention
- [ ] All commits on branch match `[TAG][TASK_ID]` format
- [ ] PR title matches format
- [ ] PR body has all required sections filled
- [ ] Manifest version was bumped (from git diff)
- [ ] README updated (if applicable)
- [ ] Changelog updated (if applicable)
- [ ] No Co-Authored-By trailers

If any fails, the agent MUST report the failure and suggest the fix.

## Human Actions Required

The agent CANNOT:
- Assign Copilot as reviewer (human action in GitHub UI)
- Assign lead developer as reviewer (human action)
- Perform the actual squash-and-merge (human action)
- Delete the remote branch (human/CI action)

The agent CAN:
- Create the branch locally with the correct name
- Generate correctly-formatted commits
- Generate the PR body
- Push the branch
- Generate the `gh pr create` command

## Example Agent Output

```markdown
## Branch Creation
Branch: `feat/12345-add-approval-workflow`
Status: ✅ Created

## Commit
```
[ADD][12345] acme_sale_custom_approval: Add approval workflow module
```

## PR Command
```bash
gh pr create \
  --title "[ADD][12345] acme_sale_custom_approval: Add approval workflow module" \
  --body "$(cat pr-body.md)" \
  --base staging-dev \
  --label "type:feature"
```

## Human Actions Required
1. Review and run the `gh pr create` command above
2. In GitHub UI:
   - Assign GitHub Copilot as reviewer
   - Assign {lead-developer} as reviewer
3. After review approval, use "Squash and Merge"
4. After merge, delete the branch
```
