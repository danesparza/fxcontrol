package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/danesparza/fxcontrol/version"
)

// Service encapsulates API service operations
type Service struct {
	// StartTime is the service start time.  We can calculate uptime based on this
	StartTime time.Time

	// DiscoverFxDevices signals discovery process (window) should start
	DiscoverFxDevices chan string

	//	StopDiscoverFxDevices signals discovery should force-stop
	StopDiscoverFxDevices chan bool
}

// SystemResponse is a response for a system request
type SystemResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ErrorResponse represents an API response
type ErrorResponse struct {
	Message string `json:"message"`
}

// Used to send back an error:
func sendErrorResponse(rw http.ResponseWriter, err error, code int) {
	//	Our return value
	response := ErrorResponse{
		Message: "Error: " + err.Error()}

	//	Serialize to JSON & return the response:
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(code)
	json.NewEncoder(rw).Encode(response)
}

// ShowUI redirects to the /ui/ url path
func (service Service) ShowUI(rw http.ResponseWriter, req *http.Request) {
	// http.Redirect(rw, req, "/ui/", 301)
	fmt.Fprintf(rw, "Hello, world - UI")
}

// ApiVersionMiddleware adds the API version information to the response header
func ApiVersionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		//	Include the version in the response headers:
		rw.Header().Set(version.Header, version.String())

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(rw, r)
	})
}
