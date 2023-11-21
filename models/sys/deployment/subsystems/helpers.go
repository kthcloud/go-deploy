package subsystems

import k8sModels "go-deploy/pkg/subsystems/k8s/models"

func (gitHub *GitHub) Created() bool {
	return gitHub.Webhook.Created()
}

func (k *K8s) GetNamespace() *k8sModels.NamespacePublic {
	return &k.Namespace
}

func (k *K8s) GetDeploymentMap() map[string]k8sModels.DeploymentPublic {
	if k.DeploymentMap == nil {
		k.DeploymentMap = make(map[string]k8sModels.DeploymentPublic)
	}

	return k.DeploymentMap
}

func (k *K8s) GetServiceMap() map[string]k8sModels.ServicePublic {
	if k.ServiceMap == nil {
		k.ServiceMap = make(map[string]k8sModels.ServicePublic)
	}

	return k.ServiceMap
}

func (k *K8s) GetIngressMap() map[string]k8sModels.IngressPublic {
	if k.IngressMap == nil {
		k.IngressMap = make(map[string]k8sModels.IngressPublic)
	}

	return k.IngressMap
}

func (k *K8s) GetPvMap() map[string]k8sModels.PvPublic {
	if k.PvMap == nil {
		k.PvMap = make(map[string]k8sModels.PvPublic)
	}

	return k.PvMap
}

func (k *K8s) GetPvcMap() map[string]k8sModels.PvcPublic {
	if k.PvcMap == nil {
		k.PvcMap = make(map[string]k8sModels.PvcPublic)
	}

	return k.PvcMap
}

func (k *K8s) GetJobMap() map[string]k8sModels.JobPublic {
	if k.JobMap == nil {
		k.JobMap = make(map[string]k8sModels.JobPublic)
	}

	return k.JobMap
}

func (k *K8s) GetSecretMap() map[string]k8sModels.SecretPublic {
	if k.SecretMap == nil {
		k.SecretMap = make(map[string]k8sModels.SecretPublic)
	}

	return k.SecretMap
}

func (k *K8s) GetHpaMap() map[string]k8sModels.HpaPublic {
	if k.HpaMap == nil {
		k.HpaMap = make(map[string]k8sModels.HpaPublic)
	}

	return k.HpaMap
}

func (k *K8s) GetDeployment(name string) *k8sModels.DeploymentPublic {
	resource, ok := k.GetDeploymentMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

func (k *K8s) GetService(name string) *k8sModels.ServicePublic {
	resource, ok := k.GetServiceMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

func (k *K8s) GetIngress(name string) *k8sModels.IngressPublic {
	resource, ok := k.GetIngressMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

func (k *K8s) GetPV(name string) *k8sModels.PvPublic {
	resource, ok := k.GetPvMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

func (k *K8s) GetPVC(name string) *k8sModels.PvcPublic {
	resource, ok := k.GetPvcMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

func (k *K8s) GetJob(name string) *k8sModels.JobPublic {
	resource, ok := k.GetJobMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

func (k *K8s) GetSecret(name string) *k8sModels.SecretPublic {
	resource, ok := k.GetSecretMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

func (k *K8s) GetHPA(name string) *k8sModels.HpaPublic {
	resource, ok := k.GetHpaMap()[name]
	if !ok {
		return nil
	}

	return &resource
}

func (k *K8s) SetNamespace(namespace k8sModels.NamespacePublic) {
	k.Namespace = namespace
}

func (k *K8s) SetDeployment(name string, deployment k8sModels.DeploymentPublic) {
	k.GetDeploymentMap()[name] = deployment
}

func (k *K8s) SetService(name string, service k8sModels.ServicePublic) {
	k.GetServiceMap()[name] = service
}

func (k *K8s) SetIngress(name string, ingress k8sModels.IngressPublic) {
	k.GetIngressMap()[name] = ingress
}

func (k *K8s) SetPV(name string, pv k8sModels.PvPublic) {
	k.GetPvMap()[name] = pv
}

func (k *K8s) SetPVC(name string, pvc k8sModels.PvcPublic) {
	k.GetPvcMap()[name] = pvc
}

func (k *K8s) SetJob(name string, job k8sModels.JobPublic) {
	k.GetJobMap()[name] = job
}

func (k *K8s) SetSecret(name string, secret k8sModels.SecretPublic) {
	k.GetSecretMap()[name] = secret
}

func (k *K8s) SetHPA(name string, hpa k8sModels.HpaPublic) {
	k.GetHpaMap()[name] = hpa
}

func (k *K8s) DeleteDeployment(name string) {
	delete(k.GetDeploymentMap(), name)
}

func (k *K8s) DeleteService(name string) {
	delete(k.GetServiceMap(), name)
}

func (k *K8s) DeleteIngress(name string) {
	delete(k.GetIngressMap(), name)
}

func (k *K8s) DeletePV(name string) {
	delete(k.GetPvMap(), name)
}

func (k *K8s) DeletePVC(name string) {
	delete(k.GetPvcMap(), name)
}

func (k *K8s) DeleteJob(name string) {
	delete(k.GetJobMap(), name)
}

func (k *K8s) DeleteSecret(name string) {
	delete(k.GetSecretMap(), name)
}

func (k *K8s) DeleteHPA(name string) {
	delete(k.GetHpaMap(), name)
}
