package harbor

import (
	"context"
	"errors"
	"fmt"
	modelv2 "github.com/mittwald/goharbor-client/v5/apiv2/model"
	harborErrors "github.com/mittwald/goharbor-client/v5/apiv2/pkg/errors"
	"go-deploy/pkg/subsystems/harbor/models"
	"strings"
)

func (client *Client) RobotCreated(public *models.RobotPublic) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if robot %s is created. details: %s", public.Name, err)
	}

	robot, err := client.HarborClient.GetRobotAccountByID(context.TODO(), int64(public.ID))
	if err != nil {
		return false, makeError(err)
	}

	return robot != nil && robot.ID != 0, nil
}

func (client *Client) ReadRobot(id int) (*models.RobotPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create robot %d. details: %s", id, err)
	}

	robot, err := client.HarborClient.GetRobotAccountByID(context.TODO(), int64(id))
	if err != nil {
		errStr := fmt.Sprintf("%s", err)
		if !strings.Contains(errStr, "NotFound") {
			return nil, makeError(err)
		}
	}

	var public *models.RobotPublic
	if robot != nil {
		project, err := client.HarborClient.GetProject(context.TODO(), robot.Permissions[0].Namespace)
		if err != nil {
			return nil, makeError(err)
		}

		public = models.CreateRobotPublicFromGet(robot, project)
	}

	return public, nil
}

func (client *Client) CreateRobot(public *models.RobotPublic) (*models.RobotCreated, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create robot %s. details: %s", public.Name, err)
	}

	if public.ProjectID == 0 {
		return nil, makeError(fmt.Errorf("project id required"))
	}

	robots, err := client.HarborClient.ListProjectRobotsV1(context.TODO(), public.ProjectName)
	if err != nil {
		return nil, makeError(err)
	}

	var robotResult *modelv2.Robot
	for _, robot := range robots {
		if robot.Name == getRobotFullName(public.ProjectName, public.Name) {
			robotResult = robot
			break
		}
	}

	if robotResult != nil {
		// delete this robot, since it would not have been requested if credentials was known
		// we could also just refresh the robot credentials
		err = client.DeleteRobot(int(robotResult.ID))
		if err != nil {
			return nil, makeError(err)
		}
	}

	robotCreatedBody, err := client.createHarborRobot(public)
	if err != nil {
		return nil, makeError(err)
	}

	return &models.RobotCreated{
		ID:     int(robotCreatedBody.ID),
		Secret: robotCreatedBody.Secret,
	}, nil
}

func (client *Client) DeleteRobot(id int) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create robot %d. details: %s", id, err)
	}

	if id == 0 {
		return makeError(fmt.Errorf("id required"))
	}

	err := client.HarborClient.DeleteRobotAccountByID(context.TODO(), int64(id))
	if err != nil {
		if err != nil {
			targetErr := &harborErrors.ErrRobotAccountUnknownResource{}
			if !errors.As(err, &targetErr) {
				return makeError(err)
			}
		}
	}

	return nil
}

func (client *Client) UpdateRobot(public *models.RobotPublic) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update robot %s. details: %s", public.Name, err)
	}

	if public.ID == 0 {
		return makeError(fmt.Errorf("id required"))
	}

	requestBody := models.CreateRobotUpdateFromPublic(public)

	err := client.HarborClient.UpdateRobotAccount(context.TODO(), requestBody)
	if err != nil {
		return makeError(err)
	}

	return nil
}
