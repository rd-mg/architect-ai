## Apply Branch Protocol [sdd-apply MANDATORY]

Git isolation for all file modifications during sdd-apply.
No remote required. No PR required. Works in any git repo.

### Pre-flight
```bash
set -euo pipefail
ORIGINAL_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null) || {
  echo "WARN: not a git repo — skipping branch isolation"
  echo "GIT_AVAILABLE=false"
  exit 0
}
CHANGE_NAME="${1:?change_name required}"
APPLY_BRANCH="apply/$(echo "${CHANGE_NAME}" | tr '[:upper:] ./_' '[:lower:]----')"
echo "original=${ORIGINAL_BRANCH} apply=${APPLY_BRANCH}"
```

### Setup (before any file edit)
```bash
# Verify clean state
git diff --quiet && git diff --cached --quiet || {
  echo "ERROR: Uncommitted changes. Commit or stash first." >&2; exit 1
}
# Clean stale apply branch if exists
git branch -D "${APPLY_BRANCH}" 2>/dev/null || true
# Create and switch
git checkout -b "${APPLY_BRANCH}"
echo "ISOLATED on ${APPLY_BRANCH} — original branch is safe"
```

### During Apply
```
- Every task from execution graph = at least 1 commit
- Format: type(scope): description  (Conventional Commits)
- git status between tasks — verify no accidental files staged
- Tests run HERE, in this branch
```

### On Success (tests pass)
```bash
git checkout "${ORIGINAL_BRANCH}"

# Try fast-forward first (clean linear history)
if git merge --ff-only "${APPLY_BRANCH}" 2>/dev/null; then
  echo "MERGED via fast-forward — clean history"
else
  # Diverged: use explicit merge commit
  git merge --no-ff "${APPLY_BRANCH}" \
    -m "feat(${CHANGE_NAME}): sdd-apply complete — merged from ${APPLY_BRANCH}"
  echo "MERGED via merge commit"
fi

git branch -d "${APPLY_BRANCH}"
echo "CLEAN: apply branch deleted"
```

### On Failure (tests fail or BLOCKED)
```bash
git checkout "${ORIGINAL_BRANCH}"
echo "BLOCKED — ${APPLY_BRANCH} preserved for inspection"
echo "  Inspect: git log ${APPLY_BRANCH} --oneline"
echo "  Discard: git branch -D ${APPLY_BRANCH}"
# STATUS: FAILED | branch: ${APPLY_BRANCH} | reason: {summary}
```

### Fallback (no git / non-git project)
```
GIT_AVAILABLE = false
WARN: "No git isolation. Changes applied directly. RISK: manual review required."
ADD risk to verify report. Continue apply in current directory.
```

### Opt-in: Remote PR (user-initiated only)
```bash
# NOT automatic. Only if user explicitly requests:
git push origin "${APPLY_BRANCH}"
gh pr create --base "${ORIGINAL_BRANCH}" --head "${APPLY_BRANCH}" --draft
```
