package contextmode

import (
	"context"
	"os/exec"

	"github.com/rd-mg/architect-ai/internal/process"
)

// Capability describes the detected context-mode environment.
type Capability struct {
	Installed       bool
	BinaryPath      string
	Version         string
	DoctorAvailable bool
}

// Detect probes the system for context-mode without running expensive health checks.
func Detect(ctx context.Context) (Capability, error) {
	// Bounded probe to avoid hanging session start.
	res, _ := process.Run(ctx, "context-mode", []string{"--version"}, process.OptionsFor(process.FastCheck))
	if res.Error != nil {
		return Capability{Installed: false}, nil
	}

	path, _ := exec.LookPath("context-mode")

	return Capability{
		Installed:  true,
		BinaryPath: path,
		Version:    string(res.Stdout),
		// Doctor is available if binary is installed.
		DoctorAvailable: true,
	}, nil
}
