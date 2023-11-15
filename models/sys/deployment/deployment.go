package deployment

import (
	"fmt"
	"go-deploy/models/sys/activity"
	"time"
)

type Deployment struct {
	ID      string `bson:"id"`
	Name    string `bson:"name"`
	Type    string `bson:"type"`
	OwnerID string `bson:"ownerId"`
	Zone    string `bson:"zone"`

	CreatedAt   time.Time `bson:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt"`
	RepairedAt  time.Time `bson:"repairedAt"`
	RestartedAt time.Time `bson:"restartedAt"`
	DeletedAt   time.Time `bson:"deletedAt"`

	Activities map[string]activity.Activity `bson:"activities"`

	Apps       map[string]App `bson:"apps"`
	Subsystems Subsystems     `bson:"subsystems"`

	StatusMessage string `bson:"statusMessage"`
	StatusCode    int    `bson:"statusCode"`

	Transfer *Transfer `bson:"transfer,omitempty"`
}

func (deployment *Deployment) GetMainApp() *App {
	app, ok := deployment.Apps["main"]
	if !ok {
		return &App{}
	}
	return &app
}

func (deployment *Deployment) SetMainApp(app *App) {
	if deployment.Apps == nil {
		deployment.Apps = map[string]App{}
	}
	deployment.Apps["main"] = *app
}

func (deployment *Deployment) GetURL() *string {
	app := deployment.GetMainApp()
	if app == nil {
		return nil
	}

	ingress := deployment.Subsystems.K8s.GetIngress(deployment.Name)
	if ingress == nil || !ingress.Created() {
		return nil
	}

	if len(ingress.Hosts) > 0 && len(ingress.Hosts[0]) > 0 {
		url := fmt.Sprintf("https://%s", ingress.Hosts[0])
		return &url
	}

	return nil
}

func (deployment *Deployment) GetCustomDomainURL() *string {
	app := deployment.GetMainApp()
	if app == nil {
		return nil
	}

	if app.CustomDomain != nil && len(*app.CustomDomain) > 0 {
		url := fmt.Sprintf("https://%s", *app.CustomDomain)
		return &url
	}

	return nil
}

func (deployment *Deployment) Ready() bool {
	return !deployment.DoingActivity(ActivityBeingCreated) && !deployment.DoingActivity(ActivityBeingDeleted)
}

func (deployment *Deployment) DoingActivity(activity string) bool {
	for _, a := range deployment.Activities {
		if a.Name == activity {
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

func (deployment *Deployment) BeingTransferred() bool {
	return deployment.Transfer != nil
}

func (deployment *Deployment) Deleted() bool {
	return deployment.DeletedAt.After(time.Time{})
}
