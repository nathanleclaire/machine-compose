package persist

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/auth"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/drivers"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/engine"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/host"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/log"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/mcnerror"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/provider"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/swarm"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/version"
)

type Filestore struct {
	Path             string
	CaCertPath       string
	CaPrivateKeyPath string
}

func (s Filestore) getMachinesDir() string {
	return filepath.Join(s.Path, "machines")
}

func (s Filestore) Save(host *host.Host) error {
	data, err := json.Marshal(host)
	if err != nil {
		return err
	}

	hostPath := filepath.Join(s.getMachinesDir(), host.Name)

	if err := os.MkdirAll(hostPath, 0700); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(hostPath, "config.json"), data, 0600); err != nil {
		return err
	}

	return nil
}

func (s Filestore) Remove(name string, force bool) error {
	h, err := s.Get(name)
	if err != nil {
		return err
	}

	if err := h.Driver.Remove(); err != nil {
		if !force {
			return err
		}
	}

	hostPath := filepath.Join(s.getMachinesDir(), name)
	return os.RemoveAll(hostPath)
}

func (s Filestore) List() ([]*host.Host, error) {
	dir, err := ioutil.ReadDir(s.getMachinesDir())
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	hosts := []*host.Host{}

	for _, file := range dir {
		if file.IsDir() && !strings.HasPrefix(file.Name(), ".") {
			host, err := s.Get(file.Name())
			if err != nil {
				log.Errorf("error loading host %q: %s", file.Name(), err)
				continue
			}
			hosts = append(hosts, host)
		}
	}
	return hosts, nil
}

func (s Filestore) Exists(name string) (bool, error) {
	_, err := os.Stat(filepath.Join(s.getMachinesDir(), name))

	if os.IsNotExist(err) {
		return false, nil
	} else if err == nil {
		return true, nil
	}

	return false, err
}

func (s Filestore) loadConfig(h *host.Host) error {
	data, err := ioutil.ReadFile(filepath.Join(s.getMachinesDir(), h.Name, "config.json"))
	if err != nil {
		return err
	}

	// Remember the machine name so we don't have to pass it through each
	// struct in the migration.
	name := h.Name

	migratedHost, migrationPerformed, err := host.MigrateHost(h, data)
	if err != nil {
		return fmt.Errorf("Error getting migrated host: %s", err)
	}

	*h = *migratedHost

	h.Name = name

	// If we end up performing a migration, we should save afterwards so we don't have to do it again on subsequent invocations.
	if migrationPerformed {
		if err := s.Save(h); err != nil {
			return fmt.Errorf("Error saving config after migration was performed: %s", err)
		}
	}

	return nil

}

func (s Filestore) Get(name string) (*host.Host, error) {
	hostPath := filepath.Join(s.getMachinesDir(), name)

	if _, err := os.Stat(hostPath); os.IsNotExist(err) {
		return nil, mcnerror.ErrHostDoesNotExist{
			Name: name,
		}
	}

	host := &host.Host{
		Name: name,
	}

	if err := s.loadConfig(host); err != nil {
		return nil, err
	}

	return host, nil
}

func (s Filestore) GetActive() (*host.Host, error) {
	hosts, err := s.List()
	if err != nil {
		return nil, err
	}

	hostListItems := host.GetHostListItems(hosts)

	for _, item := range hostListItems {
		if item.Active {
			host, err := s.Get(item.Name)
			if err != nil {
				return nil, err
			}
			return host, nil
		}
	}

	return nil, errors.New("Active host not found")
}

func (s Filestore) GetProvider(name string) (provider.Provider, error) {
	return nil, nil
}

func (s Filestore) NewHost(driver drivers.Driver) (*host.Host, error) {
	certDir := filepath.Join(s.Path, "certs")

	hostOptions := &host.HostOptions{
		AuthOptions: &auth.AuthOptions{
			CertDir:          certDir,
			CaCertPath:       filepath.Join(certDir, "ca.pem"),
			CaPrivateKeyPath: filepath.Join(certDir, "ca-key.pem"),
			ClientCertPath:   filepath.Join(certDir, "cert.pem"),
			ClientKeyPath:    filepath.Join(certDir, "key.pem"),
			ServerCertPath:   filepath.Join(s.getMachinesDir(), "server.pem"),
			ServerKeyPath:    filepath.Join(s.getMachinesDir(), "server-key.pem"),
		},
		EngineOptions: &engine.EngineOptions{
			InstallURL:    "https://get.docker.com",
			StorageDriver: "aufs",
			TlsVerify:     true,
		},
		SwarmOptions: &swarm.SwarmOptions{
			Host:     "tcp://0.0.0.0:3376",
			Image:    "swarm:latest",
			Strategy: "spread",
		},
	}

	return &host.Host{
		ConfigVersion: version.ConfigVersion,
		Name:          driver.GetMachineName(),
		Driver:        driver,
		DriverName:    driver.DriverName(),
		HostOptions:   hostOptions,
	}, nil
}
