package services

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

type SchemaHandlerService interface {
}

type JSONSchemaValidator struct{}

func NewValidatorService() *JSONSchemaValidator {
	return &JSONSchemaValidator{}
}

func (v *JSONSchemaValidator) Validate(data []byte, schemaPath string) error {

	schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)
	documentLoader := gojsonschema.NewBytesLoader(data)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("Erro na validacao do schema: %w", err)
	}

	if !result.Valid() {
		errMsg := ""
		for _, desc := range result.Errors() {
			errMsg += fmt.Sprintf("- %s\n", desc)
		}
		return fmt.Errorf("JSON Invalido:\n%s", errMsg)
	}

	return nil
}
