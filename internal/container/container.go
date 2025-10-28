package container

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/rodrigues-daniel/schema-registry/internal/config"
)

type Container struct {
	Config *config.Config
	DB     *sql.DB
}

func NewContainer(cfg *config.Config) (*Container, error) {
	//  Database connection
	db, err := connectDB(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("falhas ao tentar conectar-se ao banco de dados: %w", err)
	}

	return &Container{
		Config: cfg,
		DB:     db,
	}, nil
}

func (c *Container) Close() {
	if c.DB != nil {
		c.DB.Close()
	}
}

func connectDB(cfg config.DatabaseConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
