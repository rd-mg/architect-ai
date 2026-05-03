import glob
import re

def patch_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    if 'Eval Gate Check' in content:
        print(f"Skipping {filepath}, already patched")
        return

    # Find where to insert in Procedure
    idx = content.find('4. If OpenSpec mode')
    if idx == -1:
        idx = content.find('5. Update DAG state')
        
    if idx == -1:
        print(f"Failed to find insertion point in {filepath}")
        return
        
    block = """3b. Eval Gate Check: Verify NO tasks classified as HIGH risk lack an explicit eval step (peer review, manual validation, test evidence) in the mem_search observation history. If HIGH risk tasks lack evidence, STOP and return blocked.
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

