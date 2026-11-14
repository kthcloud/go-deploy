package harbor

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-openapi/runtime"
	"github.com/kthcloud/go-deploy/pkg/imp/harbor/sdk/v2.0/client/project"
	"github.com/kthcloud/go-deploy/pkg/imp/harbor/sdk/v2.0/client/repository"
	"github.com/kthcloud/go-deploy/pkg/imp/harbor/sdk/v2.0/client/robot"
	"github.com/kthcloud/go-deploy/pkg/imp/harbor/sdk/v2.0/client/robotv1"
	"github.com/kthcloud/go-deploy/pkg/imp/harbor/sdk/v2.0/client/webhook"
)

// IsNotFoundErr checks if the error is a not found error.
func IsNotFoundErr(err error) bool {
	if apiErr, ok := errors.AsType[*runtime.APIError](err); ok {
		return apiErr.Code == http.StatusNotFound
	}

	// Robot V1
	if _, ok := errors.AsType[*robotv1.ListRobotV1NotFound](err); ok {
		return true
	}

	if _, ok := errors.AsType[*robotv1.GetRobotByIDV1NotFound](err); ok {
		return true
	}

	if _, ok := errors.AsType[*robotv1.CreateRobotV1NotFound](err); ok {
		return true
	}

	if _, ok := errors.AsType[*robotv1.UpdateRobotV1NotFound](err); ok {
		return true
	}

	if _, ok := errors.AsType[*robotv1.DeleteRobotV1NotFound](err); ok {
		return true
	}

	// Robot V2
	if _, ok := errors.AsType[*robot.GetRobotByIDNotFound](err); ok {
		return true
	}

	if _, ok := errors.AsType[*robot.ListRobotNotFound](err); ok {
		return true
	}

	if _, ok := errors.AsType[*robot.CreateRobotNotFound](err); ok {
		return true
	}

	if _, ok := errors.AsType[*robot.UpdateRobotNotFound](err); ok {
		return true
	}

	if _, ok := errors.AsType[*robot.DeleteRobotNotFound](err); ok {
		return true
	}

	// Repository
	if _, ok := errors.AsType[*repository.GetRepositoryNotFound](err); ok {
		return true
	}

	if _, ok := errors.AsType[*repository.ListRepositoriesNotFound](err); ok {
		return true
	}

	if _, ok := errors.AsType[*repository.UpdateRepositoryNotFound](err); ok {
		return true
	}

	if _, ok := errors.AsType[*repository.DeleteRepositoryNotFound](err); ok {
		return true
	}

	// Webhook
	if _, ok := errors.AsType[*webhook.GetWebhookPolicyOfProjectNotFound](err); ok {
		return true
	}

	if _, ok := errors.AsType[*webhook.UpdateWebhookPolicyOfProjectNotFound](err); ok {
		return true
	}

	if _, ok := errors.AsType[*webhook.DeleteWebhookPolicyOfProjectNotFound](err); ok {
		return true
	}

	// Project
	if _, ok := errors.AsType[*project.UpdateProjectNotFound](err); ok {
		return true
	}

	if _, ok := errors.AsType[*project.DeleteProjectNotFound](err); ok {
		return true
	}

	// Fallback
	if strings.Contains(strings.ToLower(err.Error()), "not found") || strings.Contains(strings.ToLower(err.Error()), "404") {
		return true
	}

	return false
}

// doJSONRequest is a helper function to do a JSON request to the Harbor API.
// func (client *Client) doJSONRequest(method string, relativePath string, requestBody interface{}) (*http.Response, error) {
// 	jsonBody, err := json.Marshal(requestBody)
// 	if err != nil {
// 		return nil, err
// 	}

// 	fullURL := fmt.Sprintf("%s%s", client.url, relativePath)
// 	return requestutils.DoRequestBasicAuth(method, fullURL, jsonBody, nil, client.username, client.password)
// }

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
