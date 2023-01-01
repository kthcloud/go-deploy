package npm

import (
	"fmt"
	"go-deploy/pkg/subsystems/npm/models"
	"go-deploy/utils/requestutils"
	"io"
)

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

	resCode := apiError.Error.Code
	resMsg := apiError.Error.Message
	errorMessage := fmt.Sprintf("erroneous request (%d). details: %s", resCode, resMsg)
	return makeError(fmt.Errorf(errorMessage))
}
