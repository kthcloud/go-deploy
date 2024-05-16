package harbor

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-openapi/runtime"
	"go-deploy/pkg/imp/harbor/sdk/v2.0/client/project"
	"go-deploy/pkg/imp/harbor/sdk/v2.0/client/repository"
	"go-deploy/pkg/imp/harbor/sdk/v2.0/client/robot"
	"go-deploy/pkg/imp/harbor/sdk/v2.0/client/robotv1"
	"go-deploy/pkg/imp/harbor/sdk/v2.0/client/webhook"
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

	// Robot V1
	var listRobotV1NotFoundErr *robotv1.ListRobotV1NotFound
	if errors.As(err, &listRobotV1NotFoundErr) {
		return true
	}

	var getRobotV1NotFoundErr *robotv1.GetRobotByIDV1NotFound
	if errors.As(err, &getRobotV1NotFoundErr) {
		return true
	}

	var createRobotV1NotFoundErr *robotv1.CreateRobotV1NotFound
	if errors.As(err, &createRobotV1NotFoundErr) {
		return true
	}

	var updateRobotV1NotFoundErr *robotv1.UpdateRobotV1NotFound
	if errors.As(err, &updateRobotV1NotFoundErr) {
		return true
	}

	var deleteRobotV1NotFoundErr *robotv1.DeleteRobotV1NotFound
	if errors.As(err, &deleteRobotV1NotFoundErr) {
		return true
	}

	// Robot V2
	var getRobotByIdNotFoundErr *robot.GetRobotByIDNotFound
	if errors.As(err, &getRobotByIdNotFoundErr) {
		return true
	}

	var listRobotNotFoundErr *robot.ListRobotNotFound
	if errors.As(err, &listRobotNotFoundErr) {
		return true
	}

	var createRobotNotFoundErr *robot.CreateRobotNotFound
	if errors.As(err, &createRobotNotFoundErr) {
		return true
	}

	var updateRobotNotFoundErr *robot.UpdateRobotNotFound
	if errors.As(err, &updateRobotNotFoundErr) {
		return true
	}

	var deleteRobotNotFoundErr *robot.DeleteRobotNotFound
	if errors.As(err, &deleteRobotNotFoundErr) {
		return true
	}

	// Repository
	var getRepositoryNotFoundErr *repository.GetRepositoryNotFound
	if errors.As(err, &getRepositoryNotFoundErr) {
		return true
	}

	var listRepositoryNotFoundErr *repository.ListRepositoriesNotFound
	if errors.As(err, &listRepositoryNotFoundErr) {
		return true
	}

	var updateRepositoryNotFoundErr *repository.UpdateRepositoryNotFound
	if errors.As(err, &updateRepositoryNotFoundErr) {
		return true
	}

	var deleteRepositoryNotFoundErr *repository.DeleteRepositoryNotFound
	if errors.As(err, &deleteRepositoryNotFoundErr) {
		return true
	}

	// Webhook
	var getPoliciesProjectNotFoundErr *webhook.GetWebhookPolicyOfProjectNotFound
	if errors.As(err, &getPoliciesProjectNotFoundErr) {
		return true
	}

	var updatePoliciesProjectNotFoundErr *webhook.UpdateWebhookPolicyOfProjectNotFound
	if errors.As(err, &updatePoliciesProjectNotFoundErr) {
		return true
	}

	var deletePoliciesProjectNotFoundErr *webhook.DeleteWebhookPolicyOfProjectNotFound
	if errors.As(err, &deletePoliciesProjectNotFoundErr) {
		return true
	}

	// Project
	var updateProjectNotFoundErr *project.UpdateProjectNotFound
	if errors.As(err, &updateProjectNotFoundErr) {
		return true
	}

	var deleteProjectNotFoundErr *project.DeleteProjectNotFound
	if errors.As(err, &deleteProjectNotFoundErr) {
		return true
	}

	// Fallback
	if strings.Contains(strings.ToLower(err.Error()), "not found") || strings.Contains(strings.ToLower(err.Error()), "404") {
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
