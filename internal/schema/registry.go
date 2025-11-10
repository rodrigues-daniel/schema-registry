package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rodrigues-daniel/data-platform/internal/models"
)

type Registry struct {
	storage   StorageSchema
	validator ValidatorSchema
	js        JetStream
}

func NewRegistry(
	storageSchema StorageSchema,
	validator ValidatorSchema,
	js JetStream) *Registry {

	return &Registry{
		storage:   storageSchema,
		validator: validator,
		js:        js,
	}
}

// RegisterSchema registra um novo schema
func (r *Registry) RegisterSchema(ctx context.Context, schema *models.Schema) (*models.Schema, error) {
	// Validar schema
	validationResult := r.validator.ValidateSchema(schema)
	if !validationResult.Valid {
		return nil, fmt.Errorf("schema validation failed: %v", validationResult.Errors)
	}

	// Determinar próxima versão
	versions, err := r.storage.GetSchemaVersions(ctx, schema.Subject)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema versions: %w", err)
	}

	schema.Version = 1
	if len(versions) > 0 {
		schema.Version = versions[len(versions)-1] + 1
	}

	// Validar compatibilidade
	compatResult := r.validator.ValidateCompatibility(ctx, schema)
	if !compatResult.Valid {
		return nil, fmt.Errorf("compatibility check failed: %v", compatResult.Errors)
	}

	// Salvar schema
	if err := r.storage.SaveSchema(ctx, schema); err != nil {
		return nil, fmt.Errorf("failed to save schema: %w", err)
	}

	// Publicar evento
	if err := r.publishSchemaEvent(ctx, &models.SchemaEvent{
		Type:      "SCHEMA_CREATED",
		Subject:   schema.Subject,
		Version:   schema.Version,
		SchemaID:  schema.ID,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"schema_type": schema.SchemaType,
		},
	}); err != nil {
		log.Printf("Warning: failed to publish schema event: %v", err)
	}

	log.Printf("Schema registered: %s version %d", schema.Subject, schema.Version)
	return schema, nil
}

// GetSchema obtém um schema
func (r *Registry) GetSchema(ctx context.Context, subject string, version int) (*models.Schema, error) {
	return r.storage.GetSchema(ctx, subject, version)
}

// GetLatestSchema obtém a última versão
func (r *Registry) GetLatestSchema(ctx context.Context, subject string) (*models.Schema, error) {
	return r.storage.GetLatestSchema(ctx, subject)
}

// ListSubjects lista todos os subjects
func (r *Registry) ListSubjects(ctx context.Context) ([]string, error) {
	return r.storage.ListSubjects(ctx)
}

// ListVersions lista versões de um subject
func (r *Registry) ListVersions(ctx context.Context, subject string) ([]int, error) {
	return r.storage.GetSchemaVersions(ctx, subject)
}

// SetConfig define configuração de compatibilidade
func (r *Registry) SetConfig(ctx context.Context, config *models.SchemaConfig) error {
	return r.storage.SaveConfig(ctx, config)
}

// GetConfig obtém configuração
func (r *Registry) GetConfig(ctx context.Context, subject string) (*models.SchemaConfig, error) {
	return r.storage.GetConfig(ctx, subject)
}

// ValidateData valida dados contra schema
func (r *Registry) ValidateData(ctx context.Context, req *models.SchemaValidationRequest) (*models.SchemaValidationResult, error) {
	schema, err := r.storage.GetLatestSchema(ctx, req.Subject)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	result := r.validator.ValidateData(ctx, req.Subject, schema.Version, req.Data)
	return result, nil
}

// CheckCompatibility verifica compatibilidade entre schemas
func (r *Registry) CheckCompatibility(ctx context.Context, subject string, schemaContent string) (*models.SchemaValidationResult, error) {
	tempSchema := &models.Schema{
		Subject:    subject,
		Schema:     schemaContent,
		SchemaType: models.SchemaTypeJSON, // Ou detectar automaticamente
	}

	result := r.validator.ValidateCompatibility(ctx, tempSchema)
	return result, nil
}

// DeleteSchema deleta um schema
func (r *Registry) DeleteSchema(ctx context.Context, subject string, version int) error {
	schema, err := r.storage.GetSchema(ctx, subject, version)
	if err != nil {
		return err
	}

	if err := r.storage.DeleteSchema(ctx, subject, version); err != nil {
		return err
	}

	// Publicar evento de deleção
	r.publishSchemaEvent(ctx, &models.SchemaEvent{
		Type:      "SCHEMA_DELETED",
		Subject:   subject,
		Version:   version,
		SchemaID:  schema.ID,
		Timestamp: time.Now(),
	})

	return nil
}

// GetSchemaByID obtém schema por ID
func (r *Registry) GetSchemaByID(ctx context.Context, schemaID string) (*models.Schema, error) {
	return r.storage.GetSchemaByID(ctx, schemaID)
}

func (r *Registry) publishSchemaEvent(ctx context.Context, event *models.SchemaEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("schema.events.%s", event.Subject)
	err = r.js.Publish(subject, data)
	return err
}
