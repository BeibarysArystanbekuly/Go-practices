package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	DB_DSN    string
	JWTSecret string
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		Port:      getEnv("APP_PORT", "8080"),
		DB_DSN:    getEnv("DB_DSN", "postgres://polling_user:polling_pass@localhost:5432/polling_db?sslmode=disable"),
		JWTSecret: getEnv("JWT_SECRET", "dev-secret-change-me"),
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
