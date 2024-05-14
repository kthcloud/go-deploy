package harbor

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-openapi/runtime"
	"go-deploy/pkg/imp/harbor/sdk/v2.0/client/repository"
	"go-deploy/pkg/imp/harbor/sdk/v2.0/client/robotv1"
	"go-deploy/utils/requestutils"
	"net/http"
	"strings"
)

// IsNotFoundErr checks if the error is a not found error.
func IsNotFoundErr(err error) bool {
	var apiErr *runtime.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == http.StatusNotFound
	}

	var robotNotFoundErr *robotv1.ListRobotV1NotFound
	if errors.As(err, &robotNotFoundErr) {
		return true
	}

	var getRepositoryNotFoundErr *repository.GetRepositoryNotFound
	if errors.As(err, &getRepositoryNotFoundErr) {
		return true
	}

	if strings.Contains(strings.ToLower(err.Error()), "not found") {
		return true
	}

	return false
}

// doJSONRequest is a helper function to do a JSON request to the Harbor API.
func (client *Client) doJSONRequest(method string, relativePath string, requestBody interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf("%s%s", client.url, relativePath)
	return requestutils.DoRequestBasicAuth(method, fullURL, jsonBody, nil, client.username, client.password)
}

// createProject creates a project in Harbor.
// Needed since the installed Harbor client does not return credentials.
//func (client *Client) createHarborRobot(public *models.RobotPublic) (*models.RobotCreated, error) {
//	makeError := func(err error) error {
//		return fmt.Errorf("failed to create harbor robot %s. details: %w", public.Name, err)
//	}
//
//	robotRequestBody := models.CreateRobotCreateBody(public, client.Project)
//	res, err := client.doJSONRequest("POST", "/robots", *robotRequestBody)
//	if err != nil {
//		return nil, makeError(err)
//	}
//
//	if !requestutils.IsGoodStatusCode(res.StatusCode) {
//		return nil, makeApiError(res.Body, makeError)
//	}
//
//	var robotCreated modelv2.RobotCreated
//	err = requestutils.ParseBody(res.Body, &robotCreated)
//	if err != nil {
//		return nil, makeError(err)
//	}
//
//	return &robotCreated, nil
//}

// getRobotByNameV1 fetches a robot by name from the Harbor API.
// Needed since the installed Harbor client does not return credentials.
//func (client *Client) getRobotByNameV1(projectName string, name string) (*models.Robot, error) {
//	makeError := func(err error) error {
//		return fmt.Errorf("failed to fetch harbor robot %s by name. details: %w", name, err)
//	}
//
//	robots, err := client.HarborClient.ListProjectRobotsV1(context.TODO(), projectName)
//	if err != nil {
//		return nil, makeError(err)
//	}
//
//	var robotResult *modelv2.Robot
//	for _, robot := range robots {
//		if robot.Name == name {
//			robotResult = robot
//			break
//		}
//	}
//
//	return robotResult, nil
//}
