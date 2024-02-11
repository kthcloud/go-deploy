package subsystems

import (
	harborModels "go-deploy/pkg/subsystems/harbor/models"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"time"
)

type K8s struct {
	Namespace k8sModels.NamespacePublic `bson:"namespace"`

	DeploymentMap    map[string]k8sModels.DeploymentPublic    `bson:"deploymentMap,omitempty"`
	ServiceMap       map[string]k8sModels.ServicePublic       `bson:"serviceMap,omitempty"`
	IngressMap       map[string]k8sModels.IngressPublic       `bson:"ingressMap,omitempty"`
	PvMap            map[string]k8sModels.PvPublic            `bson:"pvMap,omitempty"`
	PvcMap           map[string]k8sModels.PvcPublic           `bson:"pvcMap,omitempty"`
	JobMap           map[string]k8sModels.JobPublic           `bson:"jobMap,omitempty"`
	SecretMap        map[string]k8sModels.SecretPublic        `bson:"secretMap,omitempty"`
	HpaMap           map[string]k8sModels.HpaPublic           `bson:"hpaMap,omitempty"`
	NetworkPolicyMap map[string]k8sModels.NetworkPolicyPublic `bson:"networkPolicyMap,omitempty"`
}

type Harbor struct {
	Project     harborModels.ProjectPublic    `bson:"project"`
	Robot       harborModels.RobotPublic      `bson:"robot"`
	Repository  harborModels.RepositoryPublic `bson:"repository"`
	Webhook     harborModels.WebhookPublic    `bson:"webhook"`
	Placeholder bool                          `bson:"placeholder"`
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
