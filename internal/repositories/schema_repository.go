package repositories

import (
	"database/sql"

	"github.com/rodrigues-daniel/schema-registry/internal/models"
)

type SchemaRepository interface {
	CreateSchema(schema *models.JSONSchema) error
	GetSchemaByName(name string) (*models.JSONSchema, error)
	GetSchemaByNameAndVersion(name, version string) (*models.JSONSchema, error)
	GetActiveSchemas() ([]models.JSONSchema, error)
	UpdateSchema(schema *models.JSONSchema) error
	DeactivateSchema(name string) error
}

type schemaRepository struct {
	db *sql.DB
}

func NewSchemaRepository(db *sql.DB) SchemaRepository {
	return &schemaRepository{db: db}
}

func (r *schemaRepository) CreateSchema(schema *models.JSONSchema) error {
	query := `
        INSERT INTO json_schemas (name, version, schema, description, is_active)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at, updated_at
    `

	return r.db.QueryRow(
		query,
		schema.Name,
		schema.Version,
		schema.Schema,
		schema.Description,
		schema.IsActive,
	).Scan(&schema.ID, &schema.CreatedAt, &schema.UpdatedAt)
}

func (r *schemaRepository) GetSchemaByName(name string) (*models.JSONSchema, error) {
	query := `
        SELECT id, name, version, schema, description, is_active, created_at, updated_at
        FROM json_schemas 
        WHERE name = $1 AND is_active = true
        ORDER BY created_at DESC
        LIMIT 1
    `

	var schema models.JSONSchema
	err := r.db.QueryRow(query, name).Scan(
		&schema.ID,
		&schema.Name,
		&schema.Version,
		&schema.Schema,
		&schema.Description,
		&schema.IsActive,
		&schema.CreatedAt,
		&schema.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &schema, err
}

func (r *schemaRepository) GetSchemaByNameAndVersion(name, version string) (*models.JSONSchema, error) {
	query := `
        SELECT id, name, version, schema, description, is_active, created_at, updated_at
        FROM json_schemas 
        WHERE name = $1 AND version = $2
    `

	var schema models.JSONSchema
	err := r.db.QueryRow(query, name, version).Scan(
		&schema.ID,
		&schema.Name,
		&schema.Version,
		&schema.Schema,
		&schema.Description,
		&schema.IsActive,
		&schema.CreatedAt,
		&schema.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &schema, err
}

func (r *schemaRepository) GetActiveSchemas() ([]models.JSONSchema, error) {
	query := `
        SELECT id, name, version, schema, description, is_active, created_at, updated_at
        FROM json_schemas 
        WHERE is_active = true
        ORDER BY name, created_at DESC
    `

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []models.JSONSchema
	for rows.Next() {
		var schema models.JSONSchema
		err := rows.Scan(
			&schema.ID,
			&schema.Name,
			&schema.Version,
			&schema.Schema,
			&schema.Description,
			&schema.IsActive,
			&schema.CreatedAt,
			&schema.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, schema)
	}

	return schemas, nil
}

func (r *schemaRepository) UpdateSchema(schema *models.JSONSchema) error {
	return nil
}

func (r *schemaRepository) DeactivateSchema(name string) error {
	return nil
}
