package middlewares

import "log/slog"

type Middleware struct {
	Logger *slog.Logger
}
