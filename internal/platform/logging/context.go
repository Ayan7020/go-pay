package logging

import (
	"context"
	"log/slog"
	"slices"
)

type contextHandler struct {
	slog.Handler
}

type attrStore struct{}

func With(ctx context.Context, attr ...slog.Attr) context.Context {
	prev, _ := ctx.Value(attrStore{}).([]slog.Attr)
	return context.WithValue(ctx, attrStore{}, slices.Concat(prev, attr))
}

func (c contextHandler) Handle(ctx context.Context, rec slog.Record) error {
	if attrs, ok := ctx.Value(attrStore{}).([]slog.Attr); ok {
		rec.AddAttrs(attrs...)
	}

	return c.Handler.Handle(ctx, rec)
}

// WithAttrs: .With override our custom contextHandler with new Slog.handler
func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{h.Handler.WithAttrs(attrs)}
}

// WithGroup: .With override our custom contextHandler with new Slog.handler
func (h contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{h.Handler.WithGroup(name)}
}
