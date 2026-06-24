package contextmode

import (
	"context"
	"os/exec"
	"strings"

	"github.com/rd-mg/architect-ai/internal/process"
)

// Capability describes the detected context-mode environment.
type Capability struct {
	Installed       bool
	BinaryPath      string
	Version         string
	MCPModeSupport  bool
	HookModeSupport bool
	DoctorClean     bool
}

func Detect(ctx context.Context) (Capability, error) {
	res, _ := process.Run(ctx, "context-mode", []string{"--version"}, process.OptionsFor(process.FastCheck))
	if res.Error != nil {
		return Capability{Installed: false}, nil
	}

	path, _ := exec.LookPath("context-mode")
	version := strings.TrimSpace(string(res.Stdout))

	helpOut, _ := process.Run(ctx, "context-mode", []string{"--help"}, process.OptionsFor(process.FastCheck))
	mcpSupport := strings.Contains(string(helpOut.Stdout), "--mcp")

	hookOut, _ := process.Run(ctx, "context-mode", []string{"hook", "--help"}, process.OptionsFor(process.FastCheck))
	hookSupport := len(hookOut.Stdout) > 0

	doctorOut, doctorErr := process.Run(ctx, "context-mode", []string{"doctor"}, process.OptionsFor(process.FastCheck))
	doctorClean := doctorErr == nil && !strings.Contains(string(doctorOut.Stdout), "ERROR")

	return Capability{
		Installed:       true,
		BinaryPath:      path,
		Version:         version,
		MCPModeSupport:  mcpSupport,
		HookModeSupport: hookSupport,
		DoctorClean:     doctorClean,
	}, nil
}
