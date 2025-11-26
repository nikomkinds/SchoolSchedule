package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"os"
)

// Config structure describes connection params
type Config struct {
	DBHost    string `mapstructure:"DB_HOST" validate:"required"`
	DBPort    string `mapstructure:"DB_PORT" validate:"required"`
	User      string `mapstructure:"DB_USER" validate:"required"`
	Password  string `mapstructure:"DB_PASSWORD" validate:"required"`
	DBName    string `mapstructure:"DB_NAME" validate:"required"`
	SSLMode   string `mapstructure:"DB_SSLMODE" validate:"required"`
	ServHost  string `mapstructure:"SERVER_HOST" validate:"required"`
	ServPort  string `mapstructure:"SERVER_PORT" validate:"required"`
	JWTSecret string `mapstructure:"JWT_SECRET" validate:"required"`
}

// LoadConfig function gets params from the environment
func LoadConfig() (*Config, error) {

	cfg := Config{
		DBHost:    os.Getenv("DB_HOST"),
		DBPort:    os.Getenv("DB_PORT"),
		User:      os.Getenv("DB_USER"),
		Password:  os.Getenv("DB_PASSWORD"),
		DBName:    os.Getenv("DB_NAME"),
		SSLMode:   os.Getenv("DB_SSLMODE"),
		ServHost:  os.Getenv("SERVER_HOST"),
		ServPort:  os.Getenv("SERVER_PORT"),
		JWTSecret: os.Getenv("JWT_SECRET"),
	}

	if err := validator.New().Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}
