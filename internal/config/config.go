package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv string

	HTTPAddr string

	PostgresHost     string
	PostgresPort     string
	PostgresDB       string
	PostgresUser     string
	PostgresPassword string
	PostgresSSLMode  string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using system environment variables")
	}

	cfg := &Config{
		AppEnv: getEnv("APP_ENV", "dev"),

		HTTPAddr: getEnv("HTTP_ADDR", ":8080"),

		PostgresHost:     os.Getenv("POSTGRES_HOST"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresDB:       os.Getenv("POSTGRES_DB"),
		PostgresUser:     os.Getenv("POSTGRES_USER"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgresSSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) PostgresDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.PostgresUser,
		c.PostgresPassword,
		c.PostgresHost,
		c.PostgresPort,
		c.PostgresDB,
		c.PostgresSSLMode,
	)
}

func (c *Config) validate() error {
	if c.PostgresHost == "" {
		return fmt.Errorf("POSTGRES_HOST is required")
	}
	if c.PostgresDB == "" {
		return fmt.Errorf("POSTGRES_DB is required")
	}
	if c.PostgresUser == "" {
		return fmt.Errorf("POSTGRES_USER is required")
	}
	if c.PostgresPassword == "" {
		return fmt.Errorf("POSTGRES_PASSWORD is required")
	}

	return nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}