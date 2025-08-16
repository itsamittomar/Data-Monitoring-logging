package prometheus

import (
	"sync"
	"time"
	"awesomeProject6/internal/models"
)

type MetricCollector struct {
	metrics map[string]*MetricSeries
	mutex   sync.RWMutex
}

type MetricSeries struct {
	Name     string
	Type     string
	Values   []DataPoint
	Labels   map[string]string
	mutex    sync.RWMutex
}

type DataPoint struct {
	Value     float64
	Timestamp time.Time
}

func NewMetricCollector() *MetricCollector {
	return &MetricCollector{
		metrics: make(map[string]*MetricSeries),
	}
}

func (mc *MetricCollector) RecordMetric(metric models.Metric) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	key := mc.generateKey(metric.Name, metric.Labels)
	
	series, exists := mc.metrics[key]
	if !exists {
		series = &MetricSeries{
			Name:   metric.Name,
			Type:   metric.Type,
			Labels: metric.Labels,
			Values: make([]DataPoint, 0),
		}
		mc.metrics[key] = series
	}

	series.mutex.Lock()
	series.Values = append(series.Values, DataPoint{
		Value:     metric.Value,
		Timestamp: metric.Timestamp,
	})
	
	mc.pruneOldValues(series)
	series.mutex.Unlock()
}

func (mc *MetricCollector) GetMetrics(name string, labels map[string]string) (*MetricSeries, bool) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	key := mc.generateKey(name, labels)
	series, exists := mc.metrics[key]
	return series, exists
}

func (mc *MetricCollector) GetAllMetrics() map[string]*MetricSeries {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	result := make(map[string]*MetricSeries)
	for key, series := range mc.metrics {
		result[key] = series
	}
	return result
}

func (mc *MetricCollector) generateKey(name string, labels map[string]string) string {
	key := name
	for k, v := range labels {
		key += "__" + k + "=" + v
	}
	return key
}

func (mc *MetricCollector) pruneOldValues(series *MetricSeries) {
	cutoff := time.Now().Add(-24 * time.Hour)
	var pruned []DataPoint
	
	for _, dp := range series.Values {
		if dp.Timestamp.After(cutoff) {
			pruned = append(pruned, dp)
		}
	}
	
	series.Values = pruned
}