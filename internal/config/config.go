package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl        string
	JWTSecret    string
	OpenAIKey    string
	Port         string
}

// LoadConfig loads environment variables from .env file and returns a Config struct
func LoadConfig() (*Config, error) {
	// Load .env file (it's okay if it doesn't exist in production)
	_ = godotenv.Load()

	cfg := &Config{
		DBUrl:     os.Getenv("DB_URL"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		OpenAIKey: os.Getenv("OPENAI_API_KEY"),
		Port:      os.Getenv("PORT"),
	}

	// Validate required fields
	if cfg.DBUrl == "" {
		return nil, fmt.Errorf("DB_URL environment variable is required")
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}

	// Set default port if not specified
	if cfg.Port == "" {
		cfg.Port = "3000"
	}

	return cfg, nil
}
