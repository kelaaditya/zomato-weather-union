package internal

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// application config
// stores for in-memory access
type AppConfig struct {
	Logger       *slog.Logger
	ENVVariables *AppENVVariables
	DBPool       *pgxpool.Pool
}
