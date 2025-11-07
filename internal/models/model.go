package models

import (
	"fmt"
	"time"
)

// Schema representa um schema no registry
type Schema struct {
	ID         string            `json:"id"`
	Subject    string            `json:"subject"`
	Version    int               `json:"version"`
	Schema     string            `json:"schema"`
	SchemaType string            `json:"schema_type"` // AVRO, JSON, PROTOBUF
	References []Reference       `json:"references,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// Reference representa dependências entre schemas
type Reference struct {
	Name    string `json:"name"`
	Subject string `json:"subject"`
	Version int    `json:"version"`
}

// SchemaConfig configuração de compatibilidade
type SchemaConfig struct {
	Subject       string `json:"subject"`
	Compatibility string `json:"compatibility"` // BACKWARD, FORWARD, FULL, NONE
}

// SchemaValidationRequest pedido de validação
type SchemaValidationRequest struct {
	Subject string      `json:"subject"`
	Schema  string      `json:"schema"`
	Data    interface{} `json:"data,omitempty"`
}

// SchemaValidationResult resultado da validação
type SchemaValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// SchemaResponse resposta da API
type SchemaResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Eventos para o JetStream
type SchemaEvent struct {
	Type      string                 `json:"type"` // SCHEMA_CREATED, SCHEMA_UPDATED, SCHEMA_DELETED
	Subject   string                 `json:"subject"`
	Version   int                    `json:"version"`
	SchemaID  string                 `json:"schema_id"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Constants
const (
	SchemaTypeAVRO     = "AVRO"
	SchemaTypeJSON     = "JSON"
	SchemaTypeProtobuf = "PROTOBUF"

	CompatibilityBackward = "BACKWARD"
	CompatibilityForward  = "FORWARD"
	CompatibilityFull     = "FULL"
	CompatibilityNone     = "NONE"
)

// Validações
func (s *Schema) Validate() error {
	if s.Subject == "" {
		return fmt.Errorf("subject is required")
	}
	if s.Schema == "" {
		return fmt.Errorf("schema content is required")
	}
	if s.SchemaType == "" {
		s.SchemaType = SchemaTypeJSON
	}

	validTypes := map[string]bool{
		SchemaTypeAVRO:     true,
		SchemaTypeJSON:     true,
		SchemaTypeProtobuf: true,
	}

	if !validTypes[s.SchemaType] {
		return fmt.Errorf("invalid schema type: %s", s.SchemaType)
	}

	return nil
}

func (c *SchemaConfig) Validate() error {
	validCompatibilities := map[string]bool{
		CompatibilityBackward: true,
		CompatibilityForward:  true,
		CompatibilityFull:     true,
		CompatibilityNone:     true,
	}

	if !validCompatibilities[c.Compatibility] {
		return fmt.Errorf("invalid compatibility: %s", c.Compatibility)
	}

	return nil
}
