package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// NewRouter configures the HTTP middleware, API routes, and Swagger UI.
func NewRouter(apiService Service) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	// Keep the socket peer address; forwarded IP headers are not trusted.
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(ApiVersionMiddleware)

	// Allow browser clients from HTTP and HTTPS origins to access the API.
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Cache successful CORS preflight responses for five minutes.
	}))

	r.Route("/v1", func(r chi.Router) {

		// Configuration routes currently return placeholder text.
		r.Route("/config", func(r chi.Router) {
			r.Get("/", apiService.ShowUI)
			r.Post("/{key}", apiService.ShowUI)
		})

		// Return the most recent background discovery snapshot.
		r.Route("/discover", func(r chi.Router) {
			r.Get("/", apiService.Discover)
		})
	})

	// Serve interactive documentation for the registered API endpoints.
	r.Mount("/swagger", httpSwagger.WrapHandler)

	return r
}
