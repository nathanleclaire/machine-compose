package drivermaker

import (
	"fmt"

	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/drivers"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/drivers/digitalocean"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/drivers/none"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/drivers/virtualbox"
)

func NewDriver(driverName, hostName, artifactPath string) (drivers.Driver, error) {
	var (
		driver drivers.Driver
	)

	switch driverName {
	case "virtualbox":
		driver = virtualbox.NewDriver(hostName, artifactPath)
	case "digitalocean":
		driver = digitalocean.NewDriver(hostName, artifactPath)
	case "none":
		driver = &none.Driver{}
	default:
		return nil, fmt.Errorf("Driver %q not recognized", driverName)
	}

	return driver, nil
}
