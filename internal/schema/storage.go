package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/rodrigues-daniel/data-platform/internal/models"

	"github.com/nats-io/nats.go"
)

type Storage struct {
	kv nats.KeyValue
}

func NewStorage(kv nats.KeyValue) *Storage {
	return &Storage{kv: kv}
}

// SaveSchema salva um schema
func (s *Storage) SaveSchema(ctx context.Context, schema *models.Schema) error {
	if err := schema.Validate(); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	// Gerar ID se não existir
	if schema.ID == "" {
		schema.ID = generateSchemaID(schema.Subject, schema.Version)
	}

	schema.UpdatedAt = time.Now()
	if schema.CreatedAt.IsZero() {
		schema.CreatedAt = schema.UpdatedAt
	}

	// Serializar schema
	data, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	// Salvar no KV store
	key := fmt.Sprintf("schemas.%s.%d", schema.Subject, schema.Version)
	_, err = s.kv.Put(key, data)
	if err != nil {
		return fmt.Errorf("failed to save schema: %w", err)
	}

	// Salvar metadata separadamente para busca rápida
	metadataKey := fmt.Sprintf("metadata.%s", schema.ID)
	metadata := map[string]interface{}{
		"subject":    schema.Subject,
		"version":    schema.Version,
		"type":       schema.SchemaType,
		"created_at": schema.CreatedAt,
	}

	metadataData, _ := json.Marshal(metadata)
	s.kv.Put(metadataKey, metadataData)

	log.Printf("Schema saved: %s version %d", schema.Subject, schema.Version)
	return nil
}

// GetSchema obtém um schema específico
func (s *Storage) GetSchema(ctx context.Context, subject string, version int) (*models.Schema, error) {
	key := fmt.Sprintf("schemas.%s.%d", subject, version)

	entry, err := s.kv.Get(key)
	if err != nil {
		if err == nats.ErrKeyNotFound {
			return nil, fmt.Errorf("schema not found: %s version %d", subject, version)
		}
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	var schema models.Schema
	if err := json.Unmarshal(entry.Value(), &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	return &schema, nil
}

// GetLatestSchema obtém a última versão de um schema
func (s *Storage) GetLatestSchema(ctx context.Context, subject string) (*models.Schema, error) {
	// Listar todas as versões
	versions, err := s.GetSchemaVersions(ctx, subject)
	if err != nil {
		return nil, err
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no schemas found for subject: %s", subject)
	}

	// Encontrar a versão mais recente
	latestVersion := 0
	for _, v := range versions {
		if v > latestVersion {
			latestVersion = v
		}
	}

	return s.GetSchema(ctx, subject, latestVersion)
}

// GetSchemaVersions lista todas as versões de um subject
func (s *Storage) GetSchemaVersions(ctx context.Context, subject string) ([]int, error) {
	// prefix := fmt.Sprintf("schemas.%s.", subject)

	// keys, err := s.kv.Keys()
	// if err != nil {
	// 	if err == nats.ErrNoKeysFound {
	// 		return []int{}, nil
	// 	}
	// 	return nil, fmt.Errorf("failed to list keys: %w", err)
	// }

	// var versions []int
	// for _, key := range keys {
	// 	if len(key) > len(prefix) && key[:len(prefix)] == prefix {
	// 		var version int
	// 		_, err := fmt.Sscanf(key, "schemas:%s:%d", &subject, &version)
	// 		if err == nil {
	// 			versions = append(versions, version)
	// 		}
	// 	}
	// }

	versions = []int{0}
	sort.Ints(versions)
	return versions, nil
}

// ListSubjects lista todos os subjects
func (s *Storage) ListSubjects(ctx context.Context) ([]string, error) {
	keys, err := s.kv.Keys()
	if err != nil {
		if err == nats.ErrNoKeysFound {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}

	subjects := make(map[string]bool)
	for _, key := range keys {
		var subject string
		if _, err := fmt.Sscanf(key, "schemas.%s.", &subject); err == nil {
			subjects[subject] = true
		}
	}

	result := make([]string, 0, len(subjects))
	for subject := range subjects {
		result = append(result, subject)
	}

	sort.Strings(result)
	return result, nil
}

// DeleteSchema deleta um schema
func (s *Storage) DeleteSchema(ctx context.Context, subject string, version int) error {
	key := fmt.Sprintf("schemas.%s.%d", subject, version)

	// Obter schema para remover metadata
	schema, err := s.GetSchema(ctx, subject, version)
	if err == nil {
		metadataKey := fmt.Sprintf("metadata.%s", schema.ID)
		s.kv.Delete(metadataKey)
	}

	return s.kv.Delete(key)
}

// SaveConfig salva configuração de compatibilidade
func (s *Storage) SaveConfig(ctx context.Context, config *models.SchemaConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}

	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	key := fmt.Sprintf("subjects.%s.config", config.Subject)
	_, err = s.kv.Put(key, data)
	return err
}

// GetConfig obtém configuração de compatibilidade
func (s *Storage) GetConfig(ctx context.Context, subject string) (*models.SchemaConfig, error) {
	key := fmt.Sprintf("subjects.%s.config", subject)

	entry, err := s.kv.Get(key)
	if err != nil {
		if err == nats.ErrKeyNotFound {
			// Retornar configuração padrão
			return &models.SchemaConfig{
				Subject:       subject,
				Compatibility: models.CompatibilityBackward,
			}, nil
		}
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	var config models.SchemaConfig
	if err := json.Unmarshal(entry.Value(), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// GetSchemaByID obtém schema por ID
func (s *Storage) GetSchemaByID(ctx context.Context, schemaID string) (*models.Schema, error) {
	// Primeiro buscar metadata
	metadataKey := fmt.Sprintf("metadata.%s", schemaID)
	entry, err := s.kv.Get(metadataKey)
	if err != nil {
		return nil, fmt.Errorf("schema not found: %s", schemaID)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(entry.Value(), &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	subject, _ := metadata["subject"].(string)
	version, _ := metadata["version"].(float64)

	return s.GetSchema(ctx, subject, int(version))
}

func generateSchemaID(subject string, version int) string {
	return fmt.Sprintf("%s-%d-%d", subject, version, time.Now().UnixNano())
}
