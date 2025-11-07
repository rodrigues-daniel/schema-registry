-- +goose Up
-- +goose StatementBegin

-- Tabela de Subjects (tópicos/entidades)
CREATE TABLE IF NOT EXISTS subjects (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    version INTEGER NOT NULL,
    schema_definition JSONB NOT NULL,
    schema_type VARCHAR(50) DEFAULT 'JSON',
    description TEXT,
    compatibility_level VARCHAR(50) DEFAULT 'BACKWARD',
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
 
-- Índices
CREATE INDEX IF NOT EXISTS idx_subjects_name ON subjects(name);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin

-- Remover trigger e função
DROP TRIGGER IF EXISTS update_subjects_updated_at ON subjects;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Remover índices
DROP INDEX IF EXISTS idx_subjects_name;

-- Remover tabelas (ordem importa por causa das FK)
DROP TABLE IF EXISTS subjects CASCADE;

-- +goose StatementEnd