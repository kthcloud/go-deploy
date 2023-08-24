package deployment

import (
	"time"
)

type Deployment struct {
	ID      string `bson:"id"`
	Name    string `bson:"name"`
	OwnerID string `bson:"ownerId"`
	Zone    string `bson:"zone"`

	CreatedAt   time.Time `bson:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt"`
	RepairedAt  time.Time `bson:"repairedAt"`
	RestartedAt time.Time `bson:"restartedAt"`

	// keep until migrated
	Private      bool     `bson:"private"`
	Envs         []Env    `bson:"envs"`
	Volumes      []Volume `bson:"volumes"`
	InitCommands []string `bson:"initCommands"`

	Apps map[string]App `bson:"apps"`

	Activities []string `bson:"activities"`

	Subsystems    Subsystems `bson:"subsystems"`
	StatusCode    int        `bson:"statusCode"`
	StatusMessage string     `bson:"statusMessage"`

	PingResult int `bson:"pingResult"`
}

func (deployment *Deployment) GetMainApp() *App {
	app, ok := deployment.Apps["main"]
	if !ok {
		return nil
	}
	return &app
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
