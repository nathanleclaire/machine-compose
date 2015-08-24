package persist

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/drivers/none"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/host"
)

func cleanup() {
	os.RemoveAll(os.Getenv("MACHINE_STORAGE_PATH"))
}

func getTestStore() Filestore {
	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Setenv("MACHINE_STORAGE_PATH", tmpDir)

	return Filestore{
		Path:             tmpDir,
		CaCertPath:       filepath.Join(tmpDir, "certs", "ca-cert.pem"),
		CaPrivateKeyPath: filepath.Join(tmpDir, "certs", "ca-key.pem"),
	}
}

func TestStoreSave(t *testing.T) {
	defer cleanup()

	store := getTestStore()

	h, err := host.GetDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(store.getMachinesDir(), h.Name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Host path doesn't exist: %s", path)
	}
}

func TestStoreRemove(t *testing.T) {
	defer cleanup()

	store := getTestStore()

	h, err := host.GetDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(store.getMachinesDir(), h.Name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Host path doesn't exist: %s", path)
	}

	err = store.Remove(h.Name, false)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); err == nil {
		t.Fatalf("Host path still exists after remove: %s", path)
	}
}

func TestStoreList(t *testing.T) {
	defer cleanup()

	store := getTestStore()

	h, err := host.GetDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	hosts, err := store.List()
	if len(hosts) != 1 {
		t.Fatalf("List returned %d items, expected 1", len(hosts))
	}

	if hosts[0].Name != h.Name {
		t.Fatalf("hosts[0] name is incorrect, got: %s", hosts[0].Name)
	}
}

func TestStoreExists(t *testing.T) {
	defer cleanup()
	store := getTestStore()

	h, err := host.GetDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	exists, err := store.Exists(h.Name)
	if exists {
		t.Fatal("Host should not exist before saving")
	}

	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	exists, err = store.Exists(h.Name)
	if err != nil {
		t.Fatal(err)
	}

	if !exists {
		t.Fatal("Host should exist after saving")
	}

	if err := store.Remove(h.Name, true); err != nil {
		t.Fatal(err)
	}

	exists, err = store.Exists(h.Name)
	if err != nil {
		t.Fatal(err)
	}

	if exists {
		t.Fatal("Host should not exist after removing")
	}
}

func TestStoreLoad(t *testing.T) {
	defer cleanup()

	expectedURL := "unix:///foo/baz"
	flags := host.GetTestDriverFlags()
	flags.Data["url"] = expectedURL

	store := getTestStore()

	h, err := host.GetDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	if err := h.Driver.SetConfigFromFlags(flags); err != nil {
		t.Fatal(err)
	}

	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	h, err = store.Get(h.Name)
	if err != nil {
		log.Fatal(err)
	}

	actualURL, err := h.GetURL()
	if err != nil {
		t.Fatal(err)
	}

	if actualURL != expectedURL {
		t.Fatalf("GetURL is not %q, got %q", expectedURL, actualURL)
	}
}

func TestStoreGetSetActive(t *testing.T) {
	defer cleanup()

	store := getTestStore()

	// No host set
	h, err := store.GetActive()
	if err == nil {
		t.Fatal("Expected an error because there is no active host set")
	}

	if h != nil {
		t.Fatalf("GetActive: Active host should not exist")
	}

	h, err = host.GetDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	// Set normal host
	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	url, err := h.GetURL()
	if err != nil {
		t.Fatal(err)
	}

	os.Setenv("DOCKER_HOST", url)

	h, err = store.GetActive()
	if err != nil {
		t.Fatal(err)
	}
	if h.Name != host.HostTestName {
		t.Fatalf("Active host is not '%s', got %s", host.HostTestName, h.Name)
	}
}
