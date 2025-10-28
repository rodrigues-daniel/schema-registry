package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	App      AppConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type ServerConfig struct {
	Port         int
	ReadTimeout  int
	WriteTimeout int
}

type AppConfig struct {
	Environment    string
	SchemaCacheTTL int // minutos
}

func Load() *Config {

	// Carrega o arquivo .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar o arquivo .env: %s", err)
	}

	return &Config{
		Database: DatabaseConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     getEnvAsInt("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
			SSLMode:  os.Getenv("DB_SSL_MODE"),
		},
		Server: ServerConfig{
			Port:         getEnvAsInt("SERVER_PORT"),
			ReadTimeout:  getEnvAsInt("READ_TIMEOUT"),
			WriteTimeout: getEnvAsInt("WRITE_TIMEOUT"),
		},
		App: AppConfig{
			Environment:    os.Getenv("APP_ENV"),
			SchemaCacheTTL: getEnvAsInt("SCHEMA_CACHE_TTL"),
		},
	}
}

func getEnvAsInt(key string) int {
	value := os.Getenv(key)
	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	return 0
}
