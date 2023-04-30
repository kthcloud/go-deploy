package test

import (
	"github.com/google/uuid"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/pkg/subsystems/k8s/models"
	"testing"
)

func TestCreateGoodK8s(t *testing.T) {
	setup(t)

	client, err := k8s.New(&k8s.ClientConf{
		K8sAuth: conf.Env.K8s.Config,
	})

	if err != nil {
		t.Fatalf(err.Error())
	}

	namespacePublic := &models.NamespacePublic{
		Name: "acc-test-" + uuid.New().String(),
	}

	servicePublic := &models.ServicePublic{
		Name:       "acc-test",
		Namespace:  namespacePublic.Name,
		Port:       11111,
		TargetPort: 22222,
	}

	deploymentPublic := &models.DeploymentPublic{
		Name:        "acc-test",
		Namespace:   namespacePublic.Name,
		DockerImage: conf.Env.DockerRegistry.Placeholder,
		EnvVars:     nil,
	}

	// Create phase

	err = client.CreateNamespace(namespacePublic)
	if err != nil {
		t.Fatalf(err.Error())
	}

	id, err := client.CreateDeployment(deploymentPublic)
	if err != nil {
		t.Fatalf(err.Error())
	}

	deploymentPublic.ID = id

	if len(deploymentPublic.ID) == 0 {
		t.Fatalf("no deployment id received from client")
	}

	id, err = client.CreateService(servicePublic)
	if err != nil {
		t.Fatalf(err.Error())
	}

	servicePublic.ID = id

	if len(servicePublic.ID) == 0 {
		t.Fatalf("no service id received from client")
	}

	// Delete phase

	err = client.DeleteNamespace(namespacePublic.Name)
	if err != nil {
		t.Fatalf(err.Error())
	}

	deletedProject, err := client.NamespaceDeleted(namespacePublic.Name)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if !deletedProject {
		t.Fatalf("failed to delete namespace")
	}
}

func TestCreateBadK8s(t *testing.T) {
	setup(t)

	client, err := k8s.New(&k8s.ClientConf{
		K8sAuth: conf.Env.K8s.Config,
	})

	if err != nil {
		t.Fatalf(err.Error())
	}

	namespacePublic := &models.NamespacePublic{
		Name: "acc-test-" + uuid.New().String(),
	}

	deploymentPublic := &models.DeploymentPublic{
		Name:        "acc-test",
		Namespace:   namespacePublic.Name,
		DockerImage: conf.Env.DockerRegistry.Placeholder,
		EnvVars:     nil,
	}

	deploymentPublicBad := &models.DeploymentPublic{
		Name:        "acc_test",
		Namespace:   namespacePublic.Name,
		DockerImage: conf.Env.DockerRegistry.Placeholder,
		EnvVars:     nil,
	}

	servicePublic := &models.ServicePublic{
		Name:       "acc-test",
		Namespace:  namespacePublic.Name,
		Port:       11111,
		TargetPort: 22222,
	}

	servicePublicBad := &models.ServicePublic{
		Name:       "acc_test",
		Namespace:  namespacePublic.Name,
		Port:       -1000,
		TargetPort: 0,
	}

	// Create phase

	// create before namespace exists is invalid
	id, err := client.CreateDeployment(deploymentPublic)
	if err == nil {
		t.Fatalf("creating a deployment without a namespace did not result in an error")
	}

	id, err = client.CreateService(servicePublic)
	if err == nil {
		t.Fatalf("creating a service without a namespace did not result in an error")
	}

	// create namespace twice
	err = client.CreateNamespace(namespacePublic)
	if err != nil {
		t.Fatalf(err.Error())
	}

	err = client.CreateNamespace(namespacePublic)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// create bad resources
	_, err = client.CreateDeployment(deploymentPublicBad)
	if err == nil {
		t.Fatalf("creating a deployment with invalid public did not result in an error")
	}

	_, err = client.CreateService(servicePublicBad)
	if err == nil {
		t.Fatalf("creating a service with invalid public did not result in an error")
	}

	id, err = client.CreateDeployment(deploymentPublic)
	if err != nil {
		t.Fatalf(err.Error())
	}

	deploymentPublic.ID = id

	if len(deploymentPublic.ID) == 0 {
		t.Fatalf("no deployment id received from client")
	}

	id, err = client.CreateService(servicePublic)
	if err != nil {
		t.Fatalf(err.Error())
	}

	servicePublic.ID = id

	if len(servicePublic.ID) == 0 {
		t.Fatalf("no service id received from client")
	}

	// Delete phase

	// check deleted before deleting
	deletedProject, err := client.NamespaceDeleted(namespacePublic.Name)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if deletedProject {
		t.Fatalf("namespace was deleted without calling delete namespace")
	}

	err = client.DeleteNamespace(namespacePublic.Name)
	if err != nil {
		t.Fatalf(err.Error())
	}

	deletedProject, err = client.NamespaceDeleted(namespacePublic.Name)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if !deletedProject {
		t.Fatalf("failed to delete namespace")
	}
}
