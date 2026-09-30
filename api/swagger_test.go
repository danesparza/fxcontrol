package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/danesparza/fxcontrol/docs"
)

func TestSwagger(t *testing.T) {
	router := NewRouter(Service{})
	for _, tc := range []struct{ path, contentType, content string }{
		{"/swagger/index.html", "text/html", "SwaggerUIBundle"},
		{"/swagger/swagger-ui.css", "text/css", ".swagger-ui"},
		{"/swagger/doc.json", "application/json", `"swagger": "2.0"`},
	} {
		t.Run(tc.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.path, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("status %d: %s", response.Code, response.Body)
			}
			if got := response.Header().Get("Content-Type"); !strings.Contains(got, tc.contentType) {
				t.Fatalf("content type %q, want %q", got, tc.contentType)
			}
			if !strings.Contains(response.Body.String(), tc.content) {
				t.Fatalf("response missing %q", tc.content)
			}
		})
	}
}

func TestSwaggerDiscoveryRoute(t *testing.T) {
	router := NewRouter(Service{})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/swagger/doc.json", nil))
	var document struct {
		BasePath string                                `json:"basePath"`
		Paths    map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &document); err != nil {
		t.Fatal(err)
	}
	if _, ok := document.Paths["/discover/"]["get"]; !ok {
		t.Fatal("discovery GET is missing from Swagger")
	}
	path := document.BasePath + "/discover/"
	if path != "/v1/discover/" {
		t.Fatalf("incorrect discovery URL %q", path)
	}
	result := httptest.NewRecorder()
	router.ServeHTTP(result, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil))
	if result.Code != http.StatusOK {
		t.Fatalf("documented route returned %d", result.Code)
	}
}
