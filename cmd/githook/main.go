package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("architect-ai", "check", "all")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "\n[architect-ai] ERROR: pre-commit hook failed.\n")
		fmt.Fprintf(os.Stderr, "Checks did not pass. Please fix the issues before committing.\n")
		os.Exit(1)
	}
	os.Exit(0)
}
