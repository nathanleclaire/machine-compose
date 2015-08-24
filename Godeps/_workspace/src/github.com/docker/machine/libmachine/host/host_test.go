package host

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/drivers/fakedriver"
	_ "github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/drivers/none"

	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/state"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/swarm"
	"github.com/stretchr/testify/assert"
)

var (
	hostTestStorePath string
	stdout            *os.File
)

func init() {
	stdout = os.Stdout
}

func cleanup() {
	os.Stdout = stdout
	os.RemoveAll(hostTestStorePath)
}

func TestValidateHostnameValid(t *testing.T) {
	hosts := []string{
		"zomg",
		"test-ing",
		"some.h0st",
	}

	for _, v := range hosts {
		isValid := ValidateHostName(v)
		if !isValid {
			t.Fatalf("Thought a valid hostname was invalid: %s", v)
		}
	}
}

func TestValidateHostnameInvalid(t *testing.T) {
	hosts := []string{
		"zom_g",
		"test$ing",
		"some😄host",
	}

	for _, v := range hosts {
		isValid := ValidateHostName(v)
		if isValid {
			t.Fatalf("Thought an invalid hostname was valid: %s", v)
		}
	}
}

func TestPrintIPEmptyGivenLocalEngine(t *testing.T) {
	defer cleanup()
	host, _ := GetDefaultTestHost()

	out, w := captureStdout()

	assert.Nil(t, host.PrintIP())
	w.Close()

	assert.Equal(t, "", strings.TrimSpace(<-out))
}

func TestPrintIPPrintsGivenRemoteEngine(t *testing.T) {
	defer cleanup()
	host, _ := GetDefaultTestHost()
	host.Driver = &fakedriver.FakeDriver{}

	out, w := captureStdout()

	assert.Nil(t, host.PrintIP())

	w.Close()

	assert.Equal(t, "1.2.3.4", strings.TrimSpace(<-out))
}

func captureStdout() (chan string, *os.File) {
	r, w, _ := os.Pipe()
	os.Stdout = w

	out := make(chan string)

	go func() {
		var testOutput bytes.Buffer
		io.Copy(&testOutput, r)
		out <- testOutput.String()
	}()

	return out, w
}

func TestGetHostListItems(t *testing.T) {
	defer cleanup()

	hostListItemsChan := make(chan HostListItem)

	hosts := []Host{
		{
			Name:       "foo",
			DriverName: "fakedriver",
			Driver: &fakedriver.FakeDriver{
				MockState: state.Running,
			},
			HostOptions: &HostOptions{
				SwarmOptions: &swarm.SwarmOptions{
					Master:    false,
					Address:   "",
					Discovery: "",
				},
			},
		},
		{
			Name:       "bar",
			DriverName: "fakedriver",
			Driver: &fakedriver.FakeDriver{
				MockState: state.Stopped,
			},
			HostOptions: &HostOptions{
				SwarmOptions: &swarm.SwarmOptions{
					Master:    false,
					Address:   "",
					Discovery: "",
				},
			},
		},
		{
			Name:       "baz",
			DriverName: "fakedriver",
			Driver: &fakedriver.FakeDriver{
				MockState: state.Running,
			},
			HostOptions: &HostOptions{
				SwarmOptions: &swarm.SwarmOptions{
					Master:    false,
					Address:   "",
					Discovery: "",
				},
			},
		},
	}

	expected := map[string]state.State{
		"foo": state.Running,
		"bar": state.Stopped,
		"baz": state.Running,
	}

	items := []HostListItem{}
	for _, host := range hosts {
		go getHostState(host, hostListItemsChan)
	}

	for i := 0; i < len(hosts); i++ {
		items = append(items, <-hostListItemsChan)
	}

	for _, item := range items {
		if expected[item.Name] != item.State {
			t.Fatal("Expected state did not match for item", item)
		}
	}
}
