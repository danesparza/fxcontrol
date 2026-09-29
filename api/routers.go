package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(apiService Service) http.Handler {
	//	Create a router and set up our REST endpoints...
	r := chi.NewRouter()

	//	Add middleware
	r.Use(middleware.RequestID)
	// Keep the socket peer address; forwarded IP headers are not trusted.
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(ApiVersionMiddleware)

	//	... including CORS middleware
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Route("/v1", func(r chi.Router) {

		//	System config
		r.Route("/config", func(r chi.Router) {
			r.Get("/", apiService.ShowUI)       // Get all system config keys and values
			r.Post("/{key}", apiService.ShowUI) // Update system config value
		})

		//	Discovery management
		r.Route("/discover", func(r chi.Router) {
			r.Get("/", apiService.ShowUI) // Discover fx devices
		})
	})

	//	SWAGGER
	r.Mount("/swagger", httpSwagger.WrapHandler)

	return r
}
