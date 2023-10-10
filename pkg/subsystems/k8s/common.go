package k8s

import (
	"errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func IsNotFoundErr(err error) bool {
	var statusError *k8sErrors.StatusError
	if errors.As(err, &statusError) {
		if statusError.ErrStatus.Reason == v1.StatusReasonNotFound {
			return true
		}
	}

	return false
}
