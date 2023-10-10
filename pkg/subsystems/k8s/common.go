package k8s

import (
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func IsNotFoundErr(err error) bool {
	if statusError, ok := err.(*k8sErrors.StatusError); ok {
		if statusError.ErrStatus.Reason == v1.StatusReasonNotFound {
			return true
		}
	}

	return false
}
