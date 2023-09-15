package k8s

import "testing"

func TestCreateIngress(t *testing.T) {
	namespace := withK8sNamespace(t)
	service := withK8sService(t, namespace)
	ingress := withK8sIngress(t, namespace, service)
	cleanUpIngress(t, namespace, ingress)
	cleanUpService(t, namespace, service)
	cleanUpNamespace(t, namespace)
}
