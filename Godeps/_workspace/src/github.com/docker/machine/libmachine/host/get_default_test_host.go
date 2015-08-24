package host

import (
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/auth"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/drivers"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/drivers/none"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/engine"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/swarm"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/version"
)

const (
	HostTestName       = "test-host"
	hostTestDriverName = "none"
	hostTestCaCert     = "test-cert"
	hostTestPrivateKey = "test-key"
)

// Helper functions for tests meant to exported and used by other packages as
// well.
func GetTestDriverFlags() drivers.DriverOptionsMock {
	name := HostTestName
	flags := drivers.DriverOptionsMock{
		Data: map[string]interface{}{
			"name":            name,
			"url":             "unix:///var/run/docker.sock",
			"swarm":           false,
			"swarm-host":      "",
			"swarm-master":    false,
			"swarm-discovery": "",
		},
	}
	return flags
}

func GetDefaultTestHost() (*Host, error) {
	hostOptions := &HostOptions{
		EngineOptions: &engine.EngineOptions{},
		SwarmOptions:  &swarm.SwarmOptions{},
		AuthOptions: &auth.AuthOptions{
			CaCertPath:       hostTestCaCert,
			CaPrivateKeyPath: hostTestPrivateKey,
		},
	}

	driver := none.NewDriver(HostTestName, "/tmp/artifacts")

	host := &Host{
		ConfigVersion: version.ConfigVersion,
		Name:          HostTestName,
		Driver:        &driver,
		DriverName:    "none",
		HostOptions:   hostOptions,
	}

	flags := GetTestDriverFlags()
	if err := host.Driver.SetConfigFromFlags(flags); err != nil {
		return nil, err
	}

	return host, nil
}
