package app

import (
	"log/slog"

	"github.com/hsm-gustavo/authentication/internal/config"
	"github.com/hsm-gustavo/authentication/internal/database"
)

type Application struct {
	Logger *slog.Logger
	DB *database.Queries
	Config *config.Config
}