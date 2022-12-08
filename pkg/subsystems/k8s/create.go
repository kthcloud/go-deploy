package k8s

import (
	"context"
	"fmt"
	"go-deploy/utils/subsystemutils"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createNamespace(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to k8s namespace %s. details: %s", name, err)
	}

	namespaceCreated, err := createdNamespace(subsystemutils.GetPrefixedName(name))
	if err != nil {
		return makeError(err)
	}

	if namespaceCreated {
		return nil
	}

	namespace := getNamespaceManifest(name)

	_, err = client.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
	if err != nil {
		return makeError(err)
	}

	return err
}

func createDeployment(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to k8s deployment %s. details: %s", name, err)
	}
	prexfixedName := subsystemutils.GetPrefixedName(name)

	deploymentCreated, err := createdDeployment(name)
	if err != nil {
		return makeError(err)
	}

	if deploymentCreated {
		return nil
	}

	namespaceCreated, err := createdNamespace(prexfixedName)
	if err != nil {
		return makeError(err)
	}

	if !namespaceCreated {
		err = fmt.Errorf("no namespace created")
		return makeError(err)
	}

	deployment := getDeploymentManifest(name)

	_, err = client.AppsV1().Deployments(prexfixedName).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return makeError(err)
	}

	return nil
}

func createService(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to k8s service %s. details: %s", name, err)
	}
	prexfixedName := subsystemutils.GetPrefixedName(name)

	serviceCreated, err := createdService(name)
	if err != nil {
		return makeError(err)
	}

	if serviceCreated {
		return nil
	}

	namespaceCreated, err := createdNamespace(prexfixedName)
	if err != nil {
		return makeError(err)
	}

	if !namespaceCreated {
		err = fmt.Errorf("no namespace created")
		return makeError(err)
	}

	service := getServiceManifest(name)

	_, err = client.CoreV1().Services(prexfixedName).Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Create(name string) error {
	log.Println("creating k8s for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to create k8s setup for deployment %s. details: %s", name, err)
	}

	err := createNamespace(name)
	if err != nil {
		return makeError(err)
	}

	err = createDeployment(name)
	if err != nil {
		return makeError(err)
	}

	err = createService(name)
	if err != nil {
		return makeError(err)
	}

	return nil
}
