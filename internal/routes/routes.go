package routes

import (

	"fem/internal/app"
	"github.com/go-chi/chi/v5"
)

// ! Handler in server configuration --> directs all req to SetupRoutes 
func SetupRoutes(app *app.Application) *chi.Mux {

	// * handler is defined and which process the client's req
	r := chi.NewRouter()

	// routes
	r.Get("/health",app.HealthCheck)
	return r

}
