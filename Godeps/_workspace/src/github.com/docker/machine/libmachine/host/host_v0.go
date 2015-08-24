package host

import "github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/drivers"

type HostV0 struct {
	Name          string `json:"-"`
	Driver        drivers.Driver
	DriverName    string
	ConfigVersion int
	HostOptions   *HostOptions

	StorePath        string
	CaCertPath       string
	CaPrivateKeyPath string
	ServerCertPath   string
	ServerKeyPath    string
	ClientCertPath   string
	SwarmHost        string
	SwarmMaster      bool
	SwarmDiscovery   string
	ClientKeyPath    string
}

type HostMetadataV0 struct {
	HostOptions   HostOptions
	DriverName    string
	ConfigVersion int

	StorePath        string
	CaCertPath       string
	CaPrivateKeyPath string
	ServerCertPath   string
	ServerKeyPath    string
	ClientCertPath   string
}
