package k8s

import (
	"errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func isNotFoundError(err error) bool {
	statusError := &k8sErrors.StatusError{}
	if errors.As(err, &statusError) {
		if statusError.Status().Reason == metav1.StatusReasonNotFound {
			return true
		}
	}
	return false
}
