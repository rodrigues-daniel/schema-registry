package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/rodrigues-daniel/data-platform/internal/models"
)

type Validator struct {
	storage *Storage
}

func NewValidator(storage *Storage) *Validator {
	return &Validator{storage: storage}
}

// ValidateSchema valida a sintaxe do schema
func (v *Validator) ValidateSchema(schema *models.Schema) *models.SchemaValidationResult {
	result := &models.SchemaValidationResult{Valid: true}

	// Validações básicas
	if schema.Subject == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "Subject cannot be empty")
	}

	if !isValidSubject(schema.Subject) {
		result.Valid = false
		result.Errors = append(result.Errors, "Invalid subject format")
	}

	if schema.Schema == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "Schema content cannot be empty")
	}

	// Validações específicas por tipo
	switch schema.SchemaType {
	case models.SchemaTypeJSON:
		if err := v.validateJSONSchema(schema.Schema); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("JSON Schema validation failed: %v", err))
		}
	case models.SchemaTypeAVRO:
		if err := v.validateAvroSchema(schema.Schema); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Avro Schema validation failed: %v", err))
		}
	case models.SchemaTypeProtobuf:
		if err := v.validateProtobufSchema(schema.Schema); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Protobuf Schema validation failed: %v", err))
		}
	}

	return result
}

// ValidateCompatibility valida compatibilidade com versão anterior
func (v *Validator) ValidateCompatibility(ctx context.Context, newSchema *models.Schema) *models.SchemaValidationResult {
	result := &models.SchemaValidationResult{Valid: true}

	// Obter configuração de compatibilidade
	config, err := v.storage.GetConfig(ctx, newSchema.Subject)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to get compatibility config: %v", err))
		return result
	}

	// Se não há requirement de compatibilidade, retornar válido
	if config.Compatibility == models.CompatibilityNone {
		return result
	}

	// Tentar obter schema anterior
	previousSchema, err := v.storage.GetLatestSchema(ctx, newSchema.Subject)
	if err != nil {
		// Se não há schema anterior, é compatível por definição
		return result
	}

	// Validar compatibilidade baseada no tipo
	switch config.Compatibility {
	case models.CompatibilityBackward:
		return v.validateBackwardCompatibility(previousSchema, newSchema)
	case models.CompatibilityForward:
		return v.validateForwardCompatibility(previousSchema, newSchema)
	case models.CompatibilityFull:
		backwardResult := v.validateBackwardCompatibility(previousSchema, newSchema)
		forwardResult := v.validateForwardCompatibility(previousSchema, newSchema)

		if !backwardResult.Valid || !forwardResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, backwardResult.Errors...)
			result.Errors = append(result.Errors, forwardResult.Errors...)
		}
	}

	return result
}

// ValidateData valida dados contra um schema
func (v *Validator) ValidateData(ctx context.Context, subject string, version int, data interface{}) *models.SchemaValidationResult {
	result := &models.SchemaValidationResult{Valid: true}

	// Obter schema
	schema, err := v.storage.GetSchema(ctx, subject, version)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Schema not found: %v", err))
		return result
	}

	// Validar dados baseado no tipo de schema
	switch schema.SchemaType {
	case models.SchemaTypeJSON:
		if err := v.validateJSONData(schema.Schema, data); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("JSON data validation failed: %v", err))
		}
	default:
		result.Warnings = append(result.Warnings, "Data validation not implemented for this schema type")
	}

	return result
}

// Validações específicas de implementação
func (v *Validator) validateJSONSchema(schemaContent string) error {
	var schema map[string]interface{}
	return json.Unmarshal([]byte(schemaContent), &schema)
}

func (v *Validator) validateAvroSchema(schemaContent string) error {
	// Implementação básica - em produção usar library Avro
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(schemaContent), &schema); err != nil {
		return err
	}

	// Verificar campos obrigatórios do Avro
	if _, ok := schema["type"]; !ok {
		return fmt.Errorf("avro schema must have 'type' field")
	}

	return nil
}

func (v *Validator) validateProtobufSchema(schemaContent string) error {
	// Verificação básica de sintaxe Protobuf
	if !strings.Contains(schemaContent, "syntax") {
		return fmt.Errorf("protobuf schema must specify syntax version")
	}
	if !strings.Contains(schemaContent, "message") {
		return fmt.Errorf("protobuf schema must contain at least one message")
	}
	return nil
}

func (v *Validator) validateJSONData(schemaContent string, data interface{}) error {
	// Em produção, usar library como go-jsonschema
	// Esta é uma implementação simplificada
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("invalid data format: %v", err)
	}

	var parsedData interface{}
	if err := json.Unmarshal(dataBytes, &parsedData); err != nil {
		return fmt.Errorf("invalid JSON data: %v", err)
	}

	return nil
}

func (v *Validator) validateBackwardCompatibility(oldSchema, newSchema *models.Schema) *models.SchemaValidationResult {
	result := &models.SchemaValidationResult{Valid: true}
	// Implementar lógica de compatibilidade backward
	// Novo schema pode ler dados escritos com old schema
	result.Warnings = append(result.Warnings, "Backward compatibility check not fully implemented")
	return result
}

func (v *Validator) validateForwardCompatibility(oldSchema, newSchema *models.Schema) *models.SchemaValidationResult {
	result := &models.SchemaValidationResult{Valid: true}
	// Implementar lógica de compatibilidade forward
	// Old schema pode ler dados escritos com novo schema
	result.Warnings = append(result.Warnings, "Forward compatibility check not fully implemented")
	return result
}

func isValidSubject(subject string) bool {
	// Validar formato do subject (ex: team.service.entity)
	pattern := `^[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)*$`
	matched, _ := regexp.MatchString(pattern, subject)
	return matched
}
