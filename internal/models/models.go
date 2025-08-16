package models

import "time"

type LogEntry struct {
	Timestamp  time.Time         `json:"timestamp"`
	Level      string           `json:"level"`
	Message    string           `json:"message"`
	Service    string           `json:"service"`
	Host       string           `json:"host"`
	Tags       map[string]string `json:"tags,omitempty"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
}

type Metric struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Timestamp time.Time         `json:"timestamp"`
	Labels    map[string]string `json:"labels"`
	Type      string           `json:"type"`
}

type AlertRule struct {
	Name        string            `json:"name"`
	Query       string           `json:"query"`
	Threshold   float64          `json:"threshold"`
	Operator    string           `json:"operator"`
	Duration    string           `json:"duration"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

type Alert struct {
	Rule        AlertRule         `json:"rule"`
	Value       float64          `json:"value"`
	Timestamp   time.Time        `json:"timestamp"`
	Status      string           `json:"status"`
	Labels      map[string]string `json:"labels"`
}