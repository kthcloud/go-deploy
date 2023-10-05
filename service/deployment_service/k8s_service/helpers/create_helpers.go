package helpers

import (
	"errors"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"log"
)

func (client *Client) CreateNamespace(id string, public *k8sModels.NamespacePublic) (*k8sModels.NamespacePublic, error) {
	createdID, err := client.SsClient.CreateNamespace(public)
	if err != nil {
		return nil, err
	}

	namespace, err := client.SsClient.ReadNamespace(createdID)
	if err != nil {
		return nil, err
	}

	if namespace == nil {
		return nil, errors.New("failed to read namespace after creation")
	}

	err = client.UpdateDB(id, "namespace", namespace)
	if err != nil {
		return nil, err
	}

	client.K8s.Namespace = *namespace

	return namespace, nil
}

func (client *Client) CreateK8sDeployment(id, name string, public *k8sModels.DeploymentPublic) (*k8sModels.DeploymentPublic, error) {
	createdID, err := client.SsClient.CreateDeployment(public)
	if err != nil {
		return nil, err
	}

	k8sDeployment, err := client.SsClient.ReadDeployment(createdID)
	if err != nil {
		return nil, err
	}

	if k8sDeployment == nil {
		log.Printf("failed to read deployment after creation. assuming it was deleted")
		return nil, nil
	}

	err = client.UpdateDB(id, "deploymentMap."+name, k8sDeployment)
	if err != nil {
		return nil, err
	}

	client.K8s.SetDeployment(name, *k8sDeployment)

	return k8sDeployment, nil
}

func (client *Client) CreateService(id, name string, public *k8sModels.ServicePublic) (*k8sModels.ServicePublic, error) {
	createdID, err := client.SsClient.CreateService(public)
	if err != nil {
		return nil, err
	}

	k8sService, err := client.SsClient.ReadService(createdID)
	if err != nil {
		return nil, err
	}

	if k8sService == nil {
		log.Printf("failed to read service after creation. assuming it was deleted")
		return nil, nil
	}

	err = client.UpdateDB(id, "serviceMap."+name, k8sService)
	if err != nil {
		return nil, err
	}

	client.K8s.SetService(name, *k8sService)

	return k8sService, nil
}

func (client *Client) CreateIngress(id, name string, public *k8sModels.IngressPublic) (*k8sModels.IngressPublic, error) {
	var ingress *k8sModels.IngressPublic

	if public.Placeholder {
		ingress = public
	} else {
		createdID, err := client.SsClient.CreateIngress(public)
		if err != nil {
			return nil, err
		}

		ingress, err = client.SsClient.ReadIngress(createdID)
		if err != nil {
			return nil, err
		}

		if ingress == nil {
			return nil, errors.New("failed to read ingress after creation")
		}
	}

	err := client.UpdateDB(id, "ingressMap."+name, ingress)
	if err != nil {
		return nil, err
	}

	client.K8s.SetIngress(name, *ingress)

	return ingress, nil
}

func (client *Client) CreatePV(id, name string, public *k8sModels.PvPublic) (*k8sModels.PvPublic, error) {
	createdID, err := client.SsClient.CreatePV(public)
	if err != nil {
		return nil, err
	}

	pv, err := client.SsClient.ReadPV(createdID)
	if err != nil {
		return nil, err
	}

	if pv == nil {
		return nil, errors.New("failed to read persistent volume after creation")
	}

	err = client.UpdateDB(id, "pvMap."+name, pv)
	if err != nil {
		return nil, err
	}

	client.K8s.SetPV(name, *pv)

	return pv, nil
}

func (client *Client) CreatePVC(id, name string, public *k8sModels.PvcPublic) (*k8sModels.PvcPublic, error) {
	createdID, err := client.SsClient.CreatePVC(public)
	if err != nil {
		return nil, err
	}

	pvc, err := client.SsClient.ReadPVC(createdID)
	if err != nil {
		return nil, err
	}

	if pvc == nil {
		return nil, errors.New("failed to read persistent volume claim after creation")
	}

	err = client.UpdateDB(id, "pvcMap."+name, pvc)
	if err != nil {
		return nil, err
	}

	client.K8s.SetPVC(name, *pvc)

	return pvc, nil
}

func (client *Client) CreateJob(id, name string, public *k8sModels.JobPublic) (*k8sModels.JobPublic, error) {
	createdID, err := client.SsClient.CreateJob(public)
	if err != nil {
		return nil, err
	}

	job, err := client.SsClient.ReadJob(createdID)
	if err != nil {
		return nil, err
	}

	if job == nil {
		return nil, errors.New("failed to read job after creation")
	}

	err = client.UpdateDB(id, "jobMap."+name, job)
	if err != nil {
		return nil, err
	}

	client.K8s.SetJob(name, *job)

	return job, nil
}

func (client *Client) CreateSecret(id, name string, public *k8sModels.SecretPublic) (*k8sModels.SecretPublic, error) {
	createdID, err := client.SsClient.CreateSecret(public)
	if err != nil {
		return nil, err
	}

	secret, err := client.SsClient.ReadSecret(createdID)
	if err != nil {
		return nil, err
	}

	if secret == nil {
		return nil, errors.New("failed to read secret after creation")
	}

	err = client.UpdateDB(id, "secretMap."+name, secret)
	if err != nil {
		return nil, err
	}

	client.K8s.SetSecret(name, *secret)

	return secret, nil
}
