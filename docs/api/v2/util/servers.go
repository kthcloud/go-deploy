package util

import (
	"fmt"
	"strings"

	docsV2 "github.com/kthcloud/go-deploy/docs/api/v2"

	"github.com/swaggo/swag/v2"
)

type ExtendedSpec struct {
	swag.Spec
	Servers []struct {
		URL         string `json:"url"`
		Description string `json:"description,omitempty"`
	} `json:"servers,omitempty"`
}

func UpdateSwaggerServers(servers ...string) error {
	// Create the server JSON part dynamically
	var serversJSON string
	for _, server := range servers {
		serversJSON += fmt.Sprintf(`
    {
      "url": "%s",
      "description": ""
    },`, server)
	}

	// Remove the last comma (to make it valid JSON) and add it in the correct location
	if len(serversJSON) > 0 {
		serversJSON = serversJSON[:len(serversJSON)-1]
	}

	// Find the last closing brace in the Swagger template and inject servers JSON before it
	swaggerTemplate := docsV2.SwaggerInfoV2.SwaggerTemplate
	closeBraceIndex := strings.LastIndex(swaggerTemplate, "}")
	if closeBraceIndex == -1 {
		return fmt.Errorf("could not find closing brace in Swagger template")
	}

	// Insert the servers JSON right before the closing brace
	swaggerTemplate = swaggerTemplate[:closeBraceIndex] + ",\n  \"servers\": [" + serversJSON + "]" + swaggerTemplate[closeBraceIndex:]

	// Update the Swagger template
	docsV2.SwaggerInfoV2.SwaggerTemplate = swaggerTemplate
	return nil
}
