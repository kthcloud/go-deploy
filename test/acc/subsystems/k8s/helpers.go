package k8s

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/pkg/subsystems/k8s/models"
	"testing"
)

func withK8sClient(t *testing.T) *k8s.Client {
	zoneName := "se-flem"
	zone := conf.Env.Deployment.GetZone(zoneName)
	if zone == nil {
		t.Fatalf("no zone with name %s found", zoneName)
	}

	client, err := k8s.New(zone.Client)

	if err != nil {
		t.Fatalf(err.Error())
	}

	return client
}

func withK8sNamespace(t *testing.T) *models.NamespacePublic {
	client := withK8sClient(t)

	namespacePublic := &models.NamespacePublic{
		Name: "acc-test-" + uuid.New().String(),
	}

	id, err := client.CreateNamespace(namespacePublic)
	assert.NoError(t, err, "failed to create namespace")
	assert.NotZero(t, id, "no namespace id received from client")

	return namespacePublic
}

func withK8sDeployment(t *testing.T, namespace *models.NamespacePublic) *models.DeploymentPublic {
	client := withK8sClient(t)

	registry := conf.Env.DockerRegistry.URL
	image := conf.Env.DockerRegistry.Placeholder.Project + "/" + conf.Env.DockerRegistry.Placeholder.Repository

	deploymentPublic := &models.DeploymentPublic{
		Name:        "acc-test-" + uuid.New().String(),
		Namespace:   namespace.FullName,
		DockerImage: registry + "/" + image + ":latest",
		EnvVars:     nil,
	}

	id, err := client.CreateDeployment(deploymentPublic)
	assert.NoError(t, err, "failed to create deployment")
	assert.NotZero(t, id, "no deployment id received from client")

	createdDeployment, err := client.ReadDeployment(namespace.FullName, id)
	assert.NoError(t, err, "failed to read deployment")
	assert.NotNil(t, createdDeployment, "deployment is nil")

	deploymentPublic.ID = id
	assert.EqualValues(t, deploymentPublic, createdDeployment, "deployment does not match")

	return deploymentPublic
}

func withK8sService(t *testing.T, namespace *models.NamespacePublic) *models.ServicePublic {
	client := withK8sClient(t)

	servicePublic := &models.ServicePublic{
		Name:       "acc-test-" + uuid.New().String(),
		Namespace:  namespace.FullName,
		Port:       11111,
		TargetPort: 22222,
	}

	id, err := client.CreateService(servicePublic)
	assert.NoError(t, err, "failed to create service")
	assert.NotZero(t, id, "no service id received from client")

	createdService, err := client.ReadService(namespace.FullName, id)
	assert.NoError(t, err, "failed to read service")
	assert.NotNil(t, createdService, "service is nil")

	servicePublic.ID = id
	assert.EqualValues(t, servicePublic, createdService, "service does not match")

	return servicePublic
}

func withK8sIngress(t *testing.T, namespace *models.NamespacePublic, service *models.ServicePublic) *models.IngressPublic {
	client := withK8sClient(t)

	ingressPublic := &models.IngressPublic{
		Name:         "acc-test-" + uuid.New().String(),
		Namespace:    namespace.FullName,
		ServiceName:  service.Name,
		ServicePort:  service.Port,
		IngressClass: "nginx",
		Hosts:        []string{"acc-test-" + uuid.New().String() + ".example.com"},
	}

	id, err := client.CreateIngress(ingressPublic)
	assert.NoError(t, err, "failed to create ingress")
	assert.NotZero(t, id, "no ingress id received from client")

	createdIngress, err := client.ReadIngress(namespace.FullName, id)
	assert.NoError(t, err, "failed to read ingress")
	assert.NotNil(t, createdIngress, "ingress is nil")

	ingressPublic.ID = id
	// for safety reasons, we don't compare the ingress class name
	createdIngress.IngressClass = "nginx"
	assert.EqualValues(t, ingressPublic, createdIngress, "ingress does not match")

	return ingressPublic
}

func cleanUpNamespace(t *testing.T, namespace *models.NamespacePublic) {
	client := withK8sClient(t)

	err := client.DeleteNamespace(namespace.FullName)
	assert.NoError(t, err, "failed to delete namespace")

	deletedNamespace, err := client.ReadNamespace(namespace.FullName)
	assert.NoError(t, err, "failed to read namespace")
	assert.Nil(t, deletedNamespace, "namespace still exists")

	err = client.DeleteNamespace(namespace.FullName)
	assert.NoError(t, err, "failed to delete namespace again")
}

func cleanUpDeployment(t *testing.T, namespace *models.NamespacePublic, deployment *models.DeploymentPublic) {
	client := withK8sClient(t)

	err := client.DeleteDeployment(namespace.FullName, deployment.ID)
	assert.NoError(t, err, "failed to delete deployment")

	deletedDeployment, err := client.ReadDeployment(namespace.FullName, deployment.ID)
	assert.NoError(t, err, "failed to read deployment")
	assert.Nil(t, deletedDeployment, "deployment still exists")

	err = client.DeleteDeployment(namespace.FullName, deployment.ID)
	assert.NoError(t, err, "failed to delete deployment again")
}

func cleanUpService(t *testing.T, namespace *models.NamespacePublic, service *models.ServicePublic) {
	client := withK8sClient(t)

	err := client.DeleteService(namespace.FullName, service.ID)
	assert.NoError(t, err, "failed to delete service")

	deletedService, err := client.ReadService(namespace.FullName, service.ID)
	assert.NoError(t, err, "failed to read service")
	assert.Nil(t, deletedService, "service still exists")

	err = client.DeleteService(namespace.FullName, service.ID)
	assert.NoError(t, err, "failed to delete service again")
}

func cleanUpIngress(t *testing.T, namespace *models.NamespacePublic, ingress *models.IngressPublic) {
	client := withK8sClient(t)

	err := client.DeleteIngress(namespace.FullName, ingress.ID)
	assert.NoError(t, err, "failed to delete ingress")

	deletedIngress, err := client.ReadIngress(namespace.FullName, ingress.ID)
	assert.NoError(t, err, "failed to read ingress")
	assert.Nil(t, deletedIngress, "ingress still exists")

	err = client.DeleteIngress(namespace.FullName, ingress.ID)
	assert.NoError(t, err, "failed to delete ingress again")
}
