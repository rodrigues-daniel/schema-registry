-- +goose Up
-- +goose StatementBegin
-- Adicionar coluna de deleted_at para soft delete
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE posts ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;

-- Criar índices para deleted_at
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_posts_deleted_at ON posts(deleted_at);

-- Função para atualizar updated_at (opcional)
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers para atualizar updated_at
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_posts_updated_at 
    BEFORE UPDATE ON posts 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_posts_updated_at ON posts;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column;

DROP INDEX IF EXISTS idx_posts_deleted_at;
DROP INDEX IF EXISTS idx_users_deleted_at;

ALTER TABLE posts DROP COLUMN deleted_at;
ALTER TABLE users DROP COLUMN deleted_at;
-- +goose StatementEnd