package rules

import (
	"context"
	"strings"
	"time"

	"awesomeProject6/internal/models"
	"awesomeProject6/pkg/prometheus"
	"github.com/sirupsen/logrus"
)

type Engine struct {
	rules      []models.AlertRule
	aggregator *prometheus.Aggregator
	alertChan  chan models.Alert
	logger     *logrus.Logger
}

func NewEngine(aggregator *prometheus.Aggregator, alertChan chan models.Alert) *Engine {
	return &Engine{
		rules:      make([]models.AlertRule, 0),
		aggregator: aggregator,
		alertChan:  alertChan,
		logger:     logrus.New(),
	}
}

func (e *Engine) AddRule(rule models.AlertRule) {
	e.rules = append(e.rules, rule)
}

func (e *Engine) LoadRules(rules []models.AlertRule) {
	e.rules = rules
}

func (e *Engine) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.evaluateRules()
		}
	}
}

func (e *Engine) evaluateRules() {
	for _, rule := range e.rules {
		if alert := e.evaluateRule(rule); alert != nil {
			select {
			case e.alertChan <- *alert:
			default:
				e.logger.Warn("Alert channel is full, dropping alert")
			}
		}
	}
}

func (e *Engine) evaluateRule(rule models.AlertRule) *models.Alert {
	duration, err := time.ParseDuration(rule.Duration)
	if err != nil {
		e.logger.Errorf("Invalid duration in rule %s: %v", rule.Name, err)
		return nil
	}

	value := e.executeQuery(rule.Query, duration)
	threshold := rule.Threshold

	triggered := false
	switch rule.Operator {
	case ">":
		triggered = value > threshold
	case ">=":
		triggered = value >= threshold
	case "<":
		triggered = value < threshold
	case "<=":
		triggered = value <= threshold
	case "==":
		triggered = value == threshold
	case "!=":
		triggered = value != threshold
	default:
		e.logger.Errorf("Unknown operator in rule %s: %s", rule.Name, rule.Operator)
		return nil
	}

	if triggered {
		return &models.Alert{
			Rule:      rule,
			Value:     value,
			Timestamp: time.Now(),
			Status:    "firing",
			Labels:    rule.Labels,
		}
	}

	return nil
}

func (e *Engine) executeQuery(query string, duration time.Duration) float64 {
	parts := strings.Fields(query)
	if len(parts) < 2 {
		return 0
	}

	function := parts[0]
	metricName := parts[1]

	labels := make(map[string]string)
	if len(parts) > 2 {
		for i := 2; i < len(parts); i++ {
			if strings.Contains(parts[i], "=") {
				kv := strings.Split(parts[i], "=")
				if len(kv) == 2 {
					labels[kv[0]] = kv[1]
				}
			}
		}
	}

	switch function {
	case "sum":
		return e.aggregator.Sum(metricName, labels, duration)
	case "avg":
		return e.aggregator.Average(metricName, labels, duration)
	case "max":
		return e.aggregator.Max(metricName, labels, duration)
	case "rate":
		return e.aggregator.Rate(metricName, labels, duration)
	case "p95":
		return e.aggregator.Percentile(metricName, labels, duration, 95)
	case "p99":
		return e.aggregator.Percentile(metricName, labels, duration, 99)
	default:
		return 0
	}
}
