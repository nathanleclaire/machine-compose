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

type DriverWrapper struct {
	DriverOptions *virtualbox.Driver
}

func bail() {
	fmt.Println("Improper usage.  Usage: moby [up|apply]")
	os.Exit(1)
}

func main() {
	if len(os.Args) != 2 {
		bail()
	}

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

	hostName := "mobydick"

	data, err := ioutil.ReadFile("moby.yml")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "up":
		driver := virtualbox.NewDriver(hostName, "/tmp/moby")

		h, err := store.NewHost(driver)
		if err != nil {
			log.Fatal(err)
		}

		driverWrapper := DriverWrapper{driver}

		if err := yaml.Unmarshal(data, h); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if err := yaml.Unmarshal(data, &driverWrapper); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		h.Driver = driverWrapper.DriverOptions
		spew.Dump(h)

		if err := libmachine.Create(store, h); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "apply":
		h, err := store.Get(hostName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		spew.Dump(h)

		if err := yaml.Unmarshal(data, h); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if err := h.Provision(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		bail()
	}

}
