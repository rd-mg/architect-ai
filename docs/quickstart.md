# Quickstart

## Prerequisites

### macOS

- Homebrew installed and available in PATH.
- `git` available.

### Ubuntu/Debian (and derivatives like Linux Mint, Pop!_OS)

- `apt-get` available (standard on these distros).
- `sudo` access for package installs.
- `git` available.

### Arch Linux (and derivatives like Manjaro, EndeavourOS)

- `pacman` available (standard on these distros).
- `sudo` access for package installs.
- `git` available.

### Fedora / RHEL family (Fedora, CentOS Stream, Rocky Linux, AlmaLinux)

- `dnf` available (standard on these distros).
- `sudo` access for package installs.
- `git` available.
- Node.js installs use NodeSource LTS setup + `dnf install -y nodejs` during dependency remediation.

### All platforms

- Go 1.24+ (for building from source).
- Node.js / npm for agent installations and MCP servers.

### Sequential Thinking MCP (MANDATORY)

The **L0 Thinking Agent** requires the official Sequential Thinking MCP server. To install it globally for your agents:

```bash
npx -y @modelcontextprotocol/server-sequential-thinking
```

Ensure your agent configuration (e.g., `~/.claude/settings.json`, `opencode.json`, etc.) includes this MCP server in its process list.

## Run

```bash
go run ./cmd/architect-ai install --dry-run
```

Use `--dry-run` first to validate selections and execution plan without applying changes. The dry-run output includes a `Platform decision` line showing the detected OS, distro, package manager, and support status.

## First real install

```bash
go run ./cmd/architect-ai install
```

The installer detects your platform automatically — no flags needed to select macOS vs Linux. Install commands are resolved through the appropriate package manager (brew, apt, pacman, or dnf) based on detection.

After completion, verify that agent configs and selected components were installed to their expected paths.

## Verification outcome

When checks pass, installer reports:

`You're ready. Run 'claude' or 'opencode' and start building.`

## Unsupported platforms

If you run the installer on an unsupported OS or Linux distro, it exits immediately with an error:

- `unsupported operating system: only macOS, Linux, and Windows are supported (detected <os>)`
- `unsupported linux distro: Linux support is limited to Ubuntu/Debian, Arch, and Fedora/RHEL family (detected <distro>)`
