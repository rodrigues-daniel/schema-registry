-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS schemas (
    id SERIAL PRIMARY KEY,
    subject TEXT NOT NULL,
    version INT NOT NULL,
    schema TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_subject_version ON schemas(subject, version);

-- Tabela para modo de compatibilidade
CREATE TABLE IF NOT EXISTS schema_compatibility (
    subject TEXT PRIMARY KEY,
    compatibility_mode TEXT NOT NULL DEFAULT 'BACKWARD'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Primeiro remove as constraints e índices explicitamente
DROP INDEX IF EXISTS idx_subject_version;

-- Depois as tabelas
DROP TABLE IF EXISTS schema_compatibility;
DROP TABLE IF EXISTS schemas;
-- +goose StatementEnd
