package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"url-shortener/internal/app"
	"url-shortener/internal/config"
	"url-shortener/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
)

func main() {
	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)
	m.Info.With(prometheus.Labels{"version": "1.0.0"}).Set(1)
	notUseDatabase := flag.Bool("d", false, "Don`t use PostgreSQL to store data")
	cfg, err := config.InitConfig()
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	logger := app.InitLogger()

	application, err := app.InitApp(cfg, logger, m, notUseDatabase)
	if err != nil {
		logger.Error("bad configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer application.Shutdown()

	pMux := http.NewServeMux()
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	pMux.Handle("/metrics", promHandler)

	eg, ctx := errgroup.WithContext(context.Background())
	sigQuit := make(chan os.Signal, 1)
	signal.Notify(sigQuit, syscall.SIGINT, syscall.SIGTERM)

	eg.Go(func() error {
		return application.Server.Run(ctx)
	})

	eg.Go(func() error {
		return (http.ListenAndServe(":8081", pMux))
	})

	eg.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case s := <-sigQuit:
			logger.Info("Captured signal", slog.String("signal", s.String()))
			return fmt.Errorf("captured signal: %v", s)
		}
	})

	err = eg.Wait()
	logger.Info("Gracefully shutting down the servers", slog.String("error", err.Error()))
}
