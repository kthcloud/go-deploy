package npm

import (
	"fmt"
	"go-deploy/pkg/conf"
	"go-deploy/utils/requestutils"
)

func makeApiError(body []byte, makeError func(error) error) error {
	apiError := npmApiError{}
	err := requestutils.ParseJsonBody(body, &apiError)
	if err != nil {
		return makeError(err)
	}

	resCode := apiError.Error.Code
	resMsg := apiError.Error.Message
	errorMessage := fmt.Sprintf("erroneous request (%d). details: %s", resCode, resMsg)
	return makeError(fmt.Errorf(errorMessage))
}

func getFqdn(name string) string {
	return fmt.Sprintf("%s.%s", name, conf.Env.ParentDomain)
}
