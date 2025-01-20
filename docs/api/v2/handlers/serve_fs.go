package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/v2"

	swaggerfiles "github.com/swaggo/files/v2"
)

// Returns a handler that serves the requested any file
// from swaggo/files/v2 embed FS with swagger UI and
// the generated doc.json (V2_swagger.json)
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
			ctx.Header("Cache-Control", "public, max-age=31536000")
			ctx.Data(http.StatusOK, "application/javascript", []byte(customScript))
			return
		}

		if strings.TrimPrefix(filePath, "/") == "doc.json" {
			doc, err := swag.ReadDoc("V2")
			if err != nil {
				ctx.String(http.StatusInternalServerError, "Failed to load documentation")
				return
			}
			ctx.Header("Cache-Control", "public, max-age=3600")
			ctx.Data(http.StatusOK, "application/json", []byte(doc))
			return
		}

		ctx.Request.URL.Path = filePath
		ctx.Header("Cache-Control", "public, max-age=31536000")
		http.FileServer(http.FS(swaggerfiles.FS)).ServeHTTP(ctx.Writer, ctx.Request)
	}
}
