package host

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/auth"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/drivers"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/engine"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/log"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/mcnutils"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/provision"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/ssh"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/state"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/swarm"
)

var (
	validHostNameChars                = `[a-zA-Z0-9\-\.]`
	validHostNamePattern              = regexp.MustCompile(`^` + validHostNameChars + `+$`)
	errMachineMustBeRunningForUpgrade = errors.New("Error: machine must be running to upgrade.")
	stateTimeoutDuration              = time.Second * 3
)

type Host struct {
	ConfigVersion int
	Driver        drivers.Driver
	DriverName    string
	HostOptions   *HostOptions
	Name          string
}

type HostOptions struct {
	Driver        string
	Memory        int
	Disk          int
	EngineOptions *engine.EngineOptions
	SwarmOptions  *swarm.SwarmOptions
	AuthOptions   *auth.AuthOptions
}

type HostMetadata struct {
	ConfigVersion int
	DriverName    string
	HostOptions   HostOptions
}

type HostListItem struct {
	Name         string
	Active       bool
	DriverName   string
	State        state.State
	URL          string
	SwarmOptions swarm.SwarmOptions
}

func ValidateHostName(name string) bool {
	return validHostNamePattern.MatchString(name)
}

func (h *Host) RunSSHCommand(command string) (string, error) {
	return drivers.RunSSHCommandFromDriver(h.Driver, command)
}

func (h *Host) CreateSSHClient() (ssh.Client, error) {
	addr, err := h.Driver.GetSSHHostname()
	if err != nil {
		return ssh.ExternalClient{}, err
	}

	port, err := h.Driver.GetSSHPort()
	if err != nil {
		return ssh.ExternalClient{}, err
	}

	auth := &ssh.Auth{
		Keys: []string{h.Driver.GetSSHKeyPath()},
	}

	return ssh.NewClient(h.Driver.GetSSHUsername(), addr, port, auth)
}

func (h *Host) CreateSSHShell() error {
	client, err := h.CreateSSHClient()
	if err != nil {
		return err
	}

	return client.Shell()
}

func (h *Host) runActionForState(action func() error, desiredState state.State) error {
	if drivers.MachineInState(h.Driver, desiredState)() {
		return fmt.Errorf("Machine %q is already %s.", h.Name, strings.ToLower(desiredState.String()))
	}

	if err := action(); err != nil {
		return err
	}

	return mcnutils.WaitFor(drivers.MachineInState(h.Driver, desiredState))
}

func (h *Host) Start() error {
	return h.runActionForState(h.Driver.Start, state.Running)
}

func (h *Host) Stop() error {
	return h.runActionForState(h.Driver.Stop, state.Stopped)
}

func (h *Host) Kill() error {
	return h.runActionForState(h.Driver.Kill, state.Stopped)
}

func (h *Host) Restart() error {
	if drivers.MachineInState(h.Driver, state.Running)() {
		if err := h.Stop(); err != nil {
			return err
		}

		if err := mcnutils.WaitFor(drivers.MachineInState(h.Driver, state.Stopped)); err != nil {
			return err
		}
	}

	if err := h.Start(); err != nil {
		return err
	}

	if err := mcnutils.WaitFor(drivers.MachineInState(h.Driver, state.Running)); err != nil {
		return err
	}

	return nil
}

func (h *Host) Upgrade() error {
	machineState, err := h.Driver.GetState()
	if err != nil {
		return err
	}

	if machineState != state.Running {
		log.Fatal(errMachineMustBeRunningForUpgrade)
	}

	provisioner, err := provision.DetectProvisioner(h.Driver)
	if err != nil {
		return err
	}

	if err := provisioner.Package("docker", pkgaction.Upgrade); err != nil {
		return err
	}

	if err := provisioner.Service("docker", pkgaction.Restart); err != nil {
		return err
	}
	return nil
}

func (h *Host) GetURL() (string, error) {
	return h.Driver.GetURL()
}

// IsActive provides a single method for determining if a host is active based
// on both the url and if the host is stopped.
func (h *Host) IsActive() (bool, error) {
	currentState, err := h.Driver.GetState()

	if err != nil {
		log.Errorf("error getting state for host %s: %s", h.Name, err)
		return false, err
	}

	url, err := h.GetURL()

	if err != nil {
		if err == drivers.ErrHostIsNotRunning {
			url = ""
		} else {
			log.Errorf("error getting URL for host %s: %s", h.Name, err)
			return false, err
		}
	}

	dockerHost := os.Getenv("DOCKER_HOST")

	notStopped := currentState != state.Stopped
	correctURL := url == dockerHost

	isActive := notStopped && correctURL

	return isActive, nil
}

func (h *Host) getProvisioner() (provision.Provisioner, error) {
	provisioner, err := provision.DetectProvisioner(h.Driver)
	if err != nil {
		return nil, fmt.Errorf("Error getting provisioner: %s", err)
	}

	return provisioner, nil
}

func (h *Host) ConfigureAuth() error {
	provisioner, err := h.getProvisioner()
	if err != nil {
		return err
	}

	// TODO: This is kind of a hack (or is it?  I'm not really sure until
	// we have more clearly defined outlook on what the responsibilities
	// and modularity of the provisioners should be).
	//
	// Call provision to re-provision the certs properly.
	if err := provisioner.Provision(swarm.SwarmOptions{}, *h.HostOptions.AuthOptions, *h.HostOptions.EngineOptions); err != nil {
		return err
	}

	return nil
}

func (h *Host) Provision() error {
	provisioner, err := h.getProvisioner()
	if err != nil {
		return err
	}

	log.Info("Running Docker Machine provisioning on host %s...", h.Name)

	if err := provisioner.Provision(*h.HostOptions.SwarmOptions, *h.HostOptions.AuthOptions, *h.HostOptions.EngineOptions); err != nil {
		return err
	}

	return nil
}

func (h *Host) PrintIP() error {
	if ip, err := h.Driver.GetIP(); err != nil {
		return err
	} else {
		fmt.Println(ip)
	}
	return nil
}

func WaitForSSH(h *Host) error {
	return drivers.WaitForSSH(h.Driver)
}

func attemptGetHostState(host Host, stateQueryChan chan<- HostListItem) {
	currentState, err := host.Driver.GetState()
	if err != nil {
		log.Errorf("error getting state for host %s: %s", host.Name, err)
	}

	url, err := host.GetURL()
	if err != nil {
		if err == drivers.ErrHostIsNotRunning {
			url = ""
		} else {
			log.Errorf("error getting URL for host %s: %s", host.Name, err)
		}
	}

	isActive, err := host.IsActive()
	if err != nil {
		log.Errorf("error determining if host is active for host %s: %s",
			host.Name, err)
	}

	stateQueryChan <- HostListItem{
		Name:         host.Name,
		Active:       isActive,
		DriverName:   host.Driver.DriverName(),
		State:        currentState,
		URL:          url,
		SwarmOptions: *host.HostOptions.SwarmOptions,
	}
}

func getHostState(host Host, hostListItemsChan chan<- HostListItem) {
	// This channel is used to communicate the properties we are querying
	// about the host in the case of a successful read.
	stateQueryChan := make(chan HostListItem)

	go attemptGetHostState(host, stateQueryChan)

	select {
	// If we get back useful information, great.  Forward it straight to
	// the original parent channel.
	case hli := <-stateQueryChan:
		hostListItemsChan <- hli

	// Otherwise, give up after a predetermined duration.
	case <-time.After(stateTimeoutDuration):
		hostListItemsChan <- HostListItem{
			Name:       host.Name,
			DriverName: host.Driver.DriverName(),
			State:      state.Timeout,
		}
	}
}

func GetHostListItems(hostList []*Host) []HostListItem {
	hostListItems := []HostListItem{}
	hostListItemsChan := make(chan HostListItem)

	for _, host := range hostList {
		go getHostState(*host, hostListItemsChan)
	}

	for range hostList {
		hostListItems = append(hostListItems, <-hostListItemsChan)
	}

	close(hostListItemsChan)
	return hostListItems
}
