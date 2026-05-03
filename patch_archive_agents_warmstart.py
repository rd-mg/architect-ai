import glob
import re

def patch_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    if 'Step 4b: Extract and Save Learned Patterns' in content:
        print(f"Skipping {filepath}, already patched")
        return

    # Find the persistence section or the end of the procedure
    idx = content.find('4. If OpenSpec mode:')
    if idx == -1:
        idx = content.find('5. Update DAG state')
    if idx == -1:
        idx = content.find('## Artifact Store')
        
    if idx == -1:
        print(f"Failed to find insertion point in {filepath}")
        return
        
    block = """
4b. Extract and Save Learned Patterns: For each SDD phase used, synthesize 3-5 bullet patterns about what worked/failed. Search `knowledge/_global/skill/{skill-name}/learned-patterns`. If found, `mem_update` with appended patterns and incremented version. If not, `mem_save` new patterns.
"""
    content = content[:idx] + block + content[idx:]

    with open(filepath, 'w') as f:
        f.write(content)
    print(f"Patched {filepath}")

files = glob.glob('internal/assets/*/sdd-phase-protocols/sdd-archive.md')
files.extend(glob.glob('internal/assets/*/agents/sdd-archive.md'))
files.extend(glob.glob('internal/assets/*/commands/sdd-archive.md'))
for f in files:
    patch_file(f)

