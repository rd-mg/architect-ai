import glob

def patch_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    if 'Eval Gate Check' in content:
        print(f"Skipping {filepath}, already patched")
        return

    # Find the Entity Tag Extraction or Step 3 to insert before
    idx = content.find('### Step 3c: Entity Tag Extraction')
    if idx == -1:
        idx = content.find('### Step 3:')
    if idx == -1:
        idx = content.find('## Step 3')
    
    if idx == -1:
        # Just append before ## Rules or ## Return Envelope
        idx = content.find('## Rules')
        
    if idx == -1:
        print(f"Failed to find insertion point in {filepath}")
        return
        
    block = """#### Step 2c: Eval Gate Check (MANDATORY)

Before completing the archive, verify that NO tasks in the change were classified as `HIGH` risk without an explicit, recorded eval step (e.g. peer review, manual validation log, test evidence) present in the archive report or in the `mem_search` observation history for the change.

If a `HIGH` risk task has no logged eval evidence:
1. Halt archiving.
2. Return status `blocked`.
3. Inform the orchestrator: "High risk task <ID> requires eval evidence before archive."

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

