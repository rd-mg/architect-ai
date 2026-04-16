---
name: ripgrep
trigger: "When the user asks for ripgrep, rg command, search files, find in files, or code search."
description: This skill MUST be used when the user or agent asks about "ripgrep", "rg command", "fast grep", "search files", "code search", "regex search in files", "find in files", "search codebase", "grep replacement", "find replacement", "ack replacement", "Read file", "Read files", "Read directory", "Read directories", "Read directory tree", "Read directory trees", "Read directory trees", or needs guidance on file content searching, pattern matching, recursive text search, or using ripgrep as a pre-filter for LLM text extraction pipelines.
---

##  MANDATORY: Replacement Standard
`ripgrep` (`rg`) IS THE MANDATORY REPLACEMENT for `grep`, `find`, and `ack` in all terminal commands, scripts, and automated pipelines across this project. No exceptions.

# ripgrep (rg) CLI Reference & Architecture

ripgrep is a line-oriented search tool that recursively searches directories for regex patterns. It respects `.gitignore` rules by default and automatically skips hidden files, directories, and binary files. 

While traditionally used for code search, **it is also the ideal "Regex Parser" for hybrid Regex-to-LLM text extraction pipelines**.

## Integration in Hybrid Pipelines (Regex -> LLM)

When parsing structured text (forms, invoices, logs, quizzes) at scale, **do not send raw text directly to an LLM**. Instead, use the "Regex First, LLM Second" architectural pattern:

1. **Regex Parser (`ripgrep`)**: Use `ripgrep` to extract the predictable 95-98% of the data deterministically and at zero token cost.
2. **Confidence Scorer**: Filter out extractions that miss expected fields programmatically.
3. **LLM Validator**: Only send the remaining 2-5% (low confidence outliers) to an LLM for evaluation.

### 1. Extracting Structured Data with JSON Output (`--json`)

`ripgrep` can act as a local parser API that emits structured matches, contexts, and capture groups, which your Python/Node pipeline can easily consume.

```bash
# Output raw search results as a stream of JSON objects
rg --json '(?P<id>\d+)\.\s*(?P<text>.+?)' data/
```

*Inside a script (Confidence Scorer):*
```python
import json, subprocess

result = subprocess.run(['rg', '--json', '(?P<id>\d+)\.\s*(?P<text>.+?)', 'data/'], capture_output=True, text=True)
for line in result.stdout.splitlines():
    data = json.loads(line)
    if data['type'] == 'match':
        # Safely access capture groups
        # data['data']['submatches'][0]['match']['text']
        print(data)
```

### 2. Advanced Extraction: PCRE2 & Multiline (`-U` + `-P`)

To extract repetitive blocks (e.g., questions with options, or log blocks), use multiline and PCRE2 capture groups.

```bash
# Extract full multiline elements using capture groups
rg -U -P '(?P<id>\d+)\.\s*(?P<text>.+?)\n(?P<choices>(?:[A-D]\..+?\n)+)Answer:\s*(?P<answer>[A-D])' docs/

# Output ONLY the matched text, ignoring surrounding file content
rg -U -P -o 'pattern'
```

### 3. Detecting LLM Candidates (Inverted Match `-v`)

Use inverted matches to identify files or lines that **failed** your regex baseline. These are your out-of-format anomalies (the `<5%`) that you should delegate to an LLM.

```bash
# Find lines that DO NOT match the expected structured format
rg -v '^\d+\.\s+[A-Z]' file.txt > outliers_for_llm.txt
```

---

## Best Practices for Sub-Agents

1. **Start with ripgrep:** Even an imperfect regex establishes a baseline. Do this before reaching for LLM tools.
2. **Paging Context:** Use `-A 5` (after) and `-B 5` (before) to feed an LLM *only* the relevant lines around an error or pattern instead of an entire 2,000-line file.
3. **Format-Aware Filtering:** Always use `--type` (e.g., `-trust`, `-tjs`) to narrow down searches instead of searching all files.

---

## Basic Usage

```bash
# Search for pattern in current directory (recursive)
rg 'pattern'

# Search specific file
rg 'pattern' file.txt

# Search specific directory
rg 'pattern' src/

# Multiple paths
rg 'pattern' src/ tests/ docs/
```

## Essential Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--ignore-case` | `-i` | Case-insensitive search |
| `--smart-case` | `-S` | Case-insensitive unless pattern has uppercase |
| `--word-regexp` | `-w` | Match whole words only |
| `--fixed-strings` | `-F` | Treat pattern as literal string (no regex) |
| `--count` | `-c` | Show count of matches per file |
| `--files-with-matches` | `-l` | Show only filenames with matches |
| `--files-without-match` | `-L` | Show only filenames without matches |
| `--line-number` | `-n` | Show line numbers (default when output is terminal) |
| `--no-line-number` | `-N` | Suppress line numbers |
| `--only-matching` | `-o` | Print only the matched text |
| `--invert-match` | `-v` | Show non-matching lines |
| `--multiline` | `-U` | Enable multiline matching (`.` matches newlines) |

## Context Lines

```bash
# Show 3 lines after each match
rg -A 3 'pattern'

# Show 3 lines before each match
rg -B 3 'pattern'

# Show 3 lines before and after (context)
rg -C 3 'pattern'
```

## File Filtering

### By File Type

```bash
# Search only in Rust files
rg 'pattern' --type rust
rg 'pattern' -trust

# Exclude JavaScript files
rg 'pattern' --type-not js
rg 'pattern' -Tjs

# List all available types
rg --type-list

# Define custom type
rg --type-add 'web:*.{html,css,js}' -tweb 'pattern'
```

Common built-in types: `py`, `js`, `ts`, `rust`, `go`, `java`, `c`, `cpp`, `html`, `css`, `json`, `yaml`, `md`, `sh`

### By Glob Pattern

```bash
# Include only .toml files
rg 'pattern' -g '*.toml'

# Include files in specific directory
rg 'pattern' -g 'src/**/*.rs'

# Exclude files (prefix with !)
rg 'pattern' -g '!*.min.js'

# Multiple globs
rg 'pattern' -g '*.rs' -g '*.toml'
```

## Automatic Filtering Control

ripgrep automatically ignores:
1. Files matching `.gitignore`, `.ignore`, `.rgignore` patterns
2. Hidden files and directories
3. Binary files

```bash
# Disable .gitignore filtering
rg --no-ignore 'pattern'

# Search hidden files
rg --hidden 'pattern'
rg -. 'pattern'

# Search binary files as text
rg --text 'pattern'
rg -a 'pattern'

# Shorthand for disabling filters
rg -u 'pattern'    # --no-ignore
rg -uu 'pattern'   # --no-ignore --hidden
rg -uuu 'pattern'  # --no-ignore --hidden --text
```

## Regex Patterns

ripgrep uses Rust regex syntax by default.

```bash
# Word characters followed by digits
rg '\w+\d+'

# Start/end of line anchors
rg '^start'
rg 'end$'

# Character classes
rg '[A-Z][a-z]+'

# Alternation
rg 'foo|bar'

# Quantifiers
rg 'colou?r'       # 0 or 1
rg 'a+'            # 1 or more
rg 'a*'            # 0 or more
rg 'a{2,4}'        # 2 to 4 times

# Non-greedy matching
rg 'a+?'

# Lookahead/lookbehind (requires PCRE2)
rg -P '(?<=prefix)\w+'    # Lookbehind
rg -P '\w+(?=suffix)'     # Lookahead
```

### PCRE2 Mode

Enable PCRE2 for advanced regex features:

```bash
# Enable PCRE2
rg -P 'pattern'
rg --pcre2 'pattern'

# Auto-detect when PCRE2 is needed
rg --engine auto 'pattern'
```

## Multiline Search

```bash
# Basic multiline (dot matches newline)
rg -U 'start.*end'

# Match across lines with PCRE2
rg -UP 'function\s+\w+\s*\([^)]*\)\s*\{'
```

## Replacements (Output Only)

ripgrep never modifies files. The `--replace` flag transforms output only.

```bash
# Simple replacement
rg 'old' -r 'new'

# With capture groups
rg '(\w+)@(\w+)' -r '$2:$1'

# Named capture groups
rg '(?P<user>\w+)@(?P<domain>\w+)' -r '$domain:$user'
```

## Output Formats

```bash
# JSON output (for parsing)
rg --json 'pattern'

# Null-separated filenames (for xargs -0)
rg -l --null 'pattern'

# Show only matching files
rg -l 'pattern'

# Show files that would be searched
rg --files

# Show files matching glob
rg --files -g '*.rs'
```

## Performance Options

```bash
# Limit search depth
rg --max-depth 3 'pattern'

# Limit results
rg --max-count 5 'pattern'    # Per file
rg 'pattern' | head -n 100    # Total

# Disable memory maps (for consistency)
rg --no-mmap 'pattern'

# Sort results by path
rg --sort path 'pattern'
```

## Common Patterns

### Find function definitions

```bash
# Rust
rg '^(pub\s+)?(async\s+)?fn\s+\w+'

# Python
rg '^def \w+|^class \w+'

# JavaScript/TypeScript
rg '(function|const|let|var)\s+\w+\s*='
```

### Find TODO/FIXME comments

```bash
rg 'TODO|FIXME|XXX|HACK' -g '!*.min.*'
```

### Find imports/requires

```bash
# Python
rg '^(import|from)\s+\w+'

# JavaScript
rg "^(import|require\()"

# Go
rg '^import\s+'
```

### Search specific patterns

```bash
# IP addresses (approximate)
rg '\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}'

# Email addresses (simple)
rg '\w+@\w+\.\w+'

# URLs
rg 'https?://[^\s]+'
```

## Configuration File

Set default options via `RIPGREP_CONFIG_PATH`:

```bash
export RIPGREP_CONFIG_PATH="$HOME/.ripgreprc"
```

Example `~/.ripgreprc`:

```
--smart-case
--hidden
--glob=!.git/*
--max-columns=150
--max-columns-preview
```

## Debugging

```bash
# Show what files would be searched and why others are ignored
rg --debug 'pattern'

# Disable config file
rg --no-config 'pattern'
```

## Key Differences from grep

| Feature | ripgrep | GNU grep |
|---------|---------|----------|
| Recursive by default | Yes | No (`-r` required) |
| Respects .gitignore | Yes | No |
| Skips binary files | Yes | No |
| Unicode by default | Yes | Depends on locale |
| Parallel search | Yes | No |
| Memory-mapped files | Yes | No |

## Reference

For advanced features, see:
- [references/advanced.md](references/advanced.md) - Preprocessors, encoding, binary handling
