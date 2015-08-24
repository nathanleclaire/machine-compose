package persist

import (
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/drivers"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/host"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/provider"
)

type Store interface {
	// Exists returns whether a machine exists or not
	Exists(name string) (bool, error)

	// NewHost will initialize a new host machine
	NewHost(driver drivers.Driver) (*host.Host, error)

	// GetActive returns the active host
	GetActive() (*host.Host, error)

	// GetProvider returns the provider with the given name
	GetProvider(name string) (provider.Provider, error)

	// List returns a list of hosts
	List() ([]*host.Host, error)

	// Get loads a host by name
	Get(name string) (*host.Host, error)

	// Remove removes a machine from the store
	Remove(name string, force bool) error

	// Save persists a machine in the store
	Save(host *host.Host) error
}
