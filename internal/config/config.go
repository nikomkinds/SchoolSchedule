package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
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

	_ = godotenv.Load(".env")

	viper.AutomaticEnv()

	viper.BindEnv("DB_HOST")
	viper.BindEnv("DB_PORT")
	viper.BindEnv("DB_USER")
	viper.BindEnv("DB_PASSWORD")
	viper.BindEnv("DB_NAME")
	viper.BindEnv("DB_SSLMODE")
	viper.BindEnv("SERVER_HOST")
	viper.BindEnv("SERVER_PORT")
	viper.BindEnv("JWT_SECRET")

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to unmarshal env config: %w", err)
	}

	val := validator.New()
	if err := val.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}
