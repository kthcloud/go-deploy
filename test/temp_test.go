package test

import (
	"errors"
	"testing"

	er "github.com/kthcloud/go-deploy/pkg/services/errors"
)

func TestList(t *testing.T) {
	if errors.Is(er.NewHostsFailedErr([]string{"t01n14", "t01n16"}), er.ErrNoHosts) {
		t.Fail()
	}
}
