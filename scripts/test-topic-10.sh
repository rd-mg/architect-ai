#!/bin/bash
set -e

REPO_ROOT=$(pwd)
CHANGE_NAME="test-conflict"
DOMAIN="auth"

echo "1. Setup baseline"
mkdir -p openspec/specs/$DOMAIN
echo "# Auth Spec" > openspec/specs/$DOMAIN/spec.md
MAIN_SHA=$(sha256sum openspec/specs/$DOMAIN/spec.md | cut -d' ' -f1)

echo "2. Create change with delta"
mkdir -p openspec/changes/$CHANGE_NAME/specs/$DOMAIN
cat <<EOF > openspec/changes/$CHANGE_NAME/specs/$DOMAIN/spec.md
---
openspec_delta:
  base_sha: "$MAIN_SHA"
  base_path: "openspec/specs/$DOMAIN/spec.md"
  base_captured_at: "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  generator: test-script
  generator_version: 1
---
# Delta content
EOF

cat <<EOF > openspec/changes/$CHANGE_NAME/state.yaml
schema_version: 1
change_name: $CHANGE_NAME
created_at: 2026-04-18T00:00:00Z
updated_at: 2026-04-18T00:00:00Z
artifact_store: hybrid
phases:
  sdd-archive:
    status: pending
EOF

echo "3. Verify clean preflight"
./architect-ai sdd-archive-preflight $CHANGE_NAME

echo "4. Modify main to cause conflict"
echo "Unauthorized edit" >> openspec/specs/$DOMAIN/spec.md

echo "5. Run preflight again (should FAIL)"
if ./architect-ai sdd-archive-preflight $CHANGE_NAME; then
  echo "FAIL: Preflight should have failed but exited 0"
  exit 1
else
  echo "SUCCESS: Preflight failed as expected"
fi

echo "6. Verify report and state"
if [ -f openspec/changes/$CHANGE_NAME/merge-conflict.md ]; then
  echo "SUCCESS: merge-conflict.md found"
  grep "Domain \`auth\`" openspec/changes/$CHANGE_NAME/merge-conflict.md
else
  echo "FAIL: merge-conflict.md NOT found"
  exit 1
fi

grep "status: failed" openspec/changes/$CHANGE_NAME/state.yaml
echo "INTEGRATION TEST PASSED"

# Cleanup
rm -rf openspec/specs/$DOMAIN
rm -rf openspec/changes/$CHANGE_NAME
