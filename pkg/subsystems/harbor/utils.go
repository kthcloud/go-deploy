package harbor

import (
	"fmt"
	"go-deploy/pkg/subsystems/harbor/models"
	"go-deploy/utils/requestutils"
	"io"
)

func getRobotFullName(projectName, name string) string {
	return fmt.Sprintf("robot$%s", getRobotName(projectName, name))
}

func getRobotName(projectName, name string) string {
	return fmt.Sprintf("%s+%s", projectName, name)
}

func makeApiError(readCloser io.ReadCloser, makeError func(error) error) error {
	body, err := requestutils.ReadBody(readCloser)
	if err != nil {
		return makeError(err)
	}
	defer requestutils.CloseBody(readCloser)

	apiError := models.ApiError{}
	err = requestutils.ParseJson(body, &apiError)
	if err != nil {
		return makeError(err)
	}

	if len(apiError.Errors) == 0 {
		requestError := fmt.Errorf("erroneous request. details: unknown")
		return makeError(requestError)
	}

	resCode := apiError.Errors[0].Code
	resMsg := apiError.Errors[0].Message
	requestError := fmt.Errorf("erroneous request (%s). details: %w", resCode, resMsg)
	return makeError(requestError)
}
