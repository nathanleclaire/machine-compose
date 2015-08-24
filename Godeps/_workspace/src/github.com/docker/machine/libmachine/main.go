package libmachine

import (
	"fmt"
	"path/filepath"

	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/cert"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/drivers"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/host"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/log"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/mcnerror"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/mcnutils"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/persist"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/provision"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/state"
)

func GetDefaultStore() *persist.Filestore {
	homeDir := mcnutils.GetHomeDir()
	certsDir := filepath.Join(homeDir, ".docker", "machine", "certs")
	return &persist.Filestore{
		Path:             homeDir,
		CaCertPath:       certsDir,
		CaPrivateKeyPath: certsDir,
	}
}

// Create is the wrapper method which covers all of the boilerplate around
// actually creating, provisioning, and persisting an instance in the store.
func Create(store persist.Store, h *host.Host) error {
	if err := cert.BootstrapCertificates(h.HostOptions.AuthOptions); err != nil {
		return fmt.Errorf("Error generating certificates: %s", err)
	}

	validName := host.ValidateHostName(h.Name)
	if !validName {
		return mcnerror.ErrInvalidHostname
	}

	exists, err := store.Exists(h.Name)
	if err != nil {
		return fmt.Errorf("Error checking if host exists: %s", err)
	}
	if exists {
		return mcnerror.ErrHostAlreadyExists{h.Name}
	}

	if err := h.Driver.PreCreateCheck(); err != nil {
		return fmt.Errorf("Error with pre-create check: %s", err)
	}

	if err := store.Save(h); err != nil {
		return fmt.Errorf("Error saving host to store before attempting creation: %s", err)
	}

	if err := h.Driver.Create(); err != nil {
		return fmt.Errorf("Error in driver during machine creation: %s", err)
	}

	if err := store.Save(h); err != nil {
		return fmt.Errorf("Error saving host to store after attempting creation: %s", err)
	}

	// TODO: Not really a fan of just checking "none" here.
	if h.Driver.DriverName() != "none" {
		if err := mcnutils.WaitFor(drivers.MachineInState(h.Driver, state.Running)); err != nil {
			return fmt.Errorf("Error waiting for machine to be running: %s", err)
		}

		if err := host.WaitForSSH(h); err != nil {
			return fmt.Errorf("Error waiting for SSH: %s", err)
		}

		provisioner, err := provision.DetectProvisioner(h.Driver)
		if err != nil {
			return fmt.Errorf("Error detecting OS: %s", err)
		}

		if err := provisioner.Provision(*h.HostOptions.SwarmOptions, *h.HostOptions.AuthOptions, *h.HostOptions.EngineOptions); err != nil {
			return fmt.Errorf("Error running provisioning: %s", err)
		}
	}

	return nil
}

func SetDebug(val bool) {
	log.IsDebug = val
}
