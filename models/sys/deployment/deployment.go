package deployment

import (
	"go-deploy/models/sys/deployment/subsystems"
	"time"
)

type Subsystems struct {
	K8s    subsystems.K8s    `bson:"k8s"`
	Harbor subsystems.Harbor `bson:"harbor"`
	GitHub subsystems.GitHub `bson:"github"`
	GitLab subsystems.GitLab `bson:"gitlab"`
}

type Deployment struct {
	ID      string `bson:"id"`
	Name    string `bson:"name"`
	OwnerID string `bson:"ownerId"`
	Zone    string `bson:"zone"`

	CreatedAt   time.Time `bson:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt"`
	RepairedAt  time.Time `bson:"repairedAt"`
	RestartedAt time.Time `bson:"restartedAt"`

	Private      bool     `bson:"private"`
	Envs         []Env    `bson:"envs"`
	ExtraDomains []string `bson:"extraDomains"`

	Activities []string `bson:"activities"`

	Subsystems    Subsystems `bson:"subsystems"`
	StatusCode    int        `bson:"statusCode"`
	StatusMessage string     `bson:"statusMessage"`

	PingResult int `bson:"pingResult"`
}

func (deployment *Deployment) Ready() bool {
	return !deployment.DoingActivity(ActivityBeingCreated) && !deployment.DoingActivity(ActivityBeingDeleted)
}

func (deployment *Deployment) DoingActivity(activity string) bool {
	for _, a := range deployment.Activities {
		if a == activity {
			return true
		}
	}
	return false
}

func (deployment *Deployment) BeingCreated() bool {
	return deployment.DoingActivity(ActivityBeingCreated)
}

func (deployment *Deployment) BeingDeleted() bool {
	return deployment.DoingActivity(ActivityBeingDeleted)
}

func (deployment *Deployment) Created() bool {
	return deployment.ID != "" &&
		deployment.Subsystems.GitHub.Created() &&
		deployment.Subsystems.Harbor.Created() &&
		deployment.Subsystems.K8s.Created()
}
