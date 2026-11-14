package v2

import (
	_ "embed"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files/v2"
)

//go:embed openapi.yaml
var OpenAPISpec []byte

// ServeDocs serves the embedded OpenAPI specification and Swagger UI at basePath.
func ServeDocs(basePath string) gin.HandlerFunc {
	specURL := strings.TrimRight(basePath, "/") + "/openapi.yaml"
	indexHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>go-deploy API</title>
  <link rel="stylesheet" type="text/css" href="./swagger-ui.css" />
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

	fileServer := http.StripPrefix(strings.TrimRight(basePath, "/"), http.FileServer(http.FS(swaggerfiles.FS)))

	return func(c *gin.Context) {
		p := strings.TrimPrefix(c.Request.URL.Path, strings.TrimRight(basePath, "/"))
		switch p {
		case "", "/", "/index.html":
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(indexHTML))
		case "/openapi.yaml":
			c.Data(http.StatusOK, "application/yaml; charset=utf-8", OpenAPISpec)
		default:
			fileServer.ServeHTTP(c.Writer, c.Request)
		}
	}
}
