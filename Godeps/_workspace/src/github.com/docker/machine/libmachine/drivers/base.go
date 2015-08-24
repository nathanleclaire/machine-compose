package drivers

import "path/filepath"

// BaseDriver - Embed this struct into drivers to provide the common set
// of fields and functions.
type BaseDriver struct {
	IPAddress      string
	SSHUser        string
	SSHPort        int
	MachineName    string
	SwarmMaster    bool
	SwarmHost      string
	SwarmDiscovery string
	ArtifactPath   string
}

// GetSSHKeyPath -
func (d *BaseDriver) GetSSHKeyPath() string {
	return filepath.Join(d.ArtifactPath, "machines", d.MachineName, "id_rsa")
}

// AuthorizePort -
func (d *BaseDriver) AuthorizePort(ports []*Port) error {
	return nil
}

// DeauthorizePort -
func (d *BaseDriver) DeauthorizePort(ports []*Port) error {
	return nil
}

// DriverName - This must be implemented in every driver
func (d *BaseDriver) DriverName() string {
	return "unknown"
}

// GetMachineName -
func (d *BaseDriver) GetMachineName() string {
	return d.MachineName
}

// LocalArtifactPath -
func (d *BaseDriver) LocalArtifactPath(file string) string {
	return filepath.Join(d.ArtifactPath, "machines", d.MachineName, file)
}

// GlobalArtifactPath -
func (d *BaseDriver) GlobalArtifactPath() string {
	return d.ArtifactPath
}

// GetSSHPort -
func (d *BaseDriver) GetSSHPort() (int, error) {
	if d.SSHPort == 0 {
		d.SSHPort = 22
	}

	return d.SSHPort, nil
}

// GetSSHUsername -
func (d *BaseDriver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "root"
	}

	return d.SSHUser
}
