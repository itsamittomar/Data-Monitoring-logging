package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"awesomeProject6/internal/config"
	"awesomeProject6/internal/models"
	"awesomeProject6/pkg/elasticsearch"
	"awesomeProject6/pkg/kafka"
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

	logChan := make(chan models.LogEntry, 1000)

	consumer, err := kafka.NewConsumer(
		cfg.Kafka.Brokers,
		"log-ingestion-group",
		[]string{cfg.Kafka.Topic},
		logChan,
	)
	if err != nil {
		logger.Fatalf("Failed to create Kafka consumer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := consumer.Start(ctx); err != nil {
			logger.Errorf("Kafka consumer error: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		batchProcessor(ctx, esClient, logChan, logger)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("Shutting down...")

	cancel()
	consumer.Close()
	wg.Wait()

	logger.Info("Shutdown complete")
}

func batchProcessor(ctx context.Context, esClient *elasticsearch.Client, logChan chan models.LogEntry, logger *logrus.Logger) {
	batch := make([]models.LogEntry, 0, 100)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				if err := esClient.BulkIndexLogs(ctx, batch); err != nil {
					logger.Errorf("Failed to index final batch: %v", err)
				}
			}
			return

		case log := <-logChan:
			batch = append(batch, log)
			if len(batch) >= 100 {
				if err := esClient.BulkIndexLogs(ctx, batch); err != nil {
					logger.Errorf("Failed to index batch: %v", err)
				}
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				if err := esClient.BulkIndexLogs(ctx, batch); err != nil {
					logger.Errorf("Failed to index timed batch: %v", err)
				}
				batch = batch[:0]
			}
		}
	}
}