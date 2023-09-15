package k8s

import "testing"

func TestCreateService(t *testing.T) {
	namespace := withK8sNamespace(t)
	service := withK8sService(t, namespace)
	cleanUpService(t, namespace, service)
	cleanUpNamespace(t, namespace)
}
