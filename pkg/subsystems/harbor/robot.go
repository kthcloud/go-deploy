package harbor

import (
	"context"
	"fmt"
	"github.com/go-openapi/strfmt"
	robotModels "go-deploy/pkg/imp/harbor/sdk/v2.0/client/robot"
	"go-deploy/pkg/imp/harbor/sdk/v2.0/client/robotv1"
	harborModels "go-deploy/pkg/imp/harbor/sdk/v2.0/models"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems/harbor/models"
	"unicode"
)

func robotPermissions(project string) []*harborModels.RobotPermission {
	return []*harborModels.RobotPermission{
		{
			Access: []*harborModels.Access{
				{
					Action:   "list",
					Resource: "repository",
				},
				{
					Action:   "pull",
					Resource: "repository",
				},
				{
					Action:   "push",
					Resource: "repository",
				},
				{
					Action:   "create",
					Resource: "tag",
				},
			},
			Kind:      "project",
			Namespace: project,
		},
	}
}

// ReadRobot reads a robot from Harbor.
func (client *Client) ReadRobot(id int) (*models.RobotPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read robot %d. details: %w", id, err)
	}

	if id == 0 {
		log.Println("ID not supplied when reading robot. Assuming it was deleted")
		return nil, nil
	}

	robot, err := client.HarborClient.V2().Robot.GetRobotByID(context.TODO(), &robotModels.GetRobotByIDParams{
		RobotID: int64(id),
	})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	var public *models.RobotPublic
	if robot != nil {
		public = models.CreateRobotPublicFromGet(robot.Payload)
	}

	return public, nil
}

// CreateRobot creates a robot in Harbor.
func (client *Client) CreateRobot(public *models.RobotPublic) (*models.RobotPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create robot %s. details: %w", public.Name, err)
	}

	robots, err := client.HarborClient.V2().Robotv1.ListRobotV1(context.TODO(), &robotv1.ListRobotV1Params{
		ProjectNameOrID: client.Project,
	})
	if err != nil {
		if !IsNotFoundErr(err) {
			return nil, makeError(err)
		}
	}

	var robot *harborModels.Robot
	if robots != nil {
		for _, r := range robots.Payload {
			if r.Name == getRobotFullName(client.Project, public.Name) {
				robot = r
				break
			}
		}
	}

	var appliedSecret string
	if robot == nil {
		robotCreated, err := client.HarborClient.V2().Robot.CreateRobot(context.TODO(), &robotModels.CreateRobotParams{
			Robot: &harborModels.RobotCreate{
				Description: "",
				Disable:     false,
				Duration:    -1,
				Level:       "project",
				Name:        public.Name,
				Permissions: robotPermissions(client.Project),
				Secret:      public.Secret,
			},
		})
		if err != nil {
			return nil, err
		}

		r, err := client.HarborClient.V2().Robotv1.GetRobotByIDV1(context.TODO(), &robotv1.GetRobotByIDV1Params{
			RobotID:         robotCreated.Payload.ID,
			ProjectNameOrID: client.Project,
		})
		if err != nil {
			return nil, makeError(err)
		}
		robot = r.Payload

		appliedSecret = robotCreated.Payload.Secret
	}

	if public.Secret != "" {
		appliedSecret = public.Secret
	}

	err = client.assertCorrectRobotSecret(robot, appliedSecret)
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreateRobotPublicFromGet(robot), nil
}

// UpdateRobot updates a robot in Harbor.
func (client *Client) UpdateRobot(public *models.RobotPublic) (*models.RobotPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update robot %s. details: %w", public.Name, err)
	}

	if public.ID == 0 {
		log.Println("ID not supplied when updating robot. Assuming it was deleted")
		return nil, nil
	}

	_, err := client.HarborClient.V2().Robot.UpdateRobot(context.TODO(), &robotModels.UpdateRobotParams{
		Robot: &harborModels.Robot{
			CreationTime: strfmt.DateTime{},
			Description:  "",
			Disable:      false,
			Duration:     -1,
			Editable:     false,
			ExpiresAt:    -1,
			ID:           int64(public.ID),
			Level:        "project",
			Name:         public.Name,
			Permissions:  robotPermissions(client.Project),
			Secret:       "",
			UpdateTime:   strfmt.DateTime{},
		},
		RobotID: int64(public.ID),
	})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	robots, err := client.HarborClient.V2().Robotv1.ListRobotV1(context.TODO(), &robotv1.ListRobotV1Params{ProjectNameOrID: client.Project})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	var robot *harborModels.Robot
	if robots != nil {
		for _, r := range robots.Payload {
			if r.ID == int64(public.ID) {
				robot = r
				break
			}
		}
	}

	if robot == nil {
		log.Println("Robot", public.Name, "not found when updating. Assuming it was deleted")
		return nil, nil
	}

	err = client.assertCorrectRobotSecret(robot, public.Secret)
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreateRobotPublicFromGet(robot), nil
}

// DeleteRobot deletes a robot in Harbor.
func (client *Client) DeleteRobot(id int) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete robot %d. details: %w", id, err)
	}

	if id == 0 {
		log.Println("ID not supplied when deleting robot. Assuming it was deleted")
		return nil
	}

	_, err := client.HarborClient.V2().Robot.DeleteRobot(context.TODO(), &robotModels.DeleteRobotParams{
		RobotID: int64(id),
	})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil
		}

		return makeError(err)
	}

	return nil
}

// assertCorrectRobotSecret asserts that the robot secret is correct.
// This is needed since the installed Harbor client does not return credentials.
// We use the description field to store the secret (which is arguably a hack, but it works).
func (client *Client) assertCorrectRobotSecret(robot *harborModels.Robot, secret string) error {
	if robot.Description != secret && isValidHarborRobotSecret(secret) {
		robot.Description = secret
		robot.Secret = secret

		_, err := client.HarborClient.V2().Robot.RefreshSec(context.TODO(), &robotModels.RefreshSecParams{
			RobotSec: &harborModels.RobotSec{
				Secret: secret,
			},
			RobotID: robot.ID,
		})
		if err != nil {
			if IsNotFoundErr(err) {
				return nil
			}

			return err
		}

		_, err = client.HarborClient.V2().Robot.UpdateRobot(context.TODO(), &robotModels.UpdateRobotParams{
			Robot:   robot,
			RobotID: robot.ID,
		})
		if err != nil {
			if IsNotFoundErr(err) {
				return nil
			}

			return err
		}
	}

	return nil
}

// isValidHarborRobotSecret asserts that a secret is a valid to be used as a Harbor robot secret.
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
