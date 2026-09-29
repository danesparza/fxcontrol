package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/danesparza/fxcontrol/internal/model"
	"github.com/danesparza/fxcontrol/version"
)

// DiscoveryCache provides the most recently completed background scan.
type DiscoveryCache interface {
	Snapshot() []model.Service
}

// Service holds the shared state used by HTTP handlers.
type Service struct {
	// StartTime records when the HTTP service was initialized.
	StartTime time.Time

	// Discovery holds the latest background network scan.
	Discovery DiscoveryCache
}

// SystemResponse wraps a message and arbitrary system response data.
type SystemResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ErrorResponse contains the message returned for an API error.
type ErrorResponse struct {
	Message string `json:"message"`
}

// sendErrorResponse writes an error message as JSON with the supplied HTTP status.
func sendErrorResponse(rw http.ResponseWriter, err error, code int) {
	response := ErrorResponse{
		Message: "Error: " + err.Error()}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(code)
	json.NewEncoder(rw).Encode(response)
}

// ShowUI writes placeholder text for the unimplemented configuration routes.
func (service Service) ShowUI(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "Hello, world - UI")
}

// ApiVersionMiddleware adds the fxcontrol build version to each response.
func ApiVersionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set(version.Header, version.String())

		next.ServeHTTP(rw, r)
	})
}
