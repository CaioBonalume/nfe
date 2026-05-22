package api

import (
	"nfe/internal/nfe"

	nfe_middleware "nfe/internal/middleware"

	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
)

func SetupRoutes(router *chi.Mux, h Handlers) {
	router.Use(chi_middleware.Logger)
	router.Use(chi_middleware.Recoverer)
	router.Use(chi_middleware.RequestID)
	router.Use(chi_middleware.Timeout(60))
	router.Use(nfe_middleware.CORSMiddleware)

	router.Route("/v1", func(r chi.Router) {
		nfe.SetupRoutes(r, h.NFE)
	})
}
