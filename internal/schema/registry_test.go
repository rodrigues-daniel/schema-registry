package schema

import (
	"testing"
)

func TestNewRegistry(t *testing.T) {
	tests := []struct {
		name        string
		storage     StorageSchema
		validator   ValidatorSchema
		js          JetStream
		wantStorage bool
		wantErr     bool
	}{
		{
			name:        "should create registry with all dependencies",
			storage:     &mockStorage{},
			validator:   &mockValidator{},
			js:          &mockJetStream{},
			wantStorage: true,
			wantErr:     false,
		},
		{
			name:        "should create registry with nil jetstream",
			storage:     &mockStorage{},
			validator:   &mockValidator{},
			js:          nil,
			wantStorage: true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry(tt.storage, tt.validator, tt.js)

			if registry == nil {
				t.Error("expected non-nil registry")
			}

			if tt.wantStorage && registry.storage == nil {
				t.Error("expected non-nil storage")
			}

			if registry.validator == nil {
				t.Error("expected non-nil validator")
			}

			// Check if dependencies were properly assigned
			if registry.storage != tt.storage {
				t.Error("storage not properly assigned")
			}
			if registry.validator != tt.validator {
				t.Error("validator not properly assigned")
			}
			if registry.js != tt.js {
				t.Error("jetstream not properly assigned")
			}
		})
	}
}

// Mock implementations for testing
type mockStorage struct {
	StorageSchema
}

type mockValidator struct {
	ValidatorSchema
}

type mockJetStream struct {
	JetStream
}
