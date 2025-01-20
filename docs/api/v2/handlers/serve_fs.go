package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/v2"

	swaggerfiles "github.com/swaggo/files/v2"
)

func ServeDocs(basePath string) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		filePath := ctx.Param("any")

		if strings.TrimPrefix(filePath, "/") == "swagger-initializer.js" {
			customScript := fmt.Sprintf(`window.onload = function() {
  window.ui = SwaggerUIBundle({
	url: "%s/doc.json",
	dom_id: '#swagger-ui',
	deepLinking: true,
	presets: [
	  SwaggerUIBundle.presets.apis,
	  SwaggerUIStandalonePreset
	],
	plugins: [
	  SwaggerUIBundle.plugins.DownloadUrl
	],
	layout: "StandaloneLayout"
  });
};`, basePath)
			ctx.Data(http.StatusOK, "application/javascript", []byte(customScript))
			return
		}

		// Serve the Swagger documentation
		if strings.TrimPrefix(filePath, "/") == "doc.json" {
			doc, err := swag.ReadDoc("V2")
			if err != nil {
				ctx.String(http.StatusInternalServerError, "Failed to load documentation")
				return
			}
			ctx.Data(http.StatusOK, "application/json", []byte(doc))
			return
		}

		ctx.Request.URL.Path = filePath
		http.FileServer(http.FS(swaggerfiles.FS)).ServeHTTP(ctx.Writer, ctx.Request)
	}
}
