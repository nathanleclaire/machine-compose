default: build

build:
	GOPATH=$(shell pwd)/Godeps/_workspace:$(GOPATH) go build
