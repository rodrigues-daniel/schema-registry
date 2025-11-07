package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/rodrigues-daniel/data-platform/internal/dtos"
	"github.com/rodrigues-daniel/data-platform/internal/mappers"
	"github.com/rodrigues-daniel/data-platform/internal/schema"

	"github.com/rodrigues-daniel/data-platform/internal/models"

	"github.com/gorilla/mux"
)

type Handlers struct {
	registry *schema.Registry
}

func NewHandlers(registry *schema.Registry) *Handlers {
	return &Handlers{registry: registry}
}

// RegisterSchemaHandler registra novo schema
func (h *Handlers) RegisterSchemaHandler(w http.ResponseWriter, r *http.Request) {
	var schemaDTO dtos.CreateSchemaRequest
	if err := json.NewDecoder(r.Body).Decode(&schemaDTO); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var schema models.Schema
	schema = mappers.MapCreateSchemaRequestToModel(schemaDTO)
	registeredSchema, err := h.registry.RegisterSchema(r.Context(), &schema)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.sendSuccess(w, http.StatusCreated, registeredSchema)
}

// GetSchemaHandler obtém schema
func (h *Handlers) GetSchemaHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subject := vars["subject"]
	versionStr := vars["version"]

	var schema *models.Schema
	var err error

	if versionStr == "latest" {
		schema, err = h.registry.GetLatestSchema(r.Context(), subject)
	} else {
		version, parseErr := strconv.Atoi(versionStr)
		if parseErr != nil {
			h.sendError(w, http.StatusBadRequest, "Invalid version")
			return
		}
		schema, err = h.registry.GetSchema(r.Context(), subject, version)
	}

	if err != nil {
		h.sendError(w, http.StatusNotFound, err.Error())
		return
	}

	h.sendSuccess(w, http.StatusOK, schema)
}

// ListSubjectsHandler lista subjects
func (h *Handlers) ListSubjectsHandler(w http.ResponseWriter, r *http.Request) {
	subjects, err := h.registry.ListSubjects(r.Context())
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccess(w, http.StatusOK, subjects)
}

// ListVersionsHandler lista versões
func (h *Handlers) ListVersionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subject := vars["subject"]

	versions, err := h.registry.ListVersions(r.Context(), subject)
	if err != nil {
		h.sendError(w, http.StatusNotFound, err.Error())
		return
	}

	h.sendSuccess(w, http.StatusOK, versions)
}

// ConfigHandler gerencia configurações
func (h *Handlers) ConfigHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subject := vars["subject"]

	switch r.Method {
	case "PUT":
		var config models.SchemaConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			h.sendError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		config.Subject = subject

		if err := h.registry.SetConfig(r.Context(), &config); err != nil {
			h.sendError(w, http.StatusBadRequest, err.Error())
			return
		}

		h.sendSuccess(w, http.StatusOK, config)

	case "GET":
		config, err := h.registry.GetConfig(r.Context(), subject)
		if err != nil {
			h.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		h.sendSuccess(w, http.StatusOK, config)

	case "DELETE":
		// Reset para padrão
		defaultConfig := &models.SchemaConfig{
			Subject:       subject,
			Compatibility: models.CompatibilityBackward,
		}
		if err := h.registry.SetConfig(r.Context(), defaultConfig); err != nil {
			h.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		h.sendSuccess(w, http.StatusOK, defaultConfig)
	}
}

// CompatibilityHandler verifica compatibilidade
func (h *Handlers) CompatibilityHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subject := vars["subject"]

	var req struct {
		Schema string `json:"schema"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.registry.CheckCompatibility(r.Context(), subject, req.Schema)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.sendSuccess(w, http.StatusOK, result)
}

// ValidateHandler valida dados
func (h *Handlers) ValidateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subject := vars["subject"]

	var req models.SchemaValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	req.Subject = subject

	result, err := h.registry.ValidateData(r.Context(), &req)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.sendSuccess(w, http.StatusOK, result)
}

func (h *Handlers) sendSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := models.SchemaResponse{
		Success: true,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := models.SchemaResponse{
		Success: false,
		Error:   message,
	}

	json.NewEncoder(w).Encode(response)
}
