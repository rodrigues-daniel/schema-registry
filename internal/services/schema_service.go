package services

import (
	"encoding/json"
	"fmt"

	"github.com/rodrigues-daniel/schema-registry/internal/models"
	"github.com/rodrigues-daniel/schema-registry/internal/repositories"
	"github.com/rodrigues-daniel/schema-registry/internal/validation"

	"github.com/xeipuuv/gojsonschema"
)

type SchemaService struct {
	schemaRepo repositories.SchemaRepository
	validator  *validation.DatabaseSchemaValidator
}

func NewSchemaService(
	schemaRepo repositories.SchemaRepository,
	validator *validation.DatabaseSchemaValidator,
) *SchemaService {
	return &SchemaService{
		schemaRepo: schemaRepo,
		validator:  validator,
	}
}

func (s *SchemaService) CreateSchema(
	name, version, description string,
	schemaData map[string]interface{},
) error {
	// Validar que o schema JSON é válido
	if err := s.validateSchemaSyntax(schemaData); err != nil {
		return fmt.Errorf("invalid schema syntax: %w", err)
	}

	// Converter para JSON
	schemaJSON, err := json.Marshal(schemaData)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	schema := &models.JSONSchema{
		Name:        name,
		Version:     version,
		Schema:      schemaJSON,
		Description: description,
		IsActive:    true,
	}

	if err := s.schemaRepo.CreateSchema(schema); err != nil {
		return fmt.Errorf("failed to save schema: %w", err)
	}

	// Recarregar validadores
	go s.validator.ReloadSchemas()

	return nil
}

func (s *SchemaService) validateSchemaSyntax(schemaData map[string]interface{}) error {
	schemaLoader := gojsonschema.NewGoLoader(schemaData)
	_, err := gojsonschema.NewSchema(schemaLoader)
	return err
}

func (s *SchemaService) DeactivateSchema(name string) error {
	// Buscar schema atual
	schema, err := s.schemaRepo.GetSchemaByName(name)
	if err != nil {
		return err
	}
	if schema == nil {
		return fmt.Errorf("schema %s not found", name)
	}

	// "Desativar" criando nova versão inativa
	schema.IsActive = false
	if err := s.schemaRepo.UpdateSchema(schema); err != nil {
		return err
	}

	// Recarregar validadores
	go s.validator.ReloadSchemas()

	return nil
}

func (s *SchemaService) GetSchema(name string) (*models.JSONSchema, error) {

	schemaResp, err := s.schemaRepo.GetSchemaByName(name)
	if err != nil {
		return nil, err
	}

	return schemaResp, nil

}
