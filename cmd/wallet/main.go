package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ayan7020/go-pay/internal/platform/config"
	"github.com/Ayan7020/go-pay/internal/platform/logging"
	"github.com/Ayan7020/go-pay/internal/platform/postgres"
)

var version = "dev"

func main() {

	c, err := loadConfig(config.New(os.LookupEnv))
	if err != nil {
		fmt.Fprintln(os.Stderr, "wallet:", err)
		os.Exit(2)
	}

	logger := logging.New(os.Stdout, logging.Options{
		Service: "wallet",
		Version: version,
		Level:   c.LogLevel,
		Format:  c.LogFormat,
	})

	pool, err := postgres.NewPool(context.Background(), c.DatabaseURL, int32(c.DBMaxConns))
	if err != nil {
		logger.Error("connect to postgres", "err", err)
		os.Exit(1)
	}

	logger.Info("connected to postgres", "max_conns", pool.Config().MaxConns)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Health check", "remote", r.RemoteAddr)
		w.Header().Set("Content-type", "application/json")
		w.Write([]byte(`{"status":"OK"}` + "\n"))
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		w.Header().Set("Content-type", "application/json")
		if err := pool.Ping(ctx); err != nil {
			logger.Warn("not ready: database unreachable", "err", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unavailable"}` + "\n"))
			return
		}
		w.Write([]byte(`{"status":"ready"}` + "\n"))
	})

	srv := &http.Server{
		Addr:              ":8081",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("wallet service listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down gracefully")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown failed", "err", err)
	}

	logger.Info("Server Stopped")
}
