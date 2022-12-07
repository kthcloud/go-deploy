package harbor

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go-deploy/pkg/conf"
	"go-deploy/utils/subsystemutils"

	"github.com/mittwald/goharbor-client/v5/apiv2"
	modelv2 "github.com/mittwald/goharbor-client/v5/apiv2/model"
	"github.com/mittwald/goharbor-client/v5/apiv2/pkg/config"
	"github.com/sethvargo/go-password/password"
	"go.mongodb.org/mongo-driver/bson"
)

func updateDatabaseRobot(name string, robotUsername string, robotPassword string) error {
	err := models.UpdateProjectByName(name, bson.D{
		{"subsystems.harbor.robotUsername", robotUsername},
		{"subsystems.harbor.robotPassword", robotPassword},
	})
	if err != nil {
		return err
	}

	return nil
}

func updateDatabaseWebhook(name string, token string) error {
	err := models.UpdateProjectByName(name, bson.D{
		{"subsystems.harbor.webhookToken", token},
	})
	if err != nil {
		return err
	}

	return nil
}

func createClient() (*apiv2.RESTClient, error) {
	harbor := conf.Env.Harbor
	return apiv2.NewRESTClientForHost(harbor.Url, harbor.Identity, harbor.Secret, &config.Options{})
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

func updateRobotCredentials(client *apiv2.RESTClient, name string) error {
	robotName := getRobotName(name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to update credentials for harbor robot %s. details: %s", robotName, err)
	}

	robot, err := getRobotByNameV1(client, subsystemutils.GetPrefixedName(name), getRobotFullName(name))
	if err != nil {
		return makeError(err)
	}

	generatedSecret, err := password.Generate(10, 2, 2, true, false)
	if err != nil {
		return makeError(err)
	}

	updatedRobot := &modelv2.Robot{
		Secret: generatedSecret,
	}

	err = client.UpdateProjectRobotV1(context.TODO(), subsystemutils.GetPrefixedName(name), robot.ID, updatedRobot)
	if err != nil {
		return makeError(err)
	}

	err = updateDatabaseRobot(name, robotName, generatedSecret)
	if err != nil {
		return makeError(err)
	}

	return nil
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
