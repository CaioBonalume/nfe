package main

import (
	"log"
	"net/http"
	"nfe/config"

	"nfe/internal/api"
	"nfe/internal/nfe"

	"github.com/go-chi/chi/v5"
)

func main() {
	config.LoadEnv()

	// Wiring
	nfeSvc := nfe.NewNFEService()
	nfeH := nfe.NewNFEHandler(nfeSvc)

	h := api.Handlers{
		NFE: nfeH,
	}

	router := chi.NewRouter()
	api.SetupRoutes(router, h)

	addr := ":8080"
	log.Printf("Servidor iniciado em http://localhost%s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
