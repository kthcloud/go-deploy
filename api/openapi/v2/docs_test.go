package v2

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestServeDocsKeycloakOAuthOverride(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("default uses embedded URLs", func(t *testing.T) {
		r := gin.New()
		r.GET("/v2/docs/*any", ServeDocs("/v2/docs"))

		req := httptest.NewRequest(http.MethodGet, "/v2/docs/openapi.yaml", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		body := w.Body.String()
		if !strings.Contains(body, "authorizationUrl: "+defaultKeycloakAuthURL) {
			t.Fatalf("expected default authorizationUrl in spec")
		}
		if !strings.Contains(body, "tokenUrl: "+defaultKeycloakTokenURL) {
			t.Fatalf("expected default tokenUrl in spec")
		}
	})

	t.Run("overrides KeycloakOAuth URLs", func(t *testing.T) {
		customAuth := "https://auth.example.com/realms/custom/protocol/openid-connect/auth"
		customToken := "https://auth.example.com/realms/custom/protocol/openid-connect/token"

		r := gin.New()
		r.GET("/v2/docs/*any", ServeDocs("/v2/docs", DocsOptions{
			KeycloakAuthURL:  customAuth,
			KeycloakTokenURL: customToken,
		}))

		req := httptest.NewRequest(http.MethodGet, "/v2/docs/openapi.yaml", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		body := w.Body.String()
		if !strings.Contains(body, "authorizationUrl: "+customAuth) {
			t.Fatalf("expected overridden authorizationUrl in spec")
		}
		if !strings.Contains(body, "tokenUrl: "+customToken) {
			t.Fatalf("expected overridden tokenUrl in spec")
		}
		if strings.Contains(body, defaultKeycloakAuthURL) || strings.Contains(body, defaultKeycloakTokenURL) {
			t.Fatalf("expected default Keycloak URLs to be replaced")
		}
	})
}
