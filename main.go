package main

import (
	"os"

	"github.com/kthcloud/go-deploy/cmd"
)

// @Title go-deploy API
// @Version 1.0
// @Description This is the API explorer for the go-deploy API. You can use it as a reference for the API endpoints.
// @TermsOfService http://swagger.io/terms/
// @Contact.name Support
// @Contact.url https://github.com/kthcloud/go-deploy
// @License.name MIT License
// @License.url https://github.com/kthcloud/go-deploy?tab=MIT-1-ov-file#readme
// @SecurityDefinitions.apikey ApiKeyAuth
// @In header
// @Name X-Api-Key
// @SecurityDefinitions.oauth2.accessCode KeycloakOAuth
// @authorizationUrl https://iam.cloud.cbh.kth.se/realms/cloud/protocol/openid-connect/auth
// @tokenUrl https://iam.cloud.cbh.kth.se/realms/cloud/protocol/openid-connect/token
// @scope.openid openid
// @scope.profile profile
// @scope.email email
func main() {
	options := cmd.ParseFlags()

	deployApp := cmd.Create(options)
	if deployApp == nil {
		panic("failed to start app")
	}
	defer deployApp.Stop()

	quit := make(chan os.Signal)
	<-quit
}
