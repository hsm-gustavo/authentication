package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	DBURL string
	Env string
}

func Load() Config {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("no .env file found, relying on environment variables")
	} else {
		slog.Info(".env file loaded")
	}

	cfg := Config{
		Port: getEnv("PORT", "8080"),
		DBURL: getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/mydb?sslmode=disable"),
		Env: getEnv("ENV", "development"),
	}
	
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}