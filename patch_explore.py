import glob

def patch_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    if '## Code Investigation (Section B' in content:
        print(f"Skipping {filepath}, already patched")
        return

    # Usually they have "Task: Explore {topic}..." or "## Procedure"
    # Let's insert before "## Procedure" or after "Task: ..."
    insertion_point = "## Procedure"
    idx = content.find(insertion_point)
    if idx == -1:
        # try something else
        insertion_point = "Task:"
        idx = content.find(insertion_point)
        if idx != -1:
            # find end of line
            idx = content.find('\n', idx) + 1
    
    if idx == -1:
        print(f"Failed to find insertion point in {filepath}")
        return
        
    block = """
## Code Investigation (Section B — Skim-First)

Follow the Code Skimming Protocol from `sdd-phase-common.md` Section B before reading any source file.

Investigation sequence:
1. ripgrep for candidate files: `rg "{topic}" --type go -l`
2. Skim each candidate: `grep -n "^func\\|^type\\|^var\\|^const" {file}`
3. Load only confirmed targets: `sed -n '{start},{end}p' {file}` or `cat` for small files
4. Do NOT `cat` any file that the skim shows is irrelevant

If FastCode MCP is available in your tool list, use `fastcode_skim_file` and `fastcode_get_function` instead of manual grep.
"""
    content = content[:idx] + block + "\n" + content[idx:]

    with open(filepath, 'w') as f:
        f.write(content)
    print(f"Patched {filepath}")

files = glob.glob('internal/assets/*/sdd-phase-protocols/sdd-explore.md')
files.extend(glob.glob('internal/assets/*/agents/sdd-explore.md'))
files.extend(glob.glob('internal/assets/*/commands/sdd-explore.md'))
for f in files:
    patch_file(f)

