package subsystems

import (
	githubModels "go-deploy/pkg/subsystems/github/models"
	harborModels "go-deploy/pkg/subsystems/harbor/models"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"time"
)

type K8s struct {
	Namespace  k8sModels.NamespacePublic  `bson:"namespace"`
	Deployment k8sModels.DeploymentPublic `bson:"deployment"`
	Service    k8sModels.ServicePublic    `bson:"service"`
	Ingress    k8sModels.IngressPublic    `bson:"ingress"`
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
