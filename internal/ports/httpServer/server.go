package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"url-shortener/internal/config"
	"url-shortener/pkg/jwt"
	"url-shortener/pkg/metrics"

	"github.com/go-redis/redis_rate/v9"
)

type Server struct {
	server          *http.Server
	logger          *slog.Logger
	shutDownTimeout time.Duration
}

func NewHTTPServer(config *config.ServerConfig, authService ServiceAuth, logger *slog.Logger, serviceURLShortener URLShortenerService, render RepresenrService, limiter *redis_rate.Limiter, metrics *metrics.PrometheusMetrics, manger jwt.TokenManager) (*Server, error) {
	httpHandler := NewHandler(logger, serviceURLShortener, render, metrics)
	authHandler := NewAuthHandler(logger, authService)
	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      InitRouter(httpHandler, authHandler, logger, limiter, manger),
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	return &Server{
		server:          server,
		logger:          logger,
		shutDownTimeout: config.ShutdownTimeout,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	errChan := make(chan error)
	go func() {
		s.logger.Info(fmt.Sprintf("starting listening: %s", s.server.Addr))

		errChan <- s.server.ListenAndServe()
	}()

	var err error
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err = <-errChan:
		
	}
	return err
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutDownTimeout)
	defer cancel()
	err := s.server.Shutdown(ctx)
	if err != nil {
		s.logger.Error("failed to shutdown HTTP Server", slog.String("error", err.Error()))
	}
}
