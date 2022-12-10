package harbor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mittwald/goharbor-client/v5/apiv2"
	modelv2 "github.com/mittwald/goharbor-client/v5/apiv2/model"
	"github.com/mittwald/goharbor-client/v5/apiv2/pkg/config"
	"go-deploy/models"
	"go-deploy/pkg/conf"
	"go-deploy/utils/requestutils"
	"go.mongodb.org/mongo-driver/bson"
)

func updateDatabaseRobot(name string, robotUsername string, robotPassword string) error {
	err := models.UpdateDeploymentByName(name, bson.D{
		{"subsystems.harbor.robotUsername", robotUsername},
		{"subsystems.harbor.robotPassword", robotPassword},
	})
	if err != nil {
		return err
	}

	return nil
}

func updateDatabaseWebhook(name string, token string) error {
	err := models.UpdateDeploymentByName(name, bson.D{
		{"subsystems.harbor.webhookToken", token},
	})
	if err != nil {
		return err
	}

	return nil
}

func createClient() (*apiv2.RESTClient, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create harbor client. details: %s", err.Error())
	}

	harbor := conf.Env.Harbor
	client, err := apiv2.NewRESTClientForHost(harbor.Url, harbor.Identity, harbor.Secret, &config.Options{})
	if err != nil {
		return nil, makeError(err)
	}

	return client, nil
}

func assertProjectExists(client *apiv2.RESTClient, projectName string) (bool, *modelv2.Project, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to assert harbor project %s exists. details: %s", projectName, err)
	}

	project, err := client.GetProject(context.TODO(), projectName)
	if err != nil {
		return false, nil, makeError(err)
	}
	return project.ProjectID != 0, project, nil
}

// Needed since the harbor client package refuses to return credentials
func createHarborRobot(name string) (*modelv2.RobotCreated, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create harbor robot %s. details: %s", name, err)
	}

	robotURL := fmt.Sprintf("%s/robots", conf.Env.Harbor.Url)

	username := conf.Env.Harbor.Identity
	password := conf.Env.Harbor.Secret

	robotRequestBody := createRobotRequestBody(name)
	robotRequestBodyJson, err := json.Marshal(robotRequestBody)
	if err != nil {
		return nil, makeError(err)
	}

	res, err := requestutils.DoRequestBasicAuth("POST", robotURL, robotRequestBodyJson, username, password)
	if err != nil {
		return nil, makeError(err)
	}

	defer requestutils.CloseBody(res.Body)

	body, err := requestutils.ReadBody(res.Body)
	if err != nil {
		return nil, makeError(err)
	}

	var robotCreated modelv2.RobotCreated
	err = requestutils.ParseJson(body, &robotCreated)
	if err != nil {
		return nil, makeError(err)
	}

	return &robotCreated, nil
}

func getRobotByNameV1(client *apiv2.RESTClient, projectName string, name string) (*modelv2.Robot, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch harbor robot %s by name. details: %s", name, err)
	}

	robots, err := client.ListProjectRobotsV1(context.TODO(), projectName)
	if err != nil {
		return nil, makeError(err)
	}

	robotResult := &modelv2.Robot{}
	for _, robot := range robots {
		if robot.Name == name {
			robotResult = robot
			break
		}
	}

	return robotResult, nil
}

func getWebhookEventTypes() []string {
	return []string{
		"PUSH_ARTIFACT",
	}
}
