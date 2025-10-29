package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/rodrigues-daniel/schema-registry/internal/services"
)

type SchemaHandler struct {
	schemaService *services.SchemaService
}

func NewSchemaHandler(schemaService *services.SchemaService) *SchemaHandler {
	return &SchemaHandler{
		schemaService: schemaService,
	}
}

func (h *SchemaHandler) CreateSchema(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name        string                 `json:"name"`
		Version     string                 `json:"version"`
		Description string                 `json:"description"`
		Schema      map[string]interface{} `json:"schema"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.schemaService.CreateSchema(
		request.Name,
		request.Version,
		request.Description,
		request.Schema,
	); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Schema created successfully",
	})
}
