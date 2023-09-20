package helpers

import (
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service"
	"log"
)

func (client *Client) RepairNamespace(id string, genPublic func() *k8sModels.NamespacePublic) error {
	dbNamespace := client.K8s.Namespace
	if service.NotCreated(&dbNamespace) {
		public := genPublic()
		if public == nil {
			log.Println("no public supplied for k8s namespace when trying to create it in the repair process")
			return nil
		}

		_, err := client.CreateNamespace(id, public)
		return err
	}

	return service.UpdateIfDiff[k8sModels.NamespacePublic](
		dbNamespace,
		func() (*k8sModels.NamespacePublic, error) {
			return client.SsClient.ReadNamespace(dbNamespace.ID)
		},
		client.SsClient.UpdateNamespace,
		func(dbResource *k8sModels.NamespacePublic) error {
			return client.RecreateNamespace(id, dbResource)
		},
	)
}

func (client *Client) RepairK8sDeployment(id, name string, genPublic func() *k8sModels.DeploymentPublic) error {
	dbDeployment := client.K8s.GetDeployment(name)
	if service.NotCreated(dbDeployment) {
		public := genPublic()
		if public == nil {
			log.Println("no public supplied for k8s deployment", name, " when trying to create it in the repair process")
			return nil
		}

		_, err := client.CreateK8sDeployment(id, name, public)
		return err
	}

	return service.UpdateIfDiff(
		*dbDeployment,
		func() (*k8sModels.DeploymentPublic, error) {
			return client.SsClient.ReadDeployment(dbDeployment.ID)
		},
		client.SsClient.UpdateDeployment,
		func(dbResource *k8sModels.DeploymentPublic) error {
			return client.RecreateK8sDeployment(id, name, dbResource)
		},
	)
}

func (client *Client) RepairService(id, name string, genPublic func() *k8sModels.ServicePublic) error {
	dbService := client.K8s.GetService(name)
	if service.NotCreated(dbService) {
		public := genPublic()
		if public == nil {
			log.Println("no public supplied for k8s service", name, " when trying to create it in the repair process")
			return nil
		}

		_, err := client.CreateService(id, name, public)
		return err
	}

	return service.UpdateIfDiff(
		*dbService,
		func() (*k8sModels.ServicePublic, error) {
			return client.SsClient.ReadService(dbService.ID)
		},
		client.SsClient.UpdateService,
		func(dbResource *k8sModels.ServicePublic) error {
			return client.RecreateService(id, name, dbResource)
		},
	)
}

func (client *Client) RepairIngress(id, name string, genPublic func() *k8sModels.IngressPublic) error {
	dbIngress := client.K8s.GetIngress(name)
	if service.NotCreated(dbIngress) {
		public := genPublic()
		if public == nil {
			log.Println("no public supplied for k8s ingress", name, " when trying to create it in the repair process")
			return nil
		}

		_, err := client.CreateIngress(id, name, public)
		return err
	}

	return service.UpdateIfDiff(
		*dbIngress,
		func() (*k8sModels.IngressPublic, error) {
			return client.SsClient.ReadIngress(dbIngress.ID)
		},
		client.SsClient.UpdateIngress,
		func(dbResource *k8sModels.IngressPublic) error {
			return client.RecreateIngress(id, name, dbResource)
		},
	)
}
