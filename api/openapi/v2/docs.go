package v2

import (
	"bytes"
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	defaultKeycloakAuthURL  = "https://iam.cloud.cbh.kth.se/realms/cloud/protocol/openid-connect/auth"
	defaultKeycloakTokenURL = "https://iam.cloud.cbh.kth.se/realms/cloud/protocol/openid-connect/token"
)

//go:embed openapi.yaml
var OpenAPISpec []byte

//go:embed swagger-ui/*
var swaggerUIDist embed.FS

// DocsOptions configures optional overrides for the served OpenAPI specification.
type DocsOptions struct {
	KeycloakAuthURL  string
	KeycloakTokenURL string
}

// ServeDocs serves the embedded OpenAPI specification and Swagger UI at basePath.
func ServeDocs(basePath string, opts ...DocsOptions) gin.HandlerFunc {
	spec := applyDocsOptions(OpenAPISpec, opts...)
	specURL := strings.TrimRight(basePath, "/") + "/openapi.yaml"
	indexHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>go-deploy API</title>
  <link rel="stylesheet" type="text/css" href="./swagger-ui.css" />
  <link rel="stylesheet" type="text/css" href="./index.css" />
  <link rel="icon" type="image/png" href="./favicon-32x32.png" sizes="32x32" />
  <link rel="icon" type="image/png" href="./favicon-16x16.png" sizes="16x16" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="./swagger-ui-bundle.js"></script>
  <script src="./swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      SwaggerUIBundle({
        url: "` + specURL + `",
        dom_id: '#swagger-ui',
        presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
        layout: "StandaloneLayout"
      });
    };
  </script>
</body>
</html>`

	subFS, err := fs.Sub(swaggerUIDist, "swagger-ui")
	if err != nil {
		panic(err)
	}
	fileServer := http.StripPrefix(strings.TrimRight(basePath, "/"), http.FileServer(http.FS(subFS)))

	return func(c *gin.Context) {
		p := strings.TrimPrefix(c.Request.URL.Path, strings.TrimRight(basePath, "/"))
		switch p {
		case "", "/", "/index.html":
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(indexHTML))
		case "/openapi.yaml":
			c.Data(http.StatusOK, "application/yaml; charset=utf-8", spec)
		default:
			fileServer.ServeHTTP(c.Writer, c.Request)
		}
	}
}

func applyDocsOptions(spec []byte, opts ...DocsOptions) []byte {
	if len(opts) == 0 {
		return spec
	}
	opt := opts[0]
	if opt.KeycloakAuthURL == "" && opt.KeycloakTokenURL == "" {
		return spec
	}

	out := spec
	if opt.KeycloakAuthURL != "" {
		out = bytes.Replace(out, []byte("authorizationUrl: "+defaultKeycloakAuthURL), []byte("authorizationUrl: "+opt.KeycloakAuthURL), 1)
	}
	if opt.KeycloakTokenURL != "" {
		out = bytes.Replace(out, []byte("tokenUrl: "+defaultKeycloakTokenURL), []byte("tokenUrl: "+opt.KeycloakTokenURL), 1)
	}
	return out
}
