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

func (client *Client) ReadRobot(id int) (*models.RobotPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read robot %d. details: %w", id, err)
	}

	if id == 0 {
		log.Println("id not supplied when reading robot. assuming it was deleted")
		return nil, nil
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

func (client *Client) CreateRobot(public *models.RobotPublic) (*models.RobotPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create robot %s. details: %w", public.Name, err)
	}

	robots, err := client.HarborClient.ListProjectRobotsV1(context.TODO(), client.Project)
	if err != nil {
		return nil, makeError(err)
	}

	var robot *harborModelsV2.Robot
	for _, r := range robots {
		if r.Name == getRobotFullName(client.Project, public.Name) {
			robot = r
			break
		}
	}

	var appliedSecret string
	if robot == nil {
		created, err := client.createHarborRobot(public)
		if err != nil {
			return nil, makeError(err)
		}

		appliedSecret = created.Secret

		robot, err = client.HarborClient.GetRobotAccountByID(context.TODO(), created.ID)
		if err != nil {
			return nil, makeError(err)
		}
	}

	if public.Secret != "" {
		appliedSecret = public.Secret
	}

	err = client.assertCorrectRobotSecret(robot, appliedSecret)
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreateRobotPublicFromGet(robot, nil), nil
}

func (client *Client) UpdateRobot(public *models.RobotPublic) (*models.RobotPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update robot %s. details: %w", public.Name, err)
	}

	if public.ID == 0 {
		log.Println("id not supplied when updating robot. assuming it was deleted")
		return nil, nil
	}

	robots, err := client.HarborClient.ListProjectRobotsV1(context.TODO(), client.Project)
	if err != nil {
		return nil, makeError(err)
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
		return nil, nil
	}

	err = client.assertCorrectRobotSecret(robot, public.Secret)
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreateRobotPublicFromGet(robot, nil), nil
}

func (client *Client) DeleteRobot(id int) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete robot %d. details: %w", id, err)
	}

	if id == 0 {
		log.Println("id not supplied when deleting robot. assuming it was deleted")
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
