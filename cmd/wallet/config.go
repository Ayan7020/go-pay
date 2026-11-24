package main

import (
	"log/slog"
	"time"

	"github.com/Ayan7020/go-pay/internal/platform/config"
	"github.com/Ayan7020/go-pay/internal/platform/logging"
)

type Config struct {
	Env             string
	HTTPAddr        string
	ShutdownTimeout time.Duration
	LogLevel        slog.Level
	LogFormat       string
	DatabaseURL     string
	DBMaxConns      int
}

func loadConfig(env *config.Env) (Config, error) {
	c := Config{
		Env:             env.OneOf("APP_ENV", "development", "development", "production"),
		HTTPAddr:        env.String("WALLET_HTTP_ADDR", ":8081"),
		ShutdownTimeout: env.Duration("WALLET_SHUTDOWN_TIMEOUT", 10*time.Second),
		LogLevel:        config.Parse(env, "LOG_LEVEL", slog.LevelInfo, logging.ParseLevel),
		DatabaseURL:     env.Required("WALLET_DATABASE_URL"),
		DBMaxConns:      env.Int("WALLET_DB_MAX_CONNS", 10),
	}

	if c.DBMaxConns < 1 || c.DBMaxConns > 100 {
		env.Invalid("WALLET_DB_MAX_CONNS", "must be between 1 and 100")
	}

	format := "text"
	if c.Env == "production" {
		format = "json"
	}

	c.LogFormat = env.OneOf("LOG_FORMAT", format, "text", "json")
	return c, env.Err()
}
