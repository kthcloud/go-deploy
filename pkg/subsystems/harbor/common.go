package harbor

import (
	"context"
	"encoding/json"
	"fmt"
	modelv2 "github.com/mittwald/goharbor-client/v5/apiv2/model"
	"go-deploy/pkg/subsystems/harbor/models"
	"go-deploy/utils/requestutils"
	"net/http"
)

func (client *Client) doJSONRequest(method string, relativePath string, requestBody interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf("%s%s", client.url, relativePath)
	return requestutils.DoRequestBasicAuth(method, fullURL, jsonBody, nil, client.username, client.password)
}

func (client *Client) assertProjectExists(name string) (bool, *modelv2.Project, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to assert harbor project %s exists. details: %w", name, err)
	}

	project, err := client.HarborClient.GetProject(context.TODO(), name)
	if err != nil {
		return false, nil, makeError(err)
	}
	return project.ProjectID != 0, project, nil
}

// Needed since the harbor client package refuses to return credentials
func (client *Client) createHarborRobot(public *models.RobotPublic) (*modelv2.RobotCreated, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create harbor robot %s. details: %w", public.Name, err)
	}

	robotRequestBody := models.CreateRobotCreateBody(public)
	res, err := client.doJSONRequest("POST", "/robots", robotRequestBody)
	if err != nil {
		return nil, makeError(err)
	}

	if !requestutils.IsGoodStatusCode(res.StatusCode) {
		return nil, makeApiError(res.Body, makeError)
	}

	var robotCreated modelv2.RobotCreated
	err = requestutils.ParseBody(res.Body, &robotCreated)
	if err != nil {
		return nil, makeError(err)
	}

	return &robotCreated, nil
}

func (client *Client) getRobotByNameV1(projectName string, name string) (*modelv2.Robot, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch harbor robot %s by name. details: %w", name, err)
	}

	robots, err := client.HarborClient.ListProjectRobotsV1(context.TODO(), projectName)
	if err != nil {
		return nil, makeError(err)
	}

	var robotResult *modelv2.Robot
	for _, robot := range robots {
		if robot.Name == name {
			robotResult = robot
			break
		}
	}

	return robotResult, nil
}
