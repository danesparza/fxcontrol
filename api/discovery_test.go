package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danesparza/fxcontrol/internal/discovery"
	"github.com/danesparza/fxcontrol/internal/model"
	"github.com/danesparza/fxcontrol/version"
)

func TestDiscoveryRoutes(t *testing.T) {
	for _, cache := range []DiscoveryCache{nil, &discovery.Cache{}} {
		for _, tc := range []struct {
			method, path string
			status       int
		}{
			{http.MethodGet, "/discover/", http.StatusNotFound},
			{http.MethodGet, "/v1/discover/", http.StatusOK},
			{http.MethodPost, "/discover/", http.StatusNotFound},
			{http.MethodPost, "/v1/discover/", http.StatusMethodNotAllowed},
		} {
			t.Run(tc.method+tc.path, func(t *testing.T) {
				router := NewRouter(Service{Discovery: cache})
				response := httptest.NewRecorder()
				router.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), tc.method, tc.path, nil))
				if response.Code != tc.status {
					t.Fatalf("status %d, want %d", response.Code, tc.status)
				}
				if tc.status != http.StatusOK {
					return
				}
				if got := response.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
					t.Fatalf("content type %q", got)
				}
				if got := response.Header().Get(version.Header); got != version.String() {
					t.Fatalf("version %q", got)
				}
				var body DiscoveryResponse
				if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if body.Message == "" || body.Data == nil || len(body.Data) != 0 {
					t.Fatalf("unexpected response: %s", response.Body)
				}
			})
		}
	}
}

func TestDiscoveryReturnsCachedServices(t *testing.T) {
	for _, path := range []string{"/v1/discover/"} {
		t.Run(path, func(t *testing.T) {
			cache := NewMockDiscoveryCache(t)
			cache.EXPECT().Snapshot().Return([]model.Service{{ID: "porch", Service: "fxpixel", Addresses: []string{"192.168.1.10"}, Port: 3030}}).Once()
			response := httptest.NewRecorder()
			NewRouter(Service{Discovery: cache}).ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil))
			var body DiscoveryResponse
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if response.Code != http.StatusOK || len(body.Data) != 1 || body.Data[0].ID != "porch" || body.Data[0].Service != "fxpixel" || body.Data[0].Port != 3030 || body.Data[0].Addresses[0] != "192.168.1.10" {
				t.Fatalf("unexpected response: %s", response.Body)
			}
		})
	}
}
