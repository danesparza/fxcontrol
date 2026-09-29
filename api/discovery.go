package api

import (
	"encoding/json"
	"net/http"

	"github.com/danesparza/fxcontrol/internal/model"
	"github.com/rs/zerolog/log"
)

// DiscoveryResponse contains the latest completed discovery snapshot.
type DiscoveryResponse struct {
	Message string          `json:"message"`
	Data    []model.Service `json:"data"`
}

// Discover returns cached services immediately; it never initiates a network scan.
// @Summary List discovered FX services
// @Produce json
// @Success 200 {object} DiscoveryResponse
// @Router /discover/ [get]
func (service Service) Discover(rw http.ResponseWriter, req *http.Request) {
	services := make([]model.Service, 0)
	if service.Discovery != nil {
		services = service.Discovery.Snapshot()
	}
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(rw).Encode(DiscoveryResponse{Message: "Discovered FX services", Data: services}); err != nil {
		log.Error().Err(err).Msg("Unable to write FX discovery response")
	}
}
