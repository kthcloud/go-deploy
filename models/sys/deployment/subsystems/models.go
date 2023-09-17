package subsystems

import (
	githubModels "go-deploy/pkg/subsystems/github/models"
	harborModels "go-deploy/pkg/subsystems/harbor/models"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"time"
)

type K8s struct {
	Namespace k8sModels.NamespacePublic `bson:"namespace"`

	DeploymentMap map[string]k8sModels.DeploymentPublic `bson:"deploymentMap"`
	ServiceMap    map[string]k8sModels.ServicePublic    `bson:"serviceMap"`
	IngressMap    map[string]k8sModels.IngressPublic    `bson:"ingressMap"`
	PvMap         map[string]k8sModels.PvPublic         `bson:"pvMap"`
	PvcMap        map[string]k8sModels.PvcPublic        `bson:"pvcMap"`
	JobMap        map[string]k8sModels.JobPublic        `bson:"jobMap"`
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

func (k *K8s) SetDeployment(name string, deployment k8sModels.DeploymentPublic) {
	k.GetDeploymentMap()[name] = deployment
}

func (k *K8s) SetService(name string, service k8sModels.ServicePublic) {
	k.GetServiceMap()[name] = service
}

func (k *K8s) SetIngress(name string, ingress k8sModels.IngressPublic) {
	k.GetIngressMap()[name] = ingress
}

func (k *K8s) SetPv(name string, pv k8sModels.PvPublic) {
	k.GetPvMap()[name] = pv
}

func (k *K8s) SetPvc(name string, pvc k8sModels.PvcPublic) {
	k.GetPvcMap()[name] = pvc
}

func (k *K8s) SetJob(name string, job k8sModels.JobPublic) {
	k.GetJobMap()[name] = job
}

type Harbor struct {
	Project    harborModels.ProjectPublic    `bson:"project"`
	Robot      harborModels.RobotPublic      `bson:"robot"`
	Repository harborModels.RepositoryPublic `bson:"repository"`
	Webhook    harborModels.WebhookPublic    `bson:"webhook"`
}

type GitHub struct {
	Webhook     githubModels.WebhookPublic `bson:"webhook"`
	Placeholder bool                       `bson:"placeholder"`
}

type GitLab struct {
	LastBuild GitLabBuild `bson:"lastBuild"`
}

type GitLabBuild struct {
	ID        int       `bson:"id"`
	ProjectID int       `bson:"projectId"`
	Trace     []string  `bson:"trace"`
	Status    string    `bson:"status"`
	Stage     string    `bson:"stage"`
	CreatedAt time.Time `bson:"createdAt"`
}
