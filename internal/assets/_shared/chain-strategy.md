## Chain Strategy [Applied when Delivery Strategy requires splitting]

When a change needs to be split into multiple PRs, choose ONE of these strategies.
Ask the user ONCE which they prefer for this change. Cache for the change (not the session).

### Strategy A: Stacked PRs to main

```
main ← PR #1 (first work unit)
main ← PR #2 (second work unit, independent of PR #1)
main ← PR #3 (third work unit, independent)
```

Use when: Each work unit can land independently on main without the others.
Benefit: Each PR is reviewable on its own. No coordination needed.
Risk: If PRs have inter-dependencies, merging order matters.

Branch naming:
```bash
git checkout -b feat/{change-name}-slice-1
git checkout -b feat/{change-name}-slice-2
git checkout -b feat/{change-name}-slice-3
```

### Strategy B: Feature Branch Chain

```
main ← tracker PR (draft, never merges)
         ↑
    feature/{change-name}   ← PR #1 (child, targets feature branch)
         ↑
    feature/{change-name}-2 ← PR #2 (targets PR #1 branch)
         ↑
    feature/{change-name}-3 ← PR #3 (targets PR #2 branch)
```

Use when: Work units must integrate before reaching main (shared state, schema migration, etc.)
Benefit: Each child PR shows only its delta (not the full feature).
Risk: Reviewers must understand the chain order. Tracker PR stays as draft.

Commands:
```bash
# Create tracker branch (draft PR — never merges)
git checkout -b feature/{change-name}
gh pr create --base main --head feature/{change-name} --draft \
  --title "feat: {change-name} [tracker — do not merge]"

# Create slice 1 from tracker
git checkout -b feature/{change-name}-slice-1 feature/{change-name}
# ... implement slice 1 ...
gh pr create --base feature/{change-name} --head feature/{change-name}-slice-1

# Create slice 2 from slice 1
git checkout -b feature/{change-name}-slice-2 feature/{change-name}-slice-1
# ... implement slice 2 ...
gh pr create --base feature/{change-name}-slice-1 --head feature/{change-name}-slice-2
```

### Decision Gate for Chain Strategy

| Condition | Strategy |
|---|---|
| Work units are independent (can land on main alone) | Stacked PRs to main |
| Work units share state / depend on each other | Feature Branch Chain |
| Auto-chain mode AND budget_risk = medium | Stacked PRs (simpler) |
| Auto-chain mode AND budget_risk = high | Feature Branch Chain (safer) |
| User explicitly requested stacked | Stacked PRs |
| User explicitly requested feature branch | Feature Branch Chain |

### Dependency Diagram (include in EVERY chained PR)

Every PR in a chain MUST include this in its description:

```markdown
## Chain Context

**Change:** {change-name}  
**This PR:** Slice {N} of {M}  
**Strategy:** Stacked PRs / Feature Branch Chain  
**Depends on:** PR #{previous} (or main)  
**Followed by:** PR #{next} (or nothing — last slice)  
**Tracker:** #{tracker_pr_number} (Feature Branch Chain only)

### Dependency Diagram
```
main
 ├── [📍 This PR] feat/{change-name}-slice-{N}  ← current
 ├── [ ] feat/{change-name}-slice-{N+1}         ← next (pending)
 └── [ ] feat/{change-name}-slice-{N+2}         ← future
```

**Out of scope for this PR:** {what is explicitly NOT in this PR}
```

### Anti-patterns (never do these)

- Mixing chain strategies mid-change (choose once, stick with it)
- Merging a Feature Branch Chain tracker PR (it's always draft)
- Reviewing child PRs against main (only against the parent branch)
- PRs with polluted diffs from parent branch (retarget or rebase if this happens)
