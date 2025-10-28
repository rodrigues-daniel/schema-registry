package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type SchemaHandler struct {
}

func NewSchemaHandler() *SchemaHandler {
	return &SchemaHandler{}
}

func (sh *SchemaHandler) CommpareSchema(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	raw, _ := json.Marshal(body)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully",
		"user":    user,
	})
}

func (sh *SchemaHandler) RegistrySchema(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	fmt.Fprintf(w, "Olá, estou registrando um schema!")

}

func (sh *SchemaHandler) UpdateSchema(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	fmt.Fprintf(w, "Olá, estou atualizando um esquema")

}
