package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"awesomeProject6/internal/config"
	"awesomeProject6/pkg/prometheus"
)

func main() {
	logger := logrus.New()

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	collector := prometheus.NewMetricCollector()
	aggregator := prometheus.NewAggregator(collector)

	router := mux.NewRouter()
	
	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/api/metrics", handleMetricsAPI(aggregator))
	router.HandleFunc("/api/query", handleQuery(aggregator))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Metrics.Port),
		Handler: router,
	}

	go func() {
		logger.Infof("Starting metrics server on port %d", cfg.Metrics.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server failed: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("Shutting down metrics server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("Server shutdown error: %v", err)
	}

	logger.Info("Metrics server shutdown complete")
}

func handleMetricsAPI(aggregator *prometheus.Aggregator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "metrics endpoint"}`))
	}
}

func handleQuery(aggregator *prometheus.Aggregator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "query endpoint"}`))
	}
}