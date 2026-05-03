import glob
import os
import sys

def patch_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    # We need to insert Step 3c after "Any open follow-ups\n"
    if 'Step 3c: Entity Tag Extraction' in content:
        print(f"Skipping {filepath}, already patched")
        return

    insertion_point = "- Any open follow-ups\n"
    if insertion_point not in content:
        # try fallback
        insertion_point = "Any open follow-ups\n"

    idx = content.find(insertion_point)
    if idx == -1:
        print(f"Failed to find insertion point in {filepath}")
        return

    idx += len(insertion_point)
    
    step_3c = """   - Step 3c: Entity Tag Extraction (scan artifacts and extract named entities: modules, types, functions, files, services, concepts, risks)
"""
    content = content[:idx] + step_3c + content[idx:]

    # Now update mem_save
    # Find mem_save(...)
    old_content_line = 'content: "{archive summary markdown}"'
    new_content_line = 'content: "{archive summary markdown with entity index at bottom}"'
    content = content.replace(old_content_line, new_content_line)

    with open(filepath, 'w') as f:
        f.write(content)
    print(f"Patched {filepath}")

# also patch .agent/skills/sdd-archive/SKILL.md
def patch_skill():
    skill_path = ".agent/skills/sdd-archive/SKILL.md"
    if not os.path.exists(skill_path):
        return
    with open(skill_path, 'r') as f:
        content = f.read()
    if 'Step 3c: Entity Tag Extraction' in content:
        return
    
    insertion_point = "### Step 3: Move to Archive"
    idx = content.find(insertion_point)
    if idx == -1:
        print("Failed to patch SKILL.md")
        return
    
    step_3c = """### Step 3c: Entity Tag Extraction (runs as part of summary generation)

Before writing the archive report, scan all artifacts for this change and extract named entities (modules, types, functions, files, services, concepts, risks).
Format as an '## Entity Index' block at the end of the archive report.

"""
    content = content[:idx] + step_3c + content[idx:]
    with open(skill_path, 'w') as f:
        f.write(content)
    print("Patched SKILL.md")


files = glob.glob('internal/assets/*/sdd-phase-protocols/sdd-archive.md')
files.extend(glob.glob('internal/assets/*/agents/sdd-archive.md'))
files.extend(glob.glob('internal/assets/*/commands/sdd-archive.md'))
for f in files:
    patch_file(f)

patch_skill()
