package k8s

import (
	"context"
	"deploy-api-go/utils/subsystemutils"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
)

func deleteNamespace(name string) error {
	prexfixedName := subsystemutils.GetPrefixedName(name)
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s namespace %s. details: %s", prexfixedName, err)
	}

	namespaceDelted, err := deletedNamespace(prexfixedName)
	if err != nil {
		return makeError(err)
	}

	if namespaceDelted {
		return nil
	}

	err = client.CoreV1().Namespaces().Delete(context.TODO(), prexfixedName, metav1.DeleteOptions{})
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Delete(name string) error {
	log.Println("deleting k8s for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete k8s setup for project %s. details: %s", name, err)
	}

	err := deleteNamespace(name)
	if err != nil {
		return makeError(err)
	}

	return nil
}
