package domain

import (
	"encoding/json"
	"time"
)

type JSONSchema struct {
	ID          int             `json:"id"`
	Name        string          `json:"name"`
	Version     string          `json:"version"`
	Schema      json.RawMessage `json:"schema"`
	Description string          `json:"description"`
	IsActive    bool            `json:"is_active"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type ValidationResult struct {
	IsValid  bool     `json:"is_valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}
