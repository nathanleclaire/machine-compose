package provision

import (
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/auth"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/engine"
)

type EngineConfigContext struct {
	DockerPort       int
	AuthOptions      auth.AuthOptions
	EngineOptions    engine.EngineOptions
	DockerOptionsDir string
}
