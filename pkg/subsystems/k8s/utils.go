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

func FindLabel(labels map[string]string, key, value string) bool {
	if labels == nil {
		return false
	}
	if v, ok := labels[key]; ok {
		return v == value
	}
	return false
}

func GetLabel(labels map[string]string, key string) string {
	if labels == nil {
		return ""
	}
	if v, ok := labels[key]; ok {
		return v
	}
	return ""
}
