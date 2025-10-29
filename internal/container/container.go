package container

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/rodrigues-daniel/schema-registry/internal/config"
	"github.com/rodrigues-daniel/schema-registry/internal/handlers"
	"github.com/rodrigues-daniel/schema-registry/internal/repositories"
	"github.com/rodrigues-daniel/schema-registry/internal/services"
	"github.com/rodrigues-daniel/schema-registry/internal/validation"
)

type Container struct {
	Config *config.Config
	DB     *sql.DB
	// UserRepo      domain.UserRepository
	SchemaRepo repositories.SchemaRepository
	Validator  *validation.DatabaseSchemaValidator
	// UserService   domain.UserService
	// SchemaService *services.SchemaService
	// UserHandler   *handlers.UserHandler
	SchemaHandler *handlers.SchemaHandler
}

func NewContainer(cfg *config.Config) (*Container, error) {

	db, err := connectDB(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("falhas ao tentar conectar-se ao banco de dados: %w", err)
	}

	// userRepo := repositories.NewUserRepository(db)
	schemaRepo := repositories.NewSchemaRepository(db)

	//Validator
	validator := validation.NewDatabaseSchemaValidator(
		schemaRepo,
		time.Duration(cfg.App.SchemaCacheTTL)*time.Minute,
	)

	// Initial schema load
	// if err := validator.Reload(context.Background()); err != nil {
	// 	log.Printf("Warning: falha ao carregar schemas: %v", err)
	// }

	if err := validator.ReloadSchemas(); err != nil {
		log.Printf("Warning: falha ao carregar schemas: %v", err)
	}

	// userService := services.NewUserService(userRepo, validator)
	schemaService := services.NewSchemaService(schemaRepo, validator)

	// userHandler := handlers.NewUserHandler(userService)
	schemaHandler := handlers.NewSchemaHandler(schemaService)

	return &Container{
		Config: cfg,
		DB:     db,
		// UserRepo:      userRepo,
		SchemaRepo: schemaRepo,
		Validator:  validator,
		// UserService:   userService,
		// SchemaService: schemaService,
		// UserHandler:   userHandler,
		SchemaHandler: schemaHandler,
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
