package errors

import (
	"fmt"
	"strings"
)

type FailedTaskErr struct {
	Hosts []string
}

func (e *FailedTaskErr) Error() string {
	return "task failed for hosts " + strings.Join(e.Hosts, ", ")
}

func NewFailedTaskErr(hosts []string) error {
	return &FailedTaskErr{Hosts: hosts}
}

var (
	// NoHostsErr is returned when no hosts are found
	NoHostsErr = fmt.Errorf("no hosts found")

	// NoClustersErr is returned when no clusters are found
	NoClustersErr = fmt.Errorf("no clusters found")
)
