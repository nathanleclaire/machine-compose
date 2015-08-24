package host

import (
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/auth"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/engine"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/swarm"
)

// In the 0.0.1 => 0.0.2 transition, the JSON representation of
// machines changed from a "flat" to a more "nested" structure
// for various options and configuration settings.  To preserve
// compatibility with existing machines, these migration functions
// have been introduced.  They preserve backwards compat at the expense
// of some duplicated information.

// validates host config and modifies if needed
// this is used for configuration updates
func MigrateHostV0ToHostV1(hostV0 *HostV0) *Host {
	hostV1 := &Host{}

	hostV1.HostOptions = &HostOptions{}
	hostV1.HostOptions.EngineOptions = &engine.EngineOptions{}
	hostV1.HostOptions.SwarmOptions = &swarm.SwarmOptions{
		Address:   "",
		Discovery: hostV0.SwarmDiscovery,
		Host:      hostV0.SwarmHost,
		Master:    hostV0.SwarmMaster,
	}
	hostV1.HostOptions.AuthOptions = &auth.AuthOptions{
		StorePath:            hostV0.StorePath,
		CaCertPath:           hostV0.CaCertPath,
		CaCertRemotePath:     "",
		ServerCertPath:       hostV0.ServerCertPath,
		ServerKeyPath:        hostV0.ServerKeyPath,
		ClientKeyPath:        hostV0.ClientKeyPath,
		ServerCertRemotePath: "",
		ServerKeyRemotePath:  "",
		CaPrivateKeyPath:     hostV0.CaPrivateKeyPath,
		ClientCertPath:       hostV0.ClientCertPath,
	}

	return hostV1
}

// fills nested host metadata and modifies if needed
// this is used for configuration updates
func MigrateHostMetadataV0ToHostMetadataV1(m *HostMetadataV0) *HostMetadata {
	hostMetadata := &HostMetadata{}
	hostMetadata.DriverName = m.DriverName
	hostMetadata.HostOptions.EngineOptions = &engine.EngineOptions{}
	hostMetadata.HostOptions.AuthOptions = &auth.AuthOptions{
		StorePath:            m.StorePath,
		CaCertPath:           m.CaCertPath,
		CaCertRemotePath:     "",
		ServerCertPath:       m.ServerCertPath,
		ServerKeyPath:        m.ServerKeyPath,
		ClientKeyPath:        "",
		ServerCertRemotePath: "",
		ServerKeyRemotePath:  "",
		CaPrivateKeyPath:     m.CaPrivateKeyPath,
		ClientCertPath:       m.ClientCertPath,
	}

	hostMetadata.ConfigVersion = m.ConfigVersion

	return hostMetadata
}
