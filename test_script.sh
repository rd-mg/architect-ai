#!/usr/bin/env bash
JSON='{"status": "completed", "phase": "p", "executive_summary": "sum", "artifacts": [], "next_recommended": "n", "risks": [], "skill_resolution": {"status": "ok"}}'
echo "$JSON" | ./validate-result-contract.sh
