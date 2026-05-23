#!/bin/bash
GATE=$(cat internal/assets/skills/_shared/adaptive-reasoning-gate.md)

for file in $(find internal/assets -name "sdd-orchestrator.md"); do
    # use awk to replace the block
    awk -v gate="$GATE" '
    /<!-- adaptive-reasoning-gate:START -->/ {
        print
        print gate
        in_block=1
        next
    }
    /<!-- adaptive-reasoning-gate:END -->/ {
        in_block=0
        print
        next
    }
    !in_block { print }
    ' "$file" > "$file.tmp" && mv "$file.tmp" "$file"
done
