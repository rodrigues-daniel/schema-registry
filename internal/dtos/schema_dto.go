package dtos

import (
	"encoding/json"
	"time"

	"github.com/rodrigues-daniel/data-platform/internal/models"
)

// Request DTOs - Recebem JSON direto

type CreateSchemaRequest struct {
	Subject    string             `json:"subject"`
	SchemaType string             `json:"schema_type"`
	Schema     json.RawMessage    `json:"schema"`
	References []models.Reference `json:"references,omitempty"`
	Metadata   map[string]string  `json:"metadata,omitempty"`
}

type UpdateSchemaRequest struct {
	Schema     json.RawMessage   `json:"schema" validate:"required"`
	References []Reference       `json:"references,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type CompatibilityCheckRequest struct {
	Schema json.RawMessage `json:"schema" validate:"required"`
}

type ValidateDataRequest struct {
	Schema json.RawMessage `json:"schema,omitempty"`
	Data   interface{}     `json:"data" validate:"required"`
}

// Config DTOs
type SchemaConfigRequest struct {
	Compatibility string `json:"compatibility" validate:"required,oneof=BACKWARD FORWARD FULL NONE"`
}

// Response DTOs
type SchemaResponse struct {
	ID         string            `json:"id"`
	Subject    string            `json:"subject"`
	Version    int               `json:"version"`
	Schema     json.RawMessage   `json:"schema"`
	SchemaType string            `json:"schema_type"`
	References []Reference       `json:"references,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

type SchemaListResponse struct {
	Schemas []SchemaResponse `json:"schemas"`
	Total   int              `json:"total"`
}

type CompatibilityResponse struct {
	IsCompatible bool     `json:"is_compatible"`
	Errors       []string `json:"errors,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
}

type ValidationResponse struct {
	IsValid  bool     `json:"is_valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// Common structs
type Reference struct {
	Name    string `json:"name"`
	Subject string `json:"subject"`
	Version int    `json:"version"`
}

////////////////////////
