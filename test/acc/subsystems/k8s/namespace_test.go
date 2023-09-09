package k8s

import "testing"

func TestCreateNamespace(t *testing.T) {
	namespace := withK8sNamespace(t)
	cleanUpNamespace(t, namespace)
}
