package model

import (
	"fmt"
	"time"
)

const (
	NLogsCache = 100
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

	Activities map[string]Activity `bson:"activities"`

	Apps       map[string]App       `bson:"apps"`
	Subsystems DeploymentSubsystems `bson:"subsystems"`
	Logs       []Log                `bson:"logs"`

	StatusMessage string `bson:"statusMessage"`
	StatusCode    int    `bson:"statusCode"`

	Transfer *DeploymentTransfer `bson:"transfer,omitempty"`
}

// GetMainApp returns the main app of the deployment.
// If the app does not exist, it will panic.
func (deployment *Deployment) GetMainApp() *App {
	app, ok := deployment.Apps["main"]
	if !ok {
		panic(fmt.Sprintf("deployment %s does not have a main app", deployment.Name))
	}
	return &app
}

// SetMainApp sets the main app of the deployment.
// If the app map is nil, it will be initialized before setting the app.
func (deployment *Deployment) SetMainApp(app *App) {
	if deployment.Apps == nil {
		deployment.Apps = map[string]App{}
	}
	deployment.Apps["main"] = *app
}

// GetURL returns the URL of the deployment.
// If the K8s ingress does not exist, it will return nil, or if the ingress does not have a host, it will return nil.
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

// GetCustomDomainURL returns the custom domain URL of the deployment.
// If the app does not have a custom domain, it will return nil.
// This method does not check whether the custom domain is active, and does
// not check if the ingress exists.
func (deployment *Deployment) GetCustomDomainURL() *string {
	app := deployment.GetMainApp()
	if app == nil {
		return nil
	}

	if app.CustomDomain != nil && len(app.CustomDomain.Domain) > 0 {
		url := fmt.Sprintf("https://%s", app.CustomDomain.Domain)
		return &url
	}

	return nil
}

// Ready returns true if the deployment is not being created or deleted.
func (deployment *Deployment) Ready() bool {
	return !deployment.DoingActivity(ActivityBeingCreated) && !deployment.DoingActivity(ActivityBeingDeleted)
}

// DoingActivity returns true if the deployment is doing the given activity.
func (deployment *Deployment) DoingActivity(activity string) bool {
	for _, a := range deployment.Activities {
		if a.Name == activity {
			return true
		}
	}
	return false
}

// BeingCreated returns true if the deployment is being created.
func (deployment *Deployment) BeingCreated() bool {
	return deployment.DoingActivity(ActivityBeingCreated)
}

// BeingDeleted returns true if the deployment is being deleted.
func (deployment *Deployment) BeingDeleted() bool {
	return deployment.DoingActivity(ActivityBeingDeleted)
}
