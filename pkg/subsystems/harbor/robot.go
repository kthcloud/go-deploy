package harbor

import (
	"context"
	"errors"
	"fmt"
	harborModelsV2 "github.com/mittwald/goharbor-client/v5/apiv2/model"
	harborErrors "github.com/mittwald/goharbor-client/v5/apiv2/pkg/errors"
	"go-deploy/pkg/subsystems/harbor/models"
	"log"
	"strings"
	"unicode"
)

func (client *Client) RobotCreated(public *models.RobotPublic) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if robot %s is created. details: %w", public.Name, err)
	}

	robot, err := client.HarborClient.GetRobotAccountByID(context.TODO(), int64(public.ID))
	if err != nil {
		return false, makeError(err)
	}

	return robot != nil && robot.ID != 0, nil
}

func (client *Client) ReadRobot(id int) (*models.RobotPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read robot %d. details: %w", id, err)
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

func (client *Client) CreateRobot(public *models.RobotPublic) (int, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create robot %s. details: %w", public.Name, err)
	}

	if public.ProjectID == 0 {
		return 0, nil
	}

	robots, err := client.HarborClient.ListProjectRobotsV1(context.TODO(), public.ProjectName)
	if err != nil {
		return 0, makeError(err)
	}

	var robot *harborModelsV2.Robot
	for _, r := range robots {
		if r.Name == getRobotFullName(public.ProjectName, public.Name) {
			robot = r
			break
		}
	}
	var appliedSecret string

	if robot == nil {
		created, err := client.createHarborRobot(public)
		if err != nil {
			return 0, makeError(err)
		}
		robot, err = client.HarborClient.GetRobotAccountByID(context.TODO(), created.ID)
		if err != nil {
			return 0, makeError(err)
		}
	}

	if public.Secret != "" {
		appliedSecret = public.Secret
	} else {
		appliedSecret = robot.Secret
	}
	err = client.assertCorrectRobotSecret(robot, appliedSecret)
	if err != nil {
		return 0, makeError(err)
	}

	return int(robot.ID), nil
}

func (client *Client) UpdateRobot(public *models.RobotPublic) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update robot %s. details: %w", public.Name, err)
	}

	if public.ProjectID == 0 {
		return nil
	}

	robots, err := client.HarborClient.ListProjectRobotsV1(context.TODO(), public.ProjectName)
	if err != nil {
		return makeError(err)
	}

	var robot *harborModelsV2.Robot
	for _, r := range robots {
		if r.ID == int64(public.ID) {
			robot = r
			break
		}
	}

	if robot == nil {
		log.Println("robot", public.Name, "not found when updating. assuming it was deleted")
		return nil
	}

	err = client.assertCorrectRobotSecret(robot, public.Secret)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (client *Client) DeleteRobot(id int) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete robot %d. details: %w", id, err)
	}

	if id == 0 {
		return nil
	}

	err := client.HarborClient.DeleteRobotAccountByID(context.TODO(), int64(id))
	if err != nil {
		if err != nil {
			targetErr := &harborErrors.ErrRobotAccountUnknownResource{}
			if !errors.As(err, &targetErr) && !strings.Contains(err.Error(), "[404] deleteRobotNotFound") {
				return makeError(err)
			}
		}
	}

	return nil
}

func (client *Client) DeleteAllRobots(projectID int) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete all robots for project %d. details: %w", projectID, err)
	}

	if projectID == 0 {
		return nil
	}

	robots, err := client.HarborClient.ListProjectRobotsV1(context.TODO(), fmt.Sprintf("%d", projectID))
	if err != nil {
		return makeError(err)
	}

	for _, robot := range robots {
		err = client.DeleteRobot(int(robot.ID))
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func (client *Client) assertCorrectRobotSecret(robot *harborModelsV2.Robot, secret string) error {
	if robot.Description != secret && isValidHarborRobotSecret(secret) {
		robot.Description = secret
		robot.Secret = secret

		_, err := client.HarborClient.RefreshRobotAccountSecretByID(context.TODO(), robot.ID, secret)
		if err != nil {
			return err
		}

		err = client.HarborClient.UpdateRobotAccount(context.TODO(), robot)
		if err != nil {
			return err
		}
	}

	return nil
}

func isValidHarborRobotSecret(secret string) bool {
	correctLength := len(secret) >= 8 && len(secret) <= 100

	var atLeastOneNumber bool
	var atLeastOneUpper bool
	var atLeastOneLower bool

	for _, c := range secret {
		if unicode.IsNumber(c) {
			atLeastOneNumber = true
		} else if unicode.IsUpper(c) {
			atLeastOneUpper = true
		} else if unicode.IsLower(c) {
			atLeastOneLower = true
		}
	}

	return correctLength && atLeastOneNumber && atLeastOneUpper && atLeastOneLower
}
