package k8s

import "testing"

func TestCreateDeployment(t *testing.T) {
	namespace := withK8sNamespace(t)
	deployment := withK8sDeployment(t, namespace)
	cleanUpDeployment(t, namespace, deployment)
	cleanUpNamespace(t, namespace)
}
