# Design: OpenSpec INDEX Auto-Generation

## Architecture Decisions

### AD1: Shell vs Go
We will use a Bash shell script embedded in the `sdd-archive` skill.
- **Rationale**: Minimal complexity, zero compilation required, and it solves the immediate problem for the primary target OS (Linux).
- **Fallback**: If the shell environment is unavailable, the index generation is skipped (advisory only).

### AD2: Overwrite Strategy
The `INDEX.md` will be fully regenerated on every successful archive.
- **Rationale**: Ensures the index is always atomic and synchronized with the filesystem. No need for complex delta-patching.

## Proposed Changes

### Indexing Script (Logic)
```bash
# Target: openspec/specs/INDEX.md
# 1. Header
echo "# OpenSpec Capability Index" > openspec/specs/INDEX.md
echo "" >> openspec/specs/INDEX.md
echo "| Domain | Title | Path |" >> openspec/specs/INDEX.md
echo "|--------|-------|------|" >> openspec/specs/INDEX.md

# 2. Row Generation
for d in openspec/specs/*/; do
  domain=$(basename "$d")
  spec_file="$d/spec.md"
  
  if [ -f "$spec_file" ]; then
    # Extract first line, strip '# ', trim whitespace
    title=$(head -1 "$spec_file" | sed 's/^# //' | xargs)
    echo "| $domain | $title | \`specs/$domain/spec.md\` |" >> openspec/specs/INDEX.md
  fi
done
```

### Skill Patch Points

#### [MODIFY] sdd-archive/SKILL.md
- **Insertion Point**: After the successful merge block (Step 2b).
- **Addition**: "Step 3 — Regenerate OpenSpec Index" with the script above.

#### [MODIFY] sdd-explore/SKILL.md & sdd-propose/SKILL.md
- **Insertion Point**: "Strategic Resumption" or "Research" sections.
- **Addition**: "MANDATORY: Check `openspec/specs/INDEX.md` before initiating any research into existing capabilities."

## Verification Plan

### Automated Tests
- N/A (Manual shell verification).

### Manual Verification
1. Run the script manually in the terminal.
2. Verify `openspec/specs/INDEX.md` exists.
3. Verify table contents match the directory structure.
4. Simulate an archive by running the script again and checking for persistence.
