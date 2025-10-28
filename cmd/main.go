package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigues-daniel/schema-registry/internal/handlers"
)

func main() {

	route := chi.NewRouter()

	h := handlers.NewSchemaHandler()

	route.Post("/schema", h.CommpareSchema)

	http.ListenAndServe(":8080", route)
}
