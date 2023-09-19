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
