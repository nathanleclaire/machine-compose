package mcnerror

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidHostname     = errors.New("Invalid hostname specified")
	ErrUnknownProviderType = errors.New("Unknown hypervisor type")
)

type ErrHostDoesNotExist struct {
	Name string
}

func (e ErrHostDoesNotExist) Error() string {
	return fmt.Sprintf("Error: Host does not exist: %s", e.Name)
}

type ErrHostAlreadyExists struct {
	Name string
}

func (e ErrHostAlreadyExists) Error() string {
	return fmt.Sprintf("Error: Host already exists: %s", e.Name)
}
