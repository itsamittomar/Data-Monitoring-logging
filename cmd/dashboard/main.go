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
	"github.com/sirupsen/logrus"
	"awesomeProject6/internal/config"
	"awesomeProject6/pkg/api"
	"awesomeProject6/pkg/elasticsearch"
	"awesomeProject6/pkg/prometheus"
)

func main() {
	logger := logrus.New()

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	esClient, err := elasticsearch.NewClient(
		cfg.Elasticsearch.URLs,
		cfg.Elasticsearch.Username,
		cfg.Elasticsearch.Password,
		cfg.Elasticsearch.Index,
	)
	if err != nil {
		logger.Fatalf("Failed to create Elasticsearch client: %v", err)
	}

	collector := prometheus.NewMetricCollector()
	aggregator := prometheus.NewAggregator(collector)

	router := mux.NewRouter()
	
	router.Use(corsMiddleware)
	router.Use(loggingMiddleware(logger))

	handlers := api.NewHandlers(esClient, aggregator)
	handlers.SetupRoutes(router)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Dashboard.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Infof("Starting dashboard server on port %d", cfg.Dashboard.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server failed: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("Shutting down dashboard server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("Server shutdown error: %v", err)
	}

	logger.Info("Dashboard server shutdown complete")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(logger *logrus.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			duration := time.Since(start)

			logger.WithFields(logrus.Fields{
				"method":   r.Method,
				"path":     r.URL.Path,
				"duration": duration,
				"remote":   r.RemoteAddr,
			}).Info("HTTP request")
		})
	}
}