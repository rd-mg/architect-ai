import os
import re

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    original = content

    # Remove `model.AgentGeminiCLI,` or `, model.AgentGeminiCLI`
    content = re.sub(r',\s*model\.AgentGeminiCLI\b', '', content)
    content = re.sub(r'\bmodel\.AgentGeminiCLI,\s*', '', content)
    # Remove from lists like `"gemini-cli",` or `, "gemini-cli"`
    content = re.sub(r',\s*"gemini-cli"', '', content)
    content = re.sub(r'"gemini-cli",\s*', '', content)
    
    # Remove case model.AgentGeminiCLI: and its block if empty, or just the case
    # This might be tricky, let's just remove `case model.AgentGeminiCLI:` lines
    # If it's part of a multi-case like `case A, B, AgentGeminiCLI:`
    content = re.sub(r',\s*model\.AgentGeminiCLI\b', '', content)
    content = re.sub(r'\bmodel\.AgentGeminiCLI,\s*', '', content)
    
    # Remove isolated cases: `case model.AgentGeminiCLI:`
    # We will remove the case line and maybe the lines inside if they are simple, but let's just flag them.

    if content != original:
        with open(filepath, 'w') as f:
            f.write(content)
        print(f"Updated {filepath}")

for root, _, files in os.walk('internal'):
    for file in files:
        if file.endswith('.go'):
            process_file(os.path.join(root, file))
