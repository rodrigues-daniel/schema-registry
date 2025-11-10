package schema

import (
	"context"

	"github.com/rodrigues-daniel/data-platform/internal/models"
)

type JetStream interface {
	Publish(subj string, data []byte) error
}

type StorageSchema interface {
	StorageConfig
	StorageCRUD
	StorageLatest
	StorageSubjects
}

type StorageConfig interface {
	SaveConfig(ctx context.Context, config *models.SchemaConfig) error
	GetConfig(ctx context.Context, subject string) (*models.SchemaConfig, error)
}

type StorageCRUD interface {
	SaveSchema(ctx context.Context, schema *models.Schema) error
	GetSchema(ctx context.Context, subject string, version int) (*models.Schema, error)
	DeleteSchema(ctx context.Context, subject string, version int) error
	GetSchemaByID(ctx context.Context, schemaID string) (*models.Schema, error)
}

type StorageLatest interface {
	GetLatestSchema(ctx context.Context, subject string) (*models.Schema, error)
	GetSchemaVersions(ctx context.Context, subject string) ([]int, error)
}

type StorageSubjects interface {
	ListSubjects(ctx context.Context) ([]string, error)
}

type ValidatorSchema interface {
	ValidateSchema(schema *models.Schema) *models.SchemaValidationResult
	ValidateCompatibility(ctx context.Context, newSchema *models.Schema) *models.SchemaValidationResult
	ValidateData(ctx context.Context, subject string, version int, data interface{}) *models.SchemaValidationResult
}
