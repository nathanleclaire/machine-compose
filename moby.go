package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers/virtualbox"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"gopkg.in/yaml.v2"
)

type FancyWriter struct {
	bufWriter *bufio.Writer
}

func NewFancyWriter(w io.Writer) (*FancyWriter, error) {
	f := &FancyWriter{bufio.NewWriter(w)}

	_, err := f.bufWriter.WriteString("=> ")
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (f FancyWriter) Write(p []byte) (int, error) {
	n := 0
	for _, b := range p {
		if err := f.bufWriter.WriteByte(b); err != nil {
			return n, err
		}
		if b == '\n' {
			if err := f.bufWriter.Flush(); err != nil {
				return n, err
			}
			n, err := f.bufWriter.WriteString("=> ")
			if err != nil {
				return n, err
			}
		}
		n++
	}

	return n, nil
}

func bail() {
	fmt.Println("Improper usage.  Usage: moby up")
	os.Exit(1)
}

func main() {
	libmachine.SetDebug(true)

	fwout, err := NewFancyWriter(os.Stdout)
	if err != nil {
		bail()
	}
	log.SetOutWriter(fwout)

	fwerr, err := NewFancyWriter(os.Stderr)
	if err != nil {
		bail()
	}
	log.SetErrWriter(fwerr)

	store := libmachine.GetDefaultStore()
	store.Path = "/tmp/moby"

	driver := virtualbox.NewDriver("mobydick", "/tmp/moby")

	h, err := store.NewHost(driver)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) != 2 {
		bail()
	}

	if os.Args[1] == "up" {
		data, err := ioutil.ReadFile("moby.yml")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		unmarshalHost := host.Host{}

		if err := yaml.Unmarshal(data, &unmarshalHost); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		spew.Dump(unmarshalHost)

		h.HostOptions = unmarshalHost.HostOptions

		spew.Dump(h)

		os.Exit(0)

		if err := libmachine.Create(store, h); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		bail()
	}
}
