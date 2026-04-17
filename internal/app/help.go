package app

import (
	"fmt"
	"io"
)

func printHelp(w io.Writer, version string) {
	fmt.Fprintf(w, `architect-ai — Architect AI Stack (%s)

USAGE
  architect-ai                     Launch interactive TUI
  architect-ai <command> [flags]

COMMANDS
  install      Configure AI coding agents on this machine
  uninstall    Remove Architect AI managed files from this machine
  sync         Sync agent configs and skills to current version
  update       Check for available updates
  upgrade      Apply updates to managed tools
  cleanup      Deep clean of temporary files and logs
  restore      Restore a config backup
  overlay      Manage specialist project overlays (e.g. Odoo)
  skill-registry Generate or refresh the project's .atl/skill-registry.md
  sdd-init     Initialize SDD context in the current project
  version      Print version

FLAGS
  --help, -h    Show this help

Run 'architect-ai help' for this message.
Documentation: %s
`, version, DocsURL)
}
