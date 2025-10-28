package validation

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

type SchemaValidator interface {
	ValidateJSON(schemaName string, document interface{}) (*ValidationResult, error)
	LoadSchema(schemaName, schemaPath string) error
}

type JSONSchemaValidator struct {
	schemas map[string]*gojsonschema.Schema
}

type ValidationResult struct {
	IsValid  bool
	Errors   []string
	Warnings []string
}

func NewJSONSchemaValidator() *JSONSchemaValidator {
	return &JSONSchemaValidator{
		schemas: make(map[string]*gojsonschema.Schema),
	}
}

func (v *JSONSchemaValidator) LoadSchema(schemaName, schemaPath string) error {
	schemaLoader := gojsonschema.NewReferenceLoader(schemaPath)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return fmt.Errorf("falha ao carregar o schema %s: %w", schemaName, err)
	}

	v.schemas[schemaName] = schema
	return nil
}

func (v *JSONSchemaValidator) ValidateJSON(schemaName string, document interface{}) (*ValidationResult, error) {
	schema, exists := v.schemas[schemaName]
	if !exists {
		return nil, fmt.Errorf("schema %s not loaded", schemaName)
	}

	documentLoader := gojsonschema.NewGoLoader(document)
	result, err := schema.Validate(documentLoader)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	validationResult := &ValidationResult{
		IsValid: result.Valid(),
	}

	if !result.Valid() {
		for _, desc := range result.Errors() {
			validationResult.Errors = append(validationResult.Errors, desc.String())
		}
	}

	return validationResult, nil
}
