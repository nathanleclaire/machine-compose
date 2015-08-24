package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/davecgh/go-spew/spew"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/drivers/digitalocean"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/drivers/drivermaker"
	dmlog "github.com/nathanleclaire/moby/Godeps/_workspace/src/github.com/docker/machine/libmachine/log"
	"github.com/nathanleclaire/moby/Godeps/_workspace/src/gopkg.in/yaml.v2"
)

type FancyTracker struct {
	r io.Reader
	w io.Writer
}

func (s *FancyTracker) Track() {
	scanner := bufio.NewScanner(s.r)
	for scanner.Scan() {
		fmt.Fprintln(s.w, "=>", scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error tracking writer:", err)
	}
}

type DriverWrapper struct {
	DriverOptions *digitalocean.Driver
}

func bail() {
	log.Fatal("Improper usage.  Usage: moby [up|apply]")
}

func main() {
	if len(os.Args) != 2 {
		bail()
	}

	libmachine.SetDebug(true)

	errR, errW := io.Pipe()
	errTracker := &FancyTracker{
		r: errR,
		w: os.Stderr,
	}

	outR, outW := io.Pipe()
	outTracker := &FancyTracker{
		r: outR,
		w: os.Stdout,
	}

	go errTracker.Track()
	go outTracker.Track()

	dmlog.SetOutWriter(errW)
	dmlog.SetErrWriter(outW)

	store := libmachine.GetDefaultStore()
	store.Path = "./store"

	hostName := "mobydick"

	data, err := ioutil.ReadFile("moby.yml")
	if err != nil {
		log.Fatal(err)
	}

	switch os.Args[1] {
	case "up":
		driver, err := drivermaker.NewDriver("digitalocean", hostName, "./store")
		if err != nil {
			log.Fatal(err)
		}

		h, err := store.NewHost(driver)
		if err != nil {
			log.Fatal(err)
		}

		if err := yaml.Unmarshal(data, h); err != nil {
			log.Fatal(err)
		}

		castedDriver, ok := driver.(*digitalocean.Driver)
		if !ok {
			log.Fatal("Fatal error, shoud be able to cast to driver type \"digitalocean\".")
		}

		driverWrapper := DriverWrapper{castedDriver}

		if err := yaml.Unmarshal(data, &driverWrapper); err != nil {
			log.Fatal(err)
		}

		h.Driver = driverWrapper.DriverOptions
		spew.Dump(h)

		if err := libmachine.Create(store, h); err != nil {
			log.Fatal(err)
		}
	case "apply":
		h, err := store.Get(hostName)
		if err != nil {
			log.Fatal(err)
		}

		spew.Dump(h)

		if err := yaml.Unmarshal(data, h); err != nil {
			log.Fatal(err)
		}

		if err := h.Provision(); err != nil {
			log.Fatal(err)
		}
	default:
		bail()
	}

}
