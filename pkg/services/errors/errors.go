package errors

import (
	"fmt"
	"strings"
)

type HostsFailedErr struct {
	Hosts []string
}

func (e *HostsFailedErr) Error() string {
	return "hosts failed " + strings.Join(e.Hosts, ", ")
}

func NewHostsFailedErr(hosts []string) error {
	return &HostsFailedErr{Hosts: hosts}
}

var (
	// ErrNoHosts is returned when no hosts are found
	ErrNoHosts = fmt.Errorf("no hosts found")

	// ErrNoClusters is returned when no clusters are found
	ErrNoClusters = fmt.Errorf("no clusters found")
)
