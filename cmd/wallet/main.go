package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ayan7020/go-pay/internal/platform/config"
	"github.com/Ayan7020/go-pay/internal/platform/httpx"
	"github.com/Ayan7020/go-pay/internal/platform/logging"
	"github.com/Ayan7020/go-pay/internal/platform/postgres"
	"github.com/Ayan7020/go-pay/internal/wallet"
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

	slog.SetDefault(logger)
	logger.Info("starting", "env", c.Env, "log_level", c.LogLevel.String())

	pool, err := postgres.NewPool(context.Background(), c.DatabaseURL, int32(c.DBMaxConns))
	if err != nil {
		logger.Error("connect to postgres", "err", err)
		os.Exit(1)
	}

	logger.Info("connected to postgres", "max_conns", pool.Config().MaxConns)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		logger.InfoContext(r.Context(), "Server Healthy")
		httpx.WriteJson(w, r, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			logger.WarnContext(ctx, "not ready: database unreachable", "err", err)
			httpx.WriteJson(w, r, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
		w.Write([]byte(`{"status":"ready"}` + "\n"))
	})

	svc := wallet.NewService(wallet.NewStore(pool))
	wallet.NewHandler(svc).Register(mux)

	mux.HandleFunc("/", httpx.NotFound)

	srv := &http.Server{
		Addr:              c.HTTPAddr,
		Handler:           httpx.Chain(mux, httpx.RequestId, httpx.AccessLog(logger)),
		ReadHeaderTimeout: 5 * time.Second,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelWarn),
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
