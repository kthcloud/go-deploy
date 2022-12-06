package k8s

import (
	"context"
	"errors"
	"fmt"
	"go-deploy/utils/subsystemutils"
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

func createdNamespace(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if k8s namespace %s is created. details: %s", name, err)
	}

	namespace, err := client.CoreV1().Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, makeError(err)
	}
	return namespace != nil, err
}

func createdDeployment(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if k8s deployment %s is created. details: %s", name, err)
	}

	deployment, err := client.AppsV1().Deployments(subsystemutils.GetPrefixedName(name)).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, makeError(err)
	}
	return deployment != nil, err
}

func createdService(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if k8s service %s is created. details: %s", name, err)
	}

	service, err := client.CoreV1().Services(subsystemutils.GetPrefixedName(name)).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, makeError(err)
	}
	return service != nil, err
}

func Created(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if k8s setup is created for project %s. details: %s", name, err)
	}

	namespaceCreated, err := createdNamespace(subsystemutils.GetPrefixedName(name))
	if err != nil {
		return false, makeError(err)
	}

	if !namespaceCreated {
		return false, nil
	}

	deploymentCreated, err := createdDeployment(name)
	if err != nil {
		return false, makeError(err)
	}

	serviceCreated, err := createdService(name)
	if err != nil {
		return false, makeError(err)
	}

	return deploymentCreated && serviceCreated, nil
}
