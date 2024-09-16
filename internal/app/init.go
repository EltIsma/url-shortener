package app

import (
	"log/slog"
	"os"
	"time"
	"url-shortener/internal/adapters/local"
	"url-shortener/internal/adapters/pgrepo"
	"url-shortener/internal/adapters/redis"
	"url-shortener/internal/config"
	httpserver "url-shortener/internal/ports/httpServer"
	"url-shortener/internal/services"
	_ "url-shortener/internal/services/encoder/base62"
	"url-shortener/internal/services/represent"
	"url-shortener/internal/services/uniqueIdGenerator/go-snowflake-master"
	"url-shortener/pkg/database"
	"url-shortener/pkg/jwt"
	"url-shortener/pkg/metrics"
)

type App struct {
	Server   *httpserver.Server
	Postgres *database.Postgres
	Redis    *redis.Redis
}

func InitApp(cfg *config.Config, logger *slog.Logger, metrics *metrics.PrometheusMetrics, noDB *bool) (*App, error) {
	postgres, err := database.NewPG(cfg.Postgres.PostgresURL)
	if err != nil {
		return nil, err
	}

	rds, limiter, err := redis.New(cfg.Redis.Hosts, cfg.Redis.Password, logger)
	if err != nil {
		return nil, err
	}

	snowflake.SetStartTime(time.Date(2021, 9, 1, 0, 0, 0, 0, time.UTC))
	snowflake.SetMachineID(1)
	var serviceURLShortener *services.URLShortener
	if *noDB {
		local := local.New()
		serviceURLShortener = services.New(logger, rds, local)
	} else {
		serviceURLShortener = services.New(logger, rds, pgrepo.NewRepositoruPG(postgres.GetConn()))
	}
	representer := represent.New(cfg.TemplatesPath, logger)

	serviceAuth, err := services.NewAuth(&cfg.Auth, pgrepo.NewRepositoruPG(postgres.GetConn()))
	if err != nil {
		return nil, err
	}
	tokenManager, err := jwt.NewManager(cfg.Auth.JWTSigningKey)
	if err != nil {
		return nil, err
	}

	httpServer, err := httpserver.NewHTTPServer(&cfg.Server, serviceAuth, logger, serviceURLShortener, representer, limiter, metrics, tokenManager)
	if err != nil {
		return nil, err
	}

	return &App{
		Server:   httpServer,
		Postgres: postgres,
		Redis:    rds,
	}, nil

}

func (a *App) Shutdown() {
	a.Server.Stop()
	a.Postgres.Close()
	a.Redis.Close()
}

func InitLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)

	logger := slog.New(handler)
	return logger
}
