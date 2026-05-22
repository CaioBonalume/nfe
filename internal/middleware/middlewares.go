package middleware

import "github.com/go-chi/cors"

var CORSMiddleware = cors.Handler(cors.Options{
	AllowedOrigins:   []string{"*"},
	AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Requested-With", "User-Agent"},
	AllowCredentials: true,
	MaxAge:           300,
})
