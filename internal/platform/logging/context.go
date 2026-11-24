package logging

import "log/slog"

type contextHandler struct {
	slog.Handler
}