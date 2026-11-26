package httpx

import (
	"log/slog"
	"net/http"
	"slices"
	"time"
)

type Middleware func(http.Handler) http.Handler

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func Chain(hh http.Handler, mw ...Middleware) http.Handler {
	h := hh
	for _, m := range slices.Backward(mw) {
		h = m(h)
	}
	return h
}

func (rec *statusRecorder) WriteHeader(status int) {
	if rec.status == 0 {
		rec.status = status
	}
	rec.ResponseWriter.WriteHeader(status)
}

func (rec *statusRecorder) Write(b []byte) (int, error) {
	if rec.status == 0 {
		rec.status = http.StatusOK
	}

	n, err := rec.ResponseWriter.Write(b)
	rec.bytes += n
	return n, err
}

func (rec *statusRecorder) Unwrap() http.ResponseWriter {
	return rec.ResponseWriter
}

func AccessLog(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w}
			
			next.ServeHTTP(rec, r)

			if rec.status == 0 {
				rec.status = http.StatusOK 
			} 

			level := slog.LevelInfo
			if rec.status >= 500 {
				level = slog.LevelError
			}

			logger.LogAttrs(r.Context(), level, "request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Duration("duration", time.Since(start)),
				slog.Int("bytes", rec.bytes),
			)
		})
	}
}
