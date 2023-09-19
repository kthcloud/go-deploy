package helpers

import (
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
)

func (client *Client) RecreateNamespace(id string, newPublic *k8sModels.NamespacePublic) error {
	err := client.DeleteNamespace(id)
	if err != nil {
		return err
	}

	_, err = client.CreateNamespace(id, newPublic)
	return err
}

func (client *Client) RecreateK8sDeployment(id, name string, newPublic *k8sModels.DeploymentPublic) error {
	err := client.DeleteK8sDeployment(id, name)
	if err != nil {
		return err
	}

	_, err = client.CreateK8sDeployment(id, name, newPublic)
	return err
}

func (client *Client) RecreateService(id, name string, newPublic *k8sModels.ServicePublic) error {
	err := client.DeleteService(id, name)
	if err != nil {
		return err
	}

	_, err = client.CreateService(id, name, newPublic)
	return err
}

func (client *Client) RecreateIngress(id, name string, newPublic *k8sModels.IngressPublic) error {
	err := client.DeleteIngress(id, name)
	if err != nil {
		return err
	}

	_, err = client.CreateIngress(id, name, newPublic)
	return err
}

func (client *Client) RecreatePV(id, name string, newPublic *k8sModels.PvPublic) error {
	err := client.DeletePV(id, name)
	if err != nil {
		return err
	}

	_, err = client.CreatePV(id, name, newPublic)
	return err
}

func (client *Client) RecreatePVC(id, name string, newPublic *k8sModels.PvcPublic) error {
	err := client.DeletePVC(id, name)
	if err != nil {
		return err
	}

	_, err = client.CreatePVC(id, name, newPublic)
	return err
}
