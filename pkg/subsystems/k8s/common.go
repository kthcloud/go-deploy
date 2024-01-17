package k8s

import (
	"errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// IsNotFoundErr returns true if the error is a Kubernetes NotFound error.
func IsNotFoundErr(err error) bool {
	var statusError *k8sErrors.StatusError
	if errors.As(err, &statusError) {
		if statusError.ErrStatus.Reason == v1.StatusReasonNotFound {
			return true
		}
	}

	return false
}

// IsImmutabilityErr returns true if the error is a Kubernetes Immutability error.
func IsImmutabilityErr(err error) bool {
	return strings.Contains(err.Error(), "field is immutable")
}
