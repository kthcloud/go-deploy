package helpers

import (
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/service"
)

func (client *Client) DeleteNamespace(id string) error {
	// never actually deleted the namespace to prevent race conditions

	err := client.UpdateDB(id, "namespace", nil)
	if err != nil {
		return err
	}

	client.K8s.Namespace = k8sModels.NamespacePublic{}

	return nil
}

func (client *Client) DeleteK8sDeployment(id, name string) error {
	deployment := client.K8s.GetDeployment(name)
	if service.Created(deployment) {
		err := client.SsClient.DeleteDeployment(deployment.ID)
		if err != nil {
			return err
		}
	}

	client.K8s.DeleteDeployment(name)
	return client.UpdateDB(id, "deploymentMap", client.K8s.DeploymentMap)
}

func (client *Client) DeleteService(id, name string) error {
	k8sService := client.K8s.GetService(name)
	if service.Created(k8sService) {
		err := client.SsClient.DeleteService(k8sService.ID)
		if err != nil {
			return err
		}
	}

	client.K8s.DeleteService(name)
	return client.UpdateDB(id, "serviceMap", client.K8s.ServiceMap)
}

func (client *Client) DeleteIngress(id, name string) error {
	ingress := client.K8s.GetIngress(name)
	if service.Created(ingress) {
		err := client.SsClient.DeleteIngress(ingress.ID)
		if err != nil {
			return err
		}
	}

	client.K8s.DeleteIngress(name)
	return client.UpdateDB(id, "ingressMap", client.K8s.IngressMap)
}

func (client *Client) DeletePV(id, name string) error {
	pv := client.K8s.GetPV(name)
	if service.Created(pv) {
		err := client.SsClient.DeletePV(pv.ID)
		if err != nil {
			return err
		}
	}

	client.K8s.DeletePV(name)
	return client.UpdateDB(id, "pvMap", client.K8s.PvMap)
}

func (client *Client) DeletePVC(id, name string) error {
	pvc := client.K8s.GetPVC(name)
	if service.Created(pvc) {
		err := client.SsClient.DeletePVC(pvc.ID)
		if err != nil {
			return err
		}
	}

	client.K8s.DeletePVC(name)
	return client.UpdateDB(id, "pvcMap", client.K8s.PvcMap)
}

func (client *Client) DeleteJob(id, name string) error {
	job := client.K8s.GetJob(name)
	if service.Created(job) {
		err := client.SsClient.DeleteJob(job.ID)
		if err != nil {
			return err
		}
	}

	client.K8s.DeleteJob(name)
	return client.UpdateDB(id, "jobMap", client.K8s.JobMap)
}

func (client *Client) DeleteSecret(id, name string) error {
	secret := client.K8s.GetSecret(name)
	if service.Created(secret) {
		err := client.SsClient.DeleteSecret(secret.ID)
		if err != nil {
			return err
		}
	}

	client.K8s.DeleteSecret(name)
	return client.UpdateDB(id, "secretMap", client.K8s.SecretMap)
}
