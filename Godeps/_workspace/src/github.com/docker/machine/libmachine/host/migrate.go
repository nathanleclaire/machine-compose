package host

import (
	"encoding/json"
	"fmt"

	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/drivers/drivermaker"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/version"
)

func getMigratedHostMetadata(data []byte) (*HostMetadata, error) {
	// HostMetadata is for a "first pass" so we can then load the driver
	var (
		hostMetadata *HostMetadataV0
	)

	if err := json.Unmarshal(data, &hostMetadata); err != nil {
		return &HostMetadata{}, err
	}

	migratedHostMetadata := MigrateHostMetadataV0ToHostMetadataV1(hostMetadata)

	return migratedHostMetadata, nil
}

func MigrateHost(h *Host, data []byte) (*Host, bool, error) {
	var (
		migrationPerformed = true
	)

	migratedHostMetadata, err := getMigratedHostMetadata(data)
	if err != nil {
		return &Host{}, false, err
	}

	// Don't need to specify store path here since it will be read from the data.
	driver, err := drivermaker.NewDriver(migratedHostMetadata.DriverName, h.Name, "")
	if err != nil {
		return &Host{}, false, err
	}

	for h.ConfigVersion = migratedHostMetadata.ConfigVersion; h.ConfigVersion <= version.ConfigVersion; h.ConfigVersion++ {
		switch h.ConfigVersion {
		case 0:
			hostV0 := &HostV0{
				Driver: driver,
			}
			if err := json.Unmarshal(data, &hostV0); err != nil {
				return &Host{}, migrationPerformed, fmt.Errorf("Error unmarshalling host config version 0: %s", err)
			}
			h = MigrateHostV0ToHostV1(hostV0)
		default:
			migrationPerformed = false
		}
	}

	h.Driver = driver
	if err := json.Unmarshal(data, &h); err != nil {
		return &Host{}, migrationPerformed, fmt.Errorf("Error unmarshalling most recent host version: %s", err)
	}

	return h, migrationPerformed, nil
}
