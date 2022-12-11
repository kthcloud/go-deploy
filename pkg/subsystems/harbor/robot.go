package harbor

import (
	"context"
	"fmt"
	modelv2 "github.com/mittwald/goharbor-client/v5/apiv2/model"
)

func (client *Client) RobotCreated(projectName, name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if harbor robot %s is created. details: %s", getRobotFullName(projectName, name), err)
	}

	robot, err := client.getRobotByNameV1(projectName, getRobotFullName(projectName, name))
	if err != nil {
		return false, makeError(err)
	}

	return robot != nil, nil
}

func (client *Client) CreateRobot(projectName, name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create harbor robot %s. details: %s", name, err)
	}

	projectExists, project, err := client.assertProjectExists(projectName)
	if err != nil {
		return makeError(err)
	}

	if !projectExists {
		err = fmt.Errorf("no project exists")
		return makeError(err)
	}

	robots, err := client.HarborClient.ListProjectRobotsV1(context.TODO(), project.Name)
	if err != nil {
		return err
	}

	var robotResult *modelv2.Robot
	for _, robot := range robots {
		if robot.Name == getRobotFullName(projectName, name) {
			robotResult = robot
			break
		}
	}

	if robotResult != nil {
		return nil
	}

	robotCreatedBody, err := client.createHarborRobot(projectName, name)
	if err != nil {
		return makeError(err)
	}

	err = updateDatabaseRobot(name, robotCreatedBody.Name, robotCreatedBody.Secret)
	if err != nil {
		return makeError(err)
	}

	return nil
}
