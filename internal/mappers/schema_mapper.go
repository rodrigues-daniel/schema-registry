package mappers

import (
	"time"

	"github.com/rodrigues-daniel/data-platform/internal/dtos"
	"github.com/rodrigues-daniel/data-platform/internal/models"
)

func MapCreateSchemaRequestToModel(req dtos.CreateSchemaRequest) models.Schema {
	return models.Schema{
		ID:         "",
		Subject:    req.Subject,
		Version:    0,
		Schema:     string(req.Schema),
		SchemaType: req.SchemaType,
		References: req.References,
		Metadata:   req.Metadata,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}
