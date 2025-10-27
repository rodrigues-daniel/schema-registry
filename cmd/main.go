package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigues-daniel/schema-registry/cmd/handlers"
)

func main() {

	r := chi.NewRouter()

	b := handlers.NewSchemaRegistryHandler()
	r.Get("/", b.RequestSchemaInfo)

	http.ListenAndServe(":8080", r)
}
