package domain

import "context"

type SchemaService interface {
	CreateSchema(ctx context.Context, name, version, description string, schemaData map[string]interface{}) error
	GetSchema(ctx context.Context, name string) (*JSONSchema, error)
	Validate(ctx context.Context, schemaName string, document interface{}) (*ValidationResult, error)
	ValidateVersion(ctx context.Context, schemaName, version string, document interface{}) (*ValidationResult, error)
	ListVersions(ctx context.Context, name string) ([]JSONSchema, error)
	MigrateSchema(ctx context.Context, name, newVersion string) error
}

// Validator Interface
type SchemaValidator interface {
	Validate(schemaName string, document interface{}) (*ValidationResult, error)
	ValidateVersion(schemaName, version string, document interface{}) (*ValidationResult, error)
	Reload(ctx context.Context) error
}
