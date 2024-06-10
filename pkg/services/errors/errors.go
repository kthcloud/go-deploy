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
	// NoHostsErr is returned when no hosts are found
	NoHostsErr = fmt.Errorf("no hosts found")

	// NoClustersErr is returned when no clusters are found
	NoClustersErr = fmt.Errorf("no clusters found")
)
