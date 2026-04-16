package gga

import (
	"github.com/rd-mg/architect-ai/internal/installcmd"
	"github.com/rd-mg/architect-ai/internal/model"
	"github.com/rd-mg/architect-ai/internal/system"
)

func InstallCommand(profile system.PlatformProfile) ([][]string, error) {
	return installcmd.NewResolver().ResolveComponentInstall(profile, model.ComponentGGA)
}

func ShouldInstall(enabled bool) bool {
	return enabled
}
