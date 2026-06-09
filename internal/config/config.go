package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type SMTPConfig struct {
	Host string
	Port string
	Username string
	Password string
	From string
}

type Config struct {
	Port string
	DBURL string
	Env string
	JWTSecret string
	SMTPConfig SMTPConfig
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
		JWTSecret: getEnv("JWT_SECRET", "your-secret-key"),
		SMTPConfig: SMTPConfig{
			Host: getEnv("SMTP_HOST", "smtp.example.com"),
			Port: getEnv("SMTP_PORT", "587"),
			Username: getEnv("SMTP_USERNAME", "user"),
			Password: getEnv("SMTP_PASSWORD", "password"),
			From: getEnv("SMTP_FROM", "noreply@localhost"),
		},
	}
	
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}