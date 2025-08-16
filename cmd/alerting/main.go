package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"awesomeProject6/internal/config"
	"awesomeProject6/internal/models"
	"awesomeProject6/pkg/prometheus"
	"awesomeProject6/pkg/rules"
)

func main() {
	logger := logrus.New()

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	collector := prometheus.NewMetricCollector()
	aggregator := prometheus.NewAggregator(collector)
	alertChan := make(chan models.Alert, 100)

	engine := rules.NewEngine(aggregator, alertChan)

	alertRules, err := loadAlertRules(cfg.Alerting.RulesPath)
	if err != nil {
		logger.Fatalf("Failed to load alert rules: %v", err)
	}

	engine.LoadRules(alertRules)

	interval, err := time.ParseDuration(cfg.Alerting.CheckInterval)
	if err != nil {
		logger.Fatalf("Invalid check interval: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go engine.Start(ctx, interval)

	go func() {
		for alert := range alertChan {
			handleAlert(alert, logger)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("Alerting system started")
	<-sigChan

	logger.Info("Shutting down alerting system...")
	cancel()
	close(alertChan)

	logger.Info("Alerting system shutdown complete")
}

func loadAlertRules(path string) ([]models.AlertRule, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var rules []models.AlertRule
	err = json.Unmarshal(data, &rules)
	if err != nil {
		return nil, err
	}

	return rules, nil
}

func handleAlert(alert models.Alert, logger *logrus.Logger) {
	logger.WithFields(logrus.Fields{
		"rule":      alert.Rule.Name,
		"value":     alert.Value,
		"threshold": alert.Rule.Threshold,
		"status":    alert.Status,
	}).Warn("Alert triggered")
}